/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package accessrule

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/rs/zerolog/log"
)

func (r *AccessRuleReconciler) UpdateStatus(ctx context.Context,
	accessRule *v1alpha1.AccessRule, statusMessage string) error {
	log.Ctx(ctx).Debug().Msgf("Reconciling AccessRule's <%s> Status", accessRule.Name)

	accessRuleStatusCopy := accessRule.Status.DeepCopy()
	accessRule.Status.Message = statusMessage
	if statusMessage == "" {
		accessRule.Status.Phase = v1alpha1.AccessRuleReady
	} else {
		accessRule.Status.Phase = v1alpha1.AccessRuleNotReady
	}

	return r.updateStatusIfNeeded(ctx, accessRule, accessRuleStatusCopy)
}

func (r *AccessRuleReconciler) updateStatusIfNeeded(
	ctx context.Context, accessRule *v1alpha1.AccessRule, accessRuleStatusCopy *v1alpha1.AccessRuleStatus) error {
	if accessRule.Status.Phase == accessRuleStatusCopy.Phase &&
		accessRule.Status.Message == accessRuleStatusCopy.Message {
		// nothing to update
		return nil
	}

	return r.updateStatus(ctx, accessRule)
}

func (r *AccessRuleReconciler) updateStatus(ctx context.Context, accessRule *v1alpha1.AccessRule) error {
	log.Ctx(ctx).Info().Msgf("Updating status of access rule <%s> to: <%s/%s>",
		accessRule.Name, accessRule.Status.Phase, accessRule.Status.Message)

	err := r.cachedClient.Status().Update(ctx, accessRule)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed updating status of access rule <%s>, error: %s",
			accessRule.Name, err.Error())
		return err
	}

	return nil
}
