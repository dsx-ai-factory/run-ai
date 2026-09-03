/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/secret-generator/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	maxRetries     = 3
	retryBaseDelay = 200 * time.Millisecond
)

type SecretGenerator struct {
	client client.Client
	config *config.ConfigData
}

func NewSecretGenerator(k8sClient client.Client, cfg *config.ConfigData) *SecretGenerator {
	return &SecretGenerator{
		client: k8sClient,
		config: cfg,
	}
}

func (g *SecretGenerator) Run(ctx context.Context) error {
	logging.Infof("Starting secret generator in namespace: %s", g.config.Namespace)

	generatorConfig, err := g.loadGeneratorConfig()
	if err != nil {
		return fmt.Errorf("failed to load generator config: %w", err)
	}
	logging.Infof("Loaded configuration with %d secret specifications", len(generatorConfig.Secrets))

	results := make([]GenerationResult, 0, len(generatorConfig.Secrets))
	for _, spec := range generatorConfig.Secrets {
		result := g.ProcessSecretSpec(ctx, spec)
		results = append(results, result)

		if result.Error != nil {
			if ctx.Err() != nil {
				logging.Errorf("Context cancelled while processing secret %s: %v (context: %v)",
					spec.Name, result.Error, ctx.Err())
				break
			}
			logging.Errorf("Failed to process secret %s: %v", spec.Name, result.Error)
		} else {
			logging.Infof("Secret %s: %s - %s", result.SecretName, result.Action, result.Message)
		}
	}

	g.reportSummary(results)

	var errs []error
	for _, r := range results {
		if r.Error != nil {
			errs = append(errs, fmt.Errorf("secret %s: %w", r.SecretName, r.Error))
		}
	}
	if len(errs) > 0 {
		if len(errs) == 1 {
			return errs[0]
		}
		return fmt.Errorf("secret generation completed with %d errors: %v", len(errs), errs)
	}

	logging.Infof("Secret generation completed successfully")
	return nil
}

func (g *SecretGenerator) loadGeneratorConfig() (*GeneratorConfig, error) {
	configData, err := os.ReadFile(g.config.ConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", g.config.ConfigFilePath, err)
	}

	var cfg GeneratorConfig
	if err := json.Unmarshal(configData, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	return &cfg, nil
}

// ProcessSecretSpec handles a single secret specification.
func (g *SecretGenerator) ProcessSecretSpec(ctx context.Context, spec SecretSpec) GenerationResult {
	result := GenerationResult{SecretName: spec.Name}

	if spec.ExistingSecretName != "" && spec.ExistingSecretName != spec.Name {
		return g.handleExistingSecret(ctx, spec)
	}

	genSecret, err := g.getSecret(ctx, spec.Name)
	if err != nil && !errors.IsNotFound(err) {
		result.Error = fmt.Errorf("failed to check existing secret: %w", err)
		return result
	}

	if genSecret != nil {
		missingKeys := findMissingKeys(spec.Keys, genSecret.Data)
		if len(missingKeys) == 0 {
			result.Action = "preserved"
			result.Message = "generator-managed secret already exists"
			return result
		}
		legacyData, err := g.extractLegacyData(ctx, spec)
		if err != nil {
			result.Error = err
			return result
		}
		return g.reconcileSecret(ctx, spec, genSecret, missingKeys, legacyData)
	}

	legacyData, err := g.extractLegacyData(ctx, spec)
	if err != nil {
		result.Error = err
		return result
	}
	return g.createNewSecret(ctx, spec, legacyData)
}

// handleExistingSecret verifies the user-provided secret exists.
func (g *SecretGenerator) handleExistingSecret(ctx context.Context, spec SecretSpec) GenerationResult {
	result := GenerationResult{SecretName: spec.ExistingSecretName}

	_, err := g.getSecret(ctx, spec.ExistingSecretName)
	if err != nil {
		if errors.IsNotFound(err) {
			result.Error = fmt.Errorf("existingSecret %s not found", spec.ExistingSecretName)
		} else {
			result.Error = fmt.Errorf("failed to get existingSecret: %w", err)
		}
		return result
	}

	result.Action = "skipped"
	result.Message = fmt.Sprintf("using external secret %s", spec.ExistingSecretName)
	return result
}

// extractLegacyData retrieves key values from a legacy secret via the appropriate migration handler.
func (g *SecretGenerator) extractLegacyData(ctx context.Context, spec SecretSpec) (map[string][]byte, error) {
	if spec.LegacySecretName == "" {
		return nil, nil
	}

	legacySecret, err := g.getSecret(ctx, spec.LegacySecretName)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to check legacy secret %s: %w", spec.LegacySecretName, err)
	}

	logging.Infof("Found legacy secret %s for spec %s", spec.LegacySecretName, spec.Name)

	handler := getMigrationHandler(spec.LegacySecretName)
	data, err := handler.Migrate(spec, legacySecret, spec.LegacySecretName)
	if err != nil {
		return nil, fmt.Errorf("migration handler failed for legacy secret %s: %w", spec.LegacySecretName, err)
	}
	return data, nil
}

