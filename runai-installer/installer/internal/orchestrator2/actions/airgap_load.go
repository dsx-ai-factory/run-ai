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
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/airgapvalues"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/artifacts"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/airgap"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/artifactlock"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

// pushParallelism caps concurrent image uploads.
const pushParallelism = 4

// ExtractBundle unpacks the air-gapped bundle into the working directory the
// later steps read from.
type ExtractBundle struct {
	BundlePath string
	WorkingDir string
}

func (a ExtractBundle) Kind() string { return "airgap.extract_bundle" }

func (a ExtractBundle) Args() map[string]any {
	return map[string]any{"bundle": a.BundlePath, "workingDir": a.WorkingDir}
}

func (a ExtractBundle) Describe() string {
	return fmt.Sprintf("extract %s into %s", a.BundlePath, a.WorkingDir)
}

func (a ExtractBundle) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	dest := artifacts.BundleDir(a.WorkingDir)
	// A bundle is never merged into an earlier one: a stale image directory
	// would otherwise be pushed as if it belonged to this release.
	if err := os.RemoveAll(dest); err != nil {
		return orchestrator2.ActionResult{}, fmt.Errorf("clear %s: %w", dest, err)
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return orchestrator2.ActionResult{}, fmt.Errorf("create %s: %w", dest, err)
	}
	if err := airgap.ExtractTar(ctx, a.BundlePath, dest); err != nil {
		return orchestrator2.ActionResult{}, err
	}

	// The lock is what every later step reads, so an archive that is not an
	// air-gapped bundle is rejected here rather than midway through the push.
	lock, err := artifacts.ReadLock(a.WorkingDir)
	if err != nil {
		return orchestrator2.ActionResult{}, fmt.Errorf("%s is not an air-gapped bundle: %w", a.BundlePath, err)
	}
	cliPaths, err := verifyLoadedCLIs(a.WorkingDir, lock)
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}
	return orchestrator2.ActionResult{
		Reason: fmt.Sprintf("%s %s: %d image(s), %d chart(s)",
			lock.Metadata.Name, lock.Metadata.Version, len(lock.Images()), len(lock.Spec.Charts)),
		Data: map[string]any{
			"version": lock.Metadata.Version,
			"images":  len(lock.Images()),
			"charts":  len(lock.Spec.Charts),
			"clis":    cliPaths,
		},
	}, nil
}

