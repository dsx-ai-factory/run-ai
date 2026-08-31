/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package knative

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/consts"
	controller_utils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// GetKnativeStatus returns the aggregated PublicStatus for knative and its sub-components
// (hpa, kourier, knative controller).
func GetKnativeStatus(ctx context.Context, kbClient client.Client) runv1.PublicStatus {
	hpaStatus := controller_utils.GetDependencyStatus(ctx, kbClient, "autoscaler", "hpa", "", nil, nil)
	kourierStatus := controller_utils.GetDependencyStatus(ctx, kbClient, "kourier", "gateway", "", nil, nil)
	knativeStatus := controller_utils.GetDependencyStatus(ctx, kbClient, "controller", "", consts.KnativeNamespace, nil, nil)
	if knativeStatus.Available != nil && *knativeStatus.Available {
		setKnativeConfigStatus(ctx, kbClient, &knativeStatus)
	}

	// Add external configuration status to kourier
	setKnativeExternalConfigStatus(ctx, kbClient, &kourierStatus)

	return runv1.PublicStatus{
		Available: knativeStatus.Available,
		Components: map[string]runv1.PublicStatus{
			"knative": knativeStatus,
			"hpa":     hpaStatus,
			"kourier": kourierStatus,
		},
	}
}

func setKnativeConfigStatus(ctx context.Context, kbClient client.Client, knativeStatus *runv1.PublicStatus) {
	logger := log.FromContext(ctx)

	featuresCM := &corev1.ConfigMap{}
	err := kbClient.Get(ctx, types.NamespacedName{
		Name:      consts.KnativeFeaturesConfigMapName,
		Namespace: consts.KnativeNamespace,
	}, featuresCM)
	if err != nil {
		logger.Error(err, "Failed to get knative features ConfigMap")
		knativeStatus.Available = ptr.To(false)
		knativeStatus.Reason = fmt.Sprintf("Failed to get knative %s config-map from %s namespace", consts.KnativeFeaturesConfigMapName, consts.KnativeNamespace)
		return
	}

	// Validate all features from the features configmap
	for feature, config := range consts.KnativeFeaturesConfigMap {
		validateKnativeFeatureEnabled(featuresCM, knativeStatus, feature, config, consts.KnativeFeaturesConfigMapName)
	}

	autoScalerCM := &corev1.ConfigMap{}
	err = kbClient.Get(ctx, types.NamespacedName{
		Name:      consts.KnativeAutoScalerConfigMapName,
		Namespace: consts.KnativeNamespace,
	}, autoScalerCM)
	if err != nil {
		logger.Error(err, "Failed to get knative auto scaler ConfigMap")
		knativeStatus.Available = ptr.To(false)
		knativeStatus.Reason = fmt.Sprintf("Failed to get knative %s config-map from %s namespace", consts.KnativeAutoScalerConfigMapName, consts.KnativeNamespace)
		return
	}

	// Validate all features from the autoscaler configmap
	for feature, config := range consts.KnativeAutoscalerConfigMap {
		validateKnativeFeatureEnabled(autoScalerCM, knativeStatus, feature, config, consts.KnativeAutoScalerConfigMapName)
	}
}

func validateKnativeFeatureEnabled(cm *corev1.ConfigMap, knativeStatus *runv1.PublicStatus, feature string, config consts.KnativeFeatureConfig, cmName string) {
	value, exists := cm.Data[feature]
	defaultValueAsExpected := config.DefaultValue != nil && *config.DefaultValue == config.ExpectedValue
	featureAsExpected := (exists && value == config.ExpectedValue) || (!exists && defaultValueAsExpected)

	// if the feature is required and not enabled, set the general knative status to unavailable
	if config.Required && !featureAsExpected {
		knativeStatus.Available = ptr.To(false)
		knativeStatus.Reason = fmt.Sprintf("Feature %s is not enabled in knative %s config-map from %s namespace, actual value <%s> expected value <%s>", feature, cmName, consts.KnativeNamespace, value, config.ExpectedValue)
	}

	if knativeStatus.Components == nil {
		knativeStatus.Components = map[string]runv1.PublicStatus{}
	}

	var reason string
	if !featureAsExpected {
		reason = fmt.Sprintf("Feature %s in knative %s config-map from %s namespace, has actual value <%s> expected value <%s>", feature, cmName, consts.KnativeNamespace, value, config.ExpectedValue)
	}

	// set the status for the specific feature
	knativeStatus.Components[feature] = runv1.PublicStatus{
		Available: ptr.To(featureAsExpected),
		Reason:    reason,
	}
}

func setKnativeExternalConfigStatus(ctx context.Context, kbClient client.Client, kourierStatus *runv1.PublicStatus) {
	// Only check external configuration if kourier is available
	if kourierStatus.Available == nil || !*kourierStatus.Available {
		return
	}

	logger := log.FromContext(ctx)

	if kourierStatus.Components == nil {
		kourierStatus.Components = map[string]runv1.PublicStatus{}
	}

	// Check domain configuration
	configMap := &corev1.ConfigMap{}
	err := kbClient.Get(ctx, types.NamespacedName{
		Name:      consts.KnativeDomainConfigMapName,
		Namespace: consts.KnativeNamespace,
	}, configMap)

	var status runv1.PublicStatus
	if err != nil {
		if kerrors.IsNotFound(err) {
			status = runv1.PublicStatus{
				Available: runv1.AvailableStatus(false),
				Reason:    fmt.Sprintf("Knative '%s' config-map not found in %s namespace", consts.KnativeDomainConfigMapName, consts.KnativeNamespace),
			}
		} else {
			logger.Error(err, "Failed to get knative domain config ConfigMap")
			status = runv1.PublicStatus{
				Available: runv1.AvailableStatus(false),
				Reason:    fmt.Sprintf("Failed to get knative '%s' config-map from %s namespace", consts.KnativeDomainConfigMapName, consts.KnativeNamespace),
			}
		}
	} else {
		// ConfigMap exists, validate its content
		status = validateKnativeDomainConfig(configMap)
	}

	// Add the status for domain configuration
	kourierStatus.Components["external-config-domain"] = status
}

func validateKnativeDomainConfig(configMap *corev1.ConfigMap) runv1.PublicStatus {
	// Check if domain configuration is properly set
	if len(configMap.Data) == 0 {
		return runv1.PublicStatus{
			Available: runv1.AvailableStatus(false),
			Reason:    "Domain configuration is empty - external access may not work properly",
		}
	}

	// Count non-example keys to check if there's actual configuration
	hasRealConfig := false
	for key := range configMap.Data {
		// Skip the _example key as it contains documentation/examples
		if key != "_example" {
			hasRealConfig = true
			break
		}
	}

	// If we only have _example data, treat it as empty configuration
	if !hasRealConfig {
		return runv1.PublicStatus{
			Available: runv1.AvailableStatus(false),
			Reason:    "Domain configuration is empty - external access may not work properly",
		}
	}

	return runv1.PublicStatus{
		Available: runv1.AvailableStatus(true),
	}
}
