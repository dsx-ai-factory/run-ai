/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

func TestConvertArrayToMap(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		desc        string
		arr         []string
		expectedMap map[string]bool
	}{
		{
			desc:        "test empty array",
			arr:         []string{},
			expectedMap: map[string]bool{},
		},
		{
			desc:        "test array with one element",
			arr:         []string{"test"},
			expectedMap: map[string]bool{"test": true},
		},
		{
			desc:        "test array with two elements",
			arr:         []string{"test1", "test2"},
			expectedMap: map[string]bool{"test1": true, "test2": true},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			tc := tc
			fmt.Printf("#### test case: %s #### \n", tc.desc)
			result := utils.ConvertArrayToMap(tc.arr)
			if !reflect.DeepEqual(result, tc.expectedMap) {
				t.Errorf("expected %v, got %v", tc.expectedMap, result)
			}
		})
	}
}

func TestContains(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		desc        string
		arr         []string
		str         string
		expectedRes bool
	}{
		{
			desc:        "test empty array",
			arr:         []string{},
			str:         "test",
			expectedRes: false,
		},
		{
			desc:        "test array with one element",
			arr:         []string{"test"},
			str:         "test",
			expectedRes: true,
		},
		{
			desc:        "test array with two elements",
			arr:         []string{"test1", "test2"},
			str:         "test2",
			expectedRes: true,
		},
		{
			desc:        "test array with two elements",
			arr:         []string{"test1", "test2"},
			str:         "test3",
			expectedRes: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			tc := tc
			fmt.Printf("#### test case: %s #### \n", tc.desc)
			result := utils.Contains(tc.arr, tc.str)
			if result != tc.expectedRes {
				t.Errorf("expected %v, got %v", tc.expectedRes, result)
			}
		})
	}
}

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
			hostTemplate: "http://{{.}}.example.com",
			name:         "test",
			expectedRes:  "http://test.example.com",
		},
		{
			desc:         "test full url",
			hostTemplate: "abc.example.com",
			name:         "test",
			expectedRes:  "abc.example.com",
		},
		{
			desc:         "test empty",
			hostTemplate: "",
			name:         "test",
			expectedRes:  "",
		},
		{
			desc:         "test uppercase name",
			hostTemplate: "http://{{.}}.example.com",
			name:         "TEST",
			expectedRes:  "http://test.example.com",
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

func TestMergeListsWithUniqueness(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		desc        string
		list1       []string
		list2       []string
		expectedRes []string
	}{
		{
			desc:        "test empty",
			list1:       []string{},
			list2:       []string{},
			expectedRes: []string{},
		},
		{
			desc:        "test one empty",
			list1:       []string{},
			list2:       []string{"test"},
			expectedRes: []string{"test"},
		},
		{
			desc:        "test unique elements",
			list1:       []string{"test1"},
			list2:       []string{"test2"},
			expectedRes: []string{"test1", "test2"},
		},
		{
			desc:        "test duplicate elements",
			list1:       []string{"test1", "test2"},
			list2:       []string{"test2", "test3"},
			expectedRes: []string{"test1", "test2", "test3"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			tc := tc
			fmt.Printf("#### test case: %s #### \n", tc.desc)
			result := utils.MergeListsWithUniqueness(tc.list1, tc.list2)
			if !reflect.DeepEqual(result, tc.expectedRes) {
				t.Errorf("expected %v, got %v", tc.expectedRes, result)
			}
		})
	}
}

func TestGetTenantNameFromHost(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		desc        string
		host        string
		expectedRes string
	}{
		{
			desc:        "test empty",
			host:        "",
			expectedRes: "",
		},
		{
			desc:        "test one empty",
			host:        "test",
			expectedRes: "test",
		},
		{
			desc:        "test default url",
			host:        "test.example.com",
			expectedRes: "test",
		},
		{
			desc:        "test test url",
			host:        "abcd.test.example.com",
			expectedRes: "abcd",
		},
		{
			desc:        "test stagin url",
			host:        "abcd.staging.example.com",
			expectedRes: "abcd",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			tc := tc
			fmt.Printf("#### test case: %s #### \n", tc.desc)
			result := utils.GetTenantNameFromHost(tc.host)
			if !reflect.DeepEqual(result, tc.expectedRes) {
				t.Errorf("expected %v, got %v", tc.expectedRes, result)
			}
		})
	}
}

