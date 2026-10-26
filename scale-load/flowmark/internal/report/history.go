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

func LoadHistory(ctx context.Context, store Store, clusterType string, limit int) (history []Document, skipped []string, err error) {
	if limit <= 0 || clusterType == "" {
		return nil, nil, nil
	}
	current, err := store.Get(ctx, ManifestName)
	if errors.Is(err, ErrNotFound) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("read manifest: %w", err)
	}
	manifest, err := parseManifest(current)
	if err != nil {
		return nil, nil, fmt.Errorf("parse manifest: %w", err)
	}

	runs := slices.Clone(manifest.Runs)
	slices.Sort(runs)
	for i := len(runs) - 1; i >= 0 && len(history) < limit; i-- {
		name := runs[i]
		data, err := store.Get(ctx, name)
		if err != nil {
			skipped = append(skipped, fmt.Sprintf("read history %s: %v", name, err))
			continue
		}
		var document Document
		if err := json.Unmarshal(data, &document); err != nil {
			skipped = append(skipped, fmt.Sprintf("parse history %s: %v", name, err))
			continue
		}
		if document.Metadata.ClusterType != clusterType {
			continue
		}
		history = append(history, document)
	}
	slices.SortFunc(history, func(a, b Document) int { return a.Metadata.Date.Compare(b.Metadata.Date) })
	return history, skipped, nil
}
