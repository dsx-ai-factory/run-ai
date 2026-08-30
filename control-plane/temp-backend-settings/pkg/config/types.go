package config

type ServicesConfig struct {
	ClusterServiceURL       string `mapstructure:"cluster_service_url"`
	TenantManagerURL        string `mapstructure:"tenant_manager_url"`
	IdentityManagerURL      string `mapstructure:"identity_manager_url"`
	BffServiceURL           string `mapstructure:"bff_service_url"`
	OrgUnitServiceURL       string `mapstructure:"org_unit_service_url"`
	CatalogServiceURL       string `mapstructure:"catalog_service_url"`
	DatavolumesServiceURL   string `mapstructure:"datavolumes_service_url"`
	AuthorizationServiceURL string `mapstructure:"authorization_service_url"`
	ShowAlphaFlags          bool   `mapstructure:"show_alpha_flags"`
	WorkloadsServiceURL     string `mapstructure:"workloads_service_url"`
}

func (s *ServicesConfig) SetDefaults() {
	s.ClusterServiceURL = "http://runai-backend-cluster-service:8080"
	s.TenantManagerURL = "http://runai-backend-tenants-manager:8080"
	s.IdentityManagerURL = "http://runai-backend-identity-manager:8080"
	s.BffServiceURL = "http://runai-backend-bff-service:8080"
	s.OrgUnitServiceURL = "http://runai-backend-org-unit-service:8080"
	s.WorkloadsServiceURL = "http://runai-backend-workloads:8080"
	s.CatalogServiceURL = "http://runai-backend-catalog-service:8080"
	s.DatavolumesServiceURL = "http://runai-backend-datavolumes-service:8080"
	s.AuthorizationServiceURL = "http://runai-backend-authorization:8080"
	s.ShowAlphaFlags = false
}
