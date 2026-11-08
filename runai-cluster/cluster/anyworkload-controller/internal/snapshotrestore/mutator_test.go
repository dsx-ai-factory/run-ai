/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshotrestore

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ai-dynamo/snapshot/api/podcontract"
	snapshotv1alpha1 "github.com/ai-dynamo/snapshot/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/internal/snapshotkey"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	jsonpatch "github.com/evanphx/json-patch/v5"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// emptyReader finds no PodSnapshot, which is the cold-start path every test
// here exercises except the restore ones below.
func emptyReader(t *testing.T) client.Reader {
	t.Helper()
	return snapshotReader(t)
}

// snapshotReader serves the given PodSnapshots and nothing else.
func snapshotReader(t *testing.T, objects ...client.Object) client.Reader {
	t.Helper()

	scheme := runtime.NewScheme()
	if err := snapshotv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("register Snapshot types: %v", err)
	}
	return ctrlfake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
}

// keyLabelPath is the key label as a JSON Pointer: a "/" inside a token is "~1".
const keyLabelPath = "/metadata/labels/run.ai~1snapshot-key"

// optedInPod is a keyable replica: opted in, owned by a ReplicaSet, one GPU
// container.
func optedInPod(name string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "runai-proj",
			Labels:    map[string]string{constants.SnapshotOptInLabel: "true"},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "apps/v1",
				Kind:       "ReplicaSet",
				Name:       "llama-00001-deployment-7d4f8b6c9d",
				UID:        "0c3f6b9a-1d2e-4f50-8a7b-6c5d4e3f2a1b",
				Controller: ptr.To(true),
			}},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:  "main",
				Image: "vllm/vllm-openai:v0.31.0",
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{"nvidia.com/gpu": resource.MustParse("1")},
				},
			}},
		},
	}
}

func request(t *testing.T, operation admissionv1.Operation, pod *corev1.Pod) admission.Request {
	t.Helper()

	raw, err := json.Marshal(pod)
	if err != nil {
		t.Fatalf("marshal pod: %v", err)
	}
	return admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			Operation: operation,
			Object:    runtime.RawExtension{Raw: raw},
		},
	}
}

// stampedKey returns the key the handler wrote, and whether it wrote one.
func stampedKey(t *testing.T, pod *corev1.Pod) (string, bool) {
	t.Helper()

	response := NewMutator(emptyReader(t), nil).Handle(context.Background(), request(t, admissionv1.Create, pod))
	if !response.Allowed {
		t.Fatalf("pod not admitted: %v", response.Result)
	}

	for _, patch := range response.Patches {
		if patch.Path != keyLabelPath {
			continue
		}
		value, ok := patch.Value.(string)
		if !ok {
			t.Fatalf("key label patched with %T, want string", patch.Value)
		}
		return value, true
	}
	return "", false
}

// The key label is the handoff to the capture plugin, which finds each key's
// first Ready replica by it. Without it on the cold replica nothing is ever
// captured.
func TestMutatorStampsTheKey(t *testing.T) {
	pod := optedInPod("replica-0")

	key, stamped := stampedKey(t, pod)
	if !stamped {
		t.Fatal("no key stamped on a keyable pod")
	}

	want, err := snapshotkey.ComputeKey(pod)
	if err != nil {
		t.Fatalf("ComputeKey: %v", err)
	}
	if key != want {
		t.Errorf("stamped %q, want %q", key, want)
	}
}

// Acceptance: replicas of one version of a workload all carry the same key, so
// one capture serves every replica.
func TestReplicasOfOneVersionGetTheSameKey(t *testing.T) {
	first, _ := stampedKey(t, optedInPod("replica-0"))

	other := optedInPod("replica-1")
	other.Spec.NodeName = "gpu-node-7"
	second, _ := stampedKey(t, other)

	if first == "" || first != second {
		t.Errorf("replicas got %q and %q, want one key", first, second)
	}
}

// Acceptance: an edit that changes the pods produces a different key, so a new
// version never restores from the old version's checkpoint.
func TestAnEditProducesADifferentKey(t *testing.T) {
	before, _ := stampedKey(t, optedInPod("replica-0"))

	edited := optedInPod("replica-0")
	edited.Spec.Containers[0].Image = "vllm/vllm-openai:v0.32.0"
	after, _ := stampedKey(t, edited)

	if before == after {
		t.Errorf("an image change left the key at %q", before)
	}
}

