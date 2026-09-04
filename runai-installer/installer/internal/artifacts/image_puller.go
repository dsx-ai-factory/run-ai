/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package artifacts

import (
	"context"
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/artifactlock"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// ImagePuller downloads container images and resolves their digests.
type ImagePuller interface {
	// Pull writes the image at ref into dir as an OCI layout.
	Pull(ctx context.Context, ref, dir string, platform artifactlock.Platform) error
	// ResolveDigests returns the manifest-list digest and the platform-specific
	// digest of ref.
	ResolveDigests(ctx context.Context, ref string, platform artifactlock.Platform) (index, digest string, err error)
}

type imagePuller struct {
	keychain *Keychain
}

func NewImagePuller(keychain *Keychain) ImagePuller {
	return &imagePuller{keychain: keychain}
}

// Pull writes an image into an OCI layout. References taken from a lock are
// digest-pinned, so every blob is verified against the digest as it is written.
func (p *imagePuller) Pull(ctx context.Context, ref, dir string, platform artifactlock.Platform) error {
	parsed, err := p.keychain.ParseReference(ref)
	if err != nil {
		return err
	}
	opts, err := p.options(ctx, parsed, platform)
	if err != nil {
		return err
	}

	image, err := remote.Image(parsed, opts...)
	if err != nil {
		return fmt.Errorf("fetch %s: %w", ref, err)
	}
	path, err := layout.Write(dir, empty.Index)
	if err != nil {
		return fmt.Errorf("create layout %s: %w", dir, err)
	}
	if err := path.AppendImage(image); err != nil {
		return fmt.Errorf("write %s: %w", ref, err)
	}
	return nil
}

func (p *imagePuller) ResolveDigests(ctx context.Context, ref string, platform artifactlock.Platform) (string, string, error) {
	parsed, err := p.keychain.ParseReference(ref)
	if err != nil {
		return "", "", err
	}
	opts, err := p.options(ctx, parsed, platform)
	if err != nil {
		return "", "", err
	}

	descriptor, err := remote.Get(parsed, opts...)
	if err != nil {
		return "", "", fmt.Errorf("resolve %s: %w", ref, err)
	}

	image, err := descriptor.Image()
	if err != nil {
		return "", "", fmt.Errorf("resolve %s for %s/%s: %w", ref, platform.OS, platform.Architecture, err)
	}
	digest, err := image.Digest()
	if err != nil {
		return "", "", fmt.Errorf("digest %s: %w", ref, err)
	}
	return descriptor.Digest.String(), digest.String(), nil
}

func (p *imagePuller) options(ctx context.Context, ref name.Reference, platform artifactlock.Platform) ([]remote.Option, error) {
	repo := ref.Context()
	transport, err := p.keychain.TransportFor(repo.RegistryStr() + "/" + repo.RepositoryStr())
	if err != nil {
		return nil, err
	}
	return []remote.Option{
		remote.WithContext(ctx),
		remote.WithAuth(p.keychain.AuthFor(ref)),
		remote.WithTransport(transport),
		remote.WithPlatform(v1.Platform{OS: platform.OS, Architecture: platform.Architecture}),
	}, nil
}
