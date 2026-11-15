// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package chart_inspection

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Values paths and visibility conditions", func() {
	It("parses array indices and quoted keys containing dots", func() {
		got, err := parsePath(`workers[0]["display.name"].enabled`)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal([]any{"workers", 0, "display.name", "enabled"}))
	})

	invalidPaths := []TableEntry{}
	for _, value := range []string{"", "a.0", "a[-1]", "a[10001]", "a.constructor", `a["prototype"]`, `a["__proto__"]`, "a..b", "a[b]"} {
		invalidPaths = append(invalidPaths, Entry(fmt.Sprintf("rejects %q", value), value))
	}
	DescribeTable("unsafe or ambiguous paths", func(value string) {
		_, err := parsePath(value)
		Expect(err).To(HaveOccurred())
	}, invalidPaths)

	validConditions := []TableEntry{}
	for _, value := range []string{"true", "false", "enabled=true", "name=longhorn", "name=", "name!=", `(enabled=true || count=2) && name!='hidden'`, `a["display.name"]=hello`} {
		validConditions = append(validConditions, Entry(fmt.Sprintf("accepts %q", value), value))
	}
	DescribeTable("supported visibility conditions", func(value string) {
		_, err := parseCondition(value)
		Expect(err).NotTo(HaveOccurred())
	}, validConditions)

	invalidConditions := []TableEntry{}
	for _, value := range []string{"", "x", "!x=true", "x == true", "x=true &&", "(x=true", "constructor=true", "a.0=true", "x=true; run()", "x=true & y=false"} {
		invalidConditions = append(invalidConditions, Entry(fmt.Sprintf("rejects %q", value), value))
	}
	DescribeTable("unsupported visibility conditions", func(value string) {
		_, err := parseCondition(value)
		Expect(err).To(HaveOccurred())
	}, invalidConditions)
})
