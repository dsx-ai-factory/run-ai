/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/util"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/test"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	knv1 "knative.dev/serving/pkg/apis/serving/v1"
)

var _ = Describe("Revision Parser", func() {
	Context("Inference Workload", func() {
		It("parse Revision requested pods - with min and max", func() {
			rev := &knv1.Revision{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "infw-1",
					Namespace: "runai-project-a",
					Annotations: map[string]string{
						"autoscaling.knative.dev/min-scale": "2",
						"autoscaling.knative.dev/max-scale": "3",
					},
				},
			}

			result := &openapi.WorkloadSyncFields{
				RequestedPods: parseRevisionRequestedPods(rev),
			}

			Expect(result.RequestedPods).ToNot(BeNil())
			Expect(result.RequestedPods.Number.IsSet()).To(BeFalse())
			Expect(result.RequestedPods.Min.IsSet()).To(BeTrue())
			Expect(result.RequestedPods.Max.IsSet()).To(BeTrue())
			Expect(*result.RequestedPods.Min.Get()).To(BeEquivalentTo(2))
			Expect(*result.RequestedPods.Max.Get()).To(BeEquivalentTo(3))
		})
		It("parse Revision requested pods - with min only", func() {
			rev := &knv1.Revision{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "infw-1",
					Namespace: "runai-project-a",
					Annotations: map[string]string{
						"autoscaling.knative.dev/min-scale": "2",
					},
				},
			}

			result := &openapi.WorkloadSyncFields{
				RequestedPods: parseRevisionRequestedPods(rev),
			}

			Expect(result.RequestedPods).ToNot(BeNil())
			Expect(result.RequestedPods.Number.IsSet()).To(BeFalse())
			Expect(result.RequestedPods.Min.IsSet()).To(BeTrue())
			Expect(result.RequestedPods.Max.IsSet()).To(BeFalse())
			Expect(*result.RequestedPods.Min.Get()).To(BeEquivalentTo(2))
		})
		It("should parse a Revision to server scheme", func() {
			var gibiResource = resource.MustParse("1Gi")
			uid := uuid.NewString()
			workloadUuid := uuid.NewString()

			proj := &util.ProjectData{
				ProjectId:    "1",
				DepartmentId: "2",
			}

			cases := []struct {
				rev      *knv1.Revision
				expected *openapi.RevisionSyncFields
			}{
				{
					rev: &knv1.Revision{
						TypeMeta: metav1.TypeMeta{
							Kind: "Revision",
						},
						ObjectMeta: metav1.ObjectMeta{
							UID:               types.UID(uid),
							Name:              "revName",
							CreationTimestamp: metav1.Time{Time: time.Unix(100, 0)},
							Labels:            map[string]string{},
							Annotations:       map[string]string{},
						},
						Spec: knv1.RevisionSpec{
							PodSpec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:    "take-first-as-default",
										Image:   "ubuntu",
										Command: []string{"sleep", "infinity"},
										Resources: corev1.ResourceRequirements{
											Requests: corev1.ResourceList{
												corev1.ResourceCPU:    resource.MustParse("800m"),
												corev1.ResourceMemory: gibiResource,
											},
										},
									},

									{
										Name: "other",
									},
								},
							},
						},
					},
					expected: &openapi.RevisionSyncFields{
						Type:         "Revision",
						Name:         "revName",
						Id:           uid,
						WorkloadId:   workloadUuid,
						ClusterId:    test.ClusterIdForTests,
						DepartmentId: "2",
						ProjectId:    "1",
						CreatedAt:    time.Unix(100, 0),
						RequestedResources: *openapi.NewNullableWorkloadRequestResources(&openapi.WorkloadRequestResources{
							Cpu: *openapi.NewNullableRequestResourceCores(
								&openapi.RequestResourceCores{
									Request: *openapi.NewNullableFloat64(ptr.To(0.8)),
								}),
							CpuMemory: *openapi.NewNullableRequestResourceQuantity(
								&openapi.RequestResourceQuantity{
									Request: *openapi.NewNullableString(ptr.To(gibiResource.String())),
								}),
						}),
						RequestedPods:        &openapi.RequestedPods{},
						Images:               []string{"ubuntu"},
						Command:              ptr.To("sleep infinity"),
						Arguments:            ptr.To(""),
						Conditions:           []openapi.Condition{},
						EnvironmentVariables: ptr.To(map[string]string{}),
						Phase:                openapi.PHASE_UNKNOWN,
					},
				},
				{
					rev: &knv1.Revision{
						TypeMeta: metav1.TypeMeta{
							Kind: "Revision",
						},
						ObjectMeta: metav1.ObjectMeta{
							UID:               types.UID(uid),
							Name:              "revName",
							CreationTimestamp: metav1.Time{Time: time.Unix(100, 0)},
						},
						Spec: knv1.RevisionSpec{
							PodSpec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "other",
									},
									{
										Name:  "user-container-0",
										Image: "ubuntu",
										Args:  []string{"sleep", "infinity"},
										Env: []corev1.EnvVar{
											{
												Name:  "key",
												Value: "value",
											},
										},
										Resources: corev1.ResourceRequirements{
											Requests: corev1.ResourceList{
												"nvidia.com/gpu": resource.MustParse("1"),
											},
											Limits: corev1.ResourceList{
												"nvidia.com/gpu": resource.MustParse("10"),
											},
										},
									},
								},
							},
						},
						Status: knv1.RevisionStatus{
							Status: duckv1.Status{
								Annotations: map[string]string{
									"run.ai/phase":         "Running",
									"run.ai/phase-message": "",
								},
							},
						},
					},
					expected: &openapi.RevisionSyncFields{
						Type:         "Revision",
						Name:         "revName",
						Id:           uid,
						WorkloadId:   workloadUuid,
						ClusterId:    test.ClusterIdForTests,
						DepartmentId: "2",
						ProjectId:    "1",
						CreatedAt:    time.Unix(100, 0),
						RequestedResources: *openapi.NewNullableWorkloadRequestResources(&openapi.WorkloadRequestResources{
							GpuRequestType: *openapi.NewNullableGpuRequestType(ptr.To(openapi.GPUREQUESTTYPE_PORTION)),
							Gpu: *openapi.NewNullableRequestResourceCores(
								&openapi.RequestResourceCores{
									Request: *openapi.NewNullableFloat64(ptr.To(float64(1))),
									Limit:   *openapi.NewNullableFloat64(ptr.To(float64(10))),
								}),
						}),
						RequestedPods: &openapi.RequestedPods{},
						Images:        []string{"ubuntu"},
						Arguments:     ptr.To("sleep infinity"),
						Command:       ptr.To(""),
						Conditions:    []openapi.Condition{},
						EnvironmentVariables: ptr.To(map[string]string{
							"key": "value",
						}),
						Phase:        openapi.PHASE_RUNNING,
						PhaseMessage: ptr.To(""),
					},
				},
				{
					rev: &knv1.Revision{
						TypeMeta: metav1.TypeMeta{
							Kind: "Revision",
						},
						ObjectMeta: metav1.ObjectMeta{
							UID:               types.UID(uid),
							Name:              "revName",
							CreationTimestamp: metav1.Time{Time: time.Unix(100, 0)},
						},
						Spec: knv1.RevisionSpec{
							PodSpec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "other",
									},
									{
										Name:  "user-container-0",
										Image: "ubuntu",
										Args:  []string{"sleep", "infinity"},
										Env: []corev1.EnvVar{
											{
												Name:  "key",
												Value: "value",
											},
										},
										Resources: corev1.ResourceRequirements{
											Requests: corev1.ResourceList{
												"nvidia.com/gpu": resource.MustParse("1"),
											},
											Limits: corev1.ResourceList{
												"nvidia.com/gpu": resource.MustParse("10"),
											},
										},
									},
								},
							},
						},
						Status: knv1.RevisionStatus{
							Status: duckv1.Status{
								Annotations: map[string]string{
									"run.ai/phase":         "wierd phase",
									"run.ai/phase-message": "some phase msg",
								},
							},
						},
					},
					expected: &openapi.RevisionSyncFields{
						Type:         "Revision",
						Name:         "revName",
						Id:           uid,
						WorkloadId:   workloadUuid,
						ClusterId:    test.ClusterIdForTests,
						DepartmentId: "2",
						ProjectId:    "1",
						CreatedAt:    time.Unix(100, 0),
						RequestedResources: *openapi.NewNullableWorkloadRequestResources(&openapi.WorkloadRequestResources{
							GpuRequestType: *openapi.NewNullableGpuRequestType(ptr.To(openapi.GPUREQUESTTYPE_PORTION)),
							Gpu: *openapi.NewNullableRequestResourceCores(
								&openapi.RequestResourceCores{
									Request: *openapi.NewNullableFloat64(ptr.To(float64(1))),
									Limit:   *openapi.NewNullableFloat64(ptr.To(float64(10))),
								}),
						}),
						RequestedPods: &openapi.RequestedPods{},
						Images:        []string{"ubuntu"},
						Arguments:     ptr.To("sleep infinity"),
						Command:       ptr.To(""),
						Conditions:    []openapi.Condition{},
						EnvironmentVariables: ptr.To(map[string]string{
							"key": "value",
						}),
						Phase:        openapi.PHASE_UNKNOWN,
						PhaseMessage: ptr.To("some phase msg"),
					},
				},
			}

			parser := NewRevisionParser()
			for _, c := range cases {
				actual, err := parser.ParseRevisionToServerScheme(c.rev, workloadUuid, proj)
				Expect(err).To(BeNil())

				Expect(actual).To(Not(BeNil()))
				Expect(*actual).To(Equal(*c.expected))
			}
		})
		It("should reject an empty workload ID", func() {
			var gibiResource = resource.MustParse("1Gi")
			uid := uuid.NewString()

			proj := &util.ProjectData{
				ProjectId:    "1",
				DepartmentId: "2",
			}

			rev := &knv1.Revision{
				TypeMeta: metav1.TypeMeta{
					Kind: "Revision",
				},
				ObjectMeta: metav1.ObjectMeta{
					UID:               types.UID(uid),
					Name:              "revName",
					CreationTimestamp: metav1.Time{Time: time.Unix(100, 0)},
					Labels:            map[string]string{},
					Annotations:       map[string]string{},
				},
				Spec: knv1.RevisionSpec{
					PodSpec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:    "take-first-as-default",
								Image:   "ubuntu",
								Command: []string{"sleep", "infinity"},
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("800m"),
										corev1.ResourceMemory: gibiResource,
									},
								},
							},
						},
					},
				},
			}

			parser := NewRevisionParser()
			_, err := parser.ParseRevisionToServerScheme(rev, "", proj)
			Expect(err).To(Not(BeNil()))
			Expect(err.Error()).To(Equal("workload ID is empty"))
		})
	})
})
