// Package installconfig defines the installer spec file (YAML) and loads it
// with strict, type-level validation. The struct mirrors the installer design
// doc; many fields are parsed but not yet consumed by the orchestrator, so the
// file format stays stable as features land.
package installconfig

type Config struct {
	Install            Install             `yaml:"install"`
	TLS                *TLS                `yaml:"tls,omitempty"`
	Registry           *Registry           `yaml:"registry,omitempty"`
	Airgap             *Airgap             `yaml:"airgap,omitempty"`
	ControlPlane       *ControlPlane       `yaml:"controlPlane,omitempty"`
	Cluster            *Cluster            `yaml:"cluster,omitempty"`
	Prerequisites      *Prerequisites      `yaml:"prerequisites,omitempty"`
	Features           *Features           `yaml:"features,omitempty"`
	ExternalConnection *ExternalConnection `yaml:"externalConnection,omitempty"`
	SystemNodes        *SystemNodes        `yaml:"systemNodes,omitempty"`
	OpenShift          *OpenShift          `yaml:"openshift,omitempty"`
}

// OpenShift carries OpenShift-only install settings. Consumed when
// install.kubernetesDistribution is "openshift".
type OpenShift struct {
	OLMCatalog *OLMCatalog `yaml:"olmCatalog,omitempty"`
}

// OLMCatalog points OperatorHub installs at a specific catalog source. Air-
// gapped clusters set this to a mirrored CatalogSource; connected clusters can
// omit it to use the public catalogs.
type OLMCatalog struct {
	Source          string `yaml:"source,omitempty"`
	SourceNamespace string `yaml:"sourceNamespace,omitempty"`
}

// SystemNodes pins the Run:ai system components (control plane + cluster
// services) to a chosen set of nodes. Disabled by default.
type SystemNodes struct {
	Enabled      bool              `yaml:"enabled"`
	Mode         string            `yaml:"mode,omitempty"` // preferred (default) | required
	Label        string            `yaml:"label,omitempty"`
	ManageLabels *bool             `yaml:"manageLabels,omitempty"`
	Nodes        []string          `yaml:"nodes,omitempty"`
	Selector     map[string]string `yaml:"selector,omitempty"`
}

// ExternalConnection selects how external traffic reaches Run:ai services.
// Required on non-OpenShift installs; must be omitted on OpenShift (which uses
// Routes). type is one of: ingress | gatewayApi | route.
type ExternalConnection struct {
	Type       string                `yaml:"type"`
	Ingress    *IngressConnection    `yaml:"ingress,omitempty"`
	GatewayAPI *GatewayAPIConnection `yaml:"gatewayApi,omitempty"`
}

type IngressConnection struct {
	ClassName string `yaml:"className,omitempty"`
}

type GatewayAPIConnection struct {
	Name      string `yaml:"name,omitempty"`
	Namespace string `yaml:"namespace,omitempty"`
}

type Install struct {
	Mode                   string `yaml:"mode"`
	Connectivity           string `yaml:"connectivity"`
	KubernetesDistribution string `yaml:"kubernetesDistribution,omitempty"`
}

// Registry holds the registry coordinates and credentials. The credential
// fields form one of three mutually exclusive auth schemes (validated by
// resolveRegistryFromConfig):
//   - token auth: token | tokenPath (JFrog/NGC connected installs)
//   - basic auth: username + (password | passwordPath) (e.g. a Harbor mirror)
//   - none: host only (an unauthenticated registry)
//
// A secret can be supplied inline (token/password) or read from a file
// (tokenPath/passwordPath).
type Registry struct {
	Type         string `yaml:"type,omitempty"`
	Host         string `yaml:"host,omitempty"`
	Username     string `yaml:"username,omitempty"`
	Password     string `yaml:"password,omitempty"`
	PasswordPath string `yaml:"passwordPath,omitempty"`
	Token        string `yaml:"token,omitempty"`
	TokenPath    string `yaml:"tokenPath,omitempty"`
	CAChainPath  string `yaml:"caChainPath,omitempty"`
	// ChartRepository overrides the NGC Helm repository.
	ChartRepository string `yaml:"chartRepository,omitempty"`
	// Insecure marks the registry as plain HTTP (no TLS).
	Insecure bool `yaml:"insecure,omitempty"`
	// InsecureSkipVerify keeps HTTPS but skips certificate verification.
	InsecureSkipVerify bool `yaml:"insecureSkipVerify,omitempty"`
}

type Airgap struct {
	BundlePath string `yaml:"bundlePath"`
	// WorkingDir is the directory the bundle is extracted into. It is mandatory:
	// the installer never creates temp dirs on the user's machine, since the
	// bundle is large and the caller must control where it lands.
	WorkingDir string `yaml:"workingDir"`
	ValuesDir  string `yaml:"valuesDir,omitempty"`
	// SkipPrepare skips bundle extraction and the image push, for re-runs where
	// the working dir is already populated and the images are already in the
	// registry. Only workingDir is required in that case.
	SkipPrepare bool `yaml:"skipPrepare"`
}

