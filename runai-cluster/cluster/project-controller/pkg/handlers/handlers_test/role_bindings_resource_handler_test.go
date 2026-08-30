package handlers_test

import (
	"context"
	"os"
	"path"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/common"
	"github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/config"
	. "github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/handlers"
	. "github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/test"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	roleBindingsTestsDir = "../../test/role_bindings_tests_dir"

	roleBindingsCmName      = "runai-project-controller-rolebindings-plugin"
	roleBindingsCmNamespace = "runai"

	projectPvcRoleBindingName = "runai-project-controller-cluster-pvc-per-project"
)

// roleBindingsConfigMap builds a ConfigMap whose data entries are the role
// binding yamls under roleBindingsTestsDir, mirroring how the plugin configmap
// is populated in the cluster.
func roleBindingsConfigMap() *corev1.ConfigMap {
	files, err := os.ReadDir(roleBindingsTestsDir)
	Expect(err).Should(Succeed())

	data := make(map[string]string)
	for _, file := range files {
		content, err := os.ReadFile(path.Join(roleBindingsTestsDir, file.Name()))
		Expect(err).Should(Succeed())
		data[file.Name()] = string(content)
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleBindingsCmName,
			Namespace: roleBindingsCmNamespace,
		},
		Data: data,
	}
}

