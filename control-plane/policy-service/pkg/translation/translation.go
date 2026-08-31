/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package translation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/itemized"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/datastructs"
	"github.com/run-ai/runai/runai-application-packages/jsonutils/conjungo"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

// sortedKeys returns the keys of a map sorted in ascending order.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

type TranslationContext struct {
	Policy *api.Policy
	Issues []string
}

func (xlateCtx *TranslationContext) AddIssue(issue string, args ...any) {
	xlateCtx.Issues = append(xlateCtx.Issues, fmt.Sprintf(issue, args...))
}

func TranslatePolicy(_ context.Context, workloadType string, metadata api.PolicyCreationFields, spec map[string]any) (*api.Policy, []string, *api.Error) {
	result := api.Policy{
		Description: convutils.Ref(fmt.Sprintf("translation of %s", metadata.Name)),
		WorkloadType: api.Gvk{
			Group:   convutils.Ref("run.ai"),
			Kind:    workloadType,
			Version: "v1",
		},
		Scope:        metadata.Scope,
		ClusterId:    metadata.ClusterId,
		DepartmentId: metadata.DepartmentId,
	}

	if metadata.ProjectId != nil {
		result.ProjectId = convutils.Ref(fmt.Sprintf("%d", *metadata.ProjectId))
	}

	defaultsAsJson := utils.MapMe(spec["defaults"])
	unified := preprocessDefaults(defaultsAsJson, "")

	rulesAsJson := utils.MapMe(spec["rules"])

	err := conjungo.Merge(&unified, rulesAsJson, nil)
	if err != nil {
		return nil, nil, api.NewError(http.StatusInternalServerError, "failed to merge defaults and rules", err)
	}

	xlateCtx := TranslationContext{
		Policy: &result,
		Issues: make([]string, 0),
	}

	defaultSelector := fmt.Sprintf("%sSpec", strings.ToLower(workloadType))

	processFields(&xlateCtx, unified, "", defaultSelector, fieldTranslator)

	return &result, xlateCtx.Issues, nil
}

// FieldHandler is called for each field during processFields traversal
type FieldHandler func(xlateCtx *TranslationContext, fieldName string, specSelector string, value map[string]any)

// processFields recursively scans the unified policy and calls the handler for each field.
// For flat fields: handler receives fieldName (e.g., "image") and its value map.
// For itemized fields: handler receives fieldName (e.g., "storage.git") with flattened instances data
// like {"value": [...], "canAdd": false, "locked": [...]}, then each attribute as fieldName
// (e.g., "storage.git.branch") with its value map.
func processFields(xlateCtx *TranslationContext, unified map[string]any, section string, specSelector string, handler FieldHandler) {
	if unified == nil {
		return
	}

	// Iterate in sorted order to keep the output deterministic across runs.
	for _, fieldName := range sortedKeys(unified) {
		fullFieldName := convutils.DotUs(section, fieldName)
		fieldPolicy := convutils.MapMe(unified[fieldName])

		// switch to sub-sections which require different selector
		switch fieldName {
		case "master", "worker":
			processFields(xlateCtx, fieldPolicy, section, fieldName+"Spec", handler)
			continue
		}

		// Recursively handle nested sections (compute, storage, etc.)
		if itemized.NestedPolicySections.Has(fullFieldName) {
			processFields(xlateCtx, fieldPolicy, fullFieldName, specSelector, handler)
			continue
		}

		// Check if the current field is itemized
		_, isItemized := itemized.GetItemizedFieldInfo(fullFieldName)
		if isItemized {
			// Build flattened instances data: { "value": [...], "canAdd": ..., "locked": ... }
			instancesData := make(map[string]any)
			if instancesValues, ok := fieldPolicy["instances-values"]; ok {
				instancesData["value"] = instancesValues
			}
			// Flatten instances rules (canAdd, locked, etc.) to top level
			if instances, ok := fieldPolicy[itemized.PolicyInstances]; ok {
				for k, v := range convutils.MapMe(instances) {
					instancesData[k] = v
				}
			}
			handler(xlateCtx, fullFieldName, specSelector, instancesData)

			// Process each attribute in sorted order
			if attributes, hasAttributes := fieldPolicy[itemized.PolicyAttributes]; hasAttributes {
				attributesMap := convutils.MapMe(attributes)
				for _, attrName := range sortedKeys(attributesMap) {
					attrFieldName := convutils.DotUs(fullFieldName, attrName)
					handler(xlateCtx, attrFieldName, specSelector, convutils.MapMe(attributesMap[attrName]))
				}
			}
			continue
		}

		// Flat field - call handler directly
		handler(xlateCtx, fullFieldName, specSelector, fieldPolicy)
	}
}

