/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package project

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants/oauth2proxy"
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/scheme-init"
	schedulingv2alpha2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	cwTestSchedulerNamespace = "runai"
	cwTestProjectNamespace   = "ns-proj-1"
	cwTestProjectName        = "proj-1"
	cwTestProjectUID         = types.UID("proj-1-uid")
	cwTestDepartment1        = "1"
	cwTestDepartment2        = "2"
	ProjectKind              = "Project"
)

func cwTestProject() *kaiorg.Project {
	return &kaiorg.Project{
		TypeMeta: metav1.TypeMeta{Kind: ProjectKind, APIVersion: kaiorg.GroupVersion.Identifier()},
		ObjectMeta: metav1.ObjectMeta{
			Name: cwTestProjectName,
			UID:  cwTestProjectUID,
			Labels: map[string]string{
				schedulingv2alpha2.DepartmentIdLabel: cwTestDepartment1,
			},
		},
		Status: kaiorg.ProjectStatus{Namespace: cwTestProjectNamespace},
	}
}

func cwProjectOwnerRef() []metav1.OwnerReference {
	return []metav1.OwnerReference{{
		APIVersion: kaiorg.GroupVersion.Identifier(),
		Kind:       ProjectKind,
		Name:       cwTestProjectName,
		UID:        cwTestProjectUID,
	}}
}

func cwPvc(name, namespace string, labels map[string]string) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{Kind: "PersistentVolumeClaim", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
		},
	}
}

func cwSecret(name, namespace string, labels map[string]string, secretType corev1.SecretType,
	data map[string][]byte) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta:   metav1.TypeMeta{Kind: "Secret", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: labels},
		Type:       secretType,
		Data:       data,
	}
}

