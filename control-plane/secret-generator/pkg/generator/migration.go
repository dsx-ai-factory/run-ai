/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package generator

import (
	corev1 "k8s.io/api/core/v1"
)

var migrationHandlers = map[string]MigrationHandler{
	"runai-backend-nats": &NatsMigrationHandler{},
}

func getMigrationHandler(legacyName string) MigrationHandler {
	if handler, ok := migrationHandlers[legacyName]; ok {
		return handler
	}
	return &DefaultMigrationHandler{}
}

// DefaultMigrationHandler extracts keys from a legacy secret using LegacyKeyName.
// Missing keys are left for the resolver chain to handle (via chartValue or generation).
type DefaultMigrationHandler struct{}

func (h *DefaultMigrationHandler) Migrate(spec SecretSpec, legacySecret *corev1.Secret, legacyName string) (map[string][]byte, error) {
	secretData := make(map[string][]byte)
	for _, keySpec := range spec.Keys {
		if keySpec.CopyFrom != "" {
			continue
		}

		legacyKey := keySpec.LegacyKeyName
		if legacyKey == "" {
			legacyKey = keySpec.Name
		}

		if value, ok := legacySecret.Data[legacyKey]; ok {
			secretData[keySpec.Name] = value
		}
	}
	return secretData, nil
}
