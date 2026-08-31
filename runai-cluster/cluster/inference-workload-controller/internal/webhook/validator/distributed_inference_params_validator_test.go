/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validator

import (
	"context"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/topology"
	mnnvltypes "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	dinfp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferencepolicy/v1alpha1"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	commontesting "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common/testing"
	"github.com/run-ai/runai/runai-cluster/common/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

const (
	DistributedWorkloadName        = "test-distributed-workload"
	distributedNamespace           = "runai-default"
	distributedImage               = "test-image"
	testTopologyAnnotationKeyValue = "name"
)

var _ = Describe("Distributed Inference Parameters Validator", func() {
	var (
		workload  *dinfw.DistributedInferenceWorkload
		validator *DistributedInferenceParamsValidator
		ctx       context.Context
		k8sClient client.Client
	)

	Context("validate basic fields", func() {
		BeforeEach(func() {
			workload = &dinfw.DistributedInferenceWorkload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      DistributedWorkloadName,
					Namespace: distributedNamespace,
				},
				Spec: dinfp.DistributedInferenceSpec{
					DistributedInferenceFields: field_groups.DistributedInferenceFields{
						ServingPorts: &field_types.ServingPortsField{
							Value: &map[string]field_types.ServingPortItemAndRules{
								"serving-port": {
									Value: &field_types.ServingPortItemValue{
										Port: ptr.To(8080),
									},
								},
							},
						},
					},
				},
				LeaderSpec: dinfp.DistributedInferenceEntitySpec{
					ContainerFields: field_groups.ContainerFields{
						Image: field_types.NewStringField(distributedImage),
					},
				},
			}
			ctx = context.Background()
			validator = NewDistributedInferenceParamsValidator(nil, true, false, false)
		})

		Context("validate spec fields", func() {
			Context("validate workers field", func() {
				It("should be valid with zero workers and no worker spec", func() {
					workload.Spec.DistributedInferenceFields.Workers = &field_types.IntField{
						Value: ptr.To(0),
					}
					workload.WorkerSpec = nil

					err := validator.Validate(ctx, workload)
					Expect(err).To(BeNil())
				})

				It("should be valid with positive workers and worker spec", func() {
					workload.Spec.DistributedInferenceFields.Workers = &field_types.IntField{
						Value: ptr.To(2),
					}
					workload.WorkerSpec = &dinfp.DistributedInferenceEntitySpec{
						ContainerFields: field_groups.ContainerFields{
							Image: field_types.NewStringField(distributedImage),
						},
					}

					err := validator.Validate(ctx, workload)
					Expect(err).To(BeNil())
				})

				It("should be invalid with negative workers", func() {
					workload.Spec.DistributedInferenceFields.Workers = &field_types.IntField{
						Value: ptr.To(-1),
					}

					err := validator.Validate(ctx, workload)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(ContainSubstring("workers must be non-negative"))
				})

				It("should be invalid with zero workers but worker spec provided", func() {
					workload.Spec.DistributedInferenceFields.Workers = &field_types.IntField{
						Value: ptr.To(0),
					}
					workload.WorkerSpec = &dinfp.DistributedInferenceEntitySpec{
						ContainerFields: field_groups.ContainerFields{
							Image: field_types.NewStringField(distributedImage),
						},
					}

					err := validator.Validate(ctx, workload)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(ContainSubstring("workers cannot be zero when workerSpec is provided"))
				})

				It("should be invalid with positive workers but no worker spec", func() {
					workload.Spec.DistributedInferenceFields.Workers = &field_types.IntField{
						Value: ptr.To(2),
					}
					workload.WorkerSpec = nil

					err := validator.Validate(ctx, workload)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(ContainSubstring("workerSpec must be provided when workers is greater than zero"))
				})
			})

			Context("validate replicas field", func() {
				It("should be valid with positive replicas", func() {
					workload.Spec.DistributedInferenceFields.Replicas = &field_types.IntField{
						Value: ptr.To(3),
					}

					err := validator.Validate(ctx, workload)
					Expect(err).To(BeNil())
				})

				It("should be invalid with negative replicas", func() {
					workload.Spec.DistributedInferenceFields.Replicas = &field_types.IntField{
						Value: ptr.To(-1),
					}

					err := validator.Validate(ctx, workload)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(ContainSubstring("replicas must be non-negative"))
				})
			})

			Context("validate labels", func() {
				It("should be invalid with empty label key", func() {
					workload.Labels = map[string]string{
						"": "value",
					}

					err := validator.Validate(ctx, workload)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(ContainSubstring("label key and value must not be empty"))
				})
			})
		})

		Context("validate serving ports", func() {
			It("should be valid with serving ports provided", func() {
				// (serving ports are set in BeforeEach)
				err := validator.Validate(ctx, workload)
				Expect(err).To(BeNil())
			})

			It("should be invalid without serving ports", func() {
				workload.Spec.DistributedInferenceFields.ServingPorts = nil

				err := validator.Validate(ctx, workload)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("servingPorts must be provided"))
			})
		})

		Context("validate entity specs", func() {
			Context("validate leader spec", func() {
				It("should be valid with proper leader spec", func() {
					// (leader spec is set in BeforeEach)
					err := validator.Validate(ctx, workload)
					Expect(err).To(BeNil())
				})

				It("should be invalid with empty image in leader spec", func() {
					workload.LeaderSpec.ContainerFields.Image = field_types.NewStringField("")

					err := validator.Validate(ctx, workload)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(ContainSubstring("failed to validate leaderSpec"))
					Expect(err.Error()).To(ContainSubstring("image must be provided"))
				})

				It("should be invalid with invalid PVCs in leader spec", func() {
					workload.LeaderSpec.StorageFields.Pvcs = &field_types.PvcsField{
						Value: &map[string]field_types.PvcsItemAndRules{
							"invalid-pvc": {
								Value: &field_types.PvcsItemValue{
									Size: utils.Ref("invalid-size"),
								},
							},
						},
					}

					err := validator.Validate(ctx, workload)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(ContainSubstring("failed to validate leaderSpec"))
					Expect(err.Error()).To(ContainSubstring("failed to validate PVCs"))
				})

				It("should be invalid with invalid pod affinity in leader spec", func() {
					workload.LeaderSpec.DistributedInferenceEntitySchedulerFields.PodAffinity = &field_types.BoolField{
						Value: ptr.To(false),
					}
					workload.LeaderSpec.DistributedInferenceEntitySchedulerFields.PodAffinityTopology = &field_types.StringField{
						Value: ptr.To("topology.kubernetes.io/zone"),
					}

					err := validator.Validate(ctx, workload)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(ContainSubstring("failed to validate leaderSpec"))
					Expect(err.Error()).To(ContainSubstring("podAffinity must be enabled"))
				})
			})

			Context("validate worker spec", func() {
				BeforeEach(func() {
					workload.Spec.DistributedInferenceFields.Workers = &field_types.IntField{
						Value: ptr.To(2),
					}
					workload.WorkerSpec = &dinfp.DistributedInferenceEntitySpec{
						ContainerFields: field_groups.ContainerFields{
							Image: field_types.NewStringField(distributedImage),
						},
					}
				})

				It("should be valid with proper worker spec", func() {
					err := validator.Validate(ctx, workload)
					Expect(err).To(BeNil())
				})

				It("should be invalid with empty image in worker spec", func() {
					workload.WorkerSpec.ContainerFields.Image = field_types.NewStringField("")

					err := validator.Validate(ctx, workload)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(ContainSubstring("failed to validate workerSpec"))
					Expect(err.Error()).To(ContainSubstring("image must be provided"))
				})

				It("should be invalid with invalid PVCs in worker spec", func() {
					workload.WorkerSpec.StorageFields.Pvcs = &field_types.PvcsField{
						Value: &map[string]field_types.PvcsItemAndRules{
							"invalid-pvc": {
								Value: &field_types.PvcsItemValue{
									Size: utils.Ref("invalid-size"),
								},
							},
						},
					}

					err := validator.Validate(ctx, workload)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(ContainSubstring("failed to validate workerSpec"))
					Expect(err.Error()).To(ContainSubstring("failed to validate PVCs"))
				})

				It("should be invalid with invalid pod affinity in worker spec", func() {
					workload.WorkerSpec.DistributedInferenceEntitySchedulerFields.PodAffinity = &field_types.BoolField{
						Value: ptr.To(false),
					}
					workload.WorkerSpec.DistributedInferenceEntitySchedulerFields.PodAffinityTopology = &field_types.StringField{
						Value: ptr.To("topology.kubernetes.io/zone"),
					}

					err := validator.Validate(ctx, workload)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(ContainSubstring("failed to validate workerSpec"))
					Expect(err.Error()).To(ContainSubstring("podAffinity must be enabled"))
				})
			})

			Context("validate resource allocation", func() {
				It("should be invalid with invalid CPU resources in leader spec", func() {
					workload.LeaderSpec.ResourceAllocationFields.Cpu = &field_types.NumberField{
						Value: ptr.To("2"),
					}
					workload.LeaderSpec.ResourceAllocationFields.CpuLimit = &field_types.NumberField{
						Value: ptr.To("1"),
					}

					err := validator.Validate(ctx, workload)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(ContainSubstring("failed to validate leaderSpec"))
					Expect(err.Error()).To(ContainSubstring("limit (1) must be equal to or higher than"))
				})

				It("should be invalid with invalid memory resources in worker spec", func() {
					workload.Spec.DistributedInferenceFields.Workers = &field_types.IntField{
						Value: ptr.To(2),
					}
					workload.WorkerSpec = &dinfp.DistributedInferenceEntitySpec{
						ContainerFields: field_groups.ContainerFields{
							Image: field_types.NewStringField(distributedImage),
						},
						ResourceAllocationFields: field_groups.ResourceAllocationFields{
							Memory: &field_types.QuantityField{
								Value: ptr.To("2Gi"),
							},
							MemoryLimit: &field_types.QuantityField{
								Value: ptr.To("1Gi"),
							},
						},
					}

					err := validator.Validate(ctx, workload)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(ContainSubstring("failed to validate workerSpec"))
					Expect(err.Error()).To(ContainSubstring("limit (1Gi) must be equal to or higher than"))
				})
			})
		})

		Context("validate overall function", func() {
			It("should be valid with all required fields", func() {
				workload.Spec.DistributedInferenceFields.Workers = &field_types.IntField{
					Value: ptr.To(2),
				}
				workload.WorkerSpec = &dinfp.DistributedInferenceEntitySpec{
					ContainerFields: field_groups.ContainerFields{
						Image: field_types.NewStringField(distributedImage),
					},
				}

				err := validator.Validate(ctx, workload)
				Expect(err).To(BeNil())
			})

			It("should be valid with only leader spec and zero workers", func() {
				workload.Spec.DistributedInferenceFields.Workers = &field_types.IntField{
					Value: ptr.To(0),
				}
				workload.WorkerSpec = nil

				err := validator.Validate(ctx, workload)
				Expect(err).To(BeNil())
			})

			It("should return first error encountered", func() {
				// Set multiple invalid fields
				workload.Spec.DistributedInferenceFields.Workers = &field_types.IntField{
					Value: ptr.To(-1),
				}
				workload.LeaderSpec.ContainerFields.Image = field_types.NewStringField("")

				err := validator.Validate(ctx, workload)
				Expect(err).ToNot(BeNil())
				// Should return workers validation error first
				Expect(err.Error()).To(ContainSubstring("workers must be non-negative"))
			})
		})
	})

	Context("validate MNNVL", func() {
		BeforeEach(func() {
			kaiFakeNodePool := kaiv1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{Name: "nodepoola"},
				Spec: kaiv1alpha1.NodePoolSpec{
					LabelKey:   "test-key",
					LabelValue: "test-value",
				},
			}

			kaiMnnvlNodepool := kaiv1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{
					Name: "mnnvl-nodepool",
					Annotations: map[string]string{
						kaiv1alpha1.AnnotationGPUNetworkAccelerationDetected: "true",
						kaiv1alpha1.AnnotationGPUNetworkAccelerationLabelKey: "nvidia.com/gpu.imex-domain",
					},
				},
				Spec: kaiv1alpha1.NodePoolSpec{
					LabelKey:   "nodepool-key-1",
					LabelValue: "nodepool-value-1",
				},
				Status: kaiv1alpha1.NodePoolStatus{Phase: kaiv1alpha1.NodePoolReady},
			}

			var t testing.T // dummy, for the fake client
			k8sClient = commontesting.NewFakeClient(&t, &kaiFakeNodePool, &kaiMnnvlNodepool)

			workload = &dinfw.DistributedInferenceWorkload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      DistributedWorkloadName,
					Namespace: distributedNamespace,
				},
				Spec: dinfp.DistributedInferenceSpec{
					DistributedInferenceFields: field_groups.DistributedInferenceFields{
						ServingPorts: &field_types.ServingPortsField{
							Value: &map[string]field_types.ServingPortItemAndRules{
								"serving-port": {
									Value: &field_types.ServingPortItemValue{
										Port: ptr.To(8080),
									},
								},
							},
						},
					},
				},
				LeaderSpec: dinfp.DistributedInferenceEntitySpec{
					ContainerFields: field_groups.ContainerFields{
						Image: field_types.NewStringField(distributedImage),
					},
					ResourceAllocationFields: field_groups.ResourceAllocationFields{
						Gpu: &field_types.NumberField{
							Value: ptr.To("1"),
						},
					},
				},
			}
			ctx = context.Background()
		})

		Context("when MNNVL label is not set", func() {
			It("should be valid", func() {
				validator = NewDistributedInferenceParamsValidator(k8sClient, true, false, false)
				workload.Labels = nil

				err := validator.Validate(ctx, workload)
				Expect(err).To(BeNil())
			})
		})

		Context("when MNNVL label is set to None", func() {
			It("should be valid", func() {
				validator = NewDistributedInferenceParamsValidator(k8sClient, true, false, false)
				workload.Labels = map[string]string{
					mnnvltypes.FeatureLabelKey: mnnvltypes.None,
				}

				err := validator.Validate(ctx, workload)
				Expect(err).To(BeNil())
			})
		})

		Context("when MNNVL label is set to an invalid value", func() {
			It("should be invalid with Preferred mode", func() {
				validator = NewDistributedInferenceParamsValidator(k8sClient, true, false, false)
				workload.Labels = map[string]string{
					mnnvltypes.FeatureLabelKey: mnnvltypes.Preferred,
				}

				err := validator.Validate(ctx, workload)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("invalid mode: Preferred"))
				Expect(err.Error()).To(ContainSubstring("only None and Required are allowed"))
			})

			It("should be invalid with completely invalid mode", func() {
				validator = NewDistributedInferenceParamsValidator(k8sClient, true, false, false)
				workload.Labels = map[string]string{
					mnnvltypes.FeatureLabelKey: "InvalidMode",
				}

				err := validator.Validate(ctx, workload)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("invalid mode: InvalidMode"))
				Expect(err.Error()).To(ContainSubstring("only None and Required are allowed"))
			})
		})

		Context("when MNNVL label is set to Required", func() {
			Context("with Nvidia GPU vendor", func() {
				Context("when MNNVL is not enabled in cluster", func() {
					It("should be invalid", func() {
						validator = NewDistributedInferenceParamsValidator(k8sClient, false, false, false)
						workload.Labels = map[string]string{
							mnnvltypes.FeatureLabelKey: mnnvltypes.Required,
						}

						err := validator.Validate(ctx, workload)
						Expect(err).ToNot(BeNil())
						Expect(err.Error()).To(ContainSubstring("failed to validate GPU network acceleration (MNNVL)"))
						Expect(err.Error()).To(ContainSubstring("the capability is not enabled in the cluster"))
					})
				})

				Context("when MNNVL is enabled in cluster", func() {
					It("should succeed if node pools support MNNVL", func() {
						validator = NewDistributedInferenceParamsValidator(k8sClient, true, false, false)
						workload.Labels = map[string]string{
							mnnvltypes.FeatureLabelKey: mnnvltypes.Required,
						}
						workload.Spec.NodePools = &field_types.StringField{
							Value: ptr.To("mnnvl-nodepool"),
						}

						err := validator.Validate(ctx, workload)
						Expect(err).To(BeNil())
					})

					It("should fail if node pools don't support MNNVL", func() {
						validator = NewDistributedInferenceParamsValidator(k8sClient, true, false, false)
						workload.Labels = map[string]string{
							mnnvltypes.FeatureLabelKey: mnnvltypes.Required,
						}
						workload.Spec.NodePools = &field_types.StringField{
							Value: ptr.To("nodepoola"),
						}

						err := validator.Validate(ctx, workload)
						Expect(err).ToNot(BeNil())
						Expect(err.Error()).To(ContainSubstring("failed to validate GPU network acceleration (MNNVL)"))
						Expect(err.Error()).To(ContainSubstring("at least one node pool"))
					})
				})
			})
		})
	})

	Context("validate topology", func() {
		BeforeEach(func() {
			validator = NewDistributedInferenceParamsValidator(k8sClient, false, true, false)
			workload = &dinfw.DistributedInferenceWorkload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      DistributedWorkloadName,
					Namespace: distributedNamespace,
				},
				Spec: dinfp.DistributedInferenceSpec{
					DistributedInferenceFields: field_groups.DistributedInferenceFields{
						ServingPorts: &field_types.ServingPortsField{
							Value: &map[string]field_types.ServingPortItemAndRules{
								"serving-port": {
									Value: &field_types.ServingPortItemValue{
										Port: ptr.To(8080),
									},
								},
							},
						},
					},
				},
				LeaderSpec: dinfp.DistributedInferenceEntitySpec{
					ContainerFields: field_groups.ContainerFields{
						Image: field_types.NewStringField(distributedImage),
					},
				},
			}
			ctx = context.Background()
		})

		Context("valid", func() {
			It("no annotations", func() {
				err := validator.Validate(ctx, workload)
				Expect(err).To(BeNil())
			})

			It("no topology annotations", func() {
				workload.Annotations = map[string]string{}
				err := validator.Validate(ctx, workload)
				Expect(err).To(BeNil())
			})

			It("topology annotations exist with name and preferred", func() {
				workload.Annotations = make(map[string]string)
				workload.Annotations[topology.KAITopologyAnnotationKey] = testTopologyAnnotationKeyValue
				workload.Annotations[topology.KAITopologyAnnotationPreferredKey] = "preferred"

				err := validator.Validate(ctx, workload)
				Expect(err).To(BeNil())
				Expect(workload.Annotations).To(HaveLen(2))
				Expect(workload.Annotations[topology.KAITopologyAnnotationKey]).To(Equal("name"))
				Expect(workload.Annotations[topology.KAITopologyAnnotationPreferredKey]).To(Equal("preferred"))
			})

			It("topology annotations exist with name and required", func() {
				workload.Annotations = make(map[string]string)
				workload.Annotations[topology.KAITopologyAnnotationKey] = testTopologyAnnotationKeyValue
				workload.Annotations[topology.KAITopologyAnnotationRequiredKey] = "required"

				err := validator.Validate(ctx, workload)
				Expect(err).To(BeNil())
				Expect(workload.Annotations).To(HaveLen(2))
				Expect(workload.Annotations[topology.KAITopologyAnnotationKey]).To(Equal("name"))
				Expect(workload.Annotations[topology.KAITopologyAnnotationRequiredKey]).To(Equal("required"))
			})

			It("topology annotations exist with name, required and preferred", func() {
				workload.Annotations = make(map[string]string)
				workload.Annotations[topology.KAITopologyAnnotationKey] = testTopologyAnnotationKeyValue
				workload.Annotations[topology.KAITopologyAnnotationPreferredKey] = "preferred"
				workload.Annotations[topology.KAITopologyAnnotationRequiredKey] = "required"

				err := validator.Validate(ctx, workload)
				Expect(err).To(BeNil())
				Expect(workload.Annotations).To(HaveLen(3))
				Expect(workload.Annotations[topology.KAITopologyAnnotationKey]).To(Equal("name"))
				Expect(workload.Annotations[topology.KAITopologyAnnotationPreferredKey]).To(Equal("preferred"))
				Expect(workload.Annotations[topology.KAITopologyAnnotationRequiredKey]).To(Equal("required"))
			})
		})

		Context("invalid", func() {
			It("topology annotations name is empty", func() {
				workload.Annotations = make(map[string]string)
				workload.Annotations[topology.KAITopologyAnnotationKey] = ""

				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("failed to validate topology: the 'kai.scheduler/topology' annotation is set with no value"))
			})

			It("topology annotations preferred is missing", func() {
				workload.Annotations = make(map[string]string)
				workload.Annotations[topology.KAITopologyAnnotationKey] = testTopologyAnnotationKeyValue

				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("failed to validate topology: topology is enabled, but no topology level is defined. " +
					"expected 'kai.scheduler/topology-preferred-placement' or 'kai.scheduler/topology-required-placement'"))
			})
		})
	})
})
