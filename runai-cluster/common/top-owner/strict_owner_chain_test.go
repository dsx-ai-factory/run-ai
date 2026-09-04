/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"
	"errors"
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var strictExpectedOwner = StrictOwner{
	GroupKind: schema.GroupKind{Group: "run.ai", Kind: "WorkloadRunner"},
	Namespace: "project",
	Name:      "runner",
	UID:       "runner-uid",
}

func TestVerifyStrictOwnerChain(t *testing.T) {
	t.Parallel()

	runnerRef := strictOwnerRef("run.ai/v1alpha1", "WorkloadRunner", "runner", "runner-uid", false)
	jobRef := strictOwnerRef("batch/v1", "Job", "job", "job-uid", true)
	replicaSetRef := strictOwnerRef("apps/v1", "ReplicaSet", "replica-set", "replica-set-uid", true)

	tests := []struct {
		name    string
		start   *metav1.PartialObjectMetadata
		objects []*metav1.PartialObjectMetadata
		valid   bool
	}{
		{
			name:  "direct WorkloadRunner owner",
			start: strictObject("v1", "Pod", "pod", "pod-uid", runnerRef),
			valid: true,
		},
		{
			name:  "live intermediate controller owner",
			start: strictObject("v1", "Pod", "pod", "pod-uid", jobRef),
			objects: []*metav1.PartialObjectMetadata{
				strictObject("batch/v1", "Job", "job", "job-uid", runnerRef),
			},
			valid: true,
		},
		{
			name: "controller owner takes precedence over metadata owner",
			start: strictObject(
				"v1", "Pod", "pod", "pod-uid",
				strictOwnerRef("v1", "ConfigMap", "metadata", "metadata-uid", false),
				jobRef,
			),
			objects: []*metav1.PartialObjectMetadata{
				strictObject("batch/v1", "Job", "job", "job-uid", runnerRef),
			},
			valid: true,
		},
		{
			name: "ambiguous non-controller owners",
			start: strictObject(
				"v1", "Pod", "pod", "pod-uid",
				strictOwnerRef("v1", "ConfigMap", "first", "first-uid", false),
				strictOwnerRef("v1", "Secret", "second", "second-uid", false),
			),
		},
		{
			name: "multiple controller owners",
			start: strictObject(
				"v1", "Pod", "pod", "pod-uid",
				jobRef,
				strictOwnerRef("apps/v1", "ReplicaSet", "other", "other-uid", true),
			),
		},
		{
			name:  "stale intermediate UID",
			start: strictObject("v1", "Pod", "pod", "pod-uid", jobRef),
			objects: []*metav1.PartialObjectMetadata{
				strictObject("batch/v1", "Job", "job", "replacement-job-uid", runnerRef),
			},
		},
		{
			name: "stale WorkloadRunner UID",
			start: strictObject(
				"v1", "Pod", "pod", "pod-uid",
				strictOwnerRef("run.ai/v1alpha1", "WorkloadRunner", "runner", "stale-runner-uid", false),
			),
		},
		{
			name:  "ownership cycle",
			start: strictObject("v1", "Pod", "pod", "pod-uid", jobRef),
			objects: []*metav1.PartialObjectMetadata{
				strictObject("batch/v1", "Job", "job", "job-uid", replicaSetRef),
				strictObject("apps/v1", "ReplicaSet", "replica-set", "replica-set-uid", jobRef),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := VerifyStrictOwnerChain(
				context.Background(), tt.start, strictExpectedOwner, newStrictOwnerReader(tt.objects...),
			)
			if tt.valid && err != nil {
				t.Fatalf("VerifyStrictOwnerChain() error = %v", err)
			}
			if !tt.valid && !errors.Is(err, ErrInvalidOwnerChain) {
				t.Fatalf("VerifyStrictOwnerChain() error = %v, want ErrInvalidOwnerChain", err)
			}
		})
	}
}

func TestVerifyStrictOwnerChainPreservesReadFailure(t *testing.T) {
	t.Parallel()

	readErr := errors.New("API unavailable")
	reader := newStrictOwnerReader()
	reader.err = readErr
	start := strictObject(
		"v1", "Pod", "pod", "pod-uid",
		strictOwnerRef("batch/v1", "Job", "job", "job-uid", true),
	)

	err := VerifyStrictOwnerChain(context.Background(), start, strictExpectedOwner, reader)
	if !errors.Is(err, readErr) {
		t.Fatalf("VerifyStrictOwnerChain() error = %v, want wrapped %v", err, readErr)
	}
	if errors.Is(err, ErrInvalidOwnerChain) {
		t.Fatalf("read failure must not be classified as ErrInvalidOwnerChain: %v", err)
	}
}

