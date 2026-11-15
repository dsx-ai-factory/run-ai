/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

const (
	itemSelAllResourceClaims          = "allResourceClaims"
	itemSelResourceClaimByName        = "resourceClaimByName"
	itemSelAllDirectResourceClaims    = "allDirectResourceClaims"
	itemSelAllTemplatedResourceClaims = "allTemplatedResourceClaims"
)

func BuildFieldGroupFieldMap(entries []SystemFgField) map[string]api.FieldGroupField {
	out := make(map[string]api.FieldGroupField, len(entries))

	for _, f := range entries {
		if len(f.ItemSelectors) > 0 {
			f.ItemSelectors = append([]string(nil), f.ItemSelectors...)
		}
		specRef := f.Name
		if f.SpecRef != nil {
			specRef = *f.SpecRef
		}
		out[f.Name] = api.FieldGroupField{
			Formatter:     f.Formatter,
			SpecRef:       specRef,
			Type:          f.Type,
			ItemSelectors: f.ItemSelectors,
		}
	}

	return out
}

func BuildItemSelectorMap(defs []SystemItemSel) map[string]api.ItemSelector {
	out := make(map[string]api.ItemSelector, len(defs))
	for _, d := range defs {
		sel := api.ItemSelector{
			Query: d.Query,
		}

		if d.Desc != "" {
			sel.Description = convutils.Ref(d.Desc)
		}

		if len(d.ArgsInfo) > 0 {
			sel.ArgsInfo = d.ArgsInfo
		}

		if len(d.ArgsEnum) > 0 {
			sel.ArgsEnum = d.ArgsEnum
		}

		out[d.Name] = sel
	}

	return out
}

func BuildSystemPodSpecFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	fields := BuildFieldGroupFieldMap(podSpecFgFields())
	selectors := BuildItemSelectorMap(podSpecItemSelectorDefs())
	return fields, selectors
}

func BuildSystemPodSpecContainerFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	fields := BuildFieldGroupFieldMap(podSpecContainerFgFields())
	selectors := BuildItemSelectorMap(podSpecContainerItemSelectorDefs())
	return fields, selectors
}

func BuildSystemPodSpecAndContainerFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	fields := BuildFieldGroupFieldMap(podSpecAndContainerFgFields())
	selectors := BuildItemSelectorMap(podSpecAndContainerItemSelectorDefs())
	return fields, selectors
}

