/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package specrun

import (
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
)

// This file holds the functions that fold a recorded event stream into the
// aggregated summary carried by a *_completed event. They are pure (no I/O) so
// the orchestration in execute.go stays free of result-parsing logic.

// preflightSummary tallies the preflight pipeline's per-step results into an
// overall verdict. A step is failed/warned/passed by its worst result status;
// a passing check carries no results (filtered as zero), so a result-free step
// counts as passed.
func preflightSummary(evs []orchestrator2.Event) orchestrator2.PreflightSummary {
	var sum orchestrator2.PreflightSummary
	for _, e := range evs {
		if e.Pipeline != "preflight" {
			continue
		}
		// A check that hard-errored (a Go error, not a pass/warn/fail verdict)
		// emits step_failed and carries no results. It must never be tallied as
		// a pass, so count it as a failure outright.
		if e.Kind == orchestrator2.EventKindStepFailed {
			sum.Failed++
			continue
		}
		if e.Kind != orchestrator2.EventKindStepSucceeded {
			continue
		}
		worst := orchestrator2.StatusOK
		for _, r := range e.Results {
			appendResourceIssues(&sum, r)
			appendClockSkew(&sum, r)
			appendClockDrift(&sum, r)
			appendRegistryIssues(&sum, r)
			appendSystemNodesAdvisory(&sum, r)
			if r.Status == orchestrator2.StatusFail {
				worst = orchestrator2.StatusFail
				break
			}
			if r.Status == orchestrator2.StatusWarn {
				worst = orchestrator2.StatusWarn
			}
		}
		switch worst {
		case orchestrator2.StatusFail:
			sum.Failed++
		case orchestrator2.StatusWarn:
			sum.Warned++
		default:
			sum.Passed++
		}
	}
	switch {
	case sum.Failed > 0:
		sum.Outcome = "failed"
	case sum.Warned > 0:
		sum.Outcome = "warning"
	default:
		sum.Outcome = "passed"
	}
	return sum
}

// appendResourceIssues extracts the per-node sizing breakdown a node-resource
// check carries in its result Data and folds it into the summary, so the full
// list can be rendered once at the end rather than inline per check.
func appendResourceIssues(sum *orchestrator2.PreflightSummary, r orchestrator2.ActionResult) {
	nodes, ok := r.Data["nodes"].([]map[string]any)
	if !ok {
		return
	}
	resource, _ := r.Data["resource"].(string)
	for _, n := range nodes {
		sum.ResourceIssues = append(sum.ResourceIssues, orchestrator2.NodeResourceIssue{
			Resource:    resource,
			Node:        asString(n["node"]),
			Role:        asString(n["role"]),
			Have:        asString(n["have"]),
			Recommended: asString(n["recommended"]),
		})
	}
}

// appendClockSkew folds the per-node heartbeat-lag breakdown the clock-skew
// check carries in its result Data into the summary.
func appendClockSkew(sum *orchestrator2.PreflightSummary, r orchestrator2.ActionResult) {
	nodes, ok := r.Data["clockSkewNodes"].([]map[string]any)
	if !ok {
		return
	}
	for _, n := range nodes {
		sum.ClockSkew = append(sum.ClockSkew, orchestrator2.NodeClockSkew{
			Node:          asString(n["node"]),
			Lag:           asString(n["lag"]),
			LastHeartbeat: asString(n["lastHeartbeat"]),
		})
	}
}

// appendClockDrift folds the per-node drift breakdown (nodes vs this machine's
// clock) the clock-drift check carries in its result Data into the summary.
func appendClockDrift(sum *orchestrator2.PreflightSummary, r orchestrator2.ActionResult) {
	nodes, ok := r.Data["driftNodes"].([]map[string]any)
	if !ok {
		return
	}
	if ref, ok := r.Data["reference"].(string); ok {
		sum.ClockDriftReference = ref
	}
	for _, n := range nodes {
		sum.ClockDrift = append(sum.ClockDrift, orchestrator2.NodeClockDrift{
			Node:          asString(n["node"]),
			Offset:        asString(n["offset"]),
			LastHeartbeat: asString(n["lastHeartbeat"]),
		})
	}
}

// appendRegistryIssues folds the registry check's findings into the summary.
func appendRegistryIssues(sum *orchestrator2.PreflightSummary, r orchestrator2.ActionResult) {
	issues, ok := r.Data["registryIssues"].([]map[string]any)
	if !ok {
		return
	}
	for _, i := range issues {
		sum.RegistryIssues = append(sum.RegistryIssues, orchestrator2.RegistryIssue{
			Registry: asString(i["registry"]),
			Issue:    asString(i["issue"]),
		})
	}
}

