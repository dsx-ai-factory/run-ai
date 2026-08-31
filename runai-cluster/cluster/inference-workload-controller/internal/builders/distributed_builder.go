/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	workloadbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	dinfp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferencepolicy/v1alpha1"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

type DistributedInferenceBuilderInterface interface {
	BuildLws(ctx context.Context, workload *dinfw.DistributedInferenceWorkload, owner client.Object) (*lwsv1.LeaderWorkerSet, error)
	BuildSubResources(ctx context.Context, workload *dinfw.DistributedInferenceWorkload, owner client.Object) ([]client.Object, error)
	GetPreview(ctx context.Context, workload *dinfw.DistributedInferenceWorkload) ([]client.Object, error)
}

type DistributedInferenceBuilder struct {
	k8sClient          client.Client
	lwsBuilder         LwsBuilder
	servingPortBuilder *ServingPortBuilder
	pvcBuilder         *workloadbuilders.PvcBuilder
	mnnvlBuilder       *MnnvlBuilder
}

func NewDistributedInferenceBuilder(k8sClient client.Client, isOpenshift bool, elasticEnabled bool) *DistributedInferenceBuilder {
	return &DistributedInferenceBuilder{
		k8sClient:          k8sClient,
		lwsBuilder:         NewLwsBuilder(k8sClient, elasticEnabled),
		servingPortBuilder: NewServingPortBuilder(k8sClient, isOpenshift),
		pvcBuilder:         workloadbuilders.NewPvcBuilder(k8sClient),
		mnnvlBuilder:       NewMnnvlBuilder(k8sClient, elasticEnabled),
	}
}

func (b *DistributedInferenceBuilder) GetPreview(ctx context.Context, workload *dinfw.DistributedInferenceWorkload) ([]client.Object, error) {
	log := logger.FromContext(ctx)
	objects := []client.Object{}

	// build lws
	lws, err := b.lwsBuilder.GetPreview(ctx, workload)
	if err != nil {
		return nil, err
	}
	objects = append(objects, lws)

	// build serving ports
	servingPortObjects, err := b.servingPortBuilder.GetPreview(ctx, workload)
	if err != nil {
		return nil, err
	}
	objects = append(objects, servingPortObjects...)

	// build leader sub resources
	leaderSubResources, err := b.previewSubResources(ctx, &workload.ObjectMeta, workload.GetName(), workload, &workload.LeaderSpec, string(LwsEntityNameLeader))
	if err != nil {
		return nil, err
	}
	objects = append(objects, leaderSubResources...)

	// build worker sub resources
	workerSpec := GetDistributedInferenceWorkerSpec(workload)
	workerSubResources, err := b.previewSubResources(ctx, &workload.ObjectMeta, workload.GetName(), workload, workerSpec, string(LwsEntityNameWorker))
	if err != nil {
		return nil, err
	}
	objects = append(objects, workerSubResources...)

	// build mnnvl resources
	mnnvlResources, err := b.mnnvlBuilder.GetPreview(workload)
	if err != nil {
		return nil, err
	}
	objects = append(objects, mnnvlResources...)

	log.Info("built workload resources", "workload", workload.Name, "resources", len(objects))
	return objects, nil
}

func (b *DistributedInferenceBuilder) Build(ctx context.Context, workload *dinfw.DistributedInferenceWorkload, owner client.Object) ([]client.Object, error) {
	// build lws
	lws, err := b.BuildLws(ctx, workload, owner)
	if err != nil {
		return nil, err
	}

	// build sub resources
	subResources, err := b.BuildSubResources(ctx, workload, owner)
	if err != nil {
		return nil, err
	}

	return append([]client.Object{lws}, subResources...), nil
}

func (b *DistributedInferenceBuilder) BuildLws(ctx context.Context, workload *dinfw.DistributedInferenceWorkload, owner client.Object) (*lwsv1.LeaderWorkerSet, error) {
	log := logger.FromContext(ctx)

	lws, err := b.lwsBuilder.Build(ctx, workload, owner)
	if err != nil {
		log.Error(err, "failed to build lws", "error", err)
		return nil, err
	}

	log.Info("built lws", "workload", workload.Name)
	return lws, nil
}

func (b *DistributedInferenceBuilder) BuildSubResources(ctx context.Context, workload *dinfw.DistributedInferenceWorkload, owner client.Object) ([]client.Object, error) {
	log := logger.FromContext(ctx)
	objects := []client.Object{}

	// build serving ports
	servingPortObjects, err := b.servingPortBuilder.Build(ctx, workload)
	if err != nil {
		return nil, err
	}
	objects = append(objects, servingPortObjects...)

	// build mnnvl resources
	mnnvlResources, err := b.mnnvlBuilder.Build(ctx, workload)
	if err != nil {
		return nil, err
	}
	objects = append(objects, mnnvlResources...)

	// build leader sub resources
	leaderSubResources, err := b.buildSubResources(ctx, &workload.ObjectMeta, workload.GetName(), owner, &workload.LeaderSpec, string(LwsEntityNameLeader))
	if err != nil {
		return nil, err
	}
	objects = append(objects, leaderSubResources...)

	// build worker sub resources
	workerSpec := GetDistributedInferenceWorkerSpec(workload)
	workerSubResources, err := b.buildSubResources(ctx, &workload.ObjectMeta, workload.GetName(), owner, workerSpec, string(LwsEntityNameWorker))
	if err != nil {
		return nil, err
	}
	objects = append(objects, workerSubResources...)

	log.Info("built sub resources", "workload", workload.Name, "resources", len(objects))
	return objects, nil
}

func (b *DistributedInferenceBuilder) previewSubResources(
	ctx context.Context, objectMeta *metav1.ObjectMeta, workloadBaseName string,
	owner client.Object, entitySpec *dinfp.DistributedInferenceEntitySpec, entityName string,
) ([]client.Object, error) {
	objects := []client.Object{}

	// preview pvcs
	pvcList, err := b.pvcBuilder.GetPreview(ctx, objectMeta, workloadBaseName, owner, entitySpec.Pvcs, entityName)
	if err != nil {
		return nil, err
	}

	for _, pvc := range pvcList {
		pvcObj := &corev1.PersistentVolumeClaim{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "PersistentVolumeClaim",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      pvc.Name,
				Namespace: objectMeta.Namespace,
			},
		}
		objects = append(objects, pvcObj)
	}

	return objects, nil
}

func (b *DistributedInferenceBuilder) buildSubResources(
	ctx context.Context, objectMeta *metav1.ObjectMeta, workloadBaseName string,
	owner client.Object, entitySpec *dinfp.DistributedInferenceEntitySpec, entityName string,
) ([]client.Object, error) {
	objects := []client.Object{}

	// build pvcs
	pvcList, err := b.pvcBuilder.Build(ctx, constants.LwsKind, objectMeta, workloadBaseName, owner, entitySpec.Pvcs, entityName)
	if err != nil {
		return nil, err
	}
	for _, pvc := range pvcList {
		objects = append(objects, &pvc)
	}

	return objects, nil
}
