/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policy_compliance

import (
	"fmt"
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"
)

var _ = Describe("graph validation", func() {
	Context("complex order", func() {
		It("should order correctly", func() {
			policyFname := "defaults_from_complex_policy"

			policyYaml, err := testutils.LoadInputFile(policyFname, ".")
			Expect(err).To(BeNil(), fmt.Sprintf("expecting to load %s.yaml", policyFname))

			var policyAsMap map[string]interface{}
			err = yaml.Unmarshal(policyYaml, &policyAsMap)
			Expect(err).To(BeNil(), fmt.Sprintf("failed to parse of %s.yaml", policyFname))

			order, err := NewDefaultsFromOrder(convutils.MapMe(policyAsMap["rules"]))
			Expect(err).To(BeNil())

			Expect(order["ports"]).To(Equal([]string{
				"external",
				"container",
			}))

			Expect(len(order[""])).To(Equal(6))

			/////////////////////////////////////////////////////////////////////////////////////////////////
			//    cpuCoreLimit(5) --> gpuPortionLimit(7.5)
			//                                                     ,---> runAsUid(10)
			//    cpuCoreRequest(3) --> gpuPortionRequest(10.5) --|
			//                                                    `---> runAsGid(2)
			/////////////////////////////////////////////////////////////////////////////////////////////////
			flatOrder := make(map[string]int)

			for idx, fldName := range order[""] {
				flatOrder[fldName] = idx + 1
			}

			Expect(flatOrder["compute.cpuCoreLimit"]).To(BeNumerically("<", flatOrder["compute.gpuPortionLimit"]))
			Expect(flatOrder["compute.cpuCoreRequest"]).To(BeNumerically("<", flatOrder["compute.gpuPortionRequest"]))
			Expect(flatOrder["compute.gpuPortionRequest"]).To(BeNumerically("<", flatOrder["security.runAsUid"]))
			Expect(flatOrder["compute.gpuPortionRequest"]).To(BeNumerically("<", flatOrder["security.runAsGid"]))
		})
	})
})
