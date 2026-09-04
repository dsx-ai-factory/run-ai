/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"

	"k8s.io/utils/ptr"
)

var _ = Describe("Datavolumes API tests", Ordered, func() {
	clusterId1 := uuid.New().String()
	clusterId2 := uuid.New().String()
	clusterId3 := uuid.New().String()
	clusterId4 := uuid.New().String()
	clusterIdOld := uuid.New().String()

	projectIdInDifferentTenant := "20"

	// a map of projectId -> [departmentId, clusterId, tenantId]
	projectToAncestors := map[string][]string{
		"1": {"1", clusterId1, TestTenant},
		"2": {"1", clusterId1, TestTenant},
		"3": {"3", clusterId1, TestTenant},
		"4": {"20", clusterId2, TestTenant},
		"5": {"21", clusterId2, TestTenant},

		"6": {"4", clusterId1, TestTenant},
		"7": {"5", clusterId1, TestTenant},
		"8": {"10", clusterId2, TestTenant},

		"9":  {"1", clusterId1, TestTenant},
		"10": {"3", clusterId1, TestTenant},
		"11": {"3", clusterId1, TestTenant},
		"12": {"6", clusterId1, TestTenant},
		"13": {"6", clusterId1, TestTenant},
		"14": {"7", clusterId1, TestTenant},
		"15": {"8", clusterId4, TestTenant},

		"200": {"200", clusterIdOld, TestTenant},

		"": {"", "", ""}, // invalid project

		projectIdInDifferentTenant: {projectIdInDifferentTenant, clusterId3, TestTenant2},
	}

	// a map of datavolumeId -> tenantId
	datavolumeIds := map[string]string{}
	var datavolumesApiService api.DatavolumesAPIServicer
	ctx := SetContext(TestTenant, TestUser)

	BeforeAll(func() {
		for projectId, ancestors := range projectToAncestors {
			if projectId == "" {
				continue
			}

			SetProjectAncestorsCache(ctx, projectId, &authorization_utils.Ancestors{
				TenantId:     ancestors[2],
				ClusterUuid:  &ancestors[1],
				DepartmentId: &ancestors[0],
			})
			SetAncestorsCache(ctx, ancestors[0], api.SCOPETYPE_DEPARTMENT, &authorization_utils.Ancestors{TenantId: ancestors[2], ClusterUuid: &ancestors[1]})
			SetAncestorsCache(ctx, ancestors[1], api.SCOPETYPE_CLUSTER, &authorization_utils.Ancestors{TenantId: ancestors[2]})
			SetScopeNameCache(ctx, projectId, api.SCOPETYPE_PROJECT, ancestors[1], getTestScopeName(projectId, api.SCOPETYPE_PROJECT))
			SetScopeNameCache(ctx, ancestors[0], api.SCOPETYPE_DEPARTMENT, ancestors[1], getTestScopeName(ancestors[0], api.SCOPETYPE_DEPARTMENT))
			SetScopeNameCache(ctx, ancestors[1], api.SCOPETYPE_CLUSTER, ancestors[1], getTestScopeName(ancestors[1], api.SCOPETYPE_CLUSTER))
		}

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(clusterId1, buildDisplayedCluster(clusterId1, "2.18.0")).
			WithCluster(clusterId2, buildDisplayedCluster(clusterId2, "2.18.0")).
			WithCluster(clusterId3, buildDisplayedCluster(clusterId3, "2.18.0")).
			WithCluster(clusterId4, buildDisplayedCluster(clusterId4, "2.22.0")).
			WithCluster(clusterIdOld, buildDisplayedCluster(clusterIdOld, "2.17.0")))
		checker.InitClusterCapabilitiesChecker("", getDefaultClustersCacheConfig(), "")
	})

	AfterAll(func() {
		for projectId, ancestors := range projectToAncestors {
			InvalidateProjectAncestorsCache(ctx, projectId)
			InvalidateAncestorsCache(ctx, ancestors[0], api.SCOPETYPE_DEPARTMENT)
			InvalidateAncestorsCache(ctx, ancestors[1], api.SCOPETYPE_CLUSTER)
			InvalidateScopeNameCache(ctx, projectId, api.SCOPETYPE_PROJECT, ancestors[1])
			InvalidateScopeNameCache(ctx, ancestors[0], api.SCOPETYPE_DEPARTMENT, ancestors[1])
			InvalidateScopeNameCache(ctx, ancestors[1], api.SCOPETYPE_CLUSTER, ancestors[1])
		}
	})

	BeforeEach(func() {
		ctx = SetContext(TestTenant, TestUser)

		datavolumesApiService = services.NewDatavolumesApiService()
	})

	AfterEach(func() {
		for datavolumeId, tenantIdStr := range datavolumeIds {
			tenantId, _ := strconv.Atoi(tenantIdStr)
			err := dao.DbUpdateStatus(ctx, uint(tenantId), uuid.MustParse(datavolumeId), models.DatavolumeStatus{Phase: api.DATAVOLUMEPHASE_CREATING})
			Expect(err).To(BeNil())
			err = dao.DbDeleteDatavolume(ctx, uint(tenantId), uuid.MustParse(datavolumeId))
			Expect(err).To(BeNil())
		}
		datavolumeIds = map[string]string{}
	})

	Context("create datavolume", func() {
		var (
			request   api.DatavolumeCreationFields
			projectId string
		)

		BeforeEach(func() {
			projectId = "1"
			request = api.DatavolumeCreationFields{
				Name:          "dv-0",
				Description:   "bla bla",
				ProjectId:     projectId,
				OriginPvcName: "some-pvc",
			}
		})

		It("create sanity", func() {
			now := time.Now()
			time.Sleep(1 * time.Millisecond)

			res, err := datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil(), "Create datavolume should succeed")

			Expect(res.Code).To(Equal(http.StatusCreated))
			created := res.Body.(api.Datavolume)
			datavolumeIds[created.Id] = TestTenant

			validateDatavolume(created, request, projectToAncestors)
			Expect(created.CreatedAt.After(now)).To(BeTrue())
			Expect(created.UpdatedAt.After(now)).To(BeTrue())
		})

		It("create with ShouldDeleteOriginalVolume", func() {
			request.ShouldDeleteOriginalVolume = true
			request.Name = "dv-10"
			res, err := datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil(), "Create datavolume should not succeed")

			Expect(res.Code).To(Equal(http.StatusBadRequest))
		})

		It("create with permissions in system", func() {
			ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{System: ptr.To(true)})

			res, err := datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil(), "Create datavolume should succeed")
			Expect(res.Code).To(Equal(http.StatusCreated))
			datavolumeIds[res.Body.(api.Datavolume).Id] = TestTenant
		})

		It("create with permissions in cluster", func() {
			ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Clusters: []string{projectToAncestors[projectId][1]}})

			res, err := datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil(), "Create datavolume should succeed")
			Expect(res.Code).To(Equal(http.StatusCreated))
			datavolumeIds[res.Body.(api.Datavolume).Id] = TestTenant
		})

		It("create with permissions in department", func() {
			ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Departments: []string{projectToAncestors[projectId][0]}})

			res, err := datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil(), "Create datavolume should succeed")
			Expect(res.Code).To(Equal(http.StatusCreated))
			datavolumeIds[res.Body.(api.Datavolume).Id] = TestTenant
		})

		It("create with permissions in project", func() {
			ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Projects: []string{projectId}})

			res, err := datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil(), "Create datavolume should succeed")
			Expect(res.Code).To(Equal(http.StatusCreated))
			datavolumeIds[res.Body.(api.Datavolume).Id] = TestTenant
		})

		It("create with invalid project", func() {
			request.ProjectId = "999"
			res, err := datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil(), "Create datavolume error should be nil")
			Expect(res.Code).To(Equal(http.StatusForbidden))
		})

		It("create with invalid tenant/subjectId", func() {
			ctx = context.Background()
			res, err := datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))

			ctx = context.WithValue(ctx, runaictx.CfTenantId, TestTenant)
			res, err = datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))

			ctx = context.Background()
			ctx = context.WithValue(ctx, runaictx.CfSubjectID, TestUser)
			res, err = datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
		})

		It("create with wrong tenantId+projectId combination", func() {
			request.ProjectId = "99" // an unexisting project
			res, err := datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusForbidden))
		})

		It("create with insufficient permissions", func() {
			ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Departments: []string{"1000"}})

			res, err := datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusForbidden))
		})

		It("create with invalid cluster uuid", func() {
			testProjectId := "99"
			SetProjectAncestorsCache(ctx, testProjectId, &authorization_utils.Ancestors{
				TenantId:     TestTenant,
				ClusterUuid:  ptr.To("2"), // invalid uuid
				DepartmentId: ptr.To("2"),
			})

			request.ProjectId = testProjectId
			res, err := datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))

			InvalidateProjectAncestorsCache(ctx, testProjectId)
		})

		It("create with conflict: multiple datavolumes with same name", func() {
			res, err := datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil(), "Create datavolume should succeed")
			Expect(res.Code).To(Equal(http.StatusCreated))
			created := res.Body.(api.Datavolume)
			datavolumeIds[created.Id] = TestTenant

			// in different cluster - should succeed (name constraint is only for the same cluster)
			request.ProjectId = "4"
			res, err = datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil(), "Create datavolume should succeed")
			Expect(res.Code).To(Equal(http.StatusCreated))
			created = res.Body.(api.Datavolume)
			datavolumeIds[created.Id] = TestTenant

			request.ProjectId = "2"
			request.OriginPvcName = "differentPvcName"
			// but same datavolume name...
			res, err = datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil(), "Create datavolume err should be nil")
			Expect(res.Code).To(Equal(http.StatusConflict))
			apiErr := res.Body.(api.Error)
			Expect(apiErr.Message).To(Equal("Failed to create datavolume: datavolume with the same name already exists"))
		})

		It("create with conflict: multiple datavolumes that point to same pvc", func() {
			res, err := datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil(), "Create datavolume should succeed")
			Expect(res.Code).To(Equal(http.StatusCreated))
			created := res.Body.(api.Datavolume)
			datavolumeIds[created.Id] = TestTenant

			request.Name = "different-datavolume-name"
			// but same pvc name and project...
			res, err = datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil(), "Create datavolume err should be nil")
			Expect(res.Code).To(Equal(http.StatusConflict))
			apiErr := res.Body.(api.Error)
			Expect(apiErr.Message).To(Equal("Failed to create datavolume: datavolume with the same project and pvc name already exists"))
		})

		It("create in old cluster - unsupported", func() {
			request.ProjectId = "200"
			res, err := datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
			apiErr := res.Body.(api.Error)
			Expect(apiErr.Message).To(ContainSubstring("need minimum version 2.18.0"))
		})
	})

	Context("delete datavolume", func() {
		var (
			projectId                                              string
			testDatavolumeId, testDatavolumeId1, testDatavolumeId2 string
		)

		BeforeEach(func() {
			projectId = "2"
			request := api.DatavolumeCreationFields{
				Name:          "dv-10",
				Description:   "Some description",
				ProjectId:     projectId,
				OriginPvcName: "pvc-10",
			}

			res, err := datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil(), "Create datavolume should succeed")

			Expect(res.Code).To(Equal(http.StatusCreated))
			testDatavolumeId = res.Body.(api.Datavolume).Id
			datavolumeIds[testDatavolumeId] = TestTenant

			// create another datavolume in a same tenant
			request.ProjectId = "1"
			request.Name = "dv-12"
			res, err = datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil(), "Create datavolume should succeed")
			Expect(res.Code).To(Equal(http.StatusCreated))
			testDatavolumeId1 = res.Body.(api.Datavolume).Id
			datavolumeIds[testDatavolumeId1] = TestTenant

			// create another datavolume in a different tenant
			request.ProjectId = projectIdInDifferentTenant
			request.OriginPvcName = "pvc-11"
			ctx2 := SetContext(TestTenant2, TestUser)
			res, err = datavolumesApiService.CreateDatavolume(ctx2, request)
			Expect(err).To(BeNil(), "Create datavolume should succeed")

			Expect(res.Code).To(Equal(http.StatusCreated))
			testDatavolumeId2 = res.Body.(api.Datavolume).Id
			datavolumeIds[testDatavolumeId2] = TestTenant2
		})

		It("delete sanity - marked as deleted", func() {
			// first - mark the datavolume as ready
			apiErr := dao.DbUpdateStatus(ctx, TestTenantUint, uuid.MustParse(testDatavolumeId), models.DatavolumeStatus{Phase: api.DATAVOLUMEPHASE_READY, PhaseMessage: "blabla"})
			Expect(apiErr).To(BeNil())

			// preform this twice - this should work twice in a row
			ctx = SetContext(TestTenant, TestUser2)
			now := time.Now()
			time.Sleep(1 * time.Millisecond)
			for i := 0; i < 2; i++ {
				res, err := datavolumesApiService.DeleteDatavolume(ctx, testDatavolumeId)
				Expect(err).To(BeNil(), "Delete datavolume should succeed")
				Expect(res.Code).To(Equal(http.StatusAccepted))

				res, err = datavolumesApiService.GetDatavolume(ctx, testDatavolumeId)
				Expect(err).To(BeNil(), "Get datavolume should succeed")
				Expect(res.Code).To(Equal(http.StatusOK))

				datavolume := res.Body.(api.Datavolume)
				Expect(datavolume.Status.Phase).To(Equal(api.DATAVOLUMEPHASE_DELETING))
				Expect(datavolume.Status.PhaseMessage).To(Equal(""))
				Expect(datavolume.UpdatedAt.After(datavolume.CreatedAt)).To(BeTrue(), fmt.Sprintf("UpdatedAt <%v> should be after CreatedAt <%v>", datavolume.UpdatedAt, datavolume.CreatedAt))
				Expect(datavolume.UpdatedAt.After(now)).To(BeTrue())
				Expect(datavolume.UpdatedBy).To(Equal(TestUser2))
				Expect(datavolume.CreatedBy).To(Equal(TestUser))
			}

			// validate the other DV was not deleted
			res, err := datavolumesApiService.GetDatavolume(ctx, testDatavolumeId1)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			datavolume := res.Body.(api.Datavolume)
			Expect(datavolume.Status.Phase).To(Equal(api.DATAVOLUMEPHASE_CREATING))
		})

		It("delete sanity - deleted cause was Creating", func() {
			res, err := datavolumesApiService.DeleteDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil(), "Delete datavolume should succeed")
			Expect(res.Code).To(Equal(http.StatusNoContent))

			res, err = datavolumesApiService.GetDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil(), "Get datavolume should succeed")
			Expect(res.Code).To(Equal(http.StatusNotFound))

			// validate the other DV was not deleted
			ctx2 := SetContext(TestTenant2, TestUser)
			res, err = datavolumesApiService.GetDatavolume(ctx2, testDatavolumeId2)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			datavolume := res.Body.(api.Datavolume)
			Expect(datavolume.Status.Phase).To(Equal(api.DATAVOLUMEPHASE_CREATING))
		})

		It("delete with invalid tenant/subjectId", func() {
			ctx = context.Background()
			res, err := datavolumesApiService.DeleteDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))

			ctx = context.WithValue(ctx, runaictx.CfTenantId, TestTenant)
			res, err = datavolumesApiService.DeleteDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))

			ctx = context.Background()
			ctx = context.WithValue(ctx, runaictx.CfSubjectID, TestUser)
			res, err = datavolumesApiService.DeleteDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
		})

		It("delete with invalid datavolume id", func() {
			res, err := datavolumesApiService.DeleteDatavolume(ctx, "999")
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
		})

		It("delete with wrong tenantId+projectId combination", func() {
			res, err := datavolumesApiService.DeleteDatavolume(ctx, testDatavolumeId2)
			Expect(err).To(BeNil())
			// NotFound because the datavolume is in a different tenant,
			// this returns from the DB when trying to get before the DELETE attempt.
			Expect(res.Code).To(Equal(http.StatusNotFound))

			// sanity to validate deletion succeeds for correct tenant
			ctx2 := SetContext(TestTenant2, TestUser)
			res, err = datavolumesApiService.DeleteDatavolume(ctx2, testDatavolumeId2)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNoContent))
		})

		It("delete with insufficient permissions to delete shared datavolumes", func() {
			// add a shared scope
			sharedScopesPatch := api.SharedScopesPatchRequest{
				Add: []api.SharedScopeBase{
					{
						ScopeId:   "1",
						ScopeType: api.SCOPETYPE_PROJECT,
					},
				},
			}
			res, err := datavolumesApiService.PatchDatavolumeSharedScopes(ctx, testDatavolumeId, sharedScopesPatch)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))

			// now try to delete as user 2 - should fail (auth mock is set this way)
			ctx = SetContext(TestTenant, TestUser2)
			res, err = datavolumesApiService.DeleteDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusForbidden))

			// now try to delete as user 1 - should succeed (auth mock is set this way)
			ctx = SetContext(TestTenant, TestUser)
			res, err = datavolumesApiService.DeleteDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNoContent))
		})

		It("delete with insufficient permissions to delete shared datavolumes - succeed cause empty shared scopes", func() {
			// don't add a shared scope

			// try to delete as user 2 - should succeed, cause sharedscopes list is empty
			ctx = SetContext(TestTenant, TestUser2)
			res, err := datavolumesApiService.DeleteDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNoContent))
		})
	})

	Context("update status", func() {
		var (
			projectId        string
			testDatavolumeId string
			request          api.DatavolumeCreationFields
		)

		BeforeEach(func() {
			projectId = "1"
			request = api.DatavolumeCreationFields{
				Name:          "dv-50",
				Description:   "some description 50",
				ProjectId:     projectId,
				OriginPvcName: "pvc-50",
			}

			res, err := datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil(), "Create datavolume should succeed")

			Expect(res.Code).To(Equal(http.StatusCreated))
			testDatavolumeId = res.Body.(api.Datavolume).Id
			datavolumeIds[testDatavolumeId] = TestTenant
		})

		It("update status sanity", func() {
			internalStatusRequest := api.DatavolumeStatusInternal{
				Phase:        api.DATAVOLUMEPHASE_READY,
				PhaseMessage: "some message",
				Conditions: []api.Condition{
					{Type: "some type", Status: "some status", Reason: "some reason", Message: "some message", LastTransitionTime: ptr.To(time.Now())},
					{Type: "some type 2", Status: "some status 2", Reason: "some reason 2", Message: "some message 2", LastTransitionTime: ptr.To(time.Now())},
				},
				DatavolumePvcName: "some-pvc-name",
				DatavolumePvName:  "some-pv-name",
			}

			res, err := datavolumesApiService.UpdateDatavolumeStatus(ctx, testDatavolumeId, internalStatusRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNoContent))

			res, err = datavolumesApiService.GetDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			datavolume := res.Body.(api.Datavolume)
			validateStatus(&internalStatusRequest, &datavolume)

			// validate other fields were not updated
			validateDatavolume2(datavolume, request, projectToAncestors)
		})

		It("update status with invalid tenant id", func() {
			internalStatusRequest := api.DatavolumeStatusInternal{
				Phase: api.DATAVOLUMEPHASE_READY,
			}

			ctx = context.WithValue(ctx, runaictx.CfTenantId, "blabla")
			res, err := datavolumesApiService.UpdateDatavolumeStatus(ctx, testDatavolumeId, internalStatusRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))

			ctx = context.WithValue(ctx, runaictx.CfTenantId, nil)
			res, err = datavolumesApiService.UpdateDatavolumeStatus(ctx, testDatavolumeId, internalStatusRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))

			ctx = context.WithValue(ctx, runaictx.CfTenantId, "999")
			res, err = datavolumesApiService.UpdateDatavolumeStatus(ctx, testDatavolumeId, internalStatusRequest)
			Expect(err).To(BeNil())
			// deleting a row that "doesn't exist" doesn't return error
			Expect(res.Code).To(Equal(http.StatusNoContent))

			// validate exists - sanity
			ctx = SetContext(TestTenant, TestUser)
			res, err = datavolumesApiService.GetDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			datavolume := res.Body.(api.Datavolume)
			Expect(datavolume.Status.Phase).To(Equal(api.DATAVOLUMEPHASE_CREATING))
		})

		It("update status with invalid datavolume id", func() {
			internalStatusRequest := api.DatavolumeStatusInternal{
				Phase: api.DATAVOLUMEPHASE_READY,
			}

			res, err := datavolumesApiService.UpdateDatavolumeStatus(ctx, "999", internalStatusRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
		})

		It("update status - delete - should succeed cause phase was Creating", func() {
			internalStatusRequest := api.DatavolumeStatusInternal{
				Phase:        api.DATAVOLUMEPHASE_READY,
				PhaseMessage: "some message",
				IsDeleted:    ptr.To(true),
			}

			res, err := datavolumesApiService.UpdateDatavolumeStatus(ctx, testDatavolumeId, internalStatusRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNoContent))

			res, err = datavolumesApiService.GetDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNotFound))
		})

		It("update status - delete - should succeed cause phase was Deleting", func() {
			// first, prepare the datavolume to be in "Deleting" phase
			apiErr := dao.DbUpdateStatus(ctx, TestTenantUint, uuid.MustParse(testDatavolumeId), models.DatavolumeStatus{Phase: api.DATAVOLUMEPHASE_READY})
			Expect(apiErr).To(BeNil())
			res, err := datavolumesApiService.DeleteDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusAccepted))

			internalStatusRequest := api.DatavolumeStatusInternal{
				Phase:        api.DATAVOLUMEPHASE_READY,
				PhaseMessage: "some message",
				IsDeleted:    ptr.To(true),
			}

			res, err = datavolumesApiService.UpdateDatavolumeStatus(ctx, testDatavolumeId, internalStatusRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNoContent))

			res, err = datavolumesApiService.GetDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNotFound))
		})

		It("update status - delete - should fail cause phase was NotReady", func() {
			apiErr := dao.DbUpdateStatus(ctx, TestTenantUint, uuid.MustParse(testDatavolumeId), models.DatavolumeStatus{Phase: api.DATAVOLUMEPHASE_NOT_READY})
			Expect(apiErr).To(BeNil())

			internalStatusRequest := api.DatavolumeStatusInternal{
				Phase:        api.DATAVOLUMEPHASE_READY,
				PhaseMessage: "some message",
				IsDeleted:    ptr.To(true),
			}

			res, err := datavolumesApiService.UpdateDatavolumeStatus(ctx, testDatavolumeId, internalStatusRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNoContent))

			res, err = datavolumesApiService.GetDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			dataVolume := res.Body.(api.Datavolume)
			Expect(dataVolume.Status.Phase).To(Equal(api.DATAVOLUMEPHASE_NOT_READY))
		})

		It("update status to deleting - should succeed cause was already deleting", func() {
			// first, prepare the datavolume to be in "Deleting" phase
			apiErr := dao.DbUpdateStatus(ctx, TestTenantUint, uuid.MustParse(testDatavolumeId), models.DatavolumeStatus{Phase: api.DATAVOLUMEPHASE_READY})
			Expect(apiErr).To(BeNil())
			res, err := datavolumesApiService.DeleteDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusAccepted))

			internalStatusRequest := api.DatavolumeStatusInternal{
				Phase:        api.DATAVOLUMEPHASE_DELETING,
				PhaseMessage: "datavolume is being deleted in the cluster",
			}

			res, err = datavolumesApiService.UpdateDatavolumeStatus(ctx, testDatavolumeId, internalStatusRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNoContent))
			res, err = datavolumesApiService.GetDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			datavolume := res.Body.(api.Datavolume)
			Expect(datavolume.Status.Phase).To(Equal(api.DATAVOLUMEPHASE_DELETING))
			Expect(datavolume.Status.PhaseMessage).To(Equal("datavolume is being deleted in the cluster"))
		})

		It("update status to deleting - should fail cause was not deleting before", func() {
			apiErr := dao.DbUpdateStatus(ctx, TestTenantUint, uuid.MustParse(testDatavolumeId), models.DatavolumeStatus{Phase: api.DATAVOLUMEPHASE_READY})
			Expect(apiErr).To(BeNil())

			internalStatusRequest := api.DatavolumeStatusInternal{
				Phase:        api.DATAVOLUMEPHASE_DELETING,
				PhaseMessage: "some message",
			}

			res, err := datavolumesApiService.UpdateDatavolumeStatus(ctx, testDatavolumeId, internalStatusRequest)
			Expect(err).To(BeNil())
			// there's no error if update "not found"
			Expect(res.Code).To(Equal(http.StatusNoContent))
			res, err = datavolumesApiService.GetDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			datavolume := res.Body.(api.Datavolume)
			Expect(datavolume.Status.Phase).To(Equal(api.DATAVOLUMEPHASE_READY))
		})
	})

	Context("sync statuses", func() {
		var (
			testDatavolumes        = []api.Datavolume{}
			datavolumesSyncRequest = api.DatavolumesSyncRequest{
				Datavolumes: []api.DatavolumeStatusForSync{},
				ClusterId:   clusterId1,
			}
			sharedScopesPatch = api.SharedScopesPatchRequest{
				Add: []api.SharedScopeBase{
					{ScopeId: "3", ScopeType: api.SCOPETYPE_DEPARTMENT},
					{ScopeId: "2", ScopeType: api.SCOPETYPE_PROJECT},
				},
			}
			datavolumeStatusForSync = api.DatavolumeStatusForSync{
				Phase:        api.DATAVOLUMEPHASE_READY,
				PhaseMessage: "some message",
				Conditions: []api.Condition{
					{Type: "some type", Status: "some status", Reason: "some reason", Message: "some message", LastTransitionTime: ptr.To(time.Now())},
				},
				DatavolumePvcName: "some-pvc-name",
				DatavolumePvName:  "some-pv-name",
			}
		)

		BeforeEach(func() {
			datavolumesSyncRequest = api.DatavolumesSyncRequest{
				Datavolumes: []api.DatavolumeStatusForSync{},
				ClusterId:   clusterId1,
			}
			testDatavolumes = []api.Datavolume{}

			request := api.DatavolumeCreationFields{
				Name:        "dv-0",
				Description: "bla bla",
			}

			for i := 10; i <= 14; i++ {
				request.ProjectId = fmt.Sprintf("%d", i)
				for j := 0; j < 2; j++ {
					// create 2 in each project
					request.Name = fmt.Sprintf("dv-%d-%d", i, j)
					request.OriginPvcName = fmt.Sprintf("pvc-%d-%d", i, j)

					res, err := datavolumesApiService.CreateDatavolume(ctx, request)
					Expect(err).To(BeNil(), "Create datavolume should succeed")
					Expect(res.Code).To(Equal(http.StatusCreated))
					dv := res.Body.(api.Datavolume)
					datavolumeIds[dv.Id] = TestTenant

					if j == 1 {
						// add shared scopes
						res, err := datavolumesApiService.PatchDatavolumeSharedScopes(ctx, dv.Id, sharedScopesPatch)
						Expect(err).To(BeNil())
						Expect(res.Code).To(Equal(http.StatusOK))

						res, err = datavolumesApiService.GetDatavolumeSharedScopes(ctx, dv.Id)
						Expect(err).To(BeNil())
						Expect(res.Code).To(Equal(http.StatusOK))
						dv.SharedScopes = res.Body.([]api.SharedScope)
					}

					testDatavolumes = append(testDatavolumes, dv)
				}
			}
		})

		It("sanity", func() {
			for _, testDv := range testDatavolumes {
				datavolumeStatusForSync.DatavolumeId = testDv.Id
				datavolumesSyncRequest.Datavolumes = append(datavolumesSyncRequest.Datavolumes, datavolumeStatusForSync)
			}

			res, err := datavolumesApiService.SyncDatavolumes(ctx, datavolumesSyncRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNoContent))

			for i, testDv := range testDatavolumes {
				res, err := datavolumesApiService.GetDatavolume(ctx, testDv.Id)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				datavolume := res.Body.(api.Datavolume)
				validateStatus(datavolumeStatusToInternal(&datavolumesSyncRequest.Datavolumes[i]), &datavolume)
				// validate other fields were not updated
				validateDatavolume2(datavolume, datavolumeToRequest(testDv), projectToAncestors)

				// validate shared scopes
				res, err = datavolumesApiService.GetDatavolumeSharedScopes(ctx, testDv.Id)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				sharedScopes := res.Body.([]api.SharedScope)
				if i%2 == 1 {
					validateSharedScopes(sharedScopesPatch.Add, sharedScopes, time.Time{})
				} else {
					Expect(sharedScopes).To(HaveLen(0))
				}
			}
		})

		It("sync with missing items - validate deleted", func() {
			// update the last DV to "Ready" so the sync without it should not delete it
			res, err := datavolumesApiService.UpdateDatavolumeStatus(ctx, testDatavolumes[len(testDatavolumes)-1].Id, api.DatavolumeStatusInternal{Phase: api.DATAVOLUMEPHASE_READY})
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNoContent))

			// only include the first DV in the sync request
			datavolumeStatusForSync.DatavolumeId = testDatavolumes[0].Id
			datavolumesSyncRequest.Datavolumes = append(datavolumesSyncRequest.Datavolumes, datavolumeStatusForSync)

			res, err = datavolumesApiService.SyncDatavolumes(ctx, datavolumesSyncRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNoContent))

			// first DV should exist
			res, err = datavolumesApiService.GetDatavolume(ctx, testDatavolumes[0].Id)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			datavolume := res.Body.(api.Datavolume)
			validateStatus(datavolumeStatusToInternal(&datavolumesSyncRequest.Datavolumes[0]), &datavolume)
			// validate other fields were not updated
			validateDatavolume2(datavolume, datavolumeToRequest(testDatavolumes[0]), projectToAncestors)

			for i := 1; i < len(testDatavolumes)-1; i++ {
				// all others should be deleted because the DVs were "Creating" before
				res, err := datavolumesApiService.GetDatavolume(ctx, testDatavolumes[i].Id)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusNotFound))
			}

			// last DV should exist
			res, err = datavolumesApiService.GetDatavolume(ctx, testDatavolumes[len(testDatavolumes)-1].Id)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			datavolume = res.Body.(api.Datavolume)
			Expect(datavolume.Status.Phase).To(Equal(api.DATAVOLUMEPHASE_READY))
		})

		It("sync with unexisting datavolumes in request", func() {
			unexisintgId := uuid.NewString()

			datavolumeStatusForSync.DatavolumeId = testDatavolumes[1].Id
			datavolumesSyncRequest.Datavolumes = append(datavolumesSyncRequest.Datavolumes, datavolumeStatusForSync)

			datavolumeStatusForSync.DatavolumeId = unexisintgId
			datavolumesSyncRequest.Datavolumes = append(datavolumesSyncRequest.Datavolumes, datavolumeStatusForSync)

			res, err := datavolumesApiService.SyncDatavolumes(ctx, datavolumesSyncRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNoContent))

			res, err = datavolumesApiService.GetDatavolume(ctx, unexisintgId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNotFound))

			res, err = datavolumesApiService.GetDatavolume(ctx, testDatavolumes[1].Id)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))

			res, err = datavolumesApiService.GetDatavolume(ctx, testDatavolumes[0].Id)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNotFound))

			wrongUuid := "not-an-uuid"
			datavolumeStatusForSync.DatavolumeId = wrongUuid
			datavolumesSyncRequest.Datavolumes = append(datavolumesSyncRequest.Datavolumes, datavolumeStatusForSync)
			res, err = datavolumesApiService.SyncDatavolumes(ctx, datavolumesSyncRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
		})

		It("sync with from and/or to", func() {
			sortDatavolumesForTest(testDatavolumes)

			// sync only the first 2 datavolumes, with from 0 and to the 4th
			for i := 0; i < 2; i++ {
				datavolumeStatusForSync.DatavolumeId = testDatavolumes[i].Id
				datavolumesSyncRequest.Datavolumes = append(datavolumesSyncRequest.Datavolumes, datavolumeStatusForSync)
			}

			datavolumesSyncRequest.From = testDatavolumes[0].Id
			datavolumesSyncRequest.To = testDatavolumes[3].Id

			res, err := datavolumesApiService.SyncDatavolumes(ctx, datavolumesSyncRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNoContent))

			for i, testDv := range testDatavolumes {
				res, err := datavolumesApiService.GetDatavolume(ctx, testDv.Id)
				Expect(err).To(BeNil())

				if i < 2 || i > 3 {
					Expect(res.Code).To(Equal(http.StatusOK))
					datavolume := res.Body.(api.Datavolume)

					if i < 2 {
						validateStatus(datavolumeStatusToInternal(&datavolumesSyncRequest.Datavolumes[i]), &datavolume)
					} else {
						validateStatus(&api.DatavolumeStatusInternal{Phase: api.DATAVOLUMEPHASE_CREATING, PhaseMessage: "Creating datavolume in the cluster"}, &datavolume)
					}
				} else {
					// 2 and 3 should be deleted, as they were not synced and they are in the from/to range
					Expect(res.Code).To(Equal(http.StatusNotFound))
				}
			}
		})

		It("sync with insufficient permissions", func() {
			ctx = SetContext(TestTenant2, TestUser)

			datavolumeStatusForSync.DatavolumeId = testDatavolumes[0].Id
			datavolumesSyncRequest.Datavolumes = append(datavolumesSyncRequest.Datavolumes, datavolumeStatusForSync)

			res, err := datavolumesApiService.SyncDatavolumes(ctx, datavolumesSyncRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNoContent))

			// the sync "succeeded" but basically should have updated nothing - the clusterId and tenantId don't match
			ctx = SetContext(TestTenant, TestUser)
			for _, testDv := range testDatavolumes {
				res, err := datavolumesApiService.GetDatavolume(ctx, testDv.Id)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				datavolume := res.Body.(api.Datavolume)
				Expect(datavolume.Status.Phase).To(Equal(api.DATAVOLUMEPHASE_CREATING))
			}
		})

		It("sync with no datavolumes", func() {
			res, err := datavolumesApiService.SyncDatavolumes(ctx, datavolumesSyncRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNoContent))

			for _, testDv := range testDatavolumes {
				res, err := datavolumesApiService.GetDatavolume(ctx, testDv.Id)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusNotFound))
			}
		})
	})

	Context("patch/get sharedscopes", func() {
		var (
			projectId        string
			testDatavolumeId string
		)

		BeforeEach(func() {
			projectId = "2"
			request := api.DatavolumeCreationFields{
				Name:          "dv-10",
				Description:   "Some description",
				ProjectId:     projectId,
				OriginPvcName: "pvc-10",
			}

			res, err := datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil(), "Create datavolume should succeed")

			Expect(res.Code).To(Equal(http.StatusCreated))
			testDatavolumeId = res.Body.(api.Datavolume).Id
			datavolumeIds[testDatavolumeId] = TestTenant
		})

		It("get sharedscopes sanity with empty sharedscopes list", func() {
			res, err := datavolumesApiService.GetDatavolumeSharedScopes(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			sharedScopes := res.Body.([]api.SharedScope)
			Expect(sharedScopes).To(HaveLen(0))

			res, err = datavolumesApiService.GetDatavolumeSharedScopes(ctx, uuid.NewString())
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNotFound))

			res, err = datavolumesApiService.GetDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			datavolumeSharedScopes := res.Body.(api.Datavolume).SharedScopes
			Expect(datavolumeSharedScopes).To(HaveLen(0))
		})

		It("patch+get sharedscopes sanity", func() {
			expectedSharedScopes := []api.SharedScopeBase{
				{
					ScopeId:   "3",
					ScopeType: api.SCOPETYPE_DEPARTMENT,
				},
				{
					ScopeId:   "3",
					ScopeType: api.SCOPETYPE_PROJECT,
				},
			}

			sharedScopesPatch := api.SharedScopesPatchRequest{Add: expectedSharedScopes[:]}

			now := time.Now()

			res, err := datavolumesApiService.PatchDatavolumeSharedScopes(ctx, testDatavolumeId, sharedScopesPatch)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			sharedScopes := res.Body.([]api.SharedScope)
			validateSharedScopes(expectedSharedScopes, sharedScopes, now)

			res, err = datavolumesApiService.GetDatavolumeSharedScopes(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			sharedScopes = res.Body.([]api.SharedScope)
			validateSharedScopes(expectedSharedScopes, sharedScopes, now)

			res, err = datavolumesApiService.GetDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			datavolumeSharedScopes := res.Body.(api.Datavolume).SharedScopes
			validateSharedScopes(expectedSharedScopes, datavolumeSharedScopes, now)

			// get straight from DB and validate ancestors
			dbDatavolume, apiErr := dao.DbGetDatavolumeById(ctx, TestTenantUint, uuid.MustParse(testDatavolumeId), true)
			dbSharedScopes := dbDatavolume.SharedScopes
			Expect(apiErr).To(BeNil())
			Expect(dbSharedScopes).To(HaveLen(2))
			if dbSharedScopes[0].ProjectId != "" {
				dbSharedScopes[0], dbSharedScopes[1] = dbSharedScopes[1], dbSharedScopes[0]
			}

			Expect(dbSharedScopes[0].ProjectId).To(Equal(""))
			Expect(dbSharedScopes[0].ScopeName).To(Equal(getTestScopeName(expectedSharedScopes[0].ScopeId, expectedSharedScopes[0].ScopeType)))
			Expect(dbSharedScopes[0].DepartmentId).To(Equal(expectedSharedScopes[0].ScopeId))
			Expect(dbSharedScopes[1].ProjectId).To(Equal(expectedSharedScopes[1].ScopeId))
			Expect(dbSharedScopes[1].ScopeName).To(Equal(getTestScopeName(expectedSharedScopes[1].ScopeId, expectedSharedScopes[1].ScopeType)))
			Expect(dbSharedScopes[1].DepartmentId).To(Equal(projectToAncestors[expectedSharedScopes[1].ScopeId][0]))

			time.Sleep(10 * time.Millisecond)
			now2 := time.Now()
			time.Sleep(10 * time.Millisecond)

			// add another scope and validate the previous ones' CreatedAt is unchanged
			sharedScopesPatch.Add = []api.SharedScopeBase{{ScopeId: clusterId1, ScopeType: api.SCOPETYPE_CLUSTER}}
			expectedSharedScopes = append(expectedSharedScopes, sharedScopesPatch.Add[0])
			res, err = datavolumesApiService.PatchDatavolumeSharedScopes(ctx, testDatavolumeId, sharedScopesPatch)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			sharedScopes = res.Body.([]api.SharedScope)
			validateSharedScopes(expectedSharedScopes, sharedScopes, now)
			for _, sharedScope := range sharedScopes {
				if sharedScope.ScopeId == clusterId1 {
					Expect(sharedScope.CreatedAt.UnixMilli()).To(BeNumerically(">", now2.UnixMilli()))
				} else {
					Expect(sharedScope.CreatedAt.UnixMilli()).To(BeNumerically("<", now2.UnixMilli()))
				}
			}

			// remove a scope
			sharedScopesPatch.Remove = []api.SharedScopeBase{expectedSharedScopes[1]}
			expectedSharedScopes = []api.SharedScopeBase{expectedSharedScopes[2], expectedSharedScopes[0]}
			res, err = datavolumesApiService.PatchDatavolumeSharedScopes(ctx, testDatavolumeId, sharedScopesPatch)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			sharedScopes = res.Body.([]api.SharedScope)
			validateSharedScopes(expectedSharedScopes, sharedScopes, now)
			for _, sharedScope := range sharedScopes {
				if sharedScope.ScopeId == clusterId1 {
					Expect(sharedScope.CreatedAt.UnixMilli()).To(BeNumerically(">", now2.UnixMilli()))
				} else {
					Expect(sharedScope.CreatedAt.UnixMilli()).To(BeNumerically("<", now2.UnixMilli()))
				}
			}

			res, err = datavolumesApiService.GetDatavolumeSharedScopes(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			sharedScopes = res.Body.([]api.SharedScope)
			validateSharedScopes(expectedSharedScopes, sharedScopes, now)

			// test add + remove together
			sharedScopesPatch.Add = []api.SharedScopeBase{{ScopeId: "1", ScopeType: api.SCOPETYPE_PROJECT}}
			sharedScopesPatch.Remove = []api.SharedScopeBase{expectedSharedScopes[1]}
			expectedSharedScopes = []api.SharedScopeBase{expectedSharedScopes[0], sharedScopesPatch.Add[0]}
			res, err = datavolumesApiService.PatchDatavolumeSharedScopes(ctx, testDatavolumeId, sharedScopesPatch)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			sharedScopes = res.Body.([]api.SharedScope)
			validateSharedScopes(expectedSharedScopes, sharedScopes, now2)

			res, err = datavolumesApiService.GetDatavolumeSharedScopes(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			sharedScopes = res.Body.([]api.SharedScope)
			validateSharedScopes(expectedSharedScopes, sharedScopes, now2)
		})

		It("patch add with scope in different cluster (same as doesn't exist)", func() {
			sharedScopesPatch := api.SharedScopesPatchRequest{Add: []api.SharedScopeBase{{ScopeId: clusterId3, ScopeType: api.SCOPETYPE_CLUSTER}}}

			res, err := datavolumesApiService.PatchDatavolumeSharedScopes(ctx, testDatavolumeId, sharedScopesPatch)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))

			// department "21" in cluster1 doesn't exist (we have department "21" in the other cluster..)
			sharedScopesPatch = api.SharedScopesPatchRequest{Add: []api.SharedScopeBase{{ScopeId: "21", ScopeType: api.SCOPETYPE_DEPARTMENT}}}
			res, err = datavolumesApiService.PatchDatavolumeSharedScopes(ctx, testDatavolumeId, sharedScopesPatch)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))

			// project "4" in cluster1 doesn't exist (we have project "4" in the other cluster..)
			sharedScopesPatch = api.SharedScopesPatchRequest{Add: []api.SharedScopeBase{{ScopeId: "4", ScopeType: api.SCOPETYPE_PROJECT}}}
			res, err = datavolumesApiService.PatchDatavolumeSharedScopes(ctx, testDatavolumeId, sharedScopesPatch)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
		})

		It("patch add with project scope that is the original project - error", func() {
			sharedScopesPatch := api.SharedScopesPatchRequest{Add: []api.SharedScopeBase{{ScopeId: projectId, ScopeType: api.SCOPETYPE_PROJECT}}}

			res, err := datavolumesApiService.PatchDatavolumeSharedScopes(ctx, testDatavolumeId, sharedScopesPatch)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
			apiErr := res.Body.(api.Error)
			Expect(apiErr.Message).To(ContainSubstring("Cannot share with original project"))
		})

		It("patch add with conflict - multiple sharedscopes with same scopeId and scopeType", func() {
			// first, sanity just so we have a scope
			sharedScopesPatch := api.SharedScopesPatchRequest{Add: []api.SharedScopeBase{{ScopeId: "1", ScopeType: api.SCOPETYPE_PROJECT}}}
			res, err := datavolumesApiService.PatchDatavolumeSharedScopes(ctx, testDatavolumeId, sharedScopesPatch)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))

			sharedScopesPatch.Add = []api.SharedScopeBase{
				{
					ScopeId:   "1",
					ScopeType: api.SCOPETYPE_PROJECT,
				},
				{
					ScopeId:   "1",
					ScopeType: api.SCOPETYPE_PROJECT,
				},
			}

			res, err = datavolumesApiService.PatchDatavolumeSharedScopes(ctx, testDatavolumeId, sharedScopesPatch)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusConflict))
			apiErr := res.Body.(api.Error)
			Expect(apiErr.Message).To(Equal("Duplicate shared scopes in requested Add"))

			sharedScopesPatch.Add = []api.SharedScopeBase{
				{
					ScopeId:   "1",
					ScopeType: api.SCOPETYPE_PROJECT,
				},
				{
					ScopeId:   "3",
					ScopeType: api.SCOPETYPE_DEPARTMENT,
				},
				{
					ScopeId:   "3",
					ScopeType: api.SCOPETYPE_DEPARTMENT,
				},
			}

			res, err = datavolumesApiService.PatchDatavolumeSharedScopes(ctx, testDatavolumeId, sharedScopesPatch)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusConflict))
			apiErr = res.Body.(api.Error)
			Expect(apiErr.Message).To(Equal("Duplicate shared scopes in requested Add"))

			sharedScopesPatch.Add = []api.SharedScopeBase{}
			sharedScopesPatch.Remove = []api.SharedScopeBase{
				{
					ScopeId:   "1",
					ScopeType: api.SCOPETYPE_PROJECT,
				},
				{
					ScopeId:   "1",
					ScopeType: api.SCOPETYPE_PROJECT,
				},
			}
			res, err = datavolumesApiService.PatchDatavolumeSharedScopes(ctx, testDatavolumeId, sharedScopesPatch)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusConflict))
			apiErr = res.Body.(api.Error)
			Expect(apiErr.Message).To(Equal("Duplicate shared scopes in requested Remove"))
		})

		Context("test patch/get shared scopes with permitted scopes", func() {
			scopeInDv := api.SharedScopeBase{ScopeId: "3", ScopeType: api.SCOPETYPE_DEPARTMENT}

			BeforeEach(func() {
				// add a scope
				sharedScopesPatch := api.SharedScopesPatchRequest{Add: []api.SharedScopeBase{scopeInDv}}
				res, err := datavolumesApiService.PatchDatavolumeSharedScopes(ctx, testDatavolumeId, sharedScopesPatch)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
			})

			It("patch add scope that is not permitted", func() {
				ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Projects: []string{"9"}})

				// first, add a permitted scope - should succeed
				sharedScopesPatch := api.SharedScopesPatchRequest{Add: []api.SharedScopeBase{{ScopeId: "9", ScopeType: api.SCOPETYPE_PROJECT}}}
				res, err := datavolumesApiService.PatchDatavolumeSharedScopes(ctx, testDatavolumeId, sharedScopesPatch)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				sharedScopes := res.Body.([]api.SharedScope)
				validateSharedScopes(sharedScopesPatch.Add, sharedScopes, time.Time{})

				// now try to add a forbidden scope
				sharedScopesPatch = api.SharedScopesPatchRequest{Add: []api.SharedScopeBase{{ScopeId: "1", ScopeType: api.SCOPETYPE_DEPARTMENT}}}
				res, err = datavolumesApiService.PatchDatavolumeSharedScopes(ctx, testDatavolumeId, sharedScopesPatch)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusForbidden))

				// now try to remove a forbidden scope
				sharedScopesPatch = api.SharedScopesPatchRequest{Remove: []api.SharedScopeBase{scopeInDv}, Add: []api.SharedScopeBase{}}
				res, err = datavolumesApiService.PatchDatavolumeSharedScopes(ctx, testDatavolumeId, sharedScopesPatch)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusForbidden))

				// now try to "add" an existing scope but one that is forbidden
				sharedScopesPatch = api.SharedScopesPatchRequest{Add: []api.SharedScopeBase{scopeInDv}, Remove: []api.SharedScopeBase{}}
				res, err = datavolumesApiService.PatchDatavolumeSharedScopes(ctx, testDatavolumeId, sharedScopesPatch)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusForbidden))
			})
		})
	})

	Context("patch datavolume", func() {
		var (
			projectId        string
			testDatavolumeId string
			request          api.DatavolumeCreationFields
			patchRequest     api.DatavolumePatchFields
		)

		BeforeEach(func() {
			projectId = "2"
			request = api.DatavolumeCreationFields{
				Name:          "dv-10",
				Description:   "Some description",
				ProjectId:     projectId,
				OriginPvcName: "pvc-10",
			}

			res, err := datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil(), "Create datavolume should succeed")

			Expect(res.Code).To(Equal(http.StatusCreated))
			testDatavolumeId = res.Body.(api.Datavolume).Id
			datavolumeIds[testDatavolumeId] = TestTenant

			patchRequest = api.DatavolumePatchFields{
				Description: "new-description",
			}
		})

		It("sanity", func() {
			time.Sleep(1 * time.Millisecond)
			now := time.Now()
			time.Sleep(1 * time.Millisecond)

			ctx = SetContext(TestTenant, TestUser2)

			patches := []api.DatavolumePatchFields{
				{Description: "new-description"},
				{Description: ""},
			}

			for _, patch := range patches {
				for i := 0; i < 2; i++ {
					res, err := datavolumesApiService.PatchDatavolume(ctx, testDatavolumeId, patch)
					Expect(err).To(BeNil())
					Expect(res.Code).To(Equal(http.StatusOK))
					datavolume := res.Body.(api.DatavolumeNoSharedScopes)
					// validate updated fields
					Expect(datavolume.Description).To(Equal(patch.Description))
					Expect(datavolume.UpdatedBy).To(Equal(TestUser2))
					Expect(datavolume.UpdatedAt.UnixMilli()).To(BeNumerically(">", now.UnixMilli()))

					// validate rest of fields weren't changed
					Expect(datavolume.Name).To(Equal(request.Name))
					Expect(datavolume.CreatedBy).To(Equal(TestUser))
					Expect(datavolume.CreatedAt.UnixMilli()).To(BeNumerically("<", now.UnixMilli()))
				}
			}
		})

		It("patch with invalid id", func() {
			res, err := datavolumesApiService.PatchDatavolume(ctx, "999", patchRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("get datavolumes", func() {
		var (
			testDatavolumes                     = []api.Datavolume{}
			testDatavolumesMap                  = map[string]api.Datavolume{}
			testDatavolumeSharedScopesPutFields = map[string][]api.SharedScopeBase{}
			request                             api.DatavolumeCreationFields
		)

		BeforeEach(func() {
			testDatavolumes = []api.Datavolume{}
			testDatavolumesMap = map[string]api.Datavolume{}
			testDatavolumeSharedScopesPutFields = map[string][]api.SharedScopeBase{}

			request = api.DatavolumeCreationFields{
				Name:        "dv-0",
				Description: "bla bla",
			}

			for i := 10; i <= 14; i++ {
				request.ProjectId = fmt.Sprintf("%d", i)
				for j := 0; j < 2; j++ {
					// create 2 in each project
					request.Name = fmt.Sprintf("dv-%d-%d", i, j)
					request.OriginPvcName = fmt.Sprintf("pvc-%d-%d", i, j)

					res, err := datavolumesApiService.CreateDatavolume(ctx, request)
					Expect(err).To(BeNil(), "Create datavolume should succeed")
					Expect(res.Code).To(Equal(http.StatusCreated))
					dv := res.Body.(api.Datavolume)
					datavolumeIds[dv.Id] = TestTenant

					// add shared scopes
					depIdToShareWith := "4"
					if j%2 == 0 {
						depIdToShareWith = "3"
					}
					sharedScopesToAdd := []api.SharedScopeBase{
						{ScopeId: depIdToShareWith, ScopeType: api.SCOPETYPE_DEPARTMENT},
						{ScopeId: fmt.Sprintf("%d", i-1), ScopeType: api.SCOPETYPE_PROJECT},
					}
					if i == 13 && j == 1 {
						sharedScopesToAdd = append(sharedScopesToAdd, api.SharedScopeBase{ScopeId: clusterId1, ScopeType: api.SCOPETYPE_CLUSTER})
					}

					sharedScopesPatch := api.SharedScopesPatchRequest{Add: sharedScopesToAdd}
					res, err = datavolumesApiService.PatchDatavolumeSharedScopes(ctx, dv.Id, sharedScopesPatch)
					Expect(err).To(BeNil())
					Expect(res.Code).To(Equal(http.StatusOK))

					res, err = datavolumesApiService.GetDatavolumeSharedScopes(ctx, dv.Id)
					Expect(err).To(BeNil())
					Expect(res.Code).To(Equal(http.StatusOK))

					dv.SharedScopes = res.Body.([]api.SharedScope)
					testDatavolumes = append(testDatavolumes, dv)
					testDatavolumesMap[dv.Id] = dv
					testDatavolumeSharedScopesPutFields[dv.Id] = sharedScopesToAdd

					// update status to ready
					internalStatusRequest := api.DatavolumeStatusInternal{
						Phase: api.DATAVOLUMEPHASE_READY,
					}
					res, err = datavolumesApiService.UpdateDatavolumeStatus(ctx, dv.Id, internalStatusRequest)
					Expect(err).To(BeNil())
					Expect(res.Code).To(Equal(http.StatusNoContent))
				}
			}

			// and create another one in a different cluster
			request.ProjectId = "5"
			request.Name = "dv-second-cluster"
			res, err := datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil(), "Create datavolume should succeed")
			Expect(res.Code).To(Equal(http.StatusCreated))
			dv := res.Body.(api.Datavolume)
			datavolumeIds[dv.Id] = TestTenant
			testDatavolumes = append(testDatavolumes, dv)
			testDatavolumesMap[dv.Id] = dv
			testDatavolumeSharedScopesPutFields[dv.Id] = []api.SharedScopeBase{}
		})

		It("sanity - get all datavolumes", func() {
			res, err := datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", 0, 1000, "", "", []string{}, "", false)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			response := res.Body.(api.GetDatavolumes200Response)
			datavolumes := response.Datavolumes
			Expect(datavolumes).To(HaveLen(len(testDatavolumes)))
			for _, dv := range datavolumes {
				expectedDv, ok := testDatavolumesMap[dv.Id]
				Expect(ok).To(BeTrue())
				validateDatavolume2(datavolumeNoSharedScopesToDatavolume(dv), datavolumeToRequest(expectedDv), projectToAncestors)
			}
		})

		Context("get datavolumes internal (by agent)", func() {
			It("sanity - get all datavolumes internal (by agent)", func() {
				res, err := datavolumesApiService.GetDatavolumesInternal(ctx, clusterId1, 0, 0)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				response := res.Body.(api.GetDatavolumesInternal200Response)
				datavolumes := response.Datavolumes
				// should be all datavolumes except one that is in another cluster.
				Expect(datavolumes).To(HaveLen(len(testDatavolumes) - 1))
				for _, dv := range datavolumes {
					expectedDv, ok := testDatavolumesMap[dv.Id]
					Expect(ok).To(BeTrue())
					validateDatavolume2(datavolumeInternalToDatavolume(dv), datavolumeToRequest(expectedDv), projectToAncestors)
					Expect(dv.Phase).To(Equal(api.DATAVOLUMEPHASE_READY))
					Expect(dv.SharedScopes).To(HaveLen(len(expectedDv.SharedScopes)))
					sharedScopesToUpsert, ok := testDatavolumeSharedScopesPutFields[dv.Id]
					Expect(ok).To(BeTrue())
					validateSharedScopes(sharedScopesToUpsert, dv.SharedScopes, time.Time{})
				}
			})

			It("with offset and limit", func() {
				res, err := datavolumesApiService.GetDatavolumesInternal(ctx, clusterId1, 0, 5)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				response := res.Body.(api.GetDatavolumesInternal200Response)
				datavolumes := response.Datavolumes
				Expect(datavolumes).To(HaveLen(5))
				Expect(response.Next).To(BeEquivalentTo(5))

				res, err = datavolumesApiService.GetDatavolumesInternal(ctx, clusterId1, 5, 4)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				response = res.Body.(api.GetDatavolumesInternal200Response)
				datavolumes = response.Datavolumes
				Expect(datavolumes).To(HaveLen(4))
				Expect(response.Next).To(BeEquivalentTo(9))

				res, err = datavolumesApiService.GetDatavolumesInternal(ctx, clusterId1, 9, 10)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				response = res.Body.(api.GetDatavolumesInternal200Response)
				datavolumes = response.Datavolumes
				Expect(datavolumes).To(HaveLen(1))
				Expect(response.Next).To(BeEquivalentTo(0))

				res, err = datavolumesApiService.GetDatavolumesInternal(ctx, clusterId1, 15, 4)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				response = res.Body.(api.GetDatavolumesInternal200Response)
				datavolumes = response.Datavolumes
				Expect(datavolumes).To(HaveLen(0))
				Expect(response.Next).To(BeEquivalentTo(0))

				res, err = datavolumesApiService.GetDatavolumesInternal(ctx, clusterId1, 0, 0)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				response = res.Body.(api.GetDatavolumesInternal200Response)
				datavolumes = response.Datavolumes
				Expect(datavolumes).To(HaveLen(10))
				Expect(response.Next).To(BeEquivalentTo(0))
			})
		})

		Context("get datavolumes with only certain permitted scopes", func() {
			It("only permissions in 1 project", func() {
				ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Projects: []string{"9"}})

				res, err := datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", 0, 0, "", "", []string{}, "", false)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				response := res.Body.(api.GetDatavolumes200Response)
				datavolumes := response.Datavolumes
				Expect(datavolumes).To(HaveLen(0))

				ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Projects: []string{"11"}})
				res, err = datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", 0, 0, "", "", []string{}, "", false)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				response = res.Body.(api.GetDatavolumes200Response)
				datavolumes = response.Datavolumes
				Expect(datavolumes).To(HaveLen(2))

				datavolumesResponseMap := datavolumesNoSharedScopesToMap(datavolumes)

				for i, dv := range testDatavolumes {
					if dv.ProjectId == "11" { // shared with project 9
						expectedDv, ok := datavolumesResponseMap[dv.Id]
						Expect(ok).To(BeTrue(), fmt.Sprintf("Datavolume ID should be in the response; i: %d, dv.ProjectId: %s", i, dv.ProjectId))
						validateDatavolume2(expectedDv, datavolumeToRequest(dv), projectToAncestors)
						continue
					}

					_, ok := datavolumesResponseMap[dv.Id]
					Expect(ok).To(BeFalse(), fmt.Sprintf("Datavolume ID should NOT be in the response; i: %d, dv.ProjectId: %s", i, dv.ProjectId))
				}
			})

			It("only permissions in 2 projects", func() {
				ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Projects: []string{"11", "14"}})

				res, err := datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", 0, 0, "", "", []string{}, "", false)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				response := res.Body.(api.GetDatavolumes200Response)
				datavolumes := response.Datavolumes
				Expect(datavolumes).To(HaveLen(4))

				datavolumesResponseMap := datavolumesNoSharedScopesToMap(datavolumes)

				for i, dv := range testDatavolumes {
					sharedScopesToUpsert, ok := testDatavolumeSharedScopesPutFields[dv.Id]
					Expect(ok).To(BeTrue())
					depIdSharedWith := ""
					if len(sharedScopesToUpsert) > 0 {
						depIdSharedWith = sharedScopesToUpsert[0].ScopeId
					}

					if dv.ProjectId == "11" || dv.ProjectId == "14" { // origin in my projects
						expectedDv, ok := datavolumesResponseMap[dv.Id]
						Expect(ok).To(BeTrue(), fmt.Sprintf("Datavolume ID should be in the response; i: %d, dv.ProjectId: %s, depIdSharedWith: %s", i, dv.ProjectId, depIdSharedWith))
						validateDatavolume2(expectedDv, datavolumeToRequest(dv), projectToAncestors)
						continue
					}

					_, ok = datavolumesResponseMap[dv.Id]
					Expect(ok).To(BeFalse(), fmt.Sprintf("Datavolume ID should NOT be in the response; i: %d, dv.ProjectId: %s, depIdSharedWith: %s", i, dv.ProjectId, depIdSharedWith))
				}
			})

			It("only permissions in 2 departments", func() {
				ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Departments: []string{"6", "4"}})

				res, err := datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", 0, 0, "", "", []string{}, "", false)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				response := res.Body.(api.GetDatavolumes200Response)
				datavolumes := response.Datavolumes
				Expect(datavolumes).To(HaveLen(4))

				datavolumesResponseMap := datavolumesNoSharedScopesToMap(datavolumes)

				for i, dv := range testDatavolumes {
					if dv.DepartmentId == "4" || dv.DepartmentId == "6" { // origin in my departments
						expectedDv, ok := datavolumesResponseMap[dv.Id]
						Expect(ok).To(BeTrue(), fmt.Sprintf("Datavolume ID should be in the response; i: %d, dv.DepartmentId: %s, dv.ProjectId: %s", i, dv.DepartmentId, dv.ProjectId))
						validateDatavolume2(expectedDv, datavolumeToRequest(dv), projectToAncestors)
						continue
					}

					_, ok := datavolumesResponseMap[dv.Id]
					Expect(ok).To(BeFalse(), fmt.Sprintf("Datavolume ID should NOT be in the response; i: %d, dv.DepartmentId: %s, dv.ProjectId: %s", i, dv.DepartmentId, dv.ProjectId))
				}
			})

			It("only permissions in 1 department", func() {
				ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Departments: []string{"1"}})

				res, err := datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", 0, 0, "", "", []string{}, "", false)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				response := res.Body.(api.GetDatavolumes200Response)
				datavolumes := response.Datavolumes
				Expect(datavolumes).To(HaveLen(0))

				ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Departments: []string{"7"}})
				res, err = datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", 0, 0, "", "", []string{}, "", false)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				response = res.Body.(api.GetDatavolumes200Response)
				datavolumes = response.Datavolumes
				Expect(datavolumes).To(HaveLen(2))

				datavolumesResponseMap := datavolumesNoSharedScopesToMap(datavolumes)

				for i, dv := range testDatavolumes {
					sharedScopesToUpsert, ok := testDatavolumeSharedScopesPutFields[dv.Id]
					Expect(ok).To(BeTrue())
					projIdSharedWith := ""
					if len(sharedScopesToUpsert) > 0 {
						projIdSharedWith = sharedScopesToUpsert[1].ScopeId
					}

					if dv.DepartmentId == "7" {
						expectedDv, ok := datavolumesResponseMap[dv.Id]
						Expect(ok).To(BeTrue(), fmt.Sprintf("Datavolume ID should be in the response; i: %d, dv.ProjectId: %s, projIdSharedWith: %s", i, dv.ProjectId, projIdSharedWith))
						validateDatavolume2(expectedDv, datavolumeToRequest(dv), projectToAncestors)
						continue
					}

					_, ok = datavolumesResponseMap[dv.Id]
					Expect(ok).To(BeFalse(), fmt.Sprintf("Datavolume ID should NOT be in the response; i: %d, dv.ProjectId: %s, projIdSharedWith: %s", i, dv.ProjectId, projIdSharedWith))
				}
			})

			It("only permissions in 2 clusters", func() {
				ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Clusters: []string{clusterId1, clusterId3}})

				res, err := datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", 0, 0, "", "", []string{}, "", false)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				response := res.Body.(api.GetDatavolumes200Response)
				datavolumes := response.Datavolumes

				// should be all datavolumes except one that is in another cluster.
				Expect(datavolumes).To(HaveLen(len(testDatavolumes) - 1))
				for _, dv := range datavolumes {
					expectedDv, ok := testDatavolumesMap[dv.Id]
					Expect(ok).To(BeTrue())
					validateDatavolume2(datavolumeNoSharedScopesToDatavolume(dv), datavolumeToRequest(expectedDv), projectToAncestors)
				}
			})

			It("only permissions in 1 cluster", func() {
				ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Clusters: []string{clusterId2}})

				res, err := datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", 0, 0, "", "", []string{}, "", false)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				response := res.Body.(api.GetDatavolumes200Response)
				datavolumes := response.Datavolumes

				// should be all datavolumes except one that is in another cluster.
				Expect(datavolumes).To(HaveLen(1))
				for _, dv := range datavolumes {
					expectedDv, ok := testDatavolumesMap[dv.Id]
					Expect(ok).To(BeTrue())
					validateDatavolume2(datavolumeNoSharedScopesToDatavolume(dv), datavolumeToRequest(expectedDv), projectToAncestors)
				}
			})

			It("only permissions in 1 project+ 1 department + 1 cluster", func() {
				ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
					Projects:    []string{"14"},
					Departments: []string{"6"},
					Clusters:    []string{clusterId2},
				})

				res, err := datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", 0, 0, "", "", []string{}, "", false)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				response := res.Body.(api.GetDatavolumes200Response)
				datavolumes := response.Datavolumes
				Expect(datavolumes).To(HaveLen(7))

				datavolumesResponseMap := datavolumesNoSharedScopesToMap(datavolumes)

				for i, dv := range testDatavolumes {
					sharedScopesToUpsert, ok := testDatavolumeSharedScopesPutFields[dv.Id]
					Expect(ok).To(BeTrue())
					depIdSharedWith := ""
					projIdSharedWith := ""
					if len(sharedScopesToUpsert) > 0 {
						depIdSharedWith = sharedScopesToUpsert[0].ScopeId
						projIdSharedWith = sharedScopesToUpsert[1].ScopeId
					}

					if (dv.DepartmentId == "6") || // origin in my departments
						(dv.ProjectId == "14" || dv.ProjectId == "12" || dv.ProjectId == "13") || // origin in my departments
						(dv.ClusterId == clusterId2) || // origin in my clusters
						(depIdSharedWith == "7" || depIdSharedWith == "6") || // shared with my departments
						(i%2 == 1 && dv.ProjectId == "13") || // shared with the whole cluster1
						(projIdSharedWith == "14" || projIdSharedWith == "12" || projIdSharedWith == "13") { // shared with my projects
						expectedDv, ok := datavolumesResponseMap[dv.Id]
						Expect(ok).To(BeTrue(), fmt.Sprintf("Datavolume ID should be in the response; i: %d, dv.DepartmentId: %s, depIdSharedWith: %s", i, dv.DepartmentId, depIdSharedWith))
						validateDatavolume2(expectedDv, datavolumeToRequest(dv), projectToAncestors)
						continue
					}

					_, ok = datavolumesResponseMap[dv.Id]
					Expect(ok).To(BeFalse(), fmt.Sprintf("Datavolume ID should NOT be in the response; i: %d, dv.DepartmentId: %s, depIdSharedWith: %s", i, dv.DepartmentId, depIdSharedWith))
				}
			})

			It("only permissions in different scopes without datavolumes in them", func() {
				ctxs := []context.Context{
					context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Projects: []string{"8"}}),
					context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Projects: []string{"4"}}),
					context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Departments: []string{"10"}}),
					context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Departments: []string{"20"}}),
					context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Clusters: []string{clusterId3}}),
					context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Clusters: []string{uuid.NewString()}}),
				}

				for _, c := range ctxs {
					res, err := datavolumesApiService.GetDatavolumes(c, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", 0, 0, "", "", []string{}, "", false)
					Expect(err).To(BeNil())
					Expect(res.Code).To(Equal(http.StatusOK))
					response := res.Body.(api.GetDatavolumes200Response)
					datavolumes := response.Datavolumes
					Expect(datavolumes).To(HaveLen(0))
				}

				tenantsForTest := []string{"10", TestTenant2}

				for _, tenantForTest := range tenantsForTest {
					ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Tenants: []string{tenantForTest}})
					ctx = SetContext(tenantForTest, TestUser)

					res, err := datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", 0, 0, "", "", []string{}, "", false)
					Expect(err).To(BeNil())
					Expect(res.Code).To(Equal(http.StatusOK))
					response := res.Body.(api.GetDatavolumes200Response)
					datavolumes := response.Datavolumes
					Expect(datavolumes).To(HaveLen(0))
				}
			})

			It("only permissions in unexisting scopes", func() {
				ctxs := []context.Context{
					context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Projects: []string{"200"}}),
					context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Departments: []string{"999"}}),
				}

				for _, c := range ctxs {
					res, err := datavolumesApiService.GetDatavolumes(c, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", 0, 0, "", "", []string{}, "", false)
					Expect(err).To(BeNil())
					Expect(res.Code).To(Equal(http.StatusOK))
					datavolumes := res.Body.(api.GetDatavolumes200Response).Datavolumes
					Expect(datavolumes).To(HaveLen(0))
				}
			})
		})

		It("get datavolumes - test query params - offset and limit", func() {
			responseMap := map[string]api.Datavolume{}

			res, err := datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", 0, 2, "", "", []string{}, "", false)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			response := res.Body.(api.GetDatavolumes200Response)
			datavolumes := response.Datavolumes
			Expect(datavolumes).To(HaveLen(2))
			Expect(response.Next).To(BeEquivalentTo(2))
			for _, dv := range datavolumes {
				responseMap[dv.Id] = datavolumeNoSharedScopesToDatavolume(dv)
			}

			res, err = datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", 2, 5, "", "", []string{}, "", false)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			response = res.Body.(api.GetDatavolumes200Response)
			datavolumes = response.Datavolumes
			Expect(datavolumes).To(HaveLen(5))
			Expect(response.Next).To(BeEquivalentTo(7))
			for _, dv := range datavolumes {
				responseMap[dv.Id] = datavolumeNoSharedScopesToDatavolume(dv)
			}

			res, err = datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", 7, 4, "", "", []string{}, "", false)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			response = res.Body.(api.GetDatavolumes200Response)
			datavolumes = response.Datavolumes
			Expect(datavolumes).To(HaveLen(4))
			Expect(response.Next).To(BeEquivalentTo(11))
			for _, dv := range datavolumes {
				responseMap[dv.Id] = datavolumeNoSharedScopesToDatavolume(dv)
			}

			Expect(responseMap).To(HaveLen(len(testDatavolumesMap)))

			res, err = datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", 8, 20, "", "", []string{}, "", false)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			response = res.Body.(api.GetDatavolumes200Response)
			datavolumes = response.Datavolumes
			Expect(datavolumes).To(HaveLen(3))
			Expect(response.Next).To(BeEquivalentTo(0))

			res, err = datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", 20, 2, "", "", []string{}, "", false)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			response = res.Body.(api.GetDatavolumes200Response)
			datavolumes = response.Datavolumes
			Expect(datavolumes).To(HaveLen(0))
			Expect(response.Next).To(BeEquivalentTo(0))

			res, err = datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", 5, 0, "", "", []string{}, "", false)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			response = res.Body.(api.GetDatavolumes200Response)
			datavolumes = response.Datavolumes
			Expect(datavolumes).To(HaveLen(6))
			Expect(response.Next).To(BeEquivalentTo(0))
		})

		It("get datavolumes - test query params - sortBy and sortOrder with offset and limit", func() {
			sortByFields := []string{"name", "projectId", "clusterId", "departmentId", "projectName", "createdAt", "createdBy", "phase"}
			for _, sortBy := range sortByFields {
				for _, offset := range []int32{0, 5} {
					for _, limit := range []int32{3, 5, 11} {
						res, err := datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", offset, limit, sortBy, "asc", []string{}, "", false)
						Expect(err).To(BeNil())
						Expect(res.Code).To(Equal(http.StatusOK))
						response := res.Body.(api.GetDatavolumes200Response)
						datavolumes := response.Datavolumes
						expectedCount := int(limit)
						if limit == 11 && offset == 5 {
							expectedCount = 6
						}
						Expect(datavolumes).To(HaveLen(expectedCount))
						lastValue := ""
						for _, dv := range datavolumes {
							value := getSortFilterByValue(dv, sortBy)
							if lastValue != "" {
								Expect(value >= lastValue).To(BeTrue(), fmt.Sprintf("asc sortBy: %s, value: %s, lastValue: %s", sortBy, value, lastValue))
							}
							lastValue = value
						}

						res, err = datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", offset, limit, sortBy, "desc", []string{}, "", false)
						Expect(err).To(BeNil())
						Expect(res.Code).To(Equal(http.StatusOK))
						response = res.Body.(api.GetDatavolumes200Response)
						datavolumes = response.Datavolumes
						Expect(datavolumes).To(HaveLen(expectedCount))
						lastValue = ""
						for _, dv := range datavolumes {
							value := getSortFilterByValue(dv, sortBy)
							if lastValue != "" {
								Expect(value <= lastValue).To(BeTrue(), fmt.Sprintf("desc sortBy: %s, value: %s, lastValue: %s", sortBy, value, lastValue))
							}
							lastValue = value
						}
					}
				}
			}

			res, err := datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", 0, 0, "invalidSortByFields", "asc", []string{}, "", false)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
		})

		It("get datavolumes - test query params - filterby", func() {
			filtersToTest := map[string]string{
				"name":         "dv-10-0",
				"projectId":    "10",
				"clusterId":    clusterId1,
				"departmentId": "1",
				"projectName":  "bla",
				"createdAt":    "2021-01-01T00:00:00Z",
				"createdBy":    TestUser,
				"phase":        "CREATING",
			}

			for filterBy, filterValue := range filtersToTest {
				for _, filterOperand := range []string{"==", "!="} {
					res, err := datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", 0, 0, "", "", []string{fmt.Sprintf("%s%s%s", filterBy, filterOperand, filterValue)}, "", false)
					Expect(err).To(BeNil())
					Expect(res.Code).To(Equal(http.StatusOK))
					response := res.Body.(api.GetDatavolumes200Response)
					datavolumes := response.Datavolumes
					for _, dv := range datavolumes {
						value := getSortFilterByValue(dv, filterBy)
						if filterOperand == "==" {
							Expect(value).To(BeEquivalentTo(filterValue))
						} else {
							Expect(value).ToNot(BeEquivalentTo(filterValue))
						}
					}
				}
			}

			invalidFilterBys := []string{"invalidFilterByFields:invalidValue", "!=blabla", "projectId:5", "blabl2"}
			for _, invalidFilterBy := range invalidFilterBys {
				res, err := datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", 0, 0, "", "", []string{invalidFilterBy}, "", false)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusBadRequest))
			}
		})

		It("get datavolumes - test query params - search", func() {
			res, err := datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", 0, 0, "", "", []string{}, "dv-10", false)
			Expect(err).To(BeNil())
			response := res.Body.(api.GetDatavolumes200Response)
			Expect(len(response.Datavolumes)).To(Equal(2))
			for _, dv := range response.Datavolumes {
				Expect(dv.Name).To(ContainSubstring("dv-10"))
			}
		})

		It("get datavolumes - test wrong query params - requestType + usableInProjectId", func() {
			wrongRequestsToTest := map[api.DatavolumeRequestType]string{
				api.DATAVOLUMEREQUESTTYPE_ORIGINATED:        "10",
				api.DATAVOLUMEREQUESTTYPE_USABLE_IN_PROJECT: "",
				api.DatavolumeRequestType(""):               "doesnt-matter",
				api.DatavolumeRequestType("blabla"):         "10",
				api.DatavolumeRequestType("originated"):     "1",
			}

			for requestType, usableInProjectId := range wrongRequestsToTest {
				res, err := datavolumesApiService.GetDatavolumes(ctx, requestType, usableInProjectId, 0, 0, "", "", []string{}, "", false)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusBadRequest))
			}
		})

		Context("get datavolumes - test query param - sharedWithProjectIdFilter", func() {
			It("sanity", func() {
				for j := 0; j < 6; j++ {
					switch j {
					case 1:
						ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Projects: []string{"10"}})
					case 2:
						ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Departments: []string{projectToAncestors["10"][0]}})
					case 3:
						ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Clusters: []string{projectToAncestors["10"][1]}})
					case 4:
						ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Tenants: []string{TestTenant}})
					case 5:
						ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{System: ptr.To(true)})
					}

					res, err := datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_USABLE_IN_PROJECT, "10", 0, 0, "", "", []string{}, "", false)
					Expect(err).To(BeNil())
					Expect(res.Code).To(Equal(http.StatusOK))
					response := res.Body.(api.GetDatavolumes200Response)
					datavolumes := response.Datavolumes
					Expect(datavolumes).To(HaveLen(8))

					datavolumesResponseMap := datavolumesNoSharedScopesToMap(datavolumes)

					for i, dv := range testDatavolumes {
						if (dv.ProjectId == "10") || // originated in project 10
							(dv.ProjectId == "11") || // shared with project 10
							(i%2 == 0 && i != 10) || // shared with department 3
							(i%2 == 1 && dv.ProjectId == "13") { // shared with the whole cluster
							expectedDv, ok := datavolumesResponseMap[dv.Id]
							Expect(ok).To(BeTrue(), fmt.Sprintf("Datavolume ID should be in the response; i: %d, dv.ProjectId: %s", i, dv.ProjectId))
							validateDatavolume2(expectedDv, datavolumeToRequest(dv), projectToAncestors)
							continue
						}

						_, ok := datavolumesResponseMap[dv.Id]
						Expect(ok).To(BeFalse(), fmt.Sprintf("Datavolume ID should NOT be in the response; i: %d, dv.ProjectId: %s", i, dv.ProjectId))
					}
				}
			})

			It("request usable in project that is not permitted", func() {
				permittedScopess := []authorization_api.PermittedScopes{
					{Projects: []string{"11"}},
					{Departments: []string{"1"}},
					{Clusters: []string{clusterId2}},
					{Tenants: []string{TestTenant2}},
					{System: ptr.To(false)},
				}

				for i := range permittedScopess {
					permittedScopes := permittedScopess[i]
					ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &permittedScopes)

					res, err := datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_USABLE_IN_PROJECT, "10", 0, 0, "", "", []string{}, "", false)
					Expect(err).To(BeNil())
					Expect(res.Code).To(Equal(http.StatusForbidden))
				}
			})

			It("more cases", func() {
				ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Projects: []string{"10"}})

				res, err := datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_USABLE_IN_PROJECT, "10", 0, 0, "", "", []string{}, "", false)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				response := res.Body.(api.GetDatavolumes200Response)
				datavolumes := response.Datavolumes
				Expect(datavolumes).To(HaveLen(8))

				datavolumesResponseMap := datavolumesNoSharedScopesToMap(datavolumes)

				for i, dv := range testDatavolumes {
					if (dv.ProjectId == "10") || // originated in project 10
						(dv.ProjectId == "11") || // shared with project 10
						(i%2 == 0 && i != 10) || // shared with department 3
						(i%2 == 1 && dv.ProjectId == "13") { // shared with the whole cluster
						expectedDv, ok := datavolumesResponseMap[dv.Id]
						Expect(ok).To(BeTrue(), fmt.Sprintf("Datavolume ID should be in the response; i: %d, dv.ProjectId: %s", i, dv.ProjectId))
						validateDatavolume2(expectedDv, datavolumeToRequest(dv), projectToAncestors)
						continue
					}

					_, ok := datavolumesResponseMap[dv.Id]
					Expect(ok).To(BeFalse(), fmt.Sprintf("Datavolume ID should NOT be in the response; i: %d, dv.ProjectId: %s", i, dv.ProjectId))
				}
			})

			It("no or almost no results", func() {
				for j := 0; j < 6; j++ {
					ctx = SetContext(TestTenant, TestUser)
					ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{System: ptr.To(true)})
					projectIdFilter := "9"
					filterBy := []string{}
					expectedLength := 0

					switch j {
					case 0:
						projectIdFilter = "8"
						expectedLength = 0
					case 1:
						filterBy = []string{"projectId==1"}
						expectedLength = 0
					case 2:
						filterBy = []string{fmt.Sprintf("clusterId!=%s", clusterId1)}
						expectedLength = 0
					case 3:
						projectIdFilter = "5" // this DV is not ready
						expectedLength = 1
					case 4:
						projectIdFilter = "5" // this DV is not ready
						filterBy = []string{"phase==Ready"}
						expectedLength = 0
					case 5:
						// sanity
						expectedLength = 3
					}

					res, err := datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_USABLE_IN_PROJECT, projectIdFilter, 0, 0, "", "", filterBy, "", false)
					Expect(err).To(BeNil())
					Expect(res.Code).To(Equal(http.StatusOK))
					response := res.Body.(api.GetDatavolumes200Response)
					datavolumes := response.Datavolumes
					Expect(datavolumes).To(HaveLen(expectedLength), fmt.Sprintf("j: %d, projectIdFilter: %s, filterBy: %v", j, projectIdFilter, filterBy))
				}
			})

			It("invalid projectId", func() {
				res, err := datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_USABLE_IN_PROJECT, "999", 0, 0, "", "", []string{}, "", false)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusForbidden))
			})
		})
	})

	Context("get datavolume", func() {
		var (
			projectId         string
			testDatavolumeId  string
			request           api.DatavolumeCreationFields
			sharedScopesToAdd []api.SharedScopeBase
		)

		BeforeEach(func() {
			projectId = "9"
			request = api.DatavolumeCreationFields{
				Name:          "dv-100",
				Description:   "Some description 100",
				ProjectId:     projectId,
				OriginPvcName: "pvc-100",
			}

			res, err := datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil(), "Create datavolume should succeed")

			Expect(res.Code).To(Equal(http.StatusCreated))
			testDatavolumeId = res.Body.(api.Datavolume).Id
			datavolumeIds[testDatavolumeId] = TestTenant

			sharedScopesToAdd = []api.SharedScopeBase{
				{
					ScopeId:   "3",
					ScopeType: api.SCOPETYPE_PROJECT,
				},
				{
					ScopeId:   "3",
					ScopeType: api.SCOPETYPE_DEPARTMENT,
				},
				{
					ScopeId:   "6",
					ScopeType: api.SCOPETYPE_DEPARTMENT,
				},
				{
					ScopeId:   clusterId1,
					ScopeType: api.SCOPETYPE_CLUSTER,
				},
			}

			sharedScopesPatch := api.SharedScopesPatchRequest{Add: sharedScopesToAdd}
			res, err = datavolumesApiService.PatchDatavolumeSharedScopes(ctx, testDatavolumeId, sharedScopesPatch)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
		})

		It("sanity", func() {
			res, err := datavolumesApiService.GetDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))

			datavoume := res.Body.(api.Datavolume)
			validateDatavolume(datavoume, request, projectToAncestors)
			validateSharedScopes(sharedScopesToAdd, datavoume.SharedScopes, time.Time{})
		})

		It("invalid tenant id", func() {
			ctx = SetContext(TestTenant2, TestUser)
			res, err := datavolumesApiService.GetDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNotFound))

			ctx = context.Background()
			res, err = datavolumesApiService.GetDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
		})

		It("invalid datavolumeId", func() {
			res, err := datavolumesApiService.GetDatavolume(ctx, "invalid-datavolume-id")
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))

			res, err = datavolumesApiService.GetDatavolume(ctx, uuid.NewString())
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNotFound))
		})

		It("filter out permitted scopes - projects", func() {
			ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
				Projects: []string{projectId},
			})

			res, err := datavolumesApiService.GetDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			sharedScopes := res.Body.(api.Datavolume).SharedScopes
			Expect(sharedScopes).To(HaveLen(0))

			ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
				Projects: []string{projectId, "3"},
			})
			res, err = datavolumesApiService.GetDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			sharedScopes = res.Body.(api.Datavolume).SharedScopes
			Expect(sharedScopes).To(HaveLen(1))
			validateSharedScope(sharedScopesToAdd[0], sharedScopes[0], time.Time{})
		})

		It("filter out permitted scopes - departments", func() {
			ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
				Departments: []string{"3", projectToAncestors[projectId][0]},
			})

			res, err := datavolumesApiService.GetDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			sharedScopes := res.Body.(api.Datavolume).SharedScopes
			Expect(sharedScopes).To(HaveLen(2))
			validateSharedScopes(sharedScopesToAdd[:2], sharedScopes, time.Time{})

			ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
				Departments: []string{"6", projectToAncestors[projectId][0]},
			})
			res, err = datavolumesApiService.GetDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			sharedScopes = res.Body.(api.Datavolume).SharedScopes
			Expect(sharedScopes).To(HaveLen(1))
			validateSharedScope(sharedScopesToAdd[2], sharedScopes[0], time.Time{})

			ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
				Departments: []string{projectToAncestors[projectId][0]},
			})
			res, err = datavolumesApiService.GetDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			sharedScopes = res.Body.(api.Datavolume).SharedScopes
			Expect(sharedScopes).To(HaveLen(0))
		})

		It("filter out permitted scopes - clusters", func() {
			ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
				Clusters: []string{clusterId1},
			})

			res, err := datavolumesApiService.GetDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			sharedScopes := res.Body.(api.Datavolume).SharedScopes
			Expect(sharedScopes).To(HaveLen(4))
			validateSharedScopes(sharedScopesToAdd, sharedScopes, time.Time{})

			ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
				Projects: []string{projectId},
				Clusters: []string{clusterId2},
			})
			res, err = datavolumesApiService.GetDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			sharedScopes = res.Body.(api.Datavolume).SharedScopes
			Expect(sharedScopes).To(HaveLen(0))
		})

		It("filter out permitted scopes - tenants and system", func() {
			ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
				Tenants: []string{TestTenant},
			})

			res, err := datavolumesApiService.GetDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			sharedScopes := res.Body.(api.Datavolume).SharedScopes
			Expect(sharedScopes).To(HaveLen(4))
			validateSharedScopes(sharedScopesToAdd, sharedScopes, time.Time{})

			ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
				Projects: []string{projectId},
				Tenants:  []string{TestTenant2},
			})
			res, err = datavolumesApiService.GetDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			sharedScopes = res.Body.(api.Datavolume).SharedScopes
			Expect(sharedScopes).To(HaveLen(0))

			ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
				Projects: []string{projectId},
				System:   ptr.To(true),
			})
			res, err = datavolumesApiService.GetDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			sharedScopes = res.Body.(api.Datavolume).SharedScopes
			Expect(sharedScopes).To(HaveLen(4))

			ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
				Projects: []string{projectId},
				System:   ptr.To(false),
			})
			res, err = datavolumesApiService.GetDatavolume(ctx, testDatavolumeId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			sharedScopes = res.Body.(api.Datavolume).SharedScopes
			Expect(sharedScopes).To(HaveLen(0))
		})
	})

	Context("count datavolumes", func() {
		var (
			testDatavolumes                     = []api.Datavolume{}
			testDatavolumesMap                  = map[string]api.Datavolume{}
			testDatavolumeSharedScopesPutFields = map[string][]api.SharedScopeBase{}
			request                             api.DatavolumeCreationFields
		)

		BeforeEach(func() {
			testDatavolumes = []api.Datavolume{}
			testDatavolumesMap = map[string]api.Datavolume{}
			testDatavolumeSharedScopesPutFields = map[string][]api.SharedScopeBase{}

			request = api.DatavolumeCreationFields{
				Name:        "dv-0",
				Description: "bla bla",
			}

			for i := 10; i <= 14; i++ {
				request.ProjectId = fmt.Sprintf("%d", i)
				for j := 0; j < 2; j++ {
					// create 2 in each project
					request.Name = fmt.Sprintf("dv-%d-%d", i, j)
					request.OriginPvcName = fmt.Sprintf("pvc-%d-%d", i, j)

					res, err := datavolumesApiService.CreateDatavolume(ctx, request)
					Expect(err).To(BeNil(), "Create datavolume should succeed")
					Expect(res.Code).To(Equal(http.StatusCreated))
					dv := res.Body.(api.Datavolume)
					datavolumeIds[dv.Id] = TestTenant

					// add shared scopes
					depIdToShareWith := "4"
					if j%2 == 0 {
						depIdToShareWith = "3"
					}
					sharedScopesToAdd := []api.SharedScopeBase{
						{ScopeId: depIdToShareWith, ScopeType: api.SCOPETYPE_DEPARTMENT},
						{ScopeId: fmt.Sprintf("%d", i-1), ScopeType: api.SCOPETYPE_PROJECT},
					}
					if i == 13 && j == 1 {
						sharedScopesToAdd = append(sharedScopesToAdd, api.SharedScopeBase{ScopeId: clusterId1, ScopeType: api.SCOPETYPE_CLUSTER})
					}

					sharedScopesPatch := api.SharedScopesPatchRequest{Add: sharedScopesToAdd}
					res, err = datavolumesApiService.PatchDatavolumeSharedScopes(ctx, dv.Id, sharedScopesPatch)
					Expect(err).To(BeNil())
					Expect(res.Code).To(Equal(http.StatusOK))

					res, err = datavolumesApiService.GetDatavolumeSharedScopes(ctx, dv.Id)
					Expect(err).To(BeNil())
					Expect(res.Code).To(Equal(http.StatusOK))

					dv.SharedScopes = res.Body.([]api.SharedScope)
					testDatavolumes = append(testDatavolumes, dv)
					testDatavolumesMap[dv.Id] = dv
					testDatavolumeSharedScopesPutFields[dv.Id] = sharedScopesToAdd

					// update status to ready
					internalStatusRequest := api.DatavolumeStatusInternal{
						Phase: api.DATAVOLUMEPHASE_READY,
					}
					res, err = datavolumesApiService.UpdateDatavolumeStatus(ctx, dv.Id, internalStatusRequest)
					Expect(err).To(BeNil())
					Expect(res.Code).To(Equal(http.StatusNoContent))
				}
			}

			// and create another one in a different cluster
			request.ProjectId = "5"
			request.Name = "dv-second-cluster"
			res, err := datavolumesApiService.CreateDatavolume(ctx, request)
			Expect(err).To(BeNil(), "Create datavolume should succeed")
			Expect(res.Code).To(Equal(http.StatusCreated))
			dv := res.Body.(api.Datavolume)
			datavolumeIds[dv.Id] = TestTenant
			testDatavolumes = append(testDatavolumes, dv)
			testDatavolumesMap[dv.Id] = dv
			testDatavolumeSharedScopesPutFields[dv.Id] = []api.SharedScopeBase{}
		})

		It("sanity - count all datavolumes", func() {
			res2, err2 := datavolumesApiService.CountDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", []string{}, "")
			Expect(err2).To(BeNil())
			Expect(res2.Code).To(Equal(http.StatusOK))
			response2 := res2.Body.(api.CountDatavolumes200Response)
			count := response2.Count
			Expect(count).To(BeEquivalentTo(len(testDatavolumes)))
		})

		Context("count datavolumes with only certain permitted scopes", func() {
			It("only permissions in 1 project", func() {
				ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Projects: []string{"9"}})

				res, err := datavolumesApiService.CountDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", []string{}, "")
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				response := res.Body.(api.CountDatavolumes200Response)
				count := response.Count
				Expect(count).To(BeEquivalentTo(0))

				ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Projects: []string{"11"}})
				res, err = datavolumesApiService.CountDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", []string{}, "")
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				response = res.Body.(api.CountDatavolumes200Response)
				count = response.Count
				Expect(count).To(BeEquivalentTo(2))
			})

			It("only permissions in 2 projects", func() {
				ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Projects: []string{"11", "14"}})

				res, err := datavolumesApiService.CountDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", []string{}, "")
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				response := res.Body.(api.CountDatavolumes200Response)
				count := response.Count
				Expect(count).To(BeEquivalentTo(4))
			})

			It("only permissions in 2 departments", func() {
				ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Departments: []string{"6", "4"}})

				res, err := datavolumesApiService.CountDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", []string{}, "")
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				response := res.Body.(api.CountDatavolumes200Response)
				count := response.Count
				Expect(count).To(BeEquivalentTo(4))
			})

			It("only permissions in 1 department", func() {
				ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Departments: []string{"1"}})

				res, err := datavolumesApiService.CountDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", []string{}, "")
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				response := res.Body.(api.CountDatavolumes200Response)
				count := response.Count
				Expect(count).To(BeEquivalentTo(0))

				ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Departments: []string{"7"}})
				res, err = datavolumesApiService.CountDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", []string{}, "")
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				response = res.Body.(api.CountDatavolumes200Response)
				count = response.Count
				Expect(count).To(BeEquivalentTo(2))
			})

			It("only permissions in 2 clusters", func() {
				ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Clusters: []string{clusterId1, clusterId3}})

				res, err := datavolumesApiService.CountDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", []string{}, "")
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				response := res.Body.(api.CountDatavolumes200Response)
				count := response.Count

				// should be all datavolumes except one that is in another cluster.
				Expect(count).To(BeEquivalentTo(len(testDatavolumes) - 1))
			})

			It("only permissions in 1 cluster", func() {
				ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Clusters: []string{clusterId2}})

				res, err := datavolumesApiService.CountDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", []string{}, "")
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				response := res.Body.(api.CountDatavolumes200Response)
				count := response.Count

				Expect(count).To(BeEquivalentTo(1))
			})

			It("only permissions in 1 project+ 1 department + 1 cluster", func() {
				ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
					Projects:    []string{"14"},
					Departments: []string{"6"},
					Clusters:    []string{clusterId2},
				})

				res, err := datavolumesApiService.CountDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", []string{}, "")
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusOK))
				response := res.Body.(api.CountDatavolumes200Response)
				count := response.Count
				Expect(count).To(BeEquivalentTo(7))
			})

			It("only permissions in different scopes without datavolumes in them", func() {
				ctxs := []context.Context{
					context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Projects: []string{"8"}}),
					context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Projects: []string{"4"}}),
					context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Departments: []string{"10"}}),
					context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Departments: []string{"20"}}),
					context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Clusters: []string{clusterId3}}),
					context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Clusters: []string{uuid.NewString()}}),
				}

				for _, c := range ctxs {
					res, err := datavolumesApiService.CountDatavolumes(c, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", []string{}, "")
					Expect(err).To(BeNil())
					Expect(res.Code).To(Equal(http.StatusOK))
					response := res.Body.(api.CountDatavolumes200Response)
					count := response.Count
					Expect(count).To(BeEquivalentTo(0))
				}

				tenantsForTest := []string{"10", TestTenant2}

				for _, tenantForTest := range tenantsForTest {
					ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Tenants: []string{tenantForTest}})
					ctx = SetContext(tenantForTest, TestUser)

					res, err := datavolumesApiService.CountDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", []string{}, "")
					Expect(err).To(BeNil())
					Expect(res.Code).To(Equal(http.StatusOK))
					response := res.Body.(api.CountDatavolumes200Response)
					count := response.Count
					Expect(count).To(BeEquivalentTo(0))
				}
			})

			It("only permissions in unexisting scopes", func() {
				ctxs := []context.Context{
					context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Projects: []string{"200"}}),
					context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Departments: []string{"999"}}),
				}

				for _, c := range ctxs {
					res, err := datavolumesApiService.CountDatavolumes(c, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", []string{}, "")
					Expect(err).To(BeNil())
					Expect(res.Code).To(Equal(http.StatusOK))
					count := res.Body.(api.CountDatavolumes200Response).Count
					Expect(count).To(BeEquivalentTo(0))
				}
			})
		})

		It("count datavolumes - test query params - filterby", func() {
			filtersToTest := map[string]string{
				"name":         "dv-10-0",
				"projectId":    "10",
				"clusterId":    clusterId1,
				"departmentId": "1",
				"projectName":  "bla",
				"createdAt":    "2021-01-01T00:00:00Z",
				"createdBy":    TestUser,
				"phase":        "CREATING",
			}

			for filterBy, filterValue := range filtersToTest {
				for _, filterOperand := range []string{"==", "!="} {
					res, err := datavolumesApiService.CountDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", []string{fmt.Sprintf("%s%s%s", filterBy, filterOperand, filterValue)}, "")
					Expect(err).To(BeNil())
					Expect(res.Code).To(Equal(http.StatusOK))
				}
			}

			invalidFilterBys := []string{"invalidFilterByFields:invalidValue", "!=blabla", "projectId:5", "blabl2"}
			for _, invalidFilterBy := range invalidFilterBys {
				res, err := datavolumesApiService.CountDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", []string{invalidFilterBy}, "")
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusBadRequest))
			}
		})

		It("count datavolumes - test query params - search", func() {
			res, err := datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", 0, 0, "", "", []string{}, "dv-10", false)
			Expect(err).To(BeNil())
			response := res.Body.(api.GetDatavolumes200Response)
			Expect(len(response.Datavolumes)).To(Equal(2))
		})

		It("count datavolumes - test wrong query params - requestType + usableInProjectId", func() {
			wrongRequestsToTest := map[api.DatavolumeRequestType]string{
				api.DATAVOLUMEREQUESTTYPE_ORIGINATED:        "10",
				api.DATAVOLUMEREQUESTTYPE_USABLE_IN_PROJECT: "",
				api.DatavolumeRequestType(""):               "doesnt-matter",
				api.DatavolumeRequestType("blabla"):         "10",
				api.DatavolumeRequestType("originated"):     "1",
			}

			for requestType, usableInProjectId := range wrongRequestsToTest {
				res, err := datavolumesApiService.CountDatavolumes(ctx, requestType, usableInProjectId, []string{}, "")
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusBadRequest))
			}
		})

		Context("count datavolumes - test query param - sharedWithProjectIdFilter", func() {
			It("sanity", func() {
				for j := 0; j < 6; j++ {
					switch j {
					case 1:
						ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Projects: []string{"10"}})
					case 2:
						ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Departments: []string{projectToAncestors["10"][0]}})
					case 3:
						ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Clusters: []string{projectToAncestors["10"][1]}})
					case 4:
						ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Tenants: []string{TestTenant}})
					case 5:
						ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{System: ptr.To(true)})
					}

					res, err := datavolumesApiService.CountDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_USABLE_IN_PROJECT, "10", []string{}, "")
					Expect(err).To(BeNil())
					Expect(res.Code).To(Equal(http.StatusOK))
					response := res.Body.(api.CountDatavolumes200Response)
					count := response.Count
					Expect(count).To(BeEquivalentTo(8))
				}
			})

			It("request usable in project that is not permitted", func() {
				permittedScopess := []authorization_api.PermittedScopes{
					{Projects: []string{"11"}},
					{Departments: []string{"1"}},
					{Clusters: []string{clusterId2}},
					{Tenants: []string{TestTenant2}},
					{System: ptr.To(false)},
				}

				for i := range permittedScopess {
					permittedScopes := permittedScopess[i]
					ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &permittedScopes)

					res, err := datavolumesApiService.CountDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_USABLE_IN_PROJECT, "10", []string{}, "")
					Expect(err).To(BeNil())
					Expect(res.Code).To(Equal(http.StatusForbidden))
				}
			})
		})
	})

	Context("internal datavolumes", func() {
		var dataVolumeIds []string
		const internalCreatedBy = "runai"
		BeforeEach(func() {
			dataVolumeIds = make([]string, 0)
		})
		AfterEach(func() {
			for _, id := range dataVolumeIds {
				datavolumesApiService.DeleteDatavolume(ctx, id)
			}
		})
		Context("create internal", func() {
			It("create internal datavolume - cluster scope", func() {
				request := api.InternalDatavolumeCreationFields{
					Name:          "dv",
					Scope:         "cluster",
					ClusterId:     clusterId4,
					OriginPvcName: "pvc-1",
				}
				res, err := datavolumesApiService.InternalCreateDatavolume(ctx, request)
				Expect(err).To(BeNil())
				dv := res.Body.(*models.Datavolume)
				dataVolumeIds = append(dataVolumeIds, dv.ID.String())
				Expect(dv.Name).To(BeEquivalentTo(request.Name))
				Expect(dv.CreatedBy).To(BeEquivalentTo("runai"))

				res, err = datavolumesApiService.GetDatavolumeSharedScopes(ctx, dv.ID.String())
				Expect(err).To(BeNil())
				sharedScopes := res.Body.([]api.SharedScope)
				Expect(sharedScopes).ToNot(BeNil())
				Expect(sharedScopes).To(HaveLen(1))
				Expect(sharedScopes[0].ScopeId).To(BeEquivalentTo(request.ClusterId))
			})
			It("create internal datavolume - department scope", func() {
				request := api.InternalDatavolumeCreationFields{
					Name:          "dv",
					Scope:         "department",
					DepartmentId:  "8",
					OriginPvcName: "pvc-1",
				}
				res, err := datavolumesApiService.InternalCreateDatavolume(ctx, request)
				Expect(err).To(BeNil())
				dv := res.Body.(*models.Datavolume)
				dataVolumeIds = append(dataVolumeIds, dv.ID.String())
				Expect(dv.Name).To(BeEquivalentTo(request.Name))
				Expect(dv.CreatedBy).To(BeEquivalentTo(internalCreatedBy))

				res, err = datavolumesApiService.GetDatavolumeSharedScopes(ctx, dv.ID.String())
				Expect(err).To(BeNil())
				sharedScopes := res.Body.([]api.SharedScope)
				Expect(sharedScopes).ToNot(BeNil())
				Expect(sharedScopes).To(HaveLen(1))
				Expect(sharedScopes[0].ScopeId).To(BeEquivalentTo(request.DepartmentId))
			})
			It("create internal datavolume - invalid cluster", func() {
				request := api.InternalDatavolumeCreationFields{
					Name:          "dv",
					Scope:         "cluster",
					ClusterId:     clusterId3,
					OriginPvcName: "pvc-1",
				}
				res, err := datavolumesApiService.InternalCreateDatavolume(ctx, request)
				Expect(err).To(BeNil())
				Expect(res.Code).To(BeEquivalentTo(http.StatusBadRequest))
			})
			It("create internal datavolume - invalid scope", func() {
				request := api.InternalDatavolumeCreationFields{
					Name:          "dv",
					Scope:         "project",
					ClusterId:     clusterId4,
					OriginPvcName: "pvc-1",
				}
				res, err := datavolumesApiService.InternalCreateDatavolume(ctx, request)
				Expect(err).To(BeNil())
				Expect(res.Code).To(BeEquivalentTo(http.StatusBadRequest))
			})
		})

		Context("get datavolumes", func() {
			It("should not return internal datavolumes", func() {
				request := api.InternalDatavolumeCreationFields{
					Name:          "internal-dv",
					Scope:         "cluster",
					ClusterId:     clusterId4,
					OriginPvcName: "pvc-1",
				}
				res, err := datavolumesApiService.InternalCreateDatavolume(ctx, request)
				Expect(err).To(BeNil())
				Expect(res.Code).To(BeEquivalentTo(http.StatusOK))
				internalDv := res.Body.(*models.Datavolume)
				dataVolumeIds = append(dataVolumeIds, internalDv.ID.String())

				res, err = datavolumesApiService.CreateDatavolume(ctx, api.DatavolumeCreationFields{
					Name:          "dv",
					ProjectId:     "8",
					OriginPvcName: "pvc-2",
				})
				Expect(err).To(BeNil())
				Expect(res.Code).To(BeEquivalentTo(http.StatusCreated))
				dv := res.Body.(api.Datavolume)
				dataVolumeIds = append(dataVolumeIds, dv.Id)

				res, err = datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", 0, 1000, "", "", []string{}, "", false)
				Expect(err).To(BeNil())
				Expect(res.Code).To(BeEquivalentTo(http.StatusOK))
				dvs := res.Body.(api.GetDatavolumes200Response).Datavolumes
				Expect(dvs).To(HaveLen(1))
				Expect(dvs[0].Id).To(BeEquivalentTo(dv.Id))
			})

			It("should return also internal datavolumes", func() {
				request := api.InternalDatavolumeCreationFields{
					Name:          "internal-dv",
					Scope:         "cluster",
					ClusterId:     clusterId4,
					OriginPvcName: "pvc-1",
				}
				res, err := datavolumesApiService.InternalCreateDatavolume(ctx, request)
				Expect(err).To(BeNil())
				Expect(res.Code).To(BeEquivalentTo(http.StatusOK))
				internalDv := res.Body.(*models.Datavolume)
				dataVolumeIds = append(dataVolumeIds, internalDv.ID.String())

				res, err = datavolumesApiService.CreateDatavolume(ctx, api.DatavolumeCreationFields{
					Name:          "dv",
					ProjectId:     "8",
					OriginPvcName: "pvc-2",
				})
				Expect(err).To(BeNil())
				Expect(res.Code).To(BeEquivalentTo(http.StatusCreated))
				dv := res.Body.(api.Datavolume)
				dataVolumeIds = append(dataVolumeIds, dv.Id)

				res, err = datavolumesApiService.GetDatavolumes(ctx, api.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", 0, 1000, "", "", []string{}, "", true)
				Expect(err).To(BeNil())
				Expect(res.Code).To(BeEquivalentTo(http.StatusOK))
				dvs := res.Body.(api.GetDatavolumes200Response).Datavolumes
				Expect(dvs).To(HaveLen(2))
			})
		})
	})
})

