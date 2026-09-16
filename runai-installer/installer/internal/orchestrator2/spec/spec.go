/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package spec

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
	pkgvalidation "github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/validation"
)

type Spec struct {
	Command                Command
	InstallMode            InstallMode
	DryRun                 bool
	NonInteractive         bool
	DevMode                bool
	OnStepFailure          FailurePolicy
	Output                 Output
	Connectivity           Connectivity
	Env                    Env
	KubernetesDistribution string
	Registry               *Registry
	Airgap                 *Airgap
	TLS                    *TLS
	SelfSignedTLS          *SelfSignedTLS
	ControlPlane           *ControlPlane
	Cluster                *Cluster
	Prerequisites          *Prerequisites
	Features               *Features

	ExternalConnection *ExternalConnection

	// SystemNodes, when set, pins the Run:ai system components (control plane
	// and cluster services) to a chosen set of nodes. Nil means the feature is
	// off (the charts keep their built-in soft defaults). It never affects user
	// workloads.
	SystemNodes *SystemNodes

	// OpenShift carries OpenShift-only knobs. Nil keeps the OpenShift defaults.
	OpenShift *OpenShift

	// AirgapPack carries the airgap pack command's settings. It is set only for
	// that command, which shares none of the install settings above.
	AirgapPack *AirgapPack
	// AirgapLoad carries the airgap load command's settings, on the same terms.
	AirgapLoad *AirgapLoad
}

type InstallMode string

const (
	InstallModeFull             InstallMode = "full_installation"
	InstallModeControlPlaneOnly InstallMode = "runai_control_plane_only"
	InstallModeClusterOnly      InstallMode = "runai_cluster_only"
)

// IsKnown reports whether the mode is a recognized value. The empty string is
// accepted and treated as the full-installation default.
func (m InstallMode) IsKnown() bool {
	switch m {
	case "", InstallModeFull, InstallModeControlPlaneOnly, InstallModeClusterOnly:
		return true
	}
	return false
}

// OpenShift holds OpenShift-specific install settings consumed when
// KubernetesDistribution is "openshift".
type OpenShift struct {
	// OLMCatalogSource overrides the OperatorHub catalog operators are pulled
	// from. Empty uses the public catalogs (certified-operators / redhat-
	// operators). Air-gapped clusters point this at a mirrored CatalogSource.
	OLMCatalogSource string
	// OLMCatalogSourceNamespace is the namespace the catalog source lives in.
	// Empty defaults to openshift-marketplace.
	OLMCatalogSourceNamespace string
}

type SystemNodeMode string

const (
	SystemNodeModePreferred SystemNodeMode = "preferred"
	SystemNodeModeRequired  SystemNodeMode = "required"
)

// SystemNodes describes where the Run:ai system components run. The chosen
// nodes (by name or label selector) are labeled with Label, and a node affinity
// targeting that label is injected into the control-plane and cluster charts.
type SystemNodes struct {
	Mode         SystemNodeMode
	Label        string
	ManageLabels bool
	Nodes        []string
	Selector     map[string]string
}

type ExternalConnectionType string

const (
	ExternalConnectionIngress    ExternalConnectionType = "ingress"
	ExternalConnectionGatewayAPI ExternalConnectionType = "gatewayApi"
	ExternalConnectionRoute      ExternalConnectionType = "route"
)

// ExternalConnection selects the cluster's external routing mechanism and the
// settings it needs. Exactly one of Ingress/GatewayAPI is populated, matching
// Type.
type ExternalConnection struct {
	Type       ExternalConnectionType
	Ingress    *IngressConnection
	GatewayAPI *GatewayAPIConnection
}

// IngressConnection configures Ingress-based routing.
type IngressConnection struct {
	ClassName string
}

// GatewayAPIConnection references an existing Gateway (the platform provides the
// Gateway controller; Run:ai/Knative only attach HTTPRoutes to it).
type GatewayAPIConnection struct {
	Name      string
	Namespace string
}

type Connectivity string

