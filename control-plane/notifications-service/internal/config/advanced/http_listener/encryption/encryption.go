/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"
)

const (
	gcmNonceSize = 12
	gcmTagSize   = 16
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
	ciphertext := ed.GCM.Seal(nil, nil, []byte(plaintext), nil)
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

	if len(encryptedData) < gcmNonceSize+gcmTagSize {
		return "", errors.New("invalid ciphertext")
	}

	plaintext, err := ed.GCM.Open(nil, nil, encryptedData, nil)
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
	gcm, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		log.Panic().Msgf("failed to create new cipher GCM")
	}
	return &EncrypterDecrypter{GCM: gcm}
}
