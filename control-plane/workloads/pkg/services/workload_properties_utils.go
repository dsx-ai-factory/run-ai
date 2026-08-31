/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
	karta "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"golang.org/x/exp/slices"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/workloadutils"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/clients/policy"
	"github.com/run-ai/runai/control-plane/workloads/pkg/common"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-clients/policy-service-client/structs"
	"github.com/run-ai/runai/runai-cluster/common/utils"
	"github.com/run-ai/runai/runai-common-packages/dao-utils/validators"

	"k8s.io/utils/ptr"
)

func WorkloadTypeConfigsDbToApi(workloadTypes []models.ExternalWorkloadTypeConfigView) []api.WorkloadTypeConfig {
	result := make([]api.WorkloadTypeConfig, len(workloadTypes))
	for i, workloadType := range workloadTypes {
		result[i] = WorkloadTypeConfigDbToApi(workloadType)
	}
	return result
}

func WorkloadTypeConfigDbToApi(workloadType models.ExternalWorkloadTypeConfigView) api.WorkloadTypeConfig {
	resourceInterfaces, kartas, err := parseKartasDbToApi(workloadType.Kartas)
	if err != nil {
		log.Error().Msgf("Failed to parse kartas for workload type %s, error: %s", workloadType.Name, err.Error())
		resourceInterfaces = []api.WorkloadTypeResourceInterfacesWithVersionInner{}
		kartas = []api.WorkloadTypeKartasWithVersionInner{}
	}

	result := api.WorkloadTypeConfig{
		CategoryId:     workloadType.CategoryId.String(),
		PriorityId:     workloadType.PriorityId.String(),
		Name:           workloadType.Name,
		Id:             workloadType.ID.String(),
		CategoryName:   workloadType.CategoryName,
		PriorityName:   workloadType.PriorityName,
		Preemptibility: api.WorkloadTypePreemptibility(workloadType.Preemptibility),
		CreatedAt:      workloadType.CreatedAt,
		CreatedBy:      utils.Deref(workloadType.CreatedBy),
		UpdatedAt:      workloadType.UpdatedAt,
		UpdatedBy:      utils.Deref(workloadType.UpdatedBy),
		Source:         workloadType.Source,

		Group:              workloadType.Group,
		ResourceInterfaces: resourceInterfaces,
		Kartas:             kartas,
	}

	if !result.Preemptibility.IsValid() {
		// default to PREEMPTIBLE if invalid
		result.Preemptibility = api.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE
	}

	if len(workloadType.Statuses) > 0 {
		clusterStatuses := make([]api.WorkloadTypeStatus, len(workloadType.Statuses))
		for i, status := range workloadType.Statuses {
			clusterStatuses[i] = api.WorkloadTypeStatus{
				Phase:      status.Phase,
				ClusterId:  status.ClusterId.String(),
				UpdatedAt:  status.UpdatedAt,
				Conditions: convutils.ConvertSliceTypes(status.Conditions, conditionModelToApi),
				Version:    status.Version,
			}
		}

		result.ClusterStatuses = &clusterStatuses
	}

	return result
}

func GetWorkloadConfigOfTenant(ctx context.Context) (*map[string]*structs.PolicyWorkloadTypeConfig, *api.Error) {
	policiesWorkloadTypeConfig, policyErr := policy.Get().GetWorkloadConfigOfTenant(ctx)
	if policyErr != nil {
		return nil, api.NewError(int(policyErr.Code), policyErr.Message, nil)
	}
	for _, defaultTypeConfig := range workloadutils.DefaultRunaiWorkloadTypesConfig {
		categoryName := defaultTypeConfig.Category
		priorityName := defaultTypeConfig.Priority
		runaiWorkloadKind := defaultTypeConfig.WorkloadKind
		config := (*policiesWorkloadTypeConfig)[runaiWorkloadKind]

		if config == nil {
			config = &structs.PolicyWorkloadTypeConfig{}
		}

		if config.CategoryName == nil {
			config.CategoryName = &categoryName
		}

		if config.PriorityName == nil {
			config.PriorityName = &priorityName
		}

		(*policiesWorkloadTypeConfig)[runaiWorkloadKind] = config
	}
	return policiesWorkloadTypeConfig, nil
}

func GetRunaiWorkloadTypeConfig(ctx context.Context, tenantId uint, workloadType *models.WorkloadType) (*api.WorkloadTypeConfig, *api.Error) {
	policiesWorkloadTypeConfig, policyErr := policy.Get().GetWorkloadConfigOfTenant(ctx)
	if policyErr != nil {
		return nil, api.NewError(int(policyErr.Code), policyErr.Message, nil)
	}

	runaiWorkloadType, err := common.NewWorkloadTypeFromKind(workloadType.Name)
	if err != nil || !runaiWorkloadType.IsNative() {
		return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("Runai workload type config for %s not found", workloadType.Name), nil)
	}
	runaiWorkloadTypeConfig, ok := workloadutils.DefaultRunaiWorkloadTypesConfig[string(runaiWorkloadType)]
	if !ok {
		return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("Runai workload type config for %s not found", workloadType.Name), nil)
	}

	policyConfig := (*policiesWorkloadTypeConfig)[runaiWorkloadTypeConfig.WorkloadKind]

	workloadTypeConfig, apiErr := convertRunaiWorkloadTypeConfigToApi(ctx, tenantId,
		policyConfig, runaiWorkloadTypeConfig,
		workloadType, map[string]string{}, map[string]string{})
	if apiErr != nil {
		return nil, apiErr
	}

	return workloadTypeConfig, nil
}

