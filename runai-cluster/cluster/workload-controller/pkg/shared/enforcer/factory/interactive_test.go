/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package factory

import (
	"context"
	"encoding/json"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer"
	corev1 "k8s.io/api/core/v1"
	"os"

	clusterCommonConstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	workloadproject "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/project"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	ipv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	iwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	job "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils"
	client2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("InteractiveWorkloadMergeCreator", func() {
	var (
		k8sClient                       client.Client
		interactiveWorkloadMergeCreator *InteractiveWorkloadMergeCreator
		scheme                          *runtime.Scheme
	)
	BeforeEach(func() {
		scheme = runtime.NewScheme()
		register.InitSchema(scheme)

	})

	Context("CreateMergeStack", func() {
		BeforeEach(func() {
			k8sClient = client2.InitFakeClientWithObjectNoStatus(scheme)
			interactiveWorkloadMergeCreator = NewInteractiveWorkloadMergeCreator(k8sClient, true)
		})
		It("cant use self as base workload", func() {
			//
			//   cannot base workload on self
			//
			selfBased := iwv2.InteractiveWorkload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mytemp",
					Namespace: "runai-team-a",
				},
				Spec: ipv2.InteractiveSpec{
					WorkloadFields: field_groups.WorkloadFields{
						IdentityFields: field_groups.IdentityFields{
							BaseWorkload: ptr.To("mytemp"),
						},
					},
				},
			}
			err := k8sClient.Create(context.Background(), &selfBased)
			Expect(err).ToNot(HaveOccurred())

			_, err = interactiveWorkloadMergeCreator.CreateMergeStack(context.TODO(), &selfBased)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("base workload cannot be the same workload"))
		})
		It("cant find base workload", func() {
			notFound := iwv2.InteractiveWorkload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testos",
					Namespace: "runai-team-a",
				},
				Spec: ipv2.InteractiveSpec{
					WorkloadFields: field_groups.WorkloadFields{
						IdentityFields: field_groups.IdentityFields{
							BaseWorkload: ptr.To("notfound"),
						},
					},
				},
			}
			_, err := interactiveWorkloadMergeCreator.CreateMergeStack(context.TODO(), &notFound)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to find base workload"))

		})
		Context("should create MergeStack", func() {
			var (
				basedOnSpec *iwv2.InteractiveWorkload
			)
			workingDirFromBase := "working-dir-from-base"

			baseName := "base-name"
			basedOnSpecName := "based-on-spec"

			BeforeEach(func() {

				//
				//   setup base template
				//
				base := iwv2.InteractiveWorkload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      baseName,
						Namespace: "runai-team-a",
					},
					Spec: ipv2.InteractiveSpec{
						WorkloadFields: field_groups.WorkloadFields{
							IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
								Name: &flds.StringField{
									Value: ptr.To("name-from-base"),
								},
							},
							ContainerFields: field_groups.ContainerFields{
								WorkingDir: flds.NewStringField(workingDirFromBase),
							},
						},
						ResourceAllocationFields: field_groups.ResourceAllocationFields{
							Gpu: flds.NewOptionalNumberField(ptr.To[float64](4.0)),
						},
					},
				}

				//
				//    setup workload
				//
				basedOnSpec = &iwv2.InteractiveWorkload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      basedOnSpecName,
						Namespace: "runai-team-a",
					},
					Spec: ipv2.InteractiveSpec{
						WorkloadFields: field_groups.WorkloadFields{
							IdentityFields: field_groups.IdentityFields{
								BaseWorkload: ptr.To(baseName),
							},
						},
					},
				}

				k8sClient = client2.InitFakeClientWithObjectNoStatus(scheme, &base, basedOnSpec)
				interactiveWorkloadMergeCreator = NewInteractiveWorkloadMergeCreator(k8sClient, true)
			})
			It("validate merge stack", func() {

				mergeStackMap, err := interactiveWorkloadMergeCreator.CreateMergeStack(context.TODO(), basedOnSpec)
				Expect(err).ToNot(HaveOccurred())
				mergeStack := mergeStackMap[api.SpecSubResource]
				err = mergeStack.MergeAndValidate(context.TODO(), enforcer.MergeAndValidateOptions{})
				Expect(err).ToNot(HaveOccurred())
				//
				//   the expected results are:
				//   name - not given (we do not take name from base)
				//   gpu - taken from the workload (2.0)
				//   workingDir - taken from the base
				//
				merged1, err := job.UnmarshalWorkload(mergeStack.MergedValues(), &basedOnSpec.TypeMeta, &basedOnSpec.ObjectMeta)
				Expect(err).ToNot(HaveOccurred())

				Expect(merged1.Spec.Name).To(BeNil(), "part1 - name - nil")
				Expect(ptr.Deref(merged1.Spec.Gpu.Value, "")).To(Equal("4"), "part1 - gpu - from workload")
				Expect(merged1.Spec.WorkingDir.Deref()).To(Equal(workingDirFromBase), "part1 - workingDir - from workload")
			})
			Context("include policy", func() {
				var (
					temp *iwv2.InteractiveWorkload
				)
				completionFromSett := 3
				completionFromBase := 5
				completionFromTemp := 7
				maxCompletionValue := 6
				BeforeEach(func() {
					policy := ipv2.InteractivePolicy{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "policy",
							Namespace: clusterCommonConstants.NamespaceRunAi,
						},
						Spec: ipv2.InteractiveSpec{
							WorkloadFields: field_groups.WorkloadFields{
								ContainerFields: field_groups.ContainerFields{
									RunAsUid: &flds.IntField{
										Value: ptr.To(completionFromSett),
										Rules: &flds.IntRules{
											Max: ptr.To(maxCompletionValue),
										},
									},
									RunAsUser: flds.NewOptionalBoolField(ptr.To(true)),
								},
							},
						},
					}

					anotherPolicy := ipv2.InteractivePolicy{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "ignored-policy",
							Namespace: clusterCommonConstants.NamespaceRunAi,
						},
						Spec: ipv2.InteractiveSpec{
							WorkloadFields: field_groups.WorkloadFields{
								ContainerFields: field_groups.ContainerFields{
									RunAsUid: &flds.IntField{
										Value: ptr.To(completionFromSett),
										Rules: &flds.IntRules{
											Max: ptr.To(1),
										},
									},
									RunAsUser: flds.NewOptionalBoolField(ptr.To(false)),
								},
							},
						},
					}

					base := &iwv2.InteractiveWorkload{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "base",
							Namespace: "runai-team-a",
						},
						Spec: ipv2.InteractiveSpec{
							WorkloadFields: field_groups.WorkloadFields{
								ContainerFields: field_groups.ContainerFields{
									RunAsUid: flds.NewOptionalIntField(ptr.To(completionFromBase)),
								},
							},
							ResourceAllocationFields: field_groups.ResourceAllocationFields{
								GpuMemory: flds.NewQuantityField("123M"),
							},
						},
					}

					temp = &iwv2.InteractiveWorkload{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "temp",
							Namespace: "runai-team-a",
						},
						Spec: ipv2.InteractiveSpec{
							WorkloadFields: field_groups.WorkloadFields{
								IdentityFields: field_groups.IdentityFields{
									Usage:        common.UsedForSubmit,
									BaseWorkload: ptr.To("base"),
								},
								ContainerFields: field_groups.ContainerFields{
									Image:    flds.NewStringField("ubuntu"),
									RunAsUid: flds.NewOptionalIntField(ptr.To(completionFromTemp)),
								},
							},
						},
					}
					err := k8sClient.Create(context.Background(), &policy)
					Expect(err).ToNot(HaveOccurred())
					err = k8sClient.Create(context.Background(), &anotherPolicy)
					Expect(err).ToNot(HaveOccurred())
					err = k8sClient.Create(context.Background(), base)
					Expect(err).ToNot(HaveOccurred())
					err = k8sClient.Create(context.Background(), temp)
					Expect(err).ToNot(HaveOccurred())
					err = k8sClient.Create(context.Background(), workloadproject.EmulateProjNs("runai-team-a", "team-a"))
					Expect(err).ToNot(HaveOccurred())
				})
				It("should failed validate, runAsUid.value is not valid by policy", func() {
					mergeStackMap, err := interactiveWorkloadMergeCreator.CreateMergeStack(context.TODO(), temp)
					Expect(err).ToNot(HaveOccurred())
					mergeStack := mergeStackMap[api.SpecSubResource]
					err = mergeStack.MergeAndValidate(context.TODO(), enforcer.MergeAndValidateOptions{})
					Expect(err).To(HaveOccurred(), "policy -> merge")
					Expect(err.Error()).To(ContainSubstring("runAsUid.value: must be no greater than 1"))

				})

			})
		})
	})

	Context("CreateMergeStack with backend policy sync", func() {
		projectName1 := "p1"
		departmentName1 := "d1"
		namespaceName1 := "runai-p1"
		basePolicyWithoutLevel := ipv2.InteractivePolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: "policy",
			},
			Spec: ipv2.InteractiveSpec{
				ResourceAllocationFields: field_groups.ResourceAllocationFields{
					Gpu: &flds.NumberField{
						Value: ptr.To("0"),
						Rules: &flds.NumberRules{
							Max: ptr.To("6"),
						},
					},
				},
			},
		}

		var namespace1, projectP1, departmentD1 client.Object

		Context("backend policy project level", func() {
			BeforeEach(func() {
				namespace1 = workloadproject.EmulateProjNs(namespaceName1, projectName1)
				projectP1 = &kaiorg.Project{
					ObjectMeta: metav1.ObjectMeta{
						Name: projectName1,
						Labels: map[string]string{
							"runai/department-id": "1",
							"runai/project-id":    "1",
						},
					},
				}
				departmentD1 = &v2alpha2.Department{
					ObjectMeta: metav1.ObjectMeta{
						Name:      departmentName1,
						Namespace: runaiv1.DefaultNamespace,
						Labels: map[string]string{
							"runai/department-id": "1",
						},
					},
				}

				os.Setenv(utils.EnvVarControlPlanePolicies, "true")
				k8sClient = client2.InitFakeClientWithObjectNoStatus(scheme, projectP1, departmentD1, namespace1)
				interactiveWorkloadMergeCreator = NewInteractiveWorkloadMergeCreator(k8sClient, true)

				// project level policy definition
				basePolicyWithoutLevel.Namespace = namespaceName1
				basePolicyWithoutLevel.Labels = map[string]string{
					clusterCommonConstants.PolicyLabel:       clusterCommonConstants.BoolTrueAsString,
					clusterCommonConstants.ProjectScopeLabel: "1",
				}
				basePolicyWithoutLevel.ResourceVersion = ""
			})

			AfterEach(func() {
				os.Unsetenv(utils.EnvVarControlPlanePolicies)
			})

			It("should failed validate, gpu is not valid by policy project level", func() {

				err := k8sClient.Create(context.Background(), &basePolicyWithoutLevel)
				Expect(err).ToNot(HaveOccurred())
				tempWorkload := &iwv2.InteractiveWorkload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "temp",
						Namespace: namespaceName1,
					},
					Spec: ipv2.InteractiveSpec{
						ResourceAllocationFields: field_groups.ResourceAllocationFields{
							Gpu: flds.NewOptionalNumberField(ptr.To[float64](7)),
						},
					},
				}

				mergeStackMap, err := interactiveWorkloadMergeCreator.CreateMergeStack(context.TODO(), tempWorkload)
				Expect(err).ToNot(HaveOccurred()) // found policy
				mergeStack := mergeStackMap[api.SpecSubResource]
				err = mergeStack.MergeAndValidate(context.TODO(), enforcer.MergeAndValidateOptions{})

				Expect(err).To(HaveOccurred(), "policy -> merge")
				Expect(err.Error()).To(ContainSubstring("gpu.value: must be no greater than 6"))

			})

			It("should success validate, gpu is not valid by policy project level", func() {
				err := k8sClient.Create(context.Background(), &basePolicyWithoutLevel)
				Expect(err).ToNot(HaveOccurred())

				tempWorkload := &iwv2.InteractiveWorkload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "temp",
						Namespace: namespaceName1,
					},
					Spec: ipv2.InteractiveSpec{
						ResourceAllocationFields: field_groups.ResourceAllocationFields{
							Gpu: flds.NewOptionalNumberField(ptr.To[float64](5)),
						},
					},
				}

				mergeStackMap, err := interactiveWorkloadMergeCreator.CreateMergeStack(context.TODO(), tempWorkload)
				Expect(err).ToNot(HaveOccurred())
				mergeStack := mergeStackMap[api.SpecSubResource]
				err = mergeStack.MergeAndValidate(context.TODO(), enforcer.MergeAndValidateOptions{})
				Expect(err).ToNot(HaveOccurred())
			})

			It("should ignore node affinity from policy", func() {
				basePolicyWithoutLevel.Spec.NodeAffinityRequired = &flds.NodeSelector{
					NodeSelectorTerms: []flds.NodeSelectorTerm{
						{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      "bla",
									Operator: corev1.NodeSelectorOpDoesNotExist,
								},
							},
						},
					},
				}

				err := k8sClient.Create(context.Background(), &basePolicyWithoutLevel)
				Expect(err).ToNot(HaveOccurred())
				tempWorkload := &iwv2.InteractiveWorkload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "temp",
						Namespace: namespaceName1,
					},
					Spec: ipv2.InteractiveSpec{
						ResourceAllocationFields: field_groups.ResourceAllocationFields{
							Gpu: flds.NewOptionalNumberField(ptr.To[float64](5)),
						},
					},
				}

				mergeStackMap, err := interactiveWorkloadMergeCreator.CreateMergeStack(context.TODO(), tempWorkload)
				Expect(err).ToNot(HaveOccurred())
				mergeStack := mergeStackMap[api.SpecSubResource]
				err = mergeStack.MergeAndValidate(context.TODO(), enforcer.MergeAndValidateOptions{})
				Expect(err).ToNot(HaveOccurred())

				values := mergeStack.MergedValues()
				jsonDump, err := json.Marshal(values.JumpToTop().Get())
				Expect(err).ToNot(HaveOccurred())

				err = json.Unmarshal(jsonDump, &tempWorkload.Spec)
				Expect(err).ToNot(HaveOccurred())

				Expect(tempWorkload.Spec.NodeAffinityRequired).To(BeNil())
			})

			Context("GetPolicy", func() {
				BeforeEach(func() {
					err := k8sClient.Create(context.Background(), &basePolicyWithoutLevel)
					Expect(err).ToNot(HaveOccurred())
				})
				It("should success get policy", func() {
					temp := &iwv2.InteractiveWorkload{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "temp",
							Namespace: namespaceName1,
						},
					}
					policy, levelName, err := interactiveWorkloadMergeCreator.GetPolicy(context.Background(), temp)
					Expect(err).ToNot(HaveOccurred())
					Expect(policy).To(Equal(&basePolicyWithoutLevel))
					Expect(levelName).To(Equal("project:temp"))
				})

			})

		})

		Context("backend policy department level", func() {
			BeforeEach(func() {
				namespace1 = workloadproject.EmulateProjNs(namespaceName1, projectName1)
				projectP1 = &kaiorg.Project{
					ObjectMeta: metav1.ObjectMeta{
						Name: projectName1,
						Labels: map[string]string{
							"runai/department-id": "1",
							"runai/project-id":    "1",
						},
					},
				}
				departmentD1 = &v2alpha2.Department{
					ObjectMeta: metav1.ObjectMeta{
						Name:      departmentName1,
						Namespace: runaiv1.DefaultNamespace,
						Labels: map[string]string{
							"runai/department-id": "1",
						},
					},
				}

				os.Setenv(utils.EnvVarControlPlanePolicies, "true")
				k8sClient = client2.InitFakeClientWithObjectNoStatus(scheme, projectP1, departmentD1, namespace1)
				interactiveWorkloadMergeCreator = NewInteractiveWorkloadMergeCreator(k8sClient, true)

				// department level policy definition
				basePolicyWithoutLevel.Namespace = clusterCommonConstants.NamespaceRunAi
				basePolicyWithoutLevel.Labels = map[string]string{
					clusterCommonConstants.PolicyLabel:          "true",
					clusterCommonConstants.DepartmentScopeLabel: "1",
				}
				basePolicyWithoutLevel.ResourceVersion = ""
			})

			AfterEach(func() {
				os.Unsetenv(utils.EnvVarControlPlanePolicies)
			})

			It("should failed validate, gpu is not valid by policy department level", func() {
				err := k8sClient.Create(context.Background(), &basePolicyWithoutLevel)
				Expect(err).ToNot(HaveOccurred())
				tempWorkload := &iwv2.InteractiveWorkload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "temp",
						Namespace: namespaceName1,
					},
					Spec: ipv2.InteractiveSpec{
						ResourceAllocationFields: field_groups.ResourceAllocationFields{
							Gpu: flds.NewOptionalNumberField(ptr.To[float64](7)),
						},
					},
				}

				mergeStackMap, err := interactiveWorkloadMergeCreator.CreateMergeStack(context.TODO(), tempWorkload)
				Expect(err).ToNot(HaveOccurred()) // found policy
				mergeStack := mergeStackMap[api.SpecSubResource]
				err = mergeStack.MergeAndValidate(context.TODO(), enforcer.MergeAndValidateOptions{})
				Expect(err).To(HaveOccurred(), "policy -> merge")
				Expect(err.Error()).To(ContainSubstring("gpu.value: must be no greater than 6"))

			})

			It("should success validate, gpu is not valid by policy department level", func() {
				err := k8sClient.Create(context.Background(), &basePolicyWithoutLevel)
				Expect(err).ToNot(HaveOccurred())
				tempWorkload := &iwv2.InteractiveWorkload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "temp",
						Namespace: namespaceName1,
					},
					Spec: ipv2.InteractiveSpec{
						ResourceAllocationFields: field_groups.ResourceAllocationFields{
							Gpu: flds.NewOptionalNumberField(ptr.To[float64](5)),
						},
					},
				}

				mergeStackMap, err := interactiveWorkloadMergeCreator.CreateMergeStack(context.TODO(), tempWorkload)
				Expect(err).ToNot(HaveOccurred())
				mergeStack := mergeStackMap[api.SpecSubResource]
				err = mergeStack.MergeAndValidate(context.TODO(), enforcer.MergeAndValidateOptions{})
				Expect(err).ToNot(HaveOccurred())
			})
			Context("GetPolicy", func() {
				BeforeEach(func() {
					err := k8sClient.Create(context.Background(), &basePolicyWithoutLevel)
					Expect(err).ToNot(HaveOccurred())
				})
				It("should success get policy", func() {
					temp := &iwv2.InteractiveWorkload{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "temp",
							Namespace: namespaceName1,
						},
					}
					policy, levelName, err := interactiveWorkloadMergeCreator.GetPolicy(context.Background(), temp)
					Expect(err).ToNot(HaveOccurred())
					Expect(policy).To(Equal(&basePolicyWithoutLevel))
					Expect(levelName).To(Equal("department:temp"))
				})
			})
		})

		Context("backend policy cluster level", func() {
			BeforeEach(func() {
				namespace1 = workloadproject.EmulateProjNs(namespaceName1, projectName1)
				projectP1 = &kaiorg.Project{
					ObjectMeta: metav1.ObjectMeta{
						Name: projectName1,
						Labels: map[string]string{
							"runai/department-id": "1",
							"runai/project-id":    "1",
						},
					},
				}
				departmentD1 = &v2alpha2.Department{
					ObjectMeta: metav1.ObjectMeta{
						Name:      departmentName1,
						Namespace: runaiv1.DefaultNamespace,
						Labels: map[string]string{
							"runai/department-id": "1",
						},
					},
				}

				os.Setenv(utils.EnvVarControlPlanePolicies, "true")
				k8sClient = client2.InitFakeClientWithObjectNoStatus(scheme, projectP1, departmentD1, namespace1)
				interactiveWorkloadMergeCreator = NewInteractiveWorkloadMergeCreator(k8sClient, true)

				// cluster level policy definition
				basePolicyWithoutLevel.Namespace = clusterCommonConstants.NamespaceRunAi
				basePolicyWithoutLevel.Labels = map[string]string{
					clusterCommonConstants.PolicyLabel:              "true",
					clusterCommonConstants.ClusterWideResourceLabel: "true",
				}
				basePolicyWithoutLevel.ResourceVersion = ""
			})

			AfterEach(func() {
				os.Unsetenv(utils.EnvVarControlPlanePolicies)
			})

			It("should failed validate, gpu is not valid by policy cluster level", func() {
				err := k8sClient.Create(context.Background(), &basePolicyWithoutLevel)
				Expect(err).ToNot(HaveOccurred())
				tempWorkload := &iwv2.InteractiveWorkload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "temp",
						Namespace: namespaceName1,
					},
					Spec: ipv2.InteractiveSpec{
						ResourceAllocationFields: field_groups.ResourceAllocationFields{
							Gpu: flds.NewOptionalNumberField(ptr.To[float64](7)),
						},
					},
				}

				mergeStackMap, err := interactiveWorkloadMergeCreator.CreateMergeStack(context.TODO(), tempWorkload)
				Expect(err).ToNot(HaveOccurred())
				mergeStack := mergeStackMap[api.SpecSubResource]
				err = mergeStack.MergeAndValidate(context.TODO(), enforcer.MergeAndValidateOptions{})
				Expect(err).To(HaveOccurred(), "policy -> merge")
				Expect(err.Error()).To(ContainSubstring("gpu.value: must be no greater than 6"))

			})

			It("should success validate, gpu is not valid by policy cluster level", func() {
				err := k8sClient.Create(context.Background(), &basePolicyWithoutLevel)
				Expect(err).ToNot(HaveOccurred())
				tempWorkload := &iwv2.InteractiveWorkload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "temp",
						Namespace: namespaceName1,
					},
					Spec: ipv2.InteractiveSpec{
						ResourceAllocationFields: field_groups.ResourceAllocationFields{
							Gpu: flds.NewOptionalNumberField(ptr.To[float64](5)),
						},
					},
				}

				mergeStackMap, err := interactiveWorkloadMergeCreator.CreateMergeStack(context.TODO(), tempWorkload)

				Expect(err).ToNot(HaveOccurred())
				mergeStack := mergeStackMap[api.SpecSubResource]
				err = mergeStack.MergeAndValidate(context.TODO(), enforcer.MergeAndValidateOptions{})
				Expect(err).ToNot(HaveOccurred())
			})
			Context("GetPolicy", func() {
				BeforeEach(func() {
					err := k8sClient.Create(context.Background(), &basePolicyWithoutLevel)
					Expect(err).ToNot(HaveOccurred())
				})
				It("should success get policy", func() {
					temp := &iwv2.InteractiveWorkload{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "temp",
							Namespace: namespaceName1,
						},
					}
					policy, levelName, err := interactiveWorkloadMergeCreator.GetPolicy(context.Background(), temp)
					Expect(err).ToNot(HaveOccurred())
					Expect(policy).To(Equal(&basePolicyWithoutLevel))
					Expect(levelName).To(Equal("cluster:temp"))
				})
			})

		})

		Context("namespace is empty", func() {
			Context("GetPolicy", func() {
				var (
					temp *iwv2.InteractiveWorkload
				)
				BeforeEach(func() {
					temp = &iwv2.InteractiveWorkload{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "temp",
							Namespace: "",
						},
					}
				})
				It("should return empty policy", func() {
					policy, levelName, err := interactiveWorkloadMergeCreator.GetPolicy(context.Background(), temp)
					Expect(err).ToNot(HaveOccurred())
					Expect(policy).To(Equal(ipv2.NewEmptyInteractivePolicy()))
					Expect(levelName).To(Equal(""))
				})
			})
		})

		Context("useControlPlanePolicies effect on policy usage", func() {
			BeforeEach(func() {
				namespace1 = workloadproject.EmulateProjNs(namespaceName1, projectName1)
				projectP1 = &kaiorg.Project{
					ObjectMeta: metav1.ObjectMeta{
						Name: projectName1,
						Labels: map[string]string{
							"runai/department-id": "1",
							"runai/project-id":    "1",
						},
					},
				}

				os.Setenv(utils.EnvVarControlPlanePolicies, "true")
				k8sClient = client2.InitFakeClientWithObjectNoStatus(scheme, projectP1, namespace1)

				interactiveWorkloadMergeCreator = NewInteractiveWorkloadMergeCreator(k8sClient, true)
				// new policy (With label) setting gpu max to 1
				basePolicyWithoutLevel.Spec.ResourceAllocationFields.Gpu = &flds.NumberField{
					Value: ptr.To("0"),
					Rules: &flds.NumberRules{
						Max: ptr.To("1"),
					},
				}

				basePolicyWithoutLevel.Name = "new-policy"
				basePolicyWithoutLevel.Namespace = clusterCommonConstants.NamespaceRunAi
				basePolicyWithoutLevel.Labels = map[string]string{
					clusterCommonConstants.PolicyLabel:              "true",
					clusterCommonConstants.ClusterWideResourceLabel: "true",
				}
				basePolicyWithoutLevel.ResourceVersion = ""

				err := k8sClient.Create(context.Background(), &basePolicyWithoutLevel)
				Expect(err).ToNot(HaveOccurred())

				// old policy (without label) setting gpu max to 2
				basePolicyWithoutLevel.Spec.ResourceAllocationFields.Gpu = &flds.NumberField{
					Value: ptr.To("0"),
					Rules: &flds.NumberRules{
						Max: ptr.To("2"),
					},
				}
				basePolicyWithoutLevel.Name = "old-policy"
				basePolicyWithoutLevel.Namespace = clusterCommonConstants.NamespaceRunAi
				basePolicyWithoutLevel.Labels = map[string]string{
					clusterCommonConstants.ClusterWideResourceLabel: "true",
				}
				basePolicyWithoutLevel.ResourceVersion = ""

				err = k8sClient.Create(context.Background(), &basePolicyWithoutLevel)
				Expect(err).ToNot(HaveOccurred())

			})

			AfterEach(func() {
				os.Unsetenv(utils.EnvVarControlPlanePolicies)
			})

			It("expecting new policy to have effect over the workload when useControlPlanePolicies is set to 'true'", func() {
				tempWorkload := &iwv2.InteractiveWorkload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "temp",
						Namespace: namespaceName1,
					},
					Spec: ipv2.InteractiveSpec{
						ResourceAllocationFields: field_groups.ResourceAllocationFields{
							Gpu: flds.NewOptionalNumberField(ptr.To[float64](5)),
						},
					},
				}

				mergeStackMap, err := interactiveWorkloadMergeCreator.CreateMergeStack(context.TODO(), tempWorkload)
				mergeStack := mergeStackMap[api.SpecSubResource]
				Expect(err).ToNot(HaveOccurred())
				err = mergeStack.MergeAndValidate(context.TODO(), enforcer.MergeAndValidateOptions{})
				Expect(err).To(HaveOccurred(), "policy -> merge")
				// Checking that the new policy rule is in effect
				Expect(err.Error()).To(ContainSubstring("gpu.value: must be no greater than 1"))

			})

			It("expecting old policy to have effect over the workload when useControlPlanePolicies is set to 'false'", func() {
				interactiveWorkloadMergeCreator = NewInteractiveWorkloadMergeCreator(k8sClient, false)

				tempWorkload := &iwv2.InteractiveWorkload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "temp",
						Namespace: namespaceName1,
					},
					Spec: ipv2.InteractiveSpec{
						ResourceAllocationFields: field_groups.ResourceAllocationFields{
							Gpu: flds.NewOptionalNumberField(ptr.To[float64](5)),
						},
					},
				}

				mergeStackMap, err := interactiveWorkloadMergeCreator.CreateMergeStack(context.TODO(), tempWorkload)
				mergeStack := mergeStackMap[api.SpecSubResource]
				Expect(err).ToNot(HaveOccurred())
				err = mergeStack.MergeAndValidate(context.TODO(), enforcer.MergeAndValidateOptions{})
				Expect(err).To(HaveOccurred(), "policy -> merge")
				// Checking that the old policy rule is in effect
				Expect(err.Error()).To(ContainSubstring("gpu.value: must be no greater than 2"))

			})

		})

	})

})
