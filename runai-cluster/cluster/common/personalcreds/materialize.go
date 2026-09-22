/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package personalcreds

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CopyLabel               = "run.ai/personal-credential-copy"
	UserCredentialsLabel    = "run.ai/user-credentials"
	AssetLabel              = "run.ai/asset"
	GenericAssetKind        = "generic"
	DockerRegistryAssetKind = "docker-registry"
)

// BuildDestination constructs a decrypted destination Secret without performing
// Kubernetes reads or writes. Controller-specific ownership is added by callers.
func BuildDestination(name, namespace string, source *corev1.Secret, key []byte) (*corev1.Secret, error) {
	if len(source.Data) == 0 {
		return nil, fmt.Errorf("source secret %s/%s has no data", source.Namespace, source.Name)
	}

	dest := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      cloneMap(source.Labels),
			Annotations: cloneMap(source.Annotations),
		},
		Type: source.Type,
		Data: make(map[string][]byte, len(source.Data)),
	}
	if dest.Labels == nil {
		dest.Labels = make(map[string]string)
	}
	delete(dest.Labels, UserCredentialsLabel)
	delete(dest.Labels, AssetLabel)
	dest.Labels[CopyLabel] = "true"
	if source.Labels[AssetLabel] == DockerRegistryAssetKind {
		dest.Type = corev1.SecretTypeDockerConfigJson
	}

	for dataKey, encrypted := range source.Data {
		decrypted, err := DecryptData(encrypted, key)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt key %s: %w", dataKey, err)
		}
		dest.Data[dataKey] = decrypted
	}
	return dest, nil
}

// DecryptData decrypts nonce-prefixed AES-GCM ciphertext.
func DecryptData(encryptedData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}
	if len(encryptedData) < gcm.NonceSize() {
		return nil, fmt.Errorf("encrypted data too short")
	}
	plaintext, err := gcm.Open(nil, encryptedData[:gcm.NonceSize()], encryptedData[gcm.NonceSize():], nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}
	return plaintext, nil
}

func cloneMap(source map[string]string) map[string]string {
	if len(source) == 0 {
		return nil
	}
	dest := make(map[string]string, len(source))
	for key, value := range source {
		dest[key] = value
	}
	return dest
}
