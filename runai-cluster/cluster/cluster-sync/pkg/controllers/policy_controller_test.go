/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers"
)

type nilInformerCachedClient struct {
	cachedclient.CachedClient
}

func (nilInformerCachedClient) GetInformer(cachedclient.SupportedType) cachedclient.CachedInformer {
	return nil
}

var _ = Describe("legacy policy API lifecycle", func() {
	It("skips policy controllers when their APIs are absent", func() {
		stub := nilInformerCachedClient{}
		policyHandler := handlers.NewPolicyHandler(nil, stub)

		Expect(func() {
			AddPolicyHandlerToCache(stub, policyHandler)
		}).NotTo(Panic())
	})
})
