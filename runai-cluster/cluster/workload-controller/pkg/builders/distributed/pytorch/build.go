/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pytorch

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	workloadCommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/common/constants"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/topology"
	workloadconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	distcommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/common"
	"github.com/ghodss/yaml"
	kubeflowcommon "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/factory"
)

type Builder struct {
	distcommon.DistributedBuilder
	Testing bool
	PyTorch *kubeflowcommon.PyTorchJob
}

var builderLog = logger.Log.WithName("builder.kubeflowcommon")

// Init register the builder to the builders factory
func Init() {
	factory.BuildersRegistry[factory.PyTorchBuilder] = NewPyTorchBuilderForTesting
}

// NewPyTorchBuilderForTesting is a constructor of PyTorchJob
// only for testing
func NewPyTorchBuilderForTesting(fromObject interface{}, toResource interface{}, k8sClient client.Client,
	options factory.BuilderOptions,
	podLabels map[string]string, podAnnotation map[string]string) factory.Builder {
	builderInterface, err := NewPyTorchBuilderWithExistingNodeAffinities(fromObject, toResource, k8sClient, options,
		nil, podLabels, podAnnotation)
	if err != nil {
		panic(err)
	}
	builder := builderInterface.(*Builder)
	builder.Testing = true
	return builder
}

func NewPyTorchBuilderWithExistingNodeAffinities(fromObject interface{}, toResource interface{},
	k8sClient client.Client, options factory.BuilderOptions,
	existingPodsSpec map[workloadCommon.ReplicaType]*corev1.PodSpec,
	podLabels map[string]string, podAnnotations map[string]string) (factory.Builder, error) {

	mergedWorkload, ok := fromObject.(*v2alpha1.DistributedWorkload)
	if !ok {
		return nil, fmt.Errorf(distcommon.WorkloadCastErrorMsg, toResource)
	}
	workload := &jobs.JobWorkload{}
	workload.MasterSpec = &jobs.JobSpec{}
	mergedWorkload.CopyTo(workload)

	workload.Labels = make(map[string]string)
	sharedResourcesBuilder := common.NewShardResourcesBuilder(common.NewServiceBuilder(k8sClient),
		common.NewExternalsUrlsBuilder(k8sClient, workloadconfig.IsOpenshift()), common.NewPvcBuilder(k8sClient))
	job, ok := toResource.(*kubeflowcommon.PyTorchJob)
	if !ok {
		return nil, fmt.Errorf(distcommon.JobCastErrorMsg, &kubeflowcommon.PyTorchJob{}, toResource)
	}
	result := &Builder{
		DistributedBuilder: distcommon.DistributedBuilder{
			PodLabels:      podLabels,
			PodAnnotations: podAnnotations,
			OldSharedBuilder: common.OldSharedBuilder{
				ObjectNaming:                   &mergedWorkload.ObjectMeta,
				SharedFields:                   &mergedWorkload.Spec.WorkloadFields,
				SharedStatus:                   &mergedWorkload.Status.WorkloadStatus,
				SharedResources:                &common.SharedResources{},
				InternalSharedResourcesBuilder: sharedResourcesBuilder,
				Workload:                       workload,
				Options:                        options,
				K8sClient:                      k8sClient,
			},
			ExisingPodsSpec: existingPodsSpec,
			ContainerName:   kubeflowcommon.PyTorchJobDefaultContainerName,
			MergedWorkload:  mergedWorkload,
		},
		PyTorch: job,
	}

	if mergedWorkload.MasterSpec != nil {
		result.DistSharedFields = &mergedWorkload.MasterSpec.WorkloadFields
	}

	return result, nil
}

func (b *Builder) GetJob() client.Object {
	return b.PyTorch
}

