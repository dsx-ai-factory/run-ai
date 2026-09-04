/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"fmt"
	"strconv"

	. "github.com/onsi/gomega"

	kainodepool "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/resourcing"
	v2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	schedulingv2alpha2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// getBasicKaiProjectObject builds a minimal (intentionally backend-mismatching) kai Project used
// to seed the "existing" state the reconcile specs start from.
func getBasicKaiProjectObject(projectName, projectId, departmentName, departmentId, statusNamespace string) *kainodepool.Project {
	project := &kainodepool.Project{
		TypeMeta: metav1.TypeMeta{Kind: "Project", APIVersion: kainodepool.GroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name: projectName,
			Labels: map[string]string{
				v2.ProjectIdLabel:                    projectId,
				schedulingv2alpha2.DepartmentIdLabel: departmentId,
			},
		},
		Spec: kainodepool.ProjectSpec{
			Parent: departmentName,
			Queues: []kainodepool.QueueConfig{{Name: projectName, Priority: ptr.To(int32(210))}},
		},
	}
	if statusNamespace != "" {
		project.Status.Namespace = statusNamespace
	}
	return project
}

func getBasicKaiDepartmentObject(departmentName, departmentId string) *kainodepool.Department {
	return &kainodepool.Department{
		TypeMeta: metav1.TypeMeta{Kind: "Department", APIVersion: kainodepool.GroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:   departmentName,
			Labels: map[string]string{schedulingv2alpha2.DepartmentIdLabel: departmentId},
		},
		Spec: kainodepool.DepartmentSpec{
			Queues: []kainodepool.QueueConfig{{Name: departmentName, Nodepool: "default", Priority: ptr.To(int32(50))}},
		},
	}
}

func validateKaiProjectExistsEventually(projectName string, k8sClient client.Client) {
	EventuallyWithOffset(1, func() *kainodepool.Project {
		return getProjectFromClient(projectName, k8sClient)
	}, DefaultTimeout).ShouldNot(BeNil())
}

func validateKaiProjectDeletedEventually(projectName string, k8sClient client.Client) {
	EventuallyWithOffset(1, func() *kainodepool.Project {
		return getProjectFromClient(projectName, k8sClient)
	}, DefaultTimeout).Should(BeNil())
}

func validateExistingKaiProjectMatchesBackendSpec(projectName string, k8sClient client.Client) {
	EventuallyWithOffset(1, func(g Gomega) {
		validateExistingKaiProjectWithBackend(g, projectName, k8sClient)
	}, 0).Should(Succeed())
}

func validateExistingKaiProjectDoesNotMatchBackendSpec(projectName string, k8sClient client.Client) {
	EventuallyWithOffset(1, func(g Gomega) {
		validateExistingKaiProjectWithBackend(g, projectName, k8sClient)
	}, 0).ShouldNot(Succeed())
}

func validateExistingKaiProjectWithBackend(g Gomega, projectName string, k8sClient client.Client) {
	clusterProject := getProjectFromClient(projectName, k8sClient)
	g.ExpectWithOffset(1, clusterProject).ToNot(BeNil())

	backendProjSpec := OrgUnitProjectsMap[projectName]
	g.ExpectWithOffset(1, backendProjSpec).ToNot(BeNil(), "Backend project spec should exist")

	g.ExpectWithOffset(1, clusterProject.Name).To(Equal(backendProjSpec.Name), "project name")
	g.ExpectWithOffset(1, clusterProject.Spec.Parent).To(Equal(backendProjSpec.Parent.Get().Name), "parent department name")
	g.ExpectWithOffset(1, clusterProject.Labels[v2.ProjectIdLabel]).To(Equal(backendProjSpec.Id), "project id")
	g.ExpectWithOffset(1, clusterProject.Labels[schedulingv2alpha2.DepartmentIdLabel]).To(Equal(backendProjSpec.Parent.Get().Id), "department id")
	g.ExpectWithOffset(1, clusterProject.Spec.DefaultNodePools).To(Equal(backendProjSpec.DefaultNodePools), "DefaultNodePools")
	g.ExpectWithOffset(1, clusterProject.Spec.Namespace).To(Equal(common.GetStringVal(backendProjSpec.RequestedNamespace.Get())), "Namespace")

	g.ExpectWithOffset(1, len(clusterProject.Spec.Queues)).To(Equal(len(backendProjSpec.Resources)), "Queues length")
	for _, queue := range clusterProject.Spec.Queues {
		validateQueue(g, queue.Name, queue.Nodepool, queue.Priority, kaiQueueResourcesValue(queue.Resources), backendProjSpec.Resources, backendProjSpec.Name)
	}
}

