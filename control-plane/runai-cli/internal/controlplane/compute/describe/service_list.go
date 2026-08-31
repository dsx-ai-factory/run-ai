/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe

import (
	"context"
	"fmt"

	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

// Lister defines the interface for listing computes
type Lister interface {
	List(ctx context.Context, projectID *int32, clusterID *string) (*openapi.ComputeListResponse, error)
}

func FindByName(ctx context.Context, lister Lister, name string, projectID *int32, clusterID *string) (*openapi.ComputeAsset, error) {
	response, err := lister.List(ctx, projectID, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to search for compute: %w", err)
	}

	for _, compute := range response.Entries {
		if compute.Meta.Name == name {
			return &compute, nil
		}
	}

	return nil, fmt.Errorf("compute '%s' not found", name)
}
