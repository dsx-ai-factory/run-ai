/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package configmaps

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

const (
	confidentialComputingComponentName = "confidential-computing"
	initdataConfigMapKey               = "initdata.toml"
)

func withConfidentialComputingStatus(gpuStackStatus, confidentialStatus runv1.PublicStatus) runv1.PublicStatus {
	if gpuStackStatus.Components == nil {
		gpuStackStatus.Components = map[string]runv1.PublicStatus{}
	}
	gpuStackStatus.Components[confidentialComputingComponentName] = confidentialStatus

	if !*confidentialStatus.Available && *gpuStackStatus.Available {
		gpuStackStatus.Available = runv1.AvailableStatus(false)
		gpuStackStatus.Reason = confidentialStatus.Reason
	}
	return gpuStackStatus
}

func getConfidentialComputingStatus(
	ctx context.Context, nodeReader, configMapReader client.Reader, initdataConfigMap string,
) runv1.PublicStatus {
	nodesProblem, nodesSummary := checkConfidentialNodes(ctx, nodeReader)
	problems := []string{}
	if nodesProblem != "" {
		problems = append(problems, nodesProblem)
	}
	if initdataProblem := checkInitdataConfigMap(ctx, configMapReader, initdataConfigMap); initdataProblem != "" {
		problems = append(problems, initdataProblem)
	}

	if len(problems) > 0 {
		return runv1.PublicStatus{
			Available: runv1.AvailableStatus(false),
			Reason:    strings.Join(problems, "; "),
		}
	}
	return runv1.PublicStatus{
		Available: runv1.AvailableStatus(true),
		Reason:    nodesSummary,
	}
}

func checkConfidentialNodes(ctx context.Context, nodeReader client.Reader) (problem string, summary string) {
	nodes := &corev1.NodeList{}
	err := nodeReader.List(ctx, nodes, client.MatchingLabels{
		clusterconstants.KataNvidiaGpuNodeLabel: clusterconstants.ConfidentialLabelEnabled,
	})
	if err != nil {
		return fmt.Sprintf("failed to list GPU nodes: %v", err), ""
	}

	if len(nodes.Items) == 0 {
		return fmt.Sprintf("no node is labeled %s=true; KRAB is not installed or found no NVIDIA GPU",
			clusterconstants.KataNvidiaGpuNodeLabel), ""
	}

	ready := 0
	for _, node := range nodes.Items {
		if isConfidentialReadyNode(node.Labels) {
			ready++
		}
	}

	summary = fmt.Sprintf("%d of %d GPU nodes are ready for confidential workloads", ready, len(nodes.Items))
	if ready == 0 {
		return summary, ""
	}
	if ready == len(nodes.Items) {
		return "", ""
	}
	return "", summary
}

func isConfidentialReadyNode(labels map[string]string) bool {
	enabled := func(key string) bool { return labels[key] == clusterconstants.ConfidentialLabelEnabled }
	return (enabled(clusterconstants.AmdSevSnpNodeLabel) || enabled(clusterconstants.IntelTdxNodeLabel)) &&
		enabled(clusterconstants.NvidiaCCReadyStateLabel) &&
		enabled(clusterconstants.KataRuntimeNodeLabel)
}

func checkInitdataConfigMap(ctx context.Context, configMapReader client.Reader, name string) string {
	if name == "" {
		return ""
	}

	configMap := &corev1.ConfigMap{}
	err := configMapReader.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, configMap)
	if kerrors.IsNotFound(err) {
		return fmt.Sprintf("initdata ConfigMap %s/%s not found", namespace, name)
	}
	if err != nil {
		return fmt.Sprintf("failed to get initdata ConfigMap %s/%s: %v", namespace, name, err)
	}

	if strings.TrimSpace(configMap.Data[initdataConfigMapKey]) == "" {
		return fmt.Sprintf("initdata ConfigMap %s/%s has no %s key or it is empty", namespace, name, initdataConfigMapKey)
	}
	return ""
}
