/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshots

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	snapshotv1alpha1 "github.com/ai-dynamo/snapshot/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/internal/snapshotkey"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/internal/snapshotrestore"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// timeoutAnnotation, on the workload, bounds a capture: scheduling, warm-up and
// the dump. Unset, Snapshot's one-hour default applies.
const timeoutAnnotation = "run.ai/snapshot-capture-timeout-seconds"

var (
	// errShapeDrift is reported: a capture would not match the replicas it restores into.
	errShapeDrift = errors.New("snapshots are off for this workload: its replicas don't share one shape")
	// errNoAgentNodes is reported, as the restore webhook treats it: a capture
	// pod with nowhere known to run would sit Pending for the whole timeout.
	errNoAgentNodes = errors.New("no nodes are known to run Snapshot's agent")
)

// captureLabel marks a capture pod.
const captureLabel = "run.ai/snapshot-capture"

// knativeLabelPrefix marks the labels Knative and Run:ai find a revision's pods by.
const knativeLabelPrefix = "serving.knative.dev/"

// droppedLabels are labels a replica carries that its capture copy must not: the
// opt-in and key, so the restore webhook and this plugin pass the copy by; the
// runner's name and UID, so it isn't counted as a replica; and the identity
// labels, so its GPU isn't attributed to the serving workload.
var droppedLabels = []string{
	constants.SnapshotOptInLabel,
	constants.SnapshotKeyLabel,
	commonconstants.WorkloadRunnerNameLabel,
	commonconstants.WorkloadRunnerUidLabel,
	commonconstants.TopOwnerWorkloadIdLabel,
	commonconstants.TopOwnerUidLabel, //nolint:staticcheck // deprecated, but metrics-exporter still attributes GPUs by it
	commonconstants.ClusterTopOwnerUidLabel,
}

// captureInput is what a capture needs besides the replica it copies.
type captureInput struct {
	key               string
	agentNodeSelector map[string]string
	serviceSelectors  []map[string]string
	timeoutSeconds    *int64
}

// HandleAdditionalResources deletes the workload's snapshots that are no longer
// needed, then captures each key from its first Ready replica, one capture at a
// time per workload, newest key first. Each capture event goes on the runner
// once per key, which puts it in the workload's history (§8.6).
//
// A problem is reported as a warning and never returned: an error here would
// stop the workload being applied.
func (p *Plugin) HandleAdditionalResources(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	resource *unstructured.Unstructured,
	karta *kartav1alpha1.Karta,
	k8sClient client.Client,
	state *reconcile.State,
) error {
	if !p.snapshotInstalled {
		return nil
	}
	// Opted in or not, so removing the opt-in deletes them too.
	if err := deleteUnused(ctx, k8sClient, runner, optedIn(resource)); err != nil {
		_ = state.AddPluginIssue(ctx, p, resource, err, false, "", issue.Compute)
	}
	if !optedIn(resource) {
		return nil
	}
	if err := checkEligibility(ctx, resource, karta); err != nil {
		return nil //nolint:nilerr // Status reports it.
	}
	events := p.captureEvents(runner)
	if err := errors.Join(p.capture(ctx, k8sClient, runner, resource, events), events.save()); err != nil {
		_ = state.AddPluginIssue(ctx, p, resource, err, false, "", issue.Compute)
	}
	return nil
}

