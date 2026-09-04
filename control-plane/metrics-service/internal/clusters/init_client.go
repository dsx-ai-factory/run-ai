/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clusters

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jarcoal/httpmock"
	api "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

func InitClient(useMocks bool) *Client {
	client := NewClustersClient()
	if useMocks {
		mockHttpClients(client)
		mockHttpEndpoints()
	}
	return client
}

func mockHttpClients(client *Client) {
	logging.Infof("Mocking clusters client http clients")
	httpmock.ActivateNonDefault(client.client.GetConfig().HTTPClient)
}

func mockHttpEndpoints() {
	conf := config.Get[*ClustersConfiguration]()

	ts := time.Now()
	cluster1Id := "cluster_1"
	cluster2Id := "cluster_2"
	cluster3Id := "cluster_3"
	apiHistory := api.ClusterHistory{
		ClusterId:       cluster1Id,
		Version:         "2.11",
		UpdateTimestamp: ts.Add(-25 * time.Hour).Unix(),
	}
	apiHistory2 := api.ClusterHistory{
		ClusterId:       cluster1Id,
		Version:         "2.12",
		UpdateTimestamp: ts.Add(-1 * time.Hour).Unix(),
	}
	histories := []api.ClusterHistory{apiHistory2, apiHistory}
	mockClusterHistories(cluster1Id, conf, http.StatusOK, histories)
	mockClusterHistories(cluster2Id, conf, http.StatusOK, []api.ClusterHistory{})
	mockClusterHistories(cluster3Id, conf, http.StatusOK, []api.ClusterHistory{})

	mockGetClusterName(cluster1Id, conf, http.StatusOK, api.GetClusterName200Response{Name: "cluster_1_name"})
}

func mockClusterHistories(clusterId string, config *ClustersConfiguration, statusCode int, body any) {
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/internal/cluster-histories?clusterId=%s", config.ServiceUrl, clusterId), httpmock.NewJsonResponderOrPanic(statusCode, body))
}

func mockGetClusterName(clusterId string, config *ClustersConfiguration, statusCode int, body any) {
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/internal/clusters/%s/name", config.ServiceUrl, clusterId), httpmock.NewJsonResponderOrPanic(statusCode, body))
}
