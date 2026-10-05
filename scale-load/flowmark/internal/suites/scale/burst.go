/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package scale

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

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
	deploymentType       = "Deployment"
	trainingType         = "Training"
	defaultImage         = "registry.k8s.io/pause:3.9"
	defaultConcurrency   = 16
	maxConcurrency       = 64
	maxRequestsPerSecond = 1000
	maxPodsPerWorkload   = 10
	quiescePoll          = 2 * time.Second
	cleanupGrace         = time.Minute
	runLabel             = "flowmark/run"
	roundLabel           = "flowmark/round"
	roundSizeLabel       = "flowmark/round-size"
	sloP90Key            = "p90_ms"
)

var knownKnobs = []string{"workloads", "workloadType", "podsPerWorkload", "image", "submitConcurrency", "kubeRequestsPerSecond", "controlPlaneRequestsPerSecond", "timeout", "setupTimeout", "settle", "cluster", "slo"}

type Suite struct{}

type settings struct {
	rounds       []int
	workloadType string
	pods         int32
	image        string
	concurrency  int
	limits       client.Limits
	timeout      time.Duration
	setupTimeout time.Duration
	settle       time.Duration
	cluster      string
	slo          map[string]int
}

type item struct {
	name      string
	namespace string
	labels    map[string]string
	image     string
	pods      int32
}

type workloadFlow interface {
	name() string
	watch(ctx context.Context, namespace string) (*measure.Watcher, error)
	create(ctx context.Context, workload item) (string, error)
	teardown(ctx context.Context, namespace string, labels map[string]string) (int, error)
}

type round struct {
	index          int
	size           int
	timings        []measure.Timing
	unmatched      int
	interruptions  int
	submitMakespan time.Duration
	quiesce        time.Duration
	err            error
}

const suiteName = "scale-burst"

func (Suite) Name() string {
	return suiteName
}

func (Suite) Category() core.Category {
	return core.CategoryScale
}

