/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policy_compliance

import (
	"fmt"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"
	"github.com/stretchr/testify/assert"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2" //nolint:staticcheck // dot-import is the idiomatic ginkgo usage in test helpers
	. "github.com/onsi/gomega"    //nolint:staticcheck // dot-import is the idiomatic gomega usage in test helpers
)

type ComplianceTest struct {
	workloadType string

	workloadFname string
	policyFname   string

	mutatedSpecFname string // this is for the updates spec that contains also the defaults from the policy

	options ValidateComplianceOptions

	expectCycle    bool
	expectedErrors ComplianceErrors
}

var ComplianceTests = []TableEntry{
	Entry("blocked field having value", ComplianceTest{
		workloadType:  "",
		workloadFname: "supplemental",
		policyFname:   "blocked_fields",

		expectedErrors: ComplianceErrors{
			{
				WorkloadField: "security.supplementalGroups",
				PolicyField:   "security.supplementalGroups",
				Details:       "value should not be provided for this field, it has been blocked by the administrator",
				Rule:          &PolicyBlocked,
			},
		},
	}),

	Entry("blocked field with default", ComplianceTest{
		workloadType: "",
		policyFname:  "blocked_fields",
		// when doing policy self validation the defaults as workload values
		workloadFname: "blocked_default",

		options: ValidateComplianceOptions{
			PolicySelfValidation: true,
		},

		expectedErrors: ComplianceErrors{
			{
				WorkloadField: "security.supplementalGroups",
				PolicyField:   "security.supplementalGroups",
				Details:       "default value should not be provided for a field with 'blocked' rule",
				Rule:          &PolicyBlocked,
			},
		},
	}),

	Entry("strict policy vs wrong workload", ComplianceTest{
		workloadType:  "",
		workloadFname: "strict_values_errors",
		policyFname:   "strict_policy",

		expectedErrors: ComplianceErrors{
			{
				WorkloadField: "args",
				PolicyField:   "args",
				Details:       "the administrator prohibited changing the value of this field",
				Rule:          &PolicyCanEdit,
			},
			{
				WorkloadField: "environmentVariables",
				PolicyField:   "environmentVariables.instances",
				Details:       "the administrator prohibited adding new items",
				Rule:          &PolicyCanAdd,
			},
			{
				WorkloadField: "environmentVariables[FIXED_ENV]",
				PolicyField:   "environmentVariables.instances",
				Details:       "the administrator prohibited modifying item 'FIXED_ENV'",
				Rule:          &PolicyLocked,
			},
			{
				WorkloadField: "ports",
				PolicyField:   "ports.instances",
				Details:       "the administrator prohibited adding new items",
				Rule:          &PolicyCanAdd,
			},
			{
				WorkloadField: "ports[FIXED_PORT]",
				PolicyField:   "ports.instances",
				Details:       "the administrator prohibited modifying item 'FIXED_PORT'",
				Rule:          &PolicyLocked,
			},
			{
				WorkloadField: "priorityClass",
				PolicyField:   "priorityClass",
				Details:       "the administrator prohibited changing the value of this field",
				Rule:          &PolicyCanEdit,
			},
			{
				WorkloadField: "storage.hostPath",
				PolicyField:   "storage.hostPath.instances",
				Details:       "the administrator prohibited adding new items",
				Rule:          &PolicyCanAdd,
			},
			{
				WorkloadField: "storage.hostPath[FIXED_HOST_PATH_ITEM]",
				PolicyField:   "storage.hostPath.instances",
				Details:       "the administrator prohibited modifying item 'FIXED_HOST_PATH_ITEM'",
				Rule:          &PolicyLocked,
			},
			{
				WorkloadField: "storage.pvc[ATTRIBUTE_DEFAULTS_CHANGED].claimInfo.accessMode.readWriteMany",
				PolicyField:   "storage.pvc.attributes.claimInfo.accessMode.readWriteMany",
				Details:       "the administrator prohibited changing the value of this field",
				Rule:          &PolicyCanEdit,
			},
			{
				WorkloadField: "storage.pvc[ATTRIBUTE_DEFAULTS_CHANGED].claimInfo.accessMode.readWriteOnce",
				PolicyField:   "storage.pvc.attributes.claimInfo.accessMode.readWriteOnce",
				Details:       "the administrator prohibited changing the value of this field",
				Rule:          &PolicyCanEdit,
			},
		},
	}),

	Entry("both instance and attributes errors", ComplianceTest{
		workloadType:  "",
		workloadFname: "attr_and_inst_values",
		policyFname:   "attr_and_inst_policy",

		expectedErrors: ComplianceErrors{
			{
				WorkloadField: "ports[FIXED_PORT]",
				PolicyField:   "ports.instances",
				Details:       "the administrator prohibited modifying item 'FIXED_PORT'",
				Rule:          &PolicyLocked,
			},
			{
				WorkloadField: "ports[ADDED_PORT].container",
				PolicyField:   "ports.attributes.container",
				Details:       "must be no greater than 9000",
				Rule:          &PolicyMax,
			},
			{
				WorkloadField: "ports[CAN_CHANGE_PORT].container",
				PolicyField:   "ports.attributes.container",
				Details:       "must be no greater than 9000",
				Rule:          &PolicyMax,
			},
			{
				WorkloadField: "ports[FIXED_PORT].container",
				PolicyField:   "ports.attributes.container",
				Details:       "must be no greater than 9000",
				Rule:          &PolicyMax,
			},
		},
	}),

	Entry("defaults_from_w1", ComplianceTest{
		workloadFname: "defaults_from_w1",
		policyFname:   "defaults_from_policy",

		expectedErrors: ComplianceErrors{
			{
				WorkloadField: "compute.cpuCoreRequest",
				PolicyField:   "compute.cpuCoreRequest",
				Details:       "must be no greater than 5",
				Rule:          &PolicyMax,
			},
		},
	}),

	Entry("defaults_from_w2", ComplianceTest{
		workloadFname: "defaults_from_w2",
		policyFname:   "defaults_from_policy",

		mutatedSpecFname: "defaults_from_w2",

		expectedErrors: ComplianceErrors{},
	}),

	Entry("defaults_from_complex", ComplianceTest{
		workloadFname: "defaults_from_two_fields",
		policyFname:   "defaults_from_complex_policy",

		mutatedSpecFname: "defaults_from_complex",

		expectedErrors: ComplianceErrors{},
	}),

	Entry("defaults_from_conversion", ComplianceTest{
		policyFname: "default_from_conversion",

		mutatedSpecFname: "defaults_from_conversion",

		expectedErrors: ComplianceErrors{},
	}),

	Entry("defaults_from_cycle", ComplianceTest{
		workloadFname: "defaults_from_two_fields",
		policyFname:   "defaults_from_cycle_policy",

		expectCycle:    true,
		expectedErrors: ComplianceErrors{},
	}),

	Entry("cant_edit_no_default", ComplianceTest{
		policyFname: "cant_edit_no_default",

		options: ValidateComplianceOptions{
			PolicySelfValidation:           true,
			CanEditWithNoDefaultValidation: true,
		},

		expectedErrors: ComplianceErrors{
			{
				WorkloadField: "image",
				PolicyField:   "image",
				Details:       "default value must be provided for policy fields having canEdit=false",
				Rule:          &PolicyCanEdit,
			},
			{
				PolicyField: "compute.extendedResources.attributes.quantity",
				Details:     "default value must be provided for policy fields having canEdit=false",
				Rule:        &PolicyCanEdit,
			},
		},
	}),

	Entry("cant_edit_with_default", ComplianceTest{
		policyFname: "cant_edit_no_default",

		// when doing policy self validation the defaults as workload values
		workloadFname: "cant_edit_with_default",

		mutatedSpecFname: "cant_edit_with_default",

		options: ValidateComplianceOptions{
			PolicySelfValidation:           true,
			CanEditWithNoDefaultValidation: true,
		},

		expectedErrors: ComplianceErrors{},
	}),

	Entry("locked unchanged", ComplianceTest{
		policyFname: "locked_policy",

		workloadFname: "locked_workload_ok",

		mutatedSpecFname: "locked_workload_ok",
	}),

	Entry("locked values changed", ComplianceTest{
		policyFname: "locked_policy",

		workloadFname: "locked_workload_not_ok",

		mutatedSpecFname: "locked_workload_not_ok",

		expectedErrors: ComplianceErrors{
			{
				WorkloadField: "annotations[a1]",
				PolicyField:   "annotations.instances",
				Details:       "the administrator prohibited modifying item 'a1'",
				Rule:          &PolicyLocked,
			},
			{
				WorkloadField: "storage.s3[s2]",
				PolicyField:   "storage.s3.instances",
				Details:       "the administrator prohibited modifying item 's2'",
				Rule:          &PolicyLocked,
			},
		},
	}),
}

