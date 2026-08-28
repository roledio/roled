package encryptionutil

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeriveKey(t *testing.T) {
	tests := []struct {
		name      string
		masterKey []byte
		info      string
		wantErr   bool
	}{
		{
			name:      "valid inputs",
			masterKey: []byte("masterkey"),
			info:      "testinfo",
			wantErr:   false,
		},
		{
			name:      "empty masterKey",
			masterKey: []byte{},
			info:      "testinfo",
			wantErr:   false,
		},
		{
			name:      "empty info",
			masterKey: []byte("masterkey"),
			info:      "",
			wantErr:   false,
		},
		{
			name:      "both empty",
			masterKey: []byte{},
			info:      "",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DeriveKey(tt.masterKey, tt.info)
			if tt.wantErr {
				assert.Error(t, err, "DeriveKey() should error")
			} else {
				assert.NoError(t, err, "DeriveKey() should not error")
				assert.Len(t, got, 32, "DeriveKey() key length should be 32")
			}
		})
	}
}

func TestDeriveKeyDeterministic(t *testing.T) {
	masterKey := []byte("masterkey")
	info := "testinfo"

	key1, err1 := DeriveKey(masterKey, info)
	key2, err2 := DeriveKey(masterKey, info)

	assert.NoError(t, err1, "DeriveKey() unexpected error for key1")
	assert.NoError(t, err2, "DeriveKey() unexpected error for key2")
	assert.True(t, bytes.Equal(key1, key2), "DeriveKey() not deterministic: keys differ")
}

func TestDeriveKeyDifferentInfo(t *testing.T) {
	masterKey := []byte("masterkey")

	key1, _ := DeriveKey(masterKey, "info1")
	key2, _ := DeriveKey(masterKey, "info2")

	assert.False(t, bytes.Equal(key1, key2), "DeriveKey() same key for different info")
}
