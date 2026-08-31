/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clusters

type ClustersConfiguration struct {
	ServiceUrl string       `mapstructure:"service_url"`
	Cache      *CacheConfig `mapstructure:"cache"`
}

type CacheConfig struct {
	VersionHistoriesTtlMinutes int `mapstructure:"version_histories_ttl_minutes"`
	ClusterNamesTtlMinutes     int `mapstructure:"get_cluster_names_ttl_minutes"`
}

func (cc *ClustersConfiguration) SetDefaults() {
	cc.ServiceUrl = "http://runai-backend-cluster-service:8080"
	cc.Cache = &CacheConfig{
		VersionHistoriesTtlMinutes: 1,
		ClusterNamesTtlMinutes:     60,
	}
}
