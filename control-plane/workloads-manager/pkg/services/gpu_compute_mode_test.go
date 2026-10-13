/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/middleware"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/stretchr/testify/require"
)

func TestNimModeDefaultAndPersistence(t *testing.T) {
	var src api.NimServiceSpec
	require.NoError(t, json.Unmarshal([]byte(`{"compute":{"gpuMemoryPortionRequest":0.25}}`), &src))
	defaults := map[string]interface{}{"compute": map[string]interface{}{"gpuComputeMode": "sm-sharing"}}
	values, err := MutateWorkload(&src, nil, defaults)
	require.Nil(t, err)
	require.NotContains(t, values["compute"], "computeMode")
	_, err = validateMutatedSpecVsPolicy(nil, defaults, values)
	require.Nil(t, err)
	var resolved api.NimServiceSpec
	require.NoError(t, convutils.CopySpecs(&resolved, values))
	service := &NimServicesImpl{}
	ctx, ctxErr := service.updateContextWithMergedSpec(context.Background(), &api.NimServiceMetadataCreateParams{Name: "model", ProjectId: "p1"}, resolved)
	require.NoError(t, ctxErr)
	var saved api.NimServiceCreateRequest
	require.NoError(t, json.Unmarshal(middleware.GetRequestBody(ctx), &saved))
	require.NotContains(t, string(middleware.GetRequestBody(ctx)), `"computeMode"`)
	require.Equal(t, api.GPUCOMPUTEMODE_SM_SHARING, *saved.Spec.Compute.GpuComputeMode)
	require.Equal(t, api.GPUCOMPUTEMODE_SM_SHARING, *legacyCompatibleNimSpec(saved.Spec).Compute.GpuComputeMode)
}
