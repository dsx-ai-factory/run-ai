/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package basicflow

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/client"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/submit"
)

const (
	pollInterval = time.Second
	cleanupGrace = time.Minute
	trainingType = "Training"
)

var workloadRunnerKind = schema.GroupVersionKind{Group: "run.ai", Version: "v1alpha1", Kind: "WorkloadRunner"}

type Suite struct{}

type settings struct {
	projects     int
	workloads    int
	workloadType string
	timeout      time.Duration
	cluster      string
}

type acceptedWorkload struct {
	ref        submit.WorkloadRef
	namespace  string
	acceptedAt time.Time
}

type arrivalCheck func(ctx context.Context, workload acceptedWorkload) (bool, error)

type measurement struct {
	syncMillis map[string]int64
	pending    []string
	lastErr    error
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

	id, err := runID()
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

	accepted := make([]acceptedWorkload, 0, knobs.workloads)
	for i := range knobs.workloads {
		project := projects[i%len(projects)]
		ref, err := submit.Training(ctx, clients, submit.WorkloadSpec{
			Name:    fmt.Sprintf("flowmark-basic-%s-w%d", id, i+1),
			Project: project,
			Image:   "ubuntu",
			Command: "sleep infinity",
		})
		acceptedAt := time.Now()
		if err != nil {
			return core.Section{}, err
		}
		accepted = append(accepted, acceptedWorkload{ref: ref, namespace: project.Namespace, acceptedAt: acceptedAt})
	}

	result := measure(ctx, accepted, runnerExists(clients.Kube), pollInterval, knobs.timeout)
	return result.section(knobs.workloads), nil
}

func runID() (string, error) {
	var raw [4]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate run id: %w", err)
	}
	return hex.EncodeToString(raw[:]), nil
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

func runnerExists(kube k8sclient.Client) arrivalCheck {
	return func(ctx context.Context, workload acceptedWorkload) (bool, error) {
		runner := &unstructured.Unstructured{}
		runner.SetGroupVersionKind(workloadRunnerKind)
		err := kube.Get(ctx, k8sclient.ObjectKey{Namespace: workload.namespace, Name: workload.ref.Name}, runner)
		switch {
		case err == nil:
			return true, nil
		case apierrors.IsNotFound(err):
			return false, nil
		}
		return false, err
	}
}

// TODO(Task 7): replace this polling loop with watch-arrival timing from the measure task.
func measure(ctx context.Context, workloads []acceptedWorkload, arrived arrivalCheck, interval, timeout time.Duration) measurement {
	result := measurement{syncMillis: map[string]int64{}}
	remaining := append([]acceptedWorkload(nil), workloads...)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for len(remaining) > 0 {
		still := remaining[:0]
		for _, workload := range remaining {
			ok, err := arrived(ctx, workload)
			if err != nil {
				result.lastErr = err
			}
			switch {
			case ok:
				result.syncMillis[workload.ref.Name] = time.Since(workload.acceptedAt).Milliseconds()
			case time.Since(workload.acceptedAt) >= timeout:
				result.pending = append(result.pending, workload.ref.Name)
			default:
				still = append(still, workload)
			}
		}
		remaining = still
		if len(remaining) == 0 {
			break
		}

		select {
		case <-ctx.Done():
			for _, workload := range remaining {
				result.pending = append(result.pending, workload.ref.Name)
			}
			result.lastErr = ctx.Err()
			return result
		case <-ticker.C:
		}
	}
	return result
}

func (m measurement) section(total int) core.Section {
	metrics := map[string]any{
		"workloads":          total,
		"sync_to_cluster_ms": m.syncMillis,
		"pending":            len(m.pending),
	}
	if len(m.syncMillis) > 0 {
		var maxMillis int64
		for _, millis := range m.syncMillis {
			maxMillis = max(maxMillis, millis)
		}
		metrics["sync_to_cluster_max_ms"] = maxMillis
	}

	section := core.Section{Category: core.CategoryScale, Metrics: metrics}
	if len(m.pending) > 0 {
		section.Failed = true
		section.Error = fmt.Sprintf("%d of %d workloads never reached the cluster within the timeout: %v", len(m.pending), total, m.pending)
		if m.lastErr != nil {
			section.Error = joinMessages(section.Error, "last cluster error: "+m.lastErr.Error())
		}
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
