/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/utils"
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Secret", func() {
	const operandName = "anyworkload-controller"

	It("should wait until the cluster-redis secret exists", func() {
		runaiConfig := utils.GetDefaultRunaiConfig()
		existingSecret := newWorkloadsAuthSecret(runaiConfig.GetNamespace(), nil)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).WithObjects(existingSecret).Build()
		resource := NewSecretResource(operandName, fakeClient, runaiConfig)

		_, err := resource.Object(context.Background())
		Expect(err).To(MatchError("redis secret /cluster-redis is not ready"))

		redisSecret := newClusterRedisSecret(runaiConfig.GetNamespace(), map[string][]byte{
			constants.WorkloadAuthRedisPasswordKey: []byte("redis-password"),
		})
		Expect(fakeClient.Create(context.Background(), redisSecret)).To(Succeed())

		obj, err := resource.Object(context.Background())
		Expect(err).NotTo(HaveOccurred())
		secret := obj.(*v1.Secret)
		Expect(secret.Data[constants.WorkloadAuthRedisPasswordKey]).To(Equal([]byte("redis-password")))
	})

	It("should wait until the cluster-redis secret contains a password", func() {
		runaiConfig := utils.GetDefaultRunaiConfig()
		workloadsAuthSecret := newWorkloadsAuthSecret(runaiConfig.GetNamespace(), nil)
		redisSecret := newClusterRedisSecret(runaiConfig.GetNamespace(), nil)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).
			WithObjects(workloadsAuthSecret, redisSecret).
			Build()
		resource := NewSecretResource(operandName, fakeClient, runaiConfig)

		_, err := resource.Object(context.Background())
		Expect(err).To(MatchError("redis secret /cluster-redis does not contain OAUTH2_PROXY_REDIS_PASSWORD"))
	})

	It("should copy the password from the cluster-redis secret", func() {
		runaiConfig := utils.GetDefaultRunaiConfig()
		workloadsAuthSecret := newWorkloadsAuthSecret(runaiConfig.GetNamespace(), map[string][]byte{
			constants.WorkloadAuthRedisPasswordKey: []byte("old-password"),
		})
		redisSecret := newClusterRedisSecret(runaiConfig.GetNamespace(), map[string][]byte{
			constants.WorkloadAuthRedisPasswordKey: []byte("redis-password"),
		})
		fakeClient := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).
			WithObjects(workloadsAuthSecret, redisSecret).
			Build()
		resource := NewSecretResource(operandName, fakeClient, runaiConfig)

		obj, err := resource.Object(context.Background())
		Expect(err).NotTo(HaveOccurred())
		secret := obj.(*v1.Secret)
		Expect(secret.Data[constants.WorkloadAuthRedisPasswordKey]).To(Equal([]byte("redis-password")))
	})

	DescribeTable("should report not ready when the live secret is missing a required key", func(missingKey string) {
		runaiConfig := utils.GetDefaultRunaiConfig()
		workloadsAuthSecret := newWorkloadsAuthSecret(runaiConfig.GetNamespace(), map[string][]byte{
			oAuthClientIDKey:           []byte("client-id"),
			oAuthProxyCookieSecretKey:  []byte("cookie-secret"),
			oAuthProxyRedisPasswordKey: []byte("redis-password"),
		})
		delete(workloadsAuthSecret.Data, missingKey)
		workloadsAuthSecret.UID = "existing-uid"
		redisSecret := newClusterRedisSecret(runaiConfig.GetNamespace(), map[string][]byte{
			constants.WorkloadAuthRedisPasswordKey: []byte("redis-password"),
		})
		fakeClient := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).
			WithObjects(workloadsAuthSecret, redisSecret).
			Build()
		resource := NewSecretResource(operandName, fakeClient, runaiConfig)

		Expect(resource.Status(context.Background())).To(MatchError("secret [/workloads-auth] does not contain " + missingKey))
	},
		Entry("client ID", oAuthClientIDKey),
		Entry("client secret", oAuthClientSecretKey),
		Entry("cookie secret", oAuthProxyCookieSecretKey),
		Entry("Redis password", oAuthProxyRedisPasswordKey),
	)

	It("should report ready when the live secret contains every required key", func() {
		runaiConfig := utils.GetDefaultRunaiConfig()
		workloadsAuthSecret := newWorkloadsAuthSecret(runaiConfig.GetNamespace(), map[string][]byte{
			oAuthClientIDKey:           []byte("client-id"),
			oAuthProxyCookieSecretKey:  []byte("cookie-secret"),
			oAuthProxyRedisPasswordKey: []byte("redis-password"),
		})
		fakeClient := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).WithObjects(workloadsAuthSecret).Build()
		resource := NewSecretResource(operandName, fakeClient, runaiConfig)

		Expect(resource.Status(context.Background())).To(Succeed())
	})
})

func newWorkloadsAuthSecret(namespace string, data map[string][]byte) *v1.Secret {
	if data == nil {
		data = map[string][]byte{}
	}
	data[oAuthClientSecretKey] = []byte("client-secret")
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.WorkloadAuthSecretName,
			Namespace: namespace,
		},
		Data: data,
	}
}

func newClusterRedisSecret(namespace string, data map[string][]byte) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterRedisSecretName,
			Namespace: namespace,
		},
		Data: data,
	}
}
