package models

import metrics_client "github.com/run-ai/runai/control-plane/metrics-service/pkg/client"

type ExportableMetric struct {
	Metric        metrics_client.Metric
	Name          string
	Description   string
	LabelsMapping map[string]string
}
