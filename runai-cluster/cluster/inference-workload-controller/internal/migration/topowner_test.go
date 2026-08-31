/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	safeFake "github.com/run-ai/runai/runai-cluster/cluster/common/utils/client"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/utils/spec_checksum"
	register "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/utils/testing_utils"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	infp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	infw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	clustercommonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	knv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

var _ = Describe("TopOwnerMigrator tests", func() {
	var (
		tom                    *TopOwnerMigrator
		k8sClient              client.Client
		k8sMigrationClientFake CachedClient
		scheme                 *runtime.Scheme
	)
	BeforeEach(func() {
		setLimit(1)
	})
	AfterEach(func() {
		setLimit(paginationLimit)
	})
	scheme = runtime.NewScheme()
	register.InitSchema(scheme)

	BeforeEach(func() {
		k8sClientFake := fake.NewClientBuilder().WithScheme(scheme).WithIndex(&v1alpha1.ExternalWorkload{},
			ExternalWorkloadByOriginalUIDIndexerName, ExternalWorkloadByOriginalWorkloadUIDIndexer).Build()
		k8sClient = safeFake.NewFakeClientConcurrentSafe(k8sClientFake)

		k8sMigrationClientFake = NewMigrationFakeCachedClient(k8sClient)

		for _, obj := range createNamespaces() {
			Expect(k8sClient.Create(context.Background(), &obj)).To(BeNil())
		}

		tom = NewTopOwnerMigrator(k8sClient, k8sMigrationClientFake)
	})

	Context("setKnativeServiceOwnerReference", func() {
		It("should do nothing if knative service is not found", func() {
			inf1 := getInferenceWorkload(namespace1, "inf1")
			ksvc := getKnativeService(namespace1, "some-other-ksvc-that-should-not-be-found")
			createObject(k8sClient, ksvc)

			found, err := tom.setKnativeServiceOwnerReference(context.Background(), inf1)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeFalse())
		})

		It("should set owner if knative service is found", func() {
			inf1 := getInferenceWorkload(namespace1, "inf1")
			ksvc := getKnativeService(namespace1, "inf1")
			createObject(k8sClient, ksvc)

			found, err := tom.setKnativeServiceOwnerReference(context.Background(), inf1)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())

			ksvcUpdated := readKnativeService(k8sClient, ksvc)
			Expect(ksvcUpdated.GetOwnerReferences()).To(HaveLen(1))
			Expect(ksvcUpdated.GetOwnerReferences()[0].Name).To(Equal(inf1.GetName()))
		})
	})

	Context("InferenceWorkload migrations", func() {
		It("treats a missing InferenceWorkload API as an empty batch", func() {
			missingAPIClient := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
				List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
					return &meta.NoKindMatchError{GroupKind: infw.GroupVersion.WithKind("InferenceWorkload").GroupKind(), SearchedVersions: []string{infw.GroupVersion.Version}}
				},
			}).Build()
			missingAPIMigrator := NewTopOwnerMigrator(missingAPIClient, nil)

			batch, continueToken, err := missingAPIMigrator.getNextWorkloadBatch(context.Background(), namespace1, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(batch).To(BeEmpty())
			Expect(continueToken).To(BeEmpty())
		})

		It("returns authorization errors while listing inference workloads", func() {
			forbiddenErr := errors.NewForbidden(schema.GroupResource{Group: infw.GroupVersion.Group, Resource: "inferenceworkloads"}, "", fmt.Errorf("denied"))
			forbiddenClient := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
				List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
					return forbiddenErr
				},
			}).Build()
			forbiddenMigrator := NewTopOwnerMigrator(forbiddenClient, nil)

			_, _, err := forbiddenMigrator.getNextWorkloadBatch(context.Background(), namespace1, "")
			Expect(err).To(MatchError(forbiddenErr))
		})

		It("should remove the finalizer and not set spec-checksum if there's no knative service", func() {
			inf1 := getInferenceWorkload(namespace1, "inf1")
			setFinalizer(inf1)
			createObject(k8sClient, inf1)

			err := tom.migrateWorkload(context.Background(), inf1)
			Expect(err).ToNot(HaveOccurred())
			inf1Updated := readInferenceWorkload(k8sClient, inf1)
			Expect(inf1Updated.Finalizers).To(BeEmpty())  // finalizer removed
			Expect(inf1Updated.Annotations).To(BeEmpty()) // spec-checksum not added
		})

		It("should remove the finalizer, set owner references and set spec-checksum if there is a knative service", func() {
			inf1 := getInferenceWorkload(namespace1, "inf1")
			setFinalizer(inf1)
			createObject(k8sClient, inf1)

			ksvc := getKnativeService(namespace1, "inf1")
			createObject(k8sClient, ksvc)

			err := tom.migrateWorkload(context.Background(), inf1)
			Expect(err).ToNot(HaveOccurred())
			inf1Updated := readInferenceWorkload(k8sClient, inf1)
			ksvcUpdated := readKnativeService(k8sClient, ksvc)

			// finalizers should be empty
			Expect(inf1Updated.Finalizers).To(BeEmpty())

			// knative service should have its owner reference set correctly
			Expect(ksvcUpdated.GetOwnerReferences()).To(HaveLen(1))
			Expect(ksvcUpdated.GetOwnerReferences()[0].Name).To(Equal(inf1.GetName()))

			// spec-checksum annotation should be set with correct value
			expectedSpecChecksum, err := spec_checksum.CalculateSpecChecksum(inf1Updated)
			Expect(err).ToNot(HaveOccurred())
			Expect(inf1Updated.Annotations).To(HaveKeyWithValue(spec_checksum.SpecChecksumAnnotation, expectedSpecChecksum))
			Expect(inf1Updated.Annotations).To(HaveKeyWithValue(spec_checksum.SpecChecksumMigratedAnnotation, "true"))
		})

		It("should set service type to serving port if there's only one port, without a type", func() {
			inf1 := getInferenceWorkload(namespace1, "inf1")
			inf1.Spec.Ports.Deref()["serving-port"].Value.ServiceType = nil
			createObject(k8sClient, inf1)

			// adding a knative service in order to be able to check the spec-checksum as well (otherwise it's not calculated)
			ksvc := getKnativeService(namespace1, "inf1")
			setWorkloadLabels(ksvc, constants.KindInferenceWorkload, inf1.GetName())
			createObject(k8sClient, ksvc)

			err := tom.migrateWorkload(context.Background(), inf1)
			Expect(err).ToNot(HaveOccurred())
			inf1Updated := readInferenceWorkload(k8sClient, inf1)
			Expect(inf1Updated.Spec.Ports.Deref()["serving-port"].Value.ServiceType).To(Equal(ptr.To(constants.ServiceTypeServingPort)))

			// the spec-checksum should be calculated on the updated workload, with the serving port set
			expectedSpecChecksum, err := spec_checksum.CalculateSpecChecksum(inf1Updated)
			Expect(err).ToNot(HaveOccurred())
			Expect(inf1Updated.Annotations).To(HaveKeyWithValue(spec_checksum.SpecChecksumAnnotation, expectedSpecChecksum))
		})

		It("should set service type to serving port only to the right port, if there's more than one", func() {
			inf1 := getInferenceWorkload(namespace1, "inf1")
			inf1.Spec.Ports.Deref()["serving-port"].Value.ServiceType = nil
			inf1.Spec.Ports.Deref()["another-port"] = flds.PortItemAndRules{
				Value: &flds.PortItemValue{
					ServiceType:   ptr.To(constants.ServiceTypeLoadBalancer),
					ContainerPort: ptr.To(8081),
				},
			}
			createObject(k8sClient, inf1)

			err := tom.migrateWorkload(context.Background(), inf1)
			Expect(err).ToNot(HaveOccurred())
			inf1Updated := readInferenceWorkload(k8sClient, inf1)
			Expect(inf1Updated.Spec.Ports.Deref()["serving-port"].Value.ServiceType).To(Equal(ptr.To(constants.ServiceTypeServingPort)))
			Expect(inf1Updated.Spec.Ports.Deref()["another-port"].Value.ServiceType).To(Equal(ptr.To(constants.ServiceTypeLoadBalancer)))
		})

		It("should not set service type if it's already set", func() {
			inf1 := getInferenceWorkload(namespace1, "inf1")
			inf1.Spec.Ports.Deref()["another-port"] = flds.PortItemAndRules{
				Value: &flds.PortItemValue{
					ServiceType:   ptr.To(constants.ServiceTypeLoadBalancer),
					ContainerPort: ptr.To(8081),
				},
			}
			createObject(k8sClient, inf1)

			err := tom.migrateWorkload(context.Background(), inf1)
			Expect(err).ToNot(HaveOccurred())
			inf1Updated := readInferenceWorkload(k8sClient, inf1)
			Expect(inf1Updated.Spec.Ports.Deref()["serving-port"].Value.ServiceType).To(Equal(ptr.To(constants.ServiceTypeServingPort)))
			Expect(inf1Updated.Spec.Ports.Deref()["another-port"].Value.ServiceType).To(Equal(ptr.To(constants.ServiceTypeLoadBalancer)))
		})

		It("should not set service type if a default service type exists", func() {
			inf1 := getInferenceWorkload(namespace1, "inf1")
			inf1.Spec.Ports.Deref()["serving-port"].Value.ServiceType = nil
			inf1.Spec.ServiceType = &flds.StringField{
				Value: ptr.To(constants.ServiceTypeLoadBalancer),
			}
			createObject(k8sClient, inf1)

			err := tom.migrateWorkload(context.Background(), inf1)
			Expect(err).ToNot(HaveOccurred())
			inf1Updated := readInferenceWorkload(k8sClient, inf1)
			Expect(inf1Updated.Spec.Ports.Deref()["serving-port"].Value.ServiceType).To(BeNil())
		})
	})

	Context("ExternalWorkload migrations", func() {
		It("should delete the matching external workload", func() {
			inf1 := getInferenceWorkload(namespace1, "inf1")
			createObject(k8sClient, inf1)

			// create three pairs of ksvc and ew resources
			ksvc1 := getKnativeService(namespace1, "inf1")
			createObject(k8sClient, ksvc1)
			ksvc2 := getKnativeService(namespace1, "inf2")
			createObject(k8sClient, ksvc2)
			ksvc3 := getKnativeService(namespace2, "inf1")
			createObject(k8sClient, ksvc3)

			ew1 := getExternalWorkload(namespace1, "ew1", ksvc1)
			createObject(k8sClient, ew1)
			ew2 := getExternalWorkload(namespace1, "ew2", ksvc2)
			createObject(k8sClient, ew2)
			ew3 := getExternalWorkload(namespace2, "ew3", ksvc3)
			createObject(k8sClient, ew3)

			// migrate the first inf resource, only the relevant ew should be deleted
			err := tom.migrateWorkload(context.Background(), inf1)
			Expect(err).ToNot(HaveOccurred())

			_ = readExternalWorkload(k8sClient, ew1, true)
			_ = readExternalWorkload(k8sClient, ew2, false)
			_ = readExternalWorkload(k8sClient, ew3, false)
		})

		It("should replace the top-owner label for the matching pods", func() {
			inf1 := getInferenceWorkload(namespace1, "inf1")
			createObject(k8sClient, inf1)

			ksvc1 := getKnativeService(namespace1, "inf1")
			createObject(k8sClient, ksvc1)
			ksvc2 := getKnativeService(namespace2, "inf2")
			createObject(k8sClient, ksvc2)

			// create pods with and without the required labels to match,
			// in different namespaces

			// two pods of the infw, with the right label
			pod1 := getPod(namespace1, "inf1-1", ksvc1.GetUID())
			pod1.Labels = map[string]string{LabelInferenceWorkload: "inf1"}
			createObject(k8sClient, pod1)
			pod2 := getPod(namespace1, "inf1-2", ksvc1.GetUID())
			pod2.Labels = map[string]string{LabelInferenceWorkload: "inf1"}
			createObject(k8sClient, pod2)

			// pod without the label
			pod3 := getPod(namespace1, "inf1-3", ksvc1.GetUID())
			createObject(k8sClient, pod3)

			// pod with the label but different namespace
			pod4 := getPod(namespace2, "inf1-4", ksvc2.GetUID())
			pod2.Labels = map[string]string{LabelInferenceWorkload: "inf1"}
			createObject(k8sClient, pod4)

			// migrate the workload, only the relevant pods should be modified
			err := tom.migrateWorkload(context.Background(), inf1)
			Expect(err).ToNot(HaveOccurred())

			// only pod 1 & 2 should have the new uid, the rest should have the original one
			pod1Updated := readPod(k8sClient, pod1)
			Expect(pod1Updated.Labels[clustercommonconstants.TopOwnerUidLabel]).To(Equal(string(inf1.GetUID())))
			pod2Updated := readPod(k8sClient, pod2)
			Expect(pod2Updated.Labels[clustercommonconstants.TopOwnerUidLabel]).To(Equal(string(inf1.GetUID())))
			pod3Updated := readPod(k8sClient, pod3)
			Expect(pod3Updated.Labels[clustercommonconstants.TopOwnerUidLabel]).To(Equal(string(ksvc1.GetUID())))
			pod4Updated := readPod(k8sClient, pod3)
			Expect(pod4Updated.Labels[clustercommonconstants.TopOwnerUidLabel]).To(Equal(string(ksvc1.GetUID())))
		})

		It("should remove the external-workload-id annotation for the matching podgroup", func() {
			inf1 := getInferenceWorkload(namespace1, "inf1")
			createObject(k8sClient, inf1)

			ksvc1 := getKnativeService(namespace1, "inf1")
			createObject(k8sClient, ksvc1)
			ksvc2 := getKnativeService(namespace2, "inf2")
			createObject(k8sClient, ksvc2)

			// create podgroups with and without the required labels to match,
			// in different namespaces

			// podgroup of the infw, with the right label
			pg1 := getPodGroup(namespace1, "inf1-1")
			pg1.Labels = map[string]string{LabelInferenceWorkload: "inf1"}
			createObject(k8sClient, pg1)

			// podgroup without the label
			pg2 := getPodGroup(namespace1, "inf1-2")
			createObject(k8sClient, pg2)

			// podgroup with the label but different namespace
			pg3 := getPodGroup(namespace2, "inf1-3")
			pg3.Labels = map[string]string{LabelInferenceWorkload: "inf1"}
			createObject(k8sClient, pg3)

			// migrate the workload, only the relevant podgroup should be modified
			err := tom.migrateWorkload(context.Background(), inf1)
			Expect(err).ToNot(HaveOccurred())

			// only podgroup 1 should be modified
			pg1Updated := readPodGroup(k8sClient, pg1)
			Expect(pg1Updated.Annotations).NotTo(HaveKey(clustercommonconstants.ExternalWorkloadName))

			pg2Updated := readPodGroup(k8sClient, pg2)
			Expect(pg2Updated.Annotations).To(HaveKey(clustercommonconstants.ExternalWorkloadName))

			pg3Updated := readPodGroup(k8sClient, pg3)
			Expect(pg3Updated.Annotations).To(HaveKey(clustercommonconstants.ExternalWorkloadName))
		})
	})
})

