/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"fmt"

	"github.com/golang/mock/gomock"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	cpmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/cluster"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli"
	"k8s.io/utils/ptr"
)

var (
	ClusterOne = openapi.ClustersMinimalData{
		TenantId: 1,
		Uuid:     cli.ClusterOneUuid,
		Name:     cli.ClusterOneName,
		Domain:   *openapi.NewNullableString(ptr.To(cli.ClusterOneDomain)),
	}
	ClusterTwo = openapi.ClustersMinimalData{
		TenantId: 1,
		Uuid:     cli.ClusterTwoUuid,
		Name:     cli.ClusterTwoName,
		Domain:   *openapi.NewNullableString(ptr.To(cli.ClusterOneDomain)),
	}
)

func MockClusters(mockCtrl *gomock.Controller) {
	clusterList := []openapi.ClustersMinimalData{ClusterOne, ClusterTwo}
	mockSingleOrMultiCluster(mockCtrl, clusterList)
}

func MockSingleCluster(mockCtrl *gomock.Controller) {
	clusterList := []openapi.ClustersMinimalData{ClusterOne}
	mockSingleOrMultiCluster(mockCtrl, clusterList)
}

func mockSingleOrMultiCluster(mockCtrl *gomock.Controller, clusterList []openapi.ClustersMinimalData) {
	mockCluster := cpmocks.NewMockClusterClient(mockCtrl)
	mockCluster.EXPECT().GetClusterList(gomock.Any()).Return(clusterList, nil).AnyTimes()
	mockCluster.EXPECT().GetClusterByName(gomock.Any(), cli.ClusterOneName).Return(&ClusterOne, nil).AnyTimes()
	mockCluster.EXPECT().GetClusterByID(gomock.Any(), cli.ClusterOneUuid).Return(&ClusterOne, nil).AnyTimes()
	cluster.SetClient(mockCluster)
}

func ClusterSet(envMap map[string]string) error {
	_, err := cli.RunCommand(fmt.Sprintf("cluster set %s", cli.ClusterOneName), envMap)
	return err
}
