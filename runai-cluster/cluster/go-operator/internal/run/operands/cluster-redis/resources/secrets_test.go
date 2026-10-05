/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/scheme-init"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Secret", func() {
	const operandName = "cluster-redis"

	It("should generate matching Redis credentials when no password exists", func() {
		runaiConfig := utils.GetDefaultRunaiConfig()
		fakeClient := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).Build()
		resource := NewSecretResource(operandName, fakeClient, runaiConfig)

		obj, err := resource.Object(context.Background())
		Expect(err).NotTo(HaveOccurred())
		secret := obj.(*v1.Secret)
		password := string(secret.Data[constants.WorkloadAuthRedisPasswordKey])

		Expect(password).NotTo(BeEmpty())
		decoded, err := base64.URLEncoding.DecodeString(password)
		Expect(err).NotTo(HaveOccurred())
		Expect(decoded).To(HaveLen(32))
		Expect(secret.Data[valkeyConfFileName]).To(Equal([]byte(fmt.Sprintf(`requirepass "%s"`, password))))
	})

	It("should migrate the existing workloads-auth password", func() {
		runaiConfig := utils.GetDefaultRunaiConfig()
		workloadsAuthSecret := newSecret(constants.WorkloadAuthSecretName, runaiConfig.GetNamespace(), map[string][]byte{
			constants.WorkloadAuthRedisPasswordKey: []byte("existing-password"),
		})
		fakeClient := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).WithObjects(workloadsAuthSecret).Build()
		resource := NewSecretResource(operandName, fakeClient, runaiConfig)

		obj, err := resource.Object(context.Background())
		Expect(err).NotTo(HaveOccurred())
		secret := obj.(*v1.Secret)

		Expect(secret.Data[constants.WorkloadAuthRedisPasswordKey]).To(Equal([]byte("existing-password")))
		Expect(secret.Data[valkeyConfFileName]).To(Equal([]byte(`requirepass "existing-password"`)))
	})

	It("should preserve its existing password", func() {
		runaiConfig := utils.GetDefaultRunaiConfig()
		redisSecret := newSecret(DeploymentName, runaiConfig.GetNamespace(), map[string][]byte{
			constants.WorkloadAuthRedisPasswordKey: []byte("redis-password"),
		})
		workloadsAuthSecret := newSecret(constants.WorkloadAuthSecretName, runaiConfig.GetNamespace(), map[string][]byte{
			constants.WorkloadAuthRedisPasswordKey: []byte("workloads-password"),
		})
		fakeClient := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).
			WithObjects(redisSecret, workloadsAuthSecret).
			Build()
		resource := NewSecretResource(operandName, fakeClient, runaiConfig)

		obj, err := resource.Object(context.Background())
		Expect(err).NotTo(HaveOccurred())
		secret := obj.(*v1.Secret)

		Expect(secret.Data[constants.WorkloadAuthRedisPasswordKey]).To(Equal([]byte("redis-password")))
		Expect(secret.Data[valkeyConfFileName]).To(Equal([]byte(`requirepass "redis-password"`)))
	})

	It("should preserve a generated password across reconciliations", func() {
		runaiConfig := utils.GetDefaultRunaiConfig()
		fakeClient := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).Build()
		resource := NewSecretResource(operandName, fakeClient, runaiConfig)

		firstObj, err := resource.Object(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(fakeClient.Create(context.Background(), firstObj)).To(Succeed())

		secondObj, err := resource.Object(context.Background())
		Expect(err).NotTo(HaveOccurred())

		firstSecret := firstObj.(*v1.Secret)
		secondSecret := secondObj.(*v1.Secret)
		Expect(secondSecret.Data[constants.WorkloadAuthRedisPasswordKey]).
			To(Equal(firstSecret.Data[constants.WorkloadAuthRedisPasswordKey]))
	})
})

func newSecret(name, namespace string, data map[string][]byte) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}
}
