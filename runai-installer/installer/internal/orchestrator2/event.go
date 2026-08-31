/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orchestrator2

import (
	"encoding/json"
	"fmt"
)

type EventKind string

const (
	EventKindPipelineStarted   EventKind = "pipeline_started"
	EventKindPipelineSucceeded EventKind = "pipeline_succeeded"
	EventKindPipelineFailed    EventKind = "pipeline_failed"
	EventKindFlowStarted       EventKind = "flow_started"
	EventKindFlowSucceeded     EventKind = "flow_succeeded"
	EventKindFlowFailed        EventKind = "flow_failed"
	EventKindStepStarted       EventKind = "step_started"
	EventKindStepSucceeded     EventKind = "step_succeeded"
	EventKindStepFailed        EventKind = "step_failed"
	EventKindStepSkipped       EventKind = "step_skipped"

	EventKindInputValidationFailed EventKind = "input_validation_failed"
	EventKindPreflightCompleted    EventKind = "preflight_completed"
	EventKindPrereqCompleted       EventKind = "prereq_completed"
	EventKindRunaiCompleted        EventKind = "runai_completed"
	EventKindFeaturesCompleted     EventKind = "features_completed"
	EventKindRunCompleted          EventKind = "run_completed"
	EventKindCompatFallback        EventKind = "compat_fallback"
	EventKindSelfSignedTLS         EventKind = "self_signed_tls"
)

// CompatFallback is carried by an EventKindCompatFallback event. It is a
// warning, not a failure: the requested Run:ai version is not in the
// compatibility matrix, so prerequisite versions were taken from the latest
// known release instead. Components lists what will actually be installed.
type CompatFallback struct {
	RequestedVersion string            `json:"requestedVersion"`
	UsingVersion     string            `json:"usingVersion"`
	Components       []PrereqComponent `json:"components"`
}

type SelfSignedTLSWarning struct {
	OutputDir string `json:"outputDir"`
	Message   string `json:"message"`
}

func NewSelfSignedTLSWarning(outputDir string) *SelfSignedTLSWarning {
	return &SelfSignedTLSWarning{
		OutputDir: outputDir,
		Message: fmt.Sprintf(
			"Self-signed certificate files will be saved to: %s. Note - using self-signed certificates comes with risks, and is typically not recommended for production environments. Use at your own risk.",
			outputDir,
		),
	}
}

// PreflightSummary is the aggregated outcome of the preflight pipeline, carried
// by an EventKindPreflightCompleted event. Outcome is the overall verdict
// derived from the per-check counts.
type PreflightSummary struct {
	Outcome string `json:"outcome"` // passed | warning | failed
	Passed  int    `json:"passed"`
	Warned  int    `json:"warned"`
	Failed  int    `json:"failed"`
	// ResourceIssues aggregates per-node sizing findings from the node-resource
	// checks so the full list can be rendered once, prettily, in the summary
	// (the per-check warning stays a single line).
	ResourceIssues []NodeResourceIssue `json:"resourceIssues,omitempty"`
	// ClockSkew aggregates the per-node heartbeat-lag findings from the
	// clock-skew check, rendered the same way as ResourceIssues.
	ClockSkew []NodeClockSkew `json:"clockSkew,omitempty"`
	// ClockDrift aggregates the per-node findings from the clock-drift check
	// (nodes vs this machine's clock); ClockDriftReference is the machine clock
	// the comparison was made against (shown in local zone and UTC).
	ClockDrift          []NodeClockDrift `json:"clockDrift,omitempty"`
	ClockDriftReference string           `json:"clockDriftReference,omitempty"`
	// RegistryIssues aggregates the findings from the registry check.
	RegistryIssues []RegistryIssue `json:"registryIssues,omitempty"`
	// SystemNodesAdvisory is the soft advisory surfaced when no system-node
	// selection is configured on a GPU install. It is rendered in the summary
	// so the per-check line can stay short.
	SystemNodesAdvisory string `json:"systemNodesAdvisory,omitempty"`
}

