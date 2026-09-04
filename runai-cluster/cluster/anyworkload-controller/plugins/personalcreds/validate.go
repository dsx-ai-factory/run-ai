/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package personalcreds

import (
	"context"

	pcreds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/personalcreds"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (p *Plugin) WebhookValidate(
	ctx context.Context,
	_ *dromosv1alpha1.WorkloadRunner,
	desiredRunner *dromosv1alpha1.WorkloadRunner,
) (admission.Warnings, field.ErrorList) {
	if !pcreds.Has(desiredRunner) {
		return nil, nil
	}

	annPath := field.NewPath("metadata", "annotations").Key(pcreds.Annotation)

	entries, err := pcreds.Get(desiredRunner)
	if err != nil {
		return nil, field.ErrorList{field.Invalid(annPath, desiredRunner.GetAnnotations()[pcreds.Annotation], err.Error())}
	}

	var errs field.ErrorList
	for i, e := range entries {
		entryPath := annPath.Index(i).Child("sourceSecret")
		var src corev1.Secret
		err := p.reader.Get(ctx, types.NamespacedName{Name: e.SourceSecret, Namespace: p.runaiNS}, &src)
		switch {
		case err == nil:
			continue
		case apierrors.IsNotFound(err):
			errs = append(errs, field.NotFound(entryPath, e.SourceSecret))
		default:
			errs = append(errs, field.InternalError(entryPath, err))
		}
	}
	return nil, errs
}
