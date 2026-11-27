/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package environment_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/environment"
)

func manyVariables(count int) map[string]string {
	variables := make(map[string]string, count)
	for i := range count {
		variables[fmt.Sprintf("K%03d", i)] = "v"
	}
	return variables
}

func entryCount(rendered string) int {
	return strings.Count(rendered, ": v")
}

var _ = Describe("Format", func() {
	const columnPadding = 10
	indent := strings.Repeat(" ", columnPadding+1)

	DescribeTable("renders environment variables for describe output",
		func(variables map[string]string, limit int, matcher types.GomegaMatcher) {
			Expect(environment.Format(variables, columnPadding, limit)).To(matcher)
		},
		Entry("nil map renders empty brackets", nil, 0, Equal("[]")),
		Entry("empty map renders empty brackets", map[string]string{}, 0, Equal("[]")),
		Entry("single variable", map[string]string{"A": "1"}, 0, Equal("[A: 1]")),
		Entry("keys sorted and padded to the longest key",
			map[string]string{"Z": "last", "AA": "first"}, 0,
			Equal("[AA: first,\n"+indent+"Z : last]")),
		Entry("no limit shows every variable without a footer", manyVariables(101), 0,
			And(
				HavePrefix("[K000: v,\n"+indent+"K001: v"),
				HaveSuffix("K100: v]"),
				WithTransform(entryCount, Equal(101)),
				Not(ContainSubstring("Displaying")),
			)),
		Entry("limit exceeded adds a displaying footer", manyVariables(101), 100,
			And(
				HavePrefix("[K000: v,\n"+indent+"K001: v"),
				ContainSubstring(indent+"K099: v"),
				HaveSuffix("]\n"+indent+"Displaying 100/101 environment variables"),
			)),
	)
})
