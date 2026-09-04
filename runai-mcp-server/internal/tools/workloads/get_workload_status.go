/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"fmt"
	"time"

	wlclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/rs/zerolog"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

const recentEventsLimit = 5

type getWorkloadStatusInput struct {
	WorkloadID string `json:"workloadId" jsonschema:"Workload ID (UUID)"`
}

func (in getWorkloadStatusInput) Validate() error {
	return common.ValidateUUID("workloadId", in.WorkloadID)
}

type workloadStatusEvent struct {
	CreatedAt string `json:"createdAt" jsonschema:"Event timestamp (RFC3339)"`
	Type      string `json:"type"      jsonschema:"Event severity (Normal | Warning | Error)"`
	Reason    string `json:"reason"    jsonschema:"Short, machine-friendly reason code"`
	Message   string `json:"message"   jsonschema:"Human-readable event message"`
}

type workloadStatusCondition struct {
	Type               string `json:"type"                         jsonschema:"Condition type reported by the workload service, such as Failed or Available"`
	Status             string `json:"status"                       jsonschema:"Condition status reported by the workload service, such as True, False, or Unknown"`
	Reason             string `json:"reason,omitempty"             jsonschema:"Machine-readable reason reported for the condition"`
	Message            string `json:"message,omitempty"            jsonschema:"Human-readable diagnostic message reported for the condition"`
	LastTransitionTime string `json:"lastTransitionTime,omitempty" jsonschema:"When the condition last transitioned (RFC3339); empty when unavailable"`
}

type getWorkloadStatusOutput struct {
	ID             string                    `json:"id"                       jsonschema:"Workload ID"`
	Name           string                    `json:"name"                     jsonschema:"Workload name"`
	ClusterID      string                    `json:"clusterId,omitempty"      jsonschema:"Cluster the workload runs on"`
	Type           string                    `json:"type"                     jsonschema:"Workload type (Training | Workspace | Distributed | Inference | DistributedInference | non-native types)"`
	Phase          string                    `json:"phase"                    jsonschema:"Current workload phase"`
	ProjectName    string                    `json:"projectName"              jsonschema:"Owning project name"`
	DepartmentName string                    `json:"departmentName,omitempty" jsonschema:"Owning department name"`
	CreatedAt      string                    `json:"createdAt"                jsonschema:"Workload creation timestamp (RFC3339)"`
	PhaseUpdatedAt string                    `json:"phaseUpdatedAt"           jsonschema:"Timestamp of the last phase transition (RFC3339)"`
	PhaseMessage   string                    `json:"phaseMessage,omitempty"   jsonschema:"Authoritative diagnostic message reported by the workload service for the current phase; quote or paraphrase this before making any inference"` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
	PhaseReason    string                    `json:"phaseReason,omitempty"    jsonschema:"Machine-readable reason reported by the workload service for the current phase"`
	K8sPhase       string                    `json:"k8sPhase,omitempty"       jsonschema:"Underlying Kubernetes phase reported by the workload service"`
	Conditions     []workloadStatusCondition `json:"conditions" jsonschema:"Authoritative workload conditions. Use reason and message as diagnostic evidence; always present, empty when the API reports none"`

	// Pointer so the encoder distinguishes a failed events call
	// (nil → JSON null) from a successful empty page (&[]).
	Events *[]workloadStatusEvent `json:"events" jsonschema:"Up to 5 most recent events, newest first; null when the events fetch failed, [] when there are none"`
}

