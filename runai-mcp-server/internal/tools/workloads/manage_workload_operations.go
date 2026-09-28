/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	workloadsapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	workloadsv2 "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

const (
	workloadLifecycleActionStop    = "stop"
	workloadLifecycleActionRestart = "restart"
	workloadLifecycleActionUpdate  = "update"
	workloadLifecycleActionDelete  = "delete"
	workloadLifecycleDeleteWarning = "deleting a workload is permanent"
)

type manageWorkloadLifecycleInput struct {
	Action        string         `json:"action"                  jsonschema:"Workload operation to perform. One of: stop, restart, update, delete. Stop and restart require suspend support. Update is supported only through the native Run:ai Inference and DistributedInference APIs, including workloads backed by AnyWorkload; it is not supported for direct source=yaml submissions."` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
	WorkloadID    string         `json:"workloadId"              jsonschema:"Workload ID (UUID). If unknown, use list_workloads to resolve it."`
	UpdateRequest map[string]any `json:"updateRequest,omitempty" jsonschema:"Update body containing a non-empty spec. Required only when action=update. Supported only through the native Run:ai Inference and DistributedInference APIs, including workloads backed by AnyWorkload; direct source=yaml submissions cannot be updated."` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
	Confirm       bool           `json:"confirm,omitempty"       jsonschema:"Required for every operation when the client cannot prompt; otherwise returns confirmation_required before changing the workload."`
}

func (in manageWorkloadLifecycleInput) Validate() error {
	switch in.Action {
	case workloadLifecycleActionUpdate:
		if len(in.UpdateRequest) == 0 {
			return fmt.Errorf("updateRequest is required when action=update")
		}
		if err := validateUpdateRequestEnvelope(in.UpdateRequest); err != nil {
			return err
		}
	case workloadLifecycleActionStop, workloadLifecycleActionRestart, workloadLifecycleActionDelete:
		if len(in.UpdateRequest) != 0 {
			return fmt.Errorf("updateRequest is supported only when action=update")
		}
	default:
		return fmt.Errorf("action must be one of: stop, restart, update, delete")
	}
	return common.ValidateUUID("workloadId", in.WorkloadID)
}

type manageWorkloadLifecycleOutput struct {
	Action     string `json:"action"     jsonschema:"The workload operation requested: stop, restart, update, or delete."`
	WorkloadID string `json:"workloadId" jsonschema:"Workload ID."`
	Applied    bool   `json:"applied"    jsonschema:"True when the upstream service accepted the requested action. For update, this does not mean cluster reconciliation has completed."`
}

func registerManageWorkloadOperations(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[manageWorkloadLifecycleInput, manageWorkloadLifecycleOutput]{
		Name: "manage_workload_operations",
		Description: "Call when the user asks to stop, restart, update, or permanently delete one existing workload. " +
			"Delete can target workloads submitted through a native Run:ai workload API or directly from a Kubernetes YAML manifest. " +
			"Stop suspends the workload and restart resumes it; for both native and direct-YAML submissions, these operations are supported only when the resolved workload reports actionsSupport.suspend=true. Delete permanently removes the workload. " + //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
			"A known Kubernetes issue can cause batch job workload types (such as Job, PyTorchJob, and RayJob) to get stuck in the Resuming state. The issue is fixed in Kubernetes v1.32.9, v1.33.6, v1.34.2, and v1.35.0 (or newer). Upgrade to a fixed Kubernetes release to resolve this. " + //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
			"Update is supported only through the native Run:ai Inference or DistributedInference API and requires updateRequest with a non-empty spec. This includes native-API submissions internally backed by AnyWorkload/Workloads V2, but not workloads submitted directly with source=yaml. " + //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
			"An accepted update is asynchronous; use get_workload_status or get_workload_spec afterward to verify the reconciled state. " +
			"If workloadId is unknown, use list_workloads to resolve it. Every operation requires confirmation. Use get_workload_status to inspect state without changing it.",
		ReadOnly:    false,
		Destructive: true,
		Handler: func(ctx context.Context, in manageWorkloadLifecycleInput) (manageWorkloadLifecycleOutput, error) {
			return manageWorkloadLifecycleRun(ctx, client, in)
		},
	})
}

