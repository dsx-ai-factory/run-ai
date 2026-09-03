/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"fmt"
	"math"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	authorization_config "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/test"
	cluster_service_client "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
)

const (
	TestGroupId       = "some-group"
	TestLocalUser     = "some-local-user"
	TestAppId         = "some-app"
	TestProject2Id    = "2"
	TestProject3Id    = "3"
	TestProject4Id    = "4"
	TestDepartment2Id = "2"
	TestDepartment3Id = "3"
	TestDepartment4Id = "4"
)

var (
	TestCluster1Id = "a418ed33-9399-48c0-a890-122cadd13c48"
	TestCluster2Id = "122cadd1-9399-48c0-a890-3bfda418ed33"
)

type GetAccessRulesQueryParamsTest struct {
	limit          int
	offset         int
	lastUpdatedKey int
	includeDeleted bool
	clusterId      string

	expectedCount           *int
	expectedCountBiggerThan *int
}

var _ = Describe("GET Access rules API tests", Ordered, func() {
	tenantId := fmt.Sprintf("%d", testing.TestTenant)
	var createdRoleId int32 = 0
	ctx := setContext(tenantId, TestUserId, []string{dao.SuperAdmins})

	tenantSettings := NewMockTenantSettingsService(map[string]bool{})

	clustersClient := cluster_service_client.NewAPIClient(&cluster_service_client.Configuration{
		Servers: cluster_service_client.ServerConfigurations{
			{
				URL: config.Get[*authorization_config.Server]().ClusterServiceUrl,
			},
		},
	})

	rolesApiService := NewRolesApiService(clustersClient, tenantSettings)
	accessRulesApiService := NewAccessRulesApiService(tenantSettings)
	createdRoles := []api.Role{}

	// create a role to be used by tests
	BeforeAll(func() {
		TestsCleanupAndInit()
		roleBody := api.RoleCreationFieldsInternal{
			Name:        "test role 3",
			Description: "can partially manage projects",
			Permissions: []api.PostPermission{
				{
					ResourceType: api.RESOURCETYPE_PROJECT,
					Actions: []api.Action{
						api.ACTION_CREATE,
						api.ACTION_READ,
						api.ACTION_UPDATE,
					},
				},
			},
		}
		createdRole := createTestRole(ctx, rolesApiService, roleBody, &createdRoles, false, http.StatusCreated)
		createdRoleId = createdRole.Id
	})

	AfterAll(func() {
		for _, role := range createdRoles {
			deleteRole, err := rolesApiService.DeleteRole(ctx, role.Id)
			Expect(err).To(BeNil())
			Expect(deleteRole.Code).To(Equal(http.StatusNoContent))
		}
		TestsCleanupAndInit()
	})

	Context("GetAccessRules query params", func() {
		const (
			lastUpdatedKey1            = 1
			lastUpdatedKey2            = 2
			lastUpdatedKey3            = 3
			lastUpdatedKey4            = 4
			lastUpdatedKey5            = 5
			lastUpdatedKeyWithDeleted6 = 6

			lastUpdated1Count = 5
		)

		var (
			existingAccessRules           = []api.AccessRule{}
			allAccessRulesDeletedIncluded = []api.AccessRule{}
			deletedAccessRules            = []api.AccessRule{}

			cluster1AccessRules = []api.AccessRule{}
			cluster2AccessRules = []api.AccessRule{}
			tenantAccessRules   = []api.AccessRule{}
			tmpAccessRules      = []api.AccessRule{}

			lastUpdatedMap = map[int]string{}
		)

		// get the existing super admin access rule
		BeforeAll(func() {
			// Create AccessRules for this test
			// 12 AccessRules to create
			accessRulesInCluster1ToCreate := []api.AccessRuleCreationFields{
				{
					ScopeId:     ProjectId,
					ScopeType:   api.SCOPETYPE_PROJECT,
					SubjectId:   TestUserId,
					SubjectType: api.SUBJECTTYPE_USER,
					RoleId:      createdRoleId,
				},
				{
					ScopeId:     ProjectId,
					ScopeType:   api.SCOPETYPE_PROJECT,
					SubjectId:   TestGroupId,
					SubjectType: api.SUBJECTTYPE_GROUP,
					RoleId:      createdRoleId,
				},
				{
					ScopeId:     TestProject2Id,
					ScopeType:   api.SCOPETYPE_PROJECT,
					SubjectId:   TestGroupId,
					SubjectType: api.SUBJECTTYPE_GROUP,
					RoleId:      createdRoleId,
				},
				{
					ScopeId:     DepartmentId,
					ScopeType:   api.SCOPETYPE_DEPARTMENT,
					SubjectId:   TestUserId,
					SubjectType: api.SUBJECTTYPE_USER,
					RoleId:      createdRoleId,
				},
				{
					ScopeId:     TestDepartment2Id,
					ScopeType:   api.SCOPETYPE_DEPARTMENT,
					SubjectId:   TestUserId,
					SubjectType: api.SUBJECTTYPE_USER,
					RoleId:      createdRoleId,
				},
				{
					ScopeId:     TestCluster1Id,
					ScopeType:   api.SCOPETYPE_CLUSTER,
					SubjectId:   TestUserId,
					SubjectType: api.SUBJECTTYPE_USER,
					RoleId:      createdRoleId,
				},
			}
			accessRulesInCluster2ToCreate := []api.AccessRuleCreationFields{
				{
					ScopeId:     TestProject3Id,
					ScopeType:   api.SCOPETYPE_PROJECT,
					SubjectId:   TestGroupId,
					SubjectType: api.SUBJECTTYPE_GROUP,
					RoleId:      createdRoleId,
				},
				{
					ScopeId:     TestProject4Id,
					ScopeType:   api.SCOPETYPE_PROJECT,
					SubjectId:   TestGroupId,
					SubjectType: api.SUBJECTTYPE_GROUP,
					RoleId:      createdRoleId,
				},
				{
					ScopeId:     TestDepartment3Id,
					ScopeType:   api.SCOPETYPE_DEPARTMENT,
					SubjectId:   TestGroupId,
					SubjectType: api.SUBJECTTYPE_GROUP,
					RoleId:      createdRoleId,
				},
				{
					ScopeId:     TestDepartment4Id,
					ScopeType:   api.SCOPETYPE_DEPARTMENT,
					SubjectId:   TestGroupId,
					SubjectType: api.SUBJECTTYPE_GROUP,
					RoleId:      createdRoleId,
				},
				{
					ScopeId:     TestCluster2Id,
					ScopeType:   api.SCOPETYPE_CLUSTER,
					SubjectId:   TestGroupId,
					SubjectType: api.SUBJECTTYPE_GROUP,
					RoleId:      createdRoleId,
				},
			}
			accessRulesForTenantToCreate := []api.AccessRuleCreationFields{
				{
					ScopeId:     tenantId,
					ScopeType:   api.SCOPETYPE_TENANT,
					SubjectId:   TestGroupId,
					SubjectType: api.SUBJECTTYPE_GROUP,
					RoleId:      createdRoleId,
				},
			}

			SetAncestorsCache(ctx, ProjectId, api.RESOURCETYPE_PROJECT, &authorization_utils.Ancestors{
				TenantId:    tenantId,
				ClusterUuid: &TestCluster1Id,
			})
			SetAncestorsCache(ctx, TestProject2Id, api.RESOURCETYPE_PROJECT, &authorization_utils.Ancestors{
				TenantId:    tenantId,
				ClusterUuid: &TestCluster1Id,
			})
			SetAncestorsCache(ctx, TestProject3Id, api.RESOURCETYPE_PROJECT, &authorization_utils.Ancestors{
				TenantId:    tenantId,
				ClusterUuid: &TestCluster2Id,
			})
			SetAncestorsCache(ctx, TestProject4Id, api.RESOURCETYPE_PROJECT, &authorization_utils.Ancestors{
				TenantId:    tenantId,
				ClusterUuid: &TestCluster2Id,
			})
			SetAncestorsCache(ctx, DepartmentId, api.RESOURCETYPE_DEPARTMENT, &authorization_utils.Ancestors{
				TenantId:    tenantId,
				ClusterUuid: &TestCluster1Id,
			})
			SetAncestorsCache(ctx, TestDepartment2Id, api.RESOURCETYPE_DEPARTMENT, &authorization_utils.Ancestors{
				TenantId:    tenantId,
				ClusterUuid: &TestCluster1Id,
			})
			SetAncestorsCache(ctx, TestDepartment3Id, api.RESOURCETYPE_DEPARTMENT, &authorization_utils.Ancestors{
				TenantId:    tenantId,
				ClusterUuid: &TestCluster2Id,
			})
			SetAncestorsCache(ctx, TestDepartment4Id, api.RESOURCETYPE_DEPARTMENT, &authorization_utils.Ancestors{
				TenantId:    tenantId,
				ClusterUuid: &TestCluster2Id,
			})
			SetAncestorsCache(ctx, TestCluster1Id, api.RESOURCETYPE_CLUSTER, &authorization_utils.Ancestors{
				TenantId: tenantId,
			})
			SetAncestorsCache(ctx, TestCluster2Id, api.RESOURCETYPE_CLUSTER, &authorization_utils.Ancestors{
				TenantId: tenantId,
			})
			SetAncestorsCache(ctx, ClusterId, api.RESOURCETYPE_CLUSTER, &authorization_utils.Ancestors{
				TenantId: tenantId,
			})

			createAccessRulesFn := func(accessRulesToCreate []api.AccessRuleCreationFields) []api.AccessRule {
				created := []api.AccessRule{}
				for _, accessRuleToCreate := range accessRulesToCreate {
					time.Sleep(10 * time.Millisecond)
					createTestAccessRule(ctx, accessRulesApiService, accessRuleToCreate, &created, false, http.StatusCreated)
				}
				return created
			}

			cluster1AccessRules = createAccessRulesFn(accessRulesInCluster1ToCreate)
			cluster2AccessRules = createAccessRulesFn(accessRulesInCluster2ToCreate)
			tenantAccessRules = createAccessRulesFn(accessRulesForTenantToCreate)

			// create and delete 2 accessrules
			tempAccessRulesCreated := []api.AccessRuleCreationFields{
				{
					ScopeId:     ProjectId,
					ScopeType:   api.SCOPETYPE_PROJECT,
					SubjectId:   TestLocalUser,
					SubjectType: api.SUBJECTTYPE_USER,
					RoleId:      createdRoleId,
				},
				{
					ScopeId:     TestProject2Id,
					ScopeType:   api.SCOPETYPE_PROJECT,
					SubjectId:   TestLocalUser,
					SubjectType: api.SUBJECTTYPE_USER,
					RoleId:      createdRoleId,
				},
			}
			tmpAccessRules = []api.AccessRule{}
			for _, accessRuleToCreate := range tempAccessRulesCreated {
				time.Sleep(10 * time.Millisecond)
				created := createTestAccessRule(ctx, accessRulesApiService, accessRuleToCreate, &tmpAccessRules, false, http.StatusCreated)
				time.Sleep(10 * time.Millisecond)
				_, err := accessRulesApiService.DeleteAccessRule(ctx, created.Id)
				Expect(err).To(BeNil())
			}

			// now get all the access rules deleted included - add max limit to make sure its sorted
			getAccessRulesResponse, err := accessRulesApiService.GetAccessRules(ctx, "", []string{},
				math.MaxInt32, 0, "", true, "", "", "", "", nil, "")
			Expect(err).To(BeNil())
			Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))

			response := httpResponseBodyToAccessRules(getAccessRulesResponse)
			allAccessRulesDeletedIncluded = response.AccessRules
			for _, accessRule := range allAccessRulesDeletedIncluded {
				if accessRule.DeletedAt == nil {
					existingAccessRules = append(existingAccessRules, accessRule)
				} else {
					deletedAccessRules = append(deletedAccessRules, accessRule)
				}
			}

			Expect(len(allAccessRulesDeletedIncluded)).To(Equal(len(existingAccessRules) + len(deletedAccessRules)))

			// we cannot use vars created in "BeforeAll" to initialize Entries - so we do it like this
			lastUpdatedMap[lastUpdatedKey1] = existingAccessRules[len(existingAccessRules)-lastUpdated1Count].UpdatedAt.Format(time.RFC3339Nano)
			lastUpdatedMap[lastUpdatedKey2] = existingAccessRules[1].UpdatedAt.Format(time.RFC3339Nano)
			lastUpdatedMap[lastUpdatedKey3] = existingAccessRules[len(existingAccessRules)-1].UpdatedAt.Format(time.RFC3339Nano)
			lastUpdatedMap[lastUpdatedKey4] = existingAccessRules[len(existingAccessRules)-7].UpdatedAt.Format(time.RFC3339Nano)
			lastUpdatedMap[lastUpdatedKey5] = existingAccessRules[len(existingAccessRules)-10].UpdatedAt.Format(time.RFC3339Nano)
			lastUpdatedMap[lastUpdatedKeyWithDeleted6] = allAccessRulesDeletedIncluded[len(allAccessRulesDeletedIncluded)-1].DeletedAt.Format(time.RFC3339Nano)
		})

		AfterAll(func() {
			CleanupAccessRules(accessRulesApiService, ctx, cluster1AccessRules)
			CleanupAccessRules(accessRulesApiService, ctx, cluster2AccessRules)
			CleanupAccessRules(accessRulesApiService, ctx, tenantAccessRules)
			CleanupAccessRules(accessRulesApiService, ctx, tmpAccessRules)
			existingAccessRules = []api.AccessRule{}

			allAccessRulesDeletedIncluded = []api.AccessRule{}
		})

		entries := []TableEntry{
			Entry("sanity offset limit", GetAccessRulesQueryParamsTest{
				offset:        0,
				limit:         2,
				expectedCount: ptr.To(2),
			}),
			Entry("offset bigger than limit", GetAccessRulesQueryParamsTest{
				offset:        5,
				limit:         4,
				expectedCount: ptr.To(4),
			}),
			Entry("very big limit", GetAccessRulesQueryParamsTest{
				offset:        1,
				limit:         50,
				expectedCount: ptr.To(12),
			}),
			Entry("offset+limit bigger than records amount", GetAccessRulesQueryParamsTest{
				offset:        9,
				limit:         10,
				expectedCount: ptr.To(4),
			}),
			Entry("very big offset", GetAccessRulesQueryParamsTest{
				offset:        50,
				limit:         0,
				expectedCount: ptr.To(0),
			}),
			Entry("sanity last updated", GetAccessRulesQueryParamsTest{
				lastUpdatedKey: lastUpdatedKey1,
				expectedCount:  ptr.To(lastUpdated1Count),
			}),
			Entry("last updated with basic offset and limit", GetAccessRulesQueryParamsTest{
				offset:         0,
				limit:          2,
				lastUpdatedKey: lastUpdatedKey2,
				expectedCount:  ptr.To(2),
			}),
			Entry("sanity include deleted", GetAccessRulesQueryParamsTest{
				includeDeleted: true,
				expectedCount:  nil,
			}),
			Entry("include deleted with offset and limit", GetAccessRulesQueryParamsTest{
				offset:                  9,
				limit:                   10,
				includeDeleted:          true,
				expectedCountBiggerThan: ptr.To(4),
			}),
			Entry("include deleted with last updated", GetAccessRulesQueryParamsTest{
				lastUpdatedKey: lastUpdatedKey3,
				includeDeleted: true,
				expectedCount:  ptr.To(3),
			}),
			Entry("last updated without include deleted", GetAccessRulesQueryParamsTest{
				lastUpdatedKey: lastUpdatedKeyWithDeleted6,
				expectedCount:  ptr.To(0),
			}),
			Entry("last updated with include deleted", GetAccessRulesQueryParamsTest{
				lastUpdatedKey: lastUpdatedKeyWithDeleted6,
				includeDeleted: true,
				expectedCount:  ptr.To(1),
			}),
			Entry("sanity cluster ID", GetAccessRulesQueryParamsTest{
				clusterId:     TestCluster1Id,
				expectedCount: ptr.To(8),
			}),
			Entry("sanity cluster ID 2", GetAccessRulesQueryParamsTest{
				clusterId:     TestCluster2Id,
				expectedCount: ptr.To(7),
			}),
			Entry("clusterId + lastUpdated + offset + limit", GetAccessRulesQueryParamsTest{
				clusterId:      TestCluster2Id,
				lastUpdatedKey: lastUpdatedKey4,
				offset:         1,
				limit:          3,
				expectedCount:  ptr.To(3),
			}),
			Entry("allllll params!", GetAccessRulesQueryParamsTest{
				clusterId:      TestCluster1Id,
				lastUpdatedKey: lastUpdatedKey5,
				includeDeleted: true,
				offset:         2,
				limit:          4,
				expectedCount:  ptr.To(4),
			}),
		}

		validateGetAccessRulesQueryParamsFn := func(params GetAccessRulesQueryParamsTest) {
			// we cannot use vars created in "BeforeAll" to initialize Entries - so we do it like this
			lastUpdated := ""
			if params.lastUpdatedKey > 0 {
				lastUpdated = lastUpdatedMap[params.lastUpdatedKey]
			}

			getAccessRulesResponse, err := accessRulesApiService.GetAccessRules(ctx, "", []string{},
				int32(params.limit), int32(params.offset), lastUpdated, params.includeDeleted, params.clusterId, "", "", "", nil, "")
			Expect(err).To(BeNil())
			Expect(getAccessRulesResponse.Code).To(Equal(http.StatusOK))

			response := httpResponseBodyToAccessRules(getAccessRulesResponse)
			accessRules := response.AccessRules

			expectedCount := len(existingAccessRules)
			expectedAccessRules := existingAccessRules[:]

			if params.clusterId != "" {
				// first access rule - super admin access rule
				expectedAccessRules = []api.AccessRule{existingAccessRules[0]}

				if params.clusterId == TestCluster1Id {
					expectedAccessRules = append(expectedAccessRules, cluster1AccessRules...)
					expectedAccessRules = append(expectedAccessRules, tenantAccessRules...)

					if params.includeDeleted {
						expectedAccessRules = append(expectedAccessRules, deletedAccessRules...)
					}
				} else if params.clusterId == TestCluster2Id {
					expectedAccessRules = append(expectedAccessRules, cluster2AccessRules...)
					expectedAccessRules = append(expectedAccessRules, tenantAccessRules...)
				}

				expectedCount = len(expectedAccessRules)
			} else if params.includeDeleted {
				expectedCount = len(allAccessRulesDeletedIncluded)
				expectedAccessRules = allAccessRulesDeletedIncluded[:]
			}

			if lastUpdated != "" {
				lastUpdatedTime, err := time.Parse(time.RFC3339Nano, lastUpdated)
				Expect(err).To(BeNil())

				expectedAccessRulesWLastUpdated := []api.AccessRule{}
				for _, createdAccessRule := range expectedAccessRules {
					if !createdAccessRule.UpdatedAt.Before(lastUpdatedTime) ||
						(createdAccessRule.DeletedAt != nil && !createdAccessRule.DeletedAt.Before(lastUpdatedTime)) {
						expectedAccessRulesWLastUpdated = append(expectedAccessRulesWLastUpdated, createdAccessRule)
					}
				}

				expectedCount = len(expectedAccessRulesWLastUpdated)
				expectedAccessRules = expectedAccessRulesWLastUpdated
			}

			if params.limit > 0 || params.offset > 0 {
				previousExpectedCount := expectedCount
				expectedCount = params.limit
				if params.limit == 0 || params.offset+params.limit > previousExpectedCount {
					expectedCount = previousExpectedCount - params.offset
				}
				if params.offset > previousExpectedCount {
					expectedCount = 0
				}

				if expectedCount > 0 {
					expectedAccessRules = expectedAccessRules[params.offset : params.offset+expectedCount]
				}
			}

			if params.expectedCountBiggerThan != nil {
				Expect(expectedCount).To(BeNumerically(">", *params.expectedCountBiggerThan))
			} else if params.expectedCount != nil {
				Expect(expectedCount).To(Equal(*params.expectedCount))
			}

			Expect(response.TotalRecords).To(Equal(int32(expectedCount)))
			Expect(response.DisplayRecords).To(Equal(int32(expectedCount)))
			Expect(len(accessRules)).To(Equal(expectedCount))

			if !params.includeDeleted && (params.limit > 0 || params.offset > 0) {
				// if offset or limit set - the result is sorted;
				// but if we have includeDeleted - the previous accessrules might have been created at the same time,
				// so if we have 2 accessrules with same timestamp,
				// they could be returned from the backend in a different order than ours.
				for idx := range accessRules {
					Expect(accessRules[idx].Id).To(Equal(expectedAccessRules[idx].Id))
				}
			} else {
				// if no offset or limit set - the result is un-sorted
				Expect(getAccessRulesIds(accessRules)).To(ContainElements(getAccessRulesIds(expectedAccessRules)))
			}
		}

		DescribeTable("GetAccessRules query params", validateGetAccessRulesQueryParamsFn, entries)
	})
})

func getAccessRulesIds(accessRules []api.AccessRule) []int32 {
	result := []int32{}
	for _, accessRule := range accessRules {
		result = append(result, accessRule.Id)
	}
	return result
}
