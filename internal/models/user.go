package models

// User contains information about user of our store...
type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	Coins        int    `json:"coins"`
}
