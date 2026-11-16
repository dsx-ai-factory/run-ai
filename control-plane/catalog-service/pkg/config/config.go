/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"log"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
)

type Services struct {
	WholeGPUCutoffPercent      int64  `mapstructure:"whole_gpu_cutoff_percent" validate:"gte=1,lte=100"`
	InventoryRequestTimeoutSec int    `mapstructure:"inventory_request_timeout_sec" validate:"gt=0"`
	ClusterServiceUrl          string `mapstructure:"cluster_service_url" validate:"required"`
	AuthorizationUrl           string `mapstructure:"authorization_service_url" validate:"required"`
	AssetsServiceUrl           string `mapstructure:"assets_service_url" validate:"required"`
	TempBackendSettingsUrl     string `mapstructure:"temp_backend_settings_url" validate:"required"` // named after the actual service: temp-backend-settings
}

func (s *Services) SetDefaults() {
	// Rajan's initial sharing policy; this is not runtime/OOM headroom.
	s.WholeGPUCutoffPercent = 90
	s.InventoryRequestTimeoutSec = 30
	s.ClusterServiceUrl = "http://runai-backend-cluster-service:8080"
	s.AuthorizationUrl = "http://runai-backend-authorization:8080"
	s.AssetsServiceUrl = "http://runai-backend-assets-service:8080"
	s.TempBackendSettingsUrl = "http://runai-backend-temp-backend-settings:8080"
}

type Server struct {
	ServicePort int `mapstructure:"service_port" validate:"required"`
}

type NGC struct {
	MetadataCacheTTLSeconds int `mapstructure:"ngc_metadata_cache_ttl_seconds" validate:"gt=0"`
	RequestTimeoutSec       int `mapstructure:"request_timeout_sec" validate:"required"`
}

func (s *Server) SetDefaults() {
	s.ServicePort = 8080
}

func (s *NGC) SetDefaults() {
	s.MetadataCacheTTLSeconds = 15 * 60
	s.RequestTimeoutSec = 30
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
	err = config.Register[*NGC]()
	if err != nil {
		log.Fatalf("Failed to register NGC config: %v", err)
	}
}
