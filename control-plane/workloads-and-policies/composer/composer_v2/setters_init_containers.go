/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer_v2

import (
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/init_containers"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	corev1 "k8s.io/api/core/v1"
)

func UrlsToK8sInitContainers(info *composer.ExposedUrlInfo, target any, _ any) *workloadsApi.Error {
	if info == nil || len(info.Items) == 0 {
		return nil
	}

	podSpec, ok := target.(*corev1.PodSpec)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("UrlsToK8sInitContainers: target type %T is not *corev1.PodSpec", target))
	}

	bundle, err := init_containers.BuildOAuth2Bundle(info)
	if err != nil {
		return err
	}
	if len(bundle.Proxies) == 0 {
		return nil
	}

	if bundle.UsernameLoader != nil {
		podSpec.InitContainers = append(podSpec.InitContainers, *bundle.UsernameLoader)
	}
	podSpec.InitContainers = append(podSpec.InitContainers, bundle.Proxies...)
	podSpec.Volumes = append(podSpec.Volumes, bundle.Volumes...)
	return nil
}

func MapS3ToWorkloadK8s(info *composer.S3Info, target any, _ any) *workloadsApi.Error {
	if info == nil || len(info.Items) == 0 {
		return nil
	}

	podSpec, ok := target.(*corev1.PodSpec)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("MapS3ToWorkloadK8s: target type %T is not *corev1.PodSpec", target))
	}

	if len(podSpec.Containers) == 0 {
		return workloadsApi.NewError(http.StatusInternalServerError, "MapS3ToWorkloadK8s: PodSpec has no main container to mount s3 buckets into")
	}

	initContainers, err := init_containers.BuildS3InitContainers(info)
	if err != nil {
		return err
	}

	podSpec.InitContainers = append(podSpec.InitContainers, initContainers...)
	podSpec.Volumes = append(podSpec.Volumes, init_containers.BuildS3Volumes(info)...)
	podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, init_containers.BuildS3MainVolumeMounts(info)...)
	return nil
}
