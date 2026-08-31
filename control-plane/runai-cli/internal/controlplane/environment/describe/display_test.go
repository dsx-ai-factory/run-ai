/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/environment/describe"
	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

var _ = Describe("EnvironmentFormatter", func() {
	Context("NewEnvironmentFormatter", func() {
		It("should create formatter successfully with valid environment", func() {
			env := &openapi.EnvironmentAsset{
				Meta: openapi.AssetMeta{
					Name: "test-env",
				},
			}

			formatter, err := describe.NewEnvironmentFormatter(env)

			Expect(err).ToNot(HaveOccurred())
			Expect(formatter).ToNot(BeNil())
		})

		It("should return error for nil environment", func() {
			formatter, err := describe.NewEnvironmentFormatter(nil)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("environment is nil"))
			Expect(formatter).To(BeNil())
		})
	})

	Context("Displayer", func() {
		It("should return valid displayer", func() {
			env := &openapi.EnvironmentAsset{
				Meta: openapi.AssetMeta{
					Name: "test-env",
				},
			}
			formatter, _ := describe.NewEnvironmentFormatter(env)

			displayer := formatter.Displayer()

			Expect(displayer).ToNot(BeNil())
		})
	})
})

var _ = Describe("EnvironmentDisplayer", func() {
	var (
		env       *openapi.EnvironmentAsset
		displayer *describe.EnvironmentDisplayer
	)

	BeforeEach(func() {
		env = newTestEnvironmentAsset()
		displayer = describe.NewEnvironmentDisplayer(env)
	})

	Context("NewEnvironmentDisplayer", func() {
		It("should create displayer successfully", func() {
			Expect(displayer).ToNot(BeNil())
		})
	})

	Context("AsTable", func() {
		It("should not return error", func() {
			err := displayer.AsTable()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("AsJSON", func() {
		It("should generate valid JSON", func() {
			err := displayer.AsJSON()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("AsYAML", func() {
		It("should generate valid YAML", func() {
			err := displayer.AsYAML()
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

var _ = Describe("Helper functions", func() {
})

// Helper function to create test environment asset for describe tests
func newTestEnvironmentAsset() *openapi.EnvironmentAsset {
	// Create connections with tool
	conn := openapi.Connection{}
	conn.SetInternalToolInfo(openapi.InternalToolInfo{
		ToolType: openapi.InternalToolType("JupyterNotebook"),
	})

	// Create spec with all fields
	spec := openapi.EnvironmentAssetSpec{}
	spec.SetImage("ubuntu:20.04")
	spec.SetConnections([]openapi.Connection{conn})
	spec.SetWorkingDir("/workspace")
	spec.SetCommand("/bin/bash")
	spec.SetArgs("-c echo hello")

	policy := openapi.ImagePullPolicy("Always")
	spec.ImagePullPolicy = *openapi.NewNullableImagePullPolicy(&policy)

	envVar := openapi.EnvironmentVariableOfAsset{}
	envVar.SetName("TEST_VAR")
	envVar.SetValue("test-value")
	spec.EnvironmentVariables = []openapi.EnvironmentVariableOfAsset{envVar}

	uidGidSource := openapi.UidGidSource("FromTheImage")
	spec.UidGidSource = *openapi.NewNullableUidGidSource(&uidGidSource)
	spec.RunAsUid = *openapi.NewNullableInt64(openapi.PtrInt64(1000))
	spec.RunAsGid = *openapi.NewNullableInt64(openapi.PtrInt64(1000))
	spec.SupplementalGroups = *openapi.NewNullableString(openapi.PtrString("100,101"))
	spec.Capabilities = []openapi.Capability{
		openapi.Capability("SYS_ADMIN"),
		openapi.Capability("NET_ADMIN"),
	}

	// Create workload types
	workloadTypes := openapi.WorkloadSupportedTypes{}
	trueVal := true
	workloadTypes.Workspace = *openapi.NewNullableBool(&trueVal)
	workloadTypes.Training = *openapi.NewNullableBool(&trueVal)

	return &openapi.EnvironmentAsset{
		Meta: openapi.AssetMeta{
			Name:                   "test-env",
			Scope:                  openapi.Scope("Cluster"),
			CreatedBy:              "test-user",
			CreatedAt:              time.Now(),
			UpdatedAt:              time.Now(),
			WorkloadSupportedTypes: *openapi.NewNullableWorkloadSupportedTypes(&workloadTypes),
		},
		Spec: spec,
	}
}
