/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package secrets

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// AES-256-GCM key size in bytes. The stored key is a 32-byte printable string
	// (base64-url characters), passed to consumers via an env var and consumed via
	// aes.NewCipher([]byte(env)); the raw ASCII of that string becomes the AES-256
	// key material.
	aesKeySize = 32
	// keyEntropyBytes is the number of raw random bytes drawn from crypto/rand.
	// Base64 (no-padding) encoding expands 3 raw bytes to 4 chars, so 24 raw bytes
	// -> exactly aesKeySize (32) chars. The resulting 32-byte key string carries
	// 24*8 = 192 bits of true entropy, well above what is required to make offline
	// brute force of AES-256-GCM infeasible.
	keyEntropyBytes = 24
	// encryptionKeyOperandName is the name of the encryption key operand
	encryptionKeyOperandName = "encryption-key"
)

// NewEncryptionKeySecretResource creates a new Resource that manages the encryption key secret
func NewEncryptionKeySecretResource(operandName string, k8sReader client.Reader, runaiConfig runv1.RunaiConfig) operands.Resource {
	return &encryptionKeySecretResource{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: &runaiConfig,
	}
}

// encryptionKeySecretResource is a Resource that manages the encryption key secret for the cluster
type encryptionKeySecretResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

// OperandName returns the name of the operand
func (s *encryptionKeySecretResource) OperandName() string {
	return encryptionKeyOperandName
}

// Object returns the encryption key secret
// It creates a new secret if it doesn't exist, or returns the existing secret if it does
func (s *encryptionKeySecretResource) Object(ctx context.Context) (client.Object, error) {
	var (
		name      = constants.RunaiAdsSecretName
		namespace = s.runaiConfig.Namespace
	)

	secret := &corev1.Secret{}

	// Get the existing secret if it exists
	err := s.k8sReader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, secret)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	// If a secret with a valid-length key already exists, keep it. The key never rotates
	// automatically: rotating it would render every already-encrypted personal-credential
	// payload in the runai namespace undecryptable and break every workload that consumes
	// them. Rotation requires an operator-driven re-encryption of existing payloads and is
	// out of scope for this operand.
	if err == nil && len(secret.Data[constants.RunaiAdsEncryptionKeyName]) == aesKeySize {
		return secret, nil
	}

	// Generate a fresh key from a cryptographically secure source. The key MUST NOT be
	// derived from any cluster identifier: ClusterUUID and friends are public (URL path
	// segments, the runaiconfig CR, the CLUSTER_ID env var on the metrics-exporter pod,
	// the CP cluster-list API), so any deterministic derivation from them is equivalent
	// to no key at all against an attacker with read access to encrypted personal-credential
	// payloads.
	//
	// The key is stored as a 32-character base64-url string rather than 32 raw bytes on
	// purpose: it is delivered to consumers via SecretKeyRef -> container env var, and env
	// var values are NUL-terminated at exec() time, so raw random bytes would be truncated.
	// Encoding 24 random bytes yields 32 printable chars with 192 bits preserves the existing
	// consumer contract
	raw := make([]byte, keyEntropyBytes)
	if _, err := rand.Read(raw); err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}
	keyStr := base64.RawURLEncoding.EncodeToString(raw)
	if len(keyStr) != aesKeySize {
		return nil, fmt.Errorf("generated encryption key has unexpected length %d, want %d", len(keyStr), aesKeySize)
	}

	secret.ObjectMeta.Name = name
	secret.ObjectMeta.Namespace = namespace
	secret.Type = corev1.SecretTypeOpaque
	secret.StringData = map[string]string{
		constants.RunaiAdsEncryptionKeyName: keyStr,
	}

	return secret, nil
}

// Status checks if the encryption key secret exists
func (s *encryptionKeySecretResource) Status(ctx context.Context) error {
	secret, err := s.Object(ctx)
	if err != nil {
		return err
	}

	if secret.GetUID() == "" {
		return fmt.Errorf("encryption key secret [%s] is missing", secret.GetName())
	}

	return nil
}

// Enabled returns true if the encryption key secret is enabled
func (s *encryptionKeySecretResource) Enabled() bool {
	return true
}
