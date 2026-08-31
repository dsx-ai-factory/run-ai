/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { DEFAULT_THRESHOLD_VALUE, thresholdMetricOptions } from "./threshold-metric-section.model";
import type { AutoScalingRules } from "@/swagger-models/assets-service-client";
import type { AutoScaling } from "@/swagger-models/workloads-client";

export const thresholdMetricUtils = {
  getThresholdMetricDefaultValue,
  getDefaultMetricOption,
};

function getThresholdMetricDefaultValue(
  metricThreshold?: number | undefined,
  autoscalingDefaults?: AutoScaling | null,
  autoscalingRules?: AutoScalingRules | null,
): number {
  if (metricThreshold !== undefined) return metricThreshold;

  if (autoscalingDefaults?.metricThreshold !== undefined && autoscalingDefaults?.metricThreshold !== null) {
    return autoscalingDefaults?.metricThreshold;
  }

  let retMetricThreshold = DEFAULT_THRESHOLD_VALUE;
  if (autoscalingRules?.metricThreshold?.min !== undefined && autoscalingRules?.metricThreshold?.min !== null) {
    retMetricThreshold =
      autoscalingRules.metricThreshold.min > retMetricThreshold
        ? autoscalingRules.metricThreshold.min
        : retMetricThreshold;
  }

  return retMetricThreshold;
}

function getDefaultMetricOption(
  metric?: string | undefined,
  autoscalingDefaults?: AutoScaling | null,
  autoscalingRules?: AutoScalingRules | null,
): string {
  if (metric) return metric;

  if (autoscalingDefaults?.metric !== undefined && autoscalingDefaults?.metric !== null) {
    return autoscalingDefaults?.metric;
  }

  const defaultMetricOptionFromRules = autoscalingRules?.metric?.options?.[0]?.value;
  if (defaultMetricOptionFromRules !== undefined && defaultMetricOptionFromRules !== null) {
    return defaultMetricOptionFromRules;
  }

  return thresholdMetricOptions[0].id;
}
