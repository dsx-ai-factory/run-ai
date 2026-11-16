/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshots

import (
	"context"
	"encoding/json"
	"strings"

	snapshotv1alpha1 "github.com/ai-dynamo/snapshot/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/internal/snapshotkey"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

var agentSelector = map[string]string{"nvidia.com/snapshot-node": "true"}

var _ = Describe("HandleAdditionalResources", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	gpu := gpuContainer("main")
	optedIn := func() *unstructured.Unstructured { return deployment(optIn, gpu) }

	It("captures the newest uncaptured replica, labelled with its runner and not owned by it", func() {
		older := replica("llama-old", "rs-old", 100, true)
		newer := replica("llama-new", "rs-new", 200, true)
		c := captureClient(older, newer)

		Expect(handle(ctx, capturePlugin(), c, optedIn())).To(BeEmpty())

		jobs := snapshotJobs(ctx, c)
		Expect(jobs).To(HaveLen(1))
		Expect(jobs[0].Name).To(Equal(snapshotkey.SnapshotName(keyOf(newer), "llama")), "want the newest key")
		Expect(jobs[0].Labels).To(HaveKeyWithValue(commonconstants.WorkloadRunnerUidLabel, "runner-uid"))
		Expect(jobs[0].OwnerReferences).To(BeEmpty())
	})

	It("runs one capture at a time", func() {
		older := replica("llama-old", "rs-old", 100, true)
		newer := replica("llama-new", "rs-new", 200, true)
		c := captureClient(older, newer, capturedJob(keyOf(older), false))

		handle(ctx, capturePlugin(), c, optedIn())

		Expect(snapshotJobs(ctx, c)).To(HaveLen(1), "want only the capture already running")
	})

	// A SnapshotJob also records a failed capture, so its key is never captured again.
	It("skips a captured key", func() {
		older := replica("llama-old", "rs-old", 100, true)
		newer := replica("llama-new", "rs-new", 200, true)
		c := captureClient(older, newer, capturedJob(keyOf(newer), true))

		handle(ctx, capturePlugin(), c, optedIn())

		Expect(hasSnapshotJob(ctx, c, keyOf(older))).To(BeTrue(), "the other key was not captured")
		Expect(snapshotJobs(ctx, c)).To(HaveLen(2), "want the finished one and one new")
	})

	// A runner deleted and resubmitted under the same name leaves captures that are
	// not this runner's, and must not hold up its first capture.
	It("ignores an earlier runner's captures", func() {
		ready := replica("llama-a", "rs-1", 100, true)
		earlier := capturedJob("old-key", false)
		earlier.Labels[commonconstants.WorkloadRunnerUidLabel] = "earlier-runner-uid"
		c := captureClient(ready, earlier)

		handle(ctx, capturePlugin(), c, optedIn())

		Expect(hasSnapshotJob(ctx, c, keyOf(ready))).To(BeTrue(), "the replica was not captured")
	})

	// The earlier runner's pods can outlive it, under the same name.
	It("ignores an earlier runner's replicas", func() {
		earlier := replica("llama-old", "rs-old", 100, true, func(pod *corev1.Pod) {
			pod.Labels[commonconstants.WorkloadRunnerUidLabel] = "earlier-runner-uid"
		})
		c := captureClient(earlier)

		handle(ctx, capturePlugin(), c, optedIn())

		Expect(snapshotJobs(ctx, c)).To(BeEmpty())
	})

	It("needs a Ready replica", func() {
		c := captureClient(replica("llama-cold", "rs-1", 100, false))

		handle(ctx, capturePlugin(), c, optedIn())

		Expect(snapshotJobs(ctx, c)).To(BeEmpty())
	})

	notInstalled := capturePlugin()
	notInstalled.snapshotInstalled = false

	DescribeTable("stays off",
		func(plugin *Plugin, workload func() *unstructured.Unstructured) {
			c := captureClient(replica("llama-a", "rs-1", 100, true))

			Expect(handle(ctx, plugin, c, workload())).To(BeEmpty())
			Expect(snapshotJobs(ctx, c)).To(BeEmpty())
		},
		Entry("Snapshot not installed", notInstalled, optedIn),
		Entry("not opted in", capturePlugin(),
			func() *unstructured.Unstructured { return deployment(nil, gpu) }),
		Entry("ineligible", capturePlugin(),
			func() *unstructured.Unstructured { return deployment(optIn, gpu, gpuContainer("sidecar")) }),
	)

	newImage := func(pod *corev1.Pod) { pod.Spec.Containers[0].Image = "vllm/vllm-openai:v0.32.0" }
	oneReady := func() []client.Object { return []client.Object{replica("llama-a", "rs-1", 100, true)} }
	noAgent := capturePlugin()
	noAgent.agentNodeSelector = nil

	DescribeTable("reports what stops it, and captures nothing",
		func(plugin *Plugin, workload func() *unstructured.Unstructured, replicas func() []client.Object, want string) {
			c := captureClient(replicas()...)

			Expect(handle(ctx, plugin, c, workload())).To(ConsistOf(ContainSubstring(want)))
			Expect(snapshotJobs(ctx, c)).To(BeEmpty())
		},
		Entry("replicas of one version with different keys", capturePlugin(), optedIn,
			func() []client.Object {
				return []client.Object{
					replica("llama-a", "rs-1", 100, true),
					replica("llama-b", "rs-1", 200, true, newImage),
				}
			}, "don't share one shape"),
		Entry("a replica changed after its key", capturePlugin(), optedIn,
			func() []client.Object {
				changed := replica("llama-a", "rs-1", 100, true)
				// A webhook after ours, adding something random to each replica.
				changed.Spec.Containers[0].Env = append(changed.Spec.Containers[0].Env,
					corev1.EnvVar{Name: "RANDOM", Value: "42"})
				return []client.Object{changed}
			}, "don't share one shape"),
		Entry("placement that excludes the agent", capturePlugin(), optedIn,
			func() []client.Object {
				return []client.Object{replica("llama-a", "rs-1", 100, true, func(pod *corev1.Pod) {
					pod.Spec.NodeSelector = map[string]string{"nvidia.com/snapshot-node": "false"}
				})}
			}, "excludes every node"),
		Entry("no agent nodes known", noAgent, optedIn, oneReady, "no nodes are known"),
		Entry("a timeout that isn't a number", capturePlugin(),
			func() *unstructured.Unstructured {
				return deployment(map[string]string{optInAnnotation: "true", timeoutAnnotation: "soon"}, gpu)
			}, oneReady, timeoutAnnotation),
	)

	It("sets the timeout", func() {
		c := captureClient(replica("llama-a", "rs-1", 100, true))
		workload := deployment(map[string]string{optInAnnotation: "true", timeoutAnnotation: "7200"}, gpu)

		handle(ctx, capturePlugin(), c, workload)

		jobs := snapshotJobs(ctx, c)
		Expect(jobs).To(HaveLen(1))
		Expect(jobs[0].Spec.ActiveDeadlineSeconds).To(HaveValue(Equal(int64(7200))))
	})
})

