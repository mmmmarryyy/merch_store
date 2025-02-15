package handlers

import (
	"encoding/json"
	"merch_store/internal/auth"
	"merch_store/internal/models"
	"net/http"
)

// SendCoinHandler handles for /api/sendCoin...
func (h *Handler) SendCoinHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ValidateToken(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.SendCoinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fromUser, err := h.DB.GetUserByUsername(claims.Username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	toUser, err := h.DB.GetUserByUsername(req.ToUser)
	if err != nil {
		http.Error(w, "Recipient not found", http.StatusNotFound)
		return
	}

	if fromUser.Coins < req.Amount {
		http.Error(w, "Insufficient coins", http.StatusBadRequest)
		return
	}

	err = h.DB.TransferCoins(fromUser.ID, toUser.ID, req.Amount)
	if err != nil {
		http.Error(w, "Failed to transfer coins", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
