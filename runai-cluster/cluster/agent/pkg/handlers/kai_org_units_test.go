/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	_ "github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/tests/helper"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	orgUnitApi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	runaiBackendClient "github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/backend-client"
	v1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	v2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"
	schedv2alpha2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// --- shared test helpers (package-wide) ------------------------------------------------

// kaiTestScheme returns a scheme registering the legacy and KAI org-unit CRD types.
func kaiTestScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	Expect(v2.AddToScheme(scheme)).To(Succeed())
	Expect(schedv2alpha2.AddToScheme(scheme)).To(Succeed())
	Expect(kaiorg.AddToScheme(scheme)).To(Succeed())
	return scheme
}

// fakeClientWithObjects builds a fake client (scheme with all org-unit CRDs) seeded with objs.
func fakeClientWithObjects(objs ...client.Object) client.Client {
	return fake.NewClientBuilder().
		WithScheme(kaiTestScheme()).
		WithObjects(objs...).
		Build()
}

func kaiDepartmentObj(name, id string) *kaiorg.Department {
	return &kaiorg.Department{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{schedv2alpha2.DepartmentIdLabel: id},
		},
	}
}

// resourcesForNodePool builds a single-nodepool ResourcesSync. Cpu/Memory are left unset so the
// conversion applies the defaults (deserved -1, overQuotaWeight 1, limit -1).
func resourcesForNodePool(nodePoolName string, deservedGPU float64) orgUnitApi.ResourcesSync {
	return orgUnitApi.ResourcesSync{
		NodePool: *orgUnitApi.NewNullableResourcesNodePool(&orgUnitApi.ResourcesNodePool{
			Id:   nodePoolName,
			Name: ptr.To(nodePoolName),
		}),
		Gpu:    orgUnitApi.ResourcesSyncGpu{Deserved: deservedGPU},
		Cpu:    *orgUnitApi.NewNullableResourcesCpu(nil),
		Memory: *orgUnitApi.NewNullableResourcesSyncMemory(nil),
		Rank:   ptr.To(int32(50)),
	}
}

