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
	resourceClaimTemplatePrefix = "cd"
	resourceClaimName           = "compute-domain"
)

func createResourceClaimTemplate(params computeDomainWorkloadParams) *nvapi.ComputeDomainResourceClaimTemplate {
	rctName := composeResourceClaimTemplateName(params.WorkloadName, params.WorkloadInitials)

	return &nvapi.ComputeDomainResourceClaimTemplate{
		Name: rctName,
	}
}

func getResourceClaimName() string {
	return resourceClaimName
}

func composeResourceClaimTemplateName(workloadName, workloadInitials string) string {
	return fmt.Sprintf("%s-%s-%s", resourceClaimTemplatePrefix, workloadInitials, workloadName)
}
