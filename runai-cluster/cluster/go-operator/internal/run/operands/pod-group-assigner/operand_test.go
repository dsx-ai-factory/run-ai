/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_group_assigner

import (
	"context"
	"fmt"
	"testing"

	runutils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type hybridClient struct{ client.Client }

func (c hybridClient) Uncached() client.Client { return c.Client }

func (c hybridClient) NewDiscoveryClient() (discovery.DiscoveryInterface, error) { return nil, nil }

func controller(apiVersion, kind string) []metav1.OwnerReference {
	isController := true
	return []metav1.OwnerReference{{
		APIVersion: apiVersion, Kind: kind, Name: "owner", UID: "uid", Controller: &isController,
	}}
}

func TestOperandDeletesOnlyObjectsKRMDoesNotControl(t *testing.T) {
	krmOwned := controller("kai.resources/v1alpha1", "KRMConfig")
	runaiOwned := controller("run.ai/v1", "RunaiConfig")

	for _, tc := range []struct {
		name   string
		owners []metav1.OwnerReference
		exists bool
		want   []string
	}{
		{name: "nothing left behind", exists: false, want: nil},
		{name: "2.x leftovers without owners", exists: true,
			want: []string{"*v1.Deployment/pod-group-assigner", "*v1.Service/pod-group-assigner"}},
		{name: "2.x leftovers controlled by the RunaiConfig", exists: true, owners: runaiOwned,
			want: []string{"*v1.Deployment/pod-group-assigner", "*v1.Service/pod-group-assigner"}},
		{name: "KRM already controls them", exists: true, owners: krmOwned, want: nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runaiConfig := runutils.GetDefaultRunaiConfig()
			builder := fake.NewClientBuilder().WithScheme(scheme.Scheme)
			if tc.exists {
				objectMeta := metav1.ObjectMeta{Name: name, Namespace: runaiConfig.Namespace, OwnerReferences: tc.owners}
				builder = builder.WithObjects(
					&appsv1.Deployment{ObjectMeta: objectMeta},
					&corev1.Service{ObjectMeta: objectMeta},
				)
			}
			op := Operand(hybridClient{builder.Build()}, runaiConfig)

			assert.False(t, op.Enabled())
			assert.Equal(t, tc.want, objectsToDelete(t, op.Resources(), runaiConfig))
		})
	}
}

func TestObjectKeepsItsFirstAnswer(t *testing.T) {
	runaiConfig := runutils.GetDefaultRunaiConfig()
	kubeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(
		&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: runaiConfig.Namespace}},
	).Build()
	resource := Operand(hybridClient{kubeClient}, runaiConfig).Resources()[1]

	first, err := resource.Object(context.Background())
	require.NoError(t, err)
	require.NotNil(t, first)

	require.NoError(t, kubeClient.Delete(context.Background(), first))
	second, err := resource.Object(context.Background())
	require.NoError(t, err)
	assert.Same(t, first, second)
}

func objectsToDelete(t *testing.T, resources []operands.Resource, runaiConfig *runv1.RunaiConfig) []string {
	t.Helper()
	var objects []string
	for _, resource := range resources {
		assert.False(t, resource.Enabled())
		object, err := resource.Object(context.Background())
		require.NoError(t, err)
		if object == nil {
			continue
		}
		assert.Equal(t, runaiConfig.Namespace, object.GetNamespace())
		objects = append(objects, fmt.Sprintf("%T/%s", object, object.GetName()))
	}
	return objects
}
