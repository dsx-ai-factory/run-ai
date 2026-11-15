/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services"
)

var _ = Describe("ClusterSnapshot", func() {
	Describe("NewClusterSnapshotFromCluster", func() {
		It("returns an empty snapshot when the cluster is nil", func() {
			snapshot := services.NewClusterSnapshotFromCluster(nil)

			Expect(snapshot).NotTo(BeNil())
			Expect(snapshot.GitSyncImage).To(BeNil())
			Expect(snapshot.CustomCAGit).To(BeNil())
		})

		It("returns an empty snapshot when the cluster has no reported status", func() {
			snapshot := services.NewClusterSnapshotFromCluster(&api.DisplayedCluster{})

			Expect(snapshot).NotTo(BeNil())
			Expect(snapshot.GitSyncImage).To(BeNil())
			Expect(snapshot.ExternalAuthUrlEnabled).To(BeNil())
			Expect(snapshot.CustomCAOauth2Proxy).To(BeNil())
		})

		It("returns an empty snapshot when status carries no reported config", func() {
			cluster := &api.DisplayedCluster{
				Status: &api.ClusterDisplayedStatus{},
			}

			snapshot := services.NewClusterSnapshotFromCluster(cluster)

			Expect(snapshot).NotTo(BeNil())
			Expect(snapshot.RunaiOauth2ProxyRedisFQDN).To(BeNil())
			Expect(snapshot.CustomCAS3).To(BeNil())
		})

		It("composes cluster-reported values when populated", func() {
			terminationPolicy := api.OPENSHIFTROUTETERMINATIONPOLICY_PASSTHROUGH
			cluster := &api.DisplayedCluster{
				Status: &api.ClusterDisplayedStatus{
					Config: &api.ClusterReportedConfig{
						GitSyncImage:                           ptr.To("reg.example.com/git-sync:v1"),
						S3FileSystemImage:                      ptr.To("reg.example.com/goofys:v2"),
						S3FileSystemMountCheckImage:            ptr.To("reg.example.com/goofys:v2"),
						RunaiOauth2ProxyImage:                  ptr.To("reg.example.com/oauth2-proxy:v3"),
						RunaiOauth2ProxyUsernameLoaderImage:    ptr.To("reg.example.com/username-loader:v4"),
						RunaiOauth2ProxyRedisFQDN:              ptr.To("redis.example.com:6379"),
						RunaiOauth2ProxyWorkloadAuthSecretName: ptr.To("workload-auth-secret"),
						RunaiOauth2ProxySessionStoreType:       ptr.To("cookie"),
						ExternalAuthUrlEnabled:                 ptr.To(false),
						OpenshiftRouteTerminationPolicy:        &terminationPolicy,
						CustomCAGit: &api.ClusterReportedCustomCa{
							Enabled:    ptr.To(true),
							SecretName: ptr.To("ca-git"),
							SecretKey:  ptr.To("ca.crt"),
							Optional:   ptr.To(false),
						},
						CustomCAS3: &api.ClusterReportedCustomCa{
							Enabled:    ptr.To(true),
							SecretName: ptr.To("ca-s3"),
							SecretKey:  ptr.To("ca.crt"),
							Optional:   ptr.To(true),
						},
						CustomCAOauth2Proxy: &api.ClusterReportedCustomCa{
							Enabled: ptr.To(false),
						},
					},
				},
			}

			snapshot := services.NewClusterSnapshotFromCluster(cluster)

			Expect(snapshot).NotTo(BeNil())
			Expect(*snapshot.GitSyncImage).To(Equal("reg.example.com/git-sync:v1"))
			Expect(*snapshot.S3FileSystemImage).To(Equal("reg.example.com/goofys:v2"))
			Expect(*snapshot.S3FileSystemMountCheckImage).To(Equal("reg.example.com/goofys:v2"))
			Expect(*snapshot.RunaiOauth2ProxyImage).To(Equal("reg.example.com/oauth2-proxy:v3"))
			Expect(*snapshot.RunaiOauth2ProxyUsernameLoaderImage).To(Equal("reg.example.com/username-loader:v4"))
			Expect(*snapshot.RunaiOauth2ProxyRedisFQDN).To(Equal("redis.example.com:6379"))
			Expect(*snapshot.RunaiOauth2ProxyWorkloadAuthSecretName).To(Equal("workload-auth-secret"))
			Expect(*snapshot.RunaiOauth2ProxySessionStoreType).To(Equal("cookie"))
			Expect(*snapshot.ExternalAuthUrlEnabled).To(BeFalse())
			Expect(*snapshot.OpenshiftRouteTerminationPolicy).To(Equal("passthrough"))

			Expect(snapshot.CustomCAGit).NotTo(BeNil())
			Expect(*snapshot.CustomCAGit.Enabled).To(BeTrue())
			Expect(*snapshot.CustomCAGit.SecretName).To(Equal("ca-git"))
			Expect(*snapshot.CustomCAGit.SecretKey).To(Equal("ca.crt"))
			Expect(*snapshot.CustomCAGit.Optional).To(BeFalse())

			Expect(snapshot.CustomCAS3).NotTo(BeNil())
			Expect(*snapshot.CustomCAS3.Enabled).To(BeTrue())
			Expect(*snapshot.CustomCAS3.SecretName).To(Equal("ca-s3"))
			Expect(*snapshot.CustomCAS3.SecretKey).To(Equal("ca.crt"))
			Expect(*snapshot.CustomCAS3.Optional).To(BeTrue())

			Expect(snapshot.CustomCAOauth2Proxy).NotTo(BeNil())
			Expect(*snapshot.CustomCAOauth2Proxy.Enabled).To(BeFalse())
			Expect(snapshot.CustomCAOauth2Proxy.SecretName).To(BeNil())
			Expect(snapshot.CustomCAOauth2Proxy.SecretKey).To(BeNil())
			Expect(snapshot.CustomCAOauth2Proxy.Optional).To(BeNil())
		})

		It("preserves nil fields for unreported values", func() {
			cluster := &api.DisplayedCluster{
				Status: &api.ClusterDisplayedStatus{
					Config: &api.ClusterReportedConfig{
						GitSyncImage: ptr.To("reg.example.com/git-sync:v1"),
					},
				},
			}

			snapshot := services.NewClusterSnapshotFromCluster(cluster)

			Expect(snapshot).NotTo(BeNil())
			Expect(*snapshot.GitSyncImage).To(Equal("reg.example.com/git-sync:v1"))
			Expect(snapshot.S3FileSystemImage).To(BeNil())
			Expect(snapshot.S3FileSystemMountCheckImage).To(BeNil())
			Expect(snapshot.RunaiOauth2ProxyImage).To(BeNil())
			Expect(snapshot.ExternalAuthUrlEnabled).To(BeNil())
			Expect(snapshot.OpenshiftRouteTerminationPolicy).To(BeNil())
			Expect(snapshot.CustomCAGit).To(BeNil())
			Expect(snapshot.CustomCAS3).To(BeNil())
			Expect(snapshot.CustomCAOauth2Proxy).To(BeNil())
		})
	})
})
