package source

import (
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"k8s.io/utils/ptr"
)

// buildRayServiceStatusDefinition creates the status definition for RayService.
// RayService uses .status.state with values: ready, unhealthy, failed.
// RayService has no native spec.suspend field, so no suspend matchers are defined.
func buildRayServiceStatusDefinition() *kartav1alpha1.StatusDefinition {
	return &kartav1alpha1.StatusDefinition{
		PhaseDefinition: &kartav1alpha1.PhaseDefinition{
			Path: ".status.state",
		},
		StatusMappings: kartav1alpha1.StatusMappings{
			Running: []kartav1alpha1.StatusMatcher{
				{ByPhase: "ready"},
			},
			Failed: []kartav1alpha1.StatusMatcher{
				{ByPhase: "failed"},
			},
		},
	}
}

// buildRayClusterStatusDefinition creates the status definition for RayCluster.
// Uses .status.state (legacy) for Running/Failed.
// Suspend tracking uses expressions for cross-version compatibility:
//   - KubeRay v0.6+ with feature gate disabled: .status.state transitions to "suspended"
//   - KubeRay with RayClusterStatusConditions feature gate: dedicated conditions are set,
//     but .spec.suspend + .status.state provides a safe fallback for both.
//
// Ref: https://github.com/ray-project/kuberay/pull/1711 (suspend support)
//
//	https://github.com/ray-project/kuberay/pull/2403 (RayClusterSuspending/Suspended conditions)
func buildRayClusterStatusDefinition() *kartav1alpha1.StatusDefinition {
	return &kartav1alpha1.StatusDefinition{
		PhaseDefinition: &kartav1alpha1.PhaseDefinition{
			Path: ".status.state",
		},
		StatusMappings: kartav1alpha1.StatusMappings{
			Running: []kartav1alpha1.StatusMatcher{
				{ByPhase: "ready"},
			},
			Failed: []kartav1alpha1.StatusMatcher{
				{ByPhase: "failed"},
			},
			Suspended: []kartav1alpha1.StatusMatcher{
				{
					ByExpression: &kartav1alpha1.ExpressionMatcher{
						Expression:     ".spec.suspend == true and (.status.state == \"suspended\" or (.status.state | not))",
						ExpectedResult: "true",
					},
				},
			},
			Suspending: []kartav1alpha1.StatusMatcher{
				{
					ByExpression: &kartav1alpha1.ExpressionMatcher{
						Expression:     ".spec.suspend == true and .status.state != null and .status.state != \"suspended\"",
						ExpectedResult: "true",
					},
				},
			},
		},
	}
}

// buildRayJobStatusDefinition creates the status definition for RayJob.
// Uses .status.jobStatus with values: PENDING, RUNNING, STOPPED, SUCCEEDED, FAILED.
// Suspend phases use .status.jobDeploymentStatus with values: Suspending, Suspended.
func buildRayJobStatusDefinition() *kartav1alpha1.StatusDefinition {
	return &kartav1alpha1.StatusDefinition{
		PhaseDefinition: &kartav1alpha1.PhaseDefinition{
			Path: ".status.jobStatus",
		},
		ConditionsDefinition: &kartav1alpha1.ConditionsDefinition{
			Path:             ".status.conditions",
			TypeFieldName:    "type",
			StatusFieldName:  "status",
			MessageFieldName: ptr.To("message"),
		},
		StatusMappings: kartav1alpha1.StatusMappings{
			Initializing: []kartav1alpha1.StatusMatcher{
				{ByPhase: "PENDING"},
			},
			Running: []kartav1alpha1.StatusMatcher{
				{ByPhase: "RUNNING"},
			},
			Completed: []kartav1alpha1.StatusMatcher{
				{ByPhase: "SUCCEEDED"},
			},
			Failed: []kartav1alpha1.StatusMatcher{
				{ByPhase: "FAILED"},
			},
			// KubeRay surfaces suspend progress via .status.jobDeploymentStatus,
			// which transitions: Running → Suspending → Suspended on suspend,
			// and Suspended → Initializing → Running on resume.
			Suspended: []kartav1alpha1.StatusMatcher{
				{
					ByExpression: &kartav1alpha1.ExpressionMatcher{
						Expression:     ".status.jobDeploymentStatus == \"Suspended\"",
						ExpectedResult: "true",
					},
				},
			},
			Suspending: []kartav1alpha1.StatusMatcher{
				{
					ByExpression: &kartav1alpha1.ExpressionMatcher{
						Expression:     ".status.jobDeploymentStatus == \"Suspending\"",
						ExpectedResult: "true",
					},
				},
			},
		},
	}
}

// buildRayJobSuspendDefinition returns the suspend/resume definition for RayJob.
// KubeRay natively supports spec.suspend since v0.6 (merged 2023-05).
// Setting spec.suspend=true deletes the RayCluster and stops the job;
// setting it back to false re-creates the cluster and resubmits.
func buildRayJobSuspendDefinition() *kartav1alpha1.SuspendDefinition {
	return &kartav1alpha1.SuspendDefinition{
		SuspendActions: []kartav1alpha1.SuspendAction{
			{Path: ".spec.suspend", Value: "true"},
		},
		ResumeActions: []kartav1alpha1.SuspendAction{
			{Path: ".spec.suspend", Value: "false"},
		},
	}
}
