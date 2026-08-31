/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package util

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/run-ai/runai/runai-cluster/cluster/common/utils/options"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	cmFile = "/runai/config/params.json"

	defaultQueueTimeoutMinutes        = 10
	defaultQueueIntervalSeconds       = 3
	defaultOperandDownIntervalSeconds = 5
)

type CMConfigParams struct {
	Affinity              *v1.Affinity    `json:"affinity,omitempty"`
	Tolerations           []v1.Toleration `json:"tolerations,omitempty"`
	MgmtPrometheusIngress *common.Ingress `json:"mgmtPrometheusIngress,omitempty"`
	// We use json.RawMessage to preserve null values
	ClusterConfig json.RawMessage `json:"clusterConfig,omitempty"`
}

type Options struct {
	PreInstall                   bool
	PostInstall                  bool
	PreUpgrade                   bool
	PreCrdUpgrade                bool
	PreResourceManagementUpgrade bool
	PostUpgrade                  bool
	PreDelete                    bool
	ValidateClusterConfig        bool
	ControlPlaneURL              string
	ControlPlaneClientID         string
	ControlPlaneClientSecret     FileContentFlag
	ControlPlaneBootstrapToken   FileContentFlag
	ClusterUID                   string
	ClusterURL                   string
	CustomCaEnabled              bool
	CustomCaSecretOptional       bool
	CustomCaSecretName           string
	CustomCaSecretKey            string
	ImageRegistry                string
	AdditionalImagePullSecrets   options.ArrayFlags
	ZapOptions                   zap.Options
	Verbosity                    int
	ClusterVersion               string
	cMFilePath                   string
	CMConfigParams               CMConfigParams
	EnableNMC                    bool
	NMCRemoteClusterURL          string
	NMCServiceAccountName        string
	NMCAllowedServiceAccounts    options.ArrayFlags

	Development                             development
	DisableInfrastructureMetricsRemoteWrite bool

	QueueValidationTimeoutMillis  int
	QueueValidationIntervalMillis int
	OperandDownIntervalMillis     int
	EngineToKAIMigrationEnabled   bool

	ResourceManagementMigrationEnabled bool
	ResourceManagementMigrationCleanup bool

	ResourceManagementMigrationNodePoolsEnabled bool
	ResourceManagementMigrationNodePoolsCleanup bool

	EnableRbacEscalation     bool
	RequiresKubernetesAccess bool
	FipsMode                 string
}

type development struct {
	// Install a headless cluster without control-plane communication
	Headless bool

	SkipDependencyChecks bool
}

func SetOptions() *Options {
	return parse(flag.CommandLine)
}

func parse(flagSet *flag.FlagSet) *Options {
	opts := &Options{}
	opts.parseCommandLineArgs(flagSet)
	opts.loadConfigFromFile()
	return opts
}

func (opts *Options) loadConfigFromFile() {
	fileContent, err := os.ReadFile(opts.cMFilePath)
	if err != nil {
		return
	}
	err = json.Unmarshal(fileContent, &opts.CMConfigParams)
	if err != nil {
		log.Printf("Failed to unmarshal configmap file content, err: [%v]. Continue with hooks run", err)
	}
}

