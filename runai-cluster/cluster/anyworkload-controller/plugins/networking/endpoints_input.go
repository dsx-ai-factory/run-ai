package networking

import (
	"fmt"

	endpoints "github.com/run-ai/runai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
)

// inputShape identifies which configuration shape drives the plugin; see README "Input shapes".
type inputShape int

const (
	inputNone inputShape = iota
	inputLegacy
	inputAnnotation
)

// Errors are user-visible misconfiguration (malformed annotation, unsupported schemaVersion, or legacy field + annotation set together).
func classifyInput(runner *dromosv1alpha1.WorkloadRunner) (inputShape, *endpoints.Config, error) {
	cfg, err := endpoints.Get(runner)
	if err != nil {
		return inputNone, nil, err
	}
	hasAnnotation := cfg != nil
	hasLegacy := extractNetworkingConfig(runner) != nil

	switch {
	case hasAnnotation && hasLegacy:
		return inputNone, cfg, fmt.Errorf(
			"spec.instructions.networking and the %s annotation are mutually exclusive; "+
				"the networking field is the legacy single-endpoint plugin and endpoints is the multi-endpoint plugin",
			endpoints.Annotation,
		)
	case hasAnnotation:
		return inputAnnotation, cfg, nil
	case hasLegacy:
		return inputLegacy, nil, nil
	default:
		return inputNone, nil, nil
	}
}
