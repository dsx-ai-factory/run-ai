/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package load

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/client"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/measure"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/submit"
)

const (
	suiteName           = "load-ttj"
	projectPrefix       = "flowmark-load"
	runLabel            = "flowmark/run"
	progressInterval    = 30 * time.Second
	timestampTolerance  = time.Second
	minSubmittedDefault = 90
)

type sessionOpener func(ctx context.Context, target core.Target, limits client.Limits, cluster, projectPrefix string, setupTimeout time.Duration) (*submit.Session, error)

type Suite struct {
	openSession sessionOpener
}

func (s Suite) session() sessionOpener {
	if s.openSession != nil {
		return s.openSession
	}
	return submit.OpenSession
}

func (Suite) Name() string {
	return suiteName
}

func (Suite) Category() core.Category {
	return core.CategoryLoad
}

func (Suite) Knobs() []string {
	return knownKnobs
}

func (Suite) Validate(scenario core.Scenario) error {
	_, err := parseOptions(scenario)
	return err
}

func (s Suite) Run(ctx context.Context, target core.Target, scenario core.Scenario) (section core.Section, err error) {
	opts, err := parseOptions(scenario)
	if err != nil {
		return core.Section{}, err
	}

	session, err := s.session()(ctx, target, opts.limits, opts.cluster, projectPrefix, opts.setupTimeout)
	if err != nil {
		return core.Section{}, err
	}
	clients, project, runID := session.Clients, session.Project, session.RunID
	defer func() {
		log.Printf("%s: cleaning up project %q", suiteName, project.Name)
		if cleanupErr := session.Close(); cleanupErr != nil {
			if err != nil {
				err = errors.Join(err, cleanupErr)
			} else {
				section.Fail("cleanup: " + cleanupErr.Error())
			}
		}
	}()
	log.Printf("%s: project %q ready, namespace %s", suiteName, project.Name, project.Namespace)

	watchers, err := measure.OpenArrivals(ctx, clients.Kube, project.Namespace)
	if err != nil {
		return core.Section{}, err
	}
	defer watchers.Stop()

	log.Printf("%s: submitting %s workloads at %d/s for %s with %d users%s", suiteName, opts.workloadType, opts.rate, opts.duration, opts.users, describeRamp(opts))
	submissions, pacing := submitWorkloads(ctx, clients, project, watchers, opts, runID)
	log.Printf("%s: submitted %d workloads at %.2f/s (target %d/s), %d rejected; waiting up to %s for arrivals", suiteName, pacing.Submitted, pacing.Rate(), opts.rate, rejected(submissions), opts.timeout)

	filled := watchers.Fill(submissions, opts.timeout)
	interrupted := ctx.Err()
	watchErr := watchers.Err()
	spans := measure.CollectSpans(filled)
	log.Printf("%s: %s", suiteName, describe(spans, pacing))
	log.Printf("%s: measurement %s", suiteName, describeMeasurement(watchers.Reconnects(), watchers.Interruptions(), measure.OutOfOrder(filled, timestampTolerance)))

	deleteErr := deleteWorkloads(ctx, clients, project, opts)
	section = buildSection(spans, pacing, filled, opts, watchers.Interruptions(), interrupted, watchErr)
	if deleteErr != nil {
		section.Fail(deleteErr.Error())
	}
	return section, nil
}

func submitWorkloads(ctx context.Context, clients *client.Clients, project submit.ProjectRef, watchers *measure.Arrivals, opts options, runID string) ([]measure.Submission, submit.Pacing) {
	create := submit.Workspace
	if opts.workloadType == submit.TrainingType {
		create = submit.Training
	}
	var mu sync.Mutex
	var submissions []measure.Submission
	var accepted, failed atomic.Int64

	progressCtx, stopProgress := context.WithCancel(ctx)
	defer stopProgress()
	go reportProgress(progressCtx, &accepted, &failed, opts)

	pacer := submit.Pacer{Rate: opts.rate, SpawnRate: opts.spawnRate, Duration: opts.duration, Users: opts.users}
	pacing := pacer.Run(ctx, func(ctx context.Context, index int) {
		name := fmt.Sprintf("flowmark-load-%s-%d", runID, index+1)
		createCtx, cancel := context.WithTimeout(ctx, opts.setupTimeout)
		defer cancel()

		record := measure.Submission{Name: name, RequestedAt: time.Now()}
		_, err := create(createCtx, clients, submit.WorkloadSpec{Name: name, Project: project, Image: opts.image, Labels: map[string]string{runLabel: runID}})
		if err != nil && ctx.Err() != nil {
			return
		}
		if err != nil {
			record.Err = err
			failed.Add(1)
		} else {
			record.ApiAcceptedAt = time.Now()
			watchers.MarkAccepted(name, record.ApiAcceptedAt)
			accepted.Add(1)
		}
		mu.Lock()
		submissions = append(submissions, record)
		mu.Unlock()
	})
	return submissions, pacing
}

