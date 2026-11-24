/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/org_unit"
	events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/google/uuid"

	k8sObjectsTracker "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/k8s-objects-tracker"
	dbTesting "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/testing"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	workloadsServiceClient "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/workloads-service"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	policies "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	urlNameAvailability = "/api/v1/asset/name-availability"

	urlModelHuggingFaceInfo = "/api/v2/huggingface-model/info-request"
)

var (
	forWorkspaces = Entry(auth.WorkloadTypeWorkspace, auth.WorkloadTypeWorkspace)
	forTrainings  = Entry(auth.WorkloadTypeTraining, auth.WorkloadTypeTraining)
)

// seedLegacyWorkload supplies historical rows to the readers that outlive the asset submission API.
func seedLegacyWorkload(r *testutils.RandomValues, workloadType, name string, assets ...api.AssetMeta) string {
	GinkgoHelper()
	rec := dbmodel.WorkloadRec{
		WorkloadType: workloadType,
		WorkloadMetaV1: api.WorkloadMetaV1{
			Id:        uuid.NewString(),
			Name:      name,
			TenantId:  int32(r.TenantId),
			ProjectId: int32(r.ProjectId),
			ClusterId: r.ClusterId,
		},
	}
	db := runaidb.DB().Actual()
	Expect(db.Create(&rec).Error).To(Succeed())
	for _, asset := range assets {
		link := dbmodel.WorkloadsAssetsRec{
			WorkloadId: rec.Id,
			AssetId:    asset.Id,
			AssetKind:  asset.Kind,
			AutoDelete: convutils.Deref(asset.AutoDelete, false),
		}
		Expect(db.Create(&link).Error).To(Succeed())
	}
	return rec.Id
}

