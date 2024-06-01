package transservice

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var transaction Transaction
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Calculate total amount
	total := 0.0
	for _, item := range transaction.CartItems {
		total += item.Price
	}
	transaction.Total = total

	// Save transaction to database
	if err := saveTransaction(transaction); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transaction)
}
func saveTransaction(transaction Transaction) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Insert transaction with status
	_, err = tx.Exec("INSERT INTO transactions (id, customer_id, total, status) VALUES (?, ?, ?, ?)", transaction.ID, transaction.Customer.ID, transaction.Total, "awaiting payment")
	if err != nil {
		tx.Rollback()
		return err
	}

	// Insert cart items
	for _, item := range transaction.CartItems {
		_, err = tx.Exec("INSERT INTO cart_items (id, transaction_id, name, price) VALUES (?, ?, ?, ?)", item.ID, transaction.ID, item.Name, item.Price)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
func HandlePayment(w http.ResponseWriter, r *http.Request) {
	var paymentForm PaymentForm
	if err := json.NewDecoder(r.Body).Decode(&paymentForm); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// For testing, we assume the payment is always successful
	// Update transaction status to 'paid'
	transactionID := r.URL.Query().Get("transaction_id")
	if transactionID == "" {
		http.Error(w, "Transaction ID is required", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("UPDATE transactions SET status = ? WHERE id = ?", "paid", transactionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Retrieve the updated transaction details
	var transaction Transaction
	row := db.QueryRow("SELECT id, customer_id, total FROM transactions WHERE id = ?", transactionID)
	if err := row.Scan(&transaction.ID, &transaction.Customer.ID, &transaction.Total); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Retrieve customer details
	row = db.QueryRow("SELECT id, name, email FROM customers WHERE id = ?", transaction.Customer.ID)
	if err := row.Scan(&transaction.Customer.ID, &transaction.Customer.Name, &transaction.Customer.Email); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Retrieve cart items
	rows, err := db.Query("SELECT id, name, price FROM cart_items WHERE transaction_id = ?", transactionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item CartItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Price); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		transaction.CartItems = append(transaction.CartItems, item)
	}

	// Generate the receipt PDF
	receiptPath, err := GenerateReceipt(transaction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send the receipt via email
	err = SendEmal(transaction.Customer.Email, "Your Fiscal Receipt", "Please find attached your fiscal receipt.", receiptPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Notify the main server to update the shopping cart status
	updateCartStatus(transactionID)

	// Mark transaction as completed
	_, err = db.Exec("UPDATE transactions SET status = ? WHERE id = ?", "completed", transactionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Payment successful, receipt sent, and transaction completed"))
}
func updateCartStatus(transactionID string) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("http://main-server-url/cart/update-status?transaction_id=%s", transactionID), nil)
	if err != nil {
		log.Println("Failed to create request to update cart status:", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Failed to update cart status:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Failed to update cart status, status code:", resp.StatusCode)
	} else {
		log.Println("Successfully updated cart status")
	}
}
