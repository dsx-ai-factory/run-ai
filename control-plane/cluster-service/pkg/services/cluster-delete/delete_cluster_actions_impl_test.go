/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster_delete

import (
	"context"
	"testing"

	authorizationclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	orgunitclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/mocks"
	"github.com/stretchr/testify/assert"
)

type Mocks struct {
	orgUnitClient       *orgunitclient.IOrgUnitClient
	authorizationClient authorizationclient.IAuthorizationClient
}

func initImplMocks() (*ClusterDeleterImpl, *Mocks) {
	mocks := &Mocks{
		authorizationClient: authorizationclient.NewAuthorizationClient(),
		orgUnitClient:       &orgunitclient.IOrgUnitClient{},
	}
	return &ClusterDeleterImpl{
		orgUnitClient:       mocks.orgUnitClient,
		authorizationClient: mocks.authorizationClient,
	}, mocks
}

func TestAsyncVersionCheckShouldBeTrue(t *testing.T) {
	clusterDeleterImpl, _ := initImplMocks()
	res := clusterDeleterImpl.clusterSupportVersionSupportGracefulShutdown(context.Background(), "2.20.0")
	assert.True(t, res)
}

func TestAsyncVersionCheckDevVersionsShouldBeTrue(t *testing.T) {
	clusterDeleterImpl, _ := initImplMocks()
	res := clusterDeleterImpl.clusterSupportVersionSupportGracefulShutdown(context.Background(), "0.0.0-10006915.testing-c-304fef6")
	assert.True(t, res)

	res = clusterDeleterImpl.clusterSupportVersionSupportGracefulShutdown(context.Background(), "0.0.0-1000005.master-8b88a8e")
	assert.True(t, res)

	res = clusterDeleterImpl.clusterSupportVersionSupportGracefulShutdown(context.Background(), "2.18.8-staging.1")
	assert.False(t, res)
}

func TestAsyncVersionCheckDevVersionsShouldBeFalse(t *testing.T) {
	clusterDeleterImpl, _ := initImplMocks()
	res := clusterDeleterImpl.clusterSupportVersionSupportGracefulShutdown(context.Background(), "2.18.8-staging.1")
	assert.False(t, res)
}

func TestAsyncVersionCheckShouldBeFalse(t *testing.T) {
	clusterDeleterImpl, _ := initImplMocks()
	res := clusterDeleterImpl.clusterSupportVersionSupportGracefulShutdown(context.Background(), "2.16.3")
	assert.False(t, res)
}

func TestAsyncEmptyVersionCheckShouldBeFalse(t *testing.T) {
	clusterDeleterImpl, _ := initImplMocks()
	res := clusterDeleterImpl.clusterSupportVersionSupportGracefulShutdown(context.Background(), "")
	assert.False(t, res)
}