func TestVerifyStrictOwnerChainDetectsCycleWhenStartGVKIsEmpty(t *testing.T) {
	t.Parallel()

	start := strictObject(
		"v1", "Pod", "pod", "pod-uid",
		strictOwnerRef("batch/v1", "Job", "job", "job-uid", true),
	)
	start.SetGroupVersionKind(schema.GroupVersionKind{})
	job := strictObject(
		"batch/v1", "Job", "job", "job-uid",
		strictOwnerRef("v1", "Pod", "pod", "pod-uid", true),
	)

	err := VerifyStrictOwnerChain(
		context.Background(), start, strictExpectedOwner, newStrictOwnerReader(job),
	)
	if !errors.Is(err, ErrInvalidOwnerChain) {
		t.Fatalf("VerifyStrictOwnerChain() error = %v, want cycle ErrInvalidOwnerChain", err)
	}
}

func TestVerifyStrictOwnerChainBoundsTraversal(t *testing.T) {
	t.Parallel()

	objects := make([]*metav1.PartialObjectMetadata, 0, strictOwnerChainMaxDepth)
	for i := range strictOwnerChainMaxDepth {
		name := fmt.Sprintf("owner-%d", i)
		uid := types.UID(fmt.Sprintf("owner-%d-uid", i))
		var next metav1.OwnerReference
		if i == strictOwnerChainMaxDepth-1 {
			next = strictOwnerRef("run.ai/v1alpha1", "WorkloadRunner", "runner", "runner-uid", false)
		} else {
			nextName := fmt.Sprintf("owner-%d", i+1)
			next = strictOwnerRef("v1", "ConfigMap", nextName, types.UID(nextName+"-uid"), true)
		}
		objects = append(objects, strictObject("v1", "ConfigMap", name, uid, next))
	}
	start := strictObject(
		"v1", "Pod", "pod", "pod-uid",
		strictOwnerRef("v1", "ConfigMap", "owner-0", "owner-0-uid", true),
	)

	err := VerifyStrictOwnerChain(
		context.Background(), start, strictExpectedOwner, newStrictOwnerReader(objects...),
	)
	if !errors.Is(err, ErrInvalidOwnerChain) {
		t.Fatalf("VerifyStrictOwnerChain() error = %v, want bounded ErrInvalidOwnerChain", err)
	}
}

func strictObject(
	apiVersion, kind, name string,
	uid types.UID,
	ownerRefs ...metav1.OwnerReference,
) *metav1.PartialObjectMetadata {
	object := &metav1.PartialObjectMetadata{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       strictExpectedOwner.Namespace,
			Name:            name,
			UID:             uid,
			OwnerReferences: ownerRefs,
		},
	}
	object.SetGroupVersionKind(schema.FromAPIVersionAndKind(apiVersion, kind))
	return object
}

func strictOwnerRef(
	apiVersion, kind, name string,
	uid types.UID,
	controller bool,
) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: apiVersion,
		Kind:       kind,
		Name:       name,
		UID:        uid,
		Controller: &controller,
	}
}

type strictOwnerReader struct {
	objects map[strictOwnerKey]*metav1.PartialObjectMetadata
	err     error
}

type strictOwnerKey struct {
	key client.ObjectKey
	gvk schema.GroupVersionKind
}

func newStrictOwnerReader(objects ...*metav1.PartialObjectMetadata) *strictOwnerReader {
	reader := &strictOwnerReader{objects: make(map[strictOwnerKey]*metav1.PartialObjectMetadata, len(objects))}
	for _, object := range objects {
		reader.objects[strictOwnerKey{
			key: client.ObjectKeyFromObject(object),
			gvk: object.GroupVersionKind(),
		}] = object.DeepCopy()
	}
	return reader
}

func (r *strictOwnerReader) Get(
	_ context.Context,
	key client.ObjectKey,
	object client.Object,
	_ ...client.GetOption,
) error {
	if r.err != nil {
		return r.err
	}
	stored, found := r.objects[strictOwnerKey{key: key, gvk: object.GetObjectKind().GroupVersionKind()}]
	if !found {
		return fmt.Errorf("object %s %s not found", object.GetObjectKind().GroupVersionKind(), key)
	}
	metadata, ok := object.(*metav1.PartialObjectMetadata)
	if !ok {
		return fmt.Errorf("unsupported object type %T", object)
	}
	stored.DeepCopyInto(metadata)
	return nil
}

func (r *strictOwnerReader) List(
	context.Context,
	client.ObjectList,
	...client.ListOption,
) error {
	return errors.New("List is not implemented")
}