func TestIsValidUrl_HappyFlow(t *testing.T) {
	assert.True(t, utils.IsValidURL("https://www.example.com/"))
}

func TestIsValidUrl_Error(t *testing.T) {
	assert.False(t, utils.IsValidURL(""))
	assert.False(t, utils.IsValidURL("24234234sdfsf"))
	assert.False(t, utils.IsValidURL("www.example.com"))
	assert.False(t, utils.IsValidURL("https://"))
}

func TestDecodeBase64ToReader_HappyFlow(t *testing.T) {
	reader, err := utils.DecodeBase64String("aGVsbG8gd29ybGQ=")
	assert.Nil(t, err)
	assert.NotNil(t, reader)
	assert.Equal(t, "hello world", string(*reader))
}

func TestDecodeBase64ToReader_InvalidBase64(t *testing.T) {
	reader, err := utils.DecodeBase64String("non_base64_string")
	assert.NotNil(t, err)
	assert.Nil(t, reader)
}

func TestValidateSamlMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata string
		wantErr  bool
	}{
		{
			name:     "entity descriptor",
			metadata: `<md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata" entityID="https://idp.example.com"></md:EntityDescriptor>`,
		},
		{
			name:     "entities descriptor aggregate",
			metadata: `<EntitiesDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata"><EntityDescriptor entityID="https://idp.example.com"/></EntitiesDescriptor>`,
		},
		{
			name:     "entities descriptor without children",
			metadata: `<EntitiesDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata"></EntitiesDescriptor>`,
			wantErr:  true,
		},
		{
			name:     "entities descriptor child without entity id",
			metadata: `<EntitiesDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata"><EntityDescriptor/></EntitiesDescriptor>`,
			wantErr:  true,
		},
		{
			name:     "leading declaration and comment",
			metadata: `<?xml version="1.0"?><!-- exported --><EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" entityID="https://idp.example.com"/>`,
		},
		{
			name:     "entity descriptor without entity id",
			metadata: `<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata"/>`,
			wantErr:  true,
		},
		{
			name:     "plain text",
			metadata: "this is not saml metadata",
			wantErr:  true,
		},
		{
			name:     "empty document",
			metadata: "",
			wantErr:  true,
		},
		{
			name:     "unclosed root element",
			metadata: `<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata">`,
			wantErr:  true,
		},
		{
			name:     "well formed but not saml",
			metadata: `<html><body>login here</body></html>`,
			wantErr:  true,
		},
		{
			name:     "wrong namespace",
			metadata: `<EntityDescriptor xmlns="urn:example:not:saml"/>`,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := utils.ValidateSamlMetadata([]byte(tt.metadata))
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReadUrlContent_HappyFlow(t *testing.T) {
	url := "https://example.com/auth/realms/workAI/protocol/saml/descriptor"
	res, err := utils.ReadUrlContent(context.Background(), url)
	assert.NotNil(t, res)
	assert.Nil(t, err)
}

func TestReadUrlContent_Error(t *testing.T) {
	url := "0"
	res, err := utils.ReadUrlContent(context.Background(), url)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestMergeMaps(t *testing.T) {
	// Define test cases
	tests := []struct {
		name     string
		dest     map[string]string
		src      map[string]string
		expected map[string]string
	}{
		{
			name: "Merge maps with no overlapping keys",
			dest: map[string]string{"a": "1", "b": "2"},
			src:  map[string]string{"c": "3", "d": "4"},
			expected: map[string]string{
				"a": "1",
				"b": "2",
				"c": "3",
				"d": "4",
			},
		},
		{
			name: "Merge maps with overlapping keys and same values",
			dest: map[string]string{"a": "1", "b": "2"},
			src:  map[string]string{"b": "2", "c": "3"},
			expected: map[string]string{
				"a": "1",
				"b": "2",
				"c": "3",
			},
		},
		{
			name: "Merge maps with overlapping keys and different values",
			dest: map[string]string{"a": "1", "b": "2"},
			src:  map[string]string{"b": "3", "c": "4"},
			expected: map[string]string{
				"a": "1",
				"b": "2", // Should not be overwritten because values are different
				"c": "4",
			},
		},
	}

	// Run tests
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			utils.MergeMaps(test.dest, test.src)
			assert.Equal(t, test.expected, test.dest)
		})
	}
}

