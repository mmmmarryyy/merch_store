package models

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type InfoResponse struct {
	Coins       int         `json:"coins"`
	Inventory   []Inventory `json:"inventory"` // TODO: change from []Inventory to []InventoryInfo or something more readable
	CoinHistory CoinHistory `json:"coinHistory"`
}

type CoinHistory struct {
	Received []Transaction `json:"received"` // TODO: change from Transaction to something more readable like at API
	Sent     []Transaction `json:"sent"`     // TODO: change from Transaction to something more readable like at API
}

type SendCoinRequest struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}
