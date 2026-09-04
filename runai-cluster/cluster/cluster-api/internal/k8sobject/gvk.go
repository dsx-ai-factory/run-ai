/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package k8sobject

import (
	"fmt"

	workloadsapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	runAIGroup = "run.ai"
	v2alpha1   = "v2alpha1"

	kubeflowGroup = "kubeflow.org"
)

var podGVK = schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"}

// Types not in this map are treated as External — Group/Version come from
// kubernetesRef and Type itself is the Kind.
var nativeWorkloadGVKs = map[string]schema.GroupVersionKind{
	"Training":             {Group: runAIGroup, Version: v2alpha1, Kind: "TrainingWorkload"},
	"Inference":            {Group: runAIGroup, Version: v2alpha1, Kind: "InferenceWorkload"},
	"Distributed":          {Group: runAIGroup, Version: v2alpha1, Kind: "DistributedWorkload"},
	"Workspace":            {Group: runAIGroup, Version: v2alpha1, Kind: "InteractiveWorkload"},
	"DistributedInference": {Group: runAIGroup, Version: "v1alpha1", Kind: "DistributedInferenceWorkload"},
}

func workloadGVK(workloadInfo *AuthorizedWorkloadInfo) (schema.GroupVersionKind, error) {
	// Transitioned v1->v2 workloads keep a native DB type but the real k8s object
	// is a batch/Job (Training/Workspace), a kubeflow job (Distributed), or a
	// Knative Service (Inference), so resolve the actual GVK instead of the
	// native run.ai CRD. DistributedInference is not migrated yet and falls through.
	if workloadInfo.SourceApi == string(workloadsapi.SOURCEAPI_WORKLOADS_V2) {
		switch workloadInfo.Kind {
		case "Training", "Workspace":
			return schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "Job"}, nil
		case "Distributed":
			if gvk, ok := distributedFrameworkGVK(workloadInfo.DistributedFramework); ok {
				return gvk, nil
			}
			// Unknown/empty framework: fall through to the native run.ai GVK.
		case "Inference":
			return schema.GroupVersionKind{Group: "serving.knative.dev", Version: "v1", Kind: "Service"}, nil
		}
	}
	if gvk, ok := nativeWorkloadGVKs[workloadInfo.Kind]; ok {
		return gvk, nil
	}
	if workloadInfo.Group == "" || workloadInfo.Version == "" {
		return schema.GroupVersionKind{}, fmt.Errorf("workload type %q is not native and kubernetesRef is empty", workloadInfo.Kind)
	}
	return schema.GroupVersionKind{Group: workloadInfo.Group, Version: workloadInfo.Version, Kind: workloadInfo.Kind}, nil
}

// distributedFrameworkGVK maps a Distributed workload's framework to its kubeflow
// GVK. The Kind is always <framework>Job (MPIJob/PyTorchJob/TFJob/XGBoostJob/JAXJob);
// MPIJob lives at kubeflow.org/v2beta1 while the rest are kubeflow.org/v1.
func distributedFrameworkGVK(framework string) (schema.GroupVersionKind, bool) {
	if framework == "" {
		return schema.GroupVersionKind{}, false
	}
	version := "v1"
	if framework == "MPI" {
		version = "v2beta1"
	}
	return schema.GroupVersionKind{Group: kubeflowGroup, Version: version, Kind: framework + "Job"}, true
}

// Elements are always CRDs (Pods don't appear at the element level), so a
// non-empty Group is required — unlike workloadGVK, this is not lenient.
func elementGvkName(ref *workloadsapi.WorkloadElementKubernetesRef) (schema.GroupVersionKind, string, bool) {
	name := ref.GetName()
	gvk := ref.GetGvk()
	group := gvk.GetGroup()
	version := gvk.GetVersion()
	kind := gvk.GetKind()
	if name == "" || group == "" || version == "" || kind == "" {
		return schema.GroupVersionKind{}, "", false
	}
	return schema.GroupVersionKind{Group: group, Version: version, Kind: kind}, name, true
}
