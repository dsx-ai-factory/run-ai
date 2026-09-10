/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

type Server struct {
	ServicePort            int    `mapstructure:"service_port"`
	WorkloadsUrl           string `mapstructure:"workloads_url"`
	OrgUnitServiceUrl      string `mapstructure:"org_unit_service_url"`
	ClusterServiceUrl      string `mapstructure:"cluster_service_url"`
	TenantServiceUrl       string `mapstructure:"tenant_service_url"`
	DatavolumesServiceUrl  string `mapstructure:"datavolumes_service_url"`
	TempBackendSettingsUrl string `mapstructure:"temp_backend_settings_url"`
	IsCSP                  bool   `mapstructure:"is_csp"`
}

func (s *Server) SetDefaults() {
	s.ServicePort = 8080
	s.WorkloadsUrl = "http://runai-backend-workloads:8080"
	s.OrgUnitServiceUrl = "http://runai-backend-org-unit-service:8080"
	s.ClusterServiceUrl = "http://runai-backend-cluster-service:8080"
	s.TenantServiceUrl = "http://runai-backend-tenants-manager:8080"
	s.DatavolumesServiceUrl = "http://runai-backend-datavolumes:8080"
	s.TempBackendSettingsUrl = "http://runai-backend-temp-backend-settings:8080"
	s.IsCSP = false
}

type SyncJob struct {
	SyncJobIntervalHours         int `mapstructure:"sync_job_interval_hours"`
	SyncJobTenantIntervalSeconds int `mapstructure:"sync_job_interval_tenant_seconds"`
}

func (s *SyncJob) SetDefaults() {
	s.SyncJobIntervalHours = 168
	s.SyncJobTenantIntervalSeconds = 5
}
