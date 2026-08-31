/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller_runtime

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateOrUpdate(ctx context.Context, c client.Client, obj client.Object) error {
	inClusterObj := obj.DeepCopyObject().(client.Object)
	err := c.Get(ctx, client.ObjectKeyFromObject(obj), inClusterObj)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get object: %w", err)
		}

		err = c.Create(ctx, obj)
		if err != nil {
			return fmt.Errorf("failed to create object: %w", err)
		}

		return nil
	}

	obj.SetResourceVersion(inClusterObj.GetResourceVersion())
	err = c.Update(ctx, obj)
	if err != nil {
		return fmt.Errorf("failed to update object: %w", err)
	}

	return nil
}
