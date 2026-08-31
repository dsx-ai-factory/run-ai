/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping_test

// In-memory breakdown builder for Layer 3+4 tests. Walks the system mapping
// (the same Go data exercised by Layer 1/2) and produces a Breakdown the
// gvk_compliance engine can resolve rules against — without going through
// the database or the seeding code path.
//
// The breakdown is built as a generic map[string]any tree that mirrors the
// JSON shape of policyApi.BreakdownBreakdown. Tests JSON-roundtrip the tree
// into the typed struct so we don't import the policy-service api types
// from non-test code; this also keeps the builder naturally compatible with
// future schema changes that only adjust JSON shapes.

import (
	"fmt"
	"sort"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	sm "github.com/run-ai/runai/control-plane/policy-service/pkg/system_mapping"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

// buildLayer3Breakdown constructs a JSON-shaped breakdown for the given
// GVK by walking every spec selector, its field groups, and the formatter
// fields each builtin / custom formatter requires.
//
// The resulting map matches policyApi.BreakdownBreakdown's JSON form:
//
//	{
//	  "fields": {
//	    "<fieldName>": [
//	      {"specSelectors": [...], "formatter": {...}, "specRef": "...",
//	       "type": "...", "itemSelectors": {...}},
//	      ...
//	    ],
//	    ...
//	  },
//	  "specSelectors": {
//	    "<selectorName>": {"description": "...", "query": "...", "argsInfo": {...}, "argsEnum": {...}},
//	    ...
//	  }
//	}
//
// One variation is emitted per (fieldName, fieldGroupName) pair; the
// variation's "specSelectors" lists every spec selector whose FieldGroups
// includes that field group. This mirrors what BreakdownGenerator.Generate
// builds at runtime (see breakdown_generator.go) but is GVK-complete (no
// pruning to "fields actually referenced by rules") because the SET-everything
// test references every field anyway.
func buildLayer3Breakdown(gvk api.Gvk) (map[string]any, error) {
	specSelectors, err := selectGvkSpecSelectors(gvk)
	if err != nil {
		return nil, err
	}

	customFormatters := map[string]map[string]api.FormatterField{
		sm.SystemPodSpecFormatterName:               sm.BuildSystemPodSpecFormatterFields(),
		sm.SystemJobSpecFormatterName:               sm.BuildSystemJobSpecFormatterFields(),
		sm.SystemCronJobSpecFormatterName:           sm.BuildSystemCronJobSpecFormatterFields(),
		sm.SystemSelectorFormatterName:              sm.BuildSystemSelectorFormatterFields(),
		sm.SystemMetadataFormatterName:              sm.BuildSystemMetadataFormatterFields(),
		sm.SystemDeploymentSpecFormatterName:        sm.BuildSystemDeploymentSpecFormatterFields(),
		sm.SystemStatefulSetSpecFormatterName:       sm.BuildSystemStatefulSetSpecFormatterFields(),
		sm.SystemNimCacheSpecFormatterName:          sm.BuildSystemNimCacheSpecFormatterFields(),
		sm.SystemNimServiceSpecFormatterName:        sm.BuildSystemNimServiceSpecFormatterFields(),
		sm.SystemKubeflowRunPolicyFormatterName:     sm.BuildSystemKubeflowRunPolicyFormatterFields(),
		sm.SystemKubeflowReplicaSpecFormatterName:   sm.BuildSystemKubeflowReplicaSpecFormatterFields(),
		sm.SystemPyTorchJobSpecFormatterName:        sm.BuildSystemPyTorchJobSpecFormatterFields(),
		sm.SystemLeaderWorkerSetFormatterName:       sm.BuildSystemLeaderWorkerSetSpecFormatterFields(),
		sm.SystemDynamoGraphDeploymentFormatterName: sm.BuildSystemDynamoGraphDeploymentSpecFormatterFields(),
		sm.SystemRayClusterSpecFormatterName:        sm.BuildSystemRayClusterSpecFormatterFields(),
		sm.SystemRayJobSpecFormatterName:            sm.BuildSystemRayJobSpecFormatterFields(),
		sm.SystemRayServiceSpecFormatterName:        sm.BuildSystemRayServiceSpecFormatterFields(),
		sm.SystemNativeSpecFormatterName:            sm.BuildSystemNativeSpecFormatterFields(),
	}

	variations := map[layer3VariationKey]*layer3PartialVariation{}

	for _, selName := range sm.SortedKeys(specSelectors) {
		sel := specSelectors[selName]
		for _, fgName := range sel.FieldGroups {
			builder, ok := systemFieldGroupBuilders[fgName]
			if !ok {
				return nil, fmt.Errorf("spec selector %q references unknown field group %q", selName, fgName)
			}
			fields, itemSels := builder()
			for _, fName := range sm.SortedKeys(fields) {
				key := layer3VariationKey{fName, fgName}
				if _, exists := variations[key]; !exists {
					variations[key] = &layer3PartialVariation{
						field:         fields[fName],
						itemSelectors: itemSels,
					}
				}
				variations[key].specSelectors = append(variations[key].specSelectors, selName)
			}
		}
	}

	fieldsJSON := map[string]any{}
	for _, key := range sortedVariationKeys(variations) {
		pv := variations[key]
		sort.Strings(pv.specSelectors)

		formatterJSON, err := buildFormatterJSON(pv.field, key.fieldName, customFormatters)
		if err != nil {
			return nil, err
		}

		itemSelectorsJSON := map[string]any{}
		for _, isName := range pv.field.ItemSelectors {
			is, ok := pv.itemSelectors[isName]
			if !ok {
				return nil, fmt.Errorf("field %q references unknown item selector %q", key.fieldName, isName)
			}
			itemSel := map[string]any{
				"query": is.Query,
			}
			if len(is.ArgsInfo) > 0 {
				itemSel["argsInfo"] = mappingFieldTypeMapToJSON(is.ArgsInfo)
			}
			if len(is.ArgsEnum) > 0 {
				itemSel["argsEnum"] = is.ArgsEnum
			}
			itemSelectorsJSON[isName] = itemSel
		}

		variation := map[string]any{
			"specSelectors": pv.specSelectors,
			"formatter":     formatterJSON,
			"specRef":       pv.field.SpecRef,
			"type":          string(pv.field.Type),
		}
		if len(itemSelectorsJSON) > 0 {
			variation["itemSelectors"] = itemSelectorsJSON
		}

		existing, _ := fieldsJSON[key.fieldName].([]any)
		fieldsJSON[key.fieldName] = append(existing, variation)
	}

	specSelectorsJSON := map[string]any{}
	for name, sel := range specSelectors {
		entry := map[string]any{"query": sel.Query}
		if sel.Description != nil {
			entry["description"] = *sel.Description
		}
		if len(sel.ArgsInfo) > 0 {
			entry["argsInfo"] = mappingFieldTypeMapToJSON(sel.ArgsInfo)
		}
		if len(sel.ArgsEnum) > 0 {
			entry["argsEnum"] = sel.ArgsEnum
		}
		if sel.Dependency != nil {
			dep := map[string]any{}
			if sel.Dependency.SpecSelector != nil {
				dep["specSelector"] = *sel.Dependency.SpecSelector
			}
			if sel.Dependency.Path != "" {
				dep["path"] = sel.Dependency.Path
			}
			entry["dependency"] = dep
		}
		specSelectorsJSON[name] = entry
	}

	return map[string]any{
		"fields":        fieldsJSON,
		"specSelectors": specSelectorsJSON,
	}, nil
}

// selectGvkSpecSelectors picks the right system mapping function based on GVK.
// Add new cases here when wiring up additional workload types.
func selectGvkSpecSelectors(gvk api.Gvk) (map[string]api.SpecSelector, error) {
	group := convutils.Deref(gvk.Group, "")
	switch {
	case group == "apps" && gvk.Kind == "Deployment" && gvk.Version == "v1":
		_, sel := sm.BuildSystemAppsDeploymentV1Mapping()
		return sel, nil
	case group == "batch" && gvk.Kind == "Job" && gvk.Version == "v1":
		_, sel := sm.BuildSystemBatchJobV1Mapping()
		return sel, nil
	case group == "apps" && gvk.Kind == "ReplicaSet" && gvk.Version == "v1":
		_, sel := sm.BuildSystemAppsReplicaSetV1Mapping()
		return sel, nil
	case group == "apps" && gvk.Kind == "StatefulSet" && gvk.Version == "v1":
		_, sel := sm.BuildSystemAppsStatefulSetV1Mapping()
		return sel, nil
	case group == "batch" && gvk.Kind == "CronJob" && gvk.Version == "v1":
		_, sel := sm.BuildSystemBatchCronJobV1Mapping()
		return sel, nil
	case group == "" && gvk.Kind == "Pod" && gvk.Version == "v1":
		_, sel := sm.BuildSystemCorePodV1Mapping()
		return sel, nil
	case group == "apps.nvidia.com" && gvk.Kind == "NIMCache" && gvk.Version == "v1alpha1":
		_, sel := sm.BuildSystemAppsNvidiaComNimCacheV1alpha1Mapping()
		return sel, nil
	case group == "apps.nvidia.com" && gvk.Kind == "NIMService" && gvk.Version == "v1alpha1":
		_, sel := sm.BuildSystemAppsNvidiaComNimServiceV1alpha1Mapping()
		return sel, nil
	case group == "kubeflow.org" && gvk.Kind == "PyTorchJob" && gvk.Version == "v1":
		_, sel := sm.BuildSystemKubeflowOrgPyTorchJobV1Mapping()
		return sel, nil
	case group == "kubeflow.org" && gvk.Kind == "MPIJob" && gvk.Version == "v2beta1":
		_, sel := sm.BuildSystemKubeflowOrgMPIJobV2beta1Mapping()
		return sel, nil
	case group == "kubeflow.org" && gvk.Kind == "TFJob" && gvk.Version == "v1":
		_, sel := sm.BuildSystemKubeflowOrgTFJobV1Mapping()
		return sel, nil
	case group == "kubeflow.org" && gvk.Kind == "XGBoostJob" && gvk.Version == "v1":
		_, sel := sm.BuildSystemKubeflowOrgXGBoostJobV1Mapping()
		return sel, nil
	case group == "kubeflow.org" && gvk.Kind == "JAXJob" && gvk.Version == "v1":
		_, sel := sm.BuildSystemKubeflowOrgJAXJobV1Mapping()
		return sel, nil
	case group == "leaderworkerset.x-k8s.io" && gvk.Kind == "LeaderWorkerSet" && gvk.Version == "v1":
		_, sel := sm.BuildSystemLeaderWorkerSetXK8sIoV1Mapping()
		return sel, nil
	case group == "nvidia.com" && gvk.Kind == "DynamoGraphDeployment" && gvk.Version == "v1alpha1":
		_, sel := sm.BuildSystemNvidiaComDynamoGraphDeploymentV1alpha1Mapping()
		return sel, nil
	case group == "ray.io" && gvk.Kind == "RayCluster" && gvk.Version == "v1":
		_, sel := sm.BuildSystemRayIoRayClusterV1Mapping()
		return sel, nil
	case group == "ray.io" && gvk.Kind == "RayJob" && gvk.Version == "v1":
		_, sel := sm.BuildSystemRayIoRayJobV1Mapping()
		return sel, nil
	case group == "ray.io" && gvk.Kind == "RayService" && gvk.Version == "v1":
		_, sel := sm.BuildSystemRayIoRayServiceV1Mapping()
		return sel, nil
	case group == "serving.knative.dev" && gvk.Kind == "Service" && gvk.Version == "v1":
		_, sel := sm.BuildSystemServingKnativeDevServiceV1Mapping()
		return sel, nil
	case group == "run.ai" && gvk.Kind == "Workspace" && gvk.Version == "v1":
		_, sel := sm.BuildSystemRunAiWorkspaceV1Mapping()
		return sel, nil
	case group == "run.ai" && gvk.Kind == "Distributed" && gvk.Version == "v1":
		_, sel := sm.BuildSystemRunAiDistributedV1Mapping()
		return sel, nil
	default:
		return nil, fmt.Errorf("layer 3 breakdown builder: unsupported GVK %s/%s/%s",
			group, gvk.Kind, gvk.Version)
	}
}

// buildFormatterJSON renders a field's formatter into the JSON shape expected
// by BreakdownFormatter. Builtin formatters (KeyValue / KeyValueMap) use just
// {"builtin": "<name>"}; custom formatters are looked up in the per-formatter
// fields map and contribute setExpr / getExpr / arrayKey / argsInfo / argsEnum.
func buildFormatterJSON(
	field api.FieldGroupField,
	fieldName string,
	customFormatters map[string]map[string]api.FormatterField,
) (map[string]any, error) {
	if field.Formatter == dbmodel.BuiltinFormatterKeyValue || field.Formatter == dbmodel.BuiltinFormatterKeyValueMap {
		return map[string]any{"builtin": field.Formatter}, nil
	}

	formatterFields, ok := customFormatters[field.Formatter]
	if !ok {
		return nil, fmt.Errorf("custom formatter %q is not registered", field.Formatter)
	}
	ff, ok := formatterFields[fieldName]
	if !ok {
		return nil, fmt.Errorf("custom formatter %q has no field %q", field.Formatter, fieldName)
	}

	out := map[string]any{}
	if ff.SetExpr != nil {
		out["setExpr"] = *ff.SetExpr
	}
	if ff.GetExpr != nil {
		out["getExpr"] = *ff.GetExpr
	}
	if ff.ArrayKey != nil {
		ak := map[string]any{"argName": ff.ArrayKey.ArgName}
		if ff.ArrayKey.SpecRef != nil {
			ak["specRef"] = *ff.ArrayKey.SpecRef
		}
		out["arrayKey"] = ak
	}
	if len(ff.ArgsInfo) > 0 {
		out["argsInfo"] = mappingFieldTypeMapToJSON(ff.ArgsInfo)
	}
	if len(ff.ArgsEnum) > 0 {
		out["argsEnum"] = ff.ArgsEnum
	}
	return out, nil
}

// mappingFieldTypeMapToJSON converts a map[string]MappingFieldType into a
// JSON-friendly map[string]string. The MappingFieldType type is a Go alias
// over string but the JSON encoder treats it specially during roundtrip,
// so we render it as a plain string here.
func mappingFieldTypeMapToJSON(m map[string]api.MappingFieldType) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = string(v)
	}
	return out
}

// layer3VariationKey identifies a single (fieldGroup, field) pair in the
// breakdown. Two field groups can declare the same field name; in that case
// the breakdown's "fields" array carries one variation per field group.
type layer3VariationKey struct {
	fieldName string
	fgName    string
}

// layer3PartialVariation accumulates the data needed to render one breakdown
// variation: the field's mapping definition, the item selectors of its field
// group (looked up by name later), and the set of spec selectors whose
// FieldGroups list contains this field group.
type layer3PartialVariation struct {
	field         api.FieldGroupField
	itemSelectors map[string]api.ItemSelector
	specSelectors []string
}

// sortedVariationKeys returns the variation keys in a deterministic
// (fieldGroup, field) order so the produced breakdown is reproducible
// run-to-run.
func sortedVariationKeys(m map[layer3VariationKey]*layer3PartialVariation) []layer3VariationKey {
	keys := make([]layer3VariationKey, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].fgName != keys[j].fgName {
			return keys[i].fgName < keys[j].fgName
		}
		return keys[i].fieldName < keys[j].fieldName
	})
	return keys
}
