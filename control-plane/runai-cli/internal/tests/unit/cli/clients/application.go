/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"
	cpmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	applicationservice "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/application"
	assetsopenapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/golang/mock/gomock"
	"k8s.io/utils/ptr"
)

const (
	AppName01 = "app-one"
	AppName02 = "app-two"
	AppName03 = "app-three"

	AppID01 = "app-id-1"

	// applications only reachable through name-filtered lookups (delete/describe flows)
	AppNameDeleting       = "app-deleting"
	AppNameDesiredDeleted = "app-desired-deleted"
	AppNameMissing        = "app-missing"
	AppNameDetailed       = "app-detailed"
	AppNameWithResources  = "app-with-resources"

	// credential fixtures (submit flows)
	CredentialName    = "chart-cred"
	CredentialAssetID = "cred-asset-1"

	// same NGC API key name at two scopes - ambiguous
	CredentialNameAmbiguous     = "dup-cred"
	CredentialAmbiguousFirstID  = "cred-asset-dup-1"
	CredentialAmbiguousSecondID = "cred-asset-dup-2"

	// non-NGC credential kinds are not usable for chart auth
	CredentialNameDockerRegistry = "docker-cred"
)

// LastAIApplicationCreationRequest captures the most recent request passed to
// the mocked creator, so tests can assert on the built spec.
var LastAIApplicationCreationRequest openapi.AIApplicationCreationRequest

var (
	appOne = openapi.AIApplication{
		Id:          ptr.To("app-id-1"),
		Name:        ptr.To(AppName01),
		Type:        ptr.To("inference"),
		ProjectName: ptr.To(ProjectOneName),
		ProjectId:   ptr.To(ProjectOneID),
		CreatedAt:   ptrTime(time.Unix(10, 0)),
	}
	appTwo = openapi.AIApplication{
		Id:          ptr.To("app-id-2"),
		Name:        ptr.To(AppName02),
		Type:        ptr.To("training"),
		ProjectName: ptr.To(ProjectOneName),
		ProjectId:   ptr.To(ProjectOneID),
		CreatedAt:   ptrTime(time.Unix(11, 0)),
	}
	appThree = openapi.AIApplication{
		Id:          ptr.To("app-id-3"),
		Name:        ptr.To(AppName03),
		Type:        ptr.To("inference"),
		ProjectName: ptr.To(ProjectTwoName),
		ProjectId:   ptr.To(ProjectTwoID),
		CreatedAt:   ptrTime(time.Unix(12, 0)),
	}

	AppList = []openapi.AIApplication{appOne, appTwo, appThree}

	appDetailed = func() openapi.AIApplication {
		app := openapi.AIApplication{
			Id:          ptr.To("app-id-detailed"),
			Name:        ptr.To(AppNameDetailed),
			Type:        ptr.To("inference"),
			Source:      ptr.To("helm"),
			ProjectName: ptr.To(ProjectOneName),
			ProjectId:   ptr.To(ProjectOneID),
			CreatedAt:   ptrTime(time.Unix(20, 0)),
			CreatedBy:   ptr.To("test@example.com"),
		}
		app.ActionStatus.Set(ptr.To("deployed"))
		app.ActionMessage.Set(ptr.To("install completed"))
		app.Spec = &openapi.AIApplicationSpec{
			ChartRef: openapi.AIApplicationSpecChartRef{
				ChartURL: ptr.To("https://charts.example.com/grafana-11.0.0.tgz"),
			},
		}
		app.Resources = []openapi.AIApplicationResource{
			{Id: ptr.To("res-id-1"), Name: ptr.To("grafana-server"), Kind: ptr.To("Deployment")},
		}
		return app
	}()

	appWithResources = func() openapi.AIApplication {
		app := openapi.AIApplication{
			Id:          ptr.To("app-id-resources"),
			Name:        ptr.To(AppNameWithResources),
			Type:        ptr.To("inference"),
			Source:      ptr.To("helm"),
			ProjectName: ptr.To(ProjectOneName),
			ProjectId:   ptr.To(ProjectOneID),
		}
		reqGpu := openapi.RequestResourceCores{}
		reqGpu.Request.Set(ptr.To(2.0))
		reqCpuMem := openapi.RequestResourceQuantity{}
		reqCpuMem.Request.Set(ptr.To("4Gi"))
		reqRes := openapi.AIApplicationRequestResources{}
		reqRes.Gpu.Set(&reqGpu)
		reqRes.CpuMemory.Set(&reqCpuMem)
		app.RequestedResources.Set(&reqRes)

		allocRes := openapi.AIApplicationAllocatedResources{}
		allocRes.Gpu.Set(ptr.To(1.5))
		allocRes.GpuMemory.Set(ptr.To("3Gi"))
		app.AllocatedResources.Set(&allocRes)
		return app
	}()

	appDeleting = func() openapi.AIApplication {
		app := openapi.AIApplication{
			Id:          ptr.To("app-id-deleting"),
			Name:        ptr.To(AppNameDeleting),
			ProjectName: ptr.To(ProjectOneName),
			ProjectId:   ptr.To(ProjectOneID),
		}
		app.ActionStatus.Set(ptr.To("deleting"))
		return app
	}()

	appDesiredDeleted = func() openapi.AIApplication {
		app := openapi.AIApplication{
			Id:          ptr.To("app-id-desired-deleted"),
			Name:        ptr.To(AppNameDesiredDeleted),
			ProjectName: ptr.To(ProjectOneName),
			ProjectId:   ptr.To(ProjectOneID),
		}
		app.DesiredPhase.Set(ptr.To("deleted"))
		return app
	}()
)

func ptrTime(t time.Time) *time.Time {
	return &t
}

