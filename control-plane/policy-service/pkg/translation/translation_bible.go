/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package translation

type TranslationGuideline struct {
	FieldAndSelector *FieldAndSelectors

	SetParamsRename map[string]string

	CantAdd *FieldAndSelectors
}

type FieldAndSelectors struct {
	Field          string
	ItemSelector   string
	SelectorParams map[string]any
}

// flat maps a scalar/object native policy field to a mapping field that lives
// directly on the spec (no item selector).
func flat(field string) []TranslationGuideline {
	return []TranslationGuideline{{FieldAndSelector: &FieldAndSelectors{Field: field}}}
}

// attr maps a single attribute of an itemized native policy field to the
// per-attribute mapping field, scoped to all items of that array (the "allXxx"
// item selector). This is the "set a default for all instances" style rule.
func attr(field, allSelector string) []TranslationGuideline {
	return []TranslationGuideline{{FieldAndSelector: &FieldAndSelectors{Field: field, ItemSelector: allSelector}}}
}

// itemArray maps the container of an itemized native policy field. Whole-instance
// defaults/sets map to the create-item field; canAdd=false is enforced against
// the "<field>-ItemKey" companion, scoped to all items via the "allXxx" selector.
func itemArray(field, keyField, allSelector string) []TranslationGuideline {
	return []TranslationGuideline{{
		FieldAndSelector: &FieldAndSelectors{Field: field},
		CantAdd:          &FieldAndSelectors{Field: keyField, ItemSelector: allSelector},
	}}
}

