package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
)

// BuyHandler handles /api/buy/{item}...
func (h *Handler) BuyHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.TokenValidator.ValidateToken(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	item, ok := vars["item"]
	if !ok {
		http.Error(w, "Item not found in URL; Item is required", http.StatusBadRequest)
		return
	}

	user, err := h.DB.GetUserByUsername(claims.Username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	merch, err := h.DB.GetMerchByName(item)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	if user.Coins < merch.Price {
		http.Error(w, "Insufficient coins", http.StatusBadRequest)
		return
	}

	err = h.DB.BuyMerch(user.ID, merch.ID, merch.Price)
	if err != nil {
		http.Error(w, "Failed to buy item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
