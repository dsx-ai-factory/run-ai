/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mnnvl

import (
	"fmt"

	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
)

const (
	ResourceClaimTemplatePrefix = "cd"
	ResourceClaimName           = "compute-domain"
)

func CreateResourceClaimTemplate(params ComputeDomainWorkloadParams) *nvapi.ComputeDomainResourceClaimTemplate {
	rctName := ComposeResourceClaimTemplateName(params.WorkloadName, params.WorkloadInitials)

	return &nvapi.ComputeDomainResourceClaimTemplate{
		Name: rctName,
	}
}

func GetResourceClaimName() string {
	return ResourceClaimName
}

func ComposeResourceClaimTemplateName(workloadName, workloadInitials string) string {
	return fmt.Sprintf("%s-%s-%s", ResourceClaimTemplatePrefix, workloadInitials, workloadName)
}
