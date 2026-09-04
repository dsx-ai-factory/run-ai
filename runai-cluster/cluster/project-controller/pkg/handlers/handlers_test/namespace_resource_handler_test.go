/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers_test

import (
	"context"
	"fmt"
	"strconv"

	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/config"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/handlers"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/test"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const vIsOpenshift = true

var _ = Describe("Namespace Resource Handler", func() {

	var (
		handler                   NamespaceResourceHandler
		client                    k8sclient.Client
		project, modifiedProject  kaiv1alpha1.Project
		anotherProject            kaiv1alpha1.Project
		namespace                 v1.Namespace
		defaultRunaiProjNamespace string
	)

	BeforeEach(func() {
		project = *TestProject.DeepCopy()
		modifiedProject = *KaiModifiedTestProject.DeepCopy()
		anotherProject = *KaiAnotherTestProject.DeepCopy()
		namespace = *TestNamespace.DeepCopy()
		defaultRunaiProjNamespace = DefaultProjectNamespaceName(&project)

		client = fake.NewClientBuilder().WithScheme(scheme).WithObjects(TestNamespace.DeepCopy()).Build()
		handler = NewNamespaceResourceHandler(client, false, true)
	})

	It("Handles a new, non-existing Namespace", func() {
		By("Creating a new Namespace matching the Project spec", func() {

			// Given
			beforeCreation, beforeCreationErr := handler.GetNamespace(defaultRunaiProjNamespace)
			Expect(beforeCreationErr).To(HaveOccurred())
			Expect(beforeCreation).To(BeZero())

			// When
			conditions, err := handler.HandleResource(project)
			Expect(err).Should(Succeed())

			// Then
			Expect(conditions).To(HaveLen(1))
			Expect(conditions[0]).ToNot(BeNil())
			Expect(conditions[0].Type).To(Equal(kaiv1alpha1.NamespaceReady))
			Expect(conditions[0].Status).To(Equal(v1.ConditionTrue))
			Expect(conditions[0].Reason).To(Equal(""))
			Expect(conditions[0].Message).To(Equal(""))

			assertNamespaceIdenticalToProjectSpec(handler, project, defaultRunaiProjNamespace)
		})
	})

	It("Handles an existing Namespace", func() {
		By("Not modifying the Namespace when the incoming project Namespace spec is identical", func() {

			// Given
			Expect(client.Create(context.TODO(), &project)).To(Succeed())

			// When
			_, err := handler.HandleResource(project)
			Expect(err).Should(Succeed())

			// Then
			expected := GetNamespaceObject(project, TestNamespace.Name, !vIsOpenshift)
			actual, err := handler.GetNamespace(TestNamespace.Name)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).ToNot(BeZero())
			// The fake client strips TypeMeta on read; restore it to mirror a real client.
			populateGVK(&actual)
			Expect(NamespacesEqual(actual, *expected)).To(BeTrue())
			Expect(NamespacesEqual(actual, namespace)).To(BeTrue())
		})
	})

	It("Reports namespaces differing only in the enforce-scheduler annotation as unequal", func() {
		// Guards the annotation clause of NamespacesEqual directly: via the handler it is
		// masked, because TypeMeta is unset on a client read and an earlier conjunct already
		// returns false. Both sides here are byte-identical apart from the annotation.
		enforced := GetNamespaceObject(project, defaultRunaiProjNamespace, !vIsOpenshift)
		enforced.Annotations[config.Get().EnforceSchedulerAnnotationKey] = strconv.FormatBool(true)

		notEnforced := enforced.DeepCopy()
		notEnforced.Annotations[config.Get().EnforceSchedulerAnnotationKey] = strconv.FormatBool(false)

		Expect(NamespacesEqual(*enforced, *notEnforced)).To(BeFalse())
		Expect(NamespacesEqual(*enforced, *enforced.DeepCopy())).To(BeTrue())
	})

	It("Propagates a flipped EnforceKaiScheduler onto the existing Namespace", func() {
		// Given a namespace already reconciled from a project with enforcement off, and
		// carrying an unrelated annotation as a live namespace would.
		project.Spec.EnforceKaiScheduler = false
		Expect(client.Create(context.TODO(), &project)).To(Succeed())
		_, err := handler.HandleResource(project)
		Expect(err).Should(Succeed())

		created, err := handler.GetNamespace(defaultRunaiProjNamespace)
		Expect(err).ToNot(HaveOccurred())
		created.Annotations["unrelated.example.com/owner"] = "someone-else"
		Expect(client.Update(context.TODO(), &created)).To(Succeed())

		// When enforcement is turned on for the same project.
		project.Spec.EnforceKaiScheduler = true
		_, err = handler.HandleResource(project)
		Expect(err).Should(Succeed())

		// Then the namespace annotation follows the spec, and foreign annotations survive.
		actual, err := handler.GetNamespace(defaultRunaiProjNamespace)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual.Annotations).To(HaveKeyWithValue(
			config.Get().EnforceSchedulerAnnotationKey, strconv.FormatBool(true)))
		Expect(actual.Annotations).To(HaveKeyWithValue("unrelated.example.com/owner", "someone-else"))
	})

	It("Handles an existing Namespace", func() {
		By("Modifying the Namespace when the incoming project Namespace spec has differences", func() {

			// Given
			Expect(client.Create(context.TODO(), &project)).To(Succeed())
			assertNamespaceIdenticalToProjectSpec(handler, project, TestNamespace.Name)

			// When
			_, err := handler.HandleResource(modifiedProject)
			Expect(err).Should(Succeed())

			// Then
			expected := GetNamespaceObject(modifiedProject, defaultRunaiProjNamespace, !vIsOpenshift)
			actual, err := handler.GetNamespace(defaultRunaiProjNamespace)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).ToNot(BeZero())
			// The fake client strips TypeMeta on read; restore it to mirror a real client.
			populateGVK(&actual)
			Expect(NamespacesEqual(actual, *expected)).To(BeTrue())
			assertNamespaceIdenticalToProjectSpec(handler, modifiedProject, defaultRunaiProjNamespace)
		})
	})

	Specify("Owner Reference Creation", func() {

		// When
		_, _ = handler.HandleResource(project)

		// Then
		currentNamespace, err := handler.GetNamespace(defaultRunaiProjNamespace)
		Expect(err).ToNot(HaveOccurred())
		Expect(currentNamespace.Name).To(Equal(defaultRunaiProjNamespace))

		Expect(currentNamespace.OwnerReferences).To(BeEmpty())
	})

	When("openshift flag is not set", func() {
		It("doesn't label namespaces with openshift's cluster-monitoring", func() {
			// When
			_, _ = handler.HandleResource(project)

			// Then
			currentNamespace, err := handler.GetNamespace(defaultRunaiProjNamespace)
			Expect(err).ToNot(HaveOccurred())
			Expect(currentNamespace.Labels).ToNot(BeNil())

			_, labelExists := currentNamespace.Labels["openshift.io/cluster-monitoring"]
			Expect(labelExists).To(BeFalse())
		})
	})

	When("Openshift flag is set", func() {
		BeforeEach(func() {
			handler = NewNamespaceResourceHandler(client, true, true)
		})

		It("Label namespaces with openshift's cluster-monitoring", func() {
			// When
			_, _ = handler.HandleResource(project)

			// Then
			currentNamespace, err := handler.GetNamespace(defaultRunaiProjNamespace)
			Expect(err).ToNot(HaveOccurred())
			Expect(currentNamespace.Labels).ToNot(BeNil())
			Expect(currentNamespace.Labels["openshift.io/cluster-monitoring"]).To(Equal("true"))
		})
	})

	When("CreateNamespaces flag is off", func() {
		BeforeEach(func() {
			handler = NewNamespaceResourceHandler(client, false, false)
		})

		It("Failing cause namespace does not exist", func() {
			Expect(client.Create(context.TODO(), &anotherProject)).To(Succeed())

			conditions, err := handler.HandleResource(anotherProject)
			Expect(err).ShouldNot(Succeed())

			Expect(conditions).To(HaveLen(1))
			Expect(conditions[0]).ToNot(BeNil())
			Expect(conditions[0].Type).To(Equal(kaiv1alpha1.NamespaceReady))
			Expect(conditions[0].Status).To(Equal(v1.ConditionFalse))
			Expect(conditions[0].Reason).To(Equal(string(NamespaceNotFound)))
			Expect(conditions[0].Message).To(Equal(ProjectsNamespaceNotFoundStatusMsg))
		})

		It("Succeeding cause namespace exists", func() {
			Expect(client.Create(context.TODO(), &project)).To(Succeed())

			_, err := handler.HandleResource(project)
			Expect(err).Should(Succeed())
		})
	})

	When("Using Existing Namespace for Project", func() {
		var (
			testHandler           NamespaceResourceHandler
			testClient            k8sclient.Client
			userExternalNamespace = &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "proj-1",
				},
			}
		)

		BeforeEach(func() {
			testClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(userExternalNamespace).Build()
			testHandler = NewNamespaceResourceHandler(testClient, false, true)
		})

		AfterEach(func() {
			Expect(testClient.Delete(context.TODO(), &project)).To(Succeed())
		})

		It("Not creating namespace if it does not exist", func() {
			project.Spec.Namespace = "random-unexisting-namespace"
			Expect(testClient.Create(context.TODO(), &project)).To(Succeed())

			conditions, err := testHandler.HandleResource(project)
			Expect(err).To(HaveOccurred())

			Expect(len(conditions)).To(Equal(1))
			Expect(conditions[0].Reason).To(Equal(string(NamespaceNotFound)))
			Expect(conditions[0].Message).To(Equal(ProjectsNamespaceNotFoundStatusMsg))

			actual, err := testHandler.GetNamespace(userExternalNamespace.Name)
			Expect(err).ToNot(HaveOccurred())

			Expect(actual.Labels[config.Get().NamespaceProjectLabelKey]).To(Equal(""))
		})

		It("Adding label to existing namespace", func() {
			project.Spec.Namespace = userExternalNamespace.Name
			Expect(testClient.Create(context.TODO(), &project)).To(Succeed())

			_, err := testHandler.HandleResource(project)
			Expect(err).ToNot(HaveOccurred())

			actual, err := testHandler.GetNamespace(userExternalNamespace.Name)
			Expect(err).ToNot(HaveOccurred())

			Expect(actual.Labels[config.Get().NamespaceProjectLabelKey]).To(Equal(project.Name))
		})

		It("Existing namespace already has correct label - do nothing", func() {
			actual, err := testHandler.GetNamespace(userExternalNamespace.Name)
			Expect(err).ToNot(HaveOccurred())
			actual.Labels = map[string]string{config.Get().NamespaceProjectLabelKey: project.Name}
			Expect(testClient.Update(context.TODO(), &actual)).To(Succeed())

			project.Spec.Namespace = userExternalNamespace.Name
			Expect(testClient.Create(context.TODO(), &project)).To(Succeed())

			conditions, err := testHandler.HandleResource(project)
			Expect(err).ToNot(HaveOccurred())

			Expect(len(conditions)).To(Equal(1))
			Expect(conditions[0].Message).To(Equal(""))

			actual, err = testHandler.GetNamespace(userExternalNamespace.Name)
			Expect(err).ToNot(HaveOccurred())

			Expect(actual.Labels[config.Get().NamespaceProjectLabelKey]).To(Equal(project.Name))
		})

		It("Error when namespace has other project in label value", func() {
			actual, err := testHandler.GetNamespace(userExternalNamespace.Name)
			Expect(err).ToNot(HaveOccurred())
			actual.Labels = map[string]string{config.Get().NamespaceProjectLabelKey: anotherProject.Name}
			Expect(testClient.Update(context.TODO(), &actual)).To(Succeed())

			project.Spec.Namespace = userExternalNamespace.Name
			Expect(testClient.Create(context.TODO(), &project)).To(Succeed())
			Expect(testClient.Create(context.TODO(), &anotherProject)).To(Succeed())

			conditions, err := testHandler.HandleResource(project)
			Expect(err).To(HaveOccurred())

			Expect(len(conditions)).To(Equal(1))
			Expect(conditions[0].Message).To(Equal(fmt.Sprintf("'%s' %s", RunaiQueueLabel, QueueLabelMissingFromNamespaceStatusMsg)))

			actual, err = testHandler.GetNamespace(userExternalNamespace.Name)
			Expect(err).ToNot(HaveOccurred())

			Expect(actual.Labels[config.Get().NamespaceProjectLabelKey]).To(Equal(anotherProject.Name))
		})

		It("Error - not updating label when namespace has un-existing project in label value", func() {
			actual, err := testHandler.GetNamespace(userExternalNamespace.Name)
			Expect(err).ToNot(HaveOccurred())
			actual.Labels = map[string]string{config.Get().NamespaceProjectLabelKey: "some-other-proj"}
			Expect(testClient.Update(context.TODO(), &actual)).To(Succeed())

			project.Spec.Namespace = userExternalNamespace.Name
			Expect(testClient.Create(context.TODO(), &project)).To(Succeed())

			conditions, err := testHandler.HandleResource(project)
			Expect(err).To(HaveOccurred())

			Expect(len(conditions)).To(Equal(1))
			Expect(conditions[0].Message).To(Equal(fmt.Sprintf("'%s' %s", RunaiQueueLabel, QueueLabelMissingFromNamespaceStatusMsg)))

			actual, err = testHandler.GetNamespace(userExternalNamespace.Name)
			Expect(err).ToNot(HaveOccurred())

			Expect(actual.Labels[config.Get().NamespaceProjectLabelKey]).To(Equal("some-other-proj"))
		})
	})
})

func assertNamespaceIdenticalToProjectSpec(
	handler NamespaceResourceHandler, project kaiv1alpha1.Project, namespaceName string,
) {
	namespace, err := handler.GetNamespace(namespaceName)
	Expect(err).ToNot(HaveOccurred())
	Expect(namespace).ToNot(BeZero())
	// The fake client strips TypeMeta on read; restore it to mirror a real client.
	populateGVK(&namespace)

	Expect(namespace.Name).To(Equal(namespaceName))
	Expect(namespace.APIVersion).To(Equal(CoreV1ApiVersion))
	Expect(namespace.Kind).To(Equal(NamespaceKind))
	Expect(namespace.Labels[config.Get().NamespaceProjectLabelKey]).To(Equal(project.Name))
	Expect(namespace.Labels[config.Get().NamespaceVersionLabelKey]).To(Equal(AgentNamespaceVersion))
	Expect(namespace.ObjectMeta.Annotations[EnforceSchedulerAnnotationName]).To(Equal(strconv.FormatBool(project.Spec.EnforceKaiScheduler)))
}
