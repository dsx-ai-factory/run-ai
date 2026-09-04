/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package platform_identifier

import (
	"context"
	"fmt"

	ocpv1 "github.com/openshift/api/config/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getVersionOCP(ctx context.Context, k8s client.Client) (string, error) {
	var clusterVersion ocpv1.ClusterVersion

	err := k8s.Get(ctx, client.ObjectKey{Name: "version"}, &clusterVersion)
	if err != nil {
		return "", fmt.Errorf("failed to get ClusterVersion: %w", err)
	}

	return clusterVersion.Status.Desired.Version, nil
}