var _ = Describe("Capture events", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	gpu := gpuContainer("main")
	optedIn := func() *unstructured.Unstructured { return deployment(optIn, gpu) }

	// Creating the SnapshotJob succeeds once, so the start needs no record.
	It("reports a capture's start when it creates the SnapshotJob", func() {
		ready := replica("llama-a", "rs-1", 100, true)
		plugin, runner := capturePlugin(), testRunner()

		handleRunner(ctx, plugin, runner, captureClient(ready), optedIn())

		Expect(emitted(plugin)).To(ConsistOf(
			"Normal SnapshotCaptureStarted Capturing snapshot " + snapshotkey.SnapshotName(keyOf(ready), "llama")))
		Expect(runner.Status.Plugins).To(BeEmpty())
	})

	DescribeTable("reports how an existing capture ended, once",
		func(job func(key string) *snapshotv1alpha1.SnapshotJob, want []string) {
			ready := replica("llama-a", "rs-1", 100, true)
			c := captureClient(ready, job(keyOf(ready)))
			plugin, runner := capturePlugin(), testRunner()

			handleRunner(ctx, plugin, runner, c, optedIn())
			Expect(reasons(emitted(plugin))).To(Equal(want))

			// The next reconcile finds the record on the runner's status.
			handleRunner(ctx, plugin, runner, c, optedIn())
			Expect(emitted(plugin)).To(BeEmpty())
		},
		Entry("running", func(key string) *snapshotv1alpha1.SnapshotJob { return capturedJob(key, false) },
			[]string(nil)),
		Entry("captured", func(key string) *snapshotv1alpha1.SnapshotJob { return jobCaptured(capturedJob(key, true)) },
			[]string{eventCaptured}),
		Entry("failed", func(key string) *snapshotv1alpha1.SnapshotJob { return jobFailed(capturedJob(key, false)) },
			[]string{eventCaptureFailed}),
	)

	It("says why a capture failed", func() {
		ready := replica("llama-a", "rs-1", 100, true)
		plugin := capturePlugin()

		handle(ctx, plugin, captureClient(ready, jobFailed(capturedJob(keyOf(ready), false))), optedIn())

		Expect(emitted(plugin)).To(ContainElement(SatisfyAll(
			HavePrefix("Warning SnapshotCaptureFailed"),
			ContainSubstring("DeadlineExceeded: the capture did not finish in time"))))
	})

	// The record is what a restarted controller starts from.
	It("repeats nothing after a restart", func() {
		ready := replica("llama-a", "rs-1", 100, true)
		job := withUID(jobCaptured(capturedJob(keyOf(ready), true)), "job-1")
		runner := withRecord(testRunner(), statusInfo{Events: map[string]keyEvents{
			keyOf(ready): {SnapshotJob: "job-1", Reasons: []string{eventCaptured}},
		}})
		plugin := capturePlugin()

		handleRunner(ctx, plugin, runner, captureClient(ready, job), optedIn())

		Expect(emitted(plugin)).To(BeEmpty())
	})

	It("reports the end of a capture created again after its SnapshotJob was deleted by hand", func() {
		ready := replica("llama-a", "rs-1", 100, true)
		runner := withRecord(testRunner(), statusInfo{Events: map[string]keyEvents{
			keyOf(ready): {SnapshotJob: "job-1", Reasons: []string{eventCaptureFailed}},
		}})
		plugin := capturePlugin()
		again := withUID(jobFailed(capturedJob(keyOf(ready), false)), "job-2")

		handleRunner(ctx, plugin, runner, captureClient(ready, again), optedIn())

		Expect(reasons(emitted(plugin))).To(Equal([]string{eventCaptureFailed}))
		Expect(recorded(runner).Events[keyOf(ready)]).To(Equal(
			keyEvents{SnapshotJob: "job-2", Reasons: []string{eventCaptureFailed}}))
	})

	It("reports a shape check turning snapshots off, once", func() {
		c := captureClient(
			replica("llama-a", "rs-1", 100, true),
			replica("llama-b", "rs-1", 200, true, func(pod *corev1.Pod) { pod.Spec.Containers[0].Image = "other" }))
		plugin, runner := capturePlugin(), testRunner()

		handleRunner(ctx, plugin, runner, c, optedIn())
		Expect(emitted(plugin)).To(ConsistOf(SatisfyAll(
			HavePrefix("Warning SnapshotDisabled"), ContainSubstring("don't share one shape"))))

		handleRunner(ctx, plugin, runner, c, optedIn())
		Expect(emitted(plugin)).To(BeEmpty())
	})

	// Only a shape check turns snapshots off; a replica whose key can't be
	// recomputed is a warning.
	It("doesn't report a key it can't recompute as turned off", func() {
		orphan := replica("llama-a", "rs-1", 100, true)
		orphan.OwnerReferences = nil
		plugin := capturePlugin()

		Expect(handle(ctx, plugin, captureClient(orphan), optedIn())).NotTo(BeEmpty())
		Expect(emitted(plugin)).To(BeEmpty())
	})

	It("says nothing of a capture being deleted", func() {
		ready := replica("llama-a", "rs-1", 100, true)
		deleting := jobCaptured(capturedJob(keyOf(ready), true))
		deleting.Finalizers = []string{"nvidia.com/snapshot-cleanup"}
		deleting.DeletionTimestamp = ptr.To(metav1.Unix(400, 0))
		plugin := capturePlugin()

		handle(ctx, plugin, captureClient(ready, deleting), optedIn())

		Expect(emitted(plugin)).To(BeEmpty())
	})

	// A stale cache can miss a SnapshotJob an earlier reconcile created, and
	// reported.
	It("says nothing of a capture another reconcile created", func() {
		ready := replica("llama-a", "rs-1", 100, true)
		c := interceptor.NewClient(captureClient(ready).(client.WithWatch), interceptor.Funcs{
			Create: func(context.Context, client.WithWatch, client.Object, ...client.CreateOption) error {
				return apierrors.NewAlreadyExists(schema.GroupResource{Group: "nvidia.com", Resource: "snapshotjobs"}, "")
			},
		})
		plugin := capturePlugin()

		Expect(handle(ctx, plugin, c, optedIn())).To(BeEmpty())
		Expect(emitted(plugin)).To(BeEmpty())
	})

	// Status rewrites the record; meanwhile the capture mustn't wait, since
	// nothing reconciles the runner again until something changes.
	It("captures, and reports, despite a record it can't read", func() {
		ready := replica("llama-a", "rs-1", 100, true)
		c := captureClient(ready)
		plugin := capturePlugin()

		handleRunner(ctx, plugin, withUnreadableRecord(testRunner()), c, optedIn())

		Expect(hasSnapshotJob(ctx, c, keyOf(ready))).To(BeTrue())
		Expect(reasons(emitted(plugin))).To(Equal([]string{eventCaptureStarted}))
	})

	It("leaves the runner's status alone when there's nothing to report", func() {
		plugin, runner := capturePlugin(), testRunner()

		handleRunner(ctx, plugin, runner, captureClient(replica("llama-cold", "rs-1", 100, false)), optedIn())

		Expect(runner.Status.Plugins).To(BeEmpty())
	})
})

