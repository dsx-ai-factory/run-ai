/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	orgUnitApi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/metrics"
	"k8s.io/apimachinery/pkg/types"

	"github.com/rs/zerolog/log"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	runaiBackendClient "github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/common"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	schedv2alpha2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DepartmentsHandler struct {
	*CommonHandler
	lastUpdated time.Time
}

func NewDepartmentsHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client,
	cachedClient client.Reader) (ResourceHandler, error) {
	var prototype client.Object = &kaiorg.Department{}
	var inspector ObjectInspector = KaiDepartmentObjectInspector(true)

	commonHandler, err := NewCommonHandler(runaiBackendClient, runaiClient, cachedClient,
		prototype, inspector, HandlingOptions{})
	if err != nil {
		return nil, err
	}

	return &DepartmentsHandler{
		CommonHandler: commonHandler,
		lastUpdated:   time.Time{},
	}, nil
}

func (h *DepartmentsHandler) Handle() error {
	log.Debug().Msgf("Starting to sync departments")

	backendDepartments, err := h.getOrgUnitDepartments()
	if err != nil {
		return err
	}

	departmentsToAddUpdate, departmentsToDelete, lastUpdated, err := h.getClusterObjects(backendDepartments)

	metrics.GetMetricsExporter().AddResourceAction(h.MetricLabel(), metrics.ActionCreateOrUpdate, len(departmentsToAddUpdate))
	metrics.GetMetricsExporter().AddResourceAction(h.MetricLabel(), metrics.ActionDelete, len(departmentsToDelete))

	_, innerErr := h.DeleteObjectsById(departmentsToDelete, schedv2alpha2.DepartmentIdLabel)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}

	_, _, innerErr = h.CreateOrUpdateObjects(departmentsToAddUpdate)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}

	if err != nil {
		log.Error().Msgf("Failed to handle departments, error: %s", err.Error())
		return err
	}

	h.lastUpdated = lastUpdated

	log.Debug().Msgf("Successfully handled departments")
	return err
}

func (h *DepartmentsHandler) MetricLabel() string {
	return "departments"
}

func (h *DepartmentsHandler) getOrgUnitDepartments() ([]orgUnitApi.DepartmentSync, error) {
	departments := []orgUnitApi.DepartmentSync{}

	offset := int32(0)
	limit := int32(config.Config().DepartmentsLimit)

	for {
		log.Debug().Msgf("Getting departments from org-unit; offset: <%d>, limit: <%d>, lastUpdated: <%v>",
			offset, limit, h.lastUpdated)

		departmentsBulk, err := h.backendClient.OrgUnitClient.GetDepartmentsSync(
			h.backendClient.Context(), config.Config().ClusterUuid, limit, offset, h.lastUpdated)

		if err != nil {
			log.Error().Msgf("Failed to get departments from org-unit: err: %v", err)
			return departments, err
		}

		departments = append(departments, departmentsBulk.Departments...)

		if len(departmentsBulk.Departments) < int(limit) {
			log.Debug().Msgf("Finished getting all departments")
			break
		}

		offset = offset + limit
	}

	return departments, nil
}

func (h *DepartmentsHandler) getClusterObjects(departments []orgUnitApi.DepartmentSync) (
	[]client.Object, map[string]types.NamespacedName, time.Time, error) {
	var err error
	departmentsToAddUpdate := []client.Object{}
	departmentsToDelete := map[string]types.NamespacedName{}
	lastUpdated := h.lastUpdated

	for _, department := range departments {
		if department.DeletedAt.IsSet() && department.DeletedAt.Get() != nil && department.Name != nil {
			departmentsToDelete[department.Id] = types.NamespacedName{Name: *department.Name, Namespace: ""}
		} else {
			clusterDep, convErr := convertOrgUnitToKaiDepartment(department)
			if convErr != nil {
				err = multierror.Append(err, convErr)
				continue
			}
			departmentsToAddUpdate = append(departmentsToAddUpdate, clusterDep)
		}

		if department.UpdatedAt.After(lastUpdated) {
			lastUpdated = department.UpdatedAt
		}
	}

	return departmentsToAddUpdate, departmentsToDelete, lastUpdated, err
}

func convertOrgUnitToKaiDepartment(backendDepartment orgUnitApi.DepartmentSync) (*kaiorg.Department, error) {
	department := &kaiorg.Department{}
	department.Name = common.GetStringVal(backendDepartment.Name)
	department.Labels = map[string]string{
		schedv2alpha2.DepartmentIdLabel: backendDepartment.Id,
	}
	department.Spec.Queues = convertToKaiQueueConfigs(fmt.Sprintf("q-%s", backendDepartment.Id), backendDepartment.Resources)

	department.Annotations = map[string]string{
		constants.UpdatedAtAnnotation: backendDepartment.UpdatedAt.UTC().Format(time.RFC3339),
	}

	return department, nil
}

// kaiDepartmentManagedAnnotations are the annotation keys the agent owns on a KAI
// Department. They are the single source of truth for both writing
// (convertOrgUnitToKaiDepartment) and drift detection (KaiDepartmentObjectInspector).
var kaiDepartmentManagedAnnotations = []string{
	constants.UpdatedAtAnnotation,
}

type KaiDepartmentObjectInspector bool

func (d KaiDepartmentObjectInspector) DeepCopySpec(from, to client.Object) {
	fromDepartment, ok := from.(*kaiorg.Department)
	if !ok {
		// shouldn't happen because we convert objects to the correct type
		return
	}
	toDepartment, ok := to.(*kaiorg.Department)
	if !ok {
		return
	}

	fromDepartment.Spec.DeepCopyInto(&toDepartment.Spec)
	copyOrgUnitIdLabel(fromDepartment, toDepartment, schedv2alpha2.DepartmentIdLabel)
	copyManagedAnnotations(fromDepartment, toDepartment, kaiDepartmentManagedAnnotations)
}

func (d KaiDepartmentObjectInspector) IsUpToDate(left, right client.Object) bool {
	leftDepartment, ok := left.(*kaiorg.Department)
	if !ok {
		// shouldn't happen because we convert objects to the correct type
		return false
	}
	rightDepartment, ok := right.(*kaiorg.Department)
	if !ok {
		return false
	}

	if !specJsonEqual(leftDepartment.Spec, rightDepartment.Spec) {
		return false
	}

	if !managedAnnotationsEqual(leftDepartment.Annotations, rightDepartment.Annotations, kaiDepartmentManagedAnnotations) {
		return false
	}

	return idLabelEqual(leftDepartment.Labels, rightDepartment.Labels, schedv2alpha2.DepartmentIdLabel)
}
