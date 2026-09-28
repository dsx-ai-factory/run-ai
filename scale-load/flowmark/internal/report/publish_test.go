/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package report

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
)

type fakeStore struct {
	objects map[string][]byte
	failPut map[string]error
	failGet map[string]error
	corrupt map[string][]byte
}

func newFakeStore() *fakeStore {
	return &fakeStore{objects: map[string][]byte{}, failPut: map[string]error{}, failGet: map[string]error{}, corrupt: map[string][]byte{}}
}

func (f *fakeStore) Put(_ context.Context, name string, data []byte) error {
	if err := f.failPut[name]; err != nil {
		return err
	}
	f.objects[name] = append([]byte(nil), data...)
	return nil
}

func (f *fakeStore) Get(_ context.Context, name string) ([]byte, error) {
	if err := f.failGet[name]; err != nil {
		return nil, err
	}
	if data, ok := f.corrupt[name]; ok {
		return data, nil
	}
	data, ok := f.objects[name]
	if !ok {
		return nil, ErrNotFound
	}
	return data, nil
}

func (f *fakeStore) String() string { return "fake" }

func document(second int) Document {
	return Build(map[string]core.Section{"s": {Category: core.CategoryScale, Metrics: map[string]any{"x": 1}}},
		RunInfo{StartedAt: time.Date(2026, 9, 28, 1, 0, second, 0, time.UTC), Provider: "local", ClusterType: "local"})
}

func TestPublishStoresResultsAndManifestAndVerifies(t *testing.T) {
	store := newFakeStore()

	receipt, err := Publish(context.Background(), store, document(0))
	if err != nil {
		t.Fatal(err)
	}

	if receipt.File != "results-20260928T010000Z.json" || receipt.Runs != 1 {
		t.Fatalf("unexpected receipt %+v", receipt)
	}
	if _, ok := store.objects[receipt.File]; !ok {
		t.Fatal("results file must be stored")
	}
	manifest, err := parseManifest(store.objects[ManifestName])
	if err != nil || !manifest.contains(receipt.File) {
		t.Fatalf("manifest must list the new file: %v %v", manifest, err)
	}
}

func TestPublishAppendsToExistingManifestAndIsIdempotent(t *testing.T) {
	store := newFakeStore()
	for _, second := range []int{0, 1, 1} {
		if _, err := Publish(context.Background(), store, document(second)); err != nil {
			t.Fatal(err)
		}
	}
	manifest, err := parseManifest(store.objects[ManifestName])
	if err != nil || len(manifest.Runs) != 2 {
		t.Fatalf("two distinct runs must give two manifest entries even after a re-run, got %v (%v)", manifest.Runs, err)
	}
}

func TestPublishFailsLoudNamingTheStep(t *testing.T) {
	name := document(0).FileName()
	cases := []struct {
		name    string
		prepare func(*fakeStore)
		wantErr string
	}{
		{name: "results write fails", prepare: func(f *fakeStore) { f.failPut[name] = errors.New("disk full") }, wantErr: "write results " + name + ": disk full"},
		{name: "manifest read fails", prepare: func(f *fakeStore) { f.failGet[ManifestName] = errors.New("permission denied") }, wantErr: "read manifest: permission denied"},
		{name: "existing manifest is corrupt", prepare: func(f *fakeStore) { f.objects[ManifestName] = []byte("{oops") }, wantErr: "parse manifest:"},
		{name: "manifest write fails", prepare: func(f *fakeStore) { f.failPut[ManifestName] = errors.New("permission denied") }, wantErr: "write manifest: permission denied"},
		{name: "results read back differ", prepare: func(f *fakeStore) { f.corrupt[name] = []byte("{}") }, wantErr: "verify results " + name + ": stored content differs"},
		{name: "manifest lost the entry", prepare: func(f *fakeStore) { f.corrupt[ManifestName] = []byte(`{"runs":[]}`) }, wantErr: "verify manifest: " + name + " is missing"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := newFakeStore()
			tc.prepare(store)
			_, err := Publish(context.Background(), store, document(0))
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestLocalStoreRoundTripAndNotFound(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "out")
	store, err := NewLocal(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(context.Background(), "missing.json"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("a missing file must be ErrNotFound, got %v", err)
	}
	if err := store.Put(context.Background(), "a.json", []byte("{}")); err != nil {
		t.Fatal(err)
	}
	got, err := store.Get(context.Background(), "a.json")
	if err != nil || string(got) != "{}" {
		t.Fatalf("round trip failed: %q %v", got, err)
	}
	if info, err := os.Stat(filepath.Join(dir, "a.json")); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("files must be private to the user: %v %v", info, err)
	}
}

func TestPublishToLocalDirectoryEndToEnd(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocal(dir)
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := Publish(context.Background(), store, document(0))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{receipt.File, ManifestName} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("%s must exist in the directory: %v", name, err)
		}
	}
}
