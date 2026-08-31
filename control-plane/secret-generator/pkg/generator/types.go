/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package generator

import (
	corev1 "k8s.io/api/core/v1"
)

// MigrationHandler extracts data from a legacy secret during migration.
type MigrationHandler interface {
	Migrate(spec SecretSpec, legacySecret *corev1.Secret, legacyName string) (map[string][]byte, error)
}

type SecretSpec struct {
	Name               string          `json:"name"`
	LegacySecretName   string          `json:"legacySecretName,omitempty"`
	ExistingSecretName string          `json:"existingSecretName,omitempty"`
	Keys               []SecretKeySpec `json:"keys"`
}

// SecretKeySpec defines a single key within a generated secret.
// Type can be "password", "alphanumeric", "hex", "uuid", or "static".
// When Type is "static", the value comes from ChartValue (or legacy migration)
// with no random-generation fallback.
type SecretKeySpec struct {
	Name          string  `json:"name"`
	LegacyKeyName string  `json:"legacyKeyName,omitempty"`
	Length        int     `json:"length,omitempty"`
	Type          string  `json:"type,omitempty"`
	CopyFrom      string  `json:"copyFrom,omitempty"`
	ChartValue    *string `json:"chartValue,omitempty"`
	ExternalOnly  bool    `json:"externalOnly,omitempty"`
}

type GeneratorConfig struct {
	Secrets []SecretSpec `json:"secrets"`
}

type GenerationResult struct {
	SecretName string
	Action     string
	Message    string
	Error      error
}
