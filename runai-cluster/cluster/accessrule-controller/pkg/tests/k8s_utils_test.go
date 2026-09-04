/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	utils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/tests/utils"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	runaiv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	schedulingv2alpha2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getNamespaceFromClient(namespaceName string, k8sClient client.Client) *corev1.Namespace {
	namespaceKey := types.NamespacedName{Name: namespaceName}
	namespaceObj := &corev1.Namespace{}
	err := k8sClient.Get(apiCtx, namespaceKey, namespaceObj)
	if err != nil {
		return nil
	}
	return namespaceObj
}

func getProjectFromClient(projectName string, k8sClient client.Client) *runaiv2.Project {
	projectKey := types.NamespacedName{Name: projectName}
	projectObj := &runaiv2.Project{}
	err := k8sClient.Get(apiCtx, projectKey, projectObj)
	if err != nil {
		return nil
	}
	return projectObj
}

func getDepartmentFromClient(departmentName string, k8sClient client.Client) *schedulingv2alpha2.Department {
	departmentKey := types.NamespacedName{Name: departmentName}
	departmentObj := &schedulingv2alpha2.Department{}
	err := k8sClient.Get(apiCtx, departmentKey, departmentObj)
	if err != nil {
		return nil
	}
	return departmentObj
}

// getKaiProjectFromClient / getKaiDepartmentFromClient are the kai.resources twins, used so the
// fixtures manage both org-unit worlds.
func getKaiProjectFromClient(projectName string, k8sClient client.Client) *kaiorg.Project {
	projectObj := &kaiorg.Project{}
	err := k8sClient.Get(apiCtx, types.NamespacedName{Name: projectName}, projectObj)
	if err != nil {
		return nil
	}
	return projectObj
}

func getKaiDepartmentFromClient(departmentName string, k8sClient client.Client) *kaiorg.Department {
	departmentObj := &kaiorg.Department{}
	err := k8sClient.Get(apiCtx, types.NamespacedName{Name: departmentName}, departmentObj)
	if err != nil {
		return nil
	}
	return departmentObj
}

func getClusterRoleFromClient(name string, k8sClient client.Client) *rbacv1.ClusterRole {
	crKey := types.NamespacedName{Name: name}
	crObj := &rbacv1.ClusterRole{}
	err := k8sClient.Get(apiCtx, crKey, crObj)
	if err != nil {
		return nil
	}
	return crObj
}

func getClusterRoleBindingFromClient(name string, k8sClient client.Client) *rbacv1.ClusterRoleBinding {
	crKey := types.NamespacedName{Name: name}
	crObj := &rbacv1.ClusterRoleBinding{}
	err := k8sClient.Get(apiCtx, crKey, crObj)
	if err != nil {
		return nil
	}
	return crObj
}

func getAccessRuleFromClient(accessRuleName string, k8sClient client.Client) *v1alpha1.AccessRule {
	accessRuleKey := types.NamespacedName{Name: accessRuleName}
	accessRuleObj := &v1alpha1.AccessRule{}
	err := k8sClient.Get(apiCtx, accessRuleKey, accessRuleObj)
	if err != nil {
		return nil
	}
	return accessRuleObj
}

func getRoleBindingFromClient(name, namespace string, k8sClient client.Client) *rbacv1.RoleBinding {
	rbKey := types.NamespacedName{Name: name, Namespace: namespace}
	rbObj := &rbacv1.RoleBinding{}
	err := k8sClient.Get(apiCtx, rbKey, rbObj)
	if err != nil {
		return nil
	}
	return rbObj
}

func listAccessRules(k8sClient client.Client) *v1alpha1.AccessRuleList {
	accessRules := &v1alpha1.AccessRuleList{}
	err := k8sClient.List(apiCtx, accessRules)
	if err != nil {
		return nil
	}
	return accessRules
}

func listRoleBindings(k8sClient client.Client) *rbacv1.RoleBindingList {
	roleBindings := &rbacv1.RoleBindingList{}
	err := k8sClient.List(apiCtx, roleBindings)
	if err != nil {
		return nil
	}
	return roleBindings
}

func createNamespaceIfNeeded(namespace, project string, k8sClient client.Client) {
	namespaceObj := getNamespaceFromClient(namespace, k8sClient)
	if namespaceObj != nil {
		return
	}

	namespaceObj = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			UID:  types.UID(namespace),
			Annotations: map[string]string{
				"suite": suite,
			},
		},
	}
	if namespace != runaiv1.DefaultNamespace && project != "" {
		namespaceObj.Labels = map[string]string{
			"runai/queue": project,
		}
	}
	utils.ExpectCreateResource(k8sClient, namespaceObj)
}