// Acceptance: scaling does not change the key. Replica count lives on the
// workload, not on the pod, so a replica added by scaling up is the same shape.
func TestScalingDoesNotChangeTheKey(t *testing.T) {
	first, _ := stampedKey(t, optedInPod("replica-0"))
	scaledUp, _ := stampedKey(t, optedInPod("replica-4"))

	if first != scaledUp {
		t.Errorf("scaling changed the key: %q then %q", first, scaledUp)
	}
}

// Every failure admits the pod unchanged: not labelling costs a cold start,
// rejecting would stop the workload running.
func TestMutatorAdmitsUnchangedOnFailure(t *testing.T) {
	notKeyable := optedInPod("replica-0")
	notKeyable.OwnerReferences = nil

	noGPU := optedInPod("replica-0")
	noGPU.Spec.Containers[0].Resources = corev1.ResourceRequirements{}

	tests := map[string]admission.Request{
		"no controller owner": request(t, admissionv1.Create, notKeyable),
		"no gpu container":    request(t, admissionv1.Create, noGPU),
		// CREATE only is registered; on anything else, still admit.
		"update": request(t, admissionv1.Update, optedInPod("replica-0")),
		"delete": request(t, admissionv1.Delete, optedInPod("replica-0")),
		"unreadable": {AdmissionRequest: admissionv1.AdmissionRequest{
			Operation: admissionv1.Create,
			Object:    runtime.RawExtension{Raw: []byte("{not json")},
		}},
	}

	for name, req := range tests {
		t.Run(name, func(t *testing.T) {
			response := NewMutator(emptyReader(t), nil).Handle(context.Background(), req)

			if !response.Allowed {
				t.Fatalf("pod not admitted: %v", response.Result)
			}
			if len(response.Patches) != 0 {
				t.Errorf("produced %d patches, want none: %v", len(response.Patches), response.Patches)
			}
		})
	}
}

// The go-operator writes this path into the webhook configuration and this
// package serves it, so the shared constant keeps them together. The literal is
// pinned as well because it is cluster-visible: changing it leaves every
// already-applied configuration pointing at a path nothing answers until
// init-ca reapplies them.
func TestRegisteredPathIsTheConfiguredOne(t *testing.T) {
	if constants.SnapshotRestoreWebhookPath != "/mutate-pod-snapshot-restore" {
		t.Fatalf("SnapshotRestoreWebhookPath = %q, want /mutate-pod-snapshot-restore",
			constants.SnapshotRestoreWebhookPath)
	}
}

// readySnapshot is a PodSnapshot for the key the given pod computes.
func readySnapshot(t *testing.T, pod *corev1.Pod, conditions ...metav1.Condition) *snapshotv1alpha1.PodSnapshot {
	t.Helper()

	key, err := snapshotkey.ComputeKey(pod)
	if err != nil {
		t.Fatalf("ComputeKey: %v", err)
	}
	return &snapshotv1alpha1.PodSnapshot{
		ObjectMeta: metav1.ObjectMeta{Name: snapshotkey.SnapshotName(key), Namespace: pod.Namespace},
		Status:     snapshotv1alpha1.PodSnapshotStatus{Conditions: conditions},
	}
}

func ready() metav1.Condition {
	return metav1.Condition{
		Type: snapshotv1alpha1.PodSnapshotConditionReady, Status: metav1.ConditionTrue,
		Reason: "Captured", LastTransitionTime: metav1.Now(),
	}
}

func failed() metav1.Condition {
	return metav1.Condition{
		Type: snapshotv1alpha1.PodSnapshotConditionFailed, Status: metav1.ConditionTrue,
		Reason: "CheckpointUnusable", LastTransitionTime: metav1.Now(),
	}
}

// restoredPod runs the handler and returns the pod as it was admitted.
func restoredPod(t *testing.T, pod *corev1.Pod, reader client.Reader,
	agentNodeSelector map[string]string) *corev1.Pod {
	t.Helper()

	raw, err := json.Marshal(pod)
	if err != nil {
		t.Fatalf("marshal pod: %v", err)
	}
	response := NewMutator(reader, agentNodeSelector).Handle(context.Background(),
		admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{
			Operation: admissionv1.Create,
			Object:    runtime.RawExtension{Raw: raw},
		}})
	if !response.Allowed {
		t.Fatalf("pod not admitted: %v", response.Result)
	}

	patch, err := jsonpatch.DecodePatch(mustMarshal(t, response.Patches))
	if err != nil {
		t.Fatalf("decode patch: %v", err)
	}
	patched, err := patch.Apply(raw)
	if err != nil {
		t.Fatalf("apply patch: %v", err)
	}

	admitted := &corev1.Pod{}
	if err := json.Unmarshal(patched, admitted); err != nil {
		t.Fatalf("unmarshal admitted pod: %v", err)
	}
	return admitted
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return raw
}

