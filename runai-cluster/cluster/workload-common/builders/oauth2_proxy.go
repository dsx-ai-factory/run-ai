/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	workloadconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	featuregate "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/feature-gate"
	commonutils "github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	"knative.dev/pkg/apis"
)

const (
	ExternalUrlAuthContainerName      = "runai-oauth2-proxy"
	ExternalUrlAuthContainerImageEnv  = "RUNAI_OAUTH2_PROXY_IMAGE"
	ExternalFQDNClusterProxyRedisEnv  = "RUNAI_OAUTH2_PROXY_REDIS_FQDN"
	WorkloadAuthSecretNameEnv         = "RUNAI_OAUTH2_PROXY_WORKLOAD_AUTH_SECRET_NAME"
	WorkloadAuthSessionStoreTypeEnv   = "RUNAI_OAUTH2_PROXY_SESSION_STORE_TYPE"
	WorkloadAuthRedisSessionStoreType = "redis"
)

var (
	// a valid group name starts and ends with alphanumeric characters and can contain alphanumeric characters, hyphens,
	// underscores, and periods. the group name must be between 1 and 64 characters long
	groupPattern = regexp.MustCompile(`^[a-zA-Z0-9][-_a-zA-Z0-9.]{0,62}[a-zA-Z0-9]$`)
)

// OAuth2ProxyConfig holds the configuration for building an OAuth2 proxy container
// This is used by both sidecar (inference) and standalone deployment (networking plugin)
type OAuth2ProxyConfig struct {
	// ContainerName is the name for the container
	ContainerName string
	// UpstreamURL is the URL to proxy to (e.g., "http://localhost:8080" or "http://my-svc:8080")
	UpstreamURL string
	// ExternalScheme is the scheme for external URL (http or https), defaults to https
	ExternalScheme string
	// ExternalHost is the external hostname (e.g., "my-app.example.com")
	ExternalHost string
	// ExternalPath is the external path (e.g., "/" or "/app")
	ExternalPath string
	// Port is the port the OAuth2 proxy listens on
	Port int32
	// CookiePrefix is the prefix for the OAuth2 cookie name
	CookiePrefix string
	// AuthorizedUsers is a comma-separated list of authorized email addresses (optional)
	AuthorizedUsers *string
	// AuthorizedGroups is a comma-separated list of authorized groups (optional)
	AuthorizedGroups *string
	// Nick is used for generating safe file names for authorized users file
	Nick string
	// OmitPortDeclaration controls whether to omit container port declaration from pod spec
	// When true, prevents "Only a single port is allowed across all containers" error in Knative
	OmitPortDeclaration bool
	// ForceJsonResponse configures JSON-formatted responses instead of HTML
	// When true, enables JSON error responses and marks this as the serving port
	ForceJsonResponse bool
	// EnableSidecarRestartPolicy enables the sidecar restart policy (for K8s 1.28+ native sidecars)
	EnableSidecarRestartPolicy bool
	// ProxyPrefix overrides the --proxy-prefix flag for oauth2-proxy.
	// Use this when the ingress strips the path prefix before forwarding to the proxy (e.g. nginx rewrite-target /$2).
	// When empty, the prefix is derived from ExternalPath as ExternalPath+/oauth2.
	ProxyPrefix string
	// ReverseProxy enables --reverse-proxy=true, which tells oauth2-proxy to trust X-Forwarded-Uri headers.
	// Required for standalone deployments where nginx rewrites the path before forwarding.
	// Not needed for sidecar deployments where the proxy receives the original path directly.
	ReverseProxy bool
	// UpstreamTimeoutSeconds sets --upstream-timeout for oauth2-proxy.
	// When non-zero, overrides the default 30s upstream timeout.
	// Should match the workload's requestTimeoutSeconds for inference workloads.
	UpstreamTimeoutSeconds int
}