func podSpecItemSelectorDefs() []SystemItemSel {
	return []SystemItemSel{
		{
			Name:  "allDnsOptions",
			Desc:  "all DNS options",
			Query: `(.dnsConfig.options // [])[]`,
		},
		{
			Name:     "dnsOptionByName",
			Desc:     "DNS option with a given name",
			Query:    `(.dnsConfig.options // [])[] | select(.name == $ARGS.named.dnsOptionName)`,
			ArgsInfo: map[string]api.MappingFieldType{"dnsOptionName": api.MAPPINGFIELDTYPE_STRING},
		},
		{
			Name:  "allHostAliases",
			Desc:  "all host aliases",
			Query: `(.hostAliases // [])[]`,
		},
		{
			Name:     "hostAliasByIp",
			Desc:     "all host aliases with a given IP address",
			Query:    `(.hostAliases // [])[] | select(.ip == $ARGS.named.hostAliasIp)`,
			ArgsInfo: map[string]api.MappingFieldType{"hostAliasIp": api.MAPPINGFIELDTYPE_IPADDRESS},
		},
		{
			Name:  "allTolerations",
			Desc:  "all tolerations",
			Query: `(.tolerations // [])[]`,
		},
		{
			Name:     "tolerationByEffect",
			Desc:     "all tolerations with a given effect",
			Query:    `(.tolerations // [])[] | select(.effect == $ARGS.named.effect)`,
			ArgsInfo: map[string]api.MappingFieldType{"effect": api.MAPPINGFIELDTYPE_STRING},
			ArgsEnum: map[string][]string{"effect": {"NoSchedule", "PreferNoSchedule", "NoExecute"}},
		},
		{
			Name:     "tolerationByOperator",
			Desc:     "all tolerations with a given operator",
			Query:    `(.tolerations // [])[] | select(.operator == $ARGS.named.operator)`,
			ArgsInfo: map[string]api.MappingFieldType{"operator": api.MAPPINGFIELDTYPE_STRING},
			ArgsEnum: map[string][]string{"operator": {"Equal", "Exists"}},
		},
		{
			Name:  "allNodeSelectors",
			Desc:  "all node selectors",
			Query: `(.nodeSelector // {}) | to_entries[]`,
		},
		{
			Name:     "nodeSelectorByKey",
			Desc:     "all node selectors with a given key",
			Query:    `((.nodeSelector // {}) | to_entries[]) | select(.key == $ARGS.named.nodeSelectorKey)`,
			ArgsInfo: map[string]api.MappingFieldType{"nodeSelectorKey": api.MAPPINGFIELDTYPE_STRING},
		},
		{
			Name:  "allResourceOverhead",
			Desc:  "all overhead resources",
			Query: `(.overhead // {}) | to_entries[]`,
		},
		{
			Name:     "resourceOverheadByKey",
			Desc:     "resource overhead with a given resource name",
			Query:    `((.overhead // {}) | to_entries[]) | select(.key == $ARGS.named.resourceOverheadKey)`,
			ArgsInfo: map[string]api.MappingFieldType{"resourceOverheadKey": api.MAPPINGFIELDTYPE_STRING},
		},
		{
			Name:  "allImagePullSecrets",
			Desc:  "all image pull secret references",
			Query: `(.imagePullSecrets // [])[]`,
		},
		{
			Name:  "allReadinessGates",
			Desc:  "all readiness gates",
			Query: `(.readinessGates // [])[]`,
		},
		{
			Name:  "allSchedulingGates",
			Desc:  "all scheduling gates",
			Query: `(.schedulingGates // [])[]`,
		},
		{
			Name:  itemSelAllResourceClaims,
			Desc:  "all resource claims",
			Query: `(.resourceClaims // [])[]`,
		},
		{
			Name:     itemSelResourceClaimByName,
			Desc:     "resource claim with a given name",
			Query:    `(.resourceClaims // [])[] | select(.name == $ARGS.named.resourceClaimPodName)`,
			ArgsInfo: map[string]api.MappingFieldType{"resourceClaimPodName": api.MAPPINGFIELDTYPE_STRING},
		},
		{
			Name:  itemSelAllDirectResourceClaims,
			Desc:  "resource claims that reference a ResourceClaim by name",
			Query: `(.resourceClaims // [])[] | select(.resourceClaimName != null and .resourceClaimName != "")`,
		},
		{
			Name:  itemSelAllTemplatedResourceClaims,
			Desc:  "resource claims produced from a ResourceClaimTemplate",
			Query: `(.resourceClaims // [])[] | select(.resourceClaimTemplateName != null and .resourceClaimTemplateName != "")`,
		},
		{
			Name:  "allVolumes",
			Desc:  "all volumes",
			Query: `(.volumes // [])[]`,
		},
		{
			Name:  "allConfigMapVolumes",
			Desc:  "all configMap volumes",
			Query: `(.volumes // [])[] | select(.configMap != null)`,
		},
		{
			Name:  "allEmptyDirVolumes",
			Desc:  "all emptyDir volumes",
			Query: `(.volumes // [])[] | select(.emptyDir != null)`,
		},
		{
			Name:  "allHostPathVolumes",
			Desc:  "all hostPath volumes",
			Query: `(.volumes // [])[] | select(.hostPath != null)`,
		},
		{
			Name:  "allNfsVolumes",
			Desc:  "all NFS volumes",
			Query: `(.volumes // [])[] | select(.nfs != null)`,
		},
		{
			Name:  "allPvcVolumes",
			Desc:  "all PVC volumes",
			Query: `(.volumes // [])[] | select(.persistentVolumeClaim != null)`,
		},
		{
			Name:  "allSecretVolumes",
			Desc:  "all secret volumes",
			Query: `(.volumes // [])[] | select(.secret != null)`,
		},
	}
}

