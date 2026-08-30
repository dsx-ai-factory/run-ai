package builders

import (
	"context"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func TestKnativeBuildNodeType(t *testing.T) {
	// Test cases
	tests := []struct {
		name      string
		nodeType  *string
		expectKey bool
	}{
		{
			name:      "With NodeType",
			nodeType:  ptr.To("gpu-node"),
			expectKey: true,
		},
		{
			name:      "Empty NodeType",
			nodeType:  ptr.To(""),
			expectKey: false,
		},
		{
			name:      "Nil NodeType",
			nodeType:  nil,
			expectKey: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a builder
			k8sClient := initClientWithResources(t, namespace)
			kb := NewKnativeBuilder(k8sClient)

			// Create a test workload
			workload := &jobs.JobWorkload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-workload",
					Namespace: namespace,
				},
				Spec: jobs.JobSpec{
					RunaiSchedulerFields: field_groups.RunaiSchedulerFields{
						NodeType: flds.NewStringField(ptr.Deref(tt.nodeType, "")),
					},
					WorkloadFields: field_groups.WorkloadFields{
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

			// Set up context
			projectName := "team-a"
			ctxWithProject := context.WithValue(context.Background(), workloadproject.ContextProjectName, projectName)

			// Build the workload
			resources, err := kb.Build(ctxWithProject, workload, nil)
			assert.NoError(t, err)
			assert.NotNil(t, resources)

			// Get the Knative service
			knativeService, ok := resources.Job.(*servingv1.Service)
			assert.True(t, ok)

			// Check if NodeType was properly set
			nodeSelector := knativeService.Spec.Template.Spec.PodSpec.NodeSelector

			if tt.expectKey {
				assert.NotNil(t, nodeSelector)
				value, exists := nodeSelector[constants.NodeTypeKey]
				assert.True(t, exists)
				assert.Equal(t, *tt.nodeType, value)
			} else {
				assert.Nil(t, nodeSelector, "NodeSelector should be nil when NodeType is not provided")
			}
		})
	}
}
