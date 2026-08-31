/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// Note to dev: when adding support for a GVK, please open tasks for both scheduler team and suricata team:
	// the GVK needs to be supported both in pod-grouper and external-workload-integrator.
	//
	// cluster/external-workload-integrator/pkg/externalworkloads/external_workload_integrator.go:
	// supportedGVKsHandlers variable
	//      schema.GroupVersionKind{
	//			Group:   "GROUP",
	//			Version: "VERSION",
	//			Kind:    "KIND",
	//		}: unimplementedTopOwnerHandler,
	// + open task for suricata team.
	//
	// engine/pod-grouper/controllers/podgrouper/podgrouper.go:
	// supportedTypes variable
	//      metav1.GroupVersionKind{
	//			Group:   "GROUP",
	//			Version: "VERSION",
	//			Kind:    "KIND",
	//		}: plugins.GetPodGroupMetadata,
	// + open task for scheduler team.
	runaiExternalIntegrationSupportedGVKs = []metav1.GroupVersionKind{
		{
			Group:   "apps",
			Version: "v1",
			Kind:    "Deployment",
		},
		{
			Group:   "machinelearning.seldon.io",
			Version: "v1alpha2",
			Kind:    "SeldonDeployment",
		},
		{
			Group:   "machinelearning.seldon.io",
			Version: "v1",
			Kind:    "SeldonDeployment",
		},
		{
			Group:   "kubevirt.io",
			Version: "v1",
			Kind:    "VirtualMachineInstance",
		},
		{
			Group:   "kubeflow.org",
			Version: "v1",
			Kind:    "TFJob",
		},
		{
			Group:   "kubeflow.org",
			Version: "v1",
			Kind:    "PyTorchJob",
		},
		{
			Group:   "kubeflow.org",
			Version: "v1",
			Kind:    "XGBoostJob",
		},
		{
			Group:   "kubeflow.org",
			Version: "v1",
			Kind:    "JAXJob",
		},
		{
			Group:   "kubeflow.org",
			Version: "v1",
			Kind:    "MPIJob",
		},
		{
			Group:   "kubeflow.org",
			Version: "v2beta1",
			Kind:    "MPIJob",
		},
		{
			Group:   "kubeflow.org",
			Version: "v1beta1",
			Kind:    "Notebook",
		},
		{
			Group:   "batch",
			Version: "v1",
			Kind:    "Job",
		},
		{
			Group:   "apps",
			Version: "v1",
			Kind:    "StatefulSet",
		},
		{
			Group:   "apps",
			Version: "v1",
			Kind:    "ReplicaSet",
		},
		{
			Group:   "",
			Version: "v1",
			Kind:    "Pod",
		},
		{
			Group:   "amlarc.azureml.com",
			Version: "v1alpha1",
			Kind:    "AmlJob",
		},
		{
			Group:   "serving.knative.dev",
			Version: "v1",
			Kind:    "Service",
		},
		{
			Group:   "batch",
			Version: "v1",
			Kind:    "CronJob",
		},
		{
			Group:   "workspace.devfile.io",
			Version: "v1alpha2",
			Kind:    "DevWorkspace",
		},
		{
			Group:   "ray.io",
			Version: "v1alpha1",
			Kind:    "RayCluster",
		},
		{
			Group:   "ray.io",
			Version: "v1alpha1",
			Kind:    "RayJob",
		},
		{
			Group:   "ray.io",
			Version: "v1alpha1",
			Kind:    "RayService",
		},
		{
			Group:   "ray.io",
			Version: "v1",
			Kind:    "RayCluster",
		},
		{
			Group:   "ray.io",
			Version: "v1",
			Kind:    "RayJob",
		},
		{
			Group:   "ray.io",
			Version: "v1",
			Kind:    "RayService",
		},
		{
			Group:   "argoproj.io",
			Version: "v1alpha1",
			Kind:    "Workflow",
		},
		{
			Group:   "kubeflow.org",
			Version: "v1alpha1",
			Kind:    "ScheduledWorkflow",
		},
		{
			Group:   "tekton.dev",
			Version: "v1",
			Kind:    "TaskRun",
		},
		{
			Group:   "tekton.dev",
			Version: "v1",
			Kind:    "PipelineRun",
		},
		{
			Group:   "apps.nvidia.com",
			Version: "v1alpha1",
			Kind:    "NIMCache",
		},
		{
			Group:   "apps.nvidia.com",
			Version: "v1alpha1",
			Kind:    "NIMService",
		},
		{
			Group:   "egx.nvidia.io",
			Version: "v1",
			Kind:    "SPOTRequest",
		}, {
			Group:   "nvidia.com",
			Version: "v1alpha1",
			Kind:    "DynamoGraphDeployment",
		}, {
			Group:   "nvidia.com",
			Version: "v1beta1",
			Kind:    "DynamoGraphDeployment",
		},
	}

	runaiInternalSupportedGVKs = []metav1.GroupVersionKind{
		{
			Group:   "run.ai",
			Version: "v1",
			Kind:    "RunaiJob",
		},
	}

	runaiExternalIntegrationSupportedGVKsMap = map[metav1.GroupVersionKind]bool{}

	runaiInternalSupportedGVKsMap = map[metav1.GroupVersionKind]bool{}
)

