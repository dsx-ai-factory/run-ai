/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package encryption

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Encryption Suite")
}

var _ = Describe("Encryption Test", func() {
	Describe("GetEnvOrDie", func() {
		Context("decrypt with non existent key", func() {
			It("should fail", func() {
				keys := TenantEncryptionKeys{ActiveKeyId: 5}
				data := "hello"
				_, err := DecryptOAEPChunks(context.Background(), &keys, 0, &data, "world")
				Expect(err).NotTo(BeNil())
				Expect(err.Message).To(ContainSubstring("cannot decrypt world: key with id 0 is not found"))
			})
		})
		Context("EncryptOAEPChunks & DecryptOAEPChunks", func() {
			It("use old Encrypt function and new DecryptOAEPChunks function ", func() {
				keys := NewTenantEncryptionKeys()

				private, public, _ := generateRSAKeyPair(KeyBitSize)
				keys.ActiveKeyId = 5
				keys.Keys[5] = TenantEncryptionKey{
					Id:      5,
					Private: private,
					Public:  public,
				}

				secretString := testutils.RandString(12)
				label := "long-string"

				encrypted, encErr := Encrypt(context.Background(), keys, &secretString, label)
				Expect(encErr).To(BeNil())

				decrypted, decErr := DecryptOAEPChunks(context.Background(), keys, keys.ActiveKeyId, encrypted, label)
				Expect(decErr).To(BeNil())
				Expect(decrypted).To(Equal(secretString))
			})
			It("should handle large string sizes", func() {
				keys := NewTenantEncryptionKeys()

				private, public, _ := generateRSAKeyPair(KeyBitSize)
				keys.ActiveKeyId = 5
				keys.Keys[5] = TenantEncryptionKey{
					Id:      5,
					Private: private,
					Public:  public,
				}

				longString := testutils.RandString(500)
				label := "long-string"

				encrypted, encErr := EncryptOAEPChunks(context.Background(), keys, &longString, label)
				Expect(encErr).To(BeNil())

				decrypted, decErr := DecryptOAEPChunks(context.Background(), keys, keys.ActiveKeyId, encrypted, label)
				Expect(decErr).To(BeNil())
				Expect(decrypted).To(Equal(longString))
			})
		})
	})
})

// GenerateRSAKeyPair generates an RSA private and public key pair
func generateRSAKeyPair(bitSize int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	// Generate a private key using the specified bit size
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, nil, err
	}

	// Extract the public key from the private key
	publicKey := &privateKey.PublicKey

	// Return both private and public keys
	return privateKey, publicKey, nil
}
