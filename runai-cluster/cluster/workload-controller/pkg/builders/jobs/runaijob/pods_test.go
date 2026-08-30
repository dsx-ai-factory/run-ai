package runaijob

import (
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	workloadbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	"reflect"
	"testing"
)

func TestFuncName(t *testing.T) {
	// Test Cases
	testCases := []struct {
		name        string
		input       map[string]flds.ExposedUrlItemAndRules
		expected    []string
		shouldError bool
	}{

		{
			name: "Multiple ExposedUrls",
			input: map[string]flds.ExposedUrlItemAndRules{
				"a": {},
				"b": {},
				"c": {},
			},
			expected: []string{"a", "b", "c"},
		},
		{
			name: "Multiple ExposedUrls reverse order",
			input: map[string]flds.ExposedUrlItemAndRules{
				"c": {},
				"b": {},
				"a": {},
			},
			expected: []string{"a", "b", "c"},
		},
		{
			name: "Multiple ExposedUrls mixed order",
			input: map[string]flds.ExposedUrlItemAndRules{
				"c": {},
				"a": {},
				"b": {},
			},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := workloadbuilders.GetSortedNicknames(tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Test Case %s failed.\nExpected: %v\nGot: %v", tc.name, tc.expected, result)
			}
		})
	}
}
