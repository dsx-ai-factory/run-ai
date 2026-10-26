// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package init_containers

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

func s3CollisionConfig(secret string) composer.StaticClusterStorageConfig {
	cfg := composer.StaticClusterStorageConfig{}
	cfg.SetS3FsImage("goofys:readiness")
	cfg.SetS3FsMountCheckImage("goofys:readiness")
	cfg.SetGitSyncImage("git-sync:test")
	cfg.SetExternalAuthUrlEnabled(true)
	cfg.SetIsCustomCAS3Enabled(true)
	cfg.SetCustomCAS3Secret(secret)
	cfg.SetCustomCAS3Key("ca.pem")
	cfg.SetIsCustomCAGitEnabled(true)
	cfg.SetCustomCAGitSecret(secret)
	cfg.SetCustomCAGitKey("ca.pem")
	cfg.SetIsCustomCAAuthEnabled(true)
	cfg.SetCustomCAAuthSecret(secret)
	cfg.SetCustomCAAuthKey("ca.pem")
	return cfg
}

var _ = Describe("S3 readiness collisions", func() {
	DescribeTable("keeps application paths separate from sidecar mounts", func(applicationPath string) {
		cfg := s3CollisionConfig("s3-ready-0")
		info := &composer.S3Info{Config: cfg, Items: map[string]*composer.S3Item{
			"data": {Bucket: "bucket", Path: applicationPath},
		}}
		containers, err := BuildS3InitContainers(info)
		Expect(err).To(BeNil())
		Expect(containers).To(HaveLen(1))
		container := containers[0]
		paths := map[string]bool{}
		for _, mount := range container.VolumeMounts {
			Expect(paths).NotTo(HaveKey(mount.MountPath), "duplicate sidecar mount path: %s", mount.MountPath)
			paths[mount.MountPath] = true
		}
		Expect(container.VolumeMounts[0].MountPath).To(Equal("/runai-s3"))
		Expect(container.Command[len(container.Command)-1]).To(Equal("/runai-s3"))
		Expect(container.StartupProbe.Exec.Command[len(container.StartupProbe.Exec.Command)-1]).To(Equal("/runai-s3"))
		Expect(BuildS3MainVolumeMounts(info)[0].MountPath).To(Equal(applicationPath))
	},
		Entry("readiness directory", "/runai-goofys"),
		Entry("readiness marker path", "/runai-goofys/ready"),
		Entry("nested readiness directory", "/runai-goofys/data"),
		Entry("CA certificate path", "/etc/ssl/certs/ca.pem"),
		Entry("internal S3 directory", "/runai-s3"),
	)

	DescribeTable("keeps CA Secret references independent of Pod volume names", func(secret string) {
		cfg := s3CollisionConfig(secret)
		info := &composer.S3Info{Config: cfg, Items: map[string]*composer.S3Item{
			"data": {Bucket: "bucket", Path: "/runai-goofys"},
		}}
		gitInfo := &composer.GitSyncInfo{Config: cfg, Items: map[string]*composer.GitSyncItem{
			"git": {Nick: "git", Repository: "https://example.invalid/repo"},
		}}
		authInfo := &composer.ExposedUrlInfo{Config: cfg, Items: map[string]*composer.ExposedUrlInfoItem{
			"url": {Nick: "url", Url: "https://example.invalid"},
		}}
		volumes := BuildS3Volumes(info)
		volumes = append(volumes, composer.BuildGitSyncVolumes(gitInfo)...)
		volumes = append(volumes, BuildOAuth2ProxyVolumes(authInfo)...)
		names := map[string]bool{}
		caNames := []string{}
		for _, volume := range volumes {
			Expect(names).NotTo(HaveKey(volume.Name), "duplicate Pod volume name: %s", volume.Name)
			names[volume.Name] = true
			if volume.Secret != nil {
				caNames = append(caNames, volume.Name)
				Expect(volume.Secret.SecretName).To(Equal(secret))
				Expect(volume.Secret.Items[0].Key).To(Equal("ca.pem"))
			}
		}
		Expect(caNames).To(ConsistOf("custom-ca-s3", "custom-ca-git", "custom-ca"))
		Expect(s3CAVolumeMount(cfg).Name).To(Equal("custom-ca-s3"))
		Expect(oauth2CAVolumeMount(cfg).Name).To(Equal("custom-ca"))
		containers, err := composer.BuildGitSyncInitContainers(gitInfo)
		Expect(err).To(BeNil())
		Expect(containers[0].VolumeMounts).To(ContainElement(corev1.VolumeMount{
			Name: "custom-ca-git", MountPath: "/etc/ssl/certs/ca.pem", SubPath: "ca.pem",
		}))
	},
		Entry("readiness volume name", "s3-ready-0"),
		Entry("another readiness volume name", "s3-ready-1"),
		Entry("S3 data volume name", "s3-0"),
		Entry("Git data volume name", "code-sync-0"),
		Entry("auth data volume name", "runai-auth"),
		Entry("internal CA alias", "custom-ca-s3"),
	)
})