var _ = Describe("Asset names and historical workload records", func() {
	var randoms *testutils.RandomValues
	var differentProject *testutils.RandomValues
	var differentDepartment *testutils.RandomValues

	var envAsset api.EnvironmentAsset

	router := CreateRouter()

	BeforeEach(func() {
		dbTesting.CleanTables()

		policies.SetMock(policies.SimpleMockClientFactory{
			SimpleMockClient: policies.SimpleMockClient{
				EffectiveUpdatedAt: time.Now(),
			},
		})

		randoms = testutils.NewRandomValues()
		differentProject = randoms.CloneWithDifferentProject()
		differentDepartment = differentProject.CloneWithDifferentProject().CloneWithDifferentDepartment()

		authMock := testing.NewRandomsBasedAuthMockClient(randoms, differentProject, differentDepartment)
		org_unit.SetMock(org_unit.NewSimpleMock(randoms, differentProject, differentDepartment))
		authorization.SetMock(authMock)
		workloadsServiceClient.SetMock(workloadsServiceClient.NewSimpleMockClientFactory())

		k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient())

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)))
		checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")

		auth.TestingAllowCreateSystemScopedAssets = true

		envAssetCreationReq := buildRandomTenantEnvironmentAsset(randoms)
		envAsset, _, _ = createEnvironmentAsset(envAssetCreationReq, router, headerForAdmin(randoms))
	})

	Context("Handle project deletion", func() {
		It("basic flow", func() {
			nonTargetId := seedLegacyWorkload(randoms, auth.WorkloadTypeInference, "inference")
			targetId := seedLegacyWorkload(differentProject, auth.WorkloadTypeInference, "inference")

			expectHistoricalWorkloadCount(nonTargetId, 1)
			expectHistoricalWorkloadCount(targetId, 1)

			ctx := context.WithValue(context.Background(), runaictx.CfTenantId, differentProject.TenantIdStr)
			err := HandleProjectDeleteRequest(ctx, &events.ProjectDeleteRequest{
				ProjectId: differentProject.ProjectIdStr,
			})
			Expect(err).To(BeNil())

			expectHistoricalWorkloadCount(nonTargetId, 1)
			expectHistoricalWorkloadCount(targetId, 0)
		})
	})

	Context("Handle cluster deletion", func() {
		It("basic flow", func() {
			workloadId := seedLegacyWorkload(randoms, auth.WorkloadTypeInference, "inference")

			expectHistoricalWorkloadCount(workloadId, 1)

			ctx := context.WithValue(context.Background(), runaictx.CfTenantId, randoms.TenantIdStr)
			err := HandleClusterDeleteEvent(ctx, &events.ClusterDelete{
				ClusterId: randoms.ClusterId,
			})
			Expect(err).To(BeNil())

			expectHistoricalWorkloadCount(workloadId, 0)
		})
	})

	Context("Handle tenant deletion", func() {
		It("basic flow", func() {
			workloadId := seedLegacyWorkload(randoms, auth.WorkloadTypeInference, "inference")

			expectHistoricalWorkloadCount(workloadId, 1)

			ctx := context.WithValue(context.Background(), runaictx.CfTenantId, randoms.TenantIdStr)
			err := HandleTenantDeleteEvent(ctx, &events.TenantDelete{
				TenantId: randoms.TenantIdStr,
			})
			Expect(err).To(BeNil())

			expectHistoricalWorkloadCount(workloadId, 0)
		})
	})

	Context("name validation - mandatory parameters", func() {
		It("should give bad request error if they are missing", func() {
			Expect(assetNameValidation(router, randoms, "assetName=any&scopeType=tenant")).To(Equal(http.StatusBadRequest), "missing kind")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=tenant")).To(Equal(http.StatusBadRequest), "missing name")
			Expect(assetNameValidation(router, randoms, "kind=environment&name=abc")).To(Equal(http.StatusBadRequest), "missing scope")
		})
	})

	Context("name validation - scope id is mandatory for all scope but system and tenant", func() {
		It("should give bad request error", func() {
			Expect(assetNameValidation(router, randoms, "kind=environment&assetName=any&scopeType=system")).To(
				Equal(http.StatusNoContent), "system + no scope id - ok")
			Expect(assetNameValidation(router, randoms, "kind=environment&assetName=any&scopeType=tenant")).To(
				Equal(http.StatusNoContent), "tenant + no scope id - ok")
			Expect(assetNameValidation(router, randoms, fmt.Sprintf("kind=environment&assetName=any&scopeType=cluster&scopeId=%s", uuid.NewString()))).To(
				Equal(http.StatusNoContent), "cluster + no scope id - ok")
			Expect(assetNameValidation(router, randoms, "kind=compute&assetName=tenant-asset&scopeType=department")).To(
				Equal(http.StatusBadRequest), "department + no scope id")
			Expect(assetNameValidation(router, randoms, "kind=compute&assetName=tenant-asset&scopeType=project")).To(
				Equal(http.StatusBadRequest), "project + no scope id")
		})
	})

	Context("name validation - system", func() {
		It("can be reused by another scope", func() {
			envAsset := buildRandomSystemEnvironmentAsset(randoms)
			envAsset.Meta.Name = "xyz"
			createEnvironmentAsset(envAsset, router, headerForAdmin(randoms))

			sameName := "&assetName=xyz"
			diffName := "&assetName=abc"

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=system"+diffName)).To(
				Equal(http.StatusNoContent), "diffName")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=system"+sameName)).To(
				Equal(http.StatusConflict), "sameName")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=tenant"+sameName)).To(
				Equal(http.StatusNoContent), "sameName")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=department&scopeId="+randoms.DepartmentId+sameName)).To(
				Equal(http.StatusNoContent), "sameName")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=project&scopeId="+randoms.ProjectIdStr+sameName)).To(
				Equal(http.StatusNoContent), "sameName")
		})
	})

	Context("name validation - tenant", func() {
		It("single occurrence entire system", func() {
			envAsset := buildRandomTenantEnvironmentAsset(randoms)
			envAsset.Meta.Name = "xyz"
			createEnvironmentAsset(envAsset, router, headerForAdmin(randoms))

			sameName := "&assetName=xyz"
			diffName := "&assetName=abc"

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=tenant"+diffName)).To(
				Equal(http.StatusNoContent), "diffName")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=system"+sameName)).To(
				Equal(http.StatusNoContent), "sameName")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=tenant"+sameName)).To(
				Equal(http.StatusConflict), "sameName")
			Expect(assetNameValidation(router, randoms, fmt.Sprintf("kind=environment&scopeType=cluster&scopeId=%s"+sameName, uuid.NewString()))).To(
				Equal(http.StatusConflict), "sameName")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=department&scopeId="+randoms.DepartmentId+sameName)).To(
				Equal(http.StatusConflict), "sameName")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=project&scopeId="+randoms.ProjectIdStr+sameName)).To(
				Equal(http.StatusConflict), "sameName")
		})
	})

	Context("name validation - cluster", func() {
		It("single occurrence entire system", func() {
			envAsset := buildRandomClusterEnvironmentAsset(randoms)
			envAsset.Meta.Name = "xyz"
			createEnvironmentAsset(envAsset, router, headerForAdmin(randoms))

			sameName := "&assetName=xyz"
			diffName := "&assetName=abc"

			sameScope := "&scopeId=" + fmt.Sprintf("%s", convutils.Deref(envAsset.Meta.ClusterId, ""))
			diffScope := "&scopeId=" + fmt.Sprintf("%s", uuid.NewString())

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=system"+sameName)).To(
				Equal(http.StatusNoContent), "sameName")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=tenant"+sameName)).To(
				Equal(http.StatusConflict), "sameName")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=tenant"+diffName)).To(
				Equal(http.StatusNoContent), "diffName")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=cluster"+sameScope+sameName)).To(
				Equal(http.StatusConflict), "sameName")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=cluster"+sameScope+diffName)).To(
				Equal(http.StatusNoContent), "diffName")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=cluster"+diffScope+sameName)).To(
				Equal(http.StatusNoContent), "sameName")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=department&scopeId="+randoms.DepartmentId+sameName)).To(
				Equal(http.StatusConflict), "sameName")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=project&scopeId="+randoms.ProjectIdStr+sameName)).To(
				Equal(http.StatusConflict), "sameName")
		})
	})

	Context("name validation - department", func() {
		It("only different dept is accepted", func() {
			envAsset := buildRandomDepartmentEnvironmentAsset(randoms)
			envAsset.Meta.Name = "xyz"
			createEnvironmentAsset(envAsset, router, headerForAdmin(randoms))

			sameName := "&assetName=xyz"
			diffName := "&assetName=abc"

			sameScopeId := fmt.Sprintf("&scopeId=%s", randoms.DepartmentId)
			diffScopeId := fmt.Sprintf("&scopeId=%s", differentDepartment.DepartmentId)

			sameDepProjectId := fmt.Sprintf("&scopeId=%v", randoms.ProjectId)
			diffDepProjectId := fmt.Sprintf("&scopeId=%v", differentDepartment.ProjectId)

			clusterScopeId := "&scopeId=" + randoms.ClusterId

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=department"+diffName+sameScopeId)).To(
				Equal(http.StatusNoContent), "diffName + sameScopeId")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=department"+sameName+diffScopeId)).To(
				Equal(http.StatusNoContent), "sameName + diffScopeId")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=system"+sameName)).To(
				Equal(http.StatusNoContent), "system + sameName")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=tenant"+sameName)).To(
				Equal(http.StatusConflict), "tenant + sameName")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=cluster"+clusterScopeId+sameName)).To(
				Equal(http.StatusConflict), "cluster + sameName")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=department"+sameName+sameScopeId)).To(
				Equal(http.StatusConflict), "department + sameName + sameScopeId")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=project"+sameName+sameDepProjectId)).To(
				Equal(http.StatusConflict), "project + sameName + sameScopeId")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=project"+sameName+diffDepProjectId)).To(
				Equal(http.StatusNoContent), "project + sameName + diffScopeId")
		})
	})

	Context("name validation - project", func() {
		It("only different project is accepted", func() {
			envAsset := buildRandomProjectEnvironmentAsset(randoms)
			envAsset.Meta.Name = "xyz"
			createEnvironmentAsset(envAsset, router, headerForAdmin(randoms))

			sameName := "&assetName=xyz"
			diffName := "&assetName=abc"

			sameScopeId := fmt.Sprintf("&scopeId=%s", randoms.ProjectIdStr)
			diffScopeId := fmt.Sprintf("&scopeId=%s", differentDepartment.ProjectIdStr)

			clusterScopeId := fmt.Sprintf("&scopeId=%s", randoms.ClusterId)

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=project"+diffName+sameScopeId)).To(
				Equal(http.StatusNoContent), "diffName + sameScopeId")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=project"+sameName+diffScopeId)).To(
				Equal(http.StatusNoContent), "sameName + diffScopeId")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=system"+sameName)).To(
				Equal(http.StatusNoContent), "system + sameName")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=tenant"+sameName)).To(
				Equal(http.StatusConflict), "tenant + sameName")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=cluster"+sameName+clusterScopeId)).To(
				Equal(http.StatusConflict), "cluster + sameName")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=project"+sameName+sameScopeId)).To(
				Equal(http.StatusConflict), "project + sameName + sameScopeId")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=project"+sameName+diffScopeId)).To(
				Equal(http.StatusNoContent), "project + sameName + diffScopeId")
		})
	})

	Context("cluster sync helpers (updateWorkloadStatus, syncWorkloadStatus, deleteWorkload)", func() {
		It("updateWorkloadStatus sets workloads.current_state", func() {
			workloadID := seedLegacyWorkload(randoms, auth.WorkloadTypeWorkspace, testutils.RandString(8), envAsset.Meta)
			ctx := adminContextForClusterSync(randoms)
			resp, err := updateWorkloadStatus(ctx, randoms.ClusterId, api.WorkloadStatusUpdate{
				Id:     workloadID,
				Status: string(api.WORKLOADSTATE_SUBMITTED),
			})
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))
			var rec dbmodel.WorkloadRec
			Expect(runaidb.DB().Actual().Where("id = ?", workloadID).First(&rec).Error).To(BeNil())
			Expect(string(rec.CurrentState)).To(Equal(string(api.WORKLOADSTATE_SUBMITTED)))
		})
		It("updateWorkloadStatus returns 404 when workload does not exist", func() {
			ctx := adminContextForClusterSync(randoms)
			resp, err := updateWorkloadStatus(ctx, randoms.ClusterId, api.WorkloadStatusUpdate{
				Id:     uuid.NewString(),
				Status: string(api.WORKLOADSTATE_SUBMITTED),
			})
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusNotFound))
		})
		It("syncWorkloadStatus updates multiple workspaces", func() {
			id1 := seedLegacyWorkload(randoms, auth.WorkloadTypeWorkspace, testutils.RandString(8), envAsset.Meta)
			id2 := seedLegacyWorkload(randoms, auth.WorkloadTypeWorkspace, testutils.RandString(8), envAsset.Meta)
			ctx := adminContextForClusterSync(randoms)
			resp, err := syncWorkloadStatus(ctx, randoms.ClusterId, []api.WorkloadStatusUpdate{
				{Id: id1, Status: string(api.WORKLOADSTATE_SUBMITTED)},
				{Id: id2, Status: string(api.WORKLOADSTATE_FAILED)},
			})
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))
			var r1, r2 dbmodel.WorkloadRec
			Expect(runaidb.DB().Actual().Where("id = ?", id1).First(&r1).Error).To(BeNil())
			Expect(runaidb.DB().Actual().Where("id = ?", id2).First(&r2).Error).To(BeNil())
			Expect(string(r1.CurrentState)).To(Equal(string(api.WORKLOADSTATE_SUBMITTED)))
			Expect(string(r2.CurrentState)).To(Equal(string(api.WORKLOADSTATE_FAILED)))
		})
		It("deleteWorkload removes row and second delete is 404", func() {
			workloadID := seedLegacyWorkload(randoms, auth.WorkloadTypeWorkspace, testutils.RandString(8), envAsset.Meta)
			ctx := adminContextForClusterSync(randoms)
			resp, err := deleteWorkload(ctx, randoms.ClusterId, workloadID)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusNoContent))
			resp2, _ := deleteWorkload(ctx, randoms.ClusterId, workloadID)
			Expect(resp2.Code).To(Equal(http.StatusNotFound))
		})
	})
})

func assetNameValidation(router chi.Router, randoms *testutils.RandomValues, filters string) int {
	validateHttpReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s?%s", urlNameAvailability, filters), nil)
	Expect(err).To(BeNil())

	validateHttpReq.Header = *headerForAdmin(randoms)

	validateHttpResp := httptest.NewRecorder()
	router.ServeHTTP(validateHttpResp, validateHttpReq)
	return validateHttpResp.Code
}

func adminContextForClusterSync(r *testutils.RandomValues) context.Context {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, r.TenantIdStr)
	return context.WithValue(ctx, runaictx.CfSubjectID, "admin")
}

func expectHistoricalWorkloadCount(id string, expected int64) {
	GinkgoHelper()
	var count int64
	Expect(runaidb.DB().Actual().Model(&dbmodel.WorkloadRec{}).Where("id = ?", id).Count(&count).Error).To(Succeed())
	Expect(count).To(Equal(expected))
}