var _ = Describe("Deleting unused snapshots", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	gpu := gpuContainer("main")
	optedIn := func() *unstructured.Unstructured { return deployment(optIn, gpu) }
	optedOut := func() *unstructured.Unstructured { return deployment(nil, gpu) }

	It("deletes the snapshots, and a capture still running, of a key no replica has", func() {
		current := replica("llama-new", "rs-new", 200, true)
		c := captureClient(current,
			capturedJob(keyOf(current), true), capturedPodSnapshot(keyOf(current), "runner-uid"),
			capturedJob("old-key", false), capturedPodSnapshot("old-key", "runner-uid"))

		Expect(handle(ctx, capturePlugin(), c, optedIn())).To(BeEmpty())

		Expect(snapshotJobs(ctx, c)).To(ConsistOf(named(keyOf(current))))
		Expect(podSnapshots(ctx, c)).To(ConsistOf(named(keyOf(current))))
	})

	It("keeps the newest key while no replica runs, so the first one after a scale to zero restores", func() {
		c := captureClient(
			capturedJob("k1", true), capturedPodSnapshot("k1", "runner-uid"),
			capturedJob("k2", true), capturedPodSnapshot("k2", "runner-uid"))

		handleRunner(ctx, capturePlugin(), withNewestKey(testRunner(), "k2"), c, optedIn())

		Expect(snapshotJobs(ctx, c)).To(ConsistOf(named("k2")))
		Expect(podSnapshots(ctx, c)).To(ConsistOf(named("k2")))
	})

	DescribeTable("keeps a key while a replica has it",
		func(workload func() *unstructured.Unstructured) {
			running := replica("llama-a", "rs-1", 100, true)
			c := captureClient(running,
				capturedJob(keyOf(running), true), capturedPodSnapshot(keyOf(running), "runner-uid"))

			handle(ctx, capturePlugin(), c, workload())

			Expect(snapshotJobs(ctx, c)).To(ConsistOf(named(keyOf(running))))
			Expect(podSnapshots(ctx, c)).To(ConsistOf(named(keyOf(running))))
		},
		Entry("opted in", optedIn),
		Entry("opt-in removed", optedOut),
	)

	It("deletes the newest key's snapshots too once the opt-in is removed", func() {
		c := captureClient(capturedJob("k2", true), capturedPodSnapshot("k2", "runner-uid"))

		handleRunner(ctx, capturePlugin(), withNewestKey(testRunner(), "k2"), c, optedOut())

		Expect(snapshotJobs(ctx, c)).To(BeEmpty())
		Expect(podSnapshots(ctx, c)).To(BeEmpty())
	})

	It("removes a failed PodSnapshot but keeps its SnapshotJob, so the key doesn't capture again", func() {
		running := replica("llama-a", "rs-1", 100, true)
		c := captureClient(running,
			capturedJob(keyOf(running), true), failed(capturedPodSnapshot(keyOf(running), "runner-uid")))

		handle(ctx, capturePlugin(), c, optedIn())

		Expect(podSnapshots(ctx, c)).To(BeEmpty())
		jobs := snapshotJobs(ctx, c)
		Expect(jobs).To(ConsistOf(named(keyOf(running))))
		Expect(snapshotv1alpha1.IsSnapshotJobTerminal(&jobs[0])).To(BeTrue(), "want the finished capture, not a new one")
	})

	// Without the newest key it can't tell which snapshot a workload at zero needs.
	It("deletes nothing while it can't read the record", func() {
		c := captureClient(capturedJob("k1", true), capturedPodSnapshot("k1", "runner-uid"))

		warnings := handleRunner(ctx, capturePlugin(), withUnreadableRecord(testRunner()), c, optedIn())

		Expect(warnings).To(ConsistOf(ContainSubstring("recorded status")))
		Expect(snapshotJobs(ctx, c)).To(HaveLen(1))
		Expect(podSnapshots(ctx, c)).To(HaveLen(1))
	})

	// The cleanup deletes them once that runner is gone.
	It("leaves an earlier runner's snapshots", func() {
		c := captureClient(capturedPodSnapshot("old-key", "earlier-runner-uid"))

		handle(ctx, capturePlugin(), c, optedIn())

		Expect(podSnapshots(ctx, c)).To(HaveLen(1))
	})

	// Listing a kind whose CRD is missing fails, so it would warn on every reconcile.
	It("does nothing without Snapshot", func() {
		plugin := capturePlugin()
		plugin.snapshotInstalled = false
		c := captureClient(capturedJob("old-key", true), capturedPodSnapshot("old-key", "runner-uid"))

		Expect(handle(ctx, plugin, c, optedIn())).To(BeEmpty())

		Expect(snapshotJobs(ctx, c)).To(HaveLen(1))
		Expect(podSnapshots(ctx, c)).To(HaveLen(1))
	})
})

