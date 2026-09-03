/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package patch

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/types"
)

func RemoveGpuResourceFromLimits(requirements *v1.ResourceRequirements) {
	_, found := requirements.Limits[types.GpuResource]
	if found {
		delete(requirements.Limits, types.GpuResource)
	}
}

func AddResourcesToRequests(requirements *v1.ResourceRequirements, resourceName v1.ResourceName,
	quantity resource.Quantity) {
	if requirements.Requests == nil {
		requirements.Requests = v1.ResourceList{}
	}
	requirements.Requests[resourceName] = quantity
}

func AddResourcesToLimits(requirements *v1.ResourceRequirements, resourceName v1.ResourceName,
	quantity resource.Quantity) {
	if requirements.Limits == nil {
		requirements.Limits = v1.ResourceList{}
	}
	requirements.Limits[resourceName] = quantity
}
