/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workflows

import (
	"github.com/argoproj/argo-workflows/v4/pkg/apis/workflow"
	wfv1 "github.com/argoproj/argo-workflows/v4/pkg/apis/workflow/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	corev1 "k8s.io/api/core/v1"
)

const (
	helloWorldTemplate = "hello-world"
	helloWorldImage    = "busybox:1.37"

	// The cluster's integrator identifies the pods of a Run:ai workflow by this label.
	workflowPodLabel = "run.ai/workflow"
)

type WorkflowComposer struct {
	fc *composer.FieldsComposer
}

func NewWorkflowComposer(workloadType string, workloadName string, projectName string, namespace string, requester string) *WorkflowComposer {
	return &WorkflowComposer{
		fc: &composer.FieldsComposer{
			WorkloadType: workloadType,
			WorkloadName: workloadName,
			Namespace:    namespace,
			SubmittedBy:  requester,
			ProjectName:  projectName,
		},
	}
}

// Compose translates the workflow spec into an Argo Workflow. Until the composer translates the spec's groups and
// tasks, it sets only the workflow's metadata, and the workflow runs a single placeholder task.
func (wc *WorkflowComposer) Compose(src *workloadsApi.WorkflowSpecSpec, target *wfv1.Workflow) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		meta := &target.ObjectMeta

		runner := &composer.StepRunner{}
		runner.Batch(
			wc.fc.ComposeFullySaasInd(&meta.Labels, composer.BoolToMapString, composer.ControlPlaneManagedLabel),
			wc.fc.CopyBoolean(convutils.Ref(true), &meta.Labels, composer.BoolToMapString, composer.ConvertedFromNativeLabel),
			wc.fc.CopyString(convutils.Ref(wc.fc.SubmittedBy), &meta.Annotations, composer.StringToMapString, composer.SubmittedByAnnotation),

			wc.fc.CopyString(convutils.Ref(workflow.WorkflowKind), &target.Kind, composer.StringToString, nil),
			wc.fc.CopyString(convutils.Ref(workflow.APIVersion), &target.APIVersion, composer.StringToString, nil),

			wc.fc.CopyString(&wc.fc.Namespace, &meta.Namespace, composer.StringToString, nil),
			wc.fc.CopyString(&wc.fc.WorkloadName, &meta.Name, composer.StringToString, nil),
			wc.fc.ComposeNameAndTypeLabels(wc.fc.WorkloadName, wc.fc.WorkloadType, workflow.WorkflowKind, wc.fc.ProjectName, &meta.Labels, composer.MapToMapString, nil),
		)

		target.Spec = wfv1.WorkflowSpec{
			Entrypoint:  helloWorldTemplate,
			PodMetadata: &wfv1.Metadata{Labels: map[string]string{workflowPodLabel: wc.fc.WorkloadName}},
			Templates: []wfv1.Template{{
				Name: helloWorldTemplate,
				Container: &corev1.Container{
					Name:    "main",
					Image:   helloWorldImage,
					Command: []string{"sh", "-c"},
					Args:    []string{"echo this is temporary until we develop the full composer"},
				},
			}},
		}

		return runner.Error()
	}
}
