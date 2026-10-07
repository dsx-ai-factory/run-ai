/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mpi

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
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"

	mpiv2beta1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/factory"
)

type Builder struct {
	distcommon.DistributedBuilder
	Testing bool

	Mpi *mpiv2beta1.MPIJob
}

var builderLog = logger.Log.WithName("builder.mpi")

// Init register the builder to the builders factory
func Init() {
	factory.BuildersRegistry[factory.MpiBuilder] = NewMpiBuilderForTesting
}

// NewMpiBuilderForTesting is a constructor of Builder
// only for testing
func NewMpiBuilderForTesting(fromObject interface{}, toResource interface{}, k8sClient client.Client,
	options factory.BuilderOptions,
	podLabels map[string]string, podAnnotations map[string]string) factory.Builder {
	builderInterface, err := NewMpiBuilderWithExistingNodeAffinity(fromObject, toResource, k8sClient, options,
		nil, podLabels, podAnnotations)
	if err != nil {
		panic(err)
	}
	builder := builderInterface.(*Builder)
	builder.Testing = true
	return builder
}
func NewMpiBuilderWithExistingNodeAffinity(fromObject interface{}, toResource interface{}, k8sClient client.Client,
	options factory.BuilderOptions,
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

	job, ok := toResource.(*mpiv2beta1.MPIJob)
	if !ok {
		return nil, fmt.Errorf(distcommon.JobCastErrorMsg, &mpiv2beta1.MPIJob{}, toResource)
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
			ContainerName:   kubeflowcommon.MPIJobDefaultContainerName,
			ExisingPodsSpec: existingPodsSpec,
			MergedWorkload:  mergedWorkload,
		},
		Mpi: job,
	}

	if mergedWorkload.MasterSpec != nil {
		result.DistSharedFields = &mergedWorkload.MasterSpec.WorkloadFields
	}

	return result, nil
}

func (b *Builder) GetJob() client.Object {
	return b.Mpi
}

func (b *Builder) Build(ctx context.Context) error {
	builderLog.Info("building all MPI related resources")

	b.buildMpiTypeMeta(ctx)

	err := b.buildMpiObjectMeta(ctx)
	if err != nil {
		return err
	}

	err = b.buildMpiSpec(ctx)
	if err != nil {
		return err
	}
	// TODO: When refactoring builders test we should remove this if statement
	// only when yamliy testing build the owned resources from this builder
	// this workaround is needed because the old builder is not aware of the new builder and refactoring
	// it will take time
	if b.Testing {
		err = b.OldSharedBuilder.Build(ctx, &b.Mpi.ObjectMeta, constants.KindMpiJob, false)
		if err != nil {
			return err
		}
	}

	builderLog.Info("completed building all the resources")

	return nil
}

func (b *Builder) buildMpiTypeMeta(_ context.Context) {
	b.Mpi.TypeMeta = metav1.TypeMeta{
		APIVersion: mpiv2beta1.SchemeGroupVersionKind.String(),
		Kind:       constants.KindMpiJob,
	}
}

func (b *Builder) buildMpiObjectMeta(ctx context.Context) error {
	if b.Mpi.ObjectMeta.Labels == nil {
		b.Mpi.ObjectMeta.Labels = make(map[string]string, 0)
	}

	err := b.GetObjectMetaLabels(ctx, b.Mpi.ObjectMeta.Labels)
	if err != nil {
		return err
	}

	b.Mpi.Labels[constants.LabelApp] = strings.ToLower(constants.AppMpiJob)
	b.Mpi.Labels[constants.LabelCreatedBy] = constants.KindMpiJob

	return b.BuildObjectMeta(ctx, &b.Mpi.ObjectMeta)
}

