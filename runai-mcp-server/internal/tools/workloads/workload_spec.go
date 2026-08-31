/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/run-ai/runai/runai-application-packages/jsonutils/gojsonq"
	workloads "github.com/run-ai/runai/runai-clients/workloads-client/generated"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

type getWorkloadSpecInput struct {
	WorkloadID string `json:"workloadId" jsonschema:"Workload ID (UUID)"`
}

func (in getWorkloadSpecInput) Validate() error {
	return common.ValidateUUID("workloadId", in.WorkloadID)
}

type workloadTopologyOutput struct {
	ActualPlacement map[string]any `json:"actualPlacement,omitempty" jsonschema:"Lowest topology level at which all running pods are colocated (level + value)"`
}

type pendingSchedulingMessageOutput struct {
	NodePool    string `json:"nodePool"              jsonschema:"Node pool the scheduler tried"`
	PhaseReason string `json:"phaseReason"           jsonschema:"Machine-readable reason class (e.g. SchedulingFailed)"`
	Reason      string `json:"reason"                jsonschema:"Short reason code from the scheduler"`
	OrgType     string `json:"orgType,omitempty"     jsonschema:"Organizational scope (e.g. project, department)"`
	UserMessage string `json:"userMessage,omitempty" jsonschema:"Human-readable message"`
}

type getWorkloadSpecOutput struct {
	WorkloadID                string                           `json:"workloadId"                          jsonschema:"Echo of input workloadId"`
	Name                      string                           `json:"name"                                jsonschema:"Workload name"`
	Type                      string                           `json:"type"                                jsonschema:"Workload type (Training | Workspace | Distributed | Inference | DistributedInference | non-native types)"` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
	ProjectName               string                           `json:"projectName"                         jsonschema:"Project name"`
	DepartmentName            string                           `json:"departmentName,omitempty"            jsonschema:"Department name"`
	Spec                      map[string]any                   `json:"spec,omitempty"                      jsonschema:"Submitted workload spec, forwarded verbatim. Shape varies by kind. Omitted when the kind-specific call failed or returned no spec"`          //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
	AllocatedResources        map[string]any                   `json:"allocatedResources,omitempty"        jsonschema:"Resources the scheduler granted. Typical fields: gpu (float), gpuMemory (quantity), cpu (float), cpuMemory (quantity), extendedResources[]"` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
	Topology                  *workloadTopologyOutput          `json:"topology,omitempty"                  jsonschema:"Network topology placement"`
	PendingSchedulingMessages []pendingSchedulingMessageOutput `json:"pendingSchedulingMessages,omitempty" jsonschema:"Scheduler messages explaining why the workload is pending"`
}

func registerGetWorkloadSpec(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[getWorkloadSpecInput, getWorkloadSpecOutput]{
		Name: "get_workload_spec",
		Description: "Call when the user asks what a workload (job) requested or how it is configured — the resources " +
			"it asked for, the image it runs, its environment variables, volumes, or node placement — or why it is " +
			"pending scheduling. Returns the workload's requested spec (forwarded as-is, so which fields appear " +
			"depends on the workload type), alongside its actually-allocated resources, its actual node placement, " +
			"and any pending-scheduling messages explaining why it has not been placed. " +
			"If workloadId is unknown, use list_workloads to resolve it. " +
			"Use this tool for the allocated or requested GPU count of one workload; use list_workloads to list or filter multiple workloads by allocatedGpu or requestedGpu. " +
			"When resolving a name, follow all list_workloads pages and ask for project or cluster context if multiple matches remain. " +
			"Do NOT call for phase transitions over time — use get_workload_history. " +
			"Do NOT call for time-series utilization — use get_workload_metrics. " +
			"Do NOT call for policy info — use get_workload_effective_policy.",
		ReadOnly:      true,
		MinAPIVersion: "2.18",
		Handler: func(ctx context.Context, in getWorkloadSpecInput) (getWorkloadSpecOutput, error) {
			return getWorkloadSpecRun(ctx, client, in)
		},
	})
}

