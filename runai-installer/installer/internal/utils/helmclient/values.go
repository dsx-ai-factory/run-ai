/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package helmclient

import (
	"maps"

	"helm.sh/helm/v3/pkg/cli/values"
)

// Values keeps installer-generated overrides separate from native YAML input.
// Custom values are merged last without parsing their keys or string contents.
type Values struct {
	Overrides []string
	Custom    map[string]any
}

func (v Values) Merge(valueFiles []string) (map[string]any, error) {
	opts := values.Options{Values: v.Overrides, ValueFiles: valueFiles}
	base, err := opts.MergeValues(nil)
	if err != nil {
		return nil, err
	}
	return mergeYAMLValues(base, v.Custom), nil
}

func mergeYAMLValues(base, custom map[string]any) map[string]any {
	out := maps.Clone(base)
	if out == nil {
		out = make(map[string]any)
	}
	for key, value := range custom {
		if nested, ok := value.(map[string]any); ok {
			defaults, _ := out[key].(map[string]any)
			out[key] = mergeYAMLValues(defaults, nested)
		} else {
			out[key] = value
		}
	}
	return out
}
