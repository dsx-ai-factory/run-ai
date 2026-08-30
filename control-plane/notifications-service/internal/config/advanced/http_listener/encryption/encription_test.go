package encryption

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var encryptionKey = []byte("0123456789abcdefghwertol")

func TestCtor_InvalidKey(t *testing.T) {
	assert.Panics(t, func() {
		NewEncrypterDecrypter([]byte("012345"))
	})
}

func TestEncodeDecode(t *testing.T) {
	encryption := NewEncrypterDecrypter(encryptionKey)
	value := "some data to encrypterDecrypterer"
	e, err := encryption.Encrypt(value)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, value, e)
	p, err := encryption.Decrypt(e)
	assert.Nil(t, err)
	assert.Equal(t, value, p)
}

func TestEncrypt_UniqueNonce(t *testing.T) {
	encryption := NewEncrypterDecrypter(encryptionKey)
	value := "my-secret-data"
	e1, err := encryption.Encrypt(value)
	assert.Nil(t, err)
	e2, err := encryption.Encrypt(value)
	assert.Nil(t, err)
	assert.NotEqual(t, e1, e2, "same plaintext must produce different ciphertext (unique nonce)")
}

func TestEncodeDecode_Struct(t *testing.T) {
	raw := map[string]interface{}{
		"field": map[string]interface{}{
			"data": "some_data",
		},
	}
	encryption := NewEncrypterDecrypter(encryptionKey)
	e, err := encryption.EncryptStruct(raw)
	assert.Nil(t, err)
	assert.NotEqual(t, raw, e)
	p, err := encryption.DecryptStruct(e)
	assert.Nil(t, err)
	assert.Equal(t, raw, p)
}
