/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package legacy_researcher_service

import (
	"context"
	"fmt"
	"testing"

	runutils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/utils"
	cu "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"
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

	// Built from a separately encoded stem, not the operand's constants, so a
	// wrong encoded name fails here instead of silently leaving an orphan.
	stem := mustDecodeBase64("cmVzZWFyY2hlcg==")
	component := stem + "-service"
	assert.ElementsMatch(t, []string{
		"*v1.Deployment/" + component,
		"*v1.Service/" + component,
		"*v1.ServiceAccount/" + component,
		"*v1.VerticalPodAutoscaler/" + component,
		"*v1.Ingress/" + component + "-ingress",
		"*v1.HTTPRoute/" + component + "-httproute",
		"*v1.Route/" + stem + "-cli-route",
		"*v1.Route/runai-" + stem + "-route",
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