const (
	ConnectivityConnected Connectivity = "connected"
	ConnectivityAirGapped Connectivity = "air_gapped"
)

type Env string

const (
	EnvProd    Env = "prod"
	EnvStaging Env = "staging"
	EnvLab     Env = "lab"
	EnvTest    Env = "test"
)

type RegistryType string

const (
	RegistryTypeJFrog RegistryType = "jfrog"
	RegistryTypeNGC   RegistryType = "ngc"
)

type Registry struct {
	Type        RegistryType
	Token       string
	Username    string
	Password    string
	Host        string
	CAChainPath string
	// ChartRepository overrides the NGC Helm repository.
	ChartRepository string
	// Insecure means the registry is served over plain HTTP (no TLS).
	Insecure bool
	// InsecureSkipVerify means the registry is served over HTTPS but its
	// certificate must not be verified.
	InsecureSkipVerify bool
}

type Airgap struct {
	BundlePath  string
	WorkingDir  string
	ValuesDir   string
	SkipPrepare bool
}

type TLS struct {
	CertPath    string
	KeyPath     string
	CAChainPath string
}

type SelfSignedTLS struct {
	Enabled           bool
	ConflictingFields []string
	OutputDir         string
}

type OutputFormat string

const (
	OutputFormatNone OutputFormat = ""
	OutputFormatJSON OutputFormat = "json"
	OutputFormatText OutputFormat = "text"
)

type Output struct {
	Format     OutputFormat
	EventsFile string
}

type FailurePolicy string

const (
	FailurePolicyAbort FailurePolicy = ""
	FailurePolicySkip  FailurePolicy = "skip"
)

type ControlPlane struct {
	Domain           string
	Version          string
	TLS              *TLS
	CustomHelmValues map[string]string
}

type Cluster struct {
	Name            string
	Domain          string
	Version         string
	TLS             *TLS
	WorkloadRouting string
	Inference       *Inference
	// ControlPlaneURL points a cluster-only install at an external control
	// plane. The cluster's identity there comes from exactly one of two
	// sources: the pre-registered UID plus its BootstrapToken/ClientSecret, or
	// ControlPlaneCredentials, with which the installer registers the cluster
	// by name and fetches the install info itself.
	ControlPlaneURL         string
	UID                     string
	BootstrapToken          string
	ClientSecret            string
	ControlPlaneCredentials *ControlPlaneCredentials
	CustomHelmValues        map[string]string
}

// ControlPlaneCredentials authenticate the installer to an external control
// plane: either Username+Password or a pre-issued bearer Token.
type ControlPlaneCredentials struct {
	Username string
	Password string
	Token    string
}

func (c *Cluster) UsesExternalControlPlane() bool {
	return c != nil && c.ControlPlaneURL != ""
}

// UsesControlPlaneCredentials reports whether the installer registers the
// cluster with the external control plane itself (get-or-create by name and
// fetch the install info) instead of relying on a user-supplied UID and
// bootstrap token / client secret.
func (c *Cluster) UsesControlPlaneCredentials() bool {
	return c != nil && c.ControlPlaneCredentials != nil
}

// UsesSuppliedClusterIdentity reports whether the user pre-registered the
// cluster and supplied its UID together with its bootstrap token or client
// secret.
func (c *Cluster) UsesSuppliedClusterIdentity() bool {
	return c != nil && c.UID != ""
}

// FieldError is a validation problem attributed to a config field.
type FieldError struct {
	Field   string
	Message string
}

func (e FieldError) Error() string { return e.Message }

