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
	"slices"
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
	projectPrefix        = "flowmark-burst"
	defaultImage         = "registry.k8s.io/pause:3.9"
	defaultConcurrency   = 16
	maxConcurrency       = 64
	maxRequestsPerSecond = 1000
	cleanupGrace         = time.Minute
	runLabel             = "flowmark/run"
	roundLabel           = "flowmark/round"
	roundSizeLabel       = "flowmark/round-size"
)

var sloKeys = []string{"p50_ms", "p90_ms", "p99_ms", "throughput_per_sec"}

var knownKnobs = []string{"workloads", "workloadType", "image", "submitConcurrency", "kubeRequestsPerSecond", "controlPlaneRequestsPerSecond", "timeout", "setupTimeout", "settle", "cluster", "slo"}

var quiescePoll = 2 * time.Second

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

type settings struct {
	rounds       []int
	workloadType string
	image        string
	concurrency  int
	limits       client.Limits
	timeout      time.Duration
	setupTimeout time.Duration
	settle       time.Duration
	poll         time.Duration
	cluster      string
	slo          map[string]int
}

type item struct {
	name      string
	namespace string
	labels    map[string]string
	image     string
}

type workloadFlow interface {
	watch(ctx context.Context, namespace string) (*measure.Watcher, error)
	create(ctx context.Context, workload item) (string, error)
	teardown(ctx context.Context, namespace string, labels map[string]string) (int, error)
}

type round struct {
	index         int
	size          int
	timings       []measure.Timing
	unmatched     int
	reopened      int
	interruptions int
	err           error
}

const suiteName = "scale-burst"

func (Suite) Name() string {
	return suiteName
}

func (Suite) Category() core.Category {
	return core.CategoryScale
}

func (Suite) Knobs() []string {
	return knownKnobs
}

func (Suite) Validate(scenario core.Scenario) error {
	_, err := parseSettings(scenario)
	return err
}

func (s Suite) Run(ctx context.Context, target core.Target, scenario core.Scenario) (section core.Section, err error) {
	knobs, err := parseSettings(scenario)
	if err != nil {
		return core.Section{}, err
	}

	session, err := s.session()(ctx, target, knobs.limits, knobs.cluster, projectPrefix, knobs.setupTimeout)
	if err != nil {
		return core.Section{}, err
	}
	clients, project, id := session.Clients, session.Project, session.RunID
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

	flow := newControlPlaneFlow(knobs.workloadType, clients, project, knobs.concurrency)
	if err := warmUp(ctx, flow, clients.Kube, project.Namespace, id, knobs); err != nil {
		return core.Section{}, err
	}
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
	return buildSection(rounds, knobs), nil
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
	if summary.NeverArrivedCount > 0 {
		text += fmt.Sprintf(", %d never arrived", summary.NeverArrivedCount)
	}
	if summary.FailedCount > 0 {
		text += fmt.Sprintf(", %d failed", summary.FailedCount)
	}
	if r.unmatched > 0 {
		text += fmt.Sprintf(", %d unmatched arrivals", r.unmatched)
	}
	switch {
	case r.interruptions > 0:
		text += fmt.Sprintf(", measurement interrupted %d times", r.interruptions)
	case r.reopened > 0:
		text += fmt.Sprintf(", watch reopened %d times without losing events", r.reopened)
	}
	if r.err != nil {
		text += ": " + r.err.Error()
	}
	return text
}

func newControlPlaneFlow(workloadType string, clients *client.Clients, project submit.ProjectRef, concurrency int) *controlPlaneFlow {
	return &controlPlaneFlow{workloadType: workloadType, clients: clients, project: project, concurrency: concurrency, created: map[string]string{}, deleteRequested: map[string]bool{}}
}

func validateSLOKey(key string) error {
	stat, round, perRound := strings.Cut(key, "@")
	if !slices.Contains(sloKeys, stat) {
		return fmt.Errorf("slo key %q is not a metric this suite reports; use one of %s, optionally with @<round size>", key, strings.Join(sloKeys, ", "))
	}
	if !perRound {
		return nil
	}
	if size, err := strconv.Atoi(round); err != nil || size < 1 {
		return fmt.Errorf("slo key %q must end with a round size, e.g. %s@500", key, stat)
	}
	return nil
}

