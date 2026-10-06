/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dra

import (
	"fmt"
	"reflect"

	resourcev1 "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/runtime"
	resourcev1defaults "k8s.io/kubernetes/pkg/apis/resource/v1"
)

// rctDefaultScheme carries the API-server defaulters for
// resource.k8s.io/v1 ResourceClaimTemplate — the same ones the
// kube-apiserver runs during storage (AllocationMode→"ExactCount",
// Count→1, Toleration.Operator→"Equal", ...). We apply them to both sides
// before comparing so that user-sent zero values on defaulted fields do
// not look like a diff against the stored object.
//
// scheme.Default is keyed by reflect.TypeOf, so only RegisterDefaults is
// needed; the type scheme itself does not need to be populated.
var rctDefaultScheme = runtime.NewScheme()

func init() {
	if err := resourcev1defaults.RegisterDefaults(rctDefaultScheme); err != nil {
		panic(fmt.Sprintf("register resource/v1 defaulters: %v", err))
	}
}

// RCTSpecDiffersAfterDefaults reports whether two ResourceClaimTemplate
// specs are equal or not, taking into account k8s defauults.
// This is needed for immutable specs (such as resource-claim-templates).
// For those, callers use this to decide whether a reconcile needs a delete-and-recreate
// (true) or an in-place metadata update is enough (false).
func RCTSpecDiffersAfterDefaults(expected, existing *resourcev1.ResourceClaimTemplate) bool {
	if expected == nil || existing == nil {
		return expected != existing
	}

	// Both inputs are DeepCopied before defaulting, so neither is mutated.
	// A nil input compares equal to nil and unequal to any non-nil input.
	expectedCopy := expected.DeepCopy()
	existingCopy := existing.DeepCopy()

	// Apply k8s defaults, this makes the comparison more accurate.
	rctDefaultScheme.Default(expectedCopy)
	rctDefaultScheme.Default(existingCopy)

	return !reflect.DeepEqual(expectedCopy.Spec, existingCopy.Spec)
}
