/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	workloadsAuthSecretName = constants.WorkloadAuthSecretName
	clusterRedisSecretName  = "cluster-redis"

	oAuthClientIDKey          = "OAUTH2_PROXY_CLIENT_ID"
	oAuthClientSecretKey      = constants.WorkloadAuthClientSecretKey
	oAuthProxyCookieSecretKey = constants.WorkloadAuthCookieSecretKey

	oAuthProxyRedisPasswordKey = constants.WorkloadAuthRedisPasswordKey
)

type SecretResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

func NewSecretResource(operandName string, k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) *SecretResource {
	return &SecretResource{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: runaiConfig,
	}
}

func (s *SecretResource) OperandName() string {
	return s.operandName
}

func (s *SecretResource) Object(ctx context.Context) (client.Object, error) {
	secretName := workloadsAuthSecretName
	secretNS := s.runaiConfig.GetNamespace()

	secret := &v1.Secret{}
	err := s.k8sReader.Get(ctx,
		types.NamespacedName{
			Namespace: secretNS,
			Name:      secretName},
		secret)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	secret.Name = secretName
	secret.Namespace = secretNS

	if secret.Labels == nil {
		secret.Labels = map[string]string{}
	}
	secret.Labels[constants.LabelName] = secretName
	secret.Labels[constants.ClusterWideResourceLabel] = "true"

	cookieSecret := ""
	if secret.Data != nil {
		cookieSecret = string(secret.Data[oAuthProxyCookieSecretKey])
	}
	if cookieSecret == "" {
		var err error
		cookieSecret, err = generateCookieSecret()
		if err != nil {
			return nil, fmt.Errorf("failed to generate cookie secret: %w", err)
		}
	}

	redisSecret := &v1.Secret{}
	err = s.k8sReader.Get(ctx, types.NamespacedName{
		Namespace: secretNS,
		Name:      clusterRedisSecretName,
	}, redisSecret)
	if errors.IsNotFound(err) {
		return nil, fmt.Errorf("redis secret %s/%s is not ready", secretNS, clusterRedisSecretName)
	}
	if err != nil {
		return nil, err
	}
	redisPassword := string(redisSecret.Data[oAuthProxyRedisPasswordKey])
	if redisPassword == "" {
		return nil, fmt.Errorf("redis secret %s/%s does not contain %s", secretNS, clusterRedisSecretName, oAuthProxyRedisPasswordKey)
	}

	workloadsAuthClient := s.runaiConfig.Spec.Internal.Global.ControlPlane.Config.GetWorkloads().Client

	clientID := workloadsAuthClient.GetName()
	if clientID == "" && secret.Data != nil {
		clientID = string(secret.Data[oAuthClientIDKey])
	}

	clientSecret := workloadsAuthClient.GetSecret()
	if clientSecret == "" && secret.Data != nil {
		clientSecret = string(secret.Data[oAuthClientSecretKey])
	}
	if clientSecret == "" {
		return nil, fmt.Errorf("workloads auth client secret is empty and %s/%s does not contain %s", secretNS, secretName, oAuthClientSecretKey)
	}

	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}
	secret.Data[oAuthClientIDKey] = []byte(clientID)
	secret.Data[oAuthClientSecretKey] = []byte(clientSecret)
	secret.Data[oAuthProxyCookieSecretKey] = []byte(cookieSecret)
	secret.Data[oAuthProxyRedisPasswordKey] = []byte(redisPassword)

	return secret, nil
}

func (s *SecretResource) Status(ctx context.Context) error {
	secret := &v1.Secret{}
	err := s.k8sReader.Get(ctx, types.NamespacedName{
		Namespace: s.runaiConfig.GetNamespace(),
		Name:      workloadsAuthSecretName,
	}, secret)
	if errors.IsNotFound(err) {
		return fmt.Errorf("secret [%s] is missing", workloadsAuthSecretName)
	}
	if err != nil {
		return err
	}

	requiredKeys := []string{
		oAuthClientIDKey,
		oAuthClientSecretKey,
		oAuthProxyCookieSecretKey,
		oAuthProxyRedisPasswordKey,
	}
	for _, key := range requiredKeys {
		if len(secret.Data[key]) == 0 {
			return fmt.Errorf("secret [%s/%s] does not contain %s", secret.Namespace, secret.Name, key)
		}
	}

	return nil
}

func (s *SecretResource) Enabled() bool {
	return true
}

func generateCookieSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
