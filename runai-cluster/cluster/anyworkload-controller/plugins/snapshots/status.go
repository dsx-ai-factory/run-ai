/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshots

import (
	"context"
	"errors"
	"fmt"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Status reports Ready for an opted-in workload that can use snapshots, and
// Disabled otherwise, with a warning issue when an opt-in can't be honoured.
// Neither moves the workload's phase: an ineligible workload just starts cold.
func (p *Plugin) Status(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	_ client.Client,
	status framework.PluginStatusSetter,
	state *reconcile.State,
) error {
	// Every path but an eligible workload returns without a phase, and the
	// framework starts a plugin at Unknown, which would hold a Running workload
	// in Initializing.
	status.SetPhase(dromosv1alpha1.PluginPhaseDisabled)

	// Opting in has no effect until the admin installs Snapshot (REQ-02-2).
	if !p.snapshotInstalled {
		return nil
	}

	template, err := p.parser.Parse(runner.Spec.Template)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	// Before the lookup, so a workload that never opted in gets no issue from it.
	if !optedIn(template) {
		return nil
	}
	gvk := template.GroupVersionKind()
	karta, err := p.kartas.GetKartaForGVK(ctx, gvk)
	if err != nil {
		return fmt.Errorf("look up karta for %s: %w", gvk, err)
	}

	err = checkEligibility(ctx, template, karta)
	switch {
	case errors.Is(err, errIneligible):
		_ = state.AddPluginIssue(ctx, p, template, err, false, "", issue.Compute)
		return nil
	case err != nil:
		return err
	}

	status.SetPhase(dromosv1alpha1.PluginPhaseReady)
	return nil
}
