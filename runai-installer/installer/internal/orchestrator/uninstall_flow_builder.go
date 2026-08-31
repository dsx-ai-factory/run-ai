/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orchestrator

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	"github.com/run-ai/runai/runai-installer/installer/internal/flow"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/uninstall"
	"github.com/run-ai/runai/runai-installer/installer/internal/uninstall/steps"
)

func buildUninstallFlowGroups(cfg *config.UninstallConfig, deps uninstall.Deps, externalEvents chan *flow.FlowStatusReportingMsg) []*FlowGroup {
	helmAndNamespaceFlows := append(getClusterUninstallFlows(cfg, deps), getCPUninstallFlows(cfg, deps)...)
	helmAndNamespaceFlows = append(helmAndNamespaceFlows, getAdditionalNamespaceFlows(cfg, deps)...)
	stripCRFinalizersFlows := getStripRunAICRFinalizersFlows(cfg, deps)
	deleteCRDsFlows := getDeleteCRDsFlows(cfg, deps)
	deleteWebhooksFlows := getDeleteWebhooksFlows(cfg, deps)
	deleteResourceInterfacesFlows := getDeleteResourceInterfacesFlows(cfg, deps)
	deleteClusterRolesFlows := getDeleteClusterRolesFlows(cfg, deps)
	deletePriorityClassesFlows := getDeletePriorityClassesFlows(cfg, deps)

	// Helm releases + namespaces MUST come first. Helm owns CRDs, webhooks, and other
	// cluster-scoped resources as part of its release manifest. Deleting any of those
	// manually before calling helm uninstall would cause Helm to hit "not found" errors
	// when it tries to clean up its own resources, potentially failing the uninstall.
	// Helm uninstall also shuts down the running controllers, which is a prerequisite
	// for safely removing the CRDs they depend on.
	//
	// Finalizer stripping MUST happen after helm uninstall (controllers are gone) and
	// before CRD deletion. Deleting a CRD while any of its CRs still carry finalizers
	// leaves the CRD stuck in Terminating state: Kubernetes sets deletionTimestamp on
	// every CR but cannot complete deletion because the finalizer owner is already gone.
	//
	// The remaining groups have no strict ordering dependency between them:
	// - Webhook configurations are placed before CRDs for cleanliness, but the order
	//   does not matter. We rely on Kubernetes CRD cascade deletion to remove all CRD
	//   instances as a side effect of deleting the CRD itself, bypassing any registered
	//   validating/mutating webhooks that would block direct instance deletion (e.g.
	//   when the backing service is already gone after namespace deletion).
	// - ResourceInterfaces, ClusterRoles, and PriorityClasses are independent of each
	//   other and of the CRD/webhook groups.
	return []*FlowGroup{
		{
			Flows:   helmAndNamespaceFlows,
			Manager: flow.NewFlowManagerWithExternalEvents(helmAndNamespaceFlows, externalEvents, false),
		},
		{
			Flows:   stripCRFinalizersFlows,
			Manager: flow.NewFlowManagerWithExternalEvents(stripCRFinalizersFlows, externalEvents, false),
		},
		{
			Flows:   deleteWebhooksFlows,
			Manager: flow.NewFlowManagerWithExternalEvents(deleteWebhooksFlows, externalEvents, false),
		},
		{
			Flows:   deleteCRDsFlows,
			Manager: flow.NewFlowManagerWithExternalEvents(deleteCRDsFlows, externalEvents, false),
		},
		{
			Flows:   deleteResourceInterfacesFlows,
			Manager: flow.NewFlowManagerWithExternalEvents(deleteResourceInterfacesFlows, externalEvents, false),
		},
		{
			Flows:   deleteClusterRolesFlows,
			Manager: flow.NewFlowManagerWithExternalEvents(deleteClusterRolesFlows, externalEvents, false),
		},
		{
			Flows:   deletePriorityClassesFlows,
			Manager: flow.NewFlowManagerWithExternalEvents(deletePriorityClassesFlows, externalEvents, false),
		},
	}
}

func getDeletePriorityClassesFlows(cfg *config.UninstallConfig, deps uninstall.Deps) []*flow.InstallFlow {
	pcStep := steps.NewDeleteRunAIPriorityClasses(cfg, deps.K8sWrapper)
	return []*flow.InstallFlow{
		flow.NewInstallFlow("Delete RunAI PriorityClasses", []step.InstallStep{pcStep}, nil),
	}
}

