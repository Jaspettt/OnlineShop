package transservice

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func InitDB() {
	var err error
	db, err = sql.Open("mysql", "jaspet:1337@tcp(localhost:3306)/vinylstrans")
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
}

func CloseDB() {
	db.Close()
}
