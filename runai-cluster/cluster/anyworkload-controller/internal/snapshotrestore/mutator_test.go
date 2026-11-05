/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshotrestore

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/internal/snapshotkey"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

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

	response := NewMutator().Handle(context.Background(), request(t, admissionv1.Create, pod))
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
			response := NewMutator().Handle(context.Background(), req)

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
