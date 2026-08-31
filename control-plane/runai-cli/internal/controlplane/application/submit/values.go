/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/strvals"
)

const (
	errValuesFileNotValidYaml = "values file %q is not valid YAML"
	errValuesFileRead         = "failed to read values file %q: %w"
	errSetValueInvalid        = "invalid --set value %q: %w"
)

// BuildValues reads an optional Helm values YAML file and merges the inline
// --set overrides on top of it, mirroring Helm's own precedence. Validation
// of the file is YAML-syntax only; the resulting map is passed to the API
// as spec.values verbatim.
func BuildValues(valuesFilePath string, setValues []string) (map[string]interface{}, error) {
	values := map[string]interface{}{}

	if valuesFilePath != "" {
		content, err := os.ReadFile(valuesFilePath)
		if err != nil {
			return nil, fmt.Errorf(errValuesFileRead, valuesFilePath, err)
		}
		if err = yaml.Unmarshal(content, &values); err != nil {
			return nil, fmt.Errorf(errValuesFileNotValidYaml, valuesFilePath)
		}
		if values == nil {
			values = map[string]interface{}{}
		}
	}

	for _, setValue := range setValues {
		if err := strvals.ParseInto(setValue, values); err != nil {
			return nil, fmt.Errorf(errSetValueInvalid, setValue, err)
		}
	}

	if len(values) == 0 {
		return nil, nil
	}
	return values, nil
}
