/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/rs/zerolog/log"
)

var once sync.Once

type applicationControllerConfig struct {
	UseDebugLogLevel     bool
	EnableLeaderElection bool

	K8sClientConfigQPS   int
	K8sClientConfigBurst int

	HelmDriver     string
	HelmCacheHome  string // base cache dir; Helm appends "repository" → HELM_CACHE_HOME
	HelmConfigHome string // base config dir; Helm appends "repositories.yaml" → HELM_CONFIG_HOME

	MetricsPort string

	AllowedServiceAccounts []string

	WebhookPort int
	CertDir     string

	// RunaiApiEndPoint is the control-plane base URL (ENDPOINT env), used to obtain a bearer token
	// for authenticated chart downloads. ClientName/ClientSecret/ClusterUuid are the per-cluster
	// agent credentials, loaded from the mounted whisper secret. All empty when the secret is not
	// mounted, in which case authenticated downloads are disabled and behavior is unchanged.
	RunaiApiEndPoint string
	ClientName       string
	ClientSecret     string
	ClusterUuid      string
}

// clientCredentials mirrors the JSON written into the whisper-application-controller-secret by the
// go-operator vault (keys CLIENT_NAME/CLIENT_SECRET/CLUSTER_ID), mounted at constants.SecretMountPath.
type clientCredentials struct {
	ClientName   string `json:"CLIENT_NAME"`
	ClientSecret string `json:"CLIENT_SECRET"`
	ClusterID    string `json:"CLUSTER_ID"`
}

func loadClientCredentials() clientCredentials {
	data, err := os.ReadFile(constants.SecretMountPath)
	if err != nil {
		log.Warn().Msgf("Control-plane credentials secret not mounted at %s; authenticated chart "+
			"downloads disabled: %v", constants.SecretMountPath, err)
		return clientCredentials{}
	}
	creds, err := parseClientCredentials(data)
	if err != nil {
		log.Warn().Msgf("Failed to parse control-plane credentials at %s; authenticated chart "+
			"downloads disabled: %v", constants.SecretMountPath, err)
		return clientCredentials{}
	}
	return creds
}

func parseClientCredentials(data []byte) (clientCredentials, error) {
	var creds clientCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return clientCredentials{}, err
	}
	return creds, nil
}

var config *applicationControllerConfig

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

func newConfig() *applicationControllerConfig {
	allowedServiceAccounts := getEnvOrDefaultString("ALLOWED_SERVICE_ACCOUNTS",
		"system:serviceaccount:runai:application-controller,system:serviceaccount:runai:runai-agent")

	ans := &applicationControllerConfig{
		K8sClientConfigQPS:     getEnvOrDefaultInt("K8S_CLIENT_CONFIG_QPS", 50),
		K8sClientConfigBurst:   getEnvOrDefaultInt("K8S_CLIENT_CONFIG_BURST", 300),
		UseDebugLogLevel:       getEnvOrDefaultBool("USE_DEBUG_LOG_LEVEL", false),
		MetricsPort:            getEnvOrDefaultString("METRICS_PORT", "9400"),
		EnableLeaderElection:   getEnvOrDefaultBool("LEADER_ELECT", false),
		HelmDriver:             getEnvOrDefaultString("HELM_DRIVER", ""),
		HelmCacheHome:          getEnvOrDefaultString("HELM_CACHE_HOME", "/tmp/helm/cache"),
		HelmConfigHome:         getEnvOrDefaultString("HELM_CONFIG_HOME", "/tmp/helm/config"),
		AllowedServiceAccounts: strings.Split(allowedServiceAccounts, ","),
		WebhookPort:            getEnvOrDefaultInt("WEBHOOK_PORT", 9443),
		CertDir:                getEnvOrDefaultString("CERT_DIR", "/etc/webhook/certs"),
		RunaiApiEndPoint:       getEnvOrDefaultString("ENDPOINT", ""),
	}

	creds := loadClientCredentials()
	ans.ClientName = creds.ClientName
	ans.ClientSecret = creds.ClientSecret
	ans.ClusterUuid = creds.ClusterID

	// Ensure Helm's helmpath.CachePath resolves to a writable location. we have to do it here because
	// settings.RepositoryCache is not visible to Helm.
	if os.Getenv("HELM_CACHE_HOME") == "" {
		_ = os.Setenv("HELM_CACHE_HOME", ans.HelmCacheHome)
	}
	if os.Getenv("HELM_CONFIG_HOME") == "" {
		_ = os.Setenv("HELM_CONFIG_HOME", ans.HelmConfigHome)
	}

	return ans
}

func Config() *applicationControllerConfig {
	once.Do(func() {
		config = newConfig()
	})
	return config
}
