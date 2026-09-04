/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package helm_deployer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	TempAuthDir = "helm-auth"

	CaCertFile = "ca.crt"

	BasicAuthUser     = "username"
	BasicAuthPassword = "password"
	TokenAuth         = "token"
	NgcAPIKey         = "NGC_API_KEY"
	NgcBasicAuthUser  = "$oauthtoken"

	PrefixTlsCert      = "tls-cert-"
	PrefixTlsKey       = "tls-key-"
	PrefixCaCert       = "ca-cert-"
	PrefixDockerConfig = "docker-config-"

	TempFilePermissions = 0700
)

func getAuthOptions(ctx context.Context, k8sClient client.Client, namespace string, authSpec *v1alpha1.HelmAuthSpec) (*AuthOptions, error) {
	if authSpec == nil {
		return nil, nil
	}

	authOpts := AuthOptions{
		InsecureSkipTLSVerify: authSpec.InsecureSkipTLSVerify,
		CAFile:                authSpec.CAFile,
	}

	// If no secret reference, return with just the static options
	if authSpec.SecretRef == nil {
		return &authOpts, nil
	}

	// Read the content of the secret
	secret := &corev1.Secret{}
	secretKey := client.ObjectKey{
		Namespace: namespace,
		Name:      authSpec.SecretRef.Name,
	}

	if err := k8sClient.Get(ctx, secretKey, secret); err != nil {
		return nil, fmt.Errorf("failed to get auth secret %s: %w", authSpec.SecretRef.Name, err)
	}

	// Extract credentials based on secret type and available keys
	switch secret.Type {
	case corev1.SecretTypeDockerConfigJson:
		// For OCI registries - write docker config to temp file
		dockerConfig, ok := secret.Data[corev1.DockerConfigJsonKey]
		if !ok || len(dockerConfig) == 0 {
			return nil, fmt.Errorf("secret %s is type %s but missing key %s", authSpec.SecretRef.Name, corev1.SecretTypeDockerConfigJson, corev1.DockerConfigJsonKey)
		}
		tempFile, err := writeDockerConfigToTempFile(dockerConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to write docker config: %w", err)
		}
		authOpts.RegistryConfigFile = tempFile
		log.Debug().Msgf("using docker config from secret %s for OCI authentication", authSpec.SecretRef.Name)

	case corev1.SecretTypeTLS:
		// For TLS client certificates
		cert, certOK := secret.Data[corev1.TLSCertKey]
		key, keyOK := secret.Data[corev1.TLSPrivateKeyKey]
		if !certOK || !keyOK {
			return nil, fmt.Errorf("secret %s is type %s but missing %s or %s", authSpec.SecretRef.Name, corev1.SecretTypeTLS, corev1.TLSCertKey, corev1.TLSPrivateKeyKey)
		}

		tempCertFile, err := writeTempFile(PrefixTlsCert, cert)
		if err != nil {
			return nil, fmt.Errorf("failed to write TLS cert: %w", err)
		}
		authOpts.CertFile = tempCertFile

		tempKeyFile, err := writeTempFile(PrefixTlsKey, key)
		if err != nil {
			return nil, fmt.Errorf("failed to write TLS key: %w", err)
		}
		authOpts.KeyFile = tempKeyFile

		if ca, ok := secret.Data[CaCertFile]; ok {
			tempFile, err := writeTempFile(PrefixCaCert, ca)
			if err != nil {
				return nil, fmt.Errorf("failed to write CA cert: %w", err)
			}
			authOpts.CAFile = tempFile
		}

		log.Debug().Msgf("using TLS certificates from secret %s", authSpec.SecretRef.Name)

	default:
		// For Opaque secrets - check for username/password
		if username, ok := secret.Data[BasicAuthUser]; ok {
			authOpts.Username = string(username)
		}
		if password, ok := secret.Data[BasicAuthPassword]; ok {
			authOpts.Password = string(password)
		}
		if authOpts.Username == "" && authOpts.Password == "" {
			if apiKey, ok := secret.Data[NgcAPIKey]; ok && len(apiKey) > 0 {
				authOpts.Username = NgcBasicAuthUser
				authOpts.Password = string(apiKey)
				log.Debug().Msgf("using NGC API key from secret %s for Helm authentication", authSpec.SecretRef.Name)
			}
		}

		if authOpts.Username == "" && authOpts.Password != "" {
			return nil, errors.New("secret of type Opaque is missing username")
		}
		if authOpts.Username != "" && authOpts.Password == "" {
			return nil, errors.New("secret of type Opaque is missing password")
		}
		if authOpts.Username != "" && authOpts.Password != "" {
			log.Debug().Msgf("using basic auth from secret %s", authSpec.SecretRef.Name)
		}

		// token is not supported yet
		if _, ok := secret.Data[TokenAuth]; ok {
			return nil, errors.New("token authentication for helm chart is not supported yet")
		}

		if ca, ok := secret.Data[CaCertFile]; ok {
			tempFile, err := writeTempFile(PrefixCaCert, ca)
			if err != nil {
				return nil, fmt.Errorf("failed to write CA cert: %w", err)
			}
			authOpts.CAFile = tempFile

			log.Debug().Msgf("using CAFile auth from secret %s", authSpec.SecretRef.Name)
		}
	}

	return &authOpts, nil
}

func writeDockerConfigToTempFile(data []byte) (string, error) {
	return writeTempFile(PrefixDockerConfig, data)
}

// writeTempFile writes data to a temporary file and returns the path
// (TODO) note that those files are not removed by the controller, only by the casual cleanup of /tmp, it's not critical but not ideal
func writeTempFile(prefix string, data []byte) (string, error) {
	tempDir := filepath.Join(os.TempDir(), TempAuthDir)
	if err := os.MkdirAll(tempDir, TempFilePermissions); err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	tempFile, err := os.CreateTemp(tempDir, prefix)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func(tempFile *os.File) {
		_ = tempFile.Close()
	}(tempFile)

	if _, err := tempFile.Write(data); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	return tempFile.Name(), nil
}
