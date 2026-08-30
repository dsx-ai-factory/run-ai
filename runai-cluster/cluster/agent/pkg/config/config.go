package config

import (
	rcpconfig "github.com/run-ai/runai/runai-common-packages/config"
)

type AgentConfig struct {
	ClusterUuid string `mapstructure:"CLUSTER_ID" config:"secret"`

	// Client details to get Application token from RUN:AI backend
	ClientName   string `mapstructure:"CLIENT_NAME" config:"secret"`
	ClientSecret string `mapstructure:"CLIENT_SECRET" config:"secret"`

	// We need it for workload composer
	ClusterVersion string `mapstructure:"CLUSTER_VERSION"`

	// Base endpoint  for requests to run:ai backend
	RunaiApiEndPoint string `mapstructure:"ENDPOINT"`

	// if the agent should sync policies from the backend
	PolicySyncEnabled bool `mapstructure:"POLICY_SYNC_ENABLED"`

	TLSRejectUnauthorized bool `mapstructure:"TLS_REJECT_UNAUTHORIZED"`

	AgentSyncIntervalMillis int `mapstructure:"AGENT_SYNC_INTERVAL_MILLIS"`

	UseDebugLogLevel bool `mapstructure:"USE_DEBUG_LOG_LEVEL"`

	MetricsPort string `mapstructure:"METRICS_PORT"`
	MetricsPath string `mapstructure:"METRICS_PATH"`

	ProjectsLimit    int `mapstructure:"PROJECTS_LIMIT"`
	DepartmentsLimit int `mapstructure:"DEPARTMENTS_LIMIT"`
	AccessRulesLimit int `mapstructure:"ACCESS_RULES_LIMIT"`

	K8sClientConfigQPS   int `mapstructure:"K8S_CLIENT_CONFIG_QPS"`
	K8sClientConfigBurst int `mapstructure:"K8S_CLIENT_CONFIG_BURST"`

	EnableLeaderElection bool `mapstructure:"LEADER_ELECT"`
	EnableWorkloadRunner bool `mapstructure:"ENABLE_WORKLOAD_RUNNER"`

	// Webhook serving configuration for the project/department validating webhook.
	CertDir     string `mapstructure:"CERT_DIR"`
	WebhookPort int    `mapstructure:"WEBHOOK_PORT"`
	// AllowedServiceAccounts is a comma-separated list of service account name
	// patterns that are allowed to perform project/department operations.
	AllowedServiceAccounts string `mapstructure:"ALLOWED_SERVICE_ACCOUNTS"`
}

func (a *AgentConfig) SetDefaults() {
	a.UseDebugLogLevel = false
	a.PolicySyncEnabled = false
	a.TLSRejectUnauthorized = false
	a.AgentSyncIntervalMillis = 10000
	a.MetricsPort = "9400"
	a.MetricsPath = "/metrics"
	a.ProjectsLimit = 200
	a.DepartmentsLimit = 400
	a.AccessRulesLimit = 400
	a.K8sClientConfigQPS = 50
	a.K8sClientConfigBurst = 300
	a.EnableLeaderElection = false
	a.EnableWorkloadRunner = true
	a.CertDir = "/etc/webhook/certs"
	a.WebhookPort = 9443
	a.AllowedServiceAccounts = ""
}

func Config() *AgentConfig {
	return rcpconfig.Get[*AgentConfig]()
}
