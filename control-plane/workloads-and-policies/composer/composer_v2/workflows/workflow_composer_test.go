/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workflows

import (
	wfv1 "github.com/argoproj/argo-workflows/v4/pkg/apis/workflow/v1alpha1"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type composerTest struct {
	specFname     string
	expectedFname string
}

var _ = Describe("workflow composer", func() {
	DescribeTable("input VS expected tests", func(testDetails composerTest) {
		specYaml, err := testutils.LoadInputFile(testDetails.specFname, ".")
		Expect(err).To(BeNil())

		var spec workloadsApi.WorkflowSpecSpec
		Expect(yaml.Unmarshal(specYaml, &spec)).To(Succeed())

		target := &wfv1.Workflow{}
		composeErr := NewWorkflowComposer("Workflow", testDetails.specFname, "test", "runai-test", "tester").Compose(&spec, target)()
		Expect(composeErr).To(BeNil())

		actualYaml, marshalErr := yaml.Marshal(target)
		Expect(marshalErr).To(BeNil())

		expectedYaml := testutils.LoadExpectedFile(testDetails.expectedFname, actualYaml, ".")
		Expect(string(actualYaml)).To(Equal(string(expectedYaml)))
	},
		Entry("basic workflow", composerTest{specFname: "basic", expectedFname: "basic"}),
	)
})
