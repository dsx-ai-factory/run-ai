package common

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const validUUID = "10859ae9-93e2-4fbf-97fa-c586bd721336"

func TestValidateUUID(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty", "", true},
		{"whitespace", "   ", true},
		{"not a uuid", "t1", true},
		{"truncated", "10859ae9-93e2-4fbf-97fa", true},
		{"non-hex char", "10859ae9-93e2-4fbf-97fa-c586bd72133g", true},
		{"unhyphenated", strings.ReplaceAll(validUUID, "-", ""), true},
		{"URN", "urn:uuid:" + validUUID, true},
		{"braced", "{" + validUUID + "}", true},
		{"valid lowercase", validUUID, false},
		{"valid uppercase", strings.ToUpper(validUUID), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateUUID("workloadId", tc.value)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestValidateNumericString(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr string
	}{
		{name: "valid", value: "4500002"},
		{name: "missing", wantErr: "projectId is required"},
		{name: "not numeric", value: "project-a", wantErr: "projectId must be a numeric string"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNumericString("projectId", tt.value)
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestRequiredString(t *testing.T) {
	value, err := RequiredString("request.name", "training")
	assert.NoError(t, err)
	assert.Equal(t, "training", value)

	for _, value := range []any{nil, "", "  ", 42} {
		_, err := RequiredString("request.name", value)
		assert.ErrorContains(t, err, "request.name is required")
	}
}