func getDeleteClusterRolesFlows(cfg *config.UninstallConfig, deps uninstall.Deps) []*flow.InstallFlow {
	crStep := steps.NewDeleteRunAIClusterRoles(cfg, deps.K8sWrapper)
	return []*flow.InstallFlow{
		flow.NewInstallFlow("Delete RunAI ClusterRoles", []step.InstallStep{crStep}, nil),
	}
}

func getDeleteResourceInterfacesFlows(cfg *config.UninstallConfig, deps uninstall.Deps) []*flow.InstallFlow {
	riStep := steps.NewDeleteRunAIResourceInterfaces(cfg, deps.K8sWrapper)
	return []*flow.InstallFlow{
		flow.NewInstallFlow("Delete RunAI ResourceInterfaces", []step.InstallStep{riStep}, nil),
	}
}

func getDeleteWebhooksFlows(cfg *config.UninstallConfig, deps uninstall.Deps) []*flow.InstallFlow {
	mutatingStep := steps.NewDeleteRunAIMutatingWebhooks(cfg, deps.K8sWrapper)
	validatingStep := steps.NewDeleteRunAIValidatingWebhooks(cfg, deps.K8sWrapper)
	return []*flow.InstallFlow{
		flow.NewInstallFlow("Delete RunAI MutatingWebhookConfigurations", []step.InstallStep{mutatingStep}, nil),
		flow.NewInstallFlow("Delete RunAI ValidatingWebhookConfigurations", []step.InstallStep{validatingStep}, nil),
	}
}

func getStripRunAICRFinalizersFlows(cfg *config.UninstallConfig, deps uninstall.Deps) []*flow.InstallFlow {
	stripStep := steps.NewStripRunAICRFinalizers(cfg, deps.K8sWrapper)
	return []*flow.InstallFlow{
		flow.NewInstallFlow("Strip RunAI CR finalizers", []step.InstallStep{stripStep}, nil),
	}
}

func getDeleteCRDsFlows(cfg *config.UninstallConfig, deps uninstall.Deps) []*flow.InstallFlow {
	deleteCRDsStep := steps.NewDeleteRunAICRDs(cfg, deps.K8sWrapper)
	return []*flow.InstallFlow{
		flow.NewInstallFlow("Delete RunAI CRDs", []step.InstallStep{deleteCRDsStep}, nil),
	}
}

func getAdditionalNamespaceFlows(cfg *config.UninstallConfig, deps uninstall.Deps) []*flow.InstallFlow {
	var flows []*flow.InstallFlow
	for _, ns := range []string{step.ReservationNamespace, step.ScaleAdjustNamespace} {
		deleteNsStep := steps.NewDeleteNamespace(ns, cfg, deps.K8sWrapper)
		flows = append(flows, flow.NewInstallFlow("Delete namespace "+ns, []step.InstallStep{deleteNsStep}, nil))
	}
	return flows
}

func getClusterUninstallFlows(cfg *config.UninstallConfig, deps uninstall.Deps) []*flow.InstallFlow {
	clusterStep := steps.NewUninstallClusterChart(cfg, deps.HelmWrapper)
	deleteNsStep := steps.NewDeleteNamespace(step.ClusterNamespace, cfg, deps.K8sWrapper)
	clusterFlow := flow.NewInstallFlow("Uninstall Cluster", []step.InstallStep{clusterStep, deleteNsStep}, nil)
	clusterFlow.MarkStepTimeoutNonFatal(clusterStep)
	return []*flow.InstallFlow{clusterFlow}
}

func getCPUninstallFlows(cfg *config.UninstallConfig, deps uninstall.Deps) []*flow.InstallFlow {
	cpStep := steps.NewUninstallCPChart(cfg, deps.HelmWrapper)
	deleteNsStep := steps.NewDeleteNamespace(step.CPNamespace, cfg, deps.K8sWrapper)
	cpFlow := flow.NewInstallFlow("Uninstall Control Plane", []step.InstallStep{cpStep, deleteNsStep}, nil)
	cpFlow.MarkStepTimeoutNonFatal(cpStep)
	return []*flow.InstallFlow{cpFlow}
}
