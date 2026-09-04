/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package decoupling holds a guardrail test that keeps dromos free of
// runai-private dependencies. dromos is meant to be open-sourceable: it must
// not import anything from the runai monorepo except its own WorkloadRunner
// CRD module (api/v1alpha1), which is decoupled separately.
//
// If this test fails, you added a dependency on runai-private code. Either
// fork the minimum surface you need into dromos (see pkg/reconcile/issue or
// pkg/utils/karta for the pattern), or reconsider the change.
//
// This guard is only meaningful while dromos lives inside the runai monorepo.
// Once dromos is extracted to its own repository there is nothing to guard
// against and this whole package should be deleted.
package decoupling

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/mod/modfile"
)

// runaiMonorepoPrefix is the import/module prefix for runai-private code.
// github.com/run-ai/karta is a separate published OSS repo and is allowed -
// only the monorepo (.../runai/...) is forbidden.
const runaiMonorepoPrefix = "github.com/dsx-ai-factory/run-ai/"

// allowedModules are the only runai-monorepo modules dromos may depend on.
// api/v1alpha1 is dromos's own WorkloadRunner CRD, vendored locally via a
// replace directive and being decoupled on its own track.
var allowedModules = map[string]bool{
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1": true,
}

// allowedImportPrefixes are import paths that resolve into dromos itself
// (its own module and the local api/v1alpha1 submodule).
var allowedImportPrefixes = []string{
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/",
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate dromos go.mod walking up from test dir")
		}
		dir = parent
	}
}

func TestGoModHasNoRunaiPrivateDeps(t *testing.T) {
	root := moduleRoot(t)
	path := filepath.Join(root, "go.mod")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	mf, err := modfile.Parse(path, data, nil)
	if err != nil {
		t.Fatalf("parse go.mod: %v", err)
	}

	for _, req := range mf.Require {
		mod := req.Mod.Path
		if strings.HasPrefix(mod, runaiMonorepoPrefix) && !allowedModules[mod] {
			t.Errorf("dromos go.mod requires runai-private module %q - dromos must stay decoupled. "+
				"Fork the minimum surface you need instead (see pkg/reconcile/issue for the pattern).", mod)
		}
	}
	for _, rep := range mf.Replace {
		mod := rep.Old.Path
		if strings.HasPrefix(mod, runaiMonorepoPrefix) && !allowedModules[mod] {
			t.Errorf("dromos go.mod replaces runai-private module %q - not allowed.", mod)
		}
	}
}

func TestNoSourceImportsRunaiPrivate(t *testing.T) {
	root := moduleRoot(t)
	fset := token.NewFileSet()

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Skip the local api/v1alpha1 submodule - it is a distinct module
			// with its own decoupling track.
			if d.Name() == "api" && filepath.Dir(path) == root {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		f, perr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if perr != nil {
			t.Errorf("parse %s: %v", path, perr)
			return nil
		}
		for _, imp := range f.Imports {
			p := strings.Trim(imp.Path.Value, `"`)
			if !strings.HasPrefix(p, runaiMonorepoPrefix) {
				continue
			}
			if isAllowedImport(p) {
				continue
			}
			rel, _ := filepath.Rel(root, path)
			t.Errorf("%s imports runai-private package %q - dromos must stay decoupled.", rel, p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
}

func isAllowedImport(p string) bool {
	for _, prefix := range allowedImportPrefixes {
		if strings.HasPrefix(p, prefix) {
			return true
		}
	}
	return false
}
