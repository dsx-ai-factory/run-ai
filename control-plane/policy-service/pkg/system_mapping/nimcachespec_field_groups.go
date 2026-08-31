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

// BuildSystemNimCacheSpecFg builds the "nimCacheSpec" field group, covering the
// non-container fields of apps.nvidia.com/v1alpha1/NIMCache .spec: source,
// storage, proxy, certConfig, runtimeClass, userID/groupID, plus nodeSelector
// and tolerations (which reuse the runai-podSpec custom formatter because the
// array/map shapes are identical to the K8s ones).
//
// Field paths are rooted at the workload .spec (the NIMCache spec selector
// "workloadSpec" points at .spec).
func BuildSystemNimCacheSpecFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	fields := BuildFieldGroupFieldMap(nimCacheSpecFgFields())
	selectors := BuildItemSelectorMap(nimCacheSpecItemSelectorDefs())
	return fields, selectors
}

// BuildSystemNimCacheContainerFg builds the "nimCacheContainer" field group,
// covering the container-like fields of NIMCache .spec that a policy might
// want to shape per-container. Because NIMCache has no explicit container
// (the image is part of the source's modelPuller), "container" here means
// the env vars and resource requirements the caching Job will use.
//
// Field paths are rooted at the workload .spec (the NIMCache spec selector
// "allContainers" points at .spec, same anchor as workloadSpec — differing
// only in the FGs they pull in).
func BuildSystemNimCacheContainerFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	fields := BuildFieldGroupFieldMap(nimCacheContainerFgFields())
	selectors := BuildItemSelectorMap(nimCacheContainerItemSelectorDefs())
	return fields, selectors
}

////////////////////////////////////////////////////////
// spec FG (non-container)
////////////////////////////////////////////////////////

func nimCacheSpecItemSelectorDefs() []SystemItemSel {
	// Reuse the shape of podSpec's item selectors for nodeSelector and
	// tolerations — they work identically here because the array/map
	// structure is the same.
	return []SystemItemSel{
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
	}
}

