/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

var _ = Describe("GitSync tests", func() {
	var (
		ctx             context.Context
		containerFields *field_groups.ContainerFields
		storageFields   *field_groups.StorageFields
		isJupyter       bool
	)

	BeforeEach(func() {
		// Setup common test variables
		ctx = context.Background()

		// Initialize containerFields with default values
		containerFields = &field_groups.ContainerFields{
			Environment: &flds.DictionaryField{
				Value: &map[string]flds.DictionaryItemAndRules{},
			},
		}

		// Initialize storageFields with empty gitsync map
		storageFields = &field_groups.StorageFields{
			GitSync: &flds.GitSyncField{
				Value: &map[string]flds.GitSyncItemAndRules{},
			},
		}

		isJupyter = false
	})

	Context("Basic functionality", func() {
		It("should return empty when no gitsync items are specified", func() {
			containers := GetGitSyncInitContainers(ctx, containerFields, storageFields, isJupyter)
			Expect(containers).To(BeEmpty())
		})

		It("should create a container with repository and branch", func() {
			// Add a git sync item with repository and branch
			repoUrl := "https://github.com/example/repo.git"
			branch := "main"

			(*storageFields.GitSync.Value)["repo1"] = flds.GitSyncItemAndRules{
				Value: &flds.GitSyncItemValue{
					Repository: &repoUrl,
					Branch:     &branch,
				},
			}

			containers := GetGitSyncInitContainers(ctx, containerFields, storageFields, isJupyter)

			Expect(containers).To(HaveLen(1))
			container := containers[0]

			Expect(container.Name).To(Equal("git-sync-repo1"))
			Expect(container.Resources).To(Equal(corev1.ResourceRequirements{}))

			// Check environment variables
			hasRepo := false
			hasRef := false
			hasRoot := false
			hasOneTime := false

			for _, env := range container.Env {
				switch env.Name {
				case "GITSYNC_REPO":
					hasRepo = true
					Expect(env.Value).To(Equal(repoUrl))
				case "GITSYNC_REF":
					hasRef = true
					Expect(env.Value).To(Equal(branch))
				case "GITSYNC_ROOT":
					hasRoot = true
					Expect(env.Value).To(Equal("/code-repo1"))
				case "GITSYNC_ONE_TIME":
					hasOneTime = true
					Expect(env.Value).To(Equal("true"))
				}
			}

			Expect(hasRepo).To(BeTrue())
			Expect(hasRef).To(BeTrue())
			Expect(hasRoot).To(BeTrue())
			Expect(hasOneTime).To(BeTrue())

			// Check volume mounts
			Expect(container.VolumeMounts).To(HaveLen(1))
			Expect(container.VolumeMounts[0].MountPath).To(Equal("/code-repo1"))
		})

		It("should create a container with repository and revision", func() {
			// Add a git sync item with repository and revision
			repoUrl := "https://github.com/example/repo.git"
			revision := "a1b2c3d4"

			(*storageFields.GitSync.Value)["repo1"] = flds.GitSyncItemAndRules{
				Value: &flds.GitSyncItemValue{
					Repository: &repoUrl,
					Revision:   &revision,
				},
			}

			containers := GetGitSyncInitContainers(ctx, containerFields, storageFields, isJupyter)

			Expect(containers).To(HaveLen(1))

			// Check for revision in environment variables
			hasRevision := false
			for _, env := range containers[0].Env {
				if env.Name == "GITSYNC_REF" && env.Value == revision {
					hasRevision = true
					break
				}
			}

			Expect(hasRevision).To(BeTrue())
		})

		It("should skip nil git sync items", func() {
			// Add a nil git sync item
			(*storageFields.GitSync.Value)["nil-item"] = flds.GitSyncItemAndRules{
				Value: nil,
			}

			// Add a valid git sync item
			repoUrl := "https://github.com/example/repo.git"
			(*storageFields.GitSync.Value)["valid-item"] = flds.GitSyncItemAndRules{
				Value: &flds.GitSyncItemValue{
					Repository: &repoUrl,
				},
			}

			containers := GetGitSyncInitContainers(ctx, containerFields, storageFields, isJupyter)

			// Should only have one container for the valid item
			Expect(containers).To(HaveLen(1))
			Expect(containers[0].Name).To(Equal("git-sync-valid-item"))
		})
	})

	Context("Authentication scenarios", func() {
		It("should handle plain text username and password", func() {
			// Add a git sync item with plain text username and password
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

			containers := GetGitSyncInitContainers(ctx, containerFields, storageFields, isJupyter)

			Expect(containers).To(HaveLen(1))

			// Check for username and password in environment variables
			hasUsername := false
			hasPassword := false

			for _, env := range containers[0].Env {
				if env.Name == "GITSYNC_USERNAME" && env.Value == username {
					hasUsername = true
				}
				if env.Name == "GITSYNC_PASSWORD" && env.Value == password {
					hasPassword = true
				}
			}

			Expect(hasUsername).To(BeTrue())
			Expect(hasPassword).To(BeTrue())
		})

		It("should handle username and password from secret", func() {
			// Add a git sync item with secret-based username and password
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

			containers := GetGitSyncInitContainers(ctx, containerFields, storageFields, isJupyter)

			Expect(containers).To(HaveLen(1))

			// Check for username and password in environment variables via secretKeyRef
			hasUsernameSecret := false
			hasPasswordSecret := false

			for _, env := range containers[0].Env {
				if env.Name == "GITSYNC_USERNAME" && env.ValueFrom != nil &&
					env.ValueFrom.SecretKeyRef != nil &&
					env.ValueFrom.SecretKeyRef.Name == secretName &&
					env.ValueFrom.SecretKeyRef.Key == usernameKey {
					hasUsernameSecret = true
				}
				if env.Name == "GITSYNC_PASSWORD" && env.ValueFrom != nil &&
					env.ValueFrom.SecretKeyRef != nil &&
					env.ValueFrom.SecretKeyRef.Name == secretName &&
					env.ValueFrom.SecretKeyRef.Key == passwordKey {
					hasPasswordSecret = true
				}
			}

			Expect(hasUsernameSecret).To(BeTrue())
			Expect(hasPasswordSecret).To(BeTrue())
		})

		It("should handle username and password from secret with ", func() {
			// Add a git sync item with secret-based username and password
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

			containers := GetGitSyncInitContainers(ctx, containerFields, storageFields, isJupyter)

			Expect(containers).To(HaveLen(1))

			// Check for username and password in environment variables via secretKeyRef
			hasUsernameSecret := false
			hasPasswordSecret := false

			for _, env := range containers[0].Env {
				if env.Name == "GITSYNC_USERNAME" && env.ValueFrom != nil &&
					env.ValueFrom.SecretKeyRef != nil &&
					env.ValueFrom.SecretKeyRef.Name == secretName &&
					env.ValueFrom.SecretKeyRef.Key == usernameKey {
					hasUsernameSecret = true
				}
				if env.Name == "GITSYNC_PASSWORD" && env.ValueFrom != nil &&
					env.ValueFrom.SecretKeyRef != nil &&
					env.ValueFrom.SecretKeyRef.Name == secretName &&
					env.ValueFrom.SecretKeyRef.Key == passwordKey {
					hasPasswordSecret = true
				}
			}

			Expect(hasUsernameSecret).To(BeTrue())
			Expect(hasPasswordSecret).To(BeTrue())
		})

		It("should handle user ssh key", func() {
			// Add a git sync item with secret-based username and password
			repoUrl := "https://github.com/example/repo.git"
			secretName := "git-ssh-creds"

			(*storageFields.GitSync.Value)["repo1"] = flds.GitSyncItemAndRules{
				Value: &flds.GitSyncItemValue{
					Repository:           &repoUrl,
					SecretName:           &secretName,
					AuthenticatingMethod: ptr.To(constants.GitAuthenticatingMethodSSHKey),
				},
			}

			containers := GetGitSyncInitContainers(ctx, containerFields, storageFields, isJupyter)

			Expect(containers).To(HaveLen(1))

			hasSSH_Envar := false
			hasKnownHostEnvar := false

			for _, env := range containers[0].Env {
				if env.Name == "GITSYNC_SSH" && env.Value == "true" {
					hasSSH_Envar = true
				}
				if env.Name == "GITSYNC_SSH_KNOWN_HOSTS" && env.Value == "false" {
					hasKnownHostEnvar = true
				}
			}

			Expect(hasSSH_Envar).To(BeTrue())
			Expect(hasKnownHostEnvar).To(BeTrue())
		})

		It("should handle custom image", func() {
			// Add a git sync item with custom image
			repoUrl := "https://github.com/example/repo.git"
			customImage := "custom-git-sync:latest"

			(*storageFields.GitSync.Value)["repo1"] = flds.GitSyncItemAndRules{
				Value: &flds.GitSyncItemValue{
					Repository: &repoUrl,
					Image:      &customImage,
				},
			}

			containers := GetGitSyncInitContainers(ctx, containerFields, storageFields, isJupyter)

			Expect(containers).To(HaveLen(1))
			Expect(containers[0].Image).To(Equal(customImage))

			// Custom image should not have a security context set
			Expect(containers[0].SecurityContext).To(BeNil())
		})
	})

	Context("Security context handling", func() {
		It("should use runAsUser/runAsGid from containerFields when available", func() {
			// Setup containerFields with runAsUser and runAsUid/runAsGid
			containerFields.RunAsUser = flds.NewOptionalBoolField(ptr.To(true))
			containerFields.RunAsUid = flds.NewOptionalIntField(ptr.To(1001))
			containerFields.RunAsGid = flds.NewOptionalIntField(ptr.To(1001))

			// Add a git sync item
			repoUrl := "https://github.com/example/repo.git"

			(*storageFields.GitSync.Value)["repo1"] = flds.GitSyncItemAndRules{
				Value: &flds.GitSyncItemValue{
					Repository: &repoUrl,
				},
			}

			containers := GetGitSyncInitContainers(ctx, containerFields, storageFields, isJupyter)

			Expect(containers).To(HaveLen(1))
			Expect(containers[0].SecurityContext).NotTo(BeNil())
			Expect(*containers[0].SecurityContext.RunAsUser).To(Equal(int64(1001)))
			Expect(*containers[0].SecurityContext.RunAsGroup).To(Equal(int64(1001)))
		})

		It("should use NB_UID/NB_GID from environment when available", func() {
			// Setup containerFields with jupyter environment variables
			(*containerFields.Environment.Value)[constants.JupyterNbUid] = flds.DictionaryItemAndRules{
				Value: ptr.To("1002"),
			}
			(*containerFields.Environment.Value)[constants.JupyterNbGid] = flds.DictionaryItemAndRules{
				Value: ptr.To("1003"),
			}

			// Add a git sync item
			repoUrl := "https://github.com/example/repo.git"

			(*storageFields.GitSync.Value)["repo1"] = flds.GitSyncItemAndRules{
				Value: &flds.GitSyncItemValue{
					Repository: &repoUrl,
				},
			}

			containers := GetGitSyncInitContainers(ctx, containerFields, storageFields, isJupyter)

			Expect(containers).To(HaveLen(1))
			Expect(containers[0].SecurityContext).NotTo(BeNil())
			Expect(*containers[0].SecurityContext.RunAsUser).To(Equal(int64(1002)))
			Expect(*containers[0].SecurityContext.RunAsGroup).To(Equal(int64(1003)))
		})

		It("should use default Jupyter UID/GID when isJupyter is true", func() {
			// Set isJupyter to true
			isJupyter = true

			// Add a git sync item
			repoUrl := "https://github.com/example/repo.git"

			(*storageFields.GitSync.Value)["repo1"] = flds.GitSyncItemAndRules{
				Value: &flds.GitSyncItemValue{
					Repository: &repoUrl,
				},
			}

			containers := GetGitSyncInitContainers(ctx, containerFields, storageFields, isJupyter)

			Expect(containers).To(HaveLen(1))
			Expect(containers[0].SecurityContext).NotTo(BeNil())
			Expect(*containers[0].SecurityContext.RunAsUser).To(Equal(DefaultJupyterUid))
			Expect(*containers[0].SecurityContext.RunAsGroup).To(Equal(DefaultJupyterGid))
		})
	})

	Context("Multiple git sync items", func() {
		It("should create multiple containers for multiple git sync items", func() {
			// Add two git sync items
			repo1 := "https://github.com/example/repo1.git"
			repo2 := "https://github.com/example/repo2.git"

			(*storageFields.GitSync.Value)["repo1"] = flds.GitSyncItemAndRules{
				Value: &flds.GitSyncItemValue{
					Repository: &repo1,
				},
			}

			(*storageFields.GitSync.Value)["repo2"] = flds.GitSyncItemAndRules{
				Value: &flds.GitSyncItemValue{
					Repository: &repo2,
				},
			}

			containers := GetGitSyncInitContainers(ctx, containerFields, storageFields, isJupyter)

			Expect(containers).To(HaveLen(2))

			// Check container names
			containerNames := []string{containers[0].Name, containers[1].Name}
			Expect(containerNames).To(ConsistOf("git-sync-repo1", "git-sync-repo2"))

			// Check repositories in environment variables
			repos := []string{}

			for _, container := range containers {
				for _, env := range container.Env {
					if env.Name == "GITSYNC_REPO" {
						repos = append(repos, env.Value)
					}
				}
			}

			Expect(repos).To(ConsistOf(repo1, repo2))
		})
	})

	Context("Resource requirements", func() {
		BeforeEach(func() {
			repository := "https://github.com/example/repo.git"
			(*storageFields.GitSync.Value)["repo"] = flds.GitSyncItemAndRules{
				Value: &flds.GitSyncItemValue{Repository: &repository},
			}
		})

		It("applies configured resources to every instance", func() {
			GinkgoT().Setenv(config.EnvVarGitSyncResources, `{"requests":{"cpu":"25m"},"limits":{"memory":"64Mi"}}`)

			containers, err := GetGitSyncInitContainersWithError(ctx, containerFields, storageFields, isJupyter)

			Expect(err).NotTo(HaveOccurred())
			Expect(containers).To(HaveLen(1))
			Expect(containers[0].Resources).To(Equal(corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("25m")},
				Limits:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("64Mi")},
			}))
		})

		It("treats explicit empty resources as empty", func() {
			GinkgoT().Setenv(config.EnvVarGitSyncResources, `{}`)

			containers, err := GetGitSyncInitContainersWithError(ctx, containerFields, storageFields, isJupyter)

			Expect(err).NotTo(HaveOccurred())
			Expect(containers[0].Resources).To(Equal(corev1.ResourceRequirements{}))
		})

		It("rejects malformed resources", func() {
			GinkgoT().Setenv(config.EnvVarGitSyncResources, `{"limits":{"memory":"invalid"}}`)

			containers, err := GetGitSyncInitContainersWithError(ctx, containerFields, storageFields, isJupyter)

			Expect(containers).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(config.EnvVarGitSyncResources)))
		})
	})
})
