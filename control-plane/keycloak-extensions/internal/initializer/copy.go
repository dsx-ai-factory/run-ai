/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package initializer

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// CopyExtensions merges each source directory tree into its destination,
// the way `cp -r src/. dst/` does.
func CopyExtensions(pairs map[string]string) error {
	for src, dst := range pairs {
		if err := copyTree(src, dst); err != nil {
			return fmt.Errorf("copy %s to %s: %w", src, dst, err)
		}
	}
	return nil
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk %s: %w", path, err)
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("relative path for %s: %w", path, err)
		}
		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}

	info, err := in.Stat()
	if err != nil {
		in.Close()
		return fmt.Errorf("stat %s: %w", src, err)
	}

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		in.Close()
		return fmt.Errorf("create %s: %w", dst, err)
	}

	// A failed flush on Close can surface a write error io.Copy didn't see.
	_, copyErr := io.Copy(out, in)
	closeOutErr := out.Close()
	closeInErr := in.Close()

	if copyErr != nil {
		return fmt.Errorf("write %s: %w", dst, copyErr)
	}
	if closeOutErr != nil {
		return fmt.Errorf("close %s: %w", dst, closeOutErr)
	}
	if closeInErr != nil {
		return fmt.Errorf("close %s: %w", src, closeInErr)
	}
	return nil
}
