/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package translation

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/itemized"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

// preprocessDefaults transforms native policy defaults into the value-wrapped
// format used downstream. the purpose of doing it is to "merge" it with the
// rules, so we have one set of values containing both the values and the rules
// "under" each field.
//
// Input (native policy defaults):
//
//	image: nginx
//	compute:
//	  gpuDevicesRequest: 1
//	environmentVariables:
//	  attributes:
//	    value: some-default
//	  instances:
//	    - name: FOO
//	      value: bar
//
// Output (value-wrapped):
//
//	image:
//	  value: nginx
//	compute:
//	  gpuDevicesRequest:
//	    value: 1
//	environmentVariables:
//	  attributes:
//	    value:
//	      value: some-default
//	  instances-values:
//	    - name: FOO
//	      value: bar
//
// As shown above: flat values are wrapped with "value"; nested sections (e.g.
// compute) are recursed into; on itemized fields the "instances" list is renamed
// to "instances-values" and each "attributes" entry is wrapped with "value".
func preprocessDefaults(defaults map[string]any, section string) map[string]any {
	if defaults == nil {
		return nil
	}

	result := make(map[string]any)

	for fieldName, fieldValue := range defaults {
		fullFieldName := convutils.DotUs(section, fieldName)

		// distributed sub-specs (master/worker) are sections that wrap a full
		// spec; recurse into them (keeping the key) instead of value-wrapping.
		// this mirrors the master/worker handling in processFields.
		switch fieldName {
		case "master", "worker":
			result[fieldName] = preprocessDefaults(convutils.MapMe(fieldValue), section)
			continue
		}

		// Recursively handle nested sections (compute, storage, etc.)
		if itemized.NestedPolicySections.Has(fullFieldName) {
			result[fieldName] = preprocessDefaults(convutils.MapMe(fieldValue), fullFieldName)
			continue
		}

		// Check if the current field is itemized (has instances/attributes structure)
		_, isItemized := itemized.GetItemizedFieldInfo(fullFieldName)
		if isItemized {
			result[fieldName] = preprocessItemizedField(fieldValue)
			continue
		}

		// Flat value - wrap with "value"
		result[fieldName] = map[string]any{
			"value": fieldValue,
		}
	}

	return result
}

// preprocessItemizedField handles fields that have instances and/or attributes sections
func preprocessItemizedField(fieldValue any) map[string]any {
	result := make(map[string]any)

	fieldMap, ok := fieldValue.(map[string]any)
	if !ok {
		return result
	}

	// Handle instances array - rename to "instances-values"
	if instances, hasInstances := fieldMap[itemized.PolicyInstances]; hasInstances {
		result["instances-values"] = convutils.ArrayMe(instances)
	}

	// Handle attributes object - wrap each attribute with "value"
	if attributes, hasAttributes := fieldMap[itemized.PolicyAttributes]; hasAttributes {
		attributesMap := convutils.MapMe(attributes)
		wrappedAttributes := make(map[string]any)

		for attrName, attrValue := range attributesMap {
			wrappedAttributes[attrName] = map[string]any{
				"value": attrValue,
			}
		}

		result[itemized.PolicyAttributes] = wrappedAttributes
	}

	return result
}
