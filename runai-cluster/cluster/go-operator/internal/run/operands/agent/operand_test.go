/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package agent

import (
	"context"
	"testing"

	runutils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/secrets"
	kubeclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	schemeinit "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestOperandOwnsIndependentAssetIdentityPolicies(t *testing.T) {
	runaiConfig := runutils.GetDefaultRunaiConfig()
	installationConfig := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secrets.InstallationConfigurationSecretName,
			Namespace: runv1.DefaultNamespace,
		},
		Data: map[string][]byte{
			secrets.InstallationConfigurationSecretName: []byte("{}"),
		},
	}
	fakeClient := fake.NewClientBuilder().
		WithScheme(schemeinit.Scheme()).
		WithObjects(installationConfig).
		Build()
	hybridClient := kubeclient.New(fakeClient, fakeClient, nil)

	operandResources := Operand(hybridClient, runaiConfig).Resources()

	policyNames := map[string]bool{}
	bindingNames := map[string]bool{}
	for _, resource := range operandResources {
		object, err := resource.Object(context.Background())
		require.NoError(t, err)
		switch typedObject := object.(type) {
		case *admissionv1.ValidatingAdmissionPolicy:
			policyNames[typedObject.Name] = true
		case *admissionv1.ValidatingAdmissionPolicyBinding:
			bindingNames[typedObject.Name] = true
		}
	}

	assert.Equal(t, map[string]bool{
		"runai-secret-identity": true,
		"runai-pvc-identity":    true,
	}, policyNames)
	assert.Equal(t, map[string]bool{
		"runai-secret-identity-binding": true,
		"runai-pvc-identity-binding":    true,
	}, bindingNames)
}