func reportProgress(ctx context.Context, accepted, failed *atomic.Int64, opts options) {
	ticker := time.NewTicker(progressInterval)
	defer ticker.Stop()
	start := time.Now()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			elapsed := time.Since(start)
			total := accepted.Load() + failed.Load()
			log.Printf("%s: %s in: %d submitted (%.2f/s, target %d/s), %d rejected", suiteName, elapsed.Round(time.Second), total, float64(total)/elapsed.Seconds(), opts.rate, failed.Load())
		}
	}
}

func describeRamp(opts options) string {
	if opts.spawnRate == 0 {
		return ""
	}
	return fmt.Sprintf(", ramping up %d users per second", opts.spawnRate)
}

func rejected(submissions []measure.Submission) int {
	count := 0
	for _, s := range submissions {
		if !s.Accepted() {
			count++
		}
	}
	return count
}

func describe(s measure.Spans, pacing submit.Pacing) string {
	ttj := measure.Stats(s.TTJ)
	text := fmt.Sprintf("%d/%d reached the cluster", ttj.SampleSize, pacing.Submitted)
	if ttj.SampleSize > 0 {
		text += fmt.Sprintf(", ttj p90 %s, api p90 %s", ttj.P90.Round(time.Millisecond), measure.Stats(s.ApiAccepted).P90.Round(time.Millisecond))
	}
	if ttj.NeverArrivedCount > 0 {
		text += fmt.Sprintf(", %d never arrived", ttj.NeverArrivedCount)
	}
	return text
}

func describeMeasurement(reopened, interruptions int, outOfOrder []string) string {
	if reopened == 0 && len(outOfOrder) == 0 {
		return "clean: the cluster watch stayed open and every workload's events were seen in order"
	}
	var parts []string
	switch {
	case interruptions > 0:
		parts = append(parts, fmt.Sprintf("interrupted %d times", interruptions))
	case reopened > 0:
		parts = append(parts, fmt.Sprintf("the cluster watch was reopened %d times and resumed without losing events", reopened))
	}
	if len(outOfOrder) > 0 {
		parts = append(parts, fmt.Sprintf("timestamps out of order for %d workloads (first: %s)", len(outOfOrder), outOfOrder[0]))
	}
	return strings.Join(parts, ", ")
}

func deleteWorkloads(ctx context.Context, clients *client.Clients, project submit.ProjectRef, opts options) error {
	listCtx, cancelList := context.WithTimeout(context.WithoutCancel(ctx), opts.setupTimeout)
	defer cancelList()
	ids, err := submit.ProjectWorkloadIDs(listCtx, clients, project.ID, opts.workloadType)
	if err != nil {
		return err
	}
	deleteCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), deleteBudget(len(ids), opts))
	defer cancel()
	_, err = submit.DeleteWorkloads(deleteCtx, clients, opts.workloadType, ids, opts.users)
	return err
}

func deleteBudget(count int, opts options) time.Duration {
	perRequest := time.Second / time.Duration(opts.limits.ControlPlaneRequestsPerSecond)
	return opts.setupTimeout + time.Duration(count)*perRequest
}

const (
	defaultImage         = "registry.k8s.io/pause:3.9"
	maxUsers             = 512
	maxRatePerSecond     = 1000
	maxRequestsPerSecond = 1000
)

