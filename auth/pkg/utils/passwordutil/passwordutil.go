package passwordutil

import (
	"github.com/gookit/goutil/x/encodes/hashutil"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword creates a bcrypt hash of a password.
//
// To accommodate passwords longer than bcrypt's 72-byte limit, this function
// first computes a SHA-256 hash of the input password. The resulting 32-byte
// SHA-256 hash is then used as the input for bcrypt.
//
// It returns the generated bcrypt hash as a string and an error if one occurred.
func HashPassword(password string) (string, error) {
	sha := hashutil.HashSum(hashutil.AlgoSHA256, password) // 32 bytes
	bytes, err := bcrypt.GenerateFromPassword(sha, 8)      // Using the cost of 8 to support low spec hardwares
	return string(bytes), err
}

// IsValidPassword compares a plaintext password with a bcrypt hash to see if they match.
//
// It follows the same hashing process as HashPassword by first computing the
// SHA-256 hash of the provided password before comparing it against the given
// bcrypt hash.
//
// It returns true if the password and hash match, and false otherwise.
func IsValidPassword(password, hash string) bool {
	sha := hashutil.HashSum(hashutil.AlgoSHA256, password)
	err := bcrypt.CompareHashAndPassword([]byte(hash), sha)
	return err == nil
}
