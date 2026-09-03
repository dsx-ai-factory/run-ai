/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

type Config struct {
	KeycloakURL       string `mapstructure:"keycloak_url"`
	ManagementRealm   string `mapstructure:"management_realm"`
	TenantsManagerURL string `mapstructure:"tenants_manager_url"`
	ClusterServiceURL string `mapstructure:"cluster_service_url"`
}

func (c *Config) SetDefaults() {
	c.KeycloakURL = "http://keycloak-http"
	c.TenantsManagerURL = "http://tenants-manager:8080"
	c.ClusterServiceURL = "http://runai-backend-cluster-service:8080"
}
