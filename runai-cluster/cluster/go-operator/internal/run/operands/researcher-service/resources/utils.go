/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetTlsSecretCertAndKey(ctx context.Context, k8sReader client.Reader, secretName string, namespace string) (string, string, error) {
	secret := &v1.Secret{}
	err := k8sReader.Get(ctx, types.NamespacedName{
		Name:      secretName,
		Namespace: namespace,
	}, secret)
	if err != nil {
		return "", "", fmt.Errorf("failed to get secret %s/%s: %w", secretName, namespace, err)
	}

	certificate := string(secret.Data["tls.crt"])
	key := string(secret.Data["tls.key"])

	return certificate, key, nil
}
