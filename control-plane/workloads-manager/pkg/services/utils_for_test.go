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

	"github.com/go-chi/chi/v5"
	"github.com/jinzhu/copier"
	. "github.com/onsi/gomega"

	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	authorization_client "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	clusterApi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	runaihttp "github.com/run-ai/runai/runai-common-packages/runaictx/http"

	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"

	"github.com/run-ai/runai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/config"
	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/database/models"
	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/middleware"
)

const adminUser = "admin"

var fixedTime = time.Now() // prevent test flakiness

var emptyClusterReports = api.ClusterSyncReport{
	Updates: make([]api.ClusterSyncReportUpdate, 0),
}

func makePayload(payloadObject interface{}) *strings.Reader {
	payloadString, err := json.Marshal(payloadObject)
	Expect(err).To(BeNil())
	return strings.NewReader(string(payloadString))
}

func headerForAdmin(randoms *testutils.RandomValues) *http.Header {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, randoms.TenantIdStr)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, adminUser)
	return runaihttp.GetHeaders(ctx)
}

func createRouterForTest() chi.Router {
	WorkloadsV2Service := NewWorkloadsV2Service()
	WorkloadsV2Controller := api.NewWorkloadsV2APIController(WorkloadsV2Service)

	router := api.NewRouter()

	router.Group(func(router chi.Router) {
		router.Use(runaihttp.ContextMiddleware)
		router.Use(middleware.RequestContextMiddleware)

		api.AddRoutes(router, WorkloadsV2Controller)
	})

	return router
}

func setContext(tenantId string, subjectId string, groups []string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantId)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, subjectId)
	ctx = context.WithValue(ctx, runaictx.CfGroups, groups)
	ctx = context.WithValue(ctx, authorization_client.PermittedScopesKey, &authorization_api.PermittedScopes{
		Tenants: []string{tenantId},
	})

	return ctx
}

func cleanTables() {
	gormDb := dao_utils.DB()
	gormDb.Unscoped().Delete(&models.WorkloadsQueueRecord{}, dao_utils.All)
}

func defaultClusterCacheConfig() *checker.ClustersCacheConfig {
	servicesConfig := convutils.Ref(config.Services{})
	servicesConfig.SetDefaults()

	return &checker.ClustersCacheConfig{
		ExpirationSec: servicesConfig.ClusterCacheExpirationSec,
		CleanupSec:    servicesConfig.ClusterCacheCleanupSec,
	}
}

func clusterInfoOf(randoms *testutils.RandomValues, version string) clusterApi.DisplayedCluster {
	result := clusterApi.DisplayedCluster{
		Uuid: randoms.ClusterId,
	}

	result.Version.Set(&version)

	return result
}

func tryToCreateWorkloadForTest(router chi.Router, randoms *testutils.RandomValues, request *api.WorkloadV2CreateRequest) *httptest.ResponseRecorder {
	createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s", testHost, urlWorkloads), makePayload(request))
	Expect(err).To(BeNil())
	createReq.Header = *headerForAdmin(randoms)

	createResp := httptest.NewRecorder()
	router.ServeHTTP(createResp, createReq)

	return createResp
}

func createWorkloadForTest(router chi.Router, randoms *testutils.RandomValues, request *api.WorkloadV2CreateRequest) (string, *api.WorkloadV2) {
	createResp := tryToCreateWorkloadForTest(router, randoms, request)
	Expect(createResp.Code).To(Equal(http.StatusAccepted), "Create workload failed: %s", createResp.Body.String())

	var result api.WorkloadV2
	jsonErr := json.Unmarshal(createResp.Body.Bytes(), &result)
	Expect(jsonErr).To(BeNil())

	id := sanitizeMeta(&result.Metadata)
	Expect(result.Metadata.DeletedAt).To(BeNil())

	return id, &result
}

func tryToCreateInternalWorkloadForTest(router chi.Router, randoms *testutils.RandomValues, request *api.WorkloadV2CreateRequest) *httptest.ResponseRecorder {
	createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s", testHost, urlInternalWorkloads), makePayload(request))
	Expect(err).To(BeNil())
	createReq.Header = *headerForAdmin(randoms)

	createResp := httptest.NewRecorder()
	router.ServeHTTP(createResp, createReq)

	return createResp
}

