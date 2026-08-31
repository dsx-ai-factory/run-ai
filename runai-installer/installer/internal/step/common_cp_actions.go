/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package step

import (
	"context"
	"fmt"
	"net/http"
	"time"

	runaiclients "github.com/run-ai/runai/runai-installer/clients/pkg/clients"
	configinputs "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
)

func EvaluateCPStatus(step *RunAIInstallStepImpl, clientProvider RunAIAPIClientProvider, runaiData *model.RunAIInstallData) {
	if step.ShouldSkipStatusEvaluation() {
		return
	}

	logging.Infof("Checking Control Plane installation status...")
	timeout := step.GetTimeout()
	url, err := GetCPUrl(runaiData)
	if err != nil {
		logging.Errorf("Failed to get Control Plane URL: %v", err)
		step.SetStatus(model.Failed, err)
		return
	}
	logging.Infof("Connecting to Control Plane at %s...", url)
	caCertPath := runaiData.GetInputValue(configinputs.InstallInputCAChainPath)
	apiClient, err := clientProvider(url, caCertPath)
	if err != nil {
		elapsed := step.GetElapsedTime()
		logging.Infof("Control Plane not yet accessible (elapsed: %v, timeout: %v): %v", elapsed, timeout, err)
		return
	}
	_, statusCode, err := apiClient.GetClusters(context.Background())
	if err != nil {
		elapsed := step.GetElapsedTime()
		logging.Infof("Control Plane API not ready (elapsed: %v, timeout: %v): %v", elapsed, timeout, err)
		return
	}
	if statusCode != 200 {
		elapsed := step.GetElapsedTime()
		logging.Infof("Control Plane API returned status %d (elapsed: %v, timeout: %v)", statusCode, elapsed, timeout)
		return
	}
	logging.Infof("Control Plane is running and accessible")
	step.SetStatus(model.Installed, nil)
}

func EvaluateClusterStatus(step *RunAIInstallStepImpl, clientProvider RunAIAPIClientProvider, clusterUuid string, runaiData *model.RunAIInstallData) {
	if step.ShouldSkipStatusEvaluation() {
		return
	}

	logging.Infof("Checking cluster chart installation status...")
	timeout := step.GetTimeout()
	url, err := GetCPUrl(runaiData)
	if err != nil {
		logging.Errorf("Failed to get Control Plane URL: %v", err)
		step.SetStatus(model.Failed, err)
		return
	}
	caCertPath := runaiData.GetInputValue(configinputs.InstallInputCAChainPath)
	apiClient, err := clientProvider(url, caCertPath)
	if err != nil {
		elapsed := step.GetElapsedTime()
		logging.Infof("Control Plane API not yet accessible for cluster status check (elapsed: %v, timeout: %v): %v", elapsed, timeout, err)
		return
	}
	if clusterUuid == "" {
		logging.Infof("Cluster UUID not yet available")
		return
	}

	logging.Infof("Checking liveness for cluster UUID: %s", clusterUuid)
	cluster, statusCode, err := apiClient.GetClusterByUUID(context.Background(), clusterUuid)
	if err != nil {
		elapsed := step.GetElapsedTime()
		logging.Infof("Failed to get cluster status (elapsed: %v, timeout: %v, will retry): %v", elapsed, timeout, err)
		return
	}
	if statusCode != 200 {
		elapsed := step.GetElapsedTime()
		logging.Infof("Failed to get cluster status, HTTP %d (elapsed: %v, timeout: %v, will retry)", statusCode, elapsed, timeout)
		return
	}

	if cluster.HasLastLiveness() {
		lastLiveness := cluster.GetLastLiveness()
		oneMinuteAgo := time.Now().Add(-1 * time.Minute)
		if lastLiveness.After(oneMinuteAgo) && lastLiveness.Before(time.Now()) {
			logging.Infof("Cluster is alive and reporting to Control Plane")
			step.SetStatus(model.Installed, nil)
			return
		}
	}

	elapsed := step.GetElapsedTime()
	logging.Infof("Cluster not yet reporting liveness (elapsed: %v, timeout: %v)", elapsed, timeout)
}

func IsFQDNReachable(url string) bool {
	client := http.Client{
		Timeout: 1 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return true
}

func GetCPUrl(runaiData *model.RunAIInstallData) (string, error) {
	fqdn := runaiData.GetInputValue(configinputs.InstallInputFQDN)
	if fqdn == "" {
		return "", fmt.Errorf("FQDN is required but not provided")
	}
	fqdn = fmt.Sprintf("https://%s", fqdn)
	return fqdn, nil
}

func GetClusterByNameViaProvider(ctx context.Context, clientProvider RunAIAPIClientProvider, runaiData *model.RunAIInstallData) (string, error) {
	clusterName := runaiData.GetInputValue(configinputs.InstallInputClusterName)
	if clusterName == "" {
		return "", fmt.Errorf("cluster name is required but not provided")
	}

	url, err := GetCPUrl(runaiData)
	if err != nil {
		return "", fmt.Errorf("failed to get Control Plane URL: %w", err)
	}

	caCertPath := runaiData.GetInputValue(configinputs.InstallInputCAChainPath)
	apiClient, err := clientProvider(url, caCertPath)
	if err != nil {
		return "", fmt.Errorf("failed to get clients: %w", err)
	}

	clusters, statusCode, err := apiClient.GetClusters(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get clusters: %w", err)
	}
	if statusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get clusters, status code: %d", statusCode)
	}

	for _, cluster := range clusters {
		if cluster.Name == clusterName {
			logging.Infof("[RunAI Assets] Found cluster '%s' with UUID: %s", clusterName, cluster.Uuid)
			return cluster.Uuid, nil
		}
	}

	return "", &runaiclients.NotFoundError{Resource: "cluster", Name: clusterName}
}
