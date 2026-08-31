/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

// ClustersCountSubdomainStatus  holds the count of subdomain enabled and disabled clusters
type ClustersCountSubdomainStatus struct {
	EnabledCount  int64 // EnabledCount is the count of clusters with subdomain enabled
	DisabledCount int64 // DisabledCount is the count of clusters with subdomain disabled
	MissingCount  int64 // MissingCount is the count of clusters with missing subdomain configuration
}

// TableName implements the Tabler interface
func (ClustersCountSubdomainStatus) TableName() string {
	return ClusterTable
}