func createObject(client client.Client, object client.Object) {
	Expect(client.Create(context.Background(), object)).To(BeNil())
}

func getInferenceWorkload(namespace, name string) *infw.InferenceWorkload {
	randomUuid, err := uuid.NewRandom()
	Expect(err).To(BeNil())

	return &infw.InferenceWorkload{
		TypeMeta: metav1.TypeMeta{
			Kind:       constants.KindInferenceWorkload,
			APIVersion: infw.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(randomUuid.String()),
		},
		Spec: infp.InferenceSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{
						Value: ptr.To("ubuntu"),
					},
				},
				NetworkingFields: field_groups.NetworkingFields{
					Ports: &flds.PortsField{
						Value: &map[string]flds.PortItemAndRules{
							"serving-port": {
								Value: &flds.PortItemValue{
									ServiceType:   ptr.To(constants.ServiceTypeServingPort),
									ContainerPort: ptr.To(8080),
								},
							},
						},
					},
				},
			},
		},
	}
}

func setFinalizer(workload *infw.InferenceWorkload) {
	workload.Finalizers = []string{"some-finalizer"}
}

func getKnativeService(namespace, name string) *knv1.Service {
	randomUuid, err := uuid.NewRandom()
	Expect(err).To(BeNil())

	return &knv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(randomUuid.String()),
		},
	}
}

