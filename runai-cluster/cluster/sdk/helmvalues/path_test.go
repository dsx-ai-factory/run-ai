// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package helmvalues

import (
	"reflect"
	"strings"
	"testing"
)

func TestSetPreservesArrayAndLiteralSiblings(t *testing.T) {
	values := map[string]interface{}{
		"apps": map[string]interface{}{
			"backend": []interface{}{
				map[string]interface{}{"name": "old", "keep": true},
				map[string]interface{}{"name": "second"},
			},
		},
		"literal.key": map[string]interface{}{"keep": "sibling"},
	}
	for _, variable := range []string{"apps.backend[0].name", `literal\.key["secret.name"]`} {
		path, err := Parse(variable)
		if err != nil {
			t.Fatal(err)
		}
		if err := path.Set(values, "selected"); err != nil {
			t.Fatal(err)
		}
	}
	want := map[string]interface{}{
		"apps": map[string]interface{}{
			"backend": []interface{}{
				map[string]interface{}{"name": "selected", "keep": true},
				map[string]interface{}{"name": "second"},
			},
		},
		"literal.key": map[string]interface{}{"keep": "sibling", "secret.name": "selected"},
	}
	if !reflect.DeepEqual(values, want) {
		t.Fatalf("unexpected values: %#v", values)
	}
}

func TestValidatePaths(t *testing.T) {
	for _, tc := range []struct {
		name  string
		paths []string
		valid bool
	}{
		{"independent array entries", []string{"pulls[0].name", "pulls[1].name"}, true},
		{"incompatible sibling containers", []string{"pulls[0].name", "pulls.name"}, false},
		{"two blueprint fields", []string{"global.imagePullSecret.name", "global.ngcApiSecret.name"}, true},
		{"normalized duplicate", []string{`a.b`, `a["b"]`}, false},
		{"normalized index duplicate", []string{"pulls[01].name", "pulls[1].name"}, false},
		{"parent collision", []string{"pulls", "pulls[0].name"}, false},
		{"companion collision", []string{"global.imagePullSecret.name", "global.imagePullSecret.create"}, false},
		{"empty path", []string{""}, false},
		{"empty key", []string{"a..b"}, false},
		{"root array", []string{"[0].name"}, false},
		{"negative index", []string{"a[-1]"}, false},
		{"excessive index", []string{"a[999999999]"}, false},
		{"trailing escape", []string{`a\`}, false},
		{"unterminated quoted key", []string{`a["key\`}, false},
		{"missing separator", []string{"a[0]name"}, false},
		{"excessive nesting", []string{strings.Repeat("a.", 32) + "b"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePaths(tc.paths)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v, error=%v", tc.valid, err)
			}
		})
	}
}

func TestSetRejectsScalarParentsAndSparseArrays(t *testing.T) {
	for _, tc := range []struct {
		variable string
		values   map[string]interface{}
	}{
		{"a.name", map[string]interface{}{"a": "scalar"}},
		{"a[0].name", map[string]interface{}{"a": map[string]interface{}{}}},
		{"a[2].name", map[string]interface{}{"a": []interface{}{}}},
	} {
		t.Run(tc.variable, func(t *testing.T) {
			path, err := Parse(tc.variable)
			if err != nil {
				t.Fatal(err)
			}
			if err := path.Set(tc.values, "selected"); err == nil {
				t.Fatal("expected incompatible destination to fail")
			}
		})
	}
}

func TestBlueprintCompanionOnlyAppliesToExactDestination(t *testing.T) {
	for variable, want := range map[string]string{
		"global.imagePullSecret.name":     "global.imagePullSecret.create",
		"global.ngcApiSecret.name":        "global.ngcApiSecret.create",
		"custom.imagePullSecret.name":     "",
		`["global.imagePullSecret.name"]`: "",
	} {
		t.Run(variable, func(t *testing.T) {
			path, err := Parse(variable)
			if err != nil {
				t.Fatal(err)
			}
			if got := path.CreateFlag(); got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}
}
