/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package source

import (
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"k8s.io/utils/ptr"
)

// buildKubeflowJobStatusDefinition creates the common status definition used by all Kubeflow training jobs.
// This includes TFJob, PyTorchJob, XGBoostJob, JAXJob, and MPIJob (both v1 and v2beta1).
func buildKubeflowJobStatusDefinition() *kartav1alpha1.StatusDefinition {
	return &kartav1alpha1.StatusDefinition{
		ConditionsDefinition: &kartav1alpha1.ConditionsDefinition{
			Path:             ".status.conditions",
			TypeFieldName:    "type",
			StatusFieldName:  "status",
			MessageFieldName: ptr.To("message"),
		},
		StatusMappings: kartav1alpha1.StatusMappings{
			Initializing: []kartav1alpha1.StatusMatcher{
				{
					ByConditions: []kartav1alpha1.ExpectedCondition{
						{Type: "Created", Status: ptr.To("True")},
					},
				},
			},
			Running: []kartav1alpha1.StatusMatcher{
				{
					ByConditions: []kartav1alpha1.ExpectedCondition{
						{Type: "Running", Status: ptr.To("True")},
					},
				},
			},
			Completed: []kartav1alpha1.StatusMatcher{
				{
					ByConditions: []kartav1alpha1.ExpectedCondition{
						{Type: "Succeeded", Status: ptr.To("True")},
					},
				},
			},
			Failed: []kartav1alpha1.StatusMatcher{
				{
					ByConditions: []kartav1alpha1.ExpectedCondition{
						{Type: "Failed", Status: ptr.To("True")},
					},
				},
			},
			Suspended: []kartav1alpha1.StatusMatcher{
				{
					ByConditions: []kartav1alpha1.ExpectedCondition{
						{Type: "Suspended", Status: ptr.To("True")},
					},
				},
			},
			Suspending: []kartav1alpha1.StatusMatcher{
				{
					// Kubeflow training operator deletes pods immediately on suspend (no graceful
					// PodsStillRunning condition like batch/v1 Job). Detect the brief window between
					// the suspend patch and pods finishing termination by checking that suspend is
					// requested via spec.runPolicy.suspend while at least one replica is still active.
					ByExpression: &kartav1alpha1.ExpressionMatcher{
						Expression:     ".spec.runPolicy.suspend == true and (.status.replicaStatuses // {} | to_entries | map(.value.active // 0) | add // 0) > 0",
						ExpectedResult: "true",
					},
				},
			},
		},
	}
}

// buildKubeflowSuspendDefinition returns the suspend/resume definition for Kubeflow training jobs.
// All Kubeflow training operators (PyTorchJob, TFJob, MPIJob, XGBoostJob, JAXJob) expose
// spec.runPolicy.suspend as the native first-class suspend primitive (operator >= v1.5).
func buildKubeflowSuspendDefinition() *kartav1alpha1.SuspendDefinition {
	return &kartav1alpha1.SuspendDefinition{
		SuspendActions: []kartav1alpha1.SuspendAction{
			{Path: ".spec.runPolicy.suspend", Value: "true"},
		},
		ResumeActions: []kartav1alpha1.SuspendAction{
			{Path: ".spec.runPolicy.suspend", Value: "false"},
		},
	}
}
