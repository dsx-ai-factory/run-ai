package installconfig

import (
	"fmt"

	"github.com/run-ai/runai/runai-installer/installer/internal/artifacts"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
)

// RunOptions are the CLI-flag-derived knobs that map into the orchestrator spec.
// Cross-cutting execution knobs (timeouts, confirmation) live in
// specrun.ExecOptions, not here.
type RunOptions struct {
	DryRun        bool
	PreflightOnly bool // accepted; not yet modeled in the plan (TODO)
	SkipPreflight bool // accepted; not yet modeled in the plan (TODO)
	Yes           bool
	Output        spec.OutputFormat // "" => auto-resolved by specrun via TTY
	EventsFile    string
	StatusFile    string // accepted; status writer deferred (TODO)
	OnFailure     spec.FailurePolicy
	Kubeconfig    string // accepted; override plumbing deferred (TODO)
	Env           spec.Env
	Registry      *spec.Registry
	// DevMode relaxes input validation (e.g. allows an out-of-matrix Run:ai
	// version). Escape hatch for development/testing.
	DevMode bool
}

// ToSpec maps the config file plus the run options into the orchestrator spec.
func (c *Config) ToSpec(opts RunOptions) *spec.Spec {
	s := &spec.Spec{
		Command:                spec.CommandInstall,
		InstallMode:            toSpecInstallMode(c.Install.Mode),
		DryRun:                 opts.DryRun,
		NonInteractive:         opts.Yes,
		DevMode:                opts.DevMode,
		OnStepFailure:          opts.OnFailure,
		Connectivity:           spec.Connectivity(c.Install.Connectivity),
		Env:                    opts.Env,
		KubernetesDistribution: toSpecKubernetesDistribution(c.Install.KubernetesDistribution),
		Output: spec.Output{
			Format:     opts.Output,
			EventsFile: opts.EventsFile,
		},
		Registry:      opts.Registry,
		TLS:           toSpecTLS(c.TLS),
		SelfSignedTLS: toSpecSelfSignedTLS(c),
	}

	if c.Airgap != nil {
		valuesDir := c.Airgap.ValuesDir
		if valuesDir == "" {
			valuesDir = c.Airgap.WorkingDir
		}
		s.Airgap = &spec.Airgap{
			BundlePath:  c.Airgap.BundlePath,
			WorkingDir:  c.Airgap.WorkingDir,
			ValuesDir:   valuesDir,
			SkipPrepare: c.Airgap.SkipPrepare,
		}
	}

	if c.ControlPlane != nil && c.ControlPlane.Enabled {
		s.ControlPlane = &spec.ControlPlane{
			Domain:           c.ControlPlane.Domain,
			Version:          c.ControlPlane.Version,
			TLS:              toSpecTLS(c.ControlPlane.TLS),
			CustomHelmValues: toStringMap(c.ControlPlane.CustomHelmValues),
		}
	}

	if c.Cluster != nil && c.Cluster.Enabled {
		s.Cluster = &spec.Cluster{
			Name:             c.Cluster.Name,
			Domain:           c.Cluster.Domain,
			Version:          c.Cluster.Version,
			WorkloadRouting:  c.Cluster.WorkloadRouting,
			TLS:              toSpecTLS(c.Cluster.TLS),
			Inference:        toSpecInference(c.Cluster.Inference),
			CustomHelmValues: toStringMap(c.Cluster.CustomHelmValues),
		}
	}

	if s.UsesAirgapLoad() && s.Airgap != nil && s.Registry != nil {
		destination := artifacts.NormalizeTarget(s.Registry.Host)
		s.AirgapLoad = &spec.AirgapLoad{
			BundlePath:  s.Airgap.BundlePath,
			WorkingDir:  s.Airgap.WorkingDir,
			ValuesDir:   s.Airgap.ValuesDir,
			Destination: destination,
			Credential: spec.RegistryCredential{
				Prefix:             destination,
				Username:           s.Registry.Username,
				Password:           s.Registry.Password,
				CAChainPath:        s.Registry.CAChainPath,
				Insecure:           s.Registry.Insecure,
				InsecureSkipVerify: s.Registry.InsecureSkipVerify,
			},
		}
	}

	if c.Prerequisites != nil {
		gpu := c.Prerequisites.GPUOperator
		s.Prerequisites = &spec.Prerequisites{
			Skip:            c.Prerequisites.Skip,
			Prometheus:      toSpecDependency(c.Prerequisites.Prometheus, true),
			GPUOperator:     toSpecDependency(gpu, true),
			GPUType:         toSpecGPUType(gpu),
			NetworkOperator: toSpecDependency(c.Prerequisites.NetworkOperator, true),
		}
	}

	if c.Features != nil {
		s.Features = &spec.Features{
			Skip:        c.Features.Skip,
			Knative:     toSpecDependency(c.Features.Knative, false),
			Kubeflow:    toSpecDependency(c.Features.Kubeflow, false),
			MPIOperator: toSpecDependency(c.Features.MPIOperator, false),
			LWS:         toSpecDependency(c.Features.LWS, false),
			NIMOperator: toSpecDependency(c.Features.NIMOperator, false),
			Grove:       toSpecDependency(c.Features.Grove, false),
			Dynamo:      toSpecDependency(c.Features.Dynamo, false),
		}
	}

	s.ExternalConnection = toSpecExternalConnection(c.ExternalConnection)

	s.SystemNodes = toSpecSystemNodes(c.SystemNodes)

	s.OpenShift = toSpecOpenShift(c.OpenShift)

	return s
}

