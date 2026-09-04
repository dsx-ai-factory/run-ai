/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package patching

import (
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type EmptyStruct struct {
}

type MixedStructure struct {
	EmptyStruct *EmptyStruct
	Number      int
}

type OnlyPointers struct {
	EmptyStruct *EmptyStruct
	Num         *int
	Str         *string
	Flag        *bool
}

type NestedStruct struct {
	Num    *int
	Str    *string
	Flag   *bool
	Inner1 *OnlyPointers
	Inner2 *OnlyPointers
}

var _ = Describe("patching test", func() {

	Context("check structure for non-nil pointers", func() {

		It("non structure -> false", func() {
			result := CheckIfAllMembersAreNil([]string{
				"aaa",
			})
			Expect(result).To(Equal(false))
		})

		It("empty structure -> true", func() {
			result := CheckIfAllMembersAreNil(&EmptyStruct{})
			Expect(result).To(Equal(true))
		})

		It("mixed structure -> false", func() {
			result := CheckIfAllMembersAreNil(&MixedStructure{})
			Expect(result).To(Equal(false))
		})

		It("only nil pointers", func() {
			result := CheckIfAllMembersAreNil(&OnlyPointers{})
			Expect(result).To(Equal(true))
		})

		It("non-nil pointers", func() {
			result := CheckIfAllMembersAreNil(&OnlyPointers{
				EmptyStruct: &EmptyStruct{},
			})
			Expect(result).To(Equal(false))
		})
	})

	Context("check patching, only pointers", func() {
		var ptrs OnlyPointers

		BeforeEach(func() {
			ptrs = OnlyPointers{}
		})

		It("no existing value, no new value", func() {
			PatchField(&ptrs.Flag, nil)
			Expect(ptrs.Flag).To(BeNil())
		})

		It("no existing value, non-nil value requested", func() {
			PatchField(&ptrs.Flag, convutils.Ref(false))
			Expect(ptrs.Flag).NotTo(BeNil())
			Expect(*ptrs.Flag).To(Equal(false))
		})

		It("existing value, with no new value", func() {
			ptrs.Flag = convutils.Ref(true)
			PatchField(&ptrs.Flag, nil)
			Expect(ptrs.Flag).NotTo(BeNil())
			Expect(*ptrs.Flag).To(Equal(true))
		})

		It("existing value, replacing it with new value", func() {
			ptrs.Flag = convutils.Ref(true)
			PatchField(&ptrs.Flag, convutils.Ref(false))
			Expect(ptrs.Flag).NotTo(BeNil())
			Expect(*ptrs.Flag).To(Equal(false))
		})
	})

	Context("check nested patching", func() {
		var existingData NestedStruct
		var patchRequest NestedStruct

		BeforeEach(func() {
			existingData = NestedStruct{}
			patchRequest = NestedStruct{}
		})

		It("empty to empty", func() {
			ApplyPatch(&existingData, &patchRequest)

			expected := NestedStruct{}

			Expect(existingData).To(Equal(expected))
		})

		It("first patch to empty", func() {
			patchRequest := NestedStruct{
				Num: convutils.Ref(5),
				Inner1: &OnlyPointers{
					Num: convutils.Ref(3),
					Str: convutils.Ref("aaa"),
				},
			}

			ApplyPatch(&existingData, &patchRequest)

			Expect(existingData).To(Equal(patchRequest))
		})

		It("patch and existing both have data", func() {
			existingData = NestedStruct{
				Num:  convutils.Ref(5),
				Flag: convutils.Ref(false),
				Inner1: &OnlyPointers{
					Num: convutils.Ref(3),
					Str: convutils.Ref("aaa"),
				},
			}

			patchRequest = NestedStruct{
				Num: convutils.Ref(6),
				Inner1: &OnlyPointers{
					Str:  convutils.Ref("bbb"),
					Flag: convutils.Ref(false),
				},
				Inner2: &OnlyPointers{
					Num: convutils.Ref(44),
				},
			}

			ApplyPatch(&existingData, &patchRequest)

			expectedData := NestedStruct{
				Num:  convutils.Ref(6),
				Flag: convutils.Ref(false),
				Inner1: &OnlyPointers{
					Num:  convutils.Ref(3),
					Flag: convutils.Ref(false),
					Str:  convutils.Ref("bbb"),
				},
				Inner2: &OnlyPointers{
					Num: convutils.Ref(44),
				},
			}

			Expect(existingData).To(Equal(expectedData))
		})

		It("clean empty structs", func() {
			existingData = NestedStruct{
				Num: convutils.Ref(5),
				Inner2: &OnlyPointers{
					Num: convutils.Ref(666),
				},
			}

			patchRequest = NestedStruct{
				Num:    convutils.Ref(6),
				Inner1: &OnlyPointers{},
			}

			ApplyPatch(&existingData, &patchRequest)

			expectedData := NestedStruct{
				Num: convutils.Ref(6),
				Inner2: &OnlyPointers{
					Num: convutils.Ref(666),
				},
			}

			Expect(existingData).To(Equal(expectedData))
		})

	})
})
