/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	kaicommon "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/common"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clustercommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	outils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	runv1internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
)

const (
	krmConfigCRName = kaiv1alpha1.KRMConfigSingletonName

	availableConditionType = string(kaiv1alpha1.KRMConfigConditionTypeAvailable)
)

// KRMConfigForRunaiConfig builds the KRMConfig singleton KRM's operator reconciles.
// The chart installs KRM in its "external" mode -- neither the deployer hook nor an
// inline render -- so this is the only writer of the object.
func KRMConfigForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runaiv1.RunaiConfig) (client.Object, error) {

	krmConfig := &kaiv1alpha1.KRMConfig{}
	err := k8sReader.Get(ctx, types.NamespacedName{Name: krmConfigCRName}, krmConfig)
	if err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("get KRMConfig %s: %w", krmConfigCRName, err)
	}

	krmConfig.Name = krmConfigCRName
	krmConfig.Spec.Namespace = runaiConfig.GetNamespace()
	krmConfig.Spec.Global = convertKRMGlobal(runaiConfig)

	return krmConfig, nil
}

func convertKRMGlobal(runaiConfig *runaiv1.RunaiConfig) *kaiv1alpha1.GlobalConfig {
	globalConfig := &kaiv1alpha1.GlobalConfig{
		SchedulerName:                 ptr.To(clustercommonconstants.RunaiScheduler),
		QueueLabelKey:                 ptr.To(clustercommonconstants.RunaiQueueLabel),
		NodePoolLabelKey:              ptr.To(clustercommonconstants.NodePoolNameLabel),
		DefaultNodePoolName:           ptr.To(runv1alpha1.DefaultNodePoolName),
		FinalizerDomain:               ptr.To(clustercommonconstants.RunaiOrgDomain),
		NamespaceProjectLabelKey:      ptr.To(clustercommonconstants.RunaiQueueLabel),
		ProjectLabelKey:               ptr.To(clustercommonconstants.RunaiProjectLabel),
		EnforceSchedulerAnnotationKey: ptr.To(clustercommonconstants.EnforceSchedulerAnnotationName),

		ReplicaCount:                      runaiConfig.Spec.Global.ReplicaCount,
		ImagePullSecrets:                  runaiConfig.Spec.Internal.Global.AdditionalImagePullSecrets,
		Tolerations:                       outils.GetTolerations(runaiConfig),
		Affinity:                          outils.GetAffinity(runaiConfig),
		RequireDefaultPodAntiAffinityTerm: ptr.To(runaiConfig.Spec.Global.RequireDefaultPodAntiAffinity),
		SecurityContext:                   (*corev1.SecurityContext)(runaiConfig.Spec.Global.SecurityContext),
		FipsMode:                          convertFipsMode(runaiConfig),
		VPA:                               convertVPA(runaiConfig.Spec.Global.VPA),
		ServiceMonitor: &kaiv1alpha1.ServiceMonitorSpec{
			Enabled:    ptr.To(true),
			Accounting: ptr.To(true),
		},
	}

	// LeaderElection stays unset: KRM turns it on for replicaCount > 1.
	if runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor == runv1internal.KubernetesFlavorOpenShift {
		globalConfig.Openshift = ptr.To(true)
	}

	return globalConfig
}

// convertFipsMode leaves the field unset for an unrecognised value so KRM applies
// its own default, rather than pinning a mode the services cannot honour.
func convertFipsMode(runaiConfig *runaiv1.RunaiConfig) *kaiv1alpha1.FipsMode {
	if runaiConfig.Spec.Internal == nil || runaiConfig.Spec.Internal.Global == nil {
		return nil
	}
	switch kaiv1alpha1.FipsMode(runaiConfig.Spec.Internal.Global.FipsMode) {
	case kaiv1alpha1.FipsModeOff, kaiv1alpha1.FipsModeOn, kaiv1alpha1.FipsModeOnly:
		return ptr.To(kaiv1alpha1.FipsMode(runaiConfig.Spec.Internal.Global.FipsMode))
	}
	return nil
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

func KRMConfigStatus(ctx context.Context, k8sReader client.Reader,
	_ *runaiv1.RunaiConfig) (operands.Status, error) {

	krmConfig := &kaiv1alpha1.KRMConfig{}
	if err := k8sReader.Get(ctx, types.NamespacedName{Name: krmConfigCRName}, krmConfig); err != nil {
		if errors.IsNotFound(err) {
			return operands.NotReadyStatus(fmt.Sprintf("KRMConfig [%s] is missing", krmConfigCRName)), nil
		}
		return operands.NotReadyStatus(), fmt.Errorf("get KRMConfig %s: %w", krmConfigCRName, err)
	}

	available := utils.GetCondition(krmConfig.Status.Conditions, availableConditionType)
	if available == nil {
		return operands.NotReadyStatus("KRMConfig has no available condition"), nil
	}
	if available.Status != metav1.ConditionTrue {
		return operands.NotReadyStatus(fmt.Sprintf("KRMConfig is not available: %s", available.Message)), nil
	}

	return operands.ReadyStatus(), nil
}
