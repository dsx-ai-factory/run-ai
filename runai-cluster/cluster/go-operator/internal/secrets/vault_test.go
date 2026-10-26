/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package secrets_test

import (
	"context"
	"encoding/json"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/secrets"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	agentDeploymentName           = "agent"
	installationConfigurationData = "global:\n" +
		"  clusterURL: https://cluster.example.com\n" +
		"  controlPlane:\n" +
		"    clusterUUID: 00000000-0000-0000-0000-000000000001\n" +
		"    config:\n" +
		"      agent:\n" +
		"        client:\n" +
		"          id: 00000000-0000-0000-0000-000000000002\n" +
		"          name: runai-internal.00000000-0000-0000-0000-000000000001.agent\n" +
		"          secret: test-secret\n" +
		"      clusterName: example-cluster"
)

var _ = Describe("Secret Handling", func() {
	var (
		ctx       context.Context
		k8sClient client.Client
	)

	BeforeEach(func() {
		ctx = context.Background()
		k8sClient = fake.NewFakeClient()
	})

	Context("Vault", Ordered, func() {
		It("create vault to empty secret and nil secret deployment", func() {
			secret, err := secrets.GetConfigurationSecret(ctx, k8sClient)
			Expect(err).ToNot(HaveOccurred())
			Expect(secret).ToNot(BeNil())

			_, err = secrets.NewVault(agentDeploymentName, secret, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("deployment secret is nil"))
		})

		It("create vault to empty secret", func() {
			secret, err := secrets.GetConfigurationSecret(ctx, k8sClient)
			Expect(err).ToNot(HaveOccurred())
			Expect(secret).ToNot(BeNil())

			_, err = secrets.NewVault(agentDeploymentName, secret, &v1.Secret{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("secret data is empty"))
		})

		It("create vault when secret data is nil", func() {
			emptySecret := createInstallationConfigSecret()
			err := k8sClient.Create(ctx, emptySecret)
			Expect(err).ToNot(HaveOccurred())

			secret, err := secrets.GetConfigurationSecret(ctx, k8sClient)
			Expect(err).ToNot(HaveOccurred())
			Expect(secret).ToNot(BeNil())

			_, err = secrets.NewVault(agentDeploymentName, secret, &v1.Secret{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("secret data is empty"))
		})

		It("create vault when secret data missing 'installation-configuration", func() {
			ctx := context.Background()
			icSecret := createInstallationConfigSecret()
			icSecret.Data = make(map[string][]byte)
			icSecret.Data["bla"] = []byte("bla1")
			err := k8sClient.Create(ctx, icSecret)
			Expect(err).ToNot(HaveOccurred())

			secret, err := secrets.GetConfigurationSecret(ctx, k8sClient)
			Expect(err).ToNot(HaveOccurred())
			Expect(secret).ToNot(BeNil())

			_, err = secrets.NewVault(agentDeploymentName, secret, &v1.Secret{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("data does not contain 'installation-configuration'"))
		})

		It("create vault when secret data exist 'installation-configuration' does have valid values", func() {
			ctx := context.Background()
			icSecret := createInstallationConfigSecret()
			icSecret.Data = make(map[string][]byte)
			icSecret.Data[secrets.InstallationConfigurationSecretName] = []byte("bla1")
			err := k8sClient.Create(ctx, icSecret)
			Expect(err).ToNot(HaveOccurred())

			secret, err := secrets.GetConfigurationSecret(ctx, k8sClient)
			Expect(err).ToNot(HaveOccurred())
			Expect(secret).ToNot(BeNil())

			_, err = secrets.NewVault(agentDeploymentName, secret, &v1.Secret{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("bla1"))
			Expect(err.Error()).To(ContainSubstring("cannot unmarshal"))
		})

		It("vault create a secret with values", func() {
			ctx := context.Background()
			icSecret := createInstallationConfigSecret()
			icSecret.Data = make(map[string][]byte)
			icSecret.Data[secrets.InstallationConfigurationSecretName] = []byte(installationConfigurationData)
			err := k8sClient.Create(ctx, icSecret)
			Expect(err).ToNot(HaveOccurred())

			secret, err := secrets.GetConfigurationSecret(ctx, k8sClient)
			Expect(err).ToNot(HaveOccurred())
			Expect(secret).ToNot(BeNil())

			vault, err := secrets.NewVault("agent", secret, &v1.Secret{})
			Expect(err).ToNot(HaveOccurred())

			vault.AddFromSecret(secrets.ClientName, "CLIENT_NAME")
			secret, err = vault.Generate()
			Expect(err).ToNot(HaveOccurred())
			Expect(secret).ToNot(BeNil())
			Expect(secret.Data).To(HaveLen(1))
			extData := map[string]string{}
			err = json.Unmarshal(secret.Data["agent"], &extData)
			Expect(err).ToNot(HaveOccurred())
			Expect(extData).To(HaveLen(1))
			Expect(extData["CLIENT_NAME"]).To(Equal("runai-internal.00000000-0000-0000-0000-000000000001.agent"))
		})

		It("vault create a secret and proprietary values", func() {
			ctx := context.Background()
			icSecret := createInstallationConfigSecret()
			icSecret.Data = make(map[string][]byte)
			icSecret.Data[secrets.InstallationConfigurationSecretName] = []byte(installationConfigurationData)
			err := k8sClient.Create(ctx, icSecret)
			Expect(err).ToNot(HaveOccurred())

			secret, err := secrets.GetConfigurationSecret(ctx, k8sClient)
			Expect(err).ToNot(HaveOccurred())
			Expect(secret).ToNot(BeNil())

			vault, err := secrets.NewVault("agent", secret, &v1.Secret{})
			Expect(err).ToNot(HaveOccurred())

			vault.AddFromSecret(secrets.ClientName, "CLIENT_NAME")
			vault.AddFromSecretByPath("global.controlPlane.clusterUUID", "CLUSTER_UUID")
			vault.AddKeyValueSecret("ANOTHER_SECRET", "justfortesting")
			secret, err = vault.Generate()
			Expect(err).ToNot(HaveOccurred())
			Expect(secret).ToNot(BeNil())
			Expect(secret.Data).To(HaveLen(1))
			extData := map[string]string{}
			err = json.Unmarshal(secret.Data["agent"], &extData)
			Expect(err).ToNot(HaveOccurred())
			Expect(extData).To(HaveLen(3))
			Expect(extData["CLIENT_NAME"]).To(Equal("runai-internal.00000000-0000-0000-0000-000000000001.agent"))
			Expect(extData["CLUSTER_UUID"]).To(Equal("00000000-0000-0000-0000-000000000001"))
			Expect(extData["ANOTHER_SECRET"]).To(Equal("justfortesting"))
		})
	})
})

func createInstallationConfigSecret() *v1.Secret {
	return &v1.Secret{ObjectMeta: metav1.ObjectMeta{Name: secrets.InstallationConfigurationSecretName, Namespace: runaiv1.DefaultNamespace}}
}
