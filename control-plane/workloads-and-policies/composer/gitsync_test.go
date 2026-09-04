/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"context"

	clustersApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/builders"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

var gitSyncEnvVars = []string{
	"GIT_SYNC_IMAGE",
	"CUSTOM_CA_GIT_ENABLED",
	"CUSTOM_CA_GIT_KEY",
	"CUSTOM_CA_GIT_SECRET",
}

func gitSyncInfoFromStorageFields(sf *field_groups.StorageFields, cfg ClusterStorageConfig) *GitSyncInfo {
	items := sf.GitSync.Deref()
	info := &GitSyncInfo{
		Items:  make(map[string]*GitSyncItem, len(items)),
		Config: cfg,
	}
	for nick, item := range items {
		v := item.Value
		if v == nil {
			info.Items[nick] = nil
			continue
		}
		info.Items[nick] = &GitSyncItem{
			Nick:                 nick,
			Repository:           convutils.Deref(v.Repository, ""),
			Branch:               convutils.Deref(v.Branch, ""),
			Revision:             convutils.Deref(v.Revision, ""),
			Directory:            convutils.Deref(v.Directory, ""),
			Image:                convutils.Deref(v.Image, ""),
			SecretName:           convutils.Deref(v.SecretName, ""),
			Username:             convutils.Deref(v.Username, ""),
			Password:             convutils.Deref(v.Password, ""),
			AuthenticatingMethod: convutils.Deref(v.AuthenticatingMethod, ""),
		}
	}
	return info
}

var _ = Describe("BuildGitSyncInitContainers resources", func() {
	newInfo := func(cfg ClusterStorageConfig) *GitSyncInfo {
		return &GitSyncInfo{
			Config: cfg,
			Items: map[string]*GitSyncItem{
				"repo": {Nick: "repo", Repository: "https://github.com/example/repo.git", Image: testGitSyncImage},
			},
		}
	}

	It("rejects an injected container without a reported or per-item image", func() {
		info := newInfo(StaticClusterStorageConfig{})
		info.Items["repo"].Image = ""
		containers, err := BuildGitSyncInitContainers(info)
		Expect(containers).To(BeNil())
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("gitSyncImage"))
	})

	It("applies configured resources to every git-sync instance", func() {
		cfg := StaticClusterStorageConfig{}
		expected := corev1.ResourceRequirements{Requests: corev1.ResourceList{
			corev1.ResourceCPU: resource.MustParse("25m"),
		}}
		cfg.SetGitSyncResources(expected)

		containers, err := BuildGitSyncInitContainers(newInfo(cfg))

		Expect(err).To(BeNil())
		Expect(containers).To(HaveLen(1))
		Expect(containers[0].Resources).To(Equal(expected))
	})

	It("uses empty resources when unset or explicitly empty", func() {
		unset, unsetErr := BuildGitSyncInitContainers(newInfo(StaticClusterStorageConfig{}))
		explicitConfig := StaticClusterStorageConfig{}
		explicitConfig.SetGitSyncResources(corev1.ResourceRequirements{})
		explicit, explicitErr := BuildGitSyncInitContainers(newInfo(explicitConfig))

		Expect(unsetErr).To(BeNil())
		Expect(explicitErr).To(BeNil())
		Expect(unset[0].Resources).To(Equal(corev1.ResourceRequirements{}))
		Expect(explicit[0].Resources).To(Equal(corev1.ResourceRequirements{}))
	})

	It("reports invalid quantities with the git-sync field name", func() {
		config := clustersApi.ClusterReportedConfig{}
		config.SetGitSyncResources(reportedResources(nil, map[string]string{"memory": "invalid"}))

		_, err := BuildGitSyncInitContainers(newInfo(NewClusterStorageConfigFromCluster(ClusterWithConfig(config))))

		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("gitSyncResources"))
		Expect(err.Message).To(ContainSubstring(`limits.memory has invalid quantity "invalid"`))
	})
})

