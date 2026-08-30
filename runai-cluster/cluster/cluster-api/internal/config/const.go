package config

import "time"

const (
	defaultPort                     = 4180
	defaultWriteTimeout             = 15
	defaultReadTimeout              = 15
	defaultIdleTimeout              = 60
	defaultGraceShutdownTimeout     = 30
	defaultK8sClientConfigQPS       = 50
	defaultK8sClientConfigBurst     = 300
	defaultUseDebugLogLevel         = false
	defaultKeepaliveIdleTimeout     = 60
	defaultCacheSyncTimeout         = 3 * time.Minute
	defaultKeepalivePingTimeout     = 5
	defaultUseEnableMetrics         = true
	defaultMetricsPort              = 9090
	defaultEnvCacheExpirationPeriod = 180
	defaultWorkloadAuthCacheTTLSec  = 60

	envPort                    = "PORT"
	envWriteTimeout            = "WRITE_TIMEOUT"
	envReadTimeout             = "READ_TIMEOUT"
	envIdleTimeout             = "IDLE_TIMEOUT"
	envK8sClientConfigQPS      = "K8S_CLIENT_CONFIG_QPS"
	envK8sClientConfigBurst    = "K8S_CLIENT_CONFIG_BURST"
	envUseDebugLogLevel        = "USE_DEBUG_LOG_LEVEL"
	envIssuerURL               = "ISSUER_URL"
	envCPURL                   = "CP_URL"
	envKeepaliveIdleTimeoutSec = "KEEPALIVE_IDLE_TIME_SEC"
	envKeepalivePingTimeoutSec = "KEEPALIVE_PING_TIME_SEC"
	envCacheExpirationPeriod   = "CACHE_EXPIRATION_PERIOD_SEC"
	envWorkloadAuthCacheTTL    = "WORKLOAD_AUTH_CACHE_TTL_SEC"
	envEnableMetrics           = "ENABLE_METRICS"
	envMetricsPort             = "METRICS_PORT"

	issuerURLFlag             = "issuer-url"
	cpURLFlag                 = "control-plane-url"
	cacheExpirationPeriodFlag = "cache-expiration-period"
	workloadAuthCacheTTLFlag  = "workload-auth-cache-ttl"
)
