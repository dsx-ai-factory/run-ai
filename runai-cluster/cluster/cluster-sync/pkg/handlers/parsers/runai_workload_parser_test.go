/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/test"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/test/fakecachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	field_types "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	dinfp "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferencepolicy/v1alpha1"
	dinfwv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	dp "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	diwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	infp "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	infwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	twv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testDeploymentKind = "Deployment"
)

var _ = Describe("Parsers", func() {
	Describe("Workload parser", func() {
		Context("parseWorkloadRequestedResources - native IW/TW fractional GPU scaling", func() {
			// Native IW/TW leave both spec.gpu (per-device fraction) and
			// spec.gpuDevices set on the CR (unlike distributed, which unifies them
			// upstream). The parser must therefore scale the reported GPU request
			// by the device count so the UI reflects the total workload GPU need.
			It("scales fractional Gpu by GpuDevices when both are set (0.5 x 3 = 1.5)", func() {
				raf := &field_groups.ResourceAllocationFields{
					Gpu:        &field_types.NumberField{Value: ptr.To("0.5")},
					GpuDevices: &field_types.IntField{Value: ptr.To(3)},
				}
				res := parseWorkloadRequestedResources(raf)
				Expect(res.IsSet()).To(BeTrue())
				gpu := res.Get().Gpu.Get()
				Expect(gpu).ToNot(BeNil())
				Expect(gpu.Request.IsSet()).To(BeTrue())
				Expect(*gpu.Request.Get()).To(BeEquivalentTo(1.5))
			})

			It("does not scale when GpuDevices <= 1 (single device)", func() {
				raf := &field_groups.ResourceAllocationFields{
					Gpu:        &field_types.NumberField{Value: ptr.To("0.5")},
					GpuDevices: &field_types.IntField{Value: ptr.To(1)},
				}
				res := parseWorkloadRequestedResources(raf)
				gpu := res.Get().Gpu.Get()
				Expect(gpu).ToNot(BeNil())
				Expect(*gpu.Request.Get()).To(BeEquivalentTo(0.5))
			})

			It("scales gpu limit alongside gpu request", func() {
				raf := &field_groups.ResourceAllocationFields{
					Gpu:        &field_types.NumberField{Value: ptr.To("0.5")},
					GpuLimit:   &field_types.NumberField{Value: ptr.To("0.75")},
					GpuDevices: &field_types.IntField{Value: ptr.To(3)},
				}
				res := parseWorkloadRequestedResources(raf)
				gpu := res.Get().Gpu.Get()
				Expect(gpu).ToNot(BeNil())
				Expect(*gpu.Request.Get()).To(BeEquivalentTo(1.5))
				Expect(*gpu.Limit.Get()).To(BeEquivalentTo(2.25))
			})

			It("treats Gpu=1 as per-device fraction and scales by device count", func() {
				// portion=1 with multiple devices means 'a whole GPU on each of N devices'.
				raf := &field_groups.ResourceAllocationFields{
					Gpu:        &field_types.NumberField{Value: ptr.To("1")},
					GpuDevices: &field_types.IntField{Value: ptr.To(3)},
				}
				res := parseWorkloadRequestedResources(raf)
				gpu := res.Get().Gpu.Get()
				Expect(gpu).ToNot(BeNil())
				Expect(*gpu.Request.Get()).To(BeEquivalentTo(3.0))
			})

			It("does not scale legacy whole-GPU values (Gpu > 1) even when GpuDevices is set", func() {
				// Backwards compatibility: Gpu > 1 historically represented the total
				// number of whole GPUs directly, not a per-device fraction.
				raf := &field_groups.ResourceAllocationFields{
					Gpu:        &field_types.NumberField{Value: ptr.To("2")},
					GpuDevices: &field_types.IntField{Value: ptr.To(3)},
				}
				res := parseWorkloadRequestedResources(raf)
				gpu := res.Get().Gpu.Get()
				Expect(gpu).ToNot(BeNil())
				Expect(*gpu.Request.Get()).To(BeEquivalentTo(2.0))
			})

			It("uses GpuDevices as the request when Gpu is unset (legacy branch)", func() {
				raf := &field_groups.ResourceAllocationFields{
					GpuDevices: &field_types.IntField{Value: ptr.To(3)},
				}
				res := parseWorkloadRequestedResources(raf)
				gpu := res.Get().Gpu.Get()
				Expect(gpu).ToNot(BeNil())
				Expect(*gpu.Request.Get()).To(BeEquivalentTo(3.0))
			})
		})

		Context("Distributed Workload", func() {
			It("parse DistributedWorkload requested resources", func() {
				dw := &diwv2.DistributedWorkload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dw-1",
						Namespace: "runai-project-a",
					},
					Spec: dp.DistributedSpec{
						ResourceAllocationFields: field_groups.ResourceAllocationFields{
							Gpu: &field_types.NumberField{
								Value: ptr.To("0.5"),
							},
							GpuDevices: &field_types.IntField{
								Value: ptr.To(1),
							},
							MemoryLimit: &field_types.QuantityField{
								Value: ptr.To("1G"),
							},
							ExtendedResources: &field_types.ExtendedResourcesField{
								Value: &map[string]field_types.ExtendedResourcesItemAndRules{
									"blabla1": {
										Value: &field_types.ExtendedResourcesItemValue{
											Name:     ptr.To("blabla1"),
											Quantity: ptr.To("100M"),
										},
									},
									"blabla2": {
										Value: &field_types.ExtendedResourcesItemValue{
											Name:     ptr.To("blabla2"),
											Quantity: ptr.To("500M"),
										},
									},
								},
							},
						},
						DistributedFields: field_groups.DistributedFields{
							Workers: &field_types.IntField{
								Value: ptr.To(3),
							},
							NoMaster: &field_types.BoolField{
								Value: ptr.To(false),
							},
						},
					},
					MasterSpec: &dp.DistributedMasterSpec{
						ResourceAllocationFields: field_groups.ResourceAllocationFields{
							GpuDevices: &field_types.IntField{
								Value: ptr.To(2),
							},
							MemoryLimit: &field_types.QuantityField{
								Value: ptr.To("10G"),
							},
							ExtendedResources: &field_types.ExtendedResourcesField{
								Value: &map[string]field_types.ExtendedResourcesItemAndRules{
									"blabla1": {
										Value: &field_types.ExtendedResourcesItemValue{
											Name:     ptr.To("blabla1"),
											Quantity: ptr.To("1G"),
										},
									},
									"blabla3": {
										Value: &field_types.ExtendedResourcesItemValue{
											Name:     ptr.To("blabla3"),
											Quantity: ptr.To("1G"),
										},
									},
								},
							},
						},
					},
				}

				result := &openapi.WorkloadSyncFields{
					RequestedResources: parseWorkloadRequestedResources(&dw.Spec.ResourceAllocationFields),
				}

				setDistributedNumOfRequestedPods(result, dw)
				setRecalculatedRequestedResourcesIfNeeded(result, dw)

				r := result.RequestedResources.Get()
				Expect(r).ToNot(BeNil())
				gpu := r.Gpu.Get()
				Expect(gpu).ToNot(BeNil())
				gpuRequest := gpu.Request.Get()
				Expect(gpuRequest).ToNot(BeNil())
				Expect(*gpuRequest).To(BeEquivalentTo(3.5))

				cpuMemory := r.CpuMemory.Get()
				Expect(cpuMemory).ToNot(BeNil())
				cpuMemoryLimit := cpuMemory.Limit.Get()
				Expect(cpuMemoryLimit).ToNot(BeNil())
				Expect(*cpuMemoryLimit).To(Equal("13G"))

				Expect(len(r.ExtendedResources)).To(Equal(3))
				Expect(r.ExtendedResources).To(ConsistOf(
					openapi.WorkloadsExtendedResource{
						Resource: *openapi.NewNullableString(ptr.To("blabla1")),
						Quantity: *openapi.NewNullableString(ptr.To("1300M")),
					},
					openapi.WorkloadsExtendedResource{
						Resource: *openapi.NewNullableString(ptr.To("blabla2")),
						Quantity: *openapi.NewNullableString(ptr.To("1500M")),
					},
					openapi.WorkloadsExtendedResource{
						Resource: *openapi.NewNullableString(ptr.To("blabla3")),
						Quantity: *openapi.NewNullableString(ptr.To("1G")),
					},
				))
				Expect(result.RequestedPods).ToNot(BeNil())
				rp := result.RequestedPods.Number.Get()
				Expect(rp).ToNot(BeNil())
				Expect(*rp).To(BeEquivalentTo(4))
			})

			Context("workload GPU accelerated present in additionalFields", func() {
				var (
					dw                  *diwv2.DistributedWorkload
					workloadFieldParser *WorkloadFieldsForParser
				)

				BeforeEach(func() {
					dw = &diwv2.DistributedWorkload{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "dw-1",
							Namespace: "runai-project-a",
						},
					}

					workloadFieldParser = &WorkloadFieldsForParser{
						Kind:            DistributedDisplayType,
						ObjectMeta:      dw.ObjectMeta,
						ContainerFields: dw.Spec.ContainerFields,
						CommonStatus:    &dw.Status.WorkloadCommonStatus,
					}
				})

				It("No condition", func() {
					result := getRunaiWorkloadAdditionalFields(workloadFieldParser)
					Expect(result).ToNot(BeNil())
					Expect(*result).To(MatchJSON(`{"GPUNetworkAccelerationStatus":"{\"used\":false}"}`))
				})

				It("Condition is true", func() {
					workloadFieldParser.CommonStatus.Conditions = []common.WorkloadCondition{{Type: common.GPUNetworkAccelerated, Status: True}}
					result := getRunaiWorkloadAdditionalFields(workloadFieldParser)

					Expect(result).ToNot(BeNil())
					Expect(*result).To(MatchJSON(`{"GPUNetworkAccelerationStatus":"{\"used\":true}"}`))
				})

				It("Condition is false", func() {
					workloadFieldParser.CommonStatus.Conditions = []common.WorkloadCondition{{Type: common.GPUNetworkAccelerated, Status: False}}
					result := getRunaiWorkloadAdditionalFields(workloadFieldParser)

					Expect(result).ToNot(BeNil())
					Expect(*result).To(MatchJSON(`{"GPUNetworkAccelerationStatus":"{\"used\":false}"}`))
				})

				Context("workload NIM label present in additionalFields", func() {
					var (
						dw                  *diwv2.DistributedWorkload
						workloadFieldParser *WorkloadFieldsForParser
					)

					BeforeEach(func() {
						dw = &diwv2.DistributedWorkload{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "dw-1",
								Namespace: "runai-project-a",
							},
						}

						workloadFieldParser = &WorkloadFieldsForParser{
							Kind:       DistributedDisplayType,
							ObjectMeta: dw.ObjectMeta,
							WorkloadFields: field_groups.WorkloadFields{
								ContainerFields: field_groups.ContainerFields{},
							},
							ContainerFields: dw.Spec.ContainerFields,
							CommonStatus:    &dw.Status.WorkloadCommonStatus,
						}
					})

					It("NIM label is true", func() {
						workloadFieldParser.WorkloadFields.Labels = &field_types.DictionaryField{
							Value: &map[string]field_types.DictionaryItemAndRules{
								constants.LabelNIMWorkload: {
									Value: ptr.To("true"),
								},
							},
						}

						result := getRunaiWorkloadAdditionalFields(workloadFieldParser)

						Expect(result).ToNot(BeNil())
						Expect(*result).To(MatchJSON(`{"GPUNetworkAccelerationStatus":"{\"used\":false}","nimWorkload":"true"}`))
					})

					It("vLLM label is true", func() {
						workloadFieldParser.WorkloadFields.Labels = &field_types.DictionaryField{
							Value: &map[string]field_types.DictionaryItemAndRules{
								constants.LabelVLLMWorkload: {
									Value: ptr.To("true"),
								},
							},
						}

						result := getRunaiWorkloadAdditionalFields(workloadFieldParser)

						Expect(result).ToNot(BeNil())
						Expect(*result).To(MatchJSON(`{"GPUNetworkAccelerationStatus":"{\"used\":false}","vllmWorkload":"true"}`))
					})

					It("does not add vLLM marker for a non-canonical label value", func() {
						workloadFieldParser.WorkloadFields.Labels = &field_types.DictionaryField{
							Value: &map[string]field_types.DictionaryItemAndRules{
								constants.LabelVLLMWorkload: {
									Value: ptr.To("false"),
								},
							},
						}

						result := getRunaiWorkloadAdditionalFields(workloadFieldParser)

						Expect(result).ToNot(BeNil())
						Expect(*result).To(MatchJSON(`{"GPUNetworkAccelerationStatus":"{\"used\":false}"}`))
					})
				})
			})
		})

		Context("Training Workload", func() {
			It("workload without issues", func() {
				tw := &twv2.TrainingWorkload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tw-1",
						Namespace: "runai-project-a",
					},
				}

				workloadFieldParser := &WorkloadFieldsForParser{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: tw.Annotations,
					},
					ContainerFields: tw.Spec.ContainerFields,
					CommonStatus:    &tw.Status.WorkloadCommonStatus,
				}

				result := getRunaiWorkloadAdditionalFields(workloadFieldParser)
				Expect(result).To(BeNil())
			})

			It("workload issues are present in additionalFields", func() {
				const (
					networkDesc = "This is a serious network issue"
					computeDesc = "What is compute?"
				)

				tw := &twv2.TrainingWorkload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tw-1",
						Namespace: "runai-project-a",
					},
					Status: twv2.TrainingWorkloadStatus{
						WorkloadCommonStatus: common.WorkloadCommonStatus{
							Issues: &common.Issues{
								Network: []common.Issue{
									{
										Kind:        constants.KindIngress,
										Name:        "ingress-1",
										Description: networkDesc,
									},
								},
								Compute: []common.Issue{
									{
										Kind:        constants.KindRunaiJob,
										Name:        "job-1",
										Description: computeDesc,
									},
								},
							},
						},
					},
				}

				workloadFieldParser := &WorkloadFieldsForParser{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: tw.Annotations,
					},
					ContainerFields: tw.Spec.ContainerFields,
					CommonStatus:    &tw.Status.WorkloadCommonStatus,
				}

				result := getRunaiWorkloadAdditionalFields(workloadFieldParser)

				Expect(result).ToNot(BeNil())
				expectedJSON := `{"issues":"{\"network\":[{\"kind\":\"Ingress\",\"name\":\"ingress-1\",` +
					`\"description\":\"This is a serious network issue\"}],\"storage\":null,` +
					`\"compute\":[{\"kind\":\"RunaiJob\",\"name\":\"job-1\",` +
					`\"description\":\"What is compute?\"}],\"internal\":null}"}`
				Expect(*result).To(MatchJSON(expectedJSON))

				//make sure the result is deserializable
				issues := getIssuesFromAdditionalFieldsStr(*result)
				Expect(issues.Compute).To(HaveLen(1))
				Expect(issues.Compute[0].Description).To(Equal(computeDesc))
				Expect(issues.Network).To(HaveLen(1))
				Expect(issues.Network[0].Description).To(Equal(networkDesc))
				Expect(issues.Storage).To(HaveLen(0))
				Expect(issues.Internal).To(HaveLen(0))
			})

			Context("workload NIM label present in additionalFields", func() {
				var (
					tw                  *twv2.TrainingWorkload
					workloadFieldParser *WorkloadFieldsForParser
				)

				BeforeEach(func() {
					tw = &twv2.TrainingWorkload{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "tw-1",
							Namespace: "runai-project-a",
						},
					}

					workloadFieldParser = &WorkloadFieldsForParser{
						ObjectMeta: tw.ObjectMeta,
						WorkloadFields: field_groups.WorkloadFields{
							ContainerFields: field_groups.ContainerFields{},
						},
						ContainerFields: tw.Spec.ContainerFields,
					}
				})

				It("NIM label is true", func() {
					workloadFieldParser.WorkloadFields.Labels = &field_types.DictionaryField{
						Value: &map[string]field_types.DictionaryItemAndRules{
							constants.LabelNIMWorkload: {
								Value: ptr.To("true"),
							},
						},
					}

					result := getRunaiWorkloadAdditionalFields(workloadFieldParser)

					Expect(result).ToNot(BeNil())
					Expect(*result).To(MatchJSON(`{"nimWorkload":"true"}`))
				})

			})
		})

		Context("TrainingWorkload end-to-end", func() {
			It("resolves workload-type-config group when TypeMeta is empty (typed client-go Get)", func() {
				const (
					ns   = "runai-project-a"
					name = "tw-rego"
				)
				nsObj := test.BuildNamespace(ns, map[string]string{constants.RunaiQueueLabel: ns})
				project := test.BuildProjectWithDepId(ns, "proj-id", ns, "dep-id")
				kaiProject := test.BuildKaiProjectWithDepId(ns, "proj-id", ns, "dep-id")
				workloadsConfig := test.BuildWorkloadTypeConfigMap()
				fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(test.GetTestScheme()).WithObjects(
					workloadsConfig, nsObj, project, kaiProject)
				fakeCachedClient := fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
				parser := NewWorkloadParser(fakeCachedClient)

				// TypeMeta intentionally left empty — this is how the K8s typed cached client
				// delivers the object in production, which is the condition that produced the
				// "workload configuration not found for groupKind <TrainingWorkload>" error.
				tw := &twv2.TrainingWorkload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: ns,
						UID:       "uid-tw-rego",
					},
				}

				result, _, err := parser.ParseTrainingWorkloadToServerScheme(tw)
				Expect(err).To(BeNil())
				Expect(result).ToNot(BeNil())

				Expect(result.Group).ToNot(BeNil())
				Expect(*result.Group).To(Equal("run.ai"))
				Expect(result.Type).To(Equal(TrainingDisplayType))
				// "low" is the PriorityName for TrainingWorkload in test.BuildWorkloadTypeConfigMap().
				// Pre-fix, the lookup failed and this fell back to priority.TrainPriorityClassName ("train").
				Expect(result.PriorityClassName).To(Equal("low"))
			})
		})

		Context("Topology annotations", func() {
			It("Runai workload with topology annotations", func() {
				const (
					ns   = "runai-project-a"
					name = "tw-1"
				)
				annotations := map[string]string{
					constants.TopologyAnnotation:                   "net-topo-test",
					constants.TopologyPreferredPlacementAnnotation: "pref-json",
					constants.TopologyRequiredPlacementAnnotation:  "req-json",
				}

				// Environment: namespace, project and workload type config
				nsObj := test.BuildNamespace(ns, map[string]string{constants.RunaiQueueLabel: ns})
				project := test.BuildProjectWithDepId(ns, "proj-id", ns, "dep-id")
				kaiProject := test.BuildKaiProjectWithDepId(ns, "proj-id", ns, "dep-id")
				workloadsConfig := test.BuildWorkloadTypeConfigMap()
				fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(test.GetTestScheme()).WithObjects(
					workloadsConfig, nsObj, project, kaiProject)
				fakeCachedClient := fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
				parser := NewWorkloadParser(fakeCachedClient)

				// Minimal workload fields for parsing
				wf := &WorkloadFieldsForParser{
					Kind: "Distributed",
					TypeMeta: metav1.TypeMeta{
						Kind:       "DistributedWorkload",
						APIVersion: twv2.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:        name,
						Namespace:   ns,
						UID:         "uid-tw",
						Annotations: annotations,
					},
					WorkloadFields:  field_groups.WorkloadFields{},
					ContainerFields: field_groups.ContainerFields{},
				}

				result, _, err := parser.parseJobWorkloadToServerScheme(wf)
				Expect(err).To(BeNil())
				Expect(result).ToNot(BeNil())

				topology, ok := result.GetTopologyOk()
				Expect(ok).To(BeTrue())

				nameVal, nameOk := topology.GetNameOk()
				prefVal, prefOk := topology.GetPreferredPlacementOk()
				reqVal, reqOk := topology.GetRequiredPlacementOk()

				Expect(nameOk).To(BeTrue())
				Expect(prefOk).To(BeTrue())
				Expect(reqOk).To(BeTrue())

				Expect(*nameVal).To(Equal("net-topo-test"))
				Expect(*prefVal).To(Equal("pref-json"))
				Expect(*reqVal).To(Equal("req-json"))
			})

			It("does not set topology when no topology annotations exist", func() {
				const (
					ns   = "runai-project-b"
					name = "tw-2"
				)
				// Environment: namespace, project and workload type config
				nsObj := test.BuildNamespace(ns, map[string]string{constants.RunaiQueueLabel: ns})
				project := test.BuildProjectWithDepId(ns, "proj-id-2", ns, "dep-id-2")
				kaiProject := test.BuildKaiProjectWithDepId(ns, "proj-id-2", ns, "dep-id-2")
				workloadsConfig := test.BuildWorkloadTypeConfigMap()
				fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(test.GetTestScheme()).WithObjects(
					workloadsConfig, nsObj, project, kaiProject)
				fakeCachedClient := fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
				parser := NewWorkloadParser(fakeCachedClient)

				// Minimal workload fields without topology annotations
				wf := &WorkloadFieldsForParser{
					Kind: "DistributedInference",
					TypeMeta: metav1.TypeMeta{
						Kind:       "DistributedInferenceWorkload",
						APIVersion: twv2.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: ns,
						UID:       "uid-tw-2",
						// no annotations
					},
					WorkloadFields:  field_groups.WorkloadFields{},
					ContainerFields: field_groups.ContainerFields{},
				}

				result, _, err := parser.parseJobWorkloadToServerScheme(wf)
				Expect(err).To(BeNil())
				Expect(result).ToNot(BeNil())

				_, ok := result.GetTopologyOk()
				Expect(ok).To(BeFalse())
				Expect(result.HasTopology()).To(BeFalse())
			})
		})

		Context("Node affinity", func() {
			It("single-spec workload with node affinity has specName worker", func() {
				const (
					ns   = "runai-project-a"
					name = "tw-na"
				)
				nsObj := test.BuildNamespace(ns, map[string]string{constants.RunaiQueueLabel: ns})
				project := test.BuildProjectWithDepId(ns, "proj-id", ns, "dep-id")
				kaiProject := test.BuildKaiProjectWithDepId(ns, "proj-id", ns, "dep-id")
				workloadsConfig := test.BuildWorkloadTypeConfigMap()
				fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(test.GetTestScheme()).WithObjects(
					workloadsConfig, nsObj, project, kaiProject)
				fakeCachedClient := fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
				parser := NewWorkloadParser(fakeCachedClient)

				wf := &WorkloadFieldsForParser{
					Kind: TrainingDisplayType,
					TypeMeta: metav1.TypeMeta{
						Kind:       "TrainingWorkload",
						APIVersion: twv2.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: ns,
						UID:       "uid-na",
					},
					RunaiSchedulerFields: field_groups.RunaiSchedulerFields{
						NodeAffinityFields: field_types.NodeAffinityFields{
							NodeAffinityRequired: &field_types.NodeSelector{
								NodeSelectorTerms: []field_types.NodeSelectorTerm{{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{Key: "gpu-type", Operator: corev1.NodeSelectorOpIn, Values: []string{"a100"}},
									},
								}},
							},
						},
					},
					WorkloadFields:  field_groups.WorkloadFields{},
					ContainerFields: field_groups.ContainerFields{},
				}

				result, _, err := parser.parseJobWorkloadToServerScheme(wf)
				Expect(err).To(BeNil())
				Expect(result.NodeAffinities).To(HaveLen(1))
				Expect(result.NodeAffinities[0].GetSpecName()).To(Equal("worker"))
				req := result.NodeAffinities[0].NodeAffinity.Required.Get()
				Expect(req).ToNot(BeNil())
				Expect(req.NodeSelectorTerms).To(HaveLen(1))
				Expect(req.NodeSelectorTerms[0].MatchExpressions).To(HaveLen(1))
				Expect(req.NodeSelectorTerms[0].MatchExpressions[0].Key).To(Equal("gpu-type"))
			})

			It("workload without node affinity still has one entry with nil affinity", func() {
				const (
					ns   = "runai-project-a"
					name = "tw-no-na"
				)
				nsObj := test.BuildNamespace(ns, map[string]string{constants.RunaiQueueLabel: ns})
				project := test.BuildProjectWithDepId(ns, "proj-id", ns, "dep-id")
				kaiProject := test.BuildKaiProjectWithDepId(ns, "proj-id", ns, "dep-id")
				workloadsConfig := test.BuildWorkloadTypeConfigMap()
				fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(test.GetTestScheme()).WithObjects(
					workloadsConfig, nsObj, project, kaiProject)
				fakeCachedClient := fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
				parser := NewWorkloadParser(fakeCachedClient)

				wf := &WorkloadFieldsForParser{
					Kind: TrainingDisplayType,
					TypeMeta: metav1.TypeMeta{
						Kind:       "TrainingWorkload",
						APIVersion: twv2.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: ns,
						UID:       "uid-no-na",
					},
					WorkloadFields:  field_groups.WorkloadFields{},
					ContainerFields: field_groups.ContainerFields{},
				}

				result, _, err := parser.parseJobWorkloadToServerScheme(wf)
				Expect(err).To(BeNil())
				Expect(result.NodeAffinities).To(HaveLen(1))
				Expect(result.NodeAffinities[0].GetSpecName()).To(Equal("worker"))
				Expect(result.NodeAffinities[0].NodeAffinity).To(BeNil())
			})

			It("distributed workload has worker and master affinities", func() {
				workerAffinity := field_types.NodeAffinityFields{
					NodeAffinityRequired: &field_types.NodeSelector{
						NodeSelectorTerms: []field_types.NodeSelectorTerm{{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{Key: "gpu-type", Operator: corev1.NodeSelectorOpIn, Values: []string{"a100"}},
							},
						}},
					},
				}
				masterAffinity := field_types.NodeAffinityFields{
					NodeAffinityRequired: &field_types.NodeSelector{
						NodeSelectorTerms: []field_types.NodeSelectorTerm{{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{Key: "role", Operator: corev1.NodeSelectorOpIn, Values: []string{"master"}},
							},
						}},
					},
				}

				result := &openapi.WorkloadSyncFields{}
				dw := &diwv2.DistributedWorkload{
					Spec: dp.DistributedSpec{
						RunaiSchedulerFields: field_groups.RunaiSchedulerFields{
							NodeAffinityFields: workerAffinity,
						},
					},
					MasterSpec: &dp.DistributedMasterSpec{
						RunaiSchedulerFields: field_groups.RunaiSchedulerFields{
							NodeAffinityFields: masterAffinity,
						},
					},
				}

				setDistributedNodeAffinities(result, dw)

				Expect(result.NodeAffinities).To(HaveLen(2))
				Expect(result.NodeAffinities[0].GetSpecName()).To(Equal("worker"))
				Expect(result.NodeAffinities[0].NodeAffinity.Required.Get().NodeSelectorTerms[0].MatchExpressions[0].Key).To(Equal("gpu-type"))
				Expect(result.NodeAffinities[1].GetSpecName()).To(Equal("master"))
				Expect(result.NodeAffinities[1].NodeAffinity.Required.Get().NodeSelectorTerms[0].MatchExpressions[0].Key).To(Equal("role"))
			})

			It("distributed workload with only worker affinity (no master spec) still has both entries", func() {
				workerAffinity := field_types.NodeAffinityFields{
					NodeAffinityRequired: &field_types.NodeSelector{
						NodeSelectorTerms: []field_types.NodeSelectorTerm{{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{Key: "gpu-type", Operator: corev1.NodeSelectorOpIn, Values: []string{"h100"}},
							},
						}},
					},
				}

				result := &openapi.WorkloadSyncFields{}
				dw := &diwv2.DistributedWorkload{
					Spec: dp.DistributedSpec{
						RunaiSchedulerFields: field_groups.RunaiSchedulerFields{
							NodeAffinityFields: workerAffinity,
						},
					},
				}

				setDistributedNodeAffinities(result, dw)

				Expect(result.NodeAffinities).To(HaveLen(2))
				Expect(result.NodeAffinities[0].GetSpecName()).To(Equal("worker"))
				Expect(result.NodeAffinities[0].NodeAffinity).ToNot(BeNil())
				Expect(result.NodeAffinities[1].GetSpecName()).To(Equal("master"))
				Expect(result.NodeAffinities[1].NodeAffinity).To(BeNil())
			})

			It("distributed inference workload with only leader (no worker spec) still has both entries", func() {
				leaderAffinity := field_types.NodeAffinityFields{
					NodeAffinityRequired: &field_types.NodeSelector{
						NodeSelectorTerms: []field_types.NodeSelectorTerm{{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{Key: "gpu-type", Operator: corev1.NodeSelectorOpIn, Values: []string{"a100"}},
							},
						}},
					},
				}

				result := &openapi.WorkloadSyncFields{}
				dinfw := &dinfwv1.DistributedInferenceWorkload{
					LeaderSpec: dinfp.DistributedInferenceEntitySpec{
						DistributedInferenceEntitySchedulerFields: field_groups.DistributedInferenceEntitySchedulerFields{
							NodeAffinityFields: leaderAffinity,
						},
					},
				}

				setDistributedInferenceNodeAffinities(dinfw, result)

				Expect(result.NodeAffinities).To(HaveLen(2))
				Expect(result.NodeAffinities[0].GetSpecName()).To(Equal("leader"))
				Expect(result.NodeAffinities[0].NodeAffinity).ToNot(BeNil())
				Expect(result.NodeAffinities[1].GetSpecName()).To(Equal("worker"))
				Expect(result.NodeAffinities[1].NodeAffinity).To(BeNil())
			})

			It("distributed inference workload has leader and worker affinities", func() {
				leaderAffinity := field_types.NodeAffinityFields{
					NodeAffinityRequired: &field_types.NodeSelector{
						NodeSelectorTerms: []field_types.NodeSelectorTerm{{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{Key: "gpu-type", Operator: corev1.NodeSelectorOpIn, Values: []string{"a100"}},
							},
						}},
					},
				}
				workerAffinity := field_types.NodeAffinityFields{
					NodeAffinityRequired: &field_types.NodeSelector{
						NodeSelectorTerms: []field_types.NodeSelectorTerm{{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{Key: "gpu-type", Operator: corev1.NodeSelectorOpIn, Values: []string{"h100"}},
							},
						}},
					},
				}

				result := &openapi.WorkloadSyncFields{}
				dinfw := &dinfwv1.DistributedInferenceWorkload{
					LeaderSpec: dinfp.DistributedInferenceEntitySpec{
						DistributedInferenceEntitySchedulerFields: field_groups.DistributedInferenceEntitySchedulerFields{
							NodeAffinityFields: leaderAffinity,
						},
					},
					WorkerSpec: &dinfp.DistributedInferenceEntitySpec{
						DistributedInferenceEntitySchedulerFields: field_groups.DistributedInferenceEntitySchedulerFields{
							NodeAffinityFields: workerAffinity,
						},
					},
				}

				setDistributedInferenceNodeAffinities(dinfw, result)

				Expect(result.NodeAffinities).To(HaveLen(2))
				Expect(result.NodeAffinities[0].GetSpecName()).To(Equal("leader"))
				Expect(result.NodeAffinities[0].NodeAffinity.Required.Get().NodeSelectorTerms[0].MatchExpressions[0].Key).To(Equal("gpu-type"))
				Expect(result.NodeAffinities[0].NodeAffinity.Required.Get().NodeSelectorTerms[0].MatchExpressions[0].Values).To(Equal([]string{"a100"}))
				Expect(result.NodeAffinities[1].GetSpecName()).To(Equal("worker"))
				Expect(result.NodeAffinities[1].NodeAffinity.Required.Get().NodeSelectorTerms[0].MatchExpressions[0].Values).To(Equal([]string{"h100"}))
			})
		})

		Context("Inference Workload", func() {
			It("parse InferenceWorkload requested pods - with min and max", func() {
				infw := &infwv2.InferenceWorkload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "infw-1",
						Namespace: "runai-project-a",
					},
					Spec: infp.InferenceSpec{
						InferenceFields: field_groups.InferenceFields{
							AutoScalingFields: field_groups.AutoScalingFields{
								MinScale: &field_types.IntField{
									Value: ptr.To(2),
								},
								MaxScale: &field_types.IntField{
									Value: ptr.To(3),
								},
							},
						},
					},
				}

				result := &openapi.WorkloadSyncFields{
					RequestedPods: parseInferenceRequestedPods(infw),
				}

				Expect(result.RequestedPods).ToNot(BeNil())
				Expect(result.RequestedPods.Number.IsSet()).To(BeFalse())
				Expect(result.RequestedPods.Min.IsSet()).To(BeTrue())
				Expect(result.RequestedPods.Max.IsSet()).To(BeTrue())
				Expect(*result.RequestedPods.Min.Get()).To(BeEquivalentTo(2))
				Expect(*result.RequestedPods.Max.Get()).To(BeEquivalentTo(3))
			})

			It("parse InferenceWorkload requested pods - with min only", func() {
				infw := &infwv2.InferenceWorkload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "infw-1",
						Namespace: "runai-project-a",
					},
					Spec: infp.InferenceSpec{
						InferenceFields: field_groups.InferenceFields{
							AutoScalingFields: field_groups.AutoScalingFields{
								MinScale: &field_types.IntField{
									Value: ptr.To(2),
								},
							},
						},
					},
				}

				result := &openapi.WorkloadSyncFields{
					RequestedPods: parseInferenceRequestedPods(infw),
				}

				Expect(result.RequestedPods).ToNot(BeNil())
				Expect(result.RequestedPods.Number.IsSet()).To(BeFalse())
				Expect(result.RequestedPods.Min.IsSet()).To(BeTrue())
				Expect(result.RequestedPods.Max.IsSet()).To(BeFalse())
				Expect(*result.RequestedPods.Min.Get()).To(BeEquivalentTo(2))
			})

			It("parse InferenceWorkload requested pods - with max only, concurrency metric", func() {
				infw := &infwv2.InferenceWorkload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "infw-1",
						Namespace: "runai-project-a",
					},
					Spec: infp.InferenceSpec{
						InferenceFields: field_groups.InferenceFields{
							AutoScalingFields: field_groups.AutoScalingFields{
								MaxScale: &field_types.IntField{
									Value: ptr.To(3),
								},
								Metric: &field_types.StringField{
									Value: ptr.To("concurrency"),
								},
							},
						},
					},
				}

				result := &openapi.WorkloadSyncFields{
					RequestedPods: parseInferenceRequestedPods(infw),
				}

				Expect(result.RequestedPods).ToNot(BeNil())
				Expect(result.RequestedPods.Number.IsSet()).To(BeFalse())
				Expect(result.RequestedPods.Min.IsSet()).To(BeTrue())
				Expect(result.RequestedPods.Max.IsSet()).To(BeTrue())
				Expect(*result.RequestedPods.Min.Get()).To(BeEquivalentTo(0))
				Expect(*result.RequestedPods.Max.Get()).To(BeEquivalentTo(3))
			})

			It("parse InferenceWorkload requested pods - with max only, latency metric", func() {
				infw := &infwv2.InferenceWorkload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "infw-1",
						Namespace: "runai-project-a",
					},
					Spec: infp.InferenceSpec{
						InferenceFields: field_groups.InferenceFields{
							AutoScalingFields: field_groups.AutoScalingFields{
								MaxScale: &field_types.IntField{
									Value: ptr.To(3),
								},
								Metric: &field_types.StringField{
									Value: ptr.To("latency"),
								},
							},
						},
					},
				}

				result := &openapi.WorkloadSyncFields{
					RequestedPods: parseInferenceRequestedPods(infw),
				}

				Expect(result.RequestedPods).ToNot(BeNil())
				Expect(result.RequestedPods.Number.IsSet()).To(BeFalse())
				Expect(result.RequestedPods.Min.IsSet()).To(BeTrue())
				Expect(result.RequestedPods.Max.IsSet()).To(BeTrue())
				Expect(*result.RequestedPods.Min.Get()).To(BeEquivalentTo(1))
				Expect(*result.RequestedPods.Max.Get()).To(BeEquivalentTo(3))
			})

			It("parse InferenceWorkload phases", func() {
				Expect(getPhaseFromInferencePhase("Creating")).To(Equal("Creating"))
				Expect(getPhaseFromInferencePhase("Initializing")).To(Equal("Initializing"))
				Expect(getPhaseFromInferencePhase("Pending")).To(Equal("Pending"))
				Expect(getPhaseFromInferencePhase("Failed")).To(Equal("Failed"))
				Expect(getPhaseFromInferencePhase("Ready")).To(Equal("Running"))
				Expect(getPhaseFromInferencePhase("Progressing")).To(Equal("Running"))
			})

			It("parse InferenceWorkload revision names - latest created and ready", func() {
				infw := &infwv2.InferenceWorkload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "infw-1",
						Namespace: "runai-project-a",
					},
					Status: infwv2.InferenceWorkloadStatus{
						InferenceStatus: common.InferenceStatus{
							LatestReadyRevisionName:   ptr.To("infw-1-00001"),
							LatestCreatedRevisionName: ptr.To("infw-1-00002"),
						},
					},
				}

				workloadFieldParser := &WorkloadFieldsForParser{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: infw.Annotations,
					},
					ContainerFields: infw.Spec.ContainerFields,
					InferenceStatus: &infw.Status.InferenceStatus,
				}

				result := getRunaiWorkloadAdditionalFields(workloadFieldParser)

				Expect(result).ToNot(BeNil())
				Expect(*result).To(MatchJSON(`{"latestReadyRevisionName":"infw-1-00001","latestCreatedRevisionName":"infw-1-00002"}`))
			})

			It("parse InferenceWorkload revision names - only latest created", func() {
				infw := &infwv2.InferenceWorkload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "infw-1",
						Namespace: "runai-project-a",
					},
					Status: infwv2.InferenceWorkloadStatus{
						InferenceStatus: common.InferenceStatus{
							LatestCreatedRevisionName: ptr.To("infw-1-00001"),
						},
					},
				}

				workloadFieldParser := &WorkloadFieldsForParser{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: infw.Annotations,
					},
					ContainerFields: infw.Spec.ContainerFields,
					InferenceStatus: &infw.Status.InferenceStatus,
				}

				result := getRunaiWorkloadAdditionalFields(workloadFieldParser)

				Expect(result).ToNot(BeNil())
				Expect(*result).To(MatchJSON(`{"latestCreatedRevisionName":"infw-1-00001"}`))
			})

			It("parse InferenceWorkload children - with service, configuration and revision", func() {
				workloadName := "infw-1"
				workloadNamespace := "runai-project-a"

				infw := &infwv2.InferenceWorkload{
					TypeMeta: metav1.TypeMeta{
						Kind:       "InferenceWorkload",
						APIVersion: infwv2.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      workloadName,
						Namespace: workloadNamespace,
						UID:       "uid-infw",
					},
					Status: infwv2.InferenceWorkloadStatus{
						WorkloadDebugStatus: common.WorkloadDebugStatus{
							OwnedResources: []common.ObservedResource{
								{
									Resource: common.Resource{
										Kind: metav1.GroupVersionKind{
											Group:   "serving.knative.dev",
											Version: "v1",
											Kind:    "Service",
										},
										Name:      workloadName,
										Namespace: workloadNamespace,
									},
									Uid: "uid-svc",
								},
							},
						},
					},
				}

				existingConfiguration := test.BuildKnativeConfiguration(workloadName, workloadNamespace, "uid-conf")
				existingRevision1 := test.BuildKnativeRevision(workloadName, workloadNamespace, "00001", "uid-rev1")
				existingRevision2 := test.BuildKnativeRevision(workloadName, workloadNamespace, "00002", "uid-rev2")
				fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(test.GetTestScheme()).WithObjects(
					existingConfiguration, existingRevision1, existingRevision2)
				fakeCachedClient := fakecachedclient.NewFakeCachedClient(fakeClientBuilder)

				workloadParser := WorkloadParser{
					cachedReader: fakeCachedClient,
				}

				infwFields := &WorkloadFieldsForParser{
					Kind:        InferenceDisplayType,
					TypeMeta:    infw.TypeMeta,
					ObjectMeta:  infw.ObjectMeta,
					DebugStatus: infw.Status.WorkloadDebugStatus,
				}

				result, err := workloadParser.parseRunaiWorkloadChildIds(infwFields, nil, nil)
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(5))
				Expect(result[0].Id).To(Equal(string(infw.UID)))
				Expect(result[1].Id).To(Equal(string(infw.Status.WorkloadDebugStatus.OwnedResources[0].Uid)))
				Expect(result[2].Id).To(Equal(string(existingConfiguration.UID)))
				Expect(result[3].Id).To(Equal(string(existingRevision1.UID)))
				Expect(result[4].Id).To(Equal(string(existingRevision2.UID)))
			})

			It("parse InferenceWorkload Spec", func() {
				workloadName := "infw-1"
				workloadNamespace := "runai-project-a"
				preemptibility := field_groups.Preemptible

				infw := &infwv2.InferenceWorkload{
					TypeMeta: metav1.TypeMeta{
						Kind:       "InferenceWorkload",
						APIVersion: infwv2.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      workloadName,
						Namespace: workloadNamespace,
						UID:       "uid-infw",
						Labels: map[string]string{
							constants.WorkloadCategoryLabel: "not-existing",
						},
					},
					Spec: infp.InferenceSpec{
						SchedulingFields: field_groups.SchedulingFields{
							Preemptibility: &field_types.StringField{Value: &preemptibility},
						},
					},
					Status: infwv2.InferenceWorkloadStatus{
						WorkloadDebugStatus: common.WorkloadDebugStatus{
							OwnedResources: []common.ObservedResource{
								{
									Resource: common.Resource{
										Kind: metav1.GroupVersionKind{
											Group:   "serving.knative.dev",
											Version: "v1",
											Kind:    "Service",
										},
										Name:      workloadName,
										Namespace: workloadNamespace,
									},
									Uid: "uid-svc",
								},
							},
						},
					},
				}
				projectIdStr1 := "1"
				depId := "1"
				ns := test.BuildNamespace(workloadNamespace, map[string]string{constants.RunaiQueueLabel: workloadNamespace})
				project := test.BuildProjectWithDepId(workloadNamespace, projectIdStr1, workloadNamespace, depId)
				kaiProject := test.BuildKaiProjectWithDepId(workloadNamespace, projectIdStr1, workloadNamespace, depId)
				workloadsConfig := test.BuildWorkloadTypeConfigMap()
				fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(test.GetTestScheme()).WithObjects(
					workloadsConfig, ns, project, kaiProject)
				fakeCachedClient := fakecachedclient.NewFakeCachedClient(fakeClientBuilder)

				workloadParser := WorkloadParser{
					cachedReader: fakeCachedClient,
				}

				result, _, err := workloadParser.ParseInferenceWorkloadToServerScheme(infw)
				Expect(err).To(BeNil())
				Expect(result.Category).ToNot(BeNil())
				Expect(*result.Category).To(Equal("Deploy"))
				Expect(result.Preemptibility.Get()).To(Equal(&preemptibility))
			})

			Context("workload NIM label present in additionalFields", func() {
				var (
					infw                *infwv2.InferenceWorkload
					workloadFieldParser *WorkloadFieldsForParser
				)

				BeforeEach(func() {
					infw = &infwv2.InferenceWorkload{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "infw-1",
							Namespace: "runai-project-a",
						},
					}

					workloadFieldParser = &WorkloadFieldsForParser{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: infw.Annotations,
						},
						WorkloadFields: field_groups.WorkloadFields{
							ContainerFields: field_groups.ContainerFields{},
						},
						ContainerFields: infw.Spec.ContainerFields,
						InferenceStatus: &infw.Status.InferenceStatus,
					}
				})

				It("NIM label is true", func() {
					workloadFieldParser.WorkloadFields.Labels = &field_types.DictionaryField{
						Value: &map[string]field_types.DictionaryItemAndRules{
							constants.LabelNIMWorkload: {
								Value: ptr.To("true"),
							},
						},
					}

					result := getRunaiWorkloadAdditionalFields(workloadFieldParser)

					Expect(result).ToNot(BeNil())
					Expect(*result).To(MatchJSON(`{"nimWorkload":"true"}`))
				})

				It("NIM label is false", func() {
					workloadFieldParser.WorkloadFields.Labels = &field_types.DictionaryField{
						Value: &map[string]field_types.DictionaryItemAndRules{
							constants.LabelNIMWorkload: {
								Value: ptr.To("false"),
							},
						},
					}

					result := getRunaiWorkloadAdditionalFields(workloadFieldParser)

					Expect(result).To(BeNil())
				})

				It("NIM label is missing", func() {
					workloadFieldParser.WorkloadFields.Labels = &field_types.DictionaryField{
						Value: &map[string]field_types.DictionaryItemAndRules{
							"other-label": {
								Value: ptr.To("some-value"),
							},
						},
					}

					result := getRunaiWorkloadAdditionalFields(workloadFieldParser)

					Expect(result).To(BeNil())
				})

				It("No labels present", func() {
					workloadFieldParser.WorkloadFields.Labels = nil

					result := getRunaiWorkloadAdditionalFields(workloadFieldParser)

					Expect(result).To(BeNil())
				})

				It("NIM label value is nil", func() {
					workloadFieldParser.WorkloadFields.Labels = &field_types.DictionaryField{
						Value: &map[string]field_types.DictionaryItemAndRules{
							constants.LabelNIMWorkload: {
								Value: nil,
							},
						},
					}

					result := getRunaiWorkloadAdditionalFields(workloadFieldParser)

					Expect(result).To(BeNil())
				})

				It("NIM label value is empty string", func() {
					workloadFieldParser.WorkloadFields.Labels = &field_types.DictionaryField{
						Value: &map[string]field_types.DictionaryItemAndRules{
							constants.LabelNIMWorkload: {
								Value: ptr.To(""),
							},
						},
					}

					result := getRunaiWorkloadAdditionalFields(workloadFieldParser)

					Expect(result).To(BeNil())
				})

				It("NIM label has non-true value", func() {
					workloadFieldParser.WorkloadFields.Labels = &field_types.DictionaryField{
						Value: &map[string]field_types.DictionaryItemAndRules{
							constants.LabelNIMWorkload: {
								Value: ptr.To("yes"),
							},
						},
					}

					result := getRunaiWorkloadAdditionalFields(workloadFieldParser)

					Expect(result).To(BeNil())
				})
			})
		})

		Describe("External Workload", func() {
			Context("DistributedFramework", func() {
				It("uses the owner reference kind for v1-converted distributed workloads", func() {
					externalWorkload := &v1alpha1.ExternalWorkload{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "ew-pytorchjob-py-1",
							Namespace: "test-namespace",
							OwnerReferences: []metav1.OwnerReference{
								{
									APIVersion: "kubeflow.org/v1",
									Kind:       "PyTorchJob",
									Name:       "py-1",
								},
							},
						},
						Spec: v1alpha1.ExternalWorkloadSpec{
							Type: "Distributed",
						},
					}
					parser := &WorkloadParser{}

					Expect(parser.distributedFrameworkForExternalWorkload(externalWorkload, "Distributed")).To(Equal("PyTorch"))
				})
			})

			Context("Logical workload name", func() {
				DescribeTable("selects the read-model identity without changing ordinary workloads",
					func(workloadType string, labels map[string]string, expected string) {
						externalWorkload := &v1alpha1.ExternalWorkload{
							ObjectMeta: metav1.ObjectMeta{Labels: labels},
							Spec: v1alpha1.ExternalWorkloadSpec{
								Type:                 workloadType,
								OriginalWorkloadName: "my-workload-abcd",
							},
						}

						Expect(workloadNameForSync(externalWorkload)).To(Equal(expected))
					},
					Entry("falls back when the logical name is missing",
						"DistributedInference",
						map[string]string{commonconstants.ConvertedFromNativeLabel: "true"},
						"my-workload-abcd"),
					Entry("keeps the physical name for a directly submitted LWS",
						"DistributedInference",
						map[string]string{constants.LabelWorkloadName: "my-workload"},
						"my-workload-abcd"),
					Entry("keeps the physical name for other converted workload types",
						"Training",
						map[string]string{
							constants.LabelWorkloadName:              "my-workload",
							commonconstants.ConvertedFromNativeLabel: "true",
						},
						"my-workload-abcd"),
				)

				It("uses the original v1 name for a converted distributed inference LWS", func() {
					externalWorkload := &v1alpha1.ExternalWorkload{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "ew-leaderworkerset-my-workload-abcd",
							Namespace: "test-namespace",
							Labels: map[string]string{
								constants.WorkloadCategoryLabel:          "Deploy",
								constants.LabelWorkloadName:              "my-workload",
								commonconstants.ConvertedFromNativeLabel: "true",
							},
						},
						Spec: v1alpha1.ExternalWorkloadSpec{
							Type:                 "DistributedInference",
							Group:                constants.LwsGroup,
							Version:              constants.LwsVersion,
							OriginalWorkloadName: "my-workload-abcd",
							OriginalWorkloadUID:  "original-uid",
							ProjectName:          "test-project",
							ProjectId:            "project-id",
							DepartmentName:       "test-department",
							DepartmentId:         "department-id",
							CreationTimestamp:    metav1.Now(),
						},
					}

					namespaceObj := test.BuildNamespace("test-namespace", map[string]string{constants.RunaiQueueLabel: "test-project"})
					project := test.BuildProjectWithDepId("test-project", "project-id", "test-namespace", "department-id")
					kaiProject := test.BuildKaiProjectWithDepId("test-project", "project-id", "test-namespace", "department-id")
					workloadsConfig := test.BuildWorkloadTypeConfigMap()
					fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(test.GetTestScheme()).WithObjects(namespaceObj, project, kaiProject, workloadsConfig)
					parser := NewWorkloadParser(fakecachedclient.NewFakeCachedClient(fakeClientBuilder))

					result, _, _, err := parser.ParseExternalWorkloadToServerScheme(externalWorkload)

					Expect(err).NotTo(HaveOccurred())
					Expect(result.Name).To(Equal("my-workload"))
					Expect(externalWorkload.Spec.OriginalWorkloadName).To(Equal("my-workload-abcd"))
				})
			})

			Context("ClusterCrName", func() {
				It("should set ClusterCrName to external workload name", func() {
					// Create external workload
					externalWorkload := &v1alpha1.ExternalWorkload{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-external-workload",
							Namespace: "test-namespace",
							UID:       "test-uid",
						},
						Spec: v1alpha1.ExternalWorkloadSpec{
							Type:                 "Deployment",
							Group:                "apps",
							OriginalWorkloadName: "original-deployment",
							OriginalWorkloadUID:  "original-uid",
							ProjectName:          "test-project",
							ProjectId:            "project-id",
							DepartmentName:       "test-department",
							DepartmentId:         "department-id",
							CreationTimestamp:    metav1.Now(),
							RequestedResources:   corev1.ResourceRequirements{},
							RequestedPods:        v1alpha1.RequestedPods{},
							RequestedNodePools:   []string{},
							ImageName:            "test-image:latest",
							Urls:                 []string{},
							AdditionalFields:     nil,
						},
						Status: v1alpha1.ExternalWorkloadStatus{
							ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
								Phase:                v1alpha1.ExWRunningPhase,
								LastStatusChangeTime: metav1.Now(),
								ChildIds:             []v1alpha1.ChildId{},
								Conditions:           []v1alpha1.ObservedCondition{},
								Message:              "",
								K8sStatus:            "",
							},
						},
					}

					// Setup test environment
					namespaceObj := test.BuildNamespace("test-namespace", map[string]string{constants.RunaiQueueLabel: "test-project"})
					project := test.BuildProjectWithDepId("test-project", "project-id", "test-namespace", "department-id")
					kaiProject := test.BuildKaiProjectWithDepId("test-project", "project-id", "test-namespace", "department-id")
					workloadsConfig := test.BuildWorkloadTypeConfigMap()
					fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(test.GetTestScheme()).WithObjects(namespaceObj, project, kaiProject, workloadsConfig)
					fakeClient := fakecachedclient.NewFakeCachedClient(fakeClientBuilder)

					parser := NewWorkloadParser(fakeClient)

					// Parse the external workload
					result, _, _, err := parser.ParseExternalWorkloadToServerScheme(externalWorkload)

					// Assertions
					Expect(err).To(BeNil())
					Expect(result).ToNot(BeNil())
					Expect(result.ExternalWorkloadName).ToNot(BeNil())
					Expect(*result.ExternalWorkloadName).To(Equal("test-external-workload"))
					Expect(result.PriorityClassName).To(Equal("very-high"))
				})

				It("should set ClusterCrName correctly for different workload names", func() {
					// Create external workload with different name
					externalWorkload := &v1alpha1.ExternalWorkload{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "different-external-workload",
							Namespace: "test-namespace",
							UID:       "different-uid",
						},
						Spec: v1alpha1.ExternalWorkloadSpec{
							Type:                 "Deployment",
							OriginalWorkloadName: "original-deployment",
							OriginalWorkloadUID:  "original-uid",
							ProjectName:          "test-project",
							ProjectId:            "project-id",
							DepartmentName:       "test-department",
							DepartmentId:         "department-id",
							CreationTimestamp:    metav1.Now(),
							RequestedResources:   corev1.ResourceRequirements{},
							RequestedPods:        v1alpha1.RequestedPods{},
							RequestedNodePools:   []string{},
							ImageName:            "test-image:latest",
							Urls:                 []string{},
							AdditionalFields:     nil,
						},
						Status: v1alpha1.ExternalWorkloadStatus{
							ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
								Phase:                v1alpha1.ExWRunningPhase,
								LastStatusChangeTime: metav1.Now(),
								ChildIds:             []v1alpha1.ChildId{},
								Conditions:           []v1alpha1.ObservedCondition{},
								Message:              "",
								K8sStatus:            "",
							},
						},
					}

					// Setup test environment
					namespaceObj := test.BuildNamespace("test-namespace", map[string]string{constants.RunaiQueueLabel: "test-project"})
					project := test.BuildProjectWithDepId("test-project", "project-id", "test-namespace", "department-id")
					kaiProject := test.BuildKaiProjectWithDepId("test-project", "project-id", "test-namespace", "department-id")
					workloadsConfig := test.BuildWorkloadTypeConfigMap()
					fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(test.GetTestScheme()).WithObjects(namespaceObj, project, kaiProject, workloadsConfig)
					fakeClient := fakecachedclient.NewFakeCachedClient(fakeClientBuilder)

					parser := NewWorkloadParser(fakeClient)

					// Parse the external workload
					result, _, _, err := parser.ParseExternalWorkloadToServerScheme(externalWorkload)

					// Assertions
					Expect(err).To(BeNil())
					Expect(result).ToNot(BeNil())
					Expect(result.ExternalWorkloadName).ToNot(BeNil())
					Expect(*result.ExternalWorkloadName).To(Equal("different-external-workload"))
				})
			})

			Context("WorkloadRunner children", func() {
				It("registers the linked WorkloadRunner UID as a child so its events join", func() {
					const wlrName = "test-runner"
					wlr := &dromosv1alpha1.WorkloadRunner{
						ObjectMeta: metav1.ObjectMeta{
							Name:      wlrName,
							Namespace: "test-namespace",
							UID:       "runner-uid-123",
						},
						Status: dromosv1alpha1.WorkloadRunnerStatus{
							Phase: dromosv1alpha1.WorkloadRunnerPhaseRunning,
						},
					}

					externalWorkload := &v1alpha1.ExternalWorkload{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "runner-external-workload",
							Namespace: "test-namespace",
							UID:       "ew-uid",
							Labels: map[string]string{
								commonconstants.WorkloadRunnerNameLabel: wlrName,
							},
						},
						Spec: v1alpha1.ExternalWorkloadSpec{
							Type:                 "Job",
							OriginalWorkloadName: "original-job",
							OriginalWorkloadUID:  "original-uid",
							ProjectName:          "test-project",
							ProjectId:            "project-id",
							DepartmentName:       "test-department",
							DepartmentId:         "department-id",
							CreationTimestamp:    metav1.Now(),
							RequestedResources:   corev1.ResourceRequirements{},
							RequestedPods:        v1alpha1.RequestedPods{},
							RequestedNodePools:   []string{},
							Urls:                 []string{},
						},
						Status: v1alpha1.ExternalWorkloadStatus{
							ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
								Phase:                v1alpha1.ExWRunningPhase,
								LastStatusChangeTime: metav1.Now(),
								ChildIds:             []v1alpha1.ChildId{},
								Conditions:           []v1alpha1.ObservedCondition{},
							},
						},
					}

					namespaceObj := test.BuildNamespace("test-namespace", map[string]string{constants.RunaiQueueLabel: "test-project"})
					project := test.BuildProjectWithDepId("test-project", "project-id", "test-namespace", "department-id")
					kaiProject := test.BuildKaiProjectWithDepId("test-project", "project-id", "test-namespace", "department-id")
					workloadsConfig := test.BuildWorkloadTypeConfigMap()
					fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(test.GetTestScheme()).WithObjects(namespaceObj, project, kaiProject, workloadsConfig, wlr)
					fakeClient := fakecachedclient.NewFakeCachedClient(fakeClientBuilder)

					parser := NewWorkloadParser(fakeClient)

					result, _, _, err := parser.ParseExternalWorkloadToServerScheme(externalWorkload)
					Expect(err).To(BeNil())
					Expect(result).ToNot(BeNil())

					found := false
					for _, child := range result.ChildrenIds {
						if child.Id == "runner-uid-123" {
							Expect(child.Type).ToNot(BeNil())
							Expect(*child.Type).To(Equal(dromosv1alpha1.WorkloadRunnerKind))
							found = true
						}
					}
					Expect(found).To(BeTrue(), "WorkloadRunner UID should be registered as a child of the external workload")
				})

				It("links a WorkloadRunner event to the external workload via the runner UID", func() {
					const wlrName = "evented-runner"
					const runnerUID = "runner-uid-evt"
					wlr := &dromosv1alpha1.WorkloadRunner{
						ObjectMeta: metav1.ObjectMeta{
							Name:      wlrName,
							Namespace: "test-namespace",
							UID:       runnerUID,
						},
						Status: dromosv1alpha1.WorkloadRunnerStatus{
							Phase: dromosv1alpha1.WorkloadRunnerPhaseRunning,
						},
					}

					externalWorkload := &v1alpha1.ExternalWorkload{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "evented-external-workload",
							Namespace: "test-namespace",
							UID:       "ew-uid",
							Labels: map[string]string{
								commonconstants.WorkloadRunnerNameLabel: wlrName,
							},
						},
						Spec: v1alpha1.ExternalWorkloadSpec{
							Type:                 "Job",
							OriginalWorkloadName: "original-job",
							OriginalWorkloadUID:  "original-uid",
							ProjectName:          "test-project",
							ProjectId:            "project-id",
							DepartmentName:       "test-department",
							DepartmentId:         "department-id",
							CreationTimestamp:    metav1.Now(),
							RequestedResources:   corev1.ResourceRequirements{},
							RequestedPods:        v1alpha1.RequestedPods{},
							RequestedNodePools:   []string{},
							Urls:                 []string{},
						},
						Status: v1alpha1.ExternalWorkloadStatus{
							ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
								Phase:                v1alpha1.ExWRunningPhase,
								LastStatusChangeTime: metav1.Now(),
								ChildIds:             []v1alpha1.ChildId{},
								Conditions:           []v1alpha1.ObservedCondition{},
							},
						},
					}

					event := corev1.Event{
						ObjectMeta: metav1.ObjectMeta{Name: "timeout-event", Namespace: "test-namespace", UID: "event-uid"},
						Type:       "Normal",
						Reason:     "WorkloadTimeoutReached",
						Message:    "Stopping the workload.",
						InvolvedObject: corev1.ObjectReference{
							Kind:      dromosv1alpha1.WorkloadRunnerKind,
							Name:      wlrName,
							Namespace: "test-namespace",
							UID:       runnerUID,
						},
					}

					namespaceObj := test.BuildNamespace("test-namespace", map[string]string{constants.RunaiQueueLabel: "test-project"})
					project := test.BuildProjectWithDepId("test-project", "project-id", "test-namespace", "department-id")
					kaiProject := test.BuildKaiProjectWithDepId("test-project", "project-id", "test-namespace", "department-id")
					workloadsConfig := test.BuildWorkloadTypeConfigMap()
					fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(test.GetTestScheme()).WithObjects(namespaceObj, project, kaiProject, workloadsConfig, wlr)
					fakeClient := fakecachedclient.NewFakeCachedClient(fakeClientBuilder)

					parser := NewWorkloadParser(fakeClient)

					result, _, _, err := parser.ParseExternalWorkloadToServerScheme(externalWorkload)
					Expect(err).To(BeNil())
					Expect(result).ToNot(BeNil())

					syncReq := ParseEventsToSyncEventsRequest([]corev1.Event{event})
					Expect(syncReq.Events).To(HaveLen(1))
					eventInvolvedUID := syncReq.Events[0].InvolvedObject.Uid
					Expect(eventInvolvedUID).To(Equal(runnerUID))

					childUIDs := []string{}
					for _, child := range result.ChildrenIds {
						childUIDs = append(childUIDs, child.Id)
					}
					// The synced event joins to the external workload by this shared UID.
					Expect(childUIDs).To(ContainElement(eventInvolvedUID))
				})
			})

			Context("NIM label AdditionalFields", func() {
				var (
					externalWorkload     *v1alpha1.ExternalWorkload
					labels               map[string]string
					additionalFields     map[string]any
					externalWorkloadSpec v1alpha1.ExternalWorkloadSpec
				)

				BeforeEach(func() {
					externalWorkload = &v1alpha1.ExternalWorkload{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-external-workload",
							Namespace: "test-namespace",
							UID:       "test-uid",
						},
						Spec: v1alpha1.ExternalWorkloadSpec{
							Type:                 "Deployment",
							OriginalWorkloadName: "original-deployment",
							OriginalWorkloadUID:  "original-uid",
							ProjectName:          "test-project",
							ProjectId:            "project-id",
							DepartmentName:       "test-department",
							DepartmentId:         "department-id",
							CreationTimestamp:    metav1.Now(),
							RequestedResources:   corev1.ResourceRequirements{},
							RequestedPods:        v1alpha1.RequestedPods{},
							RequestedNodePools:   []string{},
							ImageName:            "test-image:latest",
							Urls:                 []string{},
							AdditionalFields:     nil,
						},
					}
					labels = make(map[string]string)
					additionalFields = make(map[string]any)
					externalWorkloadSpec = externalWorkload.Spec
				})

				It("should add nimWorkload flag when NIM label is true", func() {
					labels[constants.LabelNIMWorkload] = "true"
					externalWorkloadSpec.Type = testDeploymentKind

					TestAddNIMWorkloadFlagIfPresent(labels, externalWorkloadSpec.Type, additionalFields)

					Expect(additionalFields["nimWorkload"]).To(Equal("true"))
				})

				It("should add nimWorkload flag when external workload type is NIMService", func() {
					externalWorkloadSpec.Type = nimServiceType

					TestAddNIMWorkloadFlagIfPresent(labels, externalWorkloadSpec.Type, additionalFields)

					Expect(additionalFields["nimWorkload"]).To(Equal("true"))
				})

				It("should not add nimWorkload flag when NIM conditions are not met", func() {
					testCases := []struct {
						description string
						setupFunc   func()
					}{
						{
							description: "NIM label is false",
							setupFunc: func() {
								labels[constants.LabelNIMWorkload] = "false"
								externalWorkloadSpec.Type = testDeploymentKind
							},
						},
						{
							description: "NIM label is empty",
							setupFunc: func() {
								labels[constants.LabelNIMWorkload] = ""
								externalWorkloadSpec.Type = testDeploymentKind
							},
						},
						{
							description: "NIM label is missing",
							setupFunc: func() {
								// Don't set the label at all
								externalWorkloadSpec.Type = testDeploymentKind
							},
						},
					}

					for _, tc := range testCases {
						// Reset for each test case
						labels = make(map[string]string)
						additionalFields = make(map[string]any)

						tc.setupFunc()
						TestAddNIMWorkloadFlagIfPresent(labels, externalWorkloadSpec.Type, additionalFields)

						_, exists := additionalFields["nimWorkload"]
						Expect(exists).To(BeFalse(), "should not add nimWorkload flag when "+tc.description)
					}
				})

				It("should handle nil labels map gracefully", func() {
					externalWorkloadSpec.Type = testDeploymentKind

					TestAddNIMWorkloadFlagIfPresent(nil, externalWorkloadSpec.Type, additionalFields)

					_, exists := additionalFields["nimWorkload"]
					Expect(exists).To(BeFalse())
				})

				It("should work with getExternalWorkloadAdditionalFields integration when NIM label is true", func() {
					labels[constants.LabelNIMWorkload] = "true"
					externalWorkload.Labels = labels

					result := TestGetExternalWorkloadAdditionalFields(externalWorkloadSpec, labels, nil)

					Expect(result).ToNot(BeNil())
					Expect(*result).To(MatchJSON(`{"nimWorkload":"true"}`))
				})

				It("should work with getExternalWorkloadAdditionalFields integration when type is NIMService", func() {
					externalWorkloadSpec.Type = nimServiceType
					externalWorkload.Spec.Type = nimServiceType

					result := TestGetExternalWorkloadAdditionalFields(externalWorkloadSpec, labels, nil)

					Expect(result).ToNot(BeNil())
					Expect(*result).To(MatchJSON(`{"nimWorkload":"true"}`))
				})

				It("should preserve existing additional fields when adding NIM flag", func() {
					existingAdditionalFields := `{"existingField":"existingValue","anotherField":"anotherValue"}`
					externalWorkloadSpec.AdditionalFields = &existingAdditionalFields
					labels[constants.LabelNIMWorkload] = "true"

					result := TestGetExternalWorkloadAdditionalFields(externalWorkloadSpec, labels, nil)

					Expect(result).ToNot(BeNil())
					Expect(*result).To(MatchJSON(`{"existingField":"existingValue","anotherField":"anotherValue","nimWorkload":"true"}`))
				})

				It("should return nil when no additional fields and no NIM conditions", func() {
					externalWorkloadSpec.Type = testDeploymentKind

					result := TestGetExternalWorkloadAdditionalFields(externalWorkloadSpec, labels, nil)

					Expect(result).To(BeNil())
				})

				It("should handle various NIM label values correctly", func() {
					testCases := []struct {
						labelValue    string
						shouldAddFlag bool
						description   string
					}{
						{"true", true, "should add flag when label is 'true'"},
						{"TRUE", false, "should not add flag when label is 'TRUE' (case sensitive)"},
						{"True", false, "should not add flag when label is 'True' (case sensitive)"},
						{"false", false, "should not add flag when label is 'false'"},
						{"yes", false, "should not add flag when label is 'yes'"},
						{"1", false, "should not add flag when label is '1'"},
						{" true ", false, "should not add flag when label has whitespace"},
						{"", false, "should not add flag when label is empty"},
					}

					for _, tc := range testCases {
						// Reset additionalFields for each test case
						additionalFields = make(map[string]any)
						labels = map[string]string{constants.LabelNIMWorkload: tc.labelValue}
						externalWorkloadSpec.Type = testDeploymentKind

						TestAddNIMWorkloadFlagIfPresent(labels, externalWorkloadSpec.Type, additionalFields)

						if tc.shouldAddFlag {
							Expect(additionalFields["nimWorkload"]).To(Equal("true"), tc.description)
						} else {
							_, exists := additionalFields["nimWorkload"]
							Expect(exists).To(BeFalse(), tc.description)
						}
					}
				})

				It("should handle various external workload types correctly", func() {
					testCases := []struct {
						workloadType  string
						shouldAddFlag bool
						description   string
					}{
						{nimServiceType, true, "should add flag when type is 'NIMService'"},
						{"Deployment", false, "should not add flag when type is 'Deployment'"},
					}

					for _, tc := range testCases {
						// Reset additionalFields for each test case
						additionalFields = make(map[string]any)
						labels = make(map[string]string)

						TestAddNIMWorkloadFlagIfPresent(labels, tc.workloadType, additionalFields)

						if tc.shouldAddFlag {
							Expect(additionalFields["nimWorkload"]).To(Equal("true"), tc.description)
						} else {
							_, exists := additionalFields["nimWorkload"]
							Expect(exists).To(BeFalse(), tc.description)
						}
					}
				})
			})
		})
	})

	Context("toControlPlanePhase and getApiPhase", func() {
		var parser *WorkloadParser

		BeforeEach(func() {
			parser = NewWorkloadParser(nil)
		})

		It("maps 'Suspending' to PHASE_STOPPING", func() {
			phase := parser.toControlPlanePhase("Suspending")
			Expect(phase).To(Equal(openapi.PHASE_STOPPING))
		})

		It("maps 'Suspended' to PHASE_STOPPED", func() {
			phase := parser.toControlPlanePhase("Suspended")
			Expect(phase).To(Equal(openapi.PHASE_STOPPED))
		})

		It("passes through normal phase (e.g., 'Running')", func() {
			phase := parser.toControlPlanePhase("Running")
			Expect(phase).To(Equal(openapi.PHASE_RUNNING))
		})

		It("passes through non-exiting phase (e.g., 'Cyber')", func() {
			phase := parser.toControlPlanePhase("Cyber")
			Expect(phase).To(Equal(openapi.Phase("Cyber")))
		})

		It("getApiPhase returns PHASE_UNKNOWN for invalid phase", func() {
			phase := parser.getApiPhase("NotARealPhase")
			Expect(phase).To(Equal(openapi.PHASE_UNKNOWN))
		})

		It("getApiPhase returns correct value for allowed phase", func() {
			phase := parser.getApiPhase("Running")
			Expect(phase).To(Equal(openapi.PHASE_RUNNING))
		})

		It("getApiPhase returns PHASE_STOPPING for 'Suspending'", func() {
			phase := parser.getApiPhase("Suspending")
			Expect(phase).To(Equal(openapi.PHASE_STOPPING))
		})

		It("getApiPhase returns PHASE_STOPPED for 'Suspended'", func() {
			phase := parser.getApiPhase("Suspended")
			Expect(phase).To(Equal(openapi.PHASE_STOPPED))
		})
	})
})