// BuildOAuth2ProxyContainer creates an OAuth2 proxy container from the given configuration
// This is the common builder used by both sidecar and standalone deployment patterns
func BuildOAuth2ProxyContainer(cfg OAuth2ProxyConfig) (*corev1.Container, error) {
	redirectUrlSuffix, proxyPrefixSuffix := GetOAuth2PathSuffixes(cfg.ExternalPath)

	scheme := cfg.ExternalScheme
	if scheme == "" {
		scheme = "https"
	}

	externalURL := fmt.Sprintf("%s://%s%s", scheme, cfg.ExternalHost, cfg.ExternalPath)
	redirectURL := externalURL + redirectUrlSuffix

	proxyPrefix := cfg.ExternalPath + proxyPrefixSuffix
	if cfg.ProxyPrefix != "" {
		proxyPrefix = cfg.ProxyPrefix
	}
	whitelistDomain := cfg.ExternalHost

	// Handle OpenShift subdomain trimming
	if workloadconfig.IsOpenshift() && workloadconfig.SubdomainSupportEnabled() {
		redirectURL = strings.Replace(redirectURL, workloadconfig.RunaiTrimStr, "", 1)
		whitelistDomain = strings.Replace(whitelistDomain, workloadconfig.RunaiTrimStr, "", 1)
	}
	// Scope the session cookie to the exposed host.
	cookieDomain := (&url.URL{Host: whitelistDomain}).Hostname()

	oidcIssuerUrl := getOidcIssuerUrl()

	args := []string{
		"--provider=keycloak-oidc",
		fmt.Sprintf("--upstream=%s", cfg.UpstreamURL),
		fmt.Sprintf("--redirect-url=%s", redirectURL),
		fmt.Sprintf("--oidc-issuer-url=%s", oidcIssuerUrl),
		"--scope=openid",
		fmt.Sprintf("--http-address=0.0.0.0:%d", cfg.Port),
		fmt.Sprintf("--proxy-prefix=%s", proxyPrefix),
	}
	if cfg.ReverseProxy {
		args = append(args, "--reverse-proxy=true")
	}
	args = append(args,
		"--skip-provider-button=true",
		fmt.Sprintf("--whitelist-domain=%s", whitelistDomain),
		fmt.Sprintf("--cookie-domain=%s", cookieDomain),
		"--cookie-httponly=true",
		"--cookie-secure=true",
		"--code-challenge-method=S256",
		"--cookie-refresh=10m",
		"--cookie-samesite=lax",
		"--skip-jwt-bearer-tokens=true",
		fmt.Sprintf("--extra-jwt-issuers=%s=runai", oidcIssuerUrl),
		"--custom-templates-dir=/usr/share/oauth2-proxy",
		"--pass-authorization-header=false",
		fmt.Sprintf("--cookie-name=%s", cfg.CookiePrefix),
	)

	// Redis session store support
	sessionStoreType := os.Getenv(WorkloadAuthSessionStoreTypeEnv)
	if sessionStoreType == WorkloadAuthRedisSessionStoreType {
		args = append(args, "--session-store-type=redis")
		args = append(args, fmt.Sprintf("--redis-connection-url=redis://%s", os.Getenv(ExternalFQDNClusterProxyRedisEnv)))
	}

	if cfg.UpstreamTimeoutSeconds > 0 {
		args = append(args, fmt.Sprintf("--upstream-timeout=%ds", cfg.UpstreamTimeoutSeconds))
	}

	if cfg.ForceJsonResponse {
		args = append(args, "--force-json-errors=true")
	}

	// Handle authorized users
	if cfg.AuthorizedUsers != nil {
		args = append(args, fmt.Sprintf("--authenticated-emails-file=%s/%s",
			constants.ExternalUrlAuthVolumePath, GetSafeFileNameFromNick(cfg.Nick)))
	}

	// Handle authorized groups — each group requires its own --allowed-group arg
	if cfg.AuthorizedGroups != nil {
		sanitizedGroups, err := SanitizeAuthorizedGroups(*cfg.AuthorizedGroups)
		if err != nil {
			return nil, err
		}
		for _, group := range strings.Split(sanitizedGroups, ",") {
			args = append(args, "--allowed-group="+group)
		}
	}

	// Allow all email domains if no specific users or groups are allowed
	if cfg.AuthorizedUsers == nil || cfg.AuthorizedGroups != nil {
		args = append(args, "--email-domain=*")
	}

	// Volume mounts
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      constants.ExternalUrlAuthVolumeName,
			MountPath: constants.ExternalUrlAuthVolumePath,
		},
	}

	// Custom CA support
	if workloadconfig.IsCustomCAEnabled(EnvVarPrefixCARunAi) {
		volumeMounts = append(volumeMounts, ComposeCAVolumeMount(EnvVarPrefixCARunAi))
	}

	resources, err := workloadconfig.OAuth2ProxyResources()
	if err != nil {
		return nil, fmt.Errorf("get oauth2-proxy resources: %w", err)
	}

	container := &corev1.Container{
		Name:            cfg.ContainerName,
		Image:           os.Getenv(ExternalUrlAuthContainerImageEnv),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Args:            args,
		Resources:       resources,
		Env: []corev1.EnvVar{
			{
				Name:  "OAUTH2_PROXY_COOKIE_PATH",
				Value: cfg.ExternalPath,
			},
		},
		EnvFrom: []corev1.EnvFromSource{
			{
				SecretRef: &corev1.SecretEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: os.Getenv(WorkloadAuthSecretNameEnv),
					},
				},
			},
		},
		VolumeMounts: volumeMounts,
		SecurityContext: &corev1.SecurityContext{
			AllowPrivilegeEscalation: commonutils.Ref(false),
			RunAsNonRoot:             commonutils.Ref(true),
		},
	}

	// Declare ports unless OmitPortDeclaration is true (Knative compatibility)
	// Exception: ForceJsonResponse declares port as serving port (no protocol/name)
	if !cfg.OmitPortDeclaration || cfg.ForceJsonResponse {
		container.Ports = []corev1.ContainerPort{
			{
				ContainerPort: cfg.Port,
			},
		}

		if !cfg.OmitPortDeclaration {
			container.Ports[0].Protocol = corev1.ProtocolTCP
			container.Ports[0].Name = fmt.Sprintf("port-%d", cfg.Port)
		}
	}

	// Sidecar restart policy (for K8s 1.28+ native sidecars)
	if cfg.EnableSidecarRestartPolicy {
		initContainerRestartPolicy := corev1.ContainerRestartPolicyAlways
		container.RestartPolicy = &initContainerRestartPolicy
	}

	return container, nil
}

