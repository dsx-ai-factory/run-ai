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
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog"

	workloads "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

type getWorkloadSpecInput struct {
	WorkloadID      string `json:"workloadId"                jsonschema:"Workload ID (UUID)"`
	IncludeFullSpec bool   `json:"includeFullSpec,omitempty" jsonschema:"Return the complete submitted spec instead of the default compact summary."`
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
	Spec                      map[string]any                   `json:"spec,omitempty"                      jsonschema:"Submitted workload spec. By default contains only image, compute, storage, scaling, and placement or scheduling fields; all submitted fields, including distributed master configuration, are returned when includeFullSpec is true. Omitted when unavailable."` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
	AllocatedResources        map[string]any                   `json:"allocatedResources,omitempty"        jsonschema:"Resources the scheduler granted. Typical fields: gpu (float), gpuMemory (quantity), cpu (float), cpuMemory (quantity), extendedResources[]"`                                                                                                                     //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
	Topology                  *workloadTopologyOutput          `json:"topology,omitempty"                  jsonschema:"Network topology placement"`
	PendingSchedulingMessages []pendingSchedulingMessageOutput `json:"pendingSchedulingMessages,omitempty" jsonschema:"Scheduler messages explaining why the workload is pending"`
	Warnings                  []string                         `json:"warnings,omitempty"                  jsonschema:"Non-fatal notes identifying workload spec data that could not be retrieved."`
}

