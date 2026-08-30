package legacy_controller_rbac

import (
	"context"
	"fmt"
	"testing"

	runutils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

func TestOperandSelectsOnlyDisabledControllerRBAC(t *testing.T) {
	tests := []struct {
		name     string
		expected map[string]struct{}
	}{
		{
			name:     "all controllers enabled",
			expected: map[string]struct{}{},
		},
		{
			name: "workload controller disabled",
			expected: set(
				"*v1.ClusterRole/workload-controller-rw",
				"*v1.ClusterRole/workload-controller-ro",
				"*v1.ClusterRoleBinding/workload-controller-ro",
				"*v1.Role/runai/workload-controller-migration",
				"*v1.RoleBinding/runai/workload-controller-migration",
			),
		},
		{
			name: "inference workload controller disabled",
			expected: set(
				"*v1.ClusterRole/runai-inference-workload-controller-project",
				"*v1.ClusterRole/runai-inference-workload-controller",
				"*v1.ClusterRoleBinding/runai-inference-workload-controller",
			),
		},
		{
			name: "job controller disabled",
			expected: set(
				"*v1.ClusterRole/runai-job-controller",
				"*v1.ClusterRoleBinding/runai-job-controller",
				"*v1.ClusterRole/runai-job-controller-project",
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runaiConfig := runutils.GetDefaultRunaiConfig()
			runaiConfig.Namespace = "runai"
			switch tt.name {
			case "workload controller disabled":
				runaiConfig.Spec.WorkloadController.Enabled = ptr.To(false)
			case "inference workload controller disabled":
				runaiConfig.Spec.InferenceWorkloadController.Enabled = ptr.To(false)
			case "job controller disabled":
				runaiConfig.Spec.JobController.Enabled = ptr.To(false)
			}

			op := Operand(nil, runaiConfig)
			assert.False(t, op.Enabled())

			actual := map[string]struct{}{}
			for _, resource := range op.Resources() {
				object, err := resource.Object(context.Background())
				require.NoError(t, err)
				key := fmt.Sprintf("%T/%s", object, object.GetName())
				if object.GetNamespace() != "" {
					key = fmt.Sprintf("%T/%s/%s", object, object.GetNamespace(), object.GetName())
				}
				actual[key] = struct{}{}
			}

			assert.Equal(t, tt.expected, actual)
		})
	}
}

func set(values ...string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}