func (Suite) Run(ctx context.Context, target core.Target, scenario core.Scenario) (section core.Section, err error) {
	knobs, err := parseSettings(scenario)
	if err != nil {
		return core.Section{}, err
	}

	setupCtx, cancelSetup := context.WithTimeout(ctx, knobs.setupTimeout)
	defer cancelSetup()
	clients, err := client.NewWithLimits(ctx, target, knobs.limits)
	if err != nil {
		return core.Section{}, err
	}
	if !clients.TokenExpiry.IsZero() {
		log.Printf("%s: control plane token expires in %s; longer runs will fail with 401", suiteName, time.Until(clients.TokenExpiry).Round(time.Minute))
	}
	clusterID, err := submit.ClusterID(setupCtx, clients, knobs.cluster)
	if err != nil {
		return core.Section{}, err
	}
	id, err := core.NewRunID()
	if err != nil {
		return core.Section{}, err
	}

	projectName := "flowmark-burst-" + id
	defer func() {
		log.Printf("%s: cleaning up project %q", suiteName, projectName)
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
	project, err := submit.Project(setupCtx, clients, clusterID, projectName, knobs.setupTimeout)
	if err != nil {
		return core.Section{}, err
	}
	log.Printf("%s: project %q ready, namespace %s", suiteName, project.Name, project.Namespace)

	flow := newFlow(knobs.workloadType, clients, project, knobs.concurrency)
	var rounds []round
	for index, size := range knobs.rounds {
		roundNumber := index + 1
		log.Printf("%s: round %d/%d started: %d %s workloads", suiteName, roundNumber, len(knobs.rounds), size, knobs.workloadType)
		roundStart := time.Now()
		result := runRound(ctx, flow, clients.Kube, project.Namespace, id, roundNumber, size, knobs)
		log.Printf("%s: round %d/%d finished in %s: %s", suiteName, roundNumber, len(knobs.rounds), time.Since(roundStart).Round(time.Second), describeRound(result))
		rounds = append(rounds, result)
		if result.err != nil || ctx.Err() != nil {
			break
		}
		if index < len(knobs.rounds)-1 {
			log.Printf("%s: settling %s before round %d", suiteName, knobs.settle, roundNumber+1)
			settle(ctx, knobs.settle)
			if ctx.Err() != nil {
				break
			}
		}
	}
	return buildSection(flow.name(), rounds, knobs), nil
}

func describeRound(r round) string {
	summary := measure.Stats(r.timings)
	text := fmt.Sprintf("%d/%d arrived", summary.SampleSize, r.size)
	if summary.SampleSize > 0 {
		text += fmt.Sprintf(", p90 %s", summary.P90.Round(time.Millisecond))
	}
	if summary.HasThroughput() {
		text += fmt.Sprintf(", %.1f workloads/s", summary.Throughput)
	}
	if summary.PendingCount > 0 {
		text += fmt.Sprintf(", %d pending", summary.PendingCount)
	}
	if summary.FailedCount > 0 {
		text += fmt.Sprintf(", %d failed", summary.FailedCount)
	}
	if r.unmatched > 0 {
		text += fmt.Sprintf(", %d unmatched arrivals", r.unmatched)
	}
	if r.interruptions > 0 {
		text += fmt.Sprintf(", measurement interrupted %d times", r.interruptions)
	}
	if r.err != nil {
		text += ": " + r.err.Error()
	}
	return text
}

func newFlow(workloadType string, clients *client.Clients, project submit.ProjectRef, concurrency int) workloadFlow {
	if workloadType == trainingType {
		return &trainingFlow{clients: clients, project: project, concurrency: concurrency, created: map[string]string{}, deleteRequested: map[string]bool{}}
	}
	return &deploymentFlow{kube: clients.Kube, concurrency: concurrency}
}

func parseSettings(scenario core.Scenario) (settings, error) {
	if unknown := scenario.UnknownKnobs(knownKnobs...); len(unknown) > 0 {
		return settings{}, fmt.Errorf("unknown knob %q; known knobs: %s", unknown[0], strings.Join(knownKnobs, ", "))
	}
	var knobs settings
	var err error
	if knobs.rounds, err = scenario.Ints("workloads", []int{100, 250, 500}); err != nil {
		return settings{}, err
	}
	if knobs.workloadType, err = scenario.String("workloadType", deploymentType); err != nil {
		return settings{}, err
	}
	pods, err := scenario.Int("podsPerWorkload", 1)
	if err != nil {
		return settings{}, err
	}
	if knobs.image, err = scenario.String("image", defaultImage); err != nil {
		return settings{}, err
	}
	if knobs.concurrency, err = scenario.Int("submitConcurrency", defaultConcurrency); err != nil {
		return settings{}, err
	}
	if knobs.limits.KubeRequestsPerSecond, err = scenario.Int("kubeRequestsPerSecond", client.DefaultLimits.KubeRequestsPerSecond); err != nil {
		return settings{}, err
	}
	if knobs.limits.ControlPlaneRequestsPerSecond, err = scenario.Int("controlPlaneRequestsPerSecond", client.DefaultLimits.ControlPlaneRequestsPerSecond); err != nil {
		return settings{}, err
	}
	if knobs.timeout, err = scenario.Duration("timeout", 15*time.Minute); err != nil {
		return settings{}, err
	}
	if knobs.setupTimeout, err = scenario.Duration("setupTimeout", 5*time.Minute); err != nil {
		return settings{}, err
	}
	if knobs.settle, err = scenario.Duration("settle", 90*time.Second); err != nil {
		return settings{}, err
	}
	if knobs.cluster, err = scenario.String("cluster", ""); err != nil {
		return settings{}, err
	}
	slo, err := scenario.Map("slo")
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
	if knobs.image == "" {
		return settings{}, errors.New("image must not be empty")
	}
	if knobs.concurrency < 1 || knobs.concurrency > maxConcurrency {
		return settings{}, fmt.Errorf("submitConcurrency must be between 1 and %d, got %d", maxConcurrency, knobs.concurrency)
	}
	for name, value := range map[string]int{"kubeRequestsPerSecond": knobs.limits.KubeRequestsPerSecond, "controlPlaneRequestsPerSecond": knobs.limits.ControlPlaneRequestsPerSecond} {
		if value < 1 || value > maxRequestsPerSecond {
			return settings{}, fmt.Errorf("%s must be between 1 and %d, got %d", name, maxRequestsPerSecond, value)
		}
	}
	switch knobs.workloadType {
	case deploymentType:
		if pods < 0 || pods > maxPodsPerWorkload {
			return settings{}, fmt.Errorf("podsPerWorkload must be between 0 and %d, got %d", maxPodsPerWorkload, pods)
		}
	case trainingType:
		if pods != 1 {
			return settings{}, fmt.Errorf("podsPerWorkload must be 1 for %s, got %d", trainingType, pods)
		}
	default:
		return settings{}, fmt.Errorf("workloadType %q is not supported, use %s or %s", knobs.workloadType, deploymentType, trainingType)
	}
	knobs.pods = int32(pods) //nolint:gosec
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

func runRound(ctx context.Context, flow workloadFlow, kube k8sclient.Client, namespace, id string, index, size int, knobs settings) round {
	result := round{index: index, size: size}
	watcher, err := flow.watch(ctx, namespace)
	if err != nil {
		result.err = err
		return result
	}
	defer watcher.Stop()

	labels := map[string]string{runLabel: id, roundLabel: strconv.Itoa(index), roundSizeLabel: strconv.Itoa(size)}
	var names sync.Map
	submitStart := time.Now()
	burst(ctx, flow, watcher, &names, namespace, id, index, size, labels, knobs)
	result.submitMakespan = time.Since(submitStart)

	timings := watcher.WaitAll(knobs.timeout)
	result.unmatched = watcher.Unmatched()
	result.interruptions = watcher.Reconnects()
	if ctx.Err() != nil {
		result.timings = timings
		result.err = fmt.Errorf("interrupted before teardown: %w", ctx.Err())
		return result
	}
	for i := range timings {
		if name, ok := names.Load(timings[i].Name); ok {
			timings[i].Name = name.(string)
		}
	}
	sort.Slice(timings, func(i, j int) bool { return timings[i].Name < timings[j].Name })
	result.timings = timings

	quiesceStart := time.Now()
	result.err = quiesce(ctx, flow, kube, namespace, labels, knobs.setupTimeout)
	result.quiesce = time.Since(quiesceStart)
	return result
}

func burst(ctx context.Context, flow workloadFlow, watcher *measure.Watcher, names *sync.Map, namespace, id string, index, size int, labels map[string]string, knobs settings) {
	jobs := make(chan int)
	var wg sync.WaitGroup
	for range knobs.concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				name := fmt.Sprintf("flowmark-burst-%s-r%d-%d", id, index, i+1)
				createCtx, cancel := context.WithTimeout(ctx, knobs.setupTimeout)
				key, err := flow.create(createCtx, item{name: name, namespace: namespace, labels: labels, image: knobs.image, pods: knobs.pods})
				cancel()
				acceptedAt := time.Now()
				if err != nil {
					watcher.MarkFailed(name, err)
					continue
				}
				names.Store(key, name)
				watcher.MarkCreated(key, acceptedAt)
			}
		}()
	}
	for i := range size {
		if ctx.Err() != nil {
			break
		}
		jobs <- i
	}
	close(jobs)
	wg.Wait()
}

