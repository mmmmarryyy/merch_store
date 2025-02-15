// Package models - package for defining data models
package models

// AuthRequest - Request of /api/auth...
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse - Response of /api/auth...
type AuthResponse struct {
	Token string `json:"token"`
}

// TODO: move structs below to other files

// InfoResponse - Response of /api/info...
type InfoResponse struct {
	Coins       int             `json:"coins"`
	Inventory   []InventoryInfo `json:"inventory"`
	CoinHistory CoinHistory     `json:"coinHistory"`
}

// CoinHistory - History of user transactions...
type CoinHistory struct {
	Received []TransactionInfo `json:"received"`
	Sent     []TransactionInfo `json:"sent"`
}

// SendCoinRequest - request of /api/sendCoin...
type SendCoinRequest struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}
