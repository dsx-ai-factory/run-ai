/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"

	"github.com/google/uuid"
	openapi2 "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/topology"
	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/clients/clusters"
	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/plugins"
	"github.com/run-ai/runai/runai-application-packages/convutils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openapi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
)

var _ = Describe("K8sResourceUtil", func() {
	var (
		util                    *K8sResourceUtil
		ctx                     context.Context
		validPodMap             map[string]interface{}
		validPodMapWithTopology map[string]interface{}
		invalidPodMap           map[string]interface{}
		csclient                *clusterservice.SimpleMock
	)

	BeforeEach(func() {
		util = NewK8sResourceUtil()
		ctx = setContext("1", "user-a", []string{})

		validPodMap = map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      "test-pod",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"name":  "test-container",
						"image": "nginx",
					},
				},
			},
		}

		validPodMapWithTopology = map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      "test-pod",
				"namespace": "default",
				"annotations": map[string]interface{}{
					"kai.scheduler/topology": "keep-it",
				},
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"name":  "test-container",
						"image": "nginx",
					},
				},
			},
		}

		invalidPodMap = map[string]interface{}{
			"kind": "Pod",
			"metadata": map[string]interface{}{
				"name": "test-pod",
			},
		}

		csclient = clusterservice.NewSimpleMock()
		csclient.WithDefaultNodePool("test-cluster")
		clusters.SetMock(csclient)
	})

	Describe("ValidateK8sResource", func() {
		It("should validate a valid Kubernetes resource", func() {
			err := util.ValidateK8sResource(validPodMap)
			Expect(err).To(BeNil())
		})

		It("should reject nil resource", func() {
			err := util.ValidateK8sResource(nil)
			Expect(err).ToNot(BeNil())
			Expect(err.Message).To(ContainSubstring("resource cannot be nil"))
		})

		It("should reject resource without apiVersion", func() {
			err := util.ValidateK8sResource(invalidPodMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Message).To(ContainSubstring("missing required field: apiVersion"))
		})

		It("should reject resource without kind", func() {
			noKindMap := map[string]interface{}{
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"name": "test-pod",
				},
			}
			err := util.ValidateK8sResource(noKindMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Message).To(ContainSubstring("missing required field: kind"))
		})

		It("should accept resource with generateName", func() {
			generateNameMap := map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]interface{}{
					"generateName": "test-pod-",
				},
			}
			err := util.ValidateK8sResource(generateNameMap)
			Expect(err).To(BeNil())
		})
	})

	Describe("IsValidK8sResource", func() {
		It("should return true for valid resource", func() {
			isValid := util.IsValidK8sResource(validPodMap)
			Expect(isValid).To(BeTrue())
		})

		It("should return false for invalid resource", func() {
			isValid := util.IsValidK8sResource(invalidPodMap)
			Expect(isValid).To(BeFalse())
		})
	})

	Describe("UpdateResourceMetadata", func() {
		var (
			workloadId string
			projectId  string
			project    *openapi.Project
		)

		BeforeEach(func() {
			workloadId = "workload-123"
			projectId = "project-456"
			project = &openapi.Project{
				Id:               projectId,
				Name:             "test-project",
				ClusterId:        "test-cluster",
				DefaultNodePools: []string{"default"},
			}
			project.SetRequestedNamespace("test-namespace")
		})

		It("should add workload ID to resource metadata", func() {
			modifiedResource, err := util.UpdateResourceMetadata(ctx, validPodMap, workloadId, project)
			Expect(err).To(BeNil())

			metadata := modifiedResource["metadata"].(map[string]interface{})
			labels := metadata["labels"].(map[string]interface{})
			annotations := metadata["annotations"].(map[string]interface{})

			Expect(labels[WorkloadLabel]).To(Equal("true"))
			Expect(labels[ProjectIdLabel]).To(Equal(projectId))
			Expect(annotations[WorkloadIdAnnotation]).To(Equal(workloadId))
			Expect(metadata["namespace"]).To(Equal("test-namespace"))
		})

		It("should use project name as namespace when requestedNamespace is not set", func() {
			projectWithoutNamespace := &openapi.Project{
				Id:               "project-123",
				Name:             "test-project",
				ClusterId:        "test-cluster",
				DefaultNodePools: []string{"default"},
			}

			modifiedResource, err := util.UpdateResourceMetadata(ctx, validPodMap, workloadId, projectWithoutNamespace)
			Expect(err).To(BeNil())

			metadata := modifiedResource["metadata"].(map[string]interface{})
			Expect(metadata["namespace"]).To(Equal("runai-test-project"))
		})

		It("should not modify the original resource", func() {
			originalNamespace := validPodMap["metadata"].(map[string]interface{})["namespace"]

			_, err := util.UpdateResourceMetadata(ctx, validPodMap, workloadId, project)
			Expect(err).To(BeNil())

			// Original should be unchanged
			Expect(validPodMap["metadata"].(map[string]interface{})["namespace"]).To(Equal(originalNamespace))

			// Original should not have workload labels
			metadata := validPodMap["metadata"].(map[string]interface{})
			labels, hasLabels := metadata["labels"]
			if hasLabels {
				labelsMap := labels.(map[string]interface{})
				Expect(labelsMap[WorkloadLabel]).To(Equal("true"))
			}
		})

		Context("node pool with topology", func() {
			BeforeEach(func() {
				project.DefaultNodePools = []string{"with-topology"}
				csclient.WithNodePoolsWithTopology("test-cluster", "test-topology", "with-topology")
				csclient.WithTopology("test-topology", &openapi2.NetworkTopology{Levels: []string{"level0", "level1"}})
			})

			It("should add nodepool topology to resource metadata", func() {
				modifiedResource, err := util.UpdateResourceMetadata(ctx, validPodMap, workloadId, project)
				Expect(err).To(BeNil())

				metadata := modifiedResource["metadata"].(map[string]interface{})
				labels := metadata["labels"].(map[string]interface{})
				annotations := metadata["annotations"].(map[string]interface{})

				Expect(labels[WorkloadLabel]).To(Equal("true"))
				Expect(labels[ProjectIdLabel]).To(Equal(projectId))
				Expect(annotations[WorkloadIdAnnotation]).To(Equal(workloadId))
				Expect(annotations[topology.KAITopologyAnnotationKey]).To(Equal("test-topology"))
				Expect(annotations[topology.KAITopologyAnnotationPreferredKey]).To(Equal("level1"))
				Expect(metadata["namespace"]).To(Equal("test-namespace"))
			})

			It("should keep resource topology annotations", func() {
				modifiedResource, err := util.UpdateResourceMetadata(ctx, validPodMapWithTopology, workloadId, project)
				Expect(err).To(BeNil())

				metadata := modifiedResource["metadata"].(map[string]interface{})
				labels := metadata["labels"].(map[string]interface{})
				annotations := metadata["annotations"].(map[string]interface{})

				Expect(labels[WorkloadLabel]).To(Equal("true"))
				Expect(labels[ProjectIdLabel]).To(Equal(projectId))
				Expect(annotations[WorkloadIdAnnotation]).To(Equal(workloadId))
				Expect(annotations[topology.KAITopologyAnnotationKey]).To(Equal("keep-it"))
				Expect(annotations).NotTo(HaveKey(topology.KAITopologyAnnotationPreferredKey))
				Expect(annotations).NotTo(HaveKey(topology.KAITopologyAnnotationRequiredKey))
				Expect(metadata["namespace"]).To(Equal("test-namespace"))
			})
		})
	})

	Describe("GetGVK", func() {
		It("should extract GVK from core API resource", func() {
			resource := map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]interface{}{
					"name": "test-pod",
				},
			}
			gvk, err := util.GetGVK(resource)
			Expect(err).To(BeNil())
			Expect(convutils.Deref(gvk.Group, "")).To(Equal(""))
			Expect(gvk.Version).To(Equal("v1"))
			Expect(gvk.Kind).To(Equal("Pod"))
		})

		It("should extract GVK from group API resource", func() {
			resource := map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "test-deployment",
				},
			}
			gvk, err := util.GetGVK(resource)
			Expect(err).To(BeNil())
			Expect(convutils.Deref(gvk.Group, "")).To(Equal("apps"))
			Expect(gvk.Version).To(Equal("v1"))
			Expect(gvk.Kind).To(Equal("Deployment"))
		})

		It("should extract GVK from custom resource", func() {
			resource := map[string]interface{}{
				"apiVersion": "run.ai/v1",
				"kind":       plugins.DynamoGraphDeploymentKind,
				"metadata": map[string]interface{}{
					"name": "test-dynamo",
				},
			}
			gvk, err := util.GetGVK(resource)
			Expect(err).To(BeNil())
			Expect(convutils.Deref(gvk.Group, "")).To(Equal("run.ai"))
			Expect(gvk.Version).To(Equal("v1"))
			Expect(gvk.Kind).To(Equal(plugins.DynamoGraphDeploymentKind))
		})

		It("should return error for nil resource", func() {
			_, err := util.GetGVK(nil)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("resource cannot be nil"))
		})
	})

	Describe("GetResourceInfo", func() {
		It("should extract resource information", func() {
			info, err := util.GetResourceInfo(validPodMap)
			Expect(err).To(BeNil())
			Expect(info.APIVersion).To(Equal("v1"))
			Expect(info.Kind).To(Equal("Pod"))
			Expect(info.Name).To(Equal("test-pod"))
			Expect(info.Namespace).To(Equal("default"))
			Expect(info.Group).To(Equal(""))
			Expect(info.Version).To(Equal("v1"))
		})

		It("should handle resources with group in apiVersion", func() {
			deploymentMap := map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name":      "test-deployment",
					"namespace": "default",
				},
			}

			info, err := util.GetResourceInfo(deploymentMap)
			Expect(err).To(BeNil())
			Expect(info.APIVersion).To(Equal("apps/v1"))
			Expect(info.Kind).To(Equal("Deployment"))
			Expect(info.Group).To(Equal("apps"))
			Expect(info.Version).To(Equal("v1"))
		})

		It("should return error for nil resource", func() {
			_, err := util.GetResourceInfo(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("resource cannot be nil"))
		})
	})

	Describe("JSON Conversion", func() {
		It("should convert from JSON to map", func() {
			jsonBytes := []byte(`{"apiVersion":"v1","kind":"Pod","metadata":{"name":"test"}}`)
			resource, err := util.FromJSON(jsonBytes)
			Expect(err).To(BeNil())
			Expect(resource["apiVersion"]).To(Equal("v1"))
			Expect(resource["kind"]).To(Equal("Pod"))
		})

		It("should convert from map to JSON", func() {
			resource := map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]interface{}{
					"name": "test",
				},
			}
			jsonBytes, err := util.ToJSON(resource)
			Expect(err).To(BeNil())
			Expect(string(jsonBytes)).To(ContainSubstring(`"apiVersion":"v1"`))
		})

	})

	Describe("ResourceInfo", func() {
		It("should detect workload ID presence", func() {
			workloadId := uuid.New().String()
			info := &ResourceInfo{
				Annotations: map[string]string{
					WorkloadIdAnnotation: workloadId,
				},
			}
			Expect(info.HasWorkloadId()).To(BeTrue())
			Expect(info.GetWorkloadId()).To(Equal(workloadId))
		})

		It("should handle missing workload ID", func() {
			info := &ResourceInfo{
				Labels: map[string]string{
					"other-label": "value",
				},
			}
			Expect(info.HasWorkloadId()).To(BeFalse())
			Expect(info.GetWorkloadId()).To(Equal(""))
		})

		It("should format string representation correctly", func() {
			namespacedInfo := &ResourceInfo{
				Kind:      "Pod",
				Name:      "test-pod",
				Namespace: "default",
			}
			Expect(namespacedInfo.GVKkey()).To(Equal("default/test-pod (Pod)"))

			clusterScopedInfo := &ResourceInfo{
				Kind: "Namespace",
				Name: "test-namespace",
			}
			Expect(clusterScopedInfo.GVKkey()).To(Equal("test-namespace (Namespace)"))
		})

		It("should convert to api.Gvk", func() {
			info := &ResourceInfo{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}
			gvk := info.ToGvk()
			Expect(convutils.Deref(gvk.Group, "")).To(Equal("apps"))
			Expect(gvk.Version).To(Equal("v1"))
			Expect(gvk.Kind).To(Equal("Deployment"))
		})
	})
})
