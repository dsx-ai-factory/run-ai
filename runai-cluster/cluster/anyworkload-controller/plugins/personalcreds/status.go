/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package personalcreds

import (
	"context"

	pcreds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/personalcreds"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	CredentialsReadyConditionType = "CredentialsReady"

	ReasonReady  = "Ready"
	ReasonFailed = "Failed"

	msgAllSecretsReady = "all personal-credential secrets materialized"
)

func (p *Plugin) Watchers(bldr *builder.Builder) (*builder.Builder, error) {
	return bldr.Owns(&corev1.Secret{}), nil
}

func (p *Plugin) Status(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	_ client.Client,
	status framework.PluginStatusSetter,
	state *reconcile.State,
) error {
	logger := log.FromContext(ctx).WithName(PluginName)

	if !pcreds.Has(runner) {
		status.SetPhase(dromosv1alpha1.PluginPhaseDisabled)
		return nil
	}

	for _, iss := range state.PluginIssues(p) {
		if iss.IsCritical {
			logger.V(1).Info("critical issue found in state, marking plugin status as Failed", "issue", iss)
			status.SetCondition(metav1.Condition{
				Type:    CredentialsReadyConditionType,
				Status:  metav1.ConditionFalse,
				Reason:  ReasonFailed,
				Message: iss.Error(),
			})
			status.SetPhase(dromosv1alpha1.PluginPhaseFailed)
			return nil
		}
	}

	status.SetCondition(metav1.Condition{
		Type:    CredentialsReadyConditionType,
		Status:  metav1.ConditionTrue,
		Reason:  ReasonReady,
		Message: msgAllSecretsReady,
	})
	status.SetPhase(dromosv1alpha1.PluginPhaseReady)
	return nil
}
