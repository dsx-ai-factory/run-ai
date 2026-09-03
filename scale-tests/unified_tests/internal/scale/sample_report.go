/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package scale

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dsx-ai-factory/run-ai/scale-tests/sdk/reports"

	"github.com/iancoleman/strcase"
)

type SamplesReport struct {
	MetaData        map[string]string                        `json:"metadata"`
	RunIdentifier   string                                   `json:"run_identifier"`
	Status          bool                                     `json:"status"`
	Results         any                                      `json:"results"`
	Metrics         map[MetricType]reports.PerformanceMetric `json:"metrics"`
	OperatorVersion string                                   `json:"operator_version"`
	CpVersion       string                                   `json:"cp_version"`
}

type MetricType string

const (
	MetricTypeTargetScrapesExceededSampleLimitTotal MetricType = "prometheus_target_scrapes_exceeded_sample_limit_total"
	MetricTypeTargetSyncFailedTotal                 MetricType = "prometheus_target_sync_failed_total"
	MetricTypeRemoteWriteLagAvg                     MetricType = "prometheus_remote_write_lag_seconds_average"
	MetricTypeRemoteWriteQueueBacklogIncreaseRate   MetricType = "prometheus_remote_write_queue_backlog_increase_rate"
	MetricTypeRemoteWriteSamplesFailedTotal         MetricType = "prometheus_remote_write_samples_failed_total"
	MetricTypeCrashRate                             MetricType = "prometheus_crash_rate"
)

func NewSamplesReport(status bool, results any) *SamplesReport {
	return &SamplesReport{
		MetaData: map[string]string{},
		Status:   status,
		Results:  results,
	}
}

func (s SamplesReport) UpdateMetaData(obj interface{}) {
	val := reflect.ValueOf(obj)
	typ := reflect.TypeOf(obj)

	if typ.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)
		isSensitiveField := strings.Contains(field.Name, "Password") || strings.Contains(field.Name, "Username") || field.Tag.Get("sensitive") == "true"
		if isSensitiveField {
			continue
		}
		s.MetaData[strcase.ToSnake(field.Name)] = fmt.Sprintf("%v", value)
	}

}

func (s SamplesReport) UpdateMetadataMap(metadata map[string]string) {
	for k, v := range metadata {
		s.MetaData[k] = v
	}
}
