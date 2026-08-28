package pkceutil

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

// VerifyCodeChallenge verifies the code_verifier against the code_challenge using the specified method.
// Currently only supports "S256" method.
func VerifyCodeChallenge(codeVerifier, codeChallenge, method string) bool {
	if method != "S256" {
		return false
	}
	hash := sha256.Sum256([]byte(codeVerifier))
	computedChallenge := base64.RawURLEncoding.EncodeToString(hash[:])
	return strings.EqualFold(computedChallenge, codeChallenge)
}