var _ = Describe("Role Binding Resource Handler", func() {

	var (
		handler                  RoleBindingsResourceHandler
		k8sClient                client.Client
		project, modifiedProject kaiv1alpha1.Project
		namespace                string
		jobControllerRoleBinding rbacv1.RoleBinding
	)

	BeforeEach(func() {
		project = *TestProject.DeepCopy()
		modifiedProject = *KaiModifiedTestProject.DeepCopy()
		jobControllerRoleBinding = *JobControllerRoleBinding.DeepCopy()
		namespace = TestNamespace.Name

		k8sClient = fake.NewClientBuilder().WithScheme(scheme).
			WithObjects(TestNamespace.DeepCopy(), roleBindingsConfigMap()).Build()
		handler = NewRoleBindingsResourceHandler(k8sClient, roleBindingsCmName, roleBindingsCmNamespace, false, false)
	})

	It("Handles a new, non-existing RoleBinding", func() {
		By("Creating a new RoleBinding matching the Project Permissions", func() {

			// When
			conditions, err := handler.HandleResource(project)
			Expect(err).Should(Succeed())

			// Then
			Expect(conditions).To(HaveLen(1))
			Expect(conditions[0]).ToNot(BeNil())
			Expect(conditions[0].Type).To(Equal(kaiv1alpha1.RoleBindingsReady))
			Expect(conditions[0].Status).To(Equal(corev1.ConditionTrue))
			Expect(conditions[0].Reason).To(Equal(""))
			Expect(conditions[0].Message).To(Equal(""))

			jobControllerRoleBindingFound := &rbacv1.RoleBinding{}
			Expect(k8sClient.Get(context.TODO(),
				client.ObjectKey{Name: JobControllerRoleBindingName, Namespace: namespace},
				jobControllerRoleBindingFound)).To(Succeed())
			Expect(jobControllerRoleBindingFound).ToNot(BeZero())
			Expect(jobControllerRoleBindingFound.RoleRef.Name).To(Equal(JobControllerRoleBindingName))
			Expect(jobControllerRoleBindingFound.RoleRef.Kind).To(Equal(ClusterRoleKind))
			Expect(jobControllerRoleBindingFound.RoleRef.APIGroup).To(Equal(RbacGroup))
			Expect(jobControllerRoleBindingFound.Subjects).To(HaveLen(1))
			Expect(jobControllerRoleBindingFound.Subjects[0].Kind).To(Equal(ServiceAccountKind))
			Expect(jobControllerRoleBindingFound.Subjects[0].Name).To(Equal(ProjectControllerServiceAccountName))
			Expect(jobControllerRoleBindingFound.Subjects[0].Namespace).To(Equal(config.Get().InstallNamespace))

			projectPvcRoleBindingFound := &rbacv1.RoleBinding{}
			Expect(k8sClient.Get(context.TODO(),
				client.ObjectKey{Name: projectPvcRoleBindingName, Namespace: namespace},
				projectPvcRoleBindingFound)).To(Succeed())
			Expect(projectPvcRoleBindingFound).ToNot(BeZero())
			Expect(projectPvcRoleBindingFound.RoleRef.Name).To(Equal(projectPvcRoleBindingName))
			Expect(projectPvcRoleBindingFound.RoleRef.Kind).To(Equal(ClusterRoleKind))
			Expect(projectPvcRoleBindingFound.RoleRef.APIGroup).To(Equal(RbacGroup))
			Expect(projectPvcRoleBindingFound.Subjects).To(HaveLen(1))
			Expect(projectPvcRoleBindingFound.Subjects[0].Kind).To(Equal(ServiceAccountKind))
			Expect(projectPvcRoleBindingFound.Subjects[0].Name).To(Equal(ProjectControllerServiceAccountName))
			Expect(projectPvcRoleBindingFound.Subjects[0].Namespace).To(Equal(config.Get().InstallNamespace))
		})
	})

	It("Handles an existing RoleBinding", func() {
		By("RoleBindings still exists", func() {

			// Given
			_ = k8sClient.Create(context.TODO(), &jobControllerRoleBinding)

			// When
			_, err := handler.HandleResource(modifiedProject)
			Expect(err).Should(Succeed())

			// Then
			jobControllerRoleBindingFound := &rbacv1.RoleBinding{}
			Expect(k8sClient.Get(context.TODO(),
				client.ObjectKey{Name: JobControllerRoleBindingName, Namespace: namespace},
				jobControllerRoleBindingFound)).To(Succeed())
			Expect(jobControllerRoleBindingFound).ToNot(BeZero())
			Expect(jobControllerRoleBindingFound.RoleRef.Name).To(Equal(JobControllerRoleBindingName))
			Expect(jobControllerRoleBindingFound.RoleRef.Kind).To(Equal(ClusterRoleKind))
			Expect(jobControllerRoleBindingFound.RoleRef.APIGroup).To(Equal(RbacGroup))
			Expect(jobControllerRoleBindingFound.Subjects).To(HaveLen(1))
			Expect(jobControllerRoleBindingFound.Subjects[0].Kind).To(Equal(ServiceAccountKind))
			Expect(jobControllerRoleBindingFound.Subjects[0].Name).To(Equal(ProjectControllerServiceAccountName))
			Expect(jobControllerRoleBindingFound.Subjects[0].Namespace).To(Equal(config.Get().InstallNamespace))
		})
	})

	Specify("Owner Reference Creation", func() {

		// When
		_, err := handler.HandleResource(project)
		Expect(err).Should(Succeed())

		// Then
		projectPvcRoleBindingFound := &rbacv1.RoleBinding{}
		Expect(k8sClient.Get(context.TODO(),
			client.ObjectKey{Name: projectPvcRoleBindingName, Namespace: namespace},
			projectPvcRoleBindingFound)).To(Succeed())
		Expect(projectPvcRoleBindingFound.OwnerReferences).ToNot(BeZero())
		Expect(len(projectPvcRoleBindingFound.OwnerReferences)).To(Equal(1))
		Expect(projectPvcRoleBindingFound.OwnerReferences[0].APIVersion).To(Equal(project.APIVersion))
		Expect(projectPvcRoleBindingFound.OwnerReferences[0].Kind).To(Equal(project.Kind))
		Expect(projectPvcRoleBindingFound.OwnerReferences[0].Name).To(Equal(project.Name))
		Expect(projectPvcRoleBindingFound.OwnerReferences[0].UID).To(Equal(project.UID))
		Expect(projectPvcRoleBindingFound.OwnerReferences[0].Controller).To(Equal(&TrueRef))
	})

	It("deletes a Project-controlled RoleBinding omitted from the desired ConfigMap", func() {
		obsoleteRoleBinding := &rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{
			Name:      "obsolete-project-role-binding",
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: project.APIVersion,
				Kind:       project.Kind,
				Name:       project.Name,
				UID:        project.UID,
				Controller: &TrueRef,
			}},
		}}
		Expect(k8sClient.Create(context.TODO(), obsoleteRoleBinding)).To(Succeed())

		_, err := handler.HandleResource(project)
		Expect(err).Should(Succeed())

		err = k8sClient.Get(context.TODO(), client.ObjectKey{
			Name: obsoleteRoleBinding.Name, Namespace: namespace,
		}, &rbacv1.RoleBinding{})
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
	})

	It("preserves omitted RoleBindings that are not controlled by the Project", func() {
		unownedRoleBinding := &rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{
			Name: "workload-controller-rw", Namespace: namespace,
		}}
		otherProjectRoleBinding := &rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{
			Name:      "other-project-role-binding",
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: project.APIVersion,
				Kind:       project.Kind,
				Name:       "other-project",
				UID:        "other-project-uid",
				Controller: &TrueRef,
			}},
		}}
		Expect(k8sClient.Create(context.TODO(), unownedRoleBinding)).To(Succeed())
		Expect(k8sClient.Create(context.TODO(), otherProjectRoleBinding)).To(Succeed())

		_, err := handler.HandleResource(project)
		Expect(err).Should(Succeed())

		Expect(k8sClient.Get(context.TODO(), client.ObjectKey{
			Name: unownedRoleBinding.Name, Namespace: namespace,
		}, &rbacv1.RoleBinding{})).To(Succeed())
		Expect(k8sClient.Get(context.TODO(), client.ObjectKey{
			Name: otherProjectRoleBinding.Name, Namespace: namespace,
		}, &rbacv1.RoleBinding{})).To(Succeed())
	})

	It("applies a new ConfigMap version to an existing Project", func() {
		_, err := handler.HandleResource(project)
		Expect(err).Should(Succeed())

		configMap := &corev1.ConfigMap{}
		Expect(k8sClient.Get(context.TODO(), client.ObjectKey{
			Name: roleBindingsCmName, Namespace: roleBindingsCmNamespace,
		}, configMap)).To(Succeed())
		delete(configMap.Data, "job-controller.yaml")
		Expect(k8sClient.Update(context.TODO(), configMap)).To(Succeed())

		_, err = handler.HandleResource(project)
		Expect(err).Should(Succeed())

		err = k8sClient.Get(context.TODO(), client.ObjectKey{
			Name: JobControllerRoleBindingName, Namespace: namespace,
		}, &rbacv1.RoleBinding{})
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
	})

	It("does not delete RoleBindings when any ConfigMap entry is malformed", func() {
		configMap := &corev1.ConfigMap{}
		Expect(k8sClient.Get(context.TODO(), client.ObjectKey{
			Name: roleBindingsCmName, Namespace: roleBindingsCmNamespace,
		}, configMap)).To(Succeed())
		configMap.Data = map[string]string{"broken.yaml": "metadata: ["}
		Expect(k8sClient.Update(context.TODO(), configMap)).To(Succeed())

		obsoleteRoleBinding := &rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{
			Name:      "obsolete-project-role-binding",
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: project.APIVersion,
				Kind:       project.Kind,
				Name:       project.Name,
				UID:        project.UID,
				Controller: &TrueRef,
			}},
		}}
		Expect(k8sClient.Create(context.TODO(), obsoleteRoleBinding)).To(Succeed())

		_, err := handler.HandleResource(project)
		Expect(err).To(HaveOccurred())

		Expect(k8sClient.Get(context.TODO(), client.ObjectKey{
			Name: obsoleteRoleBinding.Name, Namespace: namespace,
		}, &rbacv1.RoleBinding{})).To(Succeed())
	})
})
