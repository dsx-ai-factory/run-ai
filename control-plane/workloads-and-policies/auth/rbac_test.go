/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"context"
	"net/http"

	authClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	scope "github.com/dsx-ai-factory/run-ai/runai-application-packages/scopeinfo"
	orgUnitClient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func strPtr(s string) *string {
	return &s
}

// CrossTenantMockAuthClient is a mock that simulates a user with permissions only on a specific tenant.
// PermittedClusters: if non-nil, returned in scopes (used to simulate explicit cluster list so the
// GetAncestors-failure fallback does not apply; cross-tenant user with their own clusters gets DENY).
type CrossTenantMockAuthClient struct {
	permittedTenantId string
	PermittedClusters []string // optional: when set, scopes include these clusters (fallback then requires len(Clusters)==0)
}

func (m *CrossTenantMockAuthClient) Client() authorization.ClientInterface {
	return m
}

func (m *CrossTenantMockAuthClient) CheckPermittedScopes(_ context.Context, _ authClient.ResourceType, _ authClient.Action) (*authClient.PermittedScopes, *authClient.Error) {
	scopes := &authClient.PermittedScopes{
		Tenants: []string{m.permittedTenantId},
	}
	if len(m.PermittedClusters) > 0 {
		scopes.Clusters = m.PermittedClusters
	}
	return scopes, nil
}

// CrossTenantMockOrgUnitClient provides ancestors for resources
type CrossTenantMockOrgUnitClient struct {
	ancestorsMap map[string]*orgUnitClient.Ancestors
}

func (m *CrossTenantMockOrgUnitClient) Client() orgunit.ClientInterface {
	return m
}

func (m *CrossTenantMockOrgUnitClient) GetAncestors(_ context.Context, _ string, id string) (*orgUnitClient.Ancestors, *orgUnitClient.Error) {
	if ancestors, ok := m.ancestorsMap[id]; ok {
		return ancestors, nil
	}
	return &orgUnitClient.Ancestors{}, nil
}

func (m *CrossTenantMockOrgUnitClient) GetOrgTree(_ context.Context) ([]orgUnitClient.OrgUnit, *orgUnitClient.Error) {
	return []orgUnitClient.OrgUnit{}, nil
}

func (m *CrossTenantMockOrgUnitClient) ListProjects(_ context.Context, _ string) ([]orgUnitClient.ProjectMetadata, *orgUnitClient.Error) {
	return []orgUnitClient.ProjectMetadata{}, nil
}

// FailingAncestorsMockBackendClient always returns an error from GetAncestors,
// simulating a cross-tenant cluster lookup failure.
type FailingAncestorsMockBackendClient struct{}

func (m *FailingAncestorsMockBackendClient) Client() orgunit.ClientInterface {
	return m
}

func (m *FailingAncestorsMockBackendClient) GetAncestors(_ context.Context, _ string, _ string) (*orgUnitClient.Ancestors, *orgUnitClient.Error) {
	return nil, orgUnitClient.NewError(http.StatusNotFound, "not found")
}

func (m *FailingAncestorsMockBackendClient) GetOrgTree(_ context.Context) ([]orgUnitClient.OrgUnit, *orgUnitClient.Error) {
	return []orgUnitClient.OrgUnit{}, nil
}

func (m *FailingAncestorsMockBackendClient) ListProjects(_ context.Context, _ string) ([]orgUnitClient.ProjectMetadata, *orgUnitClient.Error) {
	return []orgUnitClient.ProjectMetadata{}, nil
}