func fieldTranslator(xlateCtx *TranslationContext, fieldName string, specSelector string, nativeDef map[string]any) {
	logging.Infof("Field: %s, Value: %+v", fieldName, nativeDef)

	guidelines, ok := translationBible[fieldName]
	if !ok {
		xlateCtx.AddIssue("No translation guideline found for field %s", fieldName)
	}

	for _, cur := range guidelines {
		rule := api.PolicyRule{}

		if cur.FieldAndSelector != nil {
			rule.Field = cur.FieldAndSelector.Field
			rule.SpecSelector = specSelector

			if cur.FieldAndSelector.ItemSelector != "" {
				rule.ItemSelector = convutils.Ref(cur.FieldAndSelector.ItemSelector)
			}

			rule.SelectorParams = cur.FieldAndSelector.SelectorParams
		}

		enforcements, allKeys, err := processEnforcements(fieldName, cur, nativeDef)
		if err != nil {
			xlateCtx.AddIssue("Failed to process enforcements of %s: %s", fieldName, err.Message)
			continue
		}

		rule.Enforcements = append(rule.Enforcements, enforcements...)

		xlateCtx.Policy.Rules = append(xlateCtx.Policy.Rules, rule)

		// Emit an additional rule against the -Items companion for
		// collection-level constraints that require an item selector
		// (currently canAdd=false → items.Max=0).
		if itemsRule, ok := buildItemsRule(cur, specSelector, allKeys, nativeDef); ok {
			xlateCtx.Policy.Rules = append(xlateCtx.Policy.Rules, itemsRule)
		}
	}
}

// buildItemsRule constructs an options rule with all the allowed keys for canAdd=false
// by limiting the set of key the workload will not be able to add new keys
func buildItemsRule(guideline TranslationGuideline, specSelector string, allKeys datastructs.Set[string], nativeDef map[string]any) (api.PolicyRule, bool) {
	cantAdd := guideline.CantAdd
	if cantAdd == nil {
		return api.PolicyRule{}, false
	}

	enforcements := make([]api.PolicyRuleEnforcement, 0)

	if canAddRaw, hasCanAdd := nativeDef["canAdd"]; hasCanAdd {
		if canAdd, isBool := canAddRaw.(bool); isBool && !canAdd {
			// Set iteration order is non-deterministic; sort for stable output.
			keys := allKeys.Keys()
			sort.Strings(keys)
			enforcements = append(enforcements, api.PolicyRuleEnforcement{
				Options: convutils.Ref(keys),
			})
		}
	}

	if len(enforcements) == 0 {
		return api.PolicyRule{}, false
	}

	rule := api.PolicyRule{
		Field:          cantAdd.Field,
		SpecSelector:   specSelector,
		SelectorParams: cantAdd.SelectorParams,
		Enforcements:   enforcements,
	}
	if cantAdd.ItemSelector != "" {
		rule.ItemSelector = convutils.Ref(cantAdd.ItemSelector)
	}
	return rule, true
}

