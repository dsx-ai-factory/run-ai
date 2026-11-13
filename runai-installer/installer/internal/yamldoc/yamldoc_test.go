/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package yamldoc

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func decodeFirst(t *testing.T, contents string) *yaml.Decoder {
	t.Helper()
	dec := yaml.NewDecoder(strings.NewReader(contents))
	var first map[string]any
	require.NoError(t, dec.Decode(&first))
	return dec
}

func TestEnsureNoExtraDocuments(t *testing.T) {
	cases := map[string]struct {
		contents string
		wantErr  bool
	}{
		"single document":        {"a: 1\n", false},
		"trailing separator":     {"a: 1\n---\n", false},
		"trailing comment":       {"a: 1\n# done\n", false},
		"comment-only document":  {"a: 1\n---\n# nothing here\n", false},
		"several empty trailers": {"a: 1\n---\n---\n# end\n", false},
		"document end marker":    {"a: 1\n...\n", false},
		"mapping":                {"a: 1\n---\nb: 2\n", true},
		"scalar":                 {"a: 1\n---\nhello\n", true},
		"explicit null":          {"a: 1\n---\nnull\n", true},
		"tilde null":             {"a: 1\n---\n~\n", true},
		"empty after content":    {"a: 1\n---\n---\nb: 2\n", true},
		"malformed yaml":         {"a: 1\n---\nfoo: [\n", true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := EnsureNoExtraDocuments(decodeFirst(t, tc.contents))
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