func manageWorkloadLifecycleRun(ctx context.Context, client *runaiclient.Client, in manageWorkloadLifecycleInput) (manageWorkloadLifecycleOutput, error) {
	// Return the requested operation details with confirmation errors.
	out := manageWorkloadLifecycleOutput{Action: in.Action, WorkloadID: in.WorkloadID}
	workload, err := getLifecycleWorkload(ctx, client, in.WorkloadID)
	if err != nil {
		return manageWorkloadLifecycleOutput{}, err
	}
	if err := validateLifecycleActionForWorkload(in.Action, workload); err != nil {
		return manageWorkloadLifecycleOutput{}, err
	}

	var preparedUpdate preparedWorkloadUpdate
	if in.Action == workloadLifecycleActionUpdate {
		if !isInferenceType(workload.Type) {
			return manageWorkloadLifecycleOutput{}, fmt.Errorf(
				"update is supported only for native Run:ai %s and %s workloads, got %q: %w",
				workloadTypeInference, workloadTypeDistribInf, workload.Type, runaimcp.ErrValidation)
		}
		// Decode the type-specific request once before confirmation, then execute
		// that same prepared request after the user approves it.
		preparedUpdate, err = prepareInferenceUpdate(workload.Type, in.UpdateRequest)
		if err != nil {
			return manageWorkloadLifecycleOutput{}, err
		}
	}
	if err := confirmWorkloadLifecycle(ctx, in, workload.Name); err != nil {
		return out, err
	}

	if in.Action == workloadLifecycleActionUpdate {
		if err := preparedUpdate(ctx, client, in.WorkloadID); err != nil {
			return manageWorkloadLifecycleOutput{}, err
		}
		out.Applied = true
		return out, nil
	}

	if isWorkloadsV2RunStateAction(in.Action, workload) {
		if err := executeWorkloadsV2RunState(ctx, client, in.WorkloadID, in.Action); err != nil {
			return manageWorkloadLifecycleOutput{}, err
		}
		out.Applied = true
		return out, nil
	}

	resp, err := executeWorkloadLifecycle(ctx, client, in.WorkloadID, in.Action)
	if err != nil {
		return manageWorkloadLifecycleOutput{}, err
	}
	// Confirm success from the workload's batch result.
	return applyWorkloadLifecycleResult(in.WorkloadID, in.Action, resp, out)
}

func validateLifecycleActionForWorkload(action string, workload *workloadsapi.WorkloadDetailed) error {
	if action == workloadLifecycleActionStop || action == workloadLifecycleActionRestart {
		if workload.ActionsSupport == nil || !workload.ActionsSupport.GetSuspend() {
			return fmt.Errorf("%s is not supported for %s workload %q: %w",
				action, workload.Type, workload.Id, runaimcp.ErrValidation)
		}
	}
	return nil
}

func isWorkloadsV2RunStateAction(action string, workload *workloadsapi.WorkloadDetailed) bool {
	if action != workloadLifecycleActionStop && action != workloadLifecycleActionRestart {
		return false
	}
	sourceAPI, ok := workload.GetSourceApiOk()
	return ok && sourceAPI != nil && *sourceAPI == workloadsapi.SOURCEAPI_WORKLOADS_V2
}

func executeWorkloadsV2RunState(ctx context.Context, client *runaiclient.Client, workloadID, action string) error {
	runState, err := lifecycleRunState(action)
	if err != nil {
		return err
	}
	_, httpResp, err := client.WorkloadsV2.WorkloadsV2API.SetWorkloadRunState(ctx, workloadID).
		WorkloadRunStateRequest(*workloadsv2.NewWorkloadRunStateRequest(runState)).
		Execute()
	if err != nil {
		mapped := mapLifecycleHTTPError(httpResp, err)
		if httpResp == nil || httpResp.StatusCode < http.StatusMultipleChoices {
			return fmt.Errorf("%s workload %q: request outcome is unknown; inspect the workload before retrying: %w",
				action, workloadID, mapped)
		}
		return fmt.Errorf("%s workload %q: %w", action, workloadID, mapped)
	}
	return nil
}

func lifecycleRunState(action string) (workloadsv2.RunState, error) {
	switch action {
	case workloadLifecycleActionStop:
		return workloadsv2.RUNSTATE_SUSPEND, nil
	case workloadLifecycleActionRestart:
		return workloadsv2.RUNSTATE_RESUME, nil
	default:
		return "", fmt.Errorf("unsupported workload run-state action %q: %w", action, runaimcp.ErrValidation)
	}
}