func registerGetWorkloadSpec(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[getWorkloadSpecInput, getWorkloadSpecOutput]{
		Name: "get_workload_spec",
		Description: "Call when the user asks what a workload (job) requested or how it is configured — the resources " +
			"it asked for, the image it runs, its environment variables, volumes, or node placement — or why it is " +
			"pending scheduling. Returns a compact requested-spec summary by default; set includeFullSpec for fields " +
			"outside image, compute, storage, scaling, placement, and scheduling. Also returns allocated resources, actual node placement, " +
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
	if wl == nil {
		return getWorkloadSpecOutput{}, fmt.Errorf("get workload %q returned a nil response: %w", in.WorkloadID, runaimcp.ErrUpstream)
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

	// Step 2 — fetch the submitted spec from the type-specific endpoint. Soft fail:
	// spec is dropped on failure, runtime fields above are kept, and the caller is warned.
	result := fetchKindSpec(ctx, client, in.WorkloadID, wl.Type)
	if result.Retrieved {
		removeWorkloadSpecSecrets(result.Spec)
		if in.IncludeFullSpec {
			out.Spec = result.Spec
		} else {
			out.Spec = summarizeWorkloadSpec(result.Spec)
		}
	} else if result.Applicable {
		out.Warnings = []string{fmt.Sprintf("%s workload spec could not be retrieved; runtime fields are still included", wl.Type)}
	}
	return out, nil
}

type kindSpecResult struct {
	Spec       map[string]any
	Applicable bool
	Retrieved  bool
}

// fetchKindSpec dispatches to the type-specific endpoint; soft-fails on errors.
func fetchKindSpec(ctx context.Context, client *runaiclient.Client, workloadID, workloadType string) kindSpecResult {
	var (
		call                string
		spec                workloads.MappedNullable
		distributedWorkload *workloads.DistributedWorkload
		httpResp            *http.Response
		err                 error
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
			func(r *workloads.DistributedWorkload) workloads.MappedNullable {
				distributedWorkload = r
				return r.Spec
			})
	case workloadTypeInference:
		call = "getInference"
		spec, httpResp, err = execKind(client.Workloads.InferencesAPI.GetInference(ctx, workloadID).Execute,
			func(r *workloads.Inference) workloads.MappedNullable { return r.Spec })
	case workloadTypeDistribInf:
		call = "getDistributedInference"
		spec, httpResp, err = execKind(client.Workloads.DistributedInferencesAPI.GetDistributedInference(ctx, workloadID).Execute,
			func(r *workloads.DistributedInference) workloads.MappedNullable { return r.Spec })
	default:
		return kindSpecResult{}
	}
	if err != nil {
		if rawSpec, ok := specFromSuccessfulResponse(httpResp, workloadType); ok {
			zerolog.Ctx(ctx).Warn().Err(err).
				Str("workload_id", workloadID).
				Str("workload_type", workloadType).
				Str("call", call).
				Msg("get_workload_spec: recovered spec from raw successful response after generated-client decode failure")
			return kindSpecResult{Spec: rawSpec, Applicable: true, Retrieved: true}
		}
		if rawSpec, ok := specFromSuccessfulDecodeError(httpResp, err, workloadType); ok {
			zerolog.Ctx(ctx).Warn().Err(err).
				Str("workload_id", workloadID).
				Str("workload_type", workloadType).
				Str("call", call).
				Msg("get_workload_spec: recovered spec from raw successful response after generated-client decode failure")
			return kindSpecResult{Spec: rawSpec, Applicable: true, Retrieved: true}
		}
		evt := zerolog.Ctx(ctx).Warn().Err(err).
			Str("workload_id", workloadID).
			Str("workload_type", workloadType).
			Str("call", call)
		if httpResp != nil {
			evt = evt.Int("status_code", httpResp.StatusCode)
		}
		evt.Msg("get_workload_spec: kind-specific spec lookup failed, omitting spec")
		return kindSpecResult{Applicable: true}
	}
	if rawSpec, ok := specFromSuccessfulResponse(httpResp, workloadType); ok {
		return kindSpecResult{Spec: rawSpec, Applicable: true, Retrieved: true}
	}
	rawSpec := mergeDistributedWorkloadSpec(toRawObject(spec), distributedWorkload)
	return kindSpecResult{Spec: rawSpec, Applicable: true, Retrieved: len(rawSpec) > 0}
}

var workloadSpecSummaryFields = map[string]struct{}{
	"image": {}, "compute": {}, "storage": {},
	"distributedFramework": {}, "cleanPodPolicy": {}, "mpiLauncherCreationPolicy": {},
	"nodePools": {}, "nodeAffinityRequired": {}, "nodeType": {}, "podAffinity": {}, "tolerations": {},
	"priorityClass": {}, "preemptibility": {}, "restartPolicy": {},
	"replicas": {}, "workers": {}, "numWorkers": {}, "minReplicas": {}, "maxReplicas": {}, "segmentSize": {},
	"parallelism": {}, "completions": {}, "slotsPerWorker": {}, "autoscaling": {},
	"servingConfiguration": {}, "servingPort": {}, "startupPolicy": {},
	"masterSpecSameAsWorker": {},
}

var workloadSpecRoleFields = map[string]struct{}{
	"leader": {}, "worker": {}, "masterSpec": {},
}

// mergeDistributedWorkloadSpec adds the submitted master configuration, which the
// distributed endpoint returns beside (rather than inside) its common worker spec.
func mergeDistributedWorkloadSpec(spec map[string]any, distributed *workloads.DistributedWorkload) map[string]any {
	if distributed == nil {
		return spec
	}
	if value, set := distributed.GetMasterSpecSameAsWorkerOk(); set {
		spec = ensureWorkloadSpec(spec)
		if value == nil {
			spec["masterSpecSameAsWorker"] = nil
		} else {
			spec["masterSpecSameAsWorker"] = *value
		}
	}
	if value, set := distributed.GetMasterSpecOk(); set {
		spec = ensureWorkloadSpec(spec)
		spec["masterSpec"] = toRawObject(value)
	}
	if value, set := distributed.GetMasterAssetsOk(); set {
		spec = ensureWorkloadSpec(spec)
		spec["masterAssets"] = toRawObject(value)
	}
	return spec
}

func ensureWorkloadSpec(spec map[string]any) map[string]any {
	if spec == nil {
		return make(map[string]any)
	}
	return spec
}

// summarizeWorkloadSpec keeps only current raw spec fields needed to understand
// images, resource requests, storage, scale, placement, and scheduling.
func summarizeWorkloadSpec(spec map[string]any) map[string]any {
	if len(spec) == 0 {
		return nil
	}
	summary := make(map[string]any)
	for key, value := range spec {
		if _, ok := workloadSpecSummaryFields[key]; ok {
			summary[key] = value
			continue
		}
		if _, ok := workloadSpecRoleFields[key]; !ok {
			continue
		}
		role, ok := value.(map[string]any)
		if !ok {
			continue
		}
		if roleSummary := summarizeWorkloadSpec(role); len(roleSummary) > 0 {
			summary[key] = roleSummary
		}
	}
	if len(summary) == 0 {
		return nil
	}
	return summary
}

// removeWorkloadSpecSecrets strips credentials that may be embedded in storage
// configuration while retaining the non-secret fields needed to understand it.
func removeWorkloadSpecSecrets(value any) {
	switch value := value.(type) {
	case map[string]any:
		for key, nested := range value {
			if strings.EqualFold(key, "awsSessionToken") {
				delete(value, key)
				continue
			}
			removeWorkloadSpecSecrets(nested)
		}
	case []any:
		for _, nested := range value {
			removeWorkloadSpecSecrets(nested)
		}
	case []map[string]any:
		for _, nested := range value {
			removeWorkloadSpecSecrets(nested)
		}
	}
}

// specFromSuccessfulResponse reads the untouched response JSON so full mode
// preserves spec fields added after the generated client was built.
func specFromSuccessfulResponse(httpResp *http.Response, workloadType string) (map[string]any, bool) {
	if httpResp == nil || httpResp.Body == nil || httpResp.StatusCode < http.StatusOK || httpResp.StatusCode >= http.StatusMultipleChoices {
		return nil, false
	}
	body, err := io.ReadAll(httpResp.Body)
	_ = httpResp.Body.Close()
	httpResp.Body = io.NopCloser(bytes.NewReader(body))
	if err != nil {
		return nil, false
	}
	return specFromResponseBody(body, workloadType)
}

func specFromResponseBody(body []byte, workloadType string) (map[string]any, bool) {
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, false
	}
	rawSpec, exists := envelope["spec"]
	if !exists || len(rawSpec) == 0 || bytes.Equal(rawSpec, []byte("null")) {
		return nil, false
	}

	decoder := json.NewDecoder(bytes.NewReader(rawSpec))
	decoder.UseNumber()
	var spec map[string]any
	if err := decoder.Decode(&spec); err != nil || len(spec) == 0 {
		return nil, false
	}
	if workloadType == workloadTypeDistributed {
		for _, field := range []string{"masterSpecSameAsWorker", "masterSpec", "masterAssets"} {
			rawValue, exists := envelope[field]
			if !exists {
				continue
			}
			valueDecoder := json.NewDecoder(bytes.NewReader(rawValue))
			valueDecoder.UseNumber()
			var value any
			if valueDecoder.Decode(&value) == nil {
				spec[field] = value
			}
		}
	}
	return spec, true
}

// specFromSuccessfulDecodeError recovers spec from a successful response the generated client could not decode.
// It preserves strict validation while preventing invalid response metadata from hiding a valid spec.
func specFromSuccessfulDecodeError(httpResp *http.Response, err error, workloadType string) (map[string]any, bool) {
	if httpResp == nil || httpResp.StatusCode < http.StatusOK || httpResp.StatusCode >= http.StatusMultipleChoices {
		return nil, false
	}

	var bodyError interface{ Body() []byte }
	if !errors.As(err, &bodyError) || len(bodyError.Body()) == 0 {
		return nil, false
	}

	return specFromResponseBody(bodyError.Body(), workloadType)
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
