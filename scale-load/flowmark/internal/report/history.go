/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package report

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
)

func LoadHistory(ctx context.Context, store Store, clusterType string, window int) ([]Document, error) {
	if window <= 0 {
		return nil, nil
	}
	current, err := store.Get(ctx, ManifestName)
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	manifest, err := parseManifest(current)
	if err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}

	var history []Document
	for i := len(manifest.Runs) - 1; i >= 0 && len(history) < window; i-- {
		name := manifest.Runs[i]
		data, err := store.Get(ctx, name)
		if err != nil {
			return nil, fmt.Errorf("read history %s: %w", name, err)
		}
		var document Document
		if err := json.Unmarshal(data, &document); err != nil {
			return nil, fmt.Errorf("parse history %s: %w", name, err)
		}
		if document.Metadata.ClusterType != clusterType {
			continue
		}
		history = append(history, document)
	}
	slices.SortFunc(history, func(a, b Document) int { return a.Metadata.Date.Compare(b.Metadata.Date) })
	return history, nil
}
