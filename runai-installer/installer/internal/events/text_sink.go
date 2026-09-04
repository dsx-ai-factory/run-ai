/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package events

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/display"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/terminal"
)

// label prefers the stamped human title and falls back to humanizing the
// machine name, so an event that carries no title (e.g. hand-built) still
// renders a readable label rather than a blank.
func label(title, name string) string {
	if title != "" {
		return title
	}
	return display.Humanize(name)
}

// textSink renders events through a Terminal. Each step gets a live spinner
// that finalizes as success/warning/failure on the matching end event.
//
// Spinner output goes to the Terminal's stdout writer, including failures, so
// the progress stream stays in one place. Use NDJSON output if you need
// stdout/stderr separation for automation.
type textSink struct {
	t  *terminal.Terminal
	sp *terminal.Spinner
	// area replaces the spinner once a step reports progress (a spinner is
	// single-line; the tail needs multiple lines redrawn in place).
	area *terminal.LiveArea
	// mu guards the live-area state below (touched by Progress and the ticker).
	mu          sync.Mutex
	activeLabel string
	tail        []string
	frame       int
	progStart   time.Time
}

func newTextSink(t *terminal.Terminal) *textSink {
	return &textSink{t: t}
}

func (s *textSink) Handle(e orchestrator2.Event) error {
	switch e.Kind {
	case orchestrator2.EventKindPipelineStarted:
		s.t.Heading(label(e.PipelineTitle, e.Pipeline))
	case orchestrator2.EventKindPipelineSucceeded:
		s.t.Success("%s complete", label(e.PipelineTitle, e.Pipeline))
	case orchestrator2.EventKindPipelineFailed:
		s.t.Error("%s failed: %s", label(e.PipelineTitle, e.Pipeline), errString(e.Error))

	case orchestrator2.EventKindFlowStarted:
		s.t.Info("%s", label(e.FlowTitle, e.Flow))
	case orchestrator2.EventKindFlowSucceeded:
		// silent — covered by step events
	case orchestrator2.EventKindFlowFailed:
		s.t.Error("%s failed (%s): %s", label(e.FlowTitle, e.Flow), label(e.StepTitle, e.Step), errString(e.Error))

	case orchestrator2.EventKindStepStarted:
		stepLabel := label(e.StepTitle, e.Step)
		sp, err := s.t.StartSpinner(stepLabel)
		if err != nil {
			return err
		}
		s.mu.Lock()
		s.sp = sp
		s.activeLabel = stepLabel
		s.mu.Unlock()

	case orchestrator2.EventKindStepSucceeded:
		stepLabel := label(e.StepTitle, e.Step)
		s.finishStep(statusSuccess, stepLabel)
		if e.Output != "" {
			s.t.Print("%s", e.Output)
		}
		s.renderResults(stepLabel, e.Results)
	case orchestrator2.EventKindStepSkipped:
		msg := fmt.Sprintf("%s skipped (%s)", label(e.StepTitle, e.Step), e.SkipReason)
		s.finishStep(statusSkip, msg)
		switch {
		case e.SkipReason == orchestrator2.SkipReasonDryRun && e.Output != "":
			s.t.Print("%s", e.Output)
		case e.Error != nil:
			s.t.Print("%s", errString(e.Error))
		}
	case orchestrator2.EventKindStepFailed:
		msg := fmt.Sprintf("%s failed (%s): %s", label(e.StepTitle, e.Step), e.Output, errString(e.Error))
		s.finishStep(statusError, msg)

	case orchestrator2.EventKindInputValidationFailed:
		s.renderInputValidationFailed(e.Findings)

	case orchestrator2.EventKindPreflightCompleted:
		s.renderPreflightSummary(e.Preflight)

	case orchestrator2.EventKindPrereqCompleted:
		s.renderGroupSummary("Prerequisites", e.Prereq)

	case orchestrator2.EventKindFeaturesCompleted:
		s.renderGroupSummary("Features", e.Features)

	case orchestrator2.EventKindRunaiCompleted:
		s.renderRunaiSummary(e.Runai)

	case orchestrator2.EventKindRunCompleted:
		s.renderRunSummary(e.Run)

	case orchestrator2.EventKindCompatFallback:
		s.renderCompatFallback(e.Compat)

	case orchestrator2.EventKindSelfSignedTLS:
		if e.SelfSignedTLS != nil {
			s.t.Warning("%s", e.SelfSignedTLS.Message)
		}
	}
	return nil
}

