/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validator

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var _ = Describe("Project garbage collector authorization", func() {
	var (
		k8sClient              client.Client
		handler                *ProjectValidator
		req                    admission.Request
		oldProject, newProject *unstructured.Unstructured
	)
	BeforeEach(func() {
		objects := []client.Object{&corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{
			Name: "generic-garbage-collector", Namespace: "kube-system", UID: "gc-uid",
		}}}
		for _, name := range []string{operands.RunaiAgentName, operands.RunaiClusterSyncName,
			operands.RunaiOperatorName, operands.KRMProjectControllerName} {
			objects = append(objects, &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{
				Name: name, Namespace: "runai", UID: types.UID(name + "-uid"),
			}})
		}
		k8sClient = fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(objects...).Build()
		handler = NewProjectValidator(k8sClient, nil)
		oldProject = &unstructured.Unstructured{Object: map[string]interface{}{
			"apiVersion": "kai.resources/v1", "kind": "Project",
			"metadata": map[string]interface{}{"name": "project", "uid": "project-uid", "resourceVersion": "1"},
			"spec":     map[string]interface{}{"description": "original"},
		}}
		now := metav1.NewTime(time.Now())
		oldProject.SetDeletionTimestamp(&now)
		oldProject.SetFinalizers([]string{metav1.FinalizerOrphanDependents})
		newProject = oldProject.DeepCopy()
		newProject.SetFinalizers(nil)
		req = admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{
			Operation: admissionv1.Update,
			Resource:  metav1.GroupVersionResource{Group: "kai.resources", Version: "v1", Resource: "projects"},
			UserInfo: authenticationv1.UserInfo{
				Username: "system:serviceaccount:kube-system:generic-garbage-collector", UID: "gc-uid",
			},
		}}
	})
	handle := func() admission.Response {
		var err error
		req.OldObject.Raw, err = json.Marshal(oldProject)
		Expect(err).NotTo(HaveOccurred())
		req.Object.Raw, err = json.Marshal(newProject)
		Expect(err).NotTo(HaveOccurred())
		return handler.Handle(context.Background(), req)
	}
	It("allows the garbage collector to remove the orphan finalizer from a deleting Project", func() {
		Expect(handle().Allowed).To(BeTrue())
	})
	It("allows legacy run.ai Projects", func() {
		req.Resource.Group = "run.ai"
		oldProject.SetAPIVersion("run.ai/v1")
		newProject.SetAPIVersion("run.ai/v1")
		Expect(handle().Allowed).To(BeTrue())
	})
	It("preserves other finalizers", func() {
		oldProject.SetFinalizers([]string{"example.com/cleanup", metav1.FinalizerOrphanDependents})
		newProject.SetFinalizers([]string{"example.com/cleanup"})
		Expect(handle().Allowed).To(BeTrue())
	})
	It("allows API server managedFields bookkeeping", func() {
		newProject.SetManagedFields([]metav1.ManagedFieldsEntry{{Manager: "kube-controller-manager"}})
		Expect(handle().Allowed).To(BeTrue())
	})
	DescribeTable("rejects requests outside orphan cleanup", func(change func()) {
		change()
		Expect(handle().Allowed).To(BeFalse())
	},
		Entry("tenant service account with the same name", func() { req.UserInfo.Username = "system:serviceaccount:tenant:generic-garbage-collector" }),
		Entry("other kube-system service account", func() { req.UserInfo.Username = "system:serviceaccount:kube-system:coredns" }),
		Entry("mismatched service account UID", func() { req.UserInfo.UID = "other-uid" }),
		Entry("empty service account UID", func() { req.UserInfo.UID = "" }),
		Entry("CREATE", func() { req.Operation = admissionv1.Create }),
		Entry("DELETE", func() { req.Operation = admissionv1.Delete }),
		Entry("status update", func() { req.SubResource = "status" }),
		Entry("other resource", func() { req.Resource.Resource = "departments" }),
		Entry("other API group", func() { req.Resource.Group = "example.com" }),
		Entry("Project not deleting", func() { oldProject.SetDeletionTimestamp(nil); newProject.SetDeletionTimestamp(nil) }),
		Entry("orphan finalizer retained", func() { newProject.SetFinalizers(oldProject.GetFinalizers()) }),
		Entry("different finalizer removed", func() { oldProject.SetFinalizers([]string{"example.com/cleanup"}) }),
		Entry("additional finalizer removed", func() { oldProject.SetFinalizers([]string{metav1.FinalizerOrphanDependents, "example.com/cleanup"}) }),
		Entry("spec change", func() { newProject.Object["spec"] = map[string]interface{}{"description": "modified"} }),
		Entry("unknown field change", func() { newProject.Object["futureField"] = "modified" }),
		Entry("label change", func() { newProject.SetLabels(map[string]string{"modified": "true"}) }),
		Entry("owner reference change", func() { newProject.SetOwnerReferences([]metav1.OwnerReference{{Name: "other"}}) }),
		Entry("deletion timestamp removed", func() { newProject.SetDeletionTimestamp(nil) }),
	)
	It("denies when the garbage collector service account is missing", func() {
		Expect(k8sClient.Delete(context.Background(), &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{
			Name: "generic-garbage-collector", Namespace: "kube-system",
		}})).To(Succeed())
		Expect(handle().Allowed).To(BeFalse())
	})
	It("denies malformed objects", func() {
		req.OldObject = runtime.RawExtension{Raw: []byte(`{"metadata":`)}
		req.Object = runtime.RawExtension{Raw: []byte(`{}`)}
		Expect(handler.Handle(context.Background(), req).Allowed).To(BeFalse())
	})
})
