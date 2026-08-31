/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package scopes

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	openapi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/config"
)

func TestScopes(t *testing.T) {
	RegisterFailHandler(Fail)

	config.MockConfig()

	RunSpecs(t, "Database Suite")
}

// OrgUnitBasedBackendMockClient - for checking scopes derivation we'll need a backend mock for the ancestors
// api that will be able to support real hierarchy of project -> department -> tenant
type OrgUnitBasedBackendMockClient struct {
	TranslationTable map[string]OrgUnit
}

func (mock *OrgUnitBasedBackendMockClient) Client() orgunit.ClientInterface {
	return mock
}

func NewOrgUnitBasedBackendMockClient() *OrgUnitBasedBackendMockClient {
	return &OrgUnitBasedBackendMockClient{
		TranslationTable: make(map[string]OrgUnit),
	}
}

// Feed the mock with child -> parent hierarchy.
func (mock *OrgUnitBasedBackendMockClient) Add(orgUnit OrgUnit, parent OrgUnit) {
	mock.TranslationTable[orgUnit.String()] = parent
}

// Return ancestors based on the OrgUnits we added to this mock
func (mock *OrgUnitBasedBackendMockClient) GetAncestors(_ context.Context, scopeType string, resourceId string) (*openapi.Ancestors, *openapi.Error) {
	orgUnit := OrgUnit{
		Scope: api.ScopeType(scopeType),
		Id:    resourceId,
	}
	parent, ok := mock.TranslationTable[orgUnit.String()]
	if !ok {
		return nil, openapi.NewError(http.StatusNotFound, fmt.Sprintf("%s %s does not exist", scopeType, resourceId))
	}

	response := openapi.Ancestors{}
	if parent.Scope == api.SCOPETYPE_CLUSTER {
		response.SetClusterUuid(parent.Id)
	}
	if parent.Scope == api.SCOPETYPE_DEPARTMENT {
		response.DepartmentId.Set(convutils.Ref(parent.Id))
		clusterOrgUnit, knownCluster := mock.TranslationTable[parent.String()]
		if knownCluster {
			response.SetClusterUuid(clusterOrgUnit.Id)
		}
	}

	return &response, nil
}

func (mock *OrgUnitBasedBackendMockClient) ListProjects(_ context.Context, _ string) ([]openapi.ProjectMetadata, *openapi.Error) {
	return nil, openapi.NewError(http.StatusNotImplemented, "OrgUnitBasedBackendMockClient does not implement list projects")
}

func (mock *OrgUnitBasedBackendMockClient) GetOrgTree(ctx context.Context) ([]openapi.OrgUnit, *openapi.Error) {
	return nil, openapi.NewError(http.StatusNotImplemented, "OrgUnitBasedBackendMockClient does not implement get tree")
}
