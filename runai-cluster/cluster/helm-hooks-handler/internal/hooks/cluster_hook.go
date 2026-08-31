/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package hooks

import (
	"context"
	"encoding/json"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"

	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	v1 "k8s.io/api/core/v1"
)

const (
	PreDeleteHookIdentifier                    = "pre-delete"
	PreInstallHookIdentifier                   = "pre-install"
	PreCrdUpgradeHookIdentifier                = "pre-crd-upgrade"
	PreResourceManagementUpgradeHookIdentifier = "pre-resource-management-upgrade"
	PreUpgradeHookIdentifier                   = "pre-upgrade"
	PostUpgradeHookIdentifier                  = "post-upgrade"
	PostInstallHookIdentifier                  = "post-install"
	RunAIControlPlaneImagePullSecretName       = "runai-reg-creds"
)

type ClusterHookParams struct {
	clients.ControlPlaneDetails
	ClusterDetails
	ImageRegistry                               string
	AdditionalImagePullSecrets                  []string
	Affinity                                    *v1.Affinity
	Tolerations                                 []v1.Toleration
	Development                                 DevelopmentParams
	ClusterVersion                              string
	NMC                                         NMCParams
	DisableInfrastructureMetricsRemoteWrite     bool
	QueueValidationIntervalMillis               int
	OperandDownIntervalMillis                   int
	ResourceManagementMigrationEnabled          bool
	ResourceManagementMigrationCleanup          bool
	ResourceManagementMigrationNodePoolsEnabled bool
	ResourceManagementMigrationNodePoolsCleanup bool
	EngineToKAIMigrationEnabled                 bool
	EnableRbacEscalation                        bool
	ClusterConfig                               json.RawMessage
	FipsMode                                    string
}

type NMCParams struct {
	Enabled                  bool
	RemoteServerAPIServerURL string
	ServiceAccountName       string
	AllowedServiceAccounts   []string
	MgmtPrometheusIngress    *common.Ingress
}

type DevelopmentParams struct {
	// Install a headless cluster
	Headless             bool
	SkipDependencyChecks bool
}

type ReleaseParams struct {
	Name      string
	Namespace string
}

type PreHooksParams struct {
	ClusterHookParams
	common.CustomCA
}

type ClusterHook interface {
	Call(context.Context, *clients.Clients) error
	GetIdentifier() string
}

func (params *ClusterHookParams) ShouldPerformControlPlaneAssistedInstallationOrUpgrade() bool {
	hasCredentials := params.BootstrapToken != "" || params.ClientSecret != ""
	return params.ControlPlaneDetails.URL != "" &&
		hasCredentials &&
		params.ClusterDetails.UID != "" &&
		params.ClusterDetails.URL != ""
}
