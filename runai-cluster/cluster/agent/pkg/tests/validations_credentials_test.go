/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"fmt"
	"os"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	configmaps "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/config-maps"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func BeCorrectlySyncedSecret(expected SecretAndOptionalUpdate) MatcherCorrectlySyncedSecret {
	return MatcherCorrectlySyncedSecret{
		Expected: expected,
	}
}

type MatcherCorrectlySyncedSecret struct {
	Expected SecretAndOptionalUpdate
}

func (matcher MatcherCorrectlySyncedSecret) Match(actual interface{}) (success bool, err error) {
	fullName := fmt.Sprintf("secret %s:%s", matcher.Expected.Secret.Namespace, matcher.Expected.Secret.Name)

	if openapi.IsNil(actual) {
		return false, fmt.Errorf("%s not created in the cluster", fullName)
	}

	actualSecret := actual.(*corev1.Secret)

	if matcher.Expected.Update != nil {
		if actualSecret.Labels == nil {
			return false, fmt.Errorf("%s: missing labels", fullName)
		}

		if actualSecret.Annotations == nil {
			return false, fmt.Errorf("%s: missing annotations", fullName)
		}

		var expectedKind string
		switch matcher.Expected.Update.Meta.Kind {
		case openapi.ASSETKIND_PASSWORD:
			expectedKind = "password"
		case openapi.ASSETKIND_ACCESS_KEY:
			expectedKind = "access-key"
		case openapi.ASSETKIND_DOCKER_REGISTRY:
			expectedKind = "docker-registry"
		case openapi.ASSETKIND_GENERIC_SECRET:
			expectedKind = "generic"
		default:
			return false, fmt.Errorf("%s: unsupported credential type %s", fullName, expectedKind)
		}

		if actualSecret.Labels["run.ai/asset"] != expectedKind {
			return false, fmt.Errorf("%s: run.ai/asset is %s, should be %s", fullName, actualSecret.Labels["run.ai/asset"], expectedKind)
		}

		if actualSecret.Labels["run.ai/asset-id"] != matcher.Expected.Update.Meta.Id {
			return false, fmt.Errorf("%s: run.ai/asset-id %v does not point to asset %s", fullName, actualSecret.Labels["run.ai/asset-id"], matcher.Expected.Update.Meta.Id)
		}

		updateCountAsString := fmt.Sprintf("%d", matcher.Expected.Update.UpdateCount)

		if actualSecret.Annotations["run.ai/update-count"] != updateCountAsString {
			return false, fmt.Errorf("%s: run.ai/update-count %v does not hold update count %s", fullName, actualSecret.Annotations["run.ai/update-count"], updateCountAsString)
		}

		// Verify the asset name is in annotations
		if actualSecret.Annotations["run.ai/asset-name"] != matcher.Expected.Update.Meta.Name {
			return false, fmt.Errorf("%s: secret should have asset name %s in annotations, but has %s",
				fullName, matcher.Expected.Update.Meta.Name, actualSecret.Annotations["run.ai/asset-name"])
		}

		// Check if the secret is private
		if convutils.BoolIsTrue(matcher.Expected.Update.Meta.Private.Get()) {
			// Verify it's in the run:ai namespace
			if actualSecret.Namespace != constants.NamespaceRunAi {
				return false, fmt.Errorf("%s: private secret should be in %s namespace, but is in %s",
					fullName, constants.NamespaceRunAi, actualSecret.Namespace)
			}

			// Verify the name follows the private secret naming convention
			expectedName := fmt.Sprintf("%s-%s", matcher.Expected.Update.Meta.Name, matcher.Expected.Update.Meta.Id[:10])
			if actualSecret.Name != expectedName {
				return false, fmt.Errorf("%s: private secret should be named %s, but is named %s",
					fullName, expectedName, actualSecret.Name)
			}

			// Verify private credentials label
			if actualSecret.Labels == nil || actualSecret.Labels["run.ai/user-credentials"] != "true" {
				return false, fmt.Errorf("%s: expected label of user-credentials was not found", fullName)
			}

			if matcher.Expected.Update.Meta.Kind == openapi.ASSETKIND_DOCKER_REGISTRY {
				if err := validateDockerRegistryCredentials(actualSecret, *matcher.Expected.Secret, fullName); err != nil {
					return false, err
				}
			} else {
				// Verify number of keys match
				if len(actualSecret.Data) != len(matcher.Expected.Secret.Data) {
					return false, fmt.Errorf("%s: number of keys in secret data does not match. Expected %d, got %d",
						fullName, len(matcher.Expected.Secret.Data), len(actualSecret.Data))
				}

				// Verify all values are encrypted and match expected values after decryption
				for key, expectedValue := range matcher.Expected.Secret.Data {
					encryptedValue, exists := actualSecret.Data[key]
					if !exists {
						return false, fmt.Errorf("%s: missing key %s in secret data", fullName, key)
					}
					// Decrypt the value for regular secrets
					decryptedValue, err := decryptData(encryptedValue)
					if err != nil {
						return false, fmt.Errorf("%s: failed to decrypt value for key %s: %w", fullName, key, err)
					}

					// Compare decrypted value with expected value
					if string(decryptedValue) != string(expectedValue) {
						return false, fmt.Errorf("%s: decrypted value for key %s does not match expected value", fullName, key)
					}
				}
			}
		} else {
			// Verify private credentials label is not generated
			if actualSecret.Labels != nil && actualSecret.Labels["run.ai/user-credentials"] == "true" {
				return false, fmt.Errorf("%s: label of user-credentials was not expected to be found", fullName)
			}
		}
	}

	return true, nil
}

