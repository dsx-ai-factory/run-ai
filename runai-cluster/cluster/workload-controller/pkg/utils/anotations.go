/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func IsAnnotationExists(annotations map[string]string, key string) bool {
	if annotations == nil {
		return false
	}
	_, ok := annotations[key]
	return ok
}

func CopyAnnotationsWithPrefix(newResource client.Object, existingResource client.Object, prefix string) {
	if existingResource.GetAnnotations() == nil {
		return
	}
	if newResource.GetAnnotations() == nil {
		newResource.SetAnnotations(map[string]string{})
	}
	for key, value := range existingResource.GetAnnotations() {
		if strings.HasPrefix(key, prefix) {
			newResource.GetAnnotations()[key] = value
		}
	}
}
