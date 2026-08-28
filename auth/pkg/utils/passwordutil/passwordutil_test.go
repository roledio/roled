package passwordutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	password := "whatever password"
	hash, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	// The primary validation is whether IsValidPassword can verify the generated hash.
	assert.True(t, IsValidPassword(password, hash))
}

func TestIsValidPassword(t *testing.T) {
	password := "my-super-secret-password-123"
	correctHash, err := HashPassword(password)
	assert.NoError(t, err)

	emptyPasswordHash, err := HashPassword("")
	assert.NoError(t, err)

	testCases := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "ValidPassword",
			password: password,
			hash:     correctHash,
			want:     true,
		},
		{
			name:     "InvalidPassword",
			password: "wrongpassword",
			hash:     correctHash,
			want:     false,
		},
		{
			name:     "EmptyPasswordWithItsCorrectHash",
			password: "",
			hash:     emptyPasswordHash,
			want:     true,
		},
		{
			name:     "ValidPasswordWithInvalidHashFormat",
			password: password,
			hash:     "this-is-not-a-bcrypt-hash",
			want:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsValidPassword(tc.password, tc.hash)
			assert.Equal(t, tc.want, got)
		})
	}
}
