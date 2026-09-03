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
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/cluster"
	orgUnitService "github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbtables"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	scope "github.com/dsx-ai-factory/run-ai/runai-application-packages/scopeinfo"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/tree"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	forWorkspaceUrl            = Entry(urlPolicyWorkspace, urlPolicyWorkspace)
	forTrainingUrl             = Entry(urlPolicyTraining, urlPolicyTraining)
	forDistributedUrl          = Entry(urlPolicyDistributed, urlPolicyDistributed)
	forInferenceUrl            = Entry(urlPolicyInference, urlPolicyInference)
	forDistributedInferenceUrl = Entry(urlPolicyDistributedInference, urlPolicyDistributedInference)
)

var _ = Describe("DeleteTest", func() {
	router := CreateRouter()

	var backendMock *BackendMockClient
	var clusterMock cluster.SimpleMockClient

	BeforeEach(func() {
		database.DeleteAllTablesForTesting()

		tree.Restart()

		backendMock = NewBackendMockClient(2, 2)
		orgUnitService.SetMock(NewPoliciesOrgUnitClient(backendMock))
		orgunit.SetMock(NewPoliciesOrgUnitClient(backendMock))
		authorization.SetMock(NewPoliciesAuthClient(backendMock))

		clusterMock = cluster.NewSimpleMockClient().WithClusters(backendMock.ClusterIdPerTenant)
		cluster.SetMock(clusterMock)
	})

	/////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Authentication
	/////////////////////////////////////////////////////////////////////////////////////////////////////////

	Context("no authentication header", func() {
		It("should fail forbidden error", func() {
			request, err := http.NewRequest("DELETE", fmt.Sprintf("%s?scope=%s&departmentId=%s", urlPolicyWorkspace, api.SCOPETYPE_DEPARTMENT, "111"), nil)
			Expect(err).To(BeNil())

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusForbidden))
		})
	})

	Context("researcher attempt to delete policy", func() {
		It("should fail forbidden error", func() {
			request, err := http.NewRequest("DELETE", fmt.Sprintf("%s?scope=%s&departmentId=%s", urlPolicyWorkspace, api.SCOPETYPE_DEPARTMENT, "111"), nil)
			Expect(err).To(BeNil())
			request.Header = *headerWithMissingInfo()

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusForbidden))
		})
	})

	/////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Behavior
	/////////////////////////////////////////////////////////////////////////////////////////////////////////

	Context("system scope", func() {
		It("should return with error", func() {
			request, err := http.NewRequest("DELETE", fmt.Sprintf("%s?scope=%s", urlPolicyWorkspace, api.SCOPETYPE_SYSTEM), nil)
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(2)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.String()).To(ContainSubstring("system scope is not supported for policies"))
		})
	})

	Context("non numeric project", func() {
		DescribeTable("delete with non numeric project", func(url string) {
			request, err := http.NewRequest("DELETE", fmt.Sprintf("%s?scope=%s&projectId=abcde", url, api.SCOPETYPE_PROJECT), nil)
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(2)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.String()).To(ContainSubstring("project id 'abcde' is not numeric"))
		}, forWorkspaceUrl, forTrainingUrl, forDistributedUrl, forInferenceUrl, forDistributedInferenceUrl)
	})

	Context("non existent policy", func() {
		It("should return error", func() {
			request, err := http.NewRequest("DELETE", fmt.Sprintf("%s?scope=%s&departmentId=%s", urlPolicyWorkspace, api.SCOPETYPE_DEPARTMENT, "111"), nil)
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(1)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("handle project deletion", func() {
		It("basic flow", func() {
			targetProjectId := int32(1111)

			cluster1 := backendMock.ClusterIdPerTenant["1"][0]
			ctx := tenantContextForTest(1)

			injectPolicy(ctx, router, api.SCOPETYPE_PROJECT, api.POLICYTYPE_WORKSPACE, 1, cluster1, targetProjectId, "p111-1")
			injectPolicy(ctx, router, api.SCOPETYPE_PROJECT, api.POLICYTYPE_TRAINING, 1, cluster1, targetProjectId, "p111-2")
			injectPolicy(ctx, router, api.SCOPETYPE_PROJECT, api.POLICYTYPE_DISTRIBUTED, 1, cluster1, targetProjectId, "p111-3")
			injectPolicy(ctx, router, api.SCOPETYPE_PROJECT, api.POLICYTYPE_INFERENCE, 1, cluster1, targetProjectId, "p111-4")
			injectPolicy(ctx, router, api.SCOPETYPE_PROJECT, api.POLICYTYPE_DISTRIBUTED_INFERENCE, 1, cluster1, targetProjectId, "p111-4")
			injectPolicy(ctx, router, api.SCOPETYPE_PROJECT, api.POLICYTYPE_FALLBACK, 1, cluster1, targetProjectId, "p111-5")

			injectPolicy(ctx, router, api.SCOPETYPE_PROJECT, api.POLICYTYPE_WORKSPACE, 1, cluster1, 1112, "p222-1")
			injectPolicy(ctx, router, api.SCOPETYPE_PROJECT, api.POLICYTYPE_TRAINING, 1, cluster1, 1112, "p222-2")
			injectPolicy(ctx, router, api.SCOPETYPE_PROJECT, api.POLICYTYPE_DISTRIBUTED, 1, cluster1, 1112, "p222-3")
			injectPolicy(ctx, router, api.SCOPETYPE_PROJECT, api.POLICYTYPE_INFERENCE, 1, cluster1, 1112, "p222-4")
			injectPolicy(ctx, router, api.SCOPETYPE_PROJECT, api.POLICYTYPE_DISTRIBUTED_INFERENCE, 1, cluster1, 1112, "p222-4")
			injectPolicy(ctx, router, api.SCOPETYPE_PROJECT, api.POLICYTYPE_FALLBACK, 1, cluster1, 1112, "p222-5")

			// Test case: with orphans noting deleted
			err := HandleProjectDeleteRequest(ctx, &events.ProjectDeleteRequest{
				ProjectId:    fmt.Sprintf("%d", targetProjectId),
				ClusterId:    cluster1,
				LeaveOrphans: true,
			})
			Expect(err).To(BeNil())
			policies, err := listPolicies(ctx)
			for _, p := range policies {
				// in non-tree policy there are empty policies that are created for department, cluster, etc
				// that are all deleted so we do not want to check them
				if p.Scope == api.SCOPETYPE_PROJECT {
					Expect(p.DeletedAt).To(BeZero())
				}
			}

			// Test case: without orphans
			err = HandleProjectDeleteRequest(ctx, &events.ProjectDeleteRequest{
				ProjectId:    fmt.Sprintf("%d", targetProjectId),
				ClusterId:    cluster1,
				LeaveOrphans: false,
			})
			Expect(err).To(BeNil())
			policies, err = listPolicies(ctx)
			for _, p := range policies {
				// Verify that the policies of the target project are deleted
				if p.Scope != api.SCOPETYPE_PROJECT {
					continue
				}
				if convutils.Deref(p.ProjectId, 0) == targetProjectId {
					Expect(p.DeletedAt).NotTo(BeZero())
					continue
				}
				// Verify that the policies of the other project are not deleted
				Expect(p.DeletedAt).To(BeZero())
			}
		})
	})

	Context("handle cluster deletion", func() {
		It("basic flow", func() {
			cluster1 := backendMock.ClusterIdPerTenant["1"][0]
			cluster2 := backendMock.ClusterIdPerTenant["1"][1]
			ctx := tenantContextForTest(1)

			injectPolicy(ctx, router, api.SCOPETYPE_PROJECT, api.POLICYTYPE_WORKSPACE, 1, cluster1, 1111, "c1-p111-1")
			injectPolicy(ctx, router, api.SCOPETYPE_PROJECT, api.POLICYTYPE_WORKSPACE, 1, cluster1, 1112, "c1-p122-1")
			injectPolicy(ctx, router, api.SCOPETYPE_PROJECT, api.POLICYTYPE_WORKSPACE, 1, cluster2, 1211, "c2-p222-1")
			injectPolicy(ctx, router, api.SCOPETYPE_CLUSTER, api.POLICYTYPE_WORKSPACE, 1, cluster1, 0, "c1-1")
			injectPolicy(ctx, router, api.SCOPETYPE_CLUSTER, api.POLICYTYPE_WORKSPACE, 1, cluster2, 0, "c2-1")

			// Test case: with orphans noting deleted
			err := HandleClusterDeleteRequest(ctx, &events.ClusterDelete{
				ClusterId: cluster1,
			})
			Expect(err).To(BeNil())
			policies, err := listPolicies(ctx)
			for _, p := range policies {
				Expect(p.ClusterId).NotTo(Equal(cluster1))
			}
		})
	})

	Context("handle tenant deletion", func() {
		It("basic flow", func() {
			cluster11 := backendMock.ClusterIdPerTenant["1"][0]
			cluster12 := backendMock.ClusterIdPerTenant["1"][1]
			ctx := tenantContextForTest(1)

			// create policies under tenant 1
			injectPolicy(ctx, router, api.SCOPETYPE_PROJECT, api.POLICYTYPE_WORKSPACE, 1, cluster11, 1111, "p1-c1-t1")
			injectPolicy(ctx, router, api.SCOPETYPE_PROJECT, api.POLICYTYPE_WORKSPACE, 1, cluster11, 1112, "p2-c1-t1")
			injectPolicy(ctx, router, api.SCOPETYPE_CLUSTER, api.POLICYTYPE_WORKSPACE, 1, cluster11, 0, "c1-t1")
			injectPolicy(ctx, router, api.SCOPETYPE_CLUSTER, api.POLICYTYPE_WORKSPACE, 1, cluster12, 0, "c2-t1")
			injectPolicy(ctx, router, api.SCOPETYPE_TENANT, api.POLICYTYPE_WORKSPACE, 1, "", 0, "t1")

			policies, err := listPolicies(ctx)
			Expect(len(policies)).NotTo(BeZero())

			// create policies under tenant 2
			cluster21 := backendMock.ClusterIdPerTenant["2"][0]
			cluster22 := backendMock.ClusterIdPerTenant["2"][1]
			ctx = tenantContextForTest(2)

			// create policies under tenant 1
			injectPolicy(ctx, router, api.SCOPETYPE_PROJECT, api.POLICYTYPE_WORKSPACE, 2, cluster21, 2111, "p1-c3-t2")
			injectPolicy(ctx, router, api.SCOPETYPE_PROJECT, api.POLICYTYPE_WORKSPACE, 2, cluster21, 2112, "p2-c3-t2")
			injectPolicy(ctx, router, api.SCOPETYPE_CLUSTER, api.POLICYTYPE_WORKSPACE, 2, cluster21, 0, "c3-t2")
			injectPolicy(ctx, router, api.SCOPETYPE_CLUSTER, api.POLICYTYPE_WORKSPACE, 2, cluster22, 0, "c4-t2")
			injectPolicy(ctx, router, api.SCOPETYPE_TENANT, api.POLICYTYPE_WORKSPACE, 2, "", 0, "t2")

			ctx = tenantContextForTest(1)
			// delete the first tenant
			err = HandleTenantDeleteRequest(ctx, &events.TenantDelete{
				TenantId: "1",
			})
			Expect(err).To(BeNil())
			policies, err = listPolicies(ctx)
			// validate policies from tenant 1 were deleted
			Expect(len(policies)).To(BeZero())

			ctx = tenantContextForTest(2)
			policies, err = listPolicies(ctx)
			// validate policies from tenant 2 still exist
			Expect(len(policies)).NotTo(BeZero())
		})
	})

	// -------------------------------------------------------------------------
	// Tenant-scoped cleanup of mappings and breakdown caches on
	// project / cluster / tenant deletion events.
	// -------------------------------------------------------------------------
	Context("mappings and cache cleanup on deletion events", func() {
		const (
			tenantA int32 = 1
			tenantB int32 = 2
		)

		// seedCacheForTenant inserts one entities_cache row and one
		// entities_dependency row owned by the given tenant. returns the
		// generated cache id so the test can verify it later.
		seedCacheForTenant := func(tenantId int32, entityId string) uuid.UUID {
			cacheId := uuid.New()
			cache := dbmodel.EntitiesCache{
				ID:         cacheId,
				TenantId:   tenantId,
				EntityType: dbmodel.EntityTypeBreakdown,
				EntityId:   entityId,
				CachedInfo: dbmodels.JsonBMap{},
			}
			Expect(runaidb.DB().Create(&cache)).Error().To(BeNil())

			dep := dbmodel.EntitiesDependency{
				CacheId:       cacheId,
				TenantId:      tenantId,
				ComponentType: dbmodel.ComponentTypeMapping,
				ComponentId:   uuid.New().String(),
			}
			Expect(runaidb.DB().Create(&dep)).Error().To(BeNil())

			return cacheId
		}

		// seedTenantMappings inserts one mapping_selectors, one
		// mapping_field_groups, one mapping_formatters, and one
		// mapping_locking row for the given tenant at tenant scope.
		seedTenantMappings := func(tenantId int32, suffix string) {
			Expect(runaidb.DB().Create(&dbmodel.MappingSelectors{
				ID:            uuid.New(),
				GvkGroup:      "test." + suffix,
				GvkKind:       "Kind" + suffix,
				GvkVersion:    "v1",
				ScopeType:     api.MAPPINGSCOPE_TENANT,
				TenantId:      tenantId,
				SpecSelectors: map[string]api.SpecSelectorsWithScopeValue{},
				CreatedBy:     "test",
				CreatedAt:     time.Now(),
				UpdatedBy:     "test",
				UpdatedAt:     time.Now(),
			})).Error().To(BeNil())

			Expect(runaidb.DB().Create(&dbmodel.MappingFieldGroups{
				ID:            uuid.New(),
				Name:          "fg-" + suffix,
				ScopeType:     api.MAPPINGSCOPE_TENANT,
				TenantId:      tenantId,
				Fields:        map[string]api.FieldGroupFieldsWithScopeValue{},
				ItemSelectors: map[string]api.ItemSelectorsWithScopeValue{},
				CreatedBy:     "test",
				CreatedAt:     time.Now(),
				UpdatedBy:     "test",
				UpdatedAt:     time.Now(),
			})).Error().To(BeNil())

			Expect(runaidb.DB().Create(&dbmodel.MappingFormatters{
				ID:        uuid.New(),
				Name:      "fmt-" + suffix,
				ScopeType: api.MAPPINGSCOPE_TENANT,
				TenantId:  tenantId,
				Fields:    map[string]api.FormatterField{},
				CreatedBy: "test",
				CreatedAt: time.Now(),
				UpdatedBy: "test",
				UpdatedAt: time.Now(),
			})).Error().To(BeNil())

			Expect(runaidb.DB().Create(&dbmodel.MappingLocking{
				TenantId: tenantId,
				LockedAt: time.Now(),
			})).Error().To(BeNil())
		}

		// seedSystemMappings inserts one row per mapping table at system
		// scope (tenant_id=0) so the tests can verify they survive a
		// tenant-deletion cleanup.
		seedSystemMappings := func() {
			Expect(runaidb.DB().Create(&dbmodel.MappingSelectors{
				ID:            uuid.New(),
				GvkGroup:      "system.example.com",
				GvkKind:       "SystemKind",
				GvkVersion:    "v1",
				ScopeType:     api.MAPPINGSCOPE_SYSTEM,
				TenantId:      0,
				SpecSelectors: map[string]api.SpecSelectorsWithScopeValue{},
				CreatedBy:     "system",
				CreatedAt:     time.Now(),
				UpdatedBy:     "system",
				UpdatedAt:     time.Now(),
			})).Error().To(BeNil())

			Expect(runaidb.DB().Create(&dbmodel.MappingFieldGroups{
				ID:            uuid.New(),
				Name:          "system-fg",
				ScopeType:     api.MAPPINGSCOPE_SYSTEM,
				TenantId:      0,
				Fields:        map[string]api.FieldGroupFieldsWithScopeValue{},
				ItemSelectors: map[string]api.ItemSelectorsWithScopeValue{},
				CreatedBy:     "system",
				CreatedAt:     time.Now(),
				UpdatedBy:     "system",
				UpdatedAt:     time.Now(),
			})).Error().To(BeNil())

			Expect(runaidb.DB().Create(&dbmodel.MappingFormatters{
				ID:        uuid.New(),
				Name:      "runai-system-fmt", // system formatter names must start with "runai-"
				ScopeType: api.MAPPINGSCOPE_SYSTEM,
				TenantId:  0,
				Fields:    map[string]api.FormatterField{},
				CreatedBy: "system",
				CreatedAt: time.Now(),
				UpdatedBy: "system",
				UpdatedAt: time.Now(),
			})).Error().To(BeNil())
		}

		// countCachesForTenant returns the number of entities_cache rows for the tenant.
		countCachesForTenant := func(tenantId int32) int64 {
			var count int64
			Expect(runaidb.DB().Model(&dbmodel.EntitiesCache{}).
				Where(&dbmodel.EntitiesCache{TenantId: tenantId}).
				Count(&count)).To(BeNil())
			return count
		}

		// countDepsForTenant returns the number of entities_dependency rows for the tenant.
		countDepsForTenant := func(tenantId int32) int64 {
			var count int64
			Expect(runaidb.DB().Model(&dbmodel.EntitiesDependency{}).
				Where(&dbmodel.EntitiesDependency{TenantId: tenantId}).
				Count(&count)).To(BeNil())
			return count
		}

		// countMappingsForTenant returns total tenant-scoped rows across all mapping tables.
		countMappingsForTenant := func(tenantId int32) int64 {
			var total, n int64
			Expect(runaidb.DB().Model(&dbmodel.MappingSelectors{}).
				Where(&dbmodel.MappingSelectors{TenantId: tenantId, ScopeType: api.MAPPINGSCOPE_TENANT}).
				Count(&n)).To(BeNil())
			total += n
			Expect(runaidb.DB().Model(&dbmodel.MappingFieldGroups{}).
				Where(&dbmodel.MappingFieldGroups{TenantId: tenantId, ScopeType: api.MAPPINGSCOPE_TENANT}).
				Count(&n)).To(BeNil())
			total += n
			Expect(runaidb.DB().Model(&dbmodel.MappingFormatters{}).
				Where(&dbmodel.MappingFormatters{TenantId: tenantId, ScopeType: api.MAPPINGSCOPE_TENANT}).
				Count(&n)).To(BeNil())
			total += n
			Expect(runaidb.DB().Model(&dbmodel.MappingLocking{}).
				Where(&dbmodel.MappingLocking{TenantId: tenantId}).
				Count(&n)).To(BeNil())
			total += n
			return total
		}

		// countSystemMappings returns total system-scoped mapping rows.
		countSystemMappings := func() int64 {
			var total, n int64
			Expect(runaidb.DB().Model(&dbmodel.MappingSelectors{}).
				Where(&dbmodel.MappingSelectors{ScopeType: api.MAPPINGSCOPE_SYSTEM}).
				Count(&n)).To(BeNil())
			total += n
			Expect(runaidb.DB().Model(&dbmodel.MappingFieldGroups{}).
				Where(&dbmodel.MappingFieldGroups{ScopeType: api.MAPPINGSCOPE_SYSTEM}).
				Count(&n)).To(BeNil())
			total += n
			Expect(runaidb.DB().Model(&dbmodel.MappingFormatters{}).
				Where(&dbmodel.MappingFormatters{ScopeType: api.MAPPINGSCOPE_SYSTEM}).
				Count(&n)).To(BeNil())
			total += n
			return total
		}

		It("project deletion drops the deleted tenant's caches but keeps other tenants' caches", func() {
			seedCacheForTenant(tenantA, "policy-a-1")
			seedCacheForTenant(tenantA, "policy-a-2")
			seedCacheForTenant(tenantB, "policy-b-1")

			Expect(countCachesForTenant(tenantA)).To(Equal(int64(2)))
			Expect(countDepsForTenant(tenantA)).To(Equal(int64(2)))
			Expect(countCachesForTenant(tenantB)).To(Equal(int64(1)))

			ctx := tenantContextForTest(tenantA)
			cluster1 := backendMock.ClusterIdPerTenant["1"][0]
			err := HandleProjectDeleteRequest(ctx, &events.ProjectDeleteRequest{
				ProjectId: "1111",
				ClusterId: cluster1,
			})
			Expect(err).To(BeNil())

			Expect(countCachesForTenant(tenantA)).To(BeZero(), "tenant A caches should be cleared")
			Expect(countDepsForTenant(tenantA)).To(BeZero(), "tenant A dependency rows should be cleared")
			Expect(countCachesForTenant(tenantB)).To(Equal(int64(1)), "tenant B caches must not be touched")
			Expect(countDepsForTenant(tenantB)).To(Equal(int64(1)), "tenant B dependencies must not be touched")
		})

		It("project deletion is a no-op (for caches) when LeaveOrphans is true", func() {
			seedCacheForTenant(tenantA, "policy-a-1")
			Expect(countCachesForTenant(tenantA)).To(Equal(int64(1)))

			ctx := tenantContextForTest(tenantA)
			cluster1 := backendMock.ClusterIdPerTenant["1"][0]
			err := HandleProjectDeleteRequest(ctx, &events.ProjectDeleteRequest{
				ProjectId:    "1111",
				ClusterId:    cluster1,
				LeaveOrphans: true,
			})
			Expect(err).To(BeNil())

			Expect(countCachesForTenant(tenantA)).To(Equal(int64(1)),
				"LeaveOrphans=true must skip the cache wipe just like the policy wipe")
		})

		It("cluster deletion drops the deleted tenant's caches but keeps other tenants' caches", func() {
			seedCacheForTenant(tenantA, "policy-a-1")
			seedCacheForTenant(tenantA, "policy-a-2")
			seedCacheForTenant(tenantB, "policy-b-1")

			ctx := tenantContextForTest(tenantA)
			cluster1 := backendMock.ClusterIdPerTenant["1"][0]
			err := HandleClusterDeleteRequest(ctx, &events.ClusterDelete{ClusterId: cluster1})
			Expect(err).To(BeNil())

			Expect(countCachesForTenant(tenantA)).To(BeZero(), "tenant A caches should be cleared")
			Expect(countDepsForTenant(tenantA)).To(BeZero(), "tenant A dependency rows should be cleared")
			Expect(countCachesForTenant(tenantB)).To(Equal(int64(1)), "tenant B caches must not be touched")
			Expect(countDepsForTenant(tenantB)).To(Equal(int64(1)), "tenant B dependencies must not be touched")
		})

		It("tenant deletion drops the tenant's mappings and caches but keeps system rows and other tenants intact", func() {
			// Count any pre-existing system mappings (suite-wide seed creates the real system
			// formatters / field groups / mappings and they persist across specs).
			systemCountBefore := countSystemMappings()

			seedSystemMappings()
			seedTenantMappings(tenantA, "a")
			seedTenantMappings(tenantB, "b")
			seedCacheForTenant(tenantA, "policy-a-1")
			seedCacheForTenant(tenantB, "policy-b-1")

			Expect(countMappingsForTenant(tenantA)).To(Equal(int64(4)))
			Expect(countMappingsForTenant(tenantB)).To(Equal(int64(4)))
			Expect(countSystemMappings()).To(Equal(systemCountBefore + 3))
			Expect(countCachesForTenant(tenantA)).To(Equal(int64(1)))
			Expect(countCachesForTenant(tenantB)).To(Equal(int64(1)))

			ctx := tenantContextForTest(tenantA)
			err := HandleTenantDeleteRequest(ctx, &events.TenantDelete{TenantId: "1"})
			Expect(err).To(BeNil())

			Expect(countMappingsForTenant(tenantA)).To(BeZero(), "tenant A mappings should be wiped")
			Expect(countCachesForTenant(tenantA)).To(BeZero(), "tenant A caches should be wiped")
			Expect(countDepsForTenant(tenantA)).To(BeZero(), "tenant A deps should be wiped")

			Expect(countSystemMappings()).To(Equal(systemCountBefore+3), "system-scoped mappings must not be touched")
			Expect(countMappingsForTenant(tenantB)).To(Equal(int64(4)), "tenant B mappings must not be touched")
			Expect(countCachesForTenant(tenantB)).To(Equal(int64(1)), "tenant B caches must not be touched")
			Expect(countDepsForTenant(tenantB)).To(Equal(int64(1)), "tenant B deps must not be touched")
		})
	})

	Context("delete department policy", func() {
		It("project effective re-calculated", func() {
			//
			//  Part 1: Tenant policy
			//
			bodyTenantA := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t-1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Rules: &api.WorkspacePolicyRulesV2{
						Security: &api.SupersetRulesAllOfSecurity{
							RunAsUid: &api.IntegerRulesOptional{
								Min: convutils.Ref(int32(1)),
							},
						},
					},
				},
			}

			requestTenantA, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyTenantA))
			Expect(err).To(BeNil())
			requestTenantA.Header = *headerForAdmin(1)

			responseTenantA := httptest.NewRecorder()
			router.ServeHTTP(responseTenantA, requestTenantA)
			Expect(responseTenantA.Code).To(Equal(http.StatusOK))

			var actualRespTenantA api.WorkspacePolicyV2
			err = json.Unmarshal(responseTenantA.Body.Bytes(), &actualRespTenantA)
			Expect(err).To(BeNil())

			expected := api.WorkspacePolicyDefaultsAndRulesV2{
				Defaults: nil,
				Rules:    bodyTenantA.Policy.Rules,
			}
			expected.Rules.Security.RunAsUid.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_TENANT}

			dao.InjectSystemPolicyForTesting(&expected)

			testutils.Expect(*actualRespTenantA.Effective).ToEqual(expected)

			//
			//   Part 2 - Department policy adds runAsGid
			//
			bodyDeptA := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:        api.SCOPETYPE_DEPARTMENT,
					DepartmentId: convutils.Ref("111"),
					Name:         "pol-d-111",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Rules: &api.WorkspacePolicyRulesV2{
						Security: &api.SupersetRulesAllOfSecurity{
							RunAsGid: &api.IntegerRulesOptional{
								Min: convutils.Ref(int32(2)),
							},
						},
					},
				},
			}

			requestDeptA, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyDeptA))
			Expect(err).To(BeNil())
			requestDeptA.Header = *headerForAdmin(1)

			responseDeptA := httptest.NewRecorder()
			router.ServeHTTP(responseDeptA, requestDeptA)
			Expect(responseDeptA.Code).To(Equal(http.StatusOK))

			var actualRespDeptA api.WorkspacePolicyV2
			err = json.Unmarshal(responseDeptA.Body.Bytes(), &actualRespDeptA)
			Expect(err).To(BeNil())

			expectedRespDeptA := api.WorkspacePolicyDefaultsAndRulesV2{
				Defaults: nil,
				Rules: &api.WorkspacePolicyRulesV2{
					Security: &api.SupersetRulesAllOfSecurity{
						RunAsUid: &api.IntegerRulesOptional{
							Min: convutils.Ref(int32(1)),
						},
						RunAsGid: &api.IntegerRulesOptional{
							Min: convutils.Ref(int32(2)),
						},
					},
				},
			}
			expectedRespDeptA.Rules.Security.RunAsUid.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_TENANT}
			expectedRespDeptA.Rules.Security.RunAsGid.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_DEPARTMENT, DepartmentId: convutils.Ref("111")}

			dao.InjectSystemPolicyForTesting(&expectedRespDeptA)

			Expect(*actualRespDeptA.Effective).To(Equal(expectedRespDeptA))

			//
			//   Part 3 - Project policy inherits both dept and tenant
			//
			bodyProjA := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_PROJECT,
					ProjectId: convutils.Ref(int32(1111)),
					Name:      "pol-p-1111",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("image-proj"),
					},
				},
			}

			requestProjA, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyProjA))
			Expect(err).To(BeNil())
			requestProjA.Header = *headerForAdmin(1)

			responseProjA := httptest.NewRecorder()
			router.ServeHTTP(responseProjA, requestProjA)
			Expect(responseProjA.Code).To(Equal(http.StatusOK))

			var actualRespSetProjA api.WorkspacePolicyV2
			err = json.Unmarshal(responseProjA.Body.Bytes(), &actualRespSetProjA)
			Expect(err).To(BeNil())

			expectedRespProjA := api.WorkspacePolicyDefaultsAndRulesV2{
				Defaults: &api.WorkspacePolicyDefaultsV2{
					Image: convutils.Ref("image-proj"),
				},
				Rules: &api.WorkspacePolicyRulesV2{
					Security: &api.SupersetRulesAllOfSecurity{
						RunAsUid: &api.IntegerRulesOptional{
							Min: convutils.Ref(int32(1)),
						},
						RunAsGid: &api.IntegerRulesOptional{
							Min: convutils.Ref(int32(2)),
						},
					},
				},
			}
			expectedRespProjA.Rules.Security.RunAsUid.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_TENANT}
			expectedRespProjA.Rules.Security.RunAsGid.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_DEPARTMENT, DepartmentId: convutils.Ref("111")}

			dao.InjectSystemPolicyForTesting(&expectedRespProjA)

			Expect(*actualRespSetProjA.Effective).To(Equal(expectedRespProjA))

			//
			//   Part 4 - Delete the department policy
			//
			deletePolicyForTest(router, 1, fmt.Sprintf("%s?scope=%s&departmentId=%s", urlPolicyWorkspace, api.SCOPETYPE_DEPARTMENT, "111"))

			//
			//   Part 5 - Get the project policy - should inherit directly from the tenant
			//
			actualRespGetProjA := getWorkspacePolicyForTest(router, 1, fmt.Sprintf("%s?scope=project&projectId=1111", urlPolicyWorkspace))

			expectedEmpty := api.WorkspacePolicyDefaultsAndRulesV2{
				Defaults: &api.WorkspacePolicyDefaultsV2{
					Image: convutils.Ref("image-proj"),
				},
				Rules: nil,
			}
			Expect(*actualRespGetProjA.Policy).To(Equal(expectedEmpty))

			// the project will have the same effective as the tenant + its own default for the image
			actualRespTenantA.Effective.Defaults = &api.WorkspacePolicyDefaultsV2{
				Image: convutils.Ref("image-proj"),
			}

			dao.InjectSystemPolicyForTesting(actualRespTenantA.Effective)

			testutils.Expect(*actualRespGetProjA.Effective).ToEqual(*actualRespTenantA.Effective)
		})
	})

	Context("distributed and inference policies are distinct", func() {
		It("delete one does not delete the other", func() {
			distributedBody := api.DistributedPolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.DistributedPolicyDefaultsAndRulesV2{
					Defaults: &api.DistributedPolicyDefaultsV2{
						Worker: &api.DistributedPolicyDefaultsV2Worker{
							Image: convutils.Ref("ubuntu"),
						},
					},
				},
			}

			inferenceBody := api.InferencePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.InferencePolicyDefaultsAndRulesV2{
					Defaults: &api.InferencePolicyDefaultsV2{
						Image: convutils.Ref("ubuntu"),
					},
				},
			}

			//
			//   Phase 1 -> add them both
			//
			requestDist, err := http.NewRequest(http.MethodPatch, urlPolicyDistributed, makePayload(distributedBody))
			Expect(err).To(BeNil())
			requestDist.Header = *headerForAdmin(2)

			responseDist := httptest.NewRecorder()
			router.ServeHTTP(responseDist, requestDist)
			Expect(responseDist.Code).To(Equal(http.StatusOK))

			requestInference, err := http.NewRequest(http.MethodPatch, urlPolicyInference, makePayload(inferenceBody))
			Expect(err).To(BeNil())
			requestInference.Header = *headerForAdmin(2)

			responseInference := httptest.NewRecorder()
			router.ServeHTTP(responseInference, requestInference)
			Expect(responseInference.Code).To(Equal(http.StatusOK))

			//
			//   Phase 2 delete each of them should succeed
			//
			deletePolicyForTest(router, 2, fmt.Sprintf("%s?scope=tenant", urlPolicyInference))

			deletePolicyForTest(router, 2, fmt.Sprintf("%s?scope=tenant", urlPolicyDistributed))

			//
			//   Phase 3 delete again now fails
			//
			requestDeleteInf, err := http.NewRequest("DELETE", fmt.Sprintf("%s?scope=tenant", urlPolicyDistributed), makePayload(inferenceBody))
			Expect(err).To(BeNil())
			requestDeleteInf.Header = *headerForAdmin(2)

			responseDeleteInf := httptest.NewRecorder()
			router.ServeHTTP(responseDeleteInf, requestDeleteInf)
			Expect(responseDeleteInf.Code).To(Equal(http.StatusNotFound))

			requestDeleteDist, err := http.NewRequest("DELETE", fmt.Sprintf("%s?scope=tenant", urlPolicyDistributed), makePayload(distributedBody))
			Expect(err).To(BeNil())
			requestDeleteDist.Header = *headerForAdmin(2)

			responseDeleteDist := httptest.NewRecorder()
			router.ServeHTTP(responseDeleteDist, requestDeleteDist)
			Expect(responseDeleteDist.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("workspace and training policies are distinct", func() {
		It("delete one does not delete the other", func() {
			workspaceBody := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("ubuntu"),
					},
				},
			}

			trainingBody := api.TrainingPolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.TrainingPolicyDefaultsAndRulesV2{
					Defaults: &api.TrainingPolicyDefaultsV2{
						Image: convutils.Ref("ubuntu"),
					},
				},
			}

			//
			//   Phase 1 -> add them both
			//
			requestWs, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(workspaceBody))
			Expect(err).To(BeNil())
			requestWs.Header = *headerForAdmin(2)

			responseWs := httptest.NewRecorder()
			router.ServeHTTP(responseWs, requestWs)
			Expect(responseWs.Code).To(Equal(http.StatusOK))

			requestTrain, err := http.NewRequest(http.MethodPatch, urlPolicyTraining, makePayload(trainingBody))
			Expect(err).To(BeNil())
			requestTrain.Header = *headerForAdmin(2)

			responseTrain := httptest.NewRecorder()
			router.ServeHTTP(responseTrain, requestTrain)
			Expect(responseTrain.Code).To(Equal(http.StatusOK))

			//
			//   Phase 2 delete each of them should succeed
			//
			deletePolicyForTest(router, 2, fmt.Sprintf("%s?scope=tenant", urlPolicyWorkspace))

			deletePolicyForTest(router, 2, fmt.Sprintf("%s?scope=tenant", urlPolicyTraining))

			//
			//   Phase 3 delete again now fails
			//
			requestDeleteWs, err := http.NewRequest("DELETE", fmt.Sprintf("%s?scope=tenant", urlPolicyWorkspace), makePayload(workspaceBody))
			Expect(err).To(BeNil())
			requestDeleteWs.Header = *headerForAdmin(2)

			responseDeleteWs := httptest.NewRecorder()
			router.ServeHTTP(responseDeleteWs, requestDeleteWs)
			Expect(responseDeleteWs.Code).To(Equal(http.StatusNotFound))

			requestDeleteTrain, err := http.NewRequest("DELETE", fmt.Sprintf("%s?scope=tenant", urlPolicyTraining), makePayload(workspaceBody))
			Expect(err).To(BeNil())
			requestDeleteTrain.Header = *headerForAdmin(2)

			responseDeleteTrain := httptest.NewRecorder()
			router.ServeHTTP(responseDeleteTrain, requestDeleteTrain)
			Expect(responseDeleteTrain.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("policy is deleted and re-added with the same name", func() {
		It("re-adding should succeed", func() {
			// this test is important especially due to our use of soft-delete,
			// as re-adding, if not implemented correctly, may fail on 'already exists'
			workspaceBody := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("image"),
					},
				},
			}

			//
			//   Phase 1 -> add the policy
			//
			requestWs, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(workspaceBody))
			Expect(err).To(BeNil())
			requestWs.Header = *headerForAdmin(2)

			responseWs := httptest.NewRecorder()
			router.ServeHTTP(responseWs, requestWs)
			Expect(responseWs.Code).To(Equal(http.StatusOK))

			//
			//   Phase 2 delete the policy
			//
			deletePolicyForTest(router, 2, fmt.Sprintf("%s?scope=tenant", urlPolicyWorkspace))

			//
			//   Phase 3 re-add the same policy
			//
			requestWs, err = http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(workspaceBody))
			Expect(err).To(BeNil())
			requestWs.Header = *headerForAdmin(2)

			responseWs = httptest.NewRecorder()
			router.ServeHTTP(responseWs, requestWs)
			Expect(responseWs.Code).To(Equal(http.StatusOK))

			//
			//   Phase 4 delete the policy again
			//
			deletePolicyForTest(router, 2, fmt.Sprintf("%s?scope=tenant", urlPolicyWorkspace))

			//
			//   Phase 5 re-add the same policy once again
			//
			requestWs, err = http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(workspaceBody))
			Expect(err).To(BeNil())
			requestWs.Header = *headerForAdmin(2)

			responseWs = httptest.NewRecorder()
			router.ServeHTTP(responseWs, requestWs)
			Expect(responseWs.Code).To(Equal(http.StatusOK))
		})
	})

	Context("project policy inherits from cluster policy after deletion", func() {
		It("create project policy -> delete project policy -> create cluster policy -> effective policy should inherit from cluster", func() {
			//
			//   Part 1 - Project policy with image default
			//
			bodyProj := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_PROJECT,
					ProjectId: convutils.Ref(int32(1111)),
					Name:      "pol-p-1111",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("image-proj"),
					},
				},
			}

			requestProj, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyProj))
			Expect(err).To(BeNil())
			requestProj.Header = *headerForAdmin(1)

			responseProj := httptest.NewRecorder()
			router.ServeHTTP(responseProj, requestProj)
			Expect(responseProj.Code).To(Equal(http.StatusOK))

			// Verify initial project policy
			actualRespGetProj := getWorkspacePolicyForTest(router, 1, fmt.Sprintf("%s?scope=project&projectId=1111", urlPolicyWorkspace))
			Expect(*actualRespGetProj.Policy).To(Equal(*bodyProj.Policy))

			var expectedEffective api.WorkspacePolicyDefaultsAndRulesV2
			copyErr := convutils.CopySpecs(&expectedEffective, bodyProj.Policy)
			Expect(copyErr).To(BeNil())

			dao.InjectSystemPolicyForTesting(&expectedEffective)

			Expect(*actualRespGetProj.Effective).To(Equal(expectedEffective))

			//
			//   Part 2 - Delete the project policy
			//
			deletePolicyForTest(router, 1, fmt.Sprintf("%s?scope=project&projectId=1111", urlPolicyWorkspace))

			// Verify project policy is empty after deletion
			actualRespGetProjAfterDelete := getWorkspacePolicyForTest(router, 1, fmt.Sprintf("%s?scope=project&projectId=1111", urlPolicyWorkspace))

			testutils.Expect(*actualRespGetProjAfterDelete.Policy).ToEqual(api.WorkspacePolicyDefaultsAndRulesV2{})
			testutils.Expect(*actualRespGetProjAfterDelete.Effective).ToEqual(dao.SystemPolicy.Policy)

			//
			//   Part 3 - Create cluster policy
			//
			clusterId := backendMock.ClusterIdPerTenant["1"][0]
			bodyCluster := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_CLUSTER,
					ClusterId: &clusterId,
					Name:      "pol-c-1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Command: convutils.Ref("command-cluster"),
					},
				},
			}

			requestCluster, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyCluster))
			Expect(err).To(BeNil())
			requestCluster.Header = *headerForAdmin(1)

			responseCluster := httptest.NewRecorder()
			router.ServeHTTP(responseCluster, requestCluster)
			Expect(responseCluster.Code).To(Equal(http.StatusOK))

			//
			//   Part 4 - Verify project inherits cluster policy
			//
			actualRespGetProjAfterCluster := getWorkspacePolicyForTest(router, 1, fmt.Sprintf("%s?scope=project&projectId=1111", urlPolicyWorkspace))

			// Project policy should still be empty
			Expect(*actualRespGetProjAfterCluster.Policy).To(Equal(api.WorkspacePolicyDefaultsAndRulesV2{}))

			// But effective policy should inherit from cluster
			expectedEffective = api.WorkspacePolicyDefaultsAndRulesV2{
				Defaults: &api.WorkspacePolicyDefaultsV2{
					Command: convutils.Ref("command-cluster"),
				},
			}

			dao.InjectSystemPolicyForTesting(&expectedEffective)

			Expect(*actualRespGetProjAfterCluster.Effective).To(Equal(expectedEffective))
		})
	})

	Context("project policy inherits from cluster policy", func() {
		It("create cluster policy -> create project policy -> effective policy should inherit from cluster", func() {
			//
			//   Part 1 - Project policy with image default
			//
			bodyProj := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_PROJECT,
					ProjectId: convutils.Ref(int32(1111)),
					Name:      "pol-p-1111",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("image-proj"),
					},
				},
			}

			requestProj, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyProj))
			Expect(err).To(BeNil())
			requestProj.Header = *headerForAdmin(1)

			responseProj := httptest.NewRecorder()
			router.ServeHTTP(responseProj, requestProj)
			Expect(responseProj.Code).To(Equal(http.StatusOK))

			// Verify initial project policy
			actualRespGetProj := getWorkspacePolicyForTest(router, 1, fmt.Sprintf("%s?scope=project&projectId=1111", urlPolicyWorkspace))
			Expect(*actualRespGetProj.Policy).To(Equal(*bodyProj.Policy))

			var expectedEffective api.WorkspacePolicyDefaultsAndRulesV2
			copyErr := convutils.CopySpecs(&expectedEffective, bodyProj.Policy)
			Expect(copyErr).To(BeNil())

			dao.InjectSystemPolicyForTesting(&expectedEffective)

			Expect(*actualRespGetProj.Effective).To(Equal(expectedEffective))

			//
			//   Part 2 - Create cluster policy
			//
			clusterId := backendMock.ClusterIdPerTenant["1"][0]
			bodyCluster := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_CLUSTER,
					ClusterId: &clusterId,
					Name:      "pol-c-1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Command: convutils.Ref("command-cluster"),
					},
				},
			}

			requestCluster, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyCluster))
			Expect(err).To(BeNil())
			requestCluster.Header = *headerForAdmin(1)

			responseCluster := httptest.NewRecorder()
			router.ServeHTTP(responseCluster, requestCluster)
			Expect(responseCluster.Code).To(Equal(http.StatusOK))

			//
			//   Part 3 - Verify project inherits cluster policy
			//
			actualRespGetProjAfterCluster := getWorkspacePolicyForTest(router, 1, fmt.Sprintf("%s?scope=project&projectId=1111", urlPolicyWorkspace))

			// effective policy should inherit from cluster
			expectedEffective = api.WorkspacePolicyDefaultsAndRulesV2{
				Defaults: &api.WorkspacePolicyDefaultsV2{
					Image:   convutils.Ref("image-proj"),
					Command: convutils.Ref("command-cluster"),
				},
			}

			dao.InjectSystemPolicyForTesting(&expectedEffective)

			Expect(*actualRespGetProjAfterCluster.Effective).To(Equal(expectedEffective))
		})
	})

	Context("project policy inherits from department and recalculates after deletion", func() {
		It("create department policy -> create project policy -> delete project policy -> effective policy should inherit from department", func() {
			//
			//   Part 1 - Department policy with image default and runAsUid rule
			//
			bodyDept := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:        api.SCOPETYPE_DEPARTMENT,
					DepartmentId: convutils.Ref("111"),
					Name:         "pol-d-111",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("dep-image"),
					},
					Rules: &api.WorkspacePolicyRulesV2{
						Security: &api.SupersetRulesAllOfSecurity{
							RunAsUid: &api.IntegerRulesOptional{
								Min: convutils.Ref(int32(1000)),
							},
						},
					},
				},
			}

			requestDept, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyDept))
			Expect(err).To(BeNil())
			requestDept.Header = *headerForAdmin(1)

			responseDept := httptest.NewRecorder()
			router.ServeHTTP(responseDept, requestDept)
			Expect(responseDept.Code).To(Equal(http.StatusOK))

			//
			//   Part 2 - Project policy with command default and runAsGid rule
			//
			bodyProj := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_PROJECT,
					ProjectId: convutils.Ref(int32(1111)),
					Name:      "pol-p-1111",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Command: convutils.Ref("command-proj"),
					},
					Rules: &api.WorkspacePolicyRulesV2{
						Security: &api.SupersetRulesAllOfSecurity{
							RunAsGid: &api.IntegerRulesOptional{
								Min: convutils.Ref(int32(2000)),
							},
						},
					},
				},
			}

			requestProj, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyProj))
			Expect(err).To(BeNil())
			requestProj.Header = *headerForAdmin(1)

			responseProj := httptest.NewRecorder()
			router.ServeHTTP(responseProj, requestProj)
			Expect(responseProj.Code).To(Equal(http.StatusOK))

			// Verify project policy inherits department policy
			actualRespGetProj := getWorkspacePolicyForTest(router, 1, fmt.Sprintf("%s?scope=project&projectId=1111", urlPolicyWorkspace))

			expectedRespProj := api.WorkspacePolicyDefaultsAndRulesV2{
				Defaults: &api.WorkspacePolicyDefaultsV2{
					Image:   convutils.Ref("dep-image"),
					Command: convutils.Ref("command-proj"),
				},
				Rules: &api.WorkspacePolicyRulesV2{
					Security: &api.SupersetRulesAllOfSecurity{
						RunAsUid: &api.IntegerRulesOptional{
							SourceOfRule: &api.SourceOfRule{
								Scope:        api.SCOPETYPE_DEPARTMENT,
								DepartmentId: convutils.Ref("111"),
							},
							Min: convutils.Ref(int32(1000)),
						},
						RunAsGid: &api.IntegerRulesOptional{
							SourceOfRule: &api.SourceOfRule{
								Scope:     api.SCOPETYPE_PROJECT,
								ProjectId: convutils.Ref(int32(1111)),
							},
							Min: convutils.Ref(int32(2000)),
						},
					},
				},
			}

			dao.InjectSystemPolicyForTesting(&expectedRespProj)

			Expect(*actualRespGetProj.Effective).To(Equal(expectedRespProj))

			//
			//   Part 3 - Delete the project policy
			//
			deletePolicyForTest(router, 1, fmt.Sprintf("%s?scope=project&projectId=1111", urlPolicyWorkspace))

			//
			//   Part 4 - Get the project policy again - should only inherit from department
			//
			actualRespGetProjAfterDelete := getWorkspacePolicyForTest(router, 1, fmt.Sprintf("%s?scope=project&projectId=1111", urlPolicyWorkspace))

			expectedRespProjAfterDelete := api.WorkspacePolicyDefaultsAndRulesV2{
				Defaults: &api.WorkspacePolicyDefaultsV2{
					Image: convutils.Ref("dep-image"),
				},
				Rules: &api.WorkspacePolicyRulesV2{
					Security: &api.SupersetRulesAllOfSecurity{
						RunAsUid: &api.IntegerRulesOptional{
							SourceOfRule: &api.SourceOfRule{
								Scope:        api.SCOPETYPE_DEPARTMENT,
								DepartmentId: convutils.Ref("111"),
							},
							Min: convutils.Ref(int32(1000)),
						},
					},
				},
			}

			dao.InjectSystemPolicyForTesting(&expectedRespProjAfterDelete)

			Expect(*actualRespGetProjAfterDelete.Effective).To(Equal(expectedRespProjAfterDelete))
		})
	})

	Context("project policy inherits from cluster and recalculates after deletion", func() {
		It("create cluster policy -> create project policy -> delete project policy -> effective policy should inherit from cluster", func() {
			//
			//   Part 1 - Cluster policy with image default and runAsUid rule
			//
			clusterId := backendMock.ClusterIdPerTenant["1"][0]
			bodyCluster := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_CLUSTER,
					ClusterId: &clusterId,
					Name:      "pol-c-1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("cluster-image"),
					},
					Rules: &api.WorkspacePolicyRulesV2{
						Security: &api.SupersetRulesAllOfSecurity{
							RunAsUid: &api.IntegerRulesOptional{
								Min: convutils.Ref(int32(1000)),
							},
						},
					},
				},
			}

			requestCluster, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyCluster))
			Expect(err).To(BeNil())
			requestCluster.Header = *headerForAdmin(1)

			responseCluster := httptest.NewRecorder()
			router.ServeHTTP(responseCluster, requestCluster)
			Expect(responseCluster.Code).To(Equal(http.StatusOK))

			//
			//   Part 2 - Project policy with command default and runAsGid rule
			//
			bodyProj := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_PROJECT,
					ProjectId: convutils.Ref(int32(1111)),
					Name:      "pol-p-1111",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Command: convutils.Ref("command-proj"),
					},
					Rules: &api.WorkspacePolicyRulesV2{
						Security: &api.SupersetRulesAllOfSecurity{
							RunAsGid: &api.IntegerRulesOptional{
								Min: convutils.Ref(int32(2000)),
							},
						},
					},
				},
			}

			requestProj, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyProj))
			Expect(err).To(BeNil())
			requestProj.Header = *headerForAdmin(1)

			responseProj := httptest.NewRecorder()
			router.ServeHTTP(responseProj, requestProj)
			Expect(responseProj.Code).To(Equal(http.StatusOK))

			// Verify project policy inherits cluster policy
			actualRespGetProj := getWorkspacePolicyForTest(router, 1, fmt.Sprintf("%s?scope=project&projectId=1111", urlPolicyWorkspace))

			expectedRespProj := api.WorkspacePolicyDefaultsAndRulesV2{
				Defaults: &api.WorkspacePolicyDefaultsV2{
					Image:   convutils.Ref("cluster-image"),
					Command: convutils.Ref("command-proj"),
				},
				Rules: &api.WorkspacePolicyRulesV2{
					Security: &api.SupersetRulesAllOfSecurity{
						RunAsUid: &api.IntegerRulesOptional{
							SourceOfRule: &api.SourceOfRule{
								Scope:     api.SCOPETYPE_CLUSTER,
								ClusterId: &clusterId,
							},
							Min: convutils.Ref(int32(1000)),
						},
						RunAsGid: &api.IntegerRulesOptional{
							SourceOfRule: &api.SourceOfRule{
								Scope:     api.SCOPETYPE_PROJECT,
								ProjectId: convutils.Ref(int32(1111)),
							},
							Min: convutils.Ref(int32(2000)),
						},
					},
				},
			}

			dao.InjectSystemPolicyForTesting(&expectedRespProj)

			Expect(*actualRespGetProj.Effective).To(Equal(expectedRespProj))

			//
			//   Part 3 - Delete the project policy
			//
			deletePolicyForTest(router, 1, fmt.Sprintf("%s?scope=project&projectId=1111", urlPolicyWorkspace))

			//
			//   Part 4 - Get the project policy again - should only inherit from cluster
			//
			actualRespGetProjAfterDelete := getWorkspacePolicyForTest(router, 1, fmt.Sprintf("%s?scope=project&projectId=1111", urlPolicyWorkspace))

			expectedRespProjAfterDelete := api.WorkspacePolicyDefaultsAndRulesV2{
				Defaults: &api.WorkspacePolicyDefaultsV2{
					Image: convutils.Ref("cluster-image"),
				},
				Rules: &api.WorkspacePolicyRulesV2{
					Security: &api.SupersetRulesAllOfSecurity{
						RunAsUid: &api.IntegerRulesOptional{
							SourceOfRule: &api.SourceOfRule{
								Scope:     api.SCOPETYPE_CLUSTER,
								ClusterId: &clusterId,
							},
							Min: convutils.Ref(int32(1000)),
						},
					},
				},
			}

			dao.InjectSystemPolicyForTesting(&expectedRespProjAfterDelete)

			Expect(*actualRespGetProjAfterDelete.Effective).To(Equal(expectedRespProjAfterDelete))
		})
	})

	Context("department policy inherits from cluster and recalculates after deletion", func() {
		It("create cluster policy -> create department policy -> delete department policy -> effective policy should inherit from cluster", func() {
			//
			//   Part 1 - Cluster policy with image default and runAsUid rule
			//
			clusterId := backendMock.ClusterIdPerTenant["1"][0]
			bodyCluster := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_CLUSTER,
					ClusterId: &clusterId,
					Name:      "pol-c-1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("cluster-image"),
					},
					Rules: &api.WorkspacePolicyRulesV2{
						Security: &api.SupersetRulesAllOfSecurity{
							RunAsUid: &api.IntegerRulesOptional{
								Min: convutils.Ref(int32(1000)),
							},
						},
					},
				},
			}

			requestCluster, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyCluster))
			Expect(err).To(BeNil())
			requestCluster.Header = *headerForAdmin(1)

			responseCluster := httptest.NewRecorder()
			router.ServeHTTP(responseCluster, requestCluster)
			Expect(responseCluster.Code).To(Equal(http.StatusOK))

			//
			//   Part 2 - Department policy with command default and runAsGid rule
			//
			bodyDept := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:        api.SCOPETYPE_DEPARTMENT,
					DepartmentId: convutils.Ref("111"),
					Name:         "pol-d-111",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Command: convutils.Ref("command-dept"),
					},
					Rules: &api.WorkspacePolicyRulesV2{
						Security: &api.SupersetRulesAllOfSecurity{
							RunAsGid: &api.IntegerRulesOptional{
								Min: convutils.Ref(int32(2000)),
							},
						},
					},
				},
			}

			requestDept, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyDept))
			Expect(err).To(BeNil())
			requestDept.Header = *headerForAdmin(1)

			responseDept := httptest.NewRecorder()
			router.ServeHTTP(responseDept, requestDept)
			Expect(responseDept.Code).To(Equal(http.StatusOK))

			// Verify department policy inherits cluster policy
			actualRespGetDept := getWorkspacePolicyForTest(router, 1, fmt.Sprintf("%s?scope=department&departmentId=111", urlPolicyWorkspace))

			expectedRespDept := api.WorkspacePolicyDefaultsAndRulesV2{
				Defaults: &api.WorkspacePolicyDefaultsV2{
					Image:   convutils.Ref("cluster-image"),
					Command: convutils.Ref("command-dept"),
				},
				Rules: &api.WorkspacePolicyRulesV2{
					Security: &api.SupersetRulesAllOfSecurity{
						RunAsUid: &api.IntegerRulesOptional{
							SourceOfRule: &api.SourceOfRule{
								Scope:     api.SCOPETYPE_CLUSTER,
								ClusterId: &clusterId,
							},
							Min: convutils.Ref(int32(1000)),
						},
						RunAsGid: &api.IntegerRulesOptional{
							SourceOfRule: &api.SourceOfRule{
								Scope:        api.SCOPETYPE_DEPARTMENT,
								DepartmentId: convutils.Ref("111"),
							},
							Min: convutils.Ref(int32(2000)),
						},
					},
				},
			}

			dao.InjectSystemPolicyForTesting(&expectedRespDept)

			Expect(*actualRespGetDept.Effective).To(Equal(expectedRespDept))

			//
			//   Part 3 - Delete the department policy
			//
			deletePolicyForTest(router, 1, fmt.Sprintf("%s?scope=department&departmentId=111", urlPolicyWorkspace))

			//
			//   Part 4 - Get the department policy again - should only inherit from cluster
			//
			actualRespGetDeptAfterDelete := getWorkspacePolicyForTest(router, 1, fmt.Sprintf("%s?scope=department&departmentId=111", urlPolicyWorkspace))

			expectedRespDeptAfterDelete := api.WorkspacePolicyDefaultsAndRulesV2{
				Defaults: &api.WorkspacePolicyDefaultsV2{
					Image: convutils.Ref("cluster-image"),
				},
				Rules: &api.WorkspacePolicyRulesV2{
					Security: &api.SupersetRulesAllOfSecurity{
						RunAsUid: &api.IntegerRulesOptional{
							SourceOfRule: &api.SourceOfRule{
								Scope:     api.SCOPETYPE_CLUSTER,
								ClusterId: &clusterId,
							},
							Min: convutils.Ref(int32(1000)),
						},
					},
				},
			}

			dao.InjectSystemPolicyForTesting(&expectedRespDeptAfterDelete)

			Expect(*actualRespGetDeptAfterDelete.Effective).To(Equal(expectedRespDeptAfterDelete))
		})
	})
	Context("orphaned department policy deletion (department deleted from backend)", func() {
		It("should successfully delete a department policy when the department no longer exists", func() {
			ctx := tenantContextForTest(1)

			bodyDept := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:        api.SCOPETYPE_DEPARTMENT,
					DepartmentId: convutils.Ref("111"),
					Name:         "orphaned-dept-policy",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("dept-image"),
					},
				},
			}

			requestCreate, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyDept))
			Expect(err).To(BeNil())
			requestCreate.Header = *headerForAdmin(1)

			responseCreate := httptest.NewRecorder()
			router.ServeHTTP(responseCreate, requestCreate.WithContext(ctx))
			Expect(responseCreate.Code).To(Equal(http.StatusOK))

			backendMock.MarkScopeEntityDeleted(scope.Department, "111")

			requestDelete, err := http.NewRequest("DELETE",
				fmt.Sprintf("%s?scope=%s&departmentId=%s", urlPolicyWorkspace, api.SCOPETYPE_DEPARTMENT, "111"), nil)
			Expect(err).To(BeNil())
			requestDelete.Header = *headerForAdmin(1)

			responseDelete := httptest.NewRecorder()
			router.ServeHTTP(responseDelete, requestDelete.WithContext(ctx))
			Expect(responseDelete.Code).To(Equal(http.StatusNoContent))
		})

		It("should return 404 when trying to delete an already-deleted orphaned policy twice", func() {
			ctx := tenantContextForTest(1)

			bodyDept := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:        api.SCOPETYPE_DEPARTMENT,
					DepartmentId: convutils.Ref("111"),
					Name:         "orphaned-dept-policy-2",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("dept-image"),
					},
				},
			}

			requestCreate, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyDept))
			Expect(err).To(BeNil())
			requestCreate.Header = *headerForAdmin(1)

			responseCreate := httptest.NewRecorder()
			router.ServeHTTP(responseCreate, requestCreate.WithContext(ctx))
			Expect(responseCreate.Code).To(Equal(http.StatusOK))

			backendMock.MarkScopeEntityDeleted(scope.Department, "111")

			requestDelete, err := http.NewRequest("DELETE",
				fmt.Sprintf("%s?scope=%s&departmentId=%s", urlPolicyWorkspace, api.SCOPETYPE_DEPARTMENT, "111"), nil)
			Expect(err).To(BeNil())
			requestDelete.Header = *headerForAdmin(1)

			responseDelete := httptest.NewRecorder()
			router.ServeHTTP(responseDelete, requestDelete.WithContext(ctx))
			Expect(responseDelete.Code).To(Equal(http.StatusNoContent))

			requestDelete2, err := http.NewRequest("DELETE",
				fmt.Sprintf("%s?scope=%s&departmentId=%s", urlPolicyWorkspace, api.SCOPETYPE_DEPARTMENT, "111"), nil)
			Expect(err).To(BeNil())
			requestDelete2.Header = *headerForAdmin(1)

			responseDelete2 := httptest.NewRecorder()
			router.ServeHTTP(responseDelete2, requestDelete2.WithContext(ctx))
			Expect(responseDelete2.Code).To(Equal(http.StatusNotFound))
		})

		It("should successfully delete a project policy when the project no longer exists", func() {
			ctx := tenantContextForTest(1)
			cluster1 := backendMock.ClusterIdPerTenant["1"][0]

			injectPolicy(ctx, router, api.SCOPETYPE_PROJECT, api.POLICYTYPE_WORKSPACE, 1, cluster1, 1111, "orphaned-proj-policy")

			backendMock.MarkScopeEntityDeleted(scope.Project, "1111")

			requestDelete, err := http.NewRequest("DELETE",
				fmt.Sprintf("%s?scope=%s&projectId=%d", urlPolicyWorkspace, api.SCOPETYPE_PROJECT, 1111), nil)
			Expect(err).To(BeNil())
			requestDelete.Header = *headerForAdmin(1)

			responseDelete := httptest.NewRecorder()
			router.ServeHTTP(responseDelete, requestDelete.WithContext(ctx))
			Expect(responseDelete.Code).To(Equal(http.StatusNoContent))
		})

		It("should successfully delete orphaned policies across multiple workload types", func() {
			ctx := tenantContextForTest(1)
			cluster1 := backendMock.ClusterIdPerTenant["1"][0]

			bodyDept := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:        api.SCOPETYPE_DEPARTMENT,
					DepartmentId: convutils.Ref("111"),
					Name:         "orphaned-ws",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("dept-image"),
					},
				},
			}
			requestCreateDept, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyDept))
			Expect(err).To(BeNil())
			requestCreateDept.Header = *headerForAdmin(1)
			responseCreateDept := httptest.NewRecorder()
			router.ServeHTTP(responseCreateDept, requestCreateDept.WithContext(ctx))
			Expect(responseCreateDept.Code).To(Equal(http.StatusOK))

			injectPolicy(ctx, router, api.SCOPETYPE_PROJECT, api.POLICYTYPE_TRAINING, 1, cluster1, 1111, "orphaned-train")

			backendMock.MarkScopeEntityDeleted(scope.Department, "111")
			backendMock.MarkScopeEntityDeleted(scope.Project, "1111")

			requestDeleteDept, err := http.NewRequest("DELETE",
				fmt.Sprintf("%s?scope=%s&departmentId=%s", urlPolicyWorkspace, api.SCOPETYPE_DEPARTMENT, "111"), nil)
			Expect(err).To(BeNil())
			requestDeleteDept.Header = *headerForAdmin(1)

			responseDeleteDept := httptest.NewRecorder()
			router.ServeHTTP(responseDeleteDept, requestDeleteDept.WithContext(ctx))
			Expect(responseDeleteDept.Code).To(Equal(http.StatusNoContent))

			requestDeleteProj, err := http.NewRequest("DELETE",
				fmt.Sprintf("%s?scope=%s&projectId=%d", urlPolicyTraining, api.SCOPETYPE_PROJECT, 1111), nil)
			Expect(err).To(BeNil())
			requestDeleteProj.Header = *headerForAdmin(1)

			responseDeleteProj := httptest.NewRecorder()
			router.ServeHTTP(responseDeleteProj, requestDeleteProj.WithContext(ctx))
			Expect(responseDeleteProj.Code).To(Equal(http.StatusNoContent))
		})
	})

	Context("project policy created first, then department policy created and deleted", func() {
		It("create project policy -> create department policy -> delete department policy -> project effective should be recalculated", func() {
			//
			//   Part 1 - Project policy with image default and runAsGid rule
			//
			bodyProj := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_PROJECT,
					ProjectId: convutils.Ref(int32(1111)),
					Name:      "pol-p-1111",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("image-proj"),
					},
					Rules: &api.WorkspacePolicyRulesV2{
						Security: &api.SupersetRulesAllOfSecurity{
							RunAsGid: &api.IntegerRulesOptional{
								Min: convutils.Ref(int32(2000)),
							},
						},
					},
				},
			}

			requestProj, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyProj))
			Expect(err).To(BeNil())
			requestProj.Header = *headerForAdmin(1)

			responseProj := httptest.NewRecorder()
			router.ServeHTTP(responseProj, requestProj)
			Expect(responseProj.Code).To(Equal(http.StatusOK))

			//
			//   Part 2 - Department policy with command default and runAsUid rule
			//
			bodyDept := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:        api.SCOPETYPE_DEPARTMENT,
					DepartmentId: convutils.Ref("111"),
					Name:         "pol-d-111",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Command: convutils.Ref("command-dept"),
					},
					Rules: &api.WorkspacePolicyRulesV2{
						Security: &api.SupersetRulesAllOfSecurity{
							RunAsUid: &api.IntegerRulesOptional{
								Min: convutils.Ref(int32(1000)),
							},
						},
					},
				},
			}

			requestDept, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyDept))
			Expect(err).To(BeNil())
			requestDept.Header = *headerForAdmin(1)

			responseDept := httptest.NewRecorder()
			router.ServeHTTP(responseDept, requestDept)
			Expect(responseDept.Code).To(Equal(http.StatusOK))

			// Verify project policy now inherits from department
			actualRespGetProjAfterDept := getWorkspacePolicyForTest(router, 1, fmt.Sprintf("%s?scope=project&projectId=1111", urlPolicyWorkspace))

			expectedRespProjAfterDept := api.WorkspacePolicyDefaultsAndRulesV2{
				Defaults: &api.WorkspacePolicyDefaultsV2{
					Image:   convutils.Ref("image-proj"),
					Command: convutils.Ref("command-dept"),
				},
				Rules: &api.WorkspacePolicyRulesV2{
					Security: &api.SupersetRulesAllOfSecurity{
						RunAsUid: &api.IntegerRulesOptional{
							SourceOfRule: &api.SourceOfRule{
								Scope:        api.SCOPETYPE_DEPARTMENT,
								DepartmentId: convutils.Ref("111"),
							},
							Min: convutils.Ref(int32(1000)),
						},
						RunAsGid: &api.IntegerRulesOptional{
							SourceOfRule: &api.SourceOfRule{
								Scope:     api.SCOPETYPE_PROJECT,
								ProjectId: convutils.Ref(int32(1111)),
							},
							Min: convutils.Ref(int32(2000)),
						},
					},
				},
			}

			dao.InjectSystemPolicyForTesting(&expectedRespProjAfterDept)
			Expect(*actualRespGetProjAfterDept.Effective).To(Equal(expectedRespProjAfterDept))

			//
			//   Part 3 - Delete the department policy
			//
			deletePolicyForTest(router, 1, fmt.Sprintf("%s?scope=department&departmentId=111", urlPolicyWorkspace))

			//
			//   Part 4 - Get the project policy again - should only have its own defaults and rules
			//
			actualRespGetProjAfterDelete := getWorkspacePolicyForTest(router, 1, fmt.Sprintf("%s?scope=project&projectId=1111", urlPolicyWorkspace))

			// Project policy should still have its own defaults and rules
			Expect(*actualRespGetProjAfterDelete.Policy).To(Equal(*bodyProj.Policy))

			// Effective policy should only have project's own defaults and rules (no inheritance)
			expectedRespProjAfterDelete := api.WorkspacePolicyDefaultsAndRulesV2{
				Defaults: &api.WorkspacePolicyDefaultsV2{
					Image: convutils.Ref("image-proj"),
				},
				Rules: &api.WorkspacePolicyRulesV2{
					Security: &api.SupersetRulesAllOfSecurity{
						RunAsGid: &api.IntegerRulesOptional{
							SourceOfRule: &api.SourceOfRule{
								Scope:     api.SCOPETYPE_PROJECT,
								ProjectId: convutils.Ref(int32(1111)),
							},
							Min: convutils.Ref(int32(2000)),
						},
					},
				},
			}

			dao.InjectSystemPolicyForTesting(&expectedRespProjAfterDelete)

			Expect(*actualRespGetProjAfterDelete.Effective).To(Equal(expectedRespProjAfterDelete))
		})
	})
})

