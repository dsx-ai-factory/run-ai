/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package environment_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/environment"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/environment/describe"
	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

var _ = Describe("Environment Describe Command", func() {
	Describe("Command Structure", func() {
		It("should have correct aliases", func() {
			Expect(environment.DescribeCmd.Aliases).To(ContainElement("desc"))
		})

		It("should require exactly one argument", func() {
			Expect(environment.DescribeCmd.Args).NotTo(BeNil())
		})

		It("should have correct use string", func() {
			Expect(environment.DescribeCmd.Use).To(ContainSubstring("ENVIRONMENT_NAME"))
		})
	})

	Describe("Flags", func() {
		It("should have project flag with shorthand", func() {
			flag := environment.DescribeCmd.Flags().Lookup("project")
			Expect(flag).NotTo(BeNil())
			Expect(flag.Shorthand).To(Equal("p"))
		})

		It("should have output flag with default value", func() {
			flag := environment.DescribeCmd.Flags().Lookup("output")
			Expect(flag).NotTo(BeNil())
			Expect(flag.Shorthand).To(Equal("o"))
			Expect(flag.DefValue).To(Equal("table"))
		})

		It("should allow parsing project flag", func() {
			err := environment.DescribeCmd.Flags().Parse([]string{"-p", "test-project"})
			Expect(err).To(BeNil())
		})

		It("should allow parsing output flag", func() {
			err := environment.DescribeCmd.Flags().Parse([]string{"-o", "json"})
			Expect(err).To(BeNil())
		})
	})

	Describe("Output Generation", func() {
		var testEnv *openapi.EnvironmentAsset

		BeforeEach(func() {
			// Create a test environment
			now := time.Now()
			testEnv = &openapi.EnvironmentAsset{
				Meta: openapi.AssetMeta{
					Name:        "test-env",
					Scope:       "Project",
					CreatedBy:   "admin@test.com",
					CreatedAt:   now,
					UpdatedAt:   now,
					Description: *openapi.NewNullableString(openapi.PtrString("Test environment")),
					WorkloadSupportedTypes: *openapi.NewNullableWorkloadSupportedTypes(&openapi.WorkloadSupportedTypes{
						Workspace:     *openapi.NewNullableBool(openapi.PtrBool(false)),
						Training:      *openapi.NewNullableBool(openapi.PtrBool(true)),
						Inference:     *openapi.NewNullableBool(openapi.PtrBool(false)),
						Distributed:   *openapi.NewNullableBool(openapi.PtrBool(true)),
						DistFramework: *openapi.NewNullableString(openapi.PtrString("PyTorch")),
					}),
				},
				Spec: openapi.EnvironmentAssetSpec{
					Image: *openapi.NewNullableString(openapi.PtrString("pytorch:latest")),
					ImagePullPolicy: func() openapi.NullableImagePullPolicy {
						p := openapi.IMAGEPULLPOLICY_IF_NOT_PRESENT
						return *openapi.NewNullableImagePullPolicy(&p)
					}(),
					WorkingDir: *openapi.NewNullableString(openapi.PtrString("/workspace")),
					Command:    *openapi.NewNullableString(openapi.PtrString("python")),
					Args:       *openapi.NewNullableString(openapi.PtrString("train.py")),
					EnvironmentVariables: []openapi.EnvironmentVariableOfAsset{
						{
							Name:  *openapi.NewNullableString(openapi.PtrString("CUDA_VISIBLE_DEVICES")),
							Value: *openapi.NewNullableString(openapi.PtrString("all")),
						},
					},
					Capabilities: []openapi.Capability{
						openapi.CAPABILITY_SYS_ADMIN,
						openapi.CAPABILITY_NET_ADMIN,
					},
					Connections: []openapi.Connection{
						{
							Name: "jupyter",
							InternalToolInfo: *openapi.NewNullableInternalToolInfo(&openapi.InternalToolInfo{
								ToolType: openapi.INTERNALTOOLTYPE_JUPYTER_NOTEBOOK,
							}),
						},
					},
					RunAsUid: *openapi.NewNullableInt64(openapi.PtrInt64(1000)),
					RunAsGid: *openapi.NewNullableInt64(openapi.PtrInt64(1000)),
					UidGidSource: func() openapi.NullableUidGidSource {
						u := openapi.UIDGIDSOURCE_CUSTOM
						return *openapi.NewNullableUidGidSource(&u)
					}(),
					SupplementalGroups: *openapi.NewNullableString(openapi.PtrString("100, 101")),
				},
				Status: *openapi.NewNullableAssetClusterStatusInfo(&openapi.AssetClusterStatusInfo{}),
			}
		})

		Describe("Displayer Output", func() {
			It("should create displayer without errors", func() {
				output, err := describe.NewEnvironmentFormatter(testEnv)
				Expect(err).NotTo(HaveOccurred())

				displayer := output.Displayer()
				Expect(displayer).NotTo(BeNil())
			})

			It("should render table without errors", func() {
				output, err := describe.NewEnvironmentFormatter(testEnv)
				Expect(err).NotTo(HaveOccurred())

				displayer := output.Displayer()
				err = displayer.AsTable()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should render JSON without errors", func() {
				output, err := describe.NewEnvironmentFormatter(testEnv)
				Expect(err).NotTo(HaveOccurred())

				displayer := output.Displayer()
				err = displayer.AsJSON()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should render YAML without errors", func() {
				output, err := describe.NewEnvironmentFormatter(testEnv)
				Expect(err).NotTo(HaveOccurred())

				displayer := output.Displayer()
				err = displayer.AsYAML()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
