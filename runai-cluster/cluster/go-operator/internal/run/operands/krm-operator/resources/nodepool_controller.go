/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	kaicommon "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/common"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clustercommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	nodepoolresources "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/nodepool-controller/resources"
	controller_utils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	nodepool_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/nodepool-controller"
)

// nodePoolMetricsPortName is the name runai publishes the second, unscraped port
// under, kept so anything selecting on it keeps working.
const nodePoolMetricsPortName = "np-metrics"

func convertNodePoolController(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runaiv1.RunaiConfig) (*kaiv1alpha1.NodePoolController, error) {
	config := runaiConfig.Spec.NodePoolController
	if config == nil {
		return nil, nil
	}

	shardArgs, err := nodepoolresources.MarshalShardArgs(
		runaiConfig.Spec.RunAIScheduler, runaiConfig.Spec.Global.NodeAffinity)
	if err != nil {
		return nil, fmt.Errorf("marshal scheduling shard args: %w", err)
	}

	return &kaiv1alpha1.NodePoolController{
		Service: &kaicommon.Service{
			Enabled:   config.Enabled,
			Resources: (*kaicommon.Resources)(config.Resources),
			Affinity:  config.Affinity,
			// This controller has no client-rate configuration in RunaiConfig.
			// Image is KRM's own, and PodDisruptionBudget stays at KRM's default
			// of disabled, matching runai, which creates none.
		},
		ControllerService: &kaiv1alpha1.NodePoolControllerService{
			Metrics:         metricsPortMapping(config.Metrics),
			NodePoolMetrics: nodePoolMetricsPortMapping(config.Service),
			// No webhook port: the validating webhook is off, see Webhooks below.
		},
		Webhooks: &kaiv1alpha1.NodePoolControllerWebhooks{
			// NodePool validation is covered by the control-plane and agent
			// reconcile. The binary defaults this on and would then start a TLS
			// server with no certificate mounted.
			EnableNodePoolValidation: ptr.To(false),
		},
		Args: &kaiv1alpha1.NodePoolControllerArgs{
			Debug:                                  nodePoolControllerDebug(config),
			MetricsNamespace:                       ptr.To(clustercommonconstants.NamespaceRunAi),
			DcgmExporterNamespace:                  ptr.To(dcgmExporterNamespace(ctx, k8sReader)),
			SchedulingShardArgs:                    ptr.To(shardArgs),
			ExcludedNodepoolName:                   ptr.To(clustercommonconstants.ExcludedNodepoolName),
			ManagedNodesConfigName:                 ptr.To(clustercommonconstants.ManagedNodesConfigName),
			ToExcludeLabel:                         ptr.To(clustercommonconstants.ShouldBeExcludedLabelKey),
			UnschedulableLabel:                     ptr.To(clustercommonconstants.LabelMarkedUnschedulableByNodePoolController),
			GroveTopologyAnnotation:                ptr.To(clustercommonconstants.GroveTopologyAnnotation),
			GroveTopologyResourceVersionAnnotation: ptr.To(clustercommonconstants.GroveTopologyResourceVersionAnnotation),
		},
		// None of these are modelled on KRMConfig. The worker-node label keys are
		// the controller's own defaults in KAI vocabulary, and the uninstall
		// detection reference has no equivalent field at all.
		ExtraArgs: []string{
			fmt.Sprintf("--scheduler-namespace=%s", clustercommonconstants.NamespaceRunAi),
			"--cpu-worker-node-label-key=node-role.kubernetes.io/runai-cpu-worker",
			"--gpu-worker-node-label-key=node-role.kubernetes.io/runai-gpu-worker",
			"--mig-worker-node-label-key=node-role.kubernetes.io/runai-mig-enabled",
			"--uninstall-detection-ref=run.ai/v1/Cluster//cluster",
		},
		Replicas: config.Replicas,
		VPA:      convertVPA(config.VPA),
	}, nil
}

// metricsPortMapping is the port both ServiceMonitors scrape, so its name has to
// match what the endpoints refer to.
func metricsPortMapping(metrics *nodepool_controller.Metrics) *kaiv1alpha1.PortMapping {
	if metrics == nil {
		return nil
	}
	return &kaiv1alpha1.PortMapping{
		Port:       ptr.To(metrics.Port),
		TargetPort: ptr.To(metrics.Port),
		Name:       ptr.To(metrics.PortName),
	}
}

func nodePoolMetricsPortMapping(service *nodepool_controller.Service) *kaiv1alpha1.PortMapping {
	if service == nil {
		return nil
	}
	mapping := portMapping(service.Port, service.TargetPort)
	if mapping == nil {
		return nil
	}
	mapping.Name = ptr.To(nodePoolMetricsPortName)
	return mapping
}

func nodePoolControllerDebug(config *nodepool_controller.NodePoolController) *bool {
	if config.Args == nil {
		return nil
	}
	return ptr.To(config.Args.DebugLog)
}

// dcgmExporterNamespace resolves the exporter's namespace from the cluster, as
// the operand does; an empty string means it was not found, which is what the
// controller's own default expects.
func dcgmExporterNamespace(ctx context.Context, k8sReader client.Reader) string {
	dcgmExporterService, err := controller_utils.GetDCGMExporterService(ctx, k8sReader)
	if err != nil {
		return ""
	}
	return dcgmExporterService.Namespace
}
