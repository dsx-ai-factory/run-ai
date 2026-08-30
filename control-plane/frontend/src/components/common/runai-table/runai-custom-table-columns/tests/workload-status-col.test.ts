import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";

// cmps

import WorkloadPendingSchedulingMessagesMenu from "@/components/workload/workload-pending-scheduling-messages-menu/workload-pending-scheduling-messages-menu.vue";
import WorkloadPhaseMessagesMenu from "@/components/workload/workload-phase-messages-menu/workload-phase-messages-menu.vue";

// models
import { PhaseReason, type Workload } from "@/swagger-models/workloads-client";
import { Phase } from "@/swagger-models/workloads-client";
import type { ITableColumn } from "@/models/table.model";
import { allWorkloadPhaseMessageColumnsMap, type Field } from "@/models/workload-phase-messages.model";
import { WorkloadStatusCol } from "@/components/common/runai-table/runai-custom-table-columns";
import { RunaiStatus } from "@/components/common/runai-status";
import {
  mockStatusOptions,
  mockWorkloadColumn,
  mockWorkloadRow,
} from "@/components/common/runai-table/runai-custom-table-columns/tests/mock/workload-status-col.mock";
import { tableFormatUtil } from "@/utils/table-format.util/table-format.util";
import { unitTestUtil } from "@/utils/unit-test.util";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
vi.mock("@/utils/table-format.util/table-format.util");
vi.mock("@/services/cluster/workload.service/workload.service");
const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });

describe("workload-status-col component", () => {
  let wrapper: VueWrapper<InstanceType<typeof WorkloadStatusCol>>;
  const createComponent = (props: { row: Workload; column: ITableColumn }): void => {
    wrapper = mount(WorkloadStatusCol, {
      global: {
        plugins: [Quasar, pinia],
        directives: {
          permission: vi.fn(),
        },
      },
      props: props,
    });
    wrapper.vm.$q.notify = vi.fn();
  };

  afterEach(() => {
    wrapper.unmount();
    vi.clearAllMocks();
  });

  describe("Running Phase without issues", () => {
    beforeEach(() => {
      vi.mocked(tableFormatUtil.getFieldContent).mockReturnValue({ ...mockStatusOptions, status: "Running" });
      createComponent({
        row: { ...mockWorkloadRow, phase: Phase.Running, additionalFields: {} },
        column: mockWorkloadColumn,
      });
    });

    test("renders RunaiStatus component", () => {
      expect(wrapper.findComponent(RunaiStatus).exists()).toBe(true);
    });

    test("does not render extended menu button", () => {
      expect(wrapper.find(".ph-info").exists()).toBe(false);
    });
  });

  describe("Pending Phase with phaseReason", () => {
    beforeEach(() => {
      vi.mocked(tableFormatUtil.getFieldContent).mockReturnValue({ ...mockStatusOptions, status: "Pending" });
      createComponent({
        row: { ...mockWorkloadRow, phase: Phase.Pending, phaseReason: PhaseReason.NonPreemptibleOverQuota },
        column: mockWorkloadColumn,
      });
    });

    test("renders extended menu button", () => {
      expect(wrapper.find(".ph-info").exists()).toBe(true);
    });
  });

  describe("Failed Phase with issues", () => {
    beforeEach(() => {
      vi.mocked(tableFormatUtil.getFieldContent).mockReturnValue({ ...mockStatusOptions, status: "Failed" });

      createComponent({
        row: {
          ...mockWorkloadRow,
          phase: Phase.Failed,
        },
        column: mockWorkloadColumn,
      });
    });

    test("renders extended menu button", () => {
      expect(wrapper.find(".ph-info").exists()).toBe(true);
    });

    test("opens WorkloadPhaseMessagesMenu on button click", async () => {
      await wrapper.find(".ph-info").trigger("click");
      await wrapper.vm.$nextTick();
      expect(wrapper.findComponent(WorkloadPhaseMessagesMenu).exists()).toBe(true);
      expect(wrapper.vm.isPhaseMessagesMenuOpen).toBe(true);
      expect(wrapper.vm.phaseMessages.length).toBe(2);
      expect(wrapper.findComponent(WorkloadPhaseMessagesMenu).props().columns).toEqual(
        allWorkloadPhaseMessageColumnsMap,
      );
    });
  });

  describe("Pending Phase without phaseReason", () => {
    beforeEach(() => {
      vi.mocked(workloadService.getWorkloadPendingSchedulingMessages).mockResolvedValue([]);
      vi.mocked(tableFormatUtil.getFieldContent).mockReturnValue({ ...mockStatusOptions, status: "Pending" });
      createComponent({ row: { ...mockWorkloadRow, phase: Phase.Pending }, column: mockWorkloadColumn });
    });

    test("does not render extended menu button", () => {
      expect(wrapper.find(".ph-info").exists()).toBe(false);
    });
  });

  describe("Failed Phase without issues with phaseMessage", () => {
    beforeEach(() => {
      vi.mocked(tableFormatUtil.getFieldContent).mockReturnValue({
        ...mockStatusOptions,
        status: "Failed",
        tooltipText: "some message",
      });
      createComponent({
        row: { ...mockWorkloadRow, additionalFields: {}, phase: Phase.Failed, phaseMessage: "some message" },
        column: mockWorkloadColumn,
      });
    });

    test("does not render extended menu button", async () => {
      expect(wrapper.find(".ph-info").exists()).toBe(true);
      await wrapper.find(".ph-info").trigger("click");
      expect(unitTestUtil.getMenuElement().find(".tooltip").text()).toBe("some message");
    });
  });

  describe("Pending Phase", () => {
    beforeEach(() => {
      vi.mocked(tableFormatUtil.getFieldContent).mockReturnValue({ ...mockStatusOptions, status: "Pending" });

      createComponent({
        row: { ...mockWorkloadRow, phaseReason: PhaseReason.OverLimit, phase: Phase.Pending },
        column: mockWorkloadColumn,
      });
    });

    test("does not render extended menu button", async () => {
      await wrapper.find(".ph-info").trigger("click");
      expect(wrapper.findComponent(WorkloadPendingSchedulingMessagesMenu).exists()).toBe(true);
    });
  });

  describe("Running Phase with issues", () => {
    beforeEach(() => {
      vi.mocked(tableFormatUtil.getFieldContent).mockReturnValue({ ...mockStatusOptions, status: "Running" });
      createComponent({
        row: {
          ...mockWorkloadRow,
          phase: Phase.Running,
        },
        column: mockWorkloadColumn,
      });
    });

    test("renders extended menu button", () => {
      expect(wrapper.find(".ph-info").exists()).toBe(true);
    });

    test("opens WorkloadPhaseMessagesMenu on button click", async () => {
      await wrapper.find(".ph-info").trigger("click");
      await wrapper.vm.$nextTick();
      expect(wrapper.findComponent(WorkloadPhaseMessagesMenu).exists()).toBe(true);
      expect(wrapper.vm.isPhaseMessagesMenuOpen).toBe(true);
      expect(wrapper.vm.phaseMessages.length).toBe(2);
    });
  });

  describe("Degraded Phase with issues", () => {
    beforeEach(() => {
      vi.mocked(tableFormatUtil.getFieldContent).mockReturnValue({ ...mockStatusOptions, status: "Degraded" });
      createComponent({
        row: {
          ...mockWorkloadRow,
          phase: Phase.Degraded,
        },
        column: mockWorkloadColumn,
      });
    });

    test("renders extended menu button", () => {
      expect(wrapper.find(".ph-info").exists()).toBe(true);
    });

    test("opens WorkloadPhaseMessagesMenu on button click", async () => {
      await wrapper.find(".ph-info").trigger("click");
      await wrapper.vm.$nextTick();
      expect(wrapper.findComponent(WorkloadPhaseMessagesMenu).exists()).toBe(true);
      expect(wrapper.vm.isPhaseMessagesMenuOpen).toBe(true);
    });
  });

  describe("Completed Phase with issues", () => {
    beforeEach(() => {
      vi.mocked(tableFormatUtil.getFieldContent).mockReturnValue({ ...mockStatusOptions, status: "Completed" });
      createComponent({
        row: {
          ...mockWorkloadRow,
          phase: Phase.Completed,
        },
        column: mockWorkloadColumn,
      });
    });

    test("does not render extended menu button", () => {
      expect(wrapper.find(".ph-info").exists()).toBe(false);
    });
  });

  describe("Stopped Phase without issues", () => {
    beforeEach(() => {
      vi.mocked(tableFormatUtil.getFieldContent).mockReturnValue({ ...mockStatusOptions, status: "Stopped" });
      createComponent({
        row: {
          ...mockWorkloadRow,
          phase: Phase.Stopped,
          additionalFields: {},
        },
        column: mockWorkloadColumn,
      });
    });

    test("does not render extended menu button", () => {
      expect(wrapper.find(".ph-info").exists()).toBe(false);
    });
  });

  describe("Creating Phase with name validation issue", () => {
    const longNameWorkload = {
      ...mockWorkloadRow,
      name: "really-aadfjvndkfvnsiduvni348hv3948vh39v8qh439vb4545b4545bs8qh45v948h34v34gv-long",
      phase: Phase.Creating,
      type: "Workspace",
      additionalFields: {
        issues:
          '{"network":null,"storage":null,"compute":[{"kind":"RunaiJob","name":"really-aadfjvndkfvnsiduvni348hv3948vh39v8qh439vb4545b4545bs8qh45v948h34v34gv-long","description":"[RunaiJob:really-aadfjvndkfvnsiduvni348hv3948vh39v8qh439vb4545b4545bs8qh45v948h34v34gv-long] RunaiJob.run.ai \\"really-aadfjvndkfvnsiduvni348hv3948vh39v8qh439vb4545b4545bs8qh45v948h34v34gv-long\\" is invalid: metadata.labels: Invalid value: \\"really-aadfjvndkfvnsiduvni348hv3948vh39v8qh439vb4545b4545bs8qh45v948h34v34gv-long\\": must be no more than 63 characters"}],"internal":null}',
      },
    };

    beforeEach(() => {
      vi.mocked(tableFormatUtil.getFieldContent).mockReturnValue({ ...mockStatusOptions, status: "Creating" });
      createComponent({
        row: longNameWorkload,
        column: mockWorkloadColumn,
      });
    });

    test("renders extended menu button for Creating phase with issues", () => {
      expect(wrapper.find(".ph-info").exists()).toBe(true);
    });

    test("opens WorkloadPhaseMessagesMenu with name validation issue", async () => {
      await wrapper.find(".ph-info").trigger("click");
      await wrapper.vm.$nextTick();
      expect(wrapper.findComponent(WorkloadPhaseMessagesMenu).exists()).toBe(true);
      expect(wrapper.vm.isPhaseMessagesMenuOpen).toBe(true);
      expect(wrapper.vm.phaseMessages.length).toBeGreaterThan(0);

      // Verify the issue contains the name length validation error
      const firstMessage = wrapper.vm.phaseMessages[0];
      expect(firstMessage.description).toContain("must be no more than 63 characters");
    });

    test("displays correct issue category (compute)", async () => {
      await wrapper.find(".ph-info").trigger("click");
      await wrapper.vm.$nextTick();

      const messages = wrapper.vm.phaseMessages;
      const computeIssue = messages.find((msg: Field) => msg.parentName === "Compute");
      expect(computeIssue).toBeDefined();
      expect(computeIssue?.kind).toBe("RunaiJob");
      expect(computeIssue?.name).toBe(
        "really-aadfjvndkfvnsiduvni348hv3948vh39v8qh439vb4545b4545bs8qh45v948h34v34gv-long",
      );
    });
  });
});