func podSpecContainerItemSelectorDefs() []SystemItemSel {
	return []SystemItemSel{
		{
			Name:  "allEnvVars",
			Desc:  "all environment variables",
			Query: `(.env // [])[]`,
		},
		{
			Name:  "envVarAsValue",
			Desc:  "all environment variables defined with a literal value",
			Query: `(.env // [])[] | select(has("value"))`,
		},
		{
			Name:  "envVarFromConfigMap",
			Desc:  "all environment variables sourced from a ConfigMap",
			Query: `(.env // [])[] | select(.valueFrom.configMapKeyRef != null)`,
		},
		{
			Name:  "envVarFromField",
			Desc:  "all environment variables sourced from a field reference",
			Query: `(.env // [])[] | select(.valueFrom.fieldRef != null)`,
		},
		{
			Name:  "envVarFromSecret",
			Desc:  "all environment variables sourced from a secret",
			Query: `(.env // [])[] | select(.valueFrom.secretKeyRef != null)`,
		},
		{
			Name:  "envVarFromResource",
			Desc:  "all environment variables sourced from a resource field",
			Query: `(.env // [])[] | select(.valueFrom.resourceFieldRef != null)`,
		},
		{
			Name:  "allEnvVarBulks",
			Desc:  "all envFrom entries",
			Query: `(.envFrom // [])[]`,
		},
		{
			Name:  "envVarBulksFromConfigMap",
			Desc:  "all envFrom entries that load from a ConfigMap",
			Query: `(.envFrom // [])[] | select(.configMapRef != null)`,
		},
		{
			Name:  "envVarBulksFromSecret",
			Desc:  "all envFrom entries that load from a Secret",
			Query: `(.envFrom // [])[] | select(.secretRef != null)`,
		},
		{
			Name:  "allPorts",
			Desc:  "all ports",
			Query: `(.ports // [])[]`,
		},
		{
			Name:     "portByContainerPort",
			Desc:     "port with a given containerPort",
			Query:    `(.ports // [])[] | select(.containerPort == $ARGS.named.containerPort)`,
			ArgsInfo: map[string]api.MappingFieldType{"containerPort": api.MAPPINGFIELDTYPE_INTEGER},
		},
		{
			Name:  "allVolumeMounts",
			Desc:  "all volume mounts",
			Query: `(.volumeMounts // [])[]`,
		},
		{
			Name:     "volumeMountByName",
			Desc:     "volume mount with a given name",
			Query:    `(.volumeMounts // [])[] | select(.name == $ARGS.named.volumeMountName)`,
			ArgsInfo: map[string]api.MappingFieldType{"volumeMountName": api.MAPPINGFIELDTYPE_STRING},
		},
		{
			Name:     "volumeMountsByReadOnly",
			Desc:     "all volume mounts with a given readOnly flag",
			Query:    `(.volumeMounts // [])[] | select((.readOnly // false) == $ARGS.named.readOnly)`,
			ArgsInfo: map[string]api.MappingFieldType{"readOnly": api.MAPPINGFIELDTYPE_BOOLEAN},
		},
	}
}

func podSpecAndContainerItemSelectorDefs() []SystemItemSel {
	return []SystemItemSel{
		{
			Name:  "allResourceRequests",
			Desc:  "all resource requests",
			Query: `(.resources.requests // {}) | to_entries[]`,
		},
		{
			Name:     "resourceRequestByKey",
			Desc:     "resource request with a given resource name",
			Query:    `((.resources.requests // {}) | to_entries[]) | select(.key == $ARGS.named.resourceRequestKey)`,
			ArgsInfo: map[string]api.MappingFieldType{"resourceRequestKey": api.MAPPINGFIELDTYPE_STRING},
		},
		{
			Name:  "allResourceLimits",
			Desc:  "all resource limits",
			Query: `(.resources.limits // {}) | to_entries[]`,
		},
		{
			Name:     "resourceLimitByKey",
			Desc:     "resource limit with a given resource name",
			Query:    `((.resources.limits // {}) | to_entries[]) | select(.key == $ARGS.named.resourceLimitKey)`,
			ArgsInfo: map[string]api.MappingFieldType{"resourceLimitKey": api.MAPPINGFIELDTYPE_STRING},
		},
	}
}

