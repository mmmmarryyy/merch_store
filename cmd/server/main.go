// Main package
package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"

	"merch_store/internal/db"
	"merch_store/internal/handlers"
)

func main() {
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		log.Fatalf("Wrong port: %v", err)
	}

	db, err := db.NewDatabase(os.Getenv("DB_HOST"), port, os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	handler := handlers.NewHandler(db)
	r := mux.NewRouter()

	r.HandleFunc("/api/auth", handler.AuthHandler)
	r.HandleFunc("/api/info", handler.InfoHandler)
	r.HandleFunc("/api/sendCoin", handler.SendCoinHandler)
	r.HandleFunc("/api/buy/{item}", handler.BuyHandler)

	log.Println("Server started on :8080")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Panicf("Failed in server: %v", err)
	}
}
