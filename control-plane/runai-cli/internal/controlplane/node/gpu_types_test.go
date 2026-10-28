// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package node

import (
	"testing"

	mocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/golang/mock/gomock"
)

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
