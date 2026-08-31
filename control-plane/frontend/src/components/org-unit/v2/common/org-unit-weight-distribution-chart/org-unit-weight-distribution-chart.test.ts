/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import { describe, it, expect, vi } from "vitest";
import OrgUnitWeightDistributionChart from "./org-unit-weight-distribution-chart.vue";
import { ORG_UNIT_STACKED_BAR_CHART_COLORS } from "@/models/org-unit-forms.model";
import { EQuotaEntity } from "@/models/resource.model";

// Mock the StackedBarChart component
vi.mock("@/components/common/charts/stacked-bar-chart/stacked-bar-chart.vue", () => ({
  default: {
    name: "StackedBarChart",
    template: "<div class='mock-stacked-bar-chart'></div>",
    props: [
      "segments",
      "entityName",
      "maxValue",
      "height",
      "showLabels",
      "showScale",
      "showPercentage",
      "labelTextBorderWidth",
    ],
  },
}));

// Gray color for "all zeros" empty state
const EMPTY_STATE_COLOR = "#BDBDBD";

interface OrgUnitWeightDistributionChartProps {
  entityName: string;
  entityType: EQuotaEntity;
  weight: number;
  otherEntitiesWeight: number;
  height?: number;
  showLabels?: boolean;
  showScale?: boolean;
}