// appendSystemNodesAdvisory folds the system-node advisory message the check
// carries in its result Data into the summary, so the full text is shown once
// at the end while the per-check line stays short.
func appendSystemNodesAdvisory(sum *orchestrator2.PreflightSummary, r orchestrator2.ActionResult) {
	if a, ok := r.Data["systemNodesAdvisory"].(string); ok {
		sum.SystemNodesAdvisory = a
	}
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

// prereqSummary aggregates the prerequisites pipeline's per-step events into one
// outcome per dependency (keyed by flow name), then an overall verdict.
func prereqSummary(evs []orchestrator2.Event) orchestrator2.PrereqSummary {
	return groupSummary(evs, "prereq")
}

func featuresSummary(evs []orchestrator2.Event) orchestrator2.PrereqSummary {
	return groupSummary(evs, "features")
}

// runaiSummary folds the runai pipeline's events plus the spec into the control
// plane / cluster summary carried by a runai_completed event. Installed flags
// come from the per-step outcomes (so a partial install is reported
// accurately); versions/domain/name come from the spec.
func runaiSummary(evs []orchestrator2.Event, s *spec.Spec) orchestrator2.RunaiSummary {
	cpVersion, clusterVersion := runaiVersions(s)
	cpInstalled := runaiStepDone(evs, "install-control-plane-chart")
	clusterInstalled := runaiStepDone(evs, "install-cluster-chart")
	clusterRegistered := runaiStepDone(evs, "register-cluster")
	if s.Cluster != nil && s.Cluster.UsesSuppliedClusterIdentity() {
		clusterRegistered = true
	}

	var sum orchestrator2.RunaiSummary
	if s.ControlPlane != nil {
		sum.ControlPlane = &orchestrator2.RunaiControlPlane{
			Installed: cpInstalled,
			Version:   cpVersion,
			Domain:    s.ControlPlane.Domain,
			URL:       httpsURL(s.ControlPlane.Domain),
		}
	}
	if s.Cluster != nil {
		sum.Cluster = &orchestrator2.RunaiCluster{
			Installed:  clusterInstalled,
			Registered: clusterRegistered,
			Name:       s.Cluster.Name,
			Version:    clusterVersion,
		}
	}

	switch {
	case s.DryRun:
		sum.Outcome = "skipped"
	case pipelineFailed(evs, "runai"):
		if cpInstalled || clusterInstalled {
			sum.Outcome = "partial"
		} else {
			sum.Outcome = "failed"
		}
	default:
		sum.Outcome = "installed"
	}
	return sum
}

// runSummary folds the whole event stream plus the spec into the final run
// summary carried by a run_completed event. installErr is non-nil on a failed
// install.
func runSummary(evs []orchestrator2.Event, s *spec.Spec, duration time.Duration, installErr error) orchestrator2.RunSummary {
	sum := orchestrator2.RunSummary{
		Outcome:         "succeeded",
		Command:         string(s.Command),
		DryRun:          s.DryRun,
		DurationSeconds: int64(duration.Seconds()),
		Warnings:        warningCount(evs),
		Pipelines:       runPipelines(evs, s),
	}
	if installErr != nil {
		sum.Outcome = "failed"
		sum.Failure = runFailure(evs, installErr)
	}
	if s.ControlPlane != nil {
		cpVersion, _ := runaiVersions(s)
		sum.ControlPlane = &orchestrator2.RunaiControlPlane{
			Installed: runaiStepDone(evs, "install-control-plane-chart"),
			Version:   cpVersion,
			Domain:    s.ControlPlane.Domain,
			URL:       httpsURL(s.ControlPlane.Domain),
		}
	}
	if s.Cluster != nil {
		_, clusterVersion := runaiVersions(s)
		clusterRegistered := runaiStepDone(evs, "register-cluster")
		if s.Cluster.UsesSuppliedClusterIdentity() {
			clusterRegistered = true
		}
		sum.Cluster = &orchestrator2.RunaiCluster{
			Installed:  runaiStepDone(evs, "install-cluster-chart"),
			Registered: clusterRegistered,
			Name:       s.Cluster.Name,
			Version:    clusterVersion,
		}
	}
	return sum
}

// runPipelines rolls up each pipeline's own summary, including only the
// pipelines that actually ran.
func runPipelines(evs []orchestrator2.Event, s *spec.Spec) *orchestrator2.RunPipelines {
	p := &orchestrator2.RunPipelines{}
	if pipelineStarted(evs, "preflight") {
		ps := preflightSummary(evs)
		p.Preflight = &ps
	}
	if pipelineStarted(evs, "prereq") {
		pr := prereqSummary(evs)
		p.Prerequisites = &pr
	}
	if pipelineStarted(evs, "runai") {
		rs := runaiSummary(evs, s)
		p.Runai = &rs
	}
	if pipelineStarted(evs, "features") {
		fs := featuresSummary(evs)
		p.Features = &fs
	}
	return p
}

// runFailure locates the failing step from the event stream, falling back to the
// raw error when no step-level failure was recorded.
func runFailure(evs []orchestrator2.Event, installErr error) *orchestrator2.RunFailure {
	var failure *orchestrator2.RunFailure
	for _, e := range evs {
		if e.Kind == orchestrator2.EventKindStepFailed {
			failure = &orchestrator2.RunFailure{
				Pipeline: e.Pipeline,
				Flow:     e.Flow,
				Step:     e.Step,
				Error:    errString(e.Error),
			}
		}
	}
	if failure == nil {
		failure = &orchestrator2.RunFailure{Error: errString(installErr)}
	}
	return failure
}

// runaiVersions returns the control-plane and cluster versions from the spec,
// the cluster falling back to the control-plane version (mirroring the plan).
func runaiVersions(s *spec.Spec) (controlPlane, cluster string) {
	if s.ControlPlane != nil {
		controlPlane = s.ControlPlane.Version
	}
	if s.Cluster != nil {
		cluster = s.Cluster.Version
		if cluster == "" {
			cluster = controlPlane
		}
	}
	return controlPlane, cluster
}

// runaiStepDone reports whether a runai-pipeline step completed: it succeeded,
// or was skipped because it was already installed.
func runaiStepDone(evs []orchestrator2.Event, step string) bool {
	for _, e := range evs {
		if e.Pipeline != "runai" || e.Step != step {
			continue
		}
		if e.Kind == orchestrator2.EventKindStepSucceeded {
			return true
		}
		if e.Kind == orchestrator2.EventKindStepSkipped && e.SkipReason == orchestrator2.SkipReasonAlreadyInstalled {
			return true
		}
	}
	return false
}

// pipelineStarted reports whether the named pipeline emitted a start event.
func pipelineStarted(evs []orchestrator2.Event, pipeline string) bool {
	for _, e := range evs {
		if e.Kind == orchestrator2.EventKindPipelineStarted && e.Pipeline == pipeline {
			return true
		}
	}
	return false
}

// pipelineFailed reports whether the named pipeline had a step fail.
func pipelineFailed(evs []orchestrator2.Event, pipeline string) bool {
	for _, e := range evs {
		if e.Kind == orchestrator2.EventKindStepFailed && e.Pipeline == pipeline {
			return true
		}
	}
	return false
}

// warningCount counts the warning verdicts an action reported across all steps
// (preflight warnings and any install-time warnings alike).
func warningCount(evs []orchestrator2.Event) int {
	n := 0
	for _, e := range evs {
		if e.Kind != orchestrator2.EventKindStepSucceeded {
			continue
		}
		for _, r := range e.Results {
			if r.Status == orchestrator2.StatusWarn {
				n++
			}
		}
	}
	return n
}

func httpsURL(domain string) string {
	if domain == "" {
		return ""
	}
	return "https://" + domain
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// groupSummary aggregates the per-flow outcome of one install pipeline (the
// prereq or features pipeline) into a single summary.
func groupSummary(evs []orchestrator2.Event, pipeline string) orchestrator2.PrereqSummary {
	type agg struct {
		failed, installed, alreadyPresent bool
	}
	var order []string
	byFlow := map[string]*agg{}
	get := func(flow string) *agg {
		a, ok := byFlow[flow]
		if !ok {
			a = &agg{}
			byFlow[flow] = a
			order = append(order, flow)
		}
		return a
	}
	for _, e := range evs {
		if e.Pipeline != pipeline || e.Flow == "" {
			continue
		}
		switch e.Kind {
		case orchestrator2.EventKindStepSucceeded:
			get(e.Flow).installed = true
		case orchestrator2.EventKindStepFailed:
			get(e.Flow).failed = true
		case orchestrator2.EventKindStepSkipped:
			if e.SkipReason == orchestrator2.SkipReasonAlreadyInstalled {
				get(e.Flow).alreadyPresent = true
			} else {
				get(e.Flow)
			}
		}
	}

	var sum orchestrator2.PrereqSummary
	for _, flow := range order {
		a := byFlow[flow]
		var status string
		switch {
		case a.failed:
			status, sum.Failed = "failed", sum.Failed+1
		case a.installed:
			status, sum.Installed = "installed", sum.Installed+1
		case a.alreadyPresent:
			status, sum.AlreadyPresent = "already_present", sum.AlreadyPresent+1
		default:
			status, sum.Skipped = "skipped", sum.Skipped+1
		}
		sum.Components = append(sum.Components, orchestrator2.PrereqComponent{Name: flow, Status: status})
	}
	switch {
	case sum.Failed > 0 && sum.Installed == 0:
		sum.Outcome = "failed"
	case sum.Failed > 0:
		sum.Outcome = "partial"
	case sum.Installed > 0:
		sum.Outcome = "installed"
	default:
		sum.Outcome = "completed"
	}
	return sum
}
