/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"fmt"

	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"

	corev1 "k8s.io/api/core/v1"
)

const (
	gitSyncContainerPrefix  = "git-sync"
	GitSyncNamePrefix       = "code-sync"
	GitSyncPathPrefix       = "/code"
	gitSyncSSHKeyNamePrefix = "git-ssh-secret"
	gitSyncSSHKeyPath       = "/etc/git-secret"

	envGitSyncSSH        = "GITSYNC_SSH"
	envGitSyncKnownHosts = "GITSYNC_SSH_KNOWN_HOSTS"
	envGitSyncRepo       = "GITSYNC_REPO"
	envGitSyncRef        = "GITSYNC_REF"
	envGitSyncUsername   = "GITSYNC_USERNAME"
	envGitSyncPassword   = "GITSYNC_PASSWORD"
	envGitSyncRoot       = "GITSYNC_ROOT"
	envGitSyncOneTime    = "GITSYNC_ONE_TIME"
	envHome              = "HOME"

	gitSyncHomeDir = "/tmp"

	gitSyncCADefaultMode int32 = 420
)

type GitSyncInfo struct {
	Items  map[string]*GitSyncItem
	Config ClusterStorageConfig
}

type GitSyncItem struct {
	Nick                 string
	Repository           string
	Branch               string
	Revision             string
	Directory            string
	Image                string
	SecretName           string
	Username             string
	Password             string
	AuthenticatingMethod string
}

// BuildGitSyncInitContainers returns one git-sync init container per non-nil
// entry in info.Items, in sorted-key order. The init container's SecurityContext
// is left as nil; in Kubernetes that means the init container inherits the
// pod-level SecurityContext, which is written from the workload spec by
// processSecurityFields in iw_tw_workloads/iw_tw_to_job.go.
func BuildGitSyncInitContainers(info *GitSyncInfo) ([]corev1.Container, *workloadsApi.Error) {
	result := make([]corev1.Container, 0)
	if info == nil || info.Config == nil || len(info.Items) == 0 {
		return result, nil
	}
	cfg := info.Config
	var resources corev1.ResourceRequirements
	resourcesLoaded := false

	for idx, key := range SortedMapKeys(info.Items) {
		item := info.Items[key]
		if item == nil {
			continue
		}
		image := cfg.GitSyncImage()
		if item.Image != "" {
			image = item.Image
		}
		if image == "" {
			return nil, ErrClusterValueMissing("gitSyncImage")
		}
		if !resourcesLoaded {
			var resourcesErr error
			resources, resourcesErr = cfg.GitSyncResources()
			if resourcesErr != nil {
				return nil, ErrClusterValueInvalid("gitSyncResources", resourcesErr)
			}
			resourcesLoaded = true
		}

		var envInfos []EnvVarInfo
		var volumeMounts []corev1.VolumeMount

		if item.Repository != "" {
			envInfos = append(envInfos, EnvVarInfo{Name: envGitSyncRepo, ValueType: EnvValueText, ValueFrom: item.Repository})
		}

		if item.Revision != "" {
			envInfos = append(envInfos, EnvVarInfo{Name: envGitSyncRef, ValueType: EnvValueText, ValueFrom: item.Revision})
		}

		if item.Branch != "" {
			envInfos = append(envInfos, EnvVarInfo{Name: envGitSyncRef, ValueType: EnvValueText, ValueFrom: item.Branch})
		}

		if item.SecretName == "" {
			if item.Username != "" {
				envInfos = append(envInfos, EnvVarInfo{Name: envGitSyncUsername, ValueType: EnvValueText, ValueFrom: item.Username})
			}
			if item.Password != "" {
				envInfos = append(envInfos, EnvVarInfo{Name: envGitSyncPassword, ValueType: EnvValueText, ValueFrom: item.Password})
			}
		} else {
			if item.Username != "" {
				envInfos = append(envInfos, EnvVarInfo{Name: envGitSyncUsername, ValueType: EnvValueSecret, ValueFrom: item.SecretName, Key: item.Username})
			}
			if item.Password != "" {
				envInfos = append(envInfos, EnvVarInfo{Name: envGitSyncPassword, ValueType: EnvValueSecret, ValueFrom: item.SecretName, Key: item.Password})
			}
		}

		envInfos = append(envInfos,
			EnvVarInfo{Name: envGitSyncRoot, ValueType: EnvValueText, ValueFrom: fmt.Sprintf("%s-%s", GitSyncPathPrefix, item.Nick)},
			EnvVarInfo{Name: envGitSyncOneTime, ValueType: EnvValueText, ValueFrom: "true"},
			EnvVarInfo{Name: envHome, ValueType: EnvValueText, ValueFrom: gitSyncHomeDir},
		)

		if isGitSSHAuthenticationSelected(item) {
			envInfos = append(envInfos,
				EnvVarInfo{Name: envGitSyncSSH, ValueType: EnvValueText, ValueFrom: "true"},
				EnvVarInfo{Name: envGitSyncKnownHosts, ValueType: EnvValueText, ValueFrom: "false"},
			)
			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      fmt.Sprintf("%s-%d", gitSyncSSHKeyNamePrefix, idx),
				MountPath: gitSyncSSHKeyPath,
			})
		}

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      fmt.Sprintf("%s-%d", GitSyncNamePrefix, idx),
			MountPath: fmt.Sprintf("%s-%s", GitSyncPathPrefix, item.Nick),
		})

		if cfg.IsCustomCAGitEnabled() {
			volumeMounts = append(volumeMounts, gitSyncCAVolumeMount(cfg))
		}

		var env []corev1.EnvVar
		if err := EnvVarsToK8sEnv(envInfos, &env, nil); err != nil {
			return nil, err
		}

		result = append(result, corev1.Container{
			Env:          env,
			Image:        image,
			Name:         fmt.Sprintf("%s-%s", gitSyncContainerPrefix, item.Nick),
			Resources:    resources,
			VolumeMounts: volumeMounts,
		})
	}

	return result, nil
}

