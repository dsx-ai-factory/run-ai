package services_crud

import (
	"context"
	"net/http"

	clients "github.com/run-ai/runai/control-plane/workloads/pkg/clients/backend"

	"github.com/run-ai/runai/runai-application-packages/jsonutils/gojsonq"

	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	workloadApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

func CreateWorkspace(ctx context.Context, tenantSettingsClient clients.TenantSettingsClient, request api.WorkspaceCreationRequest) (api.ImplResponse, error) {
	spec, jsonErr := gojsonq.LoadJson(request.Spec)
	if jsonErr != nil {
		return api.Response(http.StatusBadRequest, jsonErr), nil
	}

	createWorkloadParams := createWorkloadParams{
		WorkloadType:         Workspace,
		UseGivenNameAsPrefix: request.UseGivenNameAsPrefix,

		Spec: spec,

		TemplateId: request.TemplateId,

		AssetIds: request.Assets,

		Response: workloadApi.SupersetWorkload{
			RequestedName: request.Name,
			ClusterId:     request.ClusterId,
			ProjectId:     request.ProjectId,
		},
	}

	createErr := createWorkload(ctx, tenantSettingsClient, &createWorkloadParams, Workspace)
	if createErr != nil {
		return api.Response(int(createErr.Code), createErr), nil
	}

	return api.Response(http.StatusAccepted, createWorkloadParams.Response), nil
}