var _ = Describe("Integration", func() {
	It("gives the built plugin its event recorder", func() {
		integration := NewIntegration(true, agentSelector)
		recorder := record.NewFakeRecorder(1)

		plugin, err := integration.Build(context.Background(), framework.PluginDeps{})
		Expect(err).NotTo(HaveOccurred())
		Expect(integration.Setup(recorder)).To(Succeed())

		Expect(plugin.(*Plugin).events).To(BeIdenticalTo(recorder))
		Expect(plugin.(*Plugin).agentNodeSelector).To(Equal(agentSelector))
	})

	// Capture events are emitted mid-reconcile, where a missing recorder would panic.
	It("fails a setup that has no plugin to give it to", func() {
		Expect(NewIntegration(true, agentSelector).Setup(record.NewFakeRecorder(1))).NotTo(Succeed())
	})
})

var _ = Describe("buildSnapshotJob", func() {
	var (
		source *corev1.Pod
		job    *snapshotv1alpha1.SnapshotJob
	)

	BeforeEach(func() {
		source = replica("llama-a", "rs-1", 100, true, func(pod *corev1.Pod) {
			pod.Labels["team"] = "a"
			pod.Labels["serving.knative.dev/revision"] = "llama-00001"
			pod.Labels[commonconstants.TopOwnerWorkloadIdLabel] = "workload-1"
			pod.Labels["run.ai/element-1-id"] = "element-1"
			pod.Labels["run.ai/element-2-id"] = "element-2"
			pod.Labels["project"] = "proj"
			pod.Spec.PriorityClassName = "inference"
			pod.Spec.Priority = ptr.To[int32](100)
			pod.Spec.Containers[0].Lifecycle = &corev1.Lifecycle{
				PreStop:   &corev1.LifecycleHandler{Exec: &corev1.ExecAction{Command: []string{"drain"}}},
				PostStart: &corev1.LifecycleHandler{Exec: &corev1.ExecAction{Command: []string{"warm"}}},
			}
			pod.Spec.Containers = append(pod.Spec.Containers, corev1.Container{Name: "queue-proxy", Image: "queue"})
			pod.Spec.Affinity = &corev1.Affinity{NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{NodeSelectorTerms: []corev1.NodeSelectorTerm{{
					MatchExpressions: []corev1.NodeSelectorRequirement{{
						Key: "runai/node-pool", Operator: corev1.NodeSelectorOpIn, Values: []string{"pool-a"},
					}},
				}}},
			}}
		})

		var err error
		job, err = buildSnapshotJob(testRunner(), source, captureInput{
			key:               "k1",
			agentNodeSelector: agentSelector,
			serviceSelectors:  []map[string]string{{"app": "llama"}, {"app": "other"}, {}},
			timeoutSeconds:    ptr.To[int64](7200),
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("names the job after the runner and the key, in the replica's namespace", func() {
		Expect(job.Name).To(Equal(snapshotkey.SnapshotName("k1", "llama")))
		Expect(job.Namespace).To(Equal("runai-proj"))
	})

	// The cleanup finds them by these labels; the runner owns neither.
	It("labels the job and its PodSnapshot with the runner and key, with no owner", func() {
		identity := map[string]string{
			constants.SnapshotKeyLabel:              "k1",
			commonconstants.WorkloadRunnerNameLabel: "llama",
			commonconstants.WorkloadRunnerUidLabel:  "runner-uid",
		}
		Expect(job.Labels).To(Equal(identity))
		Expect(job.Spec.PodSnapshotTemplate.Metadata.Labels).To(Equal(identity))
		Expect(job.OwnerReferences).To(BeEmpty())
	})

	It("captures the target container within the timeout", func() {
		Expect(job.Spec.PodSnapshotTemplate.TargetContainers).To(Equal([]string{"main"}))
		Expect(job.Spec.ActiveDeadlineSeconds).To(HaveValue(Equal(int64(7200))))
	})

	It("runs only the target container, without its preStop hook", func() {
		containers := job.Spec.PodTemplate.Spec.Containers
		Expect(containers).To(HaveLen(1))
		Expect(containers[0].Name).To(Equal("main"))
		Expect(containers[0].Lifecycle.PreStop).To(BeNil())
		Expect(containers[0].Lifecycle.PostStart).NotTo(BeNil())
	})

	It("drops what scheduling and admission set for the replica", func() {
		spec := job.Spec.PodTemplate.Spec
		Expect(spec.NodeName).To(BeEmpty())
		Expect(spec.PriorityClassName).To(BeEmpty())
		Expect(spec.Priority).To(BeNil())
		Expect(spec.RestartPolicy).To(Equal(corev1.RestartPolicyNever))
	})

	// ANDed into the nodepool's term: a term of its own would be ORed, and let the
	// capture run in a nodepool its replicas can never restore in.
	It("requires an agent node within the replica's nodepool", func() {
		terms := job.Spec.PodTemplate.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
		Expect(terms).To(HaveLen(1))
		var keys []string
		for _, expression := range terms[0].MatchExpressions {
			keys = append(keys, expression.Key)
		}
		Expect(keys).To(Equal([]string{"runai/node-pool", "nvidia.com/snapshot-node"}))
	})

	It("keeps only the pod labels that don't send traffic or count it as the serving workload's", func() {
		Expect(job.Spec.PodTemplate.Labels).To(Equal(map[string]string{"team": "a", "project": "proj", captureLabel: "true"}))
	})

	It("leaves the replica unchanged", func() {
		Expect(source.Spec.Containers).To(HaveLen(2))
		Expect(source.Spec.NodeName).NotTo(BeEmpty())
	})
})

func testRunner() *dromosv1alpha1.WorkloadRunner {
	return &dromosv1alpha1.WorkloadRunner{ObjectMeta: metav1.ObjectMeta{
		Name: "llama", Namespace: "runai-proj", UID: "runner-uid",
	}}
}

func capturePlugin() *Plugin {
	plugin := newPlugin(kartaStub{}, true)
	plugin.agentNodeSelector = agentSelector
	plugin.events = record.NewFakeRecorder(100)
	return plugin
}

// replica is a keyed replica of the runner's workload. Its ReplicaSet scopes the
// key; mutate runs before the key is computed.
func replica(name, replicaSet string, created int64, ready bool, mutate ...func(*corev1.Pod)) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         "runai-proj",
			CreationTimestamp: metav1.Unix(created, 0),
			Labels: map[string]string{
				constants.SnapshotOptInLabel:            "true",
				commonconstants.WorkloadRunnerNameLabel: "llama",
				commonconstants.WorkloadRunnerUidLabel:  "runner-uid",
				"app":                                   "llama",
			},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "apps/v1", Kind: "ReplicaSet", Name: replicaSet,
				UID: types.UID(replicaSet), Controller: ptr.To(true),
			}},
		},
		Spec: corev1.PodSpec{Containers: []corev1.Container{gpuContainer("main")}, NodeName: "node-1"},
	}
	for _, m := range mutate {
		m(pod)
	}
	key, err := snapshotkey.ComputeKey(pod)
	Expect(err).NotTo(HaveOccurred())
	snapshotkey.SetSnapshotKey(pod, key)
	if ready {
		pod.Status.Conditions = []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}}
	}
	return pod
}

