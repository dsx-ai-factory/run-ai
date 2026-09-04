/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/rs/zerolog"

	workloads "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

const (
	historyDefaultLimit = int32(20)
	historyMaxLimit     = int32(50)
	revisionEventsLimit = int32(5)
)

type getWorkloadHistoryInput struct {
	WorkloadID string `json:"workloadId"           jsonschema:"Workload ID (UUID)"`
	Limit      int32  `json:"limit,omitempty"      jsonschema:"Max history records; default 20, max 50"`
	Offset     int32  `json:"offset,omitempty"     jsonschema:"Pagination offset; default 0. Pass back 'next' from a previous response to fetch the next page"`
	RevisionID string `json:"revisionId,omitempty" jsonschema:"Revision ID (UUID); inference workloads only. When set, the output's revisions[] contains only this revision with its pods and events. Errors with validation when supplied for a non-inference workload, and with not_found when the revisionId doesn't belong to this workload"` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
}

func (in getWorkloadHistoryInput) Validate() error {
	if err := common.ValidateUUID("workloadId", in.WorkloadID); err != nil {
		return err
	}
	if in.Limit < 0 || in.Limit > historyMaxLimit {
		return fmt.Errorf("limit must be between 0 and %d (got %d)", historyMaxLimit, in.Limit)
	}
	if in.Offset < 0 {
		return fmt.Errorf("offset must be >= 0 (got %d)", in.Offset)
	}
	if strings.TrimSpace(in.RevisionID) != "" {
		if err := common.ValidateUUID("revisionId", in.RevisionID); err != nil {
			return err
		}
	}
	return nil
}

type historyRecord struct {
	Meta historyRecordMeta `json:"meta" jsonschema:"Creation timestamp and record type"`
	Spec historyRecordSpec `json:"spec" jsonschema:"Record body — exactly one of phaseUpdate or event is populated, matching meta.type"`
}

type historyRecordMeta struct {
	CreationTimestamp string `json:"creationTimestamp" jsonschema:"When the record was created (RFC3339)"`
	Type              string `json:"type"              jsonschema:"Record type (PhaseUpdate | Event)"`
}

type historyRecordSpec struct {
	PhaseUpdate *historyPhaseUpdate `json:"phaseUpdate,omitempty" jsonschema:"Populated when meta.type == \"PhaseUpdate\" — phase the workload transitioned into"`
	Event       *eventOutput        `json:"event,omitempty"       jsonschema:"Populated when meta.type == \"Event\" — Kubernetes-style event the workload emitted"`
}

type historyPhaseUpdate struct {
	Phase        string `json:"phase"        jsonschema:"Phase the workload transitioned into"`
	PhaseMessage string `json:"phaseMessage" jsonschema:"Human-readable explanation for the transition"`
}

type eventOutput struct {
	CreatedAt string `json:"createdAt"        jsonschema:"Event timestamp (RFC3339)"`
	Type      string `json:"type"             jsonschema:"Event severity (Normal | Warning)"`
	Reason    string `json:"reason,omitempty" jsonschema:"Short, machine-friendly reason code"`
	Message   string `json:"message"          jsonschema:"Human-readable event message"`
	Source    string `json:"source,omitempty" jsonschema:"Reporting component (e.g. scheduler, kubelet)"`
}

type revisionRequestedPods struct {
	Min    *int32 `json:"min,omitempty"    jsonschema:"Minimum number of pods to maintain"`
	Max    *int32 `json:"max,omitempty"    jsonschema:"Maximum number of pods the autoscaler may scale to"`
	Number *int32 `json:"number,omitempty" jsonschema:"Fixed pod count (set instead of min/max for non-autoscaling workloads)"`
}