// renderRunaiSummary prints the runai pipeline verdict: a status line per
// installed side (control plane / cluster), at a level matching the outcome.
func (s *textSink) renderRunaiSummary(p *orchestrator2.RunaiSummary) {
	if p == nil {
		return
	}
	if p.Outcome == "skipped" {
		s.t.Info("Run:ai: dry run, no changes were applied")
		return
	}
	if cp := p.ControlPlane; cp != nil {
		if cp.Installed {
			s.t.Success("Run:ai control plane is up%s%s", versionSuffix(cp.Version), urlSuffix(cp.URL))
		} else {
			s.t.Error("Run:ai control plane installation did not complete")
		}
	}
	if cl := p.Cluster; cl != nil {
		if cl.Installed {
			s.t.Success("Run:ai cluster %q is registered and running%s", cl.Name, versionSuffix(cl.Version))
		} else {
			s.t.Error("Run:ai cluster %q installation did not complete", cl.Name)
		}
	}
}

// runWorkLabel names the work the final banner reports on, so a command that
// only builds a bundle does not report an installation.
func runWorkLabel(command string) string {
	switch command {
	case string(spec.CommandAirgapPack):
		return "Air-gapped bundle"
	case string(spec.CommandAirgapLoad):
		return "Air-gapped load"
	}
	return "Installation"
}

// renderRunSummary prints the terminal banner for the whole run: a success block
// with the access facts and pipeline rollup, or a failure block pointing at the
// failing step.
func (s *textSink) renderRunSummary(r *orchestrator2.RunSummary) {
	if r == nil {
		return
	}
	duration := time.Duration(r.DurationSeconds) * time.Second

	work := runWorkLabel(r.Command)

	if r.Outcome == "failed" {
		s.t.Heading(work + " failed")
		s.t.Error("NVIDIA Run:ai %s failed", strings.ToLower(work))
		if f := r.Failure; f != nil {
			if loc := failureLocation(f); loc != "" {
				s.t.Print("  Failed at: %s", loc)
			}
		}
		s.t.Print("  Ran for %s (see the errors above for details)", duration)
		return
	}

	s.t.Heading(work + " complete")
	s.t.Success("NVIDIA Run:ai %s completed%s", strings.ToLower(work), dryRunSuffix(r.DryRun))
	if cp := r.ControlPlane; cp != nil && cp.URL != "" {
		s.t.Print("  Control plane: %s%s", cp.URL, versionSuffix(cp.Version))
	}
	if cl := r.Cluster; cl != nil && cl.Name != "" {
		s.t.Print("  Cluster: %s%s", cl.Name, versionSuffix(cl.Version))
	}
	if rollup := runRollup(r.Pipelines); rollup != "" {
		s.t.Print("  %s", rollup)
	}
	s.t.Print("  Completed in %s%s", duration, warningSuffix(r.Warnings))
}

// failureLocation renders the failing step as humanized pipeline › flow › step.
func failureLocation(f *orchestrator2.RunFailure) string {
	var parts []string
	for _, name := range []string{f.Pipeline, f.Flow, f.Step} {
		if name != "" {
			parts = append(parts, display.Humanize(name))
		}
	}
	return strings.Join(parts, " › ")
}

// runRollup summarizes the per-pipeline outcomes on a single line, including
// only the pipelines that produced something worth showing.
func runRollup(p *orchestrator2.RunPipelines) string {
	if p == nil {
		return ""
	}
	var parts []string
	if p.Preflight != nil {
		parts = append(parts, fmt.Sprintf("Preflight %d passed, %d warning(s)", p.Preflight.Passed, p.Preflight.Warned))
	}
	if p.Prerequisites != nil && (p.Prerequisites.Installed > 0 || p.Prerequisites.AlreadyPresent > 0) {
		parts = append(parts, fmt.Sprintf("Prerequisites %d installed", p.Prerequisites.Installed))
	}
	if p.Features != nil && p.Features.Installed > 0 {
		parts = append(parts, fmt.Sprintf("Features %d installed", p.Features.Installed))
	}
	return strings.Join(parts, " · ")
}

func versionSuffix(version string) string {
	if version == "" {
		return ""
	}
	return " (" + version + ")"
}

func urlSuffix(url string) string {
	if url == "" {
		return ""
	}
	return " — " + url
}

func dryRunSuffix(dryRun bool) string {
	if dryRun {
		return " (dry run)"
	}
	return ""
}

func warningSuffix(warnings int) string {
	switch {
	case warnings == 1:
		return " (1 warning)"
	case warnings > 1:
		return fmt.Sprintf(" (%d warnings)", warnings)
	default:
		return ""
	}
}

// renderGroupSummary prints the aggregated verdict of a prerequisite/feature
// pipeline, rendered at a level matching the outcome (success / warning /
// error).
func (s *textSink) renderGroupSummary(label string, p *orchestrator2.PrereqSummary) {
	if p == nil {
		return
	}
	msg := fmt.Sprintf("%s: %d installed, %d already present, %d skipped, %d failed", label, p.Installed, p.AlreadyPresent, p.Skipped, p.Failed)
	switch p.Outcome {
	case "failed":
		s.t.Error("%s", msg)
	case "partial":
		s.t.Warning("%s", msg)
	default:
		s.t.Success("%s", msg)
	}
}

