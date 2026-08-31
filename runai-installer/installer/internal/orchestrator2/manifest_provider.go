/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orchestrator2

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// ManifestProvider produces raw YAML manifests from external sources. It is the
// single seam that touches the network and the kustomize renderer, so it is
// injected (and faked in tests) like HelmInstaller. Rendering is done purely in
// Go: the repo is fetched as an HTTPS tarball (no git binary) and built from an
// in-memory filesystem (no kubectl binary).
type ManifestProvider interface {
	// RenderKustomize downloads the gzipped tarball at tarballURL into an
	// in-memory filesystem and renders the kustomize overlay rooted at subPath.
	RenderKustomize(ctx context.Context, tarballURL, subPath string) ([]byte, error)
	// FetchManifest GETs a single flat YAML manifest.
	FetchManifest(ctx context.Context, url string) ([]byte, error)
}

type realManifestProvider struct {
	client   *http.Client
	attempts int
	backoff  time.Duration
}

func NewManifestProvider() ManifestProvider {
	return &realManifestProvider{
		client:   &http.Client{Timeout: 2 * time.Minute},
		attempts: 3,
		backoff:  2 * time.Second,
	}
}

func (p *realManifestProvider) FetchManifest(ctx context.Context, url string) ([]byte, error) {
	body, err := p.get(ctx, url)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	return io.ReadAll(body)
}

func (p *realManifestProvider) RenderKustomize(ctx context.Context, tarballURL, subPath string) ([]byte, error) {
	fSys, err := p.downloadIntoFs(ctx, tarballURL)
	if err != nil {
		return nil, err
	}
	resMap, err := krusty.MakeKustomizer(krusty.MakeDefaultOptions()).Run(fSys, subPath)
	if err != nil {
		return nil, fmt.Errorf("kustomize build %q: %w", subPath, err)
	}
	return resMap.AsYaml()
}

// get performs an HTTP GET with retries on transient failures (connection /
// TLS-handshake errors and 5xx responses), which are common against GitHub at
// install time. 4xx responses are not retried (they will not recover). The
// returned body is the caller's to close.
func (p *realManifestProvider) get(ctx context.Context, url string) (io.ReadCloser, error) {
	attempts := p.attempts
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt) * p.backoff):
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		resp, err := p.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode == http.StatusOK {
			return resp.Body, nil
		}
		resp.Body.Close()
		lastErr = fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
		if resp.StatusCode < http.StatusInternalServerError {
			return nil, lastErr
		}
	}
	return nil, fmt.Errorf("GET %s failed after %d attempts: %w", url, attempts, lastErr)
}

// downloadIntoFs streams the gzipped tarball into a kustomize in-memory
// filesystem, stripping the single top-level directory GitHub archives nest
// everything under (e.g. "training-operator-1.8.1/").
func (p *realManifestProvider) downloadIntoFs(ctx context.Context, tarballURL string) (filesys.FileSystem, error) {
	body, err := p.get(ctx, tarballURL)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	gz, err := gzip.NewReader(body)
	if err != nil {
		return nil, fmt.Errorf("open gzip from %s: %w", tarballURL, err)
	}
	defer gz.Close()

	fSys := filesys.MakeFsInMemory()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read tarball from %s: %w", tarballURL, err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		rel := strings.SplitN(hdr.Name, "/", 2)
		if len(rel) != 2 || rel[1] == "" {
			continue
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			return nil, fmt.Errorf("read %s from tarball: %w", hdr.Name, err)
		}
		if err := fSys.MkdirAll(path.Dir(rel[1])); err != nil {
			return nil, err
		}
		if err := fSys.WriteFile(rel[1], data); err != nil {
			return nil, err
		}
	}
	return fSys, nil
}
