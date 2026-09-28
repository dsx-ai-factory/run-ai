/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package org_unit

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"

	orgUnitService "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/structs"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

// ClientInterface lists the org unit services that the workloads service needs
type ClientInterface interface {
	GetProjectById(ctx context.Context, projectId string) (*orgUnitApi.Project, *orgUnitApi.Error)
	GetProjectsMetadataByDepartment(ctx context.Context, departmentId string) ([]orgUnitApi.ProjectMetadata, *orgUnitApi.Error)
	GetAncestorsOfScope(ctx context.Context, scopeType orgUnitApi.ScopeType, scopeId string) (*orgUnitApi.Ancestors, *orgUnitApi.Error)
	GetOrgTreeMap(ctx context.Context) (structs.OrgUnitMap, map[structs.OrgUnitKey]string, *orgUnitApi.Error)
	// GetDepartmentName resolves a department's name from its id, caching all
	// department names from a single org-tree fetch (see cacheClient).
	GetDepartmentName(ctx context.Context, departmentId string) (string, *orgUnitApi.Error)
}

var once sync.Once
var deFactoClient ClientInterface

// Get the real or mock client
func Get() ClientInterface {
	once.Do(func() {
		if err := cache.RegisterStore[departmentNameCacheValue](options.Ttl(runai_config.Get[*config.OrgUnitCache]().Expiration())); err != nil {
			logging.Warnf("Failed to register org unit department-name cache store: %v", err)
		}
		initRealClient()
	})
	return deFactoClient
}

// SetMock is used by unit test to apply mock client instead of the real client
func SetMock(mockClient ClientInterface) {
	deFactoClient = mockClient
}

// initialize the "real" client, invoked once if SetMock was not invoked by tests
func initRealClient() {
	if deFactoClient == nil {
		deFactoClient = newCacheClient(orgUnitService.NewOrgUnitClient(runai_config.Get[*config.ServicesConfig]().OrgUnitUrl))
	}
}

// departmentNameCacheValue is the cached value type for department-name lookups.
// It is a dedicated type so it gets its own TTL cache store.
type departmentNameCacheValue struct {
	Name string
}

// cacheClient wraps the org-unit client and adds a TTL cache for department-name
// resolution. Department names rarely change, so a single org-tree fetch on a
// cache miss populates every department's name at once - a sync batch spanning
// many departments therefore triggers at most one org-unit call per TTL window.
//
// NOTE: this whole department-name cache is a backward-compatibility fallback for
// clusters older than v3.0 that do not send departmentName in the sync payload.
// Once the minimum supported cluster version is >= 3.0 (departmentName always
// arrives from the cluster), this cacheClient and GetDepartmentName can be removed.
type cacheClient struct {
	client orgUnitService.IOrgUnitClient
}

func newCacheClient(client orgUnitService.IOrgUnitClient) ClientInterface {
	return &cacheClient{client: client}
}

func (c *cacheClient) GetProjectById(ctx context.Context, projectId string) (*orgUnitApi.Project, *orgUnitApi.Error) {
	return c.client.GetProjectById(ctx, projectId)
}

func (c *cacheClient) GetProjectsMetadataByDepartment(ctx context.Context, departmentId string) ([]orgUnitApi.ProjectMetadata, *orgUnitApi.Error) {
	return c.client.GetProjectsMetadataByDepartment(ctx, departmentId)
}

func (c *cacheClient) GetAncestorsOfScope(ctx context.Context, scopeType orgUnitApi.ScopeType, scopeId string) (*orgUnitApi.Ancestors, *orgUnitApi.Error) {
	return c.client.GetAncestorsOfScope(ctx, scopeType, scopeId)
}

func (c *cacheClient) GetOrgTreeMap(ctx context.Context) (structs.OrgUnitMap, map[structs.OrgUnitKey]string, *orgUnitApi.Error) {
	return c.client.GetOrgTreeMap(ctx)
}

func (c *cacheClient) GetDepartmentName(ctx context.Context, departmentId string) (string, *orgUnitApi.Error) {
	if departmentId == "" {
		return "", nil
	}

	tenantId := tenantIdFromContext(ctx)

	if cached, err := cache.Get[departmentNameCacheValue](ctx, departmentNameCacheKey(tenantId, departmentId)); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to read department name from cache for %s: %v", departmentId, err)
	} else if cached != nil {
		return cached.Name, nil
	}

	// Cache miss: fetch the whole org tree once and cache every department name.
	_, orgUnitNameMap, apiErr := c.client.GetOrgTreeMap(ctx)
	if apiErr != nil {
		return "", apiErr
	}

	result := ""
	for orgUnitKey, name := range orgUnitNameMap {
		if orgUnitKey.Scope != string(orgUnitApi.SCOPETYPE_DEPARTMENT) {
			continue
		}
		c.setDepartmentNameCache(ctx, tenantId, orgUnitKey.Id, name)
		if orgUnitKey.Id == departmentId {
			result = name
		}
	}

	if result == "" {
		return "", orgUnitApi.NewError(http.StatusNotFound, fmt.Sprintf("department %s not found in org tree", departmentId))
	}

	return result, nil
}

func (c *cacheClient) setDepartmentNameCache(ctx context.Context, tenantId string, departmentId string, name string) {
	if err := cache.Set[departmentNameCacheValue](ctx, departmentNameCacheKey(tenantId, departmentId), departmentNameCacheValue{Name: name}); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to cache department name for %s: %v", departmentId, err)
	}
}

func departmentNameCacheKey(tenantId string, departmentId string) string {
	return fmt.Sprintf("%s/%s", tenantId, departmentId)
}

func tenantIdFromContext(ctx context.Context) string {
	if t := runaictx.GetString(ctx, runaictx.CfTenantId); t != nil {
		return *t
	}
	return ""
}