type revisionOutput struct {
	ID            string                 `json:"id"                      jsonschema:"Revision ID (UUID)"`
	Name          string                 `json:"name"                    jsonschema:"Revision name"`
	CreatedAt     string                 `json:"createdAt"               jsonschema:"Revision creation timestamp (RFC3339)"`
	Phase         string                 `json:"phase"                   jsonschema:"Revision phase"`
	RequestedPods *revisionRequestedPods `json:"requestedPods,omitempty" jsonschema:"Pod-count request (min/max/number)"`

	Pods   *[]podOutput   `json:"pods,omitempty"   jsonschema:"Pods running this revision — same shape as get_workload_pods.pods (without measurements). Only present when revisionId was supplied; [] when the fetch succeeded with no pods"` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
	Events *[]eventOutput `json:"events,omitempty" jsonschema:"Up to 5 most recent revision-scoped events, newest first. Only present when revisionId was supplied; [] when the fetch succeeded with no events"`
}

type getWorkloadHistoryOutput struct {
	WorkloadID string            `json:"workloadId"          jsonschema:"Echo of input workloadId"`
	Type       string            `json:"type,omitempty"      jsonschema:"Workload type (Training | Workspace | Distributed | Inference | DistributedInference | non-native types); omitted when the workload-type lookup failed"` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
	Records    []historyRecord   `json:"records"             jsonschema:"History records"`
	Next       *int32            `json:"next"                jsonschema:"Pagination cursor — pass back as 'offset' to fetch the next page. Null when there is no next page"`
	Revisions  *[]revisionOutput `json:"revisions,omitempty" jsonschema:"Inference workload revisions. With revisionId: just the matched revision (with pods + events). Without: summaries of all revisions. Omitted for non-inference workloads or on lookup failure; [] when none exist"` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
}

func registerGetWorkloadHistory(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[getWorkloadHistoryInput, getWorkloadHistoryOutput]{
		Name: "get_workload_history",
		Description: "Call when the user asks about a workload's event history — what happened to a workload (job) " +
			"over time and the lifecycle events logged along the way — or, for inference workloads, its revisions. " +
			"Returns a chronological, paginated list of history records; each record is a phase update or an event " +
			"(a lifecycle event with its reason and message). Pass the returned next value as offset to get older " +
			"records. For inference workloads, also returns the list of revisions; pass a revisionId to fetch only " +
			"that revision with its pods and events. " +
			"If workloadId is unknown, use list_workloads to resolve it. " +
			"Do NOT call to get the current state — use get_workload_status.",
		ReadOnly:      true,
		MinAPIVersion: "2.21",
		Handler: func(ctx context.Context, in getWorkloadHistoryInput) (getWorkloadHistoryOutput, error) {
			return getWorkloadHistoryRun(ctx, client, in)
		},
	})
}

// getWorkloadHistoryRun assembles a workload's history:
//  1. fetch the page of history records (primary, hard fail);
//  2. resolve the workload type — used to gate the inference revisions
//  3. for inference workloads, attach the revisions list. When a revisionId is
//     supplied for a known non-inference workload, hard-fail with validation
//     upfront — only Inference / DistributedInference have a revisions API.
func getWorkloadHistoryRun(ctx context.Context, client *runaiclient.Client, in getWorkloadHistoryInput) (getWorkloadHistoryOutput, error) {
	// Step 1 — fetch the page of history records. Hard fail: the records list is the tool's core return value.
	limit, offset := resolveHistoryWindow(in)
	historyResp, httpResp, err := client.Workloads.EventsAPI.GetWorkloadHistory(ctx, in.WorkloadID).
		Limit(limit).
		Offset(offset).
		Execute()
	if err != nil {
		return getWorkloadHistoryOutput{}, fmt.Errorf("get workload history %q: %w", in.WorkloadID, runaimcp.MapHTTPError(httpResp, err))
	}
	out := getWorkloadHistoryOutput{
		WorkloadID: in.WorkloadID,
		Records:    mapHistoryRecords(historyResp.Records),
		Next:       historyResp.Next,
	}

	// Step 2 — resolve the workload type. Hard-fails not_found for unknown / cross-tenant IDs; soft-fails (warn-logs) on transient errors, in which case `type` is dropped and we still return the records page.
	workloadType, ok, err := resolveWorkloadType(ctx, client, in.WorkloadID)
	if err != nil {
		return getWorkloadHistoryOutput{}, err
	}
	if ok {
		out.Type = workloadType
	}

	// Step 3 — collect revisions for inference workloads.
	revisions, err := resolveRevisions(ctx, client, in, workloadType, ok)
	if err != nil {
		return getWorkloadHistoryOutput{}, err
	}
	out.Revisions = revisions
	return out, nil
}

