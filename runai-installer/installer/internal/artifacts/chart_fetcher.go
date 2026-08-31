/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package artifacts

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/run-ai/runai/runai-installer/installer/pkg/artifactlock"
	"helm.sh/helm/v3/pkg/registry"
)

// ChartFetcher downloads a chart archive to a path on disk, and checks a
// chart's source is reachable without downloading it.
type ChartFetcher interface {
	Fetch(ctx context.Context, chart artifactlock.Chart, dest string) error
	Probe(ctx context.Context, chart artifactlock.Chart) error
}

type chartFetcher struct {
	keychain *Keychain
}

func NewChartFetcher(keychain *Keychain) ChartFetcher {
	return &chartFetcher{keychain: keychain}
}

// Fetch downloads a chart archive, dispatching on the scheme of its source.
func (f *chartFetcher) Fetch(ctx context.Context, chart artifactlock.Chart, dest string) error {
	switch {
	case strings.HasPrefix(chart.Source, "oci://"):
		return f.fetchOCI(ctx, chart, dest)
	case strings.HasPrefix(chart.Source, "http://"), strings.HasPrefix(chart.Source, "https://"):
		return f.fetchHTTP(ctx, chart, dest)
	}
	return fmt.Errorf("chart %s: unsupported source %q", chart.Name, chart.Source)
}

// Probe checks a chart's source is reachable with the configured credentials,
// without downloading the archive. It verifies access, not that the exact
// archive the lock pins is served.
func (f *chartFetcher) Probe(ctx context.Context, chart artifactlock.Chart) error {
	switch {
	case strings.HasPrefix(chart.Source, "oci://"):
		// An OCI chart lives on a registry, which the registry probe covers.
		return nil
	case strings.HasPrefix(chart.Source, "http://"), strings.HasPrefix(chart.Source, "https://"):
		return f.probeHTTP(ctx, chart)
	}
	return fmt.Errorf("unsupported source %q", chart.Source)
}

func (f *chartFetcher) probeHTTP(ctx context.Context, chart artifactlock.Chart) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, chart.Source, nil)
	if err != nil {
		return err
	}
	// A ranged GET rather than a HEAD: chart repositories commonly answer HEAD
	// with 404 while serving the archive perfectly well (NGC does), and the
	// first byte is enough to tell authorized from rejected.
	req.Header.Set("Range", "bytes=0-0")
	if username, password := f.keychain.BasicAuthForURL(chart.Source); username != "" {
		req.SetBasicAuth(username, password)
	}
	transport, err := f.keychain.TransportFor(hostOf(chart.Source))
	if err != nil {
		return err
	}

	resp, err := (&http.Client{Transport: transport}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return errors.New(resp.Status)
	}
	return nil
}

func (f *chartFetcher) fetchOCI(ctx context.Context, chart artifactlock.Chart, dest string) error {
	ref := strings.TrimPrefix(chart.Source, "oci://")

	transport, err := f.keychain.TransportFor(ref)
	if err != nil {
		return err
	}
	client, err := registry.NewClient(registry.ClientOptHTTPClient(&http.Client{Transport: transport}))
	if err != nil {
		return fmt.Errorf("chart %s: %w", chart.Name, err)
	}
	if username, password := f.keychain.BasicAuthFor(ref); username != "" {
		host := ref[:strings.Index(ref+"/", "/")]
		if err := client.Login(host, registry.LoginOptBasicAuth(username, password)); err != nil {
			return fmt.Errorf("chart %s: log in to %s: %w", chart.Name, host, err)
		}
	}

	result, err := pullChart(ctx, client, ref)
	if err != nil {
		return fmt.Errorf("chart %s: pull %s: %w", chart.Name, chart.Source, err)
	}
	return os.WriteFile(dest, result.Chart.Data, 0o644)
}

// pullChart runs Helm's context-free pull under a context.
func pullChart(ctx context.Context, client *registry.Client, ref string) (*registry.PullResult, error) {
	type outcome struct {
		result *registry.PullResult
		err    error
	}
	done := make(chan outcome, 1)
	go func() {
		result, err := client.Pull(ref, registry.PullOptWithChart(true))
		done <- outcome{result: result, err: err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case o := <-done:
		return o.result, o.err
	}
}

func (f *chartFetcher) fetchHTTP(ctx context.Context, chart artifactlock.Chart, dest string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, chart.Source, nil)
	if err != nil {
		return fmt.Errorf("chart %s: %w", chart.Name, err)
	}
	if username, password := f.keychain.BasicAuthForURL(chart.Source); username != "" {
		req.SetBasicAuth(username, password)
	}
	transport, err := f.keychain.TransportFor(hostOf(chart.Source))
	if err != nil {
		return err
	}

	resp, err := (&http.Client{Transport: transport}).Do(req)
	if err != nil {
		return fmt.Errorf("chart %s: download %s: %w", chart.Name, chart.Source, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("chart %s: download %s: %s", chart.Name, chart.Source, resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("chart %s: %w", chart.Name, err)
	}
	defer out.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("chart %s: download %s: %w", chart.Name, chart.Source, err)
	}
	return nil
}

// hostOf returns the host and path of a URL, for matching against credentials.
func hostOf(source string) string {
	trimmed := source
	for _, scheme := range []string{"https://", "http://", "oci://"} {
		trimmed = strings.TrimPrefix(trimmed, scheme)
	}
	return trimmed
}
