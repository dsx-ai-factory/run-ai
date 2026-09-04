/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package translation

import (
	"context"
	"fmt"
	"os"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils"
	"github.com/stretchr/testify/assert"
)

type TranslationTest struct {
	nativeSpecFname string
	workloadType    workloadsApi.WorkloadType

	expectedError  string
	expectedFname  string
	expectedIssues []string
}

var TranslationTests = []TableEntry{
	Entry("basic workspace", TranslationTest{
		workloadType:    workloadsApi.WORKLOADTYPE_WORKSPACE,
		nativeSpecFname: "basic_ws",
	}),

	Entry("extended resources", TranslationTest{
		workloadType:    workloadsApi.WORKLOADTYPE_WORKSPACE,
		nativeSpecFname: "ext_resrc_ws",
	}),

	Entry("git storage", TranslationTest{
		workloadType:    workloadsApi.WORKLOADTYPE_WORKSPACE,
		nativeSpecFname: "git_ws",
	}),

	Entry("basic distributed", TranslationTest{
		workloadType:    workloadsApi.WORKLOADTYPE_DISTRIBUTED,
		nativeSpecFname: "basic_dist",
	}),
}

var _ = Describe("translation tests", func() {
	Context("input VS expected tests", func() {
		DescribeTable("should convert policies correctly", func(testDetails TranslationTest) {
			RunTranslationTest(testDetails, ".")
		}, TranslationTests)
	})
})

type PolicyMeta struct {
	Meta api.PolicyCreationFields
}

func RunTranslationTest(testDetails TranslationTest, basePath string) {
	testDetailsStr := fmt.Sprintf("input: %s, expected: %s", testDetails.nativeSpecFname, testDetails.expectedFname)

	// Load input file
	inputYaml, err := testutils.LoadInputFile(testDetails.nativeSpecFname, basePath)
	Expect(err).To(BeNil(), fmt.Sprintf("failed to load input file %s.yaml", testDetails.nativeSpecFname))

	var policyMeta PolicyMeta
	err = yaml.Unmarshal(inputYaml, &policyMeta)
	Expect(err).To(BeNil(), fmt.Sprintf("failed to load metadata from input file %s.yaml", testDetails.nativeSpecFname))

	var policyAsMap map[string]any
	err = yaml.Unmarshal(inputYaml, &policyAsMap)
	Expect(err).To(BeNil(), fmt.Sprintf("failed to load spec from input file %s.yaml", testDetails.nativeSpecFname))

	// Run translation
	translatedPolicy, issues, translateErr := TranslatePolicy(context.Background(), string(testDetails.workloadType), policyMeta.Meta, utils.MapMe(policyAsMap["policy"]))

	// Check for expected error
	if testDetails.expectedError != "" {
		Expect(translateErr).NotTo(BeNil(), fmt.Sprintf("expected error was not returned: %s", testDetailsStr))
		if translateErr != nil {
			Expect(translateErr.Error()).To(ContainSubstring(testDetails.expectedError), testDetailsStr)
		}
		return
	}

	Expect(translateErr).To(BeNil(), testDetailsStr)

	// Compare with expected file
	if testDetails.expectedFname == "" {
		testDetails.expectedFname = testDetails.nativeSpecFname
	}

	compareTranslationFiles(testDetailsStr, translatedPolicy, testDetails.expectedFname, basePath)

	// Compare issues
	if testDetails.expectedIssues == nil {
		testDetails.expectedIssues = []string{}
	}
	Expect(issues).To(Equal(testDetails.expectedIssues), fmt.Sprintf("issues mismatch: %s", testDetailsStr))
}

func compareTranslationFiles(testDetailsStr string, actualResults *api.Policy, expectedFname string, basePath string) {
	actualYaml, marshalErr := yaml.Marshal(actualResults)
	Expect(marshalErr).To(BeNil(), testDetailsStr)

	expectedYaml := testutils.LoadExpectedFile(expectedFname, actualYaml, basePath)

	isOk := assert.Equal(GinkgoT(), string(expectedYaml), string(actualYaml), testDetailsStr)
	if !isOk {
		tmpPath := fmt.Sprintf("%s/%s.yaml", os.TempDir(), expectedFname)
		osErr := os.WriteFile(tmpPath, actualYaml, 0600)
		if osErr != nil {
			assert.Fail(GinkgoT(), fmt.Sprintf("failed to backup result to %s: %v", tmpPath, osErr.Error()))
		} else {
			assert.Fail(GinkgoT(), fmt.Sprintf("actual result backed up to %s", tmpPath))
		}
	}
}