func validateDatavolume(actual api.Datavolume, requested api.DatavolumeCreationFields, projectToAncestors map[string][]string) {
	validateDatavolumeInner(actual, requested, projectToAncestors, true, true)
}

func validateDatavolume2(actual api.Datavolume, requested api.DatavolumeCreationFields, projectToAncestors map[string][]string) {
	validateDatavolumeInner(actual, requested, projectToAncestors, false, false)
}

func validateDatavolumeInner(actual api.Datavolume, requested api.DatavolumeCreationFields, projectToAncestors map[string][]string, validatePhase, validateUpdatedAt bool) {
	_, err := uuid.Parse(actual.Id)
	Expect(err).To(BeNil(), "Datavolume ID should be a valid UUID")
	Expect(actual.CreatedAt).ToNot(BeNil(), "CreatedAt")
	Expect(actual.UpdatedAt).ToNot(BeNil(), "UpdatedAt")
	Expect(actual.Name).To(Equal(requested.Name), "Name")
	Expect(actual.Description).To(Equal(requested.Description), "Description")
	Expect(actual.ProjectId).To(Equal(requested.ProjectId), "ProjectId")
	Expect(actual.OriginPvcName).To(Equal(requested.OriginPvcName), "OriginPvcName")
	Expect(actual.ShouldDeleteOriginalVolume).To(Equal(requested.ShouldDeleteOriginalVolume), "ShouldDeleteOriginalVolume")

	Expect(actual.CreatedBy).To(Equal(TestUser), "CreatedBy")
	Expect(actual.UpdatedBy).To(Equal(TestUser), "UpdatedBy")
	Expect(actual.ProjectName).To(Equal(getTestScopeName(requested.ProjectId, api.SCOPETYPE_PROJECT)), "ProjectName")
	Expect(actual.DepartmentId).To(Equal(projectToAncestors[requested.ProjectId][0]), "DepartmentId")
	Expect(actual.ClusterId).To(Equal(projectToAncestors[requested.ProjectId][1]), "ClusterId")

	if validateUpdatedAt {
		Expect(actual.UpdatedAt).To(Equal(actual.CreatedAt), "CreatedAt == UpdatedAt")
	}

	if validatePhase {
		Expect(actual.Status.Phase).To(Equal(api.DATAVOLUMEPHASE_CREATING), "Phase")
		Expect(actual.Status.PhaseMessage).To(Equal("Creating datavolume in the cluster"), "PhaseMessage")
	}
}

