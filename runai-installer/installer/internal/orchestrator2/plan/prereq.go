package plan

import (
	"strings"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/compat"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/steps"
)

// buildPrereqFlows resolves the third-party prerequisites Run:ai depends on
// (Prometheus, GPU operator, network operator) into one flow per selected
// dependency. These install BEFORE Run:ai. Each dependency's version is filled
// from the compatibility matrix when the spec leaves it unset. The returned
// component list feeds the compatibility-fallback warning.
func buildPrereqFlows(s *spec.Spec, deps orchestrator2.Deps) ([]*orchestrator2.Flow, []orchestrator2.PrereqComponent) {
	locks, _ := compat.Resolve(runaiVersion(s))

	var flows []*orchestrator2.Flow
	var installed []orchestrator2.PrereqComponent
	add := func(name, title, version string, step orchestrator2.Step) {
		flows = append(flows, flowTitled(name, title, step))
		installed = append(installed, orchestrator2.PrereqComponent{Name: name, Version: version})
	}
	// addMulti registers a prerequisite installed by an ordered sequence of
	// steps sharing one flow (and so counted once in the summary), used by the
	// OpenShift GPU flow (NFD operator + GPU operator + their CRs).
	addMulti := func(name, title, version string, sps ...orchestrator2.Step) {
		flows = append(flows, &orchestrator2.Flow{Name: name, Title: title, Steps: sps})
		installed = append(installed, orchestrator2.PrereqComponent{Name: name, Version: version})
	}

	if p := s.Prerequisites; p != nil && !p.Skip {
		if dep := p.Prometheus; dep.IsInstall() {
			version := locks.VersionFor(compat.Prometheus, dep.Version)
			add("prometheus", "Prometheus", version, steps.NewPrometheusChart(version, dep.CustomHelmValues, deps))
		}
		if dep := p.GPUOperator; dep.IsInstall() {
			switch {
			// The fake (simulated) GPU operator always installs via its Helm
			// chart, on every distribution including OpenShift: it is a simulator
			// with no certified OperatorHub package, and the chart already binds
			// the SCC it needs on OpenShift (environment.openshift=true). This
			// case is matched first so OpenShift never routes the fake operator
			// through OLM.
			case p.GPUType == spec.GPUTypeFake:
				version := locks.VersionFor(compat.GPUOperatorFake, "")
				add("gpu-operator", "NVIDIA GPU Operator", version, steps.NewGPUOperatorFake(version, dep.CustomHelmValues, s.KubernetesDistribution, deps))
			// The real GPU operator on OpenShift installs through OperatorHub
			// (OLM), which also pulls in Node Feature Discovery and binds the
			// SCCs the operator needs — the plain NGC Helm chart does not. The
			// ClusterPolicy CR carries any user override (openshift.customResource).
			// Both OLM channels resolve from the compat matrix: the GPU channel is
			// derived from its version (NVIDIA aligns them), and the NFD channel
			// has its own lock.
			case s.GPUOperatorUsesOLM():
				version := locks.VersionFor(compat.GPUOperator, dep.Version)
				gpuCatalog, nfdCatalog := openShiftOLMCatalogs(s)
				addMulti("gpu-operator", "NVIDIA GPU Operator", version, steps.NewGPUOperatorOpenShift(steps.GPUOpenShiftParams{
					GPUChannel:          olmChannelFromVersion(version),
					GPUVersion:          version,
					NFDChannel:          locks.VersionFor(compat.NFD, ""),
					InstallPlanApproval: olmInstallPlanApproval(dep.OpenShiftInstallPlanApproval),
					ClusterPolicySpec:   dep.OpenShiftCR,
					SubscriptionEnv:     dep.OpenShiftSubscriptionEnv,
					GPUCatalog:          gpuCatalog,
					NFDCatalog:          nfdCatalog,
				}, deps)...)
			// The real GPU operator on every other distribution installs from the
			// NGC Helm chart.
			default:
				version := locks.VersionFor(compat.GPUOperator, dep.Version)
				add("gpu-operator", "NVIDIA GPU Operator", version, steps.NewGPUOperatorReal(version, dep.CustomHelmValues, deps))
			}
		}
		// The Network Operator manages real NICs, independent of GPU flavor, so
		// it installs whenever requested — including alongside a fake GPU
		// operator (e.g. simulated GPUs on a cluster with real RDMA hardware).
		// On OpenShift it installs via OperatorHub (with NFD + SR-IOV); elsewhere
		// from the NGC Helm chart.
		if dep := p.NetworkOperator; dep.IsInstall() {
			version := locks.VersionFor(compat.NetworkOperator, dep.Version)
			if s.NetworkOperatorUsesOLM() {
				addMulti("network-operator", "NVIDIA Network Operator", version, steps.NewNetworkOperatorOpenShift(steps.NetworkOpenShiftParams{
					NetworkChannel:       olmChannelFromVersion(version),
					NetworkVersion:       version,
					NFDChannel:           locks.VersionFor(compat.NFD, ""),
					SRIOVChannel:         locks.VersionFor(compat.SRIOVOperator, ""),
					InstallPlanApproval:  olmInstallPlanApproval(dep.OpenShiftInstallPlanApproval),
					NicClusterPolicySpec: dep.OpenShiftCR,
					SubscriptionEnv:      dep.OpenShiftSubscriptionEnv,
					CertifiedCatalog:     olmCatalog(s, steps.CertifiedOperatorsCatalog()),
					RedHatCatalog:        olmCatalog(s, steps.RedHatOperatorsCatalog()),
				}, deps)...)
			} else {
				add("network-operator", "NVIDIA Network Operator", version, steps.NewNetworkOperator(version, dep.CustomHelmValues, deps))
			}
		}
	}
	return flows, installed
}

