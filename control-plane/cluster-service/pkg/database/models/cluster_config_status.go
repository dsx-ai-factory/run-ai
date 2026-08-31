/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
)

type ClusterConfigState string

const (
	ClusterStateUpdating ClusterConfigState = "updating"
	ClusterStateApplying ClusterConfigState = "applying"
	ClusterStateApplied  ClusterConfigState = "applied"
	ClusterStateFailed   ClusterConfigState = "failed"
)

type ClusterConfigStateMessage string

const (
	ClusterStateMessageUpdating ClusterConfigStateMessage = "The control plane received a request to update the cluster configuration and did not receive a status update from the cluster yet"
	ClusterStateMessageApplying ClusterConfigStateMessage = "The configuration has been stored on the cluster, but is not yet applied"
	ClusterStateMessageApplied  ClusterConfigStateMessage = "The configuration has been successfully applied in the cluster"
	ClusterStateMessageFailed   ClusterConfigStateMessage = "The cluster has failed to apply the configuration : %s"
)

const (
	ClusterConfigApplied                      string = "Applied"
	ClusterConfigApplying                     string = "Applying"
	ClusterConfigSuccessStatusIndication      string = "True"
	ClusterStatusDidNotReturnAfterMaxInterval string = "Timeout has passed and the cluster did not return a status"
)

type ClusterConfigStatus struct {
	ActualConfig ClusterConfigData
	Conditions   []api.Condition
}

func (ccs *ClusterConfigStatus) GetCondition(conditionType string) *api.Condition {
	for _, condition := range ccs.Conditions {
		if condition.Type == conditionType {
			return &condition
		}
	}
	return nil
}