func validateSharedScopes(originalPutFields []api.SharedScopeBase, actualSharedScopes []api.SharedScope, timeBeforeCreation time.Time) {
	ExpectWithOffset(1, actualSharedScopes).To(HaveLen(len(originalPutFields)))

	actualMap := sharedScopesToMap(actualSharedScopes)
	for _, original := range originalPutFields {
		actualSharedScope, ok := actualMap[original]
		Expect(ok).To(BeTrue())

		validateSharedScope(original, actualSharedScope, timeBeforeCreation)
	}
}

func validateSharedScope(original api.SharedScopeBase, actual api.SharedScope, timeBeforeCreation time.Time) {
	ExpectWithOffset(2, actual.ScopeId).To(Equal(original.ScopeId))
	ExpectWithOffset(2, actual.ScopeType).To(Equal(original.ScopeType))
	ExpectWithOffset(2, actual.ScopeName).To(Equal(getTestScopeName(original.ScopeId, original.ScopeType)))
	ExpectWithOffset(2, actual.CreatedBy).To(Equal(TestUser))
	if !timeBeforeCreation.IsZero() {
		ExpectWithOffset(2, actual.CreatedAt.After(timeBeforeCreation)).To(BeTrue())
	}
}

func validateStatus(internalStatusRequest *api.DatavolumeStatusInternal, actual *api.Datavolume) {
	Expect(actual.Status.Phase).To(Equal(internalStatusRequest.Phase), "Phase")
	Expect(actual.Status.PhaseMessage).To(Equal(internalStatusRequest.PhaseMessage), "PhaseMessage")
	Expect(actual.Status.DatavolumePvcName).To(Equal(internalStatusRequest.DatavolumePvcName), "DatavolumePvcName")
	Expect(actual.Status.DatavolumePvName).To(Equal(internalStatusRequest.DatavolumePvName), "DatavolumePvName")
	validateConditions(internalStatusRequest.Conditions, actual.Status.Conditions)
}

