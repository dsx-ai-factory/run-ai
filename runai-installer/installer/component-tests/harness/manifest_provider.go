/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package harness

import (
	"context"
	"sync"
)

// ManifestProviderCall records one ManifestProvider invocation.
type ManifestProviderCall struct {
	Op         string // "kustomize" | "fetch"
	TarballURL string
	SubPath    string
	URL        string
}

// FakeManifestProvider is a hermetic orchestrator2.ManifestProvider for tests:
// it records calls and returns canned bytes instead of hitting the network or
// running kustomize.
type FakeManifestProvider struct {
	mu        sync.Mutex
	Calls     []ManifestProviderCall
	RenderErr error
	FetchErr  error
}

func (f *FakeManifestProvider) RenderKustomize(_ context.Context, tarballURL, subPath string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Calls = append(f.Calls, ManifestProviderCall{Op: "kustomize", TarballURL: tarballURL, SubPath: subPath})
	if f.RenderErr != nil {
		return nil, f.RenderErr
	}
	return []byte("# rendered: " + tarballURL + " " + subPath), nil
}

func (f *FakeManifestProvider) FetchManifest(_ context.Context, url string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Calls = append(f.Calls, ManifestProviderCall{Op: "fetch", URL: url})
	if f.FetchErr != nil {
		return nil, f.FetchErr
	}
	return []byte("# fetched: " + url), nil
}

// CallsByOp returns the recorded calls of the given op ("kustomize" | "fetch").
func (f *FakeManifestProvider) CallsByOp(op string) []ManifestProviderCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []ManifestProviderCall
	for _, c := range f.Calls {
		if c.Op == op {
			out = append(out, c)
		}
	}
	return out
}
