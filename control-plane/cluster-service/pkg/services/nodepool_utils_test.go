/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	smocks "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/mocks/pkg/services"
	clients_local "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/clients"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
)

var _ = Describe("stripUnsupportedSchedulingFields", func() {
	var (
		ctx     context.Context
		checker *smocks.ClusterCapabilityChecker
		binpack api.PlacementStrategy
		spread  api.PlacementStrategy
	)

	BeforeEach(func() {
		ctx = context.Background()
		checker = smocks.NewClusterCapabilityChecker(GinkgoT())
		binpack = api.PLACEMENTSTRATEGY_BINPACK
		spread = api.PLACEMENTSTRATEGY_SPREAD
	})

	Context("early returns (no cluster lookup needed)", func() {
		It("does nothing when SchedulingConfiguration is nil", func() {
			err := stripUnsupportedSchedulingFields(ctx, nil, checker)
			Expect(err).NotTo(HaveOccurred())
			checker.AssertNotCalled(GinkgoT(), "ValidateClusterCapability", mock.Anything, mock.Anything)
		})

		It("does nothing when GpuPodsPlacementStrategy is nil", func() {
			sc := &api.SchedulingConfiguration{}
			err := stripUnsupportedSchedulingFields(ctx, sc, checker)
			Expect(err).NotTo(HaveOccurred())
			checker.AssertNotCalled(GinkgoT(), "ValidateClusterCapability", mock.Anything, mock.Anything)
		})

		It("skips the cluster lookup when Device is nil", func() {
			sc := &api.SchedulingConfiguration{
				GpuPodsPlacementStrategy: &api.GpuPodsPlacementStrategy{Node: &binpack},
			}
			err := stripUnsupportedSchedulingFields(ctx, sc, checker)
			Expect(err).NotTo(HaveOccurred())
			Expect(sc.GpuPodsPlacementStrategy.Node).To(Equal(&binpack))
			checker.AssertNotCalled(GinkgoT(), "ValidateClusterCapability", mock.Anything, mock.Anything)
		})
	})

	Context("with Device set", func() {
		var sc *api.SchedulingConfiguration

		BeforeEach(func() {
			sc = &api.SchedulingConfiguration{
				GpuPodsPlacementStrategy: &api.GpuPodsPlacementStrategy{
					Node:   &binpack,
					Device: &spread,
				},
			}
		})

		It("preserves Device when the cluster supports device placement", func() {
			checker.On("ValidateClusterCapability", mock.Anything, clients_local.LocalMinVersionOfGpuDevicePlacementStrategy).
				Return(true, nil)
			err := stripUnsupportedSchedulingFields(ctx, sc, checker)
			Expect(err).NotTo(HaveOccurred())
			Expect(sc.GpuPodsPlacementStrategy.Device).To(Equal(&spread))
		})

		It("strips Device when the cluster does not support device placement", func() {
			checker.On("ValidateClusterCapability", mock.Anything, clients_local.LocalMinVersionOfGpuDevicePlacementStrategy).
				Return(false, nil)
			err := stripUnsupportedSchedulingFields(ctx, sc, checker)
			Expect(err).NotTo(HaveOccurred())
			Expect(sc.GpuPodsPlacementStrategy.Device).To(BeNil())
			Expect(sc.GpuPodsPlacementStrategy.Node).To(Equal(&binpack))
		})

		It("strips Device when the cluster is missing (wrapped ErrClusterNotFound)", func() {
			wrapped := fmt.Errorf("failed to get cluster version: %w", dao.ErrClusterNotFound)
			checker.On("ValidateClusterCapability", mock.Anything, clients_local.LocalMinVersionOfGpuDevicePlacementStrategy).
				Return(false, wrapped)
			err := stripUnsupportedSchedulingFields(ctx, sc, checker)
			Expect(err).NotTo(HaveOccurred())
			Expect(sc.GpuPodsPlacementStrategy.Device).To(BeNil())
			Expect(sc.GpuPodsPlacementStrategy.Node).To(Equal(&binpack))
		})

		It("propagates capability lookup errors other than ErrClusterNotFound", func() {
			checker.On("ValidateClusterCapability", mock.Anything, clients_local.LocalMinVersionOfGpuDevicePlacementStrategy).
				Return(false, errors.New("boom"))
			err := stripUnsupportedSchedulingFields(ctx, sc, checker)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to check cluster support for gpuPodsPlacementStrategy.device"))
			Expect(errors.Is(err, dao.ErrClusterNotFound)).To(BeFalse())
		})
	})
})

var _ = Describe("validateSchedulingCapabilities on update and patch", func() {
	conflicting := api.NodepoolUpdateFields{
		SchedulingCapabilities: api.SchedulingCapabilitiesNullable{Quota: ptr.To(false), CpuQuota: ptr.To(true)},
	}

	It("refuses cpuQuota on with quota off in a PUT that sends only capabilities", func() {
		checker := smocks.NewClusterCapabilityChecker(GinkgoT())
		err := validateUpdateParams(context.Background(), conflicting, checker)
		Expect(err).To(MatchError(ContainSubstring("CpuQuota can only be enabled when SchedulingCapabilities.Quota is enabled")))
	})

	It("refuses cpuQuota on with quota off in a PATCH that sends only capabilities", func() {
		checker := smocks.NewClusterCapabilityChecker(GinkgoT())
		err := validatePatchParams(context.Background(), models.Nodepool{}, conflicting, checker)
		Expect(err).To(MatchError(ContainSubstring("CpuQuota can only be enabled when SchedulingCapabilities.Quota is enabled")))
	})
})
