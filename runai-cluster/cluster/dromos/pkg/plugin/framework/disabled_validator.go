/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package framework

import (
	"context"
	"fmt"
	"reflect"

	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// DisabledValidator rejects a create or update that adds or changes the plugin's
// instruction while the plugin is off. instructionOf returns nil when there is no ask.
func DisabledValidator(name string, instructionOf func(*dromosv1alpha1.WorkloadRunner) any, path *field.Path) Plugin {
	return disabledValidator{name: name, instructionOf: instructionOf, path: path}
}

type disabledValidator struct {
	name          string
	instructionOf func(*dromosv1alpha1.WorkloadRunner) any
	path          *field.Path
}

func (d disabledValidator) Name() string { return d.name }

func (d disabledValidator) instruction(runner *dromosv1alpha1.WorkloadRunner) any {
	if runner == nil {
		return nil
	}
	return d.instructionOf(runner)
}

func (d disabledValidator) WebhookValidate(
	_ context.Context,
	existingRunner, desiredRunner *dromosv1alpha1.WorkloadRunner,
) (admission.Warnings, field.ErrorList) {
	want := d.instruction(desiredRunner)
	if want == nil {
		return nil, nil
	}
	if reflect.DeepEqual(d.instruction(existingRunner), want) {
		return nil, nil
	}
	return nil, field.ErrorList{field.Forbidden(d.path,
		fmt.Sprintf("%s is disabled in this cluster; remove the instruction or enable the plugin", d.name))}
}
