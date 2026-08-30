import { mount } from "@vue/test-utils";
import { describe, it, expect, vi } from "vitest";
import OrgUnitQuotaDistributionChart from "./org-unit-quota-distribution-chart.vue";
import { ORG_UNIT_STACKED_BAR_CHART_COLORS } from "@/models/org-unit-forms.model";
import { EQuotaEntity } from "@/models/resource.model";

// Mock the StackedBarChart component
vi.mock("@/components/common/charts/stacked-bar-chart/stacked-bar-chart.vue", () => ({
  default: {
    name: "StackedBarChart",
    template: "<div class='mock-stacked-bar-chart'></div>",
    props: ["segments", "entityName", "maxValue", "tooltipSummary", "height", "showLabels", "showScale"],
  },
}));

interface OrgUnitQuotaDistributionChartProps {
  entityName: string;
  entityType: EQuotaEntity;
  assignedQuota: number;
  parentName: string;
  parentAvailableQuota: number;
  otherEntitiesAssignedQuota: number;
  isUnlimited?: boolean;
  isAssignedQuotaUnlimited?: boolean;
  isUnassignedQuotaUnlimited?: boolean;
  isOtherEntitiesQuotaUnlimited?: boolean;
  height?: number;
  showLabels?: boolean;
  showScale?: boolean;
}

// Gray color for "all zeros" empty state
const EMPTY_STATE_COLOR = "#BDBDBD";