// buildFeatureFlows resolves the optional Run:ai features (Knative, Kubeflow,
// MPI, LWS, NIM, Grove, Dynamo) into one flow per selected feature. These
// install AFTER Run:ai: they are capabilities layered on top of the cluster,
// and some (Dynamo, Grove) depend on the scheduler/CRDs the Run:ai cluster
// provides. The returned component list feeds the compatibility-fallback
// warning.
func buildFeatureFlows(s *spec.Spec, deps orchestrator2.Deps) ([]*orchestrator2.Flow, []orchestrator2.PrereqComponent) {
	locks, _ := compat.Resolve(runaiVersion(s))

	var flows []*orchestrator2.Flow
	var installed []orchestrator2.PrereqComponent
	add := func(name, title, version string, step orchestrator2.Step) {
		flows = append(flows, flowTitled(name, title, step))
		installed = append(installed, orchestrator2.PrereqComponent{Name: name, Version: version})
	}
	// addMulti registers a feature installed by an ordered sequence of steps
	// sharing one flow (and so counted once in the summary).
	addMulti := func(name, title, version string, steps ...orchestrator2.Step) {
		flows = append(flows, &orchestrator2.Flow{Name: name, Title: title, Steps: steps})
		installed = append(installed, orchestrator2.PrereqComponent{Name: name, Version: version})
	}

	f := s.Features
	if f == nil || f.Skip {
		return flows, installed
	}

	// Knative Serving is a multi-step install (operator Helm release, the
	// KnativeServing CR gated on its CRD, a fronting Ingress, and the
	// serving-HPA manifest), grouped into one flow so it counts as a single
	// feature. Its wildcard domain/cert come from the cluster's Inference block
	// (enforced by input validation).
	if dep := f.Knative; dep.IsInstall() {
		version := locks.VersionFor(compat.Knative, dep.Version)
		if s.KnativeUsesOLM() {
			// On OpenShift the upstream Knative operator + nginx-class fronting
			// Ingress is unsupported; install through the Red Hat OpenShift
			// Serverless operator and expose serving via OpenShift Routes.
			addMulti("knative", "Knative Serving", version, steps.NewKnativeOpenShift(steps.KnativeOpenShiftParams{
				ServerlessChannel:   locks.VersionFor(compat.Serverless, ""),
				ServerlessVersion:   locks.VersionFor(compat.ServerlessVersion, ""),
				InstallPlanApproval: olmInstallPlanApproval(dep.OpenShiftInstallPlanApproval),
				Cluster:             s.Cluster,
				CRSpec:              dep.OpenShiftCR,
				SubscriptionEnv:     dep.OpenShiftSubscriptionEnv,
				Catalog:             olmCatalog(s, steps.RedHatOperatorsCatalog()),
			}, deps)...)
		} else {
			addMulti("knative", "Knative Serving", version, steps.NewKnativeSteps(version, knativeIngressClass(s), knativeGateway(s), s.Cluster, s.InferenceTLS(), deps)...)
		}
	}
	if dep := f.Kubeflow; dep.IsInstall() {
		version := locks.VersionFor(compat.Kubeflow, dep.Version)
		add("training-operator", "Kubeflow Training Operator", version, steps.NewKubeflowTraining(version, deps))
	}
	// When both operators are requested, the coexistence step runs BETWEEN
	// them: the training operator ships a v1 mpijobs CRD that must be removed
	// before the MPI operator can install its v2beta1 mpijobs CRD (Kubernetes
	// rejects dropping a stored version otherwise). It is not a "component"
	// (no version), so it is appended directly rather than via add() to keep
	// it out of the component list.
	if f.Kubeflow.IsInstall() && f.MPIOperator.IsInstall() {
		flows = append(flows, flowTitled("training-mpi-coexistence", "Training/MPI coexistence", steps.NewMPICoexistence(deps)))
	}
	if dep := f.MPIOperator; dep.IsInstall() {
		version := locks.VersionFor(compat.MPIOperator, dep.Version)
		add("mpi-operator", "MPI Operator", version, steps.NewMPIOperator(version, deps))
	}
	// LWS and the NIM operator are lightweight, non-root controllers whose Helm
	// charts run cleanly under OpenShift's restricted-v2 SCC, so they install
	// from Helm on every distribution. Only operators that genuinely cannot
	// install via Helm on OpenShift (the GPU and network operators, which need
	// driver/SCC/NFD/SR-IOV handling) use the OperatorHub path.
	if dep := f.LWS; dep.IsInstall() {
		version := locks.VersionFor(compat.LWS, dep.Version)
		add("lws", "LeaderWorkerSet", version, steps.NewLWS(version, dep.CustomHelmValues, deps))
	}
	if dep := f.NIMOperator; dep.IsInstall() {
		version := locks.VersionFor(compat.NIMOperator, dep.Version)
		add("nim-operator", "NVIDIA NIM Operator", version, steps.NewNIMOperator(version, dep.CustomHelmValues, deps))
	}
	if dep := f.Grove; dep.IsInstall() {
		version := locks.VersionFor(compat.Grove, dep.Version)
		add("grove", "Grove", version, steps.NewGrove(version, dep.CustomHelmValues, deps))
	}
	// Dynamo installs after Grove: it enables Grove integration
	// (global.grove.enabled=true), so the Grove operator's CRDs must already be
	// present. It also integrates with the Run:ai cluster's KAI scheduler, which
	// is why features install after Run:ai. Input validation enforces Grove too.
	if dep := f.Dynamo; dep.IsInstall() {
		version := locks.VersionFor(compat.Dynamo, dep.Version)
		add("dynamo", "NVIDIA Dynamo", version, steps.NewDynamo(version, dep.CustomHelmValues, deps))
	}

	return flows, installed
}

