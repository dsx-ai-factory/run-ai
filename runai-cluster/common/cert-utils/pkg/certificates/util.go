/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package certificates

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	jp "github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils/json-patch"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// GenerateRandomString generates a random string of length 10
func GenerateRandomString() (string, error) {
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// create k8s tls secret
func createOrPatchK8sTLSSecret(
	ctx context.Context, client kubernetes.Interface, secretName, namespace string,
	createIfNotExists bool, cert, key []byte,
) error {
	_, err := client.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if k8serr.IsNotFound(err) && createIfNotExists {
			return createNewSecret(ctx, client, secretName, namespace, cert, key)
		} else {
			return fmt.Errorf("failed to get secret %s, the secret is expected to exist in the cluster, error: %w", secretName, err)
		}
	}
	return patchSecret(ctx, client, secretName, namespace, cert, key)
}

func createNewSecret(ctx context.Context, client kubernetes.Interface, secretName, namespace string, cert, key []byte) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"tls.crt": cert,
			"tls.key": key,
		},
	}
	_, err := client.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	return err
}

func patchSecret(ctx context.Context, client kubernetes.Interface, secretName, namespace string, cert, key []byte) error {
	jpOp := jp.NewPatch()
	jpOp.AddOperations(jp.NewReplaceOperation("/data/tls.crt", cert))
	jpOp.AddOperations(jp.NewReplaceOperation("/data/tls.key", key))
	certPatch, err := jpOp.Bytes()
	if err != nil {
		return err
	}
	_, err = client.CoreV1().Secrets(namespace).Patch(ctx, secretName, types.JSONPatchType, certPatch, metav1.PatchOptions{})
	return err
}
