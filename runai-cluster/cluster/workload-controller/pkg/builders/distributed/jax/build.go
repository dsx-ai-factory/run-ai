/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package jax

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	kubeflowcommon "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	workloadconfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	distcommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/common"
	workloadCommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/factory"
)

const (
	buildingMessage           = "building JAXJob job CR"
	completedResourcesMessage = "completed building JAXJob job CR"
)

var (
	MasterNotSupportedError = errors.New("JAX does not support master configuration." +
		" Please remove the master configuration from the workload spec")
	FieldsNotSupportedError = errors.New("JAX does not support the given fields")
)

type Builder struct {
	distcommon.DistributedBuilder
	Testing bool
	JAXJob  *kubeflowcommon.JAXJob
}

// Init register the builder to the builders factory
func Init() {
	factory.BuildersRegistry[factory.JAXJobBuilder] = NewJAXJobBuilderTesting
}

// NewJAXJobBuilderTesting is a constructor of JAXJob
// only for testing
func NewJAXJobBuilderTesting(fromObject interface{}, toResource interface{}, k8sClient client.Client,
	options factory.BuilderOptions,
	podLabels map[string]string, podAnnotations map[string]string) factory.Builder {
	builderInterface, err := NewJAXJobBuilderWithExistingNodeAffinities(fromObject, toResource, k8sClient, options,
		nil, podLabels, podAnnotations)
	if err != nil {
		panic(err)
	}
	builder := builderInterface.(*Builder)
	builder.Testing = true
	return builder
}

func NewJAXJobBuilderWithExistingNodeAffinities(fromObject interface{}, toResource interface{},
	k8sClient client.Client, options factory.BuilderOptions,
	existingPodsSpec map[workloadCommon.ReplicaType]*corev1.PodSpec,
	podLabels map[string]string, podAnnotations map[string]string) (factory.Builder, error) {

	mergedWorkload, ok := fromObject.(*v2alpha1.DistributedWorkload)
	if !ok {
		return nil, fmt.Errorf(distcommon.WorkloadCastErrorMsg, fromObject)
	}
	job, ok := toResource.(*kubeflowcommon.JAXJob)
	if !ok {
		return nil, fmt.Errorf(distcommon.JobCastErrorMsg, &kubeflowcommon.JAXJob{}, toResource)
	}
	workload := &jobs.JobWorkload{}
	workload.MasterSpec = &jobs.JobSpec{}
	mergedWorkload.CopyTo(workload)
	workload.Labels = make(map[string]string)
	sharedResourcesBuilder := common.NewShardResourcesBuilder(common.NewServiceBuilder(k8sClient),
		common.NewExternalsUrlsBuilder(k8sClient, workloadconfig.IsOpenshift()), common.NewPvcBuilder(k8sClient), common.NewCredentialsBuilder(k8sClient))
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
			ContainerName:   kubeflowcommon.JAXJobDefaultContainerName,
			MergedWorkload:  mergedWorkload,
		},
		JAXJob: job,
	}

	if mergedWorkload.MasterSpec != nil {
		result.DistSharedFields = &mergedWorkload.MasterSpec.WorkloadFields
	}

	return result, nil
}

func (b *Builder) Build(ctx context.Context) error {
	builderLog := logger.FromContext(ctx)

	builderLog.Info(buildingMessage)

	b.buildJAXJobTypeMeta(ctx)

	err := b.buildJAXJobObjectMeta(ctx)
	if err != nil {
		return err
	}

	err = b.buildJAXJobSpec(ctx)
	if err != nil {
		return err
	}

	if b.Testing {
		err = b.OldSharedBuilder.Build(ctx, &b.JAXJob.ObjectMeta, constants.KindJAXJob, false)
		if err != nil {
			return err
		}
	}

	builderLog.Info(completedResourcesMessage)
	return nil
}

func (b *Builder) buildJAXJobTypeMeta(_ context.Context) {
	b.JAXJob.TypeMeta = metav1.TypeMeta{
		APIVersion: kubeflowcommon.SchemeGroupVersion.String(),
		Kind:       constants.KindJAXJob,
	}
}

func (b *Builder) GetJob() client.Object {
	return b.JAXJob
}

func (b *Builder) buildJAXJobObjectMeta(ctx context.Context) error {
	if b.JAXJob.ObjectMeta.Labels == nil {
		b.JAXJob.ObjectMeta.Labels = make(map[string]string, 0)
	}

	err := b.GetObjectMetaLabels(ctx, b.JAXJob.ObjectMeta.Labels)
	if err != nil {
		return err
	}

	b.JAXJob.Labels[constants.LabelApp] = strings.ToLower(constants.AppJAXJob)
	b.JAXJob.Labels[constants.LabelCreatedBy] = constants.KindJAXJob

	return b.BuildObjectMeta(ctx, &b.JAXJob.ObjectMeta)
}

func (b *Builder) buildJAXJobSpec(ctx context.Context) error {
	err := b.BuildRunPolicy(&b.JAXJob.Spec.RunPolicy)
	if err != nil {
		return err
	}

	b.JAXJob.Spec.JAXReplicaSpecs = map[kubeflowcommon.ReplicaType]*kubeflowcommon.ReplicaSpec{}

	// jax does not support master
	err = ValidateWorkload(b.MergedWorkload)
	if err != nil {
		return err
	}
	// replica specs - worker
	workerSpec := kubeflowcommon.ReplicaSpec{}
	b.JAXJob.Spec.JAXReplicaSpecs[kubeflowcommon.JAXJobReplicaTypeWorker] = &workerSpec

	err = b.BuildWorkerReplicaSpec(ctx, &workerSpec, workloadCommon.Worker)
	if err != nil {
		return err
	}

	return nil
}

func (b *Builder) ToYaml(ctx context.Context, w *bytes.Buffer) error {

	b.JAXJob.CreationTimestamp = distcommon.FixedDate
	b.JAXJob.DeletionTimestamp = &distcommon.FixedDate

	marshalled, err := yaml.Marshal(b.JAXJob)
	if err != nil {
		return err
	}

	_, err = w.Write(marshalled)
	if err != nil {
		return err
	}

	return b.OldSharedBuilder.ToYaml(ctx, w)
}

func NewJAXJobFn() client.Object {
	return &kubeflowcommon.JAXJob{}
}
