/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package karta

import (
	"context"

	kartaresource "github.com/run-ai/karta/pkg/resource"
	corev1 "k8s.io/api/core/v1"
)

//go:generate mockgen -destination=component_mock.go -package=karta -source=component_interface.go

// Component is Karta utils interface for updating component pod specs.
type Component interface {
	UpdateFragmentedPodSpec(ctx context.Context, fragmentedPodSpecs map[string]kartaresource.FragmentedPodSpec) error
	UpdatePodSpec(ctx context.Context, podSpecs map[string]corev1.PodSpec) error
	UpdatePodTemplateSpec(ctx context.Context, podTemplateSpecs map[string]corev1.PodTemplateSpec) error
}
