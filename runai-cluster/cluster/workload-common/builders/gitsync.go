/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	commonutils "github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils"
	"k8s.io/utils/ptr"

	corev1 "k8s.io/api/core/v1"
)

const (
	gitSyncContainerPrefix = "git-sync"

	GitSyncNamePrefix = "code-sync"
	GitSyncPathPrefix = "/code"

	gitSyncSSHKeyNamePrefix = "git-ssh-secret"
	gitSyncSSHKeyPath       = "/etc/git-secret"

	DefaultJupyterUid = int64(1000) // uid of "jovyan"
	DefaultJupyterGid = int64(100)  // gid of "users"
)

var sshGitSyncEnvs = []corev1.EnvVar{
	{
		Name:  "GITSYNC_SSH",
		Value: "true",
	},
	{
		Name:  "GITSYNC_SSH_KNOWN_HOSTS",
		Value: "false",
	},
}

// GetGitSyncInitContainers composes git-sync init containers.
//
// Deprecated: use GetGitSyncInitContainersWithError. This compatibility
// wrapper panics when injected-container resource configuration is invalid.
func GetGitSyncInitContainers(ctx context.Context, containerFields *field_groups.ContainerFields, storageFields *field_groups.StorageFields, isJupyter bool) []corev1.Container {
	containers, err := GetGitSyncInitContainersWithError(ctx, containerFields, storageFields, isJupyter)
	if err != nil {
		panic(err)
	}
	return containers
}

// GetGitSyncInitContainersWithError composes git-sync init containers and
// returns resource configuration errors to the caller.
func GetGitSyncInitContainersWithError(_ context.Context, containerFields *field_groups.ContainerFields, storageFields *field_groups.StorageFields, isJupyter bool) ([]corev1.Container, error) {
	result := make([]corev1.Container, 0)

	items := storageFields.GitSync.Deref()
	if len(items) == 0 {
		return result, nil
	}

	resources, err := config.GitSyncResources()
	if err != nil {
		return nil, fmt.Errorf("get git-sync resources: %w", err)
	}

	for idx, nick := range commonutils.SortedKeys(items) {

		if items[nick].Value == nil {
			continue
		}
		params := *items[nick].Value

		var envs []corev1.EnvVar
		var volumeMounts []corev1.VolumeMount

		if ptr.Deref(params.Repository, "") != "" {
			envs = append(envs, corev1.EnvVar{
				Name:  "GITSYNC_REPO",
				Value: ptr.Deref(params.Repository, ""),
			})
		}

		if ptr.Deref(params.Revision, "") != "" {
			envs = append(envs, corev1.EnvVar{
				Name:  "GITSYNC_REF",
				Value: ptr.Deref(params.Revision, ""),
			})
		}

		if ptr.Deref(params.Branch, "") != "" {
			envs = append(envs, corev1.EnvVar{
				Name:  "GITSYNC_REF",
				Value: ptr.Deref(params.Branch, ""),
			})
		}

		secretName := ptr.Deref(params.SecretName, "")

		if secretName == "" {
			// username and password specified as plain text
			if ptr.Deref(params.Username, "") != "" {
				envs = append(envs, corev1.EnvVar{
					Name:  "GITSYNC_USERNAME",
					Value: *params.Username,
				})
			}
			if ptr.Deref(params.Password, "") != "" {
				envs = append(envs, corev1.EnvVar{
					Name:  "GITSYNC_PASSWORD",
					Value: *params.Password,
				})
			}
		} else {
			// username and password loaded from a secret
			if ptr.Deref(params.Username, "") != "" {
				envs = append(envs, corev1.EnvVar{
					Name: "GITSYNC_USERNAME",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: secretName,
							},
							Key: *params.Username,
						},
					},
				})
			}
			if ptr.Deref(params.Password, "") != "" {
				envs = append(envs, corev1.EnvVar{
					Name: "GITSYNC_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: secretName,
							},
							Key: *params.Password,
						},
					},
				})
			}
		}

		envs = append(envs, corev1.EnvVar{
			Name:  "GITSYNC_ROOT",
			Value: fmt.Sprintf("%s-%s", GitSyncPathPrefix, nick),
		})

		envs = append(envs, corev1.EnvVar{
			Name:  "GITSYNC_ONE_TIME",
			Value: "true",
		})

		envs = append(envs, corev1.EnvVar{
			Name:  "HOME",
			Value: "/tmp",
		})

		sshAuthentication := isGitSSHAuthenticationSelected(&params)
		if sshAuthentication {
			envs = append(envs, sshGitSyncEnvs...)

			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      fmt.Sprintf("%s-%d", gitSyncSSHKeyNamePrefix, idx),
				MountPath: gitSyncSSHKeyPath,
			})
		}

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      fmt.Sprintf("%s-%d", GitSyncNamePrefix, idx),
			MountPath: fmt.Sprintf("%s-%s", GitSyncPathPrefix, nick),
		})

		if config.IsCustomCAEnabled(EnvVarPrefixCAGit) {
			volumeMounts = append(volumeMounts, ComposeCAVolumeMount(EnvVarPrefixCAGit))
		}

		//
		//   when using the default image, we have to make sure uid/gid is correct. if user provides custom image,
		//   it will be his responsibility to make sure the uid/gid defined in his custom image is correct.
		//
		var sc *corev1.SecurityContext

		image := config.DefaultGitSyncImage()
		if params.Image != nil && *params.Image != "" {
			image = *params.Image
		} else {
			sc = gitSyncSecurityContext(containerFields, isJupyter)
		}

		result = append(result, corev1.Container{
			Env:             envs,
			Image:           image,
			Name:            fmt.Sprintf("%s-%s", gitSyncContainerPrefix, nick),
			Resources:       resources,
			SecurityContext: sc,
			VolumeMounts:    volumeMounts,
		})
	}

	return result, nil
}

