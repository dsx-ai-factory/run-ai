/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package business

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	CLUSTER_HOST                = "https://{{.}}.test.example.com"
	CLUSTER_UUID                = "550e8400-e29b-41d4-a716-446655440"
	TENANT_ID                   = "123"
	RUNAI_CLUSTER_DOMAIN_SUFFIX = ".example.com"
)

func TestGenerateClusterUrl_HappyFlow(t *testing.T) {
	clusterUrl, err := GenerateClusterUrl(context.Background(), CLUSTER_HOST, CLUSTER_UUID, TENANT_ID)
	assert.Equal(t, "https://runai-cluster-550e-123.test.example.com", clusterUrl)
	assert.Nil(t, err)
}

func TestGetUniqueString_HappyFlow(t *testing.T) {
	uniqueString := getUniqueString("prefix", "clusterId", "tenantId")
	assert.Equal(t, "prefix-clus-tenantId", uniqueString)
}

func TestGetClusterUrlDomainWithEnv_HappyFlow(t *testing.T) {
	res := GetClusterUrlDomainWithEnv("https://runai-cluster-550e-123.test.example.com", RUNAI_CLUSTER_DOMAIN_SUFFIX)
	assert.Equal(t, "runai-cluster-550e-123.test", res)
}
