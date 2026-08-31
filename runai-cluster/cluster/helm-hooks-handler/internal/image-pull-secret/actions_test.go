/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package image_pull_secret

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCreateImagePullSecretSkipsEmptyData(t *testing.T) {
	k8s := newFakeClient(t)

	err := CreateImagePullSecret(k8s, "runai", "test-secret", " ")
	require.NoError(t, err)

	secret := &corev1.Secret{}
	err = k8s.Get(context.Background(), client.ObjectKey{
		Namespace: "runai",
		Name:      "test-secret",
	}, secret)
	assert.True(t, k8sErrors.IsNotFound(err))
}

func TestCreateImagePullSecretCreatesDockerConfigSecret(t *testing.T) {
	k8s := newFakeClient(t)
	dockerConfigJSON := []byte(`{"auths":{}}`)

	err := CreateImagePullSecret(k8s, "runai", "test-secret", base64.StdEncoding.EncodeToString(dockerConfigJSON))
	require.NoError(t, err)

	secret := &corev1.Secret{}
	err = k8s.Get(context.Background(), client.ObjectKey{
		Namespace: "runai",
		Name:      "test-secret",
	}, secret)
	require.NoError(t, err)
	assert.Equal(t, corev1.SecretTypeDockerConfigJson, secret.Type)
	assert.Equal(t, dockerConfigJSON, secret.Data[corev1.DockerConfigJsonKey])
}

func TestCreateImagePullSecretReturnsInvalidBase64Error(t *testing.T) {
	k8s := newFakeClient(t)

	err := CreateImagePullSecret(k8s, "runai", "test-secret", "not-base64")

	require.Error(t, err)
}

func newFakeClient(t *testing.T) client.Client {
	t.Helper()

	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))

	return fake.NewClientBuilder().WithScheme(scheme).Build()
}
