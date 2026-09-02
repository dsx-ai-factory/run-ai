/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest";
import { mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";

import ThresholdMetricSection from "./threshold-metric-section.vue";
import {
  EThresholdMetricSectionEvents,
  METRIC_LATENCY,
  DEFAULT_THRESHOLD_VALUE,
} from "./threshold-metric-section.model";
import { errorMessages } from "@/common/error-message.constant";
import type { AutoScalingRules } from "@/swagger-models/assets-service-client";
import type { ISelectOption } from "@/models/global.model";

interface IThresholdMetricProps {
  modelValue: { thresholdMetric: string; thresholdValue: number };
  latencyEnabled: boolean;
  policyRules?: AutoScalingRules;
  policyDefaults?: { metric?: string; metricThreshold?: number };
}

describe("threshold-metric-section.vue", () => {
  const baseProps: IThresholdMetricProps = {
    modelValue: { thresholdMetric: "", thresholdValue: 1 },
    latencyEnabled: true,
  };

  function getMountedComponent(props: IThresholdMetricProps): VueWrapper<InstanceType<typeof ThresholdMetricSection>> {
    return mount(ThresholdMetricSection, {
      global: {
        plugins: [Quasar],
      },
      props,
    });
  }

  describe("Renders", () => {
    it("renders correctly", () => {
      const wrapper = getMountedComponent(baseProps);
      expect(wrapper.find(".threshold-metric-section").exists()).toBe(true);
      expect(wrapper.findComponent({ name: "RunaiSelect" }).exists()).toBe(true);
      expect(wrapper.findComponent({ name: "PolicyNumberInput" }).exists()).toBe(true);
    });

    it("disables latency metric option if latencyEnabled is false", () => {
      const wrapper = getMountedComponent({ ...baseProps, latencyEnabled: false });
      const selectOptions = wrapper.vm.thresholdMetricOptions;
      const latencyOption = (selectOptions as Array<{ value: string; disable: boolean; toolTip: string }>).find(
        (o) => o.value === METRIC_LATENCY.toString(),
      );
      expect(latencyOption?.disable).toBe(true);
      expect(latencyOption?.toolTip).toContain("Knative");
    });

    describe("policyDefaults.metric option handling", () => {
      it("adds policyDefaults.metric to thresholdMetricOptions if not present", () => {
        const customMetric = "custom_metric";
        const testProps: Partial<IThresholdMetricProps> = {
          ...baseProps,
          policyDefaults: { metric: customMetric },
        };
        const wrapper = getMountedComponent(testProps as IThresholdMetricProps);
        expect(wrapper.vm.thresholdMetricOptions[0]).toMatchObject({
          value: customMetric,
          label: customMetric,
          disable: false,
          toolTip: "",
        });

        const count = wrapper.vm.thresholdMetricOptions.filter(
          (opt: ISelectOption) => opt.value === customMetric,
        ).length;
        expect(count).toBe(1);
      });

      it("does not duplicate policyDefaults.metric if already present in options", () => {
        const existingMetric = "latency";
        const testProps: Partial<IThresholdMetricProps> = {
          ...baseProps,
          policyDefaults: { metric: existingMetric },
        };
        const wrapper = getMountedComponent(testProps as IThresholdMetricProps);
        const count = wrapper.vm.thresholdMetricOptions.filter(
          (opt: ISelectOption) => opt.value === existingMetric,
        ).length;
        expect(count).toBe(1);

        expect(
          wrapper.vm.thresholdMetricOptions[0].value === existingMetric &&
            wrapper.vm.thresholdMetricOptions[0].label === existingMetric,
        ).toBe(false);
      });
    });
  });

  describe("Emits", () => {
    it("emits update event when threshold metric changes", () => {
      const wrapper = getMountedComponent(baseProps);
      wrapper.vm.onThresholdMetricChanged("1");
      expect(wrapper.emitted()[EThresholdMetricSectionEvents.UpdateModelValue]).toBeTruthy();
      const emitted = wrapper.emitted()[EThresholdMetricSectionEvents.UpdateModelValue];
      expect(emitted).toEqual([[{ thresholdMetric: "1", thresholdValue: DEFAULT_THRESHOLD_VALUE }]]);
    });
    it("emits update event when threshold value changes", () => {
      const wrapper = getMountedComponent(baseProps);
      wrapper.vm.onThresholdValueChanged(5);
      const emitted = wrapper.emitted()[EThresholdMetricSectionEvents.UpdateModelValue];
      expect(emitted).toEqual([[{ thresholdMetric: "", thresholdValue: 5 }]]);
    });

    it("emits update event when threshold value is null", () => {
      const wrapper = getMountedComponent(baseProps);
      wrapper.vm.onThresholdValueChanged(null);
      const emitted = wrapper.emitted()[EThresholdMetricSectionEvents.UpdateModelValue];
      expect(emitted).toEqual([[{ thresholdMetric: "", thresholdValue: undefined }]]);
    });

    it("emits update event when threshold value is an empty string", () => {
      const wrapper = getMountedComponent(baseProps);
      wrapper.vm.onThresholdValueChanged("");
      const emitted = wrapper.emitted()[EThresholdMetricSectionEvents.UpdateModelValue];
      expect(emitted).toEqual([[{ thresholdMetric: "", thresholdValue: undefined }]]);
    });

    it("emits update event when threshold value is 0", () => {
      const wrapper = getMountedComponent(baseProps);
      wrapper.vm.onThresholdValueChanged(0);
      const emitted = wrapper.emitted()[EThresholdMetricSectionEvents.UpdateModelValue];
      expect(emitted).toEqual([[{ thresholdMetric: "", thresholdValue: 0 }]]);
    });

    it("emits correct thresholdValue when policyDefaults.metricThreshold is set", () => {
      const policyDefaults = { metricThreshold: 42 };
      const wrapper = getMountedComponent({ ...baseProps, policyDefaults });
      wrapper.vm.onThresholdMetricChanged("2");
      const emitted = wrapper.emitted()[EThresholdMetricSectionEvents.UpdateModelValue];
      expect(emitted).toEqual([[{ thresholdMetric: "2", thresholdValue: 42 }]]);
    });

    it("emits DEFAULT_THRESHOLD_VALUE when policyDefaults is undefined", () => {
      const wrapper = getMountedComponent(baseProps);
      wrapper.vm.onThresholdMetricChanged("3");
      const emitted = wrapper.emitted()[EThresholdMetricSectionEvents.UpdateModelValue];
      expect(emitted).toEqual([[{ thresholdMetric: "3", thresholdValue: DEFAULT_THRESHOLD_VALUE }]]);
    });

    it("emits DEFAULT_THRESHOLD_VALUE when policyDefaults.metricThreshold is undefined or missing", () => {
      const policyDefaults = {};
      const wrapper = getMountedComponent({ ...baseProps, policyDefaults });
      wrapper.vm.onThresholdMetricChanged("4");
      const emitted = wrapper.emitted()[EThresholdMetricSectionEvents.UpdateModelValue];
      expect(emitted).toEqual([[{ thresholdMetric: "4", thresholdValue: DEFAULT_THRESHOLD_VALUE }]]);
    });
  });

  describe("Computes", () => {
    it("computed policyMetricsRules returns undefined if no policyRules", () => {
      const wrapper = getMountedComponent(baseProps);
      expect(wrapper.vm.policyMetricsRules).toBeUndefined();
    });

    it("computed policyMetricsRules returns mapped options if policyRules.metric exists", () => {
      const policyRules = {
        metric: {
          options: [
            { value: "cpu", displayed: "CPU" },
            { value: "mem", displayed: "Memory" },
          ],
        },
      };
      const wrapper = getMountedComponent({ ...baseProps, policyRules });
      wrapper.setProps({ policyRules });
      const rules = wrapper.vm.policyMetricsRules;
      expect(rules?.options).toEqual([
        { label: "CPU", value: "cpu" },
        { label: "Memory", value: "mem" },
      ]);
    });
  });

  describe("Methods", () => {
    it("isValueEmpty returns error message if not a number", () => {
      const wrapper = getMountedComponent(baseProps);
      expect(wrapper.vm.isValueEmpty("")).toBe(errorMessages.VALUE_NOT_EMPTY);
      expect(wrapper.vm.isValueEmpty("abc")).toBe(true);
    });

    it("isGreaterThanMin returns error message if not greater than zero", () => {
      const wrapper = getMountedComponent(baseProps);
      expect(wrapper.vm.isGreaterThanMin("0")).toBe(errorMessages.VALUE_ABOVE_ZERO);
      expect(wrapper.vm.isGreaterThanMin("-1")).toBe(errorMessages.VALUE_ABOVE_ZERO);
      expect(wrapper.vm.isGreaterThanMin("1")).toBe(true);
      expect(wrapper.vm.isGreaterThanMin("2")).toBe(true);
    });
  });
});
