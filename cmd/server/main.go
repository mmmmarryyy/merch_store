package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"merch_store/internal/db"
	"merch_store/internal/handlers"
)

func main() {
	db, err := db.NewDatabase("db", 5432, "postgres", "password", "shop") // TODO: change to environment variables
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	handler := handlers.NewHandler(db)
	r := mux.NewRouter()

	r.HandleFunc("/api/auth", handler.AuthHandler)
	r.HandleFunc("/api/info", handler.InfoHandler)
	r.HandleFunc("/api/sendCoin", handler.SendCoinHandler)
	r.HandleFunc("/api/buy/{item}", handler.BuyHandler)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
