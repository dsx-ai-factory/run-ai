package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	cluster_service_client "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

// ErrClusterTenantMismatch is returned when a cluster does not belong to the
// requested tenant. Callers should map this to 401/403; other errors from
// ValidateClusterBelongsToTenant indicate an upstream service failure.
var ErrClusterTenantMismatch = errors.New("cluster does not belong to tenant")

// ValidateClusterBelongsToTenant verifies that the given cluster UUID belongs to the
// specified tenant by fetching the cluster's metadata from the cluster service and
// comparing its tenant ID. This prevents cross-tenant access when a user supplies
// a cluster ID that does not belong to their tenant.
func ValidateClusterBelongsToTenant(ctx context.Context, clusterId string, tenantId string, clusterClient cluster_service_client.IClusterExternalClient) error {
	cluster, err := clusterClient.GetClusterMinimalDataById(ctx, clusterId)
	if err != nil {
		return fmt.Errorf("failed to validate cluster ownership: %w", err)
	}
	clusterTenantId := strconv.Itoa(int(cluster.TenantId))
	if clusterTenantId != tenantId {
		return fmt.Errorf("cluster %s: %w", clusterId, ErrClusterTenantMismatch)
	}
	return nil
}

func GetClusterName(ctx context.Context, clusterId string, clustersClient cluster_service_client.IClusterExternalClient) (string, error) {
	cluster, err := clustersClient.GetClusterMinimalDataById(ctx, clusterId)
	if err != nil {
		return "", err
	}

	return cluster.Name, nil
}

func GetClusterVersion(ctx context.Context, clusterId uuid.UUID, clustersClient cluster_service_client.IClusterExternalClient) (string, error) {
	cluster, err := clustersClient.GetClusterMinimalDataById(ctx, clusterId.String())
	if err != nil {
		return "", fmt.Errorf("failed to get cluster. error: %v", err.Error())
	}

	version := cluster.Version.Get()

	return *version, nil
}
