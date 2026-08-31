/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package upgrader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type fakeCRDClient struct {
	patch func(context.Context, string, types.PatchType, []byte, metav1.PatchOptions) (*apiextensionsv1.CustomResourceDefinition, error)
	get   func(context.Context, string, metav1.GetOptions) (*apiextensionsv1.CustomResourceDefinition, error)
}

func (f *fakeCRDClient) Patch(ctx context.Context, name string, patchType types.PatchType, data []byte, options metav1.PatchOptions, _ ...string) (*apiextensionsv1.CustomResourceDefinition, error) {
	return f.patch(ctx, name, patchType, data, options)
}

func (f *fakeCRDClient) Get(ctx context.Context, name string, options metav1.GetOptions) (*apiextensionsv1.CustomResourceDefinition, error) {
	return f.get(ctx, name, options)
}

func TestRunAppliesWithForceAndWaitsForEstablished(t *testing.T) {
	directory := t.TempDir()
	writeManifest(t, directory, "b.yaml", "second.example.com")
	writeManifest(t, directory, "a.yaml", "first.example.com")

	var applied []string
	client := &fakeCRDClient{
		patch: func(_ context.Context, name string, patchType types.PatchType, data []byte, options metav1.PatchOptions) (*apiextensionsv1.CustomResourceDefinition, error) {
			if patchType != types.ApplyPatchType {
				t.Fatalf("patch type = %q, want %q", patchType, types.ApplyPatchType)
			}
			if options.FieldManager != FieldManager {
				t.Fatalf("field manager = %q, want %q", options.FieldManager, FieldManager)
			}
			if options.Force == nil || !*options.Force {
				t.Fatal("server-side apply did not force ownership")
			}
			if !strings.Contains(string(data), `"managed-by":"test"`) {
				t.Fatalf("apply payload did not preserve manifest annotations: %s", data)
			}
			applied = append(applied, name)
			return &apiextensionsv1.CustomResourceDefinition{}, nil
		},
		get: func(_ context.Context, name string, _ metav1.GetOptions) (*apiextensionsv1.CustomResourceDefinition, error) {
			return establishedCRD(name), nil
		},
	}

	u := New(client, directory, time.Second, time.Millisecond, log.New(io.Discard, "", 0))
	if err := u.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	want := []string{"first.example.com", "second.example.com"}
	if fmt.Sprint(applied) != fmt.Sprint(want) {
		t.Fatalf("apply order = %v, want %v", applied, want)
	}
}

func TestRunReportsRejectedNames(t *testing.T) {
	directory := t.TempDir()
	writeManifest(t, directory, "crd.yaml", "widgets.example.com")

	client := &fakeCRDClient{
		patch: func(_ context.Context, _ string, _ types.PatchType, _ []byte, _ metav1.PatchOptions) (*apiextensionsv1.CustomResourceDefinition, error) {
			return &apiextensionsv1.CustomResourceDefinition{}, nil
		},
		get: func(_ context.Context, name string, _ metav1.GetOptions) (*apiextensionsv1.CustomResourceDefinition, error) {
			return &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{Name: name},
				Status: apiextensionsv1.CustomResourceDefinitionStatus{
					Conditions: []apiextensionsv1.CustomResourceDefinitionCondition{{
						Type:    apiextensionsv1.NamesAccepted,
						Status:  apiextensionsv1.ConditionFalse,
						Reason:  "NameConflict",
						Message: "plural name is already in use",
					}},
				},
			}, nil
		},
	}

	u := New(client, directory, time.Second, time.Millisecond, log.New(io.Discard, "", 0))
	err := u.Run(context.Background())
	if err == nil || !strings.Contains(err.Error(), "NameConflict") || !strings.Contains(err.Error(), "widgets.example.com") {
		t.Fatalf("Run() error = %v, want CRD name and rejection reason", err)
	}
}

func TestWaitUntilEstablishedReportsLastRetryableGetError(t *testing.T) {
	getErr := apierrors.NewServiceUnavailable("apiserver overloaded")
	client := &fakeCRDClient{
		get: func(_ context.Context, _ string, _ metav1.GetOptions) (*apiextensionsv1.CustomResourceDefinition, error) {
			return nil, getErr
		},
	}

	u := New(client, "", 10*time.Millisecond, time.Millisecond, log.New(io.Discard, "", 0))
	err := u.waitUntilEstablished(context.Background(), []manifest{{name: "widgets.example.com"}})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("waitUntilEstablished() error = %v, want deadline exceeded", err)
	}
	if !errors.Is(err, getErr) {
		t.Fatalf("waitUntilEstablished() error = %v, want last Get error", err)
	}
	if !strings.Contains(err.Error(), "last Get error for CRD widgets.example.com") {
		t.Fatalf("waitUntilEstablished() error = %v, want affected CRD", err)
	}
}

