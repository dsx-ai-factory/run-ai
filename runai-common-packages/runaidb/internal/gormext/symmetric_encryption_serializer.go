package gormext

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"reflect"

	"gorm.io/gorm/schema"
)

// SymmetricEncryptionSerializer implements GORM's serializer.Interface interface
type SymmetricEncryptionSerializer struct {
	gcm cipher.AEAD
}

func NewSymmetricEncryptionSerializer(encryptionKey string) (*SymmetricEncryptionSerializer, error) {
	if len(encryptionKey) != 16 && len(encryptionKey) != 24 && len(encryptionKey) != 32 {
		return nil, fmt.Errorf("encryption key must be 16, 24, or 32 bytes, got: %d", len(encryptionKey))
	}
	block, err := aes.NewCipher([]byte(encryptionKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}
	return &SymmetricEncryptionSerializer{gcm: gcm}, nil
}

// Scan decrypts the database value and sets the result on the destination struct field.
func (s *SymmetricEncryptionSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) error {
	if dbValue == nil {
		return nil
	}

	var encrypted string
	switch v := dbValue.(type) {
	case []byte:
		encrypted = string(v)
	case string:
		encrypted = v
	default:
		return fmt.Errorf("unsupported data type for symmetric decryption: %#v", dbValue)
	}

	if encrypted == "" {
		return nil
	}

	plaintext, err := s.decrypt(encrypted)
	if err != nil {
		return fmt.Errorf("failed to decrypt field %s: %w", field.Name, err)
	}

	field.ReflectValueOf(ctx, dst).SetString(plaintext)
	return nil
}

// Value encrypts the struct field value before storing it in the database.
func (s *SymmetricEncryptionSerializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	if fieldValue == nil {
		return nil, nil
	}

	plaintext, ok := fieldValue.(string)
	if !ok {
		return nil, fmt.Errorf("unsupported data type for symmetric encryption: %#v", fieldValue)
	}

	if plaintext == "" {
		return "", nil
	}

	return s.encrypt(plaintext)
}

func (s *SymmetricEncryptionSerializer) encrypt(plaintext string) (string, error) {
	nonce := make([]byte, s.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	ciphertext := s.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *SymmetricEncryptionSerializer) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	nonceSize := s.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, encryptedData := data[:nonceSize], data[nonceSize:]
	plaintext, err := s.gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}
	return string(plaintext), nil
}
