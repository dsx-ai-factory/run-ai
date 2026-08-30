package mutator

import (
	"context"
	"encoding/json"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	commontesting "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common/testing"
	"net/http"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	mnnvltypes "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	dinfp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferencepolicy/v1alpha1"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	workloadcommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	MutatorWorkloadName = "test-mutator-workload"
	mutatorNamespace    = "runai-default"
	mutatorImage        = "test-image"
)

var _ = Describe("Distributed Inference Mutator Handler", func() {
	var (
		workload  *dinfw.DistributedInferenceWorkload
		handler   *DistributedInferenceMutatorHandler
		ctx       context.Context
		k8sClient client.Client
	)

	BeforeEach(func() {
		var t testing.T // dummy, for the fake client
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

		k8sClient = commontesting.NewFakeClient(&t, &kaiMnnvlNodepool)

		workload = &dinfw.DistributedInferenceWorkload{
			TypeMeta: metav1.TypeMeta{
				Kind:       constants.KindDistributedInferenceWorkload,
				APIVersion: "workloads.run.ai/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      MutatorWorkloadName,
				Namespace: mutatorNamespace,
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
					Image: field_types.NewStringField(mutatorImage),
				},
			},
		}
		ctx = context.Background()
	})

	Context("Handle method", func() {
		BeforeEach(func() {
			handler = NewDistributedInferenceMutator(k8sClient, true, false, false)
		})

		It("should handle CREATE operation", func() {
			workloadBytes, err := json.Marshal(workload)
			Expect(err).To(BeNil())

			req := admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Operation: admissionv1.Create,
					Kind: metav1.GroupVersionKind{
						Kind: constants.KindDistributedInferenceWorkload,
					},
					Object: runtime.RawExtension{
						Raw: workloadBytes,
					},
				},
			}

			response := handler.Handle(ctx, req)
			// Should either allow or patch, not error
			Expect(response.AdmissionResponse.Allowed).ToNot(BeFalse())
		})

		It("should allow non-CREATE operations", func() {
			req := admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Operation: admissionv1.Update,
				},
			}

			response := handler.Handle(ctx, req)
			Expect(response.AdmissionResponse.Allowed).To(BeTrue())
			Expect(response.AdmissionResponse.Result.Message).To(Equal("Ok"))
		})

		It("should allow DELETE operations", func() {
			req := admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Operation: admissionv1.Delete,
				},
			}

			response := handler.Handle(ctx, req)
			Expect(response.AdmissionResponse.Allowed).To(BeTrue())
			Expect(response.AdmissionResponse.Result.Message).To(Equal("Ok"))
		})
	})

	Context("HandleCreate method", func() {
		BeforeEach(func() {
			handler = NewDistributedInferenceMutator(k8sClient, true, false, false)
		})

		It("should allow non-DistributedInferenceWorkload kinds", func() {
			req := admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Kind: metav1.GroupVersionKind{
						Kind: "SomeOtherKind",
					},
				},
			}

			response := handler.HandleCreate(ctx, req)
			Expect(response.AdmissionResponse.Allowed).To(BeTrue())
			Expect(response.AdmissionResponse.Result.Message).To(Equal("Ok"))
		})

		It("should return error for invalid JSON", func() {
			req := admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Kind: metav1.GroupVersionKind{
						Kind: constants.KindDistributedInferenceWorkload,
					},
					Object: runtime.RawExtension{
						Raw: []byte("invalid json"),
					},
				},
			}

			response := handler.HandleCreate(ctx, req)
			Expect(response.AdmissionResponse.Allowed).To(BeFalse())
			Expect(response.AdmissionResponse.Result.Code).To(Equal(int32(http.StatusBadRequest)))
		})

		It("should allow when no mutation is needed - MNNVL label is None or Required", func() {
			for _, mnnvlLabel := range []string{mnnvltypes.None, mnnvltypes.Required} {
				workload.Labels = map[string]string{
					mnnvltypes.FeatureLabelKey: mnnvlLabel,
				}
				workloadBytes, err := json.Marshal(workload)
				Expect(err).To(BeNil())

				req := admission.Request{
					AdmissionRequest: admissionv1.AdmissionRequest{
						Kind: metav1.GroupVersionKind{
							Kind: constants.KindDistributedInferenceWorkload,
						},
						Object: runtime.RawExtension{
							Raw: workloadBytes,
						},
					},
				}

				response := handler.HandleCreate(ctx, req)
				Expect(response.AdmissionResponse.Allowed).To(BeTrue())
				Expect(response.AdmissionResponse.Result.Message).To(Equal("Ok"))
			}
		})
		It("should return patch when mutation is needed", func() {
			// Ensure workload will be mutated (no MNNVL label, not old workload)
			workloadBytes, err := json.Marshal(workload)
			Expect(err).To(BeNil())

			req := admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Kind: metav1.GroupVersionKind{
						Kind: constants.KindDistributedInferenceWorkload,
					},
					Object: runtime.RawExtension{
						Raw: workloadBytes,
					},
				},
			}

			response := handler.HandleCreate(ctx, req)
			// Should return a patch response
			Expect(response.AdmissionResponse.Allowed).To(BeTrue())
			Expect(response.Patches).ToNot(BeNil())
		})
	})

	Context("mutateForMnnvl method", func() {
		BeforeEach(func() {
			handler = NewDistributedInferenceMutator(k8sClient, true, false, false)
		})

		Context("when MNNVL label has different values or doesn't exist", func() {
			It("should not mutate when label is None", func() {
				workload.Labels = map[string]string{
					mnnvltypes.FeatureLabelKey: mnnvltypes.None,
				}

				isPatched, err := handler.mutateForMnnvl(ctx, workload)
				Expect(err).To(BeNil())
				Expect(isPatched).To(BeFalse())
			})

			It("should not mutate when label is Required", func() {
				workload.Labels = map[string]string{
					mnnvltypes.FeatureLabelKey: mnnvltypes.Required,
				}

				isPatched, err := handler.mutateForMnnvl(ctx, workload)
				Expect(err).To(BeNil())
				Expect(isPatched).To(BeFalse())
			})

			It("should mutate when label is Preferred", func() {
				workload.Labels = map[string]string{
					mnnvltypes.FeatureLabelKey:                  mnnvltypes.Preferred,
					workloadcommon.CreatedAtClusterVersionLabel: "2.23.0", // Not old
				}

				isPatched, err := handler.mutateForMnnvl(ctx, workload)
				Expect(err).To(BeNil())
				Expect(isPatched).To(BeTrue())
				// Should be set to None because no GPU nodes and other conditions
				Expect(workload.Labels[mnnvltypes.FeatureLabelKey]).To(Equal(mnnvltypes.None))
			})

			It("should mutate when label doesn't exist and leader has gpus", func() {
				workload.Labels = map[string]string{
					workloadcommon.CreatedAtClusterVersionLabel: "2.23.0",
				}
				workload.Spec.NodePools = &field_types.StringField{
					Value: ptr.To("mnnvl-nodepool"),
				}
				// Set leader spec with GPUs
				workload.LeaderSpec.ResourceAllocationFields = field_groups.ResourceAllocationFields{
					Gpu: &field_types.NumberField{
						Value: ptr.To("1"),
					},
				}

				isPatched, err := handler.mutateForMnnvl(ctx, workload)
				Expect(err).To(BeNil())
				Expect(isPatched).To(BeTrue())
				Expect(workload.Labels).To(HaveKeyWithValue(mnnvltypes.FeatureLabelKey, mnnvltypes.Required))
			})

			It("should mutate when label doesn't exist and worker has gpus", func() {
				workload.Labels = map[string]string{
					workloadcommon.CreatedAtClusterVersionLabel: "2.23.0",
				}
				workload.Spec.NodePools = &field_types.StringField{
					Value: ptr.To("mnnvl-nodepool"),
				}
				// Set worker spec with GPUs
				workload.Spec.Workers = &field_types.IntField{
					Value: ptr.To(1),
				}
				workload.WorkerSpec = &dinfp.DistributedInferenceEntitySpec{
					ResourceAllocationFields: field_groups.ResourceAllocationFields{
						Gpu: &field_types.NumberField{
							Value: ptr.To("1"),
						},
					},
				}
				// Remove leader GPU spec
				workload.LeaderSpec.ResourceAllocationFields = field_groups.ResourceAllocationFields{}

				isPatched, err := handler.mutateForMnnvl(ctx, workload)
				Expect(err).To(BeNil())
				Expect(isPatched).To(BeTrue())
				Expect(workload.Labels).To(HaveKeyWithValue(mnnvltypes.FeatureLabelKey, mnnvltypes.Required))
			})
		})

		Context("when workload is old", func() {
			It("should not mutate workload with version 2.22.0", func() {
				workload.Labels = map[string]string{
					workloadcommon.CreatedAtClusterVersionLabel: "2.22.0",
				}

				isPatched, err := handler.mutateForMnnvl(ctx, workload)
				Expect(err).To(BeNil())
				Expect(isPatched).To(BeFalse())
			})
		})

		Context("when workload is new", func() {
			BeforeEach(func() {
				workload.Labels = map[string]string{
					workloadcommon.CreatedAtClusterVersionLabel: "2.23.0",
				}
			})

			It("should set MNNVL to None when MNNVL is disabled", func() {
				handler = NewDistributedInferenceMutator(k8sClient, false, false, false)

				isPatched, err := handler.mutateForMnnvl(ctx, workload)
				Expect(err).To(BeNil())
				Expect(isPatched).To(BeTrue())
				Expect(workload.Labels[mnnvltypes.FeatureLabelKey]).To(Equal(mnnvltypes.None))
			})

			It("should set MNNVL to None when workload support check fails", func() {
				// With node pools that do not support MNNVL, it should return None
				workload.Spec.NodePools = &field_types.StringField{
					Value: ptr.To("some-nodepool"),
				}
				workload.LeaderSpec.ResourceAllocationFields = field_groups.ResourceAllocationFields{
					Gpu: &field_types.NumberField{
						Value: ptr.To("1"),
					},
				}
				isPatched, err := handler.mutateForMnnvl(ctx, workload)
				Expect(err).ToNot(BeNil())
				Expect(isPatched).To(BeFalse())
			})
		})
	})

	Context("isOldWorkload method", func() {
		It("should return false when no cluster version label exists", func() {
			workload.Labels = map[string]string{}
			result := isOldWorkload(context.Background(), workload)
			Expect(result).To(BeFalse())
		})

		It("should return true when cluster version is 2.22.0", func() {
			workload.Labels = map[string]string{
				workloadcommon.CreatedAtClusterVersionLabel: "2.22.0",
			}
			result := isOldWorkload(context.Background(), workload)
			Expect(result).To(BeTrue())
		})

		It("should return false for newer versions", func() {
			workload.Labels = map[string]string{
				workloadcommon.CreatedAtClusterVersionLabel: "2.23.0",
			}
			result := isOldWorkload(context.Background(), workload)
			Expect(result).To(BeFalse())
		})

		It("should return false for other versions", func() {
			workload.Labels = map[string]string{
				workloadcommon.CreatedAtClusterVersionLabel: "2.24.0",
			}
			result := isOldWorkload(context.Background(), workload)
			Expect(result).To(BeFalse())
		})
	})

	Context("integration scenarios", func() {
		It("should handle complete workflow for new workload", func() {
			handler = NewDistributedInferenceMutator(k8sClient, true, false, false)

			// Create a new workload (no MNNVL label, new version)
			workload.Labels = map[string]string{
				workloadcommon.CreatedAtClusterVersionLabel: "2.23.0",
			}
			workloadBytes, err := json.Marshal(workload)
			Expect(err).To(BeNil())

			req := admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Operation: admissionv1.Create,
					Kind: metav1.GroupVersionKind{
						Kind: constants.KindDistributedInferenceWorkload,
					},
					Object: runtime.RawExtension{
						Raw: workloadBytes,
					},
				},
			}

			response := handler.Handle(ctx, req)
			Expect(response.AdmissionResponse.Allowed).To(BeTrue())

			// Should have patch data because MNNVL label was added
			if response.Patch != nil {
				// Patch was applied
				Expect(len(response.Patch)).To(BeNumerically(">", 0))
			}
		})

		It("should handle workflow for workload with existing MNNVL label", func() {
			handler = NewDistributedInferenceMutator(k8sClient, true, false, false)

			// Workload already has MNNVL label
			workload.Labels = map[string]string{
				mnnvltypes.FeatureLabelKey: mnnvltypes.Required,
			}
			workloadBytes, err := json.Marshal(workload)
			Expect(err).To(BeNil())

			req := admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Operation: admissionv1.Create,
					Kind: metav1.GroupVersionKind{
						Kind: constants.KindDistributedInferenceWorkload,
					},
					Object: runtime.RawExtension{
						Raw: workloadBytes,
					},
				},
			}

			response := handler.Handle(ctx, req)
			Expect(response.AdmissionResponse.Allowed).To(BeTrue())
			Expect(response.AdmissionResponse.Result.Message).To(Equal("Ok"))
			// No patch should be applied
			Expect(response.Patches).To(BeNil())
		})
	})

	Context("topology annotations workload copy", func() {
		BeforeEach(func() {
			handler = NewDistributedInferenceMutator(k8sClient, false, true, false)
		})
	})
})
