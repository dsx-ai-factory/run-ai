/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	authorization_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/internal/generated/pkg/api"
	cacheutils "github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/pkg/cache"
	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	cacheTesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/testing"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/integdb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)

	conf, cleanup, setupErr := integdb.SetupPostgresDB(&integdb.IntegdbConfig{Schema: "datavolumes", MigrationDirPath: "file://../database/migrations"})
	if setupErr != nil {
		t.Fatalf("failed to set up test postgres: %v", setupErr)
	}
	t.Cleanup(cleanup)
	if initErr := runaidb.Init(conf); initErr != nil {
		t.Fatalf("failed to init database: %v", initErr)
	}
	t.Cleanup(func() { _ = runaidb.Close() })

	SetMockBackendServer()
	cacheTesting.EnableCache()

	authorization_client.SetMockAuthorizationClient(&SimpleMockAuthService{})

	RunSpecs(t, "Api tests")
	t.Cleanup(func() {
		cache.ClearCache()
		CleanTables()
	})
}

func SetContext(tenantId string, subjectId string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantId)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, subjectId)
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Tenants: []string{tenantId}})

	return ctx
}

func CleanTables() {
	db := runaidb.DB().Actual()
	db.Where("1 = 1").Delete(&models.Datavolume{})
	db.Where("1 = 1").Delete(&models.SharedScope{})
}

func SetProjectAncestorsCache(ctx context.Context, projectId string, ancestors *authorization_utils.Ancestors) {
	SetAncestorsCache(ctx, projectId, api.SCOPETYPE_PROJECT, ancestors)
}

func SetAncestorsCache(ctx context.Context, resourceId string, resourceType api.ScopeType, ancestors *authorization_utils.Ancestors) {
	err := cache.Set[authorization_utils.Ancestors](ctx, authorization_utils.GetKey(resourceId, string(resourceType)), *ancestors)
	Expect(err).To(BeNil())
}

func InvalidateProjectAncestorsCache(ctx context.Context, projectId string) {
	InvalidateAncestorsCache(ctx, projectId, api.SCOPETYPE_PROJECT)
}

func InvalidateAncestorsCache(ctx context.Context, resourceId string, resourceType api.ScopeType) {
	err := cache.Del[authorization_utils.Ancestors](ctx, authorization_utils.GetKey(resourceId, string(resourceType)))
	Expect(err).To(BeNil())
}

func SetScopeNameCache(ctx context.Context, scopeId string, scopeType api.ScopeType, clusterId, scopeName string) {
	err := cache.Set[string](ctx, cacheutils.GetScopeNameKey(scopeId, scopeType, clusterId), scopeName)
	Expect(err).To(BeNil())
}

func InvalidateScopeNameCache(ctx context.Context, scopeId string, scopeType api.ScopeType, clusterId string) {
	err := cache.Del[string](ctx, cacheutils.GetScopeNameKey(scopeId, scopeType, clusterId))
	Expect(err).To(BeNil())
}

func getTestScopeName(scopeId string, scopeType api.ScopeType) string {
	return fmt.Sprintf("%s//%s", scopeType, scopeId)
}

func sharedScopesToMap(sharedScopes []api.SharedScope) map[api.SharedScopeBase]api.SharedScope {
	result := make(map[api.SharedScopeBase]api.SharedScope)
	for _, sharedScope := range sharedScopes {
		result[api.SharedScopeBase{
			ScopeId:   sharedScope.ScopeId,
			ScopeType: sharedScope.ScopeType,
		}] = sharedScope
	}
	return result
}

func datavolumeStatusToInternal(status *api.DatavolumeStatusForSync) *api.DatavolumeStatusInternal {
	return &api.DatavolumeStatusInternal{
		Phase:             status.Phase,
		PhaseMessage:      status.PhaseMessage,
		Conditions:        status.Conditions,
		DatavolumePvcName: status.DatavolumePvcName,
		DatavolumePvName:  status.DatavolumePvName,
	}
}

func datavolumeToRequest(datavolume api.Datavolume) api.DatavolumeCreationFields {
	return api.DatavolumeCreationFields{
		Name:                       datavolume.Name,
		Description:                datavolume.Description,
		ProjectId:                  datavolume.ProjectId,
		OriginPvcName:              datavolume.OriginPvcName,
		ShouldDeleteOriginalVolume: datavolume.ShouldDeleteOriginalVolume,
	}
}

func datavolumeNoSharedScopesToDatavolume(datavolume api.DatavolumeNoSharedScopes) api.Datavolume {
	return api.Datavolume{
		Id:                         datavolume.Id,
		CreatedAt:                  datavolume.CreatedAt,
		UpdatedAt:                  datavolume.UpdatedAt,
		CreatedBy:                  datavolume.CreatedBy,
		UpdatedBy:                  datavolume.UpdatedBy,
		Name:                       datavolume.Name,
		Description:                datavolume.Description,
		ProjectId:                  datavolume.ProjectId,
		OriginPvcName:              datavolume.OriginPvcName,
		ShouldDeleteOriginalVolume: datavolume.ShouldDeleteOriginalVolume,
		ProjectName:                datavolume.ProjectName,
		DepartmentId:               datavolume.DepartmentId,
		ClusterId:                  datavolume.ClusterId,
		Status:                     datavolume.Status,
	}
}

func datavolumesNoSharedScopesToMap(datavolumes []api.DatavolumeNoSharedScopes) map[string]api.Datavolume {
	result := make(map[string]api.Datavolume)
	for _, datavolume := range datavolumes {
		result[datavolume.Id] = datavolumeNoSharedScopesToDatavolume(datavolume)
	}

	return result
}

func datavolumeInternalToDatavolume(datavolume api.DatavolumeInternal) api.Datavolume {
	return api.Datavolume{
		Id:                         datavolume.Id,
		CreatedAt:                  datavolume.CreatedAt,
		UpdatedAt:                  datavolume.UpdatedAt,
		CreatedBy:                  datavolume.CreatedBy,
		UpdatedBy:                  datavolume.UpdatedBy,
		Name:                       datavolume.Name,
		Description:                datavolume.Description,
		ProjectId:                  datavolume.ProjectId,
		OriginPvcName:              datavolume.OriginPvcName,
		ShouldDeleteOriginalVolume: datavolume.ShouldDeleteOriginalVolume,
		ProjectName:                datavolume.ProjectName,
		DepartmentId:               datavolume.DepartmentId,
		ClusterId:                  datavolume.ClusterId,
	}
}

func sortDatavolumesForTest(datavolumes []api.Datavolume) {
	sort.Slice(datavolumes, func(i, j int) bool {
		return datavolumes[i].Id < datavolumes[j].Id
	})
}

func getSortFilterByValue(dv api.DatavolumeNoSharedScopes, byKey string) string {
	switch byKey {
	case "name":
		return dv.Name
	case "projectId":
		return dv.ProjectId
	case "clusterId":
		return dv.ClusterId
	case "departmentId":
		return dv.DepartmentId
	case "projectName":
		return dv.ProjectName
	case "createdAt":
		return dv.CreatedAt.Format(time.RFC3339)
	case "createdBy":
		return dv.CreatedBy
	case "phase":
		return string(dv.Status.Phase)
	}

	return ""
}

func getDefaultClustersCacheConfig() *checker.ClustersCacheConfig {
	return &checker.ClustersCacheConfig{ExpirationSec: 600, CleanupSec: 10}
}

func buildDisplayedCluster(clusterId string, version string) openapi.DisplayedCluster {
	result := openapi.DisplayedCluster{
		Uuid: clusterId,
	}
	result.Version.Set(&version)
	return result
}