func nimCacheSpecFgFields() []SystemFgField {
	keyValue := dbmodel.BuiltinFormatterKeyValue
	podSpec := SystemPodSpecFormatterName
	nimCacheSpec := SystemNimCacheSpecFormatterName

	selectorsNodeSelectors := []string{"allNodeSelectors", "nodeSelectorByKey"}
	selectorsTolerations := []string{"allTolerations", "tolerationByEffect", "tolerationByOperator"}

	return []SystemFgField{
		////////////////////////////////////////////////////////
		// spec scalars
		////////////////////////////////////////////////////////
		{Name: "runtimeClassName", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "userID", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_INTEGER},
		{Name: "groupID", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_INTEGER},

		////////////////////////////////////////////////////////
		// spec.certConfig
		////////////////////////////////////////////////////////
		{Name: "certConfig-Name", Formatter: keyValue, SpecRef: convutils.Ref("certConfig.name"), Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "certConfig-MountPath", Formatter: keyValue, SpecRef: convutils.Ref("certConfig.mountPath"), Type: api.MAPPINGFIELDTYPE_STRING},

		////////////////////////////////////////////////////////
		// spec.proxy
		////////////////////////////////////////////////////////
		{Name: "proxy-HttpProxy", Formatter: keyValue, SpecRef: convutils.Ref("proxy.httpProxy"), Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "proxy-HttpsProxy", Formatter: keyValue, SpecRef: convutils.Ref("proxy.httpsProxy"), Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "proxy-NoProxy", Formatter: keyValue, SpecRef: convutils.Ref("proxy.noProxy"), Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "proxy-CertConfigMap", Formatter: keyValue, SpecRef: convutils.Ref("proxy.certConfigMap"), Type: api.MAPPINGFIELDTYPE_STRING},

		////////////////////////////////////////////////////////
		// spec.source – object-level fields (set entire source type at once)
		//
		// The three source types (ngc, dataStore, hf) are mutually exclusive
		// in a real NIMCache; here we define all three so policies can choose.
		// Each uses the runai-nimCacheSpec formatter for object-level SET.
		////////////////////////////////////////////////////////
		{Name: "ngcSource", Formatter: nimCacheSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "dataStoreSource", Formatter: nimCacheSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "hfSource", Formatter: nimCacheSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},

		////////////////////////////////////////////////////////
		// spec.storage – object-level fields (set entire storage type at once)
		//
		// pvcStorage and hostPathStorage both include the readOnly flag, which
		// maps to the same storage.readOnly path. This is intentional: setting
		// readOnly via either field has the same effect. Policy writers should
		// be aware that if both storage types are (unusually) configured,
		// setting readOnly on one will affect the other.
		////////////////////////////////////////////////////////
		{Name: "pvcStorage", Formatter: nimCacheSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "hostPathStorage", Formatter: nimCacheSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},

		////////////////////////////////////////////////////////
		// spec.nodeSelector  (reuse runai-podSpec)
		////////////////////////////////////////////////////////
		{Name: "nodeSelectors", Formatter: podSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{
			Name:          "nodeSelectors-Value",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("value"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsNodeSelectors,
		},

		////////////////////////////////////////////////////////
		// spec.tolerations (reuse runai-podSpec)
		////////////////////////////////////////////////////////
		{Name: "tolerations", Formatter: podSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
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
	}
}

////////////////////////////////////////////////////////
// container FG (env + resources)
////////////////////////////////////////////////////////

func nimCacheContainerItemSelectorDefs() []SystemItemSel {
	// Same env item selectors as runai-podSpec's container FG; here they
	// operate on .env of the NIMCache .spec (the anchor the allContainers
	// spec selector binds to).
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
	}
}

func nimCacheContainerFgFields() []SystemFgField {
	keyValue := dbmodel.BuiltinFormatterKeyValue
	podSpec := SystemPodSpecFormatterName

	selectorsEnvVars := []string{"allEnvVars", "envVarAsValue", "envVarFromConfigMap", "envVarFromField", "envVarFromSecret", "envVarFromResource"}
	selectorsEnvVarsAsValue := []string{"envVarAsValue"}
	selectorsEnvVarFromConfigMap := []string{"envVarFromConfigMap"}
	selectorsEnvVarFromField := []string{"envVarFromField"}
	selectorsEnvVarFromSecret := []string{"envVarFromSecret"}
	selectorsEnvVarFromResource := []string{"envVarFromResource"}

	return []SystemFgField{
		////////////////////////////////////////////////////////
		// spec.env (reuse runai-podSpec env formatters)
		////////////////////////////////////////////////////////
		{
			Name:          "environmentVariables-Name",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("name"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsEnvVars,
		},
		{Name: "environmentVariables-AsValue", Formatter: podSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{
			Name:          "environmentVariables-Value",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("value"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsEnvVarsAsValue,
		},
		{Name: "environmentVariables-FromConfigMap", Formatter: podSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
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
		{Name: "environmentVariables-FromField", Formatter: podSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
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
		{Name: "environmentVariables-FromSecret", Formatter: podSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
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
		{Name: "environmentVariables-FromResource", Formatter: podSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
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

		////////////////////////////////////////////////////////
		// spec.resources  (NIMCache uses its own Resources
		// {CPU, Memory} quantities — NOT corev1.ResourceRequirements;
		// path is .resources.cpu / .resources.memory, flat scalars.)
		////////////////////////////////////////////////////////
		{Name: "cpu-Resource", Formatter: keyValue, SpecRef: convutils.Ref("resources.cpu"), Type: api.MAPPINGFIELDTYPE_QUANTITY},
		{Name: "memory-Resource", Formatter: keyValue, SpecRef: convutils.Ref("resources.memory"), Type: api.MAPPINGFIELDTYPE_QUANTITY},
	}
}