// GetOAuth2PathSuffixes returns the redirect URL suffix and proxy prefix suffix based on the path
func GetOAuth2PathSuffixes(path string) (redirectSuffix, proxyPrefixSuffix string) {
	if strings.HasSuffix(path, "/") {
		return "oauth2/callback", "oauth2"
	}
	return "/oauth2/callback", "/oauth2"
}

// GetOauth2ProxyContainer creates an OAuth2 proxy sidecar container for workloads
// This is a wrapper around BuildOAuth2ProxyContainer for backwards compatibility
func GetOauth2ProxyContainer(jobUrl *apis.URL, containerPort int32, authorizedUsers *string, authorizedGroups *string, idx int, nick string, omitPortDeclaration bool, forceJsonResponse bool, cookiePrefix string, upstreamTimeoutSeconds ...int) (*corev1.Container, error) {
	featureGateUtil := featuregate.GetFeatureGateUtilInstance()

	timeout := 0
	if len(upstreamTimeoutSeconds) > 0 {
		timeout = upstreamTimeoutSeconds[0]
	}

	cfg := OAuth2ProxyConfig{
		ContainerName:              ExternalUrlAuthContainerName + "-" + strconv.Itoa(idx),
		UpstreamURL:                "http://localhost:" + strconv.Itoa(int(containerPort)),
		ExternalScheme:             jobUrl.Scheme,
		ExternalHost:               jobUrl.Host,
		ExternalPath:               jobUrl.Path,
		Port:                       constants.ExternalUrlAuthContainerDefaultPort + int32(idx),
		CookiePrefix:               cookiePrefix,
		AuthorizedUsers:            authorizedUsers,
		AuthorizedGroups:           authorizedGroups,
		Nick:                       nick,
		OmitPortDeclaration:        omitPortDeclaration,
		ForceJsonResponse:          forceJsonResponse,
		EnableSidecarRestartPolicy: featureGateUtil.FeatureGateSidecarContainersEnable,
		UpstreamTimeoutSeconds:     timeout,
	}

	return BuildOAuth2ProxyContainer(cfg)
}

