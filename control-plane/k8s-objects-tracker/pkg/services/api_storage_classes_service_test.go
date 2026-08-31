/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	runaictx_http "github.com/run-ai/runai/runai-common-packages/runaictx/http"

	authorizationApi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/authorization/pkg/client/structs"
	clustersApi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/clients/clusters"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	audit "github.com/run-ai/runai/control-plane/audit-service/pkg/client"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/clients/authorization"
	orgUnit "github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/clients/org-unit"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/database/dbtesting"
)

var clusterId = uuid.New().String()

const (
	newScUrl     = "/api/v1/storage-classes?clusterId="
	v2ScUrl      = "/api/v2/storage-classes?clusterId="
	newSyncScUrl = "/api/v1/storage-classes/sync?clusterId="

	includeNone = "&includeNone=true"

	dummyTenantId = 1111
)

const v1ScUrl = "/v1/k8s/clusters/%s/storage-classes"

var storageClass = api.StorageClass{
	StorageClassName: "sc name",
	ClusterId:        clusterId,
	Spec: api.StorageClassSpec{
		Provisioner:          "my provisioner",
		AllowVolumeExpansion: false,
		IsDefault:            true,
	},
}

var v1StorageClass = api.V1StorageClass{
	StorageClassName: "v1 sc name",
	IsDefault:        true,
}

var configService = StorageClassConfigurationAPIService{}