// compatFallback builds the compatibility-matrix fallback warning: non-nil only
// when the spec pinned a Run:ai version the matrix does not recognize (planning
// then uses the latest known release's versions) and at least one component is
// installed. An unset Run:ai version is not a fallback (latest is the default).
func compatFallback(s *spec.Spec, components []orchestrator2.PrereqComponent) *orchestrator2.CompatFallback {
	runai := runaiVersion(s)
	_, fellBack := compat.Resolve(runai)
	if !fellBack || runai == "" || len(components) == 0 {
		return nil
	}
	return &orchestrator2.CompatFallback{
		RequestedVersion: runai,
		UsingVersion:     compat.LatestVersion(),
		Components:       components,
	}
}

func flowTitled(name, title string, step orchestrator2.Step) *orchestrator2.Flow {
	return &orchestrator2.Flow{Name: name, Title: title, Steps: []orchestrator2.Step{step}}
}

// olmChannelFromVersion derives an OLM subscription channel from a semver chart
// version, e.g. "v25.3.4" -> "v25.3". It backs the GPU operator's OpenShift
// channel: NVIDIA aligns the certified operator's channel with the chart's
// major.minor, so the channel needs no separate compat lock. A version that is
// not major.minor.* falls back to the "stable" channel.
func olmChannelFromVersion(version string) string {
	const majorMinorParts = 2

	parts := strings.Split(strings.TrimPrefix(version, "v"), ".")
	if len(parts) < majorMinorParts {
		return "stable"
	}

	return "v" + parts[0] + "." + parts[1]
}

// olmInstallPlanApproval maps the spec's per-prerequisite approval override to
// the OLM Subscription value. The default (unset) is Manual: it pins the
// operator to the installed (startingCSV) version so OLM does not silently
// upgrade it in the background, matching how the Helm-installed prerequisites
// are version-locked. "automatic" opts back into OLM tracking the channel head.
func olmInstallPlanApproval(specValue string) string {
	if strings.EqualFold(specValue, spec.OLMApprovalAutomatic) {
		return "Automatic"
	}

	return "Manual"
}

// openShiftOLMCatalogs resolves the OperatorHub catalogs the GPU and NFD
// operators are pulled from, applying the air-gap mirror override when set.
func openShiftOLMCatalogs(s *spec.Spec) (gpu, nfd actions.OLMCatalog) {
	return olmCatalog(s, steps.CertifiedOperatorsCatalog()), olmCatalog(s, steps.RedHatOperatorsCatalog())
}

// olmCatalog returns the catalog an operator is pulled from: its public default,
// or the spec's OLM catalog override when set. Air-gapped clusters mirror all
// operators into one internal CatalogSource, so the override (when present)
// replaces every default, keeping its namespace when the override omits one.
func olmCatalog(s *spec.Spec, def actions.OLMCatalog) actions.OLMCatalog {
	oc := s.OpenShift
	if oc == nil || oc.OLMCatalogSource == "" {
		return def
	}

	namespace := oc.OLMCatalogSourceNamespace
	if namespace == "" {
		namespace = def.SourceNamespace
	}

	return actions.OLMCatalog{Source: oc.OLMCatalogSource, SourceNamespace: namespace}
}

// runaiVersion is the Run:ai release that drives the compatibility matrix: the
// control-plane version, or the cluster version when only a cluster installs.
func runaiVersion(s *spec.Spec) string {
	if s.ControlPlane != nil && s.ControlPlane.Version != "" {
		return s.ControlPlane.Version
	}
	if s.Cluster != nil {
		return s.Cluster.Version
	}
	return ""
}
