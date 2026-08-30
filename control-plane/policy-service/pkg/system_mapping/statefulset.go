package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

// BuildSystemAppsStatefulSetV1Mapping is the system seed for apps/StatefulSet/v1
// (spec selectors without scope).
func BuildSystemAppsStatefulSetV1Mapping() (api.Gvk, map[string]api.SpecSelector) {
	return api.Gvk{Group: convutils.Ref("apps"), Kind: "StatefulSet", Version: "v1"}, buildAppsStatefulSetV1SpecSelectors()
}

func buildAppsStatefulSetV1SpecSelectors() map[string]api.SpecSelector {
	return map[string]api.SpecSelector{
		"workloadSpec": {
			Description: convutils.Ref("the workload spec (above the pod template)"),
			Query: `(
			  .spec
			)`,
			FieldGroups: []string{SystemStatefulSetSpecFgName, SystemSelectorFgName},
			Dependency: &api.SpecSelectorDependency{
				Path: "spec",
			},
		},

		"workloadMetadata": {
			Description: convutils.Ref("the workload metadata (annotations and labels)"),
			Query: `(
			  .metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
			Dependency: &api.SpecSelectorDependency{
				Path: "metadata",
			},
		},

		"podSpec": {
			Description: convutils.Ref("the pod spec under .spec.template.spec"),
			Query: `(
			  .spec.template.spec
			)`,
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("workloadSpec"),
				Path:         "template.spec",
			},
		},

		"podMetadata": {
			Description: convutils.Ref("the pod template metadata (annotations and labels) under .spec.template.metadata"),
			Query: `(
			  .spec.template.metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("workloadSpec"),
				Path:         "template.metadata",
			},
		},

		"allContainers": {
			Description: convutils.Ref("all app and init containers"),
			Query: `(
			  .spec.template.spec.containers[]?,
			  .spec.template.spec.initContainers[]?
			)
			| select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"allInitContainers": {
			Description: convutils.Ref("all init containers"),
			Query: `(
			  .spec.template.spec.initContainers[]?
			)
			| select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"allAppContainers": {
			Description: convutils.Ref("all app containers"),
			Query: `(
			  .spec.template.spec.containers[]?
			)
			| select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"appContainerByName": {
			Description: convutils.Ref("app container with a given name"),
			Query: `(
			  .spec.template.spec.containers[]?
			)
			| select(.name == $ARGS.named.containerName)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"containerName": api.MAPPINGFIELDTYPE_STRING,
			},
			FieldGroups: []string{
				SystemPodSpecContainerFgName,
				SystemPodSpecAndContainerFgName,
			},
		},

		"initContainerByName": {
			Description: convutils.Ref("init container with a given name"),
			Query: `(
			  (.spec.template.spec.initContainers // [])[]
			)
			| select(.name == $ARGS.named.containerName)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"containerName": api.MAPPINGFIELDTYPE_STRING,
			},
			FieldGroups: []string{
				SystemPodSpecContainerFgName,
				SystemPodSpecAndContainerFgName,
			},
		},
	}
}