func validateKaiDepartmentExistsEventually(departmentName string, k8sClient client.Client) {
	EventuallyWithOffset(1, func() *kainodepool.Department {
		return getDepartmentFromClient(departmentName, k8sClient)
	}, DefaultTimeout).ShouldNot(BeNil())
}

func validateKaiDepartmentDeletedEventually(departmentName string, k8sClient client.Client) {
	EventuallyWithOffset(1, func() *kainodepool.Department {
		return getDepartmentFromClient(departmentName, k8sClient)
	}, DefaultTimeout).Should(BeNil())
}

func validateExistingKaiDepartmentMatchesBackendSpec(departmentName string, k8sClient client.Client) {
	EventuallyWithOffset(1, func(g Gomega) {
		validateExistingKaiDepartmentWithBackend(g, departmentName, k8sClient)
	}, 0).Should(Succeed())
}

func validateExistingKaiDepartmentDoesNotMatchBackendSpec(departmentName string, k8sClient client.Client) {
	EventuallyWithOffset(1, func(g Gomega) {
		validateExistingKaiDepartmentWithBackend(g, departmentName, k8sClient)
	}, 0).ShouldNot(Succeed())
}

func validateExistingKaiDepartmentWithBackend(g Gomega, departmentName string, k8sClient client.Client) {
	clusterDepartment := getDepartmentFromClient(departmentName, k8sClient)
	g.ExpectWithOffset(1, clusterDepartment).ToNot(BeNil())

	backendDepSpec := OrgUnitDepartmentsMap[departmentName]
	g.ExpectWithOffset(1, backendDepSpec).ToNot(BeNil(), "Backend department spec should exist")

	g.ExpectWithOffset(1, clusterDepartment.Name).To(Equal(common.GetStringVal(backendDepSpec.Name)), "department name")
	g.ExpectWithOffset(1, clusterDepartment.Labels[schedulingv2alpha2.DepartmentIdLabel]).To(Equal(backendDepSpec.Id), "department id")

	// kai department queues are named "q-<id>"; the legacy handler used the department name.
	queueBaseName := fmt.Sprintf("q-%s", backendDepSpec.Id)
	g.ExpectWithOffset(1, len(clusterDepartment.Spec.Queues)).To(Equal(len(backendDepSpec.Resources)), "Queues length")
	for _, queue := range clusterDepartment.Spec.Queues {
		validateQueue(g, queue.Name, queue.Nodepool, queue.Priority, kaiQueueResourcesValue(queue.Resources), backendDepSpec.Resources, queueBaseName)
	}
}

func kaiQueueResourcesValue(resources *kainodepool.QueueResourcesConfig) resourcing.QueueResources {
	if resources == nil {
		return resourcing.QueueResources{}
	}
	toSystemResource := func(r kainodepool.SystemResource) resourcing.SystemResource {
		return resourcing.SystemResource{Deserved: r.Deserved, OverQuotaWeight: r.OverQuotaWeight, Limit: r.Limit}
	}
	return resourcing.QueueResources{
		GPU:    toSystemResource(resources.GPU),
		CPU:    toSystemResource(resources.CPU),
		Memory: toSystemResource(resources.Memory),
	}
}

// seedKaiOrgUnits creates the kai "existing" org units the reconcile specs start from. The shared
// JustBeforeEach seeds the legacy copies (needed by the agent's legacy ProjectToNamespaceList),
// so the kai copies are created here instead.
func seedKaiOrgUnits(k8sClient client.Client) {
	kaiObjects := []client.Object{
		getBasicKaiProjectObject(Project1Name, strconv.Itoa(Proj1Id), Department1Name, strconv.Itoa(Dep1Id), NamespaceForSpec1),
		getBasicKaiProjectObject(Project4Name, strconv.Itoa(4), Department1Name, strconv.Itoa(Dep1Id), ""),
		getBasicKaiProjectObject(Project3Name, "22", Department1Name, strconv.Itoa(Dep1Id), ""),
		getBasicKaiDepartmentObject(Department1Name, strconv.Itoa(Dep1Id)),
		getBasicKaiDepartmentObject(Department4Name, strconv.Itoa(Dep4Id)),
	}
	for _, obj := range kaiObjects {
		ExpectWithOffset(1, k8sClient.Create(context.Background(), obj)).To(Succeed())
	}
}
