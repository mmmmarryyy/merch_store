// Package handlers - package for processing HTTP requests.
package handlers

import (
	"encoding/json"
	"net/http"

	"merch_store/internal/auth"
	"merch_store/internal/models"

	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles /api/auth...
func (h *Handler) AuthHandler(w http.ResponseWriter, r *http.Request) {
	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.DB.GetUserByUsername(req.Username)
	if err != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}

		user = &models.User{
			Username:     req.Username,
			PasswordHash: string(hashedPassword),
			Coins:        1000,
		}

		err = h.DB.CreateUser(user)
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
	} else if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(user.Username)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(models.AuthResponse{Token: token})
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