// renderCompatFallback warns that the requested Run:ai version is unknown to the
// compatibility matrix and lists the prerequisite versions that will be
// installed from the latest known release instead. It is a warning, not a
// failure.
func (s *textSink) renderCompatFallback(c *orchestrator2.CompatFallback) {
	if c == nil {
		return
	}
	s.t.Warning("Run:ai version %q is not in the compatibility matrix; using prerequisite versions from %s:", c.RequestedVersion, c.UsingVersion)
	for _, comp := range c.Components {
		s.t.Warning("  %s %s", comp.Name, comp.Version)
	}
}

// renderPreflightSummary prints the aggregated preflight verdict, rendered at a
// level matching the outcome (success / warning / error).
func (s *textSink) renderPreflightSummary(p *orchestrator2.PreflightSummary) {
	if p == nil {
		return
	}
	msg := fmt.Sprintf("Preflight checks: %d passed, %d warning(s), %d failed", p.Passed, p.Warned, p.Failed)
	switch p.Outcome {
	case "failed":
		s.t.Error("%s", msg)
	case "warning":
		s.t.Warning("%s", msg)
	default:
		s.t.Success("%s", msg)
	}
	s.renderResourceIssues(p.ResourceIssues)
	s.renderClockSkew(p.ClockSkew)
	s.renderClockDrift(p.ClockDrift, p.ClockDriftReference)
	s.renderRegistryIssues(p.RegistryIssues)
	s.renderSystemNodesAdvisory(p.SystemNodesAdvisory)
}

// renderSystemNodesAdvisory prints the system-node placement advisory under the
// preflight summary, so the per-check line can stay short.
func (s *textSink) renderSystemNodesAdvisory(advisory string) {
	if advisory == "" {
		return
	}
	s.t.Warning("system nodes: %s", advisory)
}

// renderRegistryIssues prints the aggregated registry-check findings under the
// preflight summary.
func (s *textSink) renderRegistryIssues(issues []orchestrator2.RegistryIssue) {
	if len(issues) == 0 {
		return
	}
	s.t.Warning("registry issues:")
	for _, i := range issues {
		s.t.Print("    - %s: %s", i.Registry, i.Issue)
	}
}

// renderClockDrift prints the aggregated node clock-drift findings (nodes vs
// this machine's clock) under the preflight summary. The reference makes the
// comparison basis explicit, including timezone, so a zone difference is not
// mistaken for drift.
func (s *textSink) renderClockDrift(items []orchestrator2.NodeClockDrift, reference string) {
	if len(items) == 0 {
		return
	}
	if reference != "" {
		s.t.Warning("clock drift (vs this machine's clock %s):", reference)
	} else {
		s.t.Warning("clock drift (vs this machine's clock):")
	}
	for _, i := range items {
		s.t.Print("    - %s: %s (last heartbeat %s)", i.Node, i.Offset, i.LastHeartbeat)
	}
}

// renderClockSkew prints the aggregated node clock-skew findings under the
// preflight summary, mirroring renderResourceIssues.
func (s *textSink) renderClockSkew(items []orchestrator2.NodeClockSkew) {
	if len(items) == 0 {
		return
	}
	s.t.Warning("clock skew:")
	for _, i := range items {
		s.t.Print("    - %s: %s behind (last heartbeat %s)", i.Node, i.Lag, i.LastHeartbeat)
	}
}

// renderResourceIssues prints the aggregated node-sizing findings as a grouped,
// indented list under the preflight summary, so the full per-node detail is
// shown once at the end rather than inline on each check.
func (s *textSink) renderResourceIssues(issues []orchestrator2.NodeResourceIssue) {
	if len(issues) == 0 {
		return
	}
	var order []string
	byResource := map[string][]orchestrator2.NodeResourceIssue{}
	for _, i := range issues {
		if _, seen := byResource[i.Resource]; !seen {
			order = append(order, i.Resource)
		}
		byResource[i.Resource] = append(byResource[i.Resource], i)
	}

	s.t.Warning("Resource issues detected:")
	for _, res := range order {
		s.t.Print("  %s:", res)
		for _, i := range byResource[res] {
			s.t.Print("    - %s: %s (recommended %s)", i.Node, i.Have, i.Recommended)
		}
	}
}

