/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package basicflow

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/client"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/measure"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/submit"
)

const (
	cleanupGrace = time.Minute
	trainingType = "Training"
)

type Suite struct{}

type settings struct {
	projects     int
	workloads    int
	workloadType string
	timeout      time.Duration
	cluster      string
}

func (Suite) Name() string {
	return "basic-flow"
}

func (Suite) Run(ctx context.Context, env core.Env, cfg core.Config) (section core.Section, err error) {
	knobs, err := parseSettings(cfg)
	if err != nil {
		return core.Section{}, err
	}

	clients, err := client.New(ctx, env)
	if err != nil {
		return core.Section{}, err
	}
	clusterID, err := submit.ClusterID(ctx, clients, knobs.cluster)
	if err != nil {
		return core.Section{}, err
	}

	id, err := core.RunID()
	if err != nil {
		return core.Section{}, err
	}
	var created []string
	defer func() {
		cleanupErr := cleanup(clients, clusterID, created, knobs.timeout)
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

	projects := make([]submit.ProjectRef, 0, knobs.projects)
	for i := range knobs.projects {
		name := fmt.Sprintf("flowmark-basic-%s-%d", id, i+1)
		created = append(created, name)
		project, err := submit.Project(ctx, clients, clusterID, name, knobs.timeout)
		if err != nil {
			return core.Section{}, err
		}
		projects = append(projects, project)
	}

	watchers := make([]*measure.Watcher, 0, len(projects))
	defer func() {
		for _, watcher := range watchers {
			watcher.Stop()
		}
	}()
	for _, project := range projects {
		watcher, err := measure.WatchWorkloads(ctx, clients.Kube, project.Namespace)
		if err != nil {
			return core.Section{}, err
		}
		watchers = append(watchers, watcher)
	}

	for i := range knobs.workloads {
		slot := i % len(projects)
		spec := submit.WorkloadSpec{
			Name:    fmt.Sprintf("flowmark-basic-%s-w%d", id, i+1),
			Project: projects[slot],
			Image:   "ubuntu",
			Command: "sleep infinity",
		}
		_, err := submit.Training(ctx, clients, spec)
		acceptedAt := time.Now()
		if err != nil {
			watchers[slot].MarkFailed(spec.Name, err)
			continue
		}
		watchers[slot].MarkCreated(spec.Name, acceptedAt)
	}

	var timings []measure.Timing
	for _, watcher := range watchers {
		timings = append(timings, watcher.WaitAll(knobs.timeout)...)
	}
	return buildSection(timings, knobs.workloads), nil
}

func parseSettings(cfg core.Config) (settings, error) {
	var knobs settings
	var err error
	if knobs.projects, err = cfg.Int("projects", 1); err != nil {
		return settings{}, err
	}
	if knobs.workloads, err = cfg.Int("workloads", 2); err != nil {
		return settings{}, err
	}
	if knobs.workloadType, err = cfg.String("workloadType", trainingType); err != nil {
		return settings{}, err
	}
	if knobs.timeout, err = cfg.Duration("timeout", 5*time.Minute); err != nil {
		return settings{}, err
	}
	if knobs.cluster, err = cfg.String("cluster", ""); err != nil {
		return settings{}, err
	}
	if knobs.projects < 1 || knobs.workloads < 1 {
		return settings{}, fmt.Errorf("projects and workloads must be at least 1, got %d and %d", knobs.projects, knobs.workloads)
	}
	if knobs.workloadType != trainingType {
		return settings{}, fmt.Errorf("workloadType %q is not supported yet, only %s", knobs.workloadType, trainingType)
	}
	return knobs, nil
}

func buildSection(timings []measure.Timing, total int) core.Section {
	metrics := measure.Stats(timings).Metrics()
	metrics["workloads"] = total

	sync := map[string]int64{}
	var pending, failed []string
	var lastErr error
	for _, timing := range timings {
		switch timing.Outcome {
		case measure.Arrived:
			sync[timing.Name] = timing.Duration.Milliseconds()
		case measure.Pending:
			pending = append(pending, timing.Name)
			if timing.Err != nil {
				lastErr = timing.Err
			}
		case measure.Failed:
			failed = append(failed, fmt.Sprintf("%s: %v", timing.Name, timing.Err))
		}
	}
	metrics["sync_to_cluster_ms"] = sync
	sort.Strings(pending)
	sort.Strings(failed)

	section := core.Section{Category: core.CategoryScale, Metrics: metrics}
	if len(pending) > 0 {
		section.Failed = true
		section.Error = fmt.Sprintf("%d of %d workloads never reached the cluster within the timeout: %v", len(pending), total, pending)
		if lastErr != nil {
			section.Error = joinMessages(section.Error, "watch error: "+lastErr.Error())
		}
	}
	if len(failed) > 0 {
		section.Failed = true
		section.Error = joinMessages(section.Error, fmt.Sprintf("%d of %d workload submissions failed: %s", len(failed), total, strings.Join(failed, "; ")))
	}
	return section
}

func cleanup(clients *client.Clients, clusterID string, names []string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout+cleanupGrace)
	defer cancel()

	var errs []error
	for _, name := range names {
		if err := submit.CleanupProject(ctx, clients, clusterID, name, timeout); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func joinMessages(first, second string) string {
	if first == "" {
		return second
	}
	return first + "; " + second
}
