/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datavolumes

import (
	"context"
	ginkgoTesting "testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/backend_db"
	dbTesting "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/query_types"
	authClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	dvapi "github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client/generated"
	dvclientmock "github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client/mocks"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/integdb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
)

func TestDataVolumes(t *ginkgoTesting.T) {
	config.MockConfig()
	config.Init()
	RegisterFailHandler(Fail)
	conf, cleanup, setupErr := integdb.SetupPostgresDB(&integdb.IntegdbConfig{Schema: "assets", MigrationDirPath: "file://../database/migrations"})
	if setupErr != nil {
		t.Fatalf("failed to set up test postgres: %v", setupErr)
	}
	t.Cleanup(cleanup)
	// tests run against a single-user DB; use the owner role (not runaidb's limited
	// service session) so runtime AutoMigrate can create tables.
	conf.UseServiceSession = new(bool)
	if initErr := runaidb.Init(conf); initErr != nil {
		t.Fatalf("failed to init database: %v", initErr)
	}
	t.Cleanup(func() { _ = runaidb.Close() })

	backend_db.CreateTables(runaidb.DB().Actual())

	authClient.SetMockAuthorizationClient(auth.SimpleMockAuthService{})

	RunSpecs(t, "Test data volumes package")
}

var _ = Describe("Test data volumes adapter", Ordered, func() {

	var adapter *Adapter // adapter under test
	var dvClientMock *dvclientmock.IDatavolumesClient

	Context("Test data volumes adapter", func() {

		BeforeEach(func() {
			dvClientMock = &dvclientmock.IDatavolumesClient{}
			adapter = NewAdapter(dvClientMock)
		})

		AfterEach(func() {
			dbTesting.CleanTables()
		})

		It("Basic flow", func() {
			ctx := context.Background()

			mockResp := dvapi.GetDatavolumes200Response{
				Next: nil,
				Datavolumes: []dvapi.DatavolumeNoSharedScopes{
					{
						Id:            "dv-id1",
						Name:          "dv-1",
						Description:   utils.Ref("desciprtion-1"),
						ProjectId:     "4500001",
						ClusterId:     utils.Ref("cluster-id-1"),
						CreatedAt:     utils.Ref(time.Now().AddDate(0, 0, -1)),
						CreatedBy:     utils.Ref("tester1"),
						UpdatedAt:     utils.Ref(time.Now()),
						UpdatedBy:     utils.Ref("tester2"),
						OriginPvcName: "origin-pvc-1",
						Status: dvapi.DatavolumeStatus{
							Phase:        dvapi.DATAVOLUMEPHASE_READY,
							PhaseMessage: utils.Ref("Ready for action"),
						},
					},
					{
						Id:            "dv-id2",
						Name:          "dv-2",
						Description:   utils.Ref("desciprtion-2"),
						ProjectId:     "4500002",
						ClusterId:     utils.Ref("cluster-id-1"),
						CreatedAt:     utils.Ref(time.Now().AddDate(0, 0, -1)),
						CreatedBy:     utils.Ref("tester1"),
						UpdatedAt:     utils.Ref(time.Now()),
						UpdatedBy:     utils.Ref("tester2"),
						OriginPvcName: "origin-pvc-2",
						Status: dvapi.DatavolumeStatus{
							Phase: dvapi.DATAVOLUMEPHASE_CREATING,
						},
					},
				},
			}
			expected := []api.DatasourceListResponseEntry{
				{
					Meta: api.AssetMeta{
						Name:        mockResp.Datavolumes[0].Name,
						Description: mockResp.Datavolumes[0].Description,
						Scope:       api.SCOPE_PROJECT,
						ProjectId:   utils.Ref(int32(4500001)),
						ClusterId:   mockResp.Datavolumes[0].ClusterId,
						Id:          mockResp.Datavolumes[0].Id,
						Kind:        api.ASSETKIND_DATA_VOLUME,
						CreatedBy:   convutils.Deref(mockResp.Datavolumes[0].CreatedBy, ""),
						CreatedAt:   convutils.Deref(mockResp.Datavolumes[0].CreatedAt, time.Time{}),
						UpdatedBy:   convutils.Deref(mockResp.Datavolumes[0].UpdatedBy, ""),
						UpdatedAt:   convutils.Deref(mockResp.Datavolumes[0].UpdatedAt, time.Time{}),
					},
					Spec: api.DatasourceListResponseAssetSpec{
						DataVolume: &api.DataVolumeAssetSpec{
							Id:            convutils.Ref(mockResp.Datavolumes[0].Id),
							OriginPvcName: convutils.Ref(mockResp.Datavolumes[0].OriginPvcName),
						},
					},
					Status: &api.AssetClusterStatusInfo{
						Status:  dataVolumePhaseToAssetStatus(mockResp.Datavolumes[0].Status.Phase),
						Message: convutils.Deref(mockResp.Datavolumes[0].Status.PhaseMessage, ""),
					},
				},
				{
					Meta: api.AssetMeta{
						Name:        mockResp.Datavolumes[1].Name,
						Description: mockResp.Datavolumes[1].Description,
						Scope:       api.SCOPE_PROJECT,
						ProjectId:   utils.Ref(int32(4500001)), // even that the dv is set on project id 4500002, the asset is set on the requested project 4500001
						ClusterId:   mockResp.Datavolumes[1].ClusterId,
						Id:          mockResp.Datavolumes[1].Id,
						Kind:        api.ASSETKIND_DATA_VOLUME,
						CreatedBy:   convutils.Deref(mockResp.Datavolumes[1].CreatedBy, ""),
						CreatedAt:   convutils.Deref(mockResp.Datavolumes[1].CreatedAt, time.Time{}),
						UpdatedBy:   convutils.Deref(mockResp.Datavolumes[1].UpdatedBy, ""),
						UpdatedAt:   convutils.Deref(mockResp.Datavolumes[1].UpdatedAt, time.Time{}),
					},
					Spec: api.DatasourceListResponseAssetSpec{
						DataVolume: &api.DataVolumeAssetSpec{
							Id:            convutils.Ref(mockResp.Datavolumes[1].Id),
							OriginPvcName: convutils.Ref(mockResp.Datavolumes[1].OriginPvcName),
						},
					},
					Status: &api.AssetClusterStatusInfo{
						Status:  dataVolumePhaseToAssetStatus(mockResp.Datavolumes[1].Status.Phase),
						Message: convutils.Deref(mockResp.Datavolumes[1].Status.PhaseMessage, ""),
					},
				},
			}

			dvClientMock.On("GetDatavolumes", ctx, dvapi.DATAVOLUMEREQUESTTYPE_USABLE_IN_PROJECT, "4500001", int32(LimitDvRequestSize), int32(0), []string(nil), "", "").
				Return(&mockResp, nil)
			actual, err := adapter.ListDataVolumes(ctx, &query_types.QueryParams{
				ProjectId: utils.Ref(int32(4500001)),
			})
			Expect(err).To(BeNil())
			Expect(actual).To(BeEquivalentTo(expected))

		})

		It("Pagation flow, less than 1000", func() {
			ctx := context.Background()
			var overideRequsetLimit = int32(10)
			mockResps := []dvapi.GetDatavolumes200Response{
				// sending a slice with 10 zero-value data-volumes as we don't care about the content, just as long as there are something there
				{Next: utils.Ref(int32(10)), Datavolumes: make([]dvapi.DatavolumeNoSharedScopes, overideRequsetLimit)},
				{Next: utils.Ref(int32(20)), Datavolumes: make([]dvapi.DatavolumeNoSharedScopes, overideRequsetLimit)},
				//  last resp doesn't have another page
				{Next: nil, Datavolumes: make([]dvapi.DatavolumeNoSharedScopes, 2)},
			}
			dvClientMock.On("GetDatavolumes", ctx, dvapi.DATAVOLUMEREQUESTTYPE_USABLE_IN_PROJECT, "4500001", overideRequsetLimit, int32(0), []string(nil), "", "").
				Return(&mockResps[0], nil)
			dvClientMock.On("GetDatavolumes", ctx, dvapi.DATAVOLUMEREQUESTTYPE_USABLE_IN_PROJECT, "4500001", overideRequsetLimit, int32(10), []string(nil), "", "").
				Return(&mockResps[1], nil)
			dvClientMock.On("GetDatavolumes", ctx, dvapi.DATAVOLUMEREQUESTTYPE_USABLE_IN_PROJECT, "4500001", overideRequsetLimit, int32(20), []string(nil), "", "").
				Return(&mockResps[2], nil)

			result, err := adapter.aggregateDataVolumes(ctx, "4500001", nil, MaxDvResultCount, int(overideRequsetLimit))
			Expect(err).To(BeNil())
			Expect(len(result)).To(Equal(22))
		})

		It("Pagation flow, more than max limit", func() {
			ctx := context.Background()
			var overideRequsetLimit = int32(10)
			var overrideMaxResult = 40
			mockResps := []dvapi.GetDatavolumes200Response{
				{Next: utils.Ref(int32(10)), Datavolumes: make([]dvapi.DatavolumeNoSharedScopes, overideRequsetLimit)},
				{Next: utils.Ref(int32(20)), Datavolumes: make([]dvapi.DatavolumeNoSharedScopes, overideRequsetLimit)},
				{Next: utils.Ref(int32(30)), Datavolumes: make([]dvapi.DatavolumeNoSharedScopes, overideRequsetLimit)},
				// last resp have another page, but it doesn't get called as the max result was reach
				{Next: utils.Ref(int32(40)), Datavolumes: make([]dvapi.DatavolumeNoSharedScopes, overideRequsetLimit)},
			}
			dvClientMock.On("GetDatavolumes", ctx, dvapi.DATAVOLUMEREQUESTTYPE_USABLE_IN_PROJECT, "4500001", overideRequsetLimit, int32(0), []string(nil), "", "").
				Return(&mockResps[0], nil)
			dvClientMock.On("GetDatavolumes", ctx, dvapi.DATAVOLUMEREQUESTTYPE_USABLE_IN_PROJECT, "4500001", overideRequsetLimit, int32(10), []string(nil), "", "").
				Return(&mockResps[1], nil)
			dvClientMock.On("GetDatavolumes", ctx, dvapi.DATAVOLUMEREQUESTTYPE_USABLE_IN_PROJECT, "4500001", overideRequsetLimit, int32(20), []string(nil), "", "").
				Return(&mockResps[2], nil)
			dvClientMock.On("GetDatavolumes", ctx, dvapi.DATAVOLUMEREQUESTTYPE_USABLE_IN_PROJECT, "4500001", overideRequsetLimit, int32(30), []string(nil), "", "").
				Return(&mockResps[2], nil)

			result, err := adapter.aggregateDataVolumes(ctx, "4500001", nil, overrideMaxResult, int(overideRequsetLimit))
			Expect(err).To(BeNil())
			Expect(len(result)).To(Equal(40))
		})

	})
})
