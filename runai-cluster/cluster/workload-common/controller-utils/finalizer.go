/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller_utils

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func AddFinalizer(resource client.Object, finalizerToAdd string) {
	finalizers := resource.GetFinalizers()

	found := false
	for _, finalizer := range finalizers {
		if finalizer == finalizerToAdd {
			found = true
			break
		}
	}

	if !found {
		resource.SetFinalizers(append(resource.GetFinalizers(), finalizerToAdd))
	}
}