type genPolicyBody func(scopeType api.ScopeType, tenantId int32, clusterId string, projectId int32, name string) interface{}

func genWorkspacePolicyBody(scopeType api.ScopeType, _ int32, clusterId string, projectId int32, name string) interface{} {
	result := api.WorkspacePolicyChangeRequestV2{
		Meta: api.PolicyCreationFields{
			Scope: scopeType,
			Name:  name,
		},
		Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
			Defaults: &api.WorkspacePolicyDefaultsV2{
				Image: convutils.Ref("image-proj"),
			},
		},
	}
	if clusterId != "" {
		result.Meta.ClusterId = &clusterId
	}
	if projectId != 0 {
		result.Meta.ProjectId = &projectId
	}
	return &result
}

func genTrainingPolicyBody(scopeType api.ScopeType, _ int32, clusterId string, projectId int32, name string) interface{} {
	result := api.TrainingPolicyChangeRequestV2{
		Meta: api.PolicyCreationFields{
			Scope: scopeType,
			Name:  name,
		},
		Policy: &api.TrainingPolicyDefaultsAndRulesV2{
			Defaults: &api.TrainingPolicyDefaultsV2{
				Image: convutils.Ref("image-proj"),
			},
		},
	}
	if clusterId != "" {
		result.Meta.ClusterId = &clusterId
	}
	if projectId != 0 {
		result.Meta.ProjectId = &projectId
	}
	return &result
}