// translateValue prepares a native default value for use as a Set/Default enforcement payload.
//
//   - If the value is a map, return it as-is, with keys optionally renamed
//     according to guideline.SetParamsRename (source key -> target key).
//     Keys not present in SetParamsRename are kept unchanged.
//   - Otherwise, wrap the value in a map under the "value" key.
func translateValue(guideline TranslationGuideline, value any) map[string]any {
	valueMap, isMap := value.(map[string]any)
	if !isMap {
		return map[string]any{"value": value}
	}

	if len(guideline.SetParamsRename) == 0 {
		return valueMap
	}

	renamed := make(map[string]any, len(valueMap))
	for k, v := range valueMap {
		newKey := k
		if replacement, ok := guideline.SetParamsRename[k]; ok {
			newKey = replacement
		}
		renamed[newKey] = v
	}
	return renamed
}

// itemizedLockContext returns the key attribute name of an itemized field
// (e.g., "resource" for extendedResources, "name" for git) and the set of
// locked key values pulled from nativeDef["locked"]. Returns empty results
// when the field is not itemized.
func itemizedLockContext(fieldName string, nativeDef map[string]any) (string, datastructs.Set[string]) {
	info, isItemized := itemized.GetItemizedFieldInfo(fieldName)
	if !isItemized {
		return "", nil
	}

	lockedRaw, hasLocked := nativeDef["locked"]
	if !hasLocked {
		return info.KeyAttribute, nil
	}

	lockedSet := datastructs.NewSet[string]()
	for _, k := range convutils.ArrayMe(lockedRaw) {
		if s, ok := k.(string); ok {
			lockedSet.Add(s)
		}
	}
	return info.KeyAttribute, lockedSet
}

// isItemLocked returns true if the given instance's key-attribute value is
// present in the locked set. Uses string comparison via fmt for keys that
// aren't already strings.
func isItemLocked(item any, keyAttribute string, lockedSet datastructs.Set[string]) (string, bool) {
	if keyAttribute == "" || len(lockedSet) == 0 {
		return "", false
	}

	itemMap, ok := item.(map[string]any)
	if !ok {
		return "", false
	}

	keyValue, hasKey := itemMap[keyAttribute]
	if !hasKey || keyValue == nil {
		return "", false
	}

	keyStr := fmt.Sprintf("%v", keyValue)

	return keyStr, lockedSet.Has(keyStr)
}

