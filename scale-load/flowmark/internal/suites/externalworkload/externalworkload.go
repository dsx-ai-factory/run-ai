/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package externalworkload

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/client"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/measure"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/submit"
)

const (
	deploymentType     = "Deployment"
	pauseImage         = "registry.k8s.io/pause:3.9"
	submitWorkers      = 16
	maxPodsPerWorkload = 10
	quiescePoll        = 2 * time.Second
	cleanupGrace       = time.Minute
	runLabel           = "flowmark/run"
	roundLabel         = "flowmark/round"
	sloP90Key          = "create_to_ew_p90_ms"
)

type Suite struct{}

type settings struct {
	rounds       []int
	workloadType string
	pods         int32
	timeout      time.Duration
	setupTimeout time.Duration
	settle       time.Duration
	cluster      string
	slo          map[string]int
}

type round struct {
	index          int
	size           int
	timings        []measure.Timing
	submitMakespan time.Duration
	quiesce        time.Duration
	err            error
}

func (Suite) Name() string {
	return "external-workload"
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

	projectName := "flowmark-ew-" + id
	defer func() {
		cleanupErr := cleanup(clients, clusterID, projectName, knobs.setupTimeout)
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
	project, err := submit.Project(ctx, clients, clusterID, projectName, knobs.setupTimeout)
	if err != nil {
		return core.Section{}, err
	}

	var rounds []round
	for index, size := range knobs.rounds {
		result := runRound(ctx, clients.Kube, project.Namespace, id, index+1, size, knobs)
		rounds = append(rounds, result)
		if result.err != nil || ctx.Err() != nil {
			break
		}
	}
	return buildSection(rounds, knobs), nil
}

func parseSettings(cfg core.Config) (settings, error) {
	var knobs settings
	var err error
	if knobs.rounds, err = cfg.Ints("workloads", []int{100, 250, 500}); err != nil {
		return settings{}, err
	}
	if knobs.workloadType, err = cfg.String("workloadType", deploymentType); err != nil {
		return settings{}, err
	}
	pods, err := cfg.Int("podsPerWorkload", 1)
	if err != nil {
		return settings{}, err
	}
	if knobs.timeout, err = cfg.Duration("timeout", 15*time.Minute); err != nil {
		return settings{}, err
	}
	if knobs.setupTimeout, err = cfg.Duration("setupTimeout", 5*time.Minute); err != nil {
		return settings{}, err
	}
	if knobs.settle, err = cfg.Duration("settle", 90*time.Second); err != nil {
		return settings{}, err
	}
	if knobs.cluster, err = cfg.String("cluster", ""); err != nil {
		return settings{}, err
	}
	slo, err := cfg.Map("slo")
	if err != nil {
		return settings{}, err
	}
	if knobs.slo, err = wholeNumbers("slo", slo); err != nil {
		return settings{}, err
	}

	if len(knobs.rounds) == 0 {
		return settings{}, errors.New("workloads must list at least one round size")
	}
	seen := map[int]bool{}
	for _, size := range knobs.rounds {
		if size < 1 {
			return settings{}, fmt.Errorf("every round size in workloads must be at least 1, got %d", size)
		}
		if seen[size] {
			return settings{}, fmt.Errorf("workloads must not repeat a round size, got %d twice", size)
		}
		seen[size] = true
	}
	if pods < 0 || pods > maxPodsPerWorkload {
		return settings{}, fmt.Errorf("podsPerWorkload must be between 0 and %d, got %d", maxPodsPerWorkload, pods)
	}
	knobs.pods = int32(pods)
	if knobs.workloadType != deploymentType {
		return settings{}, fmt.Errorf("workloadType %q is not supported yet, only %s", knobs.workloadType, deploymentType)
	}
	return knobs, nil
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

func runRound(ctx context.Context, kube k8sclient.WithWatch, namespace, id string, index, size int, knobs settings) round {
	result := round{index: index, size: size}
	watcher, err := measure.WatchExternalWorkloads(ctx, kube, namespace)
	if err != nil {
		result.err = err
		return result
	}
	defer watcher.Stop()

	labels := map[string]string{runLabel: id, roundLabel: strconv.Itoa(index)}
	var names sync.Map
	submitStart := time.Now()
	burst(ctx, kube, watcher, &names, namespace, id, index, size, labels, knobs.pods)
	result.submitMakespan = time.Since(submitStart)

	timings := watcher.WaitAll(knobs.timeout)
	for i := range timings {
		if name, ok := names.Load(timings[i].Name); ok {
			timings[i].Name = name.(string)
		}
	}
	sort.Slice(timings, func(i, j int) bool { return timings[i].Name < timings[j].Name })
	result.timings = timings

	quiesceStart := time.Now()
	result.err = quiesce(ctx, kube, namespace, labels, knobs.setupTimeout)
	result.quiesce = time.Since(quiesceStart)
	if result.err == nil {
		settle(ctx, knobs.settle)
	}
	return result
}

func settle(ctx context.Context, pause time.Duration) {
	timer := time.NewTimer(pause)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

func burst(ctx context.Context, kube k8sclient.Client, watcher *measure.Watcher, names *sync.Map, namespace, id string, index, size int, labels map[string]string, pods int32) {
	jobs := make(chan int)
	var wg sync.WaitGroup
	for range submitWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				name := fmt.Sprintf("flowmark-ew-%s-r%d-%d", id, index, i+1)
				deployment, err := submit.K8sDeployment(ctx, kube, submit.DeploymentSpec{
					Name: name, Namespace: namespace, Image: pauseImage, Replicas: pods, Labels: labels,
				})
				acceptedAt := time.Now()
				if err != nil {
					watcher.MarkFailed(name, err)
					continue
				}
				uid := string(deployment.UID)
				names.Store(uid, name)
				watcher.MarkCreated(uid, acceptedAt)
			}
		}()
	}
	for i := range size {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
}

func quiesce(ctx context.Context, kube k8sclient.Client, namespace string, labels map[string]string, timeout time.Duration) error {
	listKind := schema.GroupVersionKind{Group: measure.ExternalWorkloadKind.Group, Version: measure.ExternalWorkloadKind.Version, Kind: measure.ExternalWorkloadKind.Kind + "List"}
	var last string
	err := wait.PollUntilContextTimeout(ctx, quiescePoll, timeout, true, func(ctx context.Context) (bool, error) {
		remaining, err := deleteDeployments(ctx, kube, namespace, labels)
		if err != nil {
			last = err.Error()
			return false, nil
		}
		external := &unstructured.UnstructuredList{}
		external.SetGroupVersionKind(listKind)
		if err := kube.List(ctx, external, k8sclient.InNamespace(namespace)); err != nil {
			last = err.Error()
			return false, nil
		}
		var pods corev1.PodList
		if err := kube.List(ctx, &pods, k8sclient.InNamespace(namespace), k8sclient.MatchingLabels(labels)); err != nil {
			last = err.Error()
			return false, nil
		}
		last = fmt.Sprintf("%d deployments, %d external workloads and %d pods still present", remaining, len(external.Items), len(pods.Items))
		return remaining == 0 && len(external.Items) == 0 && len(pods.Items) == 0, nil
	})
	if err != nil {
		return fmt.Errorf("namespace %s did not quiesce within %s (last: %s): %w", namespace, timeout, last, err)
	}
	return nil
}

func deleteDeployments(ctx context.Context, kube k8sclient.Client, namespace string, labels map[string]string) (int, error) {
	var deployments appsv1.DeploymentList
	if err := kube.List(ctx, &deployments, k8sclient.InNamespace(namespace), k8sclient.MatchingLabels(labels)); err != nil {
		return 0, fmt.Errorf("listing the round's deployments: %w", err)
	}
	jobs := make(chan *appsv1.Deployment)
	errs := make(chan error, len(deployments.Items))
	var wg sync.WaitGroup
	for range submitWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for deployment := range jobs {
				if err := k8sclient.IgnoreNotFound(kube.Delete(ctx, deployment)); err != nil {
					errs <- fmt.Errorf("deleting deployment %s: %w", deployment.Name, err)
				}
			}
		}()
	}
	for i := range deployments.Items {
		if deployments.Items[i].DeletionTimestamp == nil {
			jobs <- &deployments.Items[i]
		}
	}
	close(jobs)
	wg.Wait()
	close(errs)
	for err := range errs {
		return len(deployments.Items), err
	}
	return len(deployments.Items), nil
}

