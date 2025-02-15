package handlers

import (
	"merch_store/internal/db"
)

type Handler struct {
	DB *db.Database
}

func NewHandler(db *db.Database) *Handler {
	return &Handler{DB: db}
}
