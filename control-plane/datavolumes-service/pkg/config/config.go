package config

import (
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

type ServerConfig struct {
	ServicePort int `mapstructure:"service_port"`
}

func (s *ServerConfig) SetDefaults() {
	s.ServicePort = 8080
}

type OrgHierarchyManagerConfig struct {
	OrgUnitServiceUrl string `mapstructure:"org_unit_service_url"`
	ClusterServiceUrl string `mapstructure:"cluster_service_url"`
	BackendUrl        string `mapstructure:"backend_url"`
}

func (c *OrgHierarchyManagerConfig) SetDefaults() {
	c.OrgUnitServiceUrl = "http://runai-backend-org-unit-service:8080"
	c.ClusterServiceUrl = "http://runai-backend-cluster-service:8080"
	c.BackendUrl = "http://runai-backend-temp-backend-settings:8080"
}

func Init() {
	if err := runai_config.Register[*ServerConfig](); err != nil {
		logging.Fatalf("failed to register server config: %v", err)
	}
	if err := runai_config.Register[*OrgHierarchyManagerConfig](); err != nil {
		logging.Fatalf("failed to register org hierarchy manager config: %v", err)
	}
}
