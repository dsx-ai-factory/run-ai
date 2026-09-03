/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package ownership

import (
	"context"
	"fmt"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// UpdateMutator restores the stored owner when an update changes it.
type UpdateMutator struct{}

// NewUpdateMutator returns an update-only ownership mutator.
func NewUpdateMutator() *UpdateMutator {
	return &UpdateMutator{}
}

func (m *UpdateMutator) Handle(_ context.Context, request admission.Request) admission.Response {
	if request.Operation != admissionv1.Update {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("unsupported operation %q", request.Operation))
	}
	if request.Kind.Kind == "WorkloadRunner" {
		return mutateWorkloadRunnerOwner(request)
	}

	workload, err := decodeLegacyWorkload(request.Object.Raw)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if workload.usage() != submitUsage {
		return admission.Allowed("")
	}

	oldWorkload, err := decodeLegacyWorkload(request.OldObject.Raw)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if workload.ownerUID() == oldWorkload.ownerUID() {
		return admission.Allowed("")
	}
	if err := workload.restoreUserID(oldWorkload); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	mutatedRaw, err := workload.marshal()
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(request.Object.Raw, mutatedRaw)
}

func mutateWorkloadRunnerOwner(request admission.Request) admission.Response {
	desired, err := decodeWorkloadRunner(request.Object.Raw)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	stored, err := decodeWorkloadRunner(request.OldObject.Raw)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if stored.ownerUID() == "" || desired.ownerUID() == stored.ownerUID() {
		return admission.Allowed("")
	}
	desired.restoreOwner(stored)
	mutatedRaw, err := desired.marshal()
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(request.Object.Raw, mutatedRaw)
}
