/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	namespace1 = "namespace1"
	namespace2 = "namespace2"
	project1   = "project1"
	project2   = "project2"
)

func createNamespaces() []v1.Namespace {
	namespaces := []v1.Namespace{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace1,
				Labels: map[string]string{
					constants.RunaiQueueLabel: project1,
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace2,
				Labels: map[string]string{
					constants.RunaiQueueLabel: project2,
				},
			},
		},
	}

	return namespaces
}
