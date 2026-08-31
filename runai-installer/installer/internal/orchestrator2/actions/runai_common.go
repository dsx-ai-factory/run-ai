/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/runai-installer/installer/internal/step"
)

func ControlPlaneURL(domain string) string {
	return "https://" + domain
}

func FindClusterUUID(ctx context.Context, apiClient step.RunAIAPIClient, name string) (string, error) {
	clusters, statusCode, err := apiClient.GetClusters(ctx)
	if err != nil {
		return "", err
	}
	if statusCode != http.StatusOK {
		return "", fmt.Errorf("list clusters: unexpected status %d", statusCode)
	}
	for _, c := range clusters {
		if c.Name == name {
			return c.Uuid, nil
		}
	}
	return "", nil
}
