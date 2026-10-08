/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

var once sync.Once

type sharedObjectsControllerConfig struct {
	UseDebugLogLevel bool
	ProjectHandler   ProjectHandlerConfig

	K8sClientConfigQPS   int
	K8sClientConfigBurst int

	BaseControllerDelay time.Duration
	MaxControllerDelay  time.Duration
}

type ProjectHandlerConfig struct {
	CreateServiceMonitors             bool
	ClusterWideSecrets                bool
	ClusterWidePvcs                   bool
	ClusterWideConfigMaps             bool
	ClusterWideResourceClaimTemplates bool
}

var config *sharedObjectsControllerConfig

//nolint:unused
func _getEnvOrDie(environmentVariableName string) string {
	env := os.Getenv(environmentVariableName)
	log.Info().Msgf("Environment variable %s: %s", environmentVariableName, env)

	if env == "" {
		log.Fatal().Msgf("Environment variable %s was not provided, agent terminating...", environmentVariableName)
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

func newConfig() *sharedObjectsControllerConfig {
	baseControllerDelayMilliseconds := getEnvOrDefaultInt("BASE_CONTROLLER_DELAY_MILLISECONDS", 500)
	maxControllerDelaySeconds := getEnvOrDefaultInt("MAX_CONTROLLER_DELAY_SECONDS", 60)

	ans := &sharedObjectsControllerConfig{
		K8sClientConfigQPS:   getEnvOrDefaultInt("K8S_CLIENT_CONFIG_QPS", 50),
		K8sClientConfigBurst: getEnvOrDefaultInt("K8S_CLIENT_CONFIG_BURST", 300),
		UseDebugLogLevel:     getEnvOrDefaultBool("USE_DEBUG_LOG_LEVEL", false),
		BaseControllerDelay:  time.Duration(baseControllerDelayMilliseconds) * time.Millisecond,
		MaxControllerDelay:   time.Duration(maxControllerDelaySeconds) * time.Second,
		ProjectHandler: ProjectHandlerConfig{
			CreateServiceMonitors:             getEnvOrDefaultBool("CREATE_SERVICE_MONITORS", false),
			ClusterWideSecrets:                getEnvOrDefaultBool("CLUSTER_WIDE_SECRETS", false),
			ClusterWidePvcs:                   getEnvOrDefaultBool("CLUSTER_WIDE_PVCS", false),
			ClusterWideConfigMaps:             getEnvOrDefaultBool("CLUSTER_WIDE_CONFIG_MAPS", true),
			ClusterWideResourceClaimTemplates: getEnvOrDefaultBool("CLUSTER_WIDE_RESOURCE_CLAIM_TEMPLATES", false),
		},
	}

	return ans
}

func Config() *sharedObjectsControllerConfig {
	once.Do(func() {
		config = newConfig()
	})
	return config
}
