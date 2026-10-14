/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package installconfig

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type HelmValues map[string]any

func (v *HelmValues) UnmarshalYAML(node *yaml.Node) error {
	var values map[string]any
	if err := node.Decode(&values); err != nil {
		return err
	}
	for key := range values {
		if strings.ContainsAny(key, ".[\\") {
			return fmt.Errorf("customHelmValues key %q uses Helm override syntax; use nested YAML maps and lists instead", key)
		}
	}
	*v = values
	return nil
}
