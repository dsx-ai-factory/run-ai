package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	sdkcommon "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	corev1 "k8s.io/api/core/v1"
)

const (
	RunaiTrimStr = "runai."

	EnvVarOpenshift                       = "IS_OPENSHIFT"
	EnvVarClusterDomain                   = "CLUSTER_DOMAIN"
	EnvVarSubdomainSupport                = "SUBDOMAIN_SUPPORT"
	EnvVarIngressAnnotations              = "INGRESS_ANNOTATIONS"
	EnvVarIngressClass                    = "INGRESS_CLASS"
	EnvVarIngressEnabled                  = "INGRESS_ENABLED"
	EnvVarGatewayAPIEnabled               = "GATEWAY_API_ENABLED"
	EnvVarGatewayName                     = "GATEWAY_NAME"
	EnvVarGatewayNamespace                = "GATEWAY_NAMESPACE"
	EnvVarExternalAuthUrlEnabled          = "EXTERNAL_AUTH_URL_ENABLED"
	EnvVarJupyterUserIdentity             = "JUPYTER_USER_IDENTITY"
	EnvVarGitSyncImage                    = "GIT_SYNC_IMAGE"
	EnvVarOpenshiftRouteTerminationPolicy = "OPENSHIFT_ROUTE_TERMINATION_POLICY"

	OpenshiftRouteTerminationPolicyEdge        = "edge"
	OpenshiftRouteTerminationPolicyPassthrough = "passthrough"

	defaultCustomCAKey    = "runai-ca.pem"
	defaultCustomCASecret = "runai-ca-cert"

	EnvVarControlPlaneEndpoint     = "ENDPOINT"
	EnvVarOauth2ProxyOidcIssuerUrl = "OAUTH2_PROXY_OIDC_ISSUER_URL"

	defaultOauth2ProxyUsernameLoaderImage = sdkcommon.WorkloadRunaiOauth2ProxyUsernameLoaderImageRef
	EnvVarOauth2ProxyUsernameLoaderImage  = "RUNAI_OAUTH2_PROXY_USERNAME_LOADER_IMAGE"
	EnvVarOauth2ProxyResources            = constants.OAuth2ProxyResourcesEnvVar
	EnvVarUsernameLoaderResources         = constants.UsernameLoaderResourcesEnvVar
	EnvVarGitSyncResources                = constants.GitSyncResourcesEnvVar
	EnvVarS3FileSystemResources           = constants.S3FileSystemResourcesEnvVar

	EnvVarS3FileSystemImage  = "S3_FILE_SYSTEM_IMAGE"
	defaultS3FileSystemImage = "runai.jfrog.io/op-containers-prod/goofys"

	EnvVarGodebug = "GODEBUG"
)

var (
	ErrMissingADSEcryptionKey = errors.New(fmt.Sprintf("ADS Ecryption Key is not set in the %s environment variable", constants.RunaiAdsEncryptionKeyEnvVar))
)

// IsOpenshift - openshift indication is injected to the controller via environment variable
func IsOpenshift() bool {
	return strings.ToLower(os.Getenv(EnvVarOpenshift)) == "true"
}

// IngressEnabled controls whether legacy routing resources (Ingress or Route) are created.
// Defaults to true when the env var is not set, preserving current behavior.
func IngressEnabled() bool {
	val := os.Getenv(EnvVarIngressEnabled)
	if val == "" {
		return true
	}
	return strings.ToLower(val) == "true"
}

// GatewayEnabled returns true when Gateway API resources (HTTPRoute/GRPCRoute) should be created.
// Reads the GATEWAY_API_ENABLED env var set by the go-operator after validating Gateway prerequisites.
func GatewayEnabled() bool {
	return strings.ToLower(os.Getenv(EnvVarGatewayAPIEnabled)) == "true"
}

// GatewayName returns the name of the Gateway resource to reference in HTTPRoute/GRPCRoute parentRefs.
func GatewayName() string {
	return os.Getenv(EnvVarGatewayName)
}

// GatewayNamespace returns the namespace where the Gateway resource exists.
func GatewayNamespace() string {
	return os.Getenv(EnvVarGatewayNamespace)
}

func SubdomainSupportEnabled() bool {
	return strings.ToLower(os.Getenv(EnvVarSubdomainSupport)) == "true"
}

// ParseWorkloadInjectedContainerResources parses an injected-container resource
// environment variable. The boolean distinguishes an unset variable from an
// explicitly configured empty object.
func ParseWorkloadInjectedContainerResources(envVarName string) (corev1.ResourceRequirements, bool, error) {
	resourcesJSON, configured := os.LookupEnv(envVarName)
	if !configured || resourcesJSON == "" {
		return corev1.ResourceRequirements{}, false, nil
	}

	var wire struct {
		Limits   corev1.ResourceList `json:"limits,omitempty"`
		Requests corev1.ResourceList `json:"requests,omitempty"`
	}
	decoder := json.NewDecoder(strings.NewReader(resourcesJSON))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&wire); err != nil {
		return corev1.ResourceRequirements{}, true, fmt.Errorf("parse %s: %w", envVarName, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return corev1.ResourceRequirements{}, true, fmt.Errorf("parse %s: trailing JSON value", envVarName)
	}
	resources := corev1.ResourceRequirements{Limits: wire.Limits, Requests: wire.Requests}
	if err := global.ValidateWorkloadInjectedContainerResourceRequirements(resources); err != nil {
		return corev1.ResourceRequirements{}, true, fmt.Errorf("validate %s: %w", envVarName, err)
	}
	return resources, true, nil
}

