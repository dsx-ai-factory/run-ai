/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package secrets_test

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/secrets"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	DeploymentName       = "agent"
	DeploymentSecretName = "whisper-agent-secret"
	DeploymentNS         = "any"
)

var _ = Describe("Secret Mounting", func() {
	var (
		ctx       context.Context
		k8sClient client.Client
	)

	BeforeEach(func() {
		ctx = context.Background()
		k8sClient = fake.NewFakeClient()
	})
	Context("Mount", Ordered, func() {
		It("no secret", func() {
			d := createDeployment()
			err := secrets.MountToDeployment(ctx, k8sClient, d)
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Spec.Template.Spec.Volumes).To(HaveLen(0))
			Expect(d.Spec.Template.Spec.Containers).To(HaveLen(0))
		})

		It("has secret but without data", func() {
			d := createDeployment()
			secret := createSecret()
			err := k8sClient.Create(ctx, secret)
			Expect(err).ToNot(HaveOccurred())

			err = secrets.MountToDeployment(ctx, k8sClient, d)
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Spec.Template.Spec.Volumes).To(HaveLen(0))
			Expect(d.Spec.Template.Spec.Containers).To(HaveLen(0))
		})

		It("has secret but without data", func() {
			d := createDeployment()
			secret := createSecret()
			err := k8sClient.Create(ctx, secret)
			Expect(err).ToNot(HaveOccurred())

			err = secrets.MountToDeployment(ctx, k8sClient, d)
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Spec.Template.Spec.Volumes).To(HaveLen(0))
			Expect(d.Spec.Template.Spec.Containers).To(HaveLen(0))
		})

		It("deployment without containers - volume added", func() {
			d := createDeployment()
			secret := createSecret()
			secret.Data = make(map[string][]byte)
			secret.Data[DeploymentName] = []byte("bla")
			err := k8sClient.Create(ctx, secret)
			Expect(err).ToNot(HaveOccurred())

			err = secrets.MountToDeployment(ctx, k8sClient, d)
			Expect(err).ToNot(HaveOccurred())
			volumes := d.Spec.Template.Spec.Volumes
			Expect(volumes).To(HaveLen(1))
			Expect(d.Spec.Template.Spec.Containers).To(HaveLen(0))
			Expect(volumes[0].Secret.Items).To(HaveLen(1))
		})

		It("deployment with containers - volume added and mounted", func() {
			d := createDeployment()
			containers := &d.Spec.Template.Spec.Containers
			*containers = make([]corev1.Container, 0)
			*containers = append(*containers, corev1.Container{
				Name:  "cont1",
				Image: "image1",
			})

			secret := createSecret()
			secret.Data = make(map[string][]byte)
			secret.Data[DeploymentName] = []byte("bla")
			err := k8sClient.Create(ctx, secret)
			Expect(err).ToNot(HaveOccurred())

			err = secrets.MountToDeployment(ctx, k8sClient, d)
			Expect(err).ToNot(HaveOccurred())
			volumes := d.Spec.Template.Spec.Volumes
			Expect(volumes).To(HaveLen(1))
			Expect(*containers).To(HaveLen(1))
			vs := volumes[0].Secret
			Expect(vs.SecretName).To(Equal(DeploymentSecretName))

			Expect((*containers)[0].VolumeMounts).To(HaveLen(1))
			vm := (*containers)[0].VolumeMounts
			Expect(vm[0].Name).To(Equal(DeploymentSecretName))
		})
	})
})

func createDeployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: DeploymentName, Namespace: DeploymentNS},
	}
}

func createSecret() *corev1.Secret {
	return &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf(secrets.SecretName, "agent"), Namespace: runaiv1.DefaultNamespace}}
}
