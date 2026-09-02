/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/clusters"
)

const (
	CROSS_CLUSTER_VERSION = "x"
)

type metricQueryPicker struct {
	ProcessedMetricConfigurations map[string]*ProcessedMetricConfiguration
}

func NewMetricQueryPicker(processedMetricConfigurations map[string]*ProcessedMetricConfiguration) *metricQueryPicker {
	for _, mc := range processedMetricConfigurations {
		sortQueriesByVersion(mc)
	}
	return &metricQueryPicker{
		ProcessedMetricConfigurations: processedMetricConfigurations,
	}
}

func (m *metricQueryPicker) GetQueryMinVersion(metricName string) (*version.Version, error) {
	metricConfig, ok := m.ProcessedMetricConfigurations[metricName]
	if !ok {
		return nil, fmt.Errorf("metric: %s not found", metricName)
	}

	return metricConfig.VersionedQueries[0].Version, nil
}

func (m *metricQueryPicker) GetQueryByVersion(metricName string, v string) (string, error) {
	if v == "" {
		return "", fmt.Errorf("version is empty")
	}
	metricConfig, ok := m.ProcessedMetricConfigurations[metricName]
	if !ok {
		return "", fmt.Errorf("metric: %s not found", metricName)
	}
	if v == CROSS_CLUSTER_VERSION {
		return metricConfig.CrossVersionQuery, nil
	}
	if strings.HasPrefix(v, clusters.DEV_VERSION_PREFIX) {
		return m.getLatestQuery(metricConfig), nil
	}
	trimmedVer := m.trimVersionSuffix(v)
	semver, err := stringToSemver(trimmedVer)
	if err != nil {
		return "", err
	}
	query := m.chooseQuery(metricConfig, semver)
	if query == "" {
		return "", fmt.Errorf("no query found for metric: %s version %s", metricName, v)
	}
	return query, nil
}

func (m *metricQueryPicker) getLatestQuery(metricConfig *ProcessedMetricConfiguration) string {
	return metricConfig.VersionedQueries[len(metricConfig.VersionedQueries)-1].Query
}

func (m *metricQueryPicker) trimVersionSuffix(v string) string {
	return strings.Split(v, "-")[0]
}

func (m *metricQueryPicker) chooseQuery(metricConfig *ProcessedMetricConfiguration, requestedVerion *version.Version) string {
	query := ""
	for _, v := range metricConfig.VersionedQueries {
		if requestedVerion.LessThan(v.Version) {
			return query
		}
		query = v.Query
	}
	return query
}

func sortQueriesByVersion(mc *ProcessedMetricConfiguration) {
	sort.Slice(mc.VersionedQueries, func(i, j int) bool {
		return mc.VersionedQueries[i].Version.LessThan(mc.VersionedQueries[j].Version)
	})
}
