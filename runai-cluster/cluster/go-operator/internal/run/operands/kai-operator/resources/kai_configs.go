/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	kaicommon "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/common"
	kaiprometheus "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/prometheus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
)

const (
	kaiConfigCRName = "kai-config"
)

func KAIConfigForRunaiConfig(
	ctx context.Context, k8sReader client.Reader,
	runaiConfig *runaiv1.RunaiConfig,
) (client.Object, error) {
	kaiConfig := &kaiv1.Config{
		ObjectMeta: metav1.ObjectMeta{
			Name: kaiConfigCRName,
		},
	}

	err := k8sReader.Get(ctx, types.NamespacedName{Name: kaiConfigCRName}, kaiConfig)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	kaiConfig.Spec.Namespace = runaiConfig.GetNamespace()
	kaiConfig.Spec.Global = convertKAIGlobal(runaiConfig)
	kaiConfig.Spec.Binder = convertBinderConfig(runaiConfig)
	kaiConfig.Spec.PodGrouper = convertPodGrouperConfig(runaiConfig)
	kaiConfig.Spec.QueueController = convertQueueControllerConfig(runaiConfig)
	kaiConfig.Spec.PodGroupController = convertPodGroupControllerConfig(runaiConfig)
	kaiConfig.Spec.Scheduler = convertSchedulerConfig(runaiConfig)
	kaiConfig.Spec.NodeScaleAdjuster = convertNodeScaleAdjusterConfig(runaiConfig)
	kaiConfig.Spec.NumaPlacementExporter = runaiConfig.Spec.NumaPlacementExporter.DeepCopy()
	kaiConfig.Spec.Admission = convertAdmissionConfig()
	kaiConfig.Spec.Prometheus = convertPrometheusConfig(runaiConfig)

	return kaiConfig, nil
}

func KAIConfigStatus(ctx context.Context,
	k8sReader client.Reader, runaiConfig *runaiv1.RunaiConfig) (operands.Status, error) {
	resource, err := KAIConfigForRunaiConfig(ctx, k8sReader, runaiConfig)
	if err != nil {
		return operands.NotReadyStatus(), err
	}

	kaiConfig := resource.(*kaiv1.Config)
	if kaiConfig.GetUID() == "" {
		return operands.NotReadyStatus(fmt.Sprintf("kaiConfig [%s] is missing", kaiConfig.GetName())), nil
	}

	if len(kaiConfig.Status.Conditions) == 0 {
		return operands.NotReadyStatus(fmt.Sprintf("kaiConfig [%s] has no conditions",
			kaiConfig.GetName())), nil
	}

	for _, condition := range kaiConfig.Status.Conditions {
		if condition.Type == string(kaiv1.ConditionTypeAvailable) &&
			condition.Status == metav1.ConditionTrue {
			return operands.ReadyStatus(), nil
		}
	}

	return operands.NotReadyStatus(fmt.Sprintf("kaiConfig [%s] has not reached the Available condition",
		kaiConfig.GetName())), nil
}

func convertVPA(vpa *global.VPASpec) *kaicommon.VPASpec {
	if vpa == nil {
		return nil
	}
	return &kaicommon.VPASpec{
		Enabled:        vpa.Enabled,
		UpdatePolicy:   vpa.UpdatePolicy,
		ResourcePolicy: vpa.ResourcePolicy,
	}
}

func convertPrometheusConfig(runaiConfig *runaiv1.RunaiConfig) *kaiprometheus.Prometheus {
	kaiPrometheusConfig := runaiConfig.Spec.HistoricalUsagePrometheus.DeepCopy()
	if kaiPrometheusConfig == nil {
		kaiPrometheusConfig = &kaiprometheus.Prometheus{}
	}
	if kaiPrometheusConfig.EnablePersistentStorage == nil {
		kaiPrometheusConfig.EnablePersistentStorage = ptr.To(false)
	}
	return kaiPrometheusConfig
}
