package services_crud

import (
	"context"
	"net/http"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	"github.com/run-ai/runai/control-plane/workloads/pkg/clients/clusterservice"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

// Uses the internal cluster endpoint, which carries no authorization, so the
// submission flow works for regular users who lack the cluster:read permission
// required by the external GetClusterByUuid endpoint.
func buildClusterStorageConfig(ctx context.Context, clusterId string) (composer.ClusterStorageConfig, *workloadsApi.SubmissionErrorV2) {
	cluster, getErr := clusterservice.Get().GetInternalCluster(ctx, clusterId)
	if getErr != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusServiceUnavailable, getErr.Error())
	}
	return composer.NewClusterStorageConfigFromCluster(cluster), nil
}
