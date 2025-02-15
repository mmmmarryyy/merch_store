package models

type Inventory struct {
	UserID   int `json:"user_id"`
	MerchID  int `json:"merch_id"`
	Quantity int `json:"quantity"`
}

type InventoryInfo struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}
