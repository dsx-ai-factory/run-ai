/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount, type VueWrapper } from "@vue/test-utils";
import { describe, it, expect } from "vitest";
import AutoScaleSection from "./auto-scale-section.vue";
import { METRIC_LATENCY } from "./threshold-metric-section/threshold-metric-section.model";
import { Quasar } from "quasar";
import type { AutoScalingRules } from "@/swagger-models/assets-service-client";
import { EAutoScaleSectionEvents, type IAutoScaleData, type IAutoScaleOptions } from "./auto-scale-section.model";
import { errorMessages } from "@/common/error-message.constant";
import { DEFAULT_SCALE_DOWN_TO_ZERO_ID } from "./scale-down-section/scale-down-section.model";

interface IAutoScaleProps {
  autoScaleData: IAutoScaleData;
  disabled: boolean;
  policyRules: AutoScalingRules | null;
  autoScaleOptions: IAutoScaleOptions;
}

describe("auto-scale-section.vue", () => {
  const baseProps: IAutoScaleProps = {
    autoScaleData: {
      minReplicas: 1,
      maxReplicas: 2,
      thresholdMetric: "cpu",
      thresholdValue: 80,
      scaleToZeroRetentionSeconds: 300,
    },
    disabled: false,
    policyRules: null,
    autoScaleOptions: {
      isAutoScaleEnabled: true,
      latencyEnabled: false,
    },
  };

  function getMountedComponent(props = {}): VueWrapper<InstanceType<typeof AutoScaleSection>> {
    return mount(AutoScaleSection, {
      global: {
        plugins: [Quasar],
      },
      props: { ...baseProps, ...props },
    });
  }

  describe("renders sections", () => {
    it("renders ReplicaScalingSection with correct props", () => {
      const wrapper = getMountedComponent();
      const cmp = wrapper.findComponent({ name: "replica-scaling-section" });
      expect(cmp.exists()).toBe(true);
      expect(cmp.props("modelValue")).toEqual({ minReplicas: 1, maxReplicas: 2 });
      expect(cmp.props("disabled")).toBe(false);
      expect(cmp.props("policyRules")).toBe(null);
      expect(cmp.props("isAutoScaleEnabled")).toBe(true);
    });

    it("conditionally renders ThresholdMetricSection when minReplicas is 0 and maxReplicas is 1", () => {
      const wrapper = getMountedComponent();
      expect(wrapper.findComponent({ name: "threshold-metric-section" }).exists()).toBe(true);
      // Should not render if shouldShowReplicaCondition returns false
      const wrapper2 = getMountedComponent({
        autoScaleData: { ...baseProps.autoScaleData, minReplicas: 0, maxReplicas: 1 },
      });
      expect(wrapper2.findComponent({ name: "threshold-metric-section" }).exists()).toBe(false);
    });

    it("conditionally renders ThresholdMetricSection when minReplicas is undefined and maxReplicas is 2", () => {
      const wrapper = getMountedComponent();
      expect(wrapper.findComponent({ name: "threshold-metric-section" }).exists()).toBe(true);
      // Should not render if shouldShowReplicaCondition returns false
      const wrapper2 = getMountedComponent({
        autoScaleData: { ...baseProps.autoScaleData, minReplicas: undefined, maxReplicas: 2 },
      });
      expect(wrapper2.findComponent({ name: "threshold-metric-section" }).exists()).toBe(false);
    });

    it("conditionally renders ThresholdMetricSection when minReplicas is 2 and maxReplicas is undefined", () => {
      const wrapper = getMountedComponent();
      expect(wrapper.findComponent({ name: "threshold-metric-section" }).exists()).toBe(true);
      // Should not render if shouldShowReplicaCondition returns false
      const wrapper2 = getMountedComponent({
        autoScaleData: { ...baseProps.autoScaleData, minReplicas: 2, maxReplicas: undefined },
      });
      expect(wrapper2.findComponent({ name: "threshold-metric-section" }).exists()).toBe(false);
    });

    it("conditionally renders ThresholdMetricSection when minReplicas and maxReplicas are equal", () => {
      const wrapper = getMountedComponent();
      expect(wrapper.findComponent({ name: "threshold-metric-section" }).exists()).toBe(true);
      // Should not render if shouldShowReplicaCondition returns false
      const wrapper2 = getMountedComponent({
        autoScaleData: { ...baseProps.autoScaleData, minReplicas: 2, maxReplicas: 2 },
      });
      expect(wrapper2.findComponent({ name: "threshold-metric-section" }).exists()).toBe(false);
    });

    it("conditionally renders ScaleDownSection", () => {
      const wrapper = getMountedComponent();
      expect(wrapper.findComponent({ name: "scale-down-section" }).exists()).toBe(true);
      // Should not render if thresholdMetric is METRIC_LATENCY
      const wrapper2 = getMountedComponent({
        autoScaleData: { ...baseProps.autoScaleData, thresholdMetric: METRIC_LATENCY },
      });
      expect(wrapper2.findComponent({ name: "scale-down-section" }).exists()).toBe(false);
    });
  });

  describe("computed properties", () => {
    it("showScaleDownToZeroValues", () => {
      const wrapper = getMountedComponent();
      expect(wrapper.vm.showScaleDownToZeroValues).toBe(true);
      const wrapper2 = getMountedComponent({
        autoScaleData: { ...baseProps.autoScaleData, thresholdMetric: METRIC_LATENCY },
      });
      expect(wrapper2.vm.showScaleDownToZeroValues).toBe(false);
    });
    it("showReplicaCondition", () => {
      const wrapper = getMountedComponent();
      expect(wrapper.vm.showReplicaCondition).toBe(true);
      // Provide all required properties for IAutoScaleData
      const wrapper2 = getMountedComponent({
        autoScaleData: {
          minReplicas: 0,
          maxReplicas: 1,
          thresholdMetric: "cpu",
          thresholdValue: 80,
          scaleToZeroRetentionSeconds: 300,
        },
      });
      expect(wrapper2.vm.showReplicaCondition).toBe(false);
      const wrapper3 = getMountedComponent({
        autoScaleData: {
          minReplicas: undefined,
          maxReplicas: 2,
          thresholdMetric: "cpu",
          thresholdValue: 80,
          scaleToZeroRetentionSeconds: 300,
        },
      });
      expect(wrapper3.vm.showReplicaCondition).toBe(false);
      const wrapper4 = getMountedComponent({
        autoScaleData: {
          minReplicas: 1,
          maxReplicas: undefined,
          thresholdMetric: "cpu",
          thresholdValue: 80,
          scaleToZeroRetentionSeconds: 300,
        },
      });
      expect(wrapper4.vm.showReplicaCondition).toBe(false);
    });
    it("minReplicasAllowed", () => {
      const wrapper = getMountedComponent();
      expect(wrapper.vm.minReplicasAllowed).toBe(0);
      const wrapper2 = getMountedComponent({
        autoScaleOptions: { ...baseProps.autoScaleOptions, isAutoScaleEnabled: false },
      });
      expect(wrapper2.vm.minReplicasAllowed).toBe(1);
    });
    it("minValueForMaxReplicas", () => {
      const wrapper = getMountedComponent();
      expect(wrapper.vm.minValueForMaxReplicas).toBe(1);
      const wrapper2 = getMountedComponent({ autoScaleData: { ...baseProps.autoScaleData, minReplicas: 3 } });
      expect(wrapper2.vm.minValueForMaxReplicas).toBe(3);
    });
  });

  describe("methods", () => {
    it("canChangeScaleDown", () => {
      const wrapper = getMountedComponent();
      expect(wrapper.vm.canChangeScaleDown(baseProps.autoScaleData)).toBe(false);
      const wrapper2 = getMountedComponent({
        autoScaleData: { ...baseProps.autoScaleData, minReplicas: 0 },
        autoScaleOptions: { ...baseProps.autoScaleOptions, isAutoScaleEnabled: true },
      });
      expect(wrapper2.vm.canChangeScaleDown(wrapper2.props("autoScaleData"))).toBe(true);
    });
    it("shouldShowReplicaCondition", () => {
      const wrapper = getMountedComponent();
      expect(wrapper.vm.shouldShowReplicaCondition(baseProps.autoScaleData)).toBe(true);
      // Provide all required properties for IAutoScaleData
      expect(
        wrapper.vm.shouldShowReplicaCondition({
          minReplicas: 0,
          maxReplicas: 1,
          thresholdMetric: "cpu",
          thresholdValue: 80,
          scaleToZeroRetentionSeconds: 300,
        }),
      ).toBe(false);
      expect(
        wrapper.vm.shouldShowReplicaCondition({
          minReplicas: undefined,
          maxReplicas: 2,
          thresholdMetric: "cpu",
          thresholdValue: 80,
          scaleToZeroRetentionSeconds: 300,
        }),
      ).toBe(false);
      expect(
        wrapper.vm.shouldShowReplicaCondition({
          minReplicas: 1,
          maxReplicas: undefined,
          thresholdMetric: "cpu",
          thresholdValue: 80,
          scaleToZeroRetentionSeconds: 300,
        }),
      ).toBe(false);
    });
    it("isMaxMorThanMin", () => {
      const wrapper = getMountedComponent();
      expect(wrapper.vm.isMaxMorThanMin()).toBe(true);
      const wrapper2 = getMountedComponent({
        autoScaleData: { ...baseProps.autoScaleData, minReplicas: 3, maxReplicas: 2 },
      });
      expect(wrapper2.vm.isMaxMorThanMin()).toBe(errorMessages.EQUAL_TO_OR_HIGHER_THAN_THE_MINIMUM);
    });
    it("isAllowScaleToZero", () => {
      const wrapper = getMountedComponent();
      expect(wrapper.vm.isAllowScaleToZero(0)).toBe(true);
      expect(wrapper.vm.isAllowScaleToZero(1)).toBe(false);
    });

    it("onScaleDownChanged emits correct event", () => {
      const wrapper = getMountedComponent();
      wrapper.vm.onScaleDownChanged({ scaleToZeroRetentionSeconds: 123 });
      const emits = wrapper.emitted()[EAutoScaleSectionEvents.OnAutoScaleChanged];
      const expected = [
        [
          {
            minReplicas: 1,
            maxReplicas: 2,
            thresholdMetric: "cpu",
            thresholdValue: 80,
            scaleToZeroRetentionSeconds: 123,
          },
        ],
      ];
      expect(emits).toEqual(expected);
    });

    describe("onReplicaScalingChanged", () => {
      it("onReplicaScalingChanged emits correct event", () => {
        const wrapper = getMountedComponent();
        wrapper.vm.onReplicaScalingChanged({ minReplicas: 2, maxReplicas: 5 });
        const emits = wrapper.emitted()[EAutoScaleSectionEvents.OnAutoScaleChanged];
        const expected = [
          [
            {
              minReplicas: 2,
              maxReplicas: 5,
              thresholdMetric: "cpu",
              thresholdValue: 80,
              scaleToZeroRetentionSeconds: undefined,
            },
          ],
        ];
        expect(emits).toEqual(expected);
      });
      it("onThresholdMetricChanged emits correct event and updates for METRIC_LATENCY", () => {
        const wrapper = getMountedComponent();
        wrapper.vm.onThresholdMetricChanged({ thresholdMetric: METRIC_LATENCY, thresholdValue: 10 });
        const emits = wrapper.emitted()[EAutoScaleSectionEvents.OnAutoScaleChanged] as unknown[][];
        const expected = [
          [
            {
              minReplicas: 1,
              maxReplicas: 2,
              thresholdMetric: METRIC_LATENCY,
              thresholdValue: 10,
              scaleToZeroRetentionSeconds: undefined,
            },
          ],
        ];
        expect(emits).toEqual(expected);
      });

      it("uses policyDefaults.scaleToZeroRetentionSeconds when scale down is enabled", () => {
        const policyDefaults = { scaleToZeroRetentionSeconds: 1234 };
        const wrapper = getMountedComponent({
          policyDefaults,
          autoScaleData: { ...baseProps.autoScaleData, minReplicas: 0 },
          autoScaleOptions: { ...baseProps.autoScaleOptions, isAutoScaleEnabled: true },
        });
        wrapper.vm.onReplicaScalingChanged({ minReplicas: 0, maxReplicas: 2 });
        const emits = wrapper.emitted()[EAutoScaleSectionEvents.OnAutoScaleChanged] as unknown[][];
        const expected = [
          [
            {
              minReplicas: 0,
              maxReplicas: 2,
              thresholdMetric: "cpu",
              thresholdValue: 80,
              scaleToZeroRetentionSeconds: 1234,
            },
          ],
        ];
        expect(emits).toEqual(expected);
      });

      it("uses default value if policyDefaults.scaleToZeroRetentionSeconds is not set", () => {
        const wrapper = getMountedComponent({
          autoScaleData: { ...baseProps.autoScaleData, minReplicas: 0 },
          autoScaleOptions: { ...baseProps.autoScaleOptions, isAutoScaleEnabled: true },
        });
        wrapper.vm.onReplicaScalingChanged({ minReplicas: 0, maxReplicas: 2 });
        const emits = wrapper.emitted()[EAutoScaleSectionEvents.OnAutoScaleChanged] as unknown[][];
        const expected = [
          [
            {
              minReplicas: 0,
              maxReplicas: 2,
              thresholdMetric: "cpu",
              thresholdValue: 80,
              scaleToZeroRetentionSeconds: Number(DEFAULT_SCALE_DOWN_TO_ZERO_ID),
            },
          ],
        ];
        expect(emits).toEqual(expected);
      });

      it("sets scaleToZeroRetentionSeconds to undefined if canChangeScaleDown is false", () => {
        const wrapper = getMountedComponent({
          autoScaleData: { ...baseProps.autoScaleData, minReplicas: 1 },
          autoScaleOptions: { ...baseProps.autoScaleOptions, isAutoScaleEnabled: true },
        });
        wrapper.vm.onReplicaScalingChanged({ minReplicas: 1, maxReplicas: 2 });
        const emits = wrapper.emitted()[EAutoScaleSectionEvents.OnAutoScaleChanged] as unknown[][];
        const expected = [
          [
            {
              minReplicas: 1,
              maxReplicas: 2,
              thresholdMetric: "cpu",
              thresholdValue: 80,
              scaleToZeroRetentionSeconds: undefined,
            },
          ],
        ];
        expect(emits).toEqual(expected);
      });

      it("sets thresholdMetric and thresholdValue from policyDefaults if autoScaleData values are undefined", () => {
        const policyDefaults = { metric: "concurrency", metricThreshold: 55 };
        const wrapper = getMountedComponent({
          policyDefaults,
          autoScaleData: { minReplicas: 1, maxReplicas: 2, thresholdMetric: undefined, thresholdValue: undefined },
        });
        wrapper.vm.onReplicaScalingChanged({ minReplicas: 1, maxReplicas: 2 });
        const emits = wrapper.emitted()[EAutoScaleSectionEvents.OnAutoScaleChanged] as unknown[][];
        const expected = [
          [
            {
              minReplicas: 1,
              maxReplicas: 2,
              thresholdMetric: "concurrency",
              thresholdValue: 55,
              scaleToZeroRetentionSeconds: undefined,
            },
          ],
        ];
        expect(emits).toEqual(expected);
      });

      it("does not override thresholdMetric and thresholdValue if autoScaleData values are present", () => {
        const policyDefaults = { metric: "concurrency", metricThreshold: 55 };
        const wrapper = getMountedComponent({
          policyDefaults,
          autoScaleData: { minReplicas: 1, maxReplicas: 2, thresholdMetric: "cpu", thresholdValue: 80 },
        });
        wrapper.vm.onReplicaScalingChanged({ minReplicas: 1, maxReplicas: 2 });
        const emits = wrapper.emitted()[EAutoScaleSectionEvents.OnAutoScaleChanged] as unknown[][];
        const expected = [
          [
            {
              minReplicas: 1,
              maxReplicas: 2,
              thresholdMetric: "cpu",
              thresholdValue: 80,
              scaleToZeroRetentionSeconds: undefined,
            },
          ],
        ];
        expect(emits).toEqual(expected);
      });
    });

    describe("getThresholdValues", () => {
      it("returns thresholdMetric and thresholdValue from autoScaleData when shouldShowReplicaCondition is true", () => {
        const wrapper = getMountedComponent();
        const autoScaleData = {
          minReplicas: 1,
          maxReplicas: 2,
          thresholdMetric: "cpu",
          thresholdValue: 80,
          scaleToZeroRetentionSeconds: 300,
        };
        const result = wrapper.vm.getThresholdValues(autoScaleData);
        expect(result).toEqual({ thresholdMetric: "cpu", thresholdValue: 80 });
      });

      it("returns default values when thresholdMetric and thresholdValue are undefined and shouldShowReplicaCondition is true", () => {
        const wrapper = getMountedComponent();
        const autoScaleData = {
          minReplicas: 1,
          maxReplicas: 2,
          thresholdMetric: undefined,
          thresholdValue: undefined,
          scaleToZeroRetentionSeconds: 300,
        };
        const result = wrapper.vm.getThresholdValues(autoScaleData);
        expect(result).toEqual({ thresholdMetric: "throughput", thresholdValue: 100 });
      });

      it("returns policyDefaults values when present and autoScaleData values are undefined", () => {
        const policyDefaults = { metric: "concurrency", metricThreshold: 55 };
        const wrapper = getMountedComponent({ policyDefaults });
        const autoScaleData = {
          minReplicas: 1,
          maxReplicas: 2,
          thresholdMetric: undefined,
          thresholdValue: undefined,
          scaleToZeroRetentionSeconds: 300,
        };
        const result = wrapper.vm.getThresholdValues(autoScaleData);
        expect(result).toEqual({ thresholdMetric: "concurrency", thresholdValue: 55 });
      });

      it("returns undefined for both values when shouldShowReplicaCondition is false", () => {
        const wrapper = getMountedComponent();
        const autoScaleData = {
          minReplicas: 0,
          maxReplicas: 1,
          thresholdMetric: "cpu",
          thresholdValue: 80,
          scaleToZeroRetentionSeconds: 300,
        };
        const result = wrapper.vm.getThresholdValues(autoScaleData);
        expect(result).toEqual({ thresholdMetric: undefined, thresholdValue: undefined });
      });

      it("returns undefined for both values when minReplicas or maxReplicas is undefined", () => {
        const wrapper = getMountedComponent();
        const autoScaleData1 = {
          minReplicas: undefined,
          maxReplicas: 2,
          thresholdMetric: "cpu",
          thresholdValue: 80,
          scaleToZeroRetentionSeconds: 300,
        };
        const autoScaleData2 = {
          minReplicas: 1,
          maxReplicas: undefined,
          thresholdMetric: "cpu",
          thresholdValue: 80,
          scaleToZeroRetentionSeconds: 300,
        };
        expect(wrapper.vm.getThresholdValues(autoScaleData1)).toEqual({
          thresholdMetric: undefined,
          thresholdValue: undefined,
        });
        expect(wrapper.vm.getThresholdValues(autoScaleData2)).toEqual({
          thresholdMetric: undefined,
          thresholdValue: undefined,
        });
      });

      it("returns values from autoScaleData when policyRules are present", () => {
        const policyRules = {
          metric: {
            options: [
              { value: "cpu", displayed: "CPU" },
              { value: "mem", displayed: "Memory" },
            ],
            canEdit: true,
            required: true,
          },
          metricThreshold: { min: 10, max: 90, step: 10, required: true, canEdit: true },
        };
        const wrapper = getMountedComponent({ policyRules });
        const autoScaleData = {
          minReplicas: 1,
          maxReplicas: 2,
          thresholdMetric: "cpu",
          thresholdValue: 80,
          scaleToZeroRetentionSeconds: 300,
        };
        const result = wrapper.vm.getThresholdValues(autoScaleData);
        expect(result).toEqual({ thresholdMetric: "cpu", thresholdValue: 80 });
      });

      it("returns default values when autoScaleData values are undefined and policyRules are present", () => {
        const policyRules = {
          metric: {
            options: [
              { value: "cpu", displayed: "CPU" },
              { value: "mem", displayed: "Memory" },
            ],
            canEdit: true,
            required: true,
          },
          metricThreshold: { min: 10, max: 90, step: 10, required: true, canEdit: true },
        };
        const wrapper = getMountedComponent({ policyRules });
        const autoScaleData = {
          minReplicas: 1,
          maxReplicas: 2,
          thresholdMetric: undefined,
          thresholdValue: undefined,
          scaleToZeroRetentionSeconds: 300,
        };
        // Should use hardcoded defaults, not policyRules
        const result = wrapper.vm.getThresholdValues(autoScaleData);
        expect(result).toEqual({ thresholdMetric: "cpu", thresholdValue: 100 });
      });

      it("returns policyDefaults values when present, even if policyRules are present", () => {
        const policyDefaults = { metric: "concurrency", metricThreshold: 55 };
        const policyRules = {
          metric: {
            options: [
              { value: "cpu", displayed: "CPU" },
              { value: "mem", displayed: "Memory" },
            ],
            canEdit: true,
            required: true,
          },
          metricThreshold: { min: 10, max: 90, step: 10, required: true, canEdit: true },
        };
        const wrapper = getMountedComponent({ policyDefaults, policyRules });
        const autoScaleData = {
          minReplicas: 1,
          maxReplicas: 2,
          thresholdMetric: undefined,
          thresholdValue: undefined,
          scaleToZeroRetentionSeconds: 300,
        };
        // Should use policyDefaults, not policyRules
        const result = wrapper.vm.getThresholdValues(autoScaleData);
        expect(result).toEqual({ thresholdMetric: "concurrency", thresholdValue: 55 });
      });

      it("returns policyDefaults.metric and default thresholdValue if only metric is set", () => {
        const policyDefaults = { metric: "custom-metric" };
        const wrapper = getMountedComponent({ policyDefaults });
        const autoScaleData = {
          minReplicas: 1,
          maxReplicas: 2,
          thresholdMetric: undefined,
          thresholdValue: undefined,
          scaleToZeroRetentionSeconds: 300,
        };
        const result = wrapper.vm.getThresholdValues(autoScaleData);
        expect(result).toEqual({ thresholdMetric: "custom-metric", thresholdValue: 100 });
      });

      it("returns default metric and policyDefaults.metricThreshold if only metricThreshold is set", () => {
        const policyDefaults = { metricThreshold: 77 };
        const wrapper = getMountedComponent({ policyDefaults });
        const autoScaleData = {
          minReplicas: 1,
          maxReplicas: 2,
          thresholdMetric: undefined,
          thresholdValue: undefined,
          scaleToZeroRetentionSeconds: 300,
        };
        const result = wrapper.vm.getThresholdValues(autoScaleData);
        expect(result).toEqual({ thresholdMetric: "throughput", thresholdValue: 77 });
      });

      it("returns hardcoded defaults if policyDefaults.metric and metricThreshold are falsy", () => {
        const policyDefaults = { metric: undefined, metricThreshold: null };
        const wrapper = getMountedComponent({ policyDefaults });
        const autoScaleData = {
          minReplicas: 1,
          maxReplicas: 2,
          thresholdMetric: undefined,
          thresholdValue: undefined,
          scaleToZeroRetentionSeconds: 300,
        };
        const result = wrapper.vm.getThresholdValues(autoScaleData);
        expect(result).toEqual({ thresholdMetric: "throughput", thresholdValue: 100 });
      });

      it("returns autoScaleData values even if policyDefaults is present", () => {
        const policyDefaults = { metric: "concurrency", metricThreshold: 55 };
        const wrapper = getMountedComponent({ policyDefaults });
        const autoScaleData = {
          minReplicas: 1,
          maxReplicas: 2,
          thresholdMetric: "cpu",
          thresholdValue: 80,
          scaleToZeroRetentionSeconds: 300,
        };
        const result = wrapper.vm.getThresholdValues(autoScaleData);
        expect(result).toEqual({ thresholdMetric: "cpu", thresholdValue: 80 });
      });

      it("returns undefined for both values if shouldShowReplicaCondition is false, even with policyDefaults", () => {
        const policyDefaults = { metric: "concurrency", metricThreshold: 55 };
        const wrapper = getMountedComponent({ policyDefaults });
        const autoScaleData = {
          minReplicas: 0,
          maxReplicas: 1,
          thresholdMetric: undefined,
          thresholdValue: undefined,
          scaleToZeroRetentionSeconds: 300,
        };
        const result = wrapper.vm.getThresholdValues(autoScaleData);
        expect(result).toEqual({ thresholdMetric: undefined, thresholdValue: undefined });
      });
    });
  });

  describe("validation helpers", () => {
    it("isValueEmpty", () => {
      const wrapper = getMountedComponent();
      expect(wrapper.vm.isValueEmpty("123")).toBe("Enter a value");
    });
    it("isGreaterThanZero", () => {
      const wrapper = getMountedComponent();
      expect(wrapper.vm.isGreaterThanZero("0")).toBe(errorMessages.VALUE_ABOVE_ZERO);
      expect(wrapper.vm.isGreaterThanZero("1")).toBe(true);
    });
    it("isMinReplicasAllowedZero", () => {
      const wrapper = getMountedComponent();
      expect(wrapper.vm.isMinReplicasAllowedZero(0)).toBe(true);
      const wrapper2 = getMountedComponent({
        autoScaleOptions: { ...baseProps.autoScaleOptions, isAutoScaleEnabled: false },
      });
      expect(wrapper2.vm.isMinReplicasAllowedZero(0)).toBe(errorMessages.MIN_REPLICAS_ALLOWED_ZERO);
      expect(wrapper2.vm.isMinReplicasAllowedZero(1)).toBe(true);
    });
  });

  describe("policyRules integration", () => {
    const policyRules = {
      minReplicas: { canEdit: false, min: 2, max: 5, step: 2, required: true },
      maxReplicas: { canEdit: true, min: 3, max: 10, step: 1, required: false },
      metricThreshold: { min: 10, max: 90, step: 10, required: true, canEdit: true },
      scaleToZeroRetentionSeconds: { min: 60, max: 600, step: 60, required: true, canEdit: true },
      metric: {
        options: [
          { value: "cpu", displayed: "CPU" },
          { value: "mem", displayed: "Memory" },
        ],
        canEdit: true,
        required: true,
      },
    };

    it("passes policyRules to ReplicaScalingSection and disables minReplicas input if canEdit is false", () => {
      const wrapper = getMountedComponent({ policyRules });
      const cmp = wrapper.findComponent({ name: "replica-scaling-section" });
      expect(cmp.exists()).toBe(true);
      const minInput = cmp.findAllComponents({ name: "policy-number-input" })[0];
      expect(minInput.vm.disableInput).toBe(true);
      expect(minInput.vm.min).toBe(2);
      expect(minInput.vm.max).toBe(5);
      expect(minInput.vm.step).toBe(2);
    });

    it("passes policyRules to ReplicaScalingSection and applies min/max/step to maxReplicas input", () => {
      const wrapper = getMountedComponent({ policyRules });
      const cmp = wrapper.findComponent({ name: "replica-scaling-section" });
      const maxInput = cmp.findAllComponents({ name: "policy-number-input" })[1];
      expect(maxInput.vm.disableInput).toBe(false);
      expect(maxInput.vm.min).toBe(3);
      expect(maxInput.vm.max).toBe(10);
      expect(maxInput.vm.step).toBe(1);
    });

    it("passes policyRules to ThresholdMetricSection and applies min/max/step to threshold value input", () => {
      const wrapper = getMountedComponent({ policyRules });
      const cmp = wrapper.findComponent({ name: "threshold-metric-section" });
      expect(cmp.exists()).toBe(true);
      const input = cmp.findComponent({ name: "policy-number-input" });
      expect(input.vm.min).toBe(10);
      expect(input.vm.max).toBe(90);
      expect(input.vm.step).toBe(10);
      expect(input.vm.disableInput).toBe(false);
    });

    it("passes metric options to ThresholdMetricSection", () => {
      const wrapper = getMountedComponent({ policyRules });
      const cmp = wrapper.findComponent({ name: "threshold-metric-section" });
      expect(cmp.exists()).toBe(true);
      const metricOptions = cmp.vm.policyMetricsRules.options;
      expect(metricOptions).toEqual([
        { label: "CPU", value: "cpu" },
        { label: "Memory", value: "mem" },
      ]);
    });

    it("passes policyRules to ScaleDownSection and applies min/max/step to custom input", async () => {
      const wrapper = getMountedComponent({
        policyRules,
        autoScaleData: { ...baseProps.autoScaleData, minReplicas: 0 },
      });
      // Enable scale down section by setting minReplicas: 0
      const cmp = wrapper.findComponent({ name: "scale-down-section" });
      expect(cmp.exists()).toBe(true);
      // Select custom option
      cmp.vm.onScaleDownChanged("custom");
      await cmp.vm.$nextTick();
      const input = cmp.findComponent({ name: "policy-number-input" });
      expect(input.vm.min).toBe(60);
      expect(input.vm.max).toBe(600);
      expect(input.vm.step).toBe(60);
      expect(input.vm.disableInput).toBe(false);
    });

    it("disables custom input in ScaleDownSection if canEdit is false", async () => {
      const customRules = {
        ...policyRules,
        scaleToZeroRetentionSeconds: { ...policyRules.scaleToZeroRetentionSeconds, canEdit: false },
      };
      const wrapper = getMountedComponent({
        policyRules: customRules,
        autoScaleData: { ...baseProps.autoScaleData, minReplicas: 0 },
      });
      const cmp = wrapper.findComponent({ name: "scale-down-section" });
      cmp.vm.onScaleDownChanged("custom");
      await cmp.vm.$nextTick();
      const input = cmp.findComponent({ name: "policy-number-input" });
      expect(input.vm.disableInput).toBe(true);
    });
  });
});