// Acceptance: with a ready snapshot, the replica is turned into a restore pod.
func TestReadySnapshotProducesARestorePod(t *testing.T) {
	pod := optedInPod("replica-0")
	reader := snapshotReader(t, readySnapshot(t, pod, ready()))

	admitted := restoredPod(t, pod, reader, map[string]string{"nvidia.com/gpu.present": "true"})

	if _, ok := admitted.Annotations[podcontract.RestoreFromAnnotation]; !ok {
		t.Errorf("no %s annotation; the pod was not built as a restore pod", podcontract.RestoreFromAnnotation)
	}
	if admitted.Labels[constants.SnapshotKeyLabel] == "" {
		t.Error("the key label was not stamped")
	}
}

// Acceptance: with no ready snapshot the replica cold-starts, carrying the key
// and nothing else.
func TestWithoutAReadySnapshotThePodOnlyGetsTheKey(t *testing.T) {
	for name, conditions := range map[string][]metav1.Condition{
		"no snapshot at all": nil,
		"not ready":          {},
		"marked failed":      {ready(), failed()},
	} {
		t.Run(name, func(t *testing.T) {
			pod := optedInPod("replica-0")
			reader := emptyReader(t)
			if conditions != nil {
				reader = snapshotReader(t, readySnapshot(t, pod, conditions...))
			}

			admitted := restoredPod(t, pod, reader, nil)

			if _, ok := admitted.Annotations[podcontract.RestoreFromAnnotation]; ok {
				t.Error("pod was restored from a snapshot that is not usable")
			}
			if admitted.Labels[constants.SnapshotKeyLabel] == "" {
				t.Error("the key label was not stamped")
			}
		})
	}
}

// The agent's selector must AND into the term runai-pod already added, not OR
// against it: nodeSelectorTerms are ORed, so a new term would let the replica
// land on a node with no agent.
func TestAgentSelectorIsMergedIntoExistingTerms(t *testing.T) {
	pod := optedInPod("replica-0")
	pod.Spec.Affinity = &corev1.Affinity{NodeAffinity: &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{{
				MatchExpressions: []corev1.NodeSelectorRequirement{{
					Key: "runai/node-pool", Operator: corev1.NodeSelectorOpIn, Values: []string{"default"},
				}},
			}},
		},
	}}

	admitted := restoredPod(t, pod, snapshotReader(t, readySnapshot(t, pod, ready())),
		map[string]string{"nvidia.com/gpu.present": "true", "kubernetes.io/hostname": "gpu-0"})

	terms := admitted.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	if len(terms) != 1 {
		t.Fatalf("got %d terms, want 1 - a second term would OR against the workload's placement", len(terms))
	}

	keys := map[string]bool{}
	for _, expression := range terms[0].MatchExpressions {
		keys[expression.Key] = true
	}
	for _, want := range []string{"runai/node-pool", "nvidia.com/gpu.present", "kubernetes.io/hostname"} {
		if !keys[want] {
			t.Errorf("term is missing %q; all three must be ANDed", want)
		}
	}
}

// podWithAffinity returns a keyable pod constrained by one node affinity term.
func podWithAffinity(expressions ...corev1.NodeSelectorRequirement) *corev1.Pod {
	pod := optedInPod("replica-0")
	pod.Spec.Affinity = &corev1.Affinity{NodeAffinity: &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{{MatchExpressions: expressions}},
		},
	}}
	return pod
}

func in(key string, values ...string) corev1.NodeSelectorRequirement {
	return corev1.NodeSelectorRequirement{Key: key, Operator: corev1.NodeSelectorOpIn, Values: values}
}

// Two In requirements on one key intersect, which is what both constraints
// together mean - the workload's expression is kept, not overridden.
func TestCompatiblePlacementIntersects(t *testing.T) {
	pod := podWithAffinity(in("kubernetes.io/hostname", "gpu-0", "gpu-1"))

	admitted := restoredPod(t, pod, snapshotReader(t, readySnapshot(t, pod, ready())),
		map[string]string{"kubernetes.io/hostname": "gpu-0"})

	terms := admitted.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	if len(terms) != 1 {
		t.Fatalf("got %d terms, want 1", len(terms))
	}

	var seen []string
	for _, expression := range terms[0].MatchExpressions {
		if expression.Key == "kubernetes.io/hostname" {
			seen = append(seen, strings.Join(expression.Values, ","))
		}
	}
	if len(seen) != 2 {
		t.Fatalf("hostname expressions = %v, want the workload's kept and ours added", seen)
	}
}