func (b *Builder) buildMpiSpec(ctx context.Context) error {
	err := b.BuildRunPolicy(&b.Mpi.Spec.RunPolicy)
	if err != nil {
		return err
	}

	// Handle SlotsPerWorker: if it was given in the spec of the DistributedWorkload, take it from there. Otherwise, set it as the number of requested GPUs.
	mpiSection := b.MergedWorkload.Spec.DistributedFields.MpiJob.Deref()
	if mpiSection.SlotsPerWorker != nil {
		b.Mpi.Spec.SlotsPerWorker = pointer.Int32(int32(mpiSection.SlotsPerWorker.Deref()))
	} else if b.MergedWorkload.Spec.Gpu.Deref() >= 1 { // Gpu is the old field we keep for backward compatibility.
		b.Mpi.Spec.SlotsPerWorker = pointer.Int32(int32(b.MergedWorkload.Spec.Gpu.Deref()))
	} else if b.MergedWorkload.Spec.GpuDevices.Deref() >= 1 {
		b.Mpi.Spec.SlotsPerWorker = pointer.Int32(int32(b.MergedWorkload.Spec.GpuDevices.Deref()))
	}

	if mpiSection.SshAuthMountPath != nil {
		b.Mpi.Spec.SSHAuthMountPath = mpiSection.SshAuthMountPath.Deref()
	}

	b.Mpi.Spec.MPIReplicaSpecs = map[mpiv2beta1.MPIReplicaType]*kubeflowcommon.ReplicaSpec{}

	// replica specs - launcher

	if b.MergedWorkload.Spec.DistributedFields.NoMaster.Deref() {
		return fmt.Errorf("noMaster is not supported for MPI, as MPI jobs must have a master pod")
	}

	// if launcher exist, we'll add it if not, keep it empty
	// this will solve Kubeflow versions that does not support this field.
	if mpiSection.LauncherCreationPolicy != nil {
		b.Mpi.Spec.LauncherCreationPolicy = mpiv2beta1.LauncherCreationPolicy(mpiSection.LauncherCreationPolicy.Deref())
	}

	replicaSpec := kubeflowcommon.ReplicaSpec{}
	b.Mpi.Spec.MPIReplicaSpecs[mpiv2beta1.MPIReplicaTypeLauncher] = &replicaSpec

	err = b.BuildMasterReplicaSpec(ctx, &replicaSpec, workloadCommon.Launcher)
	if err != nil {
		return err
	}

	// replica specs - worker

	workerSpec := kubeflowcommon.ReplicaSpec{}
	b.Mpi.Spec.MPIReplicaSpecs[mpiv2beta1.MPIReplicaTypeWorker] = &workerSpec

	err = b.BuildWorkerReplicaSpec(ctx, &workerSpec, workloadCommon.Worker)
	if err != nil {
		return err
	}
	if b.ExisingPodsSpec != nil &&
		(b.ExisingPodsSpec[workloadCommon.Worker] != nil || b.ExisingPodsSpec[workloadCommon.Master] != nil) {
		return nil
	}

	for _, podSpec := range b.Mpi.Spec.MPIReplicaSpecs {
		for idx := range podSpec.Template.Spec.Containers {
			container := &podSpec.Template.Spec.Containers[idx]
			if container.Env == nil {
				container.Env = make([]corev1.EnvVar, 0)
			}
			container.Env = append(container.Env, corev1.EnvVar{
				Name:  constants.MpiNumWorkersEnvName,
				Value: fmt.Sprintf("%d", *b.NumberOfWorkerReplicas()),
			})
		}
	}

	return nil
}

func (b *Builder) ToYaml(ctx context.Context, w *bytes.Buffer) error {

	b.Mpi.CreationTimestamp = distcommon.FixedDate
	b.Mpi.DeletionTimestamp = &distcommon.FixedDate

	marshalled, err := yaml.Marshal(b.Mpi)
	if err != nil {
		return err
	}

	_, err = w.Write(marshalled)
	if err != nil {
		return err
	}

	return b.OldSharedBuilder.ToYaml(ctx, w)
}

func NewMpiJobFn() client.Object {
	return &mpiv2beta1.MPIJob{}
}
