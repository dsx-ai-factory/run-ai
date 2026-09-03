/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clusterstatecalculator_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/config"
	models "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"

	clusterstatecalculator "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services/cluster-state-calculator"
)

func TestClusterStateCalculator(t *testing.T) {
	RegisterFailHandler(Fail)
	defer runai_config.FlushAll()

	// Register config types needed for tests
	_ = runai_config.Register[*config.App]()

	RunSpecs(t, "ClusterStateCalculator Suite")
}

var _ = Describe("ClusterStateCalculator", func() {
	const (
		clusterConnectivityThreshold = 1 * time.Hour
	)

	BeforeEach(func() {
		appConfig := runai_config.Get[*config.App]()
		appConfig.ClusterConnectivityThreshold = clusterConnectivityThreshold
	})

	Describe("CalculateState", func() {
		cases := map[string]struct {
			cluster *models.Cluster
			state   clusterstatecalculator.ClusterState
		}{
			"Connected": {
				cluster: &models.Cluster{
					LastLiveness: ptr(time.Now()),
					Status:       models.ClusterStatus{},
				},
				state: clusterstatecalculator.ClusterStateConnected,
			},
			"Disconnected": {
				cluster: &models.Cluster{
					LastLiveness: ptr(time.Now().Add(-2 * clusterConnectivityThreshold)),
					Status:       models.ClusterStatus{},
				},
				state: clusterstatecalculator.ClusterStateDisconnected,
			},
			"MissingPrerequisites": {
				cluster: &models.Cluster{
					LastLiveness: ptr(time.Now()),
					Status: models.ClusterStatus{
						Dependencies: api.ClusterDependenciesStatus{
							Required: map[string]api.ClusterDependencyStatus{
								"dependency1": {
									Available: false,
								},
							},
						},
					},
				},
				state: clusterstatecalculator.ClusterStateMissingPrerequisites,
			},
			"ServiceIssues": {
				cluster: &models.Cluster{
					LastLiveness: ptr(time.Now()),
					Status: models.ClusterStatus{
						Operands: map[string]api.ClusterReportedStatusOperandsValue{
							"operand1": {
								UnreadyThresholdCrossed: true,
							},
						},
					},
				},
				state: clusterstatecalculator.ClusterStateServiceIssues,
			},
			"WaitingToConnect": {
				cluster: &models.Cluster{
					LastLiveness: nil,
				},
				state: clusterstatecalculator.ClusterStateWaitingToConnect,
			},
		}

		for name, c := range cases {
			c := c
			It(name, func() {
				Expect(clusterstatecalculator.NewClusterStateCalculator().CalculateState(c.cluster)).To(Equal(c.state))
			})
		}
	})
})

func ptr[T any](v T) *T {
	return &v
}
