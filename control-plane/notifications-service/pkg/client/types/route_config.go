package types

type BasicRouteConfig struct {
	RouteId    string            `mapstructure:"route_id" json:"route_id"`
	ListenerId string            `mapstructure:"listener_id" json:"listener_id"`
	TenantId   string            `mapstructure:"tenant_id" json:"tenant_id"`
	Enrich     EnrichingConfig   `mapstructure:"enrich" json:"enrich"`
	Filtering  FilteringConfig   `mapstructure:"filtering" json:"filtering"`
	NotifyIds  []string          `mapstructure:"notify_ids" json:"notify_ids"`
	Metadata   map[string]string `mapstructure:"metadata" json:"metadata"`
}

type UserRouteConfig struct {
	BasicRouteConfig `mapstructure:",squash" json:",inline"`
	UserId           string `mapstructure:"user_id" json:"user_id"`
}
