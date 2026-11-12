/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inference_endpoint_controller

import (
	"context"
	"fmt"
	kerrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/inference-endpoint-controller/resources"
	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

const (
	Name = operands.RunaiInferenceEndpointControllerName
)

//+kubebuilder:rbac:namespace=runai,groups=apps,resources=deployments,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="",resources=serviceaccounts,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="",resources=services,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;create;update;watch;patch;delete;list

type operand struct {
	k8sReader   kc.HybridClient
	runaiConfig *runv1.RunaiConfig
	resources   []operands.Resource
}

func (o *operand) Name() string {
	return Name
}

func (o *operand) Resources() []operands.Resource {
	return o.resources
}

func (o *operand) Status(ctx context.Context) (operands.Status, error) {
	discoveryClient, err := o.k8sReader.NewDiscoveryClient()
	if err != nil {
		return operands.NotReadyStatus("Cannot check inference API prerequisites"), fmt.Errorf("create discovery client: %w", err)
	}
	for _, required := range []struct{ groupVersion, kind string }{
		{"inference.platform.run.ai/v1alpha1", "InferenceEndpoint"},
		{"gateway.networking.k8s.io/v1", "HTTPRoute"},
		{"inference.networking.k8s.io/v1", "InferencePool"},
		{"inference.networking.x-k8s.io/v1alpha2", "InferenceModelRewrite"},
		{"inference.networking.x-k8s.io/v1alpha2", "InferenceObjective"},
	} {
		list, err := discoveryClient.ServerResourcesForGroupVersion(required.groupVersion)
		if err != nil && !kerrors.IsNotFound(err) {
			return operands.NotReadyStatus("Cannot check inference API prerequisites"), fmt.Errorf("discover %s: %w", required.groupVersion, err)
		}
		found := false
		if list != nil {
			for _, resource := range list.APIResources {
				if resource.Kind == required.kind {
					found = true
					break
				}
			}
		}
		if !found {
			return operands.NotReadyStatus(fmt.Sprintf("Required API %s (%s) is not installed", required.kind, required.groupVersion)), nil
		}
	}
	return utils.Status(ctx, o.resources)
}

func (o *operand) Enabled() bool {
	return *o.runaiConfig.Spec.InferenceEndpointController.Enabled
}

func Operand(k8sReader kc.HybridClient, runaiConfig *runv1.RunaiConfig) operands.Operand {
	return &operand{
		runaiConfig: runaiConfig,
		k8sReader:   k8sReader,
		resources: []operands.Resource{
			resources.NewDeploymentResource(Name, k8sReader, runaiConfig),
			resources.NewServiceAccountResource(Name, k8sReader, runaiConfig),
			resources.NewServiceMonitorResource(Name, k8sReader, runaiConfig),
			resources.NewServiceResource(Name, k8sReader, runaiConfig),
		},
	}
}
