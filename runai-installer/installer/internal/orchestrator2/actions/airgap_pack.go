/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/artifacts"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/artifactlock"
	"golang.org/x/sync/errgroup"
)

// pullParallelism caps concurrent image downloads.
const pullParallelism = 4

// progress reports one live line to the text UI, when a sink is listening.
func progress(deps orchestrator2.Deps, line string) {
	if deps.Progress != nil {
		deps.Progress([]string{line})
	}
}

// ResolveArtifacts applies the user's overrides to the artifact lock and writes
// the effective lock the later steps work from.
type ResolveArtifacts struct {
	Lock           *artifactlock.ArtifactLock
	WorkDir        string
	ImageOverrides []spec.SourceOverride
}

func (a ResolveArtifacts) Kind() string { return "airgap.resolve_artifacts" }

func (a ResolveArtifacts) Args() map[string]any {
	return map[string]any{"workDir": a.WorkDir, "overrides": len(a.ImageOverrides)}
}

func (a ResolveArtifacts) Describe() string {
	return fmt.Sprintf("resolve %d image(s) and %d chart(s) with %d override(s) into %s",
		len(a.Lock.Images()), len(a.Lock.Spec.Charts), len(a.ImageOverrides), a.WorkDir)
}

func (a ResolveArtifacts) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	if deps.ImagePuller == nil {
		return orchestrator2.ActionResult{}, errors.New("no image puller configured")
	}

	merged, err := artifacts.Merge(ctx, a.Lock, toSourceOverrides(a.ImageOverrides), deps.ImagePuller)
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}
	if err := merged.Lock.ValidateImages(); err != nil {
		return orchestrator2.ActionResult{}, err
	}
	if err := artifacts.InitBundleDir(a.WorkDir); err != nil {
		return orchestrator2.ActionResult{}, err
	}
	if err := artifacts.WriteLock(a.WorkDir, merged.Lock); err != nil {
		return orchestrator2.ActionResult{}, err
	}

	return orchestrator2.ActionResult{
		Reason: fmt.Sprintf("%d image(s), %d chart(s); %d replaced, %d added",
			len(merged.Lock.Images()), len(merged.Lock.Spec.Charts), len(merged.Replaced), len(merged.Added)),
		Data: map[string]any{
			"images":   len(merged.Lock.Images()),
			"charts":   len(merged.Lock.Spec.Charts),
			"replaced": merged.Replaced,
			"added":    merged.Added,
		},
	}, nil
}

// PullImages downloads every image in the effective lock into an OCI layout. It
// attempts all of them before reporting, so the user gets one complete list of
// what failed rather than discovering the next problem on every re-run.
type PullImages struct {
	WorkDir string
}

func (a PullImages) Kind() string { return "airgap.pull_images" }

func (a PullImages) Args() map[string]any { return map[string]any{"workDir": a.WorkDir} }

func (a PullImages) Describe() string {
	return fmt.Sprintf("download the locked container images into %s", a.WorkDir)
}

func (a PullImages) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	if deps.ImagePuller == nil {
		return orchestrator2.ActionResult{}, errors.New("no image puller configured")
	}
	lock, err := artifacts.ReadLock(a.WorkDir)
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}
	images := lock.ImageEntries()

	failures := make([]string, len(images))
	var (
		group errgroup.Group
		mu    sync.Mutex
		done  int
	)
	group.SetLimit(pullParallelism)

	for i, entry := range images {
		group.Go(func() error {
			// A worker never returns an error: one failure must not cancel the
			// others, since the point is to attempt every image.
			image := entry.Image
			err := deps.ImagePuller.Pull(ctx, image.Image, artifacts.ImageDir(a.WorkDir, entry.Group, image.Name), lock.Spec.Platform)

			mu.Lock()
			defer mu.Unlock()
			done++
			if err != nil {
				failures[i] = fmt.Sprintf("%s/%s: %v", entry.Group, image.Name, err)
			}
			progress(deps, fmt.Sprintf("%d/%d · %s/%s", done, len(images), entry.Group, image.Name))
			return nil
		})
	}
	_ = group.Wait()

	if err := ctx.Err(); err != nil {
		return orchestrator2.ActionResult{}, err
	}
	if err := collectFailures("image", len(images), failures); err != nil {
		return orchestrator2.ActionResult{}, err
	}
	return orchestrator2.ActionResult{Reason: fmt.Sprintf("%d image(s) downloaded", len(images))}, nil
}

// PullCharts downloads the locked Helm charts and verifies each against the
// digest the lock pins.
type PullCharts struct {
	WorkDir string
}

func (a PullCharts) Kind() string { return "airgap.pull_charts" }

func (a PullCharts) Args() map[string]any { return map[string]any{"workDir": a.WorkDir} }

func (a PullCharts) Describe() string {
	return fmt.Sprintf("download the locked Helm charts into %s", a.WorkDir)
}

