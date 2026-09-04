/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"os"
	"strconv"
	"sync"

	"github.com/rs/zerolog/log"
)

const (
	defaultListenPort = "8080"
)

var once sync.Once

type workloadExporterConfig struct {
	ClusterUuid string

	UseDebugLogLevel bool

	K8sClientConfigQPS   int
	K8sClientConfigBurst int

	MetricsListenPort    string
	EnableLeaderElection bool
}

var config *workloadExporterConfig

func getEnvOrDie(environmentVariableName string) string {
	env := os.Getenv(environmentVariableName)
	log.Info().Msgf("Environment variable %s: %s", environmentVariableName, env)

	if env == "" {
		log.Fatal().Msgf("Environment variable %s was not provided, workload-exporter terminating...", environmentVariableName)
	}
	return env
}

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

func getEnvOrDefaultString(environmentVariableName string, defaultValue string) string {
	env := os.Getenv(environmentVariableName)
	if env == "" {
		log.Warn().Msgf("Environment variable %s was not provided, using default value instead (%s)",
			environmentVariableName, defaultValue)
		return defaultValue
	}

	return env
}

func newConfig() *workloadExporterConfig {
	ans := &workloadExporterConfig{
		ClusterUuid:          getEnvOrDie("CLUSTER_ID"),
		K8sClientConfigQPS:   getEnvOrDefaultInt("K8S_CLIENT_CONFIG_QPS", 50),
		K8sClientConfigBurst: getEnvOrDefaultInt("K8S_CLIENT_CONFIG_BURST", 300),
		UseDebugLogLevel:     getEnvOrDefaultBool("USE_DEBUG_LOG_LEVEL", false),
		MetricsListenPort:    getEnvOrDefaultString("METRICS_LISTEN_PORT", defaultListenPort),
		EnableLeaderElection: getEnvOrDefaultBool("LEADER_ELECT", false),
	}

	return ans
}

func Config() *workloadExporterConfig {
	once.Do(func() {
		config = newConfig()
	})
	return config
}
