package services_utils

import (
	"context"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
)

const (
	KnativeNotAvailableMsg           = "Inference can't be created because Knative is not installed on the selected cluster. Contact your administrator."
	KnativeFeatureNotAvailableFormat = "Inference can't be created because Knative is not configured to support %s. Contact your administrator."
)

// KnativeFeatures represents the required Knative features for inference
type KnativeFeatures struct {
	ScaleToZero bool
	NodeType    bool
	// Add more features here as needed
}

// CheckKnativeAvailable validates that knative is available in the cluster and has all required features
func CheckKnativeAvailable(ctx context.Context, clusterId string, requiredFeatures KnativeFeatures) *api.Error {
	// Try cached version first
	available, features, apiErr := getKnativeStatus(ctx, clusterId)
	if apiErr != nil {
		return apiErr
	}

	// Check if we would fail based on cached data
	wouldFailOnAvailable := !available
	wouldFailOnScaleToZero := requiredFeatures.ScaleToZero && !features.ScaleToZero
	wouldFailOnNodeType := requiredFeatures.NodeType && !features.NodeType

	// If any check would fail, perform ONE sync request and re-evaluate
	if wouldFailOnAvailable || wouldFailOnScaleToZero || wouldFailOnNodeType {
		availableSync, featuresSync, apiErrSync := getKnativeStatusSync(ctx, clusterId)
		if apiErrSync != nil {
			return apiErrSync
		}

		// Use fresh data for final validation
		available = availableSync
		features = featuresSync
	}

	// Final validation with either cached or fresh data
	if !available {
		return api.NewError(http.StatusBadRequest, KnativeNotAvailableMsg, nil)
	}

	if requiredFeatures.ScaleToZero && !features.ScaleToZero {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf(KnativeFeatureNotAvailableFormat, "scale to zero"), nil)
	}

	if requiredFeatures.NodeType && !features.NodeType {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf(KnativeFeatureNotAvailableFormat, "node type"), nil)
	}

	return nil
}

func getKnativeStatus(ctx context.Context, clusterId string) (bool, KnativeFeatures, *api.Error) {
	clusterChecker := checker.NewClusterCapabilitiesChecker(clusterId)
	config, err := clusterChecker.ClusterConfiguration(ctx)
	if err != nil {
		return false, KnativeFeatures{}, api.NewError(http.StatusServiceUnavailable, err.Error(), nil)
	}

	features := KnativeFeatures{
		ScaleToZero: config.KnativeScaleToZeroAvailable,
		NodeType:    config.KnativeNodeTypeAvailable,
	}

	return config.KnativeAvailable, features, nil
}

// getKnativeStatusSync retrieves Knative status without using cache and stores it in cache
func getKnativeStatusSync(ctx context.Context, clusterId string) (bool, KnativeFeatures, *api.Error) {
	config, err := checker.FetchClusterConfigurationSyncAndStore(ctx, clusterId)
	if err != nil {
		return false, KnativeFeatures{}, api.NewError(http.StatusServiceUnavailable, err.Error(), nil)
	}

	features := KnativeFeatures{
		ScaleToZero: config.KnativeScaleToZeroAvailable,
		NodeType:    config.KnativeNodeTypeAvailable,
	}

	return config.KnativeAvailable, features, nil
}
