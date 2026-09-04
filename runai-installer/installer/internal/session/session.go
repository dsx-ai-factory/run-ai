/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package session

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	configinputs "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	k8s "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	sessionSecretName      = "runai-installer-session"
	sessionSecretNamespace = "default"
	k8sOperationTimeout    = 5 * time.Second
)

// LoadInputs attempts to read persisted inputs from a Kubernetes Secret
func LoadInputs(k8sClient k8s.K8sClients) (map[configinputs.InstallInputType]string, error) {
	logging.Infof("[Session] Loading saved inputs from secret '%s/%s'...", sessionSecretNamespace, sessionSecretName)
	ctx, cancel := context.WithTimeout(context.Background(), k8sOperationTimeout)
	defer cancel()

	sec := &corev1.Secret{}
	err := k8sClient.GetClient().Get(ctx, client.ObjectKey{Name: sessionSecretName, Namespace: sessionSecretNamespace}, sec)
	if err != nil {
		if errors.IsNotFound(err) {
			logging.Infof("[Session] No saved session found (secret does not exist)")
			return nil, nil
		}
		logging.Errorf("[Session] Failed to load inputs from secret: %v", err)
		return nil, err
	}
	b, ok := sec.Data["inputs.json"]
	if !ok || len(b) == 0 {
		logging.Infof("[Session] Session secret exists but contains no inputs data")
		return nil, nil
	}
	var inputs map[configinputs.InstallInputType]string
	if err := json.Unmarshal(b, &inputs); err != nil {
		logging.Errorf("[Session] Failed to unmarshal inputs from secret: %v", err)
		return nil, err
	}
	logging.Infof("[Session] Successfully loaded %d saved inputs from previous session", len(inputs))
	return inputs, nil
}

// SaveInputs upserts the session secret in Kubernetes with only inputs data
func SaveInputs(k8sClient k8s.K8sClients, inputs map[configinputs.InstallInputType]string) error {
	logging.Infof("[Session] Saving %d inputs to secret '%s/%s'...", len(inputs), sessionSecretNamespace, sessionSecretName)
	ctx, cancel := context.WithTimeout(context.Background(), k8sOperationTimeout)
	defer cancel()

	payload, err := json.Marshal(inputs)
	if err != nil {
		logging.Errorf("[Session] Failed to marshal inputs: %v", err)
		return err
	}
	sec := &corev1.Secret{}
	err = k8sClient.GetClient().Get(ctx, client.ObjectKey{Name: sessionSecretName, Namespace: sessionSecretNamespace}, sec)
	if err != nil {
		if errors.IsNotFound(err) {
			logging.Infof("[Session] Creating new session secret...")
			sec = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      sessionSecretName,
					Namespace: sessionSecretNamespace,
					Labels: map[string]string{
						"app":  "runai-installer",
						"role": "session",
					},
				},
				Type: corev1.SecretTypeOpaque,
				Data: map[string][]byte{"inputs.json": payload},
			}
			if createErr := k8sClient.GetClient().Create(ctx, sec); createErr != nil {
				logging.Errorf("[Session] Failed to create session secret: %v", createErr)
				return createErr
			}
			logging.Infof("[Session] Session secret created successfully")
			return nil
		}
		logging.Errorf("[Session] Failed to get existing session secret: %v", err)
		return err
	}
	logging.Infof("[Session] Updating existing session secret...")
	if sec.Data == nil {
		sec.Data = map[string][]byte{}
	}
	sec.Data["inputs.json"] = payload
	if updateErr := k8sClient.GetClient().Update(ctx, sec); updateErr != nil {
		logging.Errorf("[Session] Failed to update session secret: %v", updateErr)
		return updateErr
	}
	logging.Infof("[Session] Session secret updated successfully")
	return nil
}

// SetupSessionLoggingIfFresh initializes a per-session log file if this is a fresh install
func SetupSessionLoggingIfFresh(opts *config.InstallConfig) (*model.RunAIInstallData, error) {
	return model.NewRunAIInstallData(opts)
}
