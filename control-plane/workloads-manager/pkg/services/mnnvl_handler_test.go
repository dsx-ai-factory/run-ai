/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"

	orgUnitApi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	clustersClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/clusters"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/org_unit"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/go-viper/mapstructure/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	resourceGvkAlpha = &schema.GroupVersionKind{
		Group:   "batch",
		Version: "v1alpha1",
		Kind:    "Job",
	}

	resourceInfoAlpha1 = &ResourceInfo{
		Version:     resourceGvkAlpha.Version,
		Kind:        resourceGvkAlpha.Kind,
		Group:       resourceGvkAlpha.Group,
		Labels:      map[string]string{},
		Annotations: map[string]string{},
		Name:        "test-job",
		Namespace:   "default",
	}

	resourceGvkV1Beta1 = &schema.GroupVersionKind{
		Group:   "batch",
		Version: "v1beta1",
		Kind:    "Job",
	}

	resourceInfoV1Beta1 = &ResourceInfo{
		Version:     resourceGvkV1Beta1.Version,
		Kind:        resourceGvkV1Beta1.Kind,
		Group:       resourceGvkV1Beta1.Group,
		Labels:      map[string]string{},
		Annotations: map[string]string{},
		Name:        "test-job",
		Namespace:   "default",
	}

	mnnvlLabelKey = "nvidia.com/gpu.clique"
)

