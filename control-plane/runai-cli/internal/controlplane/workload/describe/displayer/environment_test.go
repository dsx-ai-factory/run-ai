/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package displayer_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/describe/displayer"
)

func envVars(n int) map[string]string {
	env := make(map[string]string, n)
	for i := 0; i < n; i++ {
		env[fmt.Sprintf("VAR_%03d", i)] = "example"
	}
	return env
}

var _ = Describe("GetEnvironmentVariablesDisplay", func() {
	const notice = "Displaying 100/101 environment variables"

	DescribeTable("renders at most MaxDisplayEnvVarsCount sorted variables",
		func(isNil bool, count int, wantRendered int, wantExact string, wantLastKey string, wantNotice bool) {
			var envPtr *map[string]string
			if !isNil {
				env := envVars(count)
				envPtr = &env
			}

			out := displayer.GetEnvironmentVariablesDisplay(envPtr, 4)

			if wantExact != "" {
				Expect(out).To(Equal(wantExact))
			}
			Expect(strings.Count(out, "VAR_")).To(Equal(wantRendered))
			if wantLastKey != "" {
				Expect(out).To(ContainSubstring(wantLastKey + ": example]"))
			}
			if count > displayer.MaxDisplayEnvVarsCount {
				Expect(out).NotTo(ContainSubstring(fmt.Sprintf("VAR_%03d", displayer.MaxDisplayEnvVarsCount)))
			}
			if wantNotice {
				Expect(out).To(ContainSubstring(notice))
			} else {
				Expect(out).NotTo(ContainSubstring("Displaying"))
			}
			if envPtr != nil {
				Expect(*envPtr).To(HaveLen(count), "input map must not be mutated")
			}
		},
		Entry("nil map renders empty brackets", true, 0, 0, "[]", "", false),
		Entry("empty map renders empty brackets", false, 0, 0, "[]", "", false),
		Entry("single variable renders inline", false, 1, 1, "[VAR_000: example]", "VAR_000", false),
		Entry("exactly 100 variables renders all without notice", false, 100, 100, "", "VAR_099", false),
		Entry("101 variables renders first 100 with notice", false, 101, 100, "", "VAR_099", true),
	)

	It("aligns keys using only the displayed variables", func() {
		env := envVars(displayer.MaxDisplayEnvVarsCount)
		env["ZZZ_HIDDEN_LONGER_KEY"] = "example" // sorts last, so it is cut and must not widen the key column

		out := displayer.GetEnvironmentVariablesDisplay(&env, 4)

		Expect(out).To(HavePrefix("[VAR_000: example,\n     VAR_001: example,"))
	})

	DescribeTable("map wrapper keeps its own empty output",
		func(env map[string]string, want string) {
			Expect(displayer.GetEnvironmentVariablesDisplayForMap(env, 4)).To(Equal(want))
		},
		Entry("nil map renders nothing", nil, ""),
		Entry("empty map renders nothing", map[string]string{}, ""),
		Entry("single variable delegates", envVars(1), "[VAR_000: example]"),
	)
})
