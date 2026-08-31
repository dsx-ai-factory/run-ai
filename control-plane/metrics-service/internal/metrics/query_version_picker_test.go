/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"context"
	"testing"
	"time"

	apiClusters "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/clusters"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/services"
	"github.com/stretchr/testify/assert"
)

var (
	testingClusters        = []string{"cluster1", "cluster2"}
	clusterVersionHistory1 = clusters.NewClusterVersionHistory([]apiClusters.ClusterHistory{
		{
			ClusterId: testingClusters[0], Version: "2.13", UpdateTimestamp: time.Now().Add(-2 * time.Hour).Unix(),
		}, {
			ClusterId: testingClusters[0], Version: "2.9", UpdateTimestamp: time.Now().Add(-24 * 30 * time.Hour).Unix(),
		},
	})
)

var clusterVersionHistory2 = clusters.NewClusterVersionHistory([]apiClusters.ClusterHistory{
	{
		ClusterId: testingClusters[1], Version: "2.15", UpdateTimestamp: time.Now().Add(-2 * time.Hour).Unix(),
	}, {
		ClusterId: testingClusters[1], Version: "2.9", UpdateTimestamp: time.Now().Add(-24 * 30 * time.Hour).Unix(),
	},
})

var clusterVersionHistories = map[string]*clusters.ClusterVersionHistory{
	testingClusters[0]: clusterVersionHistory1,
	testingClusters[1]: clusterVersionHistory2,
}

var testContext = context.Background()

func TestPickVersion_EmptyHistory(t *testing.T) {
	qvp := NewQueryVersionPicker()
	clustersVersions := map[string]*clusters.ClusterVersionHistory{}
	r := &services.MetricInformationRange{
		Start: time.Now(),
		End:   time.Now(),
	}
	v := qvp.Pick(testContext, clustersVersions, r)
	assert.Equal(t, v, CROSS_CLUSTER_VERSION)
}

func TestPickVersion_EmptyVersion(t *testing.T) {
	qvp := NewQueryVersionPicker()
	empty := clusters.NewClusterVersionHistory([]apiClusters.ClusterHistory{
		{
			ClusterId: "empty", Version: "", UpdateTimestamp: time.Now().Add(-2 * time.Hour).Unix(),
		},
	})
	clustersVersions := map[string]*clusters.ClusterVersionHistory{"empty": empty}
	r := &services.MetricInformationRange{
		Start: time.Now(),
		End:   time.Now(),
	}
	v := qvp.Pick(testContext, clustersVersions, r)
	assert.Equal(t, v, CROSS_CLUSTER_VERSION)
}

func TestPickVersion_OneClusterOneVersion(t *testing.T) {
	qvp := NewQueryVersionPicker()
	r := &services.MetricInformationRange{
		Start: time.Now().Add(-24 * time.Hour),
		End:   time.Now().Add(-3 * time.Hour),
	}
	clusterVersions := map[string]*clusters.ClusterVersionHistory{testingClusters[0]: clusterVersionHistory1}
	v := qvp.Pick(testContext, clusterVersions, r)
	assert.Equal(t, v, "2.9.0")
}

func TestPickVersion_OneClusterMultipleVersions(t *testing.T) {
	qvp := NewQueryVersionPicker()
	r := &services.MetricInformationRange{
		Start: time.Now().Add(-24 * time.Hour),
		End:   time.Now(),
	}
	clusterVersions := map[string]*clusters.ClusterVersionHistory{testingClusters[0]: clusterVersionHistory1}
	v := qvp.Pick(testContext, clusterVersions, r)
	assert.Equal(t, v, CROSS_CLUSTER_VERSION)
}

func TestPickVersion_MultipleClustersOneVersion(t *testing.T) {
	qvp := NewQueryVersionPicker()
	r := &services.MetricInformationRange{
		Start: time.Now().Add(-24 * time.Hour),
		End:   time.Now().Add(-3 * time.Hour),
	}
	v := qvp.Pick(testContext, clusterVersionHistories, r)
	assert.Equal(t, v, "2.9.0")
}

func TestPickVersion_MultipleClustersMultipleVersions(t *testing.T) {
	qvp := NewQueryVersionPicker()
	r := &services.MetricInformationRange{
		Start: time.Now().Add(-1 * time.Hour),
		End:   time.Now().Add(-30 * time.Minute),
	}
	v := qvp.Pick(testContext, clusterVersionHistories, r)
	assert.Equal(t, v, CROSS_CLUSTER_VERSION)
}

func TestPickVersion_NoTimeRangeDifferentVersions(t *testing.T) {
	qvp := NewQueryVersionPicker()
	v := qvp.Pick(testContext, clusterVersionHistories, nil)
	assert.Equal(t, v, CROSS_CLUSTER_VERSION)
}

func TestPickVersion_NoTimeRangeOneVersion(t *testing.T) {
	qvp := NewQueryVersionPicker()
	clusterVersions := map[string]*clusters.ClusterVersionHistory{testingClusters[0]: clusterVersionHistory2}
	v := qvp.Pick(testContext, clusterVersions, nil)
	assert.Equal(t, v, "2.15.0")
}