// ExternalControlPlaneProblems returns every problem with the external
// control-plane settings (empty when none are set or all are consistent). It
// is shared by Validate (the hard gate, which reports the first) and the
// offline input checks (which surface all of them).
func (s *Spec) ExternalControlPlaneProblems() []FieldError {
	if s.Cluster == nil {
		return nil
	}
	c := s.Cluster
	hasExternalFields := c.ControlPlaneURL != "" || c.UID != "" || c.BootstrapToken != "" || c.ClientSecret != "" || c.ControlPlaneCredentials != nil
	if !hasExternalFields {
		return nil
	}
	var problems []FieldError
	if s.InstallMode != InstallModeClusterOnly {
		problems = append(problems, FieldError{"install.mode", "external control plane details are only supported for runai_cluster_only"})
	}
	if c.ControlPlaneURL == "" {
		problems = append(problems, FieldError{"cluster.controlPlaneUrl", "cluster.controlPlaneUrl is required"})
	} else if u, err := url.ParseRequestURI(c.ControlPlaneURL); err != nil || u.Scheme != "https" || u.Host == "" || u.Path != "" || u.RawQuery != "" || u.User != nil {
		problems = append(problems, FieldError{"cluster.controlPlaneUrl", "cluster.controlPlaneUrl must be an HTTPS URL without a path"})
	}
	if c.Domain == "" {
		problems = append(problems, FieldError{"cluster.domain", "cluster.domain is required"})
	}
	if c.ControlPlaneCredentials != nil {
		if c.UID != "" || c.BootstrapToken != "" || c.ClientSecret != "" {
			problems = append(problems, FieldError{"cluster.controlPlaneCredentials", "cluster.controlPlaneCredentials cannot be combined with cluster.uid, cluster.bootstrapTokenPath or cluster.clientSecretPath: either pre-register the cluster and supply its identity, or let the installer register it"})
		}
		problems = append(problems, c.ControlPlaneCredentials.problems()...)
		return problems
	}
	if c.UID == "" {
		problems = append(problems, FieldError{"cluster.uid", "cluster.uid is required (or set cluster.controlPlaneCredentials to let the installer register the cluster)"})
	}
	if (c.BootstrapToken == "") == (c.ClientSecret == "") {
		problems = append(problems, FieldError{"cluster.bootstrapTokenPath", "exactly one of cluster.bootstrapTokenPath or cluster.clientSecretPath is required (or set cluster.controlPlaneCredentials to let the installer register the cluster)"})
	}
	return problems
}

func (cc *ControlPlaneCredentials) problems() []FieldError {
	const field = "cluster.controlPlaneCredentials"
	usesBasic := cc.Username != "" || cc.Password != ""
	switch {
	case usesBasic && cc.Token != "":
		return []FieldError{{field, field + ": username/password and token are mutually exclusive"}}
	case usesBasic:
		var problems []FieldError
		if cc.Username == "" {
			problems = append(problems, FieldError{field + ".username", field + ": a password requires a username"})
		}
		if cc.Password == "" {
			problems = append(problems, FieldError{field + ".password", field + ": a username requires a password or passwordPath"})
		}
		return problems
	case cc.Token != "":
		return nil
	default:
		return []FieldError{{field, field + ": set either username and password (or passwordPath), or token (or tokenPath)"}}
	}
}

func (s *Spec) ControlPlaneTLS() *TLS {
	if s.ControlPlane == nil {
		return nil
	}
	if s.ControlPlane.TLS != nil {
		return s.ControlPlane.TLS
	}
	return s.TLS
}

func (s *Spec) ClusterTLS() *TLS {
	if s.Cluster == nil {
		return nil
	}
	if s.Cluster.TLS != nil {
		return s.Cluster.TLS
	}
	return s.TLS
}

func (s *Spec) InferenceTLS() *TLS {
	if s.Cluster == nil {
		return nil
	}
	if s.Cluster.Inference != nil && s.Cluster.Inference.WildcardTLS != nil {
		return s.Cluster.Inference.WildcardTLS
	}
	return s.ClusterTLS()
}

// SubdomainSupportHelmKey is the cluster chart value that toggles workload
// routing: true = host (subdomain) routing, false = path routing. It is the
// single source of truth shared by the plan (which emits it), the resolver
// below, and the input check that guards against contradicting overrides.
const SubdomainSupportHelmKey = "clusterConfig.global.subdomainSupport"