var _ = Describe("Cluster-wide PVC replication", func() {
	var (
		k8sClient client.Client
		handler   *ClusterWideReplicatedResourceHandler
		gvkCommon *ClusterWideGVKCommon
		project   *kaiorg.Project
		ctx       context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		project = cwTestProject()
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: cwTestProjectNamespace},
		}

		scheme := scheme_init.Scheme()
		Expect(kaiorg.AddToScheme(scheme)).To(Succeed())
		k8sClient = fake.NewClientBuilder().
			WithScheme(scheme).
			WithStatusSubresource(&kaiorg.Project{}).
			WithObjects(project.DeepCopy(), ns).
			Build()

		var err error
		gvkCommon, err = NewClusterWideGVKCommon(context.Background(), k8sClient, &ClusterWideReplicatorGVKConfig{ClusterWidePvcs: true})
		Expect(err).To(BeNil())
		handler = NewClusterWideReplicatedResourceHandler(k8sClient, gvkCommon)
	})

	It("replicates a cluster-wide PVC into the project namespace with an owner reference", func() {
		src := cwPvc("cluster-wide-pvc", cwTestSchedulerNamespace,
			map[string]string{commonconstants.ClusterWideResourceLabel: "true"})
		src.Spec.VolumeName = "some-volume"
		Expect(k8sClient.Create(ctx, src)).To(Succeed())

		err := handler.HandleResource(ctx, *project)
		Expect(err).To(Succeed())

		replicated := &corev1.PersistentVolumeClaim{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "cluster-wide-pvc", Namespace: cwTestProjectNamespace}, replicated)).To(Succeed())
		// VolumeName must be cleared on the replicated copy (one-to-one PV binding).
		Expect(replicated.Spec.VolumeName).To(Equal(""))
		Expect(replicated.Labels).To(HaveKeyWithValue(commonconstants.ClusterWideResourceLabel, "true"))
		Expect(replicated.Labels[replicatedProjectNameLabelKey]).To(Equal(project.Name))
		Expect(replicated.Labels[replicatedProjectIdLabelKey]).To(Equal(string(project.UID)))
	})

	It("deletes a namespaced PVC owned by the project whose source no longer exists, but keeps unlabeled ones", func() {
		orphan := cwPvc("cluster-wide-pvc", cwTestProjectNamespace,
			map[string]string{commonconstants.ClusterWideResourceLabel: "true"})
		orphan.OwnerReferences = cwProjectOwnerRef()
		Expect(k8sClient.Create(ctx, orphan)).To(Succeed())

		unlabeled := cwPvc("user-pvc", cwTestProjectNamespace, nil)
		Expect(k8sClient.Create(ctx, unlabeled)).To(Succeed())

		err := handler.HandleResource(ctx, *project)
		Expect(err).To(Succeed())

		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "cluster-wide-pvc", Namespace: cwTestProjectNamespace},
			&corev1.PersistentVolumeClaim{})).ToNot(Succeed())
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "user-pvc", Namespace: cwTestProjectNamespace},
			&corev1.PersistentVolumeClaim{})).To(Succeed())
	})

	It("replicates department-wide PVCs matching the project's department and ignores other departments", func() {
		dep1 := cwPvc("dep1-pvc", cwTestSchedulerNamespace,
			map[string]string{commonconstants.DepartmentScopeLabel: cwTestDepartment1})
		dep2 := cwPvc("dep2-pvc", cwTestSchedulerNamespace,
			map[string]string{commonconstants.DepartmentScopeLabel: cwTestDepartment2})
		Expect(k8sClient.Create(ctx, dep1)).To(Succeed())
		Expect(k8sClient.Create(ctx, dep2)).To(Succeed())

		err := handler.HandleResource(ctx, *project)
		Expect(err).To(Succeed())

		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "dep1-pvc", Namespace: cwTestProjectNamespace},
			&corev1.PersistentVolumeClaim{})).To(Succeed())
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "dep2-pvc", Namespace: cwTestProjectNamespace},
			&corev1.PersistentVolumeClaim{})).ToNot(Succeed())
	})

	It("preserves the VolumeName of an existing bound namespaced PVC on update", func() {
		src := cwPvc("cluster-wide-pvc", cwTestSchedulerNamespace,
			map[string]string{commonconstants.ClusterWideResourceLabel: "true"})
		src.Annotations = map[string]string{"run.ai/asset-name": "new"}
		Expect(k8sClient.Create(ctx, src)).To(Succeed())

		existing := cwPvc("cluster-wide-pvc", cwTestProjectNamespace,
			map[string]string{commonconstants.ClusterWideResourceLabel: "true"})
		existing.OwnerReferences = cwProjectOwnerRef()
		existing.Spec.VolumeName = "pvc-bound-volume-12345"
		existing.Annotations = map[string]string{"run.ai/asset-name": "old"}
		Expect(k8sClient.Create(ctx, existing)).To(Succeed())

		err := handler.HandleResource(ctx, *project)
		Expect(err).To(Succeed())

		updated := &corev1.PersistentVolumeClaim{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "cluster-wide-pvc", Namespace: cwTestProjectNamespace}, updated)).To(Succeed())
		Expect(updated.Spec.VolumeName).To(Equal("pvc-bound-volume-12345"))
		Expect(updated.Annotations).To(HaveKeyWithValue("run.ai/asset-name", "new"))
	})

	It("keeps an empty VolumeName for an existing pending namespaced PVC on update", func() {
		src := cwPvc("cluster-wide-pvc", cwTestSchedulerNamespace,
			map[string]string{commonconstants.ClusterWideResourceLabel: "true"})
		src.Annotations = map[string]string{"run.ai/asset-name": "new"}
		Expect(k8sClient.Create(ctx, src)).To(Succeed())

		existing := cwPvc("cluster-wide-pvc", cwTestProjectNamespace,
			map[string]string{commonconstants.ClusterWideResourceLabel: "true"})
		existing.OwnerReferences = cwProjectOwnerRef()
		existing.Spec.VolumeName = "" // not bound yet
		existing.Status.Phase = corev1.ClaimPending
		Expect(k8sClient.Create(ctx, existing)).To(Succeed())

		err := handler.HandleResource(ctx, *project)
		Expect(err).To(Succeed())

		updated := &corev1.PersistentVolumeClaim{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "cluster-wide-pvc", Namespace: cwTestProjectNamespace}, updated)).To(Succeed())
		Expect(updated.Spec.VolumeName).To(Equal(""))
		Expect(updated.Annotations).To(HaveKeyWithValue("run.ai/asset-name", "new"))
	})

	It("deletes a namespaced department-wide PVC owned by the project whose source no longer exists", func() {
		orphan := cwPvc("dep1-pvc", cwTestProjectNamespace,
			map[string]string{commonconstants.DepartmentScopeLabel: cwTestDepartment1})
		orphan.OwnerReferences = cwProjectOwnerRef()
		Expect(k8sClient.Create(ctx, orphan)).To(Succeed())

		// A department-wide PVC for a different department must not be replicated in.
		dep2 := cwPvc("dep2-pvc", cwTestSchedulerNamespace,
			map[string]string{commonconstants.DepartmentScopeLabel: cwTestDepartment2})
		Expect(k8sClient.Create(ctx, dep2)).To(Succeed())

		err := handler.HandleResource(ctx, *project)
		Expect(err).To(Succeed())

		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "dep1-pvc", Namespace: cwTestProjectNamespace},
			&corev1.PersistentVolumeClaim{})).ToNot(Succeed())
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "dep2-pvc", Namespace: cwTestProjectNamespace},
			&corev1.PersistentVolumeClaim{})).ToNot(Succeed())
	})
})