func (opts *Options) parseCommandLineArgs(flagSet *flag.FlagSet) {

	flagSet.BoolVar(&opts.PreInstall, "pre-install", false, "Run pre-install tasks")

	flagSet.BoolVar(&opts.PostInstall, "post-install", false, "Run post-install tasks")

	flagSet.BoolVar(&opts.PreUpgrade, "pre-upgrade", false, "Run pre-upgrade tasks")

	flagSet.BoolVar(&opts.PreCrdUpgrade, "pre-crd-upgrade", false, "Run pre-crd-upgrade tasks")

	flagSet.BoolVar(&opts.PreResourceManagementUpgrade, "pre-resource-management-upgrade", false, "Run pre-resource-management-upgrade tasks")

	flagSet.BoolVar(&opts.PostUpgrade, "post-upgrade", false, "Run post-upgrade tasks")

	flagSet.BoolVar(&opts.PreDelete, "pre-delete", false, "Run pre delete tasks")

	flagSet.BoolVar(&opts.ValidateClusterConfig, "validate-cluster-config", false, "Validate clusterConfig values and exit")

	// Hook configuration arguments
	flagSet.StringVar(&opts.ClusterVersion, "cluster-version", "",
		"Version of the cluster to install or upgrade")

	flagSet.StringVar(&opts.ControlPlaneURL, "control-plane-url", "",
		"URL of the control-plane to connect this cluster to")

	flagSet.StringVar(&opts.ControlPlaneClientID, "control-plane-client-id", "",
		"Client-ID to authenticate to the control-plane with")

	flagSet.Var(&opts.ControlPlaneClientSecret, "control-plane-client-secret",
		"Secret to authenticate to the control-plane with. Accepts file path when starts with @ (e.g. @/path/to/secret)")

	flagSet.Var(&opts.ControlPlaneBootstrapToken, "control-plane-bootstrap-token",
		"Bootstrap token to authenticate to the control-plane with. Accepts file path when starts with @ (e.g. @/path/to/token)")

	flagSet.StringVar(&opts.ClusterUID, "cluster-uid", "",
		"UID of the cluster to install or upgrade")

	flagSet.StringVar(&opts.ClusterURL, "cluster-url", "",
		"URL of the cluster")

	flagSet.BoolVar(&opts.CustomCaEnabled, "custom-ca-enabled", false,
		"Whether to trust a custom CA certificate")

	flagSet.StringVar(&opts.CustomCaSecretName, "custom-ca-secret-name", "",
		"Name of the configmap containing the CA certificate to be trusted")

	flagSet.StringVar(&opts.CustomCaSecretKey, "custom-ca-secret-key", "",
		"Key of the configmap containing the CA certificate to be trusted")

	flagSet.BoolVar(&opts.CustomCaSecretOptional, "custom-ca-secret-optional", false,
		"Whether the configmap containing the CA certificate to be trusted is optional")

	flagSet.IntVar(&opts.Verbosity, "log-level", 0,
		"User specified logging level")

	flagSet.StringVar(&opts.ImageRegistry, "image-registry", "",
		"User specified image registry to use across all runai-operator services")

	flagSet.StringVar(&opts.cMFilePath, "cm-file-path", cmFile,
		"Path to the configmap file")

	flagSet.Var(&opts.AdditionalImagePullSecrets, "additional-image-pull-secrets", "Additional image pull secrets names to use")

	flagSet.BoolVar(&opts.EnableNMC, "nmc-enabled", false, "Enable NMC")

	flagSet.StringVar(&opts.NMCRemoteClusterURL, "nmc-remote-cluster-api-server-url", "", "The URL of the remote cluster API server")

	flagSet.StringVar(&opts.NMCServiceAccountName, "nmc-service-account-name", "", "The name of the service account to use for NMC")

	flagSet.Var(&opts.NMCAllowedServiceAccounts, "nmc-allowed-service-accounts", "specifies a list of service account name patterns that should be allowed to perform privileged operations in NMC installation")

	flagSet.BoolVar(&opts.Development.Headless, "development-headless", false, "Install a headless cluster")

	flagSet.BoolVar(&opts.Development.SkipDependencyChecks, "development-skip-dependency-checks", false, "Skip dependency checks")

	flagSet.BoolVar(&opts.DisableInfrastructureMetricsRemoteWrite, "disable-infrastructure-metrics-remote-write", false, "Disable infrastructure metrics remote write endpoint")

	flagSet.IntVar(&opts.QueueValidationTimeoutMillis, "queue-validation-timeout-millis", defaultQueueTimeoutMinutes*60*1000,
		"Timeout in milliseconds for queue validation")
	flagSet.IntVar(&opts.QueueValidationIntervalMillis, "queue-validation-interval-millis", defaultQueueIntervalSeconds*1000,
		"Interval in milliseconds for queue validation")

	flagSet.IntVar(&opts.OperandDownIntervalMillis, "operand-down-interval-millis", defaultOperandDownIntervalSeconds*1000,
		"Interval in milliseconds for operand down validation")

	flagSet.BoolVar(&opts.ResourceManagementMigrationEnabled, "resource-management-migration-enabled", false,
		"Signals the resource-management project/department CRD migration. When set, the pre-resource-management-upgrade hook migrates legacy departments (scheduling.run.ai/v2alpha2) to kai departments (kai.resources/v1alpha1). Self-gating: no-op when the legacy department CRD is not installed.")
	flagSet.BoolVar(&opts.ResourceManagementMigrationCleanup, "resource-management-migration-cleanup", false,
		"Signals the resource-management project/department CRD migration cleanup. When set, the post-upgrade hook deletes the legacy department CRD (departments.scheduling.run.ai) once migration is complete: it errors if legacy departments still exist or the new kai department CRD is missing. Self-gating: no-op when the legacy department CRD is not installed.")
	flagSet.BoolVar(&opts.ResourceManagementMigrationNodePoolsEnabled, "resource-management-migration-nodepools-enabled", false,
		"Signals the resource-management nodepool CRD migration, independently of the department/project migration. "+
			"When set, the pre-resource-management-upgrade hook migrates legacy nodepools (run.ai/v1alpha1) to kai nodepools (kai.resources/v1alpha1). "+
			"Self-gating: no-op when the legacy nodepool CRD is not installed.")
	flagSet.BoolVar(&opts.ResourceManagementMigrationNodePoolsCleanup, "resource-management-migration-nodepools-cleanup", false,
		"Signals the resource-management nodepool CRD migration cleanup, independently of the department/project cleanup. "+
			"When set, the post-upgrade hook deletes the legacy nodepool CRD (nodepools.run.ai) once migration is complete: "+
			"it errors if legacy nodepools still exist or the new kai nodepool CRD is missing. "+
			"Self-gating: no-op when the legacy nodepool CRD is not installed.")
	flagSet.BoolVar(&opts.EngineToKAIMigrationEnabled, "engine-to-kai-migration-enabled", false,
		"Signals the engine->KAI scheduler migration. When set, the pre-crd-upgrade hook tears down leftover engine SchedulingShards (and their schedulers) if any still exist, so they don't run alongside the KAI scheduler. Self-gating: no-op when no engine SchedulingShards are present.")

	flagSet.BoolVar(&opts.EnableRbacEscalation, "enable-rbac-escalation", true,
		"Whether the relevant operands should have permissions to escalate their RBAC")

	flagSet.BoolVar(&opts.RequiresKubernetesAccess, "requires-kubernetes-access", true,
		"Whether the hook requires access to the Kubernetes cluster")

	flagSet.StringVar(&opts.FipsMode, "fips-mode", "off",
		"Go FIPS 140-3 runtime mode for cluster components (off, on, or only)")

	opts.ZapOptions = zap.Options{
		Development: true,
	}
	opts.ZapOptions.BindFlags(flagSet)
	flagSet.Parse(os.Args[1:])
}