func (p *Plugin) capture(ctx context.Context, c client.Client, runner *dromosv1alpha1.WorkloadRunner,
	workload *unstructured.Unstructured, events *captureEvents) error {
	jobs, err := runnerSnapshotJobs(ctx, c, runner)
	if err != nil {
		return err
	}
	running := false
	for i := range jobs {
		events.reportCapture(&jobs[i])
		running = running || !snapshotv1alpha1.IsSnapshotJobTerminal(&jobs[i])
	}
	// One at a time: each capture holds a GPU until it ends.
	if running {
		return nil
	}

	replicas, err := readyReplicas(ctx, c, runner)
	if err != nil {
		return err
	}
	replica, key := newestUncaptured(replicas, jobs)
	if replica == nil {
		return nil
	}
	if err := checkShape(replicas, replica); err != nil {
		if errors.Is(err, errShapeDrift) {
			events.emit(key, "", corev1.EventTypeWarning, eventDisabled, err.Error())
		}
		return err
	}
	if len(p.agentNodeSelector) == 0 {
		return errNoAgentNodes
	}

	in := captureInput{key: key, agentNodeSelector: p.agentNodeSelector}
	if in.timeoutSeconds, err = captureTimeout(workload); err != nil {
		return err
	}
	if in.serviceSelectors, err = serviceSelectors(ctx, c, runner.Namespace); err != nil {
		return err
	}
	job, err := buildSnapshotJob(runner, replica, in)
	if err != nil {
		return fmt.Errorf("capture key %s: %w", key, err)
	}
	if err := c.Create(ctx, job); err != nil {
		// Already there: an earlier reconcile created it and reported its start.
		if apierrors.IsAlreadyExists(err) {
			return nil
		}
		return fmt.Errorf("create SnapshotJob %s: %w", job.Name, err)
	}
	// Not recorded: a SnapshotJob is created once, so its start is reported once.
	p.events.Event(runner, corev1.EventTypeNormal, eventCaptureStarted, "Capturing snapshot "+job.Name)
	return nil
}

// Capture events, on the runner (§8.6).
const (
	eventCaptureStarted = "SnapshotCaptureStarted"
	eventCaptured       = "SnapshotCaptured"
	eventCaptureFailed  = "SnapshotCaptureFailed"
	eventDisabled       = "SnapshotDisabled"
)

// captureEvents emits each event a capture reaches after its start once per
// key. Its record goes in the runner's status, so a restarted controller
// neither loses nor repeats one.
type captureEvents struct {
	recorder record.EventRecorder
	runner   *dromosv1alpha1.WorkloadRunner
	info     statusInfo
	changed  bool
}

func (p *Plugin) captureEvents(runner *dromosv1alpha1.WorkloadRunner) *captureEvents {
	// An unreadable record starts over, which can repeat an event, so it
	// doesn't hold up the capture; Status rewrites it.
	info, _ := recordedStatus(runner)
	return &captureEvents{recorder: p.events, runner: runner, info: info}
}

// reportCapture emits how an existing capture ended, if it has.
func (e *captureEvents) reportCapture(job *snapshotv1alpha1.SnapshotJob) {
	// The deletion is dropping its key.
	if job.DeletionTimestamp != nil {
		return
	}
	key := job.Labels[constants.SnapshotKeyLabel]
	if meta.IsStatusConditionTrue(job.Status.Conditions, snapshotv1alpha1.SnapshotJobConditionCaptured) {
		e.emit(key, job.UID, corev1.EventTypeNormal, eventCaptured,
			"Captured snapshot "+job.Name+"; new replicas restore from it")
	}
	if snapshotv1alpha1.IsSnapshotJobFailed(job) {
		e.emit(key, job.UID, corev1.EventTypeWarning, eventCaptureFailed, captureFailure(job).Error())
	}
}

// emit emits reason for key unless the record has it. A capture with a UID
// other than the recorded one is new, and starts the key's record over.
func (e *captureEvents) emit(key string, job types.UID, eventType, reason, message string) {
	emitted := e.info.Events[key]
	if job != "" && job != emitted.SnapshotJob {
		emitted = keyEvents{SnapshotJob: job}
	}
	if slices.Contains(emitted.Reasons, reason) {
		return
	}
	e.recorder.Event(e.runner, eventType, reason, message)
	emitted.Reasons = append(emitted.Reasons, reason)
	if e.info.Events == nil {
		e.info.Events = map[string]keyEvents{}
	}
	e.info.Events[key] = emitted
	e.changed = true
}

// save puts the record on the runner's plugin status, which the framework hands
// to Status and then applies.
func (e *captureEvents) save() error {
	if !e.changed {
		return nil
	}
	return writeStatusInfo(pluginStatus(e.runner), e.info)
}

