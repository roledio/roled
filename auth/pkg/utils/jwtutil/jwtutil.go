package jwtutil

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(claims jwt.Claims, key string) (string, error) {
	jwt.MarshalSingleStringAsArray = false
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString([]byte(key))
	return s, err
}

func ParseToken(tokenString string, key string, claims jwt.Claims) error {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(key), nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("invalid token")
	}
	return nil
}
