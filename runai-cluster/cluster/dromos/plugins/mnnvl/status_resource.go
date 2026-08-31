/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mnnvl

import (
	"context"
	"fmt"
	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/reconcile"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// ComputeDomainReadyConditionName indicates whether the MNNVL ComputeDomain is ready
	ComputeDomainReadyConditionName = "ComputeDomainReady"

	ComputeDomainReasonDisabled  = "Disabled"
	ComputeDomainReasonNotFound  = "ComputeDomainNotFound"
	ComputeDomainReasonGetFailed = "ComputeDomainGetFailed"
)

// Status updates the plugin status with MNNVL ComputeDomain status information.
func (p *Plugin) Status(ctx context.Context, runner *dromosv1alpha1.WorkloadRunner, k8sClient client.Client, status framework.PluginStatusSetter, state *reconcile.State) error {
	logger := log.FromContext(ctx)

	// sanitation only, should not be reached in any case.
	if status == nil {
		logger.V(3).Info("nil status for plugin", "workloadRunner", runner.Name, "plugin", p.Name())
		return fmt.Errorf("nil status for plugin %s", p.Name())
	}

	config := extractMnnvlConfig(runner)
	if config == nil || config.Mode == dromosv1alpha1.MnnvlModeNone {
		status.SetCondition(metav1.Condition{
			Type:    ComputeDomainReadyConditionName,
			Status:  metav1.ConditionFalse,
			Reason:  ComputeDomainReasonDisabled,
			Message: "MNNVL is disabled",
		})
		status.SetPhase(dromosv1alpha1.PluginPhaseDisabled)
		return nil
	}

	// Override phase if critical issues exist from earlier reconcile phases (e.g., Karta doesn't support MNNVL)
	for _, iss := range state.PluginIssues(p) {
		if iss.IsCritical {
			status.SetPhase(dromosv1alpha1.PluginPhaseFailed)
			return nil
		}
	}

	computeDomainName := composeComputeDomainName(workloadRunnerInitials, runner.Name)
	computeDomain := &nvapi.ComputeDomain{}
	err := k8sClient.Get(ctx, types.NamespacedName{
		Name:      computeDomainName,
		Namespace: runner.Namespace,
	}, computeDomain)

	if errors.IsNotFound(err) {
		logger.V(1).Info("ComputeDomain not found for WorkloadRunner",
			"workloadRunner", runner.Name,
			"computeDomain", computeDomainName,
		)
		status.SetCondition(metav1.Condition{
			Type:    ComputeDomainReadyConditionName,
			Status:  metav1.ConditionFalse,
			Reason:  ComputeDomainReasonNotFound,
			Message: fmt.Sprintf("ComputeDomain %s not found", computeDomainName),
		})
		status.SetPhase(dromosv1alpha1.PluginPhaseCreating)
		return nil
	}

	if err != nil {
		logger.Error(err, "Failed to get ComputeDomain",
			"workloadRunner", runner.Name,
			"computeDomain", computeDomainName,
		)
		status.SetCondition(metav1.Condition{
			Type:    ComputeDomainReadyConditionName,
			Status:  metav1.ConditionUnknown,
			Reason:  ComputeDomainReasonGetFailed,
			Message: fmt.Sprintf("Failed to get ComputeDomain: %v", err),
		})
		status.SetPhase(dromosv1alpha1.PluginPhaseFailed)
		return fmt.Errorf("failed to get ComputeDomain %s: %w", computeDomainName, err)
	}

	status.SetCondition(metav1.Condition{
		Type:    ComputeDomainReadyConditionName,
		Status:  metav1.ConditionTrue,
		Reason:  ComputeDomainReadyConditionName,
		Message: fmt.Sprintf("ComputeDomain %s is ready", computeDomainName),
	})
	status.SetPhase(dromosv1alpha1.PluginPhaseReady)

	logger.V(1).Info("Updated MNNVL plugin status for WorkloadRunner",
		"workloadRunner", runner.Name,
		"computeDomain", computeDomainName,
	)

	return nil
}
