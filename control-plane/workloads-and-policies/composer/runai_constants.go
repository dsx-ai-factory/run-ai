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

	// GpuMemoryLimitEnvVar / GpuPortionLimitEnvVar are the legacy container-env
	// vars used to enforce a GPU memory limit before cluster capability
	// MinVersionGpuMemoryLimitAnnotations (< 3.0.0). Kept only for the
	// backward-compat path in fields_composer_compute.go. Do not add new usages;
	// on cluster >= 3.0.0 the composer emits the per-container CUDA/KAI annotations
	// (see CudaGpuMemoryLimitAnnotationKeyFormat and KaiGpuMemoryPortionLimitAnnotationKeyFormat).
	GpuMemoryLimitEnvVar  = "RUNAI_GPU_MEMORY_LIMIT"
	GpuPortionLimitEnvVar = "RUNAI_GPU_PORTION_LIMIT"

	GpuMemoryAnnotation          = "gpu-memory"
	GpuFractionAnnotation        = "gpu-fraction"
	GpuFractionDevicesAnnotation = "gpu-fraction-num-devices"

	// CudaGpuMemoryLimitAnnotationKeyFormat is the pod-annotation key format the
	// NVIDIA CUDA toolkit consumes to enforce an absolute GPU memory limit on a
	// specific container. The value is a k8s resource quantity (binary units like
	// "8Gi" or decimal units like "8G" are both accepted — pick one consciously,
	// since 8Gi != 8G).
	//   Full key: fmt.Sprintf(CudaGpuMemoryLimitAnnotationKeyFormat, container.Name)
	CudaGpuMemoryLimitAnnotationKeyFormat = "nvidia.com/container.%s.gpu-memory.limit"

	// KaiGpuMemoryPortionLimitAnnotationKeyFormat is the pod-annotation key format
	// the KAI scheduler consumes to express a portion-based GPU memory limit on a
	// specific container. The value is a decimal fraction between 0 and 1 (e.g.
	// "0.8"). KAI translates the portion to an absolute size on the selected GPU
	// before handing it to the CUDA toolkit — see KAI's docs for the exact
	// translation.
	//   Full key: fmt.Sprintf(KaiGpuMemoryPortionLimitAnnotationKeyFormat, container.Name)
	KaiGpuMemoryPortionLimitAnnotationKeyFormat = "kai.scheduler/container.%s.gpu-memory.portion.limit"
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
