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
	"net/http"
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
	suiteName          = "load-ttj"
	workspaceType      = "Workspace"
	trainingType       = "Training"
	runLabel           = "flowmark/run"
	progressInterval   = 30 * time.Second
	cleanupGrace       = time.Minute
	timestampTolerance = time.Second
)

type Suite struct{}

func (Suite) Name() string {
	return suiteName
}

func (Suite) Category() core.Category {
	return core.CategoryLoad
}

func (Suite) Run(ctx context.Context, target core.Target, scenario core.Scenario) (section core.Section, err error) {
	opts, err := parseOptions(scenario)
	if err != nil {
		return core.Section{}, err
	}

	setupCtx, cancelSetup := context.WithTimeout(ctx, opts.setupTimeout)
	defer cancelSetup()
	clients, err := client.NewWithLimits(ctx, target, opts.limits)
	if err != nil {
		return core.Section{}, err
	}
	if !clients.TokenExpiry.IsZero() {
		log.Printf("%s: control plane token expires in %s; longer runs will fail with 401", suiteName, time.Until(clients.TokenExpiry).Round(time.Minute))
	}
	clusterID, err := submit.ClusterID(setupCtx, clients, opts.cluster)
	if err != nil {
		return core.Section{}, err
	}
	runID, err := core.NewRunID()
	if err != nil {
		return core.Section{}, err
	}

	projectName := "flowmark-load-" + runID
	defer func() {
		log.Printf("%s: cleaning up project %q", suiteName, projectName)
		cleanupErr := cleanupProject(clients, clusterID, projectName, opts.setupTimeout)
		if cleanupErr == nil {
			return
		}
		if err != nil {
			err = errors.Join(err, cleanupErr)
			return
		}
		section.Failed = true
		section.Error = joinMessages(section.Error, "cleanup: "+cleanupErr.Error())
	}()
	project, err := submit.Project(setupCtx, clients, clusterID, projectName, opts.setupTimeout)
	if err != nil {
		return core.Section{}, err
	}
	log.Printf("%s: project %q ready, namespace %s", suiteName, project.Name, project.Namespace)

	watchers, err := measure.OpenArrivals(ctx, clients.Kube, project.Namespace)
	if err != nil {
		return core.Section{}, err
	}
	defer watchers.Stop()

	log.Printf("%s: submitting %s workloads at %d/s for %s with %d users", suiteName, opts.workloadType, opts.rate, opts.duration, opts.users)
	submissions, pacing := submitWorkloads(ctx, clients, project, watchers, opts, runID)
	log.Printf("%s: submitted %d workloads at %.2f/s (target %d/s), %d rejected; waiting up to %s for arrivals", suiteName, pacing.Submitted, pacing.Rate(), opts.rate, rejected(submissions), opts.timeout)

	filled := watchers.Fill(submissions, opts.timeout)
	interrupted := ctx.Err()
	spans := measure.CollectSpans(filled)
	log.Printf("%s: %s", suiteName, describe(spans, pacing))
	log.Printf("%s: measurement %s", suiteName, describeMeasurement(watchers.Reconnects(), measure.OutOfOrder(filled, timestampTolerance)))

	deleteErr := deleteWorkloads(ctx, clients, project, opts)
	section = buildSection(spans, pacing, filled, opts, watchers.Reconnects(), interrupted)
	if deleteErr != nil {
		section.Failed = true
		section.Error = joinMessages(section.Error, deleteErr.Error())
	}
	return section, nil
}

