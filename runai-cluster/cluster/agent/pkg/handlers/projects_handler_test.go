/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"time"

	_ "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/tests/helper"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	v2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	schedv2alpha2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// projectWithParent and deletedProject are shared by the kai (this file) and legacy
// (projects_handler_legacy_test.go) specs; they live here so they survive removal of the flag.
func projectWithParent(id, name, depId, depName string) orgUnitApi.ProjectSync {
	return orgUnitApi.ProjectSync{
		Id:   id,
		Name: name,
		Parent: *orgUnitApi.NewNullableDepartmentProjectParent(&orgUnitApi.DepartmentProjectParent{
			Id:   depId,
			Name: depName,
		}),
	}
}

func deletedProject(id, name, depId, depName string) orgUnitApi.ProjectSync {
	project := projectWithParent(id, name, depId, depName)
	project.DeletedAt = *orgUnitApi.NewNullableTime(ptr.To(time.Now()))
	return project
}

// newProjectsHandlerWithKaiDepartments builds a ProjectsHandler backed by a fake client seeded
// with the given KAI departments (used for the parent-validation path).
func newProjectsHandlerWithKaiDepartments(deps ...*kaiorg.Department) *ProjectsHandler {
	objs := make([]client.Object, 0, len(deps))
	for _, d := range deps {
		objs = append(objs, d)
	}
	return &ProjectsHandler{CommonHandler: &CommonHandler{cachedClient: fakeClientWithObjects(objs...)}}
}

// orgUnitProjectFull builds a ProjectSync populated with all fields that map onto a KAI Project
// (scheduling rules, node types, enforce flag, requested namespace and a single default-nodepool queue).
func orgUnitProjectFull(id, name, depId, depName string) orgUnitApi.ProjectSync {
	project := projectWithParent(id, name, depId, depName)
	project.RequestedNamespace = *orgUnitApi.NewNullableString(ptr.To("ns-" + name))
	project.EnforceRunaiScheduler = ptr.To(true)
	project.SchedulingRules = *orgUnitApi.NewNullableSchedulingRules(&orgUnitApi.SchedulingRules{
		InteractiveJobTimeLimitSeconds:           *orgUnitApi.NewNullableInt32(ptr.To(int32(10))),
		InteractiveJobMaxIdleDurationSeconds:     *orgUnitApi.NewNullableInt32(ptr.To(int32(20))),
		InteractiveJobPreemptIdleDurationSeconds: *orgUnitApi.NewNullableInt32(ptr.To(int32(30))),
		TrainingJobTimeLimitSeconds:              *orgUnitApi.NewNullableInt32(ptr.To(int32(40))),
		TrainingJobMaxIdleDurationSeconds:        *orgUnitApi.NewNullableInt32(ptr.To(int32(50))),
	})
	project.NodeTypes = &orgUnitApi.NodeTypesPerWorkload{
		Workspace: []string{"w1"},
		Training:  []string{"t1"},
		Names:     map[string]string{"w1": "gpu-a", "t1": "gpu-b"},
	}
	project.Resources = []orgUnitApi.ResourcesSync{resourcesForNodePool(v1alpha1.DefaultNodePoolName, 2)}
	project.DefaultNodePools = []string{"np-a", "np-b"}
	return project
}

func withProjectDeletion(project orgUnitApi.ProjectSync, force bool, deletionType orgUnitApi.DeletionTypes) orgUnitApi.ProjectSync {
	project.Deletion = *orgUnitApi.NewNullableDeletion(orgUnitApi.NewDeletion(deletionType, time.Now(), force))
	return project
}