var _ = Describe("MNNVL Handler", func() {
	var (
		ctx                  context.Context
		mockClusterClient    *clusterservice.SimpleMock
		mockOrgUnitClient    *org_unit.SimpleMock
		nodePoolWithMNNVL    = "mnnvl-nodepool"
		nodePoolWithoutMNNVL = "default"
	)
	clusterId := "cluster1"

	BeforeEach(func() {
		ctx = context.Background()
		mockClusterClient = clusterservice.NewSimpleMock()
		clustersClient.SetMock(mockClusterClient)
		mockOrgUnitClient = org_unit.NewSimpleMock()
		org_unit.SetMock(mockOrgUnitClient)
		mockClusterClient.WithNodePoolsWithMNNVL(clusterId, mnnvlLabelKey, nodePoolWithMNNVL)
		mockClusterClient.WithNodePools(clusterId, nodePoolWithoutMNNVL)
	})

	AfterEach(func() {
		clustersClient.SetMock(nil)
		org_unit.SetMock(nil)
	})

	Describe("DetermineMNNVLConfiguration", func() {
		Context("when mode is none", func() {
			It("should return none mode", func() {
				configurationNone := api.MNNVLMODE_NONE

				project := &orgUnitApi.Project{
					ClusterId: clusterId,
					Effective: orgUnitApi.ProjectEffective{
						DefaultNodePools: []string{nodePoolWithMNNVL},
					},
				}

				result, err := DetermineMNNVLConfiguration(ctx, &configurationNone, project, clusterId, createWorkloadTypeWithRi(), resourceInfoAlpha1)

				Expect(err).To(BeNil())
				Expect(result).NotTo(BeNil())
				Expect(result.Mnnvl).NotTo(BeNil())
				Expect(result.Mnnvl.Mode).To(Equal(api.MNNVLMODE_NONE))
			})
		})

		Context("when mode is required", func() {
			It("should return error when project node pools don't support it", func() {
				project := &orgUnitApi.Project{
					ClusterId: clusterId,
					Effective: orgUnitApi.ProjectEffective{
						DefaultNodePools: []string{nodePoolWithoutMNNVL},
					},
				}
				workloadType := createWorkloadTypeWithRi(createRiWithGvk(resourceGvkAlpha, &kartav1alpha1.SpecDefinition{
					PodTemplateSpecPath: convutils.Ref("spec.template"),
				}))

				result, err := DetermineMNNVLConfiguration(ctx, convutils.Ref(api.MNNVLMODE_REQUIRED), project,
					clusterId, workloadType, resourceInfoAlpha1)

				Expect(err).NotTo(BeNil())
				Expect(result).To(BeNil())
				Expect(err.Message).To(ContainSubstring("MNNVL required but project node pools does not support GPU network acceleration"))
			})
			It("should return error when RI doesn't support MNNVL", func() {
				project := &orgUnitApi.Project{
					ClusterId: clusterId,
					Effective: orgUnitApi.ProjectEffective{
						DefaultNodePools: []string{nodePoolWithMNNVL},
					},
				}
				workloadType := createWorkloadTypeWithRi(createRiWithGvk(resourceGvkAlpha, &kartav1alpha1.SpecDefinition{
					FragmentedPodSpecDefinition: &kartav1alpha1.FragmentedPodSpecDefinition{
						ContainersPath:  convutils.Ref("spec.containers"),
						PodAffinityPath: convutils.Ref("spec.affinity.podAffinity"),
					},
				}))
				result, err := DetermineMNNVLConfiguration(ctx, convutils.Ref(api.MNNVLMODE_REQUIRED), project, clusterId, workloadType, resourceInfoAlpha1)

				Expect(err).NotTo(BeNil())
				Expect(result).To(BeNil())
				Expect(err.Message).To(ContainSubstring("resource claim"))
			})
		})

		Context("when mode is invalid", func() {
			It("should return error", func() {
				invalidConfiguration := api.MnnvlMode("invalid")
				project := &orgUnitApi.Project{
					ClusterId: clusterId,
					Effective: orgUnitApi.ProjectEffective{
						DefaultNodePools: []string{nodePoolWithMNNVL},
					},
				}

				result, err := DetermineMNNVLConfiguration(ctx, &invalidConfiguration, project, clusterId, nil, resourceInfoAlpha1)

				Expect(err).NotTo(BeNil())
				Expect(result).To(BeNil())
			})
		})

		Context("RI version matching", func() {
			It("should return error  version doesn't match any RI version", func() {

				project := &orgUnitApi.Project{
					ClusterId: clusterId,
					Effective: orgUnitApi.ProjectEffective{
						DefaultNodePools: []string{nodePoolWithMNNVL},
					},
				}

				workloadType := createWorkloadTypeWithRi(createRiWithGvk(resourceGvkV1Beta1, &kartav1alpha1.SpecDefinition{
					PodTemplateSpecPath: convutils.Ref("spec.template"),
				}))

				result, err := DetermineMNNVLConfiguration(ctx, convutils.Ref(api.MNNVLMODE_REQUIRED), project, clusterId, workloadType, resourceInfoAlpha1)

				Expect(err).NotTo(BeNil())
				Expect(result).To(BeNil())
				Expect(err.Message).To(ContainSubstring("no karta found for version v1alpha1"))
			})

			It("should select correct RI from multiple versions", func() {
				configuration := api.MNNVLMODE_REQUIRED

				project := &orgUnitApi.Project{
					ClusterId: clusterId,
					Effective: orgUnitApi.ProjectEffective{
						DefaultNodePools: []string{nodePoolWithMNNVL},
					},
				}
				workloadType := createWorkloadTypeWithRi(createRiWithGvk(resourceGvkAlpha, &kartav1alpha1.SpecDefinition{
					PodTemplateSpecPath: convutils.Ref("spec.template"),
				}), createRiWithGvk(resourceGvkV1Beta1, &kartav1alpha1.SpecDefinition{
					FragmentedPodSpecDefinition: &kartav1alpha1.FragmentedPodSpecDefinition{},
				}))

				result, err := DetermineMNNVLConfiguration(ctx, &configuration, project, clusterId, workloadType, resourceInfoAlpha1)

				Expect(err).To(BeNil())
				Expect(result).NotTo(BeNil())
				Expect(result.Mnnvl).NotTo(BeNil())
				Expect(result.Mnnvl.Mode).To(Equal(api.MNNVLMODE_REQUIRED))
			})
		})

	})

})

func createWorkloadTypeWithRi(kartas ...*kartav1alpha1.Karta) *workloadsApi.WorkloadTypeConfig {
	workloadType := &workloadsApi.WorkloadTypeConfig{
		Kartas: make([]workloadsApi.WorkloadTypeKartasWithVersionInner, len(kartas)),
	}
	for i, k := range kartas {
		kartaMap := map[string]interface{}{}
		mapstructure.Decode(k, &kartaMap)
		Expect(k.Spec.StructureDefinition.RootComponent.Kind).NotTo(BeNil())
		workloadType.Kartas[i] = workloadsApi.WorkloadTypeKartasWithVersionInner{
			Version: &k.Spec.StructureDefinition.RootComponent.Kind.Version,
			Karta:   kartaMap,
		}
	}
	return workloadType
}

func createRiWithGvk(gvk *schema.GroupVersionKind, specDefinition *kartav1alpha1.SpecDefinition) *kartav1alpha1.Karta {
	return &kartav1alpha1.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "root",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   gvk.Group,
						Version: gvk.Version,
						Kind:    gvk.Kind,
					},
					SpecDefinition: specDefinition,
				},
			},
		},
	}
}
