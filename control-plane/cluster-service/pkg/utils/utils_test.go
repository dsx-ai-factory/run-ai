package utils_test

import (
	"fmt"
	"testing"

	"github.com/run-ai/runai/control-plane/cluster-service/pkg/utils"
)

func TestTemplateBaseUrl(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		desc         string
		hostTemplate string
		name         string
		expectedRes  string
	}{
		{
			desc:         "test prefix",
			hostTemplate: "https://{{.}}.example.com",
			name:         "test",
			expectedRes:  "https://test.example.com",
		},
		{
			desc:         "test empty",
			hostTemplate: "",
			name:         "test",
			expectedRes:  "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			tc := tc
			fmt.Printf("#### test case: %s #### \n", tc.desc)
			result, err := utils.TemplateBaseUrl(tc.hostTemplate, tc.name)
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if result != tc.expectedRes {
				t.Errorf("expected %v, got %v", tc.expectedRes, result)
			}
		})
	}
}
