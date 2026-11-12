/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshots

import (
	"context"

	snapshotv1alpha1 "github.com/ai-dynamo/snapshot/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/internal/snapshotkey"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
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
	return fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
}

// handle runs the capture hook and returns its warnings, failing on any critical
// one: a capture problem must never change the workload's phase.
func handle(ctx context.Context, plugin *Plugin, c client.Client, workload *unstructured.Unstructured) []string {
	state := reconcile.New()
	Expect(plugin.HandleAdditionalResources(ctx, testRunner(), workload, deploymentKarta(), c, state)).To(Succeed())
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

func keyOf(pod *corev1.Pod) string {
	key, _ := snapshotkey.SnapshotKeyOf(pod)
	return key
}
