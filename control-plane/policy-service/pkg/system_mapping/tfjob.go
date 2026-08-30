package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

// BuildSystemKubeflowOrgTFJobV1Mapping is the system seed for kubeflow.org/TFJob/v1.
//
// TFJob has replica types (Chief, Worker, PS, Evaluator) under .spec.tfReplicaSpecs.
// Each replica type has its own pod template.
//
// Spec selector hierarchy:
//   - workloadSpec/workloadMetadata — the TFJob itself (.spec, .metadata)
//   - allReplicaSpecs, chiefReplicaSpec, workerReplicaSpec, psReplicaSpec, evaluatorReplicaSpec — replica-level
//   - allPodSpecs, *PodSpec — pod templates across/per replica type
//   - allContainers, containersByReplicaType — containers across/per replica type
func BuildSystemKubeflowOrgTFJobV1Mapping() (api.Gvk, map[string]api.SpecSelector) {
	return api.Gvk{Group: convutils.Ref("kubeflow.org"), Kind: "TFJob", Version: "v1"}, buildKubeflowOrgTFJobV1SpecSelectors()
}

func buildKubeflowOrgTFJobV1SpecSelectors() map[string]api.SpecSelector {
	return map[string]api.SpecSelector{
		// ===== Workload Level =====
		"workloadSpec": {
			Description: convutils.Ref("the TFJob spec (runPolicy and tfReplicaSpecs)"),
			Query: `(
			  .spec
			)`,
			FieldGroups: []string{SystemKubeflowRunPolicyFgName},
			Dependency: &api.SpecSelectorDependency{
				Path: "spec",
			},
		},

		"workloadMetadata": {
			Description: convutils.Ref("the TFJob metadata (annotations and labels)"),
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
			Description: convutils.Ref("all replica specs (Chief, Worker, PS, Evaluator) — iterates each replica type's spec"),
			Query: `(
			  .spec.tfReplicaSpecs[]
			)`,
			FieldGroups: []string{SystemKubeflowReplicaSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("workloadSpec"),
				Path:         "tfReplicaSpecs",
			},
		},

		"chiefReplicaSpec": {
			Description: convutils.Ref("the Chief replica spec"),
			Query: `(
			  .spec.tfReplicaSpecs.Chief
			)`,
			FieldGroups: []string{SystemKubeflowReplicaSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("workloadSpec"),
				Path:         "tfReplicaSpecs.Chief",
			},
		},

		"workerReplicaSpec": {
			Description: convutils.Ref("the Worker replica spec"),
			Query: `(
			  .spec.tfReplicaSpecs.Worker
			)`,
			FieldGroups: []string{SystemKubeflowReplicaSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("workloadSpec"),
				Path:         "tfReplicaSpecs.Worker",
			},
		},

		"psReplicaSpec": {
			Description: convutils.Ref("the PS (Parameter Server) replica spec"),
			Query: `(
			  .spec.tfReplicaSpecs.PS
			)`,
			FieldGroups: []string{SystemKubeflowReplicaSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("workloadSpec"),
				Path:         "tfReplicaSpecs.PS",
			},
		},

		"evaluatorReplicaSpec": {
			Description: convutils.Ref("the Evaluator replica spec"),
			Query: `(
			  .spec.tfReplicaSpecs.Evaluator
			)`,
			FieldGroups: []string{SystemKubeflowReplicaSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("workloadSpec"),
				Path:         "tfReplicaSpecs.Evaluator",
			},
		},

		"replicaSpecByType": {
			Description: convutils.Ref("replica spec for a given replica type"),
			Query: `(
			  .spec.tfReplicaSpecs[$ARGS.named.replicaType]
			)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"replicaType": api.MAPPINGFIELDTYPE_STRING,
			},
			ArgsEnum: map[string][]string{
				"replicaType": {"Chief", "Worker", "PS", "Evaluator"},
			},
			FieldGroups: []string{SystemKubeflowReplicaSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("workloadSpec"),
				Path:         "tfReplicaSpecs",
			},
		},

		// ===== Pod Level (all replicas) =====
		"allPodSpecs": {
			Description: convutils.Ref("pod specs across all replica types"),
			Query: `(
			  .spec.tfReplicaSpecs[].template.spec
			)`,
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
		},

		"allPodMetadata": {
			Description: convutils.Ref("pod metadata (labels, annotations) across all replica types"),
			Query: `(
			  .spec.tfReplicaSpecs[].template.metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
		},

		// ===== Pod Level (per replica type shortcuts) =====
		"chiefPodSpec": {
			Description: convutils.Ref("the Chief replica's pod spec"),
			Query: `(
			  .spec.tfReplicaSpecs.Chief.template.spec
			)`,
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("chiefReplicaSpec"),
				Path:         "template.spec",
			},
		},

		"chiefPodMetadata": {
			Description: convutils.Ref("the Chief replica's pod metadata (labels, annotations)"),
			Query: `(
			  .spec.tfReplicaSpecs.Chief.template.metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("chiefReplicaSpec"),
				Path:         "template.metadata",
			},
		},

		"workerPodSpec": {
			Description: convutils.Ref("the Worker replica's pod spec"),
			Query: `(
			  .spec.tfReplicaSpecs.Worker.template.spec
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
			  .spec.tfReplicaSpecs.Worker.template.metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("workerReplicaSpec"),
				Path:         "template.metadata",
			},
		},

		"psPodSpec": {
			Description: convutils.Ref("the PS replica's pod spec"),
			Query: `(
			  .spec.tfReplicaSpecs.PS.template.spec
			)`,
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("psReplicaSpec"),
				Path:         "template.spec",
			},
		},

		"psPodMetadata": {
			Description: convutils.Ref("the PS replica's pod metadata (labels, annotations)"),
			Query: `(
			  .spec.tfReplicaSpecs.PS.template.metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("psReplicaSpec"),
				Path:         "template.metadata",
			},
		},

		"evaluatorPodSpec": {
			Description: convutils.Ref("the Evaluator replica's pod spec"),
			Query: `(
			  .spec.tfReplicaSpecs.Evaluator.template.spec
			)`,
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("evaluatorReplicaSpec"),
				Path:         "template.spec",
			},
		},

		"evaluatorPodMetadata": {
			Description: convutils.Ref("the Evaluator replica's pod metadata (labels, annotations)"),
			Query: `(
			  .spec.tfReplicaSpecs.Evaluator.template.metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("evaluatorReplicaSpec"),
				Path:         "template.metadata",
			},
		},

		// ===== Pod Level (parameterized by replica type) =====
		"podSpecByReplicaType": {
			Description: convutils.Ref("pod spec for a given replica type"),
			Query: `(
			  .spec.tfReplicaSpecs[$ARGS.named.replicaType].template.spec
			)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"replicaType": api.MAPPINGFIELDTYPE_STRING,
			},
			ArgsEnum: map[string][]string{
				"replicaType": {"Chief", "Worker", "PS", "Evaluator"},
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
			  .spec.tfReplicaSpecs[$ARGS.named.replicaType].template.metadata
			)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"replicaType": api.MAPPINGFIELDTYPE_STRING,
			},
			ArgsEnum: map[string][]string{
				"replicaType": {"Chief", "Worker", "PS", "Evaluator"},
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
			  .spec.tfReplicaSpecs[].template.spec |
			  (.containers[]?, .initContainers[]?)
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"allAppContainers": {
			Description: convutils.Ref("all app containers across all replica types"),
			Query: `(
			  .spec.tfReplicaSpecs[].template.spec.containers[]?
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"allInitContainers": {
			Description: convutils.Ref("all init containers across all replica types"),
			Query: `(
			  .spec.tfReplicaSpecs[].template.spec |
			  (.initContainers // [])[]
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		// ===== Container Level (per replica type) =====
		"containersByReplicaType": {
			Description: convutils.Ref("all containers (app and init) for a given replica type"),
			Query: `(
			  .spec.tfReplicaSpecs[$ARGS.named.replicaType].template.spec |
			  (.containers[]?, .initContainers[]?)
			) | select(.)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"replicaType": api.MAPPINGFIELDTYPE_STRING,
			},
			ArgsEnum: map[string][]string{
				"replicaType": {"Chief", "Worker", "PS", "Evaluator"},
			},
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"appContainersByReplicaType": {
			Description: convutils.Ref("app containers for a given replica type"),
			Query: `(
			  .spec.tfReplicaSpecs[$ARGS.named.replicaType].template.spec.containers[]?
			) | select(.)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"replicaType": api.MAPPINGFIELDTYPE_STRING,
			},
			ArgsEnum: map[string][]string{
				"replicaType": {"Chief", "Worker", "PS", "Evaluator"},
			},
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"initContainersByReplicaType": {
			Description: convutils.Ref("init containers for a given replica type"),
			Query: `(
			  .spec.tfReplicaSpecs[$ARGS.named.replicaType].template.spec |
			  (.initContainers // [])[]
			) | select(.)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"replicaType": api.MAPPINGFIELDTYPE_STRING,
			},
			ArgsEnum: map[string][]string{
				"replicaType": {"Chief", "Worker", "PS", "Evaluator"},
			},
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		// ===== Container Level (by replica type and name) =====
		"appContainerByReplicaTypeAndName": {
			Description: convutils.Ref("app container with a given name for a given replica type"),
			Query: `(
			  .spec.tfReplicaSpecs[$ARGS.named.replicaType].template.spec.containers[]?
			) | select(.name == $ARGS.named.containerName)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"replicaType":   api.MAPPINGFIELDTYPE_STRING,
				"containerName": api.MAPPINGFIELDTYPE_STRING,
			},
			ArgsEnum: map[string][]string{
				"replicaType": {"Chief", "Worker", "PS", "Evaluator"},
			},
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"initContainerByReplicaTypeAndName": {
			Description: convutils.Ref("init container with a given name for a given replica type"),
			Query: `(
			  (.spec.tfReplicaSpecs[$ARGS.named.replicaType].template.spec.initContainers // [])[]
			) | select(.name == $ARGS.named.containerName)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"replicaType":   api.MAPPINGFIELDTYPE_STRING,
				"containerName": api.MAPPINGFIELDTYPE_STRING,
			},
			ArgsEnum: map[string][]string{
				"replicaType": {"Chief", "Worker", "PS", "Evaluator"},
			},
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
	}
}