// deleteUnused deletes the runner's snapshots outside the keep-set. A failed
// PodSnapshot goes even when kept, but its SnapshotJob stays, so the key doesn't
// capture again until the spec changes.
func deleteUnused(ctx context.Context, c client.Client, runner *dromosv1alpha1.WorkloadRunner, optedIn bool) error {
	snapshots, err := runnerPodSnapshots(ctx, c, runner)
	if err != nil {
		return err
	}
	jobs, err := runnerSnapshotJobs(ctx, c, runner)
	if err != nil {
		return err
	}
	// Most workloads have none, and then their replicas needn't be listed.
	if len(snapshots) == 0 && len(jobs) == 0 {
		return nil
	}
	keep, err := keepSet(ctx, c, runner, optedIn)
	if err != nil {
		return err
	}

	for i := range snapshots {
		snapshot := &snapshots[i]
		if keep.Has(snapshot.Labels[constants.SnapshotKeyLabel]) && !snapshotv1alpha1.IsPodSnapshotFailed(snapshot) {
			continue
		}
		// Snapshot's finalizer removes the checkpoint from the volume.
		if err := c.Delete(ctx, snapshot); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("delete PodSnapshot %s: %w", snapshot.Name, err)
		}
	}
	for i := range jobs {
		if keep.Has(jobs[i].Labels[constants.SnapshotKeyLabel]) {
			continue
		}
		// With its PodSnapshot, not after it as the design has it: Snapshot keeps a
		// failed SnapshotJob's capture Job, so only this stops a capture still running.
		if err := c.Delete(ctx, &jobs[i]); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("delete SnapshotJob %s: %w", jobs[i].Name, err)
		}
	}
	return nil
}

// keepSet is the keys whose snapshots stay: the replicas' and, while the workload
// opts in, its newest, so the first replica after a scale to zero restores.
func keepSet(ctx context.Context, c client.Reader, runner *dromosv1alpha1.WorkloadRunner,
	optedIn bool) (datastructs.Set[string], error) {
	replicas, err := keyedReplicas(ctx, c, runner)
	if err != nil {
		return nil, err
	}
	if !optedIn {
		return keysInUse(replicas, ""), nil
	}
	info, err := recordedStatus(runner)
	if err != nil {
		return nil, err
	}
	return keysInUse(replicas, info.NewestKey), nil
}

// keysInUse is the replicas' keys, and newest unless it's empty.
func keysInUse(replicas []corev1.Pod, newest string) datastructs.Set[string] {
	keep := datastructs.NewSet[string]()
	for i := range replicas {
		key, _ := snapshotkey.SnapshotKeyOf(&replicas[i])
		keep.Add(key)
	}
	if newest != "" {
		keep.Add(newest)
	}
	return keep
}

// runnerSnapshotJobs lists the runner's captures. The UID tells them from those
// of an earlier runner of the same name.
func runnerSnapshotJobs(ctx context.Context, c client.Reader,
	runner *dromosv1alpha1.WorkloadRunner) ([]snapshotv1alpha1.SnapshotJob, error) {
	list := &snapshotv1alpha1.SnapshotJobList{}
	if err := c.List(ctx, list, client.InNamespace(runner.Namespace), client.MatchingLabels{
		commonconstants.WorkloadRunnerNameLabel: runner.Name,
		commonconstants.WorkloadRunnerUidLabel:  string(runner.UID),
	}); err != nil {
		return nil, fmt.Errorf("list SnapshotJobs: %w", err)
	}
	return list.Items, nil
}

// runnerPodSnapshots lists the runner's PodSnapshots, which Snapshot labels from
// the SnapshotJob's template. The UID leaves an earlier runner's to the cleanup.
func runnerPodSnapshots(ctx context.Context, c client.Reader,
	runner *dromosv1alpha1.WorkloadRunner) ([]snapshotv1alpha1.PodSnapshot, error) {
	list := &snapshotv1alpha1.PodSnapshotList{}
	if err := c.List(ctx, list, client.InNamespace(runner.Namespace), client.MatchingLabels{
		commonconstants.WorkloadRunnerNameLabel: runner.Name,
		commonconstants.WorkloadRunnerUidLabel:  string(runner.UID),
	}); err != nil {
		return nil, fmt.Errorf("list PodSnapshots: %w", err)
	}
	return list.Items, nil
}