func getIssuesFromAdditionalFieldsStr(additionalFieldsStr string) *common.Issues {
	var additionalFieldsMap map[string]any
	Expect(json.Unmarshal([]byte(additionalFieldsStr), &additionalFieldsMap))

	issuesStr := []byte(additionalFieldsMap[additionalIssues].(string))

	var issues common.Issues
	Expect(json.Unmarshal(issuesStr, &issues)).To(Succeed())

	return &issues
}

func TestParseDistributedInferenceWorkloadToServerScheme(t *testing.T) {
	// Test helpers for distributed inference workloads
	type ServingPortConfig struct {
		Port              int
		Protocol          string
		ExposeExternally  *bool
		AuthorizationType *string
	}

	type DistributedInferenceTestConfig struct {
		Name                string
		Namespace           string
		UID                 string
		Workers             int
		Replicas            int
		ReadyReplicas       int32
		LeaderImage         string
		WorkerImage         *string
		LeaderCommand       *string
		LeaderArguments     *string
		LeaderMemoryLimit   *string
		LeaderCpuLimit      *string
		WorkerMemoryLimit   *string
		WorkerCpuLimit      *string
		LeaderGpu           *string
		WorkerGpu           *string
		LeaderEnvironment   map[string]string
		WorkerEnvironment   map[string]string
		ServingPorts        map[string]ServingPortConfig
		LeaderLabels        map[string]string
		StartupPolicy       *string
		RestartPolicy       *string
		ExpectedPodCount    int32
		ExpectedImages      []string
		ExpectedCommand     *string
		ExpectedArguments   *string
		ExpectedEnvironment map[string]string
	}

	createDistributedInferenceWorkload := func(config DistributedInferenceTestConfig) *dinfwv1.DistributedInferenceWorkload {
		// Build serving ports
		var servingPorts *field_types.ServingPortsField
		if len(config.ServingPorts) > 0 {
			servingPortsMap := make(map[string]field_types.ServingPortItemAndRules)
			for name, portConfig := range config.ServingPorts {
				servingPortsMap[name] = field_types.ServingPortItemAndRules{
					Value: &field_types.ServingPortItemValue{
						Port:              ptr.To(portConfig.Port),
						Protocol:          ptr.To(portConfig.Protocol),
						ExposeExternally:  portConfig.ExposeExternally,
						AuthorizationType: portConfig.AuthorizationType,
					},
				}
			}
			servingPorts = &field_types.ServingPortsField{Value: &servingPortsMap}
		}

		// Build distributed inference fields
		distributedFields := field_groups.DistributedInferenceFields{
			Workers:      &field_types.IntField{Value: ptr.To(config.Workers)},
			Replicas:     &field_types.IntField{Value: ptr.To(config.Replicas)},
			ServingPorts: servingPorts,
		}
		if config.StartupPolicy != nil {
			distributedFields.StartupPolicy = &field_types.StringField{Value: config.StartupPolicy}
		}
		if config.RestartPolicy != nil {
			distributedFields.RestartPolicy = &field_types.StringField{Value: config.RestartPolicy}
		}

		// Build leader spec
		leaderSpec := dinfp.DistributedInferenceEntitySpec{
			ContainerFields: field_groups.ContainerFields{
				Image: &field_types.StringField{Value: ptr.To(config.LeaderImage)},
			},
		}
		if config.LeaderCommand != nil {
			leaderSpec.Command = &field_types.StringField{Value: config.LeaderCommand}
		}
		if config.LeaderArguments != nil {
			leaderSpec.Arguments = &field_types.StringField{Value: config.LeaderArguments}
		}
		if config.LeaderEnvironment != nil {
			envMap := make(map[string]field_types.DictionaryItemAndRules)
			for k, v := range config.LeaderEnvironment {
				envMap[k] = field_types.DictionaryItemAndRules{Value: ptr.To(v)}
			}
			leaderSpec.Environment = &field_types.DictionaryField{Value: &envMap}
		}
		if config.LeaderMemoryLimit != nil {
			leaderSpec.MemoryLimit = &field_types.QuantityField{Value: config.LeaderMemoryLimit}
		}
		if config.LeaderCpuLimit != nil {
			leaderSpec.CpuLimit = &field_types.NumberField{Value: config.LeaderCpuLimit}
		}
		if config.LeaderGpu != nil {
			leaderSpec.Gpu = &field_types.NumberField{Value: config.LeaderGpu}
		}
		if len(config.LeaderLabels) > 0 {
			// add here
			leaderSpec.Labels = &field_types.DictionaryField{Value: &map[string]field_types.DictionaryItemAndRules{}}
			for k, v := range config.LeaderLabels {
				(*leaderSpec.Labels.Value)[k] = field_types.DictionaryItemAndRules{Value: ptr.To(v)}
			}
		}
		// Build worker spec if needed
		var workerSpec *dinfp.DistributedInferenceEntitySpec
		if config.WorkerImage != nil || config.WorkerEnvironment != nil || config.WorkerMemoryLimit != nil || config.WorkerCpuLimit != nil || config.WorkerGpu != nil {
			workerSpec = &dinfp.DistributedInferenceEntitySpec{}
			if config.WorkerImage != nil {
				workerSpec.Image = &field_types.StringField{Value: config.WorkerImage}
			}
			if config.WorkerEnvironment != nil {
				envMap := make(map[string]field_types.DictionaryItemAndRules)
				for k, v := range config.WorkerEnvironment {
					envMap[k] = field_types.DictionaryItemAndRules{Value: ptr.To(v)}
				}
				workerSpec.Environment = &field_types.DictionaryField{Value: &envMap}
			}
			if config.WorkerMemoryLimit != nil {
				workerSpec.MemoryLimit = &field_types.QuantityField{Value: config.WorkerMemoryLimit}
			}
			if config.WorkerCpuLimit != nil {
				workerSpec.CpuLimit = &field_types.NumberField{Value: config.WorkerCpuLimit}
			}
			if config.WorkerGpu != nil {
				workerSpec.Gpu = &field_types.NumberField{Value: config.WorkerGpu}
			}
		}

		return &dinfwv1.DistributedInferenceWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      config.Name,
				Namespace: config.Namespace,
				UID:       types.UID(config.UID),
			},
			Spec: dinfp.DistributedInferenceSpec{
				IdentityFields: field_groups.IdentityFields{
					Usage: "Submit",
				},
				DistributedInferenceFields: distributedFields,
			},
			LeaderSpec: leaderSpec,
			WorkerSpec: workerSpec,
			Status: dinfwv1.DistributedInferenceWorkloadStatus{
				DistributedInferenceStatus: common.DistributedInferenceStatus{
					Replicas:      config.ReadyReplicas,
					ReadyReplicas: config.ReadyReplicas,
				},
			},
		}
	}

	setupTestEnvironment := func(namespace, projectName string) *WorkloadParser {
		namespaceObj := test.BuildNamespace(namespace, map[string]string{constants.RunaiQueueLabel: projectName})
		project := test.BuildProjectWithDepId(projectName, projectName+"-id", namespace, projectName+"-department-id")
		kaiProject := test.BuildKaiProjectWithDepId(projectName, projectName+"-id", namespace, projectName+"-department-id")
		fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(test.GetTestScheme()).WithObjects(namespaceObj, project, kaiProject)
		fakeClient := fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
		return NewWorkloadParser(fakeClient)
	}

	assertBasicFields := func(t *testing.T, result *openapi.WorkloadSyncFields, testConfig DistributedInferenceTestConfig) {
		assert.Equal(t, testConfig.Name, result.Name)
		assert.Equal(t, testConfig.Namespace, result.Namespace)
		assert.Equal(t, "DistributedInference", result.Type)
		assert.NotEmpty(t, result.Id)
		assert.Equal(t, config.Config().ClusterUuid, result.ClusterId)
	}

	assertRequestedPods := func(t *testing.T, result *openapi.WorkloadSyncFields, expectedCount int32) {
		assert.NotNil(t, result.RequestedPods)
		assert.NotNil(t, result.RequestedPods.Number)
		assert.True(t, result.RequestedPods.Number.IsSet())
		podCount := result.RequestedPods.Number.Get()
		assert.Equal(t, expectedCount, *podCount)
	}

	assertImages := func(t *testing.T, result *openapi.WorkloadSyncFields, expectedImages []string) {
		assert.NotNil(t, result.Images)
		assert.Equal(t, len(expectedImages), len(result.Images))
		for _, expectedImage := range expectedImages {
			assert.Contains(t, result.Images, expectedImage)
		}
	}

	assertCommandAndArguments := func(t *testing.T, result *openapi.WorkloadSyncFields, expectedCommand, expectedArguments *string) {
		if expectedCommand != nil {
			assert.NotNil(t, result.Command)
			assert.Equal(t, *expectedCommand, *result.Command)
		}
		if expectedArguments != nil {
			assert.NotNil(t, result.Arguments)
			assert.Equal(t, *expectedArguments, *result.Arguments)
		}
	}

	assertEnvironmentVariables := func(t *testing.T, result *openapi.WorkloadSyncFields, expectedEnv map[string]string) {
		if expectedEnv != nil {
			assert.NotNil(t, result.EnvironmentVariables)
			envVars := *result.EnvironmentVariables
			for k, v := range expectedEnv {
				assert.Equal(t, v, envVars[k])
			}
		}
	}

	assertFieldValue := func(t *testing.T, data map[string]interface{}, fieldName string, expectedValue interface{}) {
		field, exists := data[fieldName]
		assert.True(t, exists, "%s field should exist", fieldName)

		// If the field is a map with a 'value' key, unwrap it
		if fieldMap, ok := field.(map[string]interface{}); ok {
			if value, exists := fieldMap["value"]; exists {
				assert.Equal(t, expectedValue, value, "%s value should be %v", fieldName, expectedValue)
				return
			}
		}
		// Otherwise, compare directly
		assert.Equal(t, expectedValue, field, "%s should be %v", fieldName, expectedValue)
	}

	assertServingPorts := func(t *testing.T, data map[string]interface{}, expectedPorts map[string]ServingPortConfig) {
		servingPorts, exists := data["servingPorts"]
		assert.True(t, exists, "servingPorts field should exist")

		servingPortsMap, ok := servingPorts.(map[string]interface{})
		assert.True(t, ok, "servingPorts should be a map")
		assert.Equal(t, len(expectedPorts), len(servingPortsMap), "should have %d serving ports", len(expectedPorts))

		for portName, expectedConfig := range expectedPorts {
			port, exists := servingPortsMap[portName]
			assert.True(t, exists, "%s port should exist", portName)

			portMap, ok := port.(map[string]interface{})
			assert.True(t, ok, "%s port should be a map", portName)

			// Validate port number
			portNumber, exists := portMap["port"]
			assert.True(t, exists, "%s port should have port field", portName)
			assert.Equal(t, float64(expectedConfig.Port), portNumber, "%s port should be %d", portName, expectedConfig.Port)

			// Validate protocol
			protocol, exists := portMap["protocol"]
			assert.True(t, exists, "%s port should have protocol field", portName)
			assert.Equal(t, expectedConfig.Protocol, protocol, "%s port protocol should be %s", portName, expectedConfig.Protocol)

			// Validate optional fields+
			if expectedConfig.ExposeExternally != nil {
				exposeExternally, exists := portMap["exposeExternally"]
				assert.True(t, exists, "%s port should have exposeExternally field", portName)
				assert.Equal(t, *expectedConfig.ExposeExternally, exposeExternally, "%s port exposeExternally should be %v", portName, *expectedConfig.ExposeExternally)
			}

			if expectedConfig.AuthorizationType != nil {
				authType, exists := portMap["authorizationType"]
				assert.True(t, exists, "%s port should have authorizationType field", portName)
				assert.Equal(t, *expectedConfig.AuthorizationType, authType, "%s port authorizationType should be %s", portName, *expectedConfig.AuthorizationType)
			}
		}
	}

	assertAdditionalFields := func(t *testing.T, result *openapi.WorkloadSyncFields, testConfig DistributedInferenceTestConfig) {
		assert.NotNil(t, result.AdditionalFields)

		var diMap map[string]interface{}
		err := json.Unmarshal([]byte(*result.AdditionalFields), &diMap)
		assert.NoError(t, err)

		// Validate workers field
		assertFieldValue(t, diMap, "workers", float64(testConfig.Workers))

		// Validate specReplicas field (changed from replicas)
		assertFieldValue(t, diMap, "specReplicas", float64(testConfig.Replicas))

		// Validate statusReplicas field - only if > 0 (as per implementation)
		if testConfig.ReadyReplicas > 0 {
			assertFieldValue(t, diMap, "statusReplicas", float64(testConfig.ReadyReplicas))
		}

		// Validate servingPorts if present
		if len(testConfig.ServingPorts) > 0 {
			assertServingPorts(t, diMap, testConfig.ServingPorts)
		}

		// Validate restartPolicy if present
		if testConfig.RestartPolicy != nil {
			assertFieldValue(t, diMap, "restartPolicy", *testConfig.RestartPolicy)
		}

		// Validate startupPolicy if present
		if testConfig.StartupPolicy != nil {
			assertFieldValue(t, diMap, "startupPolicy", *testConfig.StartupPolicy)
		}

		if len(testConfig.LeaderLabels) > 0 {
			// Validate leaderLabels
			nimWorkload, exists := diMap[constants.AdditionalNIMWorkload]
			assert.True(t, exists, "nimWorkload field should exist")
			// and make sure it contains the labels and the value of nim workload is true
			assert.Equal(t, "true", nimWorkload, "nimWorkload should be true")
		}
	}

	runDistributedInferenceTest := func(t *testing.T, testConfig DistributedInferenceTestConfig) {
		// Create workload
		workload := createDistributedInferenceWorkload(testConfig)
		// Setup test environment
		parser := setupTestEnvironment(testConfig.Namespace, "test-project")
		// Parse the workload
		result, pods, err := parser.ParseDistributedInferenceWorkloadToServerScheme(workload)
		// Basic assertions
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, pods)
		// Validate all fields
		assertBasicFields(t, result, testConfig)
		assertRequestedPods(t, result, testConfig.ExpectedPodCount)
		assertImages(t, result, testConfig.ExpectedImages)
		assertCommandAndArguments(t, result, testConfig.ExpectedCommand, testConfig.ExpectedArguments)
		assertEnvironmentVariables(t, result, testConfig.ExpectedEnvironment)
		assertAdditionalFields(t, result, testConfig)
	}

	// Test cases
	t.Run("Basic", func(t *testing.T) {
		testConfig := DistributedInferenceTestConfig{
			Name:             "test-dinfw",
			Namespace:        "test-namespace",
			UID:              "test-uid",
			Workers:          3,
			Replicas:         2,
			ReadyReplicas:    1,
			LeaderImage:      "test-image:latest",
			ExpectedPodCount: 8, // (1 leader + 3 workers) * 2 replicas
			ExpectedImages:   []string{"test-image:latest"},
			ServingPorts: map[string]ServingPortConfig{
				"http": {Port: 8080, Protocol: "http"},
				"grpc": {Port: 9090, Protocol: "grpc"},
			},
		}
		runDistributedInferenceTest(t, testConfig)
	})

	t.Run("SimpleExample", func(t *testing.T) {
		testConfig := DistributedInferenceTestConfig{
			Name:              "simple-test-dist-inference",
			Namespace:         "runai-p1",
			UID:               "simple-test-uid",
			Workers:           1,
			Replicas:          1,
			ReadyReplicas:     0,
			LeaderImage:       "nginx:alpine",
			LeaderCommand:     ptr.To("nginx"),
			LeaderArguments:   ptr.To("-g 'daemon off;'"),
			LeaderMemoryLimit: ptr.To("128Mi"),
			LeaderCpuLimit:    ptr.To("0.1"),
			ExpectedPodCount:  2, // (1 leader + 1 worker) * 1 replica
			ExpectedImages:    []string{"nginx:alpine"},
			ExpectedCommand:   ptr.To("nginx"),
			ExpectedArguments: ptr.To("-g 'daemon off;'"),
			ServingPorts: map[string]ServingPortConfig{
				"http": {
					Port:              8080,
					Protocol:          "http",
					ExposeExternally:  ptr.To(true),
					AuthorizationType: ptr.To("public"),
				},
			},
			StartupPolicy: ptr.To("LeaderCreated"),
			RestartPolicy: ptr.To("RecreateGroupOnPodFailure"),
		}
		runDistributedInferenceTest(t, testConfig)
	})

	t.Run("WithWorkerSpec", func(t *testing.T) {
		testConfig := DistributedInferenceTestConfig{
			Name:              "test-dinfw-with-worker",
			Namespace:         "test-namespace",
			UID:               "test-uid-with-worker",
			Workers:           2,
			Replicas:          1,
			ReadyReplicas:     1,
			LeaderImage:       "leader-image:latest",
			WorkerImage:       ptr.To("worker-image:latest"),
			LeaderGpu:         ptr.To("1.0"),
			WorkerGpu:         ptr.To("2.0"),
			LeaderMemoryLimit: ptr.To("2Gi"),
			WorkerMemoryLimit: ptr.To("4Gi"),
			ExpectedPodCount:  3, // (1 leader + 2 workers) * 1 replica
			ExpectedImages:    []string{"leader-image:latest", "worker-image:latest"},
			LeaderEnvironment: map[string]string{
				"LEADER_ENV": "leader-value",
				"COMMON_ENV": "leader-common",
			},
			WorkerEnvironment: map[string]string{
				"WORKER_ENV": "worker-value",
				"COMMON_ENV": "worker-common", // This should be overridden by leader
			},
			ExpectedEnvironment: map[string]string{
				"LEADER_ENV": "leader-value",
				"WORKER_ENV": "worker-value",
				"COMMON_ENV": "leader-common", // Leader should override worker
			},
		}
		runDistributedInferenceTest(t, testConfig)
	})

	t.Run("Minimal", func(t *testing.T) {
		testConfig := DistributedInferenceTestConfig{
			Name:             "minimal-dinfw",
			Namespace:        "test-namespace",
			UID:              "minimal-uid",
			Workers:          0,
			Replicas:         1,
			ReadyReplicas:    0,
			LeaderImage:      "minimal-image:latest",
			ExpectedPodCount: 1, // (1 leader + 0 workers) * 1 replica
			ExpectedImages:   []string{"minimal-image:latest"},
		}
		runDistributedInferenceTest(t, testConfig)
	})

	t.Run("WithEnvironmentOnly", func(t *testing.T) {
		testConfig := DistributedInferenceTestConfig{
			Name:          "env-only-dinfw",
			Namespace:     "test-namespace",
			UID:           "env-only-uid",
			Workers:       1,
			Replicas:      1,
			ReadyReplicas: 0,
			LeaderImage:   "env-image:latest",
			LeaderEnvironment: map[string]string{
				"ENV_VAR_1": "value1",
				"ENV_VAR_2": "value2",
			},
			ExpectedPodCount: 2, // (1 leader + 1 worker) * 1 replica
			ExpectedImages:   []string{"env-image:latest"},
			ExpectedEnvironment: map[string]string{
				"ENV_VAR_1": "value1",
				"ENV_VAR_2": "value2",
			},
		}
		runDistributedInferenceTest(t, testConfig)
	})

	t.Run("WithResourceLimits", func(t *testing.T) {
		testConfig := DistributedInferenceTestConfig{
			Name:              "resource-dinfw",
			Namespace:         "test-namespace",
			UID:               "resource-uid",
			Workers:           1,
			Replicas:          1,
			ReadyReplicas:     0,
			LeaderImage:       "resource-image:latest",
			LeaderMemoryLimit: ptr.To("4Gi"),
			LeaderCpuLimit:    ptr.To("2.0"),
			LeaderGpu:         ptr.To("1.0"),
			ExpectedPodCount:  2, // (1 leader + 1 worker) * 1 replica
			ExpectedImages:    []string{"resource-image:latest"},
		}
		runDistributedInferenceTest(t, testConfig)
	})

	t.Run("WithMultipleServingPorts", func(t *testing.T) {
		testConfig := DistributedInferenceTestConfig{
			Name:             "multi-port-dinfw",
			Namespace:        "test-namespace",
			UID:              "multi-port-uid",
			Workers:          1,
			Replicas:         1,
			ReadyReplicas:    0,
			LeaderImage:      "multi-port-image:latest",
			ExpectedPodCount: 2, // (1 leader + 1 worker) * 1 replica
			ExpectedImages:   []string{"multi-port-image:latest"},
			ServingPorts: map[string]ServingPortConfig{
				"http": {
					Port:              8080,
					Protocol:          "http",
					ExposeExternally:  ptr.To(true),
					AuthorizationType: ptr.To("public"),
				},
				"grpc": {
					Port:              9090,
					Protocol:          "grpc",
					ExposeExternally:  ptr.To(false),
					AuthorizationType: ptr.To("private"),
				},
				"metrics": {
					Port:             9091,
					Protocol:         "http",
					ExposeExternally: ptr.To(false),
				},
			},
		}
		runDistributedInferenceTest(t, testConfig)
	})

	t.Run("WithWorkerOverride", func(t *testing.T) {
		testConfig := DistributedInferenceTestConfig{
			Name:            "worker-override-dinfw",
			Namespace:       "test-namespace",
			UID:             "worker-override-uid",
			Workers:         1,
			Replicas:        1,
			ReadyReplicas:   0,
			LeaderImage:     "leader-image:latest",
			WorkerImage:     ptr.To("worker-image:latest"),
			LeaderCommand:   ptr.To("leader-command"),
			LeaderArguments: ptr.To("leader-args"),
			LeaderEnvironment: map[string]string{
				"LEADER_ONLY": "leader-value",
				"SHARED":      "leader-shared",
			},
			WorkerEnvironment: map[string]string{
				"WORKER_ONLY": "worker-value",
				"SHARED":      "worker-shared", // Should be overridden by leader
			},
			ExpectedPodCount:  2, // (1 leader + 1 worker) * 1 replica
			ExpectedImages:    []string{"leader-image:latest", "worker-image:latest"},
			ExpectedCommand:   ptr.To("leader-command"), // Leader command should be used
			ExpectedArguments: ptr.To("leader-args"),    // Leader arguments should be used
			ExpectedEnvironment: map[string]string{
				"LEADER_ONLY": "leader-value",
				"WORKER_ONLY": "worker-value",
				"SHARED":      "leader-shared", // Leader should override worker
			},
		}
		runDistributedInferenceTest(t, testConfig)
	})

	t.Run("WithLargeScale", func(t *testing.T) {
		testConfig := DistributedInferenceTestConfig{
			Name:             "large-scale-dinfw",
			Namespace:        "test-namespace",
			UID:              "large-scale-uid",
			Workers:          10,
			Replicas:         5,
			ReadyReplicas:    3,
			LeaderImage:      "large-scale-image:latest",
			ExpectedPodCount: 55, // (1 leader + 10 workers) * 5 replicas
			ExpectedImages:   []string{"large-scale-image:latest"},
			ServingPorts: map[string]ServingPortConfig{
				"http": {Port: 8080, Protocol: "http"},
			},
		}
		runDistributedInferenceTest(t, testConfig)
	})

	t.Run("WithNilFields", func(t *testing.T) {
		testConfig := DistributedInferenceTestConfig{
			Name:             "nil-fields-dinfw",
			Namespace:        "test-namespace",
			UID:              "nil-fields-uid",
			Workers:          0,
			Replicas:         1,
			ReadyReplicas:    0,
			LeaderImage:      "nil-fields-image:latest",
			ExpectedPodCount: 1, // (1 leader + 0 workers) * 1 replica
			ExpectedImages:   []string{"nil-fields-image:latest"},
			// All other fields are nil/empty to test nil handling
		}
		runDistributedInferenceTest(t, testConfig)
	})

	t.Run("GPUCalculationBugFix", func(t *testing.T) {
		testConfig := DistributedInferenceTestConfig{
			Name:             "gpu-calculation-bug-fix",
			Namespace:        "test-namespace",
			UID:              "gpu-calculation-uid",
			Workers:          1,
			Replicas:         1,
			ReadyReplicas:    0,
			LeaderImage:      "gpu-test-image:latest",
			LeaderGpu:        ptr.To("1.0"),
			WorkerGpu:        ptr.To("0.5"),
			ExpectedPodCount: 2, // (1 leader + 1 worker) * 1 replica
			ExpectedImages:   []string{"gpu-test-image:latest"},
		}

		// Create workload
		workload := createDistributedInferenceWorkload(testConfig)
		// Setup test environment
		parser := setupTestEnvironment(testConfig.Namespace, "test-project")
		// Parse the workload
		result, pods, err := parser.ParseDistributedInferenceWorkloadToServerScheme(workload)

		// Basic assertions
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, pods)

		// Validate GPU calculation - should be 1.5 (1.0 leader + 0.5 worker)
		assert.NotNil(t, result.RequestedResources)
		assert.True(t, result.RequestedResources.IsSet())
		requestedResources := result.RequestedResources.Get()
		assert.NotNil(t, requestedResources.Gpu)
		assert.True(t, requestedResources.Gpu.IsSet())
		gpuCores := requestedResources.Gpu.Get()
		assert.NotNil(t, gpuCores.Request)
		assert.True(t, gpuCores.Request.IsSet())
		gpuRequest := gpuCores.Request.Get()
		assert.Equal(t, 1.5, *gpuRequest, "GPU should be 1.5 (1.0 leader + 0.5 worker)")
	})

	t.Run("EnvironmentVariablesBugFix", func(t *testing.T) {
		testConfig := DistributedInferenceTestConfig{
			Name:          "env-vars-bug-fix",
			Namespace:     "test-namespace",
			UID:           "env-vars-uid",
			Workers:       1,
			Replicas:      1,
			ReadyReplicas: 0,
			LeaderImage:   "env-test-image:latest",
			LeaderEnvironment: map[string]string{
				"LEADER_ENV": "leader-value",
				"COMMON_ENV": "leader-common",
			},
			WorkerEnvironment: map[string]string{
				"WORKER_ENV": "worker-value",
				"COMMON_ENV": "worker-common", // Should be overridden by leader
			},
			ExpectedPodCount: 2, // (1 leader + 1 worker) * 1 replica
			ExpectedImages:   []string{"env-test-image:latest"},
			ExpectedEnvironment: map[string]string{
				"LEADER_ENV": "leader-value",
				"WORKER_ENV": "worker-value",
				"COMMON_ENV": "leader-common", // Leader should override worker
			},
		}

		// Create workload
		workload := createDistributedInferenceWorkload(testConfig)
		// Setup test environment
		parser := setupTestEnvironment(testConfig.Namespace, "test-project")
		// Parse the workload
		result, pods, err := parser.ParseDistributedInferenceWorkloadToServerScheme(workload)

		// Basic assertions
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, pods)

		// Validate environment variables are properly set (not empty)
		assert.NotNil(t, result.EnvironmentVariables)
		envVars := *result.EnvironmentVariables
		assert.Equal(t, 3, len(envVars), "Should have 3 environment variables")

		// Check specific environment variables
		assert.Equal(t, "leader-value", envVars["LEADER_ENV"])
		assert.Equal(t, "worker-value", envVars["WORKER_ENV"])
		assert.Equal(t, "leader-common", envVars["COMMON_ENV"], "Leader should override worker for common env")

		// Validate that no environment variables are empty
		for key, value := range envVars {
			assert.NotEmpty(t, value, "Environment variable %s should not be empty", key)
		}
	})

	t.Run("CommandArgumentsBugFix", func(t *testing.T) {
		testConfig := DistributedInferenceTestConfig{
			Name:              "cmd-args-bug-fix",
			Namespace:         "test-namespace",
			UID:               "cmd-args-uid",
			Workers:           1,
			Replicas:          1,
			ReadyReplicas:     0,
			LeaderImage:       "cmd-test-image:latest",
			LeaderCommand:     ptr.To("python"),
			LeaderArguments:   ptr.To("app.py --mode=leader"),
			ExpectedPodCount:  2, // (1 leader + 1 worker) * 1 replica
			ExpectedImages:    []string{"cmd-test-image:latest"},
			ExpectedCommand:   ptr.To("python"),
			ExpectedArguments: ptr.To("app.py --mode=leader"),
		}

		// Create workload
		workload := createDistributedInferenceWorkload(testConfig)
		// Setup test environment
		parser := setupTestEnvironment(testConfig.Namespace, "test-project")
		// Parse the workload
		result, pods, err := parser.ParseDistributedInferenceWorkloadToServerScheme(workload)

		// Basic assertions
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, pods)

		// Validate command and arguments are properly set
		assert.NotNil(t, result.Command)
		assert.Equal(t, "python", *result.Command)
		assert.NotNil(t, result.Arguments)
		assert.Equal(t, "app.py --mode=leader", *result.Arguments)
	})

	t.Run("ServingPortsStructureBugFix", func(t *testing.T) {
		testConfig := DistributedInferenceTestConfig{
			Name:             "serving-ports-structure-bug-fix",
			Namespace:        "test-namespace",
			UID:              "serving-ports-uid",
			Workers:          1,
			Replicas:         1,
			ReadyReplicas:    0,
			LeaderImage:      "ports-test-image:latest",
			ExpectedPodCount: 2, // (1 leader + 1 worker) * 1 replica
			ExpectedImages:   []string{"ports-test-image:latest"},
			ServingPorts: map[string]ServingPortConfig{
				"http": {
					Port:              8080,
					Protocol:          "http",
					ExposeExternally:  ptr.To(true),
					AuthorizationType: ptr.To("public"),
				},
				"grpc": {
					Port:              9090,
					Protocol:          "grpc",
					ExposeExternally:  ptr.To(false),
					AuthorizationType: ptr.To("private"),
				},
			},
		}

		// Create workload
		workload := createDistributedInferenceWorkload(testConfig)
		// Setup test environment
		parser := setupTestEnvironment(testConfig.Namespace, "test-project")
		// Parse the workload
		result, pods, err := parser.ParseDistributedInferenceWorkloadToServerScheme(workload)

		// Basic assertions
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, pods)

		// Validate serving ports structure in additional fields
		assert.NotNil(t, result.AdditionalFields)
		var additionalFields map[string]interface{}
		err = json.Unmarshal([]byte(*result.AdditionalFields), &additionalFields)
		assert.NoError(t, err)

		// Check that servingPorts exists and has the correct structure
		servingPorts, exists := additionalFields["servingPorts"]
		assert.True(t, exists, "servingPorts should exist in additional fields")

		servingPortsMap, ok := servingPorts.(map[string]interface{})
		assert.True(t, ok, "servingPorts should be a map")
		assert.Equal(t, 2, len(servingPortsMap), "Should have 2 serving ports")

		// Validate http port
		httpPort, exists := servingPortsMap["http"]
		assert.True(t, exists, "http port should exist")
		httpPortMap, ok := httpPort.(map[string]interface{})
		assert.True(t, ok, "http port should be a map")
		assert.Equal(t, float64(8080), httpPortMap["port"])
		assert.Equal(t, "http", httpPortMap["protocol"])
		assert.Equal(t, true, httpPortMap["exposeExternally"])
		assert.Equal(t, "public", httpPortMap["authorizationType"])

		// Validate grpc port
		grpcPort, exists := servingPortsMap["grpc"]
		assert.True(t, exists, "grpc port should exist")
		grpcPortMap, ok := grpcPort.(map[string]interface{})
		assert.True(t, ok, "grpc port should be a map")
		assert.Equal(t, float64(9090), grpcPortMap["port"])
		assert.Equal(t, "grpc", grpcPortMap["protocol"])
		assert.Equal(t, false, grpcPortMap["exposeExternally"])
		assert.Equal(t, "private", grpcPortMap["authorizationType"])
	})

	t.Run("NoDoubleCountingBugFix", func(t *testing.T) {
		testConfig := DistributedInferenceTestConfig{
			Name:             "no-double-counting-bug-fix",
			Namespace:        "test-namespace",
			UID:              "no-double-counting-uid",
			Workers:          2,
			Replicas:         3,
			ReadyReplicas:    0,
			LeaderImage:      "double-counting-test-image:latest",
			LeaderGpu:        ptr.To("1.0"),
			WorkerGpu:        ptr.To("0.5"),
			ExpectedPodCount: 9, // (1 leader + 2 workers) * 3 replicas
			ExpectedImages:   []string{"double-counting-test-image:latest"},
		}

		// Create workload
		workload := createDistributedInferenceWorkload(testConfig)
		// Setup test environment
		parser := setupTestEnvironment(testConfig.Namespace, "test-project")
		// Parse the workload
		result, pods, err := parser.ParseDistributedInferenceWorkloadToServerScheme(workload)

		// Basic assertions
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, pods)

		// Validate GPU calculation - should be 6.0 (1.0 leader * 3 replicas + 0.5 worker * 2 workers * 3 replicas)
		assert.NotNil(t, result.RequestedResources)
		assert.True(t, result.RequestedResources.IsSet())
		requestedResources := result.RequestedResources.Get()
		assert.NotNil(t, requestedResources.Gpu)
		assert.True(t, requestedResources.Gpu.IsSet())
		gpuCores := requestedResources.Gpu.Get()
		assert.NotNil(t, gpuCores.Request)
		assert.True(t, gpuCores.Request.IsSet())
		gpuRequest := gpuCores.Request.Get()
		expectedGPU := 1.0*3 + 0.5*2*3 // leader * replicas + worker * workers * replicas
		assert.Equal(t, expectedGPU, *gpuRequest, "GPU should be %f (1.0 leader * 3 replicas + 0.5 worker * 2 workers * 3 replicas)", expectedGPU)

		// Validate that the function setDistributedInferenceRequestedResources doesn't exist in the parser
		// This is a compile-time check, but we can verify the behavior is correct
		assert.Equal(t, int32(9), *result.RequestedPods.Number.Get(), "Should have 9 pods total")
	})

	t.Run("nim label", func(t *testing.T) {
		testConfig := DistributedInferenceTestConfig{
			Name:             "test-dinfw",
			Namespace:        "test-namespace",
			UID:              "test-uid",
			Workers:          3,
			Replicas:         2,
			ReadyReplicas:    1,
			LeaderImage:      "test-image:latest",
			ExpectedPodCount: 8, // (1 leader + 3 workers) * 2 replicas
			ExpectedImages:   []string{"test-image:latest"},
			ServingPorts: map[string]ServingPortConfig{
				"http": {Port: 8080, Protocol: "http"},
			},
			LeaderLabels: map[string]string{
				constants.LabelNIMWorkload: "true",
			},
		}
		runDistributedInferenceTest(t, testConfig)
	})
}
