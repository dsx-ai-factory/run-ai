/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	_ "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/tests/helper"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	workloadsManagerApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	runaiBackendClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/backend-client"
	clusterCommonConstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	pcreds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/personalcreds"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	runaiv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("WorkloadsV2Handler", func() {
	Describe("initialWorkloadsV2SyncCounter", func() {
		It("requests a one-time replay when a control-plane-managed runner has no owner", func() {
			scheme := runtime.NewScheme()
			Expect(dromosv1alpha1.AddToScheme(scheme)).To(Succeed())
			runner := &dromosv1alpha1.WorkloadRunner{ObjectMeta: metav1.ObjectMeta{
				Name: "ownerless-runner", Namespace: "runai-project",
				Labels: map[string]string{constants.ControlPlaneManagedLabel: "true"},
			}}
			k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(runner).Build()

			counter, err := initialWorkloadsV2SyncCounter(k8sClient, 42)

			Expect(err).NotTo(HaveOccurred())
			Expect(counter).To(BeZero())
		})

		It("preserves the counter when managed runners already have owners", func() {
			scheme := runtime.NewScheme()
			Expect(dromosv1alpha1.AddToScheme(scheme)).To(Succeed())
			runner := &dromosv1alpha1.WorkloadRunner{ObjectMeta: metav1.ObjectMeta{
				Name: "owned-runner", Namespace: "runai-project",
				Labels: map[string]string{constants.ControlPlaneManagedLabel: "true"},
				Annotations: map[string]string{
					clusterCommonConstants.WorkloadOwnerUIDAnnotation: "owner-uid",
				},
			}}
			k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(runner).Build()

			counter, err := initialWorkloadsV2SyncCounter(k8sClient, 42)

			Expect(err).NotTo(HaveOccurred())
			Expect(counter).To(Equal(int64(42)))
		})
	})

	It("always uses WorkloadRunner", func() {
		handler := &WorkloadsV2Handler{}
		clusterWorkload, err := handler.convertToClusterResource(workloadsManagerApi.WorkloadV2Update{
			Id: "workload-id",
			WorkloadManifest: map[string]interface{}{
				"apiVersion": "batch/v1",
				"kind":       "Job",
				"metadata": map[string]interface{}{
					"name": "workload",
				},
			},
		}, "namespace")

		Expect(err).NotTo(HaveOccurred())
		Expect(clusterWorkload.GetObjectKind().GroupVersionKind().Kind).To(Equal(KindWorkloadRunner))
	})

	DescribeTable("idempotent workload deletion",
		func(deletionErr error, expectedComplete bool) {
			Expect(isWorkloadDeletionComplete(deletionErr)).To(Equal(expectedComplete))
		},
		Entry("treats NotFound as complete", errors.NewNotFound(schema.GroupResource{Group: "run.ai", Resource: "trainingworkloads"}, "legacy-training"), true),
		Entry("treats a missing CRD as complete", &meta.NoKindMatchError{GroupKind: schema.GroupKind{Group: "run.ai", Kind: "TrainingWorkload"}}, true),
		Entry("keeps other deletion failures", fmt.Errorf("connection refused"), false),
	)

	Context("wrapWithWorkloadRunner", func() {
		var (
			dummyWorkloadId = "dummy-workload-id"
		)
		It("should wrap an unstructured object with WorkloadRunner", func() {
			// Create a sample unstructured workload
			manifest := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "batch/v1",
					"kind":       "Job",
					"metadata": map[string]interface{}{
						"name":      "test-job",
						"namespace": "test-namespace",
					},
					"spec": map[string]interface{}{
						"template": map[string]interface{}{
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "test-container",
										"image": "test-image:latest",
									},
								},
							},
						},
					},
				},
			}

			// Wrap the manifest
			workloadRunner, err := wrapWithWorkloadRunner(manifest, dummyWorkloadId, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(workloadRunner).NotTo(BeNil())

			// Verify TypeMeta
			Expect(workloadRunner.TypeMeta.Kind).To(Equal("WorkloadRunner"))
			Expect(workloadRunner.TypeMeta.APIVersion).To(Equal("run.ai/v1alpha1"))

			// Verify the original manifest is embedded
			manifest.SetAnnotations(map[string]string{
				constants.ControlPlaneWorkloadIdAnnotationForWorkload: dummyWorkloadId,
			})
			Expect(workloadRunner.Spec.Template.Object).To(Equal(manifest))
		})

		It("should preserve the original object reference in the Template", func() {
			originalManifest := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
				},
			}

			workloadRunner, err := wrapWithWorkloadRunner(originalManifest, dummyWorkloadId, nil)

			Expect(err).NotTo(HaveOccurred())

			// Verify the Template.Object points to the original manifest
			Expect(workloadRunner.Spec.Template.Object).To(BeIdenticalTo(originalManifest))
		})

		Context("MNNVL configuration translation", func() {
			var manifest *unstructured.Unstructured

			BeforeEach(func() {
				manifest = &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "batch/v1",
						"kind":       "Job",
						"metadata": map[string]interface{}{
							"name":      "test-job",
							"namespace": "test-namespace",
						},
					},
				}
			})

			It("should translate Mode='Required' with GPU acceleration labels to MnnvlModeRequired", func() {
				labelKeys := []string{"nvidia.com/gpu.clique", "topology.kubernetes.io/rack"}
				capabilities := &workloadsManagerApi.WorkloadV2UpdateConfiguration{
					Mnnvl: &workloadsManagerApi.WorkloadV2UpdateConfigurationMnnvl{
						Mode:                            convutils.Ref(workloadsManagerApi.MNNVLMODE_REQUIRED),
						GpuNetworkAccelerationLabelKeys: labelKeys,
					},
				}

				workloadRunner, err := wrapWithWorkloadRunner(manifest, dummyWorkloadId, capabilities)

				Expect(err).NotTo(HaveOccurred())

				Expect(workloadRunner).NotTo(BeNil())
				Expect(workloadRunner.Spec.Instructions).NotTo(BeNil())
				Expect(workloadRunner.Spec.Instructions.GpuOptimizations).NotTo(BeNil())
				Expect(workloadRunner.Spec.Instructions.GpuOptimizations.Mnnvl).NotTo(BeNil())
				Expect(workloadRunner.Spec.Instructions.GpuOptimizations.Mnnvl.Mode).To(Equal(dromosv1alpha1.MnnvlModeRequired))
				Expect(workloadRunner.Spec.Instructions.GpuOptimizations.Mnnvl.GpuNetworkAccelerationLabelKeys).To(ConsistOf(labelKeys))
			})

			It("should translate Mode='Required' without GPU labels to MnnvlModeRequired with empty label keys", func() {
				capabilities := &workloadsManagerApi.WorkloadV2UpdateConfiguration{
					Mnnvl: &workloadsManagerApi.WorkloadV2UpdateConfigurationMnnvl{
						Mode: convutils.Ref(workloadsManagerApi.MNNVLMODE_REQUIRED),
					},
				}

				workloadRunner, err := wrapWithWorkloadRunner(manifest, dummyWorkloadId, capabilities)

				Expect(err).NotTo(HaveOccurred())

				Expect(workloadRunner).NotTo(BeNil())
				Expect(workloadRunner.Spec.Instructions).NotTo(BeNil())
				Expect(workloadRunner.Spec.Instructions.GpuOptimizations).NotTo(BeNil())
				Expect(workloadRunner.Spec.Instructions.GpuOptimizations.Mnnvl).NotTo(BeNil())
				Expect(workloadRunner.Spec.Instructions.GpuOptimizations.Mnnvl.Mode).To(Equal(dromosv1alpha1.MnnvlModeRequired))
				Expect(workloadRunner.Spec.Instructions.GpuOptimizations.Mnnvl.GpuNetworkAccelerationLabelKeys).To(BeNil())
			})

			It("should translate Mode='None' with GPU labels to MnnvlModeNone", func() {
				labelKeys := []string{"nvidia.com/gpu.clique"}
				capabilities := &workloadsManagerApi.WorkloadV2UpdateConfiguration{
					Mnnvl: &workloadsManagerApi.WorkloadV2UpdateConfigurationMnnvl{
						Mode:                            convutils.Ref(workloadsManagerApi.MNNVLMODE_NONE),
						GpuNetworkAccelerationLabelKeys: labelKeys,
					},
				}

				workloadRunner, err := wrapWithWorkloadRunner(manifest, dummyWorkloadId, capabilities)

				Expect(err).NotTo(HaveOccurred())

				Expect(workloadRunner).NotTo(BeNil())
				Expect(workloadRunner.Spec.Instructions).NotTo(BeNil())
				Expect(workloadRunner.Spec.Instructions.GpuOptimizations).NotTo(BeNil())
				Expect(workloadRunner.Spec.Instructions.GpuOptimizations.Mnnvl).NotTo(BeNil())
				Expect(workloadRunner.Spec.Instructions.GpuOptimizations.Mnnvl.Mode).To(Equal(dromosv1alpha1.MnnvlModeNone))
				Expect(workloadRunner.Spec.Instructions.GpuOptimizations.Mnnvl.GpuNetworkAccelerationLabelKeys).To(ConsistOf(labelKeys))
			})

			It("should translate Mode='None' without GPU labels to MnnvlModeNone with empty label keys", func() {
				capabilities := &workloadsManagerApi.WorkloadV2UpdateConfiguration{
					Mnnvl: &workloadsManagerApi.WorkloadV2UpdateConfigurationMnnvl{
						Mode: convutils.Ref(workloadsManagerApi.MNNVLMODE_NONE),
					},
				}

				workloadRunner, err := wrapWithWorkloadRunner(manifest, dummyWorkloadId, capabilities)

				Expect(err).NotTo(HaveOccurred())

				Expect(workloadRunner).NotTo(BeNil())
				Expect(workloadRunner.Spec.Instructions).NotTo(BeNil())
				Expect(workloadRunner.Spec.Instructions.GpuOptimizations).NotTo(BeNil())
				Expect(workloadRunner.Spec.Instructions.GpuOptimizations.Mnnvl).NotTo(BeNil())
				Expect(workloadRunner.Spec.Instructions.GpuOptimizations.Mnnvl.Mode).To(Equal(dromosv1alpha1.MnnvlModeNone))
				Expect(workloadRunner.Spec.Instructions.GpuOptimizations.Mnnvl.GpuNetworkAccelerationLabelKeys).To(BeNil())
			})

			It("should not add Instructions when capabilities is nil", func() {
				workloadRunner, err := wrapWithWorkloadRunner(manifest, dummyWorkloadId, nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(workloadRunner).NotTo(BeNil())
				Expect(workloadRunner.Spec.Instructions).To(BeNil())
			})

			It("should not add Instructions when capabilities.Mnnvl is nil", func() {
				capabilities := &workloadsManagerApi.WorkloadV2UpdateConfiguration{
					Mnnvl: nil,
				}

				workloadRunner, err := wrapWithWorkloadRunner(manifest, dummyWorkloadId, capabilities)

				Expect(err).NotTo(HaveOccurred())

				Expect(workloadRunner).NotTo(BeNil())
				Expect(workloadRunner.Spec.Instructions).To(BeNil())
			})

			It("should not add Instructions when capabilities.Mnnvl.Mode is nil", func() {
				capabilities := &workloadsManagerApi.WorkloadV2UpdateConfiguration{
					Mnnvl: &workloadsManagerApi.WorkloadV2UpdateConfigurationMnnvl{
						Mode: nil,
					},
				}

				workloadRunner, err := wrapWithWorkloadRunner(manifest, dummyWorkloadId, capabilities)

				Expect(err).NotTo(HaveOccurred())

				Expect(workloadRunner).NotTo(BeNil())
				Expect(workloadRunner.Spec.Instructions).To(BeNil())
			})
		})

		Context("Personal credentials translation", func() {
			var manifest *unstructured.Unstructured

			BeforeEach(func() {
				manifest = &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "batch/v1",
						"kind":       "Job",
						"metadata": map[string]interface{}{
							"name":      "test-job",
							"namespace": "test-namespace",
						},
					},
				}
			})

			It("stamps the annotation when sources are present", func() {
				capabilities := &workloadsManagerApi.WorkloadV2UpdateConfiguration{
					UserCredentials: *workloadsManagerApi.NewNullableWorkloadV2UpdateConfigurationUserCredentials(&workloadsManagerApi.WorkloadV2UpdateConfigurationUserCredentials{
						Sources: []workloadsManagerApi.WorkloadV2UpdateConfigurationUserCredentialSource{
							{Ref: "my-hf-token-abc1234567"},
						},
					}),
				}

				workloadRunner, err := wrapWithWorkloadRunner(manifest, dummyWorkloadId, capabilities)

				Expect(err).NotTo(HaveOccurred())

				Expect(workloadRunner).NotTo(BeNil())
				Expect(workloadRunner.Annotations).To(HaveKeyWithValue(
					pcreds.Annotation,
					"- sourceSecret: my-hf-token-abc1234567\n",
				))
			})

			It("sorts multiple sources for diff-stable output", func() {
				capabilities := &workloadsManagerApi.WorkloadV2UpdateConfiguration{
					UserCredentials: *workloadsManagerApi.NewNullableWorkloadV2UpdateConfigurationUserCredentials(&workloadsManagerApi.WorkloadV2UpdateConfigurationUserCredentials{
						Sources: []workloadsManagerApi.WorkloadV2UpdateConfigurationUserCredentialSource{
							{Ref: "c"},
							{Ref: "a"},
							{Ref: "b"},
						},
					}),
				}

				workloadRunner, err := wrapWithWorkloadRunner(manifest, dummyWorkloadId, capabilities)

				Expect(err).NotTo(HaveOccurred())

				Expect(workloadRunner.Annotations[pcreds.Annotation]).To(Equal(
					"- sourceSecret: a\n- sourceSecret: b\n- sourceSecret: c\n",
				))
			})

			It("does not set the annotation when UserCredentials is unset", func() {
				capabilities := &workloadsManagerApi.WorkloadV2UpdateConfiguration{}

				workloadRunner, err := wrapWithWorkloadRunner(manifest, dummyWorkloadId, capabilities)

				Expect(err).NotTo(HaveOccurred())

				Expect(workloadRunner.Annotations).NotTo(HaveKey(pcreds.Annotation))
			})

			It("does not set the annotation when Sources is empty", func() {
				capabilities := &workloadsManagerApi.WorkloadV2UpdateConfiguration{
					UserCredentials: *workloadsManagerApi.NewNullableWorkloadV2UpdateConfigurationUserCredentials(&workloadsManagerApi.WorkloadV2UpdateConfigurationUserCredentials{
						Sources: []workloadsManagerApi.WorkloadV2UpdateConfigurationUserCredentialSource{},
					}),
				}

				workloadRunner, err := wrapWithWorkloadRunner(manifest, dummyWorkloadId, capabilities)

				Expect(err).NotTo(HaveOccurred())

				Expect(workloadRunner.Annotations).NotTo(HaveKey(pcreds.Annotation))
			})

		})
	})

	Context("UnstructuredObjectInspector", func() {
		var inspector UnstructuredObjectInspector

		BeforeEach(func() {
			inspector = UnstructuredObjectInspector{}
		})

		Describe("IsUpToDate", func() {
			It("should return false when left object is nil", func() {
				right := &unstructured.Unstructured{}
				right.SetAnnotations(map[string]string{
					clusterCommonConstants.AnnotationsBackendWorkloadCounter: "10",
				})
				Expect(inspector.IsUpToDate(nil, right)).To(BeFalse())
			})

			It("should return false when right object is nil", func() {
				left := &unstructured.Unstructured{}
				left.SetAnnotations(map[string]string{
					clusterCommonConstants.AnnotationsBackendWorkloadCounter: "10",
				})
				Expect(inspector.IsUpToDate(left, nil)).To(BeFalse())
			})

			It("should return false when right object has no counter annotation", func() {
				left := &unstructured.Unstructured{}
				left.SetAnnotations(map[string]string{
					clusterCommonConstants.AnnotationsBackendWorkloadCounter: "10",
				})

				right := &unstructured.Unstructured{}
				right.SetAnnotations(map[string]string{})

				Expect(inspector.IsUpToDate(left, right)).To(BeFalse())
			})

			It("should return false when left object has no counter annotation", func() {
				left := &unstructured.Unstructured{}
				left.SetAnnotations(map[string]string{})

				right := &unstructured.Unstructured{}
				right.SetAnnotations(map[string]string{
					clusterCommonConstants.AnnotationsBackendWorkloadCounter: "10",
				})

				Expect(inspector.IsUpToDate(left, right)).To(BeFalse())
			})

			It("should return true when left counter is less than right counter", func() {
				left := &unstructured.Unstructured{}
				left.SetAnnotations(map[string]string{
					clusterCommonConstants.AnnotationsBackendWorkloadCounter: "5",
				})

				right := &unstructured.Unstructured{}
				right.SetAnnotations(map[string]string{
					clusterCommonConstants.AnnotationsBackendWorkloadCounter: "10",
				})

				Expect(inspector.IsUpToDate(left, right)).To(BeTrue())
			})

			It("should return true when left counter equals right counter", func() {
				left := &unstructured.Unstructured{}
				left.SetAnnotations(map[string]string{
					clusterCommonConstants.AnnotationsBackendWorkloadCounter: "10",
				})

				right := &unstructured.Unstructured{}
				right.SetAnnotations(map[string]string{
					clusterCommonConstants.AnnotationsBackendWorkloadCounter: "10",
				})

				Expect(inspector.IsUpToDate(left, right)).To(BeTrue())
			})

			It("should return false when left counter is greater than right counter", func() {
				left := &unstructured.Unstructured{}
				left.SetAnnotations(map[string]string{
					clusterCommonConstants.AnnotationsBackendWorkloadCounter: "15",
				})

				right := &unstructured.Unstructured{}
				right.SetAnnotations(map[string]string{
					clusterCommonConstants.AnnotationsBackendWorkloadCounter: "10",
				})

				Expect(inspector.IsUpToDate(left, right)).To(BeFalse())
			})

			It("should return false when right counter annotation is not a valid integer", func() {
				left := &unstructured.Unstructured{}
				left.SetAnnotations(map[string]string{
					clusterCommonConstants.AnnotationsBackendWorkloadCounter: "10",
				})

				right := &unstructured.Unstructured{}
				right.SetAnnotations(map[string]string{
					clusterCommonConstants.AnnotationsBackendWorkloadCounter: "invalid",
				})

				Expect(inspector.IsUpToDate(left, right)).To(BeFalse())
			})

			It("should return false when left counter annotation is not a valid integer", func() {
				left := &unstructured.Unstructured{}
				left.SetAnnotations(map[string]string{
					clusterCommonConstants.AnnotationsBackendWorkloadCounter: "invalid",
				})

				right := &unstructured.Unstructured{}
				right.SetAnnotations(map[string]string{
					clusterCommonConstants.AnnotationsBackendWorkloadCounter: "10",
				})

				Expect(inspector.IsUpToDate(left, right)).To(BeFalse())
			})
		})

		Describe("DeepCopySpec", func() {
			It("should copy spec from source to target", func() {
				from := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"spec": map[string]interface{}{
							"field1": "value1",
							"field2": int64(42),
						},
					},
				}

				to := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"spec": map[string]interface{}{
							"field1": "old-value",
						},
					},
				}

				inspector.DeepCopySpec(from, to)

				toSpec, found, err := unstructured.NestedFieldCopy(to.Object, "spec")
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
				specMap := toSpec.(map[string]interface{})
				Expect(specMap["field1"]).To(Equal("value1"))
				Expect(specMap["field2"]).To(Equal(int64(42)))
			})

			It("should copy annotations from source to target", func() {
				from := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"spec": map[string]interface{}{"dummy": "value"},
						"metadata": map[string]interface{}{
							"annotations": map[string]interface{}{
								"annotation1": "value1",
								"annotation2": "value2",
							},
						},
					},
				}

				to := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"spec": map[string]interface{}{"dummy": "old"},
						"metadata": map[string]interface{}{
							"annotations": map[string]interface{}{
								"old-annotation": "old-value",
							},
						},
					},
				}

				inspector.DeepCopySpec(from, to)

				Expect(to.GetAnnotations()).To(Equal(map[string]string{
					"annotation1": "value1",
					"annotation2": "value2",
				}))
			})

			It("should copy labels from source to target", func() {
				from := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"spec": map[string]interface{}{"dummy": "value"},
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"run.ai/category":   "inference",
								"priorityClassName": "high",
							},
						},
					},
				}

				to := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"spec": map[string]interface{}{"dummy": "old"},
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"old-label": "old-value",
							},
						},
					},
				}

				inspector.DeepCopySpec(from, to)

				Expect(to.GetLabels()).To(Equal(map[string]string{
					"run.ai/category":   "inference",
					"priorityClassName": "high",
				}))
			})

			It("should initialize labels map if nil on target", func() {
				from := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"spec": map[string]interface{}{"dummy": "value"},
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"run.ai/category": "training",
							},
						},
					},
				}

				to := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"spec": map[string]interface{}{"dummy": "old"},
					},
				}

				inspector.DeepCopySpec(from, to)

				Expect(to.GetLabels()).To(Equal(map[string]string{
					"run.ai/category": "training",
				}))
			})

			It("should handle category and priority label updates", func() {
				// Simulating an update where category and priority change
				from := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"spec": map[string]interface{}{"image": "new-image"},
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"run.ai/category":   "training",
								"priorityClassName": "low",
							},
						},
					},
				}

				to := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"spec": map[string]interface{}{"image": "old-image"},
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"run.ai/category":   "inference",
								"priorityClassName": "high",
							},
						},
					},
				}

				inspector.DeepCopySpec(from, to)

				// Verify labels were updated
				Expect(to.GetLabels()["run.ai/category"]).To(Equal("training"))
				Expect(to.GetLabels()["priorityClassName"]).To(Equal("low"))

				// Verify spec was also updated
				toSpec, _, _ := unstructured.NestedFieldCopy(to.Object, "spec")
				specMap := toSpec.(map[string]interface{})
				Expect(specMap["image"]).To(Equal("new-image"))
			})
		})
	})

	Context("ProcessWorkloads", func() {
		var (
			handler    *WorkloadsV2Handler
			workloads  []workloadsManagerApi.WorkloadV2Update
			wg         *sync.WaitGroup
			workloadId string
			projectId  string
			fakeClient client.Client
			u          *unstructured.Unstructured
		)

		BeforeEach(func() {
			workloadId = "test-workload-id"
			projectId = "test-project-id"

			// Create a scheme and add the required types
			scheme := runtime.NewScheme()
			err := runaiv2.AddToScheme(scheme)
			Expect(err).NotTo(HaveOccurred())
			// kai org units are listed by ProjectToNamespaceList
			err = kaiorg.AddToScheme(scheme)
			Expect(err).NotTo(HaveOccurred())
			// Add WorkloadRunner to scheme for submission tests
			err = dromosv1alpha1.AddToScheme(scheme)
			Expect(err).NotTo(HaveOccurred())

			// Create fake clients
			fakeClient = fake.NewClientBuilder().WithScheme(scheme).Build()

			// Create mock backend client
			mockBackendClient := &runaiBackendClient.BackendClient{}

			// Create handler with proper CommonHandler
			object := &unstructured.Unstructured{}
			objectInspector := UnstructuredObjectInspector{}
			handlingOptions := HandlingOptions{deletePropagationOrphan: false}

			commonHandler, err := NewUnstructuredHandler(mockBackendClient, fakeClient, fakeClient, object, objectInspector, handlingOptions)
			Expect(err).NotTo(HaveOccurred())

			handler = &WorkloadsV2Handler{
				CommonHandler: commonHandler,
				metricLabel:   "workloads_v2",
			}

			wg = &sync.WaitGroup{}

			u = &unstructured.Unstructured{}
			u.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "batch",
				Version: "v1",
				Kind:    "Job",
			})
		})

		It("should handle empty workload list", func() {
			workloads = []workloadsManagerApi.WorkloadV2Update{}
			wg.Add(1)

			perWorkloadErrors, err := handler.ProcessWorkloads(workloads, wg)

			// Wait group should be decremented
			// No error expected for empty list
			Expect(err).To(BeNil())
			Expect(perWorkloadErrors).To(HaveLen(0))
		})

		It("reports a project namespace lookup failure for every workload", func() {
			namespaceErr := fmt.Errorf("failed to list projects")
			handler.cachedClient = listErrorReader{Reader: handler.cachedClient, err: namespaceErr}
			workloads = []workloadsManagerApi.WorkloadV2Update{
				{Id: "workload-1"},
				{Id: "workload-2"},
			}

			clusterUpdates, perWorkloadErrors, err := handler.createClusterUpdates(workloads)

			Expect(err).To(MatchError(namespaceErr))
			Expect(perWorkloadErrors).To(ConsistOf(MatchError(namespaceErr), MatchError(namespaceErr)))
			Expect(clusterUpdates.ToUpdate).To(BeEmpty())
			Expect(clusterUpdates.ToDelete).To(BeEmpty())
		})

		It("creates WorkloadRunner objects without creating direct workloads", func() {
			runner := &dromosv1alpha1.WorkloadRunner{}
			err := fakeClient.Get(context.TODO(), client.ObjectKey{Name: "workload-1"}, runner)
			Expect(err).To(HaveOccurred())
			Expect(errors.IsNotFound(err)).To(BeTrue())

			err = fakeClient.Get(context.TODO(), client.ObjectKey{Name: "workload-2"}, runner)
			Expect(err).To(HaveOccurred())
			Expect(errors.IsNotFound(err)).To(BeTrue())

			workloads = []workloadsManagerApi.WorkloadV2Update{
				{
					Id:          workloadId + "-1",
					Name:        "workload-1",
					UpdateCount: 1,
					ProjectId:   &projectId,
					WorkloadManifest: map[string]interface{}{
						"apiVersion": "batch/v1",
						"kind":       "Job",
						"metadata": map[string]interface{}{
							"name": "workload-1",
						},
					},
				},
				{
					Id:          workloadId + "-2",
					Name:        "workload-2",
					UpdateCount: 2,
					ProjectId:   &projectId,
					WorkloadManifest: map[string]interface{}{
						"apiVersion": "batch/v1",
						"kind":       "Job",
						"metadata": map[string]interface{}{
							"name": "workload-2",
						},
					},
				},
			}
			wg.Add(1)

			perWorkloadErrors, processError := handler.ProcessWorkloads(workloads, wg)

			// Should have one error slot per workload
			Expect(perWorkloadErrors).To(HaveLen(2))
			Expect(processError).To(BeNil())

			err = fakeClient.Get(context.TODO(), client.ObjectKey{Name: "workload-1"}, runner)
			Expect(err).NotTo(HaveOccurred())
			Expect(runner.GetAnnotations()).To(HaveKey(clusterCommonConstants.AnnotationsBackendWorkloadCounter))
			Expect(runner.GetAnnotations()[clusterCommonConstants.AnnotationsBackendWorkloadCounter]).To(Equal("1"))

			runner = &dromosv1alpha1.WorkloadRunner{}
			err = fakeClient.Get(context.TODO(), client.ObjectKey{Name: "workload-2"}, runner)
			Expect(err).NotTo(HaveOccurred())
			Expect(runner.GetAnnotations()).To(HaveKey(clusterCommonConstants.AnnotationsBackendWorkloadCounter))
			Expect(runner.GetAnnotations()[clusterCommonConstants.AnnotationsBackendWorkloadCounter]).To(Equal("2"))

			err = fakeClient.Get(context.TODO(), client.ObjectKey{Name: "workload-1"}, u)
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		It("should handle workloads with DELETED desired phase", func() {
			u.SetName("workload-to-delete")
			err := fakeClient.Create(context.TODO(), u)
			Expect(err).NotTo(HaveOccurred())

			err = fakeClient.Get(context.TODO(), client.ObjectKey{Name: "workload-to-delete"}, u)
			Expect(err).NotTo(HaveOccurred())

			workloads = []workloadsManagerApi.WorkloadV2Update{
				{
					Id:           workloadId,
					Name:         "workload-to-delete",
					UpdateCount:  1,
					ProjectId:    &projectId,
					DesiredPhase: workloadsManagerApi.DESIREDPHASE_DELETED,
					Gvk: workloadsManagerApi.GVK{
						Group:   *workloadsManagerApi.NewNullableString(convutils.Ref("batch")),
						Version: "v1",
						Kind:    "Job",
					},
				},
			}
			wg.Add(1)

			perWorkloadErrors, processError := handler.ProcessWorkloads(workloads, wg)

			// Should initialize error slice for deletion
			Expect(perWorkloadErrors).To(HaveLen(1))
			Expect(processError).To(BeNil())

			err = fakeClient.Get(context.TODO(), client.ObjectKey{Name: "workload-to-delete"}, u)
			Expect(err).To(HaveOccurred())
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		It("reports an already absent workload as NotFound for tracking cleanup", func() {
			workloads = []workloadsManagerApi.WorkloadV2Update{{
				Id:           workloadId,
				Name:         "already-absent",
				UpdateCount:  1,
				ProjectId:    &projectId,
				DesiredPhase: workloadsManagerApi.DESIREDPHASE_DELETED,
				Gvk: workloadsManagerApi.GVK{
					Group:   *workloadsManagerApi.NewNullableString(convutils.Ref("batch")),
					Version: "v1",
					Kind:    "Job",
				},
			}}
			wg.Add(1)

			perWorkloadErrors, processError := handler.ProcessWorkloads(workloads, wg)

			Expect(processError).NotTo(HaveOccurred())
			Expect(perWorkloadErrors).To(HaveLen(1))
			Expect(errors.IsNotFound(perWorkloadErrors[0])).To(BeTrue())
			Expect(perWorkloadErrors[0]).To(MatchError(ContainSubstring(`Job.batch "already-absent" not found`)))
		})

		It("should handle mixed workloads (update and delete)", func() {
			workloads = []workloadsManagerApi.WorkloadV2Update{
				{
					Id:           workloadId + "-1",
					Name:         "workload-to-update",
					UpdateCount:  1,
					ProjectId:    &projectId,
					DesiredPhase: workloadsManagerApi.DESIREDPHASE_RUNNING,
					WorkloadManifest: map[string]interface{}{
						"apiVersion": "batch/v1",
						"kind":       "Job",
					},
				},
				{
					Id:           workloadId + "-2",
					Name:         "workload-to-delete",
					UpdateCount:  2,
					ProjectId:    &projectId,
					DesiredPhase: workloadsManagerApi.DESIREDPHASE_DELETED,
					Gvk: workloadsManagerApi.GVK{
						Group:   *workloadsManagerApi.NewNullableString(convutils.Ref("batch")),
						Version: "v1",
						Kind:    "Job",
					},
				},
			}
			wg.Add(1)

			perWorkloadErrors, _ := handler.ProcessWorkloads(workloads, wg)

			// Should handle both types
			Expect(perWorkloadErrors).To(HaveLen(2))
		})

		Context("WorkloadRunner submission", func() {
			It("reports WorkloadRunner conversion failure without creating a direct workload", func() {
				workloads = []workloadsManagerApi.WorkloadV2Update{{
					Id:          workloadId,
					Name:        "invalid-runner",
					UpdateCount: 1,
					ProjectId:   &projectId,
					WorkloadManifest: map[string]interface{}{
						"apiVersion": "batch/v1",
						"kind":       "Job",
					},
					Configuration: &workloadsManagerApi.WorkloadV2UpdateConfiguration{
						NetworkingEndpoints: []workloadsManagerApi.NetworkingEndpoint{{
							Services: []workloadsManagerApi.NetworkingService{{}},
						}},
					},
				}}
				wg.Add(1)

				perWorkloadErrors, processError := handler.ProcessWorkloads(workloads, wg)

				Expect(processError).To(HaveOccurred())
				Expect(perWorkloadErrors).To(HaveLen(1))
				Expect(perWorkloadErrors[0]).To(MatchError(ContainSubstring("failed to build workload runner")))

				runner := &dromosv1alpha1.WorkloadRunner{}
				err := fakeClient.Get(context.Background(), client.ObjectKey{Name: "invalid-runner"}, runner)
				Expect(errors.IsNotFound(err)).To(BeTrue())
				direct := &unstructured.Unstructured{}
				direct.SetGroupVersionKind(schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "Job"})
				err = fakeClient.Get(context.Background(), client.ObjectKey{Name: "invalid-runner"}, direct)
				Expect(errors.IsNotFound(err)).To(BeTrue())
			})

			It("stores the original submitter on the WorkloadRunner", func() {
				resource, err := handler.convertToClusterResource(workloadsManagerApi.WorkloadV2Update{
					Id:        workloadId,
					Name:      "runner-workload",
					CreatedBy: "owner-uid",
					WorkloadManifest: map[string]interface{}{
						"apiVersion": "batch/v1",
						"kind":       "Job",
					},
				}, "project-namespace")

				Expect(err).NotTo(HaveOccurred())
				Expect(resource.GetAnnotations()).To(HaveKeyWithValue(clusterCommonConstants.WorkloadOwnerUIDAnnotation, "owner-uid"))
			})

			It("does not initialize ownership when an older Control Plane omits the creator", func() {
				resource, err := handler.convertToClusterResource(workloadsManagerApi.WorkloadV2Update{
					Id:   workloadId,
					Name: "runner-workload",
					WorkloadManifest: map[string]interface{}{
						"apiVersion": "batch/v1",
						"kind":       "Job",
					},
				}, "project-namespace")

				Expect(err).NotTo(HaveOccurred())
				Expect(resource.GetAnnotations()).NotTo(HaveKey(clusterCommonConstants.WorkloadOwnerUIDAnnotation))
			})

			It("should handle workloads by creating workload runner object", func() {
				runnerWorkload := &dromosv1alpha1.WorkloadRunner{}
				err := fakeClient.Get(context.TODO(), client.ObjectKey{Name: "runner-workload"}, runnerWorkload)
				Expect(err).To(HaveOccurred())
				Expect(errors.IsNotFound(err)).To(BeTrue())

				workloads = []workloadsManagerApi.WorkloadV2Update{
					{
						Id:          workloadId,
						Name:        "runner-workload",
						UpdateCount: 1,
						ProjectId:   &projectId,
						WorkloadManifest: map[string]interface{}{
							"apiVersion": "batch/v1",
							"kind":       "Job",
						},
					},
				}
				wg.Add(1)

				perWorkloadErrors, processError := handler.ProcessWorkloads(workloads, wg)

				// Should process successfully even with wrapping
				Expect(perWorkloadErrors).To(HaveLen(1))
				Expect(processError).ToNot(HaveOccurred())

				err = fakeClient.Get(context.TODO(), client.ObjectKey{Name: "runner-workload"}, runnerWorkload)
				Expect(err).NotTo(HaveOccurred())
				Expect(runnerWorkload).NotTo(BeNil())
				byteArr, marshalErr := json.Marshal(workloads[0].WorkloadManifest)
				Expect(marshalErr).NotTo(HaveOccurred())
				Expect(runnerWorkload.Spec.Template.Raw).To(Equal(byteArr))
			})

			It("should handle mixed object type of objects", func() {
				err := fakeClient.Create(context.TODO(), &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "batch/v1",
						"kind":       "Job",
						"metadata": map[string]interface{}{
							"name": "unstructured-to-delete",
						},
					},
				})
				Expect(err).NotTo(HaveOccurred())
				err = fakeClient.Create(context.TODO(), &dromosv1alpha1.WorkloadRunner{
					TypeMeta: metav1.TypeMeta{
						Kind:       KindWorkloadRunner,
						APIVersion: ApiVersionWorkloadRunner,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "runner-to-delete",
					},
					Spec: dromosv1alpha1.WorkloadRunnerSpec{
						Template: runtime.RawExtension{
							Raw: []byte(`{"apiVersion":"batch/v1","kind":"Job","metadata":{"name":"runner-to-delete"}}`),
						},
					},
				})
				Expect(err).NotTo(HaveOccurred())

				workloads = []workloadsManagerApi.WorkloadV2Update{
					{
						Id:           workloadId + "-unstructured-delete",
						Name:         "unstructured-to-delete",
						UpdateCount:  1,
						ProjectId:    &projectId,
						DesiredPhase: workloadsManagerApi.DESIREDPHASE_DELETED,
						Gvk: workloadsManagerApi.GVK{
							Group:   *workloadsManagerApi.NewNullableString(convutils.Ref("batch")),
							Version: "v1",
							Kind:    "Job",
						},
						WorkloadManifest: map[string]interface{}{
							"apiVersion": "batch/v1",
							"kind":       "Job",
						},
					},
					{
						Id:           workloadId + "-runner-delete",
						Name:         "runner-to-delete",
						UpdateCount:  2,
						ProjectId:    &projectId,
						DesiredPhase: workloadsManagerApi.DESIREDPHASE_DELETED,
						Gvk: workloadsManagerApi.GVK{
							Group:   *workloadsManagerApi.NewNullableString(convutils.Ref("batch")),
							Version: "v1",
							Kind:    "Job",
						},
					},
				}
				wg.Add(1)

				perWorkloadErrors, _ := handler.ProcessWorkloads(workloads, wg)

				// Should handle both create and delete with wrapping
				Expect(perWorkloadErrors).To(HaveLen(2))
				err = fakeClient.Get(context.TODO(), client.ObjectKey{Name: "unstructured-to-delete"}, u)
				Expect(err).To(HaveOccurred())
				Expect(errors.IsNotFound(err)).To(BeTrue())
				runnerWorkload := &dromosv1alpha1.WorkloadRunner{}
				err = fakeClient.Get(context.TODO(), client.ObjectKey{Name: "runner-to-delete"}, runnerWorkload)
				Expect(err).To(HaveOccurred())
				Expect(errors.IsNotFound(err)).To(BeTrue())
			})
		})
	})
})
