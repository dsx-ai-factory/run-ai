/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"

	clusterapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	k8sConstants "github.com/run-ai/runai/control-plane/workloads-manager/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	clustersClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const adviceRestartControllerOperandKey = "advice-restart-controller"

// shouldStampRestartWorkflowLabel reports whether a workload should carry the Heimdall workflow-id
// label so Heimdall can monitor it and post restart Advices.
//
// convertedFromNative comes from trustedConvertedFromNative (a v1 native workload cast to v2). We
// stamp only those: they are immutable via the v2 update endpoint, so the label is set once at
// creation and a restart can never race a user update. Raw v2 PyTorchJobs are mutable and updating
// one during a restart is not handled yet, so they are not stamped.
func (w *WorkloadsV2Impl) shouldStampRestartWorkflowLabel(ctx context.Context, manifest map[string]interface{},
	clusterId string, convertedFromNative bool,
) bool {
	return manifest[k8sConstants.FieldKind] == constants.KindPyTorchJob &&
		convertedFromNative &&
		!isAJRDisabled(manifest) &&
		isAdviceRestartOperandEnabled(ctx, clusterId)
}

func (w *WorkloadsV2Impl) stampRestartWorkflowLabel(labels map[string]interface{}, workloadId string) {
	labels[constants.HeimdallWorkflowIdLabel] = workloadId
}

func isAJRDisabled(manifest map[string]interface{}) bool {
	metadata, ok := manifest[k8sConstants.FieldMetadata].(map[string]interface{})
	if !ok {
		return false
	}
	annotations, ok := metadata[k8sConstants.FieldAnnotations].(map[string]interface{})
	if !ok {
		return false
	}
	return annotations[constants.DisableAJRAnnotation] == TrueAsString
}

func isAdviceRestartOperandEnabled(ctx context.Context, clusterId string) bool {
	cluster, err := clustersClient.Get().Client().GetClusterByUuid(ctx, clusterId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf(
			"skipping restart workflow-id stamp, failed to get cluster %s: %v", clusterId, err)
		return false
	}
	if !cluster.Status.IsSet() {
		return false
	}

	status := cluster.Status.Get()
	if status.GetState() != clusterapi.CLUSTERSTATE_CONNECTED {
		return false
	}

	operand, ok := status.GetOperands()[adviceRestartControllerOperandKey]
	if ok && !operand.Ready {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf(
			"advice-restart-controller enabled but not ready on cluster %s, stamping workload optimistically", clusterId)
	}
	return ok
}