// resolveRevisions returns the revisions list, skips it, or rejects upfront
// when revisionId is supplied for a known non-inference workload.
func resolveRevisions(ctx context.Context, client *runaiclient.Client, in getWorkloadHistoryInput, workloadType string, typeKnown bool) (*[]revisionOutput, error) {
	if !isInferenceType(workloadType) {
		if in.RevisionID == "" {
			return nil, nil
		}
		if typeKnown {
			return nil, fmt.Errorf("revisionId is only valid for inference workloads (workload %q is %q): %w", in.WorkloadID, workloadType, runaimcp.ErrValidation)
		}
	}
	return collectRevisions(ctx, client, in.WorkloadID, in.RevisionID)
}

// resolveHistoryWindow defaults limit to 20 when not provided; offset defaults
// to 0 naturally. The upstream max-50 cap is enforced in Validate.
func resolveHistoryWindow(in getWorkloadHistoryInput) (int32, int32) {
	limit := in.Limit
	if limit == 0 {
		limit = historyDefaultLimit
	}
	return limit, in.Offset
}

// collectRevisions returns inference revisions. With revisionID: only that
// revision (with pods + events); hard-fails not_found if it can't resolve.
// Without: summaries of all revisions; soft-fails to nil on lookup failure.
func collectRevisions(ctx context.Context, client *runaiclient.Client, workloadID, revisionID string) (*[]revisionOutput, error) {
	listed, listErr := listInferenceRevisions(ctx, client, workloadID)

	if revisionID != "" {
		if listErr == nil {
			for _, revision := range listed {
				if strings.EqualFold(revision.Id, revisionID) {
					entry := mapRevision(revision)
					attachRevisionDetail(ctx, client, entry.ID, &entry)
					return &[]revisionOutput{entry}, nil
				}
			}
		}
		fallback, err := fetchRevisionFallback(ctx, client, workloadID, revisionID)
		if err != nil {
			return nil, err
		}
		attachRevisionDetail(ctx, client, revisionID, &fallback)
		return &[]revisionOutput{fallback}, nil
	}

	if listErr != nil {
		return nil, nil
	}
	out := make([]revisionOutput, 0, len(listed))
	for _, revision := range listed {
		out = append(out, mapRevision(revision))
	}
	return &out, nil
}

