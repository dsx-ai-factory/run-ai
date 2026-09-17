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
	"net/http/httptest"
	"strings"
	"time"

	audit "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client"

	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	authorization_config "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/test"
	orgunit_openapi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	. "github.com/onsi/gomega"
)

const (
	ProjectName    = "Project-1"
	DepartmentName = "Department-1"
	ClusterName    = "Cluster-1"
	TenantName     = "Tenant-1"
	TimeLayout     = "2006-01-02T15:04:05.999Z"
)

func addAuditToCtx(ctx context.Context) context.Context {
	req, _ := http.NewRequest(http.MethodPost, "http://whatever.com", nil)
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "1.1.1.1"

	al := audit.NewAuditLog().WithRequest(req).WithCtx(ctx)
	return audit.CtxWithAudit(ctx, al)
}

func setContext(tenantId string, subjectId string, groups []string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantId)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, subjectId)
	ctx = context.WithValue(ctx, runaictx.CfGroups, groups)
	ctx = addAuditToCtx(ctx)
	return ctx
}

func SetAncestorsCache(ctx context.Context, resourceId string, resourceType api.ResourceType, ancestors *authorization_utils.Ancestors) {
	err := cache.Set[authorization_utils.Ancestors](ctx, authorization_utils.GetKey(resourceId, string(resourceType)), *ancestors)
	Expect(err).To(BeNil())
}

func SetDescendantsCache(ctx context.Context, resourceId string, resourceType api.ResourceType, descendants *authorization_utils.Descendants) {
	err := cache.Set[authorization_utils.Descendants](ctx, authorization_utils.GetDescendantsKey(resourceId, string(resourceType)), *descendants)
	Expect(err).To(BeNil())
}

func Invalidate(ctx context.Context, resourceId string, resourceType api.ResourceType) {
	err := cache.Del[authorization_utils.Ancestors](ctx, authorization_utils.GetKey(resourceId, string(resourceType)))
	Expect(err).To(BeNil())
}