func convertRunaiWorkloadTypeConfigToApi(ctx context.Context, tenantId uint,
	policyConfig *structs.PolicyWorkloadTypeConfig, defaultTypeConfig workloadutils.NativeWorkloadConfig,
	workloadType *models.WorkloadType, categoriesNameToId, prioritiesNameToId map[string]string,
) (*api.WorkloadTypeConfig, *api.Error) {
	categoryName := defaultTypeConfig.Category
	priorityName := defaultTypeConfig.Priority
	runaiWorkloadKind := defaultTypeConfig.WorkloadKind
	preemptibility := api.WorkloadTypePreemptibility(defaultTypeConfig.Preemptibility)

	if policyConfig != nil && policyConfig.CategoryName != nil {
		categoryName = *policyConfig.CategoryName
	}

	if policyConfig != nil && policyConfig.PriorityName != nil {
		priorityName = *policyConfig.PriorityName
	}

	if policyConfig != nil && policyConfig.Preemptibility != nil && policyConfig.Preemptibility.IsValid() {
		preemptibility = api.WorkloadTypePreemptibility(*policyConfig.Preemptibility)
	}
	if !preemptibility.IsValid() {
		preemptibility = api.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE
	}

	categoryId, priorityId, err := findCategoryAndPriorityIdByName(ctx, tenantId, categoryName, priorityName, categoriesNameToId, prioritiesNameToId)
	if err != nil {
		return nil, err
	}

	return &api.WorkloadTypeConfig{
		Id:             workloadType.ID.String(),
		Name:           runaiWorkloadKind,
		CategoryId:     categoryId,
		CategoryName:   categoryName,
		PriorityId:     priorityId,
		PriorityName:   priorityName,
		Preemptibility: preemptibility,
		CreatedBy:      RunaiSource,
		CreatedAt:      workloadType.CreatedAt,
		Source:         string(workloadType.Source),
		Group:          workloadType.Group,
		ResourceInterfaces: []api.WorkloadTypeResourceInterfacesWithVersionInner{
			{
				Version: defaultTypeConfig.Version,
			},
		},
		Kartas: []api.WorkloadTypeKartasWithVersionInner{
			{
				Version: defaultTypeConfig.Version,
			},
		},
	}, nil
}

func AppendRunaiWorkloadTypeConfigs(ctx context.Context, tenantId uint, filterBy []validators.ParameterizedFilter, existingWorkloadTypeConfigs []api.WorkloadTypeConfig) ([]api.WorkloadTypeConfig, *api.Error) {
	policiesWorkloadTypeConfig, policyErr := policy.Get().GetWorkloadConfigOfTenant(ctx)
	if policyErr != nil {
		return nil, api.NewError(int(policyErr.Code), policyErr.Message, nil)
	}

	runaiWorkloadTypes, dbErr := dao.DbListWorkloadTypes(ctx, ptr.To(models.WorkloadTypeSourceNative), filterBy)
	if dbErr != nil {
		return nil, dbErr
	}

	categoriesNameToId, prioritiesNameToId := getCategoriesAndPrioritiesMapsNameToId(existingWorkloadTypeConfigs)

	runaiWorkloadTypeConfigsToAppend := []api.WorkloadTypeConfig{}
	for _, defaultTypeConfig := range workloadutils.DefaultRunaiWorkloadTypesConfig {
		policyConfig := (*policiesWorkloadTypeConfig)[defaultTypeConfig.WorkloadKind]

		workloadType, err := findRunaiWorkloadTypeByNameFromExisting(runaiWorkloadTypes, defaultTypeConfig.WorkloadKind)
		if err != nil {
			// not found - most probably was filtered using the "filterBy" field
			continue
		}

		workloadTypeConfig, err := convertRunaiWorkloadTypeConfigToApi(ctx, tenantId,
			policyConfig, defaultTypeConfig,
			workloadType, categoriesNameToId, prioritiesNameToId)
		if err != nil {
			return nil, err
		}

		runaiWorkloadTypeConfigsToAppend = append(runaiWorkloadTypeConfigsToAppend, *workloadTypeConfig)
	}

	// sort runai workload types by name to have a deterministic order
	slices.SortFunc(runaiWorkloadTypeConfigsToAppend, func(a, b api.WorkloadTypeConfig) int {
		return strings.Compare(a.Name, b.Name)
	})

	existingWorkloadTypeConfigs = append(existingWorkloadTypeConfigs, runaiWorkloadTypeConfigsToAppend...)
	return existingWorkloadTypeConfigs, nil
}