func validateConditions(expected, actual []api.Condition) {
	Expect(actual).To(HaveLen(len(expected)))

	for i := range expected {
		Expect(actual[i].Type).To(Equal(expected[i].Type), "condition type")
		Expect(actual[i].Status).To(Equal(expected[i].Status), "condition status")
		Expect(actual[i].Reason).To(Equal(expected[i].Reason), "condition reason")
		Expect(actual[i].Message).To(Equal(expected[i].Message), "condition message")
		Expect(actual[i].LastTransitionTime).ToNot(BeNil())
		Expect(expected[i].LastTransitionTime).ToNot(BeNil())
		Expect(*actual[i].LastTransitionTime).To(BeTemporally("~", *expected[i].LastTransitionTime, time.Millisecond))
	}
}

var _ = Describe("validateAndFixCreationRequest", func() {
	Context("when input is nil", func() {
		It("should return an error", func() {
			err := services.ValidateCreationRequestDefaults(nil)
			Expect(err).ToNot(BeNil())
			Expect(err.Message).To(Equal("Invalid request missing datavolume creation fields"))
		})
	})

	Context("when UseOriginalVolume is nil", func() {
		It("should set UseOriginalVolume to true", func() {
			input := &api.DatavolumeCreationFields{
				ShouldDeleteOriginalVolume: false,
				UseOriginalVolume:          nil,
			}
			err := services.ValidateCreationRequestDefaults(input)
			Expect(err).To(BeNil())
			Expect(input.UseOriginalVolume).ToNot(BeNil())
			Expect(*input.UseOriginalVolume).To(BeTrue())
		})
	})

	Context("when input is valid with UseOriginalVolume set", func() {
		It("should not modify the input", func() {
			input := &api.DatavolumeCreationFields{
				ShouldDeleteOriginalVolume: false,
				UseOriginalVolume:          convutils.Ref(false),
			}
			err := services.ValidateCreationRequestDefaults(input)
			Expect(err).To(BeNil())
			Expect(input.UseOriginalVolume).ToNot(BeNil())
			Expect(*input.UseOriginalVolume).To(BeFalse())
		})
	})
})
