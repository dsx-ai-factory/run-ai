/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"log"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"
)

type Services struct {
	AuthorizationUrl  string `mapstructure:"authorization_service_url" validate:"required"`
	OrgUnitUrl        string `mapstructure:"org_unit_url" validate:"required"`
	ClusterServiceUrl string `mapstructure:"cluster_service_url" validate:"required"`
	WorkloadsUrl      string `mapstructure:"workloads_service_url" validate:"required"`
	PolicyServiceUrl  string `mapstructure:"policy_service_url" validate:"required"`

	ClusterCacheExpirationSec int `mapstructure:"cluster_cache_expiration_sec" validate:"required"`
	ClusterCacheCleanupSec    int `mapstructure:"cluster_cache_cleanup_sec" validate:"required"`

	SuspendDefCacheTTLSec int `mapstructure:"suspend_def_cache_ttl_sec" validate:"required"`
}

func (s *Services) SetDefaults() {
	s.AuthorizationUrl = "http://runai-backend-authorization:8080"
	s.OrgUnitUrl = "http://runai-backend-org-unit-service:8080"
	s.ClusterServiceUrl = "http://runai-backend-cluster-service:8080"
	s.WorkloadsUrl = "http://runai-backend-workloads:8080"
	s.PolicyServiceUrl = "http://runai-backend-policy-service:8080"

	s.ClusterCacheExpirationSec = 600 // cluster cache expiration 10 minutes
	s.ClusterCacheCleanupSec = 1200   // cluster cache cleanup 20 minutes

	s.SuspendDefCacheTTLSec = 300 // suspend definition cache TTL 5 minutes
}

type Server struct {
	ServicePort         int `mapstructure:"service_port" validate:"required"`
	RetryMaxIntervalSec int `mapstructure:"retry_max_interval_sec"`
	RetryMaxElapsedSec  int `mapstructure:"retry_max_elapsed_sec"`
}

func (s *Server) SetDefaults() {
	s.ServicePort = 8080
	s.RetryMaxIntervalSec = 10
	s.RetryMaxElapsedSec = 60
}

func Init() {
	err := config.Register[*Services]()
	if err != nil {
		log.Fatalf("Failed to register services config: %v", err)
	}
	err = config.Register[*Server]()
	if err != nil {
		log.Fatalf("Failed to register server config: %v", err)
	}

	dao_utils.ConstructConfig("workloads_manager")
}
