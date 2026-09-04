/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_utils

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/workloadutils"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/testing_crud"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	LWSNotAvailableMsg = "NIMService can't be created because LeaderWorkerSet (LWS) is not installed on the selected cluster. Contact your administrator."
)

var _ = Describe("inference utils testing", func() {
	var ctx context.Context
	var randoms *testutils.RandomValues
	knativeClusterNoScale := uuid.NewString()
	knativeClusterNoNodeType := uuid.NewString()
	noKnativeCluster := uuid.NewString()
	lwsCluster := uuid.NewString()
	noLWSCluster := uuid.NewString()

	BeforeEach(func() {
		randoms = testutils.NewRandomValues()

		ctx = testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})

		authMock := testing_crud.NewRandomsBasedAuthMockClient().With(randoms)
		authMock.AllowedProjectsToContext(&ctx)

		orgunit.SetMock(authMock.OrgUnitMock())

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedClusterWithKnative(randoms.ClusterId, testing.TestDefaultClusterVersion, true, true, true)).
			WithCluster(knativeClusterNoScale, testing.BuildDisplayedClusterWithKnative(knativeClusterNoScale, testing.TestDefaultClusterVersion, true, false, true)).
			WithCluster(knativeClusterNoNodeType, testing.BuildDisplayedClusterWithKnative(knativeClusterNoNodeType, testing.TestDefaultClusterVersion, true, true, false)).
			WithCluster(noKnativeCluster, testing.BuildDisplayedClusterWithKnative(noKnativeCluster, testing.TestDefaultClusterVersion, false, false, false)).
			WithCluster(lwsCluster, testing.BuildDisplayedClusterWithLWS(lwsCluster, testing.TestDefaultClusterVersion, true)).
			WithCluster(noLWSCluster, testing.BuildDisplayedClusterWithLWS(noLWSCluster, testing.TestDefaultClusterVersion, false)))

		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")
	})

	Context("Check knative available", func() {
		It("All features available", func() {
			features := KnativeFeatures{
				ScaleToZero: true,
				NodeType:    true,
			}
			Expect(CheckKnativeAvailable(ctx, randoms.ClusterId, features)).To(BeNil())
		})

		It("Knative not available", func() {
			features := KnativeFeatures{
				ScaleToZero: false,
				NodeType:    false,
			}
			Expect(CheckKnativeAvailable(ctx, noKnativeCluster, features)).To(Not(BeNil()))
		})

		It("Scale to zero required but not available", func() {
			features := KnativeFeatures{
				ScaleToZero: true,
				NodeType:    false,
			}
			Expect(CheckKnativeAvailable(ctx, knativeClusterNoScale, features)).To(Not(BeNil()))
		})

		It("Scale to zero not required", func() {
			features := KnativeFeatures{
				ScaleToZero: false,
				NodeType:    false,
			}
			Expect(CheckKnativeAvailable(ctx, knativeClusterNoScale, features)).To(BeNil())
		})

		It("Node type required but not available", func() {
			features := KnativeFeatures{
				ScaleToZero: false,
				NodeType:    true,
			}
			Expect(CheckKnativeAvailable(ctx, knativeClusterNoNodeType, features)).To(Not(BeNil()))
		})

		It("Node type not required", func() {
			features := KnativeFeatures{
				ScaleToZero: false,
				NodeType:    false,
			}
			Expect(CheckKnativeAvailable(ctx, knativeClusterNoNodeType, features)).To(BeNil())
		})

		It("Unknown cluster", func() {
			features := KnativeFeatures{
				ScaleToZero: false,
				NodeType:    false,
			}
			Expect(CheckKnativeAvailable(ctx, uuid.NewString(), features)).To(Not(BeNil()))
		})
	})

	Context("Check LWS available", func() {
		It("LWS available", func() {
			Expect(workloadutils.CheckLWSAvailable(ctx, lwsCluster, LWSNotAvailableMsg)).To(BeNil())
		})

		It("LWS not available", func() {
			err := workloadutils.CheckLWSAvailable(ctx, noLWSCluster, LWSNotAvailableMsg)
			Expect(err).To(Not(BeNil()))
			Expect(err.Message).To(ContainSubstring(LWSNotAvailableMsg))
		})

		It("Unknown cluster", func() {
			Expect(workloadutils.CheckLWSAvailable(ctx, uuid.NewString(), LWSNotAvailableMsg)).To(Not(BeNil()))
		})
	})
})
