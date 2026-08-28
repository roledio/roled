package jwtutil

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	claims := jwt.MapClaims{
		"sub": "1234567890",
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token, err := GenerateToken(claims, "secret")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestParseToken(t *testing.T) {
	claims := jwt.MapClaims{
		"sub": "1234567890",
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token, _ := GenerateToken(claims, "secret")
	var parsedClaims jwt.MapClaims
	err := ParseToken(token, "secret", &parsedClaims)
	assert.NoError(t, err)
	assert.Equal(t, "1234567890", parsedClaims["sub"])
}

func TestParseToken_InvalidToken(t *testing.T) {
	var parsedClaims jwt.MapClaims
	err := ParseToken("invalid.token.here", "secret", &parsedClaims)
	assert.Error(t, err, "ParseToken should error for invalid token")
}

func TestParseToken_WrongKey(t *testing.T) {
	claims := jwt.MapClaims{
		"sub": "1234567890",
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token, _ := GenerateToken(claims, "secret")
	var parsedClaims jwt.MapClaims
	err := ParseToken(token, "wrongsecret", &parsedClaims)
	assert.Error(t, err, "ParseToken should error for wrong key")
}

func TestParseToken_Expired(t *testing.T) {
	claims := jwt.MapClaims{
		"sub": "1234567890",
		"exp": time.Now().Add(-time.Hour).Unix(), // Expired
	}
	token, _ := GenerateToken(claims, "secret")
	var parsedClaims jwt.MapClaims
	err := ParseToken(token, "secret", &parsedClaims)
	assert.Error(t, err, "ParseToken should error for expired token")
}

func TestParseToken_InvalidClaimsType(t *testing.T) {
	claims := jwt.MapClaims{
		"sub": "1234567890",
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token, _ := GenerateToken(claims, "secret")
	err := ParseToken(token, "secret", nil) // Pass nil claims, should cause error
	assert.Error(t, err, "ParseToken should error for nil claims")
}
