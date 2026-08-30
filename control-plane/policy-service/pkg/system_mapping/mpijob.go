package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

// BuildSystemKubeflowOrgMPIJobV2beta1Mapping is the system seed for kubeflow.org/MPIJob/v2beta1.
//
// MPIJob has a structure similar to PyTorchJob with replica types (Launcher, Worker)
// under .spec.mpiReplicaSpecs. Each replica type has its own pod template.
//
// Spec selector hierarchy:
//   - workloadSpec/workloadMetadata — the MPIJob itself (.spec, .metadata)
//   - allReplicaSpecs, launcherReplicaSpec, workerReplicaSpec — replica-level
//   - allPodSpecs, launcherPodSpec, workerPodSpec — pod templates across/per replica type
//   - allContainers, containersByReplicaType — containers across/per replica type
func BuildSystemKubeflowOrgMPIJobV2beta1Mapping() (api.Gvk, map[string]api.SpecSelector) {
	return api.Gvk{Group: convutils.Ref("kubeflow.org"), Kind: "MPIJob", Version: "v2beta1"}, buildKubeflowOrgMPIJobV2beta1SpecSelectors()
}

func buildKubeflowOrgMPIJobV2beta1SpecSelectors() map[string]api.SpecSelector {
	return map[string]api.SpecSelector{
		// ===== Workload Level =====
		"workloadSpec": {
			Description: convutils.Ref("the MPIJob spec (runPolicy, slotsPerWorker, and mpiReplicaSpecs)"),
			Query: `(
			  .spec
			)`,
			FieldGroups: []string{SystemKubeflowRunPolicyFgName, SystemMPIJobSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				Path: "spec",
			},
		},

		"workloadMetadata": {
			Description: convutils.Ref("the MPIJob metadata (annotations and labels)"),
			Query: `(
			  .metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
			Dependency: &api.SpecSelectorDependency{
				Path: "metadata",
			},
		},

		// ===== Replica Level =====
		"allReplicaSpecs": {
			Description: convutils.Ref("all replica specs (Launcher, Worker) — iterates each replica type's spec"),
			Query: `(
			  .spec.mpiReplicaSpecs[]
			)`,
			FieldGroups: []string{SystemKubeflowReplicaSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("workloadSpec"),
				Path:         "mpiReplicaSpecs",
			},
		},

		"launcherReplicaSpec": {
			Description: convutils.Ref("the Launcher replica spec"),
			Query: `(
			  .spec.mpiReplicaSpecs.Launcher
			)`,
			FieldGroups: []string{SystemKubeflowReplicaSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("workloadSpec"),
				Path:         "mpiReplicaSpecs.Launcher",
			},
		},

		"workerReplicaSpec": {
			Description: convutils.Ref("the Worker replica spec"),
			Query: `(
			  .spec.mpiReplicaSpecs.Worker
			)`,
			FieldGroups: []string{SystemKubeflowReplicaSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("workloadSpec"),
				Path:         "mpiReplicaSpecs.Worker",
			},
		},

		"replicaSpecByType": {
			Description: convutils.Ref("replica spec for a given replica type (Launcher or Worker)"),
			Query: `(
			  .spec.mpiReplicaSpecs[$ARGS.named.replicaType]
			)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"replicaType": api.MAPPINGFIELDTYPE_STRING,
			},
			ArgsEnum: map[string][]string{
				"replicaType": {"Launcher", "Worker"},
			},
			FieldGroups: []string{SystemKubeflowReplicaSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("workloadSpec"),
				Path:         "mpiReplicaSpecs",
			},
		},

		// ===== Pod Level (all replicas) =====
		"allPodSpecs": {
			Description: convutils.Ref("pod specs across all replica types"),
			Query: `(
			  .spec.mpiReplicaSpecs[].template.spec
			)`,
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
		},

		"allPodMetadata": {
			Description: convutils.Ref("pod metadata (labels, annotations) across all replica types"),
			Query: `(
			  .spec.mpiReplicaSpecs[].template.metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
		},

		// ===== Pod Level (per replica type shortcuts) =====
		"launcherPodSpec": {
			Description: convutils.Ref("the Launcher replica's pod spec"),
			Query: `(
			  .spec.mpiReplicaSpecs.Launcher.template.spec
			)`,
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("launcherReplicaSpec"),
				Path:         "template.spec",
			},
		},

		"launcherPodMetadata": {
			Description: convutils.Ref("the Launcher replica's pod metadata (labels, annotations)"),
			Query: `(
			  .spec.mpiReplicaSpecs.Launcher.template.metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("launcherReplicaSpec"),
				Path:         "template.metadata",
			},
		},

		"workerPodSpec": {
			Description: convutils.Ref("the Worker replica's pod spec"),
			Query: `(
			  .spec.mpiReplicaSpecs.Worker.template.spec
			)`,
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("workerReplicaSpec"),
				Path:         "template.spec",
			},
		},

		"workerPodMetadata": {
			Description: convutils.Ref("the Worker replica's pod metadata (labels, annotations)"),
			Query: `(
			  .spec.mpiReplicaSpecs.Worker.template.metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("workerReplicaSpec"),
				Path:         "template.metadata",
			},
		},

		// ===== Pod Level (parameterized by replica type) =====
		"podSpecByReplicaType": {
			Description: convutils.Ref("pod spec for a given replica type"),
			Query: `(
			  .spec.mpiReplicaSpecs[$ARGS.named.replicaType].template.spec
			)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"replicaType": api.MAPPINGFIELDTYPE_STRING,
			},
			ArgsEnum: map[string][]string{
				"replicaType": {"Launcher", "Worker"},
			},
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("replicaSpecByType"),
				Path:         "template.spec",
			},
		},

		"podMetadataByReplicaType": {
			Description: convutils.Ref("pod metadata for a given replica type"),
			Query: `(
			  .spec.mpiReplicaSpecs[$ARGS.named.replicaType].template.metadata
			)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"replicaType": api.MAPPINGFIELDTYPE_STRING,
			},
			ArgsEnum: map[string][]string{
				"replicaType": {"Launcher", "Worker"},
			},
			FieldGroups: []string{SystemMetadataFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("replicaSpecByType"),
				Path:         "template.metadata",
			},
		},

		// ===== Container Level (all replicas) =====
		"allContainers": {
			Description: convutils.Ref("all containers (app and init) across all replica types"),
			Query: `(
			  .spec.mpiReplicaSpecs[].template.spec |
			  (.containers[]?, .initContainers[]?)
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"allAppContainers": {
			Description: convutils.Ref("all app containers across all replica types"),
			Query: `(
			  .spec.mpiReplicaSpecs[].template.spec.containers[]?
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"allInitContainers": {
			Description: convutils.Ref("all init containers across all replica types"),
			Query: `(
			  .spec.mpiReplicaSpecs[].template.spec |
			  (.initContainers // [])[]
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		// ===== Container Level (per replica type) =====
		"containersByReplicaType": {
			Description: convutils.Ref("all containers (app and init) for a given replica type"),
			Query: `(
			  .spec.mpiReplicaSpecs[$ARGS.named.replicaType].template.spec |
			  (.containers[]?, .initContainers[]?)
			) | select(.)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"replicaType": api.MAPPINGFIELDTYPE_STRING,
			},
			ArgsEnum: map[string][]string{
				"replicaType": {"Launcher", "Worker"},
			},
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"appContainersByReplicaType": {
			Description: convutils.Ref("app containers for a given replica type"),
			Query: `(
			  .spec.mpiReplicaSpecs[$ARGS.named.replicaType].template.spec.containers[]?
			) | select(.)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"replicaType": api.MAPPINGFIELDTYPE_STRING,
			},
			ArgsEnum: map[string][]string{
				"replicaType": {"Launcher", "Worker"},
			},
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"initContainersByReplicaType": {
			Description: convutils.Ref("init containers for a given replica type"),
			Query: `(
			  .spec.mpiReplicaSpecs[$ARGS.named.replicaType].template.spec |
			  (.initContainers // [])[]
			) | select(.)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"replicaType": api.MAPPINGFIELDTYPE_STRING,
			},
			ArgsEnum: map[string][]string{
				"replicaType": {"Launcher", "Worker"},
			},
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		// ===== Container Level (by replica type and name) =====
		"appContainerByReplicaTypeAndName": {
			Description: convutils.Ref("app container with a given name for a given replica type"),
			Query: `(
			  .spec.mpiReplicaSpecs[$ARGS.named.replicaType].template.spec.containers[]?
			) | select(.name == $ARGS.named.containerName)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"replicaType":   api.MAPPINGFIELDTYPE_STRING,
				"containerName": api.MAPPINGFIELDTYPE_STRING,
			},
			ArgsEnum: map[string][]string{
				"replicaType": {"Launcher", "Worker"},
			},
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"initContainerByReplicaTypeAndName": {
			Description: convutils.Ref("init container with a given name for a given replica type"),
			Query: `(
			  (.spec.mpiReplicaSpecs[$ARGS.named.replicaType].template.spec.initContainers // [])[]
			) | select(.name == $ARGS.named.containerName)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"replicaType":   api.MAPPINGFIELDTYPE_STRING,
				"containerName": api.MAPPINGFIELDTYPE_STRING,
			},
			ArgsEnum: map[string][]string{
				"replicaType": {"Launcher", "Worker"},
			},
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
	}
}
