/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	knv1 "knative.dev/serving/pkg/apis/serving/v1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FakeClientWithConflictsErrors wrapper to client.Client with conflictOnUpdate flag
type FakeClientWithConflictsErrors struct {
	client.Client
	createKnativeSubResources bool
	conflictOnUpdate          bool
	unavailableOnUpdate       bool
}

func NewFakeClientWithConflictsErrors(scheme *runtime.Scheme, statusObjects []client.Object, objects ...client.Object) *FakeClientWithConflictsErrors {
	return &FakeClientWithConflictsErrors{Client: InitFakeClientWithObjectWithStatus(scheme, statusObjects, objects...)}
}

func (c *FakeClientWithConflictsErrors) SetUnavailableOnUpdate(unavailableOnUpdate bool) {
	c.unavailableOnUpdate = unavailableOnUpdate
}

func (c *FakeClientWithConflictsErrors) SetConflictOnUpdate(conflictOnUpdate bool) {
	c.conflictOnUpdate = conflictOnUpdate
}

func (c *FakeClientWithConflictsErrors) SetCreateKnativeSubResources(createKnativeSubResources bool) {
	c.createKnativeSubResources = createKnativeSubResources
}

func (c *FakeClientWithConflictsErrors) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	err := c.Client.Create(ctx, obj, opts...)
	if err != nil {
		return err
	}

	if c.createKnativeSubResources && obj.GetObjectKind().GroupVersionKind().Kind == "InferenceWorkload" {
		// create configuration
		configuration := &knv1.Configuration{
			ObjectMeta: metav1.ObjectMeta{
				Name:      obj.GetName(),
				Namespace: obj.GetNamespace(),
			},
		}
		err = c.Client.Create(ctx, configuration)
		if err != nil {
			return err
		}

		// create revision
		revision := &knv1.Revision{
			ObjectMeta: metav1.ObjectMeta{
				Name:      obj.GetName() + "-00001",
				Namespace: obj.GetNamespace(),
				Labels: map[string]string{
					constants.LabelRelease: obj.GetName(),
				},
			},
		}
		err = c.Client.Create(ctx, revision)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *FakeClientWithConflictsErrors) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if c.conflictOnUpdate {
		// return conflict error
		return &k8serrors.StatusError{ErrStatus: metav1.Status{Message: "Conflict", Code: http.StatusConflict}}
	}
	if c.unavailableOnUpdate {
		// return unavailable error
		return &k8serrors.StatusError{ErrStatus: metav1.Status{Message: "Unavailable", Code: http.StatusServiceUnavailable}}
	}
	return c.Client.Update(ctx, obj, opts...)
}

func (c *FakeClientWithConflictsErrors) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return c.Client.Patch(ctx, obj, patch, opts...)
}