func (g *SecretGenerator) createNewSecret(ctx context.Context, spec SecretSpec, legacyData map[string][]byte) GenerationResult {
	result := GenerationResult{SecretName: spec.Name}

	logging.Infof("Creating new secret %s", spec.Name)

	secretData, err := ResolveKeyData(spec.Keys, legacyData)
	if err != nil {
		result.Error = err
		return result
	}

	if err := g.createSecret(ctx, spec.Name, secretData); err != nil {
		result.Error = fmt.Errorf("failed to create secret: %w", err)
		return result
	}

	if legacyData != nil {
		result.Action = "migrated"
		result.Message = fmt.Sprintf("migrated from legacy secret %s", spec.LegacySecretName)
	} else {
		result.Action = "created"
		result.Message = "generated new secret with random values"
	}
	return result
}

func findMissingKeys(keys []SecretKeySpec, existingData map[string][]byte) []SecretKeySpec {
	var missing []SecretKeySpec
	for _, keySpec := range keys {
		if _, ok := existingData[keySpec.Name]; !ok {
			missing = append(missing, keySpec)
		}
	}
	return missing
}

// reconcileSecret adds missing keys to an existing secret, seeding values from
// legacy data when available before falling back to generation.
func (g *SecretGenerator) reconcileSecret(ctx context.Context, spec SecretSpec, existingSecret *corev1.Secret, missingKeys []SecretKeySpec, legacyData map[string][]byte) GenerationResult {
	result := GenerationResult{SecretName: spec.Name}

	missingKeyNames := make([]string, len(missingKeys))
	for i, k := range missingKeys {
		missingKeyNames[i] = k.Name
	}
	logging.Infof("Reconciling secret %s: adding missing keys %v", spec.Name, missingKeyNames)

	seedData := make(map[string][]byte)
	for k, v := range legacyData {
		if _, exists := seedData[k]; !exists {
			seedData[k] = v
		}
	}

	for k, v := range existingSecret.Data {
		seedData[k] = v
	}

	secretData, err := ResolveKeyData(missingKeys, seedData)
	if err != nil {
		result.Error = err
		return result
	}

	if g.config.DryRun {
		logging.Infof("[DRY-RUN] Would update secret %s with %d new keys", spec.Name, len(missingKeys))
		result.Action = "reconciled"
		result.Message = fmt.Sprintf("would add missing keys %v (dry-run)", missingKeyNames)
		return result
	}

	updatedSecret := existingSecret.DeepCopy()
	updatedSecret.Data = secretData
	if err := g.updateSecret(ctx, updatedSecret); err != nil {
		result.Error = fmt.Errorf("failed to update secret with missing keys: %w", err)
		return result
	}

	result.Action = "reconciled"
	result.Message = fmt.Sprintf("added missing keys: %v", missingKeyNames)
	return result
}

func (g *SecretGenerator) getSecret(ctx context.Context, name string) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	err := retryOnTransient(ctx, fmt.Sprintf("get secret %s", name), func() error {
		return g.client.Get(ctx, client.ObjectKey{
			Namespace: g.config.Namespace,
			Name:      name,
		}, secret)
	})
	if err != nil {
		return nil, err
	}
	return secret, nil
}

// createSecret creates a secret without OwnerReferences.
func (g *SecretGenerator) createSecret(ctx context.Context, name string, data map[string][]byte) error {
	if g.config.DryRun {
		logging.Infof("[DRY-RUN] Would create secret %s with %d keys", name, len(data))
		return nil
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: g.config.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "secret-generator",
				"app.kubernetes.io/part-of":    g.config.FullnameOverride,
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: data,
	}

	return retryOnTransient(ctx, fmt.Sprintf("create secret %s", name), func() error {
		return g.client.Create(ctx, secret)
	})
}

func (g *SecretGenerator) updateSecret(ctx context.Context, secret *corev1.Secret) error {
	return retryOnTransient(ctx, fmt.Sprintf("update secret %s", secret.Name), func() error {
		return g.client.Update(ctx, secret)
	})
}

func (g *SecretGenerator) reportSummary(results []GenerationResult) {
	counts := make(map[string]int)
	errorCount := 0

	for _, r := range results {
		if r.Error != nil {
			errorCount++
		} else {
			counts[r.Action]++
		}
	}

	logging.Infof("=== Secret Generation Summary ===")
	for action, count := range counts {
		logging.Infof("  %s: %d", action, count)
	}
	if errorCount > 0 {
		logging.Errorf("  errors: %d", errorCount)
	}
	logging.Infof("================================")
}

func isTransientError(err error) bool {
	if err == nil {
		return false
	}
	return errors.IsServerTimeout(err) ||
		errors.IsServiceUnavailable(err) ||
		errors.IsTooManyRequests(err) ||
		errors.IsInternalError(err) ||
		errors.IsTimeout(err)
}

// retryOnTransient retries fn on transient K8s API errors.
func retryOnTransient(ctx context.Context, operation string, fn func() error) error {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			if lastErr != nil {
				return fmt.Errorf("%s interrupted after %d attempts (last error: %v): %w", operation, attempt, lastErr, err)
			}
			return fmt.Errorf("%s interrupted: %w", operation, err)
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}
		if !isTransientError(lastErr) {
			return lastErr
		}

		if attempt < maxRetries {
			delay := retryBaseDelay * time.Duration(1<<uint(attempt))
			logging.Warnf("Transient error during %s (attempt %d/%d): %v, retrying in %v",
				operation, attempt+1, maxRetries+1, lastErr, delay)

			select {
			case <-ctx.Done():
				return fmt.Errorf("%s interrupted during retry wait (last error: %v): %w", operation, lastErr, ctx.Err())
			case <-time.After(delay):
			}
		}
	}
	return fmt.Errorf("%s failed after %d attempts: %w", operation, maxRetries+1, lastErr)
}
