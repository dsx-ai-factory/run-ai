/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

// BuildSystemNativeSpecFg builds the "nativeSpec" field group, covering the
// fields of Run.ai native workload specs (WorkspaceSpecSpec, DistributedSpecSpec).
// Field paths are rooted at the spec object (the spec selector "workspaceSpec"
// or "workerSpec" points at the spec).
func BuildSystemNativeSpecFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	fields := BuildFieldGroupFieldMap(nativeSpecFgFields())
	selectors := BuildItemSelectorMap(nativeSpecItemSelectorDefs())
	return fields, selectors
}

// BuildSystemNativeDistributedSpecFg builds the "nativeDistributedSpec" field group,
// adding distributed-only fields on top of the base nativeSpec.
func BuildSystemNativeDistributedSpecFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	fields := BuildFieldGroupFieldMap(nativeDistributedSpecFgFields())
	return fields, nil
}

func nativeSpecItemSelectorDefs() []SystemItemSel {
	return []SystemItemSel{
		{Name: "allLabels", Desc: "all labels", Query: `(.labels // [])[]`},
		{Name: "labelByName", Desc: "label with a given name",
			Query:    `(.labels // [])[] | select(.name == $ARGS.named.labelName)`,
			ArgsInfo: map[string]api.MappingFieldType{"labelName": api.MAPPINGFIELDTYPE_STRING}},

		{Name: "allAnnotations", Desc: "all annotations", Query: `(.annotations // [])[]`},
		{Name: "annotationByName", Desc: "annotation with a given name",
			Query:    `(.annotations // [])[] | select(.name == $ARGS.named.annotationName)`,
			ArgsInfo: map[string]api.MappingFieldType{"annotationName": api.MAPPINGFIELDTYPE_STRING}},

		{Name: "allEnvVars", Desc: "all environment variables", Query: `(.environmentVariables // [])[]`},
		{Name: "envVarByName", Desc: "environment variable with a given name",
			Query:    `(.environmentVariables // [])[] | select(.name == $ARGS.named.envVarName)`,
			ArgsInfo: map[string]api.MappingFieldType{"envVarName": api.MAPPINGFIELDTYPE_STRING}},

		{Name: "allImagePullSecrets", Desc: "all image pull secrets", Query: `(.imagePullSecrets // [])[]`},
		{Name: "imagePullSecretByName", Desc: "image pull secret with a given name",
			Query:    `(.imagePullSecrets // [])[] | select(.name == $ARGS.named.secretName)`,
			ArgsInfo: map[string]api.MappingFieldType{"secretName": api.MAPPINGFIELDTYPE_STRING}},

		{Name: "allTolerations", Desc: "all tolerations", Query: `(.tolerations // [])[]`},
		{Name: "tolerationByName", Desc: "toleration with a given name",
			Query:    `(.tolerations // [])[] | select(.name == $ARGS.named.tolerationName)`,
			ArgsInfo: map[string]api.MappingFieldType{"tolerationName": api.MAPPINGFIELDTYPE_STRING}},

		{Name: "allPorts", Desc: "all ports", Query: `(.ports // [])[]`},
		{Name: "portByName", Desc: "port with a given name",
			Query:    `(.ports // [])[] | select(.name == $ARGS.named.portName)`,
			ArgsInfo: map[string]api.MappingFieldType{"portName": api.MAPPINGFIELDTYPE_STRING}},

		{Name: "allExposedUrls", Desc: "all exposed URLs", Query: `(.exposedUrls // [])[]`},
		{Name: "exposedUrlByName", Desc: "exposed URL with a given name",
			Query:    `(.exposedUrls // [])[] | select(.name == $ARGS.named.urlName)`,
			ArgsInfo: map[string]api.MappingFieldType{"urlName": api.MAPPINGFIELDTYPE_STRING}},

		{Name: "allRelatedUrls", Desc: "all related URLs", Query: `(.relatedUrls // [])[]`},
		{Name: "relatedUrlByName", Desc: "related URL with a given name",
			Query:    `(.relatedUrls // [])[] | select(.name == $ARGS.named.urlName)`,
			ArgsInfo: map[string]api.MappingFieldType{"urlName": api.MAPPINGFIELDTYPE_STRING}},

		{Name: "allExtendedResources", Desc: "all extended resources", Query: `(.compute.extendedResources // [])[]`},
		{Name: "extendedResourceByResource", Desc: "extended resource with a given resource name",
			Query:    `(.compute.extendedResources // [])[] | select(.resource == $ARGS.named.resource)`,
			ArgsInfo: map[string]api.MappingFieldType{"resource": api.MAPPINGFIELDTYPE_STRING}},

		{Name: "allPvcs", Desc: "all PVC storage items", Query: `(.storage.pvc // [])[]`},
		{Name: "pvcByName", Desc: "PVC with a given name",
			Query:    `(.storage.pvc // [])[] | select(.name == $ARGS.named.pvcName)`,
			ArgsInfo: map[string]api.MappingFieldType{"pvcName": api.MAPPINGFIELDTYPE_STRING}},

		{Name: "allHostPaths", Desc: "all hostPath storage items", Query: `(.storage.hostPath // [])[]`},
		{Name: "hostPathByName", Desc: "hostPath with a given name",
			Query:    `(.storage.hostPath // [])[] | select(.name == $ARGS.named.hostPathName)`,
			ArgsInfo: map[string]api.MappingFieldType{"hostPathName": api.MAPPINGFIELDTYPE_STRING}},

		{Name: "allNfs", Desc: "all NFS storage items", Query: `(.storage.nfs // [])[]`},
		{Name: "nfsByName", Desc: "NFS with a given name",
			Query:    `(.storage.nfs // [])[] | select(.name == $ARGS.named.nfsName)`,
			ArgsInfo: map[string]api.MappingFieldType{"nfsName": api.MAPPINGFIELDTYPE_STRING}},

		{Name: "allGits", Desc: "all git storage items", Query: `(.storage.git // [])[]`},
		{Name: "gitByName", Desc: "git with a given name",
			Query:    `(.storage.git // [])[] | select(.name == $ARGS.named.gitName)`,
			ArgsInfo: map[string]api.MappingFieldType{"gitName": api.MAPPINGFIELDTYPE_STRING}},

		{Name: "allS3s", Desc: "all S3 storage items", Query: `(.storage.s3 // [])[]`},
		{Name: "s3ByName", Desc: "S3 with a given name",
			Query:    `(.storage.s3 // [])[] | select(.name == $ARGS.named.s3Name)`,
			ArgsInfo: map[string]api.MappingFieldType{"s3Name": api.MAPPINGFIELDTYPE_STRING}},

		{Name: "allConfigMapVolumes", Desc: "all configMapVolume storage items", Query: `(.storage.configMapVolume // [])[]`},
		{Name: "configMapVolumeByName", Desc: "configMapVolume with a given name",
			Query:    `(.storage.configMapVolume // [])[] | select(.name == $ARGS.named.configMapName)`,
			ArgsInfo: map[string]api.MappingFieldType{"configMapName": api.MAPPINGFIELDTYPE_STRING}},

		{Name: "allSecretVolumes", Desc: "all secretVolume storage items", Query: `(.storage.secretVolume // [])[]`},
		{Name: "secretVolumeByName", Desc: "secretVolume with a given name",
			Query:    `(.storage.secretVolume // [])[] | select(.name == $ARGS.named.secretName)`,
			ArgsInfo: map[string]api.MappingFieldType{"secretName": api.MAPPINGFIELDTYPE_STRING}},

		{Name: "allDataVolumes", Desc: "all dataVolume storage items", Query: `(.storage.dataVolume // [])[]`},
		{Name: "dataVolumeById", Desc: "dataVolume with a given id",
			Query:    `(.storage.dataVolume // [])[] | select(.id == $ARGS.named.dataVolumeId)`,
			ArgsInfo: map[string]api.MappingFieldType{"dataVolumeId": api.MAPPINGFIELDTYPE_STRING}},

		{Name: "allEmptyDirVolumes", Desc: "all emptyDirVolume storage items", Query: `(.storage.emptyDirVolume // [])[]`},
		{Name: "emptyDirVolumeByName", Desc: "emptyDirVolume with a given name",
			Query:    `(.storage.emptyDirVolume // [])[] | select(.name == $ARGS.named.emptyDirName)`,
			ArgsInfo: map[string]api.MappingFieldType{"emptyDirName": api.MAPPINGFIELDTYPE_STRING}},
	}
}

