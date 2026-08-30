package common

const (
	GitSyncImageRegistry = "registry.k8s.io"
	GitSyncImageName     = "git-sync/git-sync"
	GitSyncImageTag      = "v4.7.0"
	GitSyncImageRef      = GitSyncImageRegistry + "/" + GitSyncImageName + ":" + GitSyncImageTag

	WorkloadRunaiOauth2ProxyImageName               = "oauth2-proxy"
	WorkloadRunaiOauth2ProxyRedisImageTag           = "6.3.4"
	WorkloadRunaiOauth2ProxyRedisImageName          = "keydb"
	WorkloadRunaiOauth2ProxyWorkloadAuthSecretName  = "workloads-auth"
	WorkloadRunaiOauth2ProxyDefaultSessionStoreType = "redis"

	WorkloadRunaiOauth2ProxyUsernameLoaderImageRegistry = "registry.access.redhat.com"
	WorkloadRunaiOauth2ProxyUsernameLoaderImageName     = "ubi9/ubi-minimal"
	WorkloadRunaiOauth2ProxyUsernameLoaderImageTag      = "9.8"
	WorkloadRunaiOauth2ProxyUsernameLoaderImageRef      = WorkloadRunaiOauth2ProxyUsernameLoaderImageRegistry + "/" + WorkloadRunaiOauth2ProxyUsernameLoaderImageName + ":" + WorkloadRunaiOauth2ProxyUsernameLoaderImageTag

	WorkloadRunaiLlmGatewayImageName          = "litellm"
	WorkloadRunaiLlmGatewayResourcesCpuCore   = "100m"
	WorkloadRunaiLlmGatewayResourcesCpuMemory = "100Mi"
)
