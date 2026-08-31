/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
)

type ProjectResourceHandler interface {
	HandleResource(project kaiv1alpha1.Project) ([]kaiv1alpha1.ProjectCondition, error)
}
