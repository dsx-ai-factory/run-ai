/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	testing_pkg "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
)

var _ = Describe("determineSubmissionMethods", func() {
	var ctx context.Context
	var originalConfig config.ServicesConfig
	var testTenantId int32 = 1

	setupClusterMock := func(clusterId, version string) {
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(clusterId, testing_pkg.BuildDisplayedCluster(clusterId, version, nil)))
		checker.InitClusterCapabilitiesChecker("", testing_pkg.GetDefaultClustersCacheConfig(), "")
	}

	BeforeEach(func() {
		ctx = context.Background()
		cfg := runai_config.Get[*config.ServicesConfig]()
		originalConfig = *cfg
	})

	AfterEach(func() {
		cfg := runai_config.Get[*config.ServicesConfig]()
		*cfg = originalConfig
	})

	Context("when cluster capabilities checker is nil", func() {
		It("returns unknown (fail-closed)", func() {
			cfg := runai_config.Get[*config.ServicesConfig]()
			cfg.AnyWorkloadSubmissionEnabled = true

			result := determineSubmissionMethods(ctx, testTenantId, Workspace, nil)

			// Fail-closed: nil checker means we cannot verify cluster support, so neither
			// submission method is reported as available.
			Expect(result).To(Equal(SubmissionMethods{}))
		})
	})

	Context("when cluster does not support anyworkload for V1", func() {
		It("returns native for cluster version below 2.26.0", func() {
			clusterId := "cluster-old"
			setupClusterMock(clusterId, "2.22.0")

			cfg := runai_config.Get[*config.ServicesConfig]()
			cfg.AnyWorkloadSubmissionEnabled = true

			capChecker := checker.NewClusterCapabilitiesChecker(clusterId)
			result := determineSubmissionMethods(ctx, testTenantId, Workspace, capChecker)

			Expect(result).To(Equal(SubmissionMethodNative))
		})
	})

	Context("deployment-level config flags", func() {
		It("returns native when both global and per-type flags are disabled", func() {
			clusterId := "cluster-1"
			setupClusterMock(clusterId, "2.30.0")

			cfg := runai_config.Get[*config.ServicesConfig]()
			cfg.AnyWorkloadSubmissionEnabled = false
			cfg.AnyWorkloadWorkspaceSubmission = false

			capChecker := checker.NewClusterCapabilitiesChecker(clusterId)
			result := determineSubmissionMethods(ctx, testTenantId, Workspace, capChecker)

			Expect(result).To(Equal(SubmissionMethodNative))
		})

		It("returns native for unknown workload type even with global flag enabled", func() {
			clusterId := "cluster-unknown"
			setupClusterMock(clusterId, "2.30.0")

			cfg := runai_config.Get[*config.ServicesConfig]()
			cfg.AnyWorkloadSubmissionEnabled = true

			capChecker := checker.NewClusterCapabilitiesChecker(clusterId)
			result := determineSubmissionMethods(ctx, testTenantId, "unknown-type", capChecker)

			Expect(result).To(Equal(SubmissionMethodNative))
		})
	})

	// Note: Tests that verify per-tenant settings require mocking the setting_pkg.Get function,
	// which is beyond the scope of these unit tests. The tests above verify the deployment-level
	// config logic and cluster capability checks.
})
