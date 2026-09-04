/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
	k8s "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	configinputs "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
)

var execCommand = exec.Command

type InstallRepoAccessSecret struct {
	*step.RunAIInstallStepImpl
	k8sClient k8s.K8sClients
}

func NewInstallRepoAccessSecret(k8sClient k8s.K8sClients) step.InstallStep {
	return &InstallRepoAccessSecret{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Install repo access secret"),
		k8sClient:            k8sClient,
	}
}

func (i *InstallRepoAccessSecret) GetStatusAsync(runaiData *model.RunAIInstallData) model.InstallPhase {
	// Skip evaluation if in terminal state or Execute() is running
	if i.ShouldSkipStatusEvaluation() {
		return i.GetCachedStatus()
	}
	logging.Infof("[RunAI] Checking if registry credentials secret '%s' exists...", RunaiRegCredsSecret)
	secret := &corev1.Secret{}
	err := i.k8sClient.GetClient().Get(context.Background(), client.ObjectKey{
		Namespace: step.CPNamespace,
		Name:      RunaiRegCredsSecret,
	}, secret)

	if err != nil {
		if errors.IsNotFound(err) {
			logging.Infof("[RunAI] Registry credentials secret '%s' does not exist", RunaiRegCredsSecret)
			return i.SetStatus(model.NotInstalled, nil)
		}
		logging.Errorf("[RunAI] Failed to check registry credentials secret: %v", err)
		return i.SetStatus(model.Failed, fmt.Errorf("failed to check for existing secret: %v", err))
	}

	logging.Infof("[RunAI] Registry credentials secret '%s' exists", RunaiRegCredsSecret)
	return i.SetStatus(model.Installed, nil)
}

func (i *InstallRepoAccessSecret) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	// Check for context cancellation before starting
	if err := i.CheckContextCancellation(ctx); err != nil {
		return err
	}

	logging.Infof("[RunAI] Creating registry credentials secret '%s'...", RunaiRegCredsSecret)
	i.SetStatus(model.Installing, nil)

	secret := &corev1.Secret{}
	err := i.k8sClient.GetClient().Get(ctx, client.ObjectKey{
		Namespace: step.CPNamespace,
		Name:      RunaiRegCredsSecret,
	}, secret)

	if err == nil {
		logging.Infof("[RunAI] Registry credentials secret '%s' exists, skipping creation", RunaiRegCredsSecret)
		i.SetStatus(model.Installed, nil)
		return nil
	}

	if !errors.IsNotFound(err) {
		error := fmt.Errorf("failed to check for existing secret: %v", err)
		logging.Errorf("[RunAI] failed to check for existing secret: %v", err)
		i.SetStatus(model.Failed, error)
		return error
	}

	ngcCredentials := runaiData.GetInputValue(configinputs.InstallInputJfrogCreds)

	// If in dev mode and NGC credentials are empty, use kubectl command
	if ngcCredentials == "" && runaiData.Config.DevMode {
		// Check for context cancellation before kubectl command
		if err := i.CheckContextCancellation(ctx); err != nil {
			return err
		}

		logging.Infof("[RunAI] Dev mode: Creating registry credentials from local Docker config...")
		home := os.Getenv("HOME")
		cmd := execCommand(
			"kubectl", "create", "secret", "generic",
			"-n", step.CPNamespace, RunaiRegCredsSecret,
			"--from-file=.dockerconfigjson="+home+"/.docker/config.json",
			"--type=kubernetes.io/dockerconfigjson",
		)
		cmd.Stdout = logging.InfoWriter()
		cmd.Stderr = logging.ErrorWriter()
		logging.Infof("[RunAI] Running kubectl command: %v", cmd.Args)
		err = cmd.Run()
		if err != nil {
			error := fmt.Errorf("failed to create secret using kubectl: %v", err)
			logging.Errorf("[RunAI] failed to create secret using kubectl: %v", err)
			i.SetStatus(model.Failed, error)
			return error
		}

		logging.Infof("[RunAI] Registry credentials secret created successfully from local Docker config")
		i.SetStatus(model.Installed, nil)
		return nil
	}

	if ngcCredentials == "" {
		error := fmt.Errorf("NGC credentials are required but not provided")
		logging.Errorf("[RunAI] NGC credentials are required but not provided: %v", error)
		i.SetStatus(model.Failed, error)
		return error
	}

	// Check for context cancellation before creating secret
	if err := i.CheckContextCancellation(ctx); err != nil {
		return err
	}

	logging.Infof("[RunAI] Decoding and creating registry credentials secret...")
	decodedNgcCredentials, err := base64.StdEncoding.DecodeString(ngcCredentials)
	if err != nil {
		error := fmt.Errorf("failed to decode NGC credentials: %v", err)
		logging.Errorf("[RunAI] failed to decode NGC credentials: %v", err)
		i.SetStatus(model.Failed, error)
		return error
	}

	// Create the secret with NGC credentials
	secret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RunaiRegCredsSecret,
			Namespace: step.CPNamespace,
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			".dockerconfigjson": decodedNgcCredentials,
		},
	}

	// Create the secret in the cluster
	err = i.k8sClient.GetClient().Create(ctx, secret)
	if err != nil {
		error := fmt.Errorf("failed to create secret '%s': %v", RunaiRegCredsSecret, err)
		logging.Errorf("[RunAI] failed to create secret '%s': %v", RunaiRegCredsSecret, err)
		i.SetStatus(model.Failed, error)
		return error
	}
	logging.Infof("[RunAI] Registry credentials secret '%s' created successfully", RunaiRegCredsSecret)
	i.SetStatus(model.Installed, nil)
	return nil
}