func createProject(projectName, departmentName, namespaceName string, k8sClient client.Client) {
	if departmentName == "" {
		departmentName = "default"
	}

	projectObj := &runaiv2.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: projectName,
			UID:  types.UID(projectName),
		},
		Spec: runaiv2.ProjectSpec{
			Department: departmentName,
			Queues: []runaiv2.QueueProjectSpec{
				{
					Name: projectName,
				},
			},
		},
	}

	utils.ExpectCreateResource(k8sClient, projectObj)

	// kai.resources twin, so the reconciler resolves the same project.
	// Spec.Parent replaces the legacy Spec.Department.
	kaiProjectObj := &kaiorg.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: projectName,
			UID:  types.UID(projectName),
		},
		Spec: kaiorg.ProjectSpec{
			Parent: departmentName,
			Queues: []kaiorg.QueueConfig{{Name: projectName}},
		},
	}

	utils.ExpectCreateResource(k8sClient, kaiProjectObj)

	if namespaceName != "" {
		eventuallyUpdateProjectStatusNamespace(k8sClient, projectName, namespaceName)
	}
}

func createDepartment(departmentName string, k8sClient client.Client) {
	departmentObj := &schedulingv2alpha2.Department{
		ObjectMeta: metav1.ObjectMeta{
			Name: departmentName,
			UID:  types.UID(departmentName),
		},
		Spec: schedulingv2alpha2.DepartmentSpec{
			Queues: []schedulingv2alpha2.QueueDepartmentSpec{{
				Name:     departmentName,
				Nodepool: "default",
			}},
		},
	}

	utils.ExpectCreateResource(k8sClient, departmentObj)

	// kai.resources twin, so department-scoped rules validate the department.
	kaiDepartmentObj := &kaiorg.Department{
		ObjectMeta: metav1.ObjectMeta{
			Name: departmentName,
			UID:  types.UID(departmentName),
		},
		Spec: kaiorg.DepartmentSpec{
			Queues: []kaiorg.QueueConfig{{Name: departmentName, Nodepool: "default"}},
		},
	}

	utils.ExpectCreateResource(k8sClient, kaiDepartmentObj)
}

func createClusterRole(clusterRoleName string, k8sClient client.Client) {
	clusterRoleObj := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleName,
			UID:  types.UID(clusterRoleName),
		},
	}

	utils.ExpectCreateResource(k8sClient, clusterRoleObj)
}

func createClusterRoleBinding(clusterRoleBindingName string, k8sClient client.Client) {
	clusterRoleObj := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleBindingName,
			UID:  types.UID(clusterRoleBindingName),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     clusterRoleBindingName,
		},
	}

	utils.ExpectCreateResource(k8sClient, clusterRoleObj)
}

func createRoleBinding(roleBindingName, namespace string, k8sClient client.Client) {
	clusterRoleObj := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleBindingName,
			Namespace: namespace,
			UID:       types.UID(roleBindingName),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     roleBindingName,
		},
	}

	utils.ExpectCreateResource(k8sClient, clusterRoleObj)
}

func createAccessRule(accessRuleName string, accessRuleSpec *v1alpha1.AccessRuleSpec, k8sClient client.Client) *v1alpha1.AccessRule {
	accessRuleObj := &v1alpha1.AccessRule{
		ObjectMeta: metav1.ObjectMeta{
			Name: accessRuleName,
			UID:  types.UID(accessRuleName),
		},
		Spec: *accessRuleSpec,
	}

	utils.ExpectCreateResource(k8sClient, accessRuleObj)
	return accessRuleObj
}

func deleteNamespace(namespace string, k8sClient client.Client) {
	By(fmt.Sprintf("Deleting Namespace <%s>", namespace))

	namespaceObj := getNamespaceFromClient(namespace, k8sClient)
	if namespaceObj != nil {
		utils.DeleteAndPollUntilDeleted(k8sClient, namespaceObj)
	}
}

func deleteProject(project string, k8sClient client.Client) {
	By(fmt.Sprintf("Deleting project <%s>", project))

	projectObj := getProjectFromClient(project, k8sClient)
	if projectObj != nil {
		utils.DeleteAndPollUntilDeleted(k8sClient, projectObj)
	}
	kaiProjectObj := getKaiProjectFromClient(project, k8sClient)
	if kaiProjectObj != nil {
		utils.DeleteAndPollUntilDeleted(k8sClient, kaiProjectObj)
	}
}

func deleteDepartment(department string, k8sClient client.Client) {
	By(fmt.Sprintf("Deleting department <%s>", department))

	departmentObj := getDepartmentFromClient(department, k8sClient)
	if departmentObj != nil {
		utils.DeleteAndPollUntilDeleted(k8sClient, departmentObj)
	}
	kaiDepartmentObj := getKaiDepartmentFromClient(department, k8sClient)
	if kaiDepartmentObj != nil {
		utils.DeleteAndPollUntilDeleted(k8sClient, kaiDepartmentObj)
	}
}

