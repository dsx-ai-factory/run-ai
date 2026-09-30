/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package init_containers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants/oauth2proxy"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/builders"
	corev1 "k8s.io/api/core/v1"
)

const (
	oauth2ProxyContainerName    = "runai-oauth2-proxy"
	oauth2AuthVolumeName        = "runai-auth"
	oauth2AuthVolumePath        = "/etc/runai/auth/users"
	oauth2RedisSessionStoreType = "redis"
	oauth2RunaiTrimStr          = "runai."
	oauth2CADefaultMode         = int32(420)

	oauth2CookiePathEnvName = "OAUTH2_PROXY_COOKIE_PATH"
	oauth2DefaultScheme     = "https"

	oauth2CallbackPath = "oauth2/callback"
	oauth2ProxyPrefix  = "oauth2"
)

var oauth2GroupPattern = regexp.MustCompile(`^[a-zA-Z0-9][-_a-zA-Z0-9.]{0,62}[a-zA-Z0-9]$`)

// ShouldBuildOAuth2Proxy reports whether an exposed URL is backed by an OAuth
// proxy. Some serving URLs require the proxy even when external URL auth is
// disabled cluster-wide.
func ShouldBuildOAuth2Proxy(config composer.ClusterStorageConfig, item *composer.ExposedUrlInfoItem) bool {
	return config != nil && item != nil &&
		(config.ExternalAuthUrlEnabled() || item.BuildWhenExternalAuthURLDisabled)
}

