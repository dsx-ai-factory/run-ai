/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload_type_config

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/run-ai/runai/runai-application-packages/datastructs"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WorkloadTypeConfig struct {
	TypeName string
	Group    string
	// The name of the workload category.
	CategoryName string
	// The name of the workload priority.
	PriorityName string
	// The preemptibility level of the workload.
	Preemptibility string
}

// WorkloadConfigNotFoundError represents an error when a workload configuration is not found.
// This is a 404-type error indicating the configuration doesn't exist.
type WorkloadConfigNotFoundError struct {
	GroupKind schema.GroupKind
}

func (e *WorkloadConfigNotFoundError) Error() string {
	return fmt.Sprintf("workload configuration not found for groupKind <%s> or kind <%s>", e.GroupKind.String(), e.GroupKind.Kind)
}

// IsWorkloadConfigNotFound checks if an error is a WorkloadConfigNotFoundError
func IsWorkloadConfigNotFound(err error) bool {
	var workloadConfigNotFoundError *WorkloadConfigNotFoundError
	ok := errors.As(err, &workloadConfigNotFoundError)
	return ok
}

// GetWorkloadConfigByGroupKind - retrieves the workload configuration by its GroupKind or Kind from the provided config map.
func GetWorkloadConfigByGroupKind(ctx context.Context, readerClient client.Reader,
	workloadGroup, workloadKind string) (*WorkloadTypeConfig, error) {
	workloadConfigs, err := getWorkloadConfigsPerGroupKind(ctx, readerClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get workload configs from configmap, error: %s", err.Error())
	}

	return getWorkloadConfigByGroupKindFromConfigMap(workloadConfigs, workloadGroup, workloadKind)
}

func GetWorkloadConfigsConfigMap(ctx context.Context, readerClient client.Reader) (*corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{}
	err := readerClient.Get(ctx, types.NamespacedName{
		Name:      constants.WorkloadsConfigMapName,
		Namespace: constants.WorkloadsConfigMapNamespace}, cm)
	if err != nil {
		return nil, fmt.Errorf("failed to get default priorities configmap <%s/%s>, error: %s",
			constants.WorkloadsConfigMapNamespace, constants.WorkloadsConfigMapName, err.Error())
	}
	return cm, nil
}

func GetWorkloadCategoriesFromConfigMap(ctx context.Context, readerClient client.Reader) (datastructs.Set[string], error) {
	cm, err := GetWorkloadConfigsConfigMap(ctx, readerClient)
	if err != nil {
		return nil, err
	}

	categories, err := parseWorkloadCategories(cm)
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func ValidateWorkloadCategory(ctx context.Context, readerClient client.Reader, workloadCategory string) error {
	if workloadCategory == "" {
		return fmt.Errorf("category name is not valid - empty string")
	}

	categories, err := GetWorkloadCategoriesFromConfigMap(ctx, readerClient)
	if err != nil {
		return err
	}
	if !categories.Has(workloadCategory) {
		return fmt.Errorf("category name <%s> is not valid", workloadCategory)
	}
	return nil
}

func toMapByGroupKind(configs *[]WorkloadTypeConfig) map[string]WorkloadTypeConfig {
	res := map[string]WorkloadTypeConfig{}
	for _, config := range *configs {
		groupKind := groupKindMapKey(config.Group, config.TypeName)
		res[groupKind] = config
	}
	return res
}

func parseWorkloadConfigs(cm *corev1.ConfigMap) (map[string]WorkloadTypeConfig, error) {
	if cm == nil || cm.Data == nil {
		return nil, fmt.Errorf("configMap is empty, cannot parse Workload Configs")
	}
	data, ok := cm.Data[constants.WorkloadsConfigMapTypesKey]
	if !ok {
		return nil, fmt.Errorf("configMap Data does not contain <%s> key", constants.WorkloadsConfigMapTypesKey)
	}

	var configs []WorkloadTypeConfig
	err := json.Unmarshal([]byte(data), &configs)
	if err != nil {
		return nil, err
	}
	return toMapByGroupKind(&configs), nil
}

// getWorkloadConfigByGroupKindFromConfigMap -
// retrieves the workload configuration by its GroupKind or Kind from the provided config map.
func getWorkloadConfigByGroupKindFromConfigMap(workloadConfigs map[string]WorkloadTypeConfig,
	workloadGroup, workloadKind string) (*WorkloadTypeConfig, error) {
	gk := schema.GroupKind{Group: workloadGroup, Kind: workloadKind}
	groupKind := gk.String()

	if typeConfig, ok := workloadConfigs[groupKind]; ok {
		return &typeConfig, nil
	}
	if typeConfig, ok := workloadConfigs[workloadKind]; ok {
		return &typeConfig, nil
	}

	return nil, &WorkloadConfigNotFoundError{
		GroupKind: gk,
	}
}

func getWorkloadConfigsPerGroupKind(ctx context.Context, readerClient client.Reader) (map[string]WorkloadTypeConfig, error) {
	cm, err := GetWorkloadConfigsConfigMap(ctx, readerClient)
	if err != nil {
		return nil, err
	}
	return parseWorkloadConfigs(cm)
}

func toSet[T comparable](array []T) datastructs.Set[T] {
	set := datastructs.Set[T]{}
	for _, val := range array {
		set.Add(val)
	}
	return set
}

func parseWorkloadCategories(cm *corev1.ConfigMap) (datastructs.Set[string], error) {
	if cm == nil || cm.Data == nil {
		return nil, fmt.Errorf("configMap is empty, cannot parse Workload Categories")
	}
	data, ok := cm.Data[constants.WorkloadsConfigMapCategoriesKey]
	if !ok {
		return nil, fmt.Errorf("configMap Data does not contain <%s> key", constants.WorkloadsConfigMapCategoriesKey)
	}
	var categoriesArray []string
	err := json.Unmarshal([]byte(data), &categoriesArray)
	if err != nil {
		return nil, err
	}
	categories := toSet(categoriesArray)
	return categories, nil
}

func groupKindMapKey(group, kind string) string {
	return schema.GroupKind{Group: group, Kind: kind}.String()
}
