/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

var defaultUpdateMode = vpav1.UpdateModeInPlaceOrRecreate

func BuildVPA(name, namespace string, targetRef autoscalingv1.CrossVersionObjectReference, vpaSpec *global.VPASpec) *vpav1.VerticalPodAutoscaler {
	if vpaSpec == nil || vpaSpec.Enabled == nil || !*vpaSpec.Enabled {
		return nil
	}

	updatePolicy := vpaSpec.UpdatePolicy
	if updatePolicy == nil {
		updatePolicy = &vpav1.PodUpdatePolicy{
			UpdateMode: &defaultUpdateMode,
		}
	} else if updatePolicy.UpdateMode == nil {
		updatePolicy.UpdateMode = &defaultUpdateMode
	}

	vpa := &vpav1.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: vpav1.VerticalPodAutoscalerSpec{
			TargetRef:    &targetRef,
			UpdatePolicy: updatePolicy,
		},
	}

	if vpaSpec.ResourcePolicy != nil {
		vpa.Spec.ResourcePolicy = vpaSpec.ResourcePolicy
	}
	return vpa
}
