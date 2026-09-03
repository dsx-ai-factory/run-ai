/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"fmt"
	"k8s.io/utils/ptr"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/project"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SetObjectMetaLabels(ctx context.Context, workloadKind string, jobType string, workloadName string, objectMeta *metav1.ObjectMeta) error {
	projectName, err := project.GetProjectFromContext(ctx)
	if err != nil {
		return err
	}

	if objectMeta.Labels == nil {
		objectMeta.Labels = make(map[string]string)
	}

	objectMeta.Labels[constants.LabelWorkloadKind] = workloadKind
	objectMeta.Labels[constants.LabelRelease] = workloadName
	objectMeta.Labels[constants.LabelWorkloadName] = workloadName
	objectMeta.Labels[constants.LabelApp] = strings.ToLower(jobType)
	objectMeta.Labels[constants.LabelCreatedBy] = jobType
	objectMeta.Labels[constants.RunaiProjectLabel] = projectName

	return nil
}

func SetPriorityClassLabelOnLabels(ctx context.Context, k8sClient client.Client,
	calculatedPriorityClassNameLabel string, labels map[string]string) error {
	if calculatedPriorityClassNameLabel == "" {
		return nil
	}

	err := validatePriorityClassExists(ctx, k8sClient, calculatedPriorityClassNameLabel)
	if err != nil {
		return fmt.Errorf("failed to set priority class label '%s', error: %s",
			calculatedPriorityClassNameLabel, err.Error())
	}

	labels[constants.LabelPriorityClassName] = calculatedPriorityClassNameLabel
	return nil
}

// AddUserDefinedAnnotationsAndLabels adds user defined annotations and labels to the object
func AddUserDefinedAnnotationsAndLabels(dest *metav1.ObjectMeta, containerFields *field_groups.ContainerFields) {
	if containerFields == nil {
		return
	}

	for key, value := range containerFields.Labels.Deref() {
		dest.Labels[key] = ptr.Deref(value.Value, "")
	}
	for key, value := range containerFields.Annotations.Deref() {
		dest.Annotations[key] = ptr.Deref(value.Value, "")
	}
}

// CopyTopologyAnnotations copy workload specific annotations
func CopyTopologyAnnotations(dest *metav1.ObjectMeta, annotations map[string]string) {
	if annotations == nil {
		return
	}

	for key, value := range annotations {
		if strings.HasPrefix(key, constants.KAITopologyAnnotationPrefix) {
			dest.Annotations[key] = value
		}
	}
}