var _ = Describe("BuildGitSyncInitContainers parity", func() {
	var (
		ctx             context.Context
		containerFields *field_groups.ContainerFields
		storageFields   *field_groups.StorageFields
		cfg             ClusterStorageConfig
	)

	BeforeEach(func() {
		ctx = context.Background()
		for _, name := range gitSyncEnvVars {
			GinkgoT().Setenv(name, "")
		}
		GinkgoT().Setenv("GIT_SYNC_IMAGE", testGitSyncImage)

		containerFields = &field_groups.ContainerFields{
			Environment: &flds.DictionaryField{
				Value: &map[string]flds.DictionaryItemAndRules{},
			},
		}
		storageFields = &field_groups.StorageFields{
			GitSync: &flds.GitSyncField{
				Value: &map[string]flds.GitSyncItemAndRules{},
			},
		}
		config := DefaultClusterStorageConfig().(StaticClusterStorageConfig)
		config.SetGitSyncImage(testGitSyncImage)
		cfg = config
	})

	expectParity := func() {
		clusterOut := builders.GetGitSyncInitContainers(ctx, containerFields, storageFields, false)
		cpOut, err := BuildGitSyncInitContainers(gitSyncInfoFromStorageFields(storageFields, cfg))
		Expect(err).To(BeNil())
		Expect(cpOut).To(Equal(clusterOut))
	}

	Context("Basic functionality", func() {
		It("should return empty when no gitsync items are specified", func() {
			expectParity()
		})

		It("should mirror cluster output with repository and branch", func() {
			repoUrl := "https://github.com/example/repo.git"
			branch := "main"
			(*storageFields.GitSync.Value)["repo1"] = flds.GitSyncItemAndRules{
				Value: &flds.GitSyncItemValue{Repository: &repoUrl, Branch: &branch},
			}
			expectParity()
		})

		It("should mirror cluster output with repository and revision", func() {
			repoUrl := "https://github.com/example/repo.git"
			revision := "a1b2c3d4"
			(*storageFields.GitSync.Value)["repo1"] = flds.GitSyncItemAndRules{
				Value: &flds.GitSyncItemValue{Repository: &repoUrl, Revision: &revision},
			}
			expectParity()
		})

		It("should mirror cluster output when nil items are mixed in", func() {
			(*storageFields.GitSync.Value)["nil-item"] = flds.GitSyncItemAndRules{Value: nil}
			repoUrl := "https://github.com/example/repo.git"
			(*storageFields.GitSync.Value)["valid-item"] = flds.GitSyncItemAndRules{
				Value: &flds.GitSyncItemValue{Repository: &repoUrl},
			}
			expectParity()
		})
	})

	Context("Authentication scenarios", func() {
		It("should mirror cluster output for plain text username and password", func() {
			repoUrl := "https://github.com/example/repo.git"
			username := "user"
			password := "pass"
			(*storageFields.GitSync.Value)["repo1"] = flds.GitSyncItemAndRules{
				Value: &flds.GitSyncItemValue{
					Repository: &repoUrl,
					Username:   &username,
					Password:   &password,
				},
			}
			expectParity()
		})

		It("should mirror cluster output for username and password from secret", func() {
			repoUrl := "https://github.com/example/repo.git"
			secretName := "git-creds"
			usernameKey := "username-key"
			passwordKey := "password-key"
			(*storageFields.GitSync.Value)["repo1"] = flds.GitSyncItemAndRules{
				Value: &flds.GitSyncItemValue{
					Repository: &repoUrl,
					SecretName: &secretName,
					Username:   &usernameKey,
					Password:   &passwordKey,
				},
			}
			expectParity()
		})

		It("should mirror cluster output for password authenticating method via secret", func() {
			repoUrl := "https://github.com/example/repo.git"
			secretName := "git-creds"
			usernameKey := "username-key"
			passwordKey := "password-key"
			(*storageFields.GitSync.Value)["repo1"] = flds.GitSyncItemAndRules{
				Value: &flds.GitSyncItemValue{
					Repository:           &repoUrl,
					SecretName:           &secretName,
					Username:             &usernameKey,
					Password:             &passwordKey,
					AuthenticatingMethod: ptr.To(constants.GitAuthenticatingMethodPassword),
				},
			}
			expectParity()
		})

		It("should mirror cluster output for ssh-key authenticating method", func() {
			repoUrl := "https://github.com/example/repo.git"
			secretName := "git-ssh-creds"
			(*storageFields.GitSync.Value)["repo1"] = flds.GitSyncItemAndRules{
				Value: &flds.GitSyncItemValue{
					Repository:           &repoUrl,
					SecretName:           &secretName,
					AuthenticatingMethod: ptr.To(constants.GitAuthenticatingMethodSSHKey),
				},
			}
			expectParity()
		})

		It("should mirror cluster output for a custom image", func() {
			repoUrl := "https://github.com/example/repo.git"
			customImage := "custom-git-sync:latest"
			(*storageFields.GitSync.Value)["repo1"] = flds.GitSyncItemAndRules{
				Value: &flds.GitSyncItemValue{Repository: &repoUrl, Image: &customImage},
			}
			expectParity()
		})
	})

	Context("Documented SecurityContext divergence from the cluster builder", func() {
		expectSecurityContextDivergence := func() {
			clusterOut := builders.GetGitSyncInitContainers(ctx, containerFields, storageFields, false)
			cpOut, err := BuildGitSyncInitContainers(gitSyncInfoFromStorageFields(storageFields, cfg))
			Expect(err).To(BeNil())

			Expect(cpOut).To(HaveLen(1))
			Expect(clusterOut).To(HaveLen(1))
			Expect(cpOut[0].SecurityContext).To(BeNil(), "CP twin must not emit a SecurityContext")
			Expect(clusterOut[0].SecurityContext).NotTo(BeNil(), "cluster builder still aligns uid/gid for cloned files")

			clusterOut[0].SecurityContext = nil
			Expect(cpOut[0]).To(Equal(clusterOut[0]), "modulo SecurityContext, CP and cluster outputs must be identical")
		}

		It("should drop the SecurityContext that cluster sets from runAsUser/Uid/Gid", func() {
			containerFields.RunAsUser = flds.NewOptionalBoolField(ptr.To(true))
			containerFields.RunAsUid = flds.NewOptionalIntField(ptr.To(1001))
			containerFields.RunAsGid = flds.NewOptionalIntField(ptr.To(1001))

			repoUrl := "https://github.com/example/repo.git"
			(*storageFields.GitSync.Value)["repo1"] = flds.GitSyncItemAndRules{
				Value: &flds.GitSyncItemValue{Repository: &repoUrl},
			}
			expectSecurityContextDivergence()
		})

		It("should drop the SecurityContext that cluster sets from NB_UID/NB_GID env vars", func() {
			(*containerFields.Environment.Value)[constants.JupyterNbUid] = flds.DictionaryItemAndRules{
				Value: ptr.To("1002"),
			}
			(*containerFields.Environment.Value)[constants.JupyterNbGid] = flds.DictionaryItemAndRules{
				Value: ptr.To("1003"),
			}

			repoUrl := "https://github.com/example/repo.git"
			(*storageFields.GitSync.Value)["repo1"] = flds.GitSyncItemAndRules{
				Value: &flds.GitSyncItemValue{Repository: &repoUrl},
			}
			expectSecurityContextDivergence()
		})

		It("should drop the DefaultJupyter SecurityContext that cluster sets when isJupyter is true", func() {
			repoUrl := "https://github.com/example/repo.git"
			(*storageFields.GitSync.Value)["repo1"] = flds.GitSyncItemAndRules{
				Value: &flds.GitSyncItemValue{Repository: &repoUrl},
			}

			clusterIsJupyter := builders.GetGitSyncInitContainers(ctx, containerFields, storageFields, true)
			cpOut, err := BuildGitSyncInitContainers(gitSyncInfoFromStorageFields(storageFields, cfg))
			Expect(err).To(BeNil())

			Expect(cpOut).To(HaveLen(1))
			Expect(clusterIsJupyter).To(HaveLen(1))
			Expect(cpOut[0].SecurityContext).To(BeNil())
			Expect(clusterIsJupyter[0].SecurityContext).NotTo(BeNil())
			Expect(*clusterIsJupyter[0].SecurityContext.RunAsUser).To(Equal(builders.DefaultJupyterUid))
			Expect(*clusterIsJupyter[0].SecurityContext.RunAsGroup).To(Equal(builders.DefaultJupyterGid))
		})
	})

	Context("Multiple git sync items", func() {
		It("should mirror cluster output for multiple git sync items", func() {
			repo1 := "https://github.com/example/repo1.git"
			repo2 := "https://github.com/example/repo2.git"
			(*storageFields.GitSync.Value)["repo1"] = flds.GitSyncItemAndRules{
				Value: &flds.GitSyncItemValue{Repository: &repo1},
			}
			(*storageFields.GitSync.Value)["repo2"] = flds.GitSyncItemAndRules{
				Value: &flds.GitSyncItemValue{Repository: &repo2},
			}
			expectParity()
		})
	})

	Context("Custom CA", func() {
		It("should mirror cluster output when custom CA is enabled", func() {
			GinkgoT().Setenv("CUSTOM_CA_GIT_ENABLED", "true")
			GinkgoT().Setenv("GIT_SYNC_IMAGE", testGitSyncImage)
			cfg = StaticClusterStorageConfig{
				gitSyncImage:         testGitSyncImage,
				isCustomCAGitEnabled: true,
				customCAGitSecret:    "runai-ca-cert",
				customCAGitKey:       "runai-ca.pem",
			}

			repoUrl := "https://github.com/example/repo.git"
			(*storageFields.GitSync.Value)["repo1"] = flds.GitSyncItemAndRules{
				Value: &flds.GitSyncItemValue{Repository: &repoUrl},
			}
			expectParity()
		})
	})
})

