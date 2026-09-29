/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package report

import (
	"bytes"
	"context"
	"errors"
	"fmt"
)

type Receipt struct {
	File string
	Runs int
}

func Publish(ctx context.Context, store Store, document Document) (Receipt, error) {
	data, err := document.Encode()
	if err != nil {
		return Receipt{}, fmt.Errorf("encode results: %w", err)
	}
	name := document.FileName()

	switch existing, err := store.Get(ctx, name); {
	case err == nil && !bytes.Equal(existing, data):
		return Receipt{}, fmt.Errorf("results %s already exist in %s with different content; refusing to overwrite an earlier run", name, store)
	case err != nil && !errors.Is(err, ErrNotFound):
		return Receipt{}, fmt.Errorf("check results %s: %w", name, err)
	}
	if err := store.Put(ctx, name, data); err != nil {
		return Receipt{}, fmt.Errorf("write results %s: %w", name, err)
	}

	manifest := Manifest{}
	current, err := store.Get(ctx, ManifestName)
	switch {
	case errors.Is(err, ErrNotFound):
	case err != nil:
		return Receipt{}, fmt.Errorf("read manifest: %w", err)
	default:
		if manifest, err = parseManifest(current); err != nil {
			return Receipt{}, fmt.Errorf("parse manifest: %w", err)
		}
	}
	updated := manifest.with(name)
	encoded, err := updated.encode()
	if err != nil {
		return Receipt{}, fmt.Errorf("encode manifest: %w", err)
	}
	if err := store.Put(ctx, ManifestName, encoded); err != nil {
		return Receipt{}, fmt.Errorf("write manifest: %w", err)
	}

	stored, err := store.Get(ctx, name)
	if err != nil {
		return Receipt{}, fmt.Errorf("verify results %s: %w", name, err)
	}
	if !bytes.Equal(stored, data) {
		return Receipt{}, fmt.Errorf("verify results %s: stored content differs from what was written", name)
	}
	storedManifest, err := store.Get(ctx, ManifestName)
	if err != nil {
		return Receipt{}, fmt.Errorf("verify manifest: %w", err)
	}
	if !bytes.Equal(storedManifest, encoded) {
		return Receipt{}, fmt.Errorf("verify manifest: stored %s differs from what was written", ManifestName)
	}
	return Receipt{File: name, Runs: len(updated.Runs)}, nil
}
