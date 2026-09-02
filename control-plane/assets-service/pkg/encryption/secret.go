/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package encryption

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	kubeClient "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/kube_client"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	SecretNamePattern = "assets-key-%d" //nolint:gosec
	SecretKeyPrefix   = "key"
	SecretKeyPrivate  = "private"
	SecretKeyPublic   = "public"
	SecretKeyCreated  = "created"
	CreatedFormat     = "2006-01-02 15:04:05.000000"
)

func LoadKeysFromSecret(ctx context.Context, tenantId int32) (*TenantEncryptionKeys, *api.Error) {
	initTenantKeysCache()
	cachedSecret, err := cache.Get[corev1.Secret](ctx, string(tenantId))
	if err != nil {
		logging.Warnf("Failed to get corev1/secret from cache for tenant %d: %v", tenantId, err)
	}
	if cachedSecret != nil {
		return parseSecretData(cachedSecret)
	}

	k8sClient := kubeClient.KubeClient()
	secretName := fmt.Sprintf(SecretNamePattern, tenantId)

	var secret corev1.Secret
	err = k8sClient.Get(ctx, client.ObjectKey{
		Name:      secretName,
		Namespace: getSecretNamespace(),
	}, &secret)
	if err != nil {
		if notFound(err) {
			return NewTenantEncryptionKeys(), nil
		}
		return nil, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to load secret %s", secretName), err)
	}

	err = cache.Set[corev1.Secret](ctx, string(tenantId), secret)
	if err != nil {
		logging.Warnf("Failed to set corev1/secret in cache for tenant %d: %v", tenantId, err)
	}

	return parseSecretData(&secret)
}

func parseSecretData(secret *corev1.Secret) (*TenantEncryptionKeys, *api.Error) {
	result := NewTenantEncryptionKeys()

	for secretKey, secretData := range secret.Data {
		//
		//   the secret key is composed of key-<id>-[private|public|created]
		//
		secretKeyParts := strings.Split(secretKey, "-")
		if len(secretKeyParts) != 3 || secretKeyParts[0] != SecretKeyPrefix {
			return nil, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("malformed secret key: %s", secretKey), nil)
		}
		id, err := strconv.Atoi(secretKeyParts[1])
		if err != nil {
			return nil, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("non numeric id in secret data key: %s", secretKey), err)
		}
		curKey, ok := result.Keys[int32(id)]
		if !ok {
			curKey = TenantEncryptionKey{Id: int32(id)}
		}

		switch secretKeyParts[2] {
		case SecretKeyPrivate:
			key, _ := x509.ParsePKCS8PrivateKey(secretData)
			curKey.Private = key.(*rsa.PrivateKey)
		case SecretKeyPublic:
			key, _ := x509.ParsePKCS1PublicKey(secretData)
			curKey.Public = key
		case SecretKeyCreated:
			curKey.Created, err = time.Parse(CreatedFormat, string(secretData))
			if err != nil {
				return nil, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("invalid secret creation key: %s", secretData), err)
			}
		default:
			return nil, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("unsupported secret key: %s", secretKey), err)
		}

		result.Keys[int32(id)] = curKey
		if int32(id) > result.ActiveKeyId {
			result.ActiveKeyId = int32(id)
		}
	}

	return result, nil
}

func StoreTenantKeysInSecret(ctx context.Context, tenantId int32, keys *TenantEncryptionKeys) *api.Error {
	k8sClient := kubeClient.KubeClient()
	secretName := fmt.Sprintf(SecretNamePattern, tenantId)

	secretData := make(map[string][]byte)
	for _, key := range keys.Keys {
		privBytes, err := x509.MarshalPKCS8PrivateKey(key.Private)
		if err != nil {
			return api.NewError(http.StatusSwitchingProtocols, "failed to marshal private key", err)
		}
		pubBytes := x509.MarshalPKCS1PublicKey(key.Public)

		secretData[fmt.Sprintf("%s-%d-%s", SecretKeyPrefix, key.Id, SecretKeyPrivate)] = privBytes
		secretData[fmt.Sprintf("%s-%d-%s", SecretKeyPrefix, key.Id, SecretKeyPublic)] = pubBytes
		secretData[fmt.Sprintf("%s-%d-%s", SecretKeyPrefix, key.Id, SecretKeyCreated)] = []byte(key.Created.Format(CreatedFormat))
	}

	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: getSecretNamespace(),
		},
		Data: secretData,
	}

	err := k8sClient.Update(ctx, &secret)
	if notFound(err) {
		err = k8sClient.Create(ctx, &secret)
	}
	if err != nil {
		return api.NewError(http.StatusServiceUnavailable, "failed to persist tenant keys", err)
	}

	initTenantKeysCache()
	err = cache.Set[corev1.Secret](ctx, string(tenantId), secret)
	if err != nil {
		logging.Warnf("Failed to set corev1/secret in cache for tenant %d: %v", tenantId, err)
	}
	return nil
}