func getLifecycleWorkload(ctx context.Context, client *runaiclient.Client, workloadID string) (*workloadsapi.WorkloadDetailed, error) {
	workload, resp, err := client.Workloads.WorkloadsAPI.GetWorkload(ctx, workloadID).Execute()
	if err != nil {
		return nil, fmt.Errorf("get workload %q before lifecycle action: %w", workloadID, runaimcp.MapHTTPError(resp, err))
	}
	if workload == nil {
		return nil, fmt.Errorf("get workload %q before lifecycle action: nil response: %w", workloadID, runaimcp.ErrUpstream)
	}
	return workload, nil
}

func confirmWorkloadLifecycle(ctx context.Context, in manageWorkloadLifecycleInput, workloadName string) error {
	switch in.Action {
	case workloadLifecycleActionStop:
		return common.ConfirmWrite(ctx, in.Confirm, workloadLifecycleActionStop, "workload", workloadName, in.WorkloadID, nil)
	case workloadLifecycleActionRestart:
		return common.ConfirmWrite(ctx, in.Confirm, workloadLifecycleActionRestart, "workload", workloadName, in.WorkloadID, nil)
	case workloadLifecycleActionUpdate:
		return common.ConfirmWrite(ctx, in.Confirm, workloadLifecycleActionUpdate, "workload", workloadName, in.WorkloadID,
			updateConfirmationDetails(in.UpdateRequest))
	case workloadLifecycleActionDelete:
		details := common.WithConfirmWarning(nil, workloadLifecycleDeleteWarning)
		return common.ConfirmWrite(ctx, in.Confirm, workloadLifecycleActionDelete, "workload", workloadName, in.WorkloadID, details)
	default:
		return fmt.Errorf("unsupported workload lifecycle action %q: %w", in.Action, runaimcp.ErrValidation)
	}
}

