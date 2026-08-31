/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

const (
	WorkloadCategoryLabel = "run.ai/category"

	ControlPlaneManagedLabel = "run.ai/control-plane-managed"
	ConvertedFromNativeLabel = "run.ai/converted-from-native"

	SubmittedByLabel      = "run.ai/submitted-by"
	SubmittedByAnnotation = SubmittedByLabel // same key on Job metadata annotations in iw_tw

	// UserAnnotation is the pod-template annotation carrying the submitting user.
	// Mirrors cluster constants.AnnotationUserName ("user"), which the native
	// RunaiJob builder stamps on the pod template from spec.username.
	UserAnnotation = "user"

	RunaiScheduler = "runai-scheduler"

	AutoDeleteTimeoutAnnotation = "run.ai/auto-delete-after-completion-sec"

	Oauth2ContainerDefaultPort = int32(4180)

	S3DefaultAccessKeyIdKey     = "AccessKeyId"
	S3DefaultSecretAccessKeyKey = "SecretKey"

	GpuMemoryLimitEnvVar         = "RUNAI_GPU_MEMORY_LIMIT"
	GpuPortionLimitEnvVar        = "RUNAI_GPU_PORTION_LIMIT"
	GpuMemoryAnnotation          = "gpu-memory"
	GpuFractionAnnotation        = "gpu-fraction"
	GpuFractionDevicesAnnotation = "gpu-fraction-num-devices"
)

// Networking resource-name prefixes. These mirror the V1 cluster builder's
// "workloadInitials" convention so the V2 control-plane can produce K8s
// Service/Ingress/Route/HTTPRoute names that never collide between workload
// kinds in the same namespace.
const (
	// NetworkingResourcePrefixJob is used by the BatchJob composer
	// (interactive / training workloads).
	NetworkingResourcePrefixJob = "job"

	// NetworkingResourcePrefixKubeflow is used by the Kubeflow-based composer
	// (distributed workloads).
	NetworkingResourcePrefixKubeflow = "kf"

	// NetworkingResourcePrefixInference is used by the inference composer and
	// mirrors the legacy inference-workload controller's resource initials.
	NetworkingResourcePrefixInference = "infw"
)
