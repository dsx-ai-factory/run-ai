package workloads_test

import (
	"testing"

	"github.com/run-ai/runai/control-plane/bff-service/internal/generated/pkg/api"
	ui_workloads "github.com/run-ai/runai/control-plane/bff-service/pkg/settings/ui-workloads"
	"github.com/stretchr/testify/assert"
)

func TestDeletedWorkloadsViewEnabled_SetDefault(t *testing.T) {
	setting := &ui_workloads.DeletedWorkloadsView{}
	setting.SetDefault()
	assert.True(t, setting.Enabled)
}

func TestDeletedWorkloadsViewEnabled_Validate(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{
			name: "valid DeletedWorkloadsViewEnabled struct",
			input: ui_workloads.DeletedWorkloadsView{
				Enabled: true,
			},
			expected: true,
		},
		{
			name: "valid DeletedWorkloadsViewEnabled pointer",
			input: &ui_workloads.DeletedWorkloadsView{
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
			setting := &ui_workloads.DeletedWorkloadsView{}
			result := setting.Validate(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDeletedWorkflowsViewEnabled_TransformForMigration(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name: "valid GetDeletedWorkloadsView from API with true",
			input: api.GetDeletedWorkloadsView{
				Value: api.GetDeletedWorkloadsViewValue{
					Enabled: boolPtr(true),
				},
			},
			expected: ui_workloads.DeletedWorkloadsView{
				Enabled: true,
			},
		},
		{
			name: "valid GetDeletedWorkloadsView with false",
			input: api.GetDeletedWorkloadsView{
				Value: api.GetDeletedWorkloadsViewValue{
					Enabled: boolPtr(false),
				},
			},
			expected: ui_workloads.DeletedWorkloadsView{
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
			setting := &ui_workloads.DeletedWorkloadsView{}
			result := setting.TransformForMigration(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDeletedWorkflowsViewEnabled_TransformToResponse(t *testing.T) {
	tests := []struct {
		name     string
		setting  ui_workloads.DeletedWorkloadsView
		expected api.GetDeletedWorkloadsView
	}{
		{
			name: "transform enabled true to response",
			setting: ui_workloads.DeletedWorkloadsView{
				Enabled: true,
			},
			expected: api.GetDeletedWorkloadsView{
				Value: api.GetDeletedWorkloadsViewValue{
					Enabled: boolPtr(true),
				},
				DisplayName: "Deleted workloads view",
				Description: "Enable the deleted workloads view",
			},
		},
		{
			name: "transform enabled false to response",
			setting: ui_workloads.DeletedWorkloadsView{
				Enabled: false,
			},
			expected: api.GetDeletedWorkloadsView{
				Value: api.GetDeletedWorkloadsViewValue{
					Enabled: boolPtr(false),
				},
				DisplayName: "Deleted workloads view",
				Description: "Enable the deleted workloads view",
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

func TestDeletedWorkflowsViewEnabled_Metadata(t *testing.T) {
	setting := &ui_workloads.DeletedWorkloadsView{}

	t.Run("Category", func(t *testing.T) {
		assert.Equal(t, "UI-Workloads", setting.Category())
	})

	t.Run("Key", func(t *testing.T) {
		assert.Equal(t, string(api.WORKLOADSSETTINGSKEYENUM_DELETED_WORKLOADS_VIEW), setting.Key())
	})

	t.Run("DisplayName", func(t *testing.T) {
		assert.Equal(t, "Deleted workloads view", setting.DisplayName())
	})

	t.Run("Description", func(t *testing.T) {
		assert.Equal(t, "Enable the deleted workloads view", setting.Description())
	})

	t.Run("Stage", func(t *testing.T) {
		assert.Equal(t, "GA", string(setting.Stage()))
	})
}

func boolPtr(b bool) *bool {
	return &b
}