func createInternalWorkloadForTest(router chi.Router, randoms *testutils.RandomValues, request *api.WorkloadV2CreateRequest) (string, *api.WorkloadV2) {
	createResp := tryToCreateInternalWorkloadForTest(router, randoms, request)
	Expect(createResp.Code).To(Equal(http.StatusAccepted), "Create internal workload failed: %s", createResp.Body.String())

	var result api.WorkloadV2
	jsonErr := json.Unmarshal(createResp.Body.Bytes(), &result)
	Expect(jsonErr).To(BeNil())

	id := sanitizeMeta(&result.Metadata)
	Expect(result.Metadata.DeletedAt).To(BeNil())

	return id, &result
}

func getWorkloadForTest(router chi.Router, randoms *testutils.RandomValues, id string) *api.WorkloadV2GetResponse {
	getReq, err := http.NewRequest("GET", fmt.Sprintf("%s%s/%s", testHost, urlWorkloads, id), nil)
	Expect(err).To(BeNil())
	getReq.Header = *headerForAdmin(randoms)

	getResp := httptest.NewRecorder()
	router.ServeHTTP(getResp, getReq)
	Expect(getResp.Code).To(Equal(http.StatusOK))

	var result api.WorkloadV2GetResponse
	jsonErr := json.Unmarshal(getResp.Body.Bytes(), &result)
	Expect(jsonErr).To(BeNil())

	sanitizeMeta(&result.Workload.Metadata)
	Expect(result.Workload.Metadata.DeletedAt).To(BeNil())

	return &result
}

func verifyWorkloadNotFound(router chi.Router, randoms *testutils.RandomValues, id string) {
	getReq, err := http.NewRequest("GET", fmt.Sprintf("%s%s/%s", testHost, urlWorkloads, id), nil)
	Expect(err).To(BeNil())
	getReq.Header = *headerForAdmin(randoms)

	getResp := httptest.NewRecorder()
	router.ServeHTTP(getResp, getReq)
	Expect(getResp.Code).To(Equal(http.StatusNotFound))
}

func updateWorkloadForTest(router chi.Router, randoms *testutils.RandomValues, id string, request *api.WorkloadV2UpdateRequest) *api.WorkloadV2 {
	updateResp := tryToUpdateWorkloadForTest(router, randoms, id, request)
	Expect(updateResp.Code).To(Equal(http.StatusAccepted))

	var result api.WorkloadV2
	jsonErr := json.Unmarshal(updateResp.Body.Bytes(), &result)
	Expect(jsonErr).To(BeNil())

	sanitizeMeta(&result.Metadata)

	return &result
}

func tryToUpdateWorkloadForTest(router chi.Router, randoms *testutils.RandomValues, id string, request *api.WorkloadV2UpdateRequest) *httptest.ResponseRecorder {
	return tryToUpdateWorkloadAtPathForTest(router, randoms, urlWorkloads, id, request)
}

func tryToUpdateInternalWorkloadForTest(router chi.Router, randoms *testutils.RandomValues, id string, request *api.WorkloadV2UpdateRequest) *httptest.ResponseRecorder {
	return tryToUpdateWorkloadAtPathForTest(router, randoms, urlInternalWorkloads, id, request)
}

func tryToUpdateWorkloadAtPathForTest(router chi.Router, randoms *testutils.RandomValues, path, id string, request *api.WorkloadV2UpdateRequest) *httptest.ResponseRecorder {
	updateReq, err := http.NewRequest("PUT", fmt.Sprintf("%s%s/%s", testHost, path, id), makePayload(request))
	Expect(err).To(BeNil())
	updateReq.Header = *headerForAdmin(randoms)

	updateResp := httptest.NewRecorder()
	router.ServeHTTP(updateResp, updateReq)

	return updateResp
}

func deleteWorkloadForTest(router chi.Router, randoms *testutils.RandomValues, id string) *api.WorkloadV2 {
	deleteReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s%s/%s", testHost, urlWorkloads, id), nil)
	Expect(err).To(BeNil())
	deleteReq.Header = *headerForAdmin(randoms)

	deleteResp := httptest.NewRecorder()
	router.ServeHTTP(deleteResp, deleteReq)
	Expect(deleteResp.Code).To(Equal(http.StatusAccepted))

	var result api.WorkloadV2
	jsonErr := json.Unmarshal(deleteResp.Body.Bytes(), &result)
	Expect(jsonErr).To(BeNil())

	sanitizeMeta(&result.Metadata)

	return &result
}

