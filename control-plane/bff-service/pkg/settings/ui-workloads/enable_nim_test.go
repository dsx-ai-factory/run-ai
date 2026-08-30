package workloads

import (
	"testing"

	"github.com/run-ai/runai/control-plane/bff-service/internal/generated/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestEnableNIM_SetDefault(t *testing.T) {
	setting := &EnableNIM{}
	setting.SetDefault()
	assert.True(t, setting.Enabled)
}

func TestEnableNIM_Validate(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{
			name:     "valid EnableNIM struct",
			input:    EnableNIM{Enabled: true},
			expected: true,
		},
		{
			name:     "valid EnableNIM pointer",
			input:    &EnableNIM{Enabled: false},
			expected: true,
		},
		{
			name:     "valid any value",
			input:    "any value",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setting := &EnableNIM{}
			result := setting.Validate(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnableNIM_TransformForMigration(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name: "valid GetEnableNim from API",
			input: api.GetEnableNim{
				Value: api.GetEnableNimValue{
					Enabled: boolPtr(true),
				},
			},
			expected: EnableNIM{Enabled: true},
		},
		{
			name: "valid GetEnableNim with false",
			input: api.GetEnableNim{
				Value: api.GetEnableNimValue{
					Enabled: boolPtr(false),
				},
			},
			expected: EnableNIM{Enabled: false},
		},
		{
			name:     "invalid input type returns as-is",
			input:    "invalid",
			expected: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setting := &EnableNIM{}
			result := setting.TransformForMigration(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}
