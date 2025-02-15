package models

// Transaction contains information about one user sending coins to another user...
type Transaction struct {
	ID         int    `json:"id"`
	FromUserID int    `json:"from_user_id"`
	ToUserID   int    `json:"to_user_id"`
	Amount     int    `json:"amount"`
	CreatedAt  string `json:"created_at"`
}

// TransactionInfo contains transaction information that we send inside response to user...
type TransactionInfo struct {
	Username string `json:"username"`
	Amount   int    `json:"amount"`
}