func getExternalWorkload(namespace, name string, ksvc *knv1.Service) *v1alpha1.ExternalWorkload {
	return &v1alpha1.ExternalWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("ew-%s", name),
			Namespace: namespace,
		},
		Spec: v1alpha1.ExternalWorkloadSpec{
			OriginalWorkloadUID: ksvc.GetUID(),
		},
	}
}

func getPod(namespace, name string, topOwnerUid types.UID) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    map[string]string{clustercommonconstants.TopOwnerUidLabel: string(topOwnerUid)},
		},
	}
}

func getPodGroup(namespace, name string) *kaiv2alpha2.PodGroup {
	return &kaiv2alpha2.PodGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: map[string]string{clustercommonconstants.ExternalWorkloadName: fmt.Sprintf("ew-%s", name)},
		},
	}
}

func setWorkloadLabels(object client.Object, workloadKind, workloadName string) {
	labels := map[string]string{
		constants.LabelWorkloadKind: workloadKind,
		constants.LabelWorkloadName: workloadName,
	}

	object.SetLabels(labels)
}

func readInferenceWorkload(k8sClient client.Client, workload *infw.InferenceWorkload) *infw.InferenceWorkload {
	readWorkload := &infw.InferenceWorkload{}
	Expect(k8sClient.Get(context.Background(), client.ObjectKeyFromObject(workload), readWorkload)).To(BeNil())
	return readWorkload
}

