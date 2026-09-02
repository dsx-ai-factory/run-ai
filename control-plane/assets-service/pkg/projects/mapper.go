/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package projects

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/response_patching"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type nameToIdKey string

func newNameToIdKey(clusterId string, name string) nameToIdKey {
	return nameToIdKey(fmt.Sprintf("%s-%s", clusterId, name))
}

type BiDirectionalMapping struct {
	NameToId map[nameToIdKey]*int32
	IdToName map[int32]*string
}

func (bdm *BiDirectionalMapping) Add(projectId *int32, projectName *string, clusterId string) {
	nameKey := newNameToIdKey(clusterId, *projectName)

	bdm.NameToId[nameKey] = projectId
	bdm.IdToName[*projectId] = projectName
}

// LoadBiDirectionalMapping loads the bi-directional mapping of project names to project ids
// If
func LoadBiDirectionalMapping(ctx context.Context, clusterId string, projectId *int32, projectName *string) (*BiDirectionalMapping, *api.Error) {
	initProjectMapperCache()

	bothNameAndIdGiven := projectId != nil && projectName != nil
	nameAndIdNotGiven := projectId == nil && projectName == nil
	if bothNameAndIdGiven || nameAndIdNotGiven {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("LoadBiDirectionalMapping: either projectId or projectName should be nil")
		return nil, nil
	}

	cachedResult, err := cache.Get[BiDirectionalMapping](ctx, clusterId)
	if err != nil {
		logging.Warnf("Failed to get project bi-directional mapping from cache for cluster %s: %v", clusterId, err)
	}
	if cachedResult != nil {
		if projectId != nil {
			if cachedResult.IdToName[*projectId] != nil {
				return cachedResult, nil
			}
		}
		if projectName != nil {
			nameKey := newNameToIdKey(clusterId, *projectName)
			if cachedResult.NameToId[nameKey] != nil {
				return cachedResult, nil
			}
		}
	}

	projectsOfCluster, apiErr := orgunit.Get().Client().ListProjects(ctx, clusterId)
	if apiErr != nil {
		return nil, api.NewError(int(apiErr.Code), apiErr.Message, nil)
	}

	result := BiDirectionalMapping{
		NameToId: make(map[nameToIdKey]*int32),
		IdToName: make(map[int32]*string),
	}

	for _, project := range projectsOfCluster {
		if project.Name == "" || project.Id == "" {
			continue
		}

		projectId, err := strconv.Atoi(project.Id)
		if err != nil {
			logging.Errorf("Failed to convert project id %s to int: %v", project.Id, err)
			continue
		}

		nameKey := newNameToIdKey(clusterId, project.Name)

		result.NameToId[nameKey] = convutils.Ref(int32(projectId))
		result.IdToName[int32(projectId)] = convutils.Ref(project.Name)
	}

	err = cache.Set[BiDirectionalMapping](ctx, clusterId, result)
	if err != nil {
		logging.Warnf("Failed to set project bi-directional mapping in cache for cluster %s: %v", clusterId, err)
	}

	return &result, nil
}

func ProjectNameToId(ctx context.Context, clusterId string, projectName string) (int32, *api.Error) {
	mapping, err := LoadBiDirectionalMapping(ctx, clusterId, nil, convutils.Ref(projectName))
	if err != nil {
		return 0, nil
	}

	nameKey := newNameToIdKey(clusterId, projectName)

	result := mapping.NameToId[nameKey]
	if result == nil {
		return 0, api.NewError(http.StatusNotFound, "project not found", nil)
	}

	return *result, nil
}

func PatchProjectNames(ctx context.Context, patchingInfo map[string]response_patching.ResponsePatchingPtrs) *api.Error {
	//
	//   some sanity validations
	//
	if patchingInfo == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("patchingInfo nil, cannot patch usage times")
		return nil
	}
	if len(patchingInfo) == 0 {
		return nil
	}

	var err *api.Error
	var clusterId string
	var clusterMapping *BiDirectionalMapping

	for _, assetInfo := range patchingInfo {
		if assetInfo.Meta.Scope != api.SCOPE_PROJECT || assetInfo.Meta.ClusterId == nil || assetInfo.Meta.ProjectId == nil {
			continue
		}

		if *assetInfo.Meta.ClusterId != clusterId {
			clusterId = *assetInfo.Meta.ClusterId

			clusterMapping, err = LoadBiDirectionalMapping(ctx, clusterId, assetInfo.Meta.ProjectId, nil)
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to load projects of cluster %s: %v", clusterId, err.Message)
				return err
			}
		}

		if clusterMapping != nil {
			assetInfo.Meta.ProjectName = clusterMapping.IdToName[*assetInfo.Meta.ProjectId]
		}
	}
	return nil
}

type NamespaceProjectBiDirectionalMapping struct {
	NamespaceToProjectId map[string]*string
	ProjectIdToNamespace map[string]*string
}

func (n *NamespaceProjectBiDirectionalMapping) GetProjectIdOfNamespace(namespace string) (string, bool) {
	projectId, ok := n.NamespaceToProjectId[namespace]
	if !ok {
		return "", false
	}
	return *projectId, true
}

func (n *NamespaceProjectBiDirectionalMapping) GetNamespaceOfProjectId(projectId string) (string, bool) {
	namespace, ok := n.ProjectIdToNamespace[projectId]
	if !ok {
		return "", false
	}
	return *namespace, true
}
