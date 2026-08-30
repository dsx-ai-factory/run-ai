package services

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatMetadataForCSV_DifferentScenarios(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]string
		want     string
	}{
		{
			name:     "Empty map",
			metadata: map[string]string{},
			want:     "",
		},
		{
			name:     "Single key-value pair",
			metadata: map[string]string{"key1": "value1"},
			want:     `key1=value1`,
		},
		{
			name:     "Multiple key-value pairs",
			metadata: map[string]string{"key2": "value2", "key1": "value1"},
			want:     `key1=value1,key2=value2`,
		},
		{
			name:     "Special characters in keys and values",
			metadata: map[string]string{`key"2`: `value"2`, `key1`: `value"1`},
			want:     `key""2=value""2,key1=value""1`,
		},
		{
			name:     "Nil map",
			metadata: nil,
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatMetadataForCSV(tt.metadata)
			require.Equal(t, got, tt.want)
		})
	}
}

func TestEscapeKeyValue_DifferentScenarios(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
		want  string
	}{
		{
			name:  "Normal key-value pair",
			key:   "key1",
			value: "value1",
			want:  `key1=value1`,
		},
		{
			name:  "Key with quotes",
			key:   `key"1`,
			value: "value1",
			want:  `key""1=value1`,
		},
		{
			name:  "Value with quotes",
			key:   "key1",
			value: `value"1`,
			want:  `key1=value""1`,
		},
		{
			name:  "Key and value with quotes",
			key:   `key"1`,
			value: `value"1`,
			want:  `key""1=value""1`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeKeyValue(tt.key, tt.value)
			require.Equal(t, got, tt.want)
		})
	}
}

func TestSortKeys_DifferentScenarios(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]string
		want     []string
	}{
		{
			name:     "Empty map",
			metadata: map[string]string{},
			want:     []string{},
		},
		{
			name:     "Single key",
			metadata: map[string]string{"key1": "value1"},
			want:     []string{"key1"},
		},
		{
			name:     "Multiple keys",
			metadata: map[string]string{"key2": "value2", "key1": "value1"},
			want:     []string{"key1", "key2"},
		},
		{
			name:     "Special characters in keys",
			metadata: map[string]string{"key1": "value1", "key\"2": "value2", "_key": "value3"},
			want:     []string{"_key", "key\"2", "key1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sortKeys(tt.metadata)
			require.Equal(t, got, tt.want)
		})
	}
}
