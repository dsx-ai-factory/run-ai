/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validator

import (
	"context"
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"

	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/builders"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"

	"k8s.io/utils/ptr"
)

type InferenceParamsValidator struct{}

func NewInferenceParamsValidator() *InferenceParamsValidator {
	return &InferenceParamsValidator{}
}

func (self *InferenceParamsValidator) Validate(ctx context.Context, workload *jobs.JobWorkload) error {
	// validate that we got one and only one serving port
	_, err := builders.GetServingPort(ctx, &workload.Spec.NetworkingFields, constants.KindInferenceWorkload)
	if err != nil {
		return err
	}

	err = ValidateBasicAutoScalingFields(workload.Spec.InferenceFields)
	if err != nil {
		return err
	}

	err = ValidateServingConfigurationFields(workload.Spec.InferenceFields)
	if err != nil {
		return err
	}

	err = ValidatePvcs(workload.Spec.Pvcs)
	if err != nil {
		return err
	}

	return nil
}

func isKpaMetric(metric *field_types.StringField, metricName *field_types.StringField) bool {
	if metric == nil && metricName == nil {
		return true // the default metric in knative is concurrency, which is kpa
	}

	if metric != nil && (metric.Deref() == "concurrency" || metric.Deref() == "throughput") {
		return true
	}

	return false
}

func ValidateServingConfigurationFields(inferenceFields field_groups.InferenceFields) error {
	servingConfigurationFields := inferenceFields.ServingConfigurationFields

	initializationTimeoutSeconds := servingConfigurationFields.InitializationTimeoutSeconds
	if initializationTimeoutSeconds != nil {
		initializationTimeoutSecondsValue := ptr.Deref(initializationTimeoutSeconds.Value, 0)

		if initializationTimeoutSecondsValue < 1 {
			return fmt.Errorf("initializationTimeoutSeconds must be greater than 0")
		}
	}

	requestTimeoutSeconds := servingConfigurationFields.RequestTimeoutSeconds
	if requestTimeoutSeconds != nil {
		requestTimeoutSecondsValue := ptr.Deref(requestTimeoutSeconds.Value, 0)

		if requestTimeoutSecondsValue < 1 {
			return fmt.Errorf("requestTimeoutSeconds must be greater than 0")
		}
	}

	return nil
}

func ValidateBasicAutoScalingFields(inferenceFields field_groups.InferenceFields) error {
	autoScalingFields := inferenceFields.AutoScalingFields

	// Validate metric
	metric := autoScalingFields.Metric
	if metric != nil {
		if _, ok := builders.MetricNames[metric.Deref()]; !ok {
			return fmt.Errorf("metric '%s' is not supported", metric.Deref())
		}
	}

	isKpa := isKpaMetric(autoScalingFields.Metric, autoScalingFields.MetricName)

	// Validate minScale and maxScale
	minScale := autoScalingFields.MinScale
	maxScale := autoScalingFields.MaxScale
	if minScale != nil {
		minScaleValue := ptr.Deref(minScale.Value, 0)
		if minScaleValue < 0 {
			return fmt.Errorf("minScale must be non-negative")
		}

		if minScaleValue == 0 && !isKpa {
			return fmt.Errorf("minScale of 0 is only allowed when using the concurrency or throughput metrics")
		}

		if maxScale != nil && ptr.Deref(maxScale.Value, 0) < minScaleValue {
			return fmt.Errorf("maxScale must be greater than or equal to minScale")
		}
	}

	// Validate initialScale and activationScale
	initialScale := autoScalingFields.InitialScale
	if initialScale != nil {
		initialScaleValue := ptr.Deref(initialScale.Value, 0)

		if initialScaleValue < 0 {
			return fmt.Errorf("initialScale must be non-negative")
		}

		if minScale != nil && initialScaleValue < ptr.Deref(minScale.Value, 0) {
			return fmt.Errorf("initialScale must be greater than or equal to minScale")
		}

		if maxScale != nil && initialScaleValue > ptr.Deref(maxScale.Value, 0) {
			return fmt.Errorf("initialScale must be less than or equal to maxScale")
		}
	}

	activationScale := autoScalingFields.ActivationScale
	if activationScale != nil {
		activationScaleValue := ptr.Deref(activationScale.Value, 0)

		if activationScaleValue < 1 {
			return fmt.Errorf("activationScale must be greater than 0")
		}

		if minScale != nil && activationScaleValue < ptr.Deref(minScale.Value, 0) {
			return fmt.Errorf("activationScale must be greater than or equal to minScale")
		}

		if maxScale != nil && activationScaleValue > ptr.Deref(maxScale.Value, 0) {
			return fmt.Errorf("activationScale must be less than or equal to maxScale")
		}
	}

	// Validate targetUtilizationPercentage
	targetUtilizationPercentage := autoScalingFields.TargetUtilizationPercentage
	if targetUtilizationPercentage != nil {
		targetUtilizationPercentageValue := targetUtilizationPercentage.Deref()

		if targetUtilizationPercentageValue < 1 || targetUtilizationPercentageValue > 100 {
			return fmt.Errorf("targetUtilizationPercentage must be between 1 and 100")
		}

		if !isKpa {
			return fmt.Errorf("targetUtilizationPercentage is only allowed when using the concurrency or throughput metrics")
		}
	}

	// Validate concurrencyHardLimit
	concurrencyHardLimit := autoScalingFields.ConcurrencyHardLimit
	if concurrencyHardLimit != nil {
		concurrencyHardLimitValue := ptr.Deref(concurrencyHardLimit.Value, 0)

		if concurrencyHardLimitValue < 0 {
			return fmt.Errorf("concurrencyHardLimit must be non-negative")
		}
	}

	// Validate scaleToZeroRetentionSeconds
	scaleToZeroRetentionSeconds := autoScalingFields.ScaleToZeroRetentionSeconds
	if scaleToZeroRetentionSeconds != nil {
		scaleToZeroRetentionSecondsValue := ptr.Deref(scaleToZeroRetentionSeconds.Value, 0)

		if scaleToZeroRetentionSecondsValue < 0 {
			return fmt.Errorf("scaleToZeroRetentionSeconds must be non-negative")
		}

		if scaleToZeroRetentionSecondsValue > 3600 {
			return fmt.Errorf("scaleToZeroRetentionSeconds must be less than or equal to 3600 (1 hour)")
		}

		if minScale == nil || (ptr.Deref(minScale.Value, 0) > 0 && scaleToZeroRetentionSecondsValue > 0) {
			return fmt.Errorf("minScale must be set to 0 if scaleToZeroRetentionSeconds is set")
		}
	}

	// Validate scaleDownDelaySeconds
	scaleDownDelaySeconds := autoScalingFields.ScaleDownDelaySeconds
	if scaleDownDelaySeconds != nil {
		scaleDownDelaySecondsValue := ptr.Deref(scaleDownDelaySeconds.Value, 0)

		if scaleDownDelaySecondsValue < 0 {
			return fmt.Errorf("scaleDownDelaySeconds must be non-negative")
		}

		if scaleDownDelaySecondsValue > 3600 {
			return fmt.Errorf("scaleDownDelaySeconds must be less than or equal to 3600 (1 hour)")
		}
	}

	return nil
}

func ValidatePvcs(pvcs *field_types.PvcsField) error {
	if pvcs == nil || pvcs.Value == nil {
		return nil
	}

	for _, pvc := range *pvcs.Value {
		if pvc.Value != nil && pvc.Value.Ephemeral != nil && *pvc.Value.Ephemeral {
			return fmt.Errorf("ephemeral pvc is not supported")
		}
	}

	return nil
}