describe("OrgUnitWeightDistributionChart", () => {
  const defaultProps: OrgUnitWeightDistributionChartProps = {
    entityName: "Project-1",
    entityType: EQuotaEntity.project,
    weight: 1,
    otherEntitiesWeight: 3,
  };

  const mountComponent = (props: Partial<OrgUnitWeightDistributionChartProps> = {}) => {
    return mount(OrgUnitWeightDistributionChart, {
      props: {
        ...defaultProps,
        ...props,
      },
    });
  };

  describe("Rendering", () => {
    it("renders the component correctly", () => {
      const wrapper = mountComponent();
      expect(wrapper.exists()).toBe(true);
    });

    it("passes entityName to StackedBarChart", () => {
      const wrapper = mountComponent({ entityName: "Test-Project" });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      expect(stackedBarChart.props("entityName")).toBe("Test-Project");
    });

    it("always enables showPercentage", () => {
      const wrapper = mountComponent();
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      // show-percentage is passed as a boolean attribute, so it can be true or "" (empty string which is truthy in HTML)
      expect(stackedBarChart.props("showPercentage")).not.toBe(false);
    });

    it("uses the production-equivalent label outline width", () => {
      const wrapper = mountComponent();
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });

      expect(stackedBarChart.props("labelTextBorderWidth")).toBe(2);
    });
  });

  describe("Max Value Calculation", () => {
    it("calculates maxValue as sum of weight and otherEntitiesWeight", () => {
      const wrapper = mountComponent({
        weight: 1,
        otherEntitiesWeight: 3,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      expect(stackedBarChart.props("maxValue")).toBe(4); // 1 + 3
    });

    it("updates maxValue when weights change", async () => {
      const wrapper = mountComponent({
        weight: 2,
        otherEntitiesWeight: 2,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      expect(stackedBarChart.props("maxValue")).toBe(4);

      await wrapper.setProps({ weight: 5 });
      expect(stackedBarChart.props("maxValue")).toBe(7); // 5 + 2
    });
  });

  describe("Segments Calculation", () => {
    it("creates two segments with correct values", () => {
      const wrapper = mountComponent({
        weight: 1,
        otherEntitiesWeight: 3,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      const segments = stackedBarChart.props("segments");

      expect(segments).toHaveLength(2);
      expect(segments[0].value).toBe(1); // weight
      expect(segments[1].value).toBe(3); // other entities weight
    });

    it("uses correct colors for segments", () => {
      const wrapper = mountComponent();
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      const segments = stackedBarChart.props("segments");

      expect(segments[0].color).toBe(ORG_UNIT_STACKED_BAR_CHART_COLORS.blue9); // weight - blue
      expect(segments[1].color).toBe(ORG_UNIT_STACKED_BAR_CHART_COLORS.blue2); // other entities - light blue
    });

    it("uses correct labels for segments", () => {
      const wrapper = mountComponent({ entityType: EQuotaEntity.project });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      const segments = stackedBarChart.props("segments");

      expect(segments[0].label).toBe("Weight");
      expect(segments[1].label).toBe("Total weight of other projects");
    });

    it("uses entityType in the other entities label", () => {
      const wrapper = mountComponent({ entityType: EQuotaEntity.department });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      const segments = stackedBarChart.props("segments");

      expect(segments[1].label).toBe("Total weight of other departments");
    });
  });

  describe("Optional Props", () => {
    it("uses default height of 18", () => {
      const wrapper = mountComponent();
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      expect(stackedBarChart.props("height")).toBe(18);
    });

    it("passes custom height to StackedBarChart", () => {
      const wrapper = mountComponent({ height: 40 });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      expect(stackedBarChart.props("height")).toBe(40);
    });

    it("uses default showLabels of true", () => {
      const wrapper = mountComponent();
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      expect(stackedBarChart.props("showLabels")).toBe(true);
    });

    it("passes showLabels=false to StackedBarChart", () => {
      const wrapper = mountComponent({ showLabels: false });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      expect(stackedBarChart.props("showLabels")).toBe(false);
    });

    it("uses default showScale of true", () => {
      const wrapper = mountComponent();
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      expect(stackedBarChart.props("showScale")).toBe(true);
    });

    it("passes showScale=false to StackedBarChart", () => {
      const wrapper = mountComponent({ showScale: false });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      expect(stackedBarChart.props("showScale")).toBe(false);
    });
  });

  describe("Reactivity", () => {
    it("updates segments when weight changes", async () => {
      const wrapper = mountComponent({ weight: 1 });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });

      expect(stackedBarChart.props("segments")[0].value).toBe(1);

      await wrapper.setProps({ weight: 5 });
      expect(stackedBarChart.props("segments")[0].value).toBe(5);
    });

    it("updates segments when otherEntitiesWeight changes", async () => {
      const wrapper = mountComponent({ otherEntitiesWeight: 3 });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });

      expect(stackedBarChart.props("segments")[1].value).toBe(3);

      await wrapper.setProps({ otherEntitiesWeight: 7 });
      expect(stackedBarChart.props("segments")[1].value).toBe(7);
    });
  });

  describe("Edge Cases", () => {
    it("handles equal weights", () => {
      const wrapper = mountComponent({
        weight: 2,
        otherEntitiesWeight: 2,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      const segments = stackedBarChart.props("segments");

      expect(segments[0].value).toBe(2);
      expect(segments[1].value).toBe(2);
      expect(stackedBarChart.props("maxValue")).toBe(4);
    });

    it("handles zero other entities weight", () => {
      const wrapper = mountComponent({
        weight: 5,
        otherEntitiesWeight: 0,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      const segments = stackedBarChart.props("segments");

      expect(segments[0].value).toBe(5);
      expect(segments[1].value).toBe(0);
      expect(stackedBarChart.props("maxValue")).toBe(5);
    });

    it("handles large weight values", () => {
      const wrapper = mountComponent({
        weight: 100,
        otherEntitiesWeight: 900,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });

      expect(stackedBarChart.props("maxValue")).toBe(1000);
    });

    it("ensures maxValue is at least 1 when weight is negative", () => {
      const wrapper = mountComponent({
        weight: -1,
        otherEntitiesWeight: 0,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });

      // maxValue should be 1 (minimum), not -1
      // This prevents chart rendering issues with negative scale
      expect(stackedBarChart.props("maxValue")).toBe(1);
    });

    it("ensures maxValue is at least 1 when both weights are 0", () => {
      const wrapper = mountComponent({
        weight: 0,
        otherEntitiesWeight: 0,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });

      // maxValue should be 1 (minimum), not 0
      expect(stackedBarChart.props("maxValue")).toBe(1);
    });
  });

  describe("All Zeros State", () => {
    it("shows gray bar with displayValue '0' when both weights are zero", () => {
      const wrapper = mountComponent({
        weight: 0,
        otherEntitiesWeight: 0,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      const segments = stackedBarChart.props("segments");

      // First segment is the gray bar with "0"
      expect(segments[0].value).toBe(1);
      expect(segments[0].color).toBe(EMPTY_STATE_COLOR);
      expect(segments[0].displayValue).toBe("0");
    });

    it("hides scale when both weights are zero", () => {
      const wrapper = mountComponent({
        weight: 0,
        otherEntitiesWeight: 0,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });

      expect(stackedBarChart.props("showScale")).toBe(false);
    });

    it("sets maxValue to 1 when both weights are zero", () => {
      const wrapper = mountComponent({
        weight: 0,
        otherEntitiesWeight: 0,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });

      expect(stackedBarChart.props("maxValue")).toBe(1);
    });

    it("includes tooltip segments with zero values for all zeros state", () => {
      const wrapper = mountComponent({
        weight: 0,
        otherEntitiesWeight: 0,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      const segments = stackedBarChart.props("segments");

      // Should have 3 segments: gray bar + 2 tooltip-only segments
      expect(segments).toHaveLength(3);
      expect(segments[1].displayValue).toBe("0");
      expect(segments[2].displayValue).toBe("0");
    });

    it("does not trigger all zeros state when weight has value", () => {
      const wrapper = mountComponent({
        weight: 5,
        otherEntitiesWeight: 0,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      const segments = stackedBarChart.props("segments");

      // Should not be gray empty state
      expect(segments[0].color).not.toBe(EMPTY_STATE_COLOR);
      expect(segments[0].value).toBe(5);
    });

    it("does not trigger all zeros state when otherEntitiesWeight has value", () => {
      const wrapper = mountComponent({
        weight: 0,
        otherEntitiesWeight: 3,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      const segments = stackedBarChart.props("segments");

      // Should not be gray empty state
      expect(segments[0].color).not.toBe(EMPTY_STATE_COLOR);
    });
  });
});
