/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"time"
)

// ServiceConfig holds top-level service configuration
type Service struct {
	MaxWaitClusterRemoteConfigStatusUpdate time.Duration `mapstructure:"max_wait_for_cluster_remote_config_status_update"`
	MaxWaitForClusterGracefulDelete        time.Duration `mapstructure:"max_wait_for_cluster_graceful_delete"`
	ClusterCleanerRunTiming                time.Duration `mapstructure:"cluster_cleaner_run_timing"`
	PubsubInitialized                      bool
}

func (s *Service) SetDefaults() {
	s.MaxWaitClusterRemoteConfigStatusUpdate = 5 * time.Minute
	s.MaxWaitForClusterGracefulDelete = 30 * time.Minute
	s.ClusterCleanerRunTiming = 1 * time.Hour
	s.PubsubInitialized = false
}

// Environment holds environment-specific configuration
type Environment struct {
	EnvIdentifier string `mapstructure:"environment_identifier"`
	ReleaseName   string `mapstructure:"environment_release_name"`
	FipsMode      string `mapstructure:"fips_mode" validate:"oneof=off on only"`
}

func (e *Environment) SetDefaults() {
	e.EnvIdentifier = "production"
	e.FipsMode = "off"
}

// Server holds server configuration
type Server struct {
	ServicePort int `mapstructure:"service_port"`
}

func (s *Server) SetDefaults() {
	s.ServicePort = 8080
}

// Auth holds authentication service configuration
type Auth struct {
	IdmServiceUrl            string `mapstructure:"identity_mng_client_url"`
	KeycloakExternalEndpoint string `mapstructure:"keycloak_external_endpoint"`
}

func (a *Auth) SetDefaults() {
	a.IdmServiceUrl = "runai-backend-identity-manager:8080"
}

// Backend holds backend service configuration
type Backend struct {
	TempBackendSettingsUrl string `mapstructure:"temp_backend_settings_url"`
}

func (b *Backend) SetDefaults() {
	b.TempBackendSettingsUrl = "http://runai-backend-temp-backend-settings:8080"
}

// App holds application-level configuration
type App struct {
	Host                         string        `mapstructure:"app_host"`
	ClusterHostTemplate          string        `mapstructure:"cluster_host_template"`
	ClusterConnectivityThreshold time.Duration `mapstructure:"cluster_connectivity_threshold"`
	RunaiClusterDomainSuffix     string        `mapstructure:"runai_cluster_domain_suffix"`
	OrgUnitServiceUrl            string        `mapstructure:"org_unit_svc_url"`
	TenantsManagerServiceUrl     string        `mapstructure:"tenants_manager_svc_url"`
	WorkloadsServiceUrl          string        `mapstructure:"workloads_svc_url"`
}

func (a *App) SetDefaults() {
	// Host, ClusterHostTemplate and RunaiClusterDomainSuffix intentionally have
	// no defaults. They identify the deployment's domain and must be provided
	// explicitly via configuration (e.g. APP_HOST, CLUSTER_HOST_TEMPLATE,
	// RUNAI_CLUSTER_DOMAIN_SUFFIX).
	a.ClusterConnectivityThreshold = 1 * time.Minute
	a.OrgUnitServiceUrl = "http://runai-backend-org-unit-service:8080"
	a.TenantsManagerServiceUrl = "http://runai-backend-tenants-manager:8080"
	a.WorkloadsServiceUrl = "http://runai-backend-workloads:8080"
}

// Installation holds cluster installation configuration
type Installation struct {
	HelmRegistryUrl                 string   `mapstructure:"cluster_installation_helm_registry_url"`
	HelmRegistryUsername            string   `mapstructure:"cluster_installation_helm_registry_username"`
	HelmRegistryPasswordPlaceholder string   `mapstructure:"cluster_installation_helm_registry_password_placeholder"`
	Version                         string   `mapstructure:"version"`
	MinVersion                      string   `mapstructure:"cluster_installation_min_version"`
	MaxVersion                      string   `mapstructure:"cluster_installation_max_version"`
	ExcludeVersions                 []string `mapstructure:"cluster_installation_excluded_versions"`
	InstallationType                string   `mapstructure:"installation_type" validate:"oneof=selfhosted launchpad airgap saas"` // OpenAPI ClusterControlPlaneConfig.installationType; enforced by runai config Register
	FipsMode                        string   `mapstructure:"cluster_installation_fips_mode" validate:"oneof=off on only"`
}

func (i *Installation) SetDefaults() {
	i.Version = "development"
	i.ExcludeVersions = []string{}
	i.InstallationType = "selfhosted"
	i.FipsMode = "off"
}

// CloudFormation holds cloud load-balancer host suffix configuration
type CloudFormation struct {
	TargetHostSuffix string `mapstructure:"aws_elb_suffix"`
}

func (c *CloudFormation) SetDefaults() {
	c.TargetHostSuffix = "lb.example.com"
}

// Sync holds synchronization configuration
type Sync struct {
	IntervalMinutes        int `mapstructure:"nodepool_sync_interval_minutes"`
	TenantsInChunk         int `mapstructure:"nodepool_sync_tenants_chunk_size"`
	TenantsIntervalSeconds int `mapstructure:"nodepool_sync_tenant_interval_seconds"`
}

func (s *Sync) SetDefaults() {
	s.IntervalMinutes = 180
	s.TenantsInChunk = 2
	s.TenantsIntervalSeconds = 5
}

// AsyncClusterDelete holds async cluster deletion configuration
type AsyncClusterDelete struct {
	Enabled bool `mapstructure:"async_cluster_delete_enabled"`
}

func (a *AsyncClusterDelete) SetDefaults() {
	a.Enabled = true
}

// EditableNodepoolCapabilities stays off until the project and department forms handle the capabilities of
// each nodepool; they still assume the tenant-wide settings, so a per-nodepool change makes their saves fail.
type EditableNodepoolCapabilities struct {
	Enabled bool `mapstructure:"editable_nodepool_scheduling_capabilities_enabled"`
}

func (e *EditableNodepoolCapabilities) SetDefaults() {
	e.Enabled = false
}

// ImagePullSecret holds the static image pull secret (a docker config JSON)
// supplied via Helm (the global image-pull-secret, mounted as
// STATIC_IMAGE_PULL_SECRET). Surfaced to clusters via DockerConfigJson; no
// database or external registry is involved. Self-hosted only: a single static
// value for the whole deployment.
type ImagePullSecret struct {
	Enabled               bool   `mapstructure:"image_pull_secret_enabled"`
	StaticImagePullSecret string `mapstructure:"static_image_pull_secret" config:"secret"`
}

func (c *ImagePullSecret) SetDefaults() {}

// ApplicationMetricStore holds application metric store configuration
type ApplicationMetricStore struct {
	Url      string            `mapstructure:"application_metric_store_write_url"`
	Username string            `mapstructure:"application_metric_store_write_username" config:"secret"`
	Password string            `mapstructure:"application_metric_store_write_password" config:"secret"`
	Headers  map[string]string `mapstructure:"application_metric_store_headers" config:"secret"`
	Label    string            `mapstructure:"application_metric_store_label"`
}

func (a *ApplicationMetricStore) SetDefaults() {
	a.Headers = make(map[string]string)
	a.Label = "applicative_runai"
}
