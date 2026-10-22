/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package inferenceendpoint

import (
	"context"
	ep "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromos "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (p *Plugin) WebhookValidate(_ context.Context, old *dromos.WorkloadRunner, runner *dromos.WorkloadRunner) (admission.Warnings, field.ErrorList) {
	path := field.NewPath("metadata", "annotations").Key(ep.InferenceEndpointAnnotation)
	cfg, err := ep.GetInferenceEndpoint(runner)
	if err != nil {
		return nil, field.ErrorList{field.Invalid(path, "<inference routing intent>", err.Error())}
	}
	if cfg == nil {
		return nil, nil
	}
	// Availability is a precondition for new intent only. Finalizer removal is an
	// UPDATE, and metadata-only updates carry the same intent; neither may be
	// rejected because the InferenceEndpoint API went away after admission.
	if !runner.DeletionTimestamp.IsZero() {
		return nil, nil
	}
	if old != nil && old.Annotations[ep.InferenceEndpointAnnotation] == runner.Annotations[ep.InferenceEndpointAnnotation] {
		return nil, nil
	}
	if err := p.available(); err != nil {
		return nil, field.ErrorList{field.Forbidden(path, err.Error())}
	}
	return nil, nil
}
