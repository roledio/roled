package models

import "github.com/golang-jwt/jwt/v5"

type JWTClaims struct {
	jwt.RegisteredClaims
	ClientID string `json:"cid,omitempty"`
	UserID   string `json:"uid,omitempty"`
}