func BuildOAuth2ProxyContainers(info *composer.ExposedUrlInfo) ([]corev1.Container, *workloadsApi.Error) {
	result := make([]corev1.Container, 0)
	if info == nil || info.Config == nil || len(info.Items) == 0 {
		return result, nil
	}
	cfg := info.Config

	var resources corev1.ResourceRequirements
	resourcesLoaded := false

	// Mirror the notification's connections-pre-processor and the Ingress
	// (UrlsToWorkloadRunner): start at false so the first auto URL is
	// suffix-less, regardless of jupyter/tensorboard. Keeping these three
	// path computations aligned ensures the OAuth2 proxy-prefix matches the
	// Ingress path that the user actually hits.
	//
	// When ForceSuffixInUrls is set (distributed non-primary role), start at
	// true so every URL includes its nick suffix and avoids competing with the
	// primary role for the same Ingress host.
	noSuffixUrlGenerated := info.ForceSuffixInUrls

	for idx, key := range composer.SortedMapKeys(info.Items) {
		item := info.Items[key]
		if !ShouldBuildOAuth2Proxy(cfg, item) {
			continue
		}

		if !resourcesLoaded {
			var resourcesErr error
			resources, resourcesErr = cfg.OAuth2ProxyResources()
			if resourcesErr != nil {
				return nil, composer.ErrClusterValueInvalid("runaiOauth2ProxyResources", resourcesErr)
			}
			resourcesLoaded = true
		}

		if cfg.OAuth2ProxyImage() == "" {
			return nil, composer.ErrClusterValueMissing("runaiOauth2ProxyImage")
		}

		var parsedUrl *url.URL
		var parseErr error
		if item.Url == "" {
			urlSuffix := item.Nick
			if !noSuffixUrlGenerated {
				urlSuffix = ""
				noSuffixUrlGenerated = true
			}
			parsedUrl, parseErr = composeAutoExposedUrl(cfg.ClusterDomain(), info.ProjectName, info.WorkloadName, urlSuffix, cfg.SubdomainSupportEnabled())
		} else {
			parsedUrl, parseErr = url.Parse(item.Url)
		}
		if parseErr != nil {
			return nil, workloadsApi.NewError(http.StatusBadRequest,
				fmt.Sprintf("exposedUrls: malformed url %q: %v", item.Url, parseErr))
		}

		scheme := parsedUrl.Scheme
		if scheme == "" {
			scheme = oauth2DefaultScheme
		}

		redirectUrlSuffix, proxyPrefixSuffix := oauth2PathSuffixes(parsedUrl.Path)
		externalURL := fmt.Sprintf("%s://%s%s", scheme, parsedUrl.Host, parsedUrl.Path)
		redirectURL := externalURL + redirectUrlSuffix
		proxyPrefix := parsedUrl.Path + proxyPrefixSuffix
		whitelistDomain := parsedUrl.Host

		if cfg.IsOpenshift() && cfg.SubdomainSupportEnabled() {
			redirectURL = strings.Replace(redirectURL, oauth2RunaiTrimStr, "", 1)
			whitelistDomain = strings.Replace(whitelistDomain, oauth2RunaiTrimStr, "", 1)
		}
		// Scope the session cookie to the exposed host.
		cookieDomain := (&url.URL{Host: whitelistDomain}).Hostname()

		oidcIssuerUrl := cfg.OAuth2ProxyOidcIssuerUrl()
		port := convutils.Deref(item.OAuthPort, 0)
		cookiePrefix := item.CookiePrefix
		if cookiePrefix == "" {
			cookiePrefix = fmt.Sprintf("%s-%s-auth-proxy", info.WorkloadName, item.Nick)
		}

		args := []string{
			"--provider=keycloak-oidc",
			fmt.Sprintf("--upstream=http://localhost:%d", convutils.Deref(item.ContainerPort, 0)),
			fmt.Sprintf("--redirect-url=%s", redirectURL),
			fmt.Sprintf("--oidc-issuer-url=%s", oidcIssuerUrl),
			"--scope=openid",
			fmt.Sprintf("--http-address=0.0.0.0:%d", port),
			fmt.Sprintf("--proxy-prefix=%s", proxyPrefix),
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
			"--custom-templates-dir=" + oauth2proxy.TemplatesPath,
			"--pass-authorization-header=false",
			fmt.Sprintf("--cookie-name=%s", cookiePrefix),
		}

		if cfg.OAuth2ProxySessionStoreType() == oauth2RedisSessionStoreType {
			args = append(args,
				"--session-store-type=redis",
				fmt.Sprintf("--redis-connection-url=redis://%s", cfg.OAuth2ProxyRedisFQDN()),
			)
		}
		if item.UpstreamTimeoutSeconds > 0 {
			args = append(args, fmt.Sprintf("--upstream-timeout=%ds", item.UpstreamTimeoutSeconds))
		}
		if item.ForceJSONResponse {
			args = append(args, "--force-json-errors=true")
		}

		authorizedUsers := item.CookieAuthorizedUsers()
		authorizedGroups := item.CookieAuthorizedGroups()

		if authorizedUsers != nil {
			args = append(args, fmt.Sprintf("--authenticated-emails-file=%s/%s",
				oauth2AuthVolumePath, oauth2SafeFileNameFromNick(item.Nick)))
		}

		if authorizedGroups != nil {
			sanitizedGroups, groupsErr := sanitizeAuthorizedGroups(*authorizedGroups)
			if groupsErr != nil {
				return nil, groupsErr
			}
			for _, group := range strings.Split(sanitizedGroups, ",") {
				args = append(args, "--allowed-group="+group)
			}
		}

		if authorizedUsers == nil || authorizedGroups != nil {
			args = append(args, "--email-domain=*")
		}

		volumeMounts := []corev1.VolumeMount{
			{Name: oauth2AuthVolumeName, MountPath: oauth2AuthVolumePath},
		}
		if info.OAuth2ProxyTemplatesSupported {
			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name: oauth2proxy.VolumeName, MountPath: oauth2proxy.TemplatesPath, ReadOnly: true,
			})
		}
		if cfg.IsCustomCAAuthEnabled() {
			volumeMounts = append(volumeMounts, oauth2CAVolumeMount(cfg))
		}

		container := corev1.Container{
			Name:            fmt.Sprintf("%s-%d", oauth2ProxyContainerName, idx),
			Image:           cfg.OAuth2ProxyImage(),
			ImagePullPolicy: corev1.PullIfNotPresent,
			Args:            args,
			Env: []corev1.EnvVar{
				{Name: oauth2CookiePathEnvName, Value: parsedUrl.Path},
			},
			EnvFrom: []corev1.EnvFromSource{
				{SecretRef: &corev1.SecretEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: cfg.OAuth2ProxyAuthSecretName()},
				}},
			},
			VolumeMounts: volumeMounts,
			Resources:    resources,
			SecurityContext: &corev1.SecurityContext{
				AllowPrivilegeEscalation: convutils.Ref(false),
				RunAsNonRoot:             convutils.Ref(true),
			},
		}

		// For authenticated inference, the proxy is Knative's only
		// port-declaring container, so serving traffic reaches the proxy while
		// the metrics Service still targets the user container directly.
		if !item.OmitPortDeclaration || item.ForceJSONResponse {
			container.Ports = []corev1.ContainerPort{{ContainerPort: port}}
			if !item.OmitPortDeclaration {
				container.Ports[0].Protocol = corev1.ProtocolTCP
				container.Ports[0].Name = fmt.Sprintf("port-%d", port)
			}
		}

		if convutils.Deref(item.EnableSidecarRestartPolicy, true) {
			restartAlways := corev1.ContainerRestartPolicyAlways
			container.RestartPolicy = &restartAlways
		}
		result = append(result, container)
	}

	return result, nil
}

