package models

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

// TODO: move structs below to other files
type InfoResponse struct {
	Coins       int             `json:"coins"`
	Inventory   []InventoryInfo `json:"inventory"`
	CoinHistory CoinHistory     `json:"coinHistory"`
}

type CoinHistory struct {
	Received []TransactionInfo `json:"received"`
	Sent     []TransactionInfo `json:"sent"`
}

type SendCoinRequest struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}
