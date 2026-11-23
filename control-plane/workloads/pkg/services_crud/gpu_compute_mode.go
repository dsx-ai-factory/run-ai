/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpumode"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
)

func validateNativeGpuMode(spec any) *workloadsApi.SubmissionErrorV2 {
	values, err := gojsonq.LoadJson(spec)
	if err != nil {
		return workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, "failed to load GPU compute settings")
	}
	if hasGpuSelection(values) {
		return workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, "GPU type selection requires workloads v2 submission")
	}

	explicit := false
	gpumode.VisitCompute(values, func(compute map[string]interface{}) {
		if compute["gpuComputeMode"] != nil && gpumode.HasGPU(compute) {
			explicit = true
		}
	})
	if explicit {
		return workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, "compute.gpuComputeMode requires workloads v2 submission")
	}
	return nil
}

// The legacy controllers cannot enforce first-class accelerator placement.
func hasGpuSelection(spec map[string]interface{}) bool {
	if spec["gpuType"] != nil || spec["accelerator"] != nil {
		return true
	}
	for _, role := range []string{"leader", "worker"} {
		if nested, ok := spec[role].(map[string]interface{}); ok && (nested["gpuType"] != nil || nested["accelerator"] != nil) {
			return true
		}
	}
	return false
}
