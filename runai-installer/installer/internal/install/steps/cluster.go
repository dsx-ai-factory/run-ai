/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"errors"
	"fmt"

	cluster "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	runaiclients "github.com/dsx-ai-factory/run-ai/runai-installer/clients/pkg/clients"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	configinputs "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	pkgvalidation "github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/validation"
)

const (
	RunaiClusterNamespace     = "runai"
	RunaiClusterTlsSecretName = "runai-cluster-domain-tls-secret"
	RunaiClusterCAChainSecret = "runai-ca-cert"
)

var _runaiClusterCAChainSecretLabelsImmutableMap = map[string]string{
	"run.ai/cluster-wide": "true",
	"run.ai/name":         "runai-ca-cert",
}

// Public function to access the map (returns a copy for immutability)
func GetImmutableRunaiClusterCAChainSecretLablesMap() map[string]string {
	// Create a new map to copy the values
	copiedMap := make(map[string]string, len(_runaiClusterCAChainSecretLabelsImmutableMap))
	for k, v := range _runaiClusterCAChainSecretLabelsImmutableMap {
		copiedMap[k] = v
	}
	return copiedMap
}

type CreateCluster struct {
	*step.RunAIInstallStepImpl
	clusterUuid         string
	runaiClientProvider step.RunAIAPIClientProvider
}

func (i *CreateCluster) GetClusterUUID() string {
	return i.clusterUuid
}

func NewCreateCluster(runaiClientProvider step.RunAIAPIClientProvider) *CreateCluster {
	c := &CreateCluster{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Create Run:ai cluster"),
		runaiClientProvider:  runaiClientProvider,
	}
	c.SetStatusEvaluator(c.EvaluateStatus)
	return c
}

func (i *CreateCluster) EvaluateStatus(runaiData *model.RunAIInstallData) {
	if i.ShouldSkipStatusEvaluation() {
		return
	}

	clusterName := runaiData.GetInputValue(configinputs.InstallInputClusterName)
	logging.Infof("[RunAI] Checking if cluster exists in Control Plane...")
	clusterUuid, err := step.GetClusterByNameViaProvider(context.Background(), i.runaiClientProvider, runaiData)
	if err != nil {
		if _, ok := err.(*runaiclients.NotFoundError); ok {
			logging.Infof("[RunAI] Cluster '%s' not found in Control Plane", clusterName)
			return
		}
		logging.Errorf("[RunAI] Failed to get cluster: %v", err)
		i.SetStatus(model.Failed, err)
		return
	}

	i.clusterUuid = clusterUuid
	logging.Infof("[RunAI] Found existing cluster '%s' (UUID: %s)", clusterName, clusterUuid)
	i.SetStatus(model.Installed, nil)
}

