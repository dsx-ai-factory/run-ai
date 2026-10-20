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
