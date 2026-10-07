/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"fmt"
	commonutils "github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils"
	"regexp"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultServiceAccountName = "default"
	userCredentialPrefix      = "USERCREDENTIAL:"
)

var spaceSplitter = regexp.MustCompile(`\s+`)

type WorkloadInfo struct {
	Name      string
	Namespace string
	Initials  string
}

func (w *WorkloadInfo) composeWorkloadCredentialName(credential string) string {
	vf := ValueFromUserCredentials
	return vf.patchName(credential, w.Initials, w.Name)
}

// BuildImagePullSecrets receives secretFromWorkloads, in format: secret-1 secret-2 secret-3...
// and it adds to a podspec a list of secrets composed of this input list + the (optional) secrets
// of the serviceAccount of the namespace. if no imagePullSecrets are specified in the workload field,
// leave the pod spec untouched and allow k8s to load it automatically from the serviceAccount when the
// workload runs.
func BuildImagePullSecrets(ctx context.Context, secretsFromWorkload *field_types.StringField, k8sClient client.Client,
	namespace string, target *[]corev1.LocalObjectReference, workloadName string, workloadKindInitials string,
) error {
	if secretsFromWorkload != nil {
		unifiedImagePullSecrets, err := CreateImagePullSecrets(ctx, k8sClient, namespace, secretsFromWorkload.Deref(), workloadName, workloadKindInitials)
		if err != nil {
			return err
		}

		if unifiedImagePullSecrets != nil {
			*target = unifiedImagePullSecrets
		}
	}

	return nil
}

// CreateImagePullSecrets merges receives an optional list of imagePullSecrets from the workload. if the list
// is not empty, it attempts to load the imagePullSecrets from the default service account and merge the
// imagePullSecrets between them both.
func CreateImagePullSecrets(ctx context.Context, k8sClient client.Client, namespace string, workloadImagePullSecrets string, workloadName string, workloadKindInitials string) ([]corev1.LocalObjectReference, error) {
	info := WorkloadInfo{
		Initials:  workloadKindInitials,
		Name:      workloadName,
		Namespace: namespace,
	}

	allNames, err := ParseImagePullSecrets(workloadImagePullSecrets, info)
	if err != nil {
		return nil, err
	}

	//
	//  if no imagePullSecrets are given in the workload, do nothing and let k8s load them itself from
	//  the default service account
	//
	if len(allNames) == 0 {
		return nil, nil
	}

	//
	//   load the default service account
	//
	serviceAccount, err := GetServiceAccount(ctx, k8sClient, types.NamespacedName{
		Namespace: namespace,
		Name:      defaultServiceAccountName,
	})
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	//
	//   add the missing secrets from the service account
	//
	if serviceAccount != nil && len(serviceAccount.ImagePullSecrets) != 0 {
		for _, secret := range serviceAccount.ImagePullSecrets {
			allNames.Add(secret.Name)
		}
	}

	result := make([]corev1.LocalObjectReference, 0, len(allNames))

	for _, name := range commonutils.SortedKeys(allNames) {
		result = append(result, corev1.LocalObjectReference{Name: name})
	}

	return result, nil
}

func GetServiceAccount(ctx context.Context, k8sClient client.Client, name types.NamespacedName) (*corev1.ServiceAccount, error) {
	var serviceAccount corev1.ServiceAccount

	err := k8sClient.Get(ctx, name, &serviceAccount)
	if err != nil {
		return nil, err
	}

	return &serviceAccount, nil
}

func ParseImagePullSecrets(imagePullSecretsAsString string, info WorkloadInfo) (datastructs.Set[string], error) {
	allNames := datastructs.Set[string]{}

	trimmed := strings.TrimSpace(imagePullSecretsAsString)
	if trimmed == "" {
		return allNames, nil
	}

	secretsAsArray := spaceSplitter.Split(trimmed, -1)

	for _, name := range secretsAsArray {
		// Remove USERCREDENTIAL prefix if present
		if strings.HasPrefix(name, userCredentialPrefix) {
			name = info.composeWorkloadCredentialName(strings.TrimPrefix(name, userCredentialPrefix))
		}
		//  validate the secret name using the official k8s function
		errs := validation.IsDNS1123Subdomain(name)
		if len(errs) != 0 {
			return nil, fmt.Errorf("invalid secret name %s: %v", name, errs[0])
		}

		allNames.Add(name)
	}

	return allNames, nil
}
