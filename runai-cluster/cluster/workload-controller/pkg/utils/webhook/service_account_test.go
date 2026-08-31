/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package webhook

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeClient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	projectName        = "project-name"
	serviceAccountName = "service-account"
)

var (
	namespaceName = fmt.Sprintf("runai-%s", projectName)
)

var _ = Describe("ServiceAccount", func() {
	var (
		k8sClient client.WithWatch
	)

	Context("k8s utils", func() {
		Describe("GetServiceAccount", func() {
			BeforeEach(func() {
				namespace := &v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: namespaceName,
						Labels: map[string]string{
							"runai/queue": projectName,
						},
					},
				}
				serviceAccountObj := &v1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      serviceAccountName,
						Namespace: namespaceName,
					},
				}

				objects := []runtime.Object{namespace, serviceAccountObj}
				k8sClient = fakeClient.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(objects...).Build()
			})
			It("GetServiceAccount - Succeed", func() {
				serviceAccount, err := GetServiceAccount(
					context.TODO(), k8sClient,
					serviceAccountName, namespaceName)
				Expect(err).To(BeNil())
				Expect(serviceAccount).To(Not(BeNil()))
			})
			It("GetServiceAccount - Fail non existing", func() {
				serviceAccount, err := GetServiceAccount(
					context.TODO(), k8sClient,
					"some-service-account", namespaceName)
				Expect(err).To(Not(BeNil()))
				Expect(serviceAccount).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("not found"))
			})
		})
	})
})
