package models

// Inventory - information about user inventory...
type Inventory struct {
	UserID   int `json:"user_id"`
	MerchID  int `json:"merch_id"`
	Quantity int `json:"quantity"`
}

// InventoryInfo contains inventory information that we send inside response to user...
type InventoryInfo struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}
