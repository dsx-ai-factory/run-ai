package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

// BuildSystemKubeflowOrgJAXJobV1Mapping is the system seed for kubeflow.org/JAXJob/v1.
//
// JAXJob has replica type (Worker) under .spec.jaxReplicaSpecs.
// Each replica type has its own pod template.
//
// Spec selector hierarchy:
//   - workloadSpec/workloadMetadata — the JAXJob itself (.spec, .metadata)
//   - allReplicaSpecs, workerReplicaSpec — replica-level
//   - allPodSpecs, workerPodSpec — pod templates across/per replica type
//   - allContainers, containersByReplicaType — containers across/per replica type
func BuildSystemKubeflowOrgJAXJobV1Mapping() (api.Gvk, map[string]api.SpecSelector) {
	return api.Gvk{Group: convutils.Ref("kubeflow.org"), Kind: "JAXJob", Version: "v1"}, buildKubeflowOrgJAXJobV1SpecSelectors()
}

func buildKubeflowOrgJAXJobV1SpecSelectors() map[string]api.SpecSelector {
	return map[string]api.SpecSelector{
		// ===== Workload Level =====
		"workloadSpec": {
			Description: convutils.Ref("the JAXJob spec (runPolicy and jaxReplicaSpecs)"),
			Query: `(
			  .spec
			)`,
			FieldGroups: []string{SystemKubeflowRunPolicyFgName},
			Dependency: &api.SpecSelectorDependency{
				Path: "spec",
			},
		},

		"workloadMetadata": {
			Description: convutils.Ref("the JAXJob metadata (annotations and labels)"),
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
			Description: convutils.Ref("all replica specs (Worker) — iterates each replica type's spec"),
			Query: `(
			  .spec.jaxReplicaSpecs[]
			)`,
			FieldGroups: []string{SystemKubeflowReplicaSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("workloadSpec"),
				Path:         "jaxReplicaSpecs",
			},
		},

		"workerReplicaSpec": {
			Description: convutils.Ref("the Worker replica spec"),
			Query: `(
			  .spec.jaxReplicaSpecs.Worker
			)`,
			FieldGroups: []string{SystemKubeflowReplicaSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("workloadSpec"),
				Path:         "jaxReplicaSpecs.Worker",
			},
		},

		"replicaSpecByType": {
			Description: convutils.Ref("replica spec for a given replica type (Worker)"),
			Query: `(
			  .spec.jaxReplicaSpecs[$ARGS.named.replicaType]
			)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"replicaType": api.MAPPINGFIELDTYPE_STRING,
			},
			ArgsEnum: map[string][]string{
				"replicaType": {"Worker"},
			},
			FieldGroups: []string{SystemKubeflowReplicaSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("workloadSpec"),
				Path:         "jaxReplicaSpecs",
			},
		},

		// ===== Pod Level (all replicas) =====
		"allPodSpecs": {
			Description: convutils.Ref("pod specs across all replica types"),
			Query: `(
			  .spec.jaxReplicaSpecs[].template.spec
			)`,
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
		},

		"allPodMetadata": {
			Description: convutils.Ref("pod metadata (labels, annotations) across all replica types"),
			Query: `(
			  .spec.jaxReplicaSpecs[].template.metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
		},

		// ===== Pod Level (per replica type shortcuts) =====
		"workerPodSpec": {
			Description: convutils.Ref("the Worker replica's pod spec"),
			Query: `(
			  .spec.jaxReplicaSpecs.Worker.template.spec
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
			  .spec.jaxReplicaSpecs.Worker.template.metadata
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
			  .spec.jaxReplicaSpecs[$ARGS.named.replicaType].template.spec
			)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"replicaType": api.MAPPINGFIELDTYPE_STRING,
			},
			ArgsEnum: map[string][]string{
				"replicaType": {"Worker"},
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
			  .spec.jaxReplicaSpecs[$ARGS.named.replicaType].template.metadata
			)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"replicaType": api.MAPPINGFIELDTYPE_STRING,
			},
			ArgsEnum: map[string][]string{
				"replicaType": {"Worker"},
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
			  .spec.jaxReplicaSpecs[].template.spec |
			  (.containers[]?, .initContainers[]?)
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"allAppContainers": {
			Description: convutils.Ref("all app containers across all replica types"),
			Query: `(
			  .spec.jaxReplicaSpecs[].template.spec.containers[]?
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"allInitContainers": {
			Description: convutils.Ref("all init containers across all replica types"),
			Query: `(
			  .spec.jaxReplicaSpecs[].template.spec |
			  (.initContainers // [])[]
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		// ===== Container Level (per replica type) =====
		"containersByReplicaType": {
			Description: convutils.Ref("all containers (app and init) for a given replica type"),
			Query: `(
			  .spec.jaxReplicaSpecs[$ARGS.named.replicaType].template.spec |
			  (.containers[]?, .initContainers[]?)
			) | select(.)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"replicaType": api.MAPPINGFIELDTYPE_STRING,
			},
			ArgsEnum: map[string][]string{
				"replicaType": {"Worker"},
			},
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"appContainersByReplicaType": {
			Description: convutils.Ref("app containers for a given replica type"),
			Query: `(
			  .spec.jaxReplicaSpecs[$ARGS.named.replicaType].template.spec.containers[]?
			) | select(.)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"replicaType": api.MAPPINGFIELDTYPE_STRING,
			},
			ArgsEnum: map[string][]string{
				"replicaType": {"Worker"},
			},
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"initContainersByReplicaType": {
			Description: convutils.Ref("init containers for a given replica type"),
			Query: `(
			  .spec.jaxReplicaSpecs[$ARGS.named.replicaType].template.spec |
			  (.initContainers // [])[]
			) | select(.)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"replicaType": api.MAPPINGFIELDTYPE_STRING,
			},
			ArgsEnum: map[string][]string{
				"replicaType": {"Worker"},
			},
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		// ===== Container Level (by replica type and name) =====
		"appContainerByReplicaTypeAndName": {
			Description: convutils.Ref("app container with a given name for a given replica type"),
			Query: `(
			  .spec.jaxReplicaSpecs[$ARGS.named.replicaType].template.spec.containers[]?
			) | select(.name == $ARGS.named.containerName)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"replicaType":   api.MAPPINGFIELDTYPE_STRING,
				"containerName": api.MAPPINGFIELDTYPE_STRING,
			},
			ArgsEnum: map[string][]string{
				"replicaType": {"Worker"},
			},
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"initContainerByReplicaTypeAndName": {
			Description: convutils.Ref("init container with a given name for a given replica type"),
			Query: `(
			  (.spec.jaxReplicaSpecs[$ARGS.named.replicaType].template.spec.initContainers // [])[]
			) | select(.name == $ARGS.named.containerName)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"replicaType":   api.MAPPINGFIELDTYPE_STRING,
				"containerName": api.MAPPINGFIELDTYPE_STRING,
			},
			ArgsEnum: map[string][]string{
				"replicaType": {"Worker"},
			},
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
	}
}