// keyedReplicas lists the runner's replicas that carry a key. One without a key
// was admitted while the restore webhook was down.
func keyedReplicas(ctx context.Context, c client.Reader, runner *dromosv1alpha1.WorkloadRunner) ([]corev1.Pod, error) {
	pods := &corev1.PodList{}
	if err := c.List(ctx, pods, client.InNamespace(runner.Namespace),
		client.MatchingLabels{
			constants.SnapshotOptInLabel:            "true",
			commonconstants.WorkloadRunnerNameLabel: runner.Name,
			commonconstants.WorkloadRunnerUidLabel:  string(runner.UID),
		},
		client.HasLabels{constants.SnapshotKeyLabel}); err != nil {
		return nil, fmt.Errorf("list replicas: %w", err)
	}
	return pods.Items, nil
}

// readyReplicas lists the runner's keyed replicas that are Ready.
func readyReplicas(ctx context.Context, c client.Reader, runner *dromosv1alpha1.WorkloadRunner) ([]corev1.Pod, error) {
	replicas, err := keyedReplicas(ctx, c, runner)
	if err != nil {
		return nil, err
	}
	var ready []corev1.Pod
	for i := range replicas {
		if replicas[i].DeletionTimestamp == nil && podReady(&replicas[i]) {
			ready = append(ready, replicas[i])
		}
	}
	return ready, nil
}

// newestUncaptured picks the newest Ready replica whose key has no capture yet,
// so during a rollout the new key captures first.
func newestUncaptured(replicas []corev1.Pod, jobs []snapshotv1alpha1.SnapshotJob) (*corev1.Pod, string) {
	captured := datastructs.NewSet[string]()
	for i := range jobs {
		captured.Add(jobs[i].Labels[constants.SnapshotKeyLabel])
	}
	var newest *corev1.Pod
	for i := range replicas {
		key, _ := snapshotkey.SnapshotKeyOf(&replicas[i])
		if captured.Has(key) {
			continue
		}
		if newest == nil || replicas[i].CreationTimestamp.After(newest.CreationTimestamp.Time) {
			newest = &replicas[i]
		}
	}
	if newest == nil {
		return nil, ""
	}
	key, _ := snapshotkey.SnapshotKeyOf(newest)
	return newest, key
}

// checkShape fails when the replicas don't share the shape their keys promise:
// two of one version with different keys, or the replica about to be copied no
// longer hashing to its own key, as when a webhook after ours changed it.
func checkShape(replicas []corev1.Pod, replica *corev1.Pod) error {
	keyOfVersion := map[types.UID]string{}
	for i := range replicas {
		version, err := snapshotkey.Scope(&replicas[i])
		if err != nil {
			continue
		}
		key, _ := snapshotkey.SnapshotKeyOf(&replicas[i])
		if other, seen := keyOfVersion[version]; seen && other != key {
			return fmt.Errorf("%w: replicas of one version have keys %s and %s", errShapeDrift, other, key)
		}
		keyOfVersion[version] = key
	}

	matches, err := snapshotkey.Matches(replica)
	if err != nil {
		return err
	}
	if !matches {
		return fmt.Errorf("%w: replica %s was changed after its key was computed", errShapeDrift, replica.Name)
	}
	return nil
}

func captureTimeout(workload *unstructured.Unstructured) (*int64, error) {
	value, set := workload.GetAnnotations()[timeoutAnnotation]
	if !set {
		return nil, nil
	}
	seconds, err := strconv.ParseInt(value, 10, 64)
	if err != nil || seconds <= 0 {
		return nil, fmt.Errorf("%s must be a whole number of seconds above 0, not %q", timeoutAnnotation, value)
	}
	return &seconds, nil
}

func serviceSelectors(ctx context.Context, c client.Reader, namespace string) ([]map[string]string, error) {
	services := &corev1.ServiceList{}
	if err := c.List(ctx, services, client.InNamespace(namespace)); err != nil {
		return nil, fmt.Errorf("list Services: %w", err)
	}
	selectors := make([]map[string]string, 0, len(services.Items))
	for i := range services.Items {
		selectors = append(selectors, services.Items[i].Spec.Selector)
	}
	return selectors, nil
}

