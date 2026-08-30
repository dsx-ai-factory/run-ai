package config

import (
	"errors"
	"flag"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

var once sync.Once

type ClusterAPI struct {
	UseDebugLogLevel      bool
	K8sClientConfigQPS    int
	K8sClientConfigBurst  int
	Port                  int
	WriteTimeout          time.Duration
	ReadTimeout           time.Duration
	IdleTimeout           time.Duration
	GraceShutdownTimeout  time.Duration
	CacheSyncTimeout      time.Duration
	IssuerURL             string
	KeepaliveIdleTimeout  time.Duration
	KeepalivePingTimeout  time.Duration
	CacheExpirationPeriod time.Duration
	WorkloadAuthCacheTTL  time.Duration
	ControlPlaneURL       string
	EnableMetrics         bool
	MetricsPort           int
}

func (c *ClusterAPI) Validate() error {
	if c == nil {
		return errors.New("ClusterAPI config is nil")
	}

	if c.IssuerURL == "" {
		return errors.New("issuer URL is required")
	}

	return nil
}

var config *ClusterAPI

func getEnvOrDefaultBool(environmentVariableName string, defaultValue bool) bool {
	env := os.Getenv(environmentVariableName)
	if env == "" {
		log.Warn().Msgf("Environment variable %s was not provided, using default value <%t> instead",
			environmentVariableName, defaultValue)

		return defaultValue
	}

	ans, err := strconv.ParseBool(env)
	if err != nil {
		log.Warn().Msgf("Environment variable %s has incorrect format, using default value <%t> instead, %v",
			environmentVariableName, defaultValue, err)

		return defaultValue
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

func newConfig() *ClusterAPI {
	ans := &ClusterAPI{}

	flag.IntVar(&ans.Port, "port", getEnvOrDefaultInt(envPort, defaultPort),
		"Port to listen on")
	flag.DurationVar(&ans.WriteTimeout, "write-timeout",
		time.Duration(getEnvOrDefaultInt(envWriteTimeout, defaultWriteTimeout))*time.Second,
		"Write timeout")
	flag.DurationVar(&ans.ReadTimeout, "read-timeout",
		time.Duration(getEnvOrDefaultInt(envReadTimeout, defaultReadTimeout))*time.Second,
		"Read timeout")
	flag.DurationVar(&ans.IdleTimeout, "idle-timeout",
		time.Duration(getEnvOrDefaultInt(envIdleTimeout, defaultIdleTimeout))*time.Second,
		"Idle timeout")
	flag.DurationVar(&ans.GraceShutdownTimeout, "graceful-timeout",
		time.Second*defaultGraceShutdownTimeout, "Graceful shutdown timeout")
	flag.IntVar(&ans.K8sClientConfigQPS,
		"k8s-client-config-qps",
		getEnvOrDefaultInt(envK8sClientConfigQPS,
			defaultK8sClientConfigQPS), "K8s client config QPS")
	flag.IntVar(&ans.K8sClientConfigBurst,
		"k8s-client-config-burst",
		getEnvOrDefaultInt(envK8sClientConfigBurst,
			defaultK8sClientConfigBurst), "K8s client config burst")
	flag.BoolVar(&ans.UseDebugLogLevel, "use-debug-log-level",
		getEnvOrDefaultBool(envUseDebugLogLevel,
			defaultUseDebugLogLevel), "Use debug log level")
	flag.DurationVar(&ans.CacheSyncTimeout, "cache-sync-timeout", defaultCacheSyncTimeout, "Cache sync timeout")
	flag.StringVar(&ans.IssuerURL, issuerURLFlag, getEnvOrDefaultString(envIssuerURL, ""),
		"Issuer URL")

	flag.DurationVar(&ans.KeepaliveIdleTimeout, "keepalive-idle-time-sec",
		time.Duration(getEnvOrDefaultInt(envKeepaliveIdleTimeoutSec, defaultKeepaliveIdleTimeout))*time.Second,
		"Keepalive idle time in sec")
	flag.DurationVar(&ans.KeepalivePingTimeout, "keepalive-ping-time-sec",
		time.Duration(getEnvOrDefaultInt(envKeepalivePingTimeoutSec, defaultKeepalivePingTimeout))*time.Second,
		"Keepalive ping time in sec")
	cacheExp := getEnvOrDefaultInt(envCacheExpirationPeriod, defaultEnvCacheExpirationPeriod)
	flag.DurationVar(&ans.CacheExpirationPeriod, cacheExpirationPeriodFlag,
		time.Duration(cacheExp)*time.Second, "Cache invalidation interval")
	workloadAuthTTL := getEnvOrDefaultInt(envWorkloadAuthCacheTTL, defaultWorkloadAuthCacheTTLSec)
	flag.DurationVar(&ans.WorkloadAuthCacheTTL, workloadAuthCacheTTLFlag,
		time.Duration(workloadAuthTTL)*time.Second, "Workload-authorization cache TTL for k8s-object endpoints")
	flag.StringVar(&ans.ControlPlaneURL, cpURLFlag, getEnvOrDefaultString(envCPURL, ""),
		"Authorization service URL")

	flag.BoolVar(&ans.EnableMetrics, "enable-metrics",
		getEnvOrDefaultBool(envEnableMetrics, defaultUseEnableMetrics), "Enable prometheus metrics")
	flag.IntVar(&ans.MetricsPort, "metrics-port",
		getEnvOrDefaultInt(envMetricsPort, defaultMetricsPort), "Set prometheus metrics exposed port")
	flag.Parse()

	return ans
}

func Config() *ClusterAPI {
	once.Do(func() {
		config = newConfig()
	})

	return config
}
