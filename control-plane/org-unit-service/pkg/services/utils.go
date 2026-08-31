/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	"k8s.io/utils/ptr"
)

func roundNumber(num *float64) float64 {
	if num == nil || *num == 0 {
		return 0
	}

	rounded := int64(*num * 1000)
	return float64(rounded) / 1000
}

func getQuotaStatusesMap(projectResources []models.Resources) map[string]models.Resources {
	result := make(map[string]models.Resources)
	for _, quotaStatus := range projectResources {
		result[quotaStatus.NodePoolName] = quotaStatus
	}
	return result
}

func getNodePoolResourcesMap(nodePoolsResources []models.Resources) map[string]models.Resources {
	result := make(map[string]models.Resources)
	for _, assignedResources := range nodePoolsResources {
		result[assignedResources.NodePoolName] = assignedResources
	}
	return result
}

func ModelUnitsToApi(units models.Units) api.Units {
	switch units {
	case models.MB:
		return api.UNITS_MB
	case models.GB:
		return api.UNITS_GB
	case models.Mib:
		return api.UNITS_MIB
	}
	return api.UNITS_MIB
}

// sumDeservedResources sums the deserved resources of two resources;
// assumes their node pool is the same
func sumDeservedResources(left, right models.Resources) models.Resources {
	result := models.Resources{
		NodePoolId:   left.NodePoolId,
		NodePoolName: left.NodePoolName,
		GpuDeserved:  sumDeserved(left.GpuDeserved, right.GpuDeserved),
	}

	if left.CpuDeserved != nil && right.CpuDeserved != nil {
		result.CpuDeserved = sumDeserved(left.CpuDeserved, right.CpuDeserved)
	}
	if left.MemoryDeserved != nil && right.MemoryDeserved != nil {
		result.MemoryDeserved = sumMemoryDeserved(left.MemoryDeserved, right.MemoryDeserved, left.MemoryUnits, right.MemoryUnits)
		result.MemoryUnits = left.MemoryUnits
	}
	return result
}

func getProjectAggregatedDeservedSum(projects []models.Project) []models.Resources {
	resultMap := map[string]models.Resources{}
	for _, project := range projects {
		for _, nodePoolResources := range project.OrgUnitData.Resources {
			totalNodePoolResources, ok := resultMap[nodePoolResources.NodePoolName]
			if !ok {
				resultMap[nodePoolResources.NodePoolName] = nodePoolResources
				continue
			}
			resultMap[nodePoolResources.NodePoolName] = sumDeservedResources(totalNodePoolResources, nodePoolResources)
		}
	}

	result := []models.Resources{}
	for _, resources := range resultMap {
		result = append(result, resources)
	}
	return result
}

func sumDeserved(left, right *float64) *float64 {
	if left == nil {
		return right
	}
	if right == nil {
		return left
	}
	if isUnlimited(left) {
		return left
	}
	if isUnlimited(right) {
		return right
	}

	return ptr.To(*left + *right)
}

func sumMemoryDeserved(left, right *float64, leftUnits, rightUnits models.Units) *float64 {
	if leftUnits == rightUnits {
		return sumDeserved(left, right)
	}

	conversionFn := GetMemoryConversionFn(ModelUnitsToApi(leftUnits), ModelUnitsToApi(rightUnits))
	rightConverted := ApplyMemoryConversion(right, conversionFn)
	return sumDeserved(left, rightConverted)
}

func DereferenceFloat64(f *float64) float64 {
	if f != nil {
		return *f
	}
	return 0
}

func CopyPtr[T any](source *T) *T {
	if source == nil {
		return nil
	}

	val := *source
	return ptr.To(val)
}

func isUnlimited(deserved *float64) bool {
	return deserved == nil || *deserved == unlimited
}

func ArrayToString[T any](arr []T) []string {
	if len(arr) == 0 {
		return []string{}
	}
	result := make([]string, len(arr))
	for i, v := range arr {
		result[i] = fmt.Sprintf("%v", v)
	}
	return result
}

func ParseOrgUnitIds(ids string) ([]int64, error) {
	splitIds := strings.Split(ids, ",")

	parsedIds := []int64{}
	for _, id := range splitIds {
		parsedId, err := strconv.Atoi(id)
		if err != nil {
			return nil, err
		}

		parsedIds = append(parsedIds, int64(parsedId))
	}

	return parsedIds, nil
}

type SyncSlice[T any] struct {
	mu    sync.RWMutex
	items []T
}

func (s *SyncSlice[T]) Append(item T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
}

func (s *SyncSlice[T]) GetAll() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]T(nil), s.items...)
}