// buildSnapshotJob copies a Ready replica into the SnapshotJob that captures its
// key. The runner doesn't own it: its labels tie it to the runner, and
// NewCleanupSetup deletes it once the runner is gone.
func buildSnapshotJob(runner *dromosv1alpha1.WorkloadRunner, replica *corev1.Pod,
	in captureInput) (*snapshotv1alpha1.SnapshotJob, error) {
	target, err := snapshotkey.TargetContainer(replica)
	if err != nil {
		return nil, err
	}
	spec, err := captureSpec(replica, target, in.agentNodeSelector)
	if err != nil {
		return nil, err
	}

	identity := map[string]string{
		constants.SnapshotKeyLabel:              in.key,
		commonconstants.WorkloadRunnerNameLabel: runner.Name,
		commonconstants.WorkloadRunnerUidLabel:  string(runner.UID),
	}
	return &snapshotv1alpha1.SnapshotJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      snapshotkey.SnapshotName(in.key, runner.Name),
			Namespace: replica.Namespace,
			Labels:    identity,
		},
		Spec: snapshotv1alpha1.SnapshotJobSpec{
			PodTemplate: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: captureLabels(replica.Labels, in.serviceSelectors)},
				Spec:       *spec,
			},
			ActiveDeadlineSeconds: in.timeoutSeconds,
			PodSnapshotTemplate: snapshotv1alpha1.PodSnapshotTemplate{
				// Copied to the PodSnapshot, so it is tied to its runner and key too.
				Metadata:         &snapshotv1alpha1.PodSnapshotTemplateMetadata{Labels: maps.Clone(identity)},
				TargetContainers: []string{target},
			},
		},
	}, nil
}

// captureSpec is the replica's pod spec as its capture pod runs it.
func captureSpec(replica *corev1.Pod, target string, agentNodeSelector map[string]string) (*corev1.PodSpec, error) {
	spec := replica.Spec.DeepCopy()

	container := spec.Containers[slices.IndexFunc(spec.Containers, func(container corev1.Container) bool {
		return container.Name == target
	})]
	// It can call a container the copy leaves out, such as Knative's queue-proxy.
	if container.Lifecycle != nil {
		container.Lifecycle.PreStop = nil
	}
	spec.Containers = []corev1.Container{container}

	// The scheduler and admission set these for the replica; the capture pod
	// gets its own.
	spec.NodeName = ""
	spec.SchedulingGates = nil
	spec.Overhead = nil
	spec.EphemeralContainers = nil
	// The capture is a workload of its own, not the serving workload's priority.
	spec.PriorityClassName = ""
	spec.Priority = nil
	// A Job's pod cannot restart Always.
	spec.RestartPolicy = corev1.RestartPolicyNever

	if err := snapshotrestore.RequireNodeSelector(spec, agentNodeSelector); err != nil {
		return nil, err
	}
	return spec, nil
}

// captureLabels are the replica's labels, without those that would send the
// capture pod traffic or count it as the serving workload's.
func captureLabels(replica map[string]string, serviceSelectors []map[string]string) map[string]string {
	captured := maps.Clone(replica)
	if captured == nil {
		captured = map[string]string{}
	}
	for _, key := range droppedLabels {
		delete(captured, key)
	}

	maps.DeleteFunc(captured, func(key, _ string) bool {
		return strings.HasPrefix(key, knativeLabelPrefix) || strings.HasPrefix(key, constants.ElementLabelPrefix)
	})

	// Labels any of these select the replica by are dropped, so the capture pod
	// gets no traffic.
	for _, selector := range serviceSelectors {
		// A Service without a selector selects no pods.
		if len(selector) > 0 && labels.SelectorFromSet(selector).Matches(labels.Set(replica)) {
			for key := range selector {
				delete(captured, key)
			}
		}
	}
	captured[captureLabel] = "true"
	return captured
}

func podReady(pod *corev1.Pod) bool {
	return slices.ContainsFunc(pod.Status.Conditions, func(condition corev1.PodCondition) bool {
		return condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue
	})
}