func processEnforcements(fieldName string, guideline TranslationGuideline, nativeDef map[string]any) ([]api.PolicyRuleEnforcement, datastructs.Set[string], *api.Error) {
	enforcements := make([]api.PolicyRuleEnforcement, 0)

	// for itemized, the function also returns a set of all the keys used in the instances
	allItemizedKeys := datastructs.NewSet[string]()

	// Iterate in sorted order for stable enforcement ordering across runs.
	for _, defType := range sortedKeys(nativeDef) {
		defValue := nativeDef[defType]

		switch defType {
		case "value":
			// For itemized fields the value is an array of instance maps.
			// Emit one enforcement per instance so each becomes its own set/default.
			// If the instance's key is in the "locked" list, promote it to Set —
			// analogous to canEdit=false on a flat field.
			if arr, isArray := defValue.([]any); isArray {
				keyAttribute, lockedSet := itemizedLockContext(fieldName, nativeDef)

				for _, item := range arr {
					itemKey, isItemLocked := isItemLocked(item, keyAttribute, lockedSet)

					allItemizedKeys.Add(itemKey)

					enforcement := api.PolicyRuleEnforcement{}
					value := translateValue(guideline, item)

					if isItemLocked {
						enforcement.Set = value
					} else {
						enforcement.Default = value
					}
					enforcements = append(enforcements, enforcement)
				}

				continue
			}

			canEdit, canEditSpecified := nativeDef["canEdit"]
			flatAsSet := canEditSpecified && !canEdit.(bool)

			enforcement := api.PolicyRuleEnforcement{}
			value := translateValue(guideline, defValue)
			if flatAsSet {
				enforcement.Set = value
			} else {
				enforcement.Default = value
			}
			enforcements = append(enforcements, enforcement)

		case "options":
			options := make([]string, 0)

			for _, opt := range convutils.ArrayMe(defValue) {
				options = append(options, convutils.MapMe(opt)["value"].(string))
			}

			enforcements = append(enforcements, api.PolicyRuleEnforcement{
				Options: convutils.Ref(options),
			})

		case "required":
			enforcements = append(enforcements, api.PolicyRuleEnforcement{
				Required: convutils.Ref(defValue.(bool)),
			})

		case "blocked":
			enforcements = append(enforcements, api.PolicyRuleEnforcement{
				Blocked: convutils.Ref(defValue.(bool)),
			})

		case "min", "max", "step":
			// numeric/quantity bounds are grouped into a single Range or
			// QuantityRange enforcement, emitted after the loop.

		case "canAdd", "locked", "canEdit":
			// canAdd: emitted as a separate rule against the -Items companion
			//         via buildItemsRule (needs an item selector).
			// locked/canEdit: consumed while emitting "value" enforcements.

		case "sourceOfRule", "defaultFrom":
			// rule metadata / not-yet-supported enforcement types; skip so they
			// don't fail the translation.

		default:
			return nil, nil, api.NewError(http.StatusInternalServerError, fmt.Sprintf("unexpected native definition type '%s' for %s", defType, guideline.FieldAndSelector.Field), nil)
		}
	}

	// min/max/step describe a single bounded range; emit them as one grouped
	// enforcement (Range for numeric bounds, QuantityRange for k8s quantities).
	if boundsEnf, ok := buildBoundsEnforcement(nativeDef); ok {
		enforcements = append(enforcements, boundsEnf)
	}

	return enforcements, allItemizedKeys, nil
}

// buildBoundsEnforcement groups native min/max/step into a single enforcement.
// Numeric bounds (integer/number rules) map to Range; quantity bounds (k8s
// quantities, expressed as strings such as "1Gi") map to QuantityRange.
func buildBoundsEnforcement(nativeDef map[string]any) (api.PolicyRuleEnforcement, bool) {
	minRaw, hasMin := nativeDef["min"]
	maxRaw, hasMax := nativeDef["max"]
	stepRaw, hasStep := nativeDef["step"]

	if !hasMin && !hasMax && !hasStep {
		return api.PolicyRuleEnforcement{}, false
	}

	// Quantity bounds arrive as strings; numeric bounds arrive as JSON numbers.
	if isStringValue(minRaw) || isStringValue(maxRaw) || isStringValue(stepRaw) {
		qr := &api.PolicyEnforcementQuantityRangeQuantityRange{}
		if hasMin {
			qr.Min = convutils.Ref(toStringValue(minRaw))
		}
		if hasMax {
			qr.Max = convutils.Ref(toStringValue(maxRaw))
		}
		if hasStep {
			qr.Step = convutils.Ref(toStringValue(stepRaw))
		}
		return api.PolicyRuleEnforcement{QuantityRange: qr}, true
	}

	r := &api.PolicyEnforcementRangeRange{}
	if hasMin {
		r.Min = convutils.Ref(toFloat32Value(minRaw))
	}
	if hasMax {
		r.Max = convutils.Ref(toFloat32Value(maxRaw))
	}
	if hasStep {
		r.Step = convutils.Ref(toFloat32Value(stepRaw))
	}
	return api.PolicyRuleEnforcement{Range: r}, true
}

func isStringValue(v any) bool {
	_, ok := v.(string)
	return ok
}

func toStringValue(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func toFloat32Value(v any) float32 {
	switch n := v.(type) {
	case float64:
		return float32(n)
	case float32:
		return n
	case int:
		return float32(n)
	case int32:
		return float32(n)
	case int64:
		return float32(n)
	case json.Number:
		f, _ := n.Float64()
		return float32(f)
	default:
		return 0
	}
}
