// Package auth - package for working with authentication and JWT.
package auth

import (
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

var jwtKey = []byte("your_secret_key")

// Claims contains information about user and time when user's jwt token expires...
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type TokenValidator interface {
	ValidateToken(tokenString string) (*Claims, error)
}

type DefaultValidator struct{}

// GenerateToken generates jwt token...
func GenerateToken(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// ValidateToken checks token is valid...
func (dv *DefaultValidator) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(_ *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