var _ = Describe("Cluster-wide Secret replication", func() {
	var (
		k8sClient client.Client
		handler   *ClusterWideReplicatedResourceHandler
		project   *kaiorg.Project
		ctx       context.Context
	)

	dockerLabels := map[string]string{
		commonconstants.ClusterWideResourceLabel: "true",
		commonconstants.RunaiResourceLabel:       DockerRegistryAssetKind,
	}
	genericLabels := map[string]string{
		commonconstants.ClusterWideResourceLabel: "true",
		commonconstants.RunaiResourceLabel:       GenericSecretAssetKind,
	}

	BeforeEach(func() {
		ctx = context.Background()
		project = cwTestProject()
		defaultSA := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Name: DefaultServiceAccountName, Namespace: cwTestProjectNamespace},
		}
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: cwTestProjectNamespace},
		}

		scheme := scheme_init.Scheme()
		Expect(kaiorg.AddToScheme(scheme)).To(Succeed())
		k8sClient = fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(project.DeepCopy(), defaultSA, ns).
			Build()

		gvkCommon, err := NewClusterWideGVKCommon(context.Background(), k8sClient, &ClusterWideReplicatorGVKConfig{ClusterWideSecrets: true})
		Expect(err).To(BeNil())
		handler = NewClusterWideReplicatedResourceHandler(k8sClient, gvkCommon)
	})

	It("replicates a cluster-wide docker-registry secret and binds it to the namespace default ServiceAccount", func() {
		secret := cwSecret("cw-docker", cwTestSchedulerNamespace, dockerLabels,
			corev1.SecretTypeDockerConfigJson, map[string][]byte{".dockerconfigjson": []byte("{}")})
		Expect(k8sClient.Create(ctx, secret)).To(Succeed())

		err := handler.HandleResource(ctx, *project)
		Expect(err).To(Succeed())

		replicated := &corev1.Secret{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "cw-docker", Namespace: cwTestProjectNamespace}, replicated)).To(Succeed())
		Expect(replicated.Labels[replicatedProjectNameLabelKey]).To(Equal(project.Name))
		Expect(replicated.Labels[replicatedProjectIdLabelKey]).To(Equal(string(project.UID)))

		sa := &corev1.ServiceAccount{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: DefaultServiceAccountName, Namespace: cwTestProjectNamespace}, sa)).To(Succeed())
		Expect(sa.ImagePullSecrets).To(ContainElement(corev1.LocalObjectReference{Name: "cw-docker"}))
	})

	It("replicates a cluster-wide non-docker secret without binding it to the ServiceAccount", func() {
		secret := cwSecret("cw-generic", cwTestSchedulerNamespace, genericLabels,
			corev1.SecretTypeOpaque, map[string][]byte{"token": []byte("abc")})
		Expect(k8sClient.Create(ctx, secret)).To(Succeed())

		err := handler.HandleResource(ctx, *project)
		Expect(err).To(Succeed())

		replicated := &corev1.Secret{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "cw-generic", Namespace: cwTestProjectNamespace}, replicated)).To(Succeed())
		Expect(replicated.Data).To(HaveKeyWithValue("token", []byte("abc")))

		sa := &corev1.ServiceAccount{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: DefaultServiceAccountName, Namespace: cwTestProjectNamespace}, sa)).To(Succeed())
		Expect(sa.ImagePullSecrets).To(BeEmpty())
	})

	It("does not modify the ServiceAccount when no cluster-wide secret exists", func() {
		// Only non-cluster-wide secrets exist in the runai namespace.
		unrelated := cwSecret("unrelated", cwTestSchedulerNamespace, nil,
			corev1.SecretTypeDockerConfigJson, map[string][]byte{".dockerconfigjson": []byte("{}")})
		Expect(k8sClient.Create(ctx, unrelated)).To(Succeed())

		err := handler.HandleResource(ctx, *project)
		Expect(err).To(Succeed())

		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "unrelated", Namespace: cwTestProjectNamespace},
			&corev1.Secret{})).ToNot(Succeed())

		sa := &corev1.ServiceAccount{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: DefaultServiceAccountName, Namespace: cwTestProjectNamespace}, sa)).To(Succeed())
		Expect(sa.ImagePullSecrets).To(BeEmpty())
	})

	It("updates an existing namespaced secret with the data of the cluster-wide secret", func() {
		src := cwSecret("cw-docker", cwTestSchedulerNamespace, dockerLabels,
			corev1.SecretTypeDockerConfigJson, map[string][]byte{".dockerconfigjson": []byte(`{"new":true}`)})
		Expect(k8sClient.Create(ctx, src)).To(Succeed())

		existing := cwSecret("cw-docker", cwTestProjectNamespace, dockerLabels,
			corev1.SecretTypeDockerConfigJson, map[string][]byte{".dockerconfigjson": []byte(`{"old":true}`)})
		existing.OwnerReferences = cwProjectOwnerRef()
		Expect(k8sClient.Create(ctx, existing)).To(Succeed())

		err := handler.HandleResource(ctx, *project)
		Expect(err).To(Succeed())

		updated := &corev1.Secret{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "cw-docker", Namespace: cwTestProjectNamespace}, updated)).To(Succeed())
		Expect(updated.Data).To(HaveKeyWithValue(".dockerconfigjson", []byte(`{"new":true}`)))
	})

	It("deletes an unreferenced namespaced secret and removes it from the ServiceAccount", func() {
		orphan := cwSecret("cw-docker", cwTestProjectNamespace, dockerLabels,
			corev1.SecretTypeDockerConfigJson, map[string][]byte{".dockerconfigjson": []byte("{}")})
		orphan.OwnerReferences = cwProjectOwnerRef()
		Expect(k8sClient.Create(ctx, orphan)).To(Succeed())

		unlabeled := cwSecret("user-secret", cwTestProjectNamespace, nil,
			corev1.SecretTypeDockerConfigJson, map[string][]byte{".dockerconfigjson": []byte("{}")})
		Expect(k8sClient.Create(ctx, unlabeled)).To(Succeed())

		sa := &corev1.ServiceAccount{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: DefaultServiceAccountName, Namespace: cwTestProjectNamespace}, sa)).To(Succeed())
		sa.ImagePullSecrets = []corev1.LocalObjectReference{{Name: "cw-docker"}}
		Expect(k8sClient.Update(ctx, sa)).To(Succeed())

		err := handler.HandleResource(ctx, *project)
		Expect(err).To(Succeed())

		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "cw-docker", Namespace: cwTestProjectNamespace},
			&corev1.Secret{})).ToNot(Succeed())
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "user-secret", Namespace: cwTestProjectNamespace},
			&corev1.Secret{})).To(Succeed())

		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: DefaultServiceAccountName, Namespace: cwTestProjectNamespace}, sa)).To(Succeed())
		Expect(sa.ImagePullSecrets).To(BeEmpty())
	})
})

