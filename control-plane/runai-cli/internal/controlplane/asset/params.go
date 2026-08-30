package asset

import (
	projectservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/project"
)

// ResolveParams holds the parsed parameters for asset resolution.
type ResolveParams struct {
	ProjectID *int32
	ClusterID *string
}

// ParseResolveParams parses and validates the common parameters used for asset resolution.
// This extracts the shared logic from compute and environment packages.
func ParseResolveParams(projectId, clusterId string) (*ResolveParams, error) {
	projectID, err := projectservice.ParseProjectIdInt32(projectId)
	if err != nil {
		return nil, err
	}

	var clusterIDPtr *string
	if clusterId != "" {
		clusterIDPtr = &clusterId
	}

	return &ResolveParams{
		ProjectID: projectID,
		ClusterID: clusterIDPtr,
	}, nil
}
