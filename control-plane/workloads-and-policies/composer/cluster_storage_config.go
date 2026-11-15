/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	clustersApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	sdkglobal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	corev1 "k8s.io/api/core/v1"
)

// Cluster-sync lowercases the platform flavor (parseClusterInfoPlatformToServerScheme),
// so the reported value is "openshift", not the CR's "OpenShift".
const openshiftPlatformType = "openshift"

type ClusterStorageConfig interface {
	GitSyncImage() string
	GitSyncResources() (corev1.ResourceRequirements, error)
	IsCustomCAGitEnabled() bool
	CustomCAGitSecret() string
	CustomCAGitKey() string

	OAuth2ProxyImage() string
	OAuth2ProxyResources() (corev1.ResourceRequirements, error)
	OAuth2ProxyUserLoaderImage() string
	UsernameLoaderResources() (corev1.ResourceRequirements, error)
	OAuth2ProxyAuthSecretName() string
	OAuth2ProxySessionStoreType() string
	OAuth2ProxyRedisFQDN() string
	OAuth2ProxyOidcIssuerUrl() string
	IsOpenshift() bool
	SubdomainSupportEnabled() bool
	OpenshiftRouteTerminationPolicy() string
	ClusterDomain() string
	ExternalAuthUrlEnabled() bool
	IsCustomCAAuthEnabled() bool
	CustomCAAuthSecret() string
	CustomCAAuthKey() string

	S3FsImage() string
	S3FsMountCheckImage() string
	S3FileSystemResources() (corev1.ResourceRequirements, error)
	FipsGodebug() string
	IsCustomCAS3Enabled() bool
	CustomCAS3Secret() string
	CustomCAS3Key() string

	IngressEnabled() bool
	GatewayAPIEnabled() bool
	GatewayName() string
	GatewayNamespace() string

	// DistGPUNetworkAcceleration reports whether GPU network acceleration (MNNVL)
	// is enabled cluster-wide for the workload-controller (distributed-training).
	DistGPUNetworkAcceleration() bool
	// DinfGPUNetworkAcceleration reports whether GPU network acceleration (MNNVL)
	// is enabled cluster-wide for the inference-workload-controller (distributed-inference).
	DinfGPUNetworkAcceleration() bool
}

type StaticClusterStorageConfig struct {
	gitSyncImage         string
	gitSyncResources     configuredResourceRequirements
	isCustomCAGitEnabled bool
	customCAGitSecret    string
	customCAGitKey       string

	oauth2ProxyImage                string
	oauth2ProxyResources            configuredResourceRequirements
	oauth2ProxyUserLoaderImage      string
	usernameLoaderResources         configuredResourceRequirements
	oauth2ProxyAuthSecretName       string
	oauth2ProxySessionStoreType     string
	oauth2ProxyRedisFQDN            string
	oauth2ProxyOidcIssuerUrl        string
	isOpenshift                     bool
	subdomainSupportEnabled         bool
	openshiftRouteTerminationPolicy string
	clusterDomain                   string
	externalAuthUrlEnabled          bool
	isCustomCAAuthEnabled           bool
	customCAAuthSecret              string
	customCAAuthKey                 string

	s3FsImage             string
	s3FsMountCheckImage   string
	s3FileSystemResources configuredResourceRequirements
	fipsMode              string
	isCustomCAS3Enabled   bool
	customCAS3Secret      string
	customCAS3Key         string

	// nil preserves the default cluster behavior: Ingress/Route enabled.
	ingressEnabled    *bool
	gatewayAPIEnabled bool
	gatewayName       string
	gatewayNamespace  string

	distGPUNetworkAcceleration bool
	dinfGPUNetworkAcceleration bool
}

func (c StaticClusterStorageConfig) GitSyncImage() string { return c.gitSyncImage }
func (c StaticClusterStorageConfig) GitSyncResources() (corev1.ResourceRequirements, error) {
	return c.gitSyncResources.valueOr(corev1.ResourceRequirements{})
}
func (c StaticClusterStorageConfig) IsCustomCAGitEnabled() bool { return c.isCustomCAGitEnabled }
func (c StaticClusterStorageConfig) CustomCAGitSecret() string  { return c.customCAGitSecret }
func (c StaticClusterStorageConfig) CustomCAGitKey() string     { return c.customCAGitKey }