var _ = Describe("BuildGitSyncVolumes", func() {
	var (
		info *GitSyncInfo
		cfg  ClusterStorageConfig
	)

	BeforeEach(func() {
		for _, name := range gitSyncEnvVars {
			GinkgoT().Setenv(name, "")
		}
		cfg = DefaultClusterStorageConfig()
		info = &GitSyncInfo{Config: cfg}
	})

	It("should return empty when there are no git entries", func() {
		Expect(BuildGitSyncVolumes(info)).To(BeEmpty())
	})

	It("should emit one emptyDir volume per git entry, indexed by sorted key", func() {
		info.Items = map[string]*GitSyncItem{
			"alpha": {Nick: "alpha", Repository: "https://github.com/example/repo1.git"},
			"beta":  {Nick: "beta", Repository: "https://github.com/example/repo2.git"},
		}

		volumes := BuildGitSyncVolumes(info)

		Expect(volumes).To(HaveLen(2))
		Expect(volumes[0].Name).To(Equal("code-sync-0"))
		Expect(volumes[0].EmptyDir).NotTo(BeNil())
		Expect(volumes[1].Name).To(Equal("code-sync-1"))
		Expect(volumes[1].EmptyDir).NotTo(BeNil())
	})

	It("should add the SSH secret volume after the emptyDir when SSH authentication is selected", func() {
		info.Items = map[string]*GitSyncItem{
			"repo1": {
				Nick:                 "repo1",
				Repository:           "https://github.com/example/repo.git",
				SecretName:           "git-ssh-creds",
				AuthenticatingMethod: constants.GitAuthenticatingMethodSSHKey,
			},
		}

		volumes := BuildGitSyncVolumes(info)

		Expect(volumes).To(HaveLen(2))
		Expect(volumes[0].Name).To(Equal("code-sync-0"))
		Expect(volumes[1].Name).To(Equal("git-ssh-secret-0"))
		Expect(volumes[1].Secret).NotTo(BeNil())
		Expect(volumes[1].Secret.SecretName).To(Equal("git-ssh-creds"))
	})

	It("should append a single custom-CA volume after the per-entry volumes when CustomCAGit is enabled", func() {
		info.Config = StaticClusterStorageConfig{
			gitSyncImage:         testGitSyncImage,
			isCustomCAGitEnabled: true,
			customCAGitSecret:    "runai-ca-cert",
			customCAGitKey:       "runai-ca.pem",
		}
		info.Items = map[string]*GitSyncItem{
			"alpha": {Nick: "alpha", Repository: "https://github.com/example/repo1.git"},
			"beta":  {Nick: "beta", Repository: "https://github.com/example/repo2.git"},
		}

		volumes := BuildGitSyncVolumes(info)

		Expect(volumes).To(HaveLen(3))
		Expect(volumes[0].Name).To(Equal("code-sync-0"))
		Expect(volumes[1].Name).To(Equal("code-sync-1"))
		Expect(volumes[2].Name).To(Equal("runai-ca-cert"))
		Expect(volumes[2].Secret).NotTo(BeNil())
		Expect(volumes[2].Secret.SecretName).To(Equal("runai-ca-cert"))
		Expect(*volumes[2].Secret.DefaultMode).To(Equal(int32(420)))
		Expect(*volumes[2].Secret.Optional).To(BeTrue())
		Expect(volumes[2].Secret.Items).To(HaveLen(1))
		Expect(volumes[2].Secret.Items[0].Key).To(Equal("runai-ca.pem"))
		Expect(volumes[2].Secret.Items[0].Path).To(Equal("runai-ca.pem"))
	})

	It("should skip nil items rather than producing placeholder volumes", func() {
		info.Items = map[string]*GitSyncItem{
			"empty-slot": nil,
			"real-item":  {Nick: "real-item", Repository: "https://github.com/example/repo.git"},
		}

		volumes := BuildGitSyncVolumes(info)

		Expect(volumes).To(HaveLen(1))
		Expect(volumes[0].Name).To(Equal("code-sync-1"))
	})
})

