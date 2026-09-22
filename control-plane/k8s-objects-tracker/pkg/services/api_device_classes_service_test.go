/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	clustersApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/clients/clusters"
	orgUnit "github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/clients/org-unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/database/dbtesting"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	deviceClassesUrl     = "/api/v1/device-classes?clusterId="
	deviceClassesSyncUrl = "/api/v1/device-classes/sync?clusterId="
)

var _ = Describe("device classes test", func() {
	BeforeEach(func() {
		backendMock := NewRandomsBasedBackendMockClient()
		orgUnit.SetMock(backendMock)
		authorization.SetMock(backendMock)
		router = CreateRouter()
		dbtesting.CleanTables()

		clusters.SetMock(clusters.NewSimpleMock().WithCluster(clusterId, &clustersApi.DisplayedCluster{}))
	})

	Context("Device Class API test", func() {
		getUrl := deviceClassesUrl + clusterId
		syncUrl := deviceClassesSyncUrl + clusterId

		It("sync 2, list, sync only 1, list (stale is pruned)", func() {
			dcsList := api.DeviceClasses{Items: []api.DeviceClass{
				{DeviceClassName: "gpu.nvidia.com"},
				{DeviceClassName: "mrdma.google.com"},
			}}

			respRecorder := getDeviceClasses(getUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))
			Expect(len(unmarshalDeviceClasses(respRecorder.Body).Items)).To(Equal(0))

			respRecorder = syncDeviceClasses(dcsList, syncUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))

			respRecorder = getDeviceClasses(getUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))
			dcs := unmarshalDeviceClasses(respRecorder.Body)
			Expect(len(dcs.Items)).To(Equal(2))
			Expect(deviceClassNames(dcs)).To(ContainElements("gpu.nvidia.com", "mrdma.google.com"))
			// the cluster is taken from the query param and echoed back on read
			Expect(dcs.Items[0].ClusterId).To(Equal(clusterId))

			// Now sync only one -> the other must be pruned
			dcsList.Items = dcsList.Items[:1]
			respRecorder = syncDeviceClasses(dcsList, syncUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))

			respRecorder = getDeviceClasses(getUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))
			dcs = unmarshalDeviceClasses(respRecorder.Body)
			Expect(len(dcs.Items)).To(Equal(1))
			Expect(dcs.Items[0].DeviceClassName).To(Equal("gpu.nvidia.com"))
		})

		It("filters by name", func() {
			dcsList := api.DeviceClasses{Items: []api.DeviceClass{
				{DeviceClassName: "gpu.nvidia.com"},
				{DeviceClassName: "mrdma.google.com"},
			}}

			respRecorder := syncDeviceClasses(dcsList, syncUrl)
			Expect(respRecorder.Code).To(Equal(http.StatusOK))

			respRecorder = getDeviceClasses(getUrl + "&name=gpu.nvidia.com")
			Expect(respRecorder.Code).To(Equal(http.StatusOK))
			dcs := unmarshalDeviceClasses(respRecorder.Body)
			Expect(len(dcs.Items)).To(Equal(1))
			Expect(dcs.Items[0].DeviceClassName).To(Equal("gpu.nvidia.com"))
		})
	})
})

func syncDeviceClasses(payloadObject any, url string) *httptest.ResponseRecorder {
	request, _ := http.NewRequest("POST", url, makePayload(payloadObject))
	return sendStorageClassApiRequest(request)
}

func getDeviceClasses(url string) *httptest.ResponseRecorder {
	request, _ := http.NewRequest("GET", url, nil)
	return sendStorageClassApiRequest(request)
}

func unmarshalDeviceClasses(buf *bytes.Buffer) api.DeviceClasses {
	dcs := api.DeviceClasses{Items: make([]api.DeviceClass, 0)}
	err := json.Unmarshal(buf.Bytes(), &dcs)
	Expect(err).To(BeNil())
	return dcs
}

func deviceClassNames(dcs api.DeviceClasses) []string {
	names := make([]string, len(dcs.Items))
	for i, dc := range dcs.Items {
		names[i] = dc.DeviceClassName
	}
	return names
}