var _ = Describe("Cluster-wide OAuth2 template ConfigMap replication", func() {
	var (
		k8sClient client.Client
		handler   *ClusterWideReplicatedResourceHandler
		project   *kaiorg.Project
		source    *corev1.ConfigMap
		ctx       context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		project = cwTestProject()
		source = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: oauth2proxy.ConfigMapName, Namespace: cwTestSchedulerNamespace,
				Labels: map[string]string{commonconstants.ClusterWideResourceLabel: "true"},
			},
			Data: map[string]string{oauth2proxy.TemplateKey: "<html>Run:ai error</html>"},
		}
		scheme := scheme_init.Scheme()
		Expect(kaiorg.AddToScheme(scheme)).To(Succeed())
		k8sClient = fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&kaiorg.Project{}).
			WithObjects(project.DeepCopy(), &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{Name: cwTestProjectNamespace},
			}).Build()
		gvkCommon, err := NewClusterWideGVKCommon(ctx, k8sClient,
			&ClusterWideReplicatorGVKConfig{ClusterWideConfigMaps: true})
		Expect(err).To(Succeed())
		handler = NewClusterWideReplicatedResourceHandler(k8sClient, gvkCommon)
		Expect(k8sClient.Create(ctx, source)).To(Succeed())
	})

	It("replicates the source to existing and subsequently created project namespaces when reconciled", func() {
		Expect(handler.HandleResource(ctx, *project)).To(Succeed())
		replica := &corev1.ConfigMap{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: source.Name, Namespace: project.Status.Namespace}, replica)).To(Succeed())
		Expect(replica.Data).To(Equal(source.Data))
		Expect(replica.Labels).To(HaveKeyWithValue(replicatedProjectNameLabelKey, project.Name))
		Expect(replica.Labels).To(HaveKeyWithValue(replicatedProjectIdLabelKey, string(project.UID)))
		Expect(replica.OwnerReferences).To(BeEmpty())

		newProject := cwTestProject()
		newProject.Name, newProject.UID, newProject.Status.Namespace = "proj-2", "proj-2-uid", "ns-proj-2"
		Expect(k8sClient.Create(ctx, newProject)).To(Succeed())
		Expect(k8sClient.Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: newProject.Status.Namespace}})).To(Succeed())
		Expect(handler.HandleResource(ctx, *newProject)).To(Succeed())
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: source.Name, Namespace: newProject.Status.Namespace}, replica)).To(Succeed())
		Expect(replica.Data).To(Equal(source.Data))
		Expect(replica.Labels).To(HaveKeyWithValue(replicatedProjectNameLabelKey, newProject.Name))
	})

	It("propagates source changes on the next project reconcile", func() {
		Expect(handler.HandleResource(ctx, *project)).To(Succeed())
		source.Data[oauth2proxy.TemplateKey] = "<html>Updated Run:ai error</html>"
		Expect(k8sClient.Update(ctx, source)).To(Succeed())
		Expect(handler.HandleResource(ctx, *project)).To(Succeed())
		replica := &corev1.ConfigMap{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: source.Name, Namespace: project.Status.Namespace}, replica)).To(Succeed())
		Expect(replica.Data).To(Equal(source.Data))
	})

	It("adopts an ownerless existing copy even when its template already matches", func() {
		existing := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: source.Name, Namespace: project.Status.Namespace, UID: "existing-template-uid",
			},
			Data: source.DeepCopy().Data,
		}
		Expect(k8sClient.Create(ctx, existing)).To(Succeed())
		Expect(handler.HandleResource(ctx, *project)).To(Succeed())
		replica := &corev1.ConfigMap{}
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(existing), replica)).To(Succeed())
		Expect(replica.UID).To(Equal(existing.UID))
		Expect(replica.Data).To(Equal(source.Data))
		Expect(replica.Labels).To(HaveKeyWithValue(commonconstants.ClusterWideResourceLabel, "true"))
		Expect(replica.Labels).To(HaveKeyWithValue(replicatedProjectNameLabelKey, project.Name))
		Expect(replica.Labels).To(HaveKeyWithValue(replicatedProjectIdLabelKey, string(project.UID)))
		Expect(replica.OwnerReferences).To(BeEmpty())
	})

	It("preserves a project template marked as manually overridden", func() {
		existing := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: source.Name, Namespace: project.Status.Namespace,
				Labels: map[string]string{ResourceManualOverrideLabel: "true"},
			},
			Data: map[string]string{oauth2proxy.TemplateKey: "<html>Project error</html>"},
		}
		Expect(k8sClient.Create(ctx, existing)).To(Succeed())
		Expect(handler.HandleResource(ctx, *project)).To(Succeed())
		replica := &corev1.ConfigMap{}
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(existing), replica)).To(Succeed())
		Expect(replica.Data).To(Equal(existing.Data))
		Expect(replica.Labels).NotTo(HaveKey(replicatedProjectNameLabelKey))
	})
})

