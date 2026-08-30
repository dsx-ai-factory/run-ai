package util_test

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/util"
)

var testStr = "test"

func TestListToSet(t *testing.T) {
	testCases := []struct {
		desc        string
		list        []string
		expectedMap map[string]struct{}
	}{
		{
			desc:        "test empty list",
			list:        []string{},
			expectedMap: map[string]struct{}{},
		},
		{
			desc:        "test list with one element",
			list:        []string{"test"},
			expectedMap: map[string]struct{}{"test": {}},
		},
		{
			desc:        "test list with two elements",
			list:        []string{"test1", "test2"},
			expectedMap: map[string]struct{}{"test1": {}, "test2": {}},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			tc := tc
			fmt.Printf("#### test case: %s #### \n", tc.desc)
			result := util.ListToSet(tc.list)
			if !reflect.DeepEqual(result, tc.expectedMap) {
				t.Errorf("expected %v, got %v", tc.expectedMap, result)
			}
		})
	}
}

func TestStringPointer(t *testing.T) {
	testCases := []struct {
		desc        string
		input       string
		expectedPtr *string
	}{
		{
			desc:        "test empty string",
			input:       "",
			expectedPtr: nil,
		},
		{
			desc:        "test non empty string",
			input:       testStr,
			expectedPtr: &testStr,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			tc := tc
			fmt.Printf("#### test case: %s #### \n", tc.desc)
			result := util.StringPointer(tc.input)
			if !reflect.DeepEqual(result, tc.expectedPtr) {
				t.Errorf("expected %v, got %v", tc.expectedPtr, result)
			}
		})
	}
}

func TestAggregateRegexList(t *testing.T) {
	testCases := []struct {
		desc          string
		input         []string
		expectedRegex *regexp.Regexp
	}{
		{
			desc:          "test empty list",
			input:         []string{},
			expectedRegex: regexp.MustCompile(""),
		},
		{
			desc:          "test list with one element",
			input:         []string{"test"},
			expectedRegex: regexp.MustCompile("test"),
		},
		{
			desc:          "test list with two elements",
			input:         []string{"test1", "test2"},
			expectedRegex: regexp.MustCompile("test1|test2"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			tc := tc
			fmt.Printf("#### test case: %s #### \n", tc.desc)
			result := util.AggregateRegexList(tc.input)
			if !reflect.DeepEqual(result, tc.expectedRegex) {
				t.Errorf("expected %v, got %v", tc.expectedRegex, result)
			}
		})
	}
}
