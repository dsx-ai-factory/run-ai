/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	ccontroller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/controllers/run"
	controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/controllers/run"
	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/tests/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RunaiConfig Controller", func() {
	Context("Resources", func() {
		It("should have resources to control", func() {
			hybridClient := kc.New(k8s, k8s, nil)
			managedOperands := ccontroller.NewManagedOperands(apiCtx, hybridClient, hybridClient, runaiConfig)
			resources, err := controller.ResourcesForOperands(apiCtx,
				managedOperands.GetRemoteClusterEnabled())
			ExpectNoErr(err)
			Expect(resources.CreateOrUpdate).NotTo(BeEmpty())
		})
	})
})
