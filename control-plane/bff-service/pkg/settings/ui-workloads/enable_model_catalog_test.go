package workloads_test

import (
	"testing"

	"github.com/run-ai/runai/control-plane/bff-service/internal/generated/pkg/api"
	ui_workloads "github.com/run-ai/runai/control-plane/bff-service/pkg/settings/ui-workloads"
	"github.com/stretchr/testify/assert"
)

func TestEnableModelCatalog_SetDefault(t *testing.T) {
	setting := &ui_workloads.EnableModelCatalog{}
	setting.SetDefault()
	assert.True(t, setting.Enabled)
}

func TestEnableModelCatalog_Validate(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{
			name: "valid EnableModelCatalog struct",
			input: ui_workloads.EnableModelCatalog{
				Enabled: true,
			},
			expected: true,
		},
		{
			name: "valid EnableModelCatalog pointer",
			input: &ui_workloads.EnableModelCatalog{
				Enabled: false,
			},
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
			setting := &ui_workloads.EnableModelCatalog{}
			result := setting.Validate(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnableModelCatalog_TransformForMigration(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:  "valid bool from backend with true",
			input: true,
			expected: ui_workloads.EnableModelCatalog{
				Enabled: true,
			},
		},
		{
			name:  "valid bool with false",
			input: false,
			expected: ui_workloads.EnableModelCatalog{
				Enabled: false,
			},
		},
		{
			name:     "invalid input type returns as-is",
			input:    "invalid",
			expected: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setting := &ui_workloads.EnableModelCatalog{}
			result := setting.TransformForMigration(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnableModelCatalog_TransformToResponse(t *testing.T) {
	description := "Enable instant deployment of Hugging Face models via the Models view or the workload submission form. " +
		"For more information, see the <a href='https://run-ai-docs.nvidia.com/self-hosted/workloads-in-nvidia-run-ai/using-inference/hugging-face-inference' target='_blank'>Deploy Inference Workloads from Hugging Face</a>"

	tests := []struct {
		name     string
		setting  ui_workloads.EnableModelCatalog
		expected api.GetEnableModelCatalog
	}{
		{
			name: "transform enabled true to response",
			setting: ui_workloads.EnableModelCatalog{
				Enabled: true,
			},
			expected: api.GetEnableModelCatalog{
				Value: api.GetEnableModelCatalogValue{
					Enabled: boolPtr(true),
				},
				DisplayName: "Hugging face models",
				Description: description,
				Stage:       "GA",
			},
		},
		{
			name: "transform enabled false to response",
			setting: ui_workloads.EnableModelCatalog{
				Enabled: false,
			},
			expected: api.GetEnableModelCatalog{
				Value: api.GetEnableModelCatalogValue{
					Enabled: boolPtr(false),
				},
				DisplayName: "Hugging face models",
				Description: description,
				Stage:       "GA",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.setting.TransformToResponse()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnableModelCatalog_Metadata(t *testing.T) {
	setting := &ui_workloads.EnableModelCatalog{}

	t.Run("Category", func(t *testing.T) {
		assert.Equal(t, "UI-Workloads", setting.Category())
	})

	t.Run("Key", func(t *testing.T) {
		assert.Equal(t, string(api.WORKLOADSSETTINGSKEYENUM_ENABLE_MODEL_CATALOG), setting.Key())
	})

	t.Run("DisplayName", func(t *testing.T) {
		assert.Equal(t, "Hugging face models", setting.DisplayName())
	})

	t.Run("Description", func(t *testing.T) {
		assert.Contains(t, setting.Description(), "Hugging Face models")
	})

	t.Run("Stage", func(t *testing.T) {
		assert.Equal(t, "GA", string(setting.Stage()))
	})
}