type TLS struct {
	CertPath           string `yaml:"certPath,omitempty"`
	KeyPath            string `yaml:"keyPath,omitempty"`
	CAChainPath        string `yaml:"caChainPath,omitempty"`
	GenerateSelfSigned bool   `yaml:"generateSelfSigned,omitempty"`
}

type ControlPlane struct {
	Enabled          bool           `yaml:"enabled"`
	Version          string         `yaml:"version,omitempty"`
	MultiTenant      bool           `yaml:"multiTenant,omitempty"`
	Domain           string         `yaml:"domain"`
	TLS              *TLS           `yaml:"tls,omitempty"`
	CustomHelmValues map[string]any `yaml:"customHelmValues,omitempty"`
}

type Cluster struct {
	Enabled bool   `yaml:"enabled"`
	Version string `yaml:"version,omitempty"`
	Name    string `yaml:"name"`
	Domain  string `yaml:"domain,omitempty"`
	// WorkloadRouting selects how workloads are exposed: "host" (subdomain
	// routing, the default) or "path". It maps to the cluster chart's
	// clusterConfig.global.subdomainSupport value.
	WorkloadRouting  string         `yaml:"workloadRouting,omitempty"`
	TLS              *TLS           `yaml:"tls,omitempty"`
	Inference        *Inference     `yaml:"inference,omitempty"`
	ControlPlaneURL  string         `yaml:"controlPlaneUrl,omitempty"`
	CustomHelmValues map[string]any `yaml:"customHelmValues,omitempty"`
}

type Inference struct {
	WildcardDomain string `yaml:"wildcardDomain"`
	WildcardTLS    *TLS   `yaml:"wildcardTls,omitempty"`
}

type Dependency struct {
	Mode             string               `yaml:"mode"`
	Version          string               `yaml:"version,omitempty"`
	Class            string               `yaml:"class,omitempty"`
	Type             string               `yaml:"type,omitempty"` // gpuOperator: real | fake
	CustomHelmValues map[string]any       `yaml:"customHelmValues,omitempty"`
	OpenShift        *DependencyOpenShift `yaml:"openshift,omitempty"`
}

// DependencyOpenShift holds the OpenShift-only customization for a dependency
// that installs via OperatorHub (OLM) rather than Helm. It is the OpenShift
// counterpart of customHelmValues and is generic across operators.
type DependencyOpenShift struct {
	// InstallPlanApproval sets the OLM Subscription's upgrade strategy:
	// "manual" (the default) pins the operator to the installed version and
	// requires manual approval for later upgrades; "automatic" lets OLM upgrade
	// the operator to newer versions in the channel on its own.
	InstallPlanApproval string `yaml:"installPlanApproval,omitempty"`
	// CustomResource is merged onto the operator's primary OpenShift custom
	// resource spec (ClusterPolicy for the GPU operator, NicClusterPolicy for
	// the network operator, ...). It mirrors how an OpenShift admin customizes
	// an OLM-installed operator: by editing its CR, not Helm values.
	CustomResource map[string]any `yaml:"customResource,omitempty"`
	// SubscriptionConfig configures the operator pod itself via the OLM
	// Subscription's spec.config. It is how operator-level settings that have no
	// custom resource are set on OpenShift — e.g. the NIM operator's validation
	// webhooks, gated by its ENABLE_WEBHOOKS environment variable.
	SubscriptionConfig *SubscriptionConfig `yaml:"subscriptionConfig,omitempty"`
}

// SubscriptionConfig is the subset of an OLM Subscription's spec.config the
// installer exposes (environment variables today; extendable to resources,
// tolerations, etc.).
type SubscriptionConfig struct {
	Env map[string]string `yaml:"env,omitempty"`
}

type Prerequisites struct {
	Skip            bool        `yaml:"skip"`
	GPUOperator     *Dependency `yaml:"gpuOperator,omitempty"`
	Ingress         *Dependency `yaml:"ingress,omitempty"`
	Prometheus      *Dependency `yaml:"prometheus,omitempty"`
	NetworkOperator *Dependency `yaml:"networkOperator,omitempty"`
	MetalLB         *Dependency `yaml:"metalLB,omitempty"`
}

type Features struct {
	Skip        bool        `yaml:"skip"`
	Knative     *Dependency `yaml:"knative,omitempty"`
	Kubeflow    *Dependency `yaml:"kubeflow,omitempty"`
	MPIOperator *Dependency `yaml:"mpiOperator,omitempty"`
	LWS         *Dependency `yaml:"lws,omitempty"`
	NIMOperator *Dependency `yaml:"nimOperator,omitempty"`
	Grove       *Dependency `yaml:"grove,omitempty"`
	Dynamo      *Dependency `yaml:"dynamo,omitempty"`
}