func toSpecOpenShift(o *OpenShift) *spec.OpenShift {
	if o == nil || o.OLMCatalog == nil {
		return nil
	}

	return &spec.OpenShift{
		OLMCatalogSource:          o.OLMCatalog.Source,
		OLMCatalogSourceNamespace: o.OLMCatalog.SourceNamespace,
	}
}

const defaultSystemNodeLabel = "node-role.kubernetes.io/runai-system"

// toSpecSystemNodes maps the system-node placement config to the spec, applying
// defaults (mode=preferred, manageLabels=true, label=runai-system). A nil or
// disabled block leaves the feature off (nil spec).
func toSpecSystemNodes(c *SystemNodes) *spec.SystemNodes {
	if c == nil || !c.Enabled {
		return nil
	}
	mode := spec.SystemNodeMode(c.Mode)
	if mode == "" {
		mode = spec.SystemNodeModePreferred
	}
	manage := true
	if c.ManageLabels != nil {
		manage = *c.ManageLabels
	}
	label := c.Label
	if label == "" {
		label = defaultSystemNodeLabel
	}
	return &spec.SystemNodes{
		Mode:         mode,
		Label:        label,
		ManageLabels: manage,
		Nodes:        c.Nodes,
		Selector:     c.Selector,
	}
}

// toSpecInstallMode maps the config's install.mode string to the spec enum,
// defaulting an empty value to full installation.
func toSpecInstallMode(mode string) spec.InstallMode {
	if mode == "" {
		return spec.InstallModeFull
	}
	return spec.InstallMode(mode)
}

func toSpecKubernetesDistribution(d string) string {
	if d == "vanilla" {
		return ""
	}
	return d
}

// toSpecDependency maps a config dependency to the spec. A nil block is left
// unset (no implicit enable). When the block is present but its mode is omitted,
// defaultOn decides: mandatory dependencies default to install, optional ones to
// skip.
func toSpecDependency(d *Dependency, defaultOn bool) *spec.Dependency {
	if d == nil {
		return nil
	}
	mode := spec.DependencyMode(d.Mode)
	if mode == "" {
		mode = spec.DependencyModeSkip
		if defaultOn {
			mode = spec.DependencyModeInstall
		}
	}
	out := &spec.Dependency{
		Mode:             mode,
		Version:          d.Version,
		CustomHelmValues: toStringMap(d.CustomHelmValues),
	}
	if d.OpenShift != nil {
		out.OpenShiftInstallPlanApproval = d.OpenShift.InstallPlanApproval
		out.OpenShiftCR = d.OpenShift.CustomResource
		if d.OpenShift.SubscriptionConfig != nil {
			out.OpenShiftSubscriptionEnv = d.OpenShift.SubscriptionConfig.Env
		}
	}

	return out
}

// toSpecGPUType reads the GPU operator flavor, defaulting to real when unset.
// An invalid value is passed through unchanged so input validation can reject it
// rather than silently treating it as real.
func toSpecGPUType(d *Dependency) spec.GPUType {
	if d == nil || d.Type == "" {
		return spec.GPUTypeReal
	}
	return spec.GPUType(d.Type)
}

func toSpecInference(i *Inference) *spec.Inference {
	if i == nil {
		return nil
	}
	return &spec.Inference{
		WildcardDomain: i.WildcardDomain,
		WildcardTLS:    toSpecTLS(i.WildcardTLS),
	}
}

func toSpecExternalConnection(e *ExternalConnection) *spec.ExternalConnection {
	if e == nil {
		return nil
	}
	out := &spec.ExternalConnection{Type: spec.ExternalConnectionType(e.Type)}
	if e.Ingress != nil {
		out.Ingress = &spec.IngressConnection{ClassName: e.Ingress.ClassName}
	}
	if e.GatewayAPI != nil {
		out.GatewayAPI = &spec.GatewayAPIConnection{Name: e.GatewayAPI.Name, Namespace: e.GatewayAPI.Namespace}
	}
	return out
}

func toSpecTLS(t *TLS) *spec.TLS {
	if t == nil {
		return nil
	}
	return &spec.TLS{CertPath: t.CertPath, KeyPath: t.KeyPath, CAChainPath: t.CAChainPath}
}

func toStringMap(m map[string]any) map[string]string {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = fmt.Sprintf("%v", v)
	}
	return out
}
