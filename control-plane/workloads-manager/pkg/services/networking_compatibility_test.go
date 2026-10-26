// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	clusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
)

func TestNetworkingOwnershipCompatibility(t *testing.T) {
	values := testutils.NewRandomValues()
	networking := []api.NetworkingEndpoint{{Services: &[]api.NetworkingService{{Ref: &api.ServiceRef{Name: "workload", Port: 8080}}}}}
	empty := []api.NetworkingEndpoint{}
	for _, tc := range []struct {
		name, version, path string
		translated          bool
		config              *[]api.NetworkingEndpoint
		wantErr             string
	}{
		{name: "direct old endpoint-capable cluster", version: "2.26.0", config: &networking, wantErr: "ownership validation"},
		{name: "direct legacy-compatible config", version: "2.25.0", config: &networking, wantErr: "ownership validation"},
		{name: "direct release boundary", version: "3.0.0", config: &networking},
		{name: "direct later version", version: "3.1.0", config: &networking},
		{name: "development build convention", version: "0.0.0-test", config: &networking},
		{name: "native compatibility", version: "2.25.0", path: InternalWorkloadsAPIPath, translated: true, config: &networking},
		{name: "NIM compatibility", version: "2.25.0", path: NimServicesAPIPath, config: &networking},
		{name: "internal path alone is insufficient", version: "2.26.0", path: InternalWorkloadsAPIPath, config: &networking, wantErr: "ownership validation"},
		{name: "public origin cannot claim native", version: "2.26.0", path: urlWorkloads, translated: true, config: &networking, wantErr: "ownership validation"},
		{name: "unknown version fails closed", version: "invalid", config: &networking, wantErr: "cluster capability check"},
		{name: "no networking", version: "2.25.0"},
		{name: "clearing networking", version: "2.25.0", config: &empty},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clusterClient.SetMock(clusterClient.NewSimpleMockClient().WithCluster(values.ClusterId, clusterInfoOf(values, tc.version)))
			checker.InitClustersCache(&checker.ClustersCacheConfig{ExpirationSec: 600})
			profile := networkingSubmissionProfileForRecord(&models.WorkloadsQueueRecord{
				InputRequestPath: tc.path, TranslatedFromV1: tc.translated,
			})
			err := validateWorkloadV2NetworkingForCluster(context.Background(), values.ClusterId, "workload", tc.config, profile)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %+v", err)
				}
			} else if err == nil || !strings.Contains(err.Message, tc.wantErr) {
				t.Fatalf("got %+v, want error containing %q", err, tc.wantErr)
			}
		})
	}
}