var _ = Describe("PVC inspector helpers", func() {
	var gvkCommon *ClusterWideGVKCommon

	BeforeEach(func() {
		scheme := scheme_init.Scheme()
		Expect(kaiorg.AddToScheme(scheme)).To(Succeed())
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		var err error
		gvkCommon, err = NewClusterWideGVKCommon(context.Background(), k8sClient, &ClusterWideReplicatorGVKConfig{
			ClusterWidePvcs: true, ClusterWideSecrets: true, ClusterWideConfigMaps: true,
		})
		Expect(err).To(BeNil())
	})

	It("DeepCopySpec copies the PVC spec but clears VolumeName", func() {
		src := cwPvc("p", "runai", nil)
		src.Spec.VolumeName = "pvc-source-volume"
		target := &corev1.PersistentVolumeClaim{}
		gvkCommon.DeepCopySpec(context.Background(), src, target)
		Expect(target.Spec.AccessModes).To(Equal(src.Spec.AccessModes))
		Expect(target.Spec.VolumeName).To(Equal(""))
	})

	It("PreserveImmutableFields keeps the existing bound VolumeName", func() {
		existing := cwPvc("p", "ns", nil)
		existing.Spec.VolumeName = "pvc-existing-volume"
		newPvc := cwPvc("p", "ns", nil)
		newPvc.Spec.VolumeName = ""
		gvkCommon.PreserveImmutableFields(existing, newPvc)
		Expect(newPvc.Spec.VolumeName).To(Equal("pvc-existing-volume"))
	})

	It("PreserveImmutableFields does not set a VolumeName when the existing PVC has none", func() {
		existing := cwPvc("p", "ns", nil)
		existing.Spec.VolumeName = "" // not bound yet
		newPvc := cwPvc("p", "ns", nil)
		newPvc.Spec.VolumeName = ""
		gvkCommon.PreserveImmutableFields(existing, newPvc)
		Expect(newPvc.Spec.VolumeName).To(Equal(""))
	})

	It("PreserveImmutableFields overwrites an incorrect VolumeName on the new PVC", func() {
		existing := cwPvc("p", "ns", nil)
		existing.Spec.VolumeName = "pvc-correct-volume"
		newPvc := cwPvc("p", "ns", nil)
		newPvc.Spec.VolumeName = "pvc-wrong-volume"
		gvkCommon.PreserveImmutableFields(existing, newPvc)
		Expect(newPvc.Spec.VolumeName).To(Equal("pvc-correct-volume"))
	})

	It("PreserveImmutableFields does not affect non-PVC objects", func() {
		existing := &corev1.Secret{Data: map[string][]byte{"a": []byte("1")}}
		newSecret := &corev1.Secret{Data: map[string][]byte{"new": []byte("v")}}
		gvkCommon.PreserveImmutableFields(existing, newSecret)
		Expect(newSecret.Data).To(Equal(map[string][]byte{"new": []byte("v")}))
	})
})

