/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package plugins

import (
	orgUnitApi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GVK Plugins", func() {

	Describe("NoOpPlugin", func() {
		var (
			plugin   *NoOpPlugin
			manifest map[string]any
			project  *orgUnitApi.Project
		)

		BeforeEach(func() {
			plugin = &NoOpPlugin{}
			manifest = map[string]any{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]any{
					"name": "test-pod",
				},
			}
			project = &orgUnitApi.Project{
				Name: "test-project",
			}
		})

		It("should not modify the manifest", func() {
			err := plugin.ProcessManifest(manifest, project)
			Expect(err).ToNot(HaveOccurred())

			// Verify manifest is unchanged
			Expect(manifest["apiVersion"]).To(Equal("v1"))
			Expect(manifest["kind"]).To(Equal("Pod"))
		})
	})

	Describe("DynamoGraphDeploymentPlugin", func() {
		var (
			plugin   *DynamoGraphDeploymentPlugin
			manifest map[string]any
			project  *orgUnitApi.Project
		)

		BeforeEach(func() {
			plugin = &DynamoGraphDeploymentPlugin{}
		})

		Context("when manifest has no annotations", func() {
			BeforeEach(func() {
				manifest = map[string]any{
					"apiVersion": "run.ai/v1",
					"kind":       DynamoGraphDeploymentKind,
					"metadata": map[string]any{
						"name": "test-dynamo",
					},
				}
				project = &orgUnitApi.Project{
					Name: "my-project",
				}
			})

			It("should add queue annotation with project name only", func() {
				err := plugin.ProcessManifest(manifest, project)
				Expect(err).ToNot(HaveOccurred())

				manifestMetadata := manifest["metadata"].(map[string]any)
				annotations := manifestMetadata["annotations"].(map[string]any)
				Expect(annotations[KubeAnnotationKaiSchedulerQueue]).To(Equal("my-project"))
			})
		})

		Context("when manifest has existing annotations", func() {
			BeforeEach(func() {
				manifest = map[string]any{
					"apiVersion": "run.ai/v1",
					"kind":       DynamoGraphDeploymentKind,
					"metadata": map[string]any{
						"name": "test-dynamo",
						"annotations": map[string]any{
							"existing-key": "existing-value",
						},
					},
				}
				project = &orgUnitApi.Project{
					Name: "my-project",
				}
			})

			It("should add queue annotation alongside existing annotations", func() {
				err := plugin.ProcessManifest(manifest, project)
				Expect(err).ToNot(HaveOccurred())

				manifestMetadata := manifest["metadata"].(map[string]any)
				annotations := manifestMetadata["annotations"].(map[string]any)
				Expect(annotations[KubeAnnotationKaiSchedulerQueue]).To(Equal("my-project"))
				Expect(annotations["existing-key"]).To(Equal("existing-value"))
			})
		})

		Context("when queue annotation already exists", func() {
			BeforeEach(func() {
				manifest = map[string]any{
					"apiVersion": "run.ai/v1",
					"kind":       DynamoGraphDeploymentKind,
					"metadata": map[string]any{
						"name": "test-dynamo",
						"annotations": map[string]any{
							KubeAnnotationKaiSchedulerQueue: "existing-queue",
						},
					},
				}
				project = &orgUnitApi.Project{
					Name: "my-project",
				}
			})

			It("should not overwrite existing queue annotation", func() {
				err := plugin.ProcessManifest(manifest, project)
				Expect(err).ToNot(HaveOccurred())

				manifestMetadata := manifest["metadata"].(map[string]any)
				annotations := manifestMetadata["annotations"].(map[string]any)
				Expect(annotations[KubeAnnotationKaiSchedulerQueue]).To(Equal("existing-queue"))
			})
		})

		Context("when project has multiple default node pools", func() {
			BeforeEach(func() {
				manifest = map[string]any{
					"apiVersion": "run.ai/v1",
					"kind":       DynamoGraphDeploymentKind,
					"metadata": map[string]any{
						"name": "test-dynamo",
					},
				}
				project = &orgUnitApi.Project{
					Name: "my-project",
				}
				project.SetDefaultNodePools([]string{"gpu-pool-a100", "gpu-pool-v100", "cpu-pool"})
			})

			It("should add queue annotation with project name and first node pool", func() {
				project.SetDefaultNodePools([]string{DefaultNodepoolName, "gpu-pool-v100", "cpu-pool"})
				err := plugin.ProcessManifest(manifest, project)
				Expect(err).ToNot(HaveOccurred())

				manifestMetadata := manifest["metadata"].(map[string]any)
				annotations := manifestMetadata["annotations"].(map[string]any)
				Expect(annotations[KubeAnnotationKaiSchedulerQueue]).To(Equal("my-project"))
			})

			It("should add queue annotation with project name and first node pool", func() {
				err := plugin.ProcessManifest(manifest, project)
				Expect(err).ToNot(HaveOccurred())

				manifestMetadata := manifest["metadata"].(map[string]any)
				annotations := manifestMetadata["annotations"].(map[string]any)
				Expect(annotations[KubeAnnotationKaiSchedulerQueue]).To(Equal("my-project-gpu-pool-a100"))
			})
		})

		Context("when project has empty node pools list", func() {
			BeforeEach(func() {
				manifest = map[string]any{
					"apiVersion": "run.ai/v1",
					"kind":       DynamoGraphDeploymentKind,
					"metadata": map[string]any{
						"name": "test-dynamo",
					},
				}
				project = &orgUnitApi.Project{
					Name: "my-project",
				}
				project.SetDefaultNodePools([]string{})
			})

			It("should add queue annotation with just project name", func() {
				err := plugin.ProcessManifest(manifest, project)
				Expect(err).ToNot(HaveOccurred())

				manifestMetadata := manifest["metadata"].(map[string]any)
				annotations := manifestMetadata["annotations"].(map[string]any)
				Expect(annotations[KubeAnnotationKaiSchedulerQueue]).To(Equal("my-project"))
			})
		})

		Context("when manifest has invalid metadata", func() {
			BeforeEach(func() {
				manifest = map[string]any{
					"apiVersion": "run.ai/v1",
					"kind":       DynamoGraphDeploymentKind,
					"metadata":   "invalid-metadata", // Should be a map
				}
				project = &orgUnitApi.Project{
					Name: "my-project",
				}
			})

			It("should return an error", func() {
				err := plugin.ProcessManifest(manifest, project)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("manifest metadata is missing or not a valid map"))
			})
		})
	})
})
