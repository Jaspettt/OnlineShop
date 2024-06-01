package transservice

import (
	"fmt"
	"gopkg.in/gomail.v2"

	"github.com/jung-kurt/gofpdf/v2"
	"time"
)

func GenerateReceipt(transaction Transaction) (string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Your Project Name")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Transaction Number: %s", transaction.ID))
	pdf.Ln(10)
	pdf.Cell(40, 10, fmt.Sprintf("Date and Time: %s", time.Now().Format("2006-01-02 15:04:05")))
	pdf.Ln(10)
	pdf.Cell(40, 10, fmt.Sprintf("Customer: %s", transaction.Customer.Name))
	pdf.Ln(10)
	pdf.Cell(40, 10, fmt.Sprintf("Email: %s", transaction.Customer.Email))
	pdf.Ln(10)

	pdf.Cell(40, 10, "Items:")
	pdf.Ln(10)

	for _, item := range transaction.CartItems {
		pdf.Cell(40, 10, fmt.Sprintf("Product: %s, Unit Price: %.2f, Quantity: 1, Total: %.2f", item.Name, item.Price, item.Price))
		pdf.Ln(10)
	}

	pdf.Cell(40, 10, fmt.Sprintf("Total Amount: %.2f", transaction.Total))
	pdf.Ln(10)
	pdf.Cell(40, 10, "Payment Method: Card")

	// Save PDF
	receiptPath := fmt.Sprintf("receipts/receipt_%s.pdf", transaction.ID)
	err := pdf.OutputFileAndClose(receiptPath)
	if err != nil {
		return "", err
	}

	return receiptPath, nil
}
func SendEmal(to string, subject string, body string, attachmentPath string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "kossinovviktor@gmail.com")
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)
	m.Attach(attachmentPath)

	d := gomail.NewDialer("smtp.example.com", 587, "your-email@example.com", "swpn salo otir rbev")

	return d.DialAndSend(m)
}
