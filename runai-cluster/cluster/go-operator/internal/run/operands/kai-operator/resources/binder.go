/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	kaibinder "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/binder"
	kaicommon "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/common"
	"k8s.io/utils/ptr"

	clustercommonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/binder"
)

func convertBinderConfig(runaiConfig *runaiv1.RunaiConfig) *kaibinder.Binder {
	binderConfig := runaiConfig.Spec.Binder
	return &kaibinder.Binder{
		Service: &kaicommon.Service{
			Enabled:         binderConfig.Enabled,
			Image:           convertImage(binderConfig.Image),
			Resources:       (*kaicommon.Resources)(binderConfig.Resources),
			K8sClientConfig: convertBinderK8sClientConfig(binderConfig),
			Affinity:        binderConfig.Affinity,
		},
		ResourceReservation: &kaibinder.ResourceReservation{
			Image:              convertBinderResourceReservationImage(binderConfig),
			Namespace:          ptr.To(runaiv1.ReservationNamespace),
			ServiceAccountName: ptr.To(clustercommonconstants.ResourceReservationServiceAccountName),
			AppLabel:           ptr.To(clustercommonconstants.ResourceReservationAppLabelValue),
			RuntimeClassName:   ptr.To(""),
		},
		MaxConcurrentReconciles:     binderArgsMaxConcurrentReconciles(binderConfig),
		VolumeBindingTimeoutSeconds: binderArgsVolumeBindingTimeoutSeconds(binderConfig),
		ProbePort:                   binderWebhookProbePort(binderConfig),
		VPA:                         convertVPA(binderConfig.VPA),
	}
}

func convertBinderResourceReservationImage(binderConfig *binder.Binder) *kaicommon.Image {
	if binderConfig.ResourceReservation == nil {
		return nil
	}
	return convertImage(binderConfig.ResourceReservation.Image)
}

func convertBinderK8sClientConfig(binderConfig *binder.Binder) *kaicommon.K8sClientConfig {
	if binderConfig.Args == nil {
		return nil
	}
	if binderConfig.Args.QPS == nil && binderConfig.Args.Burst == nil {
		return nil
	}
	return &kaicommon.K8sClientConfig{
		QPS:   binderConfig.Args.QPS,
		Burst: binderConfig.Args.Burst,
	}
}

func binderArgsMaxConcurrentReconciles(binderConfig *binder.Binder) *int {
	if binderConfig.Args == nil {
		return nil
	}
	return binderConfig.Args.MaxConcurrentReconciles
}

func binderArgsVolumeBindingTimeoutSeconds(binderConfig *binder.Binder) *int {
	if binderConfig.Args == nil {
		return nil
	}
	return binderConfig.Args.VolumeBindingTimeoutSeconds
}

func binderWebhookProbePort(binderConfig *binder.Binder) *int {
	if binderConfig.Service == nil || binderConfig.Service.Webhook == nil {
		return nil
	}
	return binderConfig.Service.Webhook.ProbePort
}
