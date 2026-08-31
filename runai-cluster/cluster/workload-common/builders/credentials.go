/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"fmt"
	"strings"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"

	controllerutils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	utils "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders/utils"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/run-ai/runai/runai-application-packages/datastructs"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	secretTypeMeta     = metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"}
	secretGroupVersion = metav1.GroupVersionKind(secretTypeMeta.GetObjectKind().GroupVersionKind())
)

const dockerRegistryKindLabel = "docker-registry"

type CredentialsBuilder struct {
	k8sClient      client.Client
	runaiNamespace string
}

func NewCredentialsBuilder(k8sClient client.Client, runaiNamespace string) *CredentialsBuilder {
	return &CredentialsBuilder{
		k8sClient:      k8sClient,
		runaiNamespace: runaiNamespace,
	}
}

type WorkloadInfo struct {
	Name         string
	Namespace    string
	Initials     string
	Owner        client.Object
	WorkloadKind string
	JobType      string
}

func (b *CredentialsBuilder) Preview(ctx context.Context, info *WorkloadInfo, envs []*flds.DictionaryField, imagePullSecrets []*flds.StringField) ([]Resource, error) {
	credentials := extractAndDeduplicateCredentials(envs, imagePullSecrets, info)
	if len(credentials) == 0 {
		return nil, nil
	}

	result := make([]Resource, 0, len(credentials))
	for _, cred := range credentials {

		result = append(result, Resource{
			Kind:      secretGroupVersion,
			Name:      info.composeWorkloadCredentialName(cred),
			Namespace: info.Namespace,
		})
	}

	return result, nil
}

func (b *CredentialsBuilder) Build(ctx context.Context, info *WorkloadInfo, envs []*flds.DictionaryField, imagePullSecrets []*flds.StringField) ([]corev1.Secret, error) {
	credentials := extractAndDeduplicateCredentials(envs, imagePullSecrets, info)
	if len(credentials) == 0 {
		return nil, nil
	}

	// Get the encryption key from the cluster
	key, err := utils.GetEncryptionKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption key: %w", err)
	}

	secrets := make([]corev1.Secret, 0, len(credentials))

	for _, cred := range credentials {
		err := b.validatedCredential(ctx, info, cred)
		if err != nil {
			return nil, fmt.Errorf("failed to validate credential %s: %w", cred, err)
		}

		decryptCredential, err := b.buildCredential(ctx, info, cred, key)
		if err != nil {
			return nil, fmt.Errorf("failed to build credential for %s: %w", cred, err)
		}
		secrets = append(secrets, *decryptCredential)
	}

	return secrets, nil
}

func (b *CredentialsBuilder) validatedCredential(ctx context.Context, info *WorkloadInfo, credential string) error {
	var secret corev1.Secret
	err := b.k8sClient.Get(ctx, client.ObjectKey{
		Name:      info.composeWorkloadCredentialName(credential),
		Namespace: info.Namespace,
	}, &secret)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get secret %s: %w", info.composeWorkloadCredentialName(credential), err)
		}
		// If the secret is not found, we will create it
		return nil
	}
	if info.Owner != nil {
		if !metav1.IsControlledBy(&secret, info.Owner) {
			return fmt.Errorf("secret %s is not controlled by the workload owner %s", info.composeWorkloadCredentialName(credential), info.Owner.GetName())
		}
		return nil
	}

	return fmt.Errorf("secret %s is not owned by the workload", info.composeWorkloadCredentialName(credential))
}

func (b *CredentialsBuilder) buildCredential(ctx context.Context, info *WorkloadInfo, credential string, encryptionKey []byte) (*corev1.Secret, error) {
	var secret corev1.Secret
	err := b.k8sClient.Get(ctx, client.ObjectKey{
		Name:      credential,
		Namespace: b.runaiNamespace,
	}, &secret)
	if err != nil {
		return nil, fmt.Errorf("failed to find credential %s: %w", credential, err)
	}

	decryptCredential := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      info.composeWorkloadCredentialName(credential),
			Namespace: info.Namespace,
		},
	}

	if info.Owner != nil {
		err = controllerutils.SetControllerReference(info.Owner, decryptCredential, b.k8sClient.Scheme())
		if err != nil {
			return nil, fmt.Errorf("failed to set controller reference for credential %s: %v", info.Name, err)
		}
	}

	// copy labels and annotations from the original secret
	decryptCredential.Labels = make(map[string]string, len(secret.Labels))
	for key, value := range secret.Labels {
		decryptCredential.Labels[key] = value
	}
	decryptCredential.Annotations = make(map[string]string, len(secret.Annotations))
	for key, value := range secret.Annotations {
		decryptCredential.Annotations[key] = value
	}
	err = SetObjectMetaLabels(ctx, info.WorkloadKind, info.JobType, info.Name, &(decryptCredential.ObjectMeta))
	if err != nil {
		return nil, err
	}

	if secret.Data == nil {
		return nil, fmt.Errorf("credential %s does not contain any data", credential)
	}

	decryptCredential.Data = make(map[string][]byte, len(secret.Data))
	for key, encryptedValue := range secret.Data {
		decryptedValue, err := utils.DecryptData(encryptedValue, encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt data for key %s: %w", key, err)
		}
		decryptCredential.Data[key] = decryptedValue
	}

	if secret.Labels[constants.RunaiAssetLabel] == dockerRegistryKindLabel {
		decryptCredential.Type = corev1.SecretTypeDockerConfigJson
	}

	return decryptCredential, nil
}

func (w *WorkloadInfo) composeWorkloadCredentialName(credential string) string {
	vf := ValueFromUserCredentials
	return vf.patchName(credential, w.Initials, w.Name)
}

func extractAndDeduplicateCredentials(envs []*flds.DictionaryField, imagePullSecrets []*flds.StringField, info *WorkloadInfo) []string {
	credentials := extractCredentialsFromEnvs(envs, info)
	imagePullSecretsCredentials := extractCredentialsFromImagePullSecrets(imagePullSecrets)
	credentials = append(credentials, imagePullSecretsCredentials...)
	return datastructs.NewSet[string](credentials...).Keys()
}

func extractCredentialsFromEnvs(envs []*flds.DictionaryField, info *WorkloadInfo) []string {
	if envs == nil {
		return nil
	}

	dedupCredentials := datastructs.NewSet[string]()
	for _, env := range envs {
		for _, value := range env.Deref() {
			if !ptr.Deref(value.Deleted, false) {
				form, credential, _ := DecodeValue(ptr.Deref(value.Value, ""))
				if form != ValueFromUserCredentials {
					continue
				}
				dedupCredentials.Add(credential)
			}
		}
	}
	return dedupCredentials.Keys()
}

func extractCredentialsFromImagePullSecrets(imagePullSecrets []*flds.StringField) []string {
	if imagePullSecrets == nil {
		return nil
	}

	dedupCredentials := datastructs.NewSet[string]()
	for _, imagePullSecret := range imagePullSecrets {
		if imagePullSecret != nil {
			secretsAsArray := spaceSplitter.Split(strings.TrimSpace(imagePullSecret.Deref()), -1)
			for _, secret := range secretsAsArray {
				// Remove USERCREDENTIAL prefix if present
				if strings.HasPrefix(secret, userCredentialPrefix) {
					credential := strings.TrimPrefix(secret, userCredentialPrefix)
					dedupCredentials.Add(credential)
				}
			}
		}
	}

	return dedupCredentials.Keys()
}
