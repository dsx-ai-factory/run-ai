/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"fmt"

	"github.com/run-ai/runai/runai-common-packages/config"
)

type metricConfigurationProcessor struct{}

func NewMetricConfigurationProcessor() *metricConfigurationProcessor {
	return &metricConfigurationProcessor{}
}

func (m *metricConfigurationProcessor) Process() (map[string]*ProcessedMetricConfiguration, error) {
	conf := config.Get[*MetricsConfiguration]()

	processedMetricConfigs := make(map[string]*ProcessedMetricConfiguration)
	for metricName, metricConfig := range conf.Metrics {
		processedMetricConfig, err := processMetricConfiguration(metricConfig)
		if err != nil {
			return nil, err
		}
		processedMetricConfigs[metricName] = processedMetricConfig
	}
	return processedMetricConfigs, nil
}

func processMetricConfiguration(metricConfig *MetricConfiguration) (*ProcessedMetricConfiguration, error) {
	err := validateNeededVersionsExist(metricConfig)
	if err != nil {
		return nil, err
	}
	validSemverKeysQueries, err := transformKeysIntoValidSemvers(metricConfig.Queries)
	if err != nil {
		return nil, err
	}
	versionedQueries, err := createVersionedQueries(validSemverKeysQueries)
	if err != nil {
		return nil, err
	}
	return &ProcessedMetricConfiguration{
		CrossVersionQuery: metricConfig.CrossVersionQuery,
		VersionedQueries:  versionedQueries,
	}, nil
}

func validateNeededVersionsExist(metricConfig *MetricConfiguration) error {
	if metricConfig.CrossVersionQuery == "" {
		return fmt.Errorf("missing cross cluster version")
	}
	if len(metricConfig.Queries) == 0 {
		return fmt.Errorf("need to have at least one version other than cross cluster version")
	}
	return nil
}

func transformKeysIntoValidSemvers(queries map[string]string) (map[string]string, error) {
	validSemverKeysQueries := make(map[string]string)
	for v := range queries {
		semver, err := stringToSemver(v)
		if err != nil {
			return nil, err
		}
		// After conversion to semver, we might lose uniquness of key (e.g 2.9, 2.9.0 --> 2.9.0, 2.9.0)
		if _, ok := validSemverKeysQueries[semver.String()]; ok {
			return nil, fmt.Errorf("version: %s is not unique", v)
		}
		validSemverKeysQueries[semver.String()] = queries[v]
	}
	return validSemverKeysQueries, nil
}

func createVersionedQueries(queries map[string]string) ([]*VersionedQuery, error) {
	versionedQueries := make([]*VersionedQuery, 0)
	for v, query := range queries {
		versionedQuery, err := createVersionedQuery(v, query)
		if err != nil {
			return nil, err
		}
		versionedQueries = append(versionedQueries, versionedQuery)
	}
	return versionedQueries, nil
}

func createVersionedQuery(v string, query string) (*VersionedQuery, error) {
	semver, err := stringToSemver(v)
	if err != nil {
		return nil, err
	}
	return &VersionedQuery{Version: semver, Query: query}, nil
}