// renderResults prints an action's non-trivial results (warnings/failures with
// reason/remediation), each prefixed with the step label so it is clear which
// check the line belongs to.
func (s *textSink) renderResults(stepLabel string, results []orchestrator2.ActionResult) {
	for _, r := range results {
		msg := r.Reason
		if r.Remediation != "" {
			msg = fmt.Sprintf("%s (%s)", r.Reason, r.Remediation)
		}
		if stepLabel != "" && msg != "" {
			msg = fmt.Sprintf("%s: %s", stepLabel, msg)
		}
		switch r.Status {
		case orchestrator2.StatusFail:
			s.t.Error("%s", msg)
		case orchestrator2.StatusWarn:
			s.t.Warning("%s", msg)
		default:
			if msg != "" {
				s.t.Print("%s", msg)
			}
		}
	}
}

// renderInputValidationFailed prints a standalone block listing every input
// validation problem, so the user sees all of them at once before the run
// aborts.
func (s *textSink) renderInputValidationFailed(findings []orchestrator2.ValidationFinding) {
	s.t.Heading("Input validation failed")
	for _, f := range findings {
		if f.Field != "" {
			s.t.Error("%s: %s", f.Field, f.Message)
		} else {
			s.t.Error("%s", f.Message)
		}
	}
}

// step-completion kinds passed to finishStep.
const (
	statusSuccess = "success"
	statusWarning = "warning"
	statusSkip    = "skip"
	statusError   = "error"
)

// finishStep renders a step's final state: via the spinner, or (if it switched
// to a live area) by clearing the area and printing a plain status line.
func (s *textSink) finishStep(kind, msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sp != nil {
		s.finishSpinner(kind, msg)
		s.sp = nil
		s.activeLabel = ""
		return
	}
	if s.area != nil {
		s.area.Stop() // clears the tail; the ticker exits on area == nil
		s.area = nil
		s.activeLabel = ""
		s.tail = nil
		s.frame = 0
		s.progStart = time.Time{}
	}
	s.printStatus(kind, msg)
}

func (s *textSink) finishSpinner(kind, msg string) {
	switch kind {
	case statusError:
		s.sp.Fail(msg)
	case statusWarning:
		s.sp.Warning(msg)
	case statusSkip:
		s.sp.Skip(msg)
	default:
		s.sp.Success(msg)
	}
}

func (s *textSink) printStatus(kind, msg string) {
	switch kind {
	case statusError:
		s.t.Error("%s", msg)
	case statusWarning:
		s.t.Warning("%s", msg)
	case statusSkip:
		s.t.Skipped("%s", msg)
	default:
		s.t.Success("%s", msg)
	}
}

// Progress shows the latest output lines beneath the in-flight step. The first
// call hands the spinner off to a live area. No-op off-TTY or between steps.
func (s *textSink) Progress(lines []string) {
	if !s.t.IsTTY() {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sp == nil && s.area == nil {
		return
	}
	if s.area == nil {
		s.sp.StopSilently()
		s.sp = nil
		area, err := s.t.StartArea()
		if err != nil {
			return
		}
		s.area = area
		s.progStart = time.Now()
		s.startAnimation()
	}
	s.tail = lines
	s.renderProgress()
}

// renderProgress redraws the area: an animated spinner header and the latest
// lines. Lines are truncated to the terminal width so nothing wraps (a wrapped
// line desyncs the area's redraw). Caller must hold s.mu with s.area set.
func (s *textSink) renderProgress() {
	max := s.t.Width() - 1
	var b strings.Builder
	// Truncate only the label, not the glyph (it carries color codes); reserve
	// 3 columns for the 2-cell glyph plus a space.
	b.WriteString(terminal.SpinnerFrame(s.frame))
	b.WriteString(" ")
	b.WriteString(truncateLine(labelWithTimer(s.activeLabel, s.progStart), max-3))
	for _, ln := range s.tail {
		b.WriteString("\n")
		b.WriteString(truncateLine("    "+ln, max))
	}
	s.area.Update(b.String())
}

func labelWithTimer(label string, start time.Time) string {
	if start.IsZero() {
		return label
	}
	return fmt.Sprintf("%s (%ds)", label, int(time.Since(start).Seconds()))
}

// startAnimation advances the spinner on a ticker; it exits once the area is
// gone (step ended).
func (s *textSink) startAnimation() {
	go func() {
		ticker := time.NewTicker(terminal.SpinnerDelay)
		defer ticker.Stop()
		for range ticker.C {
			s.mu.Lock()
			if s.area == nil {
				s.mu.Unlock()
				return
			}
			s.frame++
			s.renderProgress()
			s.mu.Unlock()
		}
	}()
}

// truncateLine shortens s to at most max runes, marking the cut with an ellipsis.
func truncateLine(s string, max int) string {
	if max <= 0 {
		return s
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return string(r[:max-1]) + "…"
}

func (s *textSink) Close() error {
	if s.sp != nil {
		_ = s.sp.Stop()
		s.sp = nil
	}
	return nil
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
