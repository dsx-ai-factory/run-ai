/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PodResourceClaim", func() {
	staleClaim := func() PodResourceClaim {
		templateName := "stale-template"
		allocatedName := "stale-allocated"
		return PodResourceClaim{Name: "stale", ResourceClaimTemplateName: &templateName, AllocatedResourceClaimName: &allocatedName}
	}

	It("round-trips through Value and Scan", func() {
		templateName := "gpu-template"
		allocatedName := "worker-0-gpu-abc12"
		original := PodResourceClaim{Name: "gpu", ResourceClaimTemplateName: &templateName, AllocatedResourceClaimName: &allocatedName}

		value, err := original.Value()
		Expect(err).NotTo(HaveOccurred())

		var scanned PodResourceClaim
		Expect(scanned.Scan([]byte(value.(string)))).To(Succeed())
		Expect(scanned).To(Equal(original))
	})

	DescribeTable("resets a reused destination when the value holds no claim",
		func(value interface{}) {
			scanned := staleClaim()
			Expect(scanned.Scan(value)).To(Succeed())
			Expect(scanned).To(Equal(PodResourceClaim{}))
		},
		Entry("SQL NULL", nil),
		Entry("JSON null", []byte("null")),
	)

	It("clears optional fields of a reused destination that the payload omits", func() {
		scanned := staleClaim()
		Expect(scanned.Scan([]byte(`{"name":"fresh"}`))).To(Succeed())
		Expect(scanned).To(Equal(PodResourceClaim{Name: "fresh"}))
	})

	It("returns an error instead of panicking on an unsupported type", func() {
		var scanned PodResourceClaim
		Expect(scanned.Scan("not-bytes")).To(MatchError(ContainSubstring("unsupported type string")))
	})
})