// cwRct builds a minimal ResourceClaimTemplate with the single-GPU request the
// assets-sync agent produces. Callers pass a labels map (scope + asset label)
// so the different Describes can exercise cluster-wide, department-wide, and
// non-scoped sources from the same helper. The labels map is cloned so that
// later per-object label mutations don't leak back into the caller's map.
func cwRct(name, namespace string, labels map[string]string) *resourcev1.ResourceClaimTemplate {
	var labelsCopy map[string]string
	if labels != nil {
		labelsCopy = make(map[string]string, len(labels))
		for k, v := range labels {
			labelsCopy[k] = v
		}
	}
	return &resourcev1.ResourceClaimTemplate{
		TypeMeta: metav1.TypeMeta{Kind: "ResourceClaimTemplate", APIVersion: resourcev1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labelsCopy,
		},
		Spec: resourcev1.ResourceClaimTemplateSpec{
			Spec: resourcev1.ResourceClaimSpec{
				Devices: resourcev1.DeviceClaim{
					Requests: []resourcev1.DeviceRequest{{
						Name: "gpu",
						Exactly: &resourcev1.ExactDeviceRequest{
							DeviceClassName: "nvidia.com/gpu",
							Tolerations: []resourcev1.DeviceToleration{{
								Key:   "nvidia.com/gpu",
								Value: "present",
							}},
						},
					}},
				},
			},
		},
	}
}

// cwRctFullyDefaulted returns a copy of cwRct(...) with the server-side
// defaults the apiserver would stamp (AllocationMode=ExactCount, Count=1,
// Toleration.Operator=Equal) already set. Used by the "no churn after
// defaults" test to simulate what an on-cluster read-back of a replicated
// RCT looks like.
func cwRctFullyDefaulted(name, namespace string, labels map[string]string) *resourcev1.ResourceClaimTemplate {
	rct := cwRct(name, namespace, labels)
	rct.Spec.Spec.Devices.Requests[0].Exactly.AllocationMode = resourcev1.DeviceAllocationModeExactCount
	rct.Spec.Spec.Devices.Requests[0].Exactly.Count = 1
	rct.Spec.Spec.Devices.Requests[0].Exactly.Tolerations[0].Operator = resourcev1.DeviceTolerationOpEqual
	return rct
}

