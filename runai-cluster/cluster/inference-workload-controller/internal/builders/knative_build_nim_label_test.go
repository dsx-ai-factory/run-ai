/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func TestKnativeBuildCopiesNimLabelToRevisionTemplate(t *testing.T) {
	tests := []struct {
		name           string
		workloadLabels map[string]string
		expectNimLabel bool
		expectedValue  string
	}{
		{
			name: "With NIM label on workload",
			workloadLabels: map[string]string{
				constants.LabelNIMWorkload: "true",
			},
			expectNimLabel: true,
			expectedValue:  "true",
		},
		{
			name:           "Without NIM label on workload",
			workloadLabels: map[string]string{"app": "test"},
			expectNimLabel: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k8sClient := initClientWithResources(t, namespace)
			kb := NewKnativeBuilder(k8sClient)

			workload := &jobs.JobWorkload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-workload-nim-label",
					Namespace: namespace,
				},
				Spec: jobs.JobSpec{
					WorkloadFields: field_groups.WorkloadFields{
						ContainerFields: field_groups.ContainerFields{
							Labels: flds.NewDictionaryField(tt.workloadLabels),
						},
						NetworkingFields: field_groups.NetworkingFields{
							Ports: &flds.PortsField{
								Value: &map[string]flds.PortItemAndRules{
									"serving-port": {
										Value: &flds.PortItemValue{
											ContainerPort: ptr.To(8080),
											ServiceType:   ptr.To(constants.ServiceTypeServingPort),
										},
									},
								},
							},
						},
					},
				},
			}

			// Set up context with project name
			projectName := "team-a"
			ctxWithProject := context.WithValue(context.Background(), project.ContextProjectName, projectName)

			resources, err := kb.Build(ctxWithProject, workload, nil)
			assert.NoError(t, err)
			assert.NotNil(t, resources)

			knativeService, ok := resources.Job.(*servingv1.Service)
			assert.True(t, ok)

			labels := knativeService.Spec.Template.ObjectMeta.Labels

			if tt.expectNimLabel {
				val, exists := labels[constants.LabelNIMWorkload]
				assert.True(t, exists)
				assert.Equal(t, tt.expectedValue, val)
			} else {
				_, exists := labels[constants.LabelNIMWorkload]
				assert.False(t, exists)
			}
		})
	}
}
