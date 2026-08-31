/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Resource struct {
	Kind      metav1.GroupVersionKind
	Name      string
	Namespace string
}

func (r Resource) String() string {
	return fmt.Sprintf("%s/%s", r.Kind.String(), r.Name)
}
