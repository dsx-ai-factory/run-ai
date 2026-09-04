/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package supportedhandlers

import (
	"context"
	"errors"

	dynamov1alpha1 "github.com/ai-dynamo/dynamo/deploy/operator/api/v1alpha1"
	dynamov1beta1 "github.com/ai-dynamo/dynamo/deploy/operator/api/v1beta1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/dra"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/handlers"
	v1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/ray/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	rayv1alpha1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	unimplementedTopOwnerHandler UnimplementedTopOwnerHandler = false

	unimplementedTopOwnerHandlerWithConfig = NewTopOwnerHandlerWithConfig(
		unimplementedTopOwnerHandler,
		true,
	)
)

type TopOwnerHandler interface {
	Parse(ctx context.Context, object client.Object, externalWorkload *v1alpha1.ExternalWorkload) error

	GetObject(ctx context.Context, name, namespace string) (client.Object, error)
}

type TopOwnerHandlerWithConfig struct {
	TopOwnerHandler

	IsCrd bool
}

func NewTopOwnerHandlerWithConfig(handler TopOwnerHandler, isCrd bool) TopOwnerHandlerWithConfig {
	return TopOwnerHandlerWithConfig{
		TopOwnerHandler: handler,
		IsCrd:           isCrd,
	}
}

// UnimplementedTopOwnerHandler is a placeholder handler.
// An external-workload with correct metadata will still be created,
// just without the specific info that is derived from the original object.
type UnimplementedTopOwnerHandler bool

func (h UnimplementedTopOwnerHandler) Parse(_ context.Context, _ client.Object,
	_ *v1alpha1.ExternalWorkload) error {
	return nil
}
func (h UnimplementedTopOwnerHandler) GetObject(_ context.Context, _, _ string) (client.Object, error) {
	return nil, errors.ErrUnsupported
}

