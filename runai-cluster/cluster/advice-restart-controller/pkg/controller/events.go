/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/dsx-ai-factory/nvidia-recovery/advice-operator/api/v1alpha1"
)

const (
	// Event reasons for the advice-restart-controller.
	AdviceReceivedOnWorkloadEventReason = "AdviceReceivedOnWorkload"
	AdviceIgnoredEventReason            = "AdviceIgnored"
	AdviceProcessedLogOnlyEventReason   = "AdviceProcessedLogOnly"
	RestartRequestCreatedEventReason    = "RestartRequestCreated"
)

func (c *AdviceController) emitAdviceReceivedOnWorkloadEvent(advice *v1alpha1.Advice, topOwner *unstructured.Unstructured, nodeExclusions []string) {
	issueDesc := buildIssueDescription(advice)
	msg := fmt.Sprintf("Advice received from AJR: Restart the workload due to the following issue: %s", issueDesc)
	if nodeExclDesc := buildNodeExclusionDescription(advice, nodeExclusions); nodeExclDesc != "" {
		msg = fmt.Sprintf("Advice received from AJR: Restart the workload on different nodes due to the following issue: %s, and exclude nodes with hardware errors: %s", issueDesc, nodeExclDesc)
	}
	c.recorder.Event(topOwner, corev1.EventTypeNormal, AdviceReceivedOnWorkloadEventReason, msg)
}

func (c *AdviceController) emitAdviceIgnoredNonRestartablePhaseEvent(advice *v1alpha1.Advice, topOwner *unstructured.Unstructured, phase string) {
	msg := fmt.Sprintf("Advice from AJR ignored: Workload can't be restarted in %s state", phase)
	c.recorder.Event(advice, corev1.EventTypeWarning, AdviceIgnoredEventReason, msg)
	c.recorder.Event(topOwner, corev1.EventTypeWarning, AdviceIgnoredEventReason, msg)
}

func (c *AdviceController) emitAdviceIgnoredRestartLimitEvent(advice *v1alpha1.Advice, topOwner *unstructured.Unstructured, maxRestarts int) {
	msg := fmt.Sprintf("Advice from AJR ignored: Workload has reached its restart limit (%d)", maxRestarts)
	c.recorder.Event(advice, corev1.EventTypeWarning, AdviceIgnoredEventReason, msg)
	c.recorder.Event(topOwner, corev1.EventTypeWarning, AdviceIgnoredEventReason, msg)
}

func (c *AdviceController) emitAdviceProcessedLogOnlyEvent(advice *v1alpha1.Advice, topOwner *unstructured.Unstructured) {
	msg := "AJR Advice processed, but no action was taken (integration in log-only mode)"
	c.recorder.Event(advice, corev1.EventTypeNormal, AdviceProcessedLogOnlyEventReason, msg)
	c.recorder.Event(topOwner, corev1.EventTypeNormal, AdviceProcessedLogOnlyEventReason, msg)
}

func (c *AdviceController) emitRestartRequestCreatedEvent(advice *v1alpha1.Advice, topOwner *unstructured.Unstructured) {
	msg := "Restart request created following AJR Advice"
	c.recorder.Event(advice, corev1.EventTypeNormal, RestartRequestCreatedEventReason, msg)
	c.recorder.Event(topOwner, corev1.EventTypeNormal, RestartRequestCreatedEventReason, msg)
}

func buildIssueDescription(advice *v1alpha1.Advice) string {
	if len(advice.Spec.Issues) == 0 {
		return "unspecified issue"
	}
	types := make([]string, 0, len(advice.Spec.Issues))
	for i := range advice.Spec.Issues {
		types = append(types, advice.Spec.Issues[i].IssueType)
	}
	return strings.Join(types, ", ")
}

func buildNodeExclusionDescription(advice *v1alpha1.Advice, nodeExclusions []string) string {
	if len(nodeExclusions) == 0 {
		return ""
	}

	// Build lookup from node name to details
	nodeDetails := make(map[string]string)
	for i := range advice.Spec.Issues {
		for j := range advice.Spec.Issues[i].RemediationActions {
			for k := range advice.Spec.Issues[i].RemediationActions[j].ResourceDetails {
				rd := &advice.Spec.Issues[i].RemediationActions[j].ResourceDetails[k]
				if rd.ResourceType == "Node" && rd.Details != "" {
					nodeDetails[rd.ResourceId] = rd.Details
				}
			}
		}
	}

	parts := make([]string, 0, len(nodeExclusions))
	for _, node := range nodeExclusions {
		if detail, ok := nodeDetails[node]; ok {
			parts = append(parts, fmt.Sprintf("%s: %s", node, detail))
		} else {
			parts = append(parts, node)
		}
	}
	return strings.Join(parts, "; ")
}