func genDistributedInferencePolicyBody(scopeType api.ScopeType, _ int32, clusterId string, projectId int32, name string) interface{} {
	result := api.DistributedInferencePolicyChangeRequestV2{
		Meta: api.PolicyCreationFields{
			Scope: scopeType,
			Name:  name,
		},
		Policy: &api.DistributedInferencePolicyDefaultsAndRulesV2{
			Defaults: &api.DistributedInferenceDefaultsV2{
				Leader: &api.DistributedInferenceLeaderWorkerDefaultsV2{
					Image: convutils.Ref("image-proj"),
				},
			},
		},
	}
	if clusterId != "" {
		result.Meta.ClusterId = &clusterId
	}
	if projectId != 0 {
		result.Meta.ProjectId = &projectId
	}
	return &result
}

func genInferencePolicyBody(scopeType api.ScopeType, _ int32, clusterId string, projectId int32, name string) interface{} {
	result := api.InferencePolicyChangeRequestV2{
		Meta: api.PolicyCreationFields{
			Scope: scopeType,
			Name:  name,
		},
		Policy: &api.InferencePolicyDefaultsAndRulesV2{
			Defaults: &api.InferencePolicyDefaultsV2{
				Image: convutils.Ref("image-proj"),
			},
		},
	}
	if clusterId != "" {
		result.Meta.ClusterId = &clusterId
	}
	if projectId != 0 {
		result.Meta.ProjectId = &projectId
	}
	return &result
}

