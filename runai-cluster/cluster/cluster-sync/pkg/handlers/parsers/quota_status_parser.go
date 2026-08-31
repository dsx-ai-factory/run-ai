/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"strings"

	org_unit_openapi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

const (
	gpuResourceNameSuffix = "/gpu"
	ProjectIdLabel        = "runai/project-id"
	cpuMultiplyFactor     = 1000    // from Cores to MilliCores
	memoryDivisionFactor  = 1000000 // from bytes to MB
)

func convertResourcesListToQuotaStatusResourceList(resourceList v1.ResourceList) *org_unit_openapi.QuotaStatusResource {
	result := &org_unit_openapi.QuotaStatusResource{}
	for resourceName, quantity := range resourceList {
		if resourceName == v1.ResourceCPU {
			// convert from Cores to MilliCores
			result.Cpu = ptr.To(quantity.AsApproximateFloat64() * cpuMultiplyFactor)
		}
		if resourceName == v1.ResourceMemory {
			// convert from Bytes to MB
			result.Memory = ptr.To(quantity.AsApproximateFloat64() / memoryDivisionFactor)
		}
		if strings.HasSuffix(string(resourceName), gpuResourceNameSuffix) {
			result.Gpu = ptr.To(quantity.AsApproximateFloat64())
		}
	}

	return result
}