var (
	spanNames    = []string{"api_accepted", "cluster_synced", "pod_created", "ttj"}
	latencyStats = []string{"p50_ms", "p90_ms", "p99_ms"}
	sloKeys      = append(latencySLOKeys(), "failed_count")
)

func latencySLOKeys() []string {
	keys := make([]string, 0, len(spanNames)*len(latencyStats))
	for _, span := range spanNames {
		for _, stat := range latencyStats {
			keys = append(keys, span+"_"+stat)
		}
	}
	return keys
}

var knownKnobs = []string{"users", "spawnRate", "rate", "duration", "workloadType", "image", "timeout", "setupTimeout", "minSubmittedPercent", "kubeRequestsPerSecond", "controlPlaneRequestsPerSecond", "cluster", "slo"}

type options struct {
	users        int
	spawnRate    int
	rate         int
	duration     time.Duration
	workloadType string
	image        string
	timeout      time.Duration
	setupTimeout time.Duration
	minSubmitted int
	limits       client.Limits
	cluster      string
	slo          map[string]int
}

func parseOptions(scenario core.Scenario) (options, error) {
	if unknown := scenario.UnknownKnobs(knownKnobs...); len(unknown) > 0 {
		return options{}, fmt.Errorf("unknown knob %q; known knobs: %s", unknown[0], strings.Join(knownKnobs, ", "))
	}
	var opts options
	var err error
	if opts.users, err = scenario.Int("users", 10); err != nil {
		return options{}, err
	}
	if opts.users < 1 || opts.users > maxUsers {
		return options{}, fmt.Errorf("knob \"users\" must be between 1 and %d, got %d", maxUsers, opts.users)
	}
	if opts.spawnRate, err = scenario.Int("spawnRate", 0); err != nil {
		return options{}, err
	}
	if opts.spawnRate < 0 || opts.spawnRate > maxUsers {
		return options{}, fmt.Errorf("knob \"spawnRate\" must be between 0 and %d users per second, got %d", maxUsers, opts.spawnRate)
	}
	if opts.rate, err = scenario.Int("rate", 2); err != nil {
		return options{}, err
	}
	if opts.rate < 1 || opts.rate > maxRatePerSecond {
		return options{}, fmt.Errorf("knob \"rate\" must be between 1 and %d submissions per second, got %d", maxRatePerSecond, opts.rate)
	}
	if opts.duration, err = scenario.Duration("duration", 3*time.Minute); err != nil {
		return options{}, err
	}
	if opts.workloadType, err = scenario.String("workloadType", submit.WorkspaceType); err != nil {
		return options{}, err
	}
	if opts.workloadType != submit.WorkspaceType && opts.workloadType != submit.TrainingType {
		return options{}, fmt.Errorf("knob \"workloadType\" must be %s or %s, got %q", submit.WorkspaceType, submit.TrainingType, opts.workloadType)
	}
	if opts.image, err = scenario.String("image", defaultImage); err != nil {
		return options{}, err
	}
	if opts.timeout, err = scenario.Duration("timeout", 10*time.Minute); err != nil {
		return options{}, err
	}
	if opts.setupTimeout, err = scenario.Duration("setupTimeout", 5*time.Minute); err != nil {
		return options{}, err
	}
	if opts.minSubmitted, err = scenario.Int("minSubmittedPercent", minSubmittedDefault); err != nil {
		return options{}, err
	}
	if opts.minSubmitted < 0 || opts.minSubmitted > 100 {
		return options{}, fmt.Errorf("knob \"minSubmittedPercent\" must be between 0 and 100, got %d", opts.minSubmitted)
	}
	if opts.limits.KubeRequestsPerSecond, err = scenario.Int("kubeRequestsPerSecond", client.DefaultLimits.KubeRequestsPerSecond); err != nil {
		return options{}, err
	}
	if opts.limits.ControlPlaneRequestsPerSecond, err = scenario.Int("controlPlaneRequestsPerSecond", client.DefaultLimits.ControlPlaneRequestsPerSecond); err != nil {
		return options{}, err
	}
	for name, value := range map[string]int{"kubeRequestsPerSecond": opts.limits.KubeRequestsPerSecond, "controlPlaneRequestsPerSecond": opts.limits.ControlPlaneRequestsPerSecond} {
		if value < 1 || value > maxRequestsPerSecond {
			return options{}, fmt.Errorf("knob %q must be between 1 and %d, got %d", name, maxRequestsPerSecond, value)
		}
	}
	if opts.rate > opts.limits.ControlPlaneRequestsPerSecond {
		return options{}, fmt.Errorf("knob \"rate\" (%d) cannot exceed controlPlaneRequestsPerSecond (%d); raise the limit or lower the rate", opts.rate, opts.limits.ControlPlaneRequestsPerSecond)
	}
	if opts.cluster, err = scenario.String("cluster", ""); err != nil {
		return options{}, err
	}
	if opts.slo, err = scenario.IntMap("slo"); err != nil {
		return options{}, err
	}
	for key := range opts.slo {
		if !slices.Contains(sloKeys, key) {
			return options{}, fmt.Errorf("slo key %q is not a metric this suite reports; use failed_count or <span>_<stat> with span %s and stat %s",
				key, strings.Join(spanNames, ", "), strings.Join(latencyStats, ", "))
		}
	}
	return opts, nil
}

