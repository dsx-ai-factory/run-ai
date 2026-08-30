package services_crud

import (
	"context"
	"net/http"

	authorization_client "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"

	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/dao_crud"
	"github.com/run-ai/runai/control-plane/workloads/pkg/services_integration"
	"github.com/run-ai/runai/control-plane/workloads/pkg/services_utils"
)

func GetWorkloadById(ctx context.Context, workloadType string, workloadId string) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	permittedScopes, scopesErr := authorization_client.ParsePermittedScopes(ctx)
	if scopesErr != nil {
		permittedScopes = nil
	}

	workload, dbErr := dao_crud.DbGetWorkload(ctx, tenantId, workloadType, workloadId, permittedScopes)
	if dbErr != nil {
		if dbErr.Code != http.StatusNotFound {
			return api.Response(int(dbErr.Code), dbErr), nil
		}

		// if not found, attempt to load it from the snapshot table, assuming it might be a v2-translated workload
		workloadSnapshot, dbErr := dao_crud.DbGetWorkloadV2Snapshot(ctx, tenantId, workloadType, workloadId, permittedScopes)
		if dbErr != nil {
			return api.Response(int(dbErr.Code), dbErr), nil
		}
		workload = workloadSnapshot.ToWorkload()
	}

	actualPhase, phaseErr := services_integration.GetPhaseOfWorkload(ctx, &workload.ClusterId, workloadId)
	if phaseErr != nil {
		return api.Response(int(phaseErr.Code), phaseErr), nil
	}

	response := WorkloadFromDbRecord(workload, workloadsApi.Phase(actualPhase))

	// Apply sanitization if the requesting user is not the workload creator
	sanitizeWorkloadResponse(ctx, &response)

	return api.Response(http.StatusOK, response), nil
}