// capturedJob is testRunner's capture of key.
func capturedJob(key string, finished bool) *snapshotv1alpha1.SnapshotJob {
	job := &snapshotv1alpha1.SnapshotJob{ObjectMeta: metav1.ObjectMeta{
		Name:      snapshotkey.SnapshotName(key, "llama"),
		Namespace: "runai-proj",
		Labels: map[string]string{
			constants.SnapshotKeyLabel:              key,
			commonconstants.WorkloadRunnerNameLabel: "llama",
			commonconstants.WorkloadRunnerUidLabel:  "runner-uid",
		},
	}}
	if finished {
		job.Status.Conditions = []metav1.Condition{{
			Type: snapshotv1alpha1.SnapshotJobConditionCompleted, Status: metav1.ConditionTrue,
			Reason: "Captured", LastTransitionTime: metav1.Unix(300, 0),
		}}
	}
	return job
}

func captureClient(objects ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	Expect(corev1.AddToScheme(scheme)).To(Succeed())
	Expect(dromosv1alpha1.AddToScheme(scheme)).To(Succeed())
	Expect(snapshotv1alpha1.AddToScheme(scheme)).To(Succeed())
	// testRunner's namespace, a project's.
	project := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
		Name: "runai-proj", Labels: map[string]string{constants.RunaiQueueLabel: "proj"},
	}}
	return fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(project).WithObjects(objects...).Build()
}

