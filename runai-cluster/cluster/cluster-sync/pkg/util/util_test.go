/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package util

import (
	"testing"

	swagger "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

type TestCase struct {
	name      string
	value1    func() swagger.NullableString
	value2    func() swagger.NullableString
	different bool
}

func TestCompareNullables(t *testing.T) {

	tests := []TestCase{
		{
			name:      "Trivial",
			value1:    func() swagger.NullableString { return *swagger.NewNullableString(nil) },
			value2:    func() swagger.NullableString { return *swagger.NewNullableString(nil) },
			different: false,
		},
		{
			name:      "One set with value and one with nil",
			value1:    func() swagger.NullableString { return *swagger.NewNullableString(ptr.To("one")) },
			value2:    func() swagger.NullableString { return *swagger.NewNullableString(nil) },
			different: true,
		},
		{
			name:      "Both value set and not nil",
			value1:    func() swagger.NullableString { return *swagger.NewNullableString(ptr.To("one")) },
			value2:    func() swagger.NullableString { return *swagger.NewNullableString(ptr.To("two")) },
			different: true,
		},
		{
			name:   "One not set - one set with value",
			value1: func() swagger.NullableString { return *swagger.NewNullableString(ptr.To("one")) },
			value2: func() swagger.NullableString {
				v := swagger.NewNullableString(nil)
				v.Unset()
				return *v
			},
			different: true,
		},
		{
			name:   "Both nils - one not set",
			value1: func() swagger.NullableString { return *swagger.NewNullableString(nil) },
			value2: func() swagger.NullableString {
				v := swagger.NewNullableString(nil)
				v.Unset()
				return *v
			},
			different: false,
		},
		{
			name: "Both not set",
			value1: func() swagger.NullableString {
				v := swagger.NewNullableString(nil)
				v.Unset()
				return *v
			},
			value2: func() swagger.NullableString {
				v := swagger.NewNullableString(nil)
				v.Unset()
				return *v
			},
			different: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CompareNullables[string](tt.value1(), tt.value2()); got != tt.different {
				t.Errorf("CompareNullables() = %v, expected %v", got, tt.different)
			}
		})
	}

}

func TestIsPersonalCredentialCopy(t *testing.T) {
	tests := []struct {
		name   string
		meta   *metav1.ObjectMeta
		marked bool
	}{
		{name: "nil metadata"},
		{name: "no labels", meta: &metav1.ObjectMeta{}},
		{
			name: "canonical marker",
			meta: &metav1.ObjectMeta{Labels: map[string]string{
				PersonalCredentialCopyLabel: TrueAsString,
			}},
			marked: true,
		},
		{
			name: "non-canonical marker value",
			meta: &metav1.ObjectMeta{Labels: map[string]string{
				PersonalCredentialCopyLabel: "false",
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPersonalCredentialCopy(tt.meta); got != tt.marked {
				t.Errorf("IsPersonalCredentialCopy() = %v, expected %v", got, tt.marked)
			}
		})
	}
}
