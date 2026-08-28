package encryptionutil

import (
	"crypto/rand"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptAES_Success(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	assert.NoError(t, err, "failed to generate key")

	value := "hello world"
	info := "test info"

	encrypted, err := EncryptAES(value, key, info)
	assert.NoError(t, err, "EncryptAES failed")
	assert.NotEmpty(t, encrypted, "EncryptAES returned empty string")
}

func TestEncryptAES_InvalidKey(t *testing.T) {
	key := make([]byte, 15) // Invalid key length
	value := "test"
	info := "info"

	_, err := EncryptAES(value, key, info)
	assert.Error(t, err, "EncryptAES should fail with invalid key length")
}

func TestEncryptAES_EmptyValue(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	assert.NoError(t, err, "failed to generate key")

	value := ""
	info := "test"

	encrypted, err := EncryptAES(value, key, info)
	assert.NoError(t, err, "EncryptAES failed with empty value")
	assert.NotEmpty(t, encrypted, "EncryptAES returned empty string for empty value")
}

func TestEncryptAES_EmptyInfo(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	assert.NoError(t, err, "failed to generate key")

	value := "test"
	info := ""

	encrypted, err := EncryptAES(value, key, info)
	assert.NoError(t, err, "EncryptAES failed with empty info")
	assert.NotEmpty(t, encrypted, "EncryptAES returned empty string for empty info")
}
func TestDecryptAES_Success(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	assert.NoError(t, err, "failed to generate key")

	value := "hello world"
	info := "test info"

	encrypted, err := EncryptAES(value, key, info)
	assert.NoError(t, err, "EncryptAES failed")

	decrypted, err := DecryptAES(encrypted, key, info)
	assert.NoError(t, err, "DecryptAES failed")
	assert.Equal(t, value, decrypted, "DecryptAES returned wrong value")
}

func TestDecryptAES_InvalidKey(t *testing.T) {
	key := make([]byte, 15) // Invalid key length
	value := "test"
	info := "info"

	encrypted, _ := EncryptAES(value, make([]byte, 32), info) // Encrypt with valid key
	_, err := DecryptAES(encrypted, key, info)
	assert.Error(t, err, "DecryptAES should fail with invalid key length")
}

func TestDecryptAES_InvalidBase64(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	assert.NoError(t, err, "failed to generate key")

	invalidBase64 := "not-base64!"

	_, err = DecryptAES(invalidBase64, key, "info")
	assert.Error(t, err, "DecryptAES should fail with invalid base64")
}

func TestDecryptAES_TamperedCiphertext(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	assert.NoError(t, err, "failed to generate key")

	value := "test"
	info := "info"

	encrypted, err := EncryptAES(value, key, info)
	assert.NoError(t, err, "EncryptAES failed")

	// Tamper with the encrypted string
	tampered := encrypted[:len(encrypted)-1] + "x"

	_, err = DecryptAES(tampered, key, info)
	assert.Error(t, err, "DecryptAES should fail with tampered ciphertext")
}

func TestDecryptAES_WrongKey(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	_, err := rand.Read(key1)
	assert.NoError(t, err, "failed to generate key1")
	_, err = rand.Read(key2)
	assert.NoError(t, err, "failed to generate key2")

	value := "test"
	info := "info"

	encrypted, err := EncryptAES(value, key1, info)
	assert.NoError(t, err, "EncryptAES failed")

	_, err = DecryptAES(encrypted, key2, info)
	assert.Error(t, err, "DecryptAES should fail with wrong key")
}

func TestDecryptAES_WrongInfo(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	assert.NoError(t, err, "failed to generate key")

	value := "test"
	info1 := "info1"
	info2 := "info2"

	encrypted, err := EncryptAES(value, key, info1)
	assert.NoError(t, err, "EncryptAES failed")

	_, err = DecryptAES(encrypted, key, info2)
	assert.Error(t, err, "DecryptAES should fail with wrong info")
}

func TestDecryptAES_EmptyValue(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	assert.NoError(t, err, "failed to generate key")

	value := ""
	info := "test"

	encrypted, err := EncryptAES(value, key, info)
	assert.NoError(t, err, "EncryptAES failed")

	decrypted, err := DecryptAES(encrypted, key, info)
	assert.NoError(t, err, "DecryptAES failed with empty value")
	assert.Equal(t, value, decrypted, "DecryptAES returned wrong value")
}

func TestDecryptAES_TooShort(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	assert.NoError(t, err, "failed to generate key")

	shortEncrypted := base64.StdEncoding.EncodeToString([]byte("short"))

	_, err = DecryptAES(shortEncrypted, key, "info")
	assert.Error(t, err, "DecryptAES should fail with too short encrypted value")
	assert.Contains(t, err.Error(), "encrypted value is not valid", "Error message should indicate invalid encrypted value")
}