func submitWorkloads(ctx context.Context, clients *client.Clients, project submit.ProjectRef, watchers *measure.Arrivals, opts options, runID string) ([]measure.Submission, submit.Pacing) {
	create := submit.Workspace
	if opts.workloadType == trainingType {
		create = submit.Training
	}
	var mu sync.Mutex
	var submissions []measure.Submission
	var accepted, failed atomic.Int64

	progressCtx, stopProgress := context.WithCancel(ctx)
	defer stopProgress()
	go reportProgress(progressCtx, &accepted, &failed, opts)

	pacer := submit.Pacer{Rate: opts.rate, Duration: opts.duration, Users: opts.users}
	pacing := pacer.Run(ctx, func(ctx context.Context, index int) {
		name := fmt.Sprintf("flowmark-load-%s-%d", runID, index+1)
		createCtx, cancel := context.WithTimeout(ctx, opts.setupTimeout)
		defer cancel()

		record := measure.Submission{Name: name, RequestedAt: time.Now()}
		_, err := create(createCtx, clients, submit.WorkloadSpec{Name: name, Project: project, Image: opts.image, Labels: map[string]string{runLabel: runID}})
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

func describeMeasurement(interruptions int, outOfOrder []string) string {
	if interruptions == 0 && len(outOfOrder) == 0 {
		return "clean: the cluster watch stayed open and every workload's events were seen in order"
	}
	var parts []string
	if interruptions > 0 {
		parts = append(parts, fmt.Sprintf("interrupted %d times", interruptions))
	}
	if len(outOfOrder) > 0 {
		parts = append(parts, fmt.Sprintf("timestamps out of order for %d workloads (first: %s)", len(outOfOrder), outOfOrder[0]))
	}
	return strings.Join(parts, ", ")
}

func deleteWorkloads(ctx context.Context, clients *client.Clients, project submit.ProjectRef, opts options) error {
	deleteCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), opts.setupTimeout)
	defer cancel()
	ids, err := submit.ProjectWorkloadIDs(deleteCtx, clients, project.ID, opts.workloadType)
	if err != nil {
		return err
	}
	var errs []error
	for _, id := range ids {
		var response *http.Response
		if opts.workloadType == trainingType {
			_, response, err = clients.Workloads.TrainingsAPI.DeleteTraining(deleteCtx, id).Execute()
		} else {
			_, response, err = clients.Workloads.WorkspacesAPI.DeleteWorkspace(deleteCtx, id).Execute()
		}
		if err != nil && (response == nil || response.StatusCode != http.StatusNotFound) {
			errs = append(errs, fmt.Errorf("deleting %s %s: %w", strings.ToLower(opts.workloadType), id, err))
		}
	}
	return errors.Join(errs...)
}

func cleanupProject(clients *client.Clients, clusterID, name string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout+cleanupGrace)
	defer cancel()
	return submit.CleanupProject(ctx, clients, clusterID, name, timeout)
}

func joinMessages(first, second string) string {
	if first == "" {
		return second
	}
	return first + "; " + second
}

const (
	defaultImage         = "registry.k8s.io/pause:3.9"
	maxUsers             = 64
	maxRatePerSecond     = 1000
	maxRequestsPerSecond = 1000
)

var knownKnobs = []string{"users", "rate", "duration", "workloadType", "image", "timeout", "setupTimeout", "kubeRequestsPerSecond", "controlPlaneRequestsPerSecond", "cluster", "slo"}