var _ = Describe("BuildGitSyncMainVolumeMounts", func() {
	var info *GitSyncInfo

	BeforeEach(func() {
		info = &GitSyncInfo{}
	})

	It("should return empty when there are no git entries", func() {
		Expect(BuildGitSyncMainVolumeMounts(info)).To(BeEmpty())
	})

	It("should mount each code-sync volume at the user-specified Directory", func() {
		info.Items = map[string]*GitSyncItem{
			"repo1": {Nick: "repo1", Repository: "https://github.com/example/repo.git", Directory: "/workspace/src"},
		}

		mounts := BuildGitSyncMainVolumeMounts(info)

		Expect(mounts).To(HaveLen(1))
		Expect(mounts[0].Name).To(Equal("code-sync-0"))
		Expect(mounts[0].MountPath).To(Equal("/workspace/src"))
	})

	It("should fall back to /code-<nick> when no Directory is provided", func() {
		info.Items = map[string]*GitSyncItem{
			"repo1": {Nick: "repo1", Repository: "https://github.com/example/repo.git"},
		}

		mounts := BuildGitSyncMainVolumeMounts(info)

		Expect(mounts).To(HaveLen(1))
		Expect(mounts[0].MountPath).To(Equal("/code-repo1"))
	})

	It("should produce one mount per git entry, indexed by sorted key", func() {
		info.Items = map[string]*GitSyncItem{
			"alpha": {Nick: "alpha", Repository: "https://github.com/example/repo1.git"},
			"beta":  {Nick: "beta", Repository: "https://github.com/example/repo2.git"},
		}

		mounts := BuildGitSyncMainVolumeMounts(info)

		Expect(mounts).To(HaveLen(2))
		Expect(mounts[0].Name).To(Equal("code-sync-0"))
		Expect(mounts[0].MountPath).To(Equal("/code-alpha"))
		Expect(mounts[1].Name).To(Equal("code-sync-1"))
		Expect(mounts[1].MountPath).To(Equal("/code-beta"))
	})
})
