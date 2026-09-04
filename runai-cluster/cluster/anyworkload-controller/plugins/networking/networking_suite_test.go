/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"os"
	"testing"

	workloadbuilders "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/builders"
	workloadconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNetworking(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Networking Plugin Suite")
}

var _ = BeforeSuite(func() {
	_ = os.Setenv(workloadbuilders.ExternalUrlAuthContainerImageEnv, "quay.io/oauth2-proxy/oauth2-proxy:v7.6.0")
	_ = os.Setenv(workloadconfig.EnvVarOauth2ProxyOidcIssuerUrl, "https://auth.example.com")
	_ = os.Setenv(workloadconfig.EnvVarControlPlaneEndpoint, "https://cp.example.com")
	_ = os.Setenv(workloadbuilders.WorkloadAuthSecretNameEnv, "oauth2-proxy-secret")
	_ = os.Setenv(workloadconfig.EnvVarOauth2ProxyUsernameLoaderImage, "busybox:latest")
	_ = os.Setenv(workloadconfig.EnvVarClusterDomain, "https://cluster-domain.mock")
	_ = os.Setenv(workloadconfig.EnvVarIngressClass, "nginx")
})

var _ = AfterSuite(func() {
	_ = os.Unsetenv(workloadbuilders.ExternalUrlAuthContainerImageEnv)
	_ = os.Unsetenv(workloadconfig.EnvVarOauth2ProxyOidcIssuerUrl)
	_ = os.Unsetenv(workloadconfig.EnvVarControlPlaneEndpoint)
	_ = os.Unsetenv(workloadbuilders.WorkloadAuthSecretNameEnv)
	_ = os.Unsetenv(workloadconfig.EnvVarOauth2ProxyUsernameLoaderImage)
	_ = os.Unsetenv(workloadconfig.EnvVarClusterDomain)
	_ = os.Unsetenv(workloadconfig.EnvVarIngressClass)
})