func (b *Builder) Build(ctx context.Context) error {
	builderLog.Info("building all PyTorch related resources")

	b.buildPyTorchTypeMeta(ctx)

	err := b.buildPyTorchObjectMeta(ctx)
	if err != nil {
		return err
	}

	err = b.buildPyTorchSpec(ctx)
	if err != nil {
		return err
	}
	// TODO: When refactoring builders test we should remove this if statement
	// only when yamliy testing build the owned resources from this builder
	// this workaround is needed because the old builder is not aware of the new builder and refactoring
	// it will take time
	if b.Testing {
		return b.OldSharedBuilder.Build(ctx, &b.PyTorch.ObjectMeta, constants.KindPyTorchJob, false)
	}

	builderLog.Info("completed building all the resources")
	return nil
}

func (b *Builder) buildPyTorchTypeMeta(_ context.Context) {
	b.PyTorch.TypeMeta = metav1.TypeMeta{
		APIVersion: kubeflowcommon.SchemeGroupVersion.String(),
		Kind:       constants.KindPyTorchJob,
	}
}

func (b *Builder) buildPyTorchObjectMeta(ctx context.Context) error {
	if b.PyTorch.ObjectMeta.Labels == nil {
		b.PyTorch.ObjectMeta.Labels = make(map[string]string, 0)
	}

	err := b.GetObjectMetaLabels(ctx, b.PyTorch.ObjectMeta.Labels)
	if err != nil {
		return err
	}

	b.PyTorch.Labels[constants.LabelApp] = strings.ToLower(constants.AppPyTorchJob)
	b.PyTorch.Labels[constants.LabelCreatedBy] = constants.KindPyTorchJob

	return b.BuildObjectMeta(ctx, &b.PyTorch.ObjectMeta)
}

func (b *Builder) buildPyTorchSpec(ctx context.Context) error {
	err := b.BuildRunPolicy(&b.PyTorch.Spec.RunPolicy)
	if err != nil {
		return err
	}
	err = b.BuildPytorchElasticSpec(ctx)
	if err != nil {
		return err
	}

	// If whole GPUs were requested, put the number of GPUs in NprocPerNode.
	if b.MergedWorkload.Spec.Gpu.Deref() >= 1 { // Gpu is the old field we keep for backward compatibility.
		b.PyTorch.Spec.NprocPerNode = pointer.String(b.MergedWorkload.Spec.Gpu.String())
	} else if b.MergedWorkload.Spec.GpuDevices.Deref() >= 1 {
		b.PyTorch.Spec.NprocPerNode = pointer.String(b.MergedWorkload.Spec.GpuDevices.String())
	}

	b.PyTorch.Spec.PyTorchReplicaSpecs = map[kubeflowcommon.ReplicaType]*kubeflowcommon.ReplicaSpec{}

	// replica specs - Master
	// created only if NoMaster is not set

	if !b.MergedWorkload.Spec.DistributedFields.NoMaster.Deref() {
		replicaSpec := kubeflowcommon.ReplicaSpec{}
		b.PyTorch.Spec.PyTorchReplicaSpecs[kubeflowcommon.PyTorchJobReplicaTypeMaster] = &replicaSpec

		err = b.BuildMasterReplicaSpec(ctx, &replicaSpec, workloadCommon.Master)
		if err != nil {
			return err
		}
	}

	// replica specs - worker

	workerSpec := kubeflowcommon.ReplicaSpec{}
	b.PyTorch.Spec.PyTorchReplicaSpecs[kubeflowcommon.PyTorchJobReplicaTypeWorker] = &workerSpec

	err = b.BuildWorkerReplicaSpec(ctx, &workerSpec, workloadCommon.Worker)
	if err != nil {
		return err
	}

	b.projectSegmentedTopologyAnnotations()

	return nil
}