var _ = Describe("storage classes test", func() {
	BeforeEach(func() {
		backendMock := NewRandomsBasedBackendMockClient()
		orgUnit.SetMock(backendMock)
		authorization.SetMock(backendMock)
		router = CreateRouter()
		dbtesting.CleanTables()

		clusters.SetMock(clusters.NewSimpleMock().WithCluster(clusterId, &clustersApi.DisplayedCluster{}))
	})

	Context("Storage Class API test", func() {
		scUrl := newScUrl + clusterId
		getUrl := v2ScUrl + clusterId
		deletionUrl := scUrl + "&name=%s"
		syncUrl := newSyncScUrl + clusterId

		It("create 2, list, delete 1, list", func() {
			respRecorder := getStorageClasses(getUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))
			Expect(len(unmarshalListBody(respRecorder.Body).Items)).To(Equal(0))

			respRecorder = createStorageClass("sc1", scUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))

			respRecorder = createStorageClass("sc2", scUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))

			respRecorder = getStorageClasses(getUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))
			scs := unmarshalListBody(respRecorder.Body)
			Expect(len(scs.Items)).To(Equal(2))
			Expect(scs.Items[0].StorageClassName).To(Equal("sc1"))
			Expect(scs.Items[1].StorageClassName).To(Equal("sc2"))
			Expect(scs.Items[1].Spec.Provisioner).To(Equal("my provisioner"))
			Expect(scs.Items[1].Spec.IsDefault).To(Equal(true))
			Expect(scs.Items[1].Spec.AllowVolumeExpansion).To(Equal(false))

			respRecorder = deleteStorageClass("sc1", deletionUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))

			respRecorder = getStorageClasses(getUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))
			scs = unmarshalListBody(respRecorder.Body)
			Expect(len(scs.Items)).To(Equal(1))
			Expect(scs.Items[0].StorageClassName).To(Equal("sc2"))
		})
		It("sync 2, list, sync only 1, list", func() {
			scsList := api.StorageClasses{Items: []api.StorageClass{storageClass, storageClass}}
			scsList.Items[0].StorageClassName = "sc1"
			scsList.Items[1].StorageClassName = "sc2"
			scsList.Items[0].Spec.IsDefault = true
			scsList.Items[1].Spec.IsDefault = false

			respRecorder := getStorageClasses(getUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))
			Expect(len(unmarshalListBody(respRecorder.Body).Items)).To(Equal(0))

			respRecorder = syncStorageClass(scsList, syncUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))

			respRecorder = getStorageClasses(getUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))
			scs := unmarshalListBody(respRecorder.Body)
			Expect(len(scs.Items)).To(Equal(2))
			Expect(scs.Items[0].StorageClassName).To(Equal("sc1"))
			Expect(scs.Items[1].StorageClassName).To(Equal("sc2"))
			Expect(scs.Items[0].Spec.IsDefault).To(Equal(true))
			Expect(scs.Items[1].Spec.IsDefault).To(Equal(false))

			// Now sync only sc2 and modify it
			scsList.Items[1].Spec.IsDefault = true
			scsList.Items = scsList.Items[1:2]
			respRecorder = syncStorageClass(scsList, syncUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))

			respRecorder = getStorageClasses(getUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))
			scs = unmarshalListBody(respRecorder.Body)
			Expect(len(scs.Items)).To(Equal(1))
			Expect(scs.Items[0].StorageClassName).To(Equal("sc2"))
			Expect(scs.Items[0].Spec.IsDefault).To(Equal(true))
		})
	})

	Context("Storage Class API test with permissions", func() {
		scUrl := newScUrl + clusterId
		getUrl := v2ScUrl + clusterId

		It("create 2 storage classes, one having permissions set the other does not", func() {
			respRecorder := getStorageClasses(getUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))
			Expect(len(unmarshalListBody(respRecorder.Body).Items)).To(Equal(0))

			respRecorder = createStorageClass("sc1", scUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))
			sc1 := unmarshalGetBody(respRecorder.Body)
			Expect(sc1).NotTo(BeNil())

			respRecorder = createStorageClass("sc2", scUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))
			sc2 := unmarshalGetBody(respRecorder.Body)
			Expect(sc2).NotTo(BeNil())

			//
			//  re-create sc2 should not have any impact, but we want to see that configurations
			//  are not changed following this operation
			//
			respRecorder = createStorageClass("sc2", scUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))

			//
			//   make sure default configuration was created for it
			//
			resp, _ := configService.ListStorageClassConfiguration(adminContext(), nil, "name", "", 0, 0)
			Expect(resp.Code).To(Equal(http.StatusOK))

			expectedDefaultPermissions := api.StorageClassPermissions{
				AllowedForAssets:            convutils.Ref(true),
				AllowedForWorkloads:         convutils.Ref(true),
				AllowedForPersistentVolumes: convutils.Ref(true),
				AllowedForEphemeralVolumes:  convutils.Ref(true),
				AllowDataSharing:            convutils.Ref(false),
			}

			configs := resp.Body.(api.StorageClassConfigurationListResponse)
			Expect(len(configs.StorageClassConfigurations)).To(Equal(2))
			Expect(configs.StorageClassConfigurations[0].Name).To(Equal("sc1"))
			Expect(configs.StorageClassConfigurations[1].Name).To(Equal("sc2"))
			Expect(configs.StorageClassConfigurations[0].Permissions).NotTo(BeNil())
			Expect(configs.StorageClassConfigurations[1].Permissions).NotTo(BeNil())
			Expect(*configs.StorageClassConfigurations[0].Permissions).To(Equal(expectedDefaultPermissions))
			Expect(*configs.StorageClassConfigurations[0].Permissions).To(Equal(expectedDefaultPermissions))

			sc1Id := configs.StorageClassConfigurations[0].Id

			//
			//   change the defaults bits
			//
			sc1Permissions := api.StorageClassPermissions{
				AllowedForAssets:    convutils.Ref(false),
				AllowedForWorkloads: convutils.Ref(false),
			}

			expectedSc1Permissions := api.StorageClassPermissions{
				AllowedForAssets:            convutils.Ref(false),
				AllowedForWorkloads:         convutils.Ref(false),
				AllowedForPersistentVolumes: convutils.Ref(true),
				AllowedForEphemeralVolumes:  convutils.Ref(true),
				AllowDataSharing:            convutils.Ref(false),
			}

			resp, _ = configService.PatchStorageClassConfigurationById(adminContext(), sc1Id, api.StorageClassConfigurationPatchRequest{
				Permissions: &sc1Permissions,
			}, dummyTenantId)
			Expect(resp.Code).To(Equal(http.StatusOK))

			respRecorder = getStorageClasses(getUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))
			body := unmarshalListBody(respRecorder.Body)

			itemsMap := make(map[string]api.StorageClassV2)

			for _, item := range body.Items {
				itemsMap[item.StorageClassName] = item
			}

			Expect(len(itemsMap)).To(Equal(2))
			sc1Item, foundSc1 := itemsMap["sc1"]
			sc2Item, foundSc2 := itemsMap["sc2"]
			Expect(foundSc1).To(Equal(true))
			Expect(foundSc2).To(Equal(true))

			Expect(sc1Item.Permissions).NotTo(BeNil())
			Expect(*sc1Item.Permissions).To(Equal(expectedSc1Permissions))
			Expect(sc2Item.Permissions).NotTo(BeNil())
			Expect(*sc2Item.Permissions).To(Equal(expectedDefaultPermissions))

			//
			//  add some permission flags for None
			//
			nonePermissions := api.StorageClassPermissions{
				AllowedForPersistentVolumes: convutils.Ref(false),
				AllowedForEphemeralVolumes:  convutils.Ref(false),
			}

			resp, _ = configService.CreateStorageClassConfiguration(adminContext(), api.StorageClassConfigurationCreationRequest{
				ScopeType:   string(authorizationApi.SCOPETYPE_CLUSTER),
				ScopeId:     clusterId,
				Name:        RunaiNoneStorageClass,
				Permissions: &nonePermissions,
			})
			Expect(resp.Code).To(Equal(http.StatusOK))

			// not requesting none -> it should not be included in the list
			respRecorder = getStorageClasses(getUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))
			Expect(len(unmarshalListBody(respRecorder.Body).Items)).To(Equal(2))

			respRecorder = getStorageClasses(getUrl + includeNone)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))
			body = unmarshalListBody(respRecorder.Body)

			itemsMap = make(map[string]api.StorageClassV2)

			for _, item := range body.Items {
				itemsMap[item.StorageClassName] = item
			}

			Expect(len(itemsMap)).To(Equal(3))
			sc1Item, foundSc1 = itemsMap["sc1"]
			sc2Item, foundSc2 = itemsMap["sc2"]
			scNone, foundNone := itemsMap[RunaiNoneStorageClass]
			Expect(foundSc1).To(Equal(true))
			Expect(foundSc2).To(Equal(true))
			Expect(foundNone).To(Equal(true))

			Expect(sc1Item.Permissions).NotTo(BeNil())
			Expect(*sc1Item.Permissions).To(Equal(expectedSc1Permissions))
			Expect(sc2Item.Permissions).NotTo(BeNil())
			Expect(*sc2Item.Permissions).To(Equal(expectedDefaultPermissions))

			nonePermissionsImplicit := api.StorageClassPermissions{
				AllowedForPersistentVolumes: convutils.Ref(false),
				AllowedForEphemeralVolumes:  convutils.Ref(false),
			}

			Expect(scNone.Permissions).NotTo(BeNil())
			Expect(*scNone.Permissions).To(Equal(nonePermissionsImplicit))
		})
	})

	Context("OBSOLETE Storage Class API test", func() {
		scUrl := fmt.Sprintf(v1ScUrl, clusterId)
		getUrl := scUrl
		deletionUrl := scUrl + "/%s"
		syncUrl := scUrl + "/sync"

		It("OBSOLETE create 2, list, delete 1, list", func() {
			respRecorder := getStorageClasses(getUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))
			Expect(len(unmarshalV1Body(respRecorder.Body))).To(Equal(0))

			respRecorder = createV1StorageClass("sc1", scUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))

			respRecorder = createV1StorageClass("sc2", scUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))

			respRecorder = getStorageClasses(getUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))
			scs := unmarshalV1Body(respRecorder.Body)
			Expect(len(scs)).To(Equal(2))
			Expect(scs[0]).To(Equal("sc1"))
			Expect(scs[1]).To(Equal("sc2"))

			respRecorder = deleteStorageClass("sc1", deletionUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))

			respRecorder = getStorageClasses(getUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))
			scs = unmarshalV1Body(respRecorder.Body)
			Expect(len(scs)).To(Equal(1))
			Expect(scs[0]).To(Equal("sc2"))
		})
		It("OBSOLETE sync 2, list, sync only 1, list", func() {
			scsList := []api.V1StorageClass{v1StorageClass, v1StorageClass}
			scsList[0].StorageClassName = "sc1"
			scsList[1].StorageClassName = "sc2"
			scsList[0].IsDefault = true
			scsList[1].IsDefault = false

			respRecorder := getStorageClasses(getUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))
			Expect(len(unmarshalListBody(respRecorder.Body).Items)).To(Equal(0))

			respRecorder = syncStorageClass(scsList, syncUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))

			respRecorder = getStorageClasses(getUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))
			scs := unmarshalV1Body(respRecorder.Body)
			Expect(len(scs)).To(Equal(2))
			Expect(scs[0]).To(Equal("sc1"))
			Expect(scs[1]).To(Equal("sc2"))

			// Now sync only sc2 and modify it
			scsList[1].IsDefault = true
			scsList = scsList[1:2]
			respRecorder = syncStorageClass(scsList, syncUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))

			respRecorder = getStorageClasses(getUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))
			scs = unmarshalV1Body(respRecorder.Body)
			Expect(len(scs)).To(Equal(1))
			Expect(scs[0]).To(Equal("sc2"))
		})
	})

	Context("Storage Class API test", func() {
		It("should ensure all default permissions are properly set", func() {
			// Define the URLs
			scUrl := newScUrl + clusterId
			getUrl := v2ScUrl + clusterId

			// Create a new storage class
			respRecorder := createStorageClass("sc1", scUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))

			// Get the storage class and validate the permissions fields
			respRecorder = getStorageClasses(getUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))

			// Parse the response to get the storage class
			scs := unmarshalListBody(respRecorder.Body)
			Expect(len(scs.Items)).To(Equal(1))

			sc := scs.Items[0]
			Expect(sc.StorageClassName).To(Equal("sc1"))
			Expect(sc.Permissions).NotTo(BeNil())

			// Validate each default permission
			Expect(sc.Permissions.AllowedForAssets).NotTo(BeNil())
			Expect(*sc.Permissions.AllowedForAssets).To(BeTrue())

			Expect(sc.Permissions.AllowedForWorkloads).NotTo(BeNil())
			Expect(*sc.Permissions.AllowedForWorkloads).To(BeTrue())

			Expect(sc.Permissions.AllowedForEphemeralVolumes).NotTo(BeNil())
			Expect(*sc.Permissions.AllowedForEphemeralVolumes).To(BeTrue())

			Expect(sc.Permissions.AllowedForPersistentVolumes).NotTo(BeNil())
			Expect(*sc.Permissions.AllowedForPersistentVolumes).To(BeTrue())

			Expect(sc.Permissions.AllowDataSharing).NotTo(BeNil())
			Expect(*sc.Permissions.AllowDataSharing).To(BeFalse())
		})
	})
})

