/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package topology

import (
	"context"
	"errors"
	"fmt"

	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"k8s.io/utils/ptr"
)

const (
	levelErrorTemplate = "topology is enabled, but the '%s' annotation is set with no value"
)

var (
	errTopologyNameIsEmpty      = errors.New(fmt.Sprintf("the '%s' annotation is set with no value", KAITopologyAnnotationKey))
	errTopologyPreferredMissing = errors.New(fmt.Sprintf(levelErrorTemplate, KAITopologyAnnotationPreferredKey))
	errTopologyRequiredMissing  = errors.New(fmt.Sprintf(levelErrorTemplate, KAITopologyAnnotationRequiredKey))
	errTopologyLevelMissing     = errors.New(fmt.Sprintf("topology is enabled, but no topology level is defined. expected '%s' or '%s'", KAITopologyAnnotationPreferredKey, KAITopologyAnnotationRequiredKey))
)

func CopySpecTopologyAnnotation(ctx context.Context, annotations map[string]string, specAnnotationsField *flds.DictionaryField) error {
	specAnnotations := specAnnotationsField.Deref()
	if len(specAnnotations) == 0 {
		return nil
	}

	name, ok := specAnnotations[KAITopologyAnnotationKey]
	if !ok {
		return nil
	}

	annotations[KAITopologyAnnotationKey] = ptr.Deref(name.Value, "")
	preferred, ok := specAnnotations[KAITopologyAnnotationPreferredKey]
	if ok {
		annotations[KAITopologyAnnotationPreferredKey] = ptr.Deref(preferred.Value, "")
	}

	required, ok := specAnnotations[KAITopologyAnnotationRequiredKey]
	if ok {
		annotations[KAITopologyAnnotationRequiredKey] = ptr.Deref(required.Value, "")
	}

	return nil
}

func ValidateTopologyAnnotation(ctx context.Context, annotations map[string]string) error {
	name, ok := annotations[KAITopologyAnnotationKey]
	if !ok {
		return nil
	}

	if len(name) == 0 {
		return errTopologyNameIsEmpty
	}

	annotationPreferred, okPreferred := annotations[KAITopologyAnnotationPreferredKey]
	annotationRequired, okRequired := annotations[KAITopologyAnnotationRequiredKey]
	if okPreferred && annotationPreferred == "" {
		return errTopologyPreferredMissing
	}

	if okRequired && annotationRequired == "" {
		return errTopologyRequiredMissing
	}

	if !okRequired && !okPreferred {
		return errTopologyLevelMissing
	}

	return nil
}
