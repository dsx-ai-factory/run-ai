import { describe, it, expect, beforeEach, vi } from "vitest";
import { flushPromises, mount } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import WorkloadDrawerContent, { EWorkloadDrawerTabs } from "./workload-drawer-content.vue";
import { Phase, type Workload } from "@/swagger-models/workloads-client";
import { EWorkloadType } from "@/models/workload.model";
import { workloadService } from "@/services/cluster/workload.service/workload.service";

vi.mock("@/services/cluster/workload.service/workload.service", () => ({
  workloadService: {
    getWorkloadPods: vi.fn().mockResolvedValue([]),
  },
}));

vi.mock("@/utils/interval.util", () => ({
  intervalUtil: {
    startInterval: vi.fn(),
    stopInterval: vi.fn(),
  },
}));

vi.mock("@/utils/workload.util", () => ({
  workloadUtil: {
    getWorkloadIcon: vi.fn().mockReturnValue("workspace-new"),
    getWorkloadStatusColOptions: vi.fn().mockReturnValue({
      status: "Running",
      color: "success",
      displayAnimation: false,
    }),
    sortPods: vi.fn().mockReturnValue([]),
  },
}));

describe("WorkloadDrawerContent", () => {
  const mockWorkload: Workload = {
    id: "workload-123",
    name: "test-workload",
    type: EWorkloadType.Workspace,
    clusterId: "cluster-123",
    phase: Phase.Running,
  } as Workload;

  const createWrapper = (props = {}, stubsOverride: Record<string, unknown> = {}) => {
    return mount(WorkloadDrawerContent, {
      props: {
        drawerWidth: "400px",
        selectedWorkloadsAmount: 1,
        workload: mockWorkload,
        ...props,
      },
      global: {
        plugins: [
          [Quasar, { plugins: {} }],
          createTestingPinia({
            createSpy: vi.fn,
            stubActions: false,
          }),
        ],
        stubs: {
          RunaiDrawerContent: {
            template: `
              <div class="runai-drawer-content">
                <div class="header"><slot name="header" /></div>
                <div class="subheader"><slot name="subheader" /></div>
                <div class="body"><slot name="body" /></div>
              </div>
            `,
          },
          RunaiSvgIcon: true,
          RunaiStatus: true,
          RunaiDrawerNoData: true,
          WorkloadHistoryTab: true,
          WorkloadMetricsTab: true,
          WorkloadLogsTab: true,
          WorkloadTerminal: true,
          WorkloadDetailsTab: true,
          WorkloadPodsTab: true,
          QTabs: true,
          QTab: true,
          QTabPanels: true,
          QTabPanel: true,
          ...stubsOverride,
        },
      },
    });
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("hideStatus prop", () => {
    it("should show status when hideStatus is false", () => {
      const wrapper = createWrapper({ hideStatus: false });

      const statusComponent = wrapper.findComponent({ name: "RunaiStatus" });
      expect(statusComponent.exists()).toBe(true);
    });

    it("should hide status when hideStatus is true", () => {
      const wrapper = createWrapper({ hideStatus: true });

      const statusComponent = wrapper.findComponent({ name: "RunaiStatus" });
      expect(statusComponent.exists()).toBe(false);
    });

    it("should show status by default when hideStatus prop is not provided", () => {
      const wrapper = createWrapper();

      const statusComponent = wrapper.findComponent({ name: "RunaiStatus" });
      expect(statusComponent.exists()).toBe(true);
    });

    it("should hide status for deleted workloads view", () => {
      const wrapper = createWrapper({
        hideStatus: true,
        workload: {
          ...mockWorkload,
          phase: Phase.Running,
        },
      });

      const statusComponent = wrapper.findComponent({ name: "RunaiStatus" });
      expect(statusComponent.exists()).toBe(false);
    });
  });

  describe("Pods tab", () => {
    // Render the real tab panels (not stubbed) so the active Pods panel mounts its content.
    const renderPanelStubs = { QTabPanels: false, QTabPanel: false };

    it("should render the Pods tab content when the Pods tab is active", async () => {
      const wrapper = createWrapper({}, renderPanelStubs);
      wrapper.vm.drawerTab = EWorkloadDrawerTabs.Pods;
      await wrapper.vm.$nextTick();

      const podsTab = wrapper.findComponent({ name: "WorkloadPodsTab" });
      expect(podsTab.exists()).toBe(true);
    });

    it("should pass pods and loading state to the Pods tab", async () => {
      const wrapper = createWrapper({}, renderPanelStubs);
      wrapper.vm.drawerTab = EWorkloadDrawerTabs.Pods;
      await flushPromises();

      const podsTab = wrapper.findComponent({ name: "WorkloadPodsTab" });
      expect(podsTab.props("pods")).toEqual([]);
      expect(podsTab.props("loading")).toBe(false);
    });

    it("should load pods when switching to the Pods tab", async () => {
      const wrapper = createWrapper();
      vi.clearAllMocks();

      wrapper.vm.drawerTab = EWorkloadDrawerTabs.Pods;
      await wrapper.vm.$nextTick();

      expect(workloadService.getWorkloadPods).toHaveBeenCalled();
    });
  });
});
