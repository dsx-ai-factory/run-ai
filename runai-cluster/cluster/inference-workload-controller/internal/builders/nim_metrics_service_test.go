/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestBuildNIMMetricsService(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	ctx := context.Background()

	servingPort := &corev1.ContainerPort{
		ContainerPort: 8000,
		Protocol:      corev1.ProtocolTCP,
	}

	mockWorkload := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workload",
			Namespace: "test-ns",
		},
	}

	t.Run("creates service for NIM workload", func(t *testing.T) {
		svc, err := BuildNIMMetricsService(ctx, mockWorkload, "test-nim", "test-ns", servingPort, k8sClient)

		assert.NoError(t, err)
		assert.NotNil(t, svc)
		assert.Equal(t, "test-nim-metrics", svc.Name)
		assert.Equal(t, "test-ns", svc.Namespace)
		assert.Equal(t, corev1.ServiceTypeClusterIP, svc.Spec.Type)
		assert.Equal(t, "metrics", svc.Spec.Ports[0].Name)
		assert.Equal(t, int32(8000), svc.Spec.Ports[0].Port)
		assert.Equal(t, "test-nim", svc.Spec.Selector[constants.LabelWorkloadName])
		assert.Equal(t, constants.BoolTrueAsString, svc.Labels[constants.LabelNIMWorkload])
		assert.Equal(t, constants.BoolTrueAsString, svc.Labels[constants.LabelNIMWorkloadMetricsService])
	})

	t.Run("returns nil when serving port is nil", func(t *testing.T) {
		svc, err := BuildNIMMetricsService(ctx, mockWorkload, "test-nim", "test-ns", nil, k8sClient)

		assert.NoError(t, err)
		assert.Nil(t, svc)
	})
}