// handle runs the capture hook for testRunner and returns its warnings.
func handle(ctx context.Context, plugin *Plugin, c client.Client, workload *unstructured.Unstructured) []string {
	return handleRunner(ctx, plugin, testRunner(), c, workload)
}

// handleRunner runs the capture hook and returns its warnings, failing on any
// critical one: a capture problem must never change the workload's phase.
func handleRunner(ctx context.Context, plugin *Plugin, runner *dromosv1alpha1.WorkloadRunner, c client.Client,
	workload *unstructured.Unstructured) []string {
	state := reconcile.New()
	Expect(plugin.HandleAdditionalResources(ctx, runner, workload, deploymentKarta(), c, state)).To(Succeed())
	var warnings []string
	for _, reported := range state.PluginIssues(plugin) {
		Expect(reported.IsCritical).To(BeFalse(), "issue %q is critical", reported.Error())
		warnings = append(warnings, reported.Error())
	}
	return warnings
}

func snapshotJobs(ctx context.Context, c client.Client) []snapshotv1alpha1.SnapshotJob {
	list := &snapshotv1alpha1.SnapshotJobList{}
	Expect(c.List(ctx, list)).To(Succeed())
	return list.Items
}

func hasSnapshotJob(ctx context.Context, c client.Client, key string) bool {
	for _, job := range snapshotJobs(ctx, c) {
		if job.Name == snapshotkey.SnapshotName(key, "llama") {
			return true
		}
	}
	return false
}

