package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	transservice "project-root/transservice"
)

func main() {
	transservice.InitDB()
	defer transservice.CloseDB()

	r := mux.NewRouter()
	r.HandleFunc("/transactions", transservice.CreateTransaction).Methods("POST")
	r.HandleFunc("/transactions/payment", transservice.HandlePayment).Methods("POST")

	log.Fatal(http.ListenAndServe(":8081", r))
}
