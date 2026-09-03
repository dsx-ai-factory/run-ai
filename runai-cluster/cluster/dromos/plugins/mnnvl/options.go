/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mnnvl

import (
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// WorkloadNameLabel is the pod label the plugin stamps by default and that its
// default affinity groups a workload's GPU pods by. Value is the WorkloadRunner
// name; pod affinity is namespace-scoped, so the name is a sufficient key.
const WorkloadNameLabel = "dromos.run.ai/mnnvl-workload-name"

// PodAffinityProvider lets a consumer provide the pod affinity for a workload's
// GPU pods instead of the default; nil means none. The Plugin uses the result
// as-is and merges it into any user-authored affinity (dedup by topology key).
type PodAffinityProvider func(
	runner *dromosv1alpha1.WorkloadRunner,
	resource *unstructured.Unstructured,
) *corev1.PodAffinity

// Options holds the mnnvl plugin's host-platform hooks; build it with the Option funcs.
type Options struct {
	PodAffinityProvider       PodAffinityProvider
	DisableWorkloadLabelStamp bool
}

// Option configures the mnnvl plugin.
type Option func(*Options)

// WithPodAffinityProvider replaces the plugin's default grouping affinity with the
// consumer's provider (a nil return from it means no pod affinity).
func WithPodAffinityProvider(p PodAffinityProvider) Option {
	return func(o *Options) { o.PodAffinityProvider = p }
}

// WithWorkloadLabelStampDisabled stops the plugin from stamping WorkloadNameLabel. By
// default the plugin stamps it so its default affinity has a label to select;
// disable it when a PodAffinityProvider supplies an affinity whose label is stamped
// elsewhere (e.g. by an external controller).
func WithWorkloadLabelStampDisabled() Option {
	return func(o *Options) { o.DisableWorkloadLabelStamp = true }
}
