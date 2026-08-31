/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
)

const chartDownloadPathFormat = "%s/api/v1/ai-applications/charts/%s/content"

var errChartBaseURLNotConfigured = errors.New("chart download base url is not configured")

// chartDownloadURLBuilder builds the cluster-facing download URL for a stored chart. It is
// constructed once per sync poll and reused for every uploaded-chart application in that batch.
type chartDownloadURLBuilder struct {
	baseURL string
}

func newChartDownloadURLBuilder(conf *config.ChartUploadConfig) *chartDownloadURLBuilder {
	return &chartDownloadURLBuilder{baseURL: conf.PublicBaseURL}
}

func (c *chartDownloadURLBuilder) downloadURL(chartID string) (string, error) {
	if c.baseURL == "" {
		return "", errChartBaseURLNotConfigured
	}
	return fmt.Sprintf(chartDownloadPathFormat,
		strings.TrimSuffix(c.baseURL, "/"), url.PathEscape(chartID)), nil
}
