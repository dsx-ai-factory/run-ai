/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/clusters"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/services"
	"github.com/hashicorp/go-version"
	"k8s.io/utils/ptr"

	serviceErrors "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/errors"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/metrics/mocks"
	"github.com/stretchr/testify/assert"
)

var (
	mockCtx           = context.Background()
	testClusters      = []string{"some_cluster"}
	ver, _            = version.NewVersion("2.18.0")
	metricInformation = services.MetricInformation{
		Name: "some metric",
		Filters: map[string][]string{
			"clusterId": testClusters,
		},
		Range: &services.MetricInformationRange{
			Start: time.Now().Add(-12 * time.Hour),
			End:   time.Now(),
		},
		Quantile: ptr.To[float32](0.5),
	}
)

var clustersVersionHistoryMap = map[string]*clusters.ClusterVersionHistory{"cluster_1": {
	Versions:       []*clusters.VersionWithTimeRange{{Version: ver}},
	CurrentVersion: ver,
}}

func Test_VersionHistoryGetterReturnError(t *testing.T) {
	historiesGetter := mocks.NewClusterVersionsHistoryGetter(t)
	historiesGetter.On("GetClustersVersionsHistories", mockCtx, testClusters).Return(nil, errors.New("version histories getter failed"))
	queryVersionPicker := mocks.NewQueryVersionPicker(t)
	metricQueryPicker := mocks.NewMetricQueryPicker(t)
	injectorChain := mocks.NewQueryInjector(t)

	res, err := NewPromMetricQueryBuilder(historiesGetter, queryVersionPicker, metricQueryPicker, injectorChain).Build(mockCtx, &metricInformation)
	assert.Equal(t, err.Error(), "version histories getter failed")
	assert.Equal(t, res, "")
}

func Test_MetricVersionExceedClusterVersion(t *testing.T) {
	historiesGetter := mocks.NewClusterVersionsHistoryGetter(t)
	queryVersionPicker := mocks.NewQueryVersionPicker(t)
	historiesGetter.On("GetClustersVersionsHistories", mockCtx, testClusters).Return(clustersVersionHistoryMap, nil)

	metricQueryPicker := mocks.NewMetricQueryPicker(t)
	v, _ := version.NewVersion("2.19.0")
	metricQueryPicker.On("GetQueryMinVersion", metricInformation.Name).Return(v, nil)
	injectorChain := mocks.NewQueryInjector(t)

	res, err := NewPromMetricQueryBuilder(historiesGetter, queryVersionPicker, metricQueryPicker, injectorChain).Build(mockCtx, &metricInformation)
	assert.Equal(t, err.Error(), "cluster version unsupported for metric some metric")
	assert.True(t, serviceErrors.IsClusterVersionNotSupportedErr(err))
	assert.Equal(t, res, "")
}

func Test_ClusterVersionIsMasterExceedMetricVersion(t *testing.T) {
	historiesGetter := mocks.NewClusterVersionsHistoryGetter(t)
	queryVersionPicker := mocks.NewQueryVersionPicker(t)

	cvh := map[string]*clusters.ClusterVersionHistory{"cluster_1": {
		Versions:       []*clusters.VersionWithTimeRange{{Version: ver}, {Version: clusters.DevVersion}},
		CurrentVersion: clusters.DevVersion,
	}}
	historiesGetter.On("GetClustersVersionsHistories", mockCtx, testClusters).Return(cvh, nil)
	queryVersionPicker.On("Pick", mockCtx, cvh, metricInformation.Range).Return(CROSS_CLUSTER_VERSION)

	metricQueryPicker := mocks.NewMetricQueryPicker(t)
	metricQueryPicker.On("GetQueryMinVersion", metricInformation.Name).Return(ver, nil)
	metricQueryPicker.On("GetQueryByVersion", metricInformation.Name, CROSS_CLUSTER_VERSION).Return("some query", nil)

	injectorChain := mocks.NewQueryInjector(t)
	injectorChain.On("Inject", "some query", &metricInformation).Return("final query", nil)

	res, err := NewPromMetricQueryBuilder(historiesGetter, queryVersionPicker, metricQueryPicker, injectorChain).Build(mockCtx, &metricInformation)
	assert.Equal(t, err, nil)
	assert.Equal(t, res, "final query")
}

func Test_ClusterWithoutVersion(t *testing.T) {
	historiesGetter := mocks.NewClusterVersionsHistoryGetter(t)
	queryVersionPicker := mocks.NewQueryVersionPicker(t)
	cvh := map[string]*clusters.ClusterVersionHistory{"cluster_1": {
		Versions:       []*clusters.VersionWithTimeRange{},
		CurrentVersion: nil,
	}, "cluster_2": {
		Versions:       []*clusters.VersionWithTimeRange{},
		CurrentVersion: nil,
	}}
	historiesGetter.On("GetClustersVersionsHistories", mockCtx, testClusters).Return(cvh, nil)
	queryVersionPicker.On("Pick", mockCtx, cvh, metricInformation.Range).Return(CROSS_CLUSTER_VERSION)

	metricQueryPicker := mocks.NewMetricQueryPicker(t)
	metricQueryPicker.On("GetQueryMinVersion", metricInformation.Name).Return(ver, nil)
	metricQueryPicker.On("GetQueryByVersion", metricInformation.Name, CROSS_CLUSTER_VERSION).Return("some query", nil)

	injectorChain := mocks.NewQueryInjector(t)
	injectorChain.On("Inject", "some query", &metricInformation).Return("final query", nil)

	res, err := NewPromMetricQueryBuilder(historiesGetter, queryVersionPicker, metricQueryPicker, injectorChain).Build(mockCtx, &metricInformation)
	assert.Equal(t, err, nil)
	assert.Equal(t, res, "final query")
}

