package mainProject

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Token     string    `json:"token"`
	Confirmed bool      `json:"confirmed"`
	CreatedAt time.Time `json:"created_at"`
}

const (
	dbDriver = "mysql"
	dbUser   = "jaspet"
	dbPass   = "1337"
	dbName   = "vinylsgolang"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, "public/register.html")
		return
	} else if r.Method == "POST" {
		var user User
		user.Username = r.FormValue("username")
		user.Email = r.FormValue("email")
		password := r.FormValue("password")

		if user.Username == "" || user.Email == "" || password == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Println("Error generating hashed password:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		user.Password = string(hashedPassword)

		user.Token, err = generateToken(32)
		if err != nil {
			log.Println("Error generating token:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		dsn := fmt.Sprintf("%s:%s@tcp(localhost:3306)/%s", dbUser, dbPass, dbName)
		db, err := sql.Open(dbDriver, dsn)
		if err != nil {
			log.Println("Error opening database connection:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		query := "INSERT INTO users (username, email, password, token, confirmed) VALUES (?, ?, ?, ?, 0)"
		result, err := db.Exec(query, user.Username, user.Email, user.Password, user.Token)
		if err != nil {
			log.Println("Error executing insert query:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		userID, err := result.LastInsertId()
		if err != nil {

		}
		defaultRoleID := 2
		query = "INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)"
		_, err = db.Exec(query, userID, defaultRoleID)
		if err != nil {
		}
		err = sendConfirmationEmail(user.Email, user.Token)
		if err != nil {
			log.Println("Error sending confirmation email:", err)
			http.Error(w, "Failed to send confirmation email", http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, "Registration successful! Please check your email to confirm your account.")
		return
	}

	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func generateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func sendConfirmationEmail(email, token string) error {
	from := "KossinovViktor@gmail.com"
	password := "swpn salo otir rbev"
	to := email
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Confirm your email address\n\n" +
		"Please click the link to confirm your email address: " +
		"http://localhost:8080/confirm?token=" + token

	auth := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(msg))
	return err
}

func ConfirmHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	dsn := fmt.Sprintf("%s:%s@tcp(localhost:3306)/%s", dbUser, dbPass, dbName)
	db, err := sql.Open(dbDriver, dsn)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE token = ?", token).Scan(&userID)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE users SET confirmed = 1 WHERE id = ?", userID)
	if err != nil {
		http.Error(w, "Failed to confirm email", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login?confirmation=success", http.StatusSeeOther)
}