// SubdomainOverride reports the workload-routing intent expressed through the
// cluster chart's subdomainSupport custom helm value. host is the parsed
// boolean (true = host/subdomain routing, false = path); present is whether the
// key was set at all; valid is whether its value parsed as a boolean.
func (c *Cluster) SubdomainOverride() (host, present, valid bool) {
	if c == nil {
		return false, false, false
	}
	raw, ok := c.CustomHelmValues[SubdomainSupportHelmKey]
	if !ok {
		return false, false, false
	}
	b, err := strconv.ParseBool(strings.TrimSpace(raw))
	if err != nil {
		return false, true, false
	}
	return b, true, true
}

// HostRouting resolves the effective workload routing mode, defaulting to host
// (subdomain) routing. An explicit WorkloadRouting wins; otherwise a
// subdomainSupport custom helm value is honored; otherwise host is the default.
func (c *Cluster) HostRouting() bool {
	if c == nil {
		return true
	}
	switch c.WorkloadRouting {
	case pkgvalidation.WorkloadRoutingHost:
		return true
	case pkgvalidation.WorkloadRoutingPath:
		return false
	}
	if host, present, valid := c.SubdomainOverride(); present && valid {
		return host
	}
	return true
}

// Inference configures the wildcard domain used for inference workloads. The
// wildcard certificate is optional: when WildcardTLS is nil, the cluster's own
// certificate (TLS) is reused for the wildcard domain.
type Inference struct {
	WildcardDomain string
	WildcardTLS    *TLS
}

type DependencyMode string

const (
	DependencyModeInstall DependencyMode = "install"
	DependencyModeSkip    DependencyMode = "skip"
)

// OLM Subscription upgrade strategies, as accepted in
// Dependency.OpenShiftInstallPlanApproval. The zero value ("") means the
// default (manual).
const (
	OLMApprovalManual    = "manual"
	OLMApprovalAutomatic = "automatic"
)

// Dependency is a single prerequisite or feature. Version is optional: when
// empty, it is filled from the compatibility matrix at plan time.
type Dependency struct {
	Mode             DependencyMode
	Version          string
	CustomHelmValues map[string]string
	// OpenShiftInstallPlanApproval sets the OLM Subscription upgrade strategy
	// for an OperatorHub install: OLMApprovalManual (the default) or
	// OLMApprovalAutomatic. Empty means the default. Ignored on Helm installs.
	OpenShiftInstallPlanApproval string
	// OpenShiftCR is the structured override merged onto the operator's primary
	// OpenShift custom resource spec when the dependency installs via
	// OperatorHub (OLM) — e.g. ClusterPolicy for the GPU operator. It is the
	// OpenShift counterpart of CustomHelmValues and is ignored on Helm installs.
	OpenShiftCR map[string]any
	// OpenShiftSubscriptionEnv sets environment variables on the operator pod
	// via the OLM Subscription's spec.config, for operator-level settings that
	// have no custom resource (e.g. the NIM operator's ENABLE_WEBHOOKS). Ignored
	// on Helm installs.
	OpenShiftSubscriptionEnv map[string]string
}

// IsInstall reports whether the dependency is present and set to install.
func (d *Dependency) IsInstall() bool {
	return d != nil && d.Mode == DependencyModeInstall
}

type GPUType string

const (
	GPUTypeReal GPUType = "real"
	GPUTypeFake GPUType = "fake"
)

// Prerequisites are the third-party infra components Run:ai depends on. Skip
// short-circuits the whole group: when set, no prerequisite step is planned,
// regardless of the individual dependencies.
type Prerequisites struct {
	Skip            bool
	Prometheus      *Dependency
	GPUOperator     *Dependency
	GPUType         GPUType
	NetworkOperator *Dependency
}

// Features are optional Run:ai add-ons sharing the Dependency shape. Skip
// short-circuits the whole group: when set, no feature step is planned,
// regardless of the individual dependencies.
type Features struct {
	Skip        bool
	Knative     *Dependency
	Kubeflow    *Dependency
	MPIOperator *Dependency
	LWS         *Dependency
	NIMOperator *Dependency
	Grove       *Dependency
	Dynamo      *Dependency
}