func genDistributedPolicyBody(scopeType api.ScopeType, _ int32, clusterId string, projectId int32, name string) interface{} {
	result := api.DistributedPolicyChangeRequestV2{
		Meta: api.PolicyCreationFields{
			Scope: scopeType,
			Name:  name,
		},

		Policy: &api.DistributedPolicyDefaultsAndRulesV2{
			Defaults: &api.DistributedPolicyDefaultsV2{
				Master: &api.ReplicaDefaultsV2{
					Image: convutils.Ref("image-proj"),
				},
				Worker: &api.DistributedPolicyDefaultsV2Worker{
					Image: convutils.Ref("image-proj"),
				},
			},
		},
	}
	if clusterId != "" {
		result.Meta.ClusterId = &clusterId
	}
	if projectId != 0 {
		result.Meta.ProjectId = &projectId
	}
	return &result
}

func genFallbackPolicyBody(scopeType api.ScopeType, _ int32, clusterId string, projectId int32, name string) interface{} {
	result := api.FallbackPolicyChangeRequest{
		Meta: api.PolicyCreationFields{
			Scope: scopeType,
			Name:  name,
		},
		Policy: &api.FallbackPolicyDefaultsAndRules{
			Defaults: &api.SupersetDefaults{
				Image: convutils.Ref("image-proj"),
			},
		},
	}
	if clusterId != "" {
		result.Meta.ClusterId = &clusterId
	}
	if projectId != 0 {
		result.Meta.ProjectId = &projectId
	}
	return result
}

