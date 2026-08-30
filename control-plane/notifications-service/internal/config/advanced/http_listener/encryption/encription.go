package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"
)

type EncrypterDecrypterer interface {
	Encrypt(plaintext string) (string, error)
	EncryptStruct(s any) (string, error)
	Decrypt(ciphertext string) (string, error)
	DecryptStruct(ciphertext string) (any, error)
}

type EncrypterDecrypter struct {
	// The AesKey argument should be the AES AesKey,
	// either 16, 24, or 32 bytes to select
	GCM cipher.AEAD
}

func (ed *EncrypterDecrypter) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, ed.GCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	ciphertext := ed.GCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (ed *EncrypterDecrypter) EncryptStruct(s interface{}) (string, error) {
	sbytes, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	return ed.Encrypt(base64.StdEncoding.EncodeToString(sbytes))
}

func (ed *EncrypterDecrypter) Decrypt(ciphertext string) (string, error) {
	encryptedData, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed b64 secret decryption: %w", err)
	}

	nonceSize := ed.GCM.NonceSize()
	if len(encryptedData) < nonceSize {
		return "", errors.New("invalid ciphertext")
	}

	nonce, encryptedData := encryptedData[:nonceSize], encryptedData[nonceSize:]
	plaintext, err := ed.GCM.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func (ed *EncrypterDecrypter) DecryptStruct(ciphertext string) (interface{}, error) {
	decrypted, err := ed.Decrypt(ciphertext)
	if err != nil {
		return "", err
	}

	decryptedb, err := base64.StdEncoding.DecodeString(decrypted)
	if err != nil {
		return "", err
	}

	res := map[string]interface{}{}
	err = json.Unmarshal(decryptedb, &res)
	return res, err
}

func NewEncrypterDecrypter(key []byte) *EncrypterDecrypter {
	if !slices.Contains([]int{16, 24, 32}, len(key)) {
		log.Panic().Msgf("encryption AesKey argument should be the AES AesKey, either 16, 24 or 32 bytes len: %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Panic().Msgf("failed to create new cipher block")
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Panic().Msgf("failed to create new cipher GCM")
	}
	return &EncrypterDecrypter{GCM: gcm}
}