// A placement that excludes every agent node must cold-start. Admitting the
// restore pod would strand the replica in Pending, which is worse than not
// restoring at all.
func TestConflictingPlacementColdStarts(t *testing.T) {
	agent := map[string]string{"kubernetes.io/hostname": "gpu-0"}

	cases := map[string]*corev1.Pod{
		"affinity names other nodes": podWithAffinity(in("kubernetes.io/hostname", "gpu-7")),
		"affinity excludes the agent node": podWithAffinity(corev1.NodeSelectorRequirement{
			Key: "kubernetes.io/hostname", Operator: corev1.NodeSelectorOpNotIn, Values: []string{"gpu-0"},
		}),
		"nodeSelector pins another node": func() *corev1.Pod {
			pod := optedInPod("replica-0")
			pod.Spec.NodeSelector = map[string]string{"kubernetes.io/hostname": "gpu-7"}
			return pod
		}(),
	}

	for name, pod := range cases {
		t.Run(name, func(t *testing.T) {
			admitted := restoredPod(t, pod, snapshotReader(t, readySnapshot(t, pod, ready())), agent)

			if _, ok := admitted.Annotations[podcontract.RestoreFromAnnotation]; ok {
				t.Error("restored a pod that can never be scheduled")
			}
			if admitted.Labels[constants.SnapshotKeyLabel] == "" {
				t.Error("the key label was not stamped")
			}
		})
	}
}

// A conflicting term makes the pod cold-start, and its affinity is left
// exactly as the workload wrote it - no term is dropped or rewritten.
func TestConflictingTermLeavesAffinityUntouched(t *testing.T) {
	pod := optedInPod("replica-0")
	pod.Spec.Affinity = &corev1.Affinity{NodeAffinity: &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{MatchExpressions: []corev1.NodeSelectorRequirement{in("kubernetes.io/hostname", "gpu-7")}},
				{MatchExpressions: []corev1.NodeSelectorRequirement{in("runai/node-pool", "default")}},
			},
		},
	}}

	admitted := restoredPod(t, pod, snapshotReader(t, readySnapshot(t, pod, ready())),
		map[string]string{"kubernetes.io/hostname": "gpu-0"})

	if _, ok := admitted.Annotations[podcontract.RestoreFromAnnotation]; ok {
		t.Error("restored a pod whose placement excludes the agent's node")
	}

	terms := admitted.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	if len(terms) != 2 {
		t.Fatalf("got %d terms, want both of the workload's kept", len(terms))
	}
	for _, term := range terms {
		if len(term.MatchExpressions) != 1 {
			t.Errorf("term was rewritten: %v", term.MatchExpressions)
		}
	}
}

// An empty agent selector means the operator could not read the agent
// DaemonSet, so no node is known to be able to restore. Attempting one would
// waste a restore that fails on the node and falls back anyway.
func TestNoKnownAgentNodesColdStarts(t *testing.T) {
	pod := optedInPod("replica-0")

	admitted := restoredPod(t, pod, snapshotReader(t, readySnapshot(t, pod, ready())), nil)

	if _, ok := admitted.Annotations[podcontract.RestoreFromAnnotation]; ok {
		t.Error("restored with no node known to run the agent")
	}
	if admitted.Labels[constants.SnapshotKeyLabel] == "" {
		t.Error("the key label was not stamped")
	}
}

// podcontract replaces the target's startup probe with the restore gate, so a
// workload that defines its own would silently lose it on restored replicas.
// Such a pod cold-starts instead.
func TestOwnStartupProbeColdStarts(t *testing.T) {
	pod := optedInPod("replica-0")
	pod.Spec.Containers[0].StartupProbe = &corev1.Probe{
		ProbeHandler:     corev1.ProbeHandler{Exec: &corev1.ExecAction{Command: []string{"check"}}},
		FailureThreshold: 30,
	}

	admitted := restoredPod(t, pod, snapshotReader(t, readySnapshot(t, pod, ready())),
		map[string]string{"nvidia.com/gpu.present": "true"})

	if _, ok := admitted.Annotations[podcontract.RestoreFromAnnotation]; ok {
		t.Error("restored a pod whose own startup probe would be replaced")
	}
	if admitted.Spec.Containers[0].StartupProbe.FailureThreshold != 30 {
		t.Error("the workload's startup probe was altered")
	}
	if admitted.Labels[constants.SnapshotKeyLabel] == "" {
		t.Error("the key label was not stamped")
	}
}
