/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
	testing "github.com/run-ai/runai/control-plane/authorization/pkg/test"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"

	audit "github.com/run-ai/runai/control-plane/audit-service/pkg/client"
	auditapi "github.com/run-ai/runai/control-plane/audit-service/pkg/client/generated"
	auditmocks "github.com/run-ai/runai/control-plane/audit-service/pkg/client/mocks"
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	authorizationutils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	authorization_config "github.com/run-ai/runai/control-plane/authorization/pkg/config"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/dao"
	static_legacy_roles "github.com/run-ai/runai/control-plane/authorization/pkg/database/dao/static-legacy-roles"
	static_roles "github.com/run-ai/runai/control-plane/authorization/pkg/database/dao/static-roles"
	cluster_service_client "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/runai-common-packages/config"
)

var _ = Describe("Access rules batch API tests", Ordered, func() {
	tenantId := fmt.Sprintf("%d", testing.TestTenant)
	testAdminsGroups := []string{dao.SuperAdmins}
	var createdRoles []api.Role
	var createdAccessRules []api.AccessRule
	var createdAccessRulesIds []string
	ctx := setContext(tenantId, TestUserId, testAdminsGroups)
	anotherCtx := setContext(tenantId, TestUser2Id, []string{})

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
	accessRulesBatchApiService := NewAccessRulesApiService(tenantSettings)

	AfterAll(func() {
		CleanupAccessRules(accessRulesApiService, ctx, createdAccessRules)
		createdAccessRules = []api.AccessRule{}

		for _, role := range createdRoles {
			_, err := rolesApiService.DeleteRole(ctx, role.Id)
			Expect(err).To(BeNil())
		}
		createdRoles = []api.Role{}
		TestsCleanupAndInit()
	})

	BeforeAll(func() {
		TestsCleanupAndInit()
	})

	BeforeEach(func() {
		auditMock := auditmocks.NewWriter(GinkgoT())
		auditMock.On("Write", mock.Anything).Return(nil).Maybe()
		audit.SetMock(auditMock)

		roleBody := createRoleCreationRequest(static_legacy_roles.ResearcherL1.Permissions, "test roles", "can fully manage projects")

		createdRole := createTestRole(ctx, rolesApiService, roleBody, &createdRoles, false, http.StatusCreated)

		accessRuleBody := api.AccessRuleCreationFields{
			ScopeId:     ProjectId,
			ScopeType:   api.SCOPETYPE_PROJECT,
			SubjectId:   TestUserId,
			SubjectType: api.SUBJECTTYPE_USER,
			RoleId:      createdRole.Id,
		}

		SetAncestorsCache(ctx, ProjectId, api.RESOURCETYPE_PROJECT, &authorizationutils.Ancestors{
			TenantId: tenantId,
		})

		createdAccessRule := createTestAccessRule(ctx, accessRulesApiService, accessRuleBody, &createdAccessRules, false, http.StatusCreated)
		Expect(createdAccessRule.SubjectId).To(Equal(accessRuleBody.SubjectId))
		Expect(createdAccessRule.SubjectType).To(Equal(api.SubjectTypeExtended(accessRuleBody.SubjectType)))
		Expect(createdAccessRule.ScopeId).To(Equal(accessRuleBody.ScopeId))
		Expect(createdAccessRule.ScopeType).To(Equal(accessRuleBody.ScopeType))
		Expect(createdAccessRule.ScopeName).To(Equal(ProjectName))
		Expect(createdAccessRule.RoleId).To(Equal(createdRole.Id))
		Expect(createdAccessRule.RoleName).To(Equal(createdRole.Name))

		SetAncestorsCache(ctx, DepartmentId, api.RESOURCETYPE_DEPARTMENT, &authorizationutils.Ancestors{
			TenantId: tenantId,
		})

		SetAncestorsCache(ctx, ClusterId, api.RESOURCETYPE_CLUSTER, &authorizationutils.Ancestors{
			TenantId: tenantId,
		})

		bodyDepartment := api.AccessRuleCreationFields{
			ScopeId:     DepartmentId,
			ScopeType:   api.SCOPETYPE_DEPARTMENT,
			SubjectId:   TestUserId,
			SubjectType: api.SUBJECTTYPE_USER,
			RoleId:      createdRoles[0].Id,
		}

		createdAccessRuleDepartment := createTestAccessRule(ctx, accessRulesApiService, bodyDepartment, &createdAccessRules, false, http.StatusCreated)
		Expect(createdAccessRuleDepartment.ScopeName).To(Equal(DepartmentName))

		bodyCluster := api.AccessRuleCreationFields{
			ScopeId:     ClusterId,
			ScopeType:   api.SCOPETYPE_CLUSTER,
			SubjectId:   TestUserId,
			SubjectType: api.SUBJECTTYPE_USER,
			RoleId:      createdRoles[0].Id,
		}

		createdAccessRuleCluster := createTestAccessRule(ctx, accessRulesApiService, bodyCluster, &createdAccessRules, false, http.StatusCreated)
		Expect(createdAccessRuleCluster.ScopeName).To(Equal(ClusterName))

		bodyTenant := api.AccessRuleCreationFields{
			ScopeId:     tenantId,
			ScopeType:   api.SCOPETYPE_TENANT,
			SubjectId:   TestUserId,
			SubjectType: api.SUBJECTTYPE_USER,
			RoleId:      createdRoles[0].Id,
		}

		createdAccessRuleTenant := createTestAccessRule(ctx, accessRulesApiService, bodyTenant, &createdAccessRules, false, http.StatusCreated)
		Expect(createdAccessRuleTenant.ScopeName).To(Equal(TenantName))

		accessRuleToCreate := api.AccessRuleCreationFields{
			ScopeId:     DepartmentId,
			ScopeType:   api.SCOPETYPE_DEPARTMENT,
			SubjectId:   createdAccessRules[0].SubjectId,
			SubjectType: api.SubjectType(createdAccessRules[0].SubjectType),
			RoleId:      createdAccessRules[0].RoleId,
		}
		createTestAccessRule(ctx, accessRulesApiService, accessRuleToCreate, &createdAccessRules, false, http.StatusCreated)

		createdAccessRulesIds = getCreatedAccessRulesIds(createdAccessRules)
	})

	AfterEach(func() {
		CleanupAccessRules(accessRulesApiService, ctx, createdAccessRules)
		createdAccessRules = []api.AccessRule{}

		for _, role := range createdRoles {
			_, err := rolesApiService.DeleteRole(ctx, role.Id)
			Expect(err).To(BeNil())
		}
		createdRoles = []api.Role{}
	})

	Context("Delete", func() {
		It("Delete Batch", func() {
			response, err := accessRulesBatchApiService.AccessRulesBatch(ctx, api.AccessRulesBatchFields{
				Ids: createdAccessRulesIds, Action: "delete",
			})
			Expect(err).To(BeNil())
			Expect(response).To(Not(BeNil()))
			batchResponse := response.Body.(api.BatchResponse)

			Expect(len(batchResponse.Succeeded)).To(Equal(len(createdAccessRulesIds)))

			createdAccessRules = []api.AccessRule{}
			createdAccessRulesIds = []string{}
		})

		It("Delete Batch bigger than Max", func() {
			var ids []string

			for i := range MaxBatchDeletionSize + 1 {
				ids = append(ids, strconv.Itoa(i))
			}
			response, err := accessRulesBatchApiService.AccessRulesBatch(ctx, api.AccessRulesBatchFields{
				Ids: ids, Action: "delete",
			})
			Expect(err).To(Not(BeNil()))
			Expect(response).To(Not(BeNil()))

			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("Some failed Deletion Batch", func() {
			var almostAllPermissions []*models.Permission
			for _, permission := range static_roles.Viewer.Permissions {
				if permission.ResourceType == api.RESOURCETYPE_DEPARTMENT {
					continue
				}
				almostAllPermissions = append(almostAllPermissions, permission)
			}
			convertedPermissions := convertPermissions(almostAllPermissions)
			roleBody := api.RoleCreationFieldsInternal{
				Name:        "viewer test role",
				Description: "Viewer test role",
				Permissions: convertedPermissions,
			}
			createdRole := createTestRole(ctx, rolesApiService, roleBody, &createdRoles, false, http.StatusCreated)

			bodyTenant := api.AccessRuleCreationFields{
				ScopeId:     ProjectId,
				ScopeType:   api.SCOPETYPE_PROJECT,
				SubjectId:   TestUser2Id,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      createdRole.Id,
			}

			createdAccessRuleTenant := createTestAccessRule(ctx, accessRulesApiService, bodyTenant, &createdAccessRules, false, http.StatusCreated)
			Expect(createdAccessRuleTenant.ScopeName).To(Equal(fmt.Sprintf("Project-%s", ProjectId)))
			idsSet := map[string]bool{}

			for _, accessRule := range createdAccessRules {
				if accessRule.TenantId == int32(testing.TestTenant) {
					idsSet[strconv.Itoa(int(accessRule.Id))] = true
				}
			}
			ids := convertMapKeysToSlice(idsSet)

			response, err := accessRulesBatchApiService.AccessRulesBatch(anotherCtx, api.AccessRulesBatchFields{
				Ids: ids, Action: api.ACCESSRULESBATCHACTION_VALIDATE_DELETE,
			})
			Expect(err).To(BeNil())
			Expect(response).To(Not(BeNil()))
			batchResponse := response.Body.(api.BatchResponse)
			Expect(len(batchResponse.Succeeded)).To(Equal(1))
			Expect(len(batchResponse.Failed)).To(Equal(4))

			response, err = accessRulesBatchApiService.AccessRulesBatch(anotherCtx, api.AccessRulesBatchFields{
				Ids: ids, Action: api.ACCESSRULESBATCHACTION_DELETE,
			})
			Expect(err).To(BeNil())
			Expect(response).To(Not(BeNil()))
			batchResponse = response.Body.(api.BatchResponse)
			Expect(len(batchResponse.Succeeded)).To(Equal(1))
			Expect(len(batchResponse.Failed)).To(Equal(4))

			validateAccessRuleFetchResponse(ctx, accessRulesBatchApiService, batchResponse)
		})

		It("Batch Delete with Mixed Permission Levels and Scope Validation", func() {
			// Create a role with limited permissions (only READ)
			limitedRoleBody := api.RoleCreationFieldsInternal{
				Name:        "limited role",
				Description: "Role with limited permissions",
				Permissions: []api.PostPermission{
					{
						ResourceType: api.RESOURCETYPE_PROJECT,
						Actions:      []api.Action{api.ACTION_READ},
					},
				},
			}
			limitedRole := createTestRole(ctx, rolesApiService, limitedRoleBody, &createdRoles, false, http.StatusCreated)

			// Create a role with full permissions
			fullRoleBody := api.RoleCreationFieldsInternal{
				Name:        "full role",
				Description: "Role with full permissions",
				Permissions: []api.PostPermission{
					{
						ResourceType: api.RESOURCETYPE_PROJECT,
						Actions:      []api.Action{api.ACTION_CREATE, api.ACTION_UPDATE, api.ACTION_DELETE, api.ACTION_READ},
					},
				},
			}
			fullRole := createTestRole(ctx, rolesApiService, fullRoleBody, &createdRoles, false, http.StatusCreated)

			// Create three access rules for TestUserId:
			// 1. With limited permissions (should fail deletion)
			// 2. With full permissions (should fail deletion)
			// 3. With scope that's not in TestUser2Id's context (should fail deletion)

			// Access rule with limited permissions
			limitedAccessRuleBody := api.AccessRuleCreationFields{
				ScopeId:     ProjectId,
				ScopeType:   api.SCOPETYPE_PROJECT,
				SubjectId:   TestUser2Id,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      limitedRole.Id,
			}
			limitedAccessRule := createTestAccessRule(ctx, accessRulesApiService, limitedAccessRuleBody, &createdAccessRules, false, http.StatusCreated)

			// Access rule with full permissions
			fullAccessRuleBody := api.AccessRuleCreationFields{
				ScopeId:     ProjectId,
				ScopeType:   api.SCOPETYPE_PROJECT,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      fullRole.Id,
			}
			fullAccessRule := createTestAccessRule(ctx, accessRulesApiService, fullAccessRuleBody, &createdAccessRules, false, http.StatusCreated)

			// Access rule with a different scope that's not in TestUser2Id's context
			otherScopeAccessRuleBody := api.AccessRuleCreationFields{
				ScopeId:     DepartmentId,
				ScopeType:   api.SCOPETYPE_DEPARTMENT,
				SubjectId:   TestUserId,
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      fullRole.Id,
			}
			otherScopeAccessRule := createTestAccessRule(ctx, accessRulesApiService, otherScopeAccessRuleBody, &createdAccessRules, false, http.StatusCreated)

			// Attempt to delete all three access rules in a batch using TestUser2Id's context
			ids := []string{
				strconv.Itoa(int(limitedAccessRule.Id)),
				strconv.Itoa(int(fullAccessRule.Id)),
				strconv.Itoa(int(otherScopeAccessRule.Id)),
			}

			response, err := accessRulesBatchApiService.AccessRulesBatch(anotherCtx, api.AccessRulesBatchFields{
				Ids:    ids,
				Action: api.ACCESSRULESBATCHACTION_VALIDATE_DELETE,
			})
			Expect(err).To(BeNil())
			Expect(response).To(Not(BeNil()))
			batchResponse := response.Body.(api.BatchResponse)
			Expect(len(batchResponse.Succeeded)).To(Equal(1))
			Expect(len(batchResponse.Failed)).To(Equal(2))

			response, err = accessRulesBatchApiService.AccessRulesBatch(anotherCtx, api.AccessRulesBatchFields{
				Ids:    ids,
				Action: api.ACCESSRULESBATCHACTION_DELETE,
			})
			Expect(err).To(BeNil())
			Expect(response).To(Not(BeNil()))
			batchResponse = response.Body.(api.BatchResponse)

			// Verify that only one rule was deleted successfully (the one with limited permissions)
			Expect(len(batchResponse.Succeeded)).To(Equal(1))
			Expect(len(batchResponse.Failed)).To(Equal(2))

			// Verify that the failed rules are the ones with full permissions and department scope
			failedIds := make([]string, len(batchResponse.Failed))
			for i, failed := range batchResponse.Failed {
				failedIds[i] = failed.Id
			}
			Expect(failedIds).To(ContainElements(
				strconv.Itoa(int(fullAccessRule.Id)),
				strconv.Itoa(int(otherScopeAccessRule.Id)),
			))

			validateAccessRuleFetchResponse(ctx, accessRulesBatchApiService, batchResponse)
			succeededIds := make([]string, len(batchResponse.Succeeded))
			for i, succeeded := range batchResponse.Succeeded {
				succeededIds[i] = succeeded
			}
			Expect(succeededIds).To(ContainElements(
				strconv.Itoa(int(limitedAccessRule.Id)),
			))
		})
	})

	Context("Create", func() {
		Describe("Bad request", func() {
			It("Create Batch bigger than Max - scopes", func() {
				scopes := []api.AccessRuleScopesInner{}

				for i := range MaxBatchCreationScopesSize + 1 {
					scopes = append(scopes, api.AccessRuleScopesInner{
						ScopeId:   fmt.Sprint(i),
						ScopeType: tenantScopeType,
					})
				}
				response, err := accessRulesBatchApiService.AccessRulesBatchCreate(ctx, api.AccessRulesBatchCreateFields{
					Payload: api.AccessRulesBatchCreateFieldsPayload{
						RoleId: 5,
						Subjects: []api.AccessRuleSubjectsInner{{
							SubjectId:   "1",
							SubjectType: "user",
						}},
						Scopes: scopes,
					},
				})
				Expect(err).To(Not(BeNil()))
				Expect(response).To(Not(BeNil()))

				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})

			It("Create Batch bigger than Max - subjects", func() {
				scopes := []api.AccessRuleScopesInner{}
				subjects := []api.AccessRuleSubjectsInner{}

				for i := range 3 {
					scopes = append(scopes, api.AccessRuleScopesInner{
						ScopeId:   fmt.Sprint(i),
						ScopeType: tenantScopeType,
					})
				}
				for i := range MaxBatchCreationSubjectsSize + 1 {
					subjects = append(subjects, api.AccessRuleSubjectsInner{
						SubjectId:   fmt.Sprint(i),
						SubjectType: "user",
					})
				}
				response, err := accessRulesBatchApiService.AccessRulesBatchCreate(ctx, api.AccessRulesBatchCreateFields{
					Payload: api.AccessRulesBatchCreateFieldsPayload{
						RoleId:   5,
						Subjects: subjects,
						Scopes:   scopes,
					},
				})
				Expect(err).To(Not(BeNil()))
				Expect(response).To(Not(BeNil()))

				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})

			It("Create Batch - no scopes", func() {
				response, err := accessRulesBatchApiService.AccessRulesBatchCreate(ctx, api.AccessRulesBatchCreateFields{
					Payload: api.AccessRulesBatchCreateFieldsPayload{
						RoleId: 5,
						Subjects: []api.AccessRuleSubjectsInner{{
							SubjectId:   "1",
							SubjectType: "user",
						}},
						Scopes: []api.AccessRuleScopesInner{},
					},
				})
				Expect(err).To(Not(BeNil()))
				Expect(response).To(Not(BeNil()))

				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})

			It("Create Batch - no subjects", func() {
				response, err := accessRulesBatchApiService.AccessRulesBatchCreate(ctx, api.AccessRulesBatchCreateFields{
					Payload: api.AccessRulesBatchCreateFieldsPayload{
						RoleId:   5,
						Subjects: []api.AccessRuleSubjectsInner{},
						Scopes: []api.AccessRuleScopesInner{{
							ScopeId:   "1",
							ScopeType: "tenant",
						}},
					},
				})
				Expect(err).To(Not(BeNil()))
				Expect(response).To(Not(BeNil()))

				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Describe("Passed", func() {
			var roleId int32
			var subjectId1, subjectId2 string

			BeforeAll(func() {
				roleId = int32(dao.MapStaticRoleNameToRoleId[static_roles.Viewer.Name])

				subjectId1 = "some-unique-id"
				subjectId2 = "some-unique-id2"
				SetAncestorsCache(ctx, TestProject3Id, api.RESOURCETYPE_PROJECT, &authorizationutils.Ancestors{
					TenantId:    tenantId,
					ClusterUuid: &TestCluster2Id,
				})
				SetAncestorsCache(ctx, TestProject4Id, api.RESOURCETYPE_PROJECT, &authorizationutils.Ancestors{
					TenantId:    tenantId,
					ClusterUuid: &TestCluster2Id,
				})
			})

			It("Create Batch multi scopes - all passed", func() {
				scopes := []api.AccessRuleScopesInner{}

				scopes = append(scopes, api.AccessRuleScopesInner{
					ScopeId:   tenantId,
					ScopeType: api.SCOPETYPE_TENANT,
				}, api.AccessRuleScopesInner{
					ScopeId:   TestProject3Id,
					ScopeType: api.SCOPETYPE_PROJECT,
				}, api.AccessRuleScopesInner{
					ScopeId:   TestProject4Id,
					ScopeType: api.SCOPETYPE_PROJECT,
				})
				response, err := accessRulesBatchApiService.AccessRulesBatchCreate(ctx, api.AccessRulesBatchCreateFields{
					Payload: api.AccessRulesBatchCreateFieldsPayload{
						RoleId: roleId,
						Subjects: []api.AccessRuleSubjectsInner{{
							SubjectId:   subjectId1,
							SubjectType: api.SUBJECTTYPE_USER,
						}},
						Scopes: scopes,
					},
				})
				Expect(err).To(BeNil())
				Expect(response).To(Not(BeNil()))

				Expect(response.Code).To(Equal(http.StatusOK))

				accessRulesResponse := response.Body.(api.AccessRulesBatchCreateResponse)
				Expect(len(accessRulesResponse.Failed)).To(Equal(0))
				Expect(len(accessRulesResponse.Succeeded)).To(Equal(3))

				for i := range accessRulesResponse.Succeeded {
					Expect(accessRulesResponse.Succeeded[i].SubjectType).To(Equal(api.SUBJECTTYPEEXTENDED_USER))
					Expect(accessRulesResponse.Succeeded[i].SubjectId).To(Equal(subjectId1))
					Expect(accessRulesResponse.Succeeded[i].RoleId).To(Equal(roleId))
					Expect(accessRulesResponse.Succeeded[i].RoleName).To(Equal(static_roles.Viewer.Name))
					scopeIdApi := accessRulesResponse.Succeeded[i].ScopeId
					switch scopeIdApi {
					case tenantId:
						Expect(accessRulesResponse.Succeeded[i].ScopeType).To(Equal(api.SCOPETYPE_TENANT))
					case TestProject3Id, TestProject4Id:
						Expect(accessRulesResponse.Succeeded[i].ScopeType).To(Equal(api.SCOPETYPE_PROJECT))
					default:
						Fail(fmt.Sprintf("unknown scopeId %s", scopeIdApi))
					}
					createdAccessRules = append(createdAccessRules, accessRulesResponse.Succeeded[i])
				}
			})

			It("Create Batch multi scopes - some failed", func() {
				scopes := []api.AccessRuleScopesInner{}

				scopes = append(scopes, api.AccessRuleScopesInner{
					ScopeId:   tenantId,
					ScopeType: api.SCOPETYPE_TENANT,
				}, api.AccessRuleScopesInner{
					ScopeId:   "123123",
					ScopeType: api.SCOPETYPE_PROJECT,
				}, api.AccessRuleScopesInner{
					ScopeId:   TestProject4Id,
					ScopeType: api.SCOPETYPE_PROJECT,
				})
				response, err := accessRulesBatchApiService.AccessRulesBatchCreate(ctx, api.AccessRulesBatchCreateFields{
					Payload: api.AccessRulesBatchCreateFieldsPayload{
						RoleId: roleId,
						Subjects: []api.AccessRuleSubjectsInner{{
							SubjectId:   subjectId1,
							SubjectType: api.SUBJECTTYPE_USER,
						}},
						Scopes: scopes,
					},
				})
				Expect(err).To(BeNil())
				Expect(response).To(Not(BeNil()))

				Expect(response.Code).To(Equal(http.StatusOK))

				accessRulesResponse := response.Body.(api.AccessRulesBatchCreateResponse)
				Expect(len(accessRulesResponse.Failed)).To(Equal(1))
				Expect(len(accessRulesResponse.Succeeded)).To(Equal(2))

				failedAccessRule := accessRulesResponse.Failed[0]
				Expect(failedAccessRule.Code).To(Equal(int32(http.StatusForbidden)))
				Expect(failedAccessRule.Payload.ScopeType).To(Equal(api.SCOPETYPE_PROJECT))
				Expect(failedAccessRule.Payload.ScopeId).To(Equal("123123"))
				Expect(failedAccessRule.Payload.RoleId).To(Equal(roleId))

				for i := range accessRulesResponse.Succeeded {
					Expect(accessRulesResponse.Succeeded[i].SubjectType).To(Equal(api.SUBJECTTYPEEXTENDED_USER))
					Expect(accessRulesResponse.Succeeded[i].SubjectId).To(Equal(subjectId1))
					Expect(accessRulesResponse.Succeeded[i].RoleId).To(Equal(roleId))
					Expect(accessRulesResponse.Succeeded[i].RoleName).To(Equal(static_roles.Viewer.Name))
					scopeIdApi := accessRulesResponse.Succeeded[i].ScopeId
					switch scopeIdApi {
					case tenantId:
						Expect(accessRulesResponse.Succeeded[i].ScopeType).To(Equal(api.SCOPETYPE_TENANT))
					case TestProject4Id:
						Expect(accessRulesResponse.Succeeded[i].ScopeType).To(Equal(api.SCOPETYPE_PROJECT))
					default:
						Fail(fmt.Sprintf("unknown scopeId %s", scopeIdApi))
					}
					createdAccessRules = append(createdAccessRules, accessRulesResponse.Succeeded[i])
				}
			})

			It("Create Batch multi subjects - all passed", func() {
				subjects := []api.AccessRuleSubjectsInner{}

				subjects = append(subjects, api.AccessRuleSubjectsInner{
					SubjectId:   subjectId1,
					SubjectType: api.SUBJECTTYPE_USER,
				}, api.AccessRuleSubjectsInner{
					SubjectId:   subjectId2,
					SubjectType: api.SUBJECTTYPE_USER,
				})
				response, err := accessRulesBatchApiService.AccessRulesBatchCreate(ctx, api.AccessRulesBatchCreateFields{
					Payload: api.AccessRulesBatchCreateFieldsPayload{
						RoleId:   roleId,
						Subjects: subjects,
						Scopes: []api.AccessRuleScopesInner{{
							ScopeId:   tenantId,
							ScopeType: api.SCOPETYPE_TENANT,
						}},
					},
				})
				Expect(err).To(BeNil())
				Expect(response).To(Not(BeNil()))

				Expect(response.Code).To(Equal(http.StatusOK))

				accessRulesResponse := response.Body.(api.AccessRulesBatchCreateResponse)
				Expect(len(accessRulesResponse.Failed)).To(Equal(0))
				Expect(len(accessRulesResponse.Succeeded)).To(Equal(2))

				for i := range accessRulesResponse.Succeeded {
					Expect(accessRulesResponse.Succeeded[i].ScopeType).To(Equal(api.SCOPETYPE_TENANT))
					Expect(accessRulesResponse.Succeeded[i].ScopeId).To(Equal(tenantId))
					Expect(accessRulesResponse.Succeeded[i].RoleId).To(Equal(roleId))
					Expect(accessRulesResponse.Succeeded[i].RoleName).To(Equal(static_roles.Viewer.Name))
					subjectIdApi := accessRulesResponse.Succeeded[i].SubjectId
					switch subjectIdApi {
					case subjectId1, subjectId2:
						Expect(accessRulesResponse.Succeeded[i].SubjectType).To(Equal(api.SUBJECTTYPEEXTENDED_USER))
					default:
						Fail(fmt.Sprintf("unknown subjectId %s", subjectIdApi))
					}
					createdAccessRules = append(createdAccessRules, accessRulesResponse.Succeeded[i])
				}
			})

			It("Create Batch - all fail", func() {
				subjects := []api.AccessRuleSubjectsInner{}

				subjects = append(subjects, api.AccessRuleSubjectsInner{
					SubjectId:   subjectId1,
					SubjectType: api.SUBJECTTYPE_USER,
				}, api.AccessRuleSubjectsInner{
					SubjectId:   subjectId2,
					SubjectType: api.SUBJECTTYPE_USER,
				})
				response, err := accessRulesBatchApiService.AccessRulesBatchCreate(ctx, api.AccessRulesBatchCreateFields{
					Payload: api.AccessRulesBatchCreateFieldsPayload{
						RoleId:   roleId,
						Subjects: subjects,
						Scopes: []api.AccessRuleScopesInner{{
							ScopeId:   "123123",
							ScopeType: api.SCOPETYPE_TENANT,
						}},
					},
				})
				Expect(err).To(BeNil())
				Expect(response).To(Not(BeNil()))

				Expect(response.Code).To(Equal(http.StatusOK))

				accessRulesResponse := response.Body.(api.AccessRulesBatchCreateResponse)
				Expect(len(accessRulesResponse.Failed)).To(Equal(2))
				Expect(len(accessRulesResponse.Succeeded)).To(Equal(0))

				for i := range accessRulesResponse.Failed {
					Expect(accessRulesResponse.Failed[i].Payload.ScopeType).To(Equal(api.SCOPETYPE_TENANT))
					Expect(accessRulesResponse.Failed[i].Payload.ScopeId).To(Equal("123123"))
					Expect(accessRulesResponse.Failed[i].Payload.RoleId).To(Equal(roleId))
					subjectIdApi := accessRulesResponse.Failed[i].Payload.SubjectId
					switch subjectIdApi {
					case subjectId1, subjectId2:
						Expect(accessRulesResponse.Failed[i].Payload.SubjectType).To(Equal(api.SUBJECTTYPE_USER))
					default:
						Fail(fmt.Sprintf("unknown subjectId %s", subjectIdApi))
					}
				}
			})
		})
	})

	Context("Audit logs", func() {
		It("prints audit log on batch create", func() {
			roleId := int32(dao.MapStaticRoleNameToRoleId[static_roles.Viewer.Name])

			auditMock := auditmocks.NewWriter(GinkgoT())
			auditMock.On("Write", mock.MatchedBy(func(log auditapi.AuditLog) bool {
				return log.Action == "Create" &&
					log.GetHttpMethod() == "POST" &&
					log.GetEntityType() == "access rule" && log.GetEntityId() != ""
			})).Return(nil)
			audit.SetMock(auditMock)

			response, err := accessRulesBatchApiService.AccessRulesBatchCreate(ctx, api.AccessRulesBatchCreateFields{
				Payload: api.AccessRulesBatchCreateFieldsPayload{
					RoleId: roleId,
					Subjects: []api.AccessRuleSubjectsInner{{
						SubjectId:   "audit-subject",
						SubjectType: api.SUBJECTTYPE_USER,
					}},
					Scopes: []api.AccessRuleScopesInner{{
						ScopeId:   tenantId,
						ScopeType: api.SCOPETYPE_TENANT,
					}},
				},
			})
			Expect(err).To(BeNil())
			Expect(response).To(Not(BeNil()))
			Expect(response.Code).To(Equal(http.StatusOK))
			body := response.Body.(api.AccessRulesBatchCreateResponse)
			Expect(len(body.Succeeded)).To(Equal(1))
		})

		It("prints audit log on batch delete", func() {
			roleId := int32(dao.MapStaticRoleNameToRoleId[static_roles.Viewer.Name])
			createdAccessRule := createTestAccessRule(ctx, accessRulesApiService, api.AccessRuleCreationFields{
				ScopeId:     tenantId,
				ScopeType:   api.SCOPETYPE_TENANT,
				SubjectId:   "audit-delete-subject",
				SubjectType: api.SUBJECTTYPE_USER,
				RoleId:      roleId,
			}, &createdAccessRules, false, http.StatusCreated)

			auditMock := auditmocks.NewWriter(GinkgoT())
			auditMock.On("Write", mock.MatchedBy(func(log auditapi.AuditLog) bool {
				return log.Action == "Delete" &&
					log.GetHttpMethod() == "DELETE" &&
					log.GetEntityType() == "access rule" &&
					log.GetEntityId() == strconv.Itoa(int(createdAccessRule.Id))
			})).Return(nil)
			audit.SetMock(auditMock)
			resp, err := accessRulesBatchApiService.AccessRulesBatch(ctx, api.AccessRulesBatchFields{
				Ids:    []string{strconv.Itoa(int(createdAccessRule.Id))},
				Action: api.ACCESSRULESBATCHACTION_DELETE,
			})
			Expect(err).To(BeNil())
			Expect(resp).To(Not(BeNil()))
			Expect(resp.Code).To(Equal(http.StatusOK))
			body := resp.Body.(api.BatchResponse)
			Expect(body.Succeeded).To(ContainElement(strconv.Itoa(int(createdAccessRule.Id))))
		})
	})
})

func validateAccessRuleFetchResponse(ctx context.Context, accessRulesBatchApiService api.AccessRulesAPIServicer, batchResponse api.BatchResponse) {
	failedIds := make([]string, len(batchResponse.Failed))
	for i, failed := range batchResponse.Failed {
		failedIds[i] = failed.Id
	}
	validateAccessRuleFetchResponseForIds(ctx, accessRulesBatchApiService, failedIds, http.StatusOK)
	succeededIds := make([]string, len(batchResponse.Succeeded))
	for i, succeeded := range batchResponse.Succeeded {
		succeededIds[i] = succeeded
	}
	validateAccessRuleFetchResponseForIds(ctx, accessRulesBatchApiService, succeededIds, http.StatusNotFound)
}

func validateAccessRuleFetchResponseForIds(ctx context.Context, accessRulesBatchApiService api.AccessRulesAPIServicer, ids []string, code int) {
	for _, idStr := range ids {
		id, _ := strconv.ParseInt(idStr, 10, 32)
		response, err := accessRulesBatchApiService.GetAccessRule(ctx, int32(id))
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(code))
	}
}

func createRoleCreationRequest(inputPermissions []*models.Permission, name, description string) api.RoleCreationFieldsInternal {
	permissions := convertPermissions(inputPermissions)
	permissions = append(permissions, api.PostPermission{
		ResourceType: api.RESOURCETYPE_PROJECT,
		Actions:      []api.Action{api.ACTION_CREATE, api.ACTION_UPDATE, api.ACTION_DELETE, api.ACTION_READ},
	})

	roleBody := api.RoleCreationFieldsInternal{
		Name:                  name,
		Description:           description,
		Permissions:           permissions,
		KubernetesPermissions: api.KubernetesPermissions{PredefinedRole: ptr.To(fmt.Sprintf("%d", dao.MapStaticRoleNameToRoleId[static_legacy_roles.ResearcherL1.Name]))},
	}
	return roleBody
}

func convertPermissions(inputPermissions []*models.Permission) []api.PostPermission {
	var permissions []api.PostPermission
	for _, l1ResearcherPermission := range inputPermissions {
		if l1ResearcherPermission.ResourceType == api.RESOURCETYPE_PROJECT {
			continue
		}
		var actions []api.Action
		if l1ResearcherPermission.Read {
			actions = append(actions, api.ACTION_READ)
		}
		if l1ResearcherPermission.Create {
			actions = append(actions, api.ACTION_CREATE)
		}
		if l1ResearcherPermission.Update {
			actions = append(actions, api.ACTION_UPDATE)
		}
		if l1ResearcherPermission.Delete {
			actions = append(actions, api.ACTION_DELETE)
		}
		permissions = append(permissions, api.PostPermission{
			ResourceType: l1ResearcherPermission.ResourceType,
			Actions:      actions,
		})
	}
	return permissions
}

func getCreatedAccessRulesIds(createdAccessRules []api.AccessRule) []string {
	idsSet := map[string]bool{}

	for _, accessRule := range createdAccessRules {
		if accessRule.TenantId == int32(testing.TestTenant) {
			idsSet[strconv.Itoa(int(accessRule.Id))] = true
		}
	}
	return convertMapKeysToSlice(idsSet)
}
