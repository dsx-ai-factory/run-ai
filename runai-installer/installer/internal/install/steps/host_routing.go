/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	configinputs "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/cert"
	k8s "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	HostRoutingTlsSecretName = "runai-cluster-domain-star-tls-secret"
	HostRoutingIngressName   = "runai-cluster-domain-star-ingress"
	HostRoutingIngressClass  = "nginx"
)

type InstallHostBasedRouting struct {
	*step.RunAIInstallStepImpl
	k8sClient k8s.K8sClients
}

func NewInstallHostBasedRouting(k8sClient k8s.K8sClients) *InstallHostBasedRouting {
	s := &InstallHostBasedRouting{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Install host-based routing"),
		k8sClient:            k8sClient,
	}
	s.SetStatusEvaluator(s.EvaluateStatus)
	return s
}

func (s *InstallHostBasedRouting) EvaluateStatus(runaiData *model.RunAIInstallData) {
	if s.ShouldSkipStatusEvaluation() {
		return
	}

	logging.Infof("[RunAI] Checking host-based routing resources...")

	if err := s.checkTlsSecretExists(); err != nil {
		return
	}
	if err := s.checkIngressExists(); err != nil {
		return
	}

	logging.Infof("[RunAI] Host-based routing resources exist")
	s.SetStatus(model.Installed, nil)
}

func (s *InstallHostBasedRouting) checkTlsSecretExists() error {
	secret := &corev1.Secret{}
	err := s.k8sClient.GetClient().Get(context.Background(), client.ObjectKey{
		Namespace: RunaiClusterNamespace,
		Name:      HostRoutingTlsSecretName,
	}, secret)
	if err != nil {
		if errors.IsNotFound(err) {
			logging.Infof("[RunAI] Host routing TLS secret does not exist")
			return err
		}
		logging.Errorf("[RunAI] Failed to check host routing TLS secret: %v", err)
		s.SetStatus(model.Failed, fmt.Errorf("failed to check host routing TLS secret: %v", err))
		return err
	}
	return nil
}

func (s *InstallHostBasedRouting) checkIngressExists() error {
	dyn, err := dynamic.NewForConfig(s.k8sClient.GetConfig())
	if err != nil {
		logging.Errorf("[RunAI] Failed to create dynamic client: %v", err)
		s.SetStatus(model.Failed, fmt.Errorf("failed to create dynamic client: %v", err))
		return err
	}

	gvr := schema.GroupVersionResource{
		Group:    "networking.k8s.io",
		Version:  "v1",
		Resource: "ingresses",
	}
	_, err = dyn.Resource(gvr).Namespace(RunaiClusterNamespace).Get(context.Background(), HostRoutingIngressName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			logging.Infof("[RunAI] Host routing ingress does not exist")
			return err
		}
		logging.Errorf("[RunAI] Failed to check host routing ingress: %v", err)
		s.SetStatus(model.Failed, fmt.Errorf("failed to check host routing ingress: %v", err))
		return err
	}
	return nil
}

func (s *InstallHostBasedRouting) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	if err := s.CheckContextCancellation(ctx); err != nil {
		return err
	}

	s.SetStatus(model.Installing, nil)
	fqdn := runaiData.GetInputValue(configinputs.InstallInputFQDN)
	wildcardHost := "*." + fqdn

	logging.Infof("[RunAI] Setting up host-based routing for %s", wildcardHost)

	if err := s.createTlsSecret(ctx, runaiData); err != nil {
		s.SetStatus(model.Failed, err)
		return err
	}

	if err := s.CheckContextCancellation(ctx); err != nil {
		return err
	}

	if err := s.createIngress(ctx, wildcardHost); err != nil {
		s.SetStatus(model.Failed, err)
		return err
	}

	logging.Infof("[RunAI] Host-based routing configured successfully")
	s.SetStatus(model.Installed, nil)
	return nil
}

func (s *InstallHostBasedRouting) createTlsSecret(ctx context.Context, runaiData *model.RunAIInstallData) error {
	existing := &corev1.Secret{}
	err := s.k8sClient.GetClient().Get(ctx, client.ObjectKey{
		Namespace: RunaiClusterNamespace,
		Name:      HostRoutingTlsSecretName,
	}, existing)
	if err == nil {
		logging.Infof("[RunAI] Host routing TLS secret already exists, skipping")
		return nil
	}
	if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to check host routing TLS secret: %v", err)
	}

	certHandler := cert.GetCertHandler(
		configinputs.InstallInputHostRoutingCertPath,
		configinputs.InstallInputHostRoutingKeyPath,
		configinputs.InstallInputFQDN,
		runaiData,
	)
	certPEM, err := certHandler.GetCRT()
	if err != nil {
		return fmt.Errorf("failed to load host routing certificate: %v", err)
	}
	keyPEM, err := certHandler.GetKey()
	if err != nil {
		return fmt.Errorf("failed to load host routing key: %v", err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      HostRoutingTlsSecretName,
			Namespace: RunaiClusterNamespace,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.crt": []byte(certPEM),
			"tls.key": []byte(keyPEM),
		},
	}

	logging.Infof("[RunAI] Creating host routing TLS secret '%s'", HostRoutingTlsSecretName)
	if err := s.k8sClient.GetClient().Create(ctx, secret); err != nil {
		return fmt.Errorf("failed to create host routing TLS secret: %v", err)
	}
	return nil
}

func (s *InstallHostBasedRouting) createIngress(ctx context.Context, wildcardHost string) error {
	dyn, err := dynamic.NewForConfig(s.k8sClient.GetConfig())
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %v", err)
	}

	gvr := schema.GroupVersionResource{
		Group:    "networking.k8s.io",
		Version:  "v1",
		Resource: "ingresses",
	}

	_, err = dyn.Resource(gvr).Namespace(RunaiClusterNamespace).Get(ctx, HostRoutingIngressName, metav1.GetOptions{})
	if err == nil {
		logging.Infof("[RunAI] Host routing ingress already exists, skipping")
		return nil
	}
	if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to check host routing ingress: %v", err)
	}

	ingress := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "networking.k8s.io/v1",
			"kind":       "Ingress",
			"metadata": map[string]interface{}{
				"name":      HostRoutingIngressName,
				"namespace": RunaiClusterNamespace,
			},
			"spec": map[string]interface{}{
				"ingressClassName": HostRoutingIngressClass,
				"rules": []interface{}{
					map[string]interface{}{
						"host": wildcardHost,
					},
				},
				"tls": []interface{}{
					map[string]interface{}{
						"hosts": []interface{}{
							wildcardHost,
						},
						"secretName": HostRoutingTlsSecretName,
					},
				},
			},
		},
	}

	logging.Infof("[RunAI] Creating host routing ingress '%s'", HostRoutingIngressName)
	_, err = dyn.Resource(gvr).Namespace(RunaiClusterNamespace).Create(ctx, ingress, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create host routing ingress: %v", err)
	}
	return nil
}
