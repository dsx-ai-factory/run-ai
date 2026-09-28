/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"reflect"

	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	workloadsManagerApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("copy utils testing", func() {

	Context("cast arrays", func() {
		It("should cast correctly", func() {
			arr1 := []workloadsManagerApi.Annotation{
				{
					Name:  *workloadsManagerApi.NewNullableString(convutils.Ref("aaa")),
					Value: *workloadsManagerApi.NewNullableString(convutils.Ref("bbb")),
				},
				{
					Name:    *workloadsManagerApi.NewNullableString(convutils.Ref("ccc")),
					Exclude: *workloadsManagerApi.NewNullableBool(convutils.Ref(true)),
				},
				{
					Name:  *workloadsManagerApi.NewNullableString(convutils.Ref("ddd")),
					Value: *workloadsManagerApi.NewNullableString(convutils.Ref("eee")),
				},
			}

			arr2 := CastSlice[workloadsManagerApi.Annotation, workloadsApi.Annotation](arr1)
			expectedType := reflect.TypeOf([]workloadsApi.Annotation{})
			Expect(reflect.TypeOf(arr2)).To(Equal(expectedType))

			testutils.Expect(arr1).ToEqual(arr2)
		})
	})

	Context("cast types", func() {
		It("should cast correctly", func() {
			st1 := workloadsManagerApi.Annotation{
				Name:  *workloadsManagerApi.NewNullableString(convutils.Ref("aaa")),
				Value: *workloadsManagerApi.NewNullableString(convutils.Ref("bbb")),
			}

			st2 := CastType[workloadsManagerApi.Annotation, workloadsApi.Annotation](&st1)
			expectedType := reflect.TypeOf(&workloadsApi.Annotation{})
			Expect(reflect.TypeOf(st2)).To(Equal(expectedType))

			testutils.Expect(st1).ToEqual(*st2)
		})
	})
})
