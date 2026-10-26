/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package project_controller

import (
	"context"
	"fmt"
	"testing"

	runutils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOperandOnlyDeletesRunaiBuiltObjects(t *testing.T) {
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
		"*v1.Deployment/runai-project-controller",
		"*v1.Service/runai-project-controller",
		"*v1.ServiceAccount/runai-project-controller",
		"*v1.ServiceMonitor/runai-project-controller",
		"*v1.ConfigMap/runai-project-controller-rolebindings-plugin",
	}, objects)
}
