// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	authorizationclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/gpuinventory"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/ngcmetadata"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/services/rightsizing"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type sizingInventory struct{}

func (sizingInventory) Get(context.Context, string) ([]gpuinventory.Hardware, *api.Error) {
	capacity := int64(80_000_000_000)
	return []gpuinventory.Hardware{{Type: "A100", MemoryBytes: &capacity, MaxDevicesPerNode: 8}, {Type: "T4"}}, nil
}

type sizingMetadata struct{}

func (sizingMetadata) Get(context.Context, string, string) (ngcmetadata.Metadata, *api.Error) {
	return ngcmetadata.Metadata{}, nil
}

var _ = It("catalog recommendation route preserves null gpu and missing recommendations", func() {
	previousAuthConfig := config.Get[*authorizationclient.Config]()
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer GinkgoRecover()
		var permission struct {
			ResourceType  string `json:"resourceType"`
			RequestAction string `json:"requestAction"`
		}
		Expect(json.NewDecoder(r.Body).Decode(&permission)).To(Succeed())
		Expect(permission.ResourceType).To(Equal("inferences"))
		Expect(permission.RequestAction).To(Equal("read"))
		w.WriteHeader(http.StatusNoContent)
	}))
	DeferCleanup(authServer.Close)
	config.Set(&authorizationclient.Config{Url: authServer.URL})
	DeferCleanup(func() { config.Set(previousAuthConfig) })
	service := &RightSizingAPIService{service: rightsizing.New(sizingInventory{}, sizingMetadata{}, 90)}
	controller := api.NewRightSizingAPIController(service)
	request := httptest.NewRequest("POST", "/api/v1/catalog/right-sizing/recommendation", strings.NewReader(`{"clusterId":"4b94efde-b05a-4118-a03a-1dedff81fa8e","image":"nvcr.io/nim/nvidia/example","tag":"1.2.3","gpuType":null}`))
	response := httptest.NewRecorder()
	api.NewRouter(controller).ServeHTTP(response, request)
	Expect(response.Code).To(BeNumerically("==", http.StatusOK))
	var body struct {
		Results []map[string]interface{} `json:"results"`
	}
	{
		err := json.Unmarshal(response.Body.Bytes(), &body)
		Expect(err).To(BeNil())
	}
	Expect(body.Results).To(HaveLen(2))
	for _, result := range body.Results {
		Expect(result["status"]).To(Equal("no_recommendation"))
		Expect(result["recommendation"]).To(BeNil())
		Expect(result["minimumGpuMemoryPerDeviceGb"]).To(BeNil())
	}
})

var _ = It("does not expose recommendations under assets inference namespace", func() {
	service := &RightSizingAPIService{service: rightsizing.New(sizingInventory{}, sizingMetadata{}, 90)}
	controller := api.NewRightSizingAPIController(service)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/inference/right-sizing/recommendation", strings.NewReader(`{"clusterId":"4b94efde-b05a-4118-a03a-1dedff81fa8e","image":"nvcr.io/nim/nvidia/example","tag":"1.2.3"}`))
	response := httptest.NewRecorder()
	api.NewRouter(controller).ServeHTTP(response, request)
	Expect(response.Code).To(Equal(http.StatusNotFound))
})