// NodeResourceIssue is a single node that is below the recommended amount of a
// resource, aggregated into the preflight summary.
type NodeResourceIssue struct {
	Resource    string `json:"resource"` // human label, e.g. "disk (ephemeral storage)"
	Node        string `json:"node"`
	Role        string `json:"role"`
	Have        string `json:"have"`
	Recommended string `json:"recommended"`
}

// NodeClockSkew is a single node whose heartbeat lags behind the freshest node,
// indicating clock skew or an unreachable/dead node.
type NodeClockSkew struct {
	Node          string `json:"node"`
	Lag           string `json:"lag"`           // how far its heartbeat is behind the freshest node
	LastHeartbeat string `json:"lastHeartbeat"` // RFC3339 lease renew time
}

// NodeClockDrift is a single node whose clock differs from this installer
// machine's clock by more than the threshold.
type NodeClockDrift struct {
	Node          string `json:"node"`
	Offset        string `json:"offset"`        // e.g. "5m0s behind this machine"
	LastHeartbeat string `json:"lastHeartbeat"` // RFC3339 (UTC) lease renew time
}

// RegistryIssue is a single problem found by the registry check.
type RegistryIssue struct {
	Registry string `json:"registry"`
	Issue    string `json:"issue"`
}

// PrereqSummary is the aggregated outcome of the prerequisites pipeline,
// carried by an EventKindPrereqCompleted event. Outcome is the overall verdict
// derived from the per-component statuses.
type PrereqSummary struct {
	Outcome        string            `json:"outcome"` // installed | completed | partial | failed
	Installed      int               `json:"installed"`
	AlreadyPresent int               `json:"alreadyPresent"`
	Skipped        int               `json:"skipped"`
	Failed         int               `json:"failed"`
	Components     []PrereqComponent `json:"components"`
}

// PrereqComponent is a single prerequisite's outcome within a PrereqSummary.
type PrereqComponent struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // installed | already_present | skipped | failed
	Version string `json:"version,omitempty"`
}

// RunaiSummary is the aggregated outcome of the runai pipeline (control plane +
// cluster), carried by an EventKindRunaiCompleted event. ControlPlane/Cluster
// are nil when that side is not part of the install (CP-only or cluster-only).
type RunaiSummary struct {
	Outcome      string             `json:"outcome"` // installed | partial | failed | skipped
	ControlPlane *RunaiControlPlane `json:"controlPlane,omitempty"`
	Cluster      *RunaiCluster      `json:"cluster,omitempty"`
}

// RunaiControlPlane describes the installed control plane.
type RunaiControlPlane struct {
	Installed bool   `json:"installed"`
	Version   string `json:"version,omitempty"`
	Domain    string `json:"domain,omitempty"`
	URL       string `json:"url,omitempty"`
}

// RunaiCluster describes the installed/registered Run:ai cluster.
type RunaiCluster struct {
	Installed  bool   `json:"installed"`
	Registered bool   `json:"registered"`
	Name       string `json:"name,omitempty"`
	Version    string `json:"version,omitempty"`
}

// RunSummary is the single terminal summary of the whole run, carried by an
// EventKindRunCompleted event. It rolls up the per-pipeline summaries plus the
// key access facts, and is the extensible "final" document automation watches.
type RunSummary struct {
	Outcome         string             `json:"outcome"` // succeeded | failed
	Command         string             `json:"command,omitempty"`
	DryRun          bool               `json:"dryRun"`
	DurationSeconds int64              `json:"durationSeconds,omitempty"`
	Warnings        int                `json:"warnings,omitempty"`
	Pipelines       *RunPipelines      `json:"pipelines,omitempty"`
	ControlPlane    *RunaiControlPlane `json:"controlPlane,omitempty"`
	Cluster         *RunaiCluster      `json:"cluster,omitempty"`
	// Failure locates the failing step on a failed run (nil on success).
	Failure *RunFailure `json:"failure,omitempty"`
}

