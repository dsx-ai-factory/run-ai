package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

// BuildSystemNvidiaComDynamoGraphDeploymentV1alpha1Mapping is the system seed for
// nvidia.com/DynamoGraphDeployment/v1alpha1.
//
// DynamoGraphDeployment has a unique structure:
//   - spec.services is a map (not array) where keys are dynamic service names
//   - Each service has extraPodSpec (full PodSpec) + convenience shortcuts at service level
//   - extraPodSpec is "stronger" and overrides service-level shortcuts
//
// Because service names are dynamic, we use parameterized spec selectors:
//   - allPodSpecs, allContainers — iterate all services
//   - servicePodSpec, serviceContainers — target a specific service by name parameter
//
// Spec selector hierarchy:
//   - workloadSpec/workloadMetadata — the DynamoGraphDeployment itself (.spec, .metadata)
//   - allPodSpecs — extraPodSpec across all services
//   - allContainers — containers/initContainers within extraPodSpec of all services
//   - servicePodSpec (parameterized) — extraPodSpec of a specific service
//   - serviceContainers (parameterized) — containers of a specific service
func BuildSystemNvidiaComDynamoGraphDeploymentV1alpha1Mapping() (api.Gvk, map[string]api.SpecSelector) {
	return api.Gvk{Group: convutils.Ref("nvidia.com"), Kind: "DynamoGraphDeployment", Version: "v1alpha1"}, buildDynamoGraphDeploymentV1alpha1SpecSelectors()
}

func buildDynamoGraphDeploymentV1alpha1SpecSelectors() map[string]api.SpecSelector {
	return map[string]api.SpecSelector{
		// ===== Workload Level =====
		"workloadSpec": {
			Description: convutils.Ref("the DynamoGraphDeployment spec (backendFramework, envs, pvcs, restart, services)"),
			Query: `(
			  .spec
			)`,
			FieldGroups: []string{SystemDynamoGraphDeploymentSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				Path: "spec",
			},
		},

		"workloadMetadata": {
			Description: convutils.Ref("the DynamoGraphDeployment metadata (annotations and labels)"),
			Query: `(
			  .metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
			Dependency: &api.SpecSelectorDependency{
				Path: "metadata",
			},
		},

		// ===== Service Level (all services) =====
		"allServiceSpecs": {
			Description: convutils.Ref("all service specs (iterate over spec.services map values)"),
			Query: `(
			  .spec.services | to_entries[] | .value
			)`,
			FieldGroups: []string{SystemDynamoServiceSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("workloadSpec"),
				Path:         "services",
			},
		},

		// ===== Service Level (by name) =====
		"serviceSpec": {
			Description: convutils.Ref("spec for a specific service by name"),
			Query: `(
			  .spec.services[$ARGS.named.serviceName]
			) | select(.)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"serviceName": api.MAPPINGFIELDTYPE_STRING,
			},
			FieldGroups: []string{SystemDynamoServiceSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("workloadSpec"),
				Path:         "services",
			},
		},

		// ===== Pod Level (all services via extraPodSpec) =====
		"allPodSpecs": {
			Description: convutils.Ref("extraPodSpec across all services (the actual PodSpec)"),
			Query: `(
			  .spec.services | to_entries[] | .value.extraPodSpec
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
		},

		"allPodMetadata": {
			Description: convutils.Ref("extraPodMetadata (labels, annotations) across all services"),
			Query: `(
			  .spec.services | to_entries[] | .value.extraPodMetadata
			) | select(.)`,
			FieldGroups: []string{SystemMetadataFgName},
		},

		// ===== Pod Level (by service name) =====
		"servicePodSpec": {
			Description: convutils.Ref("extraPodSpec for a specific service by name"),
			Query: `(
			  .spec.services[$ARGS.named.serviceName].extraPodSpec
			) | select(.)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"serviceName": api.MAPPINGFIELDTYPE_STRING,
			},
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("serviceSpec"),
				Path:         "extraPodSpec",
			},
		},

		"servicePodMetadata": {
			Description: convutils.Ref("extraPodMetadata for a specific service by name"),
			Query: `(
			  .spec.services[$ARGS.named.serviceName].extraPodMetadata
			) | select(.)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"serviceName": api.MAPPINGFIELDTYPE_STRING,
			},
			FieldGroups: []string{SystemMetadataFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("serviceSpec"),
				Path:         "extraPodMetadata",
			},
		},

		// ===== Container Level (all services) =====
		"allContainers": {
			Description: convutils.Ref("all containers (app and init) across all services via extraPodSpec"),
			Query: `(
			  .spec.services | to_entries[] | .value.extraPodSpec | select(.)
			) | (.containers[]?, .initContainers[]?) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"allAppContainers": {
			Description: convutils.Ref("all app containers across all services via extraPodSpec"),
			Query: `(
			  .spec.services | to_entries[] | .value.extraPodSpec | select(.) | .containers[]?
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"allInitContainers": {
			Description: convutils.Ref("all init containers across all services via extraPodSpec"),
			Query: `(
			  .spec.services | to_entries[] | .value.extraPodSpec | select(.) | (.initContainers // [])[]?
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		// ===== Container Level (by service name) =====
		"serviceContainers": {
			Description: convutils.Ref("all containers (app and init) for a specific service"),
			Query: `(
			  .spec.services[$ARGS.named.serviceName].extraPodSpec | select(.) |
			  (.containers[]?, .initContainers[]?)
			) | select(.)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"serviceName": api.MAPPINGFIELDTYPE_STRING,
			},
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"serviceAppContainers": {
			Description: convutils.Ref("app containers for a specific service"),
			Query: `(
			  .spec.services[$ARGS.named.serviceName].extraPodSpec | select(.) | .containers[]?
			) | select(.)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"serviceName": api.MAPPINGFIELDTYPE_STRING,
			},
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"serviceInitContainers": {
			Description: convutils.Ref("init containers for a specific service"),
			Query: `(
			  .spec.services[$ARGS.named.serviceName].extraPodSpec | select(.) | (.initContainers // [])[]?
			) | select(.)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"serviceName": api.MAPPINGFIELDTYPE_STRING,
			},
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		// ===== Main Container (extraPodSpec.mainContainer) =====
		"allMainContainers": {
			Description: convutils.Ref("mainContainer override across all services (extraPodSpec.mainContainer)"),
			Query: `(
			  .spec.services | to_entries[] | .value.extraPodSpec.mainContainer | select(.)
			)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"serviceMainContainer": {
			Description: convutils.Ref("mainContainer override for a specific service"),
			Query: `(
			  .spec.services[$ARGS.named.serviceName].extraPodSpec.mainContainer
			) | select(.)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"serviceName": api.MAPPINGFIELDTYPE_STRING,
			},
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		// ===== Container by Name (within a service) =====
		"serviceAppContainerByName": {
			Description: convutils.Ref("app container with a given name in a specific service"),
			Query: `(
			  .spec.services[$ARGS.named.serviceName].extraPodSpec | select(.) | .containers[]?
			) | select(.name == $ARGS.named.containerName)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"serviceName":   api.MAPPINGFIELDTYPE_STRING,
				"containerName": api.MAPPINGFIELDTYPE_STRING,
			},
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"serviceInitContainerByName": {
			Description: convutils.Ref("init container with a given name in a specific service"),
			Query: `(
			  .spec.services[$ARGS.named.serviceName].extraPodSpec | select(.) | (.initContainers // [])[]?
			) | select(.name == $ARGS.named.containerName)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"serviceName":   api.MAPPINGFIELDTYPE_STRING,
				"containerName": api.MAPPINGFIELDTYPE_STRING,
			},
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
	}
}