func verifyLoadedCLIs(workingDir string, lock *artifactlock.ArtifactLock) ([]string, error) {
	var paths []string
	for _, cli := range lock.Spec.CLIs {
		files := cli.Files
		if len(files) == 0 && cli.SourcePath != "" {
			files = []artifactlock.CLIFile{{SourcePath: cli.SourcePath, BundlePath: cli.BundlePath, Digest: cli.Digest, Mode: 0o755}}
		}
		for _, file := range files {
			path, err := artifacts.CLIPath(workingDir, file.BundlePath)
			if err != nil {
				return nil, err
			}
			digest, err := artifacts.FileDigest(path)
			if err != nil {
				return nil, fmt.Errorf("read CLI %s file %s: %w", cli.Name, file.SourcePath, err)
			}
			if digest != file.Digest {
				return nil, fmt.Errorf("CLI %s file %s digest does not match the lock", cli.Name, file.SourcePath)
			}
			info, err := os.Stat(path)
			if err != nil {
				return nil, err
			}
			if info.Mode().Perm()&0o111 == 0 {
				return nil, fmt.Errorf("CLI %s file %s is not executable", cli.Name, file.SourcePath)
			}
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	return paths, nil
}

// PushImages uploads every bundled image to the destination registry. It
// attempts all of them before reporting, so the user gets one complete list of
// what failed rather than discovering the next problem on every re-run.
type PushImages struct {
	WorkingDir  string
	Destination string
}

type loadImage struct {
	entry  artifactlock.ImageEntry
	ref    string
	digest string
}

func (a PushImages) Kind() string { return "airgap.push_images" }

func (a PushImages) Args() map[string]any {
	return map[string]any{"workingDir": a.WorkingDir, "registry": a.Destination}
}

func (a PushImages) Describe() string {
	return fmt.Sprintf("push the bundled container images to %s", a.Destination)
}

func (a PushImages) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	if deps.ImagePusher == nil {
		return orchestrator2.ActionResult{}, errors.New("no image pusher configured")
	}
	lock, err := artifacts.ReadLock(a.WorkingDir)
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}
	images, err := loadImages(lock, a.Destination)
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}
	if err := checkPushPermissions(ctx, deps, images); err != nil {
		return orchestrator2.ActionResult{}, err
	}

	failures := make([]string, len(images))
	var (
		group errgroup.Group
		mu    sync.Mutex
		done  int
	)
	group.SetLimit(pushParallelism)

	for i, image := range images {
		group.Go(func() error {
			// A worker never returns an error: one failure must not cancel the
			// others, since the point is to attempt every image.
			err := a.push(ctx, deps, image)

			mu.Lock()
			defer mu.Unlock()
			done++
			if err != nil {
				failures[i] = fmt.Sprintf("%s/%s: %v", image.entry.Group, image.entry.Image.Name, err)
			}
			progress(deps, fmt.Sprintf("%d/%d · %s/%s", done, len(images), image.entry.Group, image.entry.Image.Name))
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
	return orchestrator2.ActionResult{
		Reason: fmt.Sprintf("%d image(s) pushed to %s", len(images), a.Destination),
	}, nil
}

func loadImages(lock *artifactlock.ArtifactLock, destination string) ([]loadImage, error) {
	entries := lock.ImageEntries()
	images := make([]loadImage, 0, len(entries))
	refs := map[string]string{}
	for _, entry := range entries {
		ref, _, err := artifacts.MirrorRef(destination, entry)
		if err != nil {
			return nil, fmt.Errorf("%s/%s: %w", entry.Group, entry.Image.Name, err)
		}
		digest, err := artifactlock.ImageDigest(entry.Image.Image)
		if err != nil {
			return nil, fmt.Errorf("%s/%s: %w", entry.Group, entry.Image.Name, err)
		}
		if existing, ok := refs[ref]; ok && existing != digest {
			return nil, fmt.Errorf("destination %s maps to both %s and %s", ref, existing, digest)
		}
		if _, ok := refs[ref]; ok {
			continue
		}
		refs[ref] = digest
		images = append(images, loadImage{entry: entry, ref: ref, digest: digest})
	}
	return images, nil
}

func checkPushPermissions(ctx context.Context, deps orchestrator2.Deps, images []loadImage) error {
	failures := make([]string, len(images))
	var group errgroup.Group
	group.SetLimit(pushParallelism)
	for i, image := range images {
		group.Go(func() error {
			if err := ctx.Err(); err != nil {
				failures[i] = err.Error()
				return nil
			}
			if err := deps.ImagePusher.CheckPushPermission(ctx, image.ref); err != nil {
				failures[i] = fmt.Sprintf("%s: %v", image.ref, err)
			}
			return nil
		})
	}
	_ = group.Wait()
	return collectFailures("repository", len(images), failures)
}

// push uploads one image and confirms the registry stored what the lock pins.
func (a PushImages) push(ctx context.Context, deps orchestrator2.Deps, image loadImage) error {
	entry := image.entry
	return deps.ImagePusher.Push(ctx,
		artifacts.ImageDir(a.WorkingDir, entry.Group, entry.Image.Name),
		image.ref,
		image.digest,
	)
}

// GenerateAirgapValues writes the Helm values that install the charts from the
// destination registry.
type GenerateAirgapValues struct {
	WorkingDir  string
	ValuesDir   string
	Destination string
}

func (a GenerateAirgapValues) Kind() string { return "airgap.generate_values" }

func (a GenerateAirgapValues) Args() map[string]any {
	return map[string]any{"valuesDir": a.ValuesDir, "registry": a.Destination}
}

func (a GenerateAirgapValues) Describe() string {
	return fmt.Sprintf("write the chart values for %s into %s", a.Destination, a.ValuesDir)
}

func (a GenerateAirgapValues) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	lock, err := artifacts.ReadLock(a.WorkingDir)
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}
	images, err := mirroredImages(lock)
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}

	if err := os.MkdirAll(a.ValuesDir, 0o755); err != nil {
		return orchestrator2.ActionResult{}, fmt.Errorf("create %s: %w", a.ValuesDir, err)
	}

	files := map[string]map[string]any{
		airgapvalues.ControlPlaneValuesFile: airgapvalues.ControlPlane(a.Destination, images),
		airgapvalues.ClusterValuesFile:      airgapvalues.Cluster(a.Destination, images),
	}
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		encoded, err := yaml.Marshal(files[name])
		if err != nil {
			return orchestrator2.ActionResult{}, fmt.Errorf("encode %s: %w", name, err)
		}
		if err := os.WriteFile(filepath.Join(a.ValuesDir, name), encoded, 0o644); err != nil {
			return orchestrator2.ActionResult{}, err
		}
	}

	return orchestrator2.ActionResult{
		Reason: strings.Join(names, ", ") + " in " + a.ValuesDir,
		Data:   map[string]any{"files": names, "valuesDir": a.ValuesDir},
	}, nil
}

// mirroredImages describes where every bundled image now lives, for the values
// to point at.
func mirroredImages(lock *artifactlock.ArtifactLock) ([]airgapvalues.Image, error) {
	images := lock.ImageEntries()
	out := make([]airgapvalues.Image, 0, len(images))
	for _, entry := range images {
		tag, err := artifacts.SourceTag(entry.Image.Source)
		if err != nil {
			return nil, fmt.Errorf("%s/%s: %w", entry.Group, entry.Image.Name, err)
		}
		out = append(out, airgapvalues.Image{
			Group: entry.Group, GroupVersion: entry.GroupVersion, Name: entry.Image.Name, Tag: tag,
		})
	}
	return out, nil
}
