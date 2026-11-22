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

func drainAfterFirst(t *testing.T, contents string) error {
	t.Helper()
	dec := yaml.NewDecoder(strings.NewReader(contents))
	var first map[string]any
	require.NoError(t, dec.Decode(&first))
	return RejectExtraDocuments(dec)
}

func TestRejectExtraDocuments_Rejected(t *testing.T) {
	cases := map[string]string{
		"mapping":        "{}\n---\nunexpectedField: true\n",
		"scalar":         "{}\n---\nhello\n",
		"explicit null":  "{}\n---\nnull\n",
		"tilde null":     "{}\n---\n~\n",
		"malformed yaml": "{}\n---\nfoo: [\n",
		"after empty":    "{}\n---\n---\nfoo: bar\n",
	}
	for name, contents := range cases {
		t.Run(name, func(t *testing.T) {
			require.Error(t, drainAfterFirst(t, contents))
		})
	}
}

func TestRejectExtraDocuments_ReportsLine(t *testing.T) {
	err := drainAfterFirst(t, "{}\n---\nunexpectedField: true\n")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "line 3")
}

func TestRejectExtraDocuments_DecodeErrorReportsDocument(t *testing.T) {
	err := drainAfterFirst(t, "{}\n---\n*missing\n")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "document 2")
	assert.Contains(t, err.Error(), "missing")

	err = drainAfterFirst(t, "{}\n---\n---\nfoo: [\n")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "document 3")
}

func TestRejectExtraDocuments_Accepted(t *testing.T) {
	cases := map[string]string{
		"single document":        "a: 1\n",
		"trailing separator":     "a: 1\n---\n",
		"trailing comment":       "a: 1\n# done\n",
		"comment-only document":  "a: 1\n---\n# nothing here\n",
		"several empty trailers": "a: 1\n---\n---\n# end\n",
		"document end marker":    "a: 1\n...\n",
	}
	for name, contents := range cases {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, drainAfterFirst(t, contents))
		})
	}
}
