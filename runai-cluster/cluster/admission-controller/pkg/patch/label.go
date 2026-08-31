/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package patch

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AddLabelToObjectMeta(objectMeta *metav1.ObjectMeta, key string, value string) {
	if objectMeta.Labels == nil {
		objectMeta.Labels = make(map[string]string)
	}
	objectMeta.Labels[key] = value
}
