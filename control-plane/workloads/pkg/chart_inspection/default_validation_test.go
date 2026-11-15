// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package chart_inspection

import (
	"encoding/json"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Question default validation", func() {

	It("reports numeric bounds without rounding large integers", func() {
		minimum := json.Number("18446744073709551617")
		maximum := json.Number("18446744073709551619")
		question := map[string]any{"type": questionInteger, propertyMin: minimum, propertyMax: maximum}
		lowErr := validateAnswer(question, json.Number("18446744073709551616"))
		Expect(lowErr).To(HaveOccurred())
		Expect(lowErr.Message).To(Equal("default is below min (18446744073709551617)"))
		highErr := validateAnswer(question, json.Number("18446744073709551620"))
		Expect(highErr).To(HaveOccurred())
		Expect(highErr.Message).To(Equal("default is above max (18446744073709551619)"))
		Expect(validateAnswer(question, minimum)).To(Succeed())
		Expect(validateAnswer(question, maximum)).To(Succeed())
	})
	It("keeps hostname case and length behavior while explaining the limit", func() {
		question := map[string]any{"type": "hostname"}
		hostname := strings.Repeat("A", 63) + "." + strings.Repeat("b", 63) + "." + strings.Repeat("C", 63) + "." + strings.Repeat("d", 61)
		Expect(validateAnswer(question, hostname)).To(Succeed())
		err := validateAnswer(question, hostname+"d")
		Expect(err).To(HaveOccurred())
		Expect(err.Message).To(Equal("hostname exceeds 253 bytes"))
	})
	It("keeps typed validation errors as warning text without an HTTP prefix", func() {
		archive := chartArchive(
			archiveEntry{name: "app/Chart.yaml", body: "name: app\nversion: '1'\n"},
			archiveEntry{name: "app/questions.yaml", body: "questions:\n- variable: replicas\n  label: Replicas\n  type: int\n  default: invalid\n"},
		)
		result, err := ParseArchive(archive)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Questions).To(BeEmpty())
		Expect(result.Warnings[len(result.Warnings)-1].Message).To(Equal("questions.yaml: questions[0] default: default must match the numeric field type"))
	})
})
