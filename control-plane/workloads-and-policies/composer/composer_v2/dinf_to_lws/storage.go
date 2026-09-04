/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dinf_to_lws

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/itemized"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	workloadsManagerApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	corev1 "k8s.io/api/core/v1"
)

func (dinfComposer *DistributedInferenceToLWSComposer) processStorageFields(
	src *workloadsApi.DistributedInferenceLeaderWorkerSpecStorage,
	createHomeDir *bool,
	largeShmRequest *bool,
	role string,
	target *corev1.PodTemplateSpec,
	config *workloadsManagerApi.WorkloadV2Configuration,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil && convutils.BoolIsFalse(createHomeDir) && convutils.BoolIsFalse(largeShmRequest) {
			return nil
		}

		volumesAndMounts := composer.VolumesAndMounts{Config: config}
		runner := &composer.StepRunner{}
		if src != nil {
			runner.Batch(
				dinfComposer.fc.ComposeConfigMap(replicaItemPrefix(itemized.ConfigMapPrefix, role), src.ConfigMapVolume, &volumesAndMounts, composer.MapConfigMapToK8sVol, nil),
				dinfComposer.fc.ComposeEmptyDir(replicaItemPrefix(itemized.EmptyDirPrefix, role), src.EmptyDirVolume, &volumesAndMounts, composer.MapEmptyDirToK8sVol, nil),
				dinfComposer.fc.ComposeSecretVolume(replicaItemPrefix(itemized.SecretStoragePrefix, role), src.SecretVolume, &volumesAndMounts, composer.MapSecretToK8sVol, nil),
				dinfComposer.fc.ComposePvc(replicaItemPrefix(itemized.PvcPrefix, role), src.Pvc, &volumesAndMounts, composer_v2.MapPvcToWorkloadRunner, nil),
			)
		}
		runner.Batch(
			dinfComposer.fc.ComposeHomeDirectory(createHomeDir, &volumesAndMounts, composer.MapEmptyDirToK8sVol, nil),
			dinfComposer.fc.ComposeLargeShm(largeShmRequest, &volumesAndMounts, composer.MapEmptyDirToK8sVol, nil),
		)

		container := &target.Spec.Containers[0]
		container.VolumeMounts = append(container.VolumeMounts, volumesAndMounts.VolumeMount...)
		target.Spec.Volumes = append(target.Spec.Volumes, volumesAndMounts.Volume...)
		return runner.Error()
	}
}