func SetMockBackendServer() {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "setting") {
			jsonResponse, err := json.Marshal(false)
			Expect(err).To(BeNil())

			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonResponse)
			return
		}

		// New internal org tree route used by OrgUnitServiceClient
		if r.URL.Path == "/internal/org-unit/org-tree" {
			jsonResponse, err := json.Marshal(orgunit_openapi.GetInternalOrgTree200Response{
				OrgUnits: []orgunit_openapi.OrgUnit{
					{
						Id:         TestProjectId1,
						Type:       orgunit_openapi.SCOPETYPE_PROJECT,
						Name:       ProjectName + TestProjectId1,
						ParentId:   &TestDepartmentId1,
						ParentType: ptr.To(orgunit_openapi.SCOPETYPE_DEPARTMENT),
					},
					{
						Id:         TestProjectId2,
						Type:       orgunit_openapi.SCOPETYPE_PROJECT,
						Name:       ProjectName + TestProjectId2,
						ParentId:   &TestDepartmentId1,
						ParentType: ptr.To(orgunit_openapi.SCOPETYPE_DEPARTMENT),
					},
					{
						Id:         TestProjectId3,
						Type:       orgunit_openapi.SCOPETYPE_PROJECT,
						Name:       ProjectName + TestProjectId3,
						ParentId:   &TestDepartmentId2,
						ParentType: ptr.To(orgunit_openapi.SCOPETYPE_DEPARTMENT),
					},
					{
						Id:         TestDepartmentId1,
						Type:       orgunit_openapi.SCOPETYPE_DEPARTMENT,
						Name:       DepartmentName + TestDepartmentId1,
						ParentId:   &TestCluster1Id,
						ParentType: ptr.To(orgunit_openapi.SCOPETYPE_CLUSTER),
					},
					{
						Id:         TestDepartmentId2,
						Type:       orgunit_openapi.SCOPETYPE_DEPARTMENT,
						Name:       DepartmentName + TestDepartmentId2,
						ParentId:   &TestCluster2Id,
						ParentType: ptr.To(orgunit_openapi.SCOPETYPE_CLUSTER),
					},

					{
						Id:         ProjectId,
						Type:       orgunit_openapi.ScopeType("project"),
						Name:       ProjectName,
						ParentId:   &DepartmentId,
						ParentType: ptr.To(orgunit_openapi.ScopeType("department")),
					},
					{
						Id:         TestScopeId,
						Type:       orgunit_openapi.ScopeType("project"),
						Name:       ProjectName,
						ParentId:   &DepartmentId,
						ParentType: ptr.To(orgunit_openapi.ScopeType("department")),
					},
					{
						Id:         TestProject2Id,
						Type:       orgunit_openapi.ScopeType("project"),
						Name:       ProjectName,
						ParentId:   &DepartmentId,
						ParentType: ptr.To(orgunit_openapi.ScopeType("department")),
					},
					{
						Id:         TestProject3Id,
						Type:       orgunit_openapi.ScopeType("project"),
						Name:       ProjectName,
						ParentId:   ptr.To(TestDepartment2Id),
						ParentType: ptr.To(orgunit_openapi.ScopeType("department")),
					},
					{
						Id:         TestProject4Id,
						Type:       orgunit_openapi.ScopeType("project"),
						Name:       ProjectName,
						ParentId:   ptr.To(TestDepartment2Id),
						ParentType: ptr.To(orgunit_openapi.ScopeType("department")),
					},
					{
						Id:         DepartmentId,
						Type:       orgunit_openapi.ScopeType("department"),
						Name:       DepartmentName,
						ParentId:   &ClusterId,
						ParentType: ptr.To(orgunit_openapi.ScopeType("cluster")),
					},
					{
						Id:         TestDepartment2Id,
						Type:       orgunit_openapi.ScopeType("department"),
						Name:       DepartmentName,
						ParentId:   &TestCluster1Id,
						ParentType: ptr.To(orgunit_openapi.ScopeType("cluster")),
					},
					{
						Id:         TestDepartment3Id,
						Type:       orgunit_openapi.ScopeType("department"),
						Name:       DepartmentName,
						ParentId:   &TestCluster2Id,
						ParentType: ptr.To(orgunit_openapi.ScopeType("cluster")),
					},
					{
						Id:         TestDepartment4Id,
						Type:       orgunit_openapi.ScopeType("department"),
						Name:       DepartmentName,
						ParentId:   &TestCluster2Id,
						ParentType: ptr.To(orgunit_openapi.ScopeType("cluster")),
					},
					{
						Id:         ClusterId,
						Type:       orgunit_openapi.ScopeType("cluster"),
						Name:       ClusterName,
						ParentId:   ptr.To(fmt.Sprintf("%d", testing.TestTenant)),
						ParentType: ptr.To(orgunit_openapi.ScopeType("tenant")),
					},
					{
						Id:         TestCluster1Id,
						Type:       orgunit_openapi.ScopeType("cluster"),
						Name:       ClusterName,
						ParentId:   ptr.To(fmt.Sprintf("%d", testing.TestTenant)),
						ParentType: ptr.To(orgunit_openapi.ScopeType("tenant")),
					},
					{
						Id:         TestCluster2Id,
						Type:       orgunit_openapi.ScopeType("cluster"),
						Name:       ClusterName,
						ParentId:   ptr.To(fmt.Sprintf("%d", testing.TestTenant)),
						ParentType: ptr.To(orgunit_openapi.ScopeType("tenant")),
					},
					{
						Id:   fmt.Sprintf("%d", testing.TestTenant),
						Type: orgunit_openapi.ScopeType("tenant"),
						Name: TenantName,
					},
				},
			})
			Expect(err).To(BeNil())

			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonResponse)
			return
		}

		if r.URL.Path == "/api/v1/org-unit" {
			type mockOrgUnit struct {
				Id         string  `json:"id"`
				Type       string  `json:"type"`
				Name       string  `json:"name"`
				ParentId   *string `json:"parentId,omitempty"`
				ParentType *string `json:"parentType,omitempty"`
			}
			jsonResponse, err := json.Marshal(struct {
				OrgUnits []mockOrgUnit `json:"orgUnits"`
			}{
				OrgUnits: []mockOrgUnit{
					{Id: ProjectId, Type: string(api.SCOPETYPE_PROJECT), Name: ProjectName, ParentId: &DepartmentId, ParentType: ptr.To(string(api.SCOPETYPE_DEPARTMENT))},
					{Id: TestScopeId, Type: string(api.SCOPETYPE_PROJECT), Name: ProjectName, ParentId: &DepartmentId, ParentType: ptr.To(string(api.SCOPETYPE_DEPARTMENT))},
					{Id: TestProject2Id, Type: string(api.SCOPETYPE_PROJECT), Name: ProjectName, ParentId: &DepartmentId, ParentType: ptr.To(string(api.SCOPETYPE_DEPARTMENT))},
					{Id: TestProject3Id, Type: string(api.SCOPETYPE_PROJECT), Name: ProjectName, ParentId: ptr.To(TestDepartment2Id), ParentType: ptr.To(string(api.SCOPETYPE_DEPARTMENT))},
					{Id: TestProject4Id, Type: string(api.SCOPETYPE_PROJECT), Name: ProjectName, ParentId: ptr.To(TestDepartment2Id), ParentType: ptr.To(string(api.SCOPETYPE_DEPARTMENT))},
					{Id: DepartmentId, Type: string(api.SCOPETYPE_DEPARTMENT), Name: DepartmentName, ParentId: &ClusterId, ParentType: ptr.To(string(api.SCOPETYPE_CLUSTER))},
					{Id: TestDepartment2Id, Type: string(api.SCOPETYPE_DEPARTMENT), Name: DepartmentName, ParentId: &TestCluster1Id, ParentType: ptr.To(string(api.SCOPETYPE_CLUSTER))},
					{Id: TestDepartment3Id, Type: string(api.SCOPETYPE_DEPARTMENT), Name: DepartmentName, ParentId: &TestCluster2Id, ParentType: ptr.To(string(api.SCOPETYPE_CLUSTER))},
					{Id: TestDepartment4Id, Type: string(api.SCOPETYPE_DEPARTMENT), Name: DepartmentName, ParentId: &TestCluster2Id, ParentType: ptr.To(string(api.SCOPETYPE_CLUSTER))},
					{Id: ClusterId, Type: string(api.SCOPETYPE_CLUSTER), Name: ClusterName, ParentId: ptr.To(fmt.Sprintf("%d", testing.TestTenant)), ParentType: ptr.To(string(api.SCOPETYPE_TENANT))},
					{Id: TestCluster1Id, Type: string(api.SCOPETYPE_CLUSTER), Name: ClusterName, ParentId: ptr.To(fmt.Sprintf("%d", testing.TestTenant)), ParentType: ptr.To(string(api.SCOPETYPE_TENANT))},
					{Id: TestCluster2Id, Type: string(api.SCOPETYPE_CLUSTER), Name: ClusterName, ParentId: ptr.To(fmt.Sprintf("%d", testing.TestTenant)), ParentType: ptr.To(string(api.SCOPETYPE_TENANT))},
					{Id: fmt.Sprintf("%d", testing.TestTenant), Type: string(api.SCOPETYPE_TENANT), Name: TenantName},
				},
			})
			Expect(err).To(BeNil())

			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonResponse)
			return
		}

		if r.URL.Path == "/v1/k8s/clusters" {
			jsonResponse, err := json.Marshal([]map[string]any{
				{
					"version": testing.ClusterVersion,
				},
			})
			Expect(err).To(BeNil())

			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonResponse)
			return
		}

		if r.URL.Path == "/api/v1/clusters" {
			jsonResponse, err := json.Marshal([]map[string]any{
				{
					"uuid":      ClusterId,
					"tenantId":  testing.TestTenant,
					"name":      ClusterName,
					"createdAt": time.Now().Format(time.RFC3339),
					"version":   testing.ClusterVersion,
				},
			})
			Expect(err).To(BeNil())

			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonResponse)
			return
		}

		allowedProjectIds := []string{ProjectId, TestScopeId, TestScope2Id, TestProject2Id, TestProject3Id, TestProject4Id}
		allowedDepartmentIds := []string{DepartmentId, TestDepartment2Id, TestDepartment3Id, TestDepartment4Id}
		allowedClusterIds := []string{ClusterId, TestCluster1Id, TestCluster2Id}

		var scopeName string

		if strings.Contains(r.URL.Path, fmt.Sprintf("tenants/%d", testing.TestTenant)) {
			jsonResponse, err := json.Marshal(map[string]interface{}{"id": testing.TestTenant, "name": TenantName, "uuid": "test-uuid"})
			Expect(err).To(BeNil())
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonResponse)
			return
		} else if isAllowedScope(r.URL.Path, "clusters/%s/", allowedClusterIds) {
			scopeName = ClusterName
		} else if strings.Contains(r.URL.Path, "org-unit/names") {
			if strings.Contains(r.URL.RawQuery, "department") && isAllowedScope(r.URL.RawQuery, "%s", allowedDepartmentIds) {
				scopeName = DepartmentName
			} else if strings.Contains(r.URL.RawQuery, "project") && isAllowedScope(r.URL.RawQuery, "%s", allowedProjectIds) {
				scopeName = ProjectName
			}
		}

		if scopeName == "" {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "Not found")
		} else {
			var jsonResponse []byte
			var err error
			if strings.Contains(r.URL.Path, "org-unit/names") {
				jsonResponse, err = json.Marshal([]map[string]string{{"name": scopeName}})
			} else {
				jsonResponse, err = json.Marshal(map[string]string{"name": scopeName})
			}
			Expect(err).To(BeNil())

			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonResponse)
		}
	}))
	serverConfig := config.Get[*authorization_config.Server]()
	serverConfig.ClusterServiceUrl = mockServer.URL
	serverConfig.OrgUnitServiceUrl = mockServer.URL
	serverConfig.TenantServiceUrl = mockServer.URL
	config.Set(serverConfig)
}

func isAllowedScope(path, urlScopeFmt string, allowedScopeIds []string) bool {
	for _, allowedScopeId := range allowedScopeIds {
		if strings.Contains(path, fmt.Sprintf(urlScopeFmt, allowedScopeId)) {
			return true
		}
	}
	return false
}

func TestsCleanupAndInit() {
	testing.CleanTables()
	testing.CreateTables()
	dao.InitStaticPermissionSets()
	dao.InitStaticLegacyRoles()
	dao.InitStaticRoles()
	dao.InitAccessRulesForTests(testing.TestTenant, testing.ManagementTestTenant)
	dao.InitRolesMap(&RolesMap)
}

func FormatTime(time time.Time) string {
	return time.UTC().Format(TimeLayout)
}
