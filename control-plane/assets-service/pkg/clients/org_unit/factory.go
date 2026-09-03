/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package org_unit

import (
	"context"
	"fmt"
	"sync"

	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	orgUnitService "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
)

// ClientInterface lists the org unit services that assets service needs
type ClientInterface interface {
	GetProjectAncestors(ctx context.Context, projectId string) (*orgUnitApi.Ancestors, *orgUnitApi.Error)
	GetProjectsMetadataByDepartment(ctx context.Context, departmentId string) ([]orgUnitApi.ProjectMetadata, *orgUnitApi.Error)
	GetProjectsMetadataByCluster(ctx context.Context, clusterId string) ([]orgUnitApi.ProjectMetadata, *orgUnitApi.Error)
	GetDepartmentAncestors(ctx context.Context, departmentId string) (*orgUnitApi.Ancestors, *orgUnitApi.Error)
	GetAncestorsOfScope(ctx context.Context, scopeType orgUnitApi.ScopeType, scopeId string) (*orgUnitApi.Ancestors, *orgUnitApi.Error)
}

var once sync.Once
var deFactoClient ClientInterface

// Get the real or mock client
func Get() ClientInterface {
	once.Do(func() {
		err := cache.RegisterStore[orgUnitApi.Ancestors](options.Ttl(runai_config.Get[*config.OrgUnitCache]().Expiration()))
		if err != nil {
			logging.Warnf("Failed to register cache store for type ancestors: %v", err)
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

// cacheClient is a implementation of the org-unit ClientInterface, with caching capabilities
type cacheClient struct {
	client ClientInterface
}

func newCacheClient(client ClientInterface) ClientInterface {
	return &cacheClient{
		client: client,
	}
}

func (c *cacheClient) GetProjectAncestors(ctx context.Context, projectId string) (*orgUnitApi.Ancestors, *orgUnitApi.Error) {
	// Check the cache for ancestors
	if ancestors, found := getAncestorsFromCache(ctx, orgUnitApi.SCOPETYPE_PROJECT, projectId); found {
		return ancestors, nil
	}
	// Call the org unit client for ancestors
	ancestors, apiErr := c.client.GetProjectAncestors(ctx, projectId)
	if apiErr == nil {
		// In case there wasn't any error, store the ancestors in the cache
		key := getAncestorsKey{
			tenantId:  auth.TenantIdOf(ctx),
			scopeType: orgUnitApi.SCOPETYPE_PROJECT,
			scopeId:   projectId,
		}
		err := cache.Set[orgUnitApi.Ancestors](ctx, key.String(), *ancestors)
		if err != nil {
			logging.Warnf("Failed to set project ancestors in cache for key %s: %v", key.String(), err)
		}
	}
	// return the result
	return ancestors, apiErr
}

func (c *cacheClient) GetDepartmentAncestors(ctx context.Context, departmentId string) (*orgUnitApi.Ancestors, *orgUnitApi.Error) {
	// Check the cache for ancestors
	if ancestors, found := getAncestorsFromCache(ctx, orgUnitApi.SCOPETYPE_DEPARTMENT, departmentId); found {
		return ancestors, nil
	}
	// Call the org unit client for ancestors
	ancestors, apiErr := c.client.GetDepartmentAncestors(ctx, departmentId)
	if apiErr == nil {
		// In case there wasn't any error, store the ancestors in the cache
		key := getAncestorsKey{
			tenantId:  auth.TenantIdOf(ctx),
			scopeType: orgUnitApi.SCOPETYPE_DEPARTMENT,
			scopeId:   departmentId,
		}
		err := cache.Set[orgUnitApi.Ancestors](ctx, key.String(), *ancestors)
		if err != nil {
			logging.Warnf("Failed to set department ancestors in cache for key %s: %v", key.String(), err)
		}
	}
	// return the result
	return ancestors, apiErr
}

func (c *cacheClient) GetProjectsMetadataByDepartment(ctx context.Context, departmentId string) ([]orgUnitApi.ProjectMetadata, *orgUnitApi.Error) {
	// No caching for this function for now, tech debt for later
	return c.client.GetProjectsMetadataByDepartment(ctx, departmentId)
}

func (c *cacheClient) GetProjectsMetadataByCluster(ctx context.Context, clusterId string) ([]orgUnitApi.ProjectMetadata, *orgUnitApi.Error) {
	// No caching for this function for now, tech debt for later
	return c.client.GetProjectsMetadataByCluster(ctx, clusterId)
}

func (c *cacheClient) GetAncestorsOfScope(ctx context.Context, scopeType orgUnitApi.ScopeType, scopeId string) (*orgUnitApi.Ancestors, *orgUnitApi.Error) {
	// Check the cache for ancestors
	if ancestors, found := getAncestorsFromCache(ctx, scopeType, scopeId); found {
		return ancestors, nil
	}
	// Call the org unit client for ancestors
	ancestors, apiErr := c.client.GetAncestorsOfScope(ctx, scopeType, scopeId)
	if apiErr == nil {
		// In case there wasn't any error, store the ancestors in the cache
		key := getAncestorsKey{
			tenantId:  auth.TenantIdOf(ctx),
			scopeType: scopeType,
			scopeId:   scopeId,
		}
		err := cache.Set[orgUnitApi.Ancestors](ctx, key.String(), *ancestors)
		if err != nil {
			logging.Warnf("Failed to set ancestors in cache for key %s: %v", key.String(), err)
		}
	}
	// return the result
	return ancestors, apiErr
}

// the following section relate to caching the result from the org-unit service,
// as most of the queries are not expected to change frequently, we can cache the result for a certain period of time

type getAncestorsKey struct {
	tenantId  int32
	scopeType orgUnitApi.ScopeType
	scopeId   string
}

func (key *getAncestorsKey) String() string {
	return fmt.Sprintf("%d %s %s", key.tenantId, key.scopeType, key.scopeId)
}

func getAncestorsFromCache(ctx context.Context, scopeType orgUnitApi.ScopeType, scopeId string) (*orgUnitApi.Ancestors, bool) {
	key := getAncestorsKey{
		tenantId:  auth.TenantIdOf(ctx),
		scopeType: scopeType,
		scopeId:   scopeId,
	}
	ancestors, err := cache.Get[orgUnitApi.Ancestors](ctx, key.String())
	if err != nil {
		logging.Warnf("Failed to get ancestors from cache for key %s: %v", key.String(), err)
		return nil, false
	}
	if ancestors != nil {
		return ancestors, true
	}
	return nil, false
}
