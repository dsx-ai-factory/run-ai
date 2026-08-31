/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mutator

import (
	"context"
	"fmt"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/mutator/gvk"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type Mutator interface {
	Name() string
	Mutate(ctx context.Context, runner *dromosv1alpha1.WorkloadRunner) (admission.Warnings, error)
}

type ControllerMutator struct {
	mutators []Mutator
}

func NewControllerMutator() *ControllerMutator {
	return &ControllerMutator{mutators: []Mutator{gvk.NewLabelMutator()}}
}

func (v *ControllerMutator) Run(ctx context.Context, runner *dromosv1alpha1.WorkloadRunner) (admission.Warnings, error) {
	logger := log.FromContext(ctx)

	logger.Info("Running ControllerMutator", "mutator", len(v.mutators))

	var allWarnings admission.Warnings
	for _, mutator := range v.mutators {
		warnings, err := mutator.Mutate(ctx, runner)
		if len(warnings) > 0 {
			logger.Info("Mutator warnings", "mutator", mutator.Name(), "warnings", warnings)
			allWarnings = append(allWarnings, warnings...)
		}
		if err != nil {
			logger.Error(err, "Mutation failed", "mutator", mutator.Name())
			return nil, fmt.Errorf("mutation failed in mutator %s: %w", mutator.Name(), err)
		}
	}

	return allWarnings, nil
}
