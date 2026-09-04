/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package application_controller

import (
	"context"
	"testing"

	runutils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/secrets"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestOperandIncludesApplicationIdentityPolicy(t *testing.T) {
	runaiConfig := runutils.GetDefaultRunaiConfig()
	scheme := scheme_init.Scheme()
	installationConfig := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secrets.InstallationConfigurationSecretName,
			Namespace: runv1.DefaultNamespace,
		},
		Data: map[string][]byte{
			secrets.InstallationConfigurationSecretName: []byte(`global:
  controlPlane:
    clusterUUID: test-cluster
    config:
      agent:
        client:
          name: test-client
          secret: test-secret`),
		},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(installationConfig).Build()
	hybridClient := kube_client.New(fakeClient, fakeClient, nil)

	resources, err := Operand().ResourcesForRunaiConfig(context.Background(), hybridClient, runaiConfig)

	require.NoError(t, err)
	require.Len(t, resources, 7)
	var policyFound bool
	var bindingFound bool
	for _, resource := range resources {
		switch object := resource.Object.(type) {
		case *admissionv1.ValidatingAdmissionPolicy:
			policyFound = object.Name == "runai-application-identity"
		case *admissionv1.ValidatingAdmissionPolicyBinding:
			bindingFound = object.Name == "runai-application-identity-binding"
		}
	}
	assert.True(t, policyFound)
	assert.True(t, bindingFound)
}
