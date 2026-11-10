// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package init_containers

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

func s3CollisionConfig(s3Secret, gitSecret, authSecret string) composer.StaticClusterStorageConfig {
	cfg := composer.StaticClusterStorageConfig{}
	cfg.SetS3FsImage("goofys:readiness")
	cfg.SetS3FsMountCheckImage("goofys:readiness")
	cfg.SetGitSyncImage("git-sync:test")
	cfg.SetExternalAuthUrlEnabled(true)
	cfg.SetIsCustomCAS3Enabled(true)
	cfg.SetCustomCAS3Secret(s3Secret)
	cfg.SetCustomCAS3Key("ca.pem")
	cfg.SetIsCustomCAGitEnabled(true)
	cfg.SetCustomCAGitSecret(gitSecret)
	cfg.SetCustomCAGitKey("ca.pem")
	cfg.SetIsCustomCAAuthEnabled(true)
	cfg.SetCustomCAAuthSecret(authSecret)
	cfg.SetCustomCAAuthKey("ca.pem")
	return cfg
}

var _ = Describe("S3 readiness collisions", func() {
	DescribeTable("keeps application paths separate from sidecar mounts", func(applicationPath string) {
		cfg := s3CollisionConfig("s3-ready-0", "git-ca", "auth-ca")
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

	DescribeTable("keeps readiness volume names distinct from existing CA volumes", func(s3Secret, gitSecret, authSecret string) {
		cfg := s3CollisionConfig(s3Secret, gitSecret, authSecret)
		info := &composer.S3Info{Config: cfg, Items: map[string]*composer.S3Item{
			"first":  {Bucket: "first", Path: "/runai-goofys"},
			"second": {Bucket: "second", Path: "/data"},
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
				Expect(volume.Name).To(Equal(volume.Secret.SecretName))
				Expect(volume.Secret.Items[0].Key).To(Equal("ca.pem"))
			}
		}
		Expect(caNames).To(ConsistOf(s3Secret, gitSecret, authSecret))
		Expect(s3CAVolumeMount(cfg).Name).To(Equal(s3Secret))
		Expect(oauth2CAVolumeMount(cfg).Name).To(Equal(authSecret))
		gitContainers, err := composer.BuildGitSyncInitContainers(gitInfo)
		Expect(err).To(BeNil())
		Expect(gitContainers[0].VolumeMounts).To(ContainElement(corev1.VolumeMount{
			Name: gitSecret, MountPath: "/etc/ssl/certs/ca.pem", SubPath: "ca.pem",
		}))
		containers, err := BuildS3InitContainers(info)
		Expect(err).To(BeNil())
		readyNames := []string{}
		for _, container := range containers {
			for _, mount := range container.VolumeMounts {
				if mount.MountPath == "/runai-goofys" {
					readyNames = append(readyNames, mount.Name)
					Expect(volumes).To(ContainElement(corev1.Volume{
						Name: mount.Name, VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
					}))
					for _, appMount := range BuildS3MainVolumeMounts(info) {
						Expect(appMount.Name).NotTo(Equal(mount.Name))
					}
				}
			}
		}
		Expect(readyNames).To(HaveLen(2))
		Expect(readyNames[0]).NotTo(Equal(readyNames[1]))
	},
		Entry("ordinary CA names", "s3-ca", "git-ca", "auth-ca"),
		Entry("S3 CA collision", "s3-ready-0", "git-ca", "auth-ca"),
		Entry("Git CA collision", "s3-ca", "s3-ready-0", "auth-ca"),
		Entry("OAuth CA collision", "s3-ca", "git-ca", "s3-ready-0"),
		Entry("second S3 volume collision", "s3-ready-1", "git-ca", "auth-ca"),
		Entry("successive candidate collisions", "s3-ready-0", "s3-ready-0-ready", "s3-ready-0-ready-ready"),
	)
})
