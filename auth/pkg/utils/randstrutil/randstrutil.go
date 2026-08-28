package randstrutil

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
)

// RandomBase64 generates a random base64-encoded string from n random bytes.
// Output length: approximately (n * 4) / 3 characters (base64 encoding).
// Uses URL-safe base64 encoding without padding.
func RandomBase64(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// RandomHex generates a random hexadecimal string from n random bytes.
// Output length: n * 2 characters (each byte becomes 2 hex characters).
func RandomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