func BuildOAuth2ProxyVolumes(info *composer.ExposedUrlInfo) []corev1.Volume {
	result := make([]corev1.Volume, 0)
	if info == nil || info.Config == nil || len(info.Items) == 0 {
		return result
	}

	result = append(result, corev1.Volume{
		Name:         oauth2AuthVolumeName,
		VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
	})
	if info.OAuth2ProxyTemplatesSupported {
		result = append(result, corev1.Volume{
			Name: oauth2proxy.VolumeName,
			VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: oauth2proxy.ConfigMapName},
				Items:                []corev1.KeyToPath{{Key: oauth2proxy.TemplateKey, Path: oauth2proxy.TemplateKey}},
			}},
		})
	}

	if info.Config.IsCustomCAAuthEnabled() {
		result = append(result, oauth2CAVolume(info.Config))
	}

	return result
}

func oauth2CAVolumeMount(cfg composer.ClusterStorageConfig) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      cfg.CustomCAAuthSecret(),
		MountPath: fmt.Sprintf("/etc/ssl/certs/%s", cfg.CustomCAAuthKey()),
		SubPath:   cfg.CustomCAAuthKey(),
	}
}

func oauth2CAVolume(cfg composer.ClusterStorageConfig) corev1.Volume {
	return corev1.Volume{
		Name: cfg.CustomCAAuthSecret(),
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				DefaultMode: convutils.Ref(oauth2CADefaultMode),
				Items: []corev1.KeyToPath{
					{Key: cfg.CustomCAAuthKey(), Path: cfg.CustomCAAuthKey()},
				},
				Optional:   convutils.Ref(true),
				SecretName: cfg.CustomCAAuthSecret(),
			},
		},
	}
}

// composeAutoExposedUrl returns the auto-generated URL for an exposed URL that
// has no custom URL set. We deliberately delegate to the cluster's shared
// builder (workload-common/builders) so the V2 composer, the V1 cluster builder
// and the notification's connections-pre-processor all derive the URL from a
// single source of truth.
//
// TODO: when the workload-common/builders package is removed, inline the
// builder's logic here. The original implementation was:
//
//	parsed, err := url.Parse(clusterDomain)
//	if err != nil || parsed == nil { return nil, ... }
//	if subdomainSupported {
//		subdomain := projectName + "-" + workloadName
//		if suffix != "" { subdomain += "-" + suffix }
//		parsed.Host = subdomain + "." + parsed.Host
//	} else {
//		path := "/" + projectName + "/" + workloadName
//		if suffix != "" { path += "/" + suffix }
//		if parsed.Path != "" { path += "/" + parsed.Path }
//		parsed.Path = path
//	}
//	if parsed.Path == "" { parsed.Path = "/" }
//	return parsed, nil
func composeAutoExposedUrl(clusterDomain, projectName, workloadName, suffix string, subdomainSupported bool) (*url.URL, error) {
	parsed, err := builders.ComposeExposedUrl(suffix, workloadName, projectName, clusterDomain, builders.ComposeUrlOptions{
		SubdomainSupported: subdomainSupported,
	})
	if err != nil {
		return nil, err
	}
	if parsed == nil {
		return nil, fmt.Errorf("ComposeExposedUrl returned nil URL for workload=%q project=%q suffix=%q domain=%q",
			workloadName, projectName, suffix, clusterDomain)
	}
	stdUrl := url.URL(*parsed)
	return &stdUrl, nil
}

// ComposeAutoExposedURL exposes the shared V1-compatible URL generation to typed V2 composers.
func ComposeAutoExposedURL(clusterDomain, projectName, workloadName, suffix string, subdomainSupported bool) (*url.URL, error) {
	return composeAutoExposedUrl(clusterDomain, projectName, workloadName, suffix, subdomainSupported)
}

func oauth2PathSuffixes(path string) (redirectSuffix, proxyPrefixSuffix string) {
	if strings.HasSuffix(path, "/") {
		return oauth2CallbackPath, oauth2ProxyPrefix
	}
	return "/" + oauth2CallbackPath, "/" + oauth2ProxyPrefix
}

func oauth2SafeFileNameFromNick(nick string) string {
	return base64.StdEncoding.EncodeToString([]byte(nick))
}

func sanitizeAuthorizedGroups(authorizedGroups string) (string, *workloadsApi.Error) {
	groups := strings.Split(authorizedGroups, ",")

	validGroups := make([]string, 0, len(groups))
	for _, group := range groups {
		trimmedGroup := strings.TrimSpace(group)
		if trimmedGroup == "" {
			continue
		}
		if !oauth2GroupPattern.MatchString(trimmedGroup) {
			return "", workloadsApi.NewError(http.StatusBadRequest,
				fmt.Sprintf("authorizedGroups: invalid group: %s", trimmedGroup))
		}
		validGroups = append(validGroups, trimmedGroup)
	}

	if len(validGroups) == 0 {
		return "", workloadsApi.NewError(http.StatusBadRequest, "authorizedGroups: no valid groups found")
	}

	return strings.Join(validGroups, ","), nil
}
