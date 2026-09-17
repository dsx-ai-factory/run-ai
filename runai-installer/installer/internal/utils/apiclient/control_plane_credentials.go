/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package apiclient

import (
	"context"
	"fmt"

	"github.com/Masterminds/semver/v3"
	runaiclient "github.com/dsx-ai-factory/run-ai/runai-installer/clients/pkg/clients"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	AdminUsernameKey = "ADMIN_USERNAME"
	AdminPasswordKey = "ADMIN_PASSWORD"
)

func AdminConfigMapName(release string) string {
	return release + "-tenants-manager"
}

func AdminSecretName(release, version string) string {
	if useGeneratedAdminSecret(version) {
		return release + "-gen-tenants-manager"
	}
	return release + "-tenants-manager"
}

type ObjectGetter interface {
	Get(ctx context.Context, key client.ObjectKey, obj client.Object) error
}

type clientObjectGetter struct {
	client.Client
}

func NewObjectGetter(k8sClient client.Client) ObjectGetter {
	return &clientObjectGetter{Client: k8sClient}
}

func (g *clientObjectGetter) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	return g.Client.Get(ctx, key, obj)
}

func LoadControlPlaneCredentials(ctx context.Context, getter ObjectGetter, namespace, release, version string) (runaiclient.Credentials, error) {
	configMap := &corev1.ConfigMap{}
	if err := getter.Get(ctx, client.ObjectKey{Namespace: namespace, Name: AdminConfigMapName(release)}, configMap); err != nil {
		return runaiclient.Credentials{}, fmt.Errorf("get control-plane admin username: %w", err)
	}

	secret := &corev1.Secret{}
	if err := getter.Get(ctx, client.ObjectKey{Namespace: namespace, Name: AdminSecretName(release, version)}, secret); err != nil {
		return runaiclient.Credentials{}, fmt.Errorf("get control-plane admin password: %w", err)
	}

	username := configMap.Data[AdminUsernameKey]
	if username == "" {
		return runaiclient.Credentials{}, fmt.Errorf("%s is empty in ConfigMap %s/%s", AdminUsernameKey, namespace, configMap.Name)
	}
	password := string(secret.Data[AdminPasswordKey])
	if password == "" {
		return runaiclient.Credentials{}, fmt.Errorf("%s is empty in Secret %s/%s", AdminPasswordKey, namespace, secret.Name)
	}

	return runaiclient.Credentials{Username: username, Password: password}, nil
}

func useGeneratedAdminSecret(version string) bool {
	v, err := semver.NewVersion(version)
	if err != nil {
		return true
	}
	return v.Major() == 0 || v.Major() > 2 || v.Major() == 2 && v.Minor() > 25
}