var _ = Describe("RBAC Cross-Tenant Authorization", func() {
	const (
		tenantA       = "tenant-a-id"
		tenantB       = "tenant-b-id"
		clusterInA    = "cluster-in-tenant-a"
		clusterInB    = "cluster-in-tenant-b"
		departmentInA = "dept-in-tenant-a"
		departmentInB = "dept-in-tenant-b"
		projectInA    = "proj-in-tenant-a"
		projectInB    = "proj-in-tenant-b"
	)

	BeforeEach(func() {
		orgUnitMock := &CrossTenantMockOrgUnitClient{
			ancestorsMap: map[string]*orgUnitClient.Ancestors{
				clusterInA:    {TenantId: tenantA},
				clusterInB:    {TenantId: tenantB},
				departmentInA: {TenantId: tenantA},
				departmentInB: {TenantId: tenantB},
				projectInA:    {TenantId: tenantA},
				projectInB:    {TenantId: tenantB},
			},
		}
		orgunit.SetMock(orgUnitMock)
	})

	Context("Cross-tenant access with tenant-level permissions", func() {
		BeforeEach(func() {
			authMock := &CrossTenantMockAuthClient{
				permittedTenantId: tenantA,
			}
			authorization.SetMock(authMock)
		})

		It("should DENY access when user from Tenant A tries to update cluster-scoped asset in Tenant B", func() {
			ctx := context.Background()

			scopeInfo := &scope.ScopeInfo{
				Scope:   scope.Cluster,
				ScopeId: clusterInB,
			}

			err := ValidateAccessByResourceTypeAndAction(
				ctx,
				scopeInfo,
				authClient.RESOURCETYPE_ENVIRONMENTS,
				authClient.ACTION_UPDATE,
				false,
			)

			Expect(err).ToNot(BeNil(), "Expected authorization to be DENIED for cross-tenant access")
			Expect(err.Code).To(Equal(int32(http.StatusForbidden)))
		})

		It("should DENY access when user from Tenant A tries to update department-scoped asset in Tenant B", func() {
			ctx := context.Background()

			scopeInfo := &scope.ScopeInfo{
				Scope:   scope.Department,
				ScopeId: departmentInB,
			}

			err := ValidateAccessByResourceTypeAndAction(
				ctx,
				scopeInfo,
				authClient.RESOURCETYPE_ENVIRONMENTS,
				authClient.ACTION_UPDATE,
				false,
			)

			Expect(err).ToNot(BeNil(), "Expected authorization to be DENIED for cross-tenant access")
			Expect(err.Code).To(Equal(int32(http.StatusForbidden)))
		})

		It("should DENY access when user from Tenant A tries to update project-scoped asset in Tenant B", func() {
			ctx := context.Background()

			scopeInfo := &scope.ScopeInfo{
				Scope:   scope.Project,
				ScopeId: projectInB,
			}

			err := ValidateAccessByResourceTypeAndAction(
				ctx,
				scopeInfo,
				authClient.RESOURCETYPE_ENVIRONMENTS,
				authClient.ACTION_UPDATE,
				false,
			)

			Expect(err).ToNot(BeNil(), "Expected authorization to be DENIED for cross-tenant access")
			Expect(err.Code).To(Equal(int32(http.StatusForbidden)))
		})

		It("should DENY access when user from Tenant A tries to delete cluster-scoped asset in Tenant B", func() {
			ctx := context.Background()

			scopeInfo := &scope.ScopeInfo{
				Scope:   scope.Cluster,
				ScopeId: clusterInB,
			}

			err := ValidateAccessByResourceTypeAndAction(
				ctx,
				scopeInfo,
				authClient.RESOURCETYPE_ENVIRONMENTS,
				authClient.ACTION_DELETE,
				false,
			)

			Expect(err).ToNot(BeNil(), "Expected authorization to be DENIED for cross-tenant access")
			Expect(err.Code).To(Equal(int32(http.StatusForbidden)))
		})

		It("should DENY access when user from Tenant A tries to create cluster-scoped asset in Tenant B", func() {
			ctx := context.Background()

			scopeInfo := &scope.ScopeInfo{
				Scope:   scope.Cluster,
				ScopeId: clusterInB,
			}

			err := ValidateAccessByResourceTypeAndAction(
				ctx,
				scopeInfo,
				authClient.RESOURCETYPE_ENVIRONMENTS,
				authClient.ACTION_CREATE,
				false,
			)

			Expect(err).ToNot(BeNil(), "Expected authorization to be DENIED for cross-tenant access")
			Expect(err.Code).To(Equal(int32(http.StatusForbidden)))
		})

		It("should ALLOW access when user from Tenant A accesses cluster-scoped asset in Tenant A (same tenant)", func() {
			ctx := context.Background()

			scopeInfo := &scope.ScopeInfo{
				Scope:   scope.Cluster,
				ScopeId: clusterInA,
			}

			err := ValidateAccessByResourceTypeAndAction(
				ctx,
				scopeInfo,
				authClient.RESOURCETYPE_ENVIRONMENTS,
				authClient.ACTION_UPDATE,
				false,
			)

			Expect(err).To(BeNil(), "Expected authorization to be ALLOWED for same-tenant access")
		})
	})

	Context("Cross-tenant access affects all asset types", func() {
		BeforeEach(func() {
			authMock := &CrossTenantMockAuthClient{
				permittedTenantId: tenantA,
			}
			authorization.SetMock(authMock)
		})

		DescribeTable("should DENY cross-tenant access for all resource types",
			func(resourceType authClient.ResourceType) {
				ctx := context.Background()

				scopeInfo := &scope.ScopeInfo{
					Scope:   scope.Cluster,
					ScopeId: clusterInB,
				}

				err := ValidateAccessByResourceTypeAndAction(
					ctx,
					scopeInfo,
					resourceType,
					authClient.ACTION_UPDATE,
					false,
				)

				Expect(err).ToNot(BeNil(), "Expected authorization to be DENIED for cross-tenant access on %s", resourceType)
				Expect(err.Code).To(Equal(int32(http.StatusForbidden)))
			},
			Entry("Environments", authClient.RESOURCETYPE_ENVIRONMENTS),
			Entry("Credentials", authClient.RESOURCETYPE_CREDENTIALS),
			Entry("Compute Resources", authClient.RESOURCETYPE_COMPUTE_RESOURCES),
			Entry("PVC Assets", authClient.RESOURCETYPE_PVC_ASSETS),
			Entry("S3 Assets", authClient.RESOURCETYPE_S3_ASSETS),
			Entry("NFS Assets", authClient.RESOURCETYPE_NFS_ASSETS),
			Entry("Host Path Assets", authClient.RESOURCETYPE_HOST_PATH_ASSETS),
			Entry("Git Assets", authClient.RESOURCETYPE_GIT_ASSETS),
			Entry("Templates", authClient.RESOURCETYPE_TEMPLATES),
			Entry("Policies", authClient.RESOURCETYPE_POLICIES),
			Entry("Workspaces", authClient.RESOURCETYPE_WORKSPACES),
			Entry("Trainings", authClient.RESOURCETYPE_TRAININGS),
			Entry("Inferences", authClient.RESOURCETYPE_INFERENCES),
		)
	})

	Context("Cross-tenant access with Tenants array (new format)", func() {
		It("should DENY access when user has multiple tenants but not the target tenant", func() {
			authMock := &CrossTenantMockAuthClient{
				permittedTenantId: tenantA,
			}
			authorization.SetMock(authMock)

			ctx := context.Background()

			scopeInfo := &scope.ScopeInfo{
				Scope:   scope.Cluster,
				ScopeId: clusterInB,
			}

			err := ValidateAccessByResourceTypeAndAction(
				ctx,
				scopeInfo,
				authClient.RESOURCETYPE_ENVIRONMENTS,
				authClient.ACTION_UPDATE,
				false,
			)

			Expect(err).ToNot(BeNil(), "Expected authorization to be DENIED")
			Expect(err.Code).To(Equal(int32(http.StatusForbidden)))
		})
	})

	Context("Same-tenant access should be ALLOWED", func() {
		BeforeEach(func() {
			authMock := &CrossTenantMockAuthClient{
				permittedTenantId: tenantA,
			}
			authorization.SetMock(authMock)
		})

		It("should ALLOW access to cluster-scoped asset in same tenant", func() {
			ctx := context.Background()
			scopeInfo := &scope.ScopeInfo{Scope: scope.Cluster, ScopeId: clusterInA}

			err := ValidateAccessByResourceTypeAndAction(ctx, scopeInfo, authClient.RESOURCETYPE_ENVIRONMENTS, authClient.ACTION_UPDATE, false)

			Expect(err).To(BeNil())
		})

		It("should ALLOW access to department-scoped asset in same tenant", func() {
			ctx := context.Background()
			scopeInfo := &scope.ScopeInfo{Scope: scope.Department, ScopeId: departmentInA}

			err := ValidateAccessByResourceTypeAndAction(ctx, scopeInfo, authClient.RESOURCETYPE_ENVIRONMENTS, authClient.ACTION_UPDATE, false)

			Expect(err).To(BeNil())
		})

		It("should ALLOW access to project-scoped asset in same tenant", func() {
			ctx := context.Background()
			scopeInfo := &scope.ScopeInfo{Scope: scope.Project, ScopeId: projectInA}

			err := ValidateAccessByResourceTypeAndAction(ctx, scopeInfo, authClient.RESOURCETYPE_ENVIRONMENTS, authClient.ACTION_UPDATE, false)

			Expect(err).To(BeNil())
		})

		It("should ALLOW access to tenant-scoped asset in same tenant", func() {
			ctx := context.Background()
			scopeInfo := &scope.ScopeInfo{Scope: scope.Tenant, ScopeId: tenantA}

			err := ValidateAccessByResourceTypeAndAction(ctx, scopeInfo, authClient.RESOURCETYPE_ENVIRONMENTS, authClient.ACTION_UPDATE, false)

			Expect(err).To(BeNil())
		})

		DescribeTable("should ALLOW same-tenant access for all actions",
			func(action authClient.Action) {
				ctx := context.Background()
				scopeInfo := &scope.ScopeInfo{Scope: scope.Cluster, ScopeId: clusterInA}

				err := ValidateAccessByResourceTypeAndAction(ctx, scopeInfo, authClient.RESOURCETYPE_ENVIRONMENTS, action, false)

				Expect(err).To(BeNil(), "Expected %s action to be ALLOWED for same-tenant access", action)
			},
			Entry("CREATE", authClient.ACTION_CREATE),
			Entry("UPDATE", authClient.ACTION_UPDATE),
			Entry("DELETE", authClient.ACTION_DELETE),
			Entry("READ", authClient.ACTION_READ),
		)

		DescribeTable("should ALLOW same-tenant access for all resource types",
			func(resourceType authClient.ResourceType) {
				ctx := context.Background()
				scopeInfo := &scope.ScopeInfo{Scope: scope.Cluster, ScopeId: clusterInA}

				err := ValidateAccessByResourceTypeAndAction(ctx, scopeInfo, resourceType, authClient.ACTION_UPDATE, false)

				Expect(err).To(BeNil(), "Expected %s to be ALLOWED for same-tenant access", resourceType)
			},
			Entry("Environments", authClient.RESOURCETYPE_ENVIRONMENTS),
			Entry("Credentials", authClient.RESOURCETYPE_CREDENTIALS),
			Entry("Compute Resources", authClient.RESOURCETYPE_COMPUTE_RESOURCES),
			Entry("PVC Assets", authClient.RESOURCETYPE_PVC_ASSETS),
			Entry("S3 Assets", authClient.RESOURCETYPE_S3_ASSETS),
			Entry("NFS Assets", authClient.RESOURCETYPE_NFS_ASSETS),
			Entry("Host Path Assets", authClient.RESOURCETYPE_HOST_PATH_ASSETS),
			Entry("Git Assets", authClient.RESOURCETYPE_GIT_ASSETS),
			Entry("Templates", authClient.RESOURCETYPE_TEMPLATES),
			Entry("Policies", authClient.RESOURCETYPE_POLICIES),
			Entry("Workspaces", authClient.RESOURCETYPE_WORKSPACES),
			Entry("Trainings", authClient.RESOURCETYPE_TRAININGS),
			Entry("Inferences", authClient.RESOURCETYPE_INFERENCES),
		)
	})

	Context("when GetAncestors fails for cluster-scoped resource", func() {
		It("returns 403 when user has explicit clusters in scopes (safe fallback does not apply)", func() {
			authorization.SetMock(&CrossTenantMockAuthClient{permittedTenantId: "100", PermittedClusters: []string{"user-tenant-cluster"}})
			orgunit.SetMock(&FailingAncestorsMockBackendClient{})
			ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "100")
			scopeInfo := &scope.ScopeInfo{Scope: scope.Cluster, ScopeId: "cluster-in-another-tenant"}

			err := ValidateAccessByResourceTypeAndAction(ctx, scopeInfo, authClient.RESOURCETYPE_ENVIRONMENTS, authClient.ACTION_CREATE, false)

			Expect(err).ToNot(BeNil())
			Expect(err.Code).To(Equal(int32(http.StatusForbidden)))
		})

		DescribeTable("returns 403 for each action when GetAncestors fails and user has explicit clusters",
			func(action authClient.Action) {
				authorization.SetMock(&CrossTenantMockAuthClient{permittedTenantId: "100", PermittedClusters: []string{"other"}})
				orgunit.SetMock(&FailingAncestorsMockBackendClient{})
				ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "100")
				scopeInfo := &scope.ScopeInfo{Scope: scope.Cluster, ScopeId: "unknown-cluster"}

				err := ValidateAccessByResourceTypeAndAction(ctx, scopeInfo, authClient.RESOURCETYPE_CREDENTIALS, action, false)

				Expect(err).ToNot(BeNil())
				Expect(err.Code).To(Equal(int32(http.StatusForbidden)))
			},
			Entry("CREATE", authClient.ACTION_CREATE),
			Entry("UPDATE", authClient.ACTION_UPDATE),
			Entry("DELETE", authClient.ACTION_DELETE),
		)

		It("allows access when user has tenant-level-only permission and no clusters in scopes", func() {
			authorization.SetMock(&CrossTenantMockAuthClient{permittedTenantId: "100"})
			orgunit.SetMock(&FailingAncestorsMockBackendClient{})
			ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "100")
			scopeInfo := &scope.ScopeInfo{Scope: scope.Cluster, ScopeId: "any-cluster"}

			err := ValidateAccessByResourceTypeAndAction(ctx, scopeInfo, authClient.RESOURCETYPE_CREDENTIALS, authClient.ACTION_CREATE, false)

			Expect(err).To(BeNil())
		})
	})

})