// RunPipelines is the per-pipeline rollup embedded in a RunSummary. Each field
// reuses the pipeline's own summary type and is nil when that pipeline did not
// run.
type RunPipelines struct {
	Preflight     *PreflightSummary `json:"preflight,omitempty"`
	Prerequisites *PrereqSummary    `json:"prerequisites,omitempty"`
	Runai         *RunaiSummary     `json:"runai,omitempty"`
	Features      *PrereqSummary    `json:"features,omitempty"`
}

// RunFailure identifies where a failed run stopped, using the machine names (the
// same stable identifiers as the other events).
type RunFailure struct {
	Pipeline string `json:"pipeline,omitempty"`
	Flow     string `json:"flow,omitempty"`
	Step     string `json:"step,omitempty"`
	Error    string `json:"error,omitempty"`
}

// ValidationFinding is a single input-validation problem carried by an
// EventKindInputValidationFailed event. It is intentionally dependency-free so
// the event stays self-contained; producers map their own findings onto it.
type ValidationFinding struct {
	Check   string `json:"check"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

type SkipReason string

const (
	SkipReasonDryRun           SkipReason = "dry_run"
	SkipReasonAlreadyInstalled SkipReason = "already_installed"
	SkipReasonUserSelected     SkipReason = "user_selected"
)

type Event struct {
	Kind     EventKind
	Pipeline string
	Flow     string
	Step     string
	// PipelineTitle, FlowTitle, and StepTitle are human-readable labels for the
	// text UI, stamped by the runner from the pipeline/flow/step. They are
	// intentionally absent from MarshalJSON: JSON/NDJSON keeps the raw machine
	// names (Pipeline/Flow/Step) as the stable contract for automation.
	PipelineTitle string
	FlowTitle     string
	StepTitle     string
	SkipReason    SkipReason
	Error         error
	Output        string
	Findings      []ValidationFinding
	Results       []ActionResult
	Preflight     *PreflightSummary
	Prereq        *PrereqSummary
	// Features carries the aggregated outcome of the features pipeline (same
	// shape as Prereq), on an EventKindFeaturesCompleted event.
	Features *PrereqSummary
	// Runai carries the runai pipeline summary on an EventKindRunaiCompleted
	// event; Run carries the whole-run summary on an EventKindRunCompleted event.
	Runai         *RunaiSummary
	Run           *RunSummary
	Compat        *CompatFallback
	SelfSignedTLS *SelfSignedTLSWarning
}

func (e Event) MarshalJSON() ([]byte, error) {
	type wire struct {
		Kind          EventKind             `json:"kind"`
		Pipeline      string                `json:"pipeline,omitempty"`
		Flow          string                `json:"flow,omitempty"`
		Step          string                `json:"step,omitempty"`
		SkipReason    SkipReason            `json:"skipReason,omitempty"`
		Error         string                `json:"error,omitempty"`
		Output        string                `json:"output,omitempty"`
		Findings      []ValidationFinding   `json:"findings,omitempty"`
		Results       []ActionResult        `json:"results,omitempty"`
		Preflight     *PreflightSummary     `json:"preflight,omitempty"`
		Prereq        *PrereqSummary        `json:"prereq,omitempty"`
		Features      *PrereqSummary        `json:"features,omitempty"`
		Runai         *RunaiSummary         `json:"runai,omitempty"`
		Run           *RunSummary           `json:"run,omitempty"`
		Compat        *CompatFallback       `json:"compat,omitempty"`
		SelfSignedTLS *SelfSignedTLSWarning `json:"selfSignedTLS,omitempty"`
	}
	w := wire{
		Kind:          e.Kind,
		Pipeline:      e.Pipeline,
		Flow:          e.Flow,
		Step:          e.Step,
		SkipReason:    e.SkipReason,
		Output:        e.Output,
		Findings:      e.Findings,
		Results:       e.Results,
		Preflight:     e.Preflight,
		Prereq:        e.Prereq,
		Features:      e.Features,
		Runai:         e.Runai,
		Run:           e.Run,
		Compat:        e.Compat,
		SelfSignedTLS: e.SelfSignedTLS,
	}
	if e.Error != nil {
		w.Error = e.Error.Error()
	}
	return json.Marshal(w)
}

type EventSink interface {
	HandleEvent(e Event) error
}
