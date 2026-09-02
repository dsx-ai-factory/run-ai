/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"
	"fmt"

	kubeflowcommon "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	workloadbuilders "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/builders"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/internal/mnnvlutils"
	workloadCommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/common/constants"
)

func (db *DistributedBuilder) BuildWorkerReplicaSpec(ctx context.Context, workerSpec *kubeflowcommon.ReplicaSpec, workerReplicaName workloadCommon.ReplicaType) error {

	workerSpec.Replicas = db.NumberOfWorkerReplicas()

	err := BuildAnnotationsAndLabels(ctx, db.K8sClient, &workerSpec.Template.ObjectMeta, db.MergedWorkload, db.MergedWorkload.Spec.Gpu,
		db.MergedWorkload.Spec.GpuMemory, db.MergedWorkload.Spec.GpuDevices,
		db.PodLabels, db.PodAnnotations, db.MergedWorkload.Spec.RunaiSchedulerFields)
	if err != nil {
		return fmt.Errorf("failed to build worker annotations and labels: %v", err)
	}

	workloadbuilders.AddUserDefinedAnnotationsAndLabels(&workerSpec.Template.ObjectMeta, &db.MergedWorkload.Spec.ContainerFields)

	mnnvlStatus := mnnvlutils.FromContext(ctx)
	newWorkerPodSpec, err := db.buildPodSpec(ctx, &workerSpec.Template.Spec,
		&db.MergedWorkload.Spec, workerReplicaName, mnnvlStatus.WorkerMode)
	if err != nil {
		return err
	}
	if newWorkerPodSpec == nil {
		return fmt.Errorf("failed to build worker pod spec is nil")
	}
	workerSpec.Template.Spec = *newWorkerPodSpec

	return nil
}