func (c StaticClusterStorageConfig) OAuth2ProxyImage() string { return c.oauth2ProxyImage }
func (c StaticClusterStorageConfig) OAuth2ProxyResources() (corev1.ResourceRequirements, error) {
	return c.oauth2ProxyResources.valueOr(corev1.ResourceRequirements{})
}
func (c StaticClusterStorageConfig) OAuth2ProxyUserLoaderImage() string {
	return c.oauth2ProxyUserLoaderImage
}
func (c StaticClusterStorageConfig) UsernameLoaderResources() (corev1.ResourceRequirements, error) {
	return c.usernameLoaderResources.valueOr(defaultUsernameLoaderResources())
}
func (c StaticClusterStorageConfig) OAuth2ProxyAuthSecretName() string {
	return c.oauth2ProxyAuthSecretName
}
func (c StaticClusterStorageConfig) OAuth2ProxySessionStoreType() string {
	return c.oauth2ProxySessionStoreType
}
func (c StaticClusterStorageConfig) OAuth2ProxyRedisFQDN() string { return c.oauth2ProxyRedisFQDN }
func (c StaticClusterStorageConfig) OAuth2ProxyOidcIssuerUrl() string {
	return c.oauth2ProxyOidcIssuerUrl
}
func (c StaticClusterStorageConfig) IsOpenshift() bool             { return c.isOpenshift }
func (c StaticClusterStorageConfig) SubdomainSupportEnabled() bool { return c.subdomainSupportEnabled }
func (c StaticClusterStorageConfig) OpenshiftRouteTerminationPolicy() string {
	if policy := clustersApi.OpenshiftRouteTerminationPolicy(c.openshiftRouteTerminationPolicy); policy.IsValid() {
		return c.openshiftRouteTerminationPolicy
	}
	return string(clustersApi.OPENSHIFTROUTETERMINATIONPOLICY_EDGE)
}
func (c StaticClusterStorageConfig) ClusterDomain() string        { return c.clusterDomain }
func (c StaticClusterStorageConfig) ExternalAuthUrlEnabled() bool { return c.externalAuthUrlEnabled }
func (c StaticClusterStorageConfig) IsCustomCAAuthEnabled() bool  { return c.isCustomCAAuthEnabled }
func (c StaticClusterStorageConfig) CustomCAAuthSecret() string   { return c.customCAAuthSecret }
func (c StaticClusterStorageConfig) CustomCAAuthKey() string      { return c.customCAAuthKey }

func (c StaticClusterStorageConfig) S3FsImage() string           { return c.s3FsImage }
func (c StaticClusterStorageConfig) S3FsMountCheckImage() string { return c.s3FsMountCheckImage }
func (c StaticClusterStorageConfig) S3FileSystemResources() (corev1.ResourceRequirements, error) {
	return c.s3FileSystemResources.valueOr(corev1.ResourceRequirements{})
}

func (c StaticClusterStorageConfig) FipsGodebug() string {
	// Empty means the cluster reported no FIPS mode (null); leave GODEBUG unset.
	// Non-empty values are constrained to off/on/only by the FipsModeEnum schema,
	// validated server-side on ingest.
	if c.fipsMode == "" {
		return ""
	}
	return "fips140=" + c.fipsMode
}
func (c StaticClusterStorageConfig) IsCustomCAS3Enabled() bool { return c.isCustomCAS3Enabled }
func (c StaticClusterStorageConfig) CustomCAS3Secret() string  { return c.customCAS3Secret }
func (c StaticClusterStorageConfig) CustomCAS3Key() string     { return c.customCAS3Key }

func (c StaticClusterStorageConfig) IngressEnabled() bool {
	return convutils.Deref(c.ingressEnabled, true)
}
func (c StaticClusterStorageConfig) GatewayAPIEnabled() bool  { return c.gatewayAPIEnabled }
func (c StaticClusterStorageConfig) GatewayName() string      { return c.gatewayName }
func (c StaticClusterStorageConfig) GatewayNamespace() string { return c.gatewayNamespace }

func (c StaticClusterStorageConfig) DistGPUNetworkAcceleration() bool {
	return c.distGPUNetworkAcceleration
}
func (c StaticClusterStorageConfig) DinfGPUNetworkAcceleration() bool {
	return c.dinfGPUNetworkAcceleration
}