var genPolicyBodyMap = map[api.PolicyType]genPolicyBody{
	api.POLICYTYPE_WORKSPACE:             genWorkspacePolicyBody,
	api.POLICYTYPE_TRAINING:              genTrainingPolicyBody,
	api.POLICYTYPE_INFERENCE:             genInferencePolicyBody,
	api.POLICYTYPE_DISTRIBUTED_INFERENCE: genDistributedInferencePolicyBody,
	api.POLICYTYPE_DISTRIBUTED:           genDistributedPolicyBody,
	api.POLICYTYPE_FALLBACK:              genFallbackPolicyBody,
}

var pathUrlMap = map[api.PolicyType]string{
	api.POLICYTYPE_WORKSPACE:             urlPolicyWorkspace,
	api.POLICYTYPE_TRAINING:              urlPolicyTraining,
	api.POLICYTYPE_INFERENCE:             urlPolicyInference,
	api.POLICYTYPE_DISTRIBUTED_INFERENCE: urlPolicyDistributedInference,
	api.POLICYTYPE_DISTRIBUTED:           urlPolicyDistributed,
	api.POLICYTYPE_FALLBACK:              urlPolicyFallback,
}

func injectPolicy(ctx context.Context, router chi.Router, scopeType api.ScopeType, policyType api.PolicyType, tenantId int32, clusterId string, projectId int32, name string) {
	body := genPolicyBodyMap[policyType](scopeType, tenantId, clusterId, projectId, name)

	request, err := http.NewRequest(http.MethodPatch, pathUrlMap[policyType], makePayload(body))
	Expect(err).To(BeNil())
	request.Header = *headerForAdmin(tenantId)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request.WithContext(ctx))
	Expect(response.Code).To(Equal(http.StatusOK))
}

func listPolicies(ctx context.Context) ([]dbmodel.Policy, error) {
	records := make([]dbmodel.Policy, 0)

	tenantID, _ := ctx.Value(runaictx.CfTenantId).(string)
	if _, err := runaidb.DB().WithContext(ctx).
		Table(dbtables.PoliciesTable).
		Where(fmt.Sprintf("%s = ?", dbtables.PoliciesColumnTenantId), tenantID).
		Scan(&records); err != nil {
		return nil, err
	}
	return records, nil
}