// the default git-sync image is built with default user (65536). we have to adopt the uid/gid
// of the cloned files to have uid/gid that the main container can read/write
func gitSyncSecurityContext(containerFields *field_groups.ContainerFields, isJupyter bool) *corev1.SecurityContext {

	result := corev1.SecurityContext{}

	//
	//   if the main container is set to runAsUser, set the security context to be identical to those
	//   of the main container.
	//
	if containerFields.RunAsUser != nil && containerFields.RunAsUser.Deref() {
		if containerFields.RunAsUid != nil {
			result.RunAsUser = ptr.To(int64(containerFields.RunAsUid.Deref()))
		}
		if containerFields.RunAsGid != nil {
			result.RunAsGroup = ptr.To(int64(containerFields.RunAsGid.Deref()))
		}
		return &result
	}

	//
	//   2nd option is user providing NB_UID/NB_GID for jupyter. this is intended for jupyter job
	//   created from jupyter environment asset.
	//
	envVars := containerFields.Environment.Deref()
	nbUidStr, uidEnvProvided := envVars[constants.JupyterNbUid]
	nbGidStr, gidEnvProvided := envVars[constants.JupyterNbGid]
	if uidEnvProvided && gidEnvProvided {
		nbUid, errUid := strconv.Atoi(PatchUidGid(ptr.Deref(nbUidStr.Value, ""), containerFields))
		nbGid, errGid := strconv.Atoi(PatchUidGid(ptr.Deref(nbGidStr.Value, ""), containerFields))
		if errUid == nil && errGid == nil {
			result.RunAsUser = ptr.To(int64(nbUid))
			result.RunAsGroup = ptr.To(int64(nbGid))
			return &result
		}
	}

	//
	//  finally, pre-assets jupyter, if the user does not specify uid/gid to runAsUser, the default
	//  container runs jupyter in uid/gid 1000:100. need to align the security context accordingly
	//
	if isJupyter && config.JupyterUserIdentity() {
		result.RunAsUser = ptr.To(DefaultJupyterUid)
		result.RunAsGroup = ptr.To(DefaultJupyterGid)
		return &result
	}

	return nil
}

func gitAuthenticationMethod(gitSync *field_types.GitSyncItemValue) string {
	if gitSync == nil || gitSync.AuthenticatingMethod == nil {
		return ""
	}
	return *gitSync.AuthenticatingMethod
}

func isGitSSHAuthenticationSelected(gitSync *field_types.GitSyncItemValue) bool {
	return gitAuthenticationMethod(gitSync) == constants.GitAuthenticatingMethodSSHKey
}
