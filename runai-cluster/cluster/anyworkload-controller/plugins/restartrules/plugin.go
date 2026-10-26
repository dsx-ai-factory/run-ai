/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package restartrules validates the WorkloadRunner restart ownership protocol
// (spec.managedBy + the run.ai/request-ref label) on the dromos validating webhook.
package restartrules

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/internal/restart"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/utils/validator"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
)

const (
	PluginName = "restartrules"

	managedByDeniedMessage = "The managedBy field is reserved for system controllers and cannot be modified manually."
)

var (
	managedByPath  = field.NewPath("spec", "managedBy")
	requestRefPath = field.NewPath("metadata", "labels").Key(restart.RequestRefLabel)
)

type Plugin struct {
	allowedServiceAccounts []string
}

// NewPlugin builds the restartrules plugin with an explicit allowed-service-account
// identity list. Useful for tests that want to wire it directly.
func NewPlugin(allowedServiceAccounts []string) *Plugin {
	return &Plugin{allowedServiceAccounts: allowedServiceAccounts}
}

func NewBuilder(allowedServiceAccounts []string) framework.PluginBuilder {
	return func(context.Context, framework.PluginDeps) (framework.Plugin, error) {
		return NewPlugin(allowedServiceAccounts), nil
	}
}

func (p *Plugin) Name() string { return PluginName }

// WebhookValidate enforces the restart ownership protocol on WorkloadRunner writes.
func (p *Plugin) WebhookValidate(
	ctx context.Context,
	existingRunner *dromosv1alpha1.WorkloadRunner,
	desiredRunner *dromosv1alpha1.WorkloadRunner,
) (admission.Warnings, field.ErrorList) {
	var existingManagedBy string
	if existingRunner != nil {
		existingManagedBy = existingRunner.Spec.ManagedBy
	}

	if desiredRunner.Spec.ManagedBy != "" && desiredRunner.Spec.ManagedBy != restart.RestartControllerManagedBy {
		return nil, field.ErrorList{field.NotSupported(managedByPath, desiredRunner.Spec.ManagedBy,
			[]string{"", restart.RestartControllerManagedBy})}
	}

	// Verify SA can write spec.managedBy
	if desiredRunner.Spec.ManagedBy != existingManagedBy || desiredRunner.Spec.ManagedBy != "" {
		req, err := admission.RequestFromContext(ctx)
		if err != nil {
			// no requester identity means the gate cannot be enforced - refuse rather than skip it
			return nil, field.ErrorList{field.InternalError(managedByPath, err)}
		}
		if !validator.AllowedServiceAccounts(req.UserInfo.Username, p.allowedServiceAccounts) {
			return nil, field.ErrorList{field.Forbidden(managedByPath, managedByDeniedMessage)}
		}
	}

	if existingRunner == nil {
		return nil, nil // On CREATE exit early
	}

	desiredRef, desiredHasRef := desiredRunner.Labels[restart.RequestRefLabel]
	existingRef, existingHasRef := existingRunner.Labels[restart.RequestRefLabel]

	var errs field.ErrorList

	if existingManagedBy != "" && desiredRunner.Spec.ManagedBy != "" && existingManagedBy != desiredRunner.Spec.ManagedBy {
		errs = append(errs, field.Invalid(managedByPath, desiredRunner.Spec.ManagedBy,
			fmt.Sprintf("cannot be changed while managed by %q", existingManagedBy)))
	}

	// request-ref label validation
	switch {
	case desiredRunner.Spec.ManagedBy == restart.RestartControllerManagedBy && desiredRef == "":
		errs = append(errs, field.Required(requestRefPath,
			fmt.Sprintf("must be set to a non-empty value when spec.managedBy is %q", restart.RestartControllerManagedBy)))
	case desiredRunner.Spec.ManagedBy != restart.RestartControllerManagedBy && desiredHasRef:
		errs = append(errs, field.Forbidden(requestRefPath,
			fmt.Sprintf("must be removed when spec.managedBy is not %q", restart.RestartControllerManagedBy)))
	}

	// adding and removing the label is only for handoff/handback
	if existingHasRef && desiredHasRef && existingRef != desiredRef {
		errs = append(errs, field.Invalid(requestRefPath, desiredRef, "cannot be modified once set"))
	}

	return nil, errs
}

var _ framework.ValidatorPlugin = &Plugin{}