// SanitizeAuthorizedGroups checks if the authorizedGroups string is in the correct format and converts it to a stripped
// format that can be safely used as a command-line argument for the oauth2-proxy container.
// each group in the comma-separated string is checked for validity. spaces between group names are allowed but trimmed
// when converting to the stripped format
func SanitizeAuthorizedGroups(authorizedGroups string) (string, error) {
	// split the input string by commas
	groups := strings.Split(authorizedGroups, ",")

	// trim spaces from each group and validate
	validGroups := make([]string, 0, len(groups))
	for _, group := range groups {
		trimmedGroup := strings.TrimSpace(group)
		if trimmedGroup == "" {
			continue // Skip empty groups
		}
		if !groupPattern.MatchString(trimmedGroup) {
			return "", fmt.Errorf("authorizedGroups: invalid group: %s", trimmedGroup)
		}
		validGroups = append(validGroups, trimmedGroup)
	}

	if len(validGroups) == 0 {
		return "", errors.New("authorizedGroups: no valid groups found")
	}

	return strings.Join(validGroups, ","), nil
}

func ComposeOauthIngressUrl(ctx context.Context, exposedUrl flds.ExposedUrlItemAndRules, nick string, noSuffixUrlGenerated *bool, name string) (*apis.URL, error) {
	// validate consistency b/w the customUrl flag and the url value
	//
	var url *apis.URL

	customUrlEnabled := ptr.Deref(exposedUrl.Value.CustomUrl, false)
	customUrlProvided := ptr.Deref(exposedUrl.Value.Url, "") != ""
	if customUrlEnabled && !customUrlProvided {
		return url, errors.New("url is mandatory when customUrl is enabled")
	}
	if !customUrlEnabled && customUrlProvided {
		return url, errors.New("url should only be provided for customUrl")
	}

	var urlErr error
	if customUrlProvided {
		url, urlErr = apis.ParseURL(*exposedUrl.Value.Url)
	} else {
		urlSuffix := nick
		if !*noSuffixUrlGenerated {
			urlSuffix = ""
			*noSuffixUrlGenerated = true
		}
		url, urlErr = ComposeExposedUrlInt(ctx, urlSuffix, name)
	}
	if urlErr != nil {
		return url, urlErr
	}
	return url, nil
}

func getOidcIssuerUrl() string {
	oidcIssuerUrl := workloadconfig.GetOauth2ProxyOidcIssuerUrl()
	if oidcIssuerUrl != "" {
		return oidcIssuerUrl
	}

	return workloadconfig.GetControlPlaneEndpoint() + "/auth/realms/runai"
}

func IsWorkloadSupportsAuth(containers []corev1.Container, initContainers []corev1.Container) bool {
	for _, container := range containers {
		if strings.Contains(container.Name, ExternalUrlAuthContainerName) {
			return true
		}
	}

	for _, container := range initContainers {
		if strings.Contains(container.Name, ExternalUrlAuthContainerName) {
			return true
		}
	}

	return false
}