var _ = Describe("Cluster-wide ResourceClaimTemplate replication", func() {
	var (
		k8sClient client.Client
		handler   *ClusterWideReplicatedResourceHandler
		project   *kaiorg.Project
		ctx       context.Context
	)

	rctClusterLabels := map[string]string{
		commonconstants.ClusterWideResourceLabel: "true",
		commonconstants.RunaiAssetLabel:          ResourceClaimTemplateAssetKind,
	}
	rctDep1Labels := map[string]string{
		commonconstants.DepartmentScopeLabel: cwTestDepartment1,
		commonconstants.RunaiAssetLabel:      ResourceClaimTemplateAssetKind,
	}
	rctDep2Labels := map[string]string{
		commonconstants.DepartmentScopeLabel: cwTestDepartment2,
		commonconstants.RunaiAssetLabel:      ResourceClaimTemplateAssetKind,
	}

	BeforeEach(func() {
		ctx = context.Background()
		project = cwTestProject()
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: cwTestProjectNamespace},
		}

		scheme := scheme_init.Scheme()
		Expect(kaiorg.AddToScheme(scheme)).To(Succeed())
		k8sClient = fake.NewClientBuilder().
			WithScheme(scheme).
			WithStatusSubresource(&kaiorg.Project{}).
			WithObjects(project.DeepCopy(), ns).
			Build()

		gvkCommon, err := NewClusterWideGVKCommon(ctx, k8sClient,
			&ClusterWideReplicatorGVKConfig{ClusterWideResourceClaimTemplates: true})
		Expect(err).To(BeNil())
		handler = NewClusterWideReplicatedResourceHandler(k8sClient, gvkCommon)
	})

	It("replicates a cluster-wide RCT into the project namespace with the replicated-project labels", func() {
		src := cwRct("cw-rct", cwTestSchedulerNamespace, rctClusterLabels)
		Expect(k8sClient.Create(ctx, src)).To(Succeed())

		Expect(handler.HandleResource(ctx, *project)).To(Succeed())

		replicated := &resourcev1.ResourceClaimTemplate{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "cw-rct", Namespace: cwTestProjectNamespace}, replicated)).To(Succeed())
		Expect(replicated.Labels).To(HaveKeyWithValue(commonconstants.ClusterWideResourceLabel, "true"))
		Expect(replicated.Labels).To(HaveKeyWithValue(replicatedProjectNameLabelKey, project.Name))
		Expect(replicated.Labels).To(HaveKeyWithValue(replicatedProjectIdLabelKey, string(project.UID)))
		Expect(replicated.Spec).To(Equal(src.Spec))
	})

	It("replicates a department-wide RCT matching the project's department and ignores other departments", func() {
		dep1 := cwRct("dep1-rct", cwTestSchedulerNamespace, rctDep1Labels)
		dep2 := cwRct("dep2-rct", cwTestSchedulerNamespace, rctDep2Labels)
		Expect(k8sClient.Create(ctx, dep1)).To(Succeed())
		Expect(k8sClient.Create(ctx, dep2)).To(Succeed())

		Expect(handler.HandleResource(ctx, *project)).To(Succeed())

		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "dep1-rct", Namespace: cwTestProjectNamespace},
			&resourcev1.ResourceClaimTemplate{})).To(Succeed())
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "dep2-rct", Namespace: cwTestProjectNamespace},
			&resourcev1.ResourceClaimTemplate{})).ToNot(Succeed())
	})

	It("does not touch a scope-wide-labeled RCT in runai that lacks the asset-sync label (safety filter)", func() {
		// Carries the cluster-wide scope label but NOT the run.ai/asset
		// label. The RCT inspector's AdditionalObjectFilter is supposed
		// to keep it out of the replicated set — otherwise a manually
		// created RCT in runai would end up in every project.
		src := cwRct("unmanaged-rct", cwTestSchedulerNamespace,
			map[string]string{commonconstants.ClusterWideResourceLabel: "true"})
		Expect(k8sClient.Create(ctx, src)).To(Succeed())

		Expect(handler.HandleResource(ctx, *project)).To(Succeed())
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "unmanaged-rct", Namespace: cwTestProjectNamespace},
			&resourcev1.ResourceClaimTemplate{})).ToNot(Succeed())
	})

	It("does not recreate the replica when the only diff is API-server defaults (no churn)", func() {
		// Source as sent by the user: zero AllocationMode/Count/Operator.
		src := cwRct("cw-rct", cwTestSchedulerNamespace, rctClusterLabels)
		Expect(k8sClient.Create(ctx, src)).To(Succeed())

		// Existing replica as it would be read back from the apiserver:
		// same request, but with server-side defaults populated. Pin a
		// known UID so a recreate would move it.
		existing := cwRctFullyDefaulted("cw-rct", cwTestProjectNamespace, rctClusterLabels)
		existing.Labels[replicatedProjectNameLabelKey] = project.Name
		existing.Labels[replicatedProjectIdLabelKey] = string(project.UID)
		existing.UID = "existing-stable-uid"
		Expect(k8sClient.Create(ctx, existing)).To(Succeed())

		Expect(handler.HandleResource(ctx, *project)).To(Succeed())

		after := &resourcev1.ResourceClaimTemplate{}
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(existing), after)).To(Succeed())
		Expect(after.UID).To(Equal(existing.UID), "RCT must not churn on a defaults-only diff")
	})

	It("delete-and-recreates the replica when the source spec actually changes", func() {
		// Source bumped to 4 GPUs.
		src := cwRct("cw-rct", cwTestSchedulerNamespace, rctClusterLabels)
		src.Spec.Spec.Devices.Requests[0].Exactly.Count = 4
		Expect(k8sClient.Create(ctx, src)).To(Succeed())

		// Existing replica is at the old (Count=1) spec, fully defaulted
		// as the apiserver would store it.
		existing := cwRctFullyDefaulted("cw-rct", cwTestProjectNamespace, rctClusterLabels)
		existing.Labels[replicatedProjectNameLabelKey] = project.Name
		existing.Labels[replicatedProjectIdLabelKey] = string(project.UID)
		existing.UID = "old-replica-uid"
		Expect(k8sClient.Create(ctx, existing)).To(Succeed())

		Expect(handler.HandleResource(ctx, *project)).To(Succeed())

		after := &resourcev1.ResourceClaimTemplate{}
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(existing), after)).To(Succeed())
		Expect(after.UID).NotTo(Equal(existing.UID), "RCT must be re-created when its immutable spec diverges")
		Expect(after.Spec.Spec.Devices.Requests[0].Exactly.Count).To(BeEquivalentTo(4))
	})

	It("deletes a replica owned by the project whose source no longer exists, but keeps unlabeled ones", func() {
		orphan := cwRct("cw-rct", cwTestProjectNamespace, rctClusterLabels)
		orphan.Labels[replicatedProjectNameLabelKey] = project.Name
		orphan.Labels[replicatedProjectIdLabelKey] = string(project.UID)
		orphan.OwnerReferences = cwProjectOwnerRef()
		Expect(k8sClient.Create(ctx, orphan)).To(Succeed())

		// A hand-authored RCT in the project namespace with no scope
		// label and no replicated-project label — must not be swept.
		userRct := cwRct("user-rct", cwTestProjectNamespace, nil)
		Expect(k8sClient.Create(ctx, userRct)).To(Succeed())

		Expect(handler.HandleResource(ctx, *project)).To(Succeed())

		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "cw-rct", Namespace: cwTestProjectNamespace},
			&resourcev1.ResourceClaimTemplate{})).ToNot(Succeed())
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "user-rct", Namespace: cwTestProjectNamespace},
			&resourcev1.ResourceClaimTemplate{})).To(Succeed())
	})
})

