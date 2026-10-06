/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (pv *ProjectValidator) allowOrphanFinalizerRemoval(ctx context.Context, req admission.Request) (bool, error) {
	if req.UserInfo.Username != "system:serviceaccount:kube-system:generic-garbage-collector" ||
		req.UserInfo.UID == "" || req.Operation != admissionv1.Update || req.SubResource != "" ||
		req.Resource.Resource != "projects" || (req.Resource.Group != "run.ai" && req.Resource.Group != "kai.resources") {
		return false, nil
	}

	// Keep unknown fields in the comparison so this exception cannot authorize changes
	// to Project fields introduced by a newer CRD version.
	oldProject, newProject := &unstructured.Unstructured{}, &unstructured.Unstructured{}
	if err := json.Unmarshal(req.OldObject.Raw, oldProject); err != nil {
		return false, fmt.Errorf("decode previous Project for orphan cleanup: %w", err)
	}
	if err := json.Unmarshal(req.Object.Raw, newProject); err != nil {
		return false, fmt.Errorf("decode updated Project for orphan cleanup: %w", err)
	}
	if oldProject.GetDeletionTimestamp() == nil ||
		!slices.Contains(oldProject.GetFinalizers(), metav1.FinalizerOrphanDependents) {
		return false, nil
	}
	remaining := slices.DeleteFunc(oldProject.GetFinalizers(), func(finalizer string) bool {
		return finalizer == metav1.FinalizerOrphanDependents
	})
	if !slices.Equal(remaining, newProject.GetFinalizers()) {
		return false, nil
	}
	oldProject.SetFinalizers(nil)
	newProject.SetFinalizers(nil)
	// The API server updates field ownership while applying the collector's patch.
	oldProject.SetManagedFields(nil)
	newProject.SetManagedFields(nil)
	if !reflect.DeepEqual(oldProject.Object, newProject.Object) {
		return false, nil
	}

	account, err := GetServiceAccount(ctx, pv.garbageCollectorReader, "generic-garbage-collector", metav1.NamespaceSystem)
	if err != nil {
		return false, fmt.Errorf("get garbage collector service account: %w", err)
	}
	return req.UserInfo.UID == string(account.UID), nil
}