func Test_ClusterVersionsEmpty(t *testing.T) {
	historiesGetter := mocks.NewClusterVersionsHistoryGetter(t)
	queryVersionPicker := mocks.NewQueryVersionPicker(t)
	cvh := map[string]*clusters.ClusterVersionHistory{}
	historiesGetter.On("GetClustersVersionsHistories", mockCtx, testClusters).Return(cvh, nil)
	queryVersionPicker.On("Pick", mockCtx, cvh, metricInformation.Range).Return(CROSS_CLUSTER_VERSION)

	metricQueryPicker := mocks.NewMetricQueryPicker(t)
	metricQueryPicker.On("GetQueryMinVersion", metricInformation.Name).Return(ver, nil)
	metricQueryPicker.On("GetQueryByVersion", metricInformation.Name, CROSS_CLUSTER_VERSION).Return("some query", nil)

	injectorChain := mocks.NewQueryInjector(t)
	injectorChain.On("Inject", "some query", &metricInformation).Return("final query", nil)

	res, err := NewPromMetricQueryBuilder(historiesGetter, queryVersionPicker, metricQueryPicker, injectorChain).Build(mockCtx, &metricInformation)
	assert.Equal(t, err, nil)
	assert.Equal(t, res, "final query")
}

func Test_metricsQueryPickerReturnError(t *testing.T) {
	historiesGetter := mocks.NewClusterVersionsHistoryGetter(t)
	queryVersionPicker := mocks.NewQueryVersionPicker(t)
	metricQueryPicker := mocks.NewMetricQueryPicker(t)

	historiesGetter.On("GetClustersVersionsHistories", mockCtx, testClusters).Return(clustersVersionHistoryMap, nil)
	queryVersionPicker.On("Pick", mockCtx, clustersVersionHistoryMap, metricInformation.Range).Return(CROSS_CLUSTER_VERSION)

	metricQueryPicker.On("GetQueryMinVersion", metricInformation.Name).Return(ver, nil)
	metricQueryPicker.On("GetQueryByVersion", metricInformation.Name, CROSS_CLUSTER_VERSION).Return("", errors.New("metricQueryPicker failed"))

	injectorChain := mocks.NewQueryInjector(t)

	res, err := NewPromMetricQueryBuilder(historiesGetter, queryVersionPicker, metricQueryPicker, injectorChain).Build(mockCtx, &metricInformation)
	assert.Equal(t, err.Error(), "metricQueryPicker failed")
	assert.Equal(t, res, "")
}

func Test_queryInjectorReturnError(t *testing.T) {
	historiesGetter := mocks.NewClusterVersionsHistoryGetter(t)
	queryVersionPicker := mocks.NewQueryVersionPicker(t)
	historiesGetter.On("GetClustersVersionsHistories", mockCtx, testClusters).Return(clustersVersionHistoryMap, nil)
	queryVersionPicker.On("Pick", mockCtx, clustersVersionHistoryMap, metricInformation.Range).Return(CROSS_CLUSTER_VERSION)

	metricQueryPicker := mocks.NewMetricQueryPicker(t)
	metricQueryPicker.On("GetQueryMinVersion", metricInformation.Name).Return(ver, nil)
	metricQueryPicker.On("GetQueryByVersion", metricInformation.Name, CROSS_CLUSTER_VERSION).Return("some query", nil)

	injectorChain := mocks.NewQueryInjector(t)
	injectorChain.On("Inject", "some query", &metricInformation).Return("", errors.New("injector failed"))

	res, err := NewPromMetricQueryBuilder(historiesGetter, queryVersionPicker, metricQueryPicker, injectorChain).Build(mockCtx, &metricInformation)
	assert.Equal(t, err.Error(), "injector failed")
	assert.Equal(t, res, "")
}

func Test_succeed(t *testing.T) {
	historiesGetter := mocks.NewClusterVersionsHistoryGetter(t)
	queryVersionPicker := mocks.NewQueryVersionPicker(t)
	historiesGetter.On("GetClustersVersionsHistories", mockCtx, testClusters).Return(clustersVersionHistoryMap, nil)
	queryVersionPicker.On("Pick", mockCtx, clustersVersionHistoryMap, metricInformation.Range).Return(CROSS_CLUSTER_VERSION)

	metricQueryPicker := mocks.NewMetricQueryPicker(t)
	metricQueryPicker.On("GetQueryMinVersion", metricInformation.Name).Return(ver, nil)
	metricQueryPicker.On("GetQueryByVersion", metricInformation.Name, CROSS_CLUSTER_VERSION).Return("some query", nil)

	injectorChain := mocks.NewQueryInjector(t)
	injectorChain.On("Inject", "some query", &metricInformation).Return("final query", nil)

	res, err := NewPromMetricQueryBuilder(historiesGetter, queryVersionPicker, metricQueryPicker, injectorChain).Build(mockCtx, &metricInformation)
	assert.Equal(t, err, nil)
	assert.Equal(t, res, "final query")
}
