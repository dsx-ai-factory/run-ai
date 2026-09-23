/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package helm_deployer

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/helmvalues"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
)

func applyCredentialBindings(loaded *chart.Chart, overrides map[string]interface{}, bindings []v1alpha1.HelmCredentialBinding) (map[string]interface{}, error) {
	if len(bindings) == 0 {
		return overrides, nil
	}
	if loaded == nil {
		return nil, fmt.Errorf("chart is required to apply credential bindings")
	}
	variables := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		variables = append(variables, binding.Variable)
		if binding.SecretName == "" {
			return nil, fmt.Errorf("credential destination %q has no resolved Secret name", binding.Variable)
		}
	}
	if err := helmvalues.ValidatePaths(variables); err != nil {
		return nil, err
	}
	// Dependency imports can supply the array defaults that an indexed binding must preserve.
	if err := chartutil.ProcessDependenciesWithMerge(loaded, overrides); err != nil {
		return nil, fmt.Errorf("process chart dependencies for credentials: %w", err)
	}
	// Keep explicit nulls until Helm's final coalescing, so deleted defaults stay deleted.
	effective, err := chartutil.MergeValues(loaded, overrides)
	if err != nil {
		return nil, fmt.Errorf("merge chart values for credentials: %w", err)
	}
	raw, err := json.Marshal(overrides)
	if err != nil {
		return nil, fmt.Errorf("copy Helm overrides: %w", err)
	}
	result := map[string]interface{}{}
	if len(overrides) > 0 {
		if err := json.Unmarshal(raw, &result); err != nil {
			return nil, fmt.Errorf("copy Helm overrides: %w", err)
		}
	}
	type assignment struct {
		binding v1alpha1.HelmCredentialBinding
		path    helmvalues.Path
	}
	assignments := make([]assignment, 0, len(bindings))
	for _, binding := range bindings {
		path, err := helmvalues.Parse(binding.Variable)
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment{binding: binding, path: path})
	}
	slices.SortFunc(assignments, func(a, b assignment) int { return a.path.Compare(b.path) })
	for _, assignment := range assignments {
		binding, path := assignment.binding, assignment.path
		if err := path.Set(effective, binding.SecretName); err != nil {
			return nil, fmt.Errorf("set credential destination %q: %w", binding.Variable, err)
		}
		if array, indexed := path.ArrayParent(); indexed {
			// Helm replaces arrays. Carry the whole effective array, including untouched chart defaults.
			value, err := array.Get(effective)
			if err != nil {
				return nil, err
			}
			if err := array.Set(result, value); err != nil {
				return nil, err
			}
		} else if err := path.Set(result, binding.SecretName); err != nil {
			return nil, err
		}
		if flag := path.CreateFlag(); flag != "" {
			companion, err := helmvalues.Parse(flag)
			if err != nil {
				return nil, err
			}
			if err := companion.Set(result, false); err != nil {
				return nil, fmt.Errorf("disable chart Secret creation: %w", err)
			}
		}
	}
	return result, nil
}
