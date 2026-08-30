package orchestrator2

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/artifacts"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Deps struct {
	K8sExecutor         K8sExecutor
	HelmInstaller       step.HelmInstaller
	RunAIClientProvider step.RunAIAPIClientProvider
	// ClusterProbe is the read-only cluster capability used by preflight checks.
	// It is optional: nil when no preflight runs (e.g. component tests that don't
	// exercise checks).
	ClusterProbe ClusterProbe
	// RegistryProber is the read-only registry capability used by the registry
	// preflight check. Optional: when nil the registry check is not assembled.
	RegistryProber RegistryProber
	// TLSProber is the read-only TLS capability used by the control-plane
	// certificate preflight check. Optional: when nil that check is not
	// assembled.
	TLSProber TLSProber
	// NodeLabeler labels worker nodes. It is optional: only the fake-GPU
	// prerequisite uses it.
	NodeLabeler NodeLabeler
	// ManifestProvider fetches and renders raw YAML / kustomize manifests. It is
	// optional: only the non-Helm prerequisites (Kubeflow training operator, MPI
	// operator) use it.
	ManifestProvider ManifestProvider
	// Progress reports a long-running action's latest output lines. Optional.
	Progress ProgressFunc
	// ImagePuller downloads container images into OCI layouts. Optional: only
	// the airgap pack command uses it.
	ImagePuller artifacts.ImagePuller
	// ChartFetcher downloads chart archives. Optional: only the airgap pack
	// command uses it.
	ChartFetcher artifacts.ChartFetcher
	CLIFetcher   artifacts.CLIFetcher
	// ImagePusher uploads bundled images to a registry. Optional: only the
	// airgap load command uses it.
	ImagePusher artifacts.ImagePusher
}

// ProgressFunc receives the most recent output lines of a long-running action.
type ProgressFunc func(lines []string)

func ProductionDeps(c client.Client, cfg *rest.Config, controlPlaneVersion string) (Deps, error) {
	probe, err := NewClusterProbe(cfg)
	if err != nil {
		return Deps{}, err
	}
	labeler, err := NewNodeLabeler(cfg)
	if err != nil {
		return Deps{}, err
	}
	k8sExecutor := NewK8sExecutor(c)
	return Deps{
		K8sExecutor:         k8sExecutor,
		HelmInstaller:       step.NewHelmInstaller(),
		RunAIClientProvider: newRunAIClientProvider(k8sExecutor, controlPlaneVersion),
		ClusterProbe:        probe,
		RegistryProber:      NewRegistryProber(),
		TLSProber:           NewTLSProber(),
		NodeLabeler:         labeler,
		ManifestProvider:    NewManifestProvider(),
	}, nil
}
