/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tfjob

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	workloadconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	distcommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/common"
	workloadCommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	"github.com/ghodss/yaml"
	kubeflowcommon "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/factory"
)

type Builder struct {
	distcommon.DistributedBuilder
	Testing bool
	TFJob   *kubeflowcommon.TFJob
}

var builderLog = logger.Log.WithName("builder.kubeflowcommon")

// Init register the builder to the builders factory
func Init() {
	factory.BuildersRegistry[factory.TFJobBuilder] = NewTFJobBuilderForTesting
}

// NewTFJobBuilderForTesting is a constructor of TFJob
// only for testing
func NewTFJobBuilderForTesting(fromObject interface{}, toResource interface{}, k8sClient client.Client,
	options factory.BuilderOptions,
	podLabels map[string]string, podAnnotations map[string]string) factory.Builder {
	builderInterface, err := NewTFJobBuilderWithExistingNodeAffinities(fromObject, toResource, k8sClient, options,
		nil, podLabels, podAnnotations)
	if err != nil {
		panic(err)
	}
	builder := builderInterface.(*Builder)
	builder.Testing = true
	return builder
}

func NewTFJobBuilderWithExistingNodeAffinities(fromObject interface{}, toResource interface{},
	k8sClient client.Client, options factory.BuilderOptions,
	existingPodsSpec map[workloadCommon.ReplicaType]*corev1.PodSpec,
	podLabels map[string]string, podAnnotations map[string]string) (factory.Builder, error) {

	mergedWorkload, ok := fromObject.(*v2alpha1.DistributedWorkload)
	if !ok {
		return nil, fmt.Errorf(distcommon.WorkloadCastErrorMsg, fromObject)
	}
	workload := &jobs.JobWorkload{}
	workload.MasterSpec = &jobs.JobSpec{}
	mergedWorkload.CopyTo(workload)
	workload.Labels = make(map[string]string)
	job, ok := toResource.(*kubeflowcommon.TFJob)
	if !ok {
		return nil, fmt.Errorf(distcommon.JobCastErrorMsg, &kubeflowcommon.TFJob{}, toResource)
	}
	sharedResourcesBuilder := common.NewShardResourcesBuilder(common.NewServiceBuilder(k8sClient),
		common.NewExternalsUrlsBuilder(k8sClient, workloadconfig.IsOpenshift()), common.NewPvcBuilder(k8sClient))
	result := &Builder{
		DistributedBuilder: distcommon.DistributedBuilder{
			PodLabels:      podLabels,
			PodAnnotations: podAnnotations,
			OldSharedBuilder: common.OldSharedBuilder{
				ObjectNaming:                   &mergedWorkload.ObjectMeta,
				SharedFields:                   &mergedWorkload.Spec.WorkloadFields,
				SharedStatus:                   &mergedWorkload.Status.WorkloadStatus,
				SharedResources:                &common.SharedResources{},
				Workload:                       workload,
				InternalSharedResourcesBuilder: sharedResourcesBuilder,
				Options:                        options,
				K8sClient:                      k8sClient,
			},
			ExisingPodsSpec: existingPodsSpec,
			ContainerName:   kubeflowcommon.TFJobDefaultContainerName,
			MergedWorkload:  mergedWorkload,
		},
		TFJob: job,
	}

	if mergedWorkload.MasterSpec != nil {
		result.DistSharedFields = &mergedWorkload.MasterSpec.WorkloadFields
	}
	return result, nil
}

func (b *Builder) GetJob() client.Object {
	return b.TFJob
}

func (b *Builder) Build(ctx context.Context) error {
	builderLog.Info("building all TFJob related resources")

	b.buildTFJobTypeMeta(ctx)

	err := b.buildTFJobObjectMeta(ctx)
	if err != nil {
		return err
	}

	err = b.buildTFJobSpec(ctx)
	if err != nil {
		return err
	}
	// TODO: When refactoring builders test we should remove this if statement
	// only when yamliy testing build the owned resources from this builder
	// this workaround is needed because the old builder is not aware of the new builder and refactoring
	// it will take time
	if b.Testing {
		return b.OldSharedBuilder.Build(ctx, &b.TFJob.ObjectMeta, constants.KindTFJob, false)
	}

	builderLog.Info("completed building all the resources")
	return nil
}

func (b *Builder) buildTFJobTypeMeta(_ context.Context) {
	b.TFJob.TypeMeta = metav1.TypeMeta{
		APIVersion: kubeflowcommon.SchemeGroupVersion.String(),
		Kind:       constants.KindTFJob,
	}
}

func (b *Builder) buildTFJobObjectMeta(ctx context.Context) error {
	if b.TFJob.ObjectMeta.Labels == nil {
		b.TFJob.ObjectMeta.Labels = make(map[string]string, 0)
	}

	err := b.GetObjectMetaLabels(ctx, b.TFJob.ObjectMeta.Labels)
	if err != nil {
		return err
	}

	b.TFJob.Labels[constants.LabelApp] = strings.ToLower(constants.AppTFJob)
	b.TFJob.Labels[constants.LabelCreatedBy] = constants.KindTFJob

	return b.BuildObjectMeta(ctx, &b.TFJob.ObjectMeta)
}

func (b *Builder) buildTFJobSpec(ctx context.Context) error {
	err := b.BuildRunPolicy(&b.TFJob.Spec.RunPolicy)
	if err != nil {
		return err
	}

	b.TFJob.Spec.TFReplicaSpecs = map[kubeflowcommon.ReplicaType]*kubeflowcommon.ReplicaSpec{}

	// replica specs - Master
	// created only if NoMaster is not set

	if !b.MergedWorkload.Spec.DistributedFields.NoMaster.Deref() {
		masterSpec := kubeflowcommon.ReplicaSpec{}
		b.TFJob.Spec.TFReplicaSpecs[kubeflowcommon.TFJobReplicaTypeChief] = &masterSpec

		err = b.BuildMasterReplicaSpec(ctx, &masterSpec, workloadCommon.Chief)
		if err != nil {
			return err
		}
	}

	// replica specs - worker

	workerSpec := kubeflowcommon.ReplicaSpec{}
	b.TFJob.Spec.TFReplicaSpecs[kubeflowcommon.TFJobReplicaTypeWorker] = &workerSpec

	err = b.BuildWorkerReplicaSpec(ctx, &workerSpec, workloadCommon.Worker)
	if err != nil {
		return err
	}

	return nil
}

func (b *Builder) ToYaml(ctx context.Context, w *bytes.Buffer) error {

	b.TFJob.CreationTimestamp = distcommon.FixedDate
	b.TFJob.DeletionTimestamp = &distcommon.FixedDate

	marshalled, err := yaml.Marshal(b.TFJob)
	if err != nil {
		return err
	}

	_, err = w.Write(marshalled)
	if err != nil {
		return err
	}

	return b.OldSharedBuilder.ToYaml(ctx, w)
}

func NewTFJobFn() client.Object {
	return &kubeflowcommon.TFJob{}

}
