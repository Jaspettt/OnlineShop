package main

import (
	"log"
	"net/http"
	"os"
	"task2"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

var limiter = rate.NewLimiter(1, 3)
var logger = logrus.New()

func main() {
	logger.SetFormatter(&logrus.JSONFormatter{})
	logFile, err := os.OpenFile("api.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		logger.Error("Failed to open log file: ", err)
		return
	}
	defer logFile.Close()

	log.SetOutput(logFile)

	r := mux.NewRouter()

	r.Use(mainProject.MethodOverrideMiddleware)
	r.HandleFunc("/", limitHandler(mainProject.MainPageHandler)).Methods("GET")
	r.HandleFunc("/json", limitHandler(mainProject.HandleJSONRequest)).Methods("POST")
	r.HandleFunc("/vinyl", limitHandler(mainProject.CreateVinylHandler)).Methods("POST")
	r.HandleFunc("/vinyl/{id}", limitHandler(mainProject.GetVinylHandler)).Methods("GET")
	r.HandleFunc("/vinyl/{id}", limitHandler(mainProject.UpdateVinylHandler)).Methods("PUT")
	r.HandleFunc("/vinyl/{id}", limitHandler(mainProject.DeleteVinylHandler)).Methods("DELETE")
	r.HandleFunc("/register", mainProject.RegisterHandler).Methods("GET", "POST")
	r.HandleFunc("/login", mainProject.LoginHandler).Methods("GET", "POST")
	r.HandleFunc("/confirm", mainProject.ConfirmHandler).Methods("GET")
	r.HandleFunc("/user", mainProject.AuthMiddleware(mainProject.UserProfileHandler)).Methods("GET")
	r.HandleFunc("/admin", mainProject.AuthMiddleware(mainProject.AdminMiddleware(mainProject.AdminProfileHandler))).Methods("GET")
	r.HandleFunc("/change-password", mainProject.AuthMiddleware(mainProject.ChangePasswordHandler)).Methods("POST")
	r.HandleFunc("/change-email", mainProject.AuthMiddleware(mainProject.ChangeEmailHandler)).Methods("POST")
	r.HandleFunc("/admin/roles", mainProject.AuthMiddleware(mainProject.AdminMiddleware(mainProject.CreateRoleHandler))).Methods("POST")
	r.HandleFunc("/admin/roles/update", mainProject.AuthMiddleware(mainProject.AdminMiddleware(mainProject.UpdateRoleHandler))).Methods("POST")
	r.HandleFunc("/admin/roles/delete", mainProject.AuthMiddleware(mainProject.AdminMiddleware(mainProject.DeleteRoleHandler))).Methods("POST")
	r.HandleFunc("/admin/send-email", mainProject.AuthMiddleware(mainProject.AdminMiddleware(mainProject.SendEmailHandler))).Methods("POST")

	logger.Info("Server listening on port 8080")
	http.ListenAndServe(":8080", r)
}

func limitHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	}
}
