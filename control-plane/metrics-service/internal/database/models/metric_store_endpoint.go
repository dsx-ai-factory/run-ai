package models

type MetricStoreEndpoint struct {
	Url     string                   `gorm:"not null;"`
	Auth    *MetricStoreEndpointAuth `gorm:"type:jsonb;not null;"`
	Headers map[string]string
}
