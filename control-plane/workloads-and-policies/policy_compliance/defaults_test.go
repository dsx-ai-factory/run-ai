/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policy_compliance

import (
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("defaults patching validations", func() {

	BeforeEach(func() {
	})

	Context("defaults itemized patching", func() {
		It("should add missing names to instances", func() {
			specYaml, err := testutils.LoadInputFile("defaults_rich", ".")
			Expect(err).To(BeNil())
			Expect(specYaml).NotTo(BeNil())

			defaults := make(map[string]interface{})
			marshalError := yaml.Unmarshal(specYaml, &defaults)
			Expect(marshalError).To(BeNil())

			attrDefaults := PrepareDefaultsForMutation(defaults)
			expectedAttrDefaults := map[PerReplicaField]interface{}{
				PerReplicaField{openapi.REPLICATYPE_WORKER, "storage.s3"}: map[string]interface{}{
					"bucket": "/default/bucket",
				},
			}
			Expect(attrDefaults).To(Equal(expectedAttrDefaults))

			actualYaml, _ := yaml.Marshal(defaults)

			expectedYaml := testutils.LoadExpectedFile("defaults_rich", actualYaml, ".")

			assert.Equal(GinkgoT(), string(expectedYaml), string(actualYaml))
		})
	})

	Context("defaults itemized patching for distributed", func() {
		It("should add missing names to instances", func() {
			specYaml, err := testutils.LoadInputFile("defaults_dist", ".")
			Expect(err).To(BeNil())
			Expect(specYaml).NotTo(BeNil())

			defaults := make(map[string]interface{})
			marshalError := yaml.Unmarshal(specYaml, &defaults)
			Expect(marshalError).To(BeNil())

			attrDefaults := PrepareDefaultsForMutation(defaults)
			expectedAttrDefaults := map[PerReplicaField]interface{}{
				PerReplicaField{openapi.REPLICATYPE_WORKER, "storage.s3"}: map[string]interface{}{
					"bucket": "/default/worker/bucket",
				},
				PerReplicaField{openapi.REPLICATYPE_MASTER, "storage.s3"}: map[string]interface{}{
					"bucket": "/default/master/bucket",
				},
			}
			Expect(attrDefaults).To(Equal(expectedAttrDefaults))

			actualYaml, _ := yaml.Marshal(defaults)

			expectedYaml := testutils.LoadExpectedFile("defaults_dist", actualYaml, ".")

			assert.Equal(GinkgoT(), string(expectedYaml), string(actualYaml))
		})
	})

	Context("defaults apply to workload", func() {
		It("instances must be correct", func() {
			defaultsYaml, err := testutils.LoadInputFile("defaults_rich", ".")
			Expect(err).To(BeNil())

			defaults := make(map[string]interface{})
			marshalError := yaml.Unmarshal(defaultsYaml, &defaults)
			Expect(marshalError).To(BeNil())

			workloadYaml, err := testutils.LoadInputFile("defaults_rich_w1", ".")
			Expect(err).To(BeNil())

			workload := make(map[string]interface{})
			marshalError = yaml.Unmarshal(workloadYaml, &workload)
			Expect(marshalError).To(BeNil())

			attributesDefaults := PrepareDefaultsForMutation(defaults)

			errApply := ApplyDefaults(workload, defaults, attributesDefaults)
			Expect(errApply).To(BeNil())

			actualYaml, _ := yaml.Marshal(defaults)

			expectedYaml := testutils.LoadExpectedFile("defaults_rich_w1", actualYaml, ".")

			assert.Equal(GinkgoT(), string(expectedYaml), string(actualYaml))
		})
	})

	Context("defaults apply to distributed workload", func() {
		It("instances must be correct", func() {
			defaultsYaml, err := testutils.LoadInputFile("defaults_dist", ".")
			Expect(err).To(BeNil())

			defaults := make(map[string]interface{})
			marshalError := yaml.Unmarshal(defaultsYaml, &defaults)
			Expect(marshalError).To(BeNil())

			workloadYaml, err := testutils.LoadInputFile("defaults_dist_w1", ".")
			Expect(err).To(BeNil())

			workload := make(map[string]interface{})
			marshalError = yaml.Unmarshal(workloadYaml, &workload)
			Expect(marshalError).To(BeNil())

			attributesDefaults := PrepareDefaultsForMutation(defaults)

			errApply := ApplyDefaults(workload, defaults, attributesDefaults)
			Expect(errApply).To(BeNil())

			actualYaml, _ := yaml.Marshal(defaults)

			expectedYaml := testutils.LoadExpectedFile("defaults_dist_w1", actualYaml, ".")

			assert.Equal(GinkgoT(), string(expectedYaml), string(actualYaml))
		})
	})

})