// named matches testRunner's capture, or its PodSnapshot, of key.
func named(key string) gomegatypes.GomegaMatcher {
	return HaveField("ObjectMeta.Name", snapshotkey.SnapshotName(key, "llama"))
}

func keyOf(pod *corev1.Pod) string {
	key, _ := snapshotkey.SnapshotKeyOf(pod)
	return key
}

// withNewestKey gives runner the status an earlier reconcile recorded for key.
func withNewestKey(runner *dromosv1alpha1.WorkloadRunner, key string) *dromosv1alpha1.WorkloadRunner {
	return withRecord(runner, statusInfo{NewestKey: key})
}

// withRecord gives runner the structured status an earlier reconcile recorded.
func withRecord(runner *dromosv1alpha1.WorkloadRunner, info statusInfo) *dromosv1alpha1.WorkloadRunner {
	raw, err := json.Marshal(info)
	Expect(err).NotTo(HaveOccurred())
	runner.Status.Plugins = map[string]*dromosv1alpha1.PluginStatus{
		PluginName: {AdditionalInfo: &runtime.RawExtension{Raw: raw}},
	}
	return runner
}

func withUnreadableRecord(runner *dromosv1alpha1.WorkloadRunner) *dromosv1alpha1.WorkloadRunner {
	runner.Status.Plugins = map[string]*dromosv1alpha1.PluginStatus{
		PluginName: {AdditionalInfo: &runtime.RawExtension{Raw: []byte(`{"newestKey":"k1","events":"not a map"}`)}},
	}
	return runner
}