var _ = Describe("KAI org-unit shared helpers", func() {

	Describe("convertToKaiQueueConfigs", func() {
		It("names a project queue after the project name for the default nodepool", func() {
			queues := convertToKaiQueueConfigs("proj-1",
				[]orgUnitApi.ResourcesSync{resourcesForNodePool(v1alpha1.DefaultNodePoolName, 2)})

			Expect(queues).To(HaveLen(1))
			Expect(queues[0].Name).To(Equal("proj-1"))
			Expect(queues[0].Nodepool).To(Equal(v1alpha1.DefaultNodePoolName))
			Expect(*queues[0].Priority).To(Equal(int32(50)))
		})

		It("appends the nodepool suffix for a non-default nodepool", func() {
			queues := convertToKaiQueueConfigs("proj-1",
				[]orgUnitApi.ResourcesSync{resourcesForNodePool("gpu-pool", 2)})

			Expect(queues).To(HaveLen(1))
			Expect(queues[0].Name).To(Equal("proj-1-gpu-pool"))
		})

		It("builds an id-based q-<id> department queue name (default and non-default nodepools)", func() {
			queues := convertToKaiQueueConfigs("q-dep-1", []orgUnitApi.ResourcesSync{
				resourcesForNodePool(v1alpha1.DefaultNodePoolName, 1),
				resourcesForNodePool("gpu-pool", 1),
			})

			Expect(queues).To(HaveLen(2))
			Expect(queues[0].Name).To(Equal("q-dep-1"))
			Expect(queues[1].Name).To(Equal("q-dep-1-gpu-pool"))
		})

		It("maps the nodepool resources into the KAI queue resources", func() {
			queues := convertToKaiQueueConfigs("proj-1",
				[]orgUnitApi.ResourcesSync{resourcesForNodePool(v1alpha1.DefaultNodePoolName, 3)})

			res := queues[0].Resources
			Expect(res).ToNot(BeNil())
			Expect(res.GPU).To(Equal(kaiorg.SystemResource{Deserved: 3, OverQuotaWeight: DefaultOverQuotaWeight, Limit: NoMaxAllowed}))
			Expect(res.CPU).To(Equal(kaiorg.SystemResource{Deserved: NoMaxAllowed, OverQuotaWeight: DefaultOverQuotaWeight, Limit: NoMaxAllowed}))
			Expect(res.Memory).To(Equal(kaiorg.SystemResource{Deserved: NoMaxAllowed, OverQuotaWeight: DefaultOverQuotaWeight, Limit: NoMaxAllowed}))
		})

		It("returns an empty (non-nil) slice when there are no resources", func() {
			Expect(convertToKaiQueueConfigs("proj-1", nil)).To(BeEmpty())
		})
	})

	Describe("convertToKaiQueueResources", func() {
		It("returns nil when the source resources are nil", func() {
			Expect(convertToKaiQueueResources(nil)).To(BeNil())
		})
	})

	Describe("copyOrgUnitIdLabel", func() {
		It("copies a present label, creating the target label map when nil", func() {
			from := &kaiorg.Project{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{v2.ProjectIdLabel: "p-1"}}}
			to := &kaiorg.Project{}

			copyOrgUnitIdLabel(from, to, v2.ProjectIdLabel)
			Expect(to.Labels).To(HaveKeyWithValue(v2.ProjectIdLabel, "p-1"))
		})

		It("is a no-op when the label is absent on the source", func() {
			from := &kaiorg.Project{}
			to := &kaiorg.Project{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}}}

			copyOrgUnitIdLabel(from, to, v2.ProjectIdLabel)
			Expect(to.Labels).ToNot(HaveKey(v2.ProjectIdLabel))
		})
	})

	Describe("specJsonEqual", func() {
		It("is true for equal specs and false for different specs", func() {
			a := kaiorg.ProjectSpec{Parent: "dep", Namespace: "ns"}
			b := kaiorg.ProjectSpec{Parent: "dep", Namespace: "ns"}
			c := kaiorg.ProjectSpec{Parent: "dep", Namespace: "other"}

			Expect(specJsonEqual(a, b)).To(BeTrue())
			Expect(specJsonEqual(a, c)).To(BeFalse())
		})
	})

	Describe("idLabelEqual", func() {
		It("matches only when both presence and value are equal", func() {
			Expect(idLabelEqual(map[string]string{"k": "v"}, map[string]string{"k": "v"}, "k")).To(BeTrue())
			Expect(idLabelEqual(map[string]string{"k": "v"}, map[string]string{"k": "x"}, "k")).To(BeFalse())
			Expect(idLabelEqual(map[string]string{"k": "v"}, map[string]string{}, "k")).To(BeFalse())
			Expect(idLabelEqual(map[string]string{}, map[string]string{}, "k")).To(BeTrue())
		})
	})

	Describe("copyManagedAnnotations", func() {
		It("copies managed keys from source and removes stale managed keys on the target", func() {
			from := &kaiorg.Project{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				constants.TrainingJobTimeLimitSecondsAnnotation: "40",
			}}}
			to := &kaiorg.Project{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				constants.ForceDeleteAnnotation: "true", // stale managed key not present on source -> removed
				"user/keep":                     "yes",  // unmanaged key -> untouched
			}}}

			copyManagedAnnotations(from, to, kaiProjectManagedAnnotations)

			Expect(to.Annotations).To(HaveKeyWithValue(constants.TrainingJobTimeLimitSecondsAnnotation, "40"))
			Expect(to.Annotations).ToNot(HaveKey(constants.ForceDeleteAnnotation))
			Expect(to.Annotations).To(HaveKeyWithValue("user/keep", "yes"))
		})
	})

	Describe("managedAnnotationsEqual", func() {
		It("compares only the managed keys", func() {
			left := map[string]string{constants.TrainingJobTimeLimitSecondsAnnotation: "40", "user/x": "1"}
			right := map[string]string{constants.TrainingJobTimeLimitSecondsAnnotation: "40", "user/x": "2"}
			Expect(managedAnnotationsEqual(left, right, kaiProjectManagedAnnotations)).To(BeTrue())

			right[constants.ForceDeleteAnnotation] = "true"
			Expect(managedAnnotationsEqual(left, right, kaiProjectManagedAnnotations)).To(BeFalse())
		})
	})

	Describe("KaiProjectObjectInspector", func() {
		inspector := KaiProjectObjectInspector(true)

		newProject := func() *kaiorg.Project {
			return &kaiorg.Project{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						v2.ProjectIdLabel:               "p-1",
						schedv2alpha2.DepartmentIdLabel: "d-1",
					},
					Annotations: map[string]string{constants.TrainingJobTimeLimitSecondsAnnotation: "40"},
				},
				Spec: kaiorg.ProjectSpec{Parent: "dep", Namespace: "ns"},
			}
		}

		It("IsUpToDate is true when spec, managed annotations and id labels match", func() {
			Expect(inspector.IsUpToDate(newProject(), newProject())).To(BeTrue())
		})

		It("IsUpToDate is false on spec drift", func() {
			right := newProject()
			right.Spec.Namespace = "changed"
			Expect(inspector.IsUpToDate(newProject(), right)).To(BeFalse())
		})

		It("IsUpToDate is false on managed-annotation drift", func() {
			right := newProject()
			right.Annotations[constants.TrainingJobTimeLimitSecondsAnnotation] = "99"
			Expect(inspector.IsUpToDate(newProject(), right)).To(BeFalse())
		})

		It("IsUpToDate is false on id-label drift", func() {
			right := newProject()
			right.Labels[schedv2alpha2.DepartmentIdLabel] = "other"
			Expect(inspector.IsUpToDate(newProject(), right)).To(BeFalse())
		})

		It("DeepCopySpec copies spec, id labels and managed annotations", func() {
			from := newProject()
			to := &kaiorg.Project{}

			inspector.DeepCopySpec(from, to)
			Expect(to.Spec).To(Equal(from.Spec))
			Expect(to.Labels).To(HaveKeyWithValue(v2.ProjectIdLabel, "p-1"))
			Expect(to.Labels).To(HaveKeyWithValue(schedv2alpha2.DepartmentIdLabel, "d-1"))
			Expect(to.Annotations).To(HaveKeyWithValue(constants.TrainingJobTimeLimitSecondsAnnotation, "40"))
		})

		It("is a safe no-op for a wrong object type", func() {
			Expect(func() { inspector.DeepCopySpec(&kaiorg.Department{}, &kaiorg.Project{}) }).ToNot(Panic())
			Expect(inspector.IsUpToDate(&kaiorg.Department{}, &kaiorg.Project{})).To(BeFalse())
		})
	})

	Describe("KaiDepartmentObjectInspector", func() {
		inspector := KaiDepartmentObjectInspector(true)

		newDepartment := func() *kaiorg.Department {
			return &kaiorg.Department{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{schedv2alpha2.DepartmentIdLabel: "d-1"}},
				Spec:       kaiorg.DepartmentSpec{Queues: []kaiorg.QueueConfig{{Name: "q-d-1"}}},
			}
		}

		It("IsUpToDate is true for equal departments", func() {
			Expect(inspector.IsUpToDate(newDepartment(), newDepartment())).To(BeTrue())
		})

		It("IsUpToDate is false on spec drift", func() {
			right := newDepartment()
			right.Spec.Queues[0].Name = "changed"
			Expect(inspector.IsUpToDate(newDepartment(), right)).To(BeFalse())
		})

		It("IsUpToDate is false on id-label drift", func() {
			right := newDepartment()
			right.Labels[schedv2alpha2.DepartmentIdLabel] = "other"
			Expect(inspector.IsUpToDate(newDepartment(), right)).To(BeFalse())
		})

		It("DeepCopySpec copies spec and the department-id label", func() {
			from := newDepartment()
			to := &kaiorg.Department{}

			inspector.DeepCopySpec(from, to)
			Expect(to.Spec).To(Equal(from.Spec))
			Expect(to.Labels).To(HaveKeyWithValue(schedv2alpha2.DepartmentIdLabel, "d-1"))
		})
	})

	Describe("handler construction", func() {
		build := func() (client.Client, *runaiBackendClient.BackendClient) {
			c := fakeClientWithObjects()
			return c, &runaiBackendClient.BackendClient{}
		}

		It("NewProjectsHandler uses the KAI Project", func() {
			c, backend := build()
			handler, err := NewProjectsHandler(backend, c, c)
			Expect(err).ToNot(HaveOccurred())
			Expect(handler.(*ProjectsHandler).gvk.Group).To(Equal(kaiorg.GroupVersion.Group))
			Expect(handler.(*ProjectsHandler).gvk.Kind).To(Equal("Project"))
		})

		It("NewDepartmentsHandler uses the KAI Department", func() {
			c, backend := build()
			handler, err := NewDepartmentsHandler(backend, c, c)
			Expect(err).ToNot(HaveOccurred())
			Expect(handler.(*DepartmentsHandler).gvk.Group).To(Equal(kaiorg.GroupVersion.Group))
			Expect(handler.(*DepartmentsHandler).gvk.Kind).To(Equal("Department"))
		})
	})
})