func deleteClusterRole(clusterRole string, k8sClient client.Client) {
	By(fmt.Sprintf("Deleting clusterRole <%s>", clusterRole))

	clusterRoleObj := getClusterRoleFromClient(clusterRole, k8sClient)
	if clusterRoleObj != nil {
		utils.DeleteAndPollUntilDeleted(k8sClient, clusterRoleObj)
	}
}

func deleteClusterRoleBinding(clusterRoleBinding string, k8sClient client.Client) {
	By(fmt.Sprintf("Deleting clusterRoleBinding <%s>", clusterRoleBinding))

	clusterRoleBindingObj := getClusterRoleBindingFromClient(clusterRoleBinding, k8sClient)
	if clusterRoleBindingObj != nil {
		utils.DeleteAndPollUntilDeleted(k8sClient, clusterRoleBindingObj)
	}
}
func deleteRoleBinding(roleBinding, namespace string, k8sClient client.Client) {
	By(fmt.Sprintf("Deleting roleBinding <%s>", roleBinding))

	roleBindingObj := getRoleBindingFromClient(roleBinding, namespace, k8sClient)
	if roleBindingObj != nil {
		utils.DeleteAndPollUntilDeleted(k8sClient, roleBindingObj)
	}
}

func deleteAccessRule(accessRule string, k8sClient client.Client) {
	By(fmt.Sprintf("Deleting accessRule <%s>", accessRule))

	accessRuleObj := getAccessRuleFromClient(accessRule, k8sClient)
	if accessRuleObj == nil {
		return
	}

	// The controller adds a finalizer, so the fake client keeps the object until
	// the controller's finalize runs. Issue the delete, then reconcile so the
	// finalizer is removed (deleting role bindings along the way) and the object
	// actually disappears.
	ExpectWithOffset(1, k8sClient.Delete(apiCtx, accessRuleObj)).To(Succeed())

	EventuallyWithOffset(1, func() *v1alpha1.AccessRule {
		reconcileAllAccessRules()
		return getAccessRuleFromClient(accessRule, k8sClient)
	}, timeout, interval).Should(BeNil(), "expected access rule <%s> to be deleted after finalizer removal", accessRule)
}

func eventuallyUpdateProjectStatusNamespace(k8sClient client.Client, projectName, namespaceToUpdate string) {
	By(fmt.Sprintf("updating status of project %s with namespace %s...", projectName, namespaceToUpdate))
	Eventually(func(g Gomega) {
		project := getProjectFromClient(projectName, k8sClient)
		g.Expect(project).ToNot(BeNil())
		project.Status.Namespace = namespaceToUpdate

		err := k8sClient.Status().Update(apiCtx, project)
		g.Expect(err).ToNot(HaveOccurred())
	}, timeout*5, interval).Should(Succeed())

	// mirror onto the kai.resources twin so the reconciler resolves the namespace under
	// kai.resources twin as well.
	Eventually(func(g Gomega) {
		kaiProject := getKaiProjectFromClient(projectName, k8sClient)
		g.Expect(kaiProject).ToNot(BeNil())
		kaiProject.Status.Namespace = namespaceToUpdate

		err := k8sClient.Status().Update(apiCtx, kaiProject)
		g.Expect(err).ToNot(HaveOccurred())
	}, timeout*5, interval).Should(Succeed())
}

func expectUpdateResourceStatus(k8sClient client.Client, resource client.Object) {
	By(fmt.Sprintf("updating status of resource %s/%s...", resource.GetNamespace(), resource.GetName()))

	ExpectWithOffset(1, k8sClient.Status().Update(apiCtx, resource)).To(Succeed())

	By(fmt.Sprintf("succesfully updated status of resource %s/%s", resource.GetNamespace(), resource.GetName()))
}

func eventuallyUpdateRoleBindingSubjects(k8sClient client.Client, name, namespace string, subjects []rbacv1.Subject) {
	Eventually(func(g Gomega) {
		roleBinding := getRoleBindingFromClient(name, namespace, k8sClient)
		g.Expect(roleBinding).ToNot(BeNil())
		roleBinding.Subjects = subjects

		err := k8sClient.Update(apiCtx, roleBinding)
		g.Expect(err).ToNot(HaveOccurred())
	}, timeout*5, interval).Should(Succeed())
}

func eventuallyUpdateRoleBindingLabels(k8sClient client.Client, name, namespace string, labels map[string]string) {
	Eventually(func(g Gomega) {
		roleBinding := getRoleBindingFromClient(name, namespace, k8sClient)
		g.Expect(roleBinding).ToNot(BeNil())
		roleBinding.Labels = labels

		err := k8sClient.Update(apiCtx, roleBinding)
		g.Expect(err).ToNot(HaveOccurred())
	}, timeout*5, interval).Should(Succeed())
}
