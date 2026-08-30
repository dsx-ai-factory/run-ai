package project_controller

import (
	"context"
	"fmt"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/integration-tests/common"
	"github.com/run-ai/runai/runai-cluster/cluster/integration-tests/config"
	schedulingv2alpha2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Project Controller Test", Ordered, func() {
	var (
		testConfig         *config.TestConfig
		kubeClient         client.Client
		ctx                context.Context
		testCreatedObjects = map[string]client.Object{}
		project            *kaiv1alpha1.Project
		namespaceName      string
	)

	projName := common.GetRandomName("project")
	projId := common.GetRandomName("projectId")

	BeforeAll(func() {
		testConfig = config.GetTestConfig(Default)
		kubeClient = testConfig.KubeClient
		ctx = testConfig.Ctx
	})

	AfterAll(func() {
		for _, obj := range testCreatedObjects {
			if err := kubeClient.Delete(ctx, obj); err != nil {
				fmt.Printf("Failed to delete resource %v: %v\n", obj, err)
			}
		}
		testConfig.Cleanup()
	})

	Context("Blocking deletion with no force delete", func() {
		BeforeEach(func() {
			project = config.BuildProjectObject(projName, testConfig.TestDepartment.Name, projId,
				testConfig.TestProject.Labels[schedulingv2alpha2.DepartmentIdLabel])
			common.SetProjectDeletion(project, true, false)
			err := kubeClient.Create(ctx, project)
			Expect(err).To(BeNil())
			testCreatedObjects[projName] = project
			common.ValidateProjectReady(ctx, kubeClient, projName)
			namespaceName = fmt.Sprintf("runai-%s", projName)
		})
		AfterEach(func() {
			kubeClient.Delete(ctx, project)
			delete(testCreatedObjects, project.Name)
		})
		It("Delete a Project - Sanity empty", func() {
			//When
			err := kubeClient.Delete(ctx, project)
			Expect(err).To(BeNil())

			// Then
			common.ValidateProjectNotFound(ctx, kubeClient, project.Name)
		})
		It("Delete a Project - Secret blocking", func() {
			var existingProj kaiv1alpha1.Project
			err := kubeClient.Get(ctx, client.ObjectKey{
				Name: project.Name,
			}, &existingProj)
			Expect(err).ToNot(HaveOccurred(), "Failed to get proj")

			//Setup
			secret, err := common.LoadSecret(common.GetRandomName("secret"), namespaceName)
			testCreatedObjects[secret.Name] = secret

			err = kubeClient.Create(ctx, secret)
			Expect(err).ToNot(HaveOccurred(), "Failed to create secret")

			var existingSecret v1.Secret
			err = kubeClient.Get(ctx, client.ObjectKey{
				Name:      secret.Name,
				Namespace: secret.Namespace,
			}, &existingSecret)
			Expect(err).ToNot(HaveOccurred(), "Failed to get secret")

			//When
			err = kubeClient.Delete(ctx, project)
			Expect(err).To(BeNil())

			conditions := common.ValidateProjectDeletionBlockedOnSecret(ctx, kubeClient, project.Name)

			// Then
			By("ProjectConditions are correct", func() {
				Expect(conditions).To(HaveLen(1))
				Expect(conditions[0]).ToNot(BeNil())
				Expect(conditions[0].Type).To(Equal(common.SecretsReadyConditionType))
				Expect(conditions[0].Status).To(Equal(v1.ConditionFalse))
				Expect(conditions[0].Reason).To(Equal("SecretsDeletionHandlerFailed"))
				Expect(conditions[0].Message).To(Equal(fmt.Sprintf("The project couldn't be deleted because the following needs to be deleted first:\nSecret %s\n", secret.Name)))
			})

			err = kubeClient.Delete(ctx, secret)
			Expect(err).To(BeNil())

			common.ValidateProjectNotFound(ctx, kubeClient, project.Name)
		})
	})

	Context("Blocking deletion with force delete", func() {
		BeforeEach(func() {
			project = config.BuildProjectObject(projName, testConfig.TestDepartment.Name, projId,
				testConfig.TestProject.Labels[schedulingv2alpha2.DepartmentIdLabel])
			common.SetProjectDeletion(project, true, true)

			err := kubeClient.Create(ctx, project)
			Expect(err).To(BeNil())
			testCreatedObjects[projName] = project
			common.ValidateProjectReady(ctx, kubeClient, projName)
			namespaceName = fmt.Sprintf("runai-%s", projName)
		})
		AfterEach(func() {
			kubeClient.Delete(ctx, project)
			delete(testCreatedObjects, project.Name)
		})
		It("Delete a Project - has secret but skip blocking", func() {
			//Setup
			secret, err := common.LoadSecret(common.GetRandomName("secret"), namespaceName)
			testCreatedObjects[secret.Name] = secret

			err = kubeClient.Create(ctx, secret)
			Expect(err).ToNot(HaveOccurred(), "Failed to create secret")

			//When
			err = kubeClient.Delete(ctx, project)
			Expect(err).To(BeNil())

			common.ValidateProjectNotFound(ctx, kubeClient, project.Name)
		})
	})
	Context("Non-blocking deletion", func() {
		BeforeEach(func() {
			project = config.BuildProjectObject(projName, testConfig.TestDepartment.Name, projId,
				testConfig.TestProject.Labels[schedulingv2alpha2.DepartmentIdLabel])
			common.SetProjectDeletion(project, false, false)

			err := kubeClient.Create(ctx, project)
			Expect(err).To(BeNil())
			testCreatedObjects[projName] = project
			common.ValidateProjectReady(ctx, kubeClient, projName)
			namespaceName = fmt.Sprintf("runai-%s", projName)
		})
		AfterEach(func() {
			kubeClient.Delete(ctx, project)
			delete(testCreatedObjects, project.Name)
		})
		It("Delete a Project - has secret but skip blocking", func() {
			//Setup
			secret, err := common.LoadSecret(common.GetRandomName("secret"), namespaceName)
			testCreatedObjects[secret.Name] = secret

			err = kubeClient.Create(ctx, secret)
			Expect(err).ToNot(HaveOccurred(), "Failed to create secret")

			//When
			err = kubeClient.Delete(ctx, project)
			Expect(err).To(BeNil())

			common.ValidateProjectNotFound(ctx, kubeClient, project.Name)
		})
	})
})
