package describe

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec/spectype"
	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

type Lister interface {
	List(ctx context.Context, projectID *int32, clusterID *string, name *string, workloadType *string) (*openapi.EnvironmentListResponse, error)
}

func FindByName(ctx context.Context, lister Lister, name string, projectID *int32, clusterID *string, workloadType string) (*openapi.EnvironmentAsset, error) {
	// Use workloadType filter if provided
	var workloadTypePtr *string
	if workloadType != "" {
		workloadTypePtr = &workloadType
	}

	response, err := lister.List(ctx, projectID, clusterID, &name, workloadTypePtr)
	if err != nil {
		return nil, fmt.Errorf("failed to search for environment: %w", err)
	}

	if len(response.Entries) == 0 {
		if workloadType != "" {
			return nil, fmt.Errorf("the environment '%s' does not support %s workloads", name, spectype.APIStringToDisplayName(workloadType))
		}
		return nil, fmt.Errorf("environment '%s' not found", name)
	}

	return &response.Entries[0], nil
}