func (c *StaticClusterStorageConfig) SetOauth2ProxyImage(image string) { c.oauth2ProxyImage = image }
func (c *StaticClusterStorageConfig) SetOauth2ProxyResources(resources corev1.ResourceRequirements) {
	c.oauth2ProxyResources.set(resources)
}
func (c *StaticClusterStorageConfig) SetGitSyncImage(image string) { c.gitSyncImage = image }
func (c *StaticClusterStorageConfig) SetGitSyncResources(resources corev1.ResourceRequirements) {
	c.gitSyncResources.set(resources)
}
func (c *StaticClusterStorageConfig) SetUsernameLoaderResources(resources corev1.ResourceRequirements) {
	c.usernameLoaderResources.set(resources)
}
func (c *StaticClusterStorageConfig) SetS3FileSystemResources(resources corev1.ResourceRequirements) {
	c.s3FileSystemResources.set(resources)
}
func (c *StaticClusterStorageConfig) SetExternalAuthUrlEnabled(enabled bool) {
	c.externalAuthUrlEnabled = enabled
}
func (c *StaticClusterStorageConfig) SetIsOpenshift(isOpenshift bool) { c.isOpenshift = isOpenshift }
func (c *StaticClusterStorageConfig) SetSubdomainSupportEnabled(enabled bool) {
	c.subdomainSupportEnabled = enabled
}
func (c *StaticClusterStorageConfig) SetOauth2ProxyOidcIssuerUrl(url string) {
	c.oauth2ProxyOidcIssuerUrl = url
}
func (c *StaticClusterStorageConfig) SetOauth2ProxyUserLoaderImage(image string) {
	c.oauth2ProxyUserLoaderImage = image
}
func (c *StaticClusterStorageConfig) SetIsCustomCAAuthEnabled(enabled bool) {
	c.isCustomCAAuthEnabled = enabled
}
func (c *StaticClusterStorageConfig) SetCustomCAAuthSecret(secret string) {
	c.customCAAuthSecret = secret
}
func (c *StaticClusterStorageConfig) SetCustomCAAuthKey(key string) { c.customCAAuthKey = key }
func (c *StaticClusterStorageConfig) SetOauth2ProxyAuthSecretName(secretName string) {
	c.oauth2ProxyAuthSecretName = secretName
}
func (c *StaticClusterStorageConfig) SetOauth2ProxySessionStoreType(storeType string) {
	c.oauth2ProxySessionStoreType = storeType
}
func (c *StaticClusterStorageConfig) SetOauth2ProxyRedisFQDN(fqdn string) {
	c.oauth2ProxyRedisFQDN = fqdn
}
func (c *StaticClusterStorageConfig) SetS3FsImage(image string) { c.s3FsImage = image }
func (c *StaticClusterStorageConfig) SetS3FsMountCheckImage(image string) {
	c.s3FsMountCheckImage = image
}
func (c *StaticClusterStorageConfig) SetFipsMode(fipsMode string) { c.fipsMode = fipsMode }
func (c *StaticClusterStorageConfig) SetIsCustomCAS3Enabled(enabled bool) {
	c.isCustomCAS3Enabled = enabled
}
func (c *StaticClusterStorageConfig) SetCustomCAS3Secret(secret string) { c.customCAS3Secret = secret }
func (c *StaticClusterStorageConfig) SetCustomCAS3Key(key string)       { c.customCAS3Key = key }
func (c *StaticClusterStorageConfig) SetIngressEnabled(enabled bool)    { c.ingressEnabled = &enabled }
func (c *StaticClusterStorageConfig) SetGatewayAPIEnabled(enabled bool) {
	c.gatewayAPIEnabled = enabled
}
func (c *StaticClusterStorageConfig) SetGatewayName(name string) { c.gatewayName = name }
func (c *StaticClusterStorageConfig) SetGatewayNamespace(namespace string) {
	c.gatewayNamespace = namespace
}
func (c *StaticClusterStorageConfig) SetClusterDomain(domain string) { c.clusterDomain = domain }
func (c *StaticClusterStorageConfig) SetCustomCAGitSecret(secret string) {
	c.customCAGitSecret = secret
}
func (c *StaticClusterStorageConfig) SetCustomCAGitKey(key string) { c.customCAGitKey = key }
func (c *StaticClusterStorageConfig) SetIsCustomCAGitEnabled(enabled bool) {
	c.isCustomCAGitEnabled = enabled
}

