/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package researcher_service

import (
	"context"
	"fmt"
	"testing"

	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/scheme-init"
	runutils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/utils"
	cu "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestOperandOnlyDeletesRetiredObjects(t *testing.T) {
	runaiConfig := runutils.GetDefaultRunaiConfig()
	op := Operand(nil, runaiConfig)

	assert.False(t, op.Enabled())

	var objects []string
	for _, resource := range op.Resources() {
		assert.False(t, resource.Enabled())
		object, err := resource.Object(context.Background())
		require.NoError(t, err)
		assert.Equal(t, runaiConfig.Namespace, object.GetNamespace())
		objects = append(objects, fmt.Sprintf("%T/%s", object, object.GetName()))
	}

	assert.ElementsMatch(t, []string{
		"*v1.Deployment/researcher-service",
		"*v1.Service/researcher-service",
		"*v1.ServiceAccount/researcher-service",
		"*v1.VerticalPodAutoscaler/researcher-service",
		"*v1.Ingress/researcher-service-ingress",
		"*v1.HTTPRoute/researcher-service-httproute",
		"*v1.Route/researcher-cli-route",
		"*v1.Route/runai-researcher-route",
	}, objects)
}

func TestOperandDeletesLeftoversOnUpgrade(t *testing.T) {
	for _, tc := range []struct {
		name          string
		seedLeftovers bool
	}{
		{name: "upgraded cluster still running the retired component", seedLeftovers: true},
		{name: "fresh install", seedLeftovers: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			op := Operand(nil, runutils.GetDefaultRunaiConfig())

			var objects []client.Object
			for _, resource := range op.Resources() {
				object, err := resource.Object(ctx)
				require.NoError(t, err)
				objects = append(objects, object)
			}

			builder := fake.NewClientBuilder().WithScheme(scheme_init.Scheme())
			if tc.seedLeftovers {
				builder = builder.WithObjects(objects...)
			}
			k8sClient := builder.Build()

			require.NoError(t, cu.DeleteResources(k8sClient, ctx, op.Resources()))

			for _, object := range objects {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(object), object.DeepCopyObject().(client.Object))
				assert.Truef(t, errors.IsNotFound(err), "%T/%s should be gone, got: %v", object, object.GetName(), err)
			}
		})
	}
}