func getRunaiSupportedGVKs() map[metav1.GroupVersionKind]bool {
	if len(runaiExternalIntegrationSupportedGVKsMap) == 0 {
		for _, runaiExternalIntegrationSupportedGVK := range runaiExternalIntegrationSupportedGVKs {
			runaiExternalIntegrationSupportedGVKsMap[runaiExternalIntegrationSupportedGVK] = true
		}
	}

	return runaiExternalIntegrationSupportedGVKsMap
}

func getRunaiInternalSupportedGVKs() map[metav1.GroupVersionKind]bool {
	if len(runaiInternalSupportedGVKsMap) == 0 {
		for _, runaiExternalIntegrationSupportedGVK := range runaiExternalIntegrationSupportedGVKs {
			runaiInternalSupportedGVKsMap[runaiExternalIntegrationSupportedGVK] = true
		}
		for _, runaiInternalSupportedGVK := range runaiInternalSupportedGVKs {
			runaiInternalSupportedGVKsMap[runaiInternalSupportedGVK] = true
		}
	}
	return runaiInternalSupportedGVKsMap
}

func getRightItemsMissingFromLeft[T, U any](
	left map[metav1.GroupVersionKind]T, right map[metav1.GroupVersionKind]U) []metav1.GroupVersionKind {
	rightItemsMissingFromLeft := []metav1.GroupVersionKind{}
	for gvk := range right {
		_, found := left[gvk]
		if !found {
			rightItemsMissingFromLeft = append(rightItemsMissingFromLeft, gvk)
		}
	}

	return rightItemsMissingFromLeft
}

func GetAllTypesThatNeedSupport() map[metav1.GroupVersionKind]bool {
	return getRunaiInternalSupportedGVKs()
}

/*
ValidateSupportedGVKs - call this func with the map of the objects you added as supported.
It will return a descriptive error if the GVKs don't match the original runaiExternalIntegrationSupportedGVKs list.
*/
func ValidateSupportedGVKs[T any](gvksToValidate map[metav1.GroupVersionKind]T, runaiExtendedSupport bool) error {
	gvksToSupport := map[metav1.GroupVersionKind]bool{}
	if runaiExtendedSupport {
		gvksToSupport = getRunaiInternalSupportedGVKs()
	} else {
		gvksToSupport = getRunaiSupportedGVKs()
	}

	unsupportedGVKs := getRightItemsMissingFromLeft(gvksToValidate, gvksToSupport)

	if len(unsupportedGVKs) == 0 && len(gvksToValidate) == len(gvksToSupport) {
		return nil
	}

	if len(unsupportedGVKs) > 0 {
		return fmt.Errorf("GVKs that require support and are missing: %+v", unsupportedGVKs)
	}

	supportedGVKsThatAreNotOnRunaiSupportList := getRightItemsMissingFromLeft(gvksToSupport, gvksToValidate)
	return fmt.Errorf(
		"GVKs that are supported but are missing from the required runai support list: %+v\n"+
			"Please add the relevant GVKs to common/supported-workloads/supported_workloads.go:runaiExternalIntegrationSupportedGVKs\n"+
			"The support implementation needs to be added to pod-grouper and to external-workload-integrator.",
		supportedGVKsThatAreNotOnRunaiSupportList)
}