func MockAIApplication(mockCtrl *gomock.Controller) {
	mockLister := cpmocks.NewMockAIApplicationLister(mockCtrl)

	allFilter := filter.NewBuilder().WithClusterID(ClusterUuid).StringList()
	proj1Filter := filter.NewBuilder().WithClusterID(ClusterUuid).WithProject(ProjectOneName).StringList()
	proj2Filter := filter.NewBuilder().WithClusterID(ClusterUuid).WithProject(ProjectTwoName).StringList()

	allResp := &openapi.ListApplications200Response{
		Applications: AppList,
	}
	proj1Resp := &openapi.ListApplications200Response{
		Applications: []openapi.AIApplication{appOne, appTwo},
	}
	proj2Resp := &openapi.ListApplications200Response{
		Applications: []openapi.AIApplication{appThree},
	}

	mockLister.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(allFilter)).Return(allResp, nil).AnyTimes()
	mockLister.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(proj1Filter)).Return(proj1Resp, nil).AnyTimes()
	mockLister.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(proj2Filter)).Return(proj2Resp, nil).AnyTimes()

	// name-scoped lookups used by the delete flow
	byName := func(name string) []string {
		return filter.NewBuilder().WithClusterID(ClusterUuid).WithProject(ProjectOneName).AppendName(name).StringList()
	}
	mockLister.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(byName(AppName01))).
		Return(&openapi.ListApplications200Response{Applications: []openapi.AIApplication{appOne}}, nil).AnyTimes()
	mockLister.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(byName(AppNameDetailed))).
		Return(&openapi.ListApplications200Response{Applications: []openapi.AIApplication{appDetailed}}, nil).AnyTimes()
	mockLister.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(byName(AppNameWithResources))).
		Return(&openapi.ListApplications200Response{Applications: []openapi.AIApplication{appWithResources}}, nil).AnyTimes()
	mockLister.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(byName(AppNameDeleting))).
		Return(&openapi.ListApplications200Response{Applications: []openapi.AIApplication{appDeleting}}, nil).AnyTimes()
	mockLister.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(byName(AppNameDesiredDeleted))).
		Return(&openapi.ListApplications200Response{Applications: []openapi.AIApplication{appDesiredDeleted}}, nil).AnyTimes()
	mockLister.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(byName(AppNameMissing))).
		Return(&openapi.ListApplications200Response{Applications: []openapi.AIApplication{}}, nil).AnyTimes()

	mockDeleter := cpmocks.NewMockAIApplicationDeleter(mockCtrl)
	mockDeleter.EXPECT().Delete(gomock.Any(), AppID01).Return(nil).AnyTimes()

	mockCreator := cpmocks.NewMockAIApplicationCreator(mockCtrl)
	mockCreator.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ interface{}, request openapi.AIApplicationCreationRequest) (*openapi.AIApplication, error) {
			LastAIApplicationCreationRequest = request
			created := openapi.AIApplication{Id: ptr.To("app-id-created"), Name: ptr.To(request.Name)}
			return &created, nil
		}).AnyTimes()

	credentialEntries := []assetsopenapi.CredentialsListResponseEntry{
		{Meta: assetsopenapi.AssetMeta{Id: CredentialAssetID, Name: CredentialName, Kind: assetsopenapi.ASSETKIND_NGC_API_KEY}},
		{Meta: assetsopenapi.AssetMeta{Id: CredentialAmbiguousFirstID, Name: CredentialNameAmbiguous, Kind: assetsopenapi.ASSETKIND_NGC_API_KEY}},
		{Meta: assetsopenapi.AssetMeta{Id: CredentialAmbiguousSecondID, Name: CredentialNameAmbiguous, Kind: assetsopenapi.ASSETKIND_NGC_API_KEY}},
		{Meta: assetsopenapi.AssetMeta{Id: "cred-asset-docker", Name: CredentialNameDockerRegistry, Kind: assetsopenapi.ASSETKIND_DOCKER_REGISTRY}},
	}

	mockCredentialLister := cpmocks.NewMockCredentialLister(mockCtrl)
	mockCredentialLister.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ interface{}, name *string, projectId *int32, clusterId *string, credType *string) (*assetsopenapi.CredentialsListResponse, error) {
			// the lookup must stay scoped to the submitted project, cluster
			// and the supported credential kind - fail loudly otherwise
			expectedProjectId, err := strconv.ParseInt(ProjectOneID, 10, 32)
			if err != nil {
				return nil, err
			}
			if projectId == nil || *projectId != int32(expectedProjectId) {
				return nil, fmt.Errorf("credential lookup lost its project scoping, got projectId %v", projectId)
			}
			if clusterId == nil || *clusterId != ClusterUuid {
				return nil, fmt.Errorf("credential lookup lost its cluster scoping, got clusterId %v", clusterId)
			}
			if credType == nil || *credType != string(assetsopenapi.ASSETKIND_NGC_API_KEY) {
				return nil, fmt.Errorf("credential lookup lost its type scoping, got credType %v", credType)
			}

			matched := []assetsopenapi.CredentialsListResponseEntry{}
			for _, entry := range credentialEntries {
				if name != nil && *name != "" && entry.Meta.Name != *name {
					continue
				}
				if string(entry.Meta.Kind) != *credType {
					continue
				}
				matched = append(matched, entry)
			}
			return &assetsopenapi.CredentialsListResponse{Entries: matched}, nil
		}).AnyTimes()

	applicationservice.SetLister(mockLister)
	applicationservice.SetDeleter(mockDeleter)
	applicationservice.SetCreator(mockCreator)
	applicationservice.SetCredentialLister(mockCredentialLister)
}
