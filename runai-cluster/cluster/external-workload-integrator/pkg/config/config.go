package config

import (
	"os"
	"strconv"
	"sync"

	"github.com/rs/zerolog/log"
)

const (
	clusterRoleReadNameForRBACDefault = "runai-additional-workload-gvks-read"
	clusterRoleCrudNameForRBACDefault = "runai-additional-workload-gvks-crud"
)

var once sync.Once

type externalWorkloadIntegratorConfig struct {
	UseDebugLogLevel bool

	K8sClientConfigQPS   int
	K8sClientConfigBurst int

	// Enable default implementation for CRDs that are not supported explicitly
	AllowNonSupportedTypes bool

	MetricsPort string `mapstructure:"METRICS_PORT"`
	MetricsPath string `mapstructure:"METRICS_PATH"`

	ServiceName          string `mapstructure:"SERVICE_NAME"`
	EnableLeaderElection bool   `mapstructure:"LEADER_ELECT"`

	// When false (default), TLS is assumed for LoadBalancer ports unless port is 80 or appProtocol is "http".
	// When true, TLS is only detected via known cloud-provider annotations, port 443, or appProtocol "https".
	LbServiceEndpointsTlsByExplicitAnnotation bool `mapstructure:"LB_SERVICE_ENDPOINTS_TLS_BY_EXPLICIT_ANNOTATION"`

	WorkloadTypeRBACHandlerEnabled bool   `mapstructure:"WORKLOAD_TYPE_RBAC_HANDLER_ENABLED"`
	ClusterRoleReadNameForRBAC     string `mapstructure:"CLUSTER_ROLE_READ_NAME_FOR_RBAC"`
	ClusterRoleCrudNameForRBAC     string `mapstructure:"CLUSTER_ROLE_CRUD_NAME_FOR_RBAC"`
}

var config *externalWorkloadIntegratorConfig

//nolint:unused // reserved for future use
func getEnvOrDie(environmentVariableName string) string {
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

func newConfig() *externalWorkloadIntegratorConfig {
	ans := &externalWorkloadIntegratorConfig{
		K8sClientConfigQPS:     getEnvOrDefaultInt("K8S_CLIENT_CONFIG_QPS", 50),
		K8sClientConfigBurst:   getEnvOrDefaultInt("K8S_CLIENT_CONFIG_BURST", 300),
		UseDebugLogLevel:       getEnvOrDefaultBool("USE_DEBUG_LOG_LEVEL", false),
		AllowNonSupportedTypes: getEnvOrDefaultBool("ALLOW_NON_SUPPORTED_TYPES", false),
		MetricsPort:            getEnvOrDefaultString("METRICS_PORT", "9400"),
		MetricsPath:            getEnvOrDefaultString("METRICS_PATH", "/metrics"),
		ServiceName:            getEnvOrDefaultString("SERVICE_NAME", "external-workload-integrator"),
		EnableLeaderElection:   getEnvOrDefaultBool("LEADER_ELECT", false),

		LbServiceEndpointsTlsByExplicitAnnotation: getEnvOrDefaultBool("LB_SERVICE_ENDPOINTS_TLS_BY_EXPLICIT_ANNOTATION", false),

		WorkloadTypeRBACHandlerEnabled: getEnvOrDefaultBool("WORKLOAD_TYPE_RBAC_HANDLER_ENABLED", true),
		ClusterRoleReadNameForRBAC:     getEnvOrDefaultString("CLUSTER_ROLE_READ_NAME_FOR_RBAC", clusterRoleReadNameForRBACDefault),
		ClusterRoleCrudNameForRBAC:     getEnvOrDefaultString("CLUSTER_ROLE_CRUD_NAME_FOR_RBAC", clusterRoleCrudNameForRBACDefault),
	}

	return ans
}

func Config() *externalWorkloadIntegratorConfig {
	once.Do(func() {
		config = newConfig()
	})
	return config
}