func listInferenceRevisions(ctx context.Context, client *runaiclient.Client, workloadID string) ([]workloads.Revision, error) {
	resp, httpResp, err := client.Workloads.RevisionsAPI.GetInferenceWorkloadRevisions(ctx, workloadID).Execute()
	if err != nil {
		evt := zerolog.Ctx(ctx).Warn().Err(err).
			Str("workload_id", workloadID).
			Str("call", "getInferenceWorkloadRevisions")
		if httpResp != nil {
			evt = evt.Int("status_code", httpResp.StatusCode)
		}
		evt.Msg("get_workload_history: revisions lookup failed")
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	return resp.Revisions, nil
}

func fetchRevisionFallback(ctx context.Context, client *runaiclient.Client, workloadID, revisionID string) (revisionOutput, error) {
	revision, httpResp, err := client.Workloads.RevisionsAPI.GetRevision(ctx, revisionID).Execute()
	notYours := fmt.Errorf("revision %q for workload %q: %w", revisionID, workloadID, runaimcp.ErrNotFound)

	if err != nil {
		if httpResp != nil && (httpResp.StatusCode == http.StatusNotFound || httpResp.StatusCode == http.StatusForbidden) {
			return revisionOutput{}, notYours
		}
		return revisionOutput{}, fmt.Errorf("get revision %q for workload %q: %w", revisionID, workloadID, runaimcp.MapHTTPError(httpResp, err))
	}
	if revision == nil || !strings.EqualFold(revision.WorkloadId, workloadID) {
		return revisionOutput{}, notYours
	}
	return mapRevision(*revision), nil
}

func attachRevisionDetail(ctx context.Context, client *runaiclient.Client, revisionID string, entry *revisionOutput) {
	if pods, ok := fetchRevisionPods(ctx, client, revisionID); ok {
		entry.Pods = &pods
	}
	if events, ok := fetchRevisionEvents(ctx, client, revisionID); ok {
		entry.Events = &events
	}
}

func fetchRevisionPods(ctx context.Context, client *runaiclient.Client, revisionID string) ([]podOutput, bool) {
	resp, httpResp, err := client.Workloads.PodsAPI.GetRevisionPods(ctx, revisionID).Execute()
	if err != nil {
		evt := zerolog.Ctx(ctx).Warn().Err(err).
			Str("revision_id", revisionID).
			Str("call", "getRevisionPods")
		if httpResp != nil {
			evt = evt.Int("status_code", httpResp.StatusCode)
		}
		evt.Msg("get_workload_history: revision pods fetch failed, omitting revision.pods")
		return nil, false
	}
	pods := make([]podOutput, 0, len(resp.Pods))
	for _, pod := range resp.Pods {
		pods = append(pods, mapPodOutput(pod))
	}
	return pods, true
}

func fetchRevisionEvents(ctx context.Context, client *runaiclient.Client, revisionID string) ([]eventOutput, bool) {
	resp, httpResp, err := client.Workloads.EventsAPI.GetRevisionEvents(ctx, revisionID).
		Limit(revisionEventsLimit).
		SortOrder("desc").
		Execute()
	if err != nil {
		evt := zerolog.Ctx(ctx).Warn().Err(err).
			Str("revision_id", revisionID).
			Str("call", "getRevisionEvents")
		if httpResp != nil {
			evt = evt.Int("status_code", httpResp.StatusCode)
		}
		evt.Msg("get_workload_history: revision events fetch failed, omitting revision.events")
		return nil, false
	}
	events := make([]eventOutput, 0, len(resp.Events))
	for _, event := range resp.Events {
		events = append(events, mapEvent(event))
	}
	return events, true
}

func mapHistoryRecords(records []workloads.HistoryRecord) []historyRecord {
	out := make([]historyRecord, 0, len(records))
	for _, record := range records {
		out = append(out, historyRecord{
			Meta: historyRecordMeta{
				CreationTimestamp: common.TimestampString(&record.Meta.CreationTimestamp),
				Type:              string(record.Meta.Type),
			},
			Spec: mapHistoryRecordSpec(record.Spec),
		})
	}
	return out
}

func mapHistoryRecordSpec(spec workloads.HistoryRecordSpec) historyRecordSpec {
	out := historyRecordSpec{}
	if phaseUpdate := spec.PhaseUpdate.Get(); phaseUpdate != nil {
		out.PhaseUpdate = &historyPhaseUpdate{
			Phase:        string(phaseUpdate.Phase),
			PhaseMessage: phaseUpdate.PhaseMessage,
		}
	}
	if event := spec.Event.Get(); event != nil {
		mapped := mapEvent(*event)
		out.Event = &mapped
	}
	return out
}

func mapEvent(event workloads.Event) eventOutput {
	out := eventOutput{
		CreatedAt: common.TimestampString(&event.CreatedAt),
		Type:      event.Type,
		Message:   event.Message,
	}
	if event.Reason != nil {
		out.Reason = *event.Reason
	}
	if event.Source != nil {
		out.Source = *event.Source
	}
	return out
}

func mapRevision(revision workloads.Revision) revisionOutput {
	return revisionOutput{
		ID:            revision.Id,
		Name:          revision.Name,
		CreatedAt:     common.TimestampString(&revision.CreatedAt),
		Phase:         string(revision.Phase),
		RequestedPods: mapRequestedPods(revision.RequestedPods),
	}
}

// mapRequestedPods exposes only {min, max, number} (drops parallelism/completions)
func mapRequestedPods(src *workloads.RequestedPods) *revisionRequestedPods {
	if src == nil {
		return nil
	}
	minPods, maxPods, number := src.Min.Get(), src.Max.Get(), src.Number.Get()
	if minPods == nil && maxPods == nil && number == nil {
		return nil
	}
	return &revisionRequestedPods{Min: minPods, Max: maxPods, Number: number}
}
