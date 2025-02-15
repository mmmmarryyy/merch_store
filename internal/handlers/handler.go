package handlers

import (
	"merch_store/internal/auth"
	"merch_store/internal/db"
)

// Handler - abstract for all handlers...
type Handler struct {
	DB             db.DB
	TokenValidator auth.TokenValidator
}

// NewHandler generates Handler...
func NewHandler(db db.DB) *Handler {
	return &Handler{DB: db, TokenValidator: &auth.DefaultValidator{}}
}