func buildSection(s measure.Spans, pacing submit.Pacing, submissions []measure.Submission, opts options, interruptions int, interrupted, watchErr error) core.Section {
	ttj := measure.Stats(s.TTJ)
	api := measure.Stats(s.ApiAccepted)
	metrics := map[string]any{
		"workload_type":   opts.workloadType,
		"failed_count":    api.FailedCount,
		"planned_count":   pacing.Planned,
		"submitted_count": pacing.Submitted,
	}
	for name, timings := range map[string][]measure.Timing{"api_accepted": s.ApiAccepted, "cluster_synced": s.ClusterSynced, "pod_created": s.PodCreated, "ttj": s.TTJ} {
		summary := measure.Stats(timings)
		if summary.SampleSize == 0 {
			continue
		}
		metrics[name+"_p50_ms"] = summary.P50.Milliseconds()
		metrics[name+"_p90_ms"] = summary.P90.Milliseconds()
		metrics[name+"_p99_ms"] = summary.P99.Milliseconds()
	}

	var problems []string
	if required := int(math.Ceil(float64(pacing.Planned) * float64(opts.minSubmitted) / 100)); pacing.Planned > 0 && pacing.Submitted < required {
		problems = append(problems, fmt.Sprintf("the pacer kept only %d of the %d planned submissions (%.2f/s against the %d/s target): "+
			"the environment could not absorb the load", pacing.Submitted, pacing.Planned, pacing.Rate(), opts.rate))
	}
	switch {
	case interrupted != nil:
		message := fmt.Sprintf("the run was interrupted after %d submissions (%v)", pacing.Submitted, interrupted)
		if ttj.NeverArrivedCount > 0 {
			message += fmt.Sprintf("; %d workloads had not reached the cluster", ttj.NeverArrivedCount)
		}
		problems = append(problems, message)
	case watchErr != nil:
		problems = append(problems, fmt.Sprintf("the cluster watch died, %d workloads could not be followed: %v", ttj.NeverArrivedCount, watchErr))
	case ttj.NeverArrivedCount > 0:
		problems = append(problems, fmt.Sprintf("%d workloads never reached the cluster within %s", ttj.NeverArrivedCount, opts.timeout))
	}
	if outOfOrder := measure.OutOfOrder(submissions, timestampTolerance); len(outOfOrder) > 0 {
		problems = append(problems, fmt.Sprintf("timestamps out of order for %d workloads (first: %s): "+
			"a WorkloadRunner or pod was seen more than %s before the step that precedes it", len(outOfOrder), outOfOrder[0], timestampTolerance))
	}
	if interruptions > 0 {
		problems = append(problems, fmt.Sprintf("the measurement was interrupted %d times: the cluster watch was reopened without its resume point; "+
			"latencies of workloads in flight during the gaps are inflated", interruptions))
	}

	section := core.Section{Metrics: metrics, SLO: opts.slo}
	if len(problems) > 0 {
		section.Fail(strings.Join(problems, "; "))
	}
	return section
}