func gitSyncCAVolumeMount(cfg ClusterStorageConfig) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      cfg.CustomCAGitSecret(),
		MountPath: fmt.Sprintf("/etc/ssl/certs/%s", cfg.CustomCAGitKey()),
		SubPath:   cfg.CustomCAGitKey(),
	}
}

func BuildGitSyncVolumes(info *GitSyncInfo) []corev1.Volume {
	result := make([]corev1.Volume, 0)
	if info == nil || info.Config == nil || len(info.Items) == 0 {
		return result
	}
	cfg := info.Config

	hasAnyGit := false
	for idx, key := range SortedMapKeys(info.Items) {
		item := info.Items[key]
		if item == nil {
			continue
		}
		hasAnyGit = true

		result = append(result, corev1.Volume{
			Name: fmt.Sprintf("%s-%d", GitSyncNamePrefix, idx),
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})

		if isGitSSHAuthenticationSelected(item) {
			result = append(result, corev1.Volume{
				Name: fmt.Sprintf("%s-%d", gitSyncSSHKeyNamePrefix, idx),
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: item.SecretName,
					},
				},
			})
		}
	}

	if hasAnyGit && cfg.IsCustomCAGitEnabled() {
		result = append(result, gitSyncCAVolume(cfg))
	}

	return result
}

func gitSyncCAVolume(cfg ClusterStorageConfig) corev1.Volume {
	return corev1.Volume{
		Name: cfg.CustomCAGitSecret(),
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				DefaultMode: convutils.Ref(gitSyncCADefaultMode),
				Items: []corev1.KeyToPath{
					{Key: cfg.CustomCAGitKey(), Path: cfg.CustomCAGitKey()},
				},
				Optional:   convutils.Ref(true),
				SecretName: cfg.CustomCAGitSecret(),
			},
		},
	}
}

func BuildGitSyncMainVolumeMounts(info *GitSyncInfo) []corev1.VolumeMount {
	result := make([]corev1.VolumeMount, 0)
	if info == nil || len(info.Items) == 0 {
		return result
	}

	for idx, key := range SortedMapKeys(info.Items) {
		item := info.Items[key]
		if item == nil {
			continue
		}

		mountPath := item.Directory
		if mountPath == "" {
			mountPath = fmt.Sprintf("%s-%s", GitSyncPathPrefix, item.Nick)
		}
		result = append(result, corev1.VolumeMount{
			Name:      fmt.Sprintf("%s-%d", GitSyncNamePrefix, idx),
			MountPath: mountPath,
		})
	}

	return result
}

func gitAuthenticationMethod(item *GitSyncItem) string {
	if item == nil {
		return ""
	}
	return item.AuthenticatingMethod
}

func isGitSSHAuthenticationSelected(item *GitSyncItem) bool {
	return gitAuthenticationMethod(item) == constants.GitAuthenticatingMethodSSHKey
}
