/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package xgboost

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
	Testing    bool
	XGBoostJob *kubeflowcommon.XGBoostJob
}

var builderLog = logger.Log.WithName("builder.kubeflowcommon")

// Init register the builder to the builders factory
func Init() {
	factory.BuildersRegistry[factory.XGBoostJobBuilder] = NewXGBoostJobBuilderTesting
}

// NewXGBoostJobBuilderTesting is a constructor of XGBoostJob
// only for testing
func NewXGBoostJobBuilderTesting(fromObject interface{}, toResource interface{}, k8sClient client.Client,
	options factory.BuilderOptions,
	podLabels map[string]string, podAnnotations map[string]string) factory.Builder {
	builderInterface, err := NewXGBoostJobBuilderWithExistingNodeAffinities(fromObject, toResource, k8sClient, options,
		nil, podLabels, podAnnotations)
	if err != nil {
		panic(err)
	}
	builder := builderInterface.(*Builder)
	builder.Testing = true
	return builder
}

func NewXGBoostJobBuilderWithExistingNodeAffinities(fromObject interface{}, toResource interface{},
	k8sClient client.Client, options factory.BuilderOptions,
	existingPodsSpec map[workloadCommon.ReplicaType]*corev1.PodSpec,
	podLabels map[string]string, podAnnotations map[string]string) (factory.Builder, error) {

	mergedWorkload, ok := fromObject.(*v2alpha1.DistributedWorkload)
	if !ok {
		return nil, fmt.Errorf(distcommon.WorkloadCastErrorMsg, fromObject)
	}
	job, ok := toResource.(*kubeflowcommon.XGBoostJob)
	if !ok {
		return nil, fmt.Errorf(distcommon.JobCastErrorMsg, &kubeflowcommon.XGBoostJob{}, toResource)
	}
	workload := &jobs.JobWorkload{}
	workload.MasterSpec = &jobs.JobSpec{}
	mergedWorkload.CopyTo(workload)
	workload.Labels = make(map[string]string)
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
			ContainerName:   kubeflowcommon.XGBoostJobDefaultContainerName,
			MergedWorkload:  mergedWorkload,
		},
		XGBoostJob: job,
	}

	if mergedWorkload.MasterSpec != nil {
		result.DistSharedFields = &mergedWorkload.MasterSpec.WorkloadFields
	}

	return result, nil
}

func (b *Builder) Build(ctx context.Context) error {
	builderLog.Info("building all XGBoostJob related resources")

	b.buildXGBoostJobTypeMeta(ctx)

	err := b.buildXGBoostJobObjectMeta(ctx)
	if err != nil {
		return err
	}

	err = b.buildXGBoostJobSpec(ctx)
	if err != nil {
		return err
	}

	// TODO: When refactoring builders test we should remove this if statement
	// only when yamliy testing build the owned resources from this builder
	// this workaround is needed because the old builder is not aware of the new builder and refactoring
	// it will take time
	if b.Testing {
		err = b.OldSharedBuilder.Build(ctx, &b.XGBoostJob.ObjectMeta, constants.KindXGBoostJob, false)
		if err != nil {
			return err
		}
	}

	builderLog.Info("completed building all the resources")
	return nil
}

func (b *Builder) buildXGBoostJobTypeMeta(_ context.Context) {
	b.XGBoostJob.TypeMeta = metav1.TypeMeta{
		APIVersion: kubeflowcommon.SchemeGroupVersion.String(),
		Kind:       constants.KindXGBoostJob,
	}
}

func (b *Builder) GetJob() client.Object {
	return b.XGBoostJob
}

func (b *Builder) buildXGBoostJobObjectMeta(ctx context.Context) error {
	if b.XGBoostJob.ObjectMeta.Labels == nil {
		b.XGBoostJob.ObjectMeta.Labels = make(map[string]string, 0)
	}

	err := b.GetObjectMetaLabels(ctx, b.XGBoostJob.ObjectMeta.Labels)
	if err != nil {
		return err
	}

	b.XGBoostJob.Labels[constants.LabelApp] = strings.ToLower(constants.AppXGBoostJob)
	b.XGBoostJob.Labels[constants.LabelCreatedBy] = constants.KindXGBoostJob

	return b.BuildObjectMeta(ctx, &b.XGBoostJob.ObjectMeta)
}

func (b *Builder) buildXGBoostJobSpec(ctx context.Context) error {
	err := b.BuildRunPolicy(&b.XGBoostJob.Spec.RunPolicy)
	if err != nil {
		return err
	}

	b.XGBoostJob.Spec.XGBReplicaSpecs = map[kubeflowcommon.ReplicaType]*kubeflowcommon.ReplicaSpec{}

	// replica specs - Master

	if b.MergedWorkload.Spec.DistributedFields.NoMaster.Deref() {
		return fmt.Errorf("noMaster is not supported for XGBoost jobs, as XGBoost jobs must have a master pod")
	}

	masterSpec := kubeflowcommon.ReplicaSpec{}
	b.XGBoostJob.Spec.XGBReplicaSpecs[kubeflowcommon.XGBoostJobReplicaTypeMaster] = &masterSpec

	err = b.BuildMasterReplicaSpec(ctx, &masterSpec, workloadCommon.Master)
	if err != nil {
		return err
	}

	// replica specs - worker

	workerSpec := kubeflowcommon.ReplicaSpec{}
	b.XGBoostJob.Spec.XGBReplicaSpecs[kubeflowcommon.XGBoostJobReplicaTypeWorker] = &workerSpec

	err = b.BuildWorkerReplicaSpec(ctx, &workerSpec, workloadCommon.Worker)
	if err != nil {
		return err
	}

	return nil
}

func (b *Builder) ToYaml(ctx context.Context, w *bytes.Buffer) error {

	b.XGBoostJob.CreationTimestamp = distcommon.FixedDate
	b.XGBoostJob.DeletionTimestamp = &distcommon.FixedDate

	marshalled, err := yaml.Marshal(b.XGBoostJob)
	if err != nil {
		return err
	}

	_, err = w.Write(marshalled)
	if err != nil {
		return err
	}

	return b.OldSharedBuilder.ToYaml(ctx, w)
}

func NewXGBoostJobFn() client.Object {
	return &kubeflowcommon.XGBoostJob{}
}
