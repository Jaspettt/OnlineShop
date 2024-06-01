package transservice

type PaymentForm struct {
	CardNumber     string `json:"cardNumber"`
	ExpirationDate string `json:"expirationDate"`
	CVV            string `json:"cvv"`
	Name           string `json:"name"`
	Address        string `json:"address"`
}

type Customer struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CartItem struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type Transaction struct {
	ID        string     `json:"id"`
	Customer  Customer   `json:"customer"`
	CartItems []CartItem `json:"cartItems"`
	Total     float64    `json:"total"`
}