func (c *StaticClusterStorageConfig) SetOpenshiftRouteTerminationPolicy(policy string) {
	c.openshiftRouteTerminationPolicy = policy
}

func (c *StaticClusterStorageConfig) SetDistGPUNetworkAcceleration(enabled bool) {
	c.distGPUNetworkAcceleration = enabled
}

func (c *StaticClusterStorageConfig) SetDinfGPUNetworkAcceleration(enabled bool) {
	c.dinfGPUNetworkAcceleration = enabled
}

func DefaultClusterStorageConfig() ClusterStorageConfig {
	return StaticClusterStorageConfig{
		customCAGitSecret: "runai-ca-cert",
		customCAGitKey:    "runai-ca.pem",

		oauth2ProxyOidcIssuerUrl: "/auth/realms/runai",
		externalAuthUrlEnabled:   true,
		customCAAuthSecret:       "runai-ca-cert",
		customCAAuthKey:          "runai-ca.pem",

		customCAS3Secret: "runai-ca-cert",
		customCAS3Key:    "runai-ca.pem",
	}
}

type ClusterStorageConfigOption func(*StaticClusterStorageConfig)

func NewClusterStorageConfigFromCluster(cluster *clustersApi.DisplayedCluster, opts ...ClusterStorageConfigOption) ClusterStorageConfig {
	defaults := DefaultClusterStorageConfig().(StaticClusterStorageConfig)
	config := reportedConfig(cluster)

	cfg := defaults
	cfg.gitSyncImage = ""
	cfg.s3FsImage = ""
	cfg.oauth2ProxyImage = ""

	if cluster != nil {
		cfg.clusterDomain = cluster.GetDomain()
		cfg.isOpenshift = reportedPlatformType(cluster) == openshiftPlatformType
	}

	var gitCA, s3CA, authCA *clustersApi.ClusterReportedCustomCA
	if config != nil {
		cfg.gitSyncImage = config.GetGitSyncImage()
		cfg.s3FsImage = config.GetS3FileSystemImage()
		cfg.s3FsMountCheckImage = config.GetS3FileSystemMountCheckImage()
		cfg.fipsMode = string(config.GetFipsMode())
		cfg.oauth2ProxyImage = config.GetRunaiOauth2ProxyImage()
		cfg.gitSyncResources = reportedResourceRequirements(config.GetGitSyncResourcesOk())
		cfg.s3FileSystemResources = reportedResourceRequirements(config.GetS3FileSystemResourcesOk())
		cfg.oauth2ProxyResources = reportedResourceRequirements(config.GetRunaiOauth2ProxyResourcesOk())
		cfg.oauth2ProxyUserLoaderImage = config.GetRunaiOauth2ProxyUsernameLoaderImage()
		cfg.usernameLoaderResources = reportedResourceRequirements(config.GetUsernameLoaderResourcesOk())
		cfg.oauth2ProxyAuthSecretName = config.GetRunaiOauth2ProxyWorkloadAuthSecretName()
		cfg.oauth2ProxySessionStoreType = config.GetRunaiOauth2ProxySessionStoreType()
		cfg.oauth2ProxyRedisFQDN = config.GetRunaiOauth2ProxyRedisFQDN()
		cfg.oauth2ProxyOidcIssuerUrl = convutils.ValueOrDefault(config.GetRunaiOauth2ProxyOidcIssuerUrl(), defaults.oauth2ProxyOidcIssuerUrl)
		cfg.subdomainSupportEnabled = convutils.Deref(config.SubdomainEnabled.Get(), defaults.subdomainSupportEnabled)
		cfg.externalAuthUrlEnabled = convutils.Deref(config.ExternalAuthUrlEnabled.Get(), defaults.externalAuthUrlEnabled)
		cfg.ingressEnabled = config.IngressEnabled.Get()
		cfg.gatewayAPIEnabled = convutils.Deref(config.GatewayAPIEnabled.Get(), config.GetGatewayName() != "")
		cfg.openshiftRouteTerminationPolicy = string(convutils.ValueOrDefault(config.GetOpenshiftRouteTerminationPolicy(), clustersApi.OPENSHIFTROUTETERMINATIONPOLICY_EDGE))
		cfg.gatewayName = config.GetGatewayName()
		cfg.gatewayNamespace = config.GetGatewayNamespace()
		cfg.distGPUNetworkAcceleration = convutils.Deref(config.DistGPUNetworkAcceleration.Get(), false)
		cfg.dinfGPUNetworkAcceleration = convutils.Deref(config.DinfGPUNetworkAcceleration.Get(), false)
		gitCA = config.CustomCAGit.Get()
		s3CA = config.CustomCAS3.Get()
		authCA = config.CustomCAOauth2Proxy.Get()
	}

	cfg.isCustomCAGitEnabled, cfg.customCAGitSecret, cfg.customCAGitKey = customCAOrDefault(gitCA,
		defaults.isCustomCAGitEnabled, defaults.customCAGitSecret, defaults.customCAGitKey)
	cfg.isCustomCAS3Enabled, cfg.customCAS3Secret, cfg.customCAS3Key = customCAOrDefault(s3CA,
		defaults.isCustomCAS3Enabled, defaults.customCAS3Secret, defaults.customCAS3Key)
	cfg.isCustomCAAuthEnabled, cfg.customCAAuthSecret, cfg.customCAAuthKey = customCAOrDefault(authCA,
		defaults.isCustomCAAuthEnabled, defaults.customCAAuthSecret, defaults.customCAAuthKey)

	for _, opt := range opts {
		opt(&cfg)
	}

	return cfg
}

