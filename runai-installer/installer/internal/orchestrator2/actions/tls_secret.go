/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/utils/cert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TLSSecret struct {
	SecretName string
	Namespace  string
	CertPath   string
	KeyPath    string
}

func (a TLSSecret) Kind() string { return "k8s.tls_secret" }

func (a TLSSecret) Args() map[string]any {
	return map[string]any{
		"name":      a.SecretName,
		"namespace": a.Namespace,
		"certPath":  a.CertPath,
		"keyPath":   a.KeyPath,
	}
}

func (a TLSSecret) Describe() string {
	return fmt.Sprintf("apply tls secret %s/%s from %s", a.Namespace, a.SecretName, a.CertPath)
}

func (a TLSSecret) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	handler := cert.NewLoadCertFromPath(a.CertPath, a.KeyPath)
	crt, err := handler.GetCRT()
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}
	key, err := handler.GetKey()
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}
	secret := &corev1.Secret{
		TypeMeta:   metav1.TypeMeta{Kind: "Secret", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{Name: a.SecretName, Namespace: a.Namespace},
		Type:       corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.crt": []byte(crt),
			"tls.key": []byte(key),
		},
	}
	return orchestrator2.ActionResult{}, deps.K8sExecutor.Apply(ctx, secret)
}
