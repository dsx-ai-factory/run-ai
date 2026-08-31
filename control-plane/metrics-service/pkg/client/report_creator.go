/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	openapi "github.com/run-ai/runai/control-plane/metrics-service/pkg/client/generated"
)

type ReportColumn struct {
	DisplayName string
	Metric      Metric
}

type reportCreator struct {
	name              string
	description       string
	metrics           []ReportColumn
	labels            []openapi.Label
	additionalFields  map[string]interface{}
	start             *time.Time
	end               *time.Time
	getClientInstance GetClientInstancer
}

type ReportCreator interface {
	SetMetrics(metrics []ReportColumn) ReportCreator
	AddMetric(metric Metric, displayName string) ReportCreator
	SetLabels(labels []openapi.Label) ReportCreator
	AddLabel(label openapi.Label) ReportCreator
	SetAdditionalFields(additionalFields map[string]interface{}) ReportCreator
	AddAdditionalField(key string, value interface{}) ReportCreator
	Create(ctx context.Context) (*openapi.Report, int, error)
}

func NewReportCreator(name string, start *time.Time, end *time.Time, description string) *reportCreator {
	return &reportCreator{
		name:              name,
		description:       description,
		start:             start,
		end:               end,
		getClientInstance: GetClientInstance,
	}
}

func (r *reportCreator) SetMetrics(metrics []ReportColumn) ReportCreator {
	r.metrics = metrics
	return r
}

func (r *reportCreator) AddMetric(metric Metric, displayName string) ReportCreator {
	if r.metrics == nil {
		r.metrics = make([]ReportColumn, 0)
	}
	r.metrics = append(r.metrics, ReportColumn{
		DisplayName: displayName,
		Metric:      metric,
	})
	return r
}

func (r *reportCreator) SetLabels(labels []openapi.Label) ReportCreator {
	r.labels = labels
	return r
}

func (r *reportCreator) AddLabel(label openapi.Label) ReportCreator {
	r.labels = append(r.labels, label)
	return r
}

func (r *reportCreator) SetAdditionalFields(additionalFields map[string]interface{}) ReportCreator {
	r.additionalFields = additionalFields
	return r
}

func (r *reportCreator) AddAdditionalField(key string, value interface{}) ReportCreator {
	if r.additionalFields == nil {
		r.additionalFields = make(map[string]interface{})
	}
	r.additionalFields[key] = value
	return r
}

func (r *reportCreator) Create(ctx context.Context) (*openapi.Report, int, error) {
	if err := validateTimes(r.start, r.end); err != nil {
		return nil, http.StatusBadRequest, err
	}
	if err := validateMetrics(r.metrics); err != nil {
		return nil, http.StatusBadRequest, err
	}
	if err := validateLabels(r.labels); err != nil {
		return nil, http.StatusBadRequest, err
	}
	fields, err := buildReportRequestFields(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	resp, code, err := r.getClientInstance().CreateReport(ctx, fields)
	return resp, code, err
}

func validateTimes(start, end *time.Time) error {
	if start == nil || end == nil {
		return fmt.Errorf("start and end times must be provided")
	}
	if start.After(*end) {
		return fmt.Errorf("start time must be before end time")
	}
	return nil
}

func validateMetrics(metrics []ReportColumn) error {
	if len(metrics) == 0 {
		return fmt.Errorf("at least one metric must be provided")
	}
	for _, m := range metrics {
		if m.DisplayName == "" {
			return fmt.Errorf("metric display name must be provided")
		}
		if err := validateMetric(m.Metric); err != nil {
			return err
		}
	}
	return nil
}

func validateMetric(metric Metric) error {
	if !metric.IsAggregation() {
		return fmt.Errorf("metric %v must be an aggregation", metric.GetName())
	}
	if err := validateQuantile(metric.GetName(), metric.IsQuantile(), metric.GetQuantile()); err != nil {
		return err
	}
	return nil
}

func validateLabels(labels []openapi.Label) error {
	for _, label := range labels {
		if label.Name == "" {
			return fmt.Errorf("label name must be provided")
		}
		if label.DisplayName == "" {
			return fmt.Errorf("label display name must be provided")
		}
	}
	return nil
}

func buildReportRequestFields(rc *reportCreator) (*openapi.ReportRequestFields, error) {
	metrics, err := buildReportMetrics(rc.metrics)
	if err != nil {
		return nil, err
	}
	fields := &openapi.ReportRequestFields{
		Name:             rc.name,
		Description:      &rc.description,
		Metrics:          metrics,
		Labels:           rc.labels,
		AdditionalFields: rc.additionalFields,
	}
	fields.SetRange(*openapi.NewRange(*rc.start, *rc.end))
	return fields, nil
}

func buildReportMetrics(metrics []ReportColumn) ([]openapi.MetricInformationForReport, error) {
	reportMetrics := make([]openapi.MetricInformationForReport, 0, len(metrics))
	for _, m := range metrics {
		filters, err := filtersToMap(m.Metric)
		if err != nil {
			return nil, fmt.Errorf("failed processing %v metrics filters: %v", m.DisplayName, err)
		}
		reportMetric := openapi.MetricInformationForReport{
			Name:        m.Metric.GetName(),
			DisplayName: m.DisplayName,
			Filters:     &filters,
			Quantile:    *openapi.NewNullableFloat32(m.Metric.GetQuantile()),
		}
		reportMetrics = append(reportMetrics, reportMetric)
	}
	return reportMetrics, nil
}
