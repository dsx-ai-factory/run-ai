/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package packconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeConfig(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))
	return path
}

func TestLoad_ExtraDocumentRejected(t *testing.T) {
	cases := map[string]string{
		"mapping":        "{}\n---\nunexpectedField: true\n",
		"scalar":         "{}\n---\nhello\n",
		"explicit null":  "{}\n---\nnull\n",
		"tilde null":     "{}\n---\n~\n",
		"malformed yaml": "{}\n---\nfoo: [\n",
	}
	for name, contents := range cases {
		t.Run(name, func(t *testing.T) {
			path := writeConfig(t, contents)
			_, err := Load(path)
			require.Error(t, err)
			assert.Contains(t, err.Error(), path)
		})
	}
}

func TestLoad_TrailingEmptyDocumentsAccepted(t *testing.T) {
	const doc = "settings:\n  output: /tmp/out\n"
	cases := map[string]string{
		"single document":        doc,
		"trailing separator":     doc + "---\n",
		"trailing comment":       doc + "# done\n",
		"comment-only document":  doc + "---\n# nothing here\n",
		"several empty trailers": doc + "---\n---\n# end\n",
		"document end marker":    doc + "...\n",
	}
	for name, contents := range cases {
		t.Run(name, func(t *testing.T) {
			cfg, err := Load(writeConfig(t, contents))
			require.NoError(t, err)
			assert.Equal(t, "/tmp/out", cfg.Settings.Output)
		})
	}
}