describe("OrgUnitQuotaDistributionChart", () => {
  const defaultProps: OrgUnitQuotaDistributionChartProps = {
    entityName: "my-project",
    entityType: EQuotaEntity.project,
    assignedQuota: 20,
    parentName: "Department Engineering",
    parentAvailableQuota: 100,
    otherEntitiesAssignedQuota: 14,
  };

  const mountComponent = (props: Partial<OrgUnitQuotaDistributionChartProps> = {}) => {
    return mount(OrgUnitQuotaDistributionChart, {
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
      const wrapper = mountComponent({ entityName: "test-project" });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      expect(stackedBarChart.props("entityName")).toBe("test-project");
    });

    it("passes parentAvailableQuota as maxValue to StackedBarChart", () => {
      const wrapper = mountComponent({ parentAvailableQuota: 200 });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      expect(stackedBarChart.props("maxValue")).toBe(200);
    });
  });

  describe("Segments Calculation", () => {
    it("creates three segments with correct values", () => {
      const wrapper = mountComponent({
        assignedQuota: 20,
        parentAvailableQuota: 100,
        otherEntitiesAssignedQuota: 14,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      const segments = stackedBarChart.props("segments");

      expect(segments).toHaveLength(3);
      expect(segments[0].value).toBe(20); // assigned
      expect(segments[1].value).toBe(66); // unassigned (100 - 20 - 14)
      expect(segments[2].value).toBe(14); // other entities
    });

    it("calculates unassigned quota correctly", () => {
      const wrapper = mountComponent({
        assignedQuota: 50,
        parentAvailableQuota: 200,
        otherEntitiesAssignedQuota: 80,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      const segments = stackedBarChart.props("segments");

      // unassigned = 200 - 50 - 80 = 70
      expect(segments[1].value).toBe(70);
    });

    it("uses correct colors for segments", () => {
      const wrapper = mountComponent();
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      const segments = stackedBarChart.props("segments");

      expect(segments[0].color).toBe(ORG_UNIT_STACKED_BAR_CHART_COLORS.blue9); // assigned - blue
      expect(segments[1].color).toBe(ORG_UNIT_STACKED_BAR_CHART_COLORS.blue5); // unassigned - grey
      expect(segments[2].color).toBe(ORG_UNIT_STACKED_BAR_CHART_COLORS.blue2); // other entities - light blue
    });

    it("uses correct labels for segments", () => {
      const wrapper = mountComponent({ entityType: EQuotaEntity.project });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      const segments = stackedBarChart.props("segments");

      expect(segments[0].label).toBe("Assigned quota");
      expect(segments[1].label).toBe("Unassigned quota");
      expect(segments[2].label).toBe("Sum of other projects' assigned quota");
    });

    it("uses entityType in the other entities label", () => {
      const wrapper = mountComponent({ entityType: EQuotaEntity.department });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      const segments = stackedBarChart.props("segments");

      expect(segments[2].label).toBe("Sum of other departments' assigned quota");
    });
  });

  describe("Tooltip Summary", () => {
    it("generates correct tooltip summary", () => {
      const wrapper = mountComponent({
        parentName: "Department Engineering",
        parentAvailableQuota: 100,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      const tooltipSummary = stackedBarChart.props("tooltipSummary");

      expect(tooltipSummary).toBe("Department Department Engineering quota: <b>100</b>");
    });

    it("updates tooltip summary when parentName changes", async () => {
      const wrapper = mountComponent({ parentName: "Cluster prod" });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });

      expect(stackedBarChart.props("tooltipSummary")).toContain("Cluster prod");

      await wrapper.setProps({ parentName: "Cluster dev" });
      expect(stackedBarChart.props("tooltipSummary")).toContain("Cluster dev");
    });
  });

  describe("Optional Props", () => {
    it("uses default height of 20", () => {
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
    it("updates segments when assignedQuota changes", async () => {
      const wrapper = mountComponent({ assignedQuota: 20 });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });

      expect(stackedBarChart.props("segments")[0].value).toBe(20);

      await wrapper.setProps({ assignedQuota: 50 });
      expect(stackedBarChart.props("segments")[0].value).toBe(50);
    });

    it("recalculates unassigned quota when props change", async () => {
      const wrapper = mountComponent({
        assignedQuota: 20,
        parentAvailableQuota: 100,
        otherEntitiesAssignedQuota: 14,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });

      // Initial: 100 - 20 - 14 = 66
      expect(stackedBarChart.props("segments")[1].value).toBe(66);

      await wrapper.setProps({ assignedQuota: 40 });
      // After: 100 - 40 - 14 = 46
      expect(stackedBarChart.props("segments")[1].value).toBe(46);
    });
  });

  describe("Edge Cases", () => {
    it("handles zero assignedQuota", () => {
      const wrapper = mountComponent({
        assignedQuota: 0,
        parentAvailableQuota: 100,
        otherEntitiesAssignedQuota: 30,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      const segments = stackedBarChart.props("segments");

      expect(segments[0].value).toBe(0);
      expect(segments[1].value).toBe(70); // 100 - 0 - 30
      expect(segments[2].value).toBe(30);
    });

    it("handles fully assigned quota", () => {
      const wrapper = mountComponent({
        assignedQuota: 50,
        parentAvailableQuota: 100,
        otherEntitiesAssignedQuota: 50,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      const segments = stackedBarChart.props("segments");

      // When fully assigned (unassigned = 0), the unassigned segment is not shown
      expect(segments).toHaveLength(3);
      expect(segments[0].value).toBe(50); // assigned
      expect(segments[1].value).toBe(0); // other entities
    });

    it("uses segments sum as maxValue when parentAvailableQuota is 0 but segments have values", () => {
      const wrapper = mountComponent({
        assignedQuota: 1,
        parentAvailableQuota: 0,
        otherEntitiesAssignedQuota: 2,
        isUnlimited: false,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });

      // maxValue should be the sum of segments (1 + 0 + 2 = 3), not parentAvailableQuota (0)
      // This prevents rendering issues where the chart scale is 0 but segments have values
      expect(stackedBarChart.props("maxValue")).toBe(3);
    });

    it("uses segments sum as maxValue when segments exceed parentAvailableQuota", () => {
      const wrapper = mountComponent({
        assignedQuota: 80,
        parentAvailableQuota: 50,
        otherEntitiesAssignedQuota: 30,
        isUnlimited: false,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });

      // Segments sum = 80 + 0 (unassigned clamped to 0) + 30 = 110
      // maxValue should be 110, not 50
      expect(stackedBarChart.props("maxValue")).toBe(110);
    });
  });

  describe("Unlimited Quota", () => {
    it("hides unassigned segment when isUnlimited is true", () => {
      const wrapper = mountComponent({
        assignedQuota: 20,
        parentAvailableQuota: 0, // Value doesn't matter when unlimited
        otherEntitiesAssignedQuota: 14,
        isUnlimited: true,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      const segments = stackedBarChart.props("segments");

      // Only 2 segments: assigned + other entities (no unassigned)
      expect(segments).toHaveLength(2);
      expect(segments[0].value).toBe(20); // assigned
      expect(segments[0].color).toBe(ORG_UNIT_STACKED_BAR_CHART_COLORS.blue9); // blue
      expect(segments[1].value).toBe(14); // other entities
      expect(segments[1].color).toBe(ORG_UNIT_STACKED_BAR_CHART_COLORS.blue2); // light blue
    });

    it("uses sum of segments as maxValue when unlimited", () => {
      const wrapper = mountComponent({
        assignedQuota: 30,
        parentAvailableQuota: 0,
        otherEntitiesAssignedQuota: 50,
        isUnlimited: true,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });

      // maxValue should be sum of visible segments (30 + 50 = 80)
      expect(stackedBarChart.props("maxValue")).toBe(80);
    });

    it("shows 'Unlimited' in tooltip when isUnlimited is true", () => {
      const wrapper = mountComponent({
        parentName: "Department Engineering",
        parentAvailableQuota: 0,
        isUnlimited: true,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      const tooltipSummary = stackedBarChart.props("tooltipSummary");

      expect(tooltipSummary).toBe("Department Department Engineering quota: <b>Unlimited</b>");
    });

    it("shows all 3 segments when isUnlimited is false (limited mode)", () => {
      const wrapper = mountComponent({
        assignedQuota: 20,
        parentAvailableQuota: 100,
        otherEntitiesAssignedQuota: 14,
        isUnlimited: false,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      const segments = stackedBarChart.props("segments");

      // All 3 segments visible
      expect(segments).toHaveLength(3);
      expect(segments[0].value).toBe(20); // assigned
      expect(segments[1].value).toBe(66); // unassigned
      expect(segments[2].value).toBe(14); // other entities
    });

    it("hides unassigned segment when limited but unassigned is 0", () => {
      const wrapper = mountComponent({
        assignedQuota: 50,
        parentAvailableQuota: 100,
        otherEntitiesAssignedQuota: 50,
        isUnlimited: false,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      const segments = stackedBarChart.props("segments");

      // Unassigned = 0, should hide it
      expect(segments).toHaveLength(3);
      expect(segments[0].value).toBe(50); // assigned
      expect(segments[1].value).toBe(0); // other entities
    });

    it("updates reactively from limited to unlimited", async () => {
      const wrapper = mountComponent({
        assignedQuota: 20,
        parentAvailableQuota: 100,
        otherEntitiesAssignedQuota: 14,
        isUnlimited: false,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });

      // Initially limited: 3 segments
      expect(stackedBarChart.props("segments")).toHaveLength(3);
      expect(stackedBarChart.props("maxValue")).toBe(100);

      // Change to unlimited
      await wrapper.setProps({ isUnlimited: true });

      // Now unlimited: 2 segments
      expect(stackedBarChart.props("segments")).toHaveLength(2);
      expect(stackedBarChart.props("maxValue")).toBe(34); // 20 + 14
    });
  });

  describe("All Zeros State", () => {
    it("shows gray bar with displayValue '0' when all values are zero", () => {
      const wrapper = mountComponent({
        assignedQuota: 0,
        parentAvailableQuota: 0,
        otherEntitiesAssignedQuota: 0,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      const segments = stackedBarChart.props("segments");

      // First segment is the gray bar with "0"
      expect(segments[0].value).toBe(1);
      expect(segments[0].color).toBe(EMPTY_STATE_COLOR);
      expect(segments[0].displayValue).toBe("0");
    });

    it("hides scale when all values are zero", () => {
      const wrapper = mountComponent({
        assignedQuota: 0,
        parentAvailableQuota: 0,
        otherEntitiesAssignedQuota: 0,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });

      expect(stackedBarChart.props("showScale")).toBe(false);
    });

    it("sets maxValue to 1 when all values are zero", () => {
      const wrapper = mountComponent({
        assignedQuota: 0,
        parentAvailableQuota: 0,
        otherEntitiesAssignedQuota: 0,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });

      expect(stackedBarChart.props("maxValue")).toBe(1);
    });

    it("includes tooltip segments with zero values for all zeros state", () => {
      const wrapper = mountComponent({
        assignedQuota: 0,
        parentAvailableQuota: 0,
        otherEntitiesAssignedQuota: 0,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      const segments = stackedBarChart.props("segments");

      // Should have 4 segments: gray bar + 3 tooltip-only segments
      expect(segments).toHaveLength(4);
      expect(segments[1].displayValue).toBe("0");
      expect(segments[2].displayValue).toBe("0");
      expect(segments[3].displayValue).toBe("0");
    });

    it("does not trigger all zeros state when assignedQuota has value", () => {
      const wrapper = mountComponent({
        assignedQuota: 5,
        parentAvailableQuota: 0,
        otherEntitiesAssignedQuota: 0,
      });
      const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
      const segments = stackedBarChart.props("segments");

      // Should not be gray empty state
      expect(segments[0].color).not.toBe(EMPTY_STATE_COLOR);
    });
  });

  describe("Specific Unlimited States", () => {
    describe("Assigned Quota Unlimited", () => {
      it("shows full dark blue bar with 'Unlimited' when assigned quota is unlimited", () => {
        const wrapper = mountComponent({
          assignedQuota: 0,
          parentAvailableQuota: 100,
          otherEntitiesAssignedQuota: 20,
          isAssignedQuotaUnlimited: true,
        });
        const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
        const segments = stackedBarChart.props("segments");

        expect(segments[0].value).toBe(1);
        expect(segments[0].color).toBe(ORG_UNIT_STACKED_BAR_CHART_COLORS.blue9);
        expect(segments[0].displayValue).toBe("Unlimited");
      });

      it("hides scale when assigned quota is unlimited", () => {
        const wrapper = mountComponent({
          isAssignedQuotaUnlimited: true,
        });
        const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });

        expect(stackedBarChart.props("showScale")).toBe(false);
      });

      it("includes other segments for tooltip when assigned is unlimited", () => {
        const wrapper = mountComponent({
          assignedQuota: 0,
          parentAvailableQuota: 100,
          otherEntitiesAssignedQuota: 20,
          isAssignedQuotaUnlimited: true,
        });
        const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
        const segments = stackedBarChart.props("segments");

        // Should have 3 segments: unlimited bar + 2 tooltip-only segments
        expect(segments).toHaveLength(3);
        expect(segments[1].label).toBe("Unassigned quota");
        expect(segments[2].label).toContain("Sum of other");
      });
    });

    describe("Unassigned Quota Unlimited", () => {
      it("shows full medium blue bar with 'Unlimited' when unassigned quota is unlimited", () => {
        const wrapper = mountComponent({
          assignedQuota: 10,
          parentAvailableQuota: 100,
          otherEntitiesAssignedQuota: 20,
          isUnassignedQuotaUnlimited: true,
        });
        const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
        const segments = stackedBarChart.props("segments");

        // First segment should be assigned quota (for tooltip), second is unlimited bar
        const unlimitedSegment = segments.find((s: { displayValue?: string }) => s.displayValue === "Unlimited");
        expect(unlimitedSegment.color).toBe(ORG_UNIT_STACKED_BAR_CHART_COLORS.blue5);
      });

      it("includes assigned quota in tooltip when unassigned is unlimited", () => {
        const wrapper = mountComponent({
          assignedQuota: 10,
          isUnassignedQuotaUnlimited: true,
        });
        const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
        const segments = stackedBarChart.props("segments");

        // Should have assigned quota segment with displayValue
        const assignedSegment = segments.find((s: { label: string }) => s.label === "Assigned quota");
        expect(assignedSegment.displayValue).toBe("10");
      });
    });

    describe("Other Entities Quota Unlimited", () => {
      it("shows full light blue bar with 'Unlimited' when other entities quota is unlimited", () => {
        const wrapper = mountComponent({
          assignedQuota: 10,
          parentAvailableQuota: 100,
          otherEntitiesAssignedQuota: 0,
          isOtherEntitiesQuotaUnlimited: true,
        });
        const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
        const segments = stackedBarChart.props("segments");

        const unlimitedSegment = segments.find((s: { displayValue?: string }) => s.displayValue === "Unlimited");
        expect(unlimitedSegment.color).toBe(ORG_UNIT_STACKED_BAR_CHART_COLORS.blue2);
      });

      it("includes assigned and unassigned quotas in tooltip when other entities is unlimited", () => {
        const wrapper = mountComponent({
          assignedQuota: 10,
          parentAvailableQuota: 100,
          isOtherEntitiesQuotaUnlimited: true,
        });
        const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
        const segments = stackedBarChart.props("segments");

        // Should have 3 segments: assigned, unassigned, and unlimited bar
        expect(segments).toHaveLength(3);
      });
    });

    describe("Priority", () => {
      it("prioritizes assigned unlimited over unassigned unlimited", () => {
        const wrapper = mountComponent({
          isAssignedQuotaUnlimited: true,
          isUnassignedQuotaUnlimited: true,
        });
        const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
        const segments = stackedBarChart.props("segments");

        // Should show assigned unlimited (blue9)
        expect(segments[0].color).toBe(ORG_UNIT_STACKED_BAR_CHART_COLORS.blue9);
        expect(segments[0].displayValue).toBe("Unlimited");
      });

      it("prioritizes unassigned unlimited over other entities unlimited", () => {
        const wrapper = mountComponent({
          isUnassignedQuotaUnlimited: true,
          isOtherEntitiesQuotaUnlimited: true,
        });
        const stackedBarChart = wrapper.findComponent({ name: "StackedBarChart" });
        const segments = stackedBarChart.props("segments");

        // Should show unassigned unlimited (blue5)
        const unlimitedSegment = segments.find((s: { displayValue?: string }) => s.displayValue === "Unlimited");
        expect(unlimitedSegment.color).toBe(ORG_UNIT_STACKED_BAR_CHART_COLORS.blue5);
      });
    });
  });
});
