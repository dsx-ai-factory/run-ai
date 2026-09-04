/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"sort"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
)

// SystemFormatterFieldsBuilder builds the field map for a system formatter (e.g. runai-*).
type SystemFormatterFieldsBuilder func() map[string]api.FormatterField

// SystemFieldGroupBuilder returns field and item-selector maps without scope; patchFieldGroup adds MAPPINGSCOPE_SYSTEM.
type SystemFieldGroupBuilder func() (map[string]api.FieldGroupField, map[string]api.ItemSelector)

// SystemWorkloadMappingBuilder returns the workload GVK and spec selectors without scope; addScopeToSpecSelectors adds MAPPINGSCOPE_SYSTEM.
type SystemWorkloadMappingBuilder func() (api.Gvk, map[string]api.SpecSelector)

// Formatter names — custom JQ snippets that write one spec field, with
// optional enum validation. Add one only for enum fields (e.g. submissionMode);
// ordinary scalars (string, int, bool) use the built-in KeyValue formatter.
const (
	SystemPodSpecFormatterName               = "runai-podSpec"
	SystemJobSpecFormatterName               = "runai-jobSpec"
	SystemCronJobSpecFormatterName           = "runai-cronJobSpec"
	SystemSelectorFormatterName              = "runai-selector"
	SystemMetadataFormatterName              = "runai-metadata"
	SystemDeploymentSpecFormatterName        = "runai-deploymentSpec"
	SystemStatefulSetSpecFormatterName       = "runai-statefulSetSpec"
	SystemNimCacheSpecFormatterName          = "runai-nimCacheSpec"
	SystemNimServiceSpecFormatterName        = "runai-nimServiceSpec"
	SystemPyTorchJobSpecFormatterName        = "runai-pytorchJobSpec"
	SystemMPIJobSpecFormatterName            = "runai-mpiJobSpec"
	SystemKubeflowRunPolicyFormatterName     = "runai-kubeflowRunPolicy"
	SystemKubeflowReplicaSpecFormatterName   = "runai-kubeflowReplicaSpec"
	SystemDynamoGraphDeploymentFormatterName = "runai-dynamoGraphDeploymentSpec"
	SystemRayJobSpecFormatterName            = "runai-rayJobSpec"
	SystemRayServiceSpecFormatterName        = "runai-rayServiceSpec"
	SystemRayClusterSpecFormatterName        = "runai-rayClusterSpec"
	// runai-nativeSpec holds enum-enforced fields of Run.ai native workload specs
	// (WorkspaceSpecSpec, DistributedSpecSpec, MasterSpec) used by fictive GVKs
	// run.ai/Workspace/v1 and run.ai/Distributed/v1.
	SystemNativeSpecFormatterName = "runai-nativeSpec"
)

// Field group names — each names a set of policy-settable fields anchored
// at a SpecSelector. Shared groups (podSpec, podSpecContainer, metadata,
// selector) already cover pod-template, container, and metadata fields for
// every workload kind; a workload-specific group only adds the scalars
// unique to that workload (e.g. rayJobSpec adds entrypoint, runtimeEnvYAML,
// shutdownAfterJobFinishes, ...).
const (
	SystemPodSpecFgName             = "podSpec"
	SystemPodSpecContainerFgName    = "podSpecContainer"
	SystemPodSpecAndContainerFgName = "podSpecAndContainer"
	SystemJobSpecFgName             = "jobSpec"
	SystemCronJobSpecFgName         = "cronJobSpec"
	SystemSelectorFgName            = "selector"
	SystemDeploymentSpecFgName      = "deploymentSpec"
	SystemReplicaSetSpecFgName      = "replicaSetSpec"
	SystemStatefulSetSpecFgName     = "statefulSetSpec"
	SystemMetadataFgName            = "metadata"
	SystemNimCacheSpecFgName        = "nimCacheSpec"
	SystemNimCacheContainerFgName   = "nimCacheContainer"
	SystemNimServiceSpecFgName      = "nimServiceSpec"
	SystemNimServiceContainerFgName = "nimServiceContainer"
	SystemPyTorchJobSpecFgName      = "pytorchJobSpec"
	SystemMPIJobSpecFgName          = "mpiJobSpec"
	SystemKubeflowRunPolicyFgName   = "kubeflowRunPolicy"
	SystemKubeflowReplicaSpecFgName = "kubeflowReplicaSpec"
	SystemRayClusterSpecFgName      = "rayClusterSpec"
	SystemRayWorkerGroupSpecFgName  = "rayWorkerGroupSpec"
	SystemRayJobSpecFgName          = "rayJobSpec"
	SystemRayServiceSpecFgName      = "rayServiceSpec"
	// nativeSpec exposes the fields of Run.ai native workload specs
	// (WorkspaceSpecSpec, DistributedSpecSpec). The anchor is the spec object,
	// reached via a spec selector like "workspaceSpec" or "workerSpec".
	SystemNativeSpecFgName = "nativeSpec"
	// nativeDistributedSpec adds distributed-only fields (numWorkers, cleanPodPolicy,
	// distributedFramework, etc.) on top of nativeSpec.
	SystemNativeDistributedSpecFgName = "nativeDistributedSpec"
)

// SystemFgField is one map entry: the field name (key) plus formatter metadata.
// SpecRef is nil when the API spec path equals name; use convutils.Ref("") for an explicit empty spec ref.
type SystemFgField struct {
	Name          string
	Formatter     string
	SpecRef       *string
	Type          api.MappingFieldType
	ItemSelectors []string
}

type SystemItemSel struct {
	Name     string
	Desc     string
	Query    string
	ArgsInfo map[string]api.MappingFieldType
	ArgsEnum map[string][]string
}

// SortedKeys returns the keys of a map sorted in ascending order.
func SortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
