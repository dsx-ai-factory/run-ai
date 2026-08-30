package imagelock

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/run-ai/runai/build/airgap/internal/metadata"
	"github.com/run-ai/runai/build/airgap/internal/thirdparty"
	"github.com/run-ai/runai/build/airgap/pkg/lock"
)

type Resolver interface {
	Resolve(context.Context, string, lock.Platform) (string, string, error)
}

type RemoteResolver struct {
	Options []remote.Option
}

type Config struct {
	Name       string
	Version    string
	Registry   string
	Platform   lock.Platform
	Roots      []string
	MaxWorkers int
	ThirdParty []thirdparty.Image
}

func Build(ctx context.Context, config Config, resolver Resolver) (lock.ImageLock, error) {
	if config.Name == "" {
		return lock.ImageLock{}, errors.New("image lock name is required")
	}
	if config.Version == "" || config.Version == "latest" {
		return lock.ImageLock{}, fmt.Errorf("invalid image lock version %q", config.Version)
	}
	if config.Registry == "" {
		return lock.ImageLock{}, errors.New("registry is required")
	}
	if config.Platform.OS == "" || config.Platform.Architecture == "" {
		return lock.ImageLock{}, errors.New("platform OS and architecture are required")
	}
	if config.MaxWorkers < 1 {
		return lock.ImageLock{}, errors.New("max workers must be at least one")
	}
	artifacts, err := metadata.Discover(config.Roots, "container")
	if err != nil {
		return lock.ImageLock{}, err
	}
	if len(artifacts) == 0 {
		return lock.ImageLock{}, errors.New("no container artifacts discovered")
	}

	type candidate struct {
		name       string
		repository string
		source     string
	}
	candidates := make([]candidate, 0, len(artifacts)+len(config.ThirdParty))
	names := map[string]struct{}{}
	for _, artifact := range artifacts {
		repository := strings.TrimRight(config.Registry, "/") + "/" + artifact.ServiceName
		candidates = append(candidates, candidate{
			name:       artifact.ServiceName,
			repository: repository,
			source:     repository + ":" + config.Version,
		})
		names[artifact.ServiceName] = struct{}{}
	}
	for _, image := range config.ThirdParty {
		if _, exists := names[image.Name]; exists {
			return lock.ImageLock{}, fmt.Errorf("duplicate image name %q", image.Name)
		}
		if image.Name == "" || image.Repository == "" || image.Tag == "" {
			return lock.ImageLock{}, errors.New("third-party image name, repository, and tag are required")
		}
		candidates = append(candidates, candidate{
			name:       image.Name,
			repository: image.Repository,
			source:     image.Source(),
		})
		names[image.Name] = struct{}{}
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].name < candidates[j].name
	})

	type resolutionResult struct {
		image lock.Image
		err   error
	}
	results := make([]resolutionResult, len(candidates))
	semaphore := make(chan struct{}, config.MaxWorkers)
	var wait sync.WaitGroup
	for index, image := range candidates {
		wait.Add(1)
		go func() {
			defer wait.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			indexDigest, platformDigest, resolveErr := resolver.Resolve(ctx, image.source, config.Platform)
			if resolveErr != nil {
				results[index].err = fmt.Errorf("%s: %w", image.source, resolveErr)
				return
			}
			results[index].image = lock.Image{
				Name:        image.name,
				Image:       image.repository + "@" + platformDigest,
				Source:      image.source,
				IndexDigest: indexDigest,
			}
		}()
	}
	wait.Wait()

	var failures []error
	images := make([]lock.Image, 0, len(results))
	for _, result := range results {
		if result.err != nil {
			failures = append(failures, result.err)
			continue
		}
		images = append(images, result.image)
	}
	if len(failures) > 0 {
		return lock.ImageLock{}, fmt.Errorf("resolve images: %w", errors.Join(failures...))
	}
	sort.Slice(images, func(i, j int) bool {
		return images[i].Name < images[j].Name
	})
	imageLock := lock.ImageLock{
		APIVersion: lock.APIVersion,
		Kind:       lock.ImageKind,
		Metadata: lock.Metadata{
			Name:    config.Name,
			Version: config.Version,
		},
		Spec: lock.ImageLockSpec{
			Profile:  "standard",
			Platform: config.Platform,
			Images:   images,
		},
	}
	if err := imageLock.Validate(); err != nil {
		return lock.ImageLock{}, err
	}
	return imageLock, nil
}

func (resolver RemoteResolver) Resolve(ctx context.Context, source string, platform lock.Platform) (string, string, error) {
	reference, err := name.ParseReference(source)
	if err != nil {
		return "", "", err
	}
	options := []remote.Option{
		remote.WithContext(ctx),
		remote.WithAuthFromKeychain(authn.DefaultKeychain),
	}
	options = append(options, resolver.Options...)
	descriptor, err := remote.Get(reference, options...)
	if err != nil {
		return "", "", err
	}
	if !descriptor.MediaType.IsIndex() {
		image, err := descriptor.Image()
		if err != nil {
			return "", "", err
		}
		config, err := image.ConfigFile()
		if err != nil {
			return "", "", err
		}
		if config.OS != platform.OS || config.Architecture != platform.Architecture {
			return "", "", fmt.Errorf("image platform is %s/%s, expected %s", config.OS, config.Architecture, platform)
		}
		digest := descriptor.Digest.String()
		return digest, digest, nil
	}
	index, err := descriptor.ImageIndex()
	if err != nil {
		return "", "", err
	}
	manifest, err := index.IndexManifest()
	if err != nil {
		return "", "", err
	}
	var digest string
	for _, candidate := range manifest.Manifests {
		if candidate.Platform == nil || candidate.Platform.OS != platform.OS || candidate.Platform.Architecture != platform.Architecture {
			continue
		}
		if digest != "" {
			return "", "", fmt.Errorf("multiple manifests match %s", platform)
		}
		digest = candidate.Digest.String()
	}
	if digest == "" {
		return "", "", fmt.Errorf("no manifest matches %s", platform)
	}
	return descriptor.Digest.String(), digest, nil
}
