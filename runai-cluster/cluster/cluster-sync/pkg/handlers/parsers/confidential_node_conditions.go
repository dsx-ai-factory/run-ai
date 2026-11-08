/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	corev1 "k8s.io/api/core/v1"

	swagger "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils"
)

const (
	ConfidentialComputingMissingTEE     = "ConfidentialComputingMissingTEE"
	ConfidentialComputingGpuNotReady    = "ConfidentialComputingGpuNotReady"
	ConfidentialComputingMissingRuntime = "ConfidentialComputingMissingRuntime"
)

func calculateConfidentialComputingConditions(node *corev1.Node) []swagger.NodeStatusConditionDetails {
	if !isConfidentialComputingGpuNode(node.Labels) {
		return []swagger.NodeStatusConditionDetails{}
	}

	result := []swagger.NodeStatusConditionDetails{}
	if !hasEnabledLabel(node.Labels, clusterconstants.AmdSevSnpNodeLabel) &&
		!hasEnabledLabel(node.Labels, clusterconstants.IntelTdxNodeLabel) {
		result = append(result, confidentialCondition(ConfidentialComputingMissingTEE,
			"Neither AMD SEV-SNP nor Intel TDX is enabled on the node (no "+
				clusterconstants.AmdSevSnpNodeLabel+" or "+clusterconstants.IntelTdxNodeLabel+" label)"))
	}

	if !hasEnabledLabel(node.Labels, clusterconstants.NvidiaCCReadyStateLabel) {
		result = append(result, confidentialCondition(ConfidentialComputingGpuNotReady,
			"The node's GPUs are not provisioned for confidential computing ("+
				clusterconstants.NvidiaCCReadyStateLabel+" is not true)"))
	}

	if !hasEnabledLabel(node.Labels, clusterconstants.KataRuntimeNodeLabel) {
		result = append(result, confidentialCondition(ConfidentialComputingMissingRuntime,
			"Kata Containers is not installed on the node (no "+clusterconstants.KataRuntimeNodeLabel+" label)"))
	}

	return result
}

func confidentialCondition(conditionType, message string) swagger.NodeStatusConditionDetails {
	return swagger.NodeStatusConditionDetails{
		Type:    conditionType,
		Reason:  conditionType,
		Message: utils.Ref(message),
	}
}

func isConfidentialComputingGpuNode(nodeLabels map[string]string) bool {
	return hasEnabledLabel(nodeLabels, clusterconstants.KataNvidiaGpuNodeLabel) ||
		shouldNodeHaveNvidiaComponents(nodeLabels)
}

func hasEnabledLabel(nodeLabels map[string]string, key string) bool {
	return nodeLabels[key] == clusterconstants.ConfidentialLabelEnabled
}