func getUpdatesForTest(router chi.Router, randoms *testutils.RandomValues, after int64, limit int32, clusterReport api.ClusterSyncReport) *api.WorkloadsV2Updates {
	updatesReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s/updates?clusterId=%s&updatesAfter=%d&limit=%d",
		testHost, urlWorkloads, randoms.ClusterId, after, limit), makePayload(clusterReport))
	Expect(err).To(BeNil())
	updatesReq.Header = *headerForAdmin(randoms)

	updatesResp := httptest.NewRecorder()
	router.ServeHTTP(updatesResp, updatesReq)
	Expect(updatesResp.Code).To(Equal(http.StatusOK))

	var result api.WorkloadsV2Updates
	jsonErr := json.Unmarshal(updatesResp.Body.Bytes(), &result)
	Expect(jsonErr).To(BeNil())

	return &result
}

func findWorkloadUpdate(updates *api.WorkloadsV2Updates, id string) *api.WorkloadV2Update {
	for i := range updates.Workloads {
		if updates.Workloads[i].Id == id {
			return &updates.Workloads[i]
		}
	}
	return nil
}

func sanitizeMeta(meta *api.WorkloadV2Metadata) string {
	// sanitize times
	Expect(meta.CreatedAt).NotTo(BeNil())
	meta.CreatedAt = fixedTime

	Expect(meta.UpdatedAt).NotTo(BeNil())
	meta.UpdatedAt = fixedTime

	// sanitize id but keep it aside for the caller
	id := meta.Id
	Expect(id).NotTo(Equal(""))
	meta.Id = ""

	return id
}

// since the request meta is missing many details, complement it with the missing data so it can
// be compared with complete meta. sanitize all the fields that can make the test flaky
func requestToResponseMeta(reqMeta *api.WorkloadV2MetadataCreateParams, randoms *testutils.RandomValues,
	subjectId string, gvk api.Gvk, configuration *api.WorkloadV2Configuration) *api.WorkloadV2Metadata {
	var result api.WorkloadV2Metadata

	err := copier.Copy(&result, *reqMeta)
	Expect(err).To(BeNil())

	// scope
	result.TenantId = int32(randoms.TenantId)
	result.ClusterId = randoms.ClusterId
	result.DepartmentId = randoms.DepartmentId
	result.DepartmentName = randoms.DepartmentId
	result.ProjectId = randoms.ProjectIdStr
	result.ProjectName = randoms.ProjectIdStr

	// preemptibility
	if reqMeta.Preemptibility != nil {
		result.Preemptibility = reqMeta.Preemptibility
	} else if convutils.Deref(reqMeta.Priority, "") == "high" {
		result.Preemptibility = convutils.Ref(api.PREEMPTIBILITY_NON_PREEMPTIBLE)
	} else {
		result.Preemptibility = convutils.Ref(api.PREEMPTIBILITY_PREEMPTIBLE)
	}
	// configuration
	if configuration != nil {
		result.Configuration = *configuration
	} else {
		result.Configuration = api.WorkloadV2Configuration{
			Mnnvl: convutils.Ref(api.MNNVLMODE_NONE),
		}
	}
	// gvk
	result.Gvk = gvk

	// user
	result.CreatedBy = subjectId
	result.UpdatedBy = subjectId

	// sanitize times
	Expect(result.CreatedAt).NotTo(BeNil())
	result.CreatedAt = fixedTime

	Expect(result.UpdatedAt).NotTo(BeNil())
	result.UpdatedAt = fixedTime

	Expect(result.DeletedAt).To(BeNil())
	Expect(result.DeletedBy).To(BeNil())

	return &result
}

func metaToUpdate(meta *api.WorkloadV2Metadata, id string, updateCount int64, phase api.DesiredPhase, effectiveManifest map[string]interface{}) api.WorkloadV2Update {
	return api.WorkloadV2Update{
		Id:               id,
		CreatedBy:        meta.CreatedBy,
		ProjectId:        meta.ProjectId,
		Name:             meta.Name,
		Gvk:              meta.Gvk,
		UpdateCount:      updateCount,
		DesiredPhase:     phase,
		WorkloadManifest: effectiveManifest,
		Configuration: api.WorkloadV2UpdateConfiguration{
			Mnnvl: api.WorkloadV2UpdateConfigurationMnnvl{
				Mode: *meta.Configuration.Mnnvl,
			},
		},
	}
}

func responseInfoForTest(low, high int) api.ResponseInfo {
	return api.ResponseInfo{
		LowestUpdate:  int64(low),
		HighestUpdate: int64(high),
	}
}
