/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package translation

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("preprocessDefaults", func() {
	It("wraps flat values, recurses nested sections, and transforms itemized fields", func() {
		// This is the exact example documented in the preprocessDefaults comment.
		const inputYAML = `
image: nginx
compute:
  gpuDevicesRequest: 1
environmentVariables:
  attributes:
    value: some-default
  instances:
    - name: FOO
      value: bar
`
		const expectedYAML = `
image:
  value: nginx
compute:
  gpuDevicesRequest:
    value: 1
environmentVariables:
  attributes:
    value:
      value: some-default
  instances-values:
    - name: FOO
      value: bar
`

		var input map[string]any
		Expect(yaml.Unmarshal([]byte(inputYAML), &input)).To(Succeed())

		actual := preprocessDefaults(input, "")

		// Normalize both sides through the same marshaller so key ordering and
		// scalar types don't cause spurious diffs.
		actualYAML, err := yaml.Marshal(actual)
		Expect(err).NotTo(HaveOccurred())

		var expected map[string]any
		Expect(yaml.Unmarshal([]byte(expectedYAML), &expected)).To(Succeed())
		expectedYAMLNormalized, err := yaml.Marshal(expected)
		Expect(err).NotTo(HaveOccurred())

		Expect(string(actualYAML)).To(Equal(string(expectedYAMLNormalized)))
	})
})
