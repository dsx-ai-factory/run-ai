/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshotrestore

import (
	"context"
	"encoding/json"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/internal/snapshotkey"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// Mutator stamps the snapshot key on every opted-in pod, the first cold one
// included. The key label is the handoff to the capture plugin, which finds
// each key's first Ready replica by it - without it on the cold replica there
// is nothing to capture from and no snapshot is ever taken.
//
// Turning a replica into a restore pod comes later; this handler only labels.
type Mutator struct{}

// NewMutator returns the restore webhook handler.
func NewMutator() *Mutator {
	return &Mutator{}
}

// Handle stamps the key, or admits the pod unchanged.
//
// Every pod reaching this handler is opted in - the configuration's
// objectSelector matched run.ai/snapshot before the API server called us - so
// the label is not rechecked.
func (m *Mutator) Handle(_ context.Context, request admission.Request) admission.Response {
	if request.Operation != admissionv1.Create {
		// The configuration registers CREATE only; anything else means it and
		// this handler have drifted apart.
		return admission.Allowed("unexpected operation, admitted unchanged")
	}

	pod := &corev1.Pod{}
	if err := json.Unmarshal(request.Object.Raw, pod); err != nil {
		return admission.Allowed("pod could not be read, admitted unchanged")
	}

	// Not keyable is an ordinary outcome, not a failure: a pod with no
	// controller owner, or without exactly one GPU container, cold-starts.
	key, err := snapshotkey.ComputeKey(pod)
	if err != nil {
		return admission.Allowed("pod is not keyable, admitted unchanged")
	}

	snapshotkey.SetSnapshotKey(pod, key)

	mutated, err := json.Marshal(pod)
	if err != nil {
		return admission.Allowed("pod could not be written, admitted unchanged")
	}
	return admission.PatchResponseFromRaw(request.Object.Raw, mutated)
}