func (s *Spec) IsAirGapped() bool {
	return s.Connectivity == ConnectivityAirGapped
}

// UsesAirgapLoad reports whether installation uses the artifact bundle flow.
func (s *Spec) UsesAirgapLoad() bool {
	if !s.IsAirGapped() {
		return false
	}
	version := ""
	if s.ControlPlane != nil {
		version = s.ControlPlane.Version
	} else if s.Cluster != nil {
		version = s.Cluster.Version
	}
	parsed, err := semver.NewVersion(strings.TrimSpace(version))
	if err != nil {
		return s.DevMode
	}
	if parsed.Major() >= 3 {
		return true
	}
	return s.DevMode && parsed.Major() == 0
}

// IsOpenShift reports whether the target distribution is OpenShift.
func (s *Spec) IsOpenShift() bool {
	return s.KubernetesDistribution == pkgvalidation.KubernetesDistributionOpenShift
}

// GPUOperatorUsesOLM is the single source of truth for whether the GPU operator
// installs via OperatorHub (OLM) rather than Helm: the real operator on
// OpenShift. The fake (simulated) operator and every non-OpenShift install use
// Helm. The planner uses it to route the install; input validation uses it to
// decide which customization surface (customHelmValues vs openshift.customResource)
// is valid.
func (s *Spec) GPUOperatorUsesOLM() bool {
	if !s.IsOpenShift() || s.Prerequisites == nil {
		return false
	}

	return s.Prerequisites.GPUOperator.IsInstall() && s.Prerequisites.GPUType == GPUTypeReal
}

// NetworkOperatorUsesOLM reports whether the network operator installs via
// OperatorHub: any time it is requested on OpenShift (unlike GPU there is no
// fake variant). Off OpenShift it installs from the NGC Helm chart.
func (s *Spec) NetworkOperatorUsesOLM() bool {
	return s.IsOpenShift() && s.Prerequisites != nil && s.Prerequisites.NetworkOperator.IsInstall()
}

// KnativeUsesOLM reports whether Knative installs via OperatorHub: on OpenShift
// it installs through the Red Hat OpenShift Serverless operator (the upstream
// Knative operator + Kourier/Ingress front-door is unsupported there). Off
// OpenShift it installs from the upstream Knative operator Helm flow.
func (s *Spec) KnativeUsesOLM() bool {
	return s.IsOpenShift() && s.Features != nil && s.Features.Knative.IsInstall()
}

func (s *Spec) Validate() error {
	if s == nil {
		return errors.New("spec is required")
	}
	if s.Command == "" {
		return errors.New("command is required")
	}
	if !s.Command.IsKnown() {
		return fmt.Errorf("unknown command: %q", s.Command)
	}
	if err := s.Output.validate(); err != nil {
		return err
	}
	if err := s.OnStepFailure.validate(); err != nil {
		return err
	}
	if err := s.Connectivity.validate(); err != nil {
		return err
	}
	if err := s.Env.validate(); err != nil {
		return err
	}
	switch s.Command {
	case CommandInstall:
		return s.validateInstall()
	case CommandUpgrade:
		return s.validateUpgrade()
	case CommandUninstall:
		return s.validateUninstall()
	case CommandAirgapPack:
		return s.AirgapPack.validate()
	case CommandAirgapLoad:
		return s.AirgapLoad.validate()
	}
	return nil
}

func (s *Spec) validateInstall() error {
	if s.ControlPlane == nil && s.Cluster == nil {
		return errors.New("at least one of control plane or cluster must be enabled")
	}
	if s.ControlPlane != nil && s.ControlPlane.Domain == "" {
		return errors.New("control plane domain is required")
	}
	if s.Cluster != nil && s.Cluster.Name == "" {
		return errors.New("cluster name is required")
	}
	if err := s.validateExternalControlPlane(); err != nil {
		return err
	}
	if s.SelfSignedTLSEnabled() {
		if len(s.SelfSignedTLS.ConflictingFields) > 0 {
			return fmt.Errorf("tls.generateSelfSigned cannot be combined with %s", strings.Join(s.SelfSignedTLS.ConflictingFields, ", "))
		}
		if reason := s.SelfSignedTLSUnsupportedReason(); reason != "" {
			return errors.New(reason)
		}
		if len(s.SelfSignedTLSDNSNames()) == 0 {
			return errors.New("tls.generateSelfSigned requires a control-plane or cluster domain")
		}
	}
	if s.IsAirGapped() {
		return s.Airgap.validate()
	}
	return s.Registry.validate()
}

