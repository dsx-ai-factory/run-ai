/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package artifactlock

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const minimalLock = "apiVersion: artifacts.run.ai/v1alpha1\nkind: ArtifactLock\nmetadata:\n  version: 2.26.0\n"

func TestParse_ExtraDocumentRejected(t *testing.T) {
	cases := map[string]string{
		"mapping":        minimalLock + "---\nunexpectedField: true\n",
		"second lock":    minimalLock + "---\n" + minimalLock,
		"explicit null":  minimalLock + "---\nnull\n",
		"tilde null":     minimalLock + "---\n~\n",
		"malformed yaml": minimalLock + "---\nfoo: [\n",
	}
	for name, contents := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := Parse(strings.NewReader(contents), "lock.yaml")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "lock.yaml")
		})
	}
}

func TestParse_TrailingEmptyDocumentsAccepted(t *testing.T) {
	cases := map[string]string{
		"single document":        minimalLock,
		"trailing separator":     minimalLock + "---\n",
		"trailing comment":       minimalLock + "# done\n",
		"comment-only document":  minimalLock + "---\n# nothing here\n",
		"several empty trailers": minimalLock + "---\n---\n# end\n",
		"document end marker":    minimalLock + "...\n",
	}
	for name, contents := range cases {
		t.Run(name, func(t *testing.T) {
			lock, err := Parse(strings.NewReader(contents), "lock.yaml")
			require.NoError(t, err)
			assert.Equal(t, "2.26.0", lock.Metadata.Version)
		})
	}
}
