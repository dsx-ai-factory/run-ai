/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	multierror "github.com/hashicorp/go-multierror"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	CPUDefault                             = openapi.PLACEMENTSTRATEGY_BINPACK
	GPUDefault                             = openapi.PLACEMENTSTRATEGY_BINPACK
	GPUNetworkAccelerationDetectionDefault = openapi.NODEPOOLGPUNETWORKACCELERATIONDETECTION_AUTO

	DefaultTimeBasedFairShareHistoricalUsageWeight = float32(1.0)
	DefaultTimeBasedFairShareDecayHalfLife         = "0h"
	DefaultTimeBasedFairShareWindowDuration        = "7d"
	DefaultTimeBasedFairShareWindowType            = openapi.TIMEBASEDFAIRSHAREWINDOWTYPE_SLIDING
)

type Nullable[T any] interface {
	Get() *T

	IsSet() bool
}

func GetFromNullable[T any](obj Nullable[T], defaultValue T) T {
	value := defaultValue
	if obj.IsSet() {
		value = ptr.Deref(obj.Get(), defaultValue)
	}
	return value
}

func convertNullableNodepoolGPUNetworkAccelerationDetection(dbValue openapi.NullableNodepoolGPUNetworkAccelerationDetection) v1alpha1.GPUNetworkAccelerationDetection {
	extractedValue := GetFromNullable(dbValue, GPUNetworkAccelerationDetectionDefault)
	switch extractedValue {
	case openapi.NODEPOOLGPUNETWORKACCELERATIONDETECTION_USE:
		return v1alpha1.UseGPUNetworkAcceleration
	case openapi.NODEPOOLGPUNETWORKACCELERATIONDETECTION_AUTO:
		return v1alpha1.AutoGPUNetworkAccelerationDetection
	case openapi.NODEPOOLGPUNETWORKACCELERATIONDETECTION_DONT_USE:
		return v1alpha1.DontUseGPUNetworkAcceleration
	}
	return v1alpha1.AutoGPUNetworkAccelerationDetection
}

func isNodePoolDeleting(nodepool *openapi.Nodepool) bool {
	return nodepool.Phase != nil && *nodepool.Phase == openapi.NODEPOOLPHASE_DELETING
}

func extractNodePoolsWithConverter(
	nodepools []openapi.Nodepool, convert func(*openapi.Nodepool) (client.Object, error),
) ([]client.Object, []types.NamespacedName, error) {
	clusterNodePools := []client.Object{}
	nodePoolNamesToDelete := []types.NamespacedName{}
	var err error

	for _, nodepool := range nodepools {
		if isNodePoolDeleting(&nodepool) {
			nodePoolNamesToDelete = append(nodePoolNamesToDelete, types.NamespacedName{Name: nodepool.Name, Namespace: ""})
			continue
		}

		clusterNodePool, innerErr := convert(&nodepool)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
			continue
		}

		clusterNodePools = append(clusterNodePools, clusterNodePool)
	}

	return clusterNodePools, nodePoolNamesToDelete, err
}
