/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package integrity

import (
	"os"
	"testing"

	"sigs.k8s.io/yaml"
)

// externalSecretKeys are values-paths whose values name a user-managed Secret.
// Anything found at one of these keys (anywhere in the values tree) is treated
// as an opaque external secret -- the user declared it, the test trusts them
// regardless of whether the chart's params CM round-trips that intent.
var externalSecretKeys = map[string]bool{
	"existingSecret":     true,
	"extraEnvVarsSecret": true,
}

// collectExternalsFromValues walks the provided values files and returns every
// secret name found at an externalSecretKeys path.
func collectExternalsFromValues(t *testing.T, valuesFiles []string) map[string]struct{} {
	t.Helper()
	out := map[string]struct{}{}
	for _, f := range valuesFiles {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read values %q: %v", f, err)
		}
		var v map[string]any
		if err := yaml.Unmarshal(data, &v); err != nil {
			t.Fatalf("parse values %q: %v", f, err)
		}
		walkForExternals(v, out)
	}
	return out
}

func walkForExternals(node any, out map[string]struct{}) {
	switch v := node.(type) {
	case map[string]any:
		for k, child := range v {
			if externalSecretKeys[k] {
				if s, ok := child.(string); ok && s != "" {
					out[s] = struct{}{}
				}
			}
			walkForExternals(child, out)
		}
	case []any:
		for _, item := range v {
			walkForExternals(item, out)
		}
	}
}