func applicableSLOs(slo map[string]int, rounds []int) map[string]int {
	if slo == nil {
		return nil
	}
	applicable := make(map[string]int, len(slo))
	for key, value := range slo {
		if _, round, perRound := strings.Cut(key, "@"); perRound {
			if size, _ := strconv.Atoi(round); !slices.Contains(rounds, size) {
				continue
			}
		}
		applicable[key] = value
	}
	return applicable
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
	if knobs.workloadType, err = scenario.String("workloadType", submit.TrainingType); err != nil {
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
	knobs.poll = quiescePoll
	if knobs.slo, err = scenario.IntMap("slo"); err != nil {
		return settings{}, err
	}
	for key := range knobs.slo {
		if err := validateSLOKey(key); err != nil {
			return settings{}, err
		}
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
	if knobs.workloadType != submit.TrainingType && knobs.workloadType != submit.WorkspaceType {
		return settings{}, fmt.Errorf("workloadType %q is not supported, use %s or %s", knobs.workloadType, submit.TrainingType, submit.WorkspaceType)
	}
	return knobs, nil
}

func warmUp(ctx context.Context, flow workloadFlow, kube k8sclient.Client, namespace, id string, knobs settings) error {
	log.Printf("%s: warming up the new project with one %s workload before round 1", suiteName, knobs.workloadType)
	probe := runRound(ctx, flow, kube, namespace, id, 0, 1, knobs)
	summary := measure.Stats(probe.timings)
	switch {
	case ctx.Err() != nil:
		return fmt.Errorf("warm-up: %w", ctx.Err())
	case probe.err != nil:
		return fmt.Errorf("warm-up: %w", probe.err)
	case summary.NeverArrivedCount > 0:
		return fmt.Errorf("warm-up: the probe workload never reached the cluster within %s", knobs.timeout)
	case summary.FailedCount > 0:
		log.Printf("%s: warm-up absorbed a rejected create: %v", suiteName, lastError(probe.timings, measure.Failed))
	}
	return nil
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
	burst(ctx, flow, watcher, &names, namespace, id, index, size, labels, knobs)

	timings := watcher.WaitAll(knobs.timeout)
	result.unmatched = watcher.Unmatched()
	result.reopened = watcher.Reconnects()
	result.interruptions = watcher.Interruptions()
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

	result.err = quiesce(ctx, flow, kube, namespace, labels, knobs.setupTimeout, knobs.poll)
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
				key, err := flow.create(createCtx, item{name: name, namespace: namespace, labels: labels, image: knobs.image})
				cancel()
				acceptedAt := time.Now()
				if err != nil && ctx.Err() != nil {
					continue
				}
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

func quiesce(ctx context.Context, flow workloadFlow, kube k8sclient.Client, namespace string, labels map[string]string, timeout, poll time.Duration) error {
	var last string
	err := wait.PollUntilContextTimeout(ctx, poll, timeout, true, func(ctx context.Context) (bool, error) {
		leftover, err := flow.teardown(ctx, namespace, labels)
		if err != nil {
			if ctx.Err() == nil {
				last = err.Error()
			}
			return false, nil
		}
		var pods corev1.PodList
		if err := kube.List(ctx, &pods, k8sclient.InNamespace(namespace)); err != nil {
			if ctx.Err() == nil {
				last = err.Error()
			}
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

func buildSection(rounds []round, knobs settings) core.Section {
	metrics := map[string]any{"workload_type": knobs.workloadType}

	var problems []string
	for _, r := range rounds {
		at := fmt.Sprintf("@%d", r.size)
		summary := measure.Stats(r.timings)
		if summary.SampleSize > 0 {
			metrics["p50_ms"+at] = summary.P50.Milliseconds()
			metrics["p90_ms"+at] = summary.P90.Milliseconds()
			metrics["p99_ms"+at] = summary.P99.Milliseconds()
		}
		if summary.HasThroughput() {
			metrics["throughput_per_sec"+at] = summary.Throughput
		}
		problems = append(problems, roundProblems(r, summary)...)
	}
	if len(rounds) < len(knobs.rounds) {
		problems = append(problems, fmt.Sprintf("run stopped after %d of %d rounds", len(rounds), len(knobs.rounds)))
	}

	section := core.Section{Metrics: metrics, SLO: applicableSLOs(knobs.slo, knobs.rounds)}
	if len(problems) > 0 {
		section.Failed = true
		section.Error = strings.Join(problems, "; ")
	}
	return section
}

func roundProblems(r round, summary measure.Summary) []string {
	var problems []string
	if summary.NeverArrivedCount > 0 {
		message := fmt.Sprintf("round %d (%d workloads): %d never reached the cluster within the timeout", r.index, r.size, summary.NeverArrivedCount)
		if err := lastError(r.timings, measure.NeverArrived); errors.Is(err, measure.ErrWatchClosed) {
			message = fmt.Sprintf("round %d (%d workloads): the watch ended before %d arrived (%v)", r.index, r.size, summary.NeverArrivedCount, err)
		} else if err != nil {
			message += " (" + err.Error() + ")"
		}
		if r.unmatched > 0 && summary.FailedCount == 0 {
			message += fmt.Sprintf(" (%d objects arrived that matched no submitted workload; is the correlation field missing?)", r.unmatched)
		}
		problems = append(problems, message)
	}
	if r.interruptions > 0 {
		problems = append(problems, fmt.Sprintf("round %d (%d workloads): the measurement was interrupted %d times: the cluster watch was reopened without its resume point; "+
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

func countAllInNamespace(ctx context.Context, kube k8sclient.Client, kind schema.GroupVersionKind, namespace string) (int, error) {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(schema.GroupVersionKind{Group: kind.Group, Version: kind.Version, Kind: kind.Kind + "List"})
	if err := kube.List(ctx, list, k8sclient.InNamespace(namespace)); err != nil {
		return 0, fmt.Errorf("listing %s objects: %w", kind.Kind, err)
	}
	return len(list.Items), nil
}

type controlPlaneFlow struct {
	workloadType string
	clients      *client.Clients
	project      submit.ProjectRef
	concurrency  int

	mu              sync.Mutex
	created         map[string]string
	deleteRequested map[string]bool
}

func (f *controlPlaneFlow) watch(ctx context.Context, namespace string) (*measure.Watcher, error) {
	return measure.WatchWorkloads(ctx, f.clients.Kube, namespace)
}

func (f *controlPlaneFlow) create(ctx context.Context, workload item) (string, error) {
	spec := submit.WorkloadSpec{Name: workload.name, Project: f.project, Image: workload.image, Labels: workload.labels}
	var ref submit.WorkloadRef
	var err error
	if f.workloadType == submit.WorkspaceType {
		ref, err = submit.Workspace(ctx, f.clients, spec)
	} else {
		ref, err = submit.Training(ctx, f.clients, spec)
	}
	if err != nil {
		return "", err
	}
	f.mu.Lock()
	f.created[ref.ID] = ref.Name
	f.mu.Unlock()
	return ref.Name, nil
}

func (f *controlPlaneFlow) teardown(ctx context.Context, namespace string, _ map[string]string) (int, error) {
	listed, err := submit.ProjectWorkloadIDs(ctx, f.clients, f.project.ID, f.workloadType)
	if err != nil {
		return 0, err
	}
	f.mu.Lock()
	for _, id := range listed {
		if _, known := f.created[id]; !known {
			f.created[id] = ""
		}
	}
	var pending []string
	for id := range f.created {
		if !f.deleteRequested[id] {
			pending = append(pending, id)
		}
	}
	f.mu.Unlock()

	undeleted, err := submit.DeleteWorkloads(ctx, f.clients, f.workloadType, pending, f.concurrency)
	f.mu.Lock()
	for _, id := range pending {
		if !slices.Contains(undeleted, id) {
			f.deleteRequested[id] = true
		}
	}
	remaining := len(f.created) - len(f.deleteRequested)
	f.mu.Unlock()
	if err != nil {
		return remaining, err
	}
	runners, err := countAllInNamespace(ctx, f.clients.Kube, measure.WorkloadRunnerKind, namespace)
	if err != nil {
		return remaining, err
	}
	return remaining + runners, nil
}
