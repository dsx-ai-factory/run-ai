package models

import (
	"github.com/google/uuid"
)

const MetricsTable = "report_metrics_queries"

type Metric struct {
	ReportID    uuid.UUID
	DisplayName string
	Query       string
}

func (m *Metric) TableName() string {
	return MetricsTable
}
