/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package constants

const (
	NodePoolNameLabel        = "runai/node-pool"
	DefaultNodePoolName      = "default"
	PodGroupAnnotationForPod = "pod-group-name"
	AppLabelName             = "app"
	ScalingPodAppLabelValue  = "scaling-pod"
	AnnotationNodepools      = "runai-nodepools"

	ControlPlaneWorkloadIdAnnotationForWorkload = "workloadId"
	ControlPlaneWorkloadNameLabelForWorkload    = "workloadName"
	ControlPlaneWorkloadKindLabelForWorkload    = "workloadKind"

	// Deprecated: use TopOwnerWorkloadIdLabel instead
	TopOwnerUidLabel = "run.ai/top-owner-uid"
	// TopOwnerWorkloadIdLabel - The control-plane workload ID of the top owner of the resource
	// if the workload was created in the CP - would be the control-plane workload ID
	// if created in the cluster - would be the k8s UID of the top owner workload object, and will be equal to ClusterTopOwnerUidLabel value
	TopOwnerWorkloadIdLabel = "run.ai/workload-id"
	// ClusterTopOwnerUidLabel - The k8s UID of the top owner workload object of the resource
	ClusterTopOwnerUidLabel = "run.ai/cluster-top-owner-uid"

	RunaiSchedulerName = "runai-scheduler"

	EnforceSchedulerAnnotationName = "runai/enforce-scheduler-name"

	// ControlPlaneManagedLabel indicates that the resource is managed by the control plane
	ControlPlaneManagedLabel = "run.ai/control-plane-managed"

	WorkloadCategoryLabel = "run.ai/category"

	ExternalWorkloadName = "run.ai/external-workload-name"

	KnativeRevisionLabel = "serving.knative.dev/revision"

	WorkloadRunnerNameLabel = "run.ai/workload-runner-name"

	// EwiStatusAnnotation: set "true" by EWI so cluster-sync skips the wlr phase override.
	EwiStatusAnnotation = "run.ai/ewi-status"

	// WorkloadIssueConditionType marks an ObservedCondition EWI adds to carry a runtime issue
	// (Reason = issue group, Message = detail); cluster-sync maps it to the workload's issues.
	WorkloadIssueConditionType = "WorkloadIssue"
	// WorkloadCliCommandAnnotation holds the original CLI submit command. Set by the
	// run:ai CLI on the workload, it is surfaced as additionalFields.cliCommand in the
	// workloads-service read model (both the native and the redirected ANY_WORKLOAD paths).
	WorkloadCliCommandAnnotation = "runai-submit-command"

	// ConvertedFromNativeLabel marks a WorkloadRunner that was produced by the
	// control-plane's v1->v2 cast (see iw_tw_to_job in composer_v2). The
	// runtime-rules uses presence of this label as the scope gate for
	// enforcement: runners without it are out of scope and never reach the
	// time-limit / idle queue.
	ConvertedFromNativeLabel = "run.ai/converted-from-native"
)
