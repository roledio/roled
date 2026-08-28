package encryptionutil

import (
	"crypto/sha256"
	"io"

	"golang.org/x/crypto/hkdf"
)

func DeriveKey(masterKey []byte, info string) ([]byte, error) {
	hkdf := hkdf.New(sha256.New, masterKey, nil, []byte(info))
	key := make([]byte, 32)
	if _, err := io.ReadFull(hkdf, key); err != nil {
		return nil, err
	}
	return key, nil
}
