/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	featuregate "github.com/run-ai/runai/runai-cluster/cluster/workload-common/feature-gate"
	commonutils "github.com/run-ai/runai/runai-cluster/common/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

const (
	EnvVarAccessKeyID     = "AWS_ACCESS_KEY_ID"
	EnvVarSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	EnvVarSessionToken    = "AWS_SESSION_TOKEN"
)

// GetS3container composes S3 filesystem containers.
//
// Deprecated: use GetS3containerWithError. This compatibility wrapper panics
// when injected-container resource configuration is invalid.
func GetS3container(ctx context.Context, sharedFields *field_groups.WorkloadFields) []corev1.Container {
	containers, err := GetS3containerWithError(ctx, sharedFields)
	if err != nil {
		panic(err)
	}
	return containers
}

// GetS3containerWithError composes S3 filesystem containers and returns
// resource configuration errors to the caller.
func GetS3containerWithError(_ context.Context, sharedFields *field_groups.WorkloadFields) ([]corev1.Container, error) {
	s3buckets := sharedFields.S3buckets.Deref()
	if len(s3buckets) == 0 {
		return nil, nil
	}

	resources, err := config.S3FileSystemResources()
	if err != nil {
		return nil, fmt.Errorf("get S3 filesystem resources: %w", err)
	}

	containers := make([]corev1.Container, 0, len(s3buckets))

	sc := corev1.SecurityContext{
		Privileged: ptr.To(true),
	}

	for idx, s3bucketId := range commonutils.SortedKeys(s3buckets) {

		s3bucket := s3buckets[s3bucketId].Value
		if s3bucket == nil {
			continue
		}

		var envs []corev1.EnvVar

		secretName := ptr.Deref(s3bucket.SecretName, "")

		if secretName == "" {
			// access key id and secret access key in plain text
			if ptr.Deref(s3bucket.Key, "") != "" {
				envs = append(envs, corev1.EnvVar{
					Name:  EnvVarAccessKeyID,
					Value: *s3bucket.Key,
				})
			}
			if ptr.Deref(s3bucket.Secret, "") != "" {
				envs = append(envs, corev1.EnvVar{
					Name:  EnvVarSecretAccessKey,
					Value: *s3bucket.Secret,
				})
			}
		} else {
			// access key id and secret access key taken from secret
			if ptr.Deref(s3bucket.Key, "") != "" {
				envs = append(envs, corev1.EnvVar{
					Name: EnvVarAccessKeyID,
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: secretName,
							},
							Key: *s3bucket.Key,
						},
					},
				})
			}
			if ptr.Deref(s3bucket.Secret, "") != "" {
				envs = append(envs, corev1.EnvVar{
					Name: EnvVarSecretAccessKey,
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: secretName,
							},
							Key: *s3bucket.Secret,
						},
					},
				})
			}
		}

		if ptr.Deref(s3bucket.SessionToken, "") != "" {
			envs = append(envs, corev1.EnvVar{
				Name:  EnvVarSessionToken,
				Value: *s3bucket.SessionToken,
			})
		}

		if godebug := config.GetGodebug(); godebug != "" {
			envs = append(envs, corev1.EnvVar{
				Name:  config.EnvVarGodebug,
				Value: godebug,
			})
		}

		command := []string{"goofys", "-f", "--dir-mode", "0777", "--file-mode", "0777", "-o", "allow_other", "--debug_s3"}

		if sharedFields.RunAsUid != nil {
			command = append(command, "--uid", fmt.Sprintf("%d", sharedFields.RunAsUid.Deref()))
		}
		if sharedFields.RunAsGid != nil {
			command = append(command, "--gid", fmt.Sprintf("%d", sharedFields.RunAsGid.Deref()))
		}
		if ptr.Deref(s3bucket.Url, "") != "" {
			command = append(command, "--endpoint", ptr.Deref(s3bucket.Url, ""))
		}
		command = append(command, ptr.Deref(s3bucket.Bucket, "''"))
		command = append(command, ptr.Deref(s3bucket.Path, "''"))

		mountBidirectional := corev1.MountPropagationBidirectional

		image := config.DefaultS3FileSystemImage()
		if ptr.Deref(s3bucket.Image, "") != "" {
			image = *s3bucket.Image
		}

		c := corev1.Container{
			Image:           image,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Name:            fmt.Sprintf("s3-fs-%d", idx),
			Resources:       resources,
			SecurityContext: &sc,
			Command:         command,
			Env:             envs,
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:             fmt.Sprintf("%s-%d", s3mountPrefix, idx),
					MountPath:        ptr.Deref(s3bucket.Path, ""),
					MountPropagation: &mountBidirectional,
				},
			},
		}

		if config.IsCustomCAEnabled(EnvVarPrefixCAS3) {
			c.VolumeMounts = append(c.VolumeMounts, ComposeCAVolumeMount(EnvVarPrefixCAS3))
		}

		featureGateUtil := featuregate.GetFeatureGateUtilInstance()

		if featureGateUtil.FeatureGateSidecarContainersEnable {
			initContainerRestartPolicy := corev1.ContainerRestartPolicyAlways
			c.RestartPolicy = &initContainerRestartPolicy
		}

		containers = append(containers, c)
	}

	return containers, nil
}
