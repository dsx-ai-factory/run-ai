/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package step

import (
	"context"
	"fmt"

	configinputs "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	"github.com/run-ai/runai/runai-installer/installer/internal/utils/cert"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type InstallTlsSecret struct {
	*RunAIInstallStepImpl
	secretName       string
	namespace        string
	sslCertPathParam configinputs.InstallInputType
	sslKeyPathParam  configinputs.InstallInputType
	hostParam        configinputs.InstallInputType
	k8sClient        k8s.K8sClients
}

func NewInstallTlsSecret(secretName, namespace string, sslCertPathParam, sslKeyPathParam, hostParam configinputs.InstallInputType, k8sClient k8s.K8sClients) *InstallTlsSecret {
	name := fmt.Sprintf("Install %s secret", secretName)
	return &InstallTlsSecret{
		RunAIInstallStepImpl: NewRunAIInstallStepImpl(name),
		secretName:           secretName,
		namespace:            namespace,
		sslCertPathParam:     sslCertPathParam,
		sslKeyPathParam:      sslKeyPathParam,
		hostParam:            hostParam,
		k8sClient:            k8sClient,
	}
}

func (i *InstallTlsSecret) GetStatusAsync(runaiData *model.RunAIInstallData) model.InstallPhase {
	// Skip evaluation if in terminal state or Execute() is running
	if i.ShouldSkipStatusEvaluation() {
		return i.GetCachedStatus()
	}
	logging.Infof("[RunAI] Checking if TLS secret '%s' exists in namespace '%s'...", i.secretName, i.namespace)
	secret := &corev1.Secret{}
	err := i.k8sClient.GetClient().Get(context.Background(), client.ObjectKey{
		Namespace: i.namespace,
		Name:      i.secretName,
	}, secret)

	if err != nil {
		if errors.IsNotFound(err) {
			logging.Infof("[RunAI] TLS secret '%s' does not exist", i.secretName)
			return i.SetStatus(model.NotInstalled, nil)
		}
		logging.Errorf("[RunAI] Failed to check TLS secret '%s': %v", i.secretName, err)
		return i.SetStatus(model.Failed, fmt.Errorf("failed to check if secret: %s in namespace: %s exist: [%v]", i.secretName, i.namespace, err))
	}
	logging.Infof("[RunAI] TLS secret '%s' exists", i.secretName)
	return i.SetStatus(model.Installed, nil)
}

func (i *InstallTlsSecret) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	// Check for context cancellation before starting
	if err := i.CheckContextCancellation(ctx); err != nil {
		return err
	}

	logging.Infof("[RunAI] Creating TLS secret '%s' in namespace '%s'...", i.secretName, i.namespace)
	i.SetStatus(model.Installing, nil)

	// Check for context cancellation before k8s operations
	if err := i.CheckContextCancellation(ctx); err != nil {
		return err
	}

	secret := &corev1.Secret{}
	err := i.k8sClient.GetClient().Get(ctx, client.ObjectKey{
		Namespace: i.namespace,
		Name:      i.secretName,
	}, secret)

	if err == nil {
		logging.Infof("[RunAI] TLS secret '%s' exists, skipping creation", i.secretName)
		i.SetStatus(model.Installed, nil)
		return nil
	}

	if !errors.IsNotFound(err) {
		error := fmt.Errorf("failed to check if secret: %s in namespace: %s exist: [%v]", i.secretName, i.namespace, err)
		logging.Errorf("[RunAI] failed to check if secret: %s in namespace: %s exist: [%v]", i.secretName, i.namespace, err)
		i.SetStatus(model.Failed, error)
		return error
	}

	// Check for context cancellation before cert operations
	if err := i.CheckContextCancellation(ctx); err != nil {
		return err
	}

	logging.Infof("[RunAI] Loading TLS certificate and key...")
	certHandler := cert.GetCertHandler(i.sslCertPathParam, i.sslKeyPathParam, i.hostParam, runaiData)
	certPEM, err := certHandler.GetCRT()
	if err != nil {
		logging.Errorf("[RunAI] Failed to get TLS certificate: %v", err)
		i.SetStatus(model.Failed, err)
		return err
	}
	keyPEM, err := certHandler.GetKey()
	if err != nil {
		logging.Errorf("[RunAI] Failed to get TLS key: %v", err)
		i.SetStatus(model.Failed, err)
		return err
	}

	certData := []byte(certPEM)
	keyData := []byte(keyPEM)

	// Check for context cancellation before creating secret
	if err := i.CheckContextCancellation(ctx); err != nil {
		return err
	}

	// Create the TLS secret
	secret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      i.secretName,
			Namespace: i.namespace,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.crt": certData,
			"tls.key": keyData,
		},
	}

	// Create the secret in the cluster
	err = i.k8sClient.GetClient().Create(ctx, secret)
	if err != nil {
		error := fmt.Errorf("failed to create TLS secret '%s', in namespace: %s : %v", i.secretName, i.namespace, err)
		logging.Errorf("[RunAI] failed to create TLS secret '%s', in namespace: %s : %v", i.secretName, i.namespace, err)
		i.SetStatus(model.Failed, error)
		return error
	}
	logging.Infof("[RunAI] TLS secret '%s' created successfully", i.secretName)
	i.SetStatus(model.Installed, nil)
	return nil
}