func syncStorageClass(payloadObject any, scUrl string) *httptest.ResponseRecorder {
	request, _ := http.NewRequest("POST", scUrl, makePayload(payloadObject))
	return sendStorageClassApiRequest(request)
}

func createStorageClass(name, scUrl string) *httptest.ResponseRecorder {
	clone := storageClass
	clone.StorageClassName = name
	request, _ := http.NewRequest("POST", scUrl, makePayload(clone))
	return sendStorageClassApiRequest(request)
}

func createV1StorageClass(name, scUrl string) *httptest.ResponseRecorder {
	clone := v1StorageClass
	clone.StorageClassName = name
	clone.IsDefault = true
	request, _ := http.NewRequest("POST", scUrl, makePayload(clone))
	return sendStorageClassApiRequest(request)
}

func getStorageClasses(scUrl string) *httptest.ResponseRecorder {
	request, _ := http.NewRequest("GET", scUrl, nil)
	return sendStorageClassApiRequest(request)
}

func deleteStorageClass(name, scUrl string) *httptest.ResponseRecorder {
	request, _ := http.NewRequest("DELETE", fmt.Sprintf(scUrl, name), nil)
	return sendStorageClassApiRequest(request)
}

func sendStorageClassApiRequest(request *http.Request) *httptest.ResponseRecorder {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", dummyTenantId))
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "admin")

	request.Header = *runaictx_http.GetHeaders(ctx)
	respRecorder := httptest.NewRecorder()

	audit.Middleware(router).ServeHTTP(respRecorder, request)
	return respRecorder
}

func unmarshalListBody(buf *bytes.Buffer) api.StorageClassesV2 {
	scs := api.StorageClassesV2{
		Items: make([]api.StorageClassV2, 0),
	}
	err := json.Unmarshal(buf.Bytes(), &scs)
	Expect(err).To(BeNil())
	return scs
}

func unmarshalGetBody(buf *bytes.Buffer) api.StorageClass {
	scs := api.StorageClass{}
	err := json.Unmarshal(buf.Bytes(), &scs)
	Expect(err).To(BeNil())
	return scs
}

func unmarshalV1Body(buf *bytes.Buffer) []string {
	scs := api.V1GetStorageClasses200Response{
		StorageClasses: make([]string, 0),
	}
	err := json.Unmarshal(buf.Bytes(), &scs)
	Expect(err).To(BeNil())
	return scs.StorageClasses
}

func adminContext() context.Context {
	permittedScopes := authorizationApi.PermittedScopes{
		System: convutils.Ref(true),
	}

	ctx := context.WithValue(context.Background(), structs.PermittedScopesKey, &permittedScopes)
	ctx = context.WithValue(ctx, runaictx.CfTenantId, fmt.Sprintf("%d", dummyTenantId))

	return ctx
}
