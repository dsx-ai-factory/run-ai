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

// Publish writes the results file, adds it to manifest.json and then reads
// both back from the store; any difference or a missing manifest entry is an
// error, so a run whose numbers did not land fails like any other failed run.
// It assumes a single writer per destination: two publishers updating the
// manifest at the same moment could lose one entry.
func Publish(ctx context.Context, store Store, document Document) (Receipt, error) {
	data, err := document.Encode()
	if err != nil {
		return Receipt{}, fmt.Errorf("encode results: %w", err)
	}
	name := document.FileName()

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