func (s *Spec) validateExternalControlPlane() error {
	if problems := s.ExternalControlPlaneProblems(); len(problems) > 0 {
		return problems[0]
	}
	return nil
}

func (s *Spec) validateUpgrade() error {
	return nil
}

func (s *Spec) validateUninstall() error {
	return nil
}

func (o Output) validate() error {
	switch o.Format {
	case OutputFormatNone, OutputFormatJSON, OutputFormatText:
		return nil
	}
	return fmt.Errorf("unknown output format: %q", o.Format)
}

func (p FailurePolicy) validate() error {
	switch p {
	case FailurePolicyAbort, FailurePolicySkip:
		return nil
	}
	return fmt.Errorf("unknown failure policy: %q", p)
}

func (c Connectivity) validate() error {
	switch c {
	case "", ConnectivityConnected, ConnectivityAirGapped:
		return nil
	}
	return fmt.Errorf("unknown connectivity: %q", c)
}

func (e Env) validate() error {
	switch e {
	case "", EnvProd, EnvStaging, EnvLab, EnvTest:
		return nil
	}
	return fmt.Errorf("unknown env: %q", e)
}

func (r *Registry) validate() error {
	if r == nil {
		return errors.New("registry is required (jfrog or ngc)")
	}
	switch r.Type {
	case RegistryTypeJFrog:
		if r.Token == "" {
			return errors.New("registry token is required")
		}
	case RegistryTypeNGC:
	default:
		return fmt.Errorf("unknown registry type: %q", r.Type)
	}
	return nil
}

const ngcTokenUser = "$oauthtoken"

const ngcImageHost = "nvcr.io"

// ImagePullConfig returns an encoded pull secret when credentials exist.
func (r *Registry) ImagePullConfig() (string, bool) {
	if r == nil {
		return "", false
	}
	if r.Type == RegistryTypeJFrog {
		return r.Token, r.Token != ""
	}
	username, password := r.Username, r.Password
	if username == "" {
		username, password = r.ChartCredentials()
	}
	if username == "" {
		return "", false
	}
	host := r.Host
	if host == "" {
		host = ngcImageHost
	}
	if i := strings.Index(host, "://"); i >= 0 {
		host = host[i+3:]
	}
	host = strings.SplitN(host, "/", 2)[0]
	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	dockerConfig := map[string]any{
		"auths": map[string]any{
			host: map[string]any{
				"username": username,
				"password": password,
				"auth":     auth,
			},
		},
	}
	raw, err := json.Marshal(dockerConfig)
	if err != nil {
		return "", false
	}
	return base64.StdEncoding.EncodeToString(raw), true
}

// ChartCredentials returns NGC chart credentials.
func (r *Registry) ChartCredentials() (username, password string) {
	if r == nil || r.Type != RegistryTypeNGC {
		return "", ""
	}
	if r.Username != "" {
		return r.Username, r.Password
	}
	if r.Token != "" {
		return ngcTokenUser, r.Token
	}
	return "", ""
}

func (a *Airgap) validate() error {
	if a == nil {
		return errors.New("airgap bundle path is required for air-gapped installation")
	}
	if a.WorkingDir == "" {
		return errors.New("airgap working directory is required for air-gapped installation")
	}
	// When prepare is skipped the bundle is not touched (it may even be gone);
	// the charts install from the already-extracted working dir.
	if !a.SkipPrepare && a.BundlePath == "" {
		return errors.New("airgap bundle path is required for air-gapped installation")
	}
	return nil
}
