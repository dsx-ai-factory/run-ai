/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/clusterservice"
	orgunit "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/org_unit"
	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
)

const (
	gpuInventoryUnavailable     = "GPU_INVENTORY_UNAVAILABLE"
	gpuTypeNotFound             = "GPU_TYPE_NOT_FOUND_IN_PROJECT_POOLS"
	unavailableInventoryMessage = "Project accelerator inventory could not be verified. The requested placement was preserved."
)

type acceleratorSpec interface {
	GetGpuTypeOk() (*string, bool)
	UnsetGpuType()
	GetAcceleratorOk() (*workloadsApi.AcceleratorSelection, bool)
	UnsetAccelerator()
}

func resolveWorkloadAccelerators(ctx context.Context, projectID, clusterID string, spec *workloadsApi.SupersetSpec, master *workloadsApi.MasterSpec) ([]workloadsApi.SubmissionWarning, *workloadsApi.SubmissionErrorV2) {
	if spec == nil {
		return nil, nil
	}
	type roleSpec struct {
		field string
		spec  acceleratorSpec
	}
	roles := []roleSpec{{"spec", spec}}
	if master != nil {
		roles = append(roles, roleSpec{"masterSpec", master})
	}
	if leader := spec.Leader.Get(); leader != nil {
		roles = append(roles, roleSpec{"spec.leader", leader})
	}
	if worker := spec.Worker.Get(); worker != nil {
		roles = append(roles, roleSpec{"spec.worker", worker})
	}
	var warnings []workloadsApi.SubmissionWarning
	getPools := sync.OnceValues(func() ([]string, error) {
		return orgunit.Get().GetProjectGPUPoolIDs(ctx, projectID)
	})
	for _, role := range roles {
		model := ""
		if value, _ := role.spec.GetGpuTypeOk(); value != nil {
			if *value == "" {
				return nil, workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, role.field+".gpuType must be a nonempty Kubernetes label value")
			}
			model = *value
		}
		selection, _ := role.spec.GetAcceleratorOk()
		if selection != nil {
			// Validate explicit fields before clearing an unused model key.
			if _, _, err := composer.ResolveGPUSelection(nil, selection); err != nil {
				return nil, workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, role.field+"."+err.Message)
			}
		}
		if strings.EqualFold(model, "any") {
			role.spec.UnsetGpuType()
			model = ""
		}
		if model == "" {
			if selection == nil || selection.ResourceName.Get() == nil {
				role.spec.UnsetAccelerator()
				continue
			}
			selection.UnsetDeviceLabelKey()
		}
		var modelValue *string
		if model != "" {
			modelValue = &model
		}
		resource, requirement, validationErr := composer.ResolveGPUSelection(modelValue, selection)
		if validationErr != nil {
			return nil, workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, role.field+"."+validationErr.Message)
		}
		labels := map[string]string{clusterconstants.AcceleratorResourceNodeLabel(resource): "true"}
		warning := workloadsApi.SubmissionWarning{Field: role.field + ".accelerator.resourceName", Code: gpuTypeNotFound,
			Message: fmt.Sprintf("Resource %s was not found in the project's accessible pools.", resource)}
		if requirement != nil {
			warning.Field = role.field + ".gpuType"
			warning.Message = fmt.Sprintf("No %s with label %s=%s was found in the project's accessible pools.", resource, requirement.Key, model)
			if existing, ok := labels[requirement.Key]; ok && existing != model {
				warnings = append(warnings, warning)
				continue
			}
			labels[requirement.Key] = model
		}
		pools, err := getPools()
		if err == nil && len(pools) != 0 {
			var matches bool
			matches, err = clusterservice.Get().MatchGPUPlacement(ctx, clusterID, pools, labels)
			if err == nil && matches {
				continue
			}
		}
		if err != nil {
			warning.Code = gpuInventoryUnavailable
			warning.Message = unavailableInventoryMessage
		}
		warnings = append(warnings, warning)
	}
	return warnings, nil
}
