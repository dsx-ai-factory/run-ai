/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package constants

const (
	RunaiQueueLabel       = "runai/queue"
	RunaiSchedulerName    = "runai-scheduler"
	RunaiProjectNameLabel = "project"

	LabelJobUrl            string = "JobUrl"
	LabelJupyterJob        string = "jupyter"
	LabelTensorboard       string = "tensorboard"
	LabelRevisionId        string = "serving.knative.dev/revisionUID"
	LabelServiceId         string = "serving.knative.dev/serviceUID"
	LabelInferenceworkload string = "run.ai/inferenceworkload"

	AnnotationUser            string = "user"

	IngressJobType string = "ingress"

	WorkloadKindLabelKey = "run.ai/workload-kind"
	WorkloadNameLabelKey = "run.ai/workload-name"

	LabelWorkloadKind = "workloadKind"
	LabelWorkloadName = "workloadName"

	// LeaderWorkerSet labels for distributed workloads
	LwsWorkerIndexLabelKey = "leaderworkerset.sigs.k8s.io/worker-index"

	// Pod task roles for distributed workloads
	PodTaskLeader = "leader"
	PodTaskWorker = "worker"
)