var _ = Describe("ProjectsHandler KAI conversion", func() {
	Describe("convertOrgUnitToKaiProject", func() {
		It("maps identity, spec and product-specific annotations", func() {
			project, err := convertOrgUnitToKaiProject(orgUnitProjectFull("p-1", "proj-1", "d-1", "dep-1"))
			Expect(err).ToNot(HaveOccurred())

			Expect(project.Name).To(Equal("proj-1"))
			Expect(project.Labels).To(HaveKeyWithValue(v2.ProjectIdLabel, "p-1"))
			Expect(project.Labels).To(HaveKeyWithValue(schedv2alpha2.DepartmentIdLabel, "d-1"))
			Expect(project.Spec.Parent).To(Equal("dep-1"))
			Expect(project.Spec.Namespace).To(Equal("ns-proj-1"))
			Expect(project.Spec.EnforceKaiScheduler).To(BeTrue())
			Expect(project.Spec.Queues).To(HaveLen(1))
			Expect(project.Spec.Queues[0].Name).To(Equal("proj-1"))
			Expect(project.Spec.DefaultNodePools).To(Equal([]string{"np-a", "np-b"}))

			Expect(project.Annotations).To(HaveKeyWithValue(constants.TrainingJobTimeLimitSecondsAnnotation, "40"))
			Expect(project.Annotations).To(HaveKeyWithValue(constants.TrainingJobMaxIdleDurationSecondsAnnotation, "50"))
			Expect(project.Annotations).To(HaveKeyWithValue(constants.InteractiveJobTimeLimitSecondsAnnotation, "10"))
			Expect(project.Annotations).To(HaveKeyWithValue(constants.InteractiveJobMaxIdleDurationSecondsAnnotation, "20"))
			Expect(project.Annotations).To(HaveKeyWithValue(constants.InteractivePreemptibleJobMaxIdleDurationSecondsAnnotation, "30"))
			Expect(project.Annotations).To(HaveKeyWithValue(constants.NodeAffinityInteractiveAnnotation, `["gpu-a"]`))
			Expect(project.Annotations).To(HaveKeyWithValue(constants.NodeAffinityTrainAnnotation, `["gpu-b"]`))
		})

		It("leaves the deletion type unset and produces only the updated-at annotation for a bare project", func() {
			project, err := convertOrgUnitToKaiProject(projectWithParent("p-1", "proj-1", "d-1", "dep-1"))
			Expect(err).ToNot(HaveOccurred())
			Expect(project.Spec.DeletionType).To(BeNil())
			Expect(project.Annotations).To(HaveKey(constants.UpdatedAtAnnotation))
			Expect(project.Annotations).To(HaveLen(1))
		})

		It("maps the deletion type onto the spec and force onto an annotation", func() {
			project, err := convertOrgUnitToKaiProject(
				withProjectDeletion(projectWithParent("p-1", "proj-1", "d-1", "dep-1"), true, orgUnitApi.DELETIONTYPES_CASCADE))
			Expect(err).ToNot(HaveOccurred())
			Expect(project.Spec.DeletionType).ToNot(BeNil())
			Expect(string(*project.Spec.DeletionType)).To(Equal(string(orgUnitApi.DELETIONTYPES_CASCADE)))
			Expect(project.Annotations).To(HaveKeyWithValue(constants.ForceDeleteAnnotation, "true"))
		})

		It("omits the department-id label and parent when the project has no parent", func() {
			project, err := convertOrgUnitToKaiProject(orgUnitApi.ProjectSync{Id: "p-1", Name: "proj-1"})
			Expect(err).ToNot(HaveOccurred())
			Expect(project.Spec.Parent).To(BeEmpty())
			Expect(project.Labels).ToNot(HaveKey(schedv2alpha2.DepartmentIdLabel))
		})
	})

	Describe("buildKaiProjectAnnotations", func() {
		It("returns only the updated-at annotation when there are no scheduling rules, node types or deletion", func() {
			annotations := buildKaiProjectAnnotations(projectWithParent("p-1", "proj-1", "d-1", "dep-1"))
			Expect(annotations).To(HaveKey(constants.UpdatedAtAnnotation))
			Expect(annotations).To(HaveLen(1))
		})
	})

	Describe("setSecondsAnnotation", func() {
		It("writes the value when set and skips it when unset", func() {
			annotations := map[string]string{}
			setSecondsAnnotation(annotations, "k", *orgUnitApi.NewNullableInt32(ptr.To(int32(7))))
			setSecondsAnnotation(annotations, "skipped", *orgUnitApi.NewNullableInt32(nil))
			Expect(annotations).To(HaveKeyWithValue("k", "7"))
			Expect(annotations).ToNot(HaveKey("skipped"))
		})
	})

	Describe("setStringSliceAnnotation", func() {
		It("JSON-encodes a non-empty slice and skips an empty one", func() {
			annotations := map[string]string{}
			setStringSliceAnnotation(annotations, "k", []string{"a", "b"})
			setStringSliceAnnotation(annotations, "skipped", nil)
			Expect(annotations).To(HaveKeyWithValue("k", `["a","b"]`))
			Expect(annotations).ToNot(HaveKey("skipped"))
		})
	})

	Describe("isKaiDepartmentPresent", func() {
		It("returns true when the department-id is in the provided set", func() {
			handler := newProjectsHandlerWithKaiDepartments()
			project := &kaiorg.Project{ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{schedv2alpha2.DepartmentIdLabel: "d-1"}}}
			Expect(handler.isKaiDepartmentPresent(project, map[string]bool{"d-1": true})).To(BeTrue())
		})

		It("falls back to matching the KAI department by name", func() {
			handler := newProjectsHandlerWithKaiDepartments(kaiDepartmentObj("dep-1", "different-id"))
			project := &kaiorg.Project{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{schedv2alpha2.DepartmentIdLabel: "stale"}},
				Spec:       kaiorg.ProjectSpec{Parent: "dep-1"},
			}
			Expect(handler.isKaiDepartmentPresent(project, map[string]bool{})).To(BeTrue())
		})

		It("returns false when neither the id nor the name match", func() {
			handler := newProjectsHandlerWithKaiDepartments(kaiDepartmentObj("dep-1", "d-1"))
			project := &kaiorg.Project{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{schedv2alpha2.DepartmentIdLabel: "missing"}},
				Spec:       kaiorg.ProjectSpec{Parent: "missing-dep"},
			}
			Expect(handler.isKaiDepartmentPresent(project, map[string]bool{})).To(BeFalse())
		})
	})

	Describe("getClusterObjects", func() {
		It("adds a KAI project whose parent department exists", func() {
			handler := newProjectsHandlerWithKaiDepartments(kaiDepartmentObj("dep-1", "d-1"))

			toAddUpdate, _, _, _, err := handler.getClusterObjects(
				[]orgUnitApi.ProjectSync{projectWithParent("p-1", "proj-1", "d-1", "dep-1")})

			Expect(err).ToNot(HaveOccurred())
			Expect(toAddUpdate).To(HaveLen(1))
			Expect(toAddUpdate[0]).To(BeAssignableToTypeOf(&kaiorg.Project{}))
			Expect(toAddUpdate[0].GetName()).To(Equal("proj-1"))
		})

		It("skips a KAI project with a missing parent department and returns an error", func() {
			handler := newProjectsHandlerWithKaiDepartments()

			toAddUpdate, _, _, _, err := handler.getClusterObjects(
				[]orgUnitApi.ProjectSync{projectWithParent("p-1", "proj-1", "d-1", "dep-1")})

			Expect(err).To(HaveOccurred())
			Expect(toAddUpdate).To(BeEmpty())
		})

		It("routes deleted KAI projects to the delete map", func() {
			handler := newProjectsHandlerWithKaiDepartments()

			_, toDelete, _, _, err := handler.getClusterObjects(
				[]orgUnitApi.ProjectSync{deletedProject("p-1", "proj-1", "d-1", "dep-1")})

			Expect(err).ToNot(HaveOccurred())
			Expect(toDelete).To(HaveKey("p-1"))
		})

		It("routes projects with a deletion request to the deletion-request list", func() {
			handler := newProjectsHandlerWithKaiDepartments(kaiDepartmentObj("dep-1", "d-1"))

			_, _, withDeletionRequest, _, err := handler.getClusterObjects([]orgUnitApi.ProjectSync{
				withProjectDeletion(projectWithParent("p-1", "proj-1", "d-1", "dep-1"), false, orgUnitApi.DELETIONTYPES_CASCADE)})

			Expect(err).ToNot(HaveOccurred())
			Expect(withDeletionRequest).To(HaveLen(1))
		})
	})

	Describe("getClusterObjects delegation", func() {
		It("produces KAI Projects when the toggle is on", func() {
			handler := newProjectsHandlerWithKaiDepartments(kaiDepartmentObj("dep-1", "d-1"))

			toAddUpdate, _, _, _, err := handler.getClusterObjects(
				[]orgUnitApi.ProjectSync{projectWithParent("p-1", "proj-1", "d-1", "dep-1")})

			Expect(err).ToNot(HaveOccurred())
			Expect(toAddUpdate).To(HaveLen(1))
			Expect(toAddUpdate[0]).To(BeAssignableToTypeOf(&kaiorg.Project{}))
		})
	})
})
