package artifacts

import (
	"context"
	"crypto/sha256"
	"fmt"
	"slices"
	"strings"

	"github.com/run-ai/runai/runai-installer/installer/pkg/artifactlock"
	"gopkg.in/yaml.v3"
)

// customImageLockName groups the images a user added that the release lock does
// not carry, keeping them distinguishable from Run:ai's own.
const customImageLockName = "custom"

// SourceOverride redirects a named image to a different source.
type SourceOverride struct {
	Group  string
	Name   string
	Source string
}

// MergeResult is the effective lock and what the overrides did to it.
type MergeResult struct {
	Lock     *artifactlock.ArtifactLock
	Replaced []string
	Added    []string
}

// Merge applies the image overrides to a copy of the lock, resolving each
// override to its digests. A name matching a locked image replaces it; any
// other name is added, since packing an extra image is legal.
func Merge(ctx context.Context, lock *artifactlock.ArtifactLock, overrides []SourceOverride, puller ImagePuller) (MergeResult, error) {
	out := cloneLock(lock)
	result := MergeResult{Lock: out}

	var additions []artifactlock.Image
	for _, o := range overrides {
		index, digest, err := puller.ResolveDigests(ctx, o.Source, out.Spec.Platform)
		if err != nil {
			return MergeResult{}, fmt.Errorf("image override %q: %w", o.Name, err)
		}
		image := artifactlock.Image{
			Name:        o.Name,
			Image:       repository(o.Source) + "@" + digest,
			Source:      o.Source,
			IndexDigest: index,
		}
		group, replaced, err := replaceImage(out, o.Group, image)
		if err != nil {
			return MergeResult{}, err
		}
		if replaced {
			result.Replaced = append(result.Replaced, group+"/"+o.Name)
			continue
		}
		additions = append(additions, image)
		result.Added = append(result.Added, customImageLockName+"/"+o.Name)
	}

	if len(additions) > 0 {
		custom := -1
		for i := range out.Spec.ImageLocks {
			if out.Spec.ImageLocks[i].Name == customImageLockName {
				custom = i
				break
			}
		}
		if custom < 0 {
			out.Spec.ImageLocks = append(out.Spec.ImageLocks, artifactlock.ImageLock{
				Name: customImageLockName, Version: out.Metadata.Version, Images: additions,
			})
			custom = len(out.Spec.ImageLocks) - 1
		} else {
			out.Spec.ImageLocks[custom].Images = append(out.Spec.ImageLocks[custom].Images, additions...)
		}
		digest, err := contentDigest(out.Spec.ImageLocks[custom].Images)
		if err != nil {
			return MergeResult{}, err
		}
		out.Spec.ImageLocks[custom].Digest = digest
	}
	return result, nil
}

func replaceImage(lock *artifactlock.ArtifactLock, group string, image artifactlock.Image) (string, bool, error) {
	var matches [][2]int
	for i := range lock.Spec.ImageLocks {
		if group != "" && lock.Spec.ImageLocks[i].Name != group {
			continue
		}
		for j := range lock.Spec.ImageLocks[i].Images {
			if lock.Spec.ImageLocks[i].Images[j].Name == image.Name {
				matches = append(matches, [2]int{i, j})
			}
		}
	}
	if len(matches) == 0 {
		if group != "" {
			return "", false, fmt.Errorf("image override %q does not match %s/%s", image.Name, group, image.Name)
		}
		return "", false, nil
	}
	if len(matches) > 1 {
		var groups []string
		for _, match := range matches {
			groups = append(groups, lock.Spec.ImageLocks[match[0]].Name)
		}
		return "", false, fmt.Errorf("image override %q is ambiguous across groups %s", image.Name, strings.Join(groups, ", "))
	}
	match := matches[0]
	matchedGroup := lock.Spec.ImageLocks[match[0]].Name
	lock.Spec.ImageLocks[match[0]].Images[match[1]] = image
	return matchedGroup, true, nil
}

// contentDigest derives a digest for the added images, which have no upstream
// document to hash.
func contentDigest(images []artifactlock.Image) (string, error) {
	encoded, err := yaml.Marshal(images)
	if err != nil {
		return "", fmt.Errorf("digest added images: %w", err)
	}
	return fmt.Sprintf("sha256:%x", sha256.Sum256(encoded)), nil
}

// repository strips the tag or digest from a reference, keeping any registry
// port intact.
func repository(ref string) string {
	lastSlash := strings.LastIndex(ref, "/")
	name := ref[lastSlash+1:]
	if i := strings.IndexAny(name, ":@"); i >= 0 {
		return ref[:lastSlash+1] + name[:i]
	}
	return ref
}

func cloneLock(lock *artifactlock.ArtifactLock) *artifactlock.ArtifactLock {
	out := *lock
	out.Spec.ImageLocks = slices.Clone(lock.Spec.ImageLocks)
	for i := range out.Spec.ImageLocks {
		out.Spec.ImageLocks[i].Images = slices.Clone(lock.Spec.ImageLocks[i].Images)
	}
	out.Spec.Charts = slices.Clone(lock.Spec.Charts)
	out.Spec.CLIs = slices.Clone(lock.Spec.CLIs)
	return &out
}
