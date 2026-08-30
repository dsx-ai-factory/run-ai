package policies

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	policyServiceClient "github.com/run-ai/runai/runai-clients/policy-service-client"
	policyService "github.com/run-ai/runai/runai-clients/policy-service-client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const ProjectScope = "project"

type Client struct {
	apiClient policyServiceClient.IPolicyClient
}

func (c Client) GetWorkspacePolicyOfProjectV2(ctx context.Context, projectId int32, options *PolicyClientOptions) (*policyService.WorkspacePolicyV2, *policyService.Error) {
	orgUnit := fmt.Sprintf("project: %d", projectId)
	projId := fmt.Sprintf("%d", projectId)

	resp, r, err := c.apiClient.GetWorkspacePolicy(ctx, ProjectScope, nil, nil, &projId, options != nil && options.NonCreationMode)
	if r != nil {
		if r.StatusCode == http.StatusOK {
			return resp, nil
		}
		if r.StatusCode == http.StatusNotFound {
			return nil, policyService.NewError(http.StatusNotFound, fmt.Sprintf("%s does not exist", orgUnit))
		}
		data, _ := io.ReadAll(r.Body)

		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get workspace policy of %s: %s", orgUnit, data)
		return nil, policyService.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to get workspace policy of %s", orgUnit))
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to request workspace policy of %s: %v", orgUnit, err)
	return nil, policyService.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request workspace policy of %s", orgUnit))
}

func (c Client) GetTrainingPolicyOfProjectV2(ctx context.Context, projectId int32, options *PolicyClientOptions) (*policyService.TrainingPolicyV2, *policyService.Error) {
	orgUnit := fmt.Sprintf("project: %d", projectId)
	projId := fmt.Sprintf("%d", projectId)

	resp, r, err := c.apiClient.GetTrainingPolicy(ctx, ProjectScope, nil, nil, &projId, options != nil && options.NonCreationMode)
	if r != nil {
		if r.StatusCode == http.StatusOK {
			return resp, nil
		}
		if r.StatusCode == http.StatusNotFound {
			return nil, policyService.NewError(http.StatusNotFound, fmt.Sprintf("%s does not exist", orgUnit))
		}

		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get training policy of %s: %d", orgUnit, r.Body)
		return nil, policyService.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to get training policy of %s", orgUnit))
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to request training policy of %s: %v", orgUnit, err)
	return nil, policyService.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request training policy of %s", orgUnit))
}

func (c Client) GetDistributedPolicyOfProjectV2(ctx context.Context, projectId int32, options *PolicyClientOptions) (*policyService.DistributedPolicyV2, *policyService.Error) {
	orgUnit := fmt.Sprintf("project: %d", projectId)
	projId := fmt.Sprintf("%d", projectId)

	resp, r, err := c.apiClient.GetDistributedPolicy(ctx, ProjectScope, nil, nil, &projId, options != nil && options.NonCreationMode)
	if r != nil {
		if r.StatusCode == http.StatusOK {
			return resp, nil
		}
		if r.StatusCode == http.StatusNotFound {
			return nil, policyService.NewError(http.StatusNotFound, fmt.Sprintf("%s does not exist", orgUnit))
		}

		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get distributed policy of %s: %d", orgUnit, r.Body)
		return nil, policyService.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to get distributed policy of %s", orgUnit))
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to request distributed policy of %s: %v", orgUnit, err)
	return nil, policyService.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request distribured policy of %s", orgUnit))
}

func (c Client) GetInferencePolicyOfProjectV2(ctx context.Context, projectId int32, options *PolicyClientOptions) (*policyService.InferencePolicyV2, *policyService.Error) {
	orgUnit := fmt.Sprintf("project: %d", projectId)
	projId := fmt.Sprintf("%d", projectId)

	resp, r, err := c.apiClient.GetInferencePolicy(ctx, ProjectScope, nil, nil, &projId, options != nil && options.NonCreationMode)
	if r != nil {
		if r.StatusCode == http.StatusOK {
			return resp, nil
		}
		if r.StatusCode == http.StatusNotFound {
			return nil, policyService.NewError(http.StatusNotFound, fmt.Sprintf("%s does not exist", orgUnit))
		}
		data, _ := io.ReadAll(r.Body)

		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get inference policy of %s: %s", orgUnit, data)
		return nil, policyService.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to get inference policy of %s", orgUnit))
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to request inference policy of %s: %v", orgUnit, err)
	return nil, policyService.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request inference policy of %s", orgUnit))
}

func (c Client) GetDistributedInferencePolicyOfProjectV2(ctx context.Context, projectId int32, options *PolicyClientOptions) (*policyService.DistributedInferencePolicyV2, *policyService.Error) {
	orgUnit := fmt.Sprintf("project: %d", projectId)
	projId := fmt.Sprintf("%d", projectId)

	resp, r, err := c.apiClient.GetDistributedInferencePolicy(ctx, ProjectScope, nil, nil, &projId, options != nil && options.NonCreationMode)
	if r != nil {
		if r.StatusCode == http.StatusOK {
			return resp, nil
		}
		if r.StatusCode == http.StatusNotFound {
			return nil, policyService.NewError(http.StatusNotFound, fmt.Sprintf("%s does not exist", orgUnit))
		}
		data, _ := io.ReadAll(r.Body)

		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get distributed inference policy of %s: %s", orgUnit, data)
		return nil, policyService.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to get distributed inference policy of %s", orgUnit))
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to request distributed inference policy of %s: %v", orgUnit, err)
	return nil, policyService.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request distributed inference policy of %s", orgUnit))
}

func (c Client) GetNimServicePolicyOfProject(ctx context.Context, projectId int32, options *PolicyClientOptions) (*policyService.NimServicePolicy, *policyService.Error) {
	orgUnit := fmt.Sprintf("project: %d", projectId)
	projId := fmt.Sprintf("%d", projectId)

	resp, r, err := c.apiClient.GetNimServicePolicy(ctx, ProjectScope, nil, nil, &projId, options != nil && options.NonCreationMode)
	if r != nil {
		if r.StatusCode == http.StatusOK {
			return resp, nil
		}
		if r.StatusCode == http.StatusNotFound {
			return nil, policyService.NewError(http.StatusNotFound, fmt.Sprintf("%s does not exist", orgUnit))
		}
		data, _ := io.ReadAll(r.Body)

		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get NIM service policy of %s: %s", orgUnit, data)
		return nil, policyService.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to get NIM service policy of %s", orgUnit))
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to request NUM service policy of %s: %v", orgUnit, err)
	return nil, policyService.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request NIM service policy of %s", orgUnit))
}

func (c Client) NotifyAssetModified(ctx context.Context, assetId string, scopeType policyService.ScopeType, scopeId string) *policyService.Error {
	return c.notifyAssetChanged(ctx, assetId, policyService.ASSETCHANGETYPE_UPDATED, scopeType, scopeId)
}

func (c Client) NotifyAssetDeleted(ctx context.Context, assetId string, scopeType policyService.ScopeType, scopeId string) *policyService.Error {
	return c.notifyAssetChanged(ctx, assetId, policyService.ASSETCHANGETYPE_DELETED, scopeType, scopeId)
}

func (c Client) notifyAssetChanged(ctx context.Context, assetId string, assetChangeType policyService.AssetChangeType,
	scopeType policyService.ScopeType, scopeId string) *policyService.Error {

	r, err := c.apiClient.GetAssetChangedHook(ctx, assetId, &policyService.AssetChangeNotification{
		ChangeType: &assetChangeType,
		ScopeType:  &scopeType,
		ScopeId:    *policyService.NewNullableString(&scopeId),
	})
	if r != nil {
		if r.StatusCode == http.StatusNoContent {
			return nil
		}
		if r.StatusCode == http.StatusBadRequest {
			data, _ := io.ReadAll(r.Body)

			var policyErr *policyService.Error
			jsonErr := json.Unmarshal(data, &policyErr)
			if jsonErr != nil {
				return policyService.NewError(http.StatusBadRequest, string(data))
			}
			return policyService.NewError(http.StatusBadRequest, fmt.Sprintf("%s: %v", policyErr.Message, convutils.Deref(policyErr.Details, "")))
		}

		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to notify asset modification to policy service: %v", err)
		return policyService.NewError(http.StatusServiceUnavailable, "failed to notify asset modification to policy service")
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to request sending asset modification event: %v", err)
	return policyService.NewError(http.StatusServiceUnavailable, "failed to request asset modification event")
}

func (c Client) GetGvkPolicyOfProject(ctx context.Context, projectId string, workloadType policyService.GVK) (*policyService.GetGvkPolicyResp, *policyService.Error) {
	return c.apiClient.GetGvkPolicyOfProject(ctx, projectId, workloadType)
}
