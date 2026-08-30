package resource_actions

import (
	"encoding/json"
	"testing"

	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/agent"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func TestPatchRunaiConfig(t *testing.T) {
	tests := []struct {
		name           string
		original       *runaiv1.RunaiConfig
		patch          json.RawMessage
		expectedSpec   *runaiv1.RunaiConfigSpec
		expectedError  bool
		errorSubstring string
	}{
		{
			name: "simple global field patch",
			original: &runaiv1.RunaiConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "runai",
				},
				Spec: runaiv1.RunaiConfigSpec{
					Global: &global.Global{
						SubdomainSupport: ptr.To(false),
					},
				},
			},
			patch: marshalRunaiConfigSpec(t, &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					SubdomainSupport: ptr.To(true),
				},
			}),
			expectedSpec: &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					SubdomainSupport: ptr.To(true),
				},
			},
			expectedError: false,
		},
		{
			name: "patch null on non-existent field (no-op)",
			original: &runaiv1.RunaiConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "runai",
				},
				Spec: runaiv1.RunaiConfigSpec{
					Global: &global.Global{
						SubdomainSupport: ptr.To(false),
					},
					// Agent field is absent
				},
			},
			patch: json.RawMessage(`{"agent": null}`),
			expectedSpec: &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					SubdomainSupport: ptr.To(false),
				},
				// Agent should remain absent
			},
			expectedError: false,
		},
		{
			name: "add new field to existing config",
			original: &runaiv1.RunaiConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "runai",
				},
				Spec: runaiv1.RunaiConfigSpec{
					Global: &global.Global{
						SubdomainSupport: ptr.To(false),
					},
				},
			},
			patch: marshalRunaiConfigSpec(t, &runaiv1.RunaiConfigSpec{
				Agent: &agent.Agent{
					Enabled: ptr.To(true),
				},
			}),
			expectedSpec: &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					SubdomainSupport: ptr.To(false),
				},
				Agent: &agent.Agent{
					Enabled: ptr.To(true),
				},
			},
			expectedError: false,
		},
		{
			name: "patch empty config",
			original: &runaiv1.RunaiConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "runai",
				},
				Spec: runaiv1.RunaiConfigSpec{},
			},
			patch: marshalRunaiConfigSpec(t, &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					SubdomainSupport: ptr.To(false),
				},
			}),
			expectedSpec: &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					SubdomainSupport: ptr.To(false),
				},
			},
			expectedError: false,
		},
		{
			name: "patch with multiple fields",
			original: &runaiv1.RunaiConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "runai",
				},
				Spec: runaiv1.RunaiConfigSpec{
					Global: &global.Global{
						SubdomainSupport: ptr.To(false),
					},
				},
			},
			patch: marshalRunaiConfigSpec(t, &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					SubdomainSupport: ptr.To(true),
				},
				Agent: &agent.Agent{
					Enabled: ptr.To(true),
				},
			}),
			expectedSpec: &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					SubdomainSupport: ptr.To(true),
				},
				Agent: &agent.Agent{
					Enabled: ptr.To(true),
				},
			},
			expectedError: false,
		},
		{
			name: "patch with nil field - deletes existing Agent according to RFC 7396",
			original: &runaiv1.RunaiConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "runai",
				},
				Spec: runaiv1.RunaiConfigSpec{
					Global: &global.Global{
						SubdomainSupport: ptr.To(true),
					},
					Agent: &agent.Agent{
						Enabled: ptr.To(true),
					},
				},
			},
			patch: json.RawMessage(`{"agent": null}`),
			expectedSpec: &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					SubdomainSupport: ptr.To(true),
				},
				// Agent should be deleted (set to nil) according to RFC 7396
				Agent: nil,
			},
			expectedError: false,
		},
		{
			name: "nested field merge with pointer fields",
			original: &runaiv1.RunaiConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "runai",
				},
				Spec: runaiv1.RunaiConfigSpec{
					Global: &global.Global{
						EnableWorkloadOwnershipProtection: ptr.To(false),
						SubdomainSupport:                  ptr.To(true),
					},
				},
			},
			patch: marshalRunaiConfigSpec(t, &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					EnableWorkloadOwnershipProtection: ptr.To(true),
					// SubdomainSupport should remain from original
				},
			}),
			expectedSpec: &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					EnableWorkloadOwnershipProtection: ptr.To(true),
					// SubdomainSupport is not in the patch (omitempty omits nil), so it remains from original
					SubdomainSupport: ptr.To(true),
				},
			},
			expectedError: false,
		},
		{
			name: "patch with metadata preservation",
			original: &runaiv1.RunaiConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "runai",
					Labels: map[string]string{
						"environment": "test",
					},
					Annotations: map[string]string{
						"version": "1.0.0",
					},
				},
				Spec: runaiv1.RunaiConfigSpec{
					Global: &global.Global{
						SubdomainSupport: ptr.To(false),
					},
				},
			},
			patch: marshalRunaiConfigSpec(t, &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					SubdomainSupport: ptr.To(true),
				},
			}),
			expectedSpec: &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					SubdomainSupport: ptr.To(true),
				},
			},
			expectedError: false,
		},
		{
			name: "empty patch - no changes",
			original: &runaiv1.RunaiConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "runai",
				},
				Spec: runaiv1.RunaiConfigSpec{
					Global: &global.Global{
						SubdomainSupport: ptr.To(false),
					},
				},
			},
			patch: marshalRunaiConfigSpec(t, &runaiv1.RunaiConfigSpec{}),
			expectedSpec: &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					SubdomainSupport: ptr.To(false),
				},
			},
			expectedError: false,
		},
		{
			name: "nil original runai config - should handle gracefully",
			original: &runaiv1.RunaiConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "runai",
				},
				Spec: runaiv1.RunaiConfigSpec{},
			},
			patch: marshalRunaiConfigSpec(t, &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					SubdomainSupport: ptr.To(false),
				},
			}),
			expectedSpec: &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					SubdomainSupport: ptr.To(false),
				},
			},
			expectedError: false,
		},
		{
			name: "patch with null field value - should delete field from original",
			original: &runaiv1.RunaiConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "runai",
				},
				Spec: runaiv1.RunaiConfigSpec{
					Global: &global.Global{
						SubdomainSupport:                  ptr.To(false),
						EnableWorkloadOwnershipProtection: ptr.To(true),
					},
					Agent: &agent.Agent{
						Enabled: ptr.To(false),
					},
				},
			},
			patch: json.RawMessage(`{
				"global": {
					"subdomainSupport": true,
					"enableWorkloadOwnershipProtection": null
				},
				"agent": null
			}`),
			expectedSpec: &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					SubdomainSupport: ptr.To(true),
					// EnableWorkloadOwnershipProtection should be deleted (not present)
				},
				// Agent should be deleted (not present)
			},
			expectedError: false,
		},

		{
			name: "patch empty list with new tolerations - should replace completely",
			original: &runaiv1.RunaiConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "runai",
				},
				Spec: runaiv1.RunaiConfigSpec{
					Global: &global.Global{
						SubdomainSupport: ptr.To(true),
						// Tolerations is initially empty/nil
					},
				},
			},
			patch: marshalRunaiConfigSpec(t, &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					Tolerations: []v1.Toleration{
						{
							Key:      "node-type",
							Operator: v1.TolerationOpEqual,
							Value:    "gpu",
							Effect:   v1.TaintEffectNoSchedule,
						},
					},
				},
			}),
			expectedSpec: &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					SubdomainSupport: ptr.To(true),
					Tolerations: []v1.Toleration{
						{
							Key:      "node-type",
							Operator: v1.TolerationOpEqual,
							Value:    "gpu",
							Effect:   v1.TaintEffectNoSchedule,
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "patch existing tolerations list - should replace completely with new list",
			original: &runaiv1.RunaiConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "runai",
				},
				Spec: runaiv1.RunaiConfigSpec{
					Global: &global.Global{
						SubdomainSupport: ptr.To(true),
						Tolerations: []v1.Toleration{
							{
								Key:      "old-key",
								Operator: v1.TolerationOpEqual,
								Value:    "old-value",
								Effect:   v1.TaintEffectNoSchedule,
							},
						},
					},
				},
			},
			patch: marshalRunaiConfigSpec(t, &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					Tolerations: []v1.Toleration{
						{
							Key:      "new-key",
							Operator: v1.TolerationOpEqual,
							Value:    "new-value",
							Effect:   v1.TaintEffectNoExecute,
						},
						{
							Key:      "another-key",
							Operator: v1.TolerationOpExists,
							Effect:   v1.TaintEffectPreferNoSchedule,
						},
					},
				},
			}),
			expectedSpec: &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					SubdomainSupport: ptr.To(true),
					Tolerations: []v1.Toleration{
						{
							Key:      "new-key",
							Operator: v1.TolerationOpEqual,
							Value:    "new-value",
							Effect:   v1.TaintEffectNoExecute,
						},
						{
							Key:      "another-key",
							Operator: v1.TolerationOpExists,
							Effect:   v1.TaintEffectPreferNoSchedule,
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "patch tolerations list with null - should delete the list",
			original: &runaiv1.RunaiConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "runai",
				},
				Spec: runaiv1.RunaiConfigSpec{
					Global: &global.Global{
						SubdomainSupport: ptr.To(true),
						Tolerations: []v1.Toleration{
							{
								Key:      "to-be-deleted",
								Operator: v1.TolerationOpEqual,
								Value:    "delete-me",
								Effect:   v1.TaintEffectNoSchedule,
							},
						},
					},
				},
			},
			patch: json.RawMessage(`{
				"global": {
					"tolerations": null
				}
			}`),
			expectedSpec: &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					SubdomainSupport: ptr.To(true),
					// Tolerations should be nil after setting to null
				},
			},
			expectedError: false,
		},
		{
			name: "patch multiple list fields simultaneously",
			original: &runaiv1.RunaiConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "runai",
				},
				Spec: runaiv1.RunaiConfigSpec{
					Global: &global.Global{
						SubdomainSupport: ptr.To(true),
						Tolerations: []v1.Toleration{
							{
								Key:      "old-toleration",
								Operator: v1.TolerationOpEqual,
								Value:    "old",
								Effect:   v1.TaintEffectNoSchedule,
							},
						},
						DaemonsetsTolerations: []v1.Toleration{
							{
								Key:      "old-daemonset",
								Operator: v1.TolerationOpExists,
								Effect:   v1.TaintEffectNoExecute,
							},
						},
					},
				},
			},
			patch: marshalRunaiConfigSpec(t, &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					Tolerations: []v1.Toleration{
						{
							Key:      "new-toleration",
							Operator: v1.TolerationOpEqual,
							Value:    "new",
							Effect:   v1.TaintEffectPreferNoSchedule,
						},
					},
					DaemonsetsTolerations: []v1.Toleration{
						{
							Key:      "new-daemonset-1",
							Operator: v1.TolerationOpExists,
							Effect:   v1.TaintEffectNoSchedule,
						},
						{
							Key:      "new-daemonset-2",
							Operator: v1.TolerationOpEqual,
							Value:    "daemon-value",
							Effect:   v1.TaintEffectNoExecute,
						},
					},
				},
			}),
			expectedSpec: &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					SubdomainSupport: ptr.To(true),
					Tolerations: []v1.Toleration{
						{
							Key:      "new-toleration",
							Operator: v1.TolerationOpEqual,
							Value:    "new",
							Effect:   v1.TaintEffectPreferNoSchedule,
						},
					},
					DaemonsetsTolerations: []v1.Toleration{
						{
							Key:      "new-daemonset-1",
							Operator: v1.TolerationOpExists,
							Effect:   v1.TaintEffectNoSchedule,
						},
						{
							Key:      "new-daemonset-2",
							Operator: v1.TolerationOpEqual,
							Value:    "daemon-value",
							Effect:   v1.TaintEffectNoExecute,
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "patch list field without specifying in patch - should preserve original list",
			original: &runaiv1.RunaiConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "runai",
				},
				Spec: runaiv1.RunaiConfigSpec{
					Global: &global.Global{
						SubdomainSupport: ptr.To(false),
						Tolerations: []v1.Toleration{
							{
								Key:      "preserve-me",
								Operator: v1.TolerationOpEqual,
								Value:    "preserved",
								Effect:   v1.TaintEffectNoSchedule,
							},
						},
					},
				},
			},
			patch: marshalRunaiConfigSpec(t, &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					SubdomainSupport: ptr.To(true),
					// Tolerations is not specified in patch, should be preserved
				},
			}),
			expectedSpec: &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					SubdomainSupport: ptr.To(true),
					Tolerations: []v1.Toleration{
						{
							Key:      "preserve-me",
							Operator: v1.TolerationOpEqual,
							Value:    "preserved",
							Effect:   v1.TaintEffectNoSchedule,
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "nil patch - should handle gracefully",
			original: &runaiv1.RunaiConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "runai",
				},
				Spec: runaiv1.RunaiConfigSpec{},
			},
			patch:         nil,
			expectedSpec:  &runaiv1.RunaiConfigSpec{},
			expectedError: false,
		},
		// Case for patch that contains fields that do not exist in the original struct
		{
			name: "patch with unknown fields - should ignore unknown fields",
			original: &runaiv1.RunaiConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "runai",
				},
				Spec: runaiv1.RunaiConfigSpec{
					Global: &global.Global{
						SubdomainSupport: ptr.To(false),
					},
				},
			},
			patch: json.RawMessage(`{
				"global": {
					"subdomainSupport": true,
					"unknownField": "should be ignored"
				},
				"nonExistentSection": {
					"someField": 123
				}
			}`),
			expectedSpec: &runaiv1.RunaiConfigSpec{
				Global: &global.Global{
					SubdomainSupport: ptr.To(true),
				},
				// No new fields should be added
			},
			expectedError: false,
		},
		{
			name: "invalid JSON patch - should return error",
			original: &runaiv1.RunaiConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "runai",
				},
				Spec: runaiv1.RunaiConfigSpec{
					Global: &global.Global{
						SubdomainSupport: ptr.To(false),
					},
				},
			},
			patch:          json.RawMessage(`{"global": {"subdomainSupport": true,`), // Malformed JSON
			expectedSpec:   nil,
			expectedError:  true,
			errorSubstring: "failed to apply merge patch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a deep copy of the original config to avoid mutation between tests
			originalCopy := tt.original.DeepCopy()

			// Apply the patch
			err := PatchRunaiConfig(originalCopy, tt.patch)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorSubstring != "" {
					assert.Contains(t, err.Error(), tt.errorSubstring)
				}
				return
			}

			// Verify no error occurred
			assert.NoError(t, err)

			// Verify the spec was patched correctly
			assert.Equal(t, tt.expectedSpec, &originalCopy.Spec)

			// Verify that metadata was preserved (ObjectMeta should not be affected by spec patch)
			assert.Equal(t, tt.original.ObjectMeta, originalCopy.ObjectMeta)
		})
	}
}

func marshalRunaiConfigSpec(t *testing.T, spec *runaiv1.RunaiConfigSpec) json.RawMessage {
	t.Helper()

	specBytes, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("failed to marshal runai config spec: %v", err)
	}
	return json.RawMessage(specBytes)
}