var _ = Describe("RCT inspector helpers", func() {
	var gvkCommon *ClusterWideGVKCommon

	BeforeEach(func() {
		scheme := scheme_init.Scheme()
		Expect(kaiorg.AddToScheme(scheme)).To(Succeed())
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		var err error
		gvkCommon, err = NewClusterWideGVKCommon(context.Background(), k8sClient,
			&ClusterWideReplicatorGVKConfig{ClusterWideResourceClaimTemplates: true})
		Expect(err).To(BeNil())
	})

	It("IsSpecImmutable is true for RCT and false for PVC/Secret/ConfigMap", func() {
		Expect(gvkCommon.IsSpecImmutable(&resourcev1.ResourceClaimTemplate{})).To(BeTrue())
		Expect(gvkCommon.IsSpecImmutable(&corev1.PersistentVolumeClaim{})).To(BeFalse())
		Expect(gvkCommon.IsSpecImmutable(&corev1.Secret{})).To(BeFalse())
		Expect(gvkCommon.IsSpecImmutable(&corev1.ConfigMap{})).To(BeFalse())
	})

	It("AreEqual treats a defaults-only diff as equal", func() {
		undefaulted := cwRct("r", "runai", nil)
		defaulted := cwRctFullyDefaulted("r", "runai", nil)
		Expect(gvkCommon.AreEqual(context.Background(), undefaulted, defaulted)).To(BeTrue())
	})

	It("AreEqual flags a real spec diff (Count change)", func() {
		a := cwRct("r", "runai", nil)
		b := cwRct("r", "runai", nil)
		b.Spec.Spec.Devices.Requests[0].Exactly.Count = 4
		Expect(gvkCommon.AreEqual(context.Background(), a, b)).To(BeFalse())
	})
})

var _ = Describe("ShouldPatchAnnotations", func() {
	It("does not patch when clearing an annotation that is already cleared", func() {
		current := map[string]string{"k2": "v2"}
		patch := map[string]interface{}{"k1": nil}
		Expect(ShouldPatchAnnotations(current, patch)).To(BeFalse())
		Expect(ShouldPatchAnnotations(nil, patch)).To(BeFalse())
	})

	It("patches when there are no current annotations", func() {
		patch := map[string]interface{}{"k1": "v1"}
		Expect(ShouldPatchAnnotations(nil, patch)).To(BeTrue())
	})

	It("patches when clearing an annotation that is currently set", func() {
		current := map[string]string{"k1": "v1"}
		patch := map[string]interface{}{"k1": nil}
		Expect(ShouldPatchAnnotations(current, patch)).To(BeTrue())
	})

	It("does not patch when the value is unchanged", func() {
		current := map[string]string{"k1": "v1"}
		patch := map[string]interface{}{"k1": "v1"}
		Expect(ShouldPatchAnnotations(current, patch)).To(BeFalse())
	})

	It("patches when the value differs", func() {
		current := map[string]string{"k1": "v1"}
		patch := map[string]interface{}{"k1": "v2"}
		Expect(ShouldPatchAnnotations(current, patch)).To(BeTrue())
	})
})