func quiesce(ctx context.Context, flow workloadFlow, kube k8sclient.Client, namespace string, labels map[string]string, timeout time.Duration) error {
	var last string
	err := wait.PollUntilContextTimeout(ctx, quiescePoll, timeout, true, func(ctx context.Context) (bool, error) {
		leftover, err := flow.teardown(ctx, namespace, labels)
		if err != nil {
			last = err.Error()
			return false, nil
		}
		var pods corev1.PodList
		if err := kube.List(ctx, &pods, k8sclient.InNamespace(namespace)); err != nil {
			last = err.Error()
			return false, nil
		}
		last = fmt.Sprintf("%d workload objects and %d pods still present", leftover, len(pods.Items))
		return leftover == 0 && len(pods.Items) == 0, nil
	})
	if err != nil {
		return fmt.Errorf("namespace %s did not quiesce within %s (last: %s): %w", namespace, timeout, last, err)
	}
	return nil
}

func settle(ctx context.Context, pause time.Duration) {
	timer := time.NewTimer(pause)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

func buildSection(flowName string, rounds []round, knobs settings) core.Section {
	metrics := map[string]any{
		"flow":                           flowName,
		"rounds":                         knobs.rounds,
		"workload_type":                  knobs.workloadType,
		"pods_per_workload":              knobs.pods,
		"image":                          knobs.image,
		"submit_concurrency":             knobs.concurrency,
		"kube_requests_per_sec":          knobs.limits.KubeRequestsPerSecond,
		"control_plane_requests_per_sec": knobs.limits.ControlPlaneRequestsPerSecond,
		"settle_ms":                      knobs.settle.Milliseconds(),
	}
	if knobs.slo != nil {
		metrics["slo"] = knobs.slo
		if value, ok := knobs.slo[sloP90Key]; ok {
			metrics["slo_p90_ms"] = value
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
		metrics["unmatched_arrivals"+at] = r.unmatched
		metrics["measurement_interruptions"+at] = r.interruptions
		metrics["submit_makespan_ms"+at] = r.submitMakespan.Milliseconds()
		metrics["quiesce_ms"+at] = r.quiesce.Milliseconds()
		if summary.SampleSize > 0 {
			metrics["p50_ms"+at] = summary.P50.Milliseconds()
			metrics["p90_ms"+at] = summary.P90.Milliseconds()
			metrics["p99_ms"+at] = summary.P99.Milliseconds()
			metrics["mean_ms"+at] = summary.Mean.Milliseconds()
			metrics["max_ms"+at] = summary.Max.Milliseconds()
		}
		if summary.HasThroughput() {
			metrics["makespan_ms"+at] = summary.Makespan.Milliseconds()
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
		message := fmt.Sprintf("round %d (%d workloads): %d never reached the cluster within the timeout", r.index, r.size, summary.PendingCount)
		if err := lastError(r.timings, measure.Pending); errors.Is(err, measure.ErrWatchClosed) {
			message = fmt.Sprintf("round %d (%d workloads): the watch ended before %d arrived (%v)", r.index, r.size, summary.PendingCount, err)
		} else if err != nil {
			message += " (" + err.Error() + ")"
		}
		if r.unmatched > 0 && summary.FailedCount == 0 {
			message += fmt.Sprintf(" (%d objects arrived that matched no submitted workload; is the correlation field missing?)", r.unmatched)
		}
		problems = append(problems, message)
	}
	if r.interruptions > 0 {
		problems = append(problems, fmt.Sprintf("round %d (%d workloads): the measurement was interrupted %d times: the cluster watch closed and was reopened; "+
			"latencies of workloads in flight during the gaps are inflated", r.index, r.size, r.interruptions))
	}
	if summary.FailedCount > 0 {
		message := fmt.Sprintf("round %d (%d workloads): %d creates failed", r.index, r.size, summary.FailedCount)
		if err := lastError(r.timings, measure.Failed); err != nil {
			message += " (" + err.Error() + ")"
		}
		problems = append(problems, message)
	}
	if r.err != nil {
		problems = append(problems, fmt.Sprintf("round %d (%d workloads): %v", r.index, r.size, r.err))
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

func deleteAll(ctx context.Context, concurrency int, targets []k8sclient.Object, remove func(context.Context, k8sclient.Object) error) error {
	jobs := make(chan k8sclient.Object)
	errs := make(chan error, len(targets))
	var wg sync.WaitGroup
	for range concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for obj := range jobs {
				if err := remove(ctx, obj); err != nil {
					errs <- fmt.Errorf("deleting %s: %w", obj.GetName(), err)
				}
			}
		}()
	}
	for _, obj := range targets {
		jobs <- obj
	}
	close(jobs)
	wg.Wait()
	close(errs)
	var failures []error
	for err := range errs {
		failures = append(failures, err)
	}
	return errors.Join(failures...)
}

func countAllInNamespace(ctx context.Context, kube k8sclient.Client, kind schema.GroupVersionKind, namespace string) (int, error) {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(schema.GroupVersionKind{Group: kind.Group, Version: kind.Version, Kind: kind.Kind + "List"})
	if err := kube.List(ctx, list, k8sclient.InNamespace(namespace)); err != nil {
		return 0, fmt.Errorf("listing %s objects: %w", kind.Kind, err)
	}
	return len(list.Items), nil
}