// updateConfirmationDetails deliberately includes field names, not values. Update
// specs may contain environment-variable secrets, and confirmation errors can be
// surfaced through model context and logs.
func updateConfirmationDetails(raw map[string]any) map[string]any {
	spec, _ := raw["spec"].(map[string]any)
	fields := make([]string, 0, len(spec))
	for field := range spec {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	return map[string]any{"specFields": fields}
}

func validateUpdateRequestEnvelope(raw map[string]any) error {
	var envelope struct {
		Spec map[string]any `json:"spec"`
	}
	if err := common.DecodeTypedRequest(raw, &envelope); err != nil {
		return fmt.Errorf("decode updateRequest: %v: %w", err, runaimcp.ErrValidation)
	}
	if len(envelope.Spec) == 0 {
		return fmt.Errorf("updateRequest.spec must contain at least one field: %w", runaimcp.ErrValidation)
	}
	return nil
}

type preparedWorkloadUpdate func(context.Context, *runaiclient.Client, string) error

func prepareInferenceUpdate(workloadType string, raw map[string]any) (preparedWorkloadUpdate, error) {
	switch workloadType {
	case workloadTypeInference:
		var request workloadsapi.InferenceUpdateRequest
		if err := decodeTypedUpdateRequest(raw, &request); err != nil {
			return nil, err
		}
		return func(ctx context.Context, client *runaiclient.Client, workloadID string) error {
			_, httpResp, err := client.Workloads.InferencesAPI.UpdateInferenceSpec(ctx, workloadID).
				InferenceUpdateRequest(request).
				Execute()
			return handleInferenceUpdateResponse(workloadType, workloadID, httpResp, err)
		}, nil
	case workloadTypeDistribInf:
		var request workloadsapi.UpdateRequest
		if err := decodeTypedUpdateRequest(raw, &request); err != nil {
			return nil, err
		}
		return func(ctx context.Context, client *runaiclient.Client, workloadID string) error {
			_, httpResp, err := client.Workloads.DistributedInferencesAPI.UpdateDistributedInferenceSpec(ctx, workloadID).
				UpdateRequest(request).
				Execute()
			return handleInferenceUpdateResponse(workloadType, workloadID, httpResp, err)
		}, nil
	default:
		return nil, fmt.Errorf("update is unsupported for workload type %q: %w", workloadType, runaimcp.ErrValidation)
	}
}

func decodeTypedUpdateRequest(raw map[string]any, request any) error {
	if err := validateUpdateRequestEnvelope(raw); err != nil {
		return err
	}
	if err := common.DecodeTypedRequest(raw, request); err != nil {
		return fmt.Errorf("decode updateRequest: %v: %w", err, runaimcp.ErrValidation)
	}
	return nil
}

func handleInferenceUpdateResponse(workloadType, workloadID string, httpResp *http.Response, err error) error {
	if err != nil {
		mapped := mapLifecycleHTTPError(httpResp, err)
		if httpResp == nil || httpResp.StatusCode < http.StatusMultipleChoices {
			return fmt.Errorf("update %s workload %q: request outcome is unknown; inspect the workload before retrying: %w",
				workloadType, workloadID, mapped)
		}
		return fmt.Errorf("update %s workload %q: %w", workloadType, workloadID, mapped)
	}
	return nil
}

func executeWorkloadLifecycle(ctx context.Context, client *runaiclient.Client, workloadID, action string) (*workloadsapi.BatchResponse, error) {
	// Operation actions use the batch endpoint, even for one workload.
	batchAction, err := lifecycleBatchAction(action)
	if err != nil {
		return nil, err
	}
	resp, httpResp, err := client.Workloads.WorkloadsBatchAPI.BatchWorkloads(ctx).
		WorkloadBatch(*workloadsapi.NewWorkloadBatch([]string{workloadID}, batchAction)).
		Execute()
	if err != nil {
		mapped := mapLifecycleHTTPError(httpResp, err)
		if httpResp == nil || httpResp.StatusCode < http.StatusMultipleChoices {
			return nil, fmt.Errorf("%s workload %q: request outcome is unknown; inspect the workload before retrying: %w", action, workloadID, mapped)
		}
		return nil, fmt.Errorf("%s workload %q: %w", action, workloadID, mapped)
	}
	if resp == nil {
		return nil, fmt.Errorf("%s workload %q: nil response; request outcome is unknown, inspect the workload before retrying: %w", action, workloadID, runaimcp.ErrUpstream)
	}
	return resp, nil
}

func mapLifecycleHTTPError(resp *http.Response, err error) error {
	mapped := runaimcp.MapHTTPError(resp, err)
	if resp != nil && resp.StatusCode == http.StatusBadRequest {
		return fmt.Errorf("%w: %v", runaimcp.ErrValidation, mapped)
	}
	return mapped
}

func applyWorkloadLifecycleResult(workloadID, action string, resp *workloadsapi.BatchResponse, out manageWorkloadLifecycleOutput) (manageWorkloadLifecycleOutput, error) {
	// A successful HTTP response only means the batch was processed. The outcome
	// for this workload is reported separately in Failed or Succeeded.
	for _, failure := range resp.Failed {
		if strings.EqualFold(failure.Id, workloadID) {
			itemResp := &http.Response{StatusCode: int(failure.Code)}
			return manageWorkloadLifecycleOutput{}, fmt.Errorf("%s workload %q: %w", action, workloadID,
				mapLifecycleHTTPError(itemResp, errors.New(failure.Message)))
		}
	}
	for _, succeededID := range resp.Succeeded {
		if strings.EqualFold(succeededID, workloadID) {
			out.Applied = true
			return out, nil
		}
	}

	// Fail if the batch response does not report an outcome for the workload.
	return manageWorkloadLifecycleOutput{}, fmt.Errorf("%s workload %q: upstream batch response reported neither success nor failure; request outcome is unknown, inspect the workload before retrying: %w", action, workloadID, runaimcp.ErrUpstream) //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
}

func lifecycleBatchAction(action string) (workloadsapi.Action, error) {
	switch action {
	case workloadLifecycleActionStop:
		return workloadsapi.ACTION_SUSPEND, nil
	case workloadLifecycleActionRestart:
		return workloadsapi.ACTION_RESUME, nil
	case workloadLifecycleActionDelete:
		return workloadsapi.ACTION_DELETE, nil
	default:
		return "", fmt.Errorf("unsupported workload lifecycle action %q: %w", action, runaimcp.ErrValidation)
	}
}