func GetSupportedGVKsHandlersWithConfig(cachedClient cachedclient.CachedClient, draParser *dra.DRAParser) map[schema.GroupVersionKind]TopOwnerHandlerWithConfig {
	return map[schema.GroupVersionKind]TopOwnerHandlerWithConfig{
		{
			Group:   "apps",
			Version: "v1",
			Kind:    "Deployment",
		}: NewTopOwnerHandlerWithConfig(handlers.NewDeploymentTopOwnerHandler(cachedClient, draParser), false),
		{
			Group:   "machinelearning.seldon.io",
			Version: "v1alpha2",
			Kind:    "SeldonDeployment",
		}: NewTopOwnerHandlerWithConfig(handlers.NewSeldonTopOwnerHandler(cachedClient), true),
		{
			Group:   "machinelearning.seldon.io",
			Version: "v1",
			Kind:    "SeldonDeployment",
		}: NewTopOwnerHandlerWithConfig(handlers.NewSeldonTopOwnerHandler(cachedClient), true),
		{
			Group:   "kubevirt.io",
			Version: "v1",
			Kind:    "VirtualMachineInstance",
		}: NewTopOwnerHandlerWithConfig(handlers.NewKubevirtVMITopOwnerHandler(cachedClient), true),
		{
			Group:   "kubeflow.org",
			Version: "v1",
			Kind:    "TFJob",
		}: NewTopOwnerHandlerWithConfig(handlers.NewKubeflowJobTopOwnerHandler(cachedClient, handlers.KubeflowJobKindNameTFJob), true),
		{
			Group:   "kubeflow.org",
			Version: "v1",
			Kind:    "PyTorchJob",
		}: NewTopOwnerHandlerWithConfig(handlers.NewKubeflowJobTopOwnerHandler(cachedClient, handlers.KubeflowJobKindNamePyTorchJob), true),
		{
			Group:   "kubeflow.org",
			Version: "v1",
			Kind:    "XGBoostJob",
		}: NewTopOwnerHandlerWithConfig(handlers.NewKubeflowJobTopOwnerHandler(cachedClient, handlers.KubeflowJobKindNameXGBoostJob), true),
		{
			Group:   "kubeflow.org",
			Version: "v1",
			Kind:    "JAXJob",
		}: NewTopOwnerHandlerWithConfig(handlers.NewKubeflowJobTopOwnerHandler(cachedClient, handlers.KubeflowJobKindNameJAXJob), true),
		{
			Group:   "kubeflow.org",
			Version: "v1",
			Kind:    "MPIJob",
		}: NewTopOwnerHandlerWithConfig(handlers.NewKubeflowJobTopOwnerHandler(cachedClient, handlers.KubeflowJobKindNameMPIJobV1), true),
		{
			Group:   "kubeflow.org",
			Version: "v2beta1",
			Kind:    "MPIJob",
		}: NewTopOwnerHandlerWithConfig(handlers.NewKubeflowJobTopOwnerHandler(cachedClient, handlers.KubeflowJobKindNameMPIJobV2), true),
		{
			Group:   "kubeflow.org",
			Version: "v1beta1",
			Kind:    "Notebook",
		}: NewTopOwnerHandlerWithConfig(handlers.NewKubeflowNotebookTopOwnerHandler(cachedClient, draParser), true),
		{
			Group:   "batch",
			Version: "v1",
			Kind:    "Job",
		}: NewTopOwnerHandlerWithConfig(handlers.NewJobTopOwnerHandler(cachedClient, draParser), false),
		{
			Group:   "apps",
			Version: "v1",
			Kind:    "StatefulSet",
		}: NewTopOwnerHandlerWithConfig(handlers.NewStatefulSetTopOwnerHandler(cachedClient, draParser), false),
		{
			Group:   "apps",
			Version: "v1",
			Kind:    "ReplicaSet",
		}: NewTopOwnerHandlerWithConfig(handlers.NewReplicaSetTopOwnerHandler(cachedClient, draParser), false),
		{
			Group:   "",
			Version: "v1",
			Kind:    "Pod",
		}: NewTopOwnerHandlerWithConfig(handlers.NewPodTopOwnerHandler(cachedClient, draParser), false),
		{
			// Note: will add full support in the future if needed; only metadata will be supported for now.
			Group:   "amlarc.azureml.com",
			Version: "v1alpha1",
			Kind:    "AmlJob",
		}: unimplementedTopOwnerHandlerWithConfig,
		{
			Group:   "serving.knative.dev",
			Version: "v1",
			Kind:    "Service",
		}: NewTopOwnerHandlerWithConfig(handlers.NewKNativeTopOwnerHandler(cachedClient, draParser), true),
		{
			Group:   "batch",
			Version: "v1",
			Kind:    "CronJob",
		}: NewTopOwnerHandlerWithConfig(handlers.NewCronJobTopOwnerHandler(cachedClient, draParser), false),
		{
			Group:   "workspace.devfile.io",
			Version: "v1alpha2",
			Kind:    "DevWorkspace",
		}: NewTopOwnerHandlerWithConfig(handlers.NewDevWorkspaceTopOwnerHandler(cachedClient), true),
		{
			Group:   "ray.io",
			Version: "v1alpha1",
			Kind:    "RayCluster",
		}: NewTopOwnerHandlerWithConfig(handlers.NewRayClusterTopOwnerHandler(cachedClient, draParser, func() client.Object { return &rayv1alpha1.RayCluster{} }), true),
		{
			Group:   "ray.io",
			Version: "v1alpha1",
			Kind:    "RayJob",
		}: NewTopOwnerHandlerWithConfig(handlers.NewRayJobTopOwnerHandler(cachedClient, draParser, func() client.Object { return &rayv1alpha1.RayJob{} }), true),
		{
			Group:   "ray.io",
			Version: "v1alpha1",
			Kind:    "RayService",
		}: NewTopOwnerHandlerWithConfig(handlers.NewRayServiceTopOwnerHandler(cachedClient, draParser), true),
		{
			Group:   "ray.io",
			Version: "v1",
			Kind:    "RayCluster",
		}: NewTopOwnerHandlerWithConfig(handlers.NewRayClusterTopOwnerHandler(cachedClient, draParser, func() client.Object { return &v1.RayCluster{} }), true),
		{
			Group:   "ray.io",
			Version: "v1",
			Kind:    "RayJob",
		}: NewTopOwnerHandlerWithConfig(handlers.NewRayJobTopOwnerHandler(cachedClient, draParser, func() client.Object { return &v1.RayJob{} }), true),
		{
			Group:   "ray.io",
			Version: "v1",
			Kind:    "RayService",
		}: NewTopOwnerHandlerWithConfig(handlers.NewRayServiceTopOwnerHandler(cachedClient, draParser), true),
		{
			Group:   "tekton.dev",
			Version: "v1",
			Kind:    "TaskRun",
		}: NewTopOwnerHandlerWithConfig(handlers.NewTaskRunTopOwnerHandler(cachedClient), true),
		{
			Group:   "tekton.dev",
			Version: "v1",
			Kind:    "PipelineRun",
		}: NewTopOwnerHandlerWithConfig(handlers.NewPipelineRunTopOwnerHandler(cachedClient), true),
		{
			Group:   "argoproj.io",
			Version: "v1alpha1",
			Kind:    "Workflow",
		}: NewTopOwnerHandlerWithConfig(handlers.NewArgoWorkflowTopOwnerHandler(cachedClient), true),
		{
			Group:   "kubeflow.org",
			Version: "v1alpha1",
			Kind:    "ScheduledWorkflow",
		}: unimplementedTopOwnerHandlerWithConfig,
		{
			Group:   "apps.nvidia.com",
			Version: "v1alpha1",
			Kind:    "NIMCache",
		}: NewTopOwnerHandlerWithConfig(handlers.NewNimCacheTopOwnerHandler(cachedClient), true),
		{
			Group:   "apps.nvidia.com",
			Version: "v1alpha1",
			Kind:    "NIMService",
		}: NewTopOwnerHandlerWithConfig(handlers.NewNimServiceTopOwnerHandler(cachedClient, draParser), true),
		{
			Group:   "egx.nvidia.io",
			Version: "v1",
			Kind:    "SPOTRequest",
		}: NewTopOwnerHandlerWithConfig(handlers.NewNVCFTopOwnerHandler(cachedClient, draParser), true),
		{
			Group:   "nvidia.com",
			Version: "v1alpha1",
			Kind:    "DynamoGraphDeployment",
		}: NewTopOwnerHandlerWithConfig(handlers.NewDynamoGraphDeploymentTopOwnerHandler(cachedClient, draParser,
			func() client.Object { return &dynamov1alpha1.DynamoGraphDeployment{} }), true),
		{
			Group:   "nvidia.com",
			Version: "v1beta1",
			Kind:    "DynamoGraphDeployment",
		}: NewTopOwnerHandlerWithConfig(handlers.NewDynamoGraphDeploymentTopOwnerHandler(cachedClient, draParser,
			func() client.Object { return &dynamov1beta1.DynamoGraphDeployment{} }), true),
	}
}

func GetSupportedGVKsHandlers(cachedClient cachedclient.CachedClient, draParser *dra.DRAParser) map[schema.GroupVersionKind]TopOwnerHandler {
	supportedWithConfig := GetSupportedGVKsHandlersWithConfig(cachedClient, draParser)
	handlers := make(map[schema.GroupVersionKind]TopOwnerHandler)
	for gvk, handlerWithConfig := range supportedWithConfig {
		handlers[gvk] = handlerWithConfig.TopOwnerHandler
	}

	return handlers
}

func GetSupportedGVKs() map[schema.GroupVersionKind]bool {
	supportedWithConfig := GetSupportedGVKsHandlersWithConfig(nil, nil)
	handlers := make(map[schema.GroupVersionKind]bool)
	for gvk := range supportedWithConfig {
		handlers[gvk] = true
	}

	return handlers
}