func (i *CreateCluster) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	if err := i.CheckContextCancellation(ctx); err != nil {
		return err
	}

	i.SetStatus(model.Installing, nil)
	clusterName := runaiData.GetInputValue(configinputs.InstallInputClusterName)
	url, err := step.GetCPUrl(runaiData)
	if err != nil {
		logging.Errorf("[RunAI] Failed to get Control Plane URL: %v", err)
		return err
	}

	if err := i.CheckContextCancellation(ctx); err != nil {
		return err
	}

	caCertPath := runaiData.GetInputValue(configinputs.InstallInputCAChainPath)
	apiClient, err := i.runaiClientProvider(url, caCertPath)
	if err != nil {
		logging.Errorf("[RunAI] Failed to get API clients: %v", err)
		return err
	}

	logging.Infof("[RunAI] Checking if cluster '%s' exists in Control Plane...", clusterName)
	existingUuid, err := step.GetClusterByNameViaProvider(ctx, i.runaiClientProvider, runaiData)
	if err == nil {
		i.clusterUuid = existingUuid
		logging.Infof("[RunAI] Cluster '%s' exists (UUID: %s), skipping creation", clusterName, existingUuid)
		i.SetStatus(model.Installed, nil)
		return nil
	}
	if errors.Is(err, &runaiclients.NotFoundError{}) {
		logging.Infof("[RunAI] Cluster '%s' not found, will create it", clusterName)
	} else {
		err := fmt.Errorf("[RunAI] Failed to check if cluster '%s' exists: %v, will attempt creation anyway", clusterName, err)
		logging.Errorf("[RunAI] failed to create cluster: %v", err)
		i.SetStatus(model.Failed, err)
		return i.GetError()
	}

	logging.Infof("[RunAI] Creating cluster '%s' in Control Plane...", clusterName)
	clusterCreationRequest := cluster.ClusterCreationRequest{
		Name:   clusterName,
		Domain: &url,
	}
	displayedCluster, statusCode, err := apiClient.CreateCluster(ctx, clusterCreationRequest)
	if err != nil {
		err := fmt.Errorf("failed to create cluster: %v", err)
		logging.Errorf("[RunAI] failed to create cluster: %v", err)
		i.SetStatus(model.Failed, err)
		return i.GetError()
	}
	if statusCode < 200 || statusCode >= 300 {
		err := fmt.Errorf("failed to create cluster: %v", statusCode)
		logging.Errorf("[RunAI] failed to create cluster: %v, http status: %d", err, statusCode)
		i.SetStatus(model.Failed, err)
		return i.GetError()
	}
	i.clusterUuid = displayedCluster.Uuid
	logging.Infof("[RunAI] Cluster '%s' created successfully (UUID: %s)", clusterName, displayedCluster.Uuid)
	i.SetStatus(model.Installed, nil)
	return nil
}

type GetClusterInstallationDetails struct {
	*step.RunAIInstallStepImpl
	createClusterStep   *CreateCluster
	installInfo         *cluster.ClusterInstallationInfoResponse
	runaiClientProvider step.RunAIAPIClientProvider
}

func NewGetClusterInstallationDetails(createClusterStep *CreateCluster, runaiClientProvider step.RunAIAPIClientProvider) *GetClusterInstallationDetails {
	return &GetClusterInstallationDetails{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Get cluster installation details"),
		createClusterStep:    createClusterStep,
		runaiClientProvider:  runaiClientProvider,
	}
}

func (i *GetClusterInstallationDetails) GetStatusAsync(runaiData *model.RunAIInstallData) model.InstallPhase {
	return i.GetCachedStatus()
}

func (i *GetClusterInstallationDetails) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	if err := i.CheckContextCancellation(ctx); err != nil {
		return err
	}

	logging.Infof("[RunAI] Getting cluster installation details...")
	i.SetStatus(model.Installing, nil)
	url, err := step.GetCPUrl(runaiData)
	if err != nil {
		logging.Errorf("[RunAI] Failed to get Control Plane URL: %v", err)
		i.SetStatus(model.Failed, err)
		return err
	}
	domain := runaiData.GetInputValue(configinputs.InstallInputFQDN)

	if err := i.CheckContextCancellation(ctx); err != nil {
		return err
	}

	caCertPath := runaiData.GetInputValue(configinputs.InstallInputCAChainPath)
	apiClient, err := i.runaiClientProvider(url, caCertPath)
	if err != nil {
		logging.Errorf("[RunAI] Failed to get API clients: %v", err)
		return err
	}
	logging.Infof("[RunAI] Fetching installation info for cluster UUID: %s", i.createClusterStep.clusterUuid)
	installInfo, statusCode, err := apiClient.GetClusterInstallInfoByUUID(ctx, i.createClusterStep.clusterUuid, runaiData.Config.ClusterVersion, domain)
	if err != nil {
		logging.Errorf("[RunAI] Failed to get cluster installation info: %v", err)
		i.SetStatus(model.Failed, fmt.Errorf("failed to get cluster installation info: %v", err))
		return i.GetError()
	}
	if statusCode != 200 {
		logging.Errorf("[RunAI] Failed to get cluster installation info, status: %d", statusCode)
		i.SetStatus(model.Failed, fmt.Errorf("failed to get cluster installation info: %v", statusCode))
		return i.GetError()
	}
	i.installInfo = installInfo
	logging.Infof("[RunAI] Successfully retrieved cluster installation details")
	i.SetStatus(model.Installed, nil)
	return nil
}