func findCategoryAndPriorityIdByName(ctx context.Context, tenantId uint, categoryName string, priorityName string, categoriesNameToId, prioritiesNameToId map[string]string) (string, string, *api.Error) {
	categoryId, ok := categoriesNameToId[categoryName]
	if !ok {
		category, err := dao.DbGetCategoryByName(ctx, tenantId, categoryName)
		if err != nil {
			return "", "", err
		}
		categoryId = category.Id
	}

	priorityId, ok := prioritiesNameToId[priorityName]
	if !ok {
		priority, err := dao.GetPriorityByName(ctx, fmt.Sprint(tenantId), priorityName)
		if err != nil {
			return "", "", err
		}
		id := priority.ID.String()
		priorityId = id
	}

	return categoryId, priorityId, nil
}

func getCategoriesAndPrioritiesMapsNameToId(externalWorkloadTypes []api.WorkloadTypeConfig) (map[string]string, map[string]string) {
	categoriesMap := map[string]string{}
	prioritiesMap := map[string]string{}
	for _, externalWorkloadType := range externalWorkloadTypes {
		if externalWorkloadType.CategoryId != "" && externalWorkloadType.CategoryName != "" {
			categoriesMap[externalWorkloadType.CategoryName] = externalWorkloadType.CategoryId
		}
		if externalWorkloadType.PriorityId != "" && externalWorkloadType.PriorityName != "" {
			prioritiesMap[externalWorkloadType.PriorityName] = externalWorkloadType.PriorityId
		}
	}
	return categoriesMap, prioritiesMap
}

func findRunaiWorkloadTypeByNameFromExisting(allRunaiWorkloadTypes []models.WorkloadType, workloadTypeName string) (*models.WorkloadType, *api.Error) {
	for _, workloadType := range allRunaiWorkloadTypes {
		if workloadType.Name == workloadTypeName {
			return &workloadType, nil
		}
	}
	return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("Workload type %s not found", workloadTypeName), nil)
}

func parseWorkloadTypeKartasApiToDb(resourceInterfaces *[]map[string]interface{}) (*[]models.Karta, error) {
	if resourceInterfaces == nil {
		return nil, nil
	}
	if len(*resourceInterfaces) == 0 {
		return ptr.To([]models.Karta{}), nil
	}

	versionsMap := map[string]bool{}
	result := make([]models.Karta, len(*resourceInterfaces))
	for i, resourceInterface := range *resourceInterfaces {
		parsedResourceInterface, err := parseWorkloadTypeKartaApiToDb(resourceInterface)
		if err != nil {
			return nil, err
		}
		result[i] = parsedResourceInterface

		gvk := parsedResourceInterface.GroupVersionKind()
		if _, exists := versionsMap[gvk.Version]; exists {
			return nil, fmt.Errorf("kartas contain duplicate version %s", gvk.Version)
		}
		versionsMap[gvk.Version] = true
	}
	return &result, nil
}

func parseWorkloadTypeKartaApiToDb(resourceInterface map[string]interface{}) (models.Karta, error) {
	parsedResourceInterface, err := parseKartaApiToDb(resourceInterface)
	if err != nil {
		return models.Karta{}, err
	}

	return parsedResourceInterface, nil
}

func parseKartaApiToDb(resourceInterface map[string]interface{}) (models.Karta, error) {
	if resourceInterface == nil {
		return models.Karta{}, nil
	}

	bytes, err := json.Marshal(resourceInterface)
	if err != nil {
		return models.Karta{}, fmt.Errorf("failed to marshal karta: %w", err)
	}

	result := models.Karta{}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return models.Karta{}, fmt.Errorf("karta has invalid spec: %w", err)
	}

	converted := karta.Karta(result)
	validator := karta.NewKartaValidator(&converted)
	err = validator.Validate()
	if err != nil {
		return models.Karta{}, fmt.Errorf("karta validation failed: %w", err)
	}

	return result, nil
}

// parseKartasDbToApi converts DB karta records to both API response formats (dual-field for backward compatibility).
func parseKartasDbToApi(kartaRecords []models.WorkloadTypeKarta) ([]api.WorkloadTypeResourceInterfacesWithVersionInner, []api.WorkloadTypeKartasWithVersionInner, error) {
	riResult := make([]api.WorkloadTypeResourceInterfacesWithVersionInner, len(kartaRecords))
	kartaResult := make([]api.WorkloadTypeKartasWithVersionInner, len(kartaRecords))
	for i, record := range kartaRecords {
		parsed, err := parseKartaSpecDbToApi(record.Karta)
		if err != nil {
			return nil, nil, err
		}
		riResult[i] = api.WorkloadTypeResourceInterfacesWithVersionInner{
			Version:           record.Version,
			UpdatedAt:         record.UpdatedAt,
			ResourceInterface: parsed,
		}
		kartaResult[i] = api.WorkloadTypeKartasWithVersionInner{
			Version:   record.Version,
			UpdatedAt: record.UpdatedAt,
			Karta:     parsed,
		}
	}
	return riResult, kartaResult, nil
}

func parseKartaSpecDbToApi(karta models.Karta) (map[string]interface{}, error) {
	if karta.IsEmpty() {
		return nil, nil
	}

	bytes, err := json.Marshal(karta)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
