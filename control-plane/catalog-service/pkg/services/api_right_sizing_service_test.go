// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/gpuinventory"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/ngcmetadata"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/services/rightsizing"
)

type sizingInventory struct{}

func (sizingInventory) Get(context.Context, string) ([]gpuinventory.Hardware, error) {
	capacity := int64(80_000_000_000)
	return []gpuinventory.Hardware{{Type: "A100", MemoryBytes: &capacity}, {Type: "T4"}}, nil
}

type sizingMetadata struct{}

func (sizingMetadata) Get(context.Context, string, string) (ngcmetadata.Metadata, error) {
	return ngcmetadata.Metadata{}, nil
}

func TestRightSizingControllerPreservesNullGPUAndMissingRecommendations(t *testing.T) {
	service := &RightSizingAPIService{service: rightsizing.New(sizingInventory{}, sizingMetadata{})}
	controller := api.NewRightSizingAPIController(service)
	request := httptest.NewRequest("POST", "/api/v1/inference/right-sizing/recommendation", strings.NewReader(`{"clusterId":"4b94efde-b05a-4118-a03a-1dedff81fa8e","image":"nvcr.io/nim/nvidia/example","tag":"1.2.3","gpuType":null}`))
	response := httptest.NewRecorder()
	controller.RecommendInferenceResources(response, request)
	if response.Code != 200 {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var body struct {
		Results []map[string]interface{} `json:"results"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Results) != 2 {
		t.Fatalf("results=%s", response.Body.String())
	}
	for _, result := range body.Results {
		if result["status"] != "no_recommendation" || result["recommendation"] != nil || result["minimumGpuMemoryPerDeviceGb"] != nil {
			t.Fatalf("invented values: %+v", result)
		}
	}
}