type options struct {
	users        int
	rate         int
	duration     time.Duration
	workloadType string
	image        string
	timeout      time.Duration
	setupTimeout time.Duration
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
	if opts.rate, err = scenario.Int("rate", 2); err != nil {
		return options{}, err
	}
	if opts.rate < 1 || opts.rate > maxRatePerSecond {
		return options{}, fmt.Errorf("knob \"rate\" must be between 1 and %d submissions per second, got %d", maxRatePerSecond, opts.rate)
	}
	if opts.duration, err = scenario.Duration("duration", 5*time.Minute); err != nil {
		return options{}, err
	}
	if opts.workloadType, err = scenario.String("workloadType", workspaceType); err != nil {
		return options{}, err
	}
	if opts.workloadType != workspaceType && opts.workloadType != trainingType {
		return options{}, fmt.Errorf("knob \"workloadType\" must be %s or %s, got %q", workspaceType, trainingType, opts.workloadType)
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
	slo, err := scenario.Map("slo")
	if err != nil {
		return options{}, err
	}
	if opts.slo, err = wholeNumbers("slo", slo); err != nil {
		return options{}, err
	}
	return opts, nil
}

func wholeNumbers(knob string, raw map[string]any) (map[string]int, error) {
	if raw == nil {
		return nil, nil
	}
	values := make(map[string]int, len(raw))
	for key, value := range raw {
		number, ok := value.(int)
		if !ok {
			return nil, fmt.Errorf("knob %q entry %q must be a whole number of milliseconds, got %T", knob, key, value)
		}
		values[key] = number
	}
	return values, nil
}

func buildSection(s measure.Spans, pacing submit.Pacing, submissions []measure.Submission, opts options, interruptions int, interrupted error) core.Section {
	ttj := measure.Stats(s.TTJ)
	api := measure.Stats(s.ApiAccepted)
	metrics := map[string]any{
		"flow":                           strings.ToLower(opts.workloadType) + "_to_pod",
		"workload_type":                  opts.workloadType,
		"users":                          opts.users,
		"target_rate_per_sec":            opts.rate,
		"actual_rate_per_sec":            pacing.Rate(),
		"duration_ms":                    opts.duration.Milliseconds(),
		"image":                          opts.image,
		"kube_requests_per_sec":          opts.limits.KubeRequestsPerSecond,
		"control_plane_requests_per_sec": opts.limits.ControlPlaneRequestsPerSecond,
		"submitted":                      pacing.Submitted,
		"failed_count":                   api.FailedCount,
		"never_arrived_count":            ttj.NeverArrivedCount,
		"success_rate":                   measure.SuccessRate(submissions),
	}
	for name, timings := range map[string][]measure.Timing{"api_accepted": s.ApiAccepted, "cluster_synced": s.ClusterSynced, "pod_created": s.PodCreated, "ttj": s.TTJ} {
		summary := measure.Stats(timings)
		if summary.SampleSize == 0 {
			continue
		}
		metrics[name+"_p50_ms"] = summary.P50.Milliseconds()
		metrics[name+"_p90_ms"] = summary.P90.Milliseconds()
		metrics[name+"_p99_ms"] = summary.P99.Milliseconds()
		metrics[name+"_mean_ms"] = summary.Mean.Milliseconds()
		metrics[name+"_max_ms"] = summary.Max.Milliseconds()
	}
	if opts.slo != nil {
		metrics["slo"] = opts.slo
		for key, value := range opts.slo {
			metrics["slo_"+key] = value
		}
	}

	var problems []string
	if ttj.NeverArrivedCount > 0 {
		if interrupted != nil {
			problems = append(problems, fmt.Sprintf("%d workloads had not reached the cluster when the run was interrupted: %v", ttj.NeverArrivedCount, interrupted))
		} else {
			problems = append(problems, fmt.Sprintf("%d workloads never reached the cluster within %s", ttj.NeverArrivedCount, opts.timeout))
		}
	}
	if outOfOrder := measure.OutOfOrder(submissions, timestampTolerance); len(outOfOrder) > 0 {
		problems = append(problems, fmt.Sprintf("timestamps out of order for %d workloads (first: %s): a WorkloadRunner or pod was seen more than %s before the step that precedes it", len(outOfOrder), outOfOrder[0], timestampTolerance))
	}
	if interruptions > 0 {
		problems = append(problems, fmt.Sprintf("the measurement was interrupted %d times: the cluster watch closed and was reopened; latencies of workloads in flight during the gaps are inflated", interruptions))
	}

	section := core.Section{Category: core.CategoryLoad, Metrics: metrics}
	if len(problems) > 0 {
		section.Failed = true
		section.Error = strings.Join(problems, "; ")
	}
	return section
}
