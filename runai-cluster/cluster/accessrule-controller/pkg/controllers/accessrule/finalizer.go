/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package accessrule

import (
	"context"
	"encoding/json"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/accessrule-controller/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
)

func (c *AccessRuleController) addControllerAsFinalizerIfNeeded(ctx context.Context, accessRule *v1alpha1.AccessRule) error {
	if utils.IsItemInList(accessRule.Finalizers, common.FinalizerName) {
		return nil
	}

	log.Info().Msgf("Adding controller to finalizers list of access rule <%s>", accessRule.Name)
	accessRule.Finalizers = append(accessRule.Finalizers, common.FinalizerName)

	if err := c.updateAccessRuleFinalizers(ctx, accessRule); err != nil {
		log.Error().Msgf("Error adding controller to finalizers list of access rule <%s>, error: %s",
			accessRule.Name, err.Error())
		return err
	}
	return nil
}

func (c *AccessRuleController) finalize(ctx context.Context, accessRule *v1alpha1.AccessRule) error {
	log.Info().Msgf("Finalizing deletion of accessrule <%v>", accessRule.Name)

	if !utils.IsItemInList(accessRule.Finalizers, common.FinalizerName) {
		log.Info().Msgf("Got a finalization event for AccessRule <%v>, "+
			"but this controller is not in its finalizers list. Skipping finalization.", accessRule.Name)
		return nil
	}

	statusMessage, err := c.accessRuleReconciler.HandleDeletion(ctx, accessRule)
	if err != nil {
		innerErr := c.accessRuleReconciler.UpdateStatus(ctx, accessRule, statusMessage)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
		}
		return err
	}

	err = c.deleteFinalizer(ctx, accessRule)
	if err != nil {
		return err
	}

	log.Info().Msgf("Done handling deletion of accessrule <%v>", accessRule.Name)
	return err
}

func (c *AccessRuleController) deleteFinalizer(ctx context.Context, accessRule *v1alpha1.AccessRule) error {
	log.Info().Msgf("Removing finalizer <%v> from accessrule's <%v> finalizers list",
		common.FinalizerName, accessRule.Name)
	accessRule.Finalizers = utils.DeleteFirstFromList(accessRule.Finalizers, common.FinalizerName)

	if err := c.updateAccessRuleFinalizers(ctx, accessRule); err != nil {
		log.Error().Msgf("Error removing controller from finalizer list for accessrule <%s>; err: %s",
			accessRule.Name, err.Error())
		return err
	}
	return nil
}

func (c *AccessRuleController) updateAccessRuleFinalizers(ctx context.Context, accessRule *v1alpha1.AccessRule) (err error) {
	patchBytes, err := updateFinalizersPatchBytes(accessRule.Finalizers)
	if err != nil {
		log.Error().Msgf("Failed to json.Marshal patch - update finalizers of accessRule <%v>, err: %v",
			accessRule.Name, err)
		return err
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)
	err = c.cachedClient.Patch(ctx, accessRule, patch)
	if err != nil {
		return err
	}
	return nil
}

func updateFinalizersPatchBytes(finalizers []string) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"metadata": map[string]interface{}{
		"finalizers": finalizers}})
}
