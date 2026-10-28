// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package node

import (
	"context"
	"errors"
	"reflect"
	"testing"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	mocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/golang/mock/gomock"
)

func TestGPUTypesInventory(t *testing.T) {
	client := mocks.NewMockNodeClient(gomock.NewController(t))
	gpu := func(label string) openapi.Node {
		return openapi.Node{GpuInfo: *openapi.NewNullableGpuInfo(openapi.NewGpuInfo(label, 1))}
	}
	first := make([]openapi.Node, 500)
	first[0] = gpu("NVIDIA-A100-SXM4-40GB")
	first[1] = gpu("NVIDIA-A100-SXM4-40GB")
	client.EXPECT().GetNodeList(gomock.Any(), "selected-cluster", int32(0), int32(500)).Return(&openapi.Nodes{Nodes: first}, nil)
	client.EXPECT().GetNodeList(gomock.Any(), "selected-cluster", int32(500), int32(500)).Return(&openapi.Nodes{Nodes: []openapi.Node{gpu("Tesla-V100")}}, nil)
	labels, err := NewBackendNodeService(client).GPUTypes(context.Background(), "selected-cluster")
	if err != nil || !reflect.DeepEqual(labels, []string{"NVIDIA-A100-SXM4-40GB", "Tesla-V100"}) {
		t.Fatalf("labels=%v err=%v", labels, err)
	}
}
func TestGPUTypesFailure(t *testing.T) {
	client := mocks.NewMockNodeClient(gomock.NewController(t))
	failure := errors.New("inventory unavailable")
	client.EXPECT().GetNodeList(gomock.Any(), "cluster", int32(0), int32(500)).Return(nil, failure)
	_, err := NewBackendNodeService(client).GPUTypes(context.Background(), "cluster")
	if !errors.Is(err, failure) {
		t.Fatalf("got %v", err)
	}
}

func TestGetNodeListerDoesNotCacheRealClient(t *testing.T) {
	previous := lister
	lister = nil
	t.Cleanup(func() { lister = previous })
	first, err := GetNodeLister("https://first.example", "first-token")
	if err != nil {
		t.Fatal(err)
	}
	second, err := GetNodeLister("https://second.example", "second-token")
	if err != nil {
		t.Fatal(err)
	}
	if first.nodeClient == second.nodeClient || lister != nil {
		t.Fatal("real inventory client retained across credentials/domains")
	}
	mock := mocks.NewMockNodeClient(gomock.NewController(t))
	lister = mock
	service, err := GetNodeLister("https://ignored.example", "ignored")
	if err != nil || service.nodeClient != mock {
		t.Fatalf("mock injection lost: %v", err)
	}
}