func readKnativeService(k8sClient client.Client, ksvc *knv1.Service) *knv1.Service {
	readKsvc := &knv1.Service{}
	Expect(k8sClient.Get(context.Background(), client.ObjectKeyFromObject(ksvc), readKsvc)).To(BeNil())
	return readKsvc
}

func readExternalWorkload(k8sClient client.Client, ew *v1alpha1.ExternalWorkload, expectNotFound bool) *v1alpha1.ExternalWorkload {
	readEw := &v1alpha1.ExternalWorkload{}
	err := k8sClient.Get(context.Background(), client.ObjectKeyFromObject(ew), readEw)
	if expectNotFound {
		Expect(errors.IsNotFound(err)).To(BeTrue())
	} else {
		Expect(err).To(BeNil())
	}

	return readEw
}

func readPod(k8sClient client.Client, pod *v1.Pod) *v1.Pod {
	readPod := &v1.Pod{}
	Expect(k8sClient.Get(context.Background(), client.ObjectKeyFromObject(pod), readPod)).To(BeNil())
	return readPod
}

func readPodGroup(k8sClient client.Client, pod *kaiv2alpha2.PodGroup) *kaiv2alpha2.PodGroup {
	readPg := &kaiv2alpha2.PodGroup{}
	Expect(k8sClient.Get(context.Background(), client.ObjectKeyFromObject(pod), readPg)).To(BeNil())
	return readPg
}
