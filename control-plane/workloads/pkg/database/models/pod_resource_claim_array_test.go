/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PodResourceClaimArray", func() {
	It("round-trips a populated array through Value and Scan", func() {
		templateName := "gpu-template"
		allocatedName := "worker-0-gpu-abc12"
		standaloneName := "shared-gpu"
		original := PodResourceClaimArray{
			{Name: "gpu", ResourceClaimTemplateName: &templateName, AllocatedResourceClaimName: &allocatedName},
			{Name: "shared", ResourceClaimName: &standaloneName, AllocatedResourceClaimName: &standaloneName},
		}

		value, err := original.Value()
		Expect(err).NotTo(HaveOccurred())

		var scanned PodResourceClaimArray
		Expect(scanned.Scan([]byte(value.(string)))).To(Succeed())
		Expect(scanned).To(Equal(original))
	})

	DescribeTable("clears a reused destination when the column holds no claims",
		func(value interface{}) {
			scanned := PodResourceClaimArray{{Name: "stale"}}
			Expect(scanned.Scan(value)).To(Succeed())
			Expect(scanned).To(BeNil())
		},
		Entry("SQL NULL", nil),
		Entry("JSON null", []byte("null")),
		Entry("empty bytes", []byte{}),
	)

	It("replaces, rather than appends to, a reused destination", func() {
		scanned := PodResourceClaimArray{{Name: "stale-1"}, {Name: "stale-2"}}
		Expect(scanned.Scan([]byte(`[{"name":"fresh"}]`))).To(Succeed())
		Expect(scanned).To(Equal(PodResourceClaimArray{{Name: "fresh"}}))
	})
})