func TestIntersection(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		desc string
		arr1 []string
		arr2 []string
		res  []string
	}{
		{
			desc: "test empty array",
			arr1: []string{},
			arr2: []string{},
			res:  []string{},
		},
		{
			desc: "test nil",
			arr1: []string{"test"},
			arr2: nil,
			res:  []string{},
		},
		{
			desc: "test empty",
			arr1: []string{"test"},
			arr2: []string{"nottest"},
			res:  []string{},
		},
		{
			desc: "test match",
			arr1: []string{"test", "nottest"},
			arr2: []string{"nottest"},
			res:  []string{"nottest"},
		},
		{
			desc: "test 2 match",
			arr1: []string{"a", "test"},
			arr2: []string{"nottest", "test"},
			res:  []string{"test"},
		},
		{
			desc: "test capital match",
			arr1: []string{"a", "TesT"},
			arr2: []string{"nottesT", "TesT", "a"},
			res:  []string{"TesT", "a"},
		},
		{
			desc: "test capital match",
			arr1: []string{"a", "TesT"},
			arr2: []string{"nottesT", "test", "a"},
			res:  []string{"a"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			tc := tc
			fmt.Printf("#### test case: %s #### \n", tc.desc)
			result := utils.Intersection(tc.arr1, tc.arr2)
			assert.Equal(t, tc.res, result)
		})
	}
}

func TestIso8601ToEpoch(t *testing.T) {
	tests := []struct {
		name     string
		isoTime  string
		expected string
		wantErr  bool
	}{
		{
			name:     "Valid ISO time",
			isoTime:  "2023-09-24T10:15:30Z",
			expected: "1695550530", // This is the epoch time for the above ISO 8601 time.
			wantErr:  false,
		},
		{
			name:    "Invalid ISO time",
			isoTime: "invalid-time-format",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.Iso8601ToEpoch(tt.isoTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("Iso8601ToEpoch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected && !tt.wantErr {
				t.Errorf("Iso8601ToEpoch() got = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestJsonToMap_HappyFlow(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}
	person := Person{
		Name: "John",
		Age:  30,
	}
	res, err := utils.JsonToMap(person)
	assert.Len(t, res, 2)
	assert.Equal(t, "John", res["Name"])
	assert.Equal(t, float64(30), res["Age"])
	assert.Nil(t, err)
}

func TestJsonToMap_HappyFlow_WithPtr(t *testing.T) {
	type Person struct {
		Name string
		Age  *int
	}
	person := Person{
		Name: "John",
		Age:  ptr.To(30),
	}
	res, err := utils.JsonToMap(person)
	assert.Len(t, res, 2)
	assert.Equal(t, "John", res["Name"])
	assert.Equal(t, float64(30), res["Age"])
	assert.Nil(t, err)
}

func TestJsonToMap_HappyFlow_WithNill(t *testing.T) {
	type Person struct {
		Name string
		Age  *int
	}
	person := Person{
		Name: "John",
	}
	res, err := utils.JsonToMap(person)
	assert.Len(t, res, 2)
	assert.Equal(t, "John", res["Name"])
	assert.Nil(t, res["Age"])
	assert.Nil(t, err)
}

func TestJsonToMap_InvalidInput(t *testing.T) {
	res, err := utils.JsonToMap("invalid")
	assert.Nil(t, res)
	assert.NotNil(t, err)
}