// getWorkloadSpecRun returns what a workload requested (spec) alongside what it actually got
// at runtime (allocatedResources, topology, pendingSchedulingMessages):
//  1. GetWorkload (hard fail) — establishes type + identity and yields the runtime fields.
//  2. Kind-specific GET (soft fail) — adds `spec` for native types with a dedicated endpoint;
//     for other types the call is skipped and `spec` is omitted from the response.
func getWorkloadSpecRun(ctx context.Context, client *runaiclient.Client, in getWorkloadSpecInput) (getWorkloadSpecOutput, error) {
	// Step 1 — identify the workload. Hard fail: nothing useful to return without type + identity.
	wl, httpResp, err := client.Workloads.WorkloadsAPI.GetWorkload(ctx, in.WorkloadID).Execute()
	if err != nil {
		if httpResp != nil && (httpResp.StatusCode == http.StatusNotFound || httpResp.StatusCode == http.StatusForbidden) {
			return getWorkloadSpecOutput{}, fmt.Errorf("get workload %q: %w", in.WorkloadID, runaimcp.ErrNotFound)
		}
		return getWorkloadSpecOutput{}, fmt.Errorf("get workload %q: %w", in.WorkloadID, runaimcp.MapHTTPError(httpResp, err))
	}
	out := getWorkloadSpecOutput{
		WorkloadID:                in.WorkloadID,
		Name:                      wl.Name,
		Type:                      wl.Type,
		ProjectName:               wl.ProjectName,
		DepartmentName:            wl.DepartmentName,
		AllocatedResources:        toRawObject(wl.AllocatedResources.Get()),
		Topology:                  mapWorkloadTopology(wl.Topology.Get()),
		PendingSchedulingMessages: mapPendingSchedulingMessages(wl.PendingSchedulingMessages),
	}

	// Step 2 — fetch the submitted spec from the type-specific endpoint. Soft fail: spec is dropped on failure, runtime fields above are kept.
	if spec, ok := fetchKindSpec(ctx, client, in.WorkloadID, wl.Type); ok {
		out.Spec = spec
	}
	return out, nil
}

// fetchKindSpec dispatches to the type-specific endpoint; soft-fails on errors.
func fetchKindSpec(ctx context.Context, client *runaiclient.Client, workloadID, workloadType string) (map[string]any, bool) {
	var (
		call     string
		spec     workloads.MappedNullable
		httpResp *http.Response
		err      error
	)
	switch workloadType {
	case workloadTypeTraining:
		call = "getTraining"
		spec, httpResp, err = execKind(client.Workloads.TrainingsAPI.GetTraining(ctx, workloadID).Execute,
			func(r *workloads.Training) workloads.MappedNullable { return r.Spec })
	case workloadTypeWorkspace:
		call = "getWorkspace"
		spec, httpResp, err = execKind(client.Workloads.WorkspacesAPI.GetWorkspace(ctx, workloadID).Execute,
			func(r *workloads.Workspace) workloads.MappedNullable { return r.Spec })
	case workloadTypeDistributed:
		call = "getDistributed"
		spec, httpResp, err = execKind(client.Workloads.DistributedAPI.GetDistributed(ctx, workloadID).Execute,
			func(r *workloads.DistributedWorkload) workloads.MappedNullable { return r.Spec })
	case workloadTypeInference:
		call = "getInference"
		spec, httpResp, err = execKind(client.Workloads.InferencesAPI.GetInference(ctx, workloadID).Execute,
			func(r *workloads.Inference) workloads.MappedNullable { return r.Spec })
	case workloadTypeDistribInf:
		call = "getDistributedInference"
		spec, httpResp, err = execKind(client.Workloads.DistributedInferencesAPI.GetDistributedInference(ctx, workloadID).Execute,
			func(r *workloads.DistributedInference) workloads.MappedNullable { return r.Spec })
	default:
		return nil, false
	}
	if err != nil {
		if rawSpec, ok := specFromSuccessfulDecodeError(httpResp, err); ok {
			zerolog.Ctx(ctx).Warn().Err(err).
				Str("workload_id", workloadID).
				Str("workload_type", workloadType).
				Str("call", call).
				Msg("get_workload_spec: recovered spec from raw successful response after generated-client decode failure")
			return rawSpec, true
		}
		evt := zerolog.Ctx(ctx).Warn().Err(err).
			Str("workload_id", workloadID).
			Str("workload_type", workloadType).
			Str("call", call)
		if httpResp != nil {
			evt = evt.Int("status_code", httpResp.StatusCode)
		}
		evt.Msg("get_workload_spec: kind-specific spec lookup failed, omitting spec")
		return nil, false
	}
	return toRawObject(spec), true
}