// OAuth2ProxyResources returns the resource requirements configured for injected oauth2-proxy containers.
func OAuth2ProxyResources() (corev1.ResourceRequirements, error) {
	resources, _, err := ParseWorkloadInjectedContainerResources(EnvVarOauth2ProxyResources)
	return resources, err
}

// UsernameLoaderResources returns configured resources or the supplied legacy
// defaults when the environment variable is unset. An explicit empty object
// overrides the defaults with empty resource requirements.
func UsernameLoaderResources(defaultResources corev1.ResourceRequirements) (corev1.ResourceRequirements, error) {
	resources, configured, err := ParseWorkloadInjectedContainerResources(EnvVarUsernameLoaderResources)
	if err != nil {
		return corev1.ResourceRequirements{}, err
	}
	if !configured {
		return defaultResources, nil
	}
	return resources, nil
}

// GitSyncResources returns the resource requirements configured for injected git-sync containers.
func GitSyncResources() (corev1.ResourceRequirements, error) {
	resources, _, err := ParseWorkloadInjectedContainerResources(EnvVarGitSyncResources)
	return resources, err
}

// S3FileSystemResources returns the resource requirements configured for injected S3 filesystem containers.
func S3FileSystemResources() (corev1.ResourceRequirements, error) {
	resources, _, err := ParseWorkloadInjectedContainerResources(EnvVarS3FileSystemResources)
	return resources, err
}

func GetClusterDomain() string {
	return os.Getenv(EnvVarClusterDomain)
}

func ExternalAuthUrlEnabled() bool {
	envVar := os.Getenv(EnvVarExternalAuthUrlEnabled)
	return envVar != "false"
}

func OpenshiftRouteTerminationPolicy() string {
	switch strings.ToLower(os.Getenv(EnvVarOpenshiftRouteTerminationPolicy)) {
	case OpenshiftRouteTerminationPolicyPassthrough:
		return OpenshiftRouteTerminationPolicyPassthrough
	default:
		return OpenshiftRouteTerminationPolicyEdge
	}
}

// IngressAnnotations are injected to the ingresses created by the workload controller (for example, the jupyter ingresses)
func IngressAnnotations() map[string]string {
	annotationsStr := os.Getenv(EnvVarIngressAnnotations)
	if annotationsStr == "" {
		return nil
	}
	var annotations map[string]string
	err := json.Unmarshal([]byte(annotationsStr), &annotations)
	if err != nil {
		return nil
	}
	return annotations
}

// IngressClass is injected to the ingresses created by the workload controller as the ingress class name
func IngressClass() string {
	return os.Getenv(EnvVarIngressClass)
}

func JupyterUserIdentity() bool {
	envVar := os.Getenv(EnvVarJupyterUserIdentity)
	return envVar != "false"
}

func DefaultGitSyncImage() string {
	envVar := os.Getenv(EnvVarGitSyncImage)
	if envVar != "" {
		return envVar
	}
	return sdkcommon.GitSyncImageRef
}

// CustomCAEnabled returns true if the custom CA is enabled
func IsCustomCAEnabled(envPrefix string) bool {
	envName := fmt.Sprintf("%s_ENABLED", envPrefix)
	return strings.ToLower(os.Getenv(envName)) == "true"
}

// GetCustomCAKey returns the key of the custom CA
func GetCustomCAKey(envPrefix string) string {
	envName := fmt.Sprintf("%s_KEY", envPrefix)

	envVar := os.Getenv(envName)
	if envVar != "" {
		return envVar
	}
	return defaultCustomCAKey
}

// GetCustomCASecret returns the secret of the custom CA
func GetCustomCASecret(envPrefix string) string {
	envName := fmt.Sprintf("%s_SECRET", envPrefix)

	envVar := os.Getenv(envName)
	if envVar != "" {
		return envVar
	}
	return defaultCustomCASecret
}

func GetControlPlaneEndpoint() string {
	return os.Getenv(EnvVarControlPlaneEndpoint)
}

func GetOauth2ProxyOidcIssuerUrl() string {
	return os.Getenv(EnvVarOauth2ProxyOidcIssuerUrl)
}

func GetOauth2ProxyUsernameLoaderImage() string {
	envVar := os.Getenv(EnvVarOauth2ProxyUsernameLoaderImage)
	if envVar != "" {
		return envVar
	}
	return defaultOauth2ProxyUsernameLoaderImage
}

func DefaultS3FileSystemImage() string {
	envVar := os.Getenv(EnvVarS3FileSystemImage)
	if envVar != "" {
		return envVar
	}
	return defaultS3FileSystemImage
}

// GetGodebug returns the GODEBUG value of the current process, which the
// operator sets to fips140=<mode> on cluster components. Injected sidecars
// (e.g. goofys) propagate it so they run with the cluster's FIPS mode.
func GetGodebug() string {
	return os.Getenv(EnvVarGodebug)
}

// GetCustomCAConfigMapName returns encryption key for private assets
func GetAdsEncryptionKey() ([]byte, error) {
	adsEncryptionKey := os.Getenv(constants.RunaiAdsEncryptionKeyEnvVar)
	if adsEncryptionKey != "" {
		return []byte(adsEncryptionKey), nil
	}
	return nil, ErrMissingADSEcryptionKey
}
