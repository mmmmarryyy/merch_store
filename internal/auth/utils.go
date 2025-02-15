package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// CheckPasswordHash checks password is correct...
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