// translationBible maps native policy field paths to the corresponding native-spec
// mapping fields (run.ai/Workspace/v1 and run.ai/Distributed/v1). The map is shared
// across the workspace spec and the distributed worker/master specs: the translation
// engine strips the "worker."/"master." prefix before lookup, so paths here are rooted
// at the spec. Distributed-only fields (numWorkers, distributedFramework, ...) live on
// the worker spec and resolve under the workerSpec selector.
//
// Conventions:
//   - Sub-objects (configMap, secret, secretRef, claimInfo, ...) are mapped as whole
//     objects; their nested leaves are not mapped because the engine flattens itemized
//     attributes one level only.
//   - The item key attribute (name / id / resource) is not mapped as an attribute; it is
//     handled by the container entry and its CantAdd companion.
var translationBible = map[string][]TranslationGuideline{
	// ---------------------------------------------------------------------
	// Container
	// ---------------------------------------------------------------------
	"image":           flat("image"),
	"command":         flat("command"),
	"args":            flat("args"),
	"workingDir":      flat("workingDir"),
	"createHomeDir":   flat("createHomeDir"),
	"stdin":           flat("stdin"),
	"tty":             flat("tty"),
	"imagePullPolicy": flat("imagePullPolicy"),
	"probes":          flat("probes"),

	// ---------------------------------------------------------------------
	// Scheduling
	// ---------------------------------------------------------------------
	"category":             flat("category"),
	"nodeType":             flat("nodeType"),
	"nodePools":            flat("nodePools"),
	"nodeAffinityRequired": flat("nodeAffinityRequired"),
	"podAffinity":          flat("podAffinity"),
	"priorityClass":        flat("priorityClass"),
	"preemptibility":       flat("preemptibility"),

	// ---------------------------------------------------------------------
	// Lifecycle
	// ---------------------------------------------------------------------
	"backoffLimit":                           flat("backoffLimit"),
	"autoDeletionTimeAfterCompletionSeconds": flat("autoDeletionTimeAfterCompletionSeconds"),
	"terminateAfterPreemption":               flat("terminateAfterPreemption"),
	"terminationGracePeriodSeconds":          flat("terminationGracePeriodSeconds"),
	"restartPolicy":                          flat("restartPolicy"),

	// ---------------------------------------------------------------------
	// Compute (scalars). Numeric fields (cpuCore*, gpuPortion*, gpuDevicesRequest)
	// support min/max/step -> Range; quantity fields (cpuMemory*, gpuMemory*)
	// support min/max -> QuantityRange.
	// ---------------------------------------------------------------------
	"compute.cpuCoreRequest":    flat("compute-CpuCoreRequest"),
	"compute.cpuCoreLimit":      flat("compute-CpuCoreLimit"),
	"compute.cpuMemoryRequest":  flat("compute-CpuMemoryRequest"),
	"compute.cpuMemoryLimit":    flat("compute-CpuMemoryLimit"),
	"compute.gpuDevicesRequest": flat("compute-GpuDevicesRequest"),
	"compute.gpuPortionRequest": flat("compute-GpuPortionRequest"),
	"compute.gpuPortionLimit":   flat("compute-GpuPortionLimit"),
	"compute.gpuMemoryRequest":  flat("compute-GpuMemoryRequest"),
	"compute.gpuMemoryLimit":    flat("compute-GpuMemoryLimit"),
	"compute.gpuRequestType":    flat("compute-GpuRequestType"),
	"compute.largeShmRequest":   flat("compute-LargeShmRequest"),

	// compute.extendedResources is itemized keyed by "resource".
	"compute.extendedResources":          itemArray("compute-ExtendedResources", "compute-ExtendedResources-ItemKey", "allExtendedResources"),
	"compute.extendedResources.quantity": attr("compute-ExtendedResources-Quantity", "allExtendedResources"),

	// ---------------------------------------------------------------------
	// Security (scalars)
	// ---------------------------------------------------------------------
	"security.allowPrivilegeEscalation": flat("security-AllowPrivilegeEscalation"),
	"security.capabilities":             flat("security-Capabilities"),
	"security.hostIpc":                  flat("security-HostIpc"),
	"security.hostNetwork":              flat("security-HostNetwork"),
	"security.privileged":               flat("security-Privileged"),
	"security.readOnlyRootFilesystem":   flat("security-ReadOnlyRootFilesystem"),
	"security.runAsGid":                 flat("security-RunAsGid"),
	"security.runAsNonRoot":             flat("security-RunAsNonRoot"),
	"security.runAsUid":                 flat("security-RunAsUid"),
	"security.supplementalGroups":       flat("security-SupplementalGroups"),
	"security.seccompProfileType":       flat("security-SeccompProfileType"),
	"security.uidGidSource":             flat("security-UidGidSource"),

	// ---------------------------------------------------------------------
	// Labels / Annotations (itemized key-value maps keyed by "name")
	// ---------------------------------------------------------------------
	"labels":       itemArray("labels", "labels-ItemKey", "allLabels"),
	"labels.value": attr("labels-Value", "allLabels"),

	"annotations":       itemArray("annotations", "annotations-ItemKey", "allAnnotations"),
	"annotations.value": attr("annotations-Value", "allAnnotations"),

	// ---------------------------------------------------------------------
	// Environment variables (itemized keyed by "name")
	// ---------------------------------------------------------------------
	"environmentVariables":                itemArray("environmentVariables", "environmentVariables-ItemKey", "allEnvVars"),
	"environmentVariables.value":          attr("environmentVariables-Value", "allEnvVars"),
	"environmentVariables.description":    attr("environmentVariables-Description", "allEnvVars"),
	"environmentVariables.configMap":      attr("environmentVariables-ConfigMap", "allEnvVars"),
	"environmentVariables.secret":         attr("environmentVariables-Secret", "allEnvVars"),
	"environmentVariables.podFieldRef":    attr("environmentVariables-PodFieldRef", "allEnvVars"),
	"environmentVariables.userCredential": attr("environmentVariables-UserCredential", "allEnvVars"),

	// ---------------------------------------------------------------------
	// Image pull secrets (itemized keyed by "name")
	// ---------------------------------------------------------------------
	"imagePullSecrets":                itemArray("imagePullSecrets", "imagePullSecrets-ItemKey", "allImagePullSecrets"),
	"imagePullSecrets.userCredential": attr("imagePullSecrets-UserCredential", "allImagePullSecrets"),

	// ---------------------------------------------------------------------
	// Tolerations (itemized keyed by "name"). The toleration also has its own
	// "key" attribute, so the item-key companion is the dedicated
	// "tolerations-ItemKey" field while "tolerations-Key" carries the toleration
	// key value.
	// ---------------------------------------------------------------------
	"tolerations":          itemArray("tolerations", "tolerations-ItemKey", "allTolerations"),
	"tolerations.key":      attr("tolerations-Key", "allTolerations"),
	"tolerations.value":    attr("tolerations-Value", "allTolerations"),
	"tolerations.effect":   attr("tolerations-Effect", "allTolerations"),
	"tolerations.operator": attr("tolerations-Operator", "allTolerations"),
	"tolerations.seconds":  attr("tolerations-Seconds", "allTolerations"),

	// ---------------------------------------------------------------------
	// Ports (itemized keyed by "name")
	// ---------------------------------------------------------------------
	"ports":             itemArray("ports", "ports-ItemKey", "allPorts"),
	"ports.container":   attr("ports-Container", "allPorts"),
	"ports.external":    attr("ports-External", "allPorts"),
	"ports.serviceType": attr("ports-ServiceType", "allPorts"),
	"ports.toolName":    attr("ports-ToolName", "allPorts"),
	"ports.toolType":    attr("ports-ToolType", "allPorts"),

	// ---------------------------------------------------------------------
	// Exposed URLs (itemized keyed by "name")
	// ---------------------------------------------------------------------
	"exposedUrls":                   itemArray("exposedUrls", "exposedUrls-ItemKey", "allExposedUrls"),
	"exposedUrls.container":         attr("exposedUrls-Container", "allExposedUrls"),
	"exposedUrls.url":               attr("exposedUrls-Url", "allExposedUrls"),
	"exposedUrls.authorizationType": attr("exposedUrls-AuthorizationType", "allExposedUrls"),
	"exposedUrls.authorizedUsers":   attr("exposedUrls-AuthorizedUsers", "allExposedUrls"),
	"exposedUrls.authorizedGroups":  attr("exposedUrls-AuthorizedGroups", "allExposedUrls"),
	"exposedUrls.toolName":          attr("exposedUrls-ToolName", "allExposedUrls"),
	"exposedUrls.toolType":          attr("exposedUrls-ToolType", "allExposedUrls"),

	// ---------------------------------------------------------------------
	// Related URLs (itemized keyed by "name")
	// ---------------------------------------------------------------------
	"relatedUrls":      itemArray("relatedUrls", "relatedUrls-ItemKey", "allRelatedUrls"),
	"relatedUrls.url":  attr("relatedUrls-Url", "allRelatedUrls"),
	"relatedUrls.type": attr("relatedUrls-Type", "allRelatedUrls"),

	// ---------------------------------------------------------------------
	// Storage: PVC (itemized keyed by "name")
	// ---------------------------------------------------------------------
	"storage.pvc":             itemArray("storage-Pvc", "storage-Pvc-ItemKey", "allPvcs"),
	"storage.pvc.path":        attr("storage-Pvc-Path", "allPvcs"),
	"storage.pvc.claimName":   attr("storage-Pvc-ClaimName", "allPvcs"),
	"storage.pvc.existingPvc": attr("storage-Pvc-ExistingPvc", "allPvcs"),
	"storage.pvc.readOnly":    attr("storage-Pvc-ReadOnly", "allPvcs"),
	"storage.pvc.ephemeral":   attr("storage-Pvc-Ephemeral", "allPvcs"),
	"storage.pvc.dataSharing": attr("storage-Pvc-DataSharing", "allPvcs"),
	"storage.pvc.claimInfo":   attr("storage-Pvc-ClaimInfo", "allPvcs"),

	// ---------------------------------------------------------------------
	// Storage: hostPath (itemized keyed by "name")
	// ---------------------------------------------------------------------
	"storage.hostPath":                  itemArray("storage-HostPath", "storage-HostPath-ItemKey", "allHostPaths"),
	"storage.hostPath.mountPath":        attr("storage-HostPath-MountPath", "allHostPaths"),
	"storage.hostPath.path":             attr("storage-HostPath-Path", "allHostPaths"),
	"storage.hostPath.readOnly":         attr("storage-HostPath-ReadOnly", "allHostPaths"),
	"storage.hostPath.mountPropagation": attr("storage-HostPath-MountPropagation", "allHostPaths"),

	// ---------------------------------------------------------------------
	// Storage: NFS (itemized keyed by "name")
	// ---------------------------------------------------------------------
	"storage.nfs":           itemArray("storage-Nfs", "storage-Nfs-ItemKey", "allNfs"),
	"storage.nfs.path":      attr("storage-Nfs-Path", "allNfs"),
	"storage.nfs.server":    attr("storage-Nfs-Server", "allNfs"),
	"storage.nfs.mountPath": attr("storage-Nfs-MountPath", "allNfs"),
	"storage.nfs.readOnly":  attr("storage-Nfs-ReadOnly", "allNfs"),

	// ---------------------------------------------------------------------
	// Storage: git (itemized keyed by "name")
	// ---------------------------------------------------------------------
	"storage.git":                     itemArray("storage-Git", "storage-Git-ItemKey", "allGits"),
	"storage.git.repository":          attr("storage-Git-Repository", "allGits"),
	"storage.git.branch":              attr("storage-Git-Branch", "allGits"),
	"storage.git.revision":            attr("storage-Git-Revision", "allGits"),
	"storage.git.path":                attr("storage-Git-Path", "allGits"),
	"storage.git.passwordSecret":      attr("storage-Git-PasswordSecret", "allGits"),
	"storage.git.secretKeyOfUser":     attr("storage-Git-SecretKeyOfUser", "allGits"),
	"storage.git.secretKeyOfPassword": attr("storage-Git-SecretKeyOfPassword", "allGits"),
	"storage.git.secretRef":           attr("storage-Git-SecretRef", "allGits"),

	// ---------------------------------------------------------------------
	// Storage: S3 (itemized keyed by "name")
	// ---------------------------------------------------------------------
	"storage.s3":                        itemArray("storage-S3", "storage-S3-ItemKey", "allS3s"),
	"storage.s3.bucket":                 attr("storage-S3-Bucket", "allS3s"),
	"storage.s3.path":                   attr("storage-S3-Path", "allS3s"),
	"storage.s3.url":                    attr("storage-S3-Url", "allS3s"),
	"storage.s3.accessKeySecret":        attr("storage-S3-AccessKeySecret", "allS3s"),
	"storage.s3.awsSessionToken":        attr("storage-S3-AwsSessionToken", "allS3s"),
	"storage.s3.secretKeyOfAccessKeyId": attr("storage-S3-SecretKeyOfAccessKeyId", "allS3s"),
	"storage.s3.secretKeyOfSecretKey":   attr("storage-S3-SecretKeyOfSecretKey", "allS3s"),

	// ---------------------------------------------------------------------
	// Storage: configMapVolume (itemized keyed by "name")
	// ---------------------------------------------------------------------
	"storage.configMapVolume":             itemArray("storage-ConfigMapVolume", "storage-ConfigMapVolume-ItemKey", "allConfigMapVolumes"),
	"storage.configMapVolume.configMap":   attr("storage-ConfigMapVolume-ConfigMap", "allConfigMapVolumes"),
	"storage.configMapVolume.mountPath":   attr("storage-ConfigMapVolume-MountPath", "allConfigMapVolumes"),
	"storage.configMapVolume.subPath":     attr("storage-ConfigMapVolume-SubPath", "allConfigMapVolumes"),
	"storage.configMapVolume.defaultMode": attr("storage-ConfigMapVolume-DefaultMode", "allConfigMapVolumes"),

	// ---------------------------------------------------------------------
	// Storage: secretVolume (itemized keyed by "name")
	// ---------------------------------------------------------------------
	"storage.secretVolume":             itemArray("storage-SecretVolume", "storage-SecretVolume-ItemKey", "allSecretVolumes"),
	"storage.secretVolume.secret":      attr("storage-SecretVolume-Secret", "allSecretVolumes"),
	"storage.secretVolume.mountPath":   attr("storage-SecretVolume-MountPath", "allSecretVolumes"),
	"storage.secretVolume.defaultMode": attr("storage-SecretVolume-DefaultMode", "allSecretVolumes"),

	// ---------------------------------------------------------------------
	// Storage: dataVolume (itemized keyed by "id")
	// ---------------------------------------------------------------------
	"storage.dataVolume":           itemArray("storage-DataVolume", "storage-DataVolume-ItemKey", "allDataVolumes"),
	"storage.dataVolume.mountPath": attr("storage-DataVolume-MountPath", "allDataVolumes"),

	// ---------------------------------------------------------------------
	// Storage: emptyDirVolume (itemized keyed by "name")
	// ---------------------------------------------------------------------
	"storage.emptyDirVolume":           itemArray("storage-EmptyDirVolume", "storage-EmptyDirVolume-ItemKey", "allEmptyDirVolumes"),
	"storage.emptyDirVolume.path":      attr("storage-EmptyDirVolume-Path", "allEmptyDirVolumes"),
	"storage.emptyDirVolume.medium":    attr("storage-EmptyDirVolume-Medium", "allEmptyDirVolumes"),
	"storage.emptyDirVolume.sizeLimit": attr("storage-EmptyDirVolume-SizeLimit", "allEmptyDirVolumes"),

	// ---------------------------------------------------------------------
	// Distributed-only (worker spec). These resolve under the workerSpec selector.
	// ---------------------------------------------------------------------
	"cleanPodPolicy":            flat("cleanPodPolicy"),
	"distributedFramework":      flat("distributedFramework"),
	"mpiLauncherCreationPolicy": flat("mpiLauncherCreationPolicy"),
	"numWorkers":                flat("numWorkers"),
	"segmentSize":               flat("segmentSize"),
	"slotsPerWorker":            flat("slotsPerWorker"),
	"sshAuthMountPath":          flat("sshAuthMountPath"),
	"minReplicas":               flat("minReplicas"),
	"maxReplicas":               flat("maxReplicas"),
}
