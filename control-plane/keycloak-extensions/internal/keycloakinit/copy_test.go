/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package keycloakinit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyExtensions(t *testing.T) {
	t.Run("recursively copies nested files", func(t *testing.T) {
		src := t.TempDir()
		dst := t.TempDir()
		mustWriteFile(t, filepath.Join(src, "a.jar"), "a")
		mustWriteFile(t, filepath.Join(src, "nested", "b.jar"), "b")

		if err := CopyExtensions(map[string]string{src: dst}); err != nil {
			t.Fatalf("CopyExtensions() error = %v", err)
		}

		assertFileContent(t, filepath.Join(dst, "a.jar"), "a")
		assertFileContent(t, filepath.Join(dst, "nested", "b.jar"), "b")
	})

	t.Run("empty source directory produces an empty destination directory", func(t *testing.T) {
		src := t.TempDir()
		dst := t.TempDir()

		if err := CopyExtensions(map[string]string{src: dst}); err != nil {
			t.Fatalf("CopyExtensions() error = %v", err)
		}

		entries, err := os.ReadDir(dst)
		if err != nil {
			t.Fatalf("read dst: %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("dst = %v, want empty", entries)
		}
	})

	t.Run("existing unrelated files in the destination are preserved", func(t *testing.T) {
		src := t.TempDir()
		dst := t.TempDir()
		mustWriteFile(t, filepath.Join(src, "a.jar"), "a")
		mustWriteFile(t, filepath.Join(dst, "preexisting.jar"), "keep-me")

		if err := CopyExtensions(map[string]string{src: dst}); err != nil {
			t.Fatalf("CopyExtensions() error = %v", err)
		}

		assertFileContent(t, filepath.Join(dst, "a.jar"), "a")
		assertFileContent(t, filepath.Join(dst, "preexisting.jar"), "keep-me")
	})

	t.Run("multiple source-to-destination pairs are all copied", func(t *testing.T) {
		providersSrc, providersDst := t.TempDir(), t.TempDir()
		themesSrc, themesDst := t.TempDir(), t.TempDir()
		mustWriteFile(t, filepath.Join(providersSrc, "p.jar"), "p")
		mustWriteFile(t, filepath.Join(themesSrc, "login", "theme.ftl"), "t")

		err := CopyExtensions(map[string]string{
			providersSrc: providersDst,
			themesSrc:    themesDst,
		})
		if err != nil {
			t.Fatalf("CopyExtensions() error = %v", err)
		}

		assertFileContent(t, filepath.Join(providersDst, "p.jar"), "p")
		assertFileContent(t, filepath.Join(themesDst, "login", "theme.ftl"), "t")
	})
}