func buildSection(rounds []round, knobs settings) core.Section {
	metrics := map[string]any{
		"rounds":            knobs.rounds,
		"workload_type":     knobs.workloadType,
		"pods_per_workload": knobs.pods,
		"settle_ms":         knobs.settle.Milliseconds(),
	}
	if knobs.slo != nil {
		metrics["slo"] = knobs.slo
		if value, ok := knobs.slo[sloP90Key]; ok {
			metrics["create_to_ew_slo_ms"] = value
		}
	}

	var problems []string
	for _, r := range rounds {
		at := fmt.Sprintf("@%d", r.size)
		summary := measure.Stats(r.timings)
		metrics["sample_size"+at] = summary.SampleSize
		metrics["pending_count"+at] = summary.PendingCount
		metrics["failed_count"+at] = summary.FailedCount
		metrics["success_rate"+at] = summary.SuccessRate
		metrics["submit_makespan_ms"+at] = r.submitMakespan.Milliseconds()
		metrics["quiesce_ms"+at] = r.quiesce.Milliseconds()
		if summary.SampleSize > 0 {
			metrics["create_to_ew_p50_ms"+at] = summary.P50.Milliseconds()
			metrics["create_to_ew_p90_ms"+at] = summary.P90.Milliseconds()
			metrics["create_to_ew_p99_ms"+at] = summary.P99.Milliseconds()
			metrics["create_to_ew_max_ms"+at] = summary.Max.Milliseconds()
			metrics["ew_makespan_ms"+at] = summary.Makespan.Milliseconds()
			metrics["throughput_per_sec"+at] = summary.Throughput
		}
		problems = append(problems, roundProblems(r, summary)...)
	}
	if len(rounds) < len(knobs.rounds) {
		problems = append(problems, fmt.Sprintf("run stopped after %d of %d rounds", len(rounds), len(knobs.rounds)))
	}

	section := core.Section{Category: core.CategoryScale, Metrics: metrics}
	if len(problems) > 0 {
		section.Failed = true
		section.Error = strings.Join(problems, "; ")
	}
	return section
}

func roundProblems(r round, summary measure.Summary) []string {
	var problems []string
	if summary.PendingCount > 0 {
		message := fmt.Sprintf("round %d (%d deployments): %d never got an ExternalWorkload within the timeout", r.index, r.size, summary.PendingCount)
		if err := lastError(r.timings, measure.Pending); err != nil {
			message += " (" + err.Error() + ")"
		}
		problems = append(problems, message)
	}
	if summary.FailedCount > 0 {
		message := fmt.Sprintf("round %d (%d deployments): %d creates failed", r.index, r.size, summary.FailedCount)
		if err := lastError(r.timings, measure.Failed); err != nil {
			message += " (" + err.Error() + ")"
		}
		problems = append(problems, message)
	}
	if r.err != nil {
		problems = append(problems, fmt.Sprintf("round %d (%d deployments): %v", r.index, r.size, r.err))
	}
	return problems
}

func lastError(timings []measure.Timing, outcome measure.Outcome) error {
	var last error
	for _, timing := range timings {
		if timing.Outcome == outcome && timing.Err != nil {
			last = timing.Err
		}
	}
	return last
}

func cleanup(clients *client.Clients, clusterID, name string, timeout time.Duration) error {
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
