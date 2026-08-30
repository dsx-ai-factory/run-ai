package datasources

type DatasourcesConfiguration struct {
	DatasourceUrl string `mapstructure:"datasource_url" validate:"required,url"`
}

func (dc *DatasourcesConfiguration) SetDefaults() {
	dc.DatasourceUrl = "http://runai-backend-thanos-query:9090"
}
