/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
)

var _ = Describe("helper tests", func() {
	Context("null record", func() {
		It("NoWSTGiven = false", func() {
			Expect(NoWSTIsGiven(nil)).To(Equal(false))
		})
	})

	Context("empty record", func() {
		It("NoWSTGiven = true", func() {
			Expect(NoWSTIsGiven(&api.WorkloadSupportedTypes{})).To(Equal(true))
		})
	})

	Context("NoWSTIsGiven", func() {
		It("NoWSTIsGiven check all WorkloadSupportedTypes types", func() {
			// the lightest way to check if all fields are nil is to check each fields are nil one by one
			// that exactly what NoWSTIsGiven does

			// so if you add new field to WorkloadSupportedTypes you should add it to NoWSTIsGiven check
			// and also update the test

			t := reflect.TypeOf(api.WorkloadSupportedTypes{})
			numFields := t.NumField()
			Expect(numFields).To(Equal(5))
		})
	})

	Context("defaults", func() {
		var meta api.AssetMeta

		setDefaultWST(&meta)
		Expect(meta.WorkloadSupportedTypes).NotTo(BeNil())
		Expect(*meta.WorkloadSupportedTypes.Workspace).To(Equal(true))
		Expect(*meta.WorkloadSupportedTypes.Training).To(Equal(true))
		Expect(*meta.WorkloadSupportedTypes.Distributed).To(Equal(false))
		Expect(*meta.WorkloadSupportedTypes.Inference).To(Equal(false))
		Expect(meta.WorkloadSupportedTypes.DistFramework).To(BeNil())
	})
})
