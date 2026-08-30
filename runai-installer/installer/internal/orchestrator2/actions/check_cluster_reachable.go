package actions

import (
	"context"
	"fmt"

	semver "github.com/Masterminds/semver/v3"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
)

const minimumKubernetesVersion = "1.30.0"

var minimumSupportedKubernetesVersion = semver.MustParse("1.30.0-0")

// CheckClusterReachable confirms the API server is reachable and its Kubernetes
// version is supported. Failed verdicts do not prevent other preflight checks.
type CheckClusterReachable struct{}

func (a CheckClusterReachable) Kind() string { return "preflight.cluster_reachable" }

func (a CheckClusterReachable) Args() map[string]any { return map[string]any{} }

func (a CheckClusterReachable) Describe() string {
	return "verify the Kubernetes API server is reachable and supported"
}

func (a CheckClusterReachable) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	version, err := deps.ClusterProbe.ServerVersion(ctx)
	if err != nil {
		return orchestrator2.ActionResult{
			Status:      orchestrator2.StatusFail,
			Reason:      fmt.Sprintf("could not reach the Kubernetes API server: %v", err),
			Remediation: "check the kubeconfig/context and that the API server is reachable from this machine",
		}, nil
	}

	data := map[string]any{"kubernetesVersion": version}
	parsedVersion, err := semver.NewVersion(version)
	if err != nil {
		return orchestrator2.ActionResult{
			Status:      orchestrator2.StatusFail,
			Reason:      fmt.Sprintf("could not parse Kubernetes server version %q: %v", version, err),
			Remediation: "verify the cluster reports a valid semantic Kubernetes version",
			Data:        data,
		}, nil
	}
	if parsedVersion.LessThan(minimumSupportedKubernetesVersion) {
		return orchestrator2.ActionResult{
			Status: orchestrator2.StatusFail,
			Reason: fmt.Sprintf("Kubernetes %s is unsupported; minimum required version is %s",
				version, minimumKubernetesVersion),
			Remediation: fmt.Sprintf("upgrade the cluster to Kubernetes %s or newer", minimumKubernetesVersion),
			Data:        data,
		}, nil
	}

	return orchestrator2.ActionResult{Data: data}, nil
}
