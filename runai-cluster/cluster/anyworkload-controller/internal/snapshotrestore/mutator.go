/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshotrestore

import (
	"context"

	admissionv1 "k8s.io/api/admission/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// Mutator admits opted-in pods. A pass-through for now: it ships with the
// webhook so a live configuration never points at a path nothing serves. Next
// come stamping the snapshot key, then rewriting the pod into a restore pod.
type Mutator struct{}

// NewMutator returns the restore webhook handler.
func NewMutator() *Mutator {
	return &Mutator{}
}

// Handle admits the pod unchanged. It returns no patch: the snapshot key hashes
// the pod's shape, so anything written here would change what later stages hash.
func (m *Mutator) Handle(_ context.Context, request admission.Request) admission.Response {
	if request.Operation != admissionv1.Create {
		// The configuration registers CREATE only; anything else means it and
		// this handler have drifted apart.
		return admission.Allowed("unexpected operation, admitted unchanged")
	}

	return admission.Allowed("")
}