func RunComplianceTest(testDetails ComplianceTest, basePath string) {
	testBrief := fmt.Sprintf("policy: %s, workload: %s", testDetails.policyFname, testDetails.workloadFname)

	workloadAsMap := map[string]interface{}{}
	policyAsMap := map[string]interface{}{}

	if testDetails.workloadFname != testutils.NA {
		workloadYaml, err := testutils.LoadInputFile(testDetails.workloadFname, basePath)
		Expect(err).To(BeNil(), fmt.Sprintf("failed to load of %s.yaml", testDetails.workloadFname))

		err = yaml.Unmarshal(workloadYaml, &workloadAsMap)
		Expect(err).To(BeNil(), fmt.Sprintf("failed to parse of %s.yaml", testDetails.workloadFname))
	}

	var order DefaultFromOrder

	if testDetails.policyFname != testutils.NA {
		policyYaml, err := testutils.LoadInputFile(testDetails.policyFname, basePath)
		Expect(err).To(BeNil(), fmt.Sprintf("expecting to load %s.yaml", testDetails.policyFname))

		err = yaml.Unmarshal(policyYaml, &policyAsMap)
		Expect(err).To(BeNil(), fmt.Sprintf("failed to parse of %s.yaml", testDetails.policyFname))

		order, err = NewDefaultsFromOrder(convutils.MapMe(policyAsMap["rules"]))
		if testDetails.expectCycle {
			Expect(err).NotTo(BeNil())
			return
		} else {
			Expect(err).To(BeNil())
		}
	}

	fieldsWithCanEdit := make([]string, 0)
	var actualErrors ComplianceErrors
	ValidateCompliance("", convutils.MapMe(policyAsMap["rules"]), convutils.MapMe(policyAsMap["defaults"]),
		workloadAsMap, order, &fieldsWithCanEdit, &testDetails.options, &actualErrors)

	actualArr := actualErrors.AsSortedArray()
	expectedArr := testDetails.expectedErrors.AsSortedArray()

	if len(expectedArr) != 0 {
		Expect(actualArr).To(Equal(expectedArr), testBrief)
		return
	}

	Expect(len(actualErrors)).To(Equal(0), testBrief)

	actualYaml, _ := yaml.Marshal(workloadAsMap)
	expectedYaml := testutils.LoadExpectedFile(testDetails.mutatedSpecFname, actualYaml, basePath)
	assert.Equal(GinkgoT(), string(expectedYaml), string(actualYaml))
}
