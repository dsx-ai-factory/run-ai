/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clusters

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hashicorp/go-version"
	api "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	cache_testing "github.com/run-ai/runai/runai-common-packages/cache/v2/testing"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	v1, _       = version.NewSemver("2.11")
	testContext = context.Background()
)

func TestClient_GetHistoriesFromCache(t *testing.T) {
	cfg := ClustersConfiguration{
		ServiceUrl: "url",
		Cache: &CacheConfig{
			VersionHistoriesTtlMinutes: 60,
		},
	}
	startTime := time.Now()
	version := VersionWithTimeRange{
		Version: v1,
		Start:   &startTime,
	}
	versionArr := []*VersionWithTimeRange{&version}
	history := &ClusterVersionHistory{versionArr, v1}

	config.Set(&cfg)
	defer config.FlushAll()

	mockCacher := cache_testing.SetMock(t)
	mockCacher.On("Get", testContext, cache_testing.KeyFromType[ClusterVersionHistory](), "cluster_1").Return(history, nil)

	client := Client{
		client: api.NewAPIClient(&api.Configuration{}),
	}

	h, err := client.GetClustersVersionsHistories(testContext, []string{"cluster_1"})
	assert.NoError(t, err)
	assert.Equal(t, history, h["cluster_1"])
}

func TestClient_GetHistoriesFromClient(t *testing.T) {
	cfg := ClustersConfiguration{
		ServiceUrl: "url",
		Cache: &CacheConfig{
			VersionHistoriesTtlMinutes: 60,
		},
	}

	ts := time.Now()
	apiHistory := api.ClusterHistory{
		ClusterId:       "cluster_1",
		Version:         "2.11",
		UpdateTimestamp: ts.Unix(),
	}
	apiHistory2 := api.ClusterHistory{
		ClusterId:       "cluster_1",
		Version:         "2.12",
		UpdateTimestamp: ts.Add(12 * time.Minute).Unix(),
	}

	config.Set(&cfg)
	defer config.FlushAll()

	mockCacher := cache_testing.SetMock(t)
	mockCacher.On("Get", testContext, cache_testing.KeyFromType[ClusterVersionHistory](), "cluster_1").Return(nil, nil)
	mockCacher.On("Set", testContext, cache_testing.KeyFromType[ClusterVersionHistory](), "cluster_1", mock.Anything).Return(nil)

	client := Client{
		client: api.NewAPIClient(&api.Configuration{}),
	}
	client.client.InternalAPI = NewInternalApiServiceMock([]api.ClusterHistory{apiHistory2, apiHistory}, nil)

	h, err := client.GetClustersVersionsHistories(testContext, []string{"cluster_1"})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(h["cluster_1"].Versions))
	assert.Equal(t, "2.11.0", (h["cluster_1"].Versions)[0].Version.String())
	assert.Equal(t, "2.12.0", (h["cluster_1"].Versions)[1].Version.String())
	assert.Equal(t, "2.12.0", h["cluster_1"].CurrentVersion.String())
}

func TestClient_GetHistoriesFromClientFailed(t *testing.T) {
	cfg := ClustersConfiguration{
		ServiceUrl: "url",
		Cache: &CacheConfig{
			VersionHistoriesTtlMinutes: 60,
		},
	}

	history := api.ClusterHistory{
		ClusterId:       "cluster_1",
		Version:         "2.11",
		UpdateTimestamp: time.Now().Unix(),
	}

	config.Set(&cfg)
	defer config.FlushAll()

	mockCacher := cache_testing.SetMock(t)
	mockCacher.On("Get", testContext, cache_testing.KeyFromType[ClusterVersionHistory](), "cluster_1").Return(nil, nil)

	client := Client{
		client: api.NewAPIClient(&api.Configuration{}),
	}
	client.client.InternalAPI = NewInternalApiServiceMock([]api.ClusterHistory{history}, errors.New("failed :("))
	h, err := client.GetClustersVersionsHistories(testContext, []string{"cluster_1"})
	assert.Error(t, err)
	assert.Nil(t, h)
}

func TestClient_GetClusterNameFromCache(t *testing.T) {
	cfg := ClustersConfiguration{
		ServiceUrl: "url",
		Cache: &CacheConfig{
			ClusterNamesTtlMinutes: 60,
		},
	}
	clusterName := "cluster_1_name"

	config.Set(&cfg)
	defer config.FlushAll()

	mockCacher := cache_testing.SetMock(t)
	mockCacher.On("Get", testContext, cache_testing.KeyFromType[api.GetClusterName200Response](), "cluster_1").Return(&api.GetClusterName200Response{Name: clusterName}, nil)

	client := &Client{}

	c, err := client.GetClusterName(testContext, "cluster_1")
	assert.NoError(t, err)
	assert.Equal(t, c, clusterName)
}

func TestClient_RegisterCacheStores(t *testing.T) {
	cfg := ClustersConfiguration{
		ServiceUrl: "url",
		Cache: &CacheConfig{
			ClusterNamesTtlMinutes: 60,
		},
	}
	config.Set(&cfg)
	defer config.FlushAll()

	mockCacher := cache_testing.SetMock(t)
	mockCacher.On("RegisterStore", cache_testing.KeyFromType[api.GetClusterName200Response](), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockCacher.On("RegisterStore", cache_testing.KeyFromType[ClusterVersionHistory](), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	registerCacheStores()
}
