/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest";
import { thresholdMetricUtils } from "./threshold-metric.util";
import { DEFAULT_THRESHOLD_VALUE, thresholdMetricOptions } from "./threshold-metric-section.model";
import type { AutoScalingRules } from "@/swagger-models/assets-service-client";
import type { AutoScaling } from "@/swagger-models/workloads-client";

describe("threshold-metric-section.util", () => {
  describe("getThresholdMetricDefaultValue", () => {
    it("returns autoscalingPolicyDefaults.metricThreshold if defined", () => {
      const autoscalingPolicyDefaults: AutoScaling = { metricThreshold: 42 };
      const autoscalingRules: AutoScalingRules = {};
      expect(
        thresholdMetricUtils.getThresholdMetricDefaultValue(undefined, autoscalingPolicyDefaults, autoscalingRules),
      ).toBe(42);
    });

    it("returns DEFAULT_THRESHOLD_VALUE if both inputs are undefined", () => {
      expect(thresholdMetricUtils.getThresholdMetricDefaultValue(undefined, undefined, undefined)).toBe(
        DEFAULT_THRESHOLD_VALUE,
      );
    });

    it("returns DEFAULT_THRESHOLD_VALUE if autoscalingPolicyDefaults.metricThreshold is null and rules are missing", () => {
      const autoscalingPolicyDefaults: AutoScaling = { metricThreshold: null };
      expect(thresholdMetricUtils.getThresholdMetricDefaultValue(undefined, autoscalingPolicyDefaults, undefined)).toBe(
        DEFAULT_THRESHOLD_VALUE,
      );
    });

    it("returns autoscalingRules.metricThreshold.min if it is greater than DEFAULT_THRESHOLD_VALUE", () => {
      const autoscalingRules: AutoScalingRules = { metricThreshold: { min: 200 } };
      expect(thresholdMetricUtils.getThresholdMetricDefaultValue(undefined, undefined, autoscalingRules)).toBe(200);
    });

    it("returns DEFAULT_THRESHOLD_VALUE if autoscalingRules.metricThreshold.min is less than DEFAULT_THRESHOLD_VALUE", () => {
      const autoscalingRules: AutoScalingRules = { metricThreshold: { min: 50 } };
      expect(thresholdMetricUtils.getThresholdMetricDefaultValue(undefined, undefined, autoscalingRules)).toBe(
        DEFAULT_THRESHOLD_VALUE,
      );
    });

    it("returns DEFAULT_THRESHOLD_VALUE if autoscalingRules.metricThreshold.min is undefined", () => {
      const autoscalingRules: AutoScalingRules = { metricThreshold: {} };
      expect(thresholdMetricUtils.getThresholdMetricDefaultValue(undefined, undefined, autoscalingRules)).toBe(
        DEFAULT_THRESHOLD_VALUE,
      );
    });

    it("returns DEFAULT_THRESHOLD_VALUE if autoscalingRules.metricThreshold is null", () => {
      const autoscalingRules: AutoScalingRules = { metricThreshold: null };
      expect(thresholdMetricUtils.getThresholdMetricDefaultValue(undefined, undefined, autoscalingRules)).toBe(
        DEFAULT_THRESHOLD_VALUE,
      );
    });

    it("returns metricThreshold if provided (number)", () => {
      expect(thresholdMetricUtils.getThresholdMetricDefaultValue(123, undefined, undefined)).toBe(123);
    });

    it("returns metricThreshold if provided (zero)", () => {
      expect(thresholdMetricUtils.getThresholdMetricDefaultValue(0, undefined, undefined)).toBe(0);
    });

    it("returns metricThreshold if provided (negative value)", () => {
      expect(thresholdMetricUtils.getThresholdMetricDefaultValue(-10, undefined, undefined)).toBe(-10);
    });
  });

  describe("getDefaultMetricOption", () => {
    it("returns metric if provided", () => {
      expect(thresholdMetricUtils.getDefaultMetricOption("custom-metric", undefined, undefined)).toBe("custom-metric");
    });

    it("returns autoscalingPolicyDefaults.metric if metric is not provided", () => {
      const autoscalingPolicyDefaults: AutoScaling = { metric: "default-metric" };
      expect(thresholdMetricUtils.getDefaultMetricOption(undefined, autoscalingPolicyDefaults, undefined)).toBe(
        "default-metric",
      );
    });

    it("returns autoscalingRules.metric.options[0].value if metric and autoscalingPolicyDefaults.metric are not provided", () => {
      const autoscalingRules: AutoScalingRules = {
        metric: {
          options: [{ value: "option-metric" }, { value: "other-metric" }],
        },
      };
      expect(thresholdMetricUtils.getDefaultMetricOption(undefined, undefined, autoscalingRules)).toBe("option-metric");
    });

    it("returns thresholdMetricOptions[0].id if nothing else is provided", () => {
      expect(thresholdMetricUtils.getDefaultMetricOption(undefined, undefined, undefined)).toBe(
        thresholdMetricOptions[0].id,
      );
    });

    it("returns thresholdMetricOptions[0].id if autoscalingRules.metric.options is empty or null", () => {
      expect(thresholdMetricUtils.getDefaultMetricOption(undefined, undefined, { metric: { options: [] } })).toBe(
        thresholdMetricOptions[0].id,
      );
      expect(thresholdMetricUtils.getDefaultMetricOption(undefined, undefined, { metric: { options: null } })).toBe(
        thresholdMetricOptions[0].id,
      );
      expect(thresholdMetricUtils.getDefaultMetricOption(undefined, undefined, { metric: null })).toBe(
        thresholdMetricOptions[0].id,
      );
    });
  });
});
