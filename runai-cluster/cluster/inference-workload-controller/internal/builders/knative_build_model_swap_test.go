package builders

import (
	"testing"
)

func TestKnativeBuildModelSwap(t *testing.T) {
	kbms := &KnativeBuilderModelSwap{}

	tests := []struct {
		name        string
		annotation  string
		expected    *ModelSwapParams
		expectError string
	}{
		{
			name:        "Empty annotation value",
			annotation:  "",
			expected:    nil,
			expectError: "model swap annotation is empty",
		},
		{
			name:       "Valid single model annotation",
			annotation: `{"models": ["modelA"], "server-type": "vllm"}`,
			expected: &ModelSwapParams{
				Models:     []string{"modelA"},
				ServerType: "vllm",
			},
			expectError: "",
		},
		{
			name:       "Valid multiple model annotations",
			annotation: `{"models": ["modelA", "modelB", "modelC"], "server-type": "tgi"}`,
			expected: &ModelSwapParams{
				Models:     []string{"modelA", "modelB", "modelC"},
				ServerType: "tgi",
			},
			expectError: "",
		},
		{
			name:        "Invalid model name",
			annotation:  `{"models": ["modelA", "model@B"], "server-type": "nim"}`,
			expected:    nil,
			expectError: "model swap annotation contains invalid model name: 'model@B'",
		},
		{
			name:        "Missing server type",
			annotation:  `{"models": ["modelA", "modelB"]}`,
			expected:    nil,
			expectError: "model swap annotation is missing server type",
		},
		{
			name:        "Invalid server type",
			annotation:  `{"models": ["modelA", "modelB"], "server-type": "aaa"}`,
			expected:    nil,
			expectError: "model swap annotation contains invalid server type: 'aaa'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := kbms.getModelSwapParamsFromAnnotation(tt.annotation)

			if tt.expectError == "" && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if tt.expectError != "" && (err == nil || err.Error() != tt.expectError) {
				t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
			}

			if !equalParams(params, tt.expected) {
				t.Errorf("Expected params: %v, got: %v", tt.expected, params)
			}
		})
	}
}

func equalParams(a, b *ModelSwapParams) bool {
	if a == nil || b == nil {
		return a == b
	}

	if a.ServerType != b.ServerType {
		return false
	}

	if len(a.Models) != len(b.Models) {
		return false
	}

	for i := range a.Models {
		if a.Models[i] != b.Models[i] {
			return false
		}
	}

	return true
}