func podSpecFgFields() []SystemFgField {
	keyValue := dbmodel.BuiltinFormatterKeyValue
	podSpec := SystemPodSpecFormatterName
	selectorsDnsOptions := []string{"allDnsOptions", "dnsOptionByName"}
	selectorsHostAliases := []string{"allHostAliases", "hostAliasByIp"}
	selectorsTolerations := []string{"allTolerations", "tolerationByEffect", "tolerationByOperator"}
	selectorsNodeSelectors := []string{"allNodeSelectors", "nodeSelectorByKey"}
	selectorsResourceOverhead := []string{"allResourceOverhead", "resourceOverheadByKey"}
	selectorsImagePullSecrets := []string{"allImagePullSecrets"}
	selectorsReadinessGates := []string{"allReadinessGates"}
	selectorsSchedulingGates := []string{"allSchedulingGates"}
	selectorsResourceClaimsAll := []string{itemSelAllResourceClaims, itemSelResourceClaimByName, itemSelAllDirectResourceClaims, itemSelAllTemplatedResourceClaims}
	selectorsResourceClaimsDirect := []string{itemSelAllResourceClaims, itemSelResourceClaimByName, itemSelAllDirectResourceClaims}
	selectorsResourceClaimsTemplated := []string{itemSelAllResourceClaims, itemSelResourceClaimByName, itemSelAllTemplatedResourceClaims}
	selectorsVolumesAllKinds := []string{"allVolumes", "allConfigMapVolumes", "allEmptyDirVolumes", "allHostPathVolumes", "allNfsVolumes", "allPvcVolumes", "allSecretVolumes"}
	selectorsVolumesConfigMap := []string{"allConfigMapVolumes"}
	selectorsVolumesEmptyDir := []string{"allEmptyDirVolumes"}
	selectorsVolumesHostPath := []string{"allHostPathVolumes"}
	selectorsVolumesNfs := []string{"allNfsVolumes"}
	selectorsVolumesPvc := []string{"allPvcVolumes"}
	selectorsVolumesSecret := []string{"allSecretVolumes"}

	return []SystemFgField{
		{
			Name:      "activeDeadlineSeconds",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "automountServiceAccountToken",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		{
			Name:      "dnsNameservers",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("dnsConfig.nameservers"),
			Type:      api.MAPPINGFIELDTYPE_STRINGS,
		},
		{
			Name:      "dnsSearches",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("dnsConfig.searches"),
			Type:      api.MAPPINGFIELDTYPE_STRINGS,
		},
		{
			Name:      "dnsOptions",
			Formatter: podSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "dnsOptions-Value",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("value"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsDnsOptions,
		},
		{
			Name:      "hostAliases",
			Formatter: podSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "hostAliases-Hostnames",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("hostnames"),
			Type:          api.MAPPINGFIELDTYPE_STRINGS,
			ItemSelectors: selectorsHostAliases,
		},
		{
			Name:      "dnsPolicy",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "enableServiceLinks",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		{
			Name:      "hostIPC",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		{
			Name:      "hostNetwork",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		{
			Name:      "hostPID",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		{
			Name:      "hostUsers",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		{
			Name:      "hostname",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "setHostnameAsFQDN",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		{
			Name:      "shareProcessNamespace",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		{
			Name:      "subdomain",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "terminationGracePeriodSeconds",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "nodeAffinity",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("affinity.nodeAffinity"),
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:      "podAffinity",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("affinity.podAffinity"),
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:      "podAntiAffinity",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("affinity.podAntiAffinity"),
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:      "tolerations",
			Formatter: podSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "tolerations-Effect",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("effect"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsTolerations,
		},
		{
			Name:          "tolerations-Operator",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("operator"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsTolerations,
		},
		{
			Name:          "tolerations-Key",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("key"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsTolerations,
		},
		{
			Name:          "tolerations-Value",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("value"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsTolerations,
		},
		{
			Name:          "tolerations-Seconds",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("tolerationSeconds"),
			Type:          api.MAPPINGFIELDTYPE_INTEGER,
			ItemSelectors: selectorsTolerations,
		},
		{
			Name:      "nodeSelectors",
			Formatter: podSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "nodeSelectors-Value",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("value"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsNodeSelectors,
		},
		{
			Name:      "resourceOverheads",
			Formatter: podSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "resourceOverheads-Quantity",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("value"),
			Type:          api.MAPPINGFIELDTYPE_QUANTITY,
			ItemSelectors: selectorsResourceOverhead,
		},
		{
			Name:          "cpu-Overhead",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("overhead.cpu"),
			Type:          api.MAPPINGFIELDTYPE_QUANTITY,
			ItemSelectors: selectorsResourceOverhead,
		},
		{
			Name:          "memory-Overhead",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("overhead.memory"),
			Type:          api.MAPPINGFIELDTYPE_QUANTITY,
			ItemSelectors: selectorsResourceOverhead,
		},
		{
			Name:      "imagePullSecrets",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_NAMES_ARRAY,
		},
		{
			Name:          "imagePullSecrets-Name",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("name"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsImagePullSecrets,
		},
		{
			Name:      "readinessGates",
			Formatter: podSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "readinessGates-ConditionType",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("conditionType"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsReadinessGates,
		},
		{
			Name:      "schedulingGates",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_NAMES_ARRAY,
		},
		{
			Name:          "schedulingGates-Name",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("name"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsSchedulingGates,
		},
		{
			Name:      "resourceClaims",
			Formatter: podSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "resourceClaims-name",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("name"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsResourceClaimsAll,
		},
		{
			Name:          "resourceClaims-resourceClaimName",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("resourceClaimName"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsResourceClaimsDirect,
		},
		{
			Name:          "resourceClaims-resourceClaimTemplateName",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("resourceClaimTemplateName"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsResourceClaimsTemplated,
		},
		{
			Name:      "nodeName",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "preemptionPolicy",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "priority",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "priorityClassName",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "restartPolicy",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "runtimeClassName",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "serviceAccountName",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "fsGroup",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("securityContext.fsGroup"),
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "fsGroupChangePolicy",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("securityContext.fsGroupChangePolicy"),
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "supplementalGroups",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("securityContext.supplementalGroups"),
			Type:      api.MAPPINGFIELDTYPE_INTEGERS,
		},
		{
			Name:      "supplementalGroupsPolicy",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("securityContext.supplementalGroupsPolicy"),
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:          "volumes-Name",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("name"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsVolumesAllKinds,
		},
		{
			Name:      "configMapVolumes",
			Formatter: podSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "configMapVolumes-ConfigMapName",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("configMap.name"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsVolumesConfigMap,
		},
		{
			Name:          "configMapVolumes-DefaultMode",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("configMap.defaultMode"),
			Type:          api.MAPPINGFIELDTYPE_INTEGER,
			ItemSelectors: selectorsVolumesConfigMap,
		},
		{
			Name:          "configMapVolumes-Items",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("configMap.items"),
			Type:          api.MAPPINGFIELDTYPE_OBJECT,
			ItemSelectors: selectorsVolumesConfigMap,
		},
		{
			Name:          "configMapVolumes-Optional",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("configMap.optional"),
			Type:          api.MAPPINGFIELDTYPE_BOOLEAN,
			ItemSelectors: selectorsVolumesConfigMap,
		},
		{
			Name:      "emptyDirVolumes",
			Formatter: podSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "emptyDirVolumes-Medium",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("emptyDir.medium"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsVolumesEmptyDir,
		},
		{
			Name:          "emptyDirVolumes-SizeLimit",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("emptyDir.sizeLimit"),
			Type:          api.MAPPINGFIELDTYPE_QUANTITY,
			ItemSelectors: selectorsVolumesEmptyDir,
		},
		{
			Name:      "hostPathVolumes",
			Formatter: podSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "hostPathVolumes-Path",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("hostPath.path"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsVolumesHostPath,
		},
		{
			Name:          "hostPathVolumes-Type",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("hostPath.type"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsVolumesHostPath,
		},
		{
			Name:      "nfsVolumes",
			Formatter: podSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "nfsVolumes-Server",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("nfs.server"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsVolumesNfs,
		},
		{
			Name:          "nfsVolumes-Path",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("nfs.path"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsVolumesNfs,
		},
		{
			Name:          "nfsVolumes-ReadOnly",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("nfs.readOnly"),
			Type:          api.MAPPINGFIELDTYPE_BOOLEAN,
			ItemSelectors: selectorsVolumesNfs,
		},
		{
			Name:      "pvcVolumes",
			Formatter: podSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "pvcVolumes-ClaimName",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("persistentVolumeClaim.claimName"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsVolumesPvc,
		},
		{
			Name:          "pvcVolumes-ReadOnly",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("persistentVolumeClaim.readOnly"),
			Type:          api.MAPPINGFIELDTYPE_BOOLEAN,
			ItemSelectors: selectorsVolumesPvc,
		},
		{
			Name:      "secretVolumes",
			Formatter: podSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "secretVolumes-SecretName",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("secret.secretName"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsVolumesSecret,
		},
		{
			Name:          "secretVolumes-DefaultMode",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("secret.defaultMode"),
			Type:          api.MAPPINGFIELDTYPE_INTEGER,
			ItemSelectors: selectorsVolumesSecret,
		},
		{
			Name:          "secretVolumes-Items",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("secret.items"),
			Type:          api.MAPPINGFIELDTYPE_OBJECT,
			ItemSelectors: selectorsVolumesSecret,
		},
		{
			Name:          "secretVolumes-Optional",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("secret.optional"),
			Type:          api.MAPPINGFIELDTYPE_BOOLEAN,
			ItemSelectors: selectorsVolumesSecret,
		},
	}
}

func podSpecContainerFgFields() []SystemFgField {
	keyValue := dbmodel.BuiltinFormatterKeyValue
	podSpec := SystemPodSpecFormatterName
	selectorsEnvVars := []string{"allEnvVars", "envVarAsValue", "envVarFromConfigMap", "envVarFromField", "envVarFromSecret", "envVarFromResource"}
	selectorsEnvVarsAsValue := []string{"envVarAsValue"}
	selectorsEnvVarFromConfigMap := []string{"envVarFromConfigMap"}
	selectorsEnvVarFromField := []string{"envVarFromField"}
	selectorsEnvVarFromSecret := []string{"envVarFromSecret"}
	selectorsEnvVarFromResource := []string{"envVarFromResource"}
	selectorsEnvVarBulks := []string{"allEnvVarBulks", "envVarBulksFromConfigMap", "envVarBulksFromSecret"}
	selectorsEnvVarBulksFromConfigMap := []string{"envVarBulksFromConfigMap"}
	selectorsEnvVarBulksFromSecret := []string{"envVarBulksFromSecret"}
	selectorsPorts := []string{"allPorts", "portByContainerPort"}
	selectorsVolumeMounts := []string{"allVolumeMounts", "volumeMountByName", "volumeMountsByReadOnly"}

	return []SystemFgField{
		{
			Name:      "args",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRINGS,
		},
		{
			Name:      "command",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRINGS,
		},
		{
			Name:          "environmentVariables-Name",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("name"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsEnvVars,
		},
		{
			Name:      "environmentVariables-AsValue",
			Formatter: podSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "environmentVariables-Value",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("value"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsEnvVarsAsValue,
		},
		{
			Name:      "environmentVariables-FromConfigMap",
			Formatter: podSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "environmentVariables-ConfigMapName",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("valueFrom.configMapKeyRef.name"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsEnvVarFromConfigMap,
		},
		{
			Name:          "environmentVariables-ConfigMapKey",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("valueFrom.configMapKeyRef.key"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsEnvVarFromConfigMap,
		},
		{
			Name:          "environmentVariables-ConfigMapOptional",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("valueFrom.configMapKeyRef.optional"),
			Type:          api.MAPPINGFIELDTYPE_BOOLEAN,
			ItemSelectors: selectorsEnvVarFromConfigMap,
		},
		{
			Name:      "environmentVariables-FromField",
			Formatter: podSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "environmentVariables-FieldApiVersion",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("valueFrom.fieldRef.apiVersion"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsEnvVarFromField,
		},
		{
			Name:          "environmentVariables-FieldPath",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("valueFrom.fieldRef.fieldPath"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsEnvVarFromField,
		},
		{
			Name:      "environmentVariables-FromSecret",
			Formatter: podSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "environmentVariables-SecretName",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("valueFrom.secretKeyRef.name"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsEnvVarFromSecret,
		},
		{
			Name:          "environmentVariables-SecretKey",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("valueFrom.secretKeyRef.key"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsEnvVarFromSecret,
		},
		{
			Name:          "environmentVariables-SecretOptional",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("valueFrom.secretKeyRef.optional"),
			Type:          api.MAPPINGFIELDTYPE_BOOLEAN,
			ItemSelectors: selectorsEnvVarFromSecret,
		},
		{
			Name:      "environmentVariables-FromResource",
			Formatter: podSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "environmentVariables-ResourceContainerName",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("valueFrom.resourceFieldRef.containerName"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsEnvVarFromResource,
		},
		{
			Name:          "environmentVariables-ResourceDivisor",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("valueFrom.resourceFieldRef.divisor"),
			Type:          api.MAPPINGFIELDTYPE_QUANTITY,
			ItemSelectors: selectorsEnvVarFromResource,
		},
		{
			Name:          "environmentVariables-ResourceKey",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("valueFrom.resourceFieldRef.resource"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsEnvVarFromResource,
		},
		{
			Name:          "environmentVariableBulks-Prefix",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("prefix"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsEnvVarBulks,
		},
		{
			Name:      "environmentVariableBulks-FromConfigMap",
			Formatter: podSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "environmentVariableBulks-ConfigMapName",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("configMapRef.name"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsEnvVarBulksFromConfigMap,
		},
		{
			Name:          "environmentVariableBulks-ConfigMapOptional",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("configMapRef.optional"),
			Type:          api.MAPPINGFIELDTYPE_BOOLEAN,
			ItemSelectors: selectorsEnvVarBulksFromConfigMap,
		},
		{
			Name:      "environmentVariableBulks-FromSecret",
			Formatter: podSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "environmentVariableBulks-SecretName",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("secretRef.name"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsEnvVarBulksFromSecret,
		},
		{
			Name:          "environmentVariableBulks-SecretOptional",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("secretRef.optional"),
			Type:          api.MAPPINGFIELDTYPE_BOOLEAN,
			ItemSelectors: selectorsEnvVarBulksFromSecret,
		},
		{
			Name:      "image",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "imagePullPolicy",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "allowPrivilegeEscalation",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("securityContext.allowPrivilegeEscalation"),
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		{
			Name:      "capabilities-Add",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("securityContext.capabilities.add"),
			Type:      api.MAPPINGFIELDTYPE_STRINGS,
		},
		{
			Name:      "capabilities-Drop",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("securityContext.capabilities.drop"),
			Type:      api.MAPPINGFIELDTYPE_STRINGS,
		},
		{
			Name:      "privileged",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("securityContext.privileged"),
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		{
			Name:      "procMount",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("securityContext.procMount"),
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "readOnlyRootFilesystem",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("securityContext.readOnlyRootFilesystem"),
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		{
			Name:      "ports",
			Formatter: podSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "ports-Name",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("name"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsPorts,
		},
		{
			Name:          "ports-ContainerPort",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("containerPort"),
			Type:          api.MAPPINGFIELDTYPE_INTEGER,
			ItemSelectors: selectorsPorts,
		},
		{
			Name:          "ports-Protocol",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("protocol"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsPorts,
		},
		{
			Name:          "ports-HostPort",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("hostPort"),
			Type:          api.MAPPINGFIELDTYPE_INTEGER,
			ItemSelectors: selectorsPorts,
		},
		{
			Name:          "ports-HostIp",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("hostIP"),
			Type:          api.MAPPINGFIELDTYPE_IPADDRESS,
			ItemSelectors: selectorsPorts,
		},
		{
			Name:      "volumeMounts",
			Formatter: podSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "volumeMounts-Name",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("name"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsVolumeMounts,
		},
		{
			Name:          "volumeMounts-MountPath",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("mountPath"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsVolumeMounts,
		},
		{
			Name:          "volumeMounts-MountPropagation",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("mountPropagation"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsVolumeMounts,
		},
		{
			Name:          "volumeMounts-ReadOnly",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("readOnly"),
			Type:          api.MAPPINGFIELDTYPE_BOOLEAN,
			ItemSelectors: selectorsVolumeMounts,
		},
		{
			Name:          "volumeMounts-RecursiveReadOnly",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("recursiveReadOnly"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsVolumeMounts,
		},
		{
			Name:          "volumeMounts-SubPath",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("subPath"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsVolumeMounts,
		},
		{
			Name:          "volumeMounts-SubPathExpr",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("subPathExpr"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsVolumeMounts,
		},
		{
			Name:      "stdin",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		{
			Name:      "stdinOnce",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		{
			Name:      "terminationMessage-Path",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("terminationMessagePath"),
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "terminationMessage-Policy",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("terminationMessagePolicy"),
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "tty",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		{
			Name:      "workingDir",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
	}
}

func podSpecAndContainerFgFields() []SystemFgField {
	keyValue := dbmodel.BuiltinFormatterKeyValue
	podSpec := SystemPodSpecFormatterName
	selectorsResourceRequests := []string{"allResourceRequests", "resourceRequestByKey"}
	selectorsResourceLimits := []string{"allResourceLimits", "resourceLimitByKey"}

	return []SystemFgField{
		{
			Name:      "cpu-Limit",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("resources.limits.cpu"),
			Type:      api.MAPPINGFIELDTYPE_QUANTITY,
		},
		{
			Name:      "memory-Limit",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("resources.limits.memory"),
			Type:      api.MAPPINGFIELDTYPE_QUANTITY,
		},
		{
			Name:      "cpu-Request",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("resources.requests.cpu"),
			Type:      api.MAPPINGFIELDTYPE_QUANTITY,
		},
		{
			Name:      "memory-Request",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("resources.requests.memory"),
			Type:      api.MAPPINGFIELDTYPE_QUANTITY,
		},
		{
			Name:      "resourceRequests",
			Formatter: podSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "resourceRequests-Quantity",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("value"),
			Type:          api.MAPPINGFIELDTYPE_QUANTITY,
			ItemSelectors: selectorsResourceRequests,
		},
		{
			Name:      "resourceLimits",
			Formatter: podSpec,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "resourceLimits-Quantity",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("value"),
			Type:          api.MAPPINGFIELDTYPE_QUANTITY,
			ItemSelectors: selectorsResourceLimits,
		},
		{
			Name:      "appArmorProfile-Type",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("securityContext.appArmorProfile.type"),
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "appArmorProfile-LocalhostProfile",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("securityContext.appArmorProfile.localhostProfile"),
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "runAsGroup",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("securityContext.runAsGroup"),
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "runAsNonRoot",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("securityContext.runAsNonRoot"),
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		{
			Name:      "runAsUser",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("securityContext.runAsUser"),
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "seccompProfile-Localhost",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("securityContext.seccompProfile.localhostProfile"),
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "seccompProfile-Type",
			Formatter: keyValue,
			SpecRef:   convutils.Ref("securityContext.seccompProfile.type"),
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
	}
}
