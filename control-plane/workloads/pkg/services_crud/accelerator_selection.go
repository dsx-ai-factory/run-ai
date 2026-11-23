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

	orgapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
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
	SetAccelerator(workloadsApi.AcceleratorSelection)
	UnsetAccelerator()
}

func resolveWorkloadAccelerators(ctx context.Context, projectID string, spec *workloadsApi.SupersetSpec, master *workloadsApi.MasterSpec) ([]workloadsApi.SubmissionWarning, *workloadsApi.SubmissionErrorV2) {
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
	var inventory *orgapi.ProjectAccelerators
	loaded := false
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
			// Validate explicit fields before an omitted model or an inferred key can replace them.
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
		if !loaded {
			var inventoryErr *orgapi.Error
			inventory, inventoryErr = orgunit.Get().GetProjectAccelerators(ctx, projectID)
			if inventoryErr != nil {
				inventory = nil
			}
			loaded = true
		}
		warning, err := resolveAccelerator(role.field, role.spec, selection, model, inventory)
		if err != nil {
			return nil, workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, err.Error())
		}
		if warning == nil {
			continue
		}
		// The picker only lists identifiable GPUs. An explicit selection may
		// refer to a different extended resource or a custom model label.
		if warning.Code == gpuTypeNotFound && selection != nil && (model == "" || selection.GetDeviceLabelKey() != "") {
			resource := selection.GetResourceName()
			if resource == "" {
				resource = composer.NvidiaResourceName
			}
			key, value := selection.GetDeviceLabelKey(), model
			if model == "" {
				key, value = clusterconstants.AcceleratorResourceNodeLabel(resource), "true"
			}
			match, lookupErr := orgunit.Get().MatchProjectAccelerator(ctx, projectID, resource, key, value)
			if lookupErr != nil {
				warning.Code = gpuInventoryUnavailable
				warning.Message = unavailableInventoryMessage
			} else if match {
				continue
			}
		}
		warnings = append(warnings, *warning)
	}
	return warnings, nil
}

func resolveAccelerator(field string, spec acceleratorSpec, selection *workloadsApi.AcceleratorSelection, model string, inventory *orgapi.ProjectAccelerators) (*workloadsApi.SubmissionWarning, error) {
	resourceName := composer.NvidiaResourceName
	if selection != nil && selection.ResourceName.Get() != nil {
		resourceName = selection.GetResourceName()
	}
	key := selection.GetDeviceLabelKey()
	var option *orgapi.AcceleratorOption
	if inventory != nil {
		for i := range inventory.Items {
			if inventory.Items[i].ResourceName == resourceName {
				option = &inventory.Items[i]
				break
			}
		}
	}
	if model != "" && key == "" && resourceName != composer.NvidiaResourceName {
		key = discoveredDeviceLabelKey(option, model)
		if key == "" {
			return nil, fmt.Errorf("%s.accelerator.deviceLabelKey is required because the label key for %q on %s cannot be determined", field, model, resourceName)
		}
		selection.SetDeviceLabelKey(key)
		spec.SetAccelerator(*selection)
	}
	var modelValue *string
	if model != "" {
		modelValue = &model
	}
	_, predicate, validationErr := composer.ResolveGPUSelection(modelValue, selection)
	if validationErr != nil {
		return nil, fmt.Errorf("%s: %s", field, validationErr.Message)
	}
	modelSeen := false
	if predicate != nil {
		key = predicate.Key
		if option != nil {
			for _, device := range option.Devices {
				if device.Key == key && device.Value == model {
					modelSeen = true
					break
				}
			}
		}
	}
	warning := workloadsApi.SubmissionWarning{Field: field + ".accelerator.resourceName"}
	if model != "" {
		warning.Field = field + ".gpuType"
	}
	switch {
	case inventory == nil:
		warning.Code = gpuInventoryUnavailable
		warning.Message = unavailableInventoryMessage
	case option == nil || model != "" && !modelSeen:
		warning.Code = gpuTypeNotFound
		warning.Message = fmt.Sprintf("Resource %s was not found in the project's accessible pools.", resourceName)
		if model != "" {
			warning.Message = fmt.Sprintf("No %s with label %s=%s was found in the project's accessible pools.", resourceName, key, model)
		}
	default:
		return nil, nil
	}
	return &warning, nil
}

// Prefer evidence for the exact model. A resource with one observed product label
// also provides an unambiguous key for a model that has not appeared in inventory yet.
func discoveredDeviceLabelKey(option *orgapi.AcceleratorOption, model string) string {
	if option == nil {
		return ""
	}
	all, matching := map[string]bool{}, map[string]bool{}
	for _, device := range option.Devices {
		all[device.Key] = true
		if device.Value == model {
			matching[device.Key] = true
		}
	}
	if len(matching) > 0 {
		all = matching
	}
	if len(all) == 1 {
		for key := range all {
			return key
		}
	}
	return ""
}