func TestLoadBundledCRDs(t *testing.T) {
	loaded, err := loadManifests(filepath.Join("..", "..", "crds"))
	if err != nil {
		t.Fatalf("loadManifests() error = %v", err)
	}
	if len(loaded.manifests) == 0 {
		t.Fatal("loadManifests() returned no bundled CRDs")
	}
	if len(loaded.ignored) != 0 {
		t.Fatalf("loadManifests() ignored bundled documents: %v", loaded.ignored)
	}

	for _, crd := range loaded.manifests {
		if crd.name == "" || len(crd.data) == 0 {
			t.Fatalf("invalid decoded manifest: %+v", crd)
		}
	}
}

func TestReleaseBundlesExcludeLegacyCRDs(t *testing.T) {
	bundles := map[string]string{
		"chart":    filepath.Join("..", "..", "..", "runai-chart", "charts", "crds", "crds"),
		"upgrader": filepath.Join("..", "..", "crds"),
	}

	for bundleName, directory := range bundles {
		loaded, err := loadManifests(directory)
		if err != nil {
			t.Fatalf("loadManifests(%s bundle) error = %v", bundleName, err)
		}
		if len(loaded.manifests) == 0 {
			t.Fatalf("%s release bundle contains no CRDs", bundleName)
		}
		if len(loaded.ignored) != 0 {
			t.Fatalf("%s release bundle ignored documents: %v", bundleName, loaded.ignored)
		}
		for _, crd := range loaded.manifests {
			if legacyCRDNames[crd.name] {
				t.Errorf("%s release bundle contains legacy CRD %q from %s", bundleName, crd.name, crd.source)
			}
		}
	}
}

func TestBundledCRDsDoNotApplyLegacyNames(t *testing.T) {
	client := &fakeCRDClient{
		patch: func(_ context.Context, name string, _ types.PatchType, _ []byte, _ metav1.PatchOptions) (*apiextensionsv1.CustomResourceDefinition, error) {
			if legacyCRDNames[name] {
				t.Fatalf("upgrader attempted to apply legacy CRD %q", name)
			}
			return establishedCRD(name), nil
		},
		get: func(_ context.Context, name string, _ metav1.GetOptions) (*apiextensionsv1.CustomResourceDefinition, error) {
			return establishedCRD(name), nil
		},
	}

	u := New(client, filepath.Join("..", "..", "crds"), time.Second, time.Millisecond, log.New(io.Discard, "", 0))
	if err := u.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

var legacyCRDNames = map[string]bool{
	"distributedinferencepolicies.run.ai":  true,
	"distributedinferenceworkloads.run.ai": true,
	"distributedpolicies.run.ai":           true,
	"distributedworkloads.run.ai":          true,
	"inferencepolicies.run.ai":             true,
	"inferenceworkloads.run.ai":            true,
	"interactivepolicies.run.ai":           true,
	"interactiveworkloads.run.ai":          true,
	"runaijobs.run.ai":                     true,
	"trainingpolicies.run.ai":              true,
	"trainingworkloads.run.ai":             true,
}

func TestLoadManifestsIgnoresNonCRD(t *testing.T) {
	directory := t.TempDir()
	writeManifest(t, directory, "crd.yaml", "widgets.example.com")
	contents := []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: not-a-crd\n")
	if err := os.WriteFile(filepath.Join(directory, "configmap.yaml"), contents, 0o600); err != nil {
		t.Fatal(err)
	}

	loaded, err := loadManifests(directory)
	if err != nil {
		t.Fatalf("loadManifests() error = %v", err)
	}
	if len(loaded.manifests) != 1 || len(loaded.ignored) != 1 || !strings.Contains(loaded.ignored[0], "ConfigMap") {
		t.Fatalf("loadManifests() = %+v, want one CRD and one ignored ConfigMap", loaded)
	}
}

func TestLoadManifestsRejectsUnsupportedCRDVersion(t *testing.T) {
	directory := t.TempDir()
	contents := []byte("apiVersion: apiextensions.k8s.io/v1beta1\nkind: CustomResourceDefinition\nmetadata:\n  name: widgets.example.com\n")
	if err := os.WriteFile(filepath.Join(directory, "crd.yaml"), contents, 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := loadManifests(directory)
	if err == nil || !strings.Contains(err.Error(), `unsupported CRD apiVersion "apiextensions.k8s.io/v1beta1"`) {
		t.Fatalf("loadManifests() error = %v", err)
	}
}

func writeManifest(t *testing.T, directory, filename, name string) {
	t.Helper()
	contents := fmt.Sprintf(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: %s
  annotations:
    managed-by: test
spec:
  group: example.com
  names:
    kind: Widget
    plural: widgets
  scope: Cluster
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
`, name)
	if err := os.WriteFile(filepath.Join(directory, filename), []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}

func establishedCRD(name string) *apiextensionsv1.CustomResourceDefinition {
	return &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status: apiextensionsv1.CustomResourceDefinitionStatus{
			Conditions: []apiextensionsv1.CustomResourceDefinitionCondition{{
				Type:   apiextensionsv1.Established,
				Status: apiextensionsv1.ConditionTrue,
			}},
		},
	}
}
