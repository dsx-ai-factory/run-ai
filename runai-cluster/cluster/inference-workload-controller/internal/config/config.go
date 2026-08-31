/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"flag"
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/common/utils/options"
	"os"
	"strconv"
	"sync"
)

var once sync.Once

type InferenceWorkloadControllerConfig struct {
	UseDebugLogLevel                  bool
	DisableMetrics                    bool
	EnableWorkloadOwnershipProtection bool
	EnableMnnvlSupport                bool
	TopologyEnabled                   bool
	ElasticEnabled                    bool
	MaxConcurrentReconciles           int
	SkipMigrations                    bool
	MigrationMaxWorkers               int
	K8sClientConfigQPS                int
	K8sClientConfigBurst              int
	MetricExporterBindAddr            string
	WebhookPort                       int
	CertDir                           string
	HealthProbeBindAddress            string
	EnableLeaderElection              bool
	AllowedServiceAccounts            options.ArrayFlags
}

var config *InferenceWorkloadControllerConfig

func getEnvOrDefaultBool(environmentVariableName string, defaultValue bool) bool {
	env := os.Getenv(environmentVariableName)
	ans, err := strconv.ParseBool(env)
	if env == "" {
		log.Warn().Msgf("Environment variable %s was not provided, using default value <%t> instead",
			environmentVariableName, defaultValue)
		ans = defaultValue
	} else if err != nil {
		log.Warn().Msgf("Environment variable %s has incorrect format, using default value <%t> instead, %v",
			environmentVariableName, defaultValue, err)
		ans = defaultValue
	}
	return ans
}

func getEnvOrDefaultInt(environmentVariableName string, defaultValue int) int {
	env := os.Getenv(environmentVariableName)
	ans, err := strconv.Atoi(env)
	if env == "" {
		log.Warn().Msgf("Environment variable %s was not provided, using default value <%d> instead",
			environmentVariableName, defaultValue)
		ans = defaultValue
	} else if err != nil {
		log.Warn().Msgf("Environment variable %s has incorrect format, using default value <%d> instead, %v",
			environmentVariableName, defaultValue, err)
		ans = defaultValue
	}
	return ans
}

func newConfig() *InferenceWorkloadControllerConfig {

	ans := &InferenceWorkloadControllerConfig{
		K8sClientConfigQPS:   getEnvOrDefaultInt("K8S_CLIENT_CONFIG_QPS", 50),
		K8sClientConfigBurst: getEnvOrDefaultInt("K8S_CLIENT_CONFIG_BURST", 300),
		UseDebugLogLevel:     getEnvOrDefaultBool("USE_DEBUG_LOG_LEVEL", false),
	}
	flag.BoolVar(&ans.DisableMetrics, "disable-metrics", false, "Disable metric exporter")
	flag.BoolVar(&ans.EnableWorkloadOwnershipProtection, "enable-workloads-ownership-protection", false, "Enable workload ownership protection")
	flag.BoolVar(&ans.EnableMnnvlSupport, "enable-mnnvl-support", false, "Enable MNNVL support")
	flag.BoolVar(&ans.TopologyEnabled, "enable-topology-support", false, "Enable scheduling topology support")
	flag.BoolVar(&ans.ElasticEnabled, "enable-elastic-compute-domain", false, "Enable support for the scheduling topology feature")
	flag.StringVar(&ans.MetricExporterBindAddr, "metrics-bind-address", ":8383", "Metric bind address for metric exporter")
	flag.StringVar(&ans.HealthProbeBindAddress, "health-probe-bind-address", ":8080", "Health probe bind address")
	flag.StringVar(&ans.CertDir, "cert-dir", "/etc/webhook/certs", "Directory that contains the server key and certificate for the webhooks")
	flag.IntVar(&ans.WebhookPort, "webhook-port", 9443, "Webhook port")
	flag.IntVar(&ans.MaxConcurrentReconciles, "max-concurrent-reconciles", 10, "Max concurrent reconciles")
	flag.BoolVar(&ans.SkipMigrations, "skip-migrations", false, "Skip the migration flow")
	flag.IntVar(&ans.MigrationMaxWorkers, "migration-max-workers", 50, "Max concurrent migration workers")
	flag.BoolVar(&ans.EnableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Var(&ans.AllowedServiceAccounts, "allowed-service-accounts", "Comma separated list of allowed service accounts")
	flag.Parse()

	return ans
}

func Config() *InferenceWorkloadControllerConfig {
	once.Do(func() {
		config = newConfig()
	})
	return config
}
