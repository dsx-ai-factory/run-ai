/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package init_containers

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	corev1 "k8s.io/api/core/v1"
)

const (
	s3DataMountPath   = "/runai-s3"
	s3ContainerPrefix = "s3-fs"
	s3VolumePrefix    = "s3"

	envS3AccessKeyID     = "AWS_ACCESS_KEY_ID"
	envS3SecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	envS3SessionToken    = "AWS_SESSION_TOKEN"

	s3CADefaultMode int32 = 420

	s3CACertsDir        = "/etc/ssl/certs"
	s3ReadyMountPath    = "/runai-goofys"
	s3ReadyFile         = s3ReadyMountPath + "/ready"
	s3ReadyVolumePrefix = "s3-ready"

	// goofys gets at most these extra args: --uid <v>, --gid <v>, --endpoint <v>, --ready-file <v>, bucket, path.
	maxGoofysPerItemArgs = 10
)

// goofysBaseCommand are the flags goofys is always launched with; uid/gid,
// endpoint, bucket and path are appended per workload in BuildS3InitContainers.
var goofysBaseCommand = []string{"goofys", "-f", "--dir-mode", "0777", "--file-mode", "0777", "-o", "allow_other", "--debug_s3"}

func BuildS3InitContainers(info *composer.S3Info) ([]corev1.Container, *workloadsApi.Error) {
	result := make([]corev1.Container, 0)
	if info == nil || info.Config == nil || len(info.Items) == 0 {
		return result, nil
	}
	cfg := info.Config
	var resources corev1.ResourceRequirements
	resourcesLoaded := false

	mountBidirectional := corev1.MountPropagationBidirectional
	restartAlways := corev1.ContainerRestartPolicyAlways

	for idx, key := range composer.SortedMapKeys(info.Items) {
		item := info.Items[key]
		if item == nil {
			continue
		}
		image := cfg.S3FsImage()
		if image == "" {
			return nil, composer.ErrClusterValueMissing("s3FileSystemImage")
		}
		if !resourcesLoaded {
			var resourcesErr error
			resources, resourcesErr = cfg.S3FileSystemResources()
			if resourcesErr != nil {
				return nil, composer.ErrClusterValueInvalid("s3FileSystemResources", resourcesErr)
			}
			resourcesLoaded = true
		}

		var envInfos []composer.EnvVarInfo
		if item.SecretName == "" {
			if item.AccessKeyIdKey != "" {
				envInfos = append(envInfos, composer.EnvVarInfo{Name: envS3AccessKeyID, ValueType: composer.EnvValueText, ValueFrom: item.AccessKeyIdKey})
			}
			if item.SecretAccessKeyKey != "" {
				envInfos = append(envInfos, composer.EnvVarInfo{Name: envS3SecretAccessKey, ValueType: composer.EnvValueText, ValueFrom: item.SecretAccessKeyKey})
			}
		} else {
			if item.AccessKeyIdKey != "" {
				envInfos = append(envInfos, composer.EnvVarInfo{Name: envS3AccessKeyID, ValueType: composer.EnvValueSecret, ValueFrom: item.SecretName, Key: item.AccessKeyIdKey})
			}
			if item.SecretAccessKeyKey != "" {
				envInfos = append(envInfos, composer.EnvVarInfo{Name: envS3SecretAccessKey, ValueType: composer.EnvValueSecret, ValueFrom: item.SecretName, Key: item.SecretAccessKeyKey})
			}
		}
		if item.SessionToken != "" {
			envInfos = append(envInfos, composer.EnvVarInfo{Name: envS3SessionToken, ValueType: composer.EnvValueText, ValueFrom: item.SessionToken})
		}

		command := make([]string, len(goofysBaseCommand), len(goofysBaseCommand)+maxGoofysPerItemArgs)
		copy(command, goofysBaseCommand)
		if info.RunAsUid != nil {
			command = append(command, "--uid", fmt.Sprintf("%d", *info.RunAsUid))
		}
		if info.RunAsGid != nil {
			command = append(command, "--gid", fmt.Sprintf("%d", *info.RunAsGid))
		}
		if item.Url != "" {
			command = append(command, "--endpoint", item.Url)
		}
		// Older and custom images may not implement the readiness flags.
		mountCheckEnabled := s3MountCheckEnabled(cfg)
		mountPath := item.Path
		var startupProbe *corev1.Probe
		if mountCheckEnabled {
			// Keep the sidecar's mount points independent of the application path.
			mountPath = s3DataMountPath
			command = append(command, "--ready-file", s3ReadyFile)
			startupProbe = &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{Exec: &corev1.ExecAction{
					Command: []string{"goofys", "--check-ready", s3ReadyFile, mountPath},
				}},
				PeriodSeconds:    2,
				TimeoutSeconds:   2,
				FailureThreshold: 150,
			}
		}
		command = append(command, item.Bucket, mountPath)

		volumeMounts := []corev1.VolumeMount{
			{
				Name:             fmt.Sprintf("%s-%d", s3VolumePrefix, idx),
				MountPath:        mountPath,
				MountPropagation: &mountBidirectional,
			},
		}
		if mountCheckEnabled {
			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      s3ReadyVolumeName(idx, cfg),
				MountPath: s3ReadyMountPath,
			})
		}
		if cfg.IsCustomCAS3Enabled() {
			volumeMounts = append(volumeMounts, s3CAVolumeMount(cfg))
		}

		var env []corev1.EnvVar
		if err := composer.EnvVarsToK8sEnv(envInfos, &env, nil); err != nil {
			return nil, err
		}
		// Propagate the cluster's FIPS mode to the goofys sidecar. The goofys image
		// is built with GOFIPS140, so with no GODEBUG it defaults to fips140=on
		// (not enforced); set it explicitly to match the cluster (e.g. only).
		if godebug := cfg.FipsGodebug(); godebug != "" {
			env = append(env, corev1.EnvVar{Name: "GODEBUG", Value: godebug})
		}

		result = append(result, corev1.Container{
			Image:           image,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Name:            fmt.Sprintf("%s-%d", s3ContainerPrefix, idx),
			SecurityContext: &corev1.SecurityContext{Privileged: convutils.Ref(true)},
			Command:         command,
			Env:             env,
			Resources:       resources,
			VolumeMounts:    volumeMounts,
			RestartPolicy:   &restartAlways,
			StartupProbe:    startupProbe,
		})
	}

	return result, nil
}

