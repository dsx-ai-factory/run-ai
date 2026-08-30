package metricstores

type DatasourceType string

const (
	DatasourceTypeGrafanaLabs DatasourceType = "grafanalabs"
	DatasourceTypeThanos      DatasourceType = "thanos"
	DatasourceTypeMimir       DatasourceType = "mimir"
)

type AuthType string

const (
	AuthTypeBasic AuthType = "basic"
)

type AuthBasic struct {
	User     string `mapstructure:"user" config:"secret"`
	Password string `mapstructure:"password" config:"secret"`
}

type DatasourceAuth struct {
	Type  AuthType   `mapstructure:"type"`
	Basic *AuthBasic `mapstructure:"basic"`
}

type DatasourceEndpoint struct {
	Url  string          `mapstructure:"url"`
	Auth *DatasourceAuth `mapstructure:"auth"`
}

type Datasource struct {
	Type  DatasourceType     `mapstructure:"type"`
	Read  DatasourceEndpoint `mapstructure:"read"`
	Write DatasourceEndpoint `mapstructure:"write"`
}
