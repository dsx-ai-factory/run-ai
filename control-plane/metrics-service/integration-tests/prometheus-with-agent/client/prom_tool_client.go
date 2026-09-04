/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

const (
	port                     = 8090
	prometheusPort           = 9090
	createMetricsEndpoint    = "tsdb/create-blocks-from/openmetrics"
	clearMetricsMemEndpoint  = "api/v1/admin/tsdb/delete_series"
	clearMetricsDiskEndpoint = "api/v1/admin/tsdb/clean_tombstones"
)

func ClearMetrics(metricNames []string) error {
	err := clearMetricsFromMem(metricNames)
	if err != nil {
		return fmt.Errorf("failed to clear metrics from memory: %w", err)
	}
	err = clearMetricsFromDisk()
	if err != nil {
		return fmt.Errorf("failed to clear metrics from disk: %w", err)
	}
	return nil
}

func clearMetricsFromMem(metricNames []string) error {
	host := getEnvOrDefault("PROMETHEUS_HOST", "localhost")
	req, err := http.NewRequest("POST", //nolint:noctx // integration test helper, no context available
		fmt.Sprintf("http://%s:%d/%s", host, prometheusPort, clearMetricsMemEndpoint),
		nil)
	if err != nil {
		return err
	}
	query := req.URL.Query()
	for _, metricName := range metricNames {
		query.Add("match[]", metricName)
	}
	req.URL.RawQuery = query.Encode()
	_, err = sendRequestAndCheckResponse(req, http.StatusNoContent)
	return err
}

func clearMetricsFromDisk() error {
	host := getEnvOrDefault("PROMETHEUS_HOST", "localhost")
	req, err := http.NewRequest("POST", //nolint:noctx // integration test helper, no context available
		fmt.Sprintf("http://%s:%d/%s", host, prometheusPort, clearMetricsDiskEndpoint),
		nil)
	if err != nil {
		return err
	}
	_, err = sendRequestAndCheckResponse(req, http.StatusNoContent)
	return err
}

func WaitForMetrics(metricNames []string, pollIntervalSeconds int, totalSeconds int) error {
	ticker := time.NewTicker(time.Duration(pollIntervalSeconds) * time.Second)
	endTime := time.Now().Add(time.Duration(totalSeconds) * time.Second)
	defer ticker.Stop()
	for ; time.Now().Before(endTime); <-ticker.C {
		valid, err := ValidateMetrics(metricNames)
		if err != nil {
			return err
		}
		if valid {
			return nil
		}
	}
	return fmt.Errorf("metrics did not appear in %d seconds", totalSeconds)
}

func ValidateMetrics(metricNames []string) (bool, error) {
	host := getEnvOrDefault("PROMETHEUS_HOST", "localhost")
	req, _ := http.NewRequest("GET", //nolint:noctx // integration test helper, no context available
		fmt.Sprintf("http://%s:%d/api/v1/label/__name__/values", host, prometheusPort),
		nil)
	res, err := sendRequestAndCheckResponse(req, http.StatusOK)
	if err != nil {
		return false, err
	}
	for _, metricName := range metricNames {
		if !strings.Contains(res, metricName) {
			return false, nil
		}
	}
	return true, nil
}

func PostMetrics(openMetrics string) error {
	host := getEnvOrDefault("PROMETHEUS_HOST", "localhost")
	req, _ := http.NewRequest("POST", //nolint:noctx // integration test helper, no context available
		fmt.Sprintf("http://%s:%d/%s", host, port, createMetricsEndpoint),
		strings.NewReader(openMetrics))
	_, err := sendRequestAndCheckResponse(req, http.StatusOK)
	return err
}

func sendRequestAndCheckResponse(req *http.Request, expectedStatusCode int) (responseBody string, err error) {
	client := http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed sending request: %w", err)
	}

	responseBody, err = bodyToString(response.Body)
	logging.Infof("got response: %s", responseBody)
	if expectedStatusCode != response.StatusCode {
		return "", fmt.Errorf("got an unexpected error code. Recieved - %d, expected - %d", response.StatusCode, expectedStatusCode)
	}
	return responseBody, err
}

func bodyToString(r io.ReadCloser) (string, error) {
	defer r.Close()
	bodyBytes, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("failed to read capturedData: %w", err)
	}
	return string(bodyBytes), nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.Trim(value, " ")
	}
	return defaultVal
}
