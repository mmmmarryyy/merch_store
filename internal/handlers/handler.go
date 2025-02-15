package handlers

import (
	"merch_store/internal/db"
)

// Handler - abstract for all handlers...
type Handler struct {
	DB *db.Database
}

// NewHandler generates Handler...
func NewHandler(db *db.Database) *Handler {
	return &Handler{DB: db}
}