func (b *Builder) projectSegmentedTopologyAnnotations() {
	workerSpec := b.PyTorch.Spec.PyTorchReplicaSpecs[kubeflowcommon.PyTorchJobReplicaTypeWorker]
	if workerSpec == nil {
		return
	}

	workerAnnotations := workerSpec.Template.Annotations
	if _, segmented := workerAnnotations[topology.KAISegmentSizeAnnotationKey]; !segmented {
		return
	}

	delete(b.PyTorch.Annotations, topology.KAISegmentSizeAnnotationKey)
	delete(b.PyTorch.Annotations, topology.KAISegmentTopologyPreferredAnnotationKey)
	delete(b.PyTorch.Annotations, topology.KAISegmentTopologyRequiredAnnotationKey)

	delete(workerAnnotations, topology.KAITopologyAnnotationPreferredKey)
	delete(workerAnnotations, topology.KAITopologyAnnotationRequiredKey)

	masterSpec := b.PyTorch.Spec.PyTorchReplicaSpecs[kubeflowcommon.PyTorchJobReplicaTypeMaster]
	if masterSpec == nil {
		return
	}

	masterAnnotations := masterSpec.Template.Annotations
	if masterAnnotations == nil {
		masterAnnotations = map[string]string{}
		masterSpec.Template.Annotations = masterAnnotations
	}
	if topologyName, exists := workerAnnotations[topology.KAITopologyAnnotationKey]; exists {
		masterAnnotations[topology.KAITopologyAnnotationKey] = topologyName
	}
	delete(masterAnnotations, topology.KAITopologyAnnotationPreferredKey)
	delete(masterAnnotations, topology.KAITopologyAnnotationRequiredKey)
	delete(masterAnnotations, topology.KAISegmentSizeAnnotationKey)
	delete(masterAnnotations, topology.KAISegmentTopologyPreferredAnnotationKey)
	delete(masterAnnotations, topology.KAISegmentTopologyRequiredAnnotationKey)
}

func (b *Builder) BuildPytorchElasticSpec(_ context.Context) error {
	pytorchFields := b.MergedWorkload.Spec.DistributedFields.PyTorchJob.Deref()
	if pytorchFields.MinReplicas != nil || pytorchFields.MaxReplicas != nil {
		b.PyTorch.Spec.ElasticPolicy = &kubeflowcommon.ElasticPolicy{}
		if pytorchFields.MinReplicas != nil {
			b.PyTorch.Spec.ElasticPolicy.MinReplicas = pointer.Int32(int32(pytorchFields.MinReplicas.Deref()))
		}
		if pytorchFields.MaxReplicas != nil {
			b.PyTorch.Spec.ElasticPolicy.MaxReplicas = pointer.Int32(int32(pytorchFields.MaxReplicas.Deref()))
		}
		replicas := pointer.Int32Deref(b.NumberOfWorkerReplicas(), constants.DefaultWorkers)
		minReplicasDeFacto := pointer.Int32Deref(b.PyTorch.Spec.ElasticPolicy.MinReplicas, replicas)
		maxReplicasDeFacto := pointer.Int32Deref(b.PyTorch.Spec.ElasticPolicy.MaxReplicas, replicas)

		if minReplicasDeFacto > replicas {
			return fmt.Errorf("minReplicas (%d) must be no greater than replicas (%d)", minReplicasDeFacto, replicas)
		}
		if maxReplicasDeFacto < replicas {
			return fmt.Errorf("maxReplicas (%d) must be no less than replicas (%d)", maxReplicasDeFacto, replicas)
		}
	}
	return nil
}

func (b *Builder) ToYaml(ctx context.Context, w *bytes.Buffer) error {

	b.PyTorch.CreationTimestamp = distcommon.FixedDate
	b.PyTorch.DeletionTimestamp = &distcommon.FixedDate

	marshalled, err := yaml.Marshal(b.PyTorch)
	if err != nil {
		return err
	}

	_, err = w.Write(marshalled)
	if err != nil {
		return err
	}

	return b.OldSharedBuilder.ToYaml(ctx, w)
}

func NewPytorchJobFn() client.Object {
	return &kubeflowcommon.PyTorchJob{}
}
