/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package operands

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// deleteOnlyResource is a Resource that is only ever routed to deletion: it
// builds a bare object (name+namespace+type) so the reconcile loop can Get/Delete
// it by key, and reports disabled so it is never created or updated. Use it in a
// disabled operand to clean up orphaned or renamed resources.
type deleteOnlyResource struct {
	operandName string
	build       func() client.Object
}

// NewDeleteOnlyResource returns a Resource that deletes the object built by build.
func NewDeleteOnlyResource(operandName string, build func() client.Object) Resource {
	return &deleteOnlyResource{operandName: operandName, build: build}
}

func (r *deleteOnlyResource) OperandName() string { return r.operandName }

func (r *deleteOnlyResource) Object(context.Context) (client.Object, error) { return r.build(), nil }

func (r *deleteOnlyResource) Status(context.Context) error { return nil }

func (r *deleteOnlyResource) Enabled() bool { return false }