type configuredResourceRequirements struct {
	value      corev1.ResourceRequirements
	err        error
	configured bool
}

func (r configuredResourceRequirements) valueOr(fallback corev1.ResourceRequirements) (corev1.ResourceRequirements, error) {
	if r.err != nil {
		return corev1.ResourceRequirements{}, r.err
	}
	if !r.configured {
		return fallback, nil
	}
	return r.value, nil
}

func (r *configuredResourceRequirements) set(value corev1.ResourceRequirements) {
	r.value = value
	r.err = nil
	r.configured = true
}

func reportedResourceRequirements(reported *clustersApi.ClusterReportedResourceRequirements, configured bool) configuredResourceRequirements {
	if !configured || reported == nil {
		return configuredResourceRequirements{}
	}
	value, err := parseReportedResourceRequirements(reported)
	return configuredResourceRequirements{value: value, err: err, configured: true}
}

func parseReportedResourceRequirements(reported *clustersApi.ClusterReportedResourceRequirements) (corev1.ResourceRequirements, error) {
	return sdkglobal.ResourceRequirementsFromStringMaps(reported.GetLimits(), reported.GetRequests())
}

func defaultUsernameLoaderResources() corev1.ResourceRequirements {
	return sdkglobal.DefaultUsernameLoaderInitContainerResources()
}

func reportedConfig(cluster *clustersApi.DisplayedCluster) *clustersApi.ClusterReportedConfig {
	if cluster == nil {
		return nil
	}
	status := cluster.Status.Get()
	if status == nil {
		return nil
	}
	return status.Config.Get()
}

func reportedPlatformType(cluster *clustersApi.DisplayedCluster) string {
	if cluster == nil {
		return ""
	}
	platform := cluster.Status.Get().GetPlatform()
	return platform.GetType()
}

func customCAOrDefault(ca *clustersApi.ClusterReportedCustomCA, defEnabled bool, defSecret, defKey string) (bool, string, string) {
	if ca == nil {
		return defEnabled, defSecret, defKey
	}
	return convutils.Deref(ca.Enabled.Get(), defEnabled),
		convutils.Deref(ca.SecretName.Get(), defSecret),
		convutils.Deref(ca.SecretKey.Get(), defKey)
}

func ClusterWithConfig(config clustersApi.ClusterReportedConfig) *clustersApi.DisplayedCluster {
	status := clustersApi.ClusterDisplayedStatus{}
	status.SetConfig(config)

	cluster := clustersApi.DisplayedCluster{}
	cluster.Status.Set(&status)
	return &cluster
}