func s3MountCheckEnabled(cfg composer.ClusterStorageConfig) bool {
	image := cfg.S3FsImage()
	return image != "" && cfg.S3FsMountCheckImage() == image
}

// CA Secret names are also Pod volume names; keep readiness volumes distinct.
func s3ReadyVolumeName(idx int, cfg composer.ClusterStorageConfig) string {
	name := fmt.Sprintf("%s-%d", s3ReadyVolumePrefix, idx)
	for name == cfg.CustomCAS3Secret() || name == cfg.CustomCAGitSecret() || name == cfg.CustomCAAuthSecret() {
		name += "-ready"
	}
	return name
}

func s3CAVolumeMount(cfg composer.ClusterStorageConfig) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      cfg.CustomCAS3Secret(),
		MountPath: fmt.Sprintf("%s/%s", s3CACertsDir, cfg.CustomCAS3Key()),
		SubPath:   cfg.CustomCAS3Key(),
	}
}

func BuildS3Volumes(info *composer.S3Info) []corev1.Volume {
	result := make([]corev1.Volume, 0)
	if info == nil || info.Config == nil || len(info.Items) == 0 {
		return result
	}
	cfg := info.Config

	hasAnyS3 := false
	for idx, key := range composer.SortedMapKeys(info.Items) {
		item := info.Items[key]
		if item == nil {
			continue
		}
		hasAnyS3 = true

		result = append(result, corev1.Volume{
			Name:         fmt.Sprintf("%s-%d", s3VolumePrefix, idx),
			VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
		})
		if s3MountCheckEnabled(cfg) {
			result = append(result, corev1.Volume{
				Name:         s3ReadyVolumeName(idx, cfg),
				VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
			})
		}
	}

	if hasAnyS3 && cfg.IsCustomCAS3Enabled() {
		result = append(result, s3CAVolume(cfg))
	}

	return result
}

func s3CAVolume(cfg composer.ClusterStorageConfig) corev1.Volume {
	return corev1.Volume{
		Name: cfg.CustomCAS3Secret(),
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				DefaultMode: convutils.Ref(s3CADefaultMode),
				Items: []corev1.KeyToPath{
					{Key: cfg.CustomCAS3Key(), Path: cfg.CustomCAS3Key()},
				},
				Optional:   convutils.Ref(true),
				SecretName: cfg.CustomCAS3Secret(),
			},
		},
	}
}

func BuildS3MainVolumeMounts(info *composer.S3Info) []corev1.VolumeMount {
	result := make([]corev1.VolumeMount, 0)
	if info == nil || len(info.Items) == 0 {
		return result
	}

	mountHostToContainer := corev1.MountPropagationHostToContainer

	for idx, key := range composer.SortedMapKeys(info.Items) {
		item := info.Items[key]
		if item == nil {
			continue
		}
		result = append(result, corev1.VolumeMount{
			Name:             fmt.Sprintf("%s-%d", s3VolumePrefix, idx),
			MountPath:        item.Path,
			MountPropagation: &mountHostToContainer,
		})
	}

	return result
}