type InstallClusterChart struct {
	*step.RunAIInstallStepImpl
	installInfoStep     *GetClusterInstallationDetails
	helmInstaller       step.HelmInstaller
	runaiClientProvider step.RunAIAPIClientProvider
}

func NewInstallClusterChart(installInfoStep *GetClusterInstallationDetails, helmInstaller step.HelmInstaller, runaiClientProvider step.RunAIAPIClientProvider) step.InstallStep {
	i := &InstallClusterChart{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Install cluster chart"),
		installInfoStep:      installInfoStep,
		helmInstaller:        helmInstaller,
		runaiClientProvider:  runaiClientProvider,
	}
	i.SetStatusEvaluator(i.EvaluateStatus)
	return i
}

func (i *InstallClusterChart) EvaluateStatus(runaiData *model.RunAIInstallData) {
	step.EvaluateClusterStatus(i.RunAIInstallStepImpl, i.runaiClientProvider, i.installInfoStep.createClusterStep.clusterUuid, runaiData)
}

func (i *InstallClusterChart) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	// Check for context cancellation before starting
	if err := i.CheckContextCancellation(ctx); err != nil {
		return err
	}

	// Update status to installing and start timer
	i.SetStatus(model.Installing, nil)
	if i.installInfoStep.installInfo == nil {
		error := fmt.Errorf("installInfo is not available")
		logging.Errorf("[RunAI] installInfo is not available: %v", error)
		i.SetStatus(model.Failed, error)
		return error
	}

	// Check for context cancellation before helm operation
	if err := i.CheckContextCancellation(ctx); err != nil {
		return err
	}

	fqdn := runaiData.GetInputValue(configinputs.InstallInputFQDN)
	logging.Infof("[RunAI] Configuring cluster chart for FQDN: %s", fqdn)
	values := []string{}
	values = append(values, "controlPlane.url="+fqdn)
	values = append(values, step.ClusterAuthValue(i.installInfoStep.installInfo))
	values = append(values, "cluster.uid="+i.installInfoStep.createClusterStep.clusterUuid)
	values = append(values, "cluster.url="+fqdn)
	if runaiData.GetInputValue(configinputs.InstallInputCAChainPath) != "" {
		values = append(values, "global.customCA.enabled=true")
		logging.Infof("[RunAI] User provided CA Chain Path: %s, Custom CA enabled for cluster", runaiData.GetInputValue(configinputs.InstallInputCAChainPath))
	}
	if runaiData.GetInputValue(configinputs.InstallInputWorkloadRouting) == pkgvalidation.WorkloadRoutingHost {
		values = append(values, "clusterConfig.global.subdomainSupport=true")
	} else {
		values = append(values, "clusterConfig.global.subdomainSupport=false")
	}
	version := runaiData.Config.ClusterVersion
	repo := i.getRepoName(runaiData.Config.EffectiveClusterEnv())
	logging.Infof("[RunAI] Installing cluster chart version %s from %s...", version, repo)
	err := i.helmInstaller.InstallOrUpgradeRelease(ctx, repo,
		version, step.ClusterReleaseName, step.ClusterChartName, step.ClusterNamespace, false, values)
	if err != nil {
		logging.Errorf("[RunAI] Failed to install cluster chart: %v", err)
		i.SetStatus(model.Failed, err)
		return err
	}
	logging.Infof("[RunAI] Run:ai cluster Helm chart deployed, waiting for readiness")
	return nil
}

func (i *InstallClusterChart) getRepoName(env config.RunaiEnvType) string {
	return step.ClusterChartRepo(env)
}

// GetTimeoutHint returns troubleshooting guidance when the step times out.
func (i *InstallClusterChart) GetTimeoutHint() string {
	return "Cluster did not report liveness within timeout - this may indicate invalid registry credentials or other deployment issues. Check pod status with: kubectl get pods -n runai"
}
