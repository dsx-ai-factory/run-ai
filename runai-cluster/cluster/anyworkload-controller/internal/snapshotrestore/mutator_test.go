/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshotrestore

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func optedInPod(t *testing.T) []byte {
	t.Helper()

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "replica-0",
			Namespace: "runai-proj",
			Labels:    map[string]string{constants.SnapshotOptInLabel: "true"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Name: "main", Image: "vllm/vllm-openai:v0.31.0"}},
		},
	}
	raw, err := json.Marshal(pod)
	if err != nil {
		t.Fatalf("marshal pod: %v", err)
	}
	return raw
}

func request(t *testing.T, operation admissionv1.Operation) admission.Request {
	t.Helper()

	return admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			Operation: operation,
			Object:    runtime.RawExtension{Raw: optedInPod(t)},
		},
	}
}

// The handler must stay inert until the keying story lands: a patch now would
// change the pod shape that later stages hash.
func TestMutatorAdmitsPodsUnchanged(t *testing.T) {
	tests := []struct {
		name      string
		operation admissionv1.Operation
	}{
		{name: "create", operation: admissionv1.Create},
		// CREATE only is registered; on anything else, still admit.
		{name: "update", operation: admissionv1.Update},
		{name: "delete", operation: admissionv1.Delete},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			response := NewMutator().Handle(context.Background(), request(t, tc.operation))

			if !response.Allowed {
				t.Fatalf("pod not admitted: %v", response.Result)
			}
			if len(response.Patches) != 0 {
				t.Errorf("handler produced %d patches, want none: %v", len(response.Patches), response.Patches)
			}
			if response.PatchType != nil {
				t.Errorf("patchType = %v, want nil", *response.PatchType)
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
