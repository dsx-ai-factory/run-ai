/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package project

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	schedulingv2alpha2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	corev1 "k8s.io/api/core/v1"
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