func (a PullCharts) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	if deps.ChartFetcher == nil {
		return orchestrator2.ActionResult{}, errors.New("no chart fetcher configured")
	}
	lock, err := artifacts.ReadLock(a.WorkDir)
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}

	failures := make([]string, len(lock.Spec.Charts))
	for i, chart := range lock.Spec.Charts {
		if err := ctx.Err(); err != nil {
			return orchestrator2.ActionResult{}, err
		}
		if err := fetchChart(ctx, deps, chart, artifacts.ChartPath(a.WorkDir, chart.BundlePath)); err != nil {
			failures[i] = err.Error()
		}
		progress(deps, fmt.Sprintf("%d/%d · %s", i+1, len(lock.Spec.Charts), chart.Name))
	}

	if err := collectFailures("chart", len(lock.Spec.Charts), failures); err != nil {
		return orchestrator2.ActionResult{}, err
	}
	return orchestrator2.ActionResult{Reason: fmt.Sprintf("%d chart(s) downloaded", len(lock.Spec.Charts))}, nil
}

type PullCLIs struct {
	WorkDir string
}

func (a PullCLIs) Kind() string { return "airgap.pull_clis" }

func (a PullCLIs) Args() map[string]any { return map[string]any{"workDir": a.WorkDir} }

func (a PullCLIs) Describe() string {
	return fmt.Sprintf("download the locked CLI artifacts into %s", a.WorkDir)
}

func (a PullCLIs) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	if deps.CLIFetcher == nil {
		return orchestrator2.ActionResult{}, errors.New("no CLI fetcher configured")
	}
	lock, err := artifacts.ReadLock(a.WorkDir)
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}
	failures := make([]string, len(lock.Spec.CLIs))
	for i, cli := range lock.Spec.CLIs {
		if err := deps.CLIFetcher.Fetch(ctx, cli, a.WorkDir); err != nil {
			failures[i] = fmt.Sprintf("%s: %v", cli.Name, err)
		}
		progress(deps, fmt.Sprintf("%d/%d · %s", i+1, len(lock.Spec.CLIs), cli.Name))
	}
	if err := collectFailures("CLI", len(lock.Spec.CLIs), failures); err != nil {
		return orchestrator2.ActionResult{}, err
	}
	return orchestrator2.ActionResult{Reason: fmt.Sprintf("%d CLI artifact(s) downloaded", len(lock.Spec.CLIs))}, nil
}

// fetchChart downloads one chart and rejects an archive that is not the one the
// lock pins.
func fetchChart(ctx context.Context, deps orchestrator2.Deps, chart artifactlock.Chart, dest string) error {
	if err := deps.ChartFetcher.Fetch(ctx, chart, dest); err != nil {
		return err
	}
	if chart.Digest == "" {
		return nil
	}
	digest, err := artifacts.FileDigest(dest)
	if err != nil {
		return fmt.Errorf("chart %s: %w", chart.Name, err)
	}
	if digest != chart.Digest {
		return fmt.Errorf("chart %s: the lock pins %s but %s served %s", chart.Name, chart.Digest, chart.Source, digest)
	}
	return nil
}

// AssembleBundle writes the staged artifacts to the bundle tar and its checksum.
type AssembleBundle struct {
	WorkDir    string
	OutputPath string
}

func (a AssembleBundle) Kind() string { return "airgap.assemble_bundle" }

func (a AssembleBundle) Args() map[string]any {
	return map[string]any{"workDir": a.WorkDir, "output": a.OutputPath}
}

func (a AssembleBundle) Describe() string {
	return fmt.Sprintf("write the bundle to %s", a.OutputPath)
}

func (a AssembleBundle) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	digest, err := artifacts.TarDir(ctx, artifacts.BundleDir(a.WorkDir), a.OutputPath)
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}
	if err := artifacts.WriteChecksum(a.OutputPath, digest); err != nil {
		return orchestrator2.ActionResult{}, err
	}
	// The staging directory is large and only useful for diagnosing a failure,
	// so it goes once the bundle exists.
	if err := os.RemoveAll(a.WorkDir); err != nil {
		return orchestrator2.ActionResult{}, err
	}
	return orchestrator2.ActionResult{Reason: a.OutputPath}, nil
}

// collectFailures reports every artifact that could not be packed as one error.
func collectFailures(kind string, total int, failures []string) error {
	var collected []string
	for _, f := range failures {
		if f != "" {
			collected = append(collected, f)
		}
	}
	if len(collected) == 0 {
		return nil
	}
	return fmt.Errorf("%d of %d %s(s) failed:\n  %s", len(collected), total, kind, strings.Join(collected, "\n  "))
}

// ToCredentials flattens the spec's registry credentials into the engine's credential set.
func ToCredentials(in []spec.RegistryCredential) artifacts.Credentials {
	out := make(artifacts.Credentials, 0, len(in))
	for _, c := range in {
		out = append(out, artifacts.Credential{
			Prefix:             c.Prefix,
			Username:           c.Username,
			Password:           c.Password,
			BearerToken:        c.BearerToken,
			CAChainPath:        c.CAChainPath,
			Insecure:           c.Insecure,
			InsecureSkipVerify: c.InsecureSkipVerify,
		})
	}
	return out
}

func toSourceOverrides(in []spec.SourceOverride) []artifacts.SourceOverride {
	out := make([]artifacts.SourceOverride, 0, len(in))
	for _, o := range in {
		out = append(out, artifacts.SourceOverride{Group: o.Group, Name: o.Name, Source: o.Source})
	}
	return out
}
