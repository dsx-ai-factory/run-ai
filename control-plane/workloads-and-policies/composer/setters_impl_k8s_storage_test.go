/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	sdkcommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("MapGitToWorkloadK8s", func() {
	var podSpec *corev1.PodSpec

	BeforeEach(func() {
		podSpec = &corev1.PodSpec{
			Containers: []corev1.Container{{Name: "main"}},
		}
	})

	It("should be a no-op for nil info", func() {
		err := MapGitToWorkloadK8s(nil, podSpec, nil)
		Expect(err).To(BeNil())
		Expect(podSpec.InitContainers).To(BeEmpty())
		Expect(podSpec.Volumes).To(BeEmpty())
		Expect(podSpec.Containers[0].VolumeMounts).To(BeEmpty())
	})

	It("should be a no-op for info with no items", func() {
		err := MapGitToWorkloadK8s(&GitSyncInfo{Config: DefaultClusterStorageConfig()}, podSpec, nil)
		Expect(err).To(BeNil())
		Expect(podSpec.InitContainers).To(BeEmpty())
		Expect(podSpec.Volumes).To(BeEmpty())
		Expect(podSpec.Containers[0].VolumeMounts).To(BeEmpty())
	})

	It("should reject a target that is not a PodSpec pointer", func() {
		info := &GitSyncInfo{
			Config: DefaultClusterStorageConfig(),
			Items: map[string]*GitSyncItem{
				"repo": {Nick: "repo", Repository: "https://github.com/example/repo.git"},
			},
		}

		err := MapGitToWorkloadK8s(info, &struct{}{}, nil)

		Expect(err).NotTo(BeNil())
		Expect(err.Code).To(Equal(int32(http.StatusInternalServerError)))
	})

	It("should reject a PodSpec that has no main container", func() {
		info := &GitSyncInfo{
			Config: DefaultClusterStorageConfig(),
			Items: map[string]*GitSyncItem{
				"repo": {Nick: "repo", Repository: "https://github.com/example/repo.git"},
			},
		}

		err := MapGitToWorkloadK8s(info, &corev1.PodSpec{}, nil)

		Expect(err).NotTo(BeNil())
		Expect(err.Code).To(Equal(int32(http.StatusInternalServerError)))
	})

	It("should append an init container, the matching emptyDir volume, and the main-container mount", func() {
		info := &GitSyncInfo{
			Config: DefaultClusterStorageConfig(),
			Items: map[string]*GitSyncItem{
				"repo1": {Nick: "repo1", Repository: "https://github.com/example/repo.git"},
			},
		}

		err := MapGitToWorkloadK8s(info, podSpec, nil)

		Expect(err).To(BeNil())

		Expect(podSpec.InitContainers).To(HaveLen(1))
		Expect(podSpec.InitContainers[0].Name).To(Equal("git-sync-repo1"))
		Expect(podSpec.InitContainers[0].Image).To(Equal(sdkcommon.GitSyncImageRef))
		Expect(podSpec.InitContainers[0].SecurityContext).To(BeNil())

		Expect(podSpec.Volumes).To(HaveLen(1))
		Expect(podSpec.Volumes[0].Name).To(Equal("code-sync-0"))
		Expect(podSpec.Volumes[0].EmptyDir).NotTo(BeNil())

		Expect(podSpec.Containers[0].VolumeMounts).To(HaveLen(1))
		Expect(podSpec.Containers[0].VolumeMounts[0].Name).To(Equal("code-sync-0"))
		Expect(podSpec.Containers[0].VolumeMounts[0].MountPath).To(Equal("/code-repo1"))
	})

	It("should append to existing init containers / volumes / mounts rather than replacing them", func() {
		podSpec.InitContainers = append(podSpec.InitContainers, corev1.Container{Name: "preexisting-init"})
		podSpec.Volumes = append(podSpec.Volumes, corev1.Volume{Name: "preexisting-vol"})
		podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, corev1.VolumeMount{Name: "preexisting-mount"})

		info := &GitSyncInfo{
			Config: DefaultClusterStorageConfig(),
			Items: map[string]*GitSyncItem{
				"repo": {Nick: "repo", Repository: "https://github.com/example/repo.git"},
			},
		}

		err := MapGitToWorkloadK8s(info, podSpec, nil)

		Expect(err).To(BeNil())
		Expect(podSpec.InitContainers).To(HaveLen(2))
		Expect(podSpec.InitContainers[0].Name).To(Equal("preexisting-init"))
		Expect(podSpec.InitContainers[1].Name).To(Equal("git-sync-repo"))
		Expect(podSpec.Volumes).To(HaveLen(2))
		Expect(podSpec.Volumes[0].Name).To(Equal("preexisting-vol"))
		Expect(podSpec.Volumes[1].Name).To(Equal("code-sync-0"))
		Expect(podSpec.Containers[0].VolumeMounts).To(HaveLen(2))
		Expect(podSpec.Containers[0].VolumeMounts[0].Name).To(Equal("preexisting-mount"))
		Expect(podSpec.Containers[0].VolumeMounts[1].Name).To(Equal("code-sync-0"))
	})
})
