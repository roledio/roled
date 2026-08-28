package randstrutil

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomBase64_Length(t *testing.T) {
	tests := []struct {
		n        int
		expected int
	}{
		{0, 0},
		{1, 2},
		{2, 3},
		{3, 4},
		{4, 6},
		{5, 7},
		{6, 8},
		{10, 14},
		{15, 20},
	}

	for _, tt := range tests {
		result := RandomBase64(tt.n)
		assert.Equal(t, tt.expected, len(result), "RandomBase64(%d) should have length %d", tt.n, tt.expected)
	}
}

func TestRandomBase64_Characters(t *testing.T) {
	result := RandomBase64(10)
	validChars := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	assert.Regexp(t, validChars, result, "RandomBase64(10) should only contain valid base64 characters")
}

func TestRandomBase64_Randomness(t *testing.T) {
	result1 := RandomBase64(16)
	result2 := RandomBase64(16)
	assert.NotEqual(t, result1, result2, "RandomBase64(16) should produce different results")
}

func TestRandomBase64_NoPadding(t *testing.T) {
	result := RandomBase64(5)
	assert.NotContains(t, result, "=", "RandomBase64(5) should not contain padding '='")
}

func TestRandomHex_Length(t *testing.T) {
	tests := []struct {
		n        int
		expected int
	}{
		{0, 0},
		{1, 2},
		{2, 4},
		{5, 10},
		{10, 20},
	}

	for _, tt := range tests {
		result := RandomHex(tt.n)
		assert.Len(t, result, tt.expected, "RandomHex(%d) should have length %d", tt.n, tt.expected)
	}
}

func TestRandomHex_Characters(t *testing.T) {
	result := RandomHex(10)
	validChars := regexp.MustCompile(`^[0-9a-f]+$`)
	assert.Regexp(t, validChars, result, "RandomHex(10) should only contain hexadecimal characters")
}

func TestRandomHex_Randomness(t *testing.T) {
	result1 := RandomHex(16)
	result2 := RandomHex(16)
	assert.NotEqual(t, result1, result2, "RandomHex(16) should produce different results")
}
