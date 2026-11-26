/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"

	orgapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/clusterservice"
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

func resolveWorkloadAccelerators(ctx context.Context, getProject func() *orgapi.Project, clusterID string, spec *workloadsApi.SupersetSpec, master *workloadsApi.MasterSpec) ([]workloadsApi.SubmissionWarning, *workloadsApi.SubmissionErrorV2) {
	if spec == nil {
		return nil, nil
	}
	type roleSpec struct {
		field string
		spec  acceleratorSpec
		pools []string
	}
	roles := []roleSpec{{"spec", spec, spec.NodePools}}
	if master != nil {
		roles = append(roles, roleSpec{"masterSpec", master, master.NodePools})
	}
	if leader := spec.Leader.Get(); leader != nil {
		roles = append(roles, roleSpec{"spec.leader", leader, spec.NodePools})
	}
	if worker := spec.Worker.Get(); worker != nil {
		roles = append(roles, roleSpec{"spec.worker", worker, spec.NodePools})
	}
	var warnings []workloadsApi.SubmissionWarning
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
			warnings = append(warnings, workloadsApi.SubmissionWarning{
				Field: role.field + ".accelerator.resourceName", Code: gpuInventoryUnavailable,
				Message: "Project inventory cannot verify a resource-only request. The requested resource was preserved.",
			})
			continue
		}
		_, requirement, validationErr := composer.ResolveGPUSelection(&model, selection)
		if validationErr != nil {
			return nil, workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, role.field+"."+validationErr.Message)
		}
		labels := map[string]string{requirement.Key: model}
		warning := workloadsApi.SubmissionWarning{
			Field: role.field + ".gpuType", Code: gpuTypeNotFound,
			Message: fmt.Sprintf("No node with label %s=%s was found in the selected project pools.", requirement.Key, model),
		}
		project := getProject()
		if project == nil {
			warning.Code = gpuInventoryUnavailable
			warning.Message = unavailableInventoryMessage
		} else if pools := selectedGPUPoolIDs(project, role.pools); len(pools) > 0 {
			matches, err := clusterservice.Get().MatchGPUPlacement(ctx, clusterID, pools, labels)
			if err != nil {
				warning.Code = gpuInventoryUnavailable
				warning.Message = unavailableInventoryMessage
			} else if matches {
				continue
			}
		}
		warnings = append(warnings, warning)
	}
	return warnings, nil
}

func selectedGPUPoolIDs(project *orgapi.Project, names []string) []string {
	if len(names) == 0 {
		names = getDefaultNodepools(project)
	}
	var ids []string
	for _, resource := range project.Resources {
		pool := resource.NodePool.Get()
		if pool != nil && pool.Id != "" && slices.Contains(names, pool.GetName()) {
			ids = append(ids, pool.Id)
		}
	}
	return ids
}