// specFromSuccessfulDecodeError recovers spec from a successful response the generated client could not decode.
// It preserves strict validation while preventing invalid response metadata from hiding a valid spec.
func specFromSuccessfulDecodeError(httpResp *http.Response, err error) (map[string]any, bool) {
	if httpResp == nil || httpResp.StatusCode < http.StatusOK || httpResp.StatusCode >= http.StatusMultipleChoices {
		return nil, false
	}

	var bodyError interface{ Body() []byte }
	if !errors.As(err, &bodyError) || len(bodyError.Body()) == 0 {
		return nil, false
	}

	var envelope struct {
		Spec json.RawMessage `json:"spec"`
	}
	if err := json.Unmarshal(bodyError.Body(), &envelope); err != nil || len(envelope.Spec) == 0 || bytes.Equal(envelope.Spec, []byte("null")) {
		return nil, false
	}

	decoder := json.NewDecoder(bytes.NewReader(envelope.Spec))
	decoder.UseNumber()
	var spec map[string]any
	if err := decoder.Decode(&spec); err != nil || len(spec) == 0 {
		return nil, false
	}
	return spec, true
}

func execKind[T comparable](exec func() (T, *http.Response, error), extract func(T) workloads.MappedNullable) (workloads.MappedNullable, *http.Response, error) {
	resp, httpResp, err := exec()
	var zero T
	if resp == zero {
		return nil, httpResp, err
	}
	return extract(resp), httpResp, err
}

// mapWorkloadTopology surfaces only actualPlacement; spec-side constraints are dropped.
func mapWorkloadTopology(topology *workloads.WorkloadTopology) *workloadTopologyOutput {
	if topology == nil {
		return nil
	}
	actualPlacement := toRawObject(topology.ActualPlacement.Get())
	if actualPlacement == nil {
		return nil
	}
	return &workloadTopologyOutput{ActualPlacement: actualPlacement}
}

func mapPendingSchedulingMessages(messages []workloads.PendingSchedulingMessage) []pendingSchedulingMessageOutput {
	if len(messages) == 0 {
		return nil
	}
	out := make([]pendingSchedulingMessageOutput, 0, len(messages))
	for _, message := range messages {
		entry := pendingSchedulingMessageOutput{
			NodePool:    message.NodePool,
			PhaseReason: string(message.PhaseReason),
			Reason:      message.Reason,
		}
		if orgType := message.OrgType.Get(); orgType != nil {
			entry.OrgType = string(*orgType)
		}
		if userMessage := message.UserMessage.Get(); userMessage != nil {
			entry.UserMessage = *userMessage
		}
		out = append(out, entry)
	}
	return out
}

func toRawObject(v workloads.MappedNullable) map[string]any {
	if workloads.IsNil(v) {
		return nil
	}
	m, err := gojsonq.LoadJson(v)
	if err != nil || len(m) == 0 {
		return nil
	}
	return m
}