func registerGetWorkloadStatus(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[getWorkloadStatusInput, getWorkloadStatusOutput]{
		Name: "get_workload_status",
		Description: "Call when the user asks about the state of a workload or job (e.g. is it running, pending, or failed and why) or about its recent events. " +
			"Returns phaseMessage, phaseReason, k8sPhase, conditions, and the 5 most recent events. For 'why' questions, check phaseMessage first, then condition reason/message and event reason/message. If they do not identify a cause, use get_workload_history for earlier phase messages and events; use get_workload_pods when pod phase or placement is relevant. " + //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
			"If workloadId is unknown, use list_workloads to resolve it. " +
			"Do NOT call to get GPU/CPU utilization — use get_workload_metrics. " +
			"Do NOT call to see the submitted spec — use get_workload_spec. Do NOT call to get pod-level state — use get_workload_pods.",
		ReadOnly:      true,
		MinAPIVersion: "2.18",
		Handler: func(ctx context.Context, in getWorkloadStatusInput) (getWorkloadStatusOutput, error) {
			return getWorkloadStatusRun(ctx, client, in)
		},
	})
}

func getWorkloadStatusRun(ctx context.Context, client *runaiclient.Client, in getWorkloadStatusInput) (getWorkloadStatusOutput, error) {
	wl, httpResp, err := client.Workloads.WorkloadsAPI.GetWorkload(ctx, in.WorkloadID).Execute()
	if err != nil {
		return getWorkloadStatusOutput{}, fmt.Errorf("get workload %q: %w", in.WorkloadID, runaimcp.MapHTTPError(httpResp, err))
	}
	if wl == nil {
		return getWorkloadStatusOutput{}, fmt.Errorf("nil workload response from upstream: %w", runaimcp.ErrUpstream)
	}
	phaseReason := ""
	if reason := wl.PhaseReason.Get(); reason != nil {
		phaseReason = string(*reason)
	}
	out := getWorkloadStatusOutput{
		ID:             wl.Id,
		Name:           wl.Name,
		ClusterID:      wl.ClusterId,
		Type:           wl.Type,
		Phase:          string(wl.Phase),
		ProjectName:    wl.ProjectName,
		DepartmentName: wl.DepartmentName,
		CreatedAt:      wl.CreatedAt.UTC().Format(time.RFC3339),
		PhaseUpdatedAt: wl.PhaseUpdatedAt.UTC().Format(time.RFC3339),
		PhaseMessage:   wl.GetPhaseMessage(),
		PhaseReason:    phaseReason,
		K8sPhase:       wl.K8sPhase,
		Conditions:     mapWorkloadStatusConditions(wl.Conditions),
	}
	attachRecentEvents(ctx, client, in.WorkloadID, &out)
	return out, nil
}

func mapWorkloadStatusConditions(conditions []wlclient.Condition) []workloadStatusCondition {
	mapped := make([]workloadStatusCondition, 0, len(conditions))
	for _, condition := range conditions {
		mapped = append(mapped, workloadStatusCondition{
			Type:               condition.Type,
			Status:             condition.Status,
			Reason:             condition.GetReason(),
			Message:            condition.GetMessage(),
			LastTransitionTime: common.TimestampString(condition.LastTransitionTime.Get()),
		})
	}
	return mapped
}

func attachRecentEvents(ctx context.Context, client *runaiclient.Client, workloadID string, out *getWorkloadStatusOutput) {
	resp, _, err := client.Workloads.EventsAPI.GetWorkloadEvents(ctx, workloadID).
		Limit(int32(recentEventsLimit)).
		SortOrder("desc").
		Execute()
	if err != nil {
		zerolog.Ctx(ctx).Warn().
			Err(err).
			Str("workload_id", workloadID).
			Msg("get_workload_status: events fetch failed, returning events=null")
		return
	}
	if resp == nil {
		zerolog.Ctx(ctx).Warn().
			Err(runaimcp.ErrUpstream).
			Str("workload_id", workloadID).
			Msg("get_workload_status: events fetch returned nil, returning events=null")
		return
	}
	mapped := make([]workloadStatusEvent, 0, len(resp.Events))
	for _, e := range resp.Events {
		reason := ""
		if e.Reason != nil {
			reason = *e.Reason
		}
		mapped = append(mapped, workloadStatusEvent{
			CreatedAt: e.CreatedAt.UTC().Format(time.RFC3339),
			Type:      e.Type,
			Reason:    reason,
			Message:   e.Message,
		})
	}
	out.Events = &mapped
}
