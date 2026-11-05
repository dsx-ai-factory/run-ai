/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_utils

import (
	"context"
	"net/http"

	clusterApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	cache_testing "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/testing"

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

	Context("zero initial replicas", func() {
		setReportedCapability := func(available *bool) {
			cluster := testing.BuildDisplayedClusterWithKnative(randoms.ClusterId, testing.TestDefaultClusterVersion, true, true, true)
			components := (*cluster.Status.Get().GetDependencies().Optional[checker.KnativeDependency].Components)[checker.KnativeDependency].Components
			if available != nil {
				(*components)[checker.ZeroInitialScaleDependency] = clusterApi.ClusterDependencyStatus{Available: *available}
			}
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().WithCluster(randoms.ClusterId, cluster))
		}

		DescribeTable("checks explicit zero against the reported capability",
			func(initial *int32, available *bool, rejected bool) {
				setReportedCapability(available)
				autoscaling := workloadsApi.NewAutoScaling()
				if initial != nil {
					autoscaling.SetInitialReplicas(*initial)
				}
				err := CheckZeroInitialReplicas(ctx, randoms.ClusterId, autoscaling)
				if rejected {
					Expect(err).NotTo(BeNil())
					Expect(int(err.Code)).To(Equal(http.StatusBadRequest))
					Expect(err.Message).To(ContainSubstring("allow-zero-initial-scale=true"))
				} else {
					Expect(err).To(BeNil())
				}
			},
			Entry("disabled rejects explicit zero", convutils.Ref(int32(0)), convutils.Ref(false), true),
			Entry("enabled accepts explicit zero", convutils.Ref(int32(0)), convutils.Ref(true), false),
			Entry("older cluster keeps admission behavior", convutils.Ref(int32(0)), nil, false),
			Entry("unset does not require the feature", nil, convutils.Ref(false), false),
			Entry("positive does not require the feature", convutils.Ref(int32(1)), convutils.Ref(false), false),
		)

		It("does not check the capability when autoscaling is absent", func() {
			Expect(CheckZeroInitialReplicas(ctx, uuid.NewString(), nil)).To(BeNil())
		})

		It("refreshes a disabled cached capability before rejecting", func() {
			cache_testing.EnableCache()
			DeferCleanup(cache_testing.DisableCache)
			setReportedCapability(convutils.Ref(false))
			configuration, err := checker.NewClusterCapabilitiesChecker(randoms.ClusterId).ClusterConfiguration(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(configuration.KnativeZeroInitialScaleAvailable).To(Equal(convutils.Ref(false)))
			setReportedCapability(convutils.Ref(true))
			autoscaling := workloadsApi.NewAutoScaling()
			autoscaling.SetInitialReplicas(0)
			Expect(CheckZeroInitialReplicas(ctx, randoms.ClusterId, autoscaling)).To(BeNil())
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