// recorded is the plugin's structured status on runner.
func recorded(runner *dromosv1alpha1.WorkloadRunner) statusInfo {
	info, err := recordedStatus(runner)
	Expect(err).NotTo(HaveOccurred())
	return info
}

// emitted drains the events plugin emitted, each as "type reason message".
func emitted(plugin *Plugin) []string {
	recorder, ok := plugin.events.(*record.FakeRecorder)
	Expect(ok).To(BeTrue())
	var events []string
	for {
		select {
		case event := <-recorder.Events:
			events = append(events, event)
		default:
			return events
		}
	}
}

func reasons(events []string) []string {
	var got []string
	for _, event := range events {
		got = append(got, strings.Fields(event)[1])
	}
	return got
}

func jobCaptured(job *snapshotv1alpha1.SnapshotJob) *snapshotv1alpha1.SnapshotJob {
	return withJobCondition(job, snapshotv1alpha1.SnapshotJobConditionCaptured,
		snapshotv1alpha1.ReasonCaptureCompleted, "")
}

func jobFailed(job *snapshotv1alpha1.SnapshotJob) *snapshotv1alpha1.SnapshotJob {
	return withJobCondition(job, snapshotv1alpha1.SnapshotJobConditionFailed,
		snapshotv1alpha1.ReasonDeadlineExceeded, "the capture did not finish in time")
}

func withJobCondition(job *snapshotv1alpha1.SnapshotJob, conditionType, reason,
	message string) *snapshotv1alpha1.SnapshotJob {
	job.Status.Conditions = append(job.Status.Conditions, metav1.Condition{
		Type: conditionType, Status: metav1.ConditionTrue, Reason: reason, Message: message,
		LastTransitionTime: metav1.Unix(300, 0),
	})
	return job
}

func withUID(job *snapshotv1alpha1.SnapshotJob, uid types.UID) *snapshotv1alpha1.SnapshotJob {
	job.UID = uid
	return job
}

func failed(snapshot *snapshotv1alpha1.PodSnapshot) *snapshotv1alpha1.PodSnapshot {
	snapshot.Status.Conditions = []metav1.Condition{{
		Type: snapshotv1alpha1.PodSnapshotConditionFailed, Status: metav1.ConditionTrue,
		Reason: "CaptureFailed", LastTransitionTime: metav1.Unix(300, 0),
	}}
	return snapshot
}
