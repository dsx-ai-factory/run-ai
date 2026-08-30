package database

import (
	"github.com/run-ai/runai/control-plane/metrics-service/internal/database/models"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/generated/pkg/api"
)

func NewReportFromApi(request api.ReportRequestFields, metricQueries map[string]string) models.Report {
	additionalFields := request.AdditionalFields
	if additionalFields == nil {
		additionalFields = map[string]interface{}{}
	}

	metrics := []*models.Metric{}
	for _, requestMetric := range request.Metrics {
		metrics = append(metrics, &models.Metric{
			DisplayName: requestMetric.DisplayName,
			Query:       metricQueries[requestMetric.DisplayName],
		})
	}

	report := models.Report{
		Name:             request.Name,
		Description:      request.Description,
		Labels:           newLabelsFromApi(request.Labels),
		Start:            request.Range.Start,
		End:              request.Range.End,
		AdditionalFields: additionalFields,
		Metrics:          metrics,
	}

	return report
}

func newLabelsFromApi(request []api.Label) models.LabelArray {
	labels := []models.Label{}

	for _, label := range request {
		labels = append(labels, models.Label{
			Name:        label.Name,
			DisplayName: label.DisplayName,
		})
	}

	return labels
}
