package tenants

type TenantsConfiguration struct {
	ServiceUrl string `mapstructure:"service_url" validate:"required,url"`
}

func (tc *TenantsConfiguration) SetDefaults() {
	tc.ServiceUrl = "http://runai-backend-tenants-manager:8080"
}
