/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package externalworkloads

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("ExternalWorkloadHandler.getTopResourceOwnerWithHandler", func() {
	const (
		namespace      = "test-ns"
		deploymentName = "inference-epp"
	)

	ownerRef := func(apiVersion, kind, name string, uid types.UID) metav1.OwnerReference {
		return metav1.OwnerReference{APIVersion: apiVersion, Kind: kind, Name: name, UID: uid}
	}
	endpointRef := ownerRef("inference.platform.run.ai/v1alpha1", "InferenceEndpoint", "inference", "endpoint-uid")

	newHandler := func(deploymentOwners []metav1.OwnerReference) (*ExternalWorkloadHandler, *corev1.Pod) {
		scheme := runtime.NewScheme()
		Expect(appsv1.AddToScheme(scheme)).To(Succeed())
		Expect(corev1.AddToScheme(scheme)).To(Succeed())

		deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName, Namespace: namespace, UID: "deployment-uid", OwnerReferences: deploymentOwners,
		}}
		replicaSet := &appsv1.ReplicaSet{ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName + "-rs", Namespace: namespace, UID: "rs-uid",
			OwnerReferences: []metav1.OwnerReference{ownerRef("apps/v1", "Deployment", deploymentName, "deployment-uid")},
		}}
		pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName + "-pod", Namespace: namespace, UID: "pod-uid",
			OwnerReferences: []metav1.OwnerReference{ownerRef("apps/v1", "ReplicaSet", replicaSet.Name, "rs-uid")},
		}}

		c := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(deployment, replicaSet).Build()
		return &ExternalWorkloadHandler{cachedClient: c}, pod
	}

	It("drops a Deployment owned by an InferenceEndpoint without reading the endpoint", func() {
		handler, _ := newHandler([]metav1.OwnerReference{endpointRef})
		deployment := &appsv1.Deployment{}
		Expect(handler.cachedClient.Get(context.Background(),
			client.ObjectKey{Namespace: namespace, Name: deploymentName}, deployment)).To(Succeed())

		_, _, err := handler.getTopResourceOwnerWithHandler(context.Background(), deployment)

		Expect(isOwnedByDroppedKind(err)).To(BeTrue(), "got: %v", err)
	})

	It("drops a pod whose Deployment is owned by an InferenceEndpoint", func() {
		handler, pod := newHandler([]metav1.OwnerReference{endpointRef})

		_, _, err := handler.getTopResourceOwnerWithHandler(context.Background(), pod)

		Expect(isOwnedByDroppedKind(err)).To(BeTrue(), "got: %v", err)
	})

	It("keeps resolving the top owner of a pod under an ordinary Deployment", func() {
		handler, pod := newHandler(nil)

		topOwner, _, err := handler.getTopResourceOwnerWithHandler(context.Background(), pod)

		Expect(err).ToNot(HaveOccurred())
		Expect(topOwner.GetName()).To(Equal(deploymentName))
	})
})
