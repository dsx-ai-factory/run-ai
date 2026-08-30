package config

type Configuration struct {
	UseMocks bool `mapstructure:"use_mocks"`
}

func (c *Configuration) SetDefaults() {
	c.UseMocks = false
}

type General struct {
	ReleaseName string `mapstructure:"release_name"`
}

func (c *General) SetDefaults() {}

type Server struct {
	ServicePort int `mapstructure:"service_port"`
}

func (c *Server) SetDefaults() {
	c.ServicePort = 8080
}

type Auth struct {
	ManagementRealm    string
	ManagementGroups   []string
	IdentityServiceUrl string `mapstructure:"identity_mng_url"`
	ServiceUrl         string `mapstructure:"auth_service_url"`
}

func (c *Auth) SetDefaults() {
	c.IdentityServiceUrl = "http://identity-manager:8080"
	c.ServiceUrl = "https://localhost.com"
}

type Tenant struct {
	Name            string `mapstructure:"tenant_name"`
	Username        string `mapstructure:"admin_username" validate:"required_with=Name"`
	Password        string `mapstructure:"admin_password" validate:"required_with=Name" config:"secret"`
	MngName         string `mapstructure:"tenant_mng_name"`
	MngUsername     string `mapstructure:"tenant_mng_user" validate:"required_with=MngName" config:"secret"`
	MngPassword     string `mapstructure:"tenant_mng_password" validate:"required_with=MngName" config:"secret"`
	DomainTemplate  string `mapstructure:"tenant_domain_template" validate:"required"`
	CacheTtlMinutes int    `mapstructure:"cache_ttl_minutes"`
}

func (c *Tenant) SetDefaults() {
	c.CacheTtlMinutes = 60 * 24 // 24 hours
}

type Realm struct {
	Name    string   `json:"realm_name"`
	Domain  string   `json:"domain"`
	Id      string   `json:"id"`
	Clients []Client `json:"clients"`
}
type Client struct {
	ClientId string `json:"clientId"`
	Id       string `json:"id"`
	Secret   string `json:"secret"`
}

type BffService struct {
	ServiceUrl string `mapstructure:"bff_service_url"`
}

func (c *BffService) SetDefaults() {
	c.ServiceUrl = "http://runai-backend-bff-service:8080"
}