func (matcher MatcherCorrectlySyncedSecret) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "secret %s:%s is not as expected", matcher.Expected.Secret.Namespace, matcher.Expected.Secret.Name)
}

func (matcher MatcherCorrectlySyncedSecret) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "secret %s:%s match the expected, failing the not(...) condition", matcher.Expected.Secret.Namespace, matcher.Expected.Secret.Name)
}

func validateSecretExistsEventually(secretAndOptionalUpdate SecretAndOptionalUpdate, k8sClient client.Client) {
	var secretResource *corev1.Secret = nil

	if secretAndOptionalUpdate.Update == nil {
		EventuallyWithOffset(1, func() *corev1.Secret {
			secretResource = getSecretFromClient(secretAndOptionalUpdate.Secret.Name, secretAndOptionalUpdate.Secret.Namespace, k8sClient)
			return secretResource
		}, DefaultTimeout).ShouldNot(BeNil())
	} else {
		EventuallyWithOffset(1, func() *corev1.Secret {
			secretResource = getSecretFromClient(secretAndOptionalUpdate.Secret.Name, secretAndOptionalUpdate.Secret.Namespace, k8sClient)
			return secretResource
		}, DefaultTimeout).Should(BeCorrectlySyncedSecret(secretAndOptionalUpdate))
	}
}

func validateSecretDoesNotExistsEventually(secret *corev1.Secret, k8sClient client.Client) {
	var secretResource *corev1.Secret = nil

	EventuallyWithOffset(1, func() *corev1.Secret {
		secretResource = getSecretFromClient(secret.Name, secret.Namespace, k8sClient)
		return secretResource
	}, DefaultTimeout).Should(BeNil())
}

func validateAgentConfigWithCredentialsCounter(k8sClient client.Client, counterValue string) {
	validateAgentConfigWithCounter(k8sClient, configmaps.CredentialsSyncCounter, counterValue)
}

// decryptData decrypts the encrypted data using the encryption key from environment
func decryptData(encryptedData []byte) ([]byte, error) {
	key := []byte(os.Getenv(constants.RunaiAdsEncryptionKeyEnvVar))
	if len(key) == 0 {
		return nil, fmt.Errorf("encryption key not found in environment variable %s", constants.RunaiAdsEncryptionKeyEnvVar)
	}

	// Create new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce from the beginning of the encrypted data
	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("encrypted data too short")
	}

	nonce := encryptedData[:nonceSize]
	ciphertext := encryptedData[nonceSize:]

	// Decrypt data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}

// validateDockerRegistryCredentials validates Docker registry secret data
func validateDockerRegistryCredentials(actualSecret *corev1.Secret, expectedSecret corev1.Secret, fullName string) error {
	secretData, ok := actualSecret.Data[".dockerconfigjson"]
	if !ok {
		return fmt.Errorf("%s: .dockerconfigjson not found in secret data", fullName)
	}

	decryptedData, err := decryptData(secretData)
	if err != nil {
		return fmt.Errorf("%s: failed to decrypt .dockerconfigjson: %w", fullName, err)
	}

	var decryptedConfig map[string]interface{}
	if err := json.Unmarshal(decryptedData, &decryptedConfig); err != nil {
		return fmt.Errorf("%s: failed to parse decrypted .dockerconfigjson: %w", fullName, err)
	}

	auths, ok := decryptedConfig["auths"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("%s: decrypted .dockerconfigjson missing 'auths' section", fullName)
	}

	for registryUrl, registryData := range auths {
		registryDataMap := registryData.(map[string]interface{})

		if registryUrl != string(expectedSecret.Data["url"]) {
			return fmt.Errorf("%s: url in registry %s does not match expected value", fullName, registryUrl)
		}

		username, ok := registryDataMap["username"].(string)
		if !ok {
			return fmt.Errorf("%s: missing username in registry %s", fullName, registryUrl)
		}
		if username != string(expectedSecret.Data["user"]) {
			return fmt.Errorf("%s: username in registry %s does not match expected value", fullName, registryUrl)
		}

		password, ok := registryDataMap["password"].(string)
		if !ok {
			return fmt.Errorf("%s: missing password in registry %s", fullName, registryUrl)
		}
		if password != string(expectedSecret.Data["password"]) {
			return fmt.Errorf("%s: password in registry %s does not match expected value", fullName, registryUrl)
		}
	}

	return nil
}
