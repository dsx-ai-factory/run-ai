/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package project

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"

	runaiv2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeClient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	projectName = "project-name"
)

func init() {
	_ = runaiv2.AddToScheme(scheme.Scheme)
}

func TestJobSuspend(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "namespace project tests")
}

var _ = Describe("namespace project tests", func() {
	var (
		k8sClient client.WithWatch
	)
	Context("namespace project", func() {
		Describe("GetProject", func() {
			BeforeEach(func() {
				projectObj := &runaiv2.Project{
					ObjectMeta: metav1.ObjectMeta{
						Name: projectName,
					},
				}

				objects := []runtime.Object{projectObj}
				k8sClient = fakeClient.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(objects...).Build()
			})
			It("GetProject - Succeed", func() {
				project, err := GetProject(
					context.TODO(), k8sClient, projectName)
				Expect(err).To(BeNil())
				Expect(project).To(Not(BeNil()))
			})
			It("GetProject - Fail non existing", func() {
				project, err := GetProject(
					context.TODO(), k8sClient, "some-project")
				Expect(err).To(Not(BeNil()))
				Expect(project).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("not found"))
			})
		})
	})
})
