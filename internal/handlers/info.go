package handlers

import (
	"encoding/json"
	"merch_store/internal/auth"
	"merch_store/internal/models"
	"net/http"
)

func (h *Handler) InfoHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ValidateToken(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.DB.GetUserByUsername(claims.Username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	inventory, err := h.DB.GetUserInventory(user.ID)
	if err != nil {
		http.Error(w, "Failed to get inventory", http.StatusInternalServerError)
		return
	}

	transactions, err := h.DB.GetUserTransactions(user.ID)
	if err != nil {
		http.Error(w, "Failed to get transactions", http.StatusInternalServerError)
		return
	}

	response := models.InfoResponse{
		Coins:       user.Coins,
		Inventory:   inventory,
		CoinHistory: transactions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