func nativeSpecFgFields() []SystemFgField {
	keyValue := dbmodel.BuiltinFormatterKeyValue
	nativeSpec := SystemNativeSpecFormatterName

	selLabels := []string{"allLabels", "labelByName"}
	selAnnotations := []string{"allAnnotations", "annotationByName"}
	selEnvVars := []string{"allEnvVars", "envVarByName"}
	selImagePullSecrets := []string{"allImagePullSecrets", "imagePullSecretByName"}
	selTolerations := []string{"allTolerations", "tolerationByName"}
	selPorts := []string{"allPorts", "portByName"}
	selExposedUrls := []string{"allExposedUrls", "exposedUrlByName"}
	selRelatedUrls := []string{"allRelatedUrls", "relatedUrlByName"}
	selExtendedResources := []string{"allExtendedResources", "extendedResourceByResource"}
	selPvcs := []string{"allPvcs", "pvcByName"}
	selHostPaths := []string{"allHostPaths", "hostPathByName"}
	selNfs := []string{"allNfs", "nfsByName"}
	selGits := []string{"allGits", "gitByName"}
	selS3s := []string{"allS3s", "s3ByName"}
	selConfigMapVolumes := []string{"allConfigMapVolumes", "configMapVolumeByName"}
	selSecretVolumes := []string{"allSecretVolumes", "secretVolumeByName"}
	selDataVolumes := []string{"allDataVolumes", "dataVolumeById"}
	selEmptyDirVolumes := []string{"allEmptyDirVolumes", "emptyDirVolumeByName"}

	return []SystemFgField{
		// Container fields
		{Name: "image", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "command", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "args", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "workingDir", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "createHomeDir", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_BOOLEAN},
		{Name: "stdin", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_BOOLEAN},
		{Name: "tty", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_BOOLEAN},
		{Name: "imagePullPolicy", Formatter: nativeSpec, Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "probes", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_OBJECT},

		// Scheduling
		{Name: "category", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "nodeType", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "nodePools", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_STRINGS},
		{Name: "nodeAffinityRequired", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "podAffinity", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "priorityClass", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "preemptibility", Formatter: nativeSpec, Type: api.MAPPINGFIELDTYPE_STRING},

		// Lifecycle
		{Name: "backoffLimit", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_INTEGER},
		{Name: "autoDeletionTimeAfterCompletionSeconds", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_INTEGER},
		{Name: "terminateAfterPreemption", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_BOOLEAN},
		{Name: "terminationGracePeriodSeconds", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_INTEGER},
		{Name: "restartPolicy", Formatter: nativeSpec, Type: api.MAPPINGFIELDTYPE_STRING},

		// Compute
		{Name: "compute-CpuCoreRequest", Formatter: keyValue, SpecRef: convutils.Ref("compute.cpuCoreRequest"), Type: api.MAPPINGFIELDTYPE_NUMBER},
		{Name: "compute-CpuCoreLimit", Formatter: keyValue, SpecRef: convutils.Ref("compute.cpuCoreLimit"), Type: api.MAPPINGFIELDTYPE_NUMBER},
		{Name: "compute-CpuMemoryRequest", Formatter: keyValue, SpecRef: convutils.Ref("compute.cpuMemoryRequest"), Type: api.MAPPINGFIELDTYPE_QUANTITY},
		{Name: "compute-CpuMemoryLimit", Formatter: keyValue, SpecRef: convutils.Ref("compute.cpuMemoryLimit"), Type: api.MAPPINGFIELDTYPE_QUANTITY},
		{Name: "compute-GpuDevicesRequest", Formatter: keyValue, SpecRef: convutils.Ref("compute.gpuDevicesRequest"), Type: api.MAPPINGFIELDTYPE_INTEGER},
		{Name: "compute-GpuPortionRequest", Formatter: keyValue, SpecRef: convutils.Ref("compute.gpuPortionRequest"), Type: api.MAPPINGFIELDTYPE_NUMBER},
		{Name: "compute-GpuPortionLimit", Formatter: keyValue, SpecRef: convutils.Ref("compute.gpuPortionLimit"), Type: api.MAPPINGFIELDTYPE_NUMBER},
		{Name: "compute-GpuMemoryRequest", Formatter: keyValue, SpecRef: convutils.Ref("compute.gpuMemoryRequest"), Type: api.MAPPINGFIELDTYPE_QUANTITY},
		{Name: "compute-GpuMemoryLimit", Formatter: keyValue, SpecRef: convutils.Ref("compute.gpuMemoryLimit"), Type: api.MAPPINGFIELDTYPE_QUANTITY},
		{Name: "compute-GpuRequestType", Formatter: nativeSpec, SpecRef: convutils.Ref("compute.gpuRequestType"), Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "compute-LargeShmRequest", Formatter: keyValue, SpecRef: convutils.Ref("compute.largeShmRequest"), Type: api.MAPPINGFIELDTYPE_BOOLEAN},
		{Name: "compute-ExtendedResources", Formatter: nativeSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},

		// Security
		{Name: "security-AllowPrivilegeEscalation", Formatter: keyValue, SpecRef: convutils.Ref("security.allowPrivilegeEscalation"), Type: api.MAPPINGFIELDTYPE_BOOLEAN},
		{Name: "security-Capabilities", Formatter: keyValue, SpecRef: convutils.Ref("security.capabilities"), Type: api.MAPPINGFIELDTYPE_STRINGS},
		{Name: "security-HostIpc", Formatter: keyValue, SpecRef: convutils.Ref("security.hostIpc"), Type: api.MAPPINGFIELDTYPE_BOOLEAN},
		{Name: "security-HostNetwork", Formatter: keyValue, SpecRef: convutils.Ref("security.hostNetwork"), Type: api.MAPPINGFIELDTYPE_BOOLEAN},
		{Name: "security-Privileged", Formatter: keyValue, SpecRef: convutils.Ref("security.privileged"), Type: api.MAPPINGFIELDTYPE_BOOLEAN},
		{Name: "security-ReadOnlyRootFilesystem", Formatter: keyValue, SpecRef: convutils.Ref("security.readOnlyRootFilesystem"), Type: api.MAPPINGFIELDTYPE_BOOLEAN},
		{Name: "security-RunAsGid", Formatter: keyValue, SpecRef: convutils.Ref("security.runAsGid"), Type: api.MAPPINGFIELDTYPE_INTEGER},
		{Name: "security-RunAsNonRoot", Formatter: keyValue, SpecRef: convutils.Ref("security.runAsNonRoot"), Type: api.MAPPINGFIELDTYPE_BOOLEAN},
		{Name: "security-RunAsUid", Formatter: keyValue, SpecRef: convutils.Ref("security.runAsUid"), Type: api.MAPPINGFIELDTYPE_INTEGER},
		{Name: "security-SupplementalGroups", Formatter: keyValue, SpecRef: convutils.Ref("security.supplementalGroups"), Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "security-SeccompProfileType", Formatter: nativeSpec, SpecRef: convutils.Ref("security.seccompProfileType"), Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "security-UidGidSource", Formatter: nativeSpec, SpecRef: convutils.Ref("security.uidGidSource"), Type: api.MAPPINGFIELDTYPE_STRING},

		// Arrays with item selectors
		{Name: "labels", Formatter: nativeSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "labels-ItemKey", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selLabels},
		{Name: "labels-Name", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selLabels},
		{Name: "labels-Value", Formatter: keyValue, SpecRef: convutils.Ref("value"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selLabels},

		{Name: "annotations", Formatter: nativeSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "annotations-ItemKey", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selAnnotations},
		{Name: "annotations-Name", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selAnnotations},
		{Name: "annotations-Value", Formatter: keyValue, SpecRef: convutils.Ref("value"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selAnnotations},

		{Name: "environmentVariables", Formatter: nativeSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "environmentVariables-ItemKey", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selEnvVars},
		{Name: "environmentVariables-Name", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selEnvVars},
		{Name: "environmentVariables-Value", Formatter: keyValue, SpecRef: convutils.Ref("value"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selEnvVars},
		{Name: "environmentVariables-Description", Formatter: keyValue, SpecRef: convutils.Ref("description"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selEnvVars},
		{Name: "environmentVariables-ConfigMap", Formatter: keyValue, SpecRef: convutils.Ref("configMap"), Type: api.MAPPINGFIELDTYPE_OBJECT, ItemSelectors: selEnvVars},
		{Name: "environmentVariables-ConfigMap-Name", Formatter: keyValue, SpecRef: convutils.Ref("configMap.name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selEnvVars},
		{Name: "environmentVariables-ConfigMap-Key", Formatter: keyValue, SpecRef: convutils.Ref("configMap.key"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selEnvVars},
		{Name: "environmentVariables-Secret", Formatter: keyValue, SpecRef: convutils.Ref("secret"), Type: api.MAPPINGFIELDTYPE_OBJECT, ItemSelectors: selEnvVars},
		{Name: "environmentVariables-Secret-Name", Formatter: keyValue, SpecRef: convutils.Ref("secret.name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selEnvVars},
		{Name: "environmentVariables-Secret-Key", Formatter: keyValue, SpecRef: convutils.Ref("secret.key"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selEnvVars},
		{Name: "environmentVariables-PodFieldRef", Formatter: keyValue, SpecRef: convutils.Ref("podFieldRef"), Type: api.MAPPINGFIELDTYPE_OBJECT, ItemSelectors: selEnvVars},
		{Name: "environmentVariables-PodFieldRef-Path", Formatter: keyValue, SpecRef: convutils.Ref("podFieldRef.path"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selEnvVars},
		{Name: "environmentVariables-UserCredential", Formatter: keyValue, SpecRef: convutils.Ref("userCredential"), Type: api.MAPPINGFIELDTYPE_OBJECT, ItemSelectors: selEnvVars},
		{Name: "environmentVariables-UserCredential-Name", Formatter: keyValue, SpecRef: convutils.Ref("userCredential.name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selEnvVars},
		{Name: "environmentVariables-UserCredential-Key", Formatter: keyValue, SpecRef: convutils.Ref("userCredential.key"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selEnvVars},

		{Name: "imagePullSecrets", Formatter: nativeSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "imagePullSecrets-ItemKey", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selImagePullSecrets},
		{Name: "imagePullSecrets-Name", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selImagePullSecrets},
		{Name: "imagePullSecrets-UserCredential", Formatter: keyValue, SpecRef: convutils.Ref("userCredential"), Type: api.MAPPINGFIELDTYPE_BOOLEAN, ItemSelectors: selImagePullSecrets},

		{Name: "tolerations", Formatter: nativeSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		// Every itemized array exposes its item-key companion as "<field>-ItemKey".
		// For tolerations this also avoids clashing with the toleration's own "key"
		// attribute (below), which keeps the plain "-Key" suffix.
		{Name: "tolerations-ItemKey", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selTolerations},
		{Name: "tolerations-Name", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selTolerations},
		{Name: "tolerations-Key", Formatter: keyValue, SpecRef: convutils.Ref("key"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selTolerations},
		{Name: "tolerations-Value", Formatter: keyValue, SpecRef: convutils.Ref("value"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selTolerations},
		{Name: "tolerations-Effect", Formatter: keyValue, SpecRef: convutils.Ref("effect"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selTolerations},
		{Name: "tolerations-Operator", Formatter: keyValue, SpecRef: convutils.Ref("operator"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selTolerations},
		{Name: "tolerations-Seconds", Formatter: keyValue, SpecRef: convutils.Ref("seconds"), Type: api.MAPPINGFIELDTYPE_INTEGER, ItemSelectors: selTolerations},

		{Name: "ports", Formatter: nativeSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "ports-ItemKey", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selPorts},
		{Name: "ports-Name", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selPorts},
		{Name: "ports-Container", Formatter: keyValue, SpecRef: convutils.Ref("container"), Type: api.MAPPINGFIELDTYPE_INTEGER, ItemSelectors: selPorts},
		{Name: "ports-External", Formatter: keyValue, SpecRef: convutils.Ref("external"), Type: api.MAPPINGFIELDTYPE_INTEGER, ItemSelectors: selPorts},
		{Name: "ports-ServiceType", Formatter: keyValue, SpecRef: convutils.Ref("serviceType"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selPorts},
		{Name: "ports-ToolName", Formatter: keyValue, SpecRef: convutils.Ref("toolName"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selPorts},
		{Name: "ports-ToolType", Formatter: keyValue, SpecRef: convutils.Ref("toolType"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selPorts},

		{Name: "exposedUrls", Formatter: nativeSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "exposedUrls-ItemKey", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selExposedUrls},
		{Name: "exposedUrls-Name", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selExposedUrls},
		{Name: "exposedUrls-Container", Formatter: keyValue, SpecRef: convutils.Ref("container"), Type: api.MAPPINGFIELDTYPE_INTEGER, ItemSelectors: selExposedUrls},
		{Name: "exposedUrls-Url", Formatter: keyValue, SpecRef: convutils.Ref("url"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selExposedUrls},
		{Name: "exposedUrls-AuthorizationType", Formatter: keyValue, SpecRef: convutils.Ref("authorizationType"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selExposedUrls},
		{Name: "exposedUrls-AuthorizedUsers", Formatter: keyValue, SpecRef: convutils.Ref("authorizedUsers"), Type: api.MAPPINGFIELDTYPE_STRINGS, ItemSelectors: selExposedUrls},
		{Name: "exposedUrls-AuthorizedGroups", Formatter: keyValue, SpecRef: convutils.Ref("authorizedGroups"), Type: api.MAPPINGFIELDTYPE_STRINGS, ItemSelectors: selExposedUrls},
		{Name: "exposedUrls-ToolName", Formatter: keyValue, SpecRef: convutils.Ref("toolName"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selExposedUrls},
		{Name: "exposedUrls-ToolType", Formatter: keyValue, SpecRef: convutils.Ref("toolType"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selExposedUrls},

		{Name: "relatedUrls", Formatter: nativeSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "relatedUrls-ItemKey", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selRelatedUrls},
		{Name: "relatedUrls-Name", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selRelatedUrls},
		{Name: "relatedUrls-Url", Formatter: keyValue, SpecRef: convutils.Ref("url"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selRelatedUrls},
		{Name: "relatedUrls-Type", Formatter: keyValue, SpecRef: convutils.Ref("type"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selRelatedUrls},

		{Name: "compute-ExtendedResources-ItemKey", Formatter: keyValue, SpecRef: convutils.Ref("resource"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selExtendedResources},
		{Name: "compute-ExtendedResources-Resource", Formatter: keyValue, SpecRef: convutils.Ref("resource"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selExtendedResources},
		{Name: "compute-ExtendedResources-Quantity", Formatter: keyValue, SpecRef: convutils.Ref("quantity"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selExtendedResources},
		// Storage arrays
		{Name: "storage-Pvc", Formatter: nativeSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "storage-Pvc-ItemKey", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selPvcs},
		{Name: "storage-Pvc-Name", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selPvcs},
		{Name: "storage-Pvc-Path", Formatter: keyValue, SpecRef: convutils.Ref("path"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selPvcs},
		{Name: "storage-Pvc-ClaimName", Formatter: keyValue, SpecRef: convutils.Ref("claimName"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selPvcs},
		{Name: "storage-Pvc-ExistingPvc", Formatter: keyValue, SpecRef: convutils.Ref("existingPvc"), Type: api.MAPPINGFIELDTYPE_BOOLEAN, ItemSelectors: selPvcs},
		{Name: "storage-Pvc-ReadOnly", Formatter: keyValue, SpecRef: convutils.Ref("readOnly"), Type: api.MAPPINGFIELDTYPE_BOOLEAN, ItemSelectors: selPvcs},
		{Name: "storage-Pvc-Ephemeral", Formatter: keyValue, SpecRef: convutils.Ref("ephemeral"), Type: api.MAPPINGFIELDTYPE_BOOLEAN, ItemSelectors: selPvcs},
		{Name: "storage-Pvc-DataSharing", Formatter: keyValue, SpecRef: convutils.Ref("dataSharing"), Type: api.MAPPINGFIELDTYPE_BOOLEAN, ItemSelectors: selPvcs},
		{Name: "storage-Pvc-ClaimInfo", Formatter: keyValue, SpecRef: convutils.Ref("claimInfo"), Type: api.MAPPINGFIELDTYPE_OBJECT, ItemSelectors: selPvcs},
		{Name: "storage-Pvc-ClaimInfo-Size", Formatter: keyValue, SpecRef: convutils.Ref("claimInfo.size"), Type: api.MAPPINGFIELDTYPE_QUANTITY, ItemSelectors: selPvcs},
		{Name: "storage-Pvc-ClaimInfo-StorageClass", Formatter: keyValue, SpecRef: convutils.Ref("claimInfo.storageClass"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selPvcs},
		{Name: "storage-Pvc-ClaimInfo-VolumeMode", Formatter: keyValue, SpecRef: convutils.Ref("claimInfo.volumeMode"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selPvcs},
		{Name: "storage-Pvc-ClaimInfo-AccessModes", Formatter: keyValue, SpecRef: convutils.Ref("claimInfo.accessModes"), Type: api.MAPPINGFIELDTYPE_OBJECT, ItemSelectors: selPvcs},
		{Name: "storage-Pvc-ClaimInfo-AccessModes-ReadWriteOnce", Formatter: keyValue, SpecRef: convutils.Ref("claimInfo.accessModes.readWriteOnce"), Type: api.MAPPINGFIELDTYPE_BOOLEAN, ItemSelectors: selPvcs},
		{Name: "storage-Pvc-ClaimInfo-AccessModes-ReadOnlyMany", Formatter: keyValue, SpecRef: convutils.Ref("claimInfo.accessModes.readOnlyMany"), Type: api.MAPPINGFIELDTYPE_BOOLEAN, ItemSelectors: selPvcs},
		{Name: "storage-Pvc-ClaimInfo-AccessModes-ReadWriteMany", Formatter: keyValue, SpecRef: convutils.Ref("claimInfo.accessModes.readWriteMany"), Type: api.MAPPINGFIELDTYPE_BOOLEAN, ItemSelectors: selPvcs},
		{Name: "storage-Pvc-ClaimInfo-AddedAttrValues", Formatter: keyValue, SpecRef: convutils.Ref("claimInfo.addedAttrValues"), Type: api.MAPPINGFIELDTYPE_OBJECT, ItemSelectors: selPvcs},

		{Name: "storage-HostPath", Formatter: nativeSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "storage-HostPath-ItemKey", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selHostPaths},
		{Name: "storage-HostPath-Name", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selHostPaths},
		{Name: "storage-HostPath-MountPath", Formatter: keyValue, SpecRef: convutils.Ref("mountPath"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selHostPaths},
		{Name: "storage-HostPath-Path", Formatter: keyValue, SpecRef: convutils.Ref("path"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selHostPaths},
		{Name: "storage-HostPath-ReadOnly", Formatter: keyValue, SpecRef: convutils.Ref("readOnly"), Type: api.MAPPINGFIELDTYPE_BOOLEAN, ItemSelectors: selHostPaths},
		{Name: "storage-HostPath-MountPropagation", Formatter: keyValue, SpecRef: convutils.Ref("mountPropagation"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selHostPaths},

		{Name: "storage-Nfs", Formatter: nativeSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "storage-Nfs-ItemKey", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selNfs},
		{Name: "storage-Nfs-Name", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selNfs},
		{Name: "storage-Nfs-Path", Formatter: keyValue, SpecRef: convutils.Ref("path"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selNfs},
		{Name: "storage-Nfs-Server", Formatter: keyValue, SpecRef: convutils.Ref("server"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selNfs},
		{Name: "storage-Nfs-MountPath", Formatter: keyValue, SpecRef: convutils.Ref("mountPath"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selNfs},
		{Name: "storage-Nfs-ReadOnly", Formatter: keyValue, SpecRef: convutils.Ref("readOnly"), Type: api.MAPPINGFIELDTYPE_BOOLEAN, ItemSelectors: selNfs},

		{Name: "storage-Git", Formatter: nativeSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "storage-Git-ItemKey", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selGits},
		{Name: "storage-Git-Name", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selGits},
		{Name: "storage-Git-Repository", Formatter: keyValue, SpecRef: convutils.Ref("repository"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selGits},
		{Name: "storage-Git-Branch", Formatter: keyValue, SpecRef: convutils.Ref("branch"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selGits},
		{Name: "storage-Git-Revision", Formatter: keyValue, SpecRef: convutils.Ref("revision"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selGits},
		{Name: "storage-Git-Path", Formatter: keyValue, SpecRef: convutils.Ref("path"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selGits},
		{Name: "storage-Git-PasswordSecret", Formatter: keyValue, SpecRef: convutils.Ref("passwordSecret"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selGits},
		{Name: "storage-Git-SecretKeyOfPassword", Formatter: keyValue, SpecRef: convutils.Ref("secretKeyOfPassword"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selGits},
		{Name: "storage-Git-SecretKeyOfUser", Formatter: keyValue, SpecRef: convutils.Ref("secretKeyOfUser"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selGits},
		{Name: "storage-Git-SecretRef", Formatter: keyValue, SpecRef: convutils.Ref("secretRef"), Type: api.MAPPINGFIELDTYPE_OBJECT, ItemSelectors: selGits},
		{Name: "storage-Git-SecretRef-Name", Formatter: keyValue, SpecRef: convutils.Ref("secretRef.name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selGits},
		{Name: "storage-Git-SecretRef-AuthenticationMethod", Formatter: keyValue, SpecRef: convutils.Ref("secretRef.authenticationMethod"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selGits},
		{Name: "storage-Git-SecretRef-SecretKeyOfUser", Formatter: keyValue, SpecRef: convutils.Ref("secretRef.secretKeyOfUser"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selGits},
		{Name: "storage-Git-SecretRef-SecretKeyOfPassword", Formatter: keyValue, SpecRef: convutils.Ref("secretRef.secretKeyOfPassword"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selGits},

		{Name: "storage-S3", Formatter: nativeSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "storage-S3-ItemKey", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selS3s},
		{Name: "storage-S3-Name", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selS3s},
		{Name: "storage-S3-Bucket", Formatter: keyValue, SpecRef: convutils.Ref("bucket"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selS3s},
		{Name: "storage-S3-Path", Formatter: keyValue, SpecRef: convutils.Ref("path"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selS3s},
		{Name: "storage-S3-Url", Formatter: keyValue, SpecRef: convutils.Ref("url"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selS3s},
		{Name: "storage-S3-AccessKeySecret", Formatter: keyValue, SpecRef: convutils.Ref("accessKeySecret"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selS3s},
		{Name: "storage-S3-AwsSessionToken", Formatter: keyValue, SpecRef: convutils.Ref("awsSessionToken"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selS3s},
		{Name: "storage-S3-SecretKeyOfAccessKeyId", Formatter: keyValue, SpecRef: convutils.Ref("secretKeyOfAccessKeyId"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selS3s},
		{Name: "storage-S3-SecretKeyOfSecretKey", Formatter: keyValue, SpecRef: convutils.Ref("secretKeyOfSecretKey"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selS3s},

		{Name: "storage-ConfigMapVolume", Formatter: nativeSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "storage-ConfigMapVolume-ItemKey", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selConfigMapVolumes},
		{Name: "storage-ConfigMapVolume-Name", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selConfigMapVolumes},
		{Name: "storage-ConfigMapVolume-ConfigMap", Formatter: keyValue, SpecRef: convutils.Ref("configMap"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selConfigMapVolumes},
		{Name: "storage-ConfigMapVolume-MountPath", Formatter: keyValue, SpecRef: convutils.Ref("mountPath"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selConfigMapVolumes},
		{Name: "storage-ConfigMapVolume-SubPath", Formatter: keyValue, SpecRef: convutils.Ref("subPath"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selConfigMapVolumes},
		{Name: "storage-ConfigMapVolume-DefaultMode", Formatter: keyValue, SpecRef: convutils.Ref("defaultMode"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selConfigMapVolumes},

		{Name: "storage-SecretVolume", Formatter: nativeSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "storage-SecretVolume-ItemKey", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selSecretVolumes},
		{Name: "storage-SecretVolume-Name", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selSecretVolumes},
		{Name: "storage-SecretVolume-Secret", Formatter: keyValue, SpecRef: convutils.Ref("secret"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selSecretVolumes},
		{Name: "storage-SecretVolume-MountPath", Formatter: keyValue, SpecRef: convutils.Ref("mountPath"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selSecretVolumes},
		{Name: "storage-SecretVolume-DefaultMode", Formatter: keyValue, SpecRef: convutils.Ref("defaultMode"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selSecretVolumes},

		{Name: "storage-DataVolume", Formatter: nativeSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "storage-DataVolume-ItemKey", Formatter: keyValue, SpecRef: convutils.Ref("id"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selDataVolumes},
		{Name: "storage-DataVolume-Id", Formatter: keyValue, SpecRef: convutils.Ref("id"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selDataVolumes},
		{Name: "storage-DataVolume-MountPath", Formatter: keyValue, SpecRef: convutils.Ref("mountPath"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selDataVolumes},

		{Name: "storage-EmptyDirVolume", Formatter: nativeSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "storage-EmptyDirVolume-ItemKey", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selEmptyDirVolumes},
		{Name: "storage-EmptyDirVolume-Name", Formatter: keyValue, SpecRef: convutils.Ref("name"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selEmptyDirVolumes},
		{Name: "storage-EmptyDirVolume-Path", Formatter: keyValue, SpecRef: convutils.Ref("path"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selEmptyDirVolumes},
		{Name: "storage-EmptyDirVolume-Medium", Formatter: keyValue, SpecRef: convutils.Ref("medium"), Type: api.MAPPINGFIELDTYPE_STRING, ItemSelectors: selEmptyDirVolumes},
		{Name: "storage-EmptyDirVolume-SizeLimit", Formatter: keyValue, SpecRef: convutils.Ref("sizeLimit"), Type: api.MAPPINGFIELDTYPE_QUANTITY, ItemSelectors: selEmptyDirVolumes},
	}
}

func nativeDistributedSpecFgFields() []SystemFgField {
	keyValue := dbmodel.BuiltinFormatterKeyValue
	nativeSpec := SystemNativeSpecFormatterName

	return []SystemFgField{
		{Name: "cleanPodPolicy", Formatter: nativeSpec, Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "distributedFramework", Formatter: nativeSpec, Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "mpiLauncherCreationPolicy", Formatter: nativeSpec, Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "numWorkers", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_INTEGER},
		{Name: "segmentSize", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_INTEGER},
		{Name: "slotsPerWorker", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_INTEGER},
		{Name: "sshAuthMountPath", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "minReplicas", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_INTEGER},
		{Name: "maxReplicas", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_INTEGER},
	}
}
