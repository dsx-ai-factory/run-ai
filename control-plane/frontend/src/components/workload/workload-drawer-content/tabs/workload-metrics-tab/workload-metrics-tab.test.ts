import type { Mock } from "vitest";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import WorkloadMetricsTab from "./workload-metrics-tab.vue";
import { Phase, HistoryRecordType, type Workload } from "@/swagger-models/workloads-client";
import { PREDEFINED_DATES } from "@/models/date.model";
import { intervalUtil } from "@/utils/interval.util";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { EMetricsType } from "@/models/node.model";
import { EWorkloadType } from "@/models/workload.model";

const { mockIsClusterVersionSufficient } = vi.hoisted(() => ({
  mockIsClusterVersionSufficient: vi.fn(),
}));

vi.mock("@/utils/interval.util", () => ({
  intervalUtil: {
    startInterval: vi.fn(),
    stopInterval: vi.fn(),
  },
}));

vi.mock("@/services/workloads.service", () => ({
  default: {
    getWorkloadPods: vi.fn().mockResolvedValue([]),
  },
  workloadService: {
    getWorkloadPods: vi.fn().mockResolvedValue([]),
  },
}));

vi.mock("@/services/cluster/workload.service/workload.service", () => ({
  workloadService: {
    getWorkloadPods: vi.fn().mockResolvedValue([]),
    getWorkloadHistory: vi.fn().mockResolvedValue({ records: [] }),
  },
}));

vi.mock("@/stores/cluster.store", () => ({
  useClusterStore: () => ({
    cluster: {
      uuid: "test-cluster-uuid",
    },
    getClusterById: vi.fn().mockReturnValue({ version: "3.0" }),
    isClusterVersionSufficient: mockIsClusterVersionSufficient,
  }),
}));

describe("WorkloadMetricsTab", () => {
  let wrapper: VueWrapper<InstanceType<typeof WorkloadMetricsTab>>;
  let mockWorkload: Workload;

  let mockStopInterval: Mock;

  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
    mockIsClusterVersionSufficient.mockReturnValue(false);

    mockStopInterval = intervalUtil.stopInterval as Mock;

    mockWorkload = {
      id: "test-workload-id",
      name: "test-workload",
      phase: Phase.Running,
      type: "Training",
      createdAt: new Date("2024-01-01T00:00:00Z").toISOString(),
      phaseUpdatedAt: new Date("2024-01-02T00:00:00Z").toISOString(),
      clusterId: "test-cluster",
    } as Workload;
  });

  const createWrapper = (props = {}) => {
    return mount(WorkloadMetricsTab, {
      props: {
        workload: mockWorkload,
        drawerWidth: "800",
        ...props,
      },
      global: {
        stubs: {
          RunaiDatepickerWithPredefined: true,
          RunaiSelect: true,
          DefaultWorkloadWidgets: true,
          LatencyThroughputAndReplicasWidget: true,
          AdvancedWorkloadMetricWidget: true,
          NimWorkloadWidgets: true,
          VllmWorkloadWidgets: true,
        },
        mocks: {
          $q: {
            notify: vi.fn(),
          },
        },
      },
    });
  };

  describe("Component Initialization", () => {
    it("should initialize with correct default values", () => {
      wrapper = createWrapper();

      expect(wrapper.vm.liveEndDate).toBeInstanceOf(Date);
      expect(wrapper.vm.userActionTimestamp).toBe(0);
      expect(wrapper.vm.startDate).toBeInstanceOf(Date);
      expect(wrapper.vm.endDate).toBeInstanceOf(Date);
    });

    it("should stop intervals on unmount", () => {
      wrapper = createWrapper();
      wrapper.unmount();

      expect(mockStopInterval).toHaveBeenCalled();
    });
  });

  describe("vLLM metrics", () => {
    const createVllmWorkload = (): Workload =>
      ({
        ...mockWorkload,
        type: EWorkloadType.Inference,
        images: ["vllm/vllm-openai:latest"],
        additionalFields: { vllmWorkload: "true" },
      }) as Workload;

    it("shows vLLM metrics for supported vLLM inference workloads", () => {
      mockIsClusterVersionSufficient.mockImplementation((_clusterId, version) => version === "3.0");
      wrapper = createWrapper({ workload: createVllmWorkload() });

      expect(wrapper.vm.metricTypeOptions).toContain(EMetricsType.Vllm);
    });

    it("hides vLLM metrics below cluster 3.0", () => {
      wrapper = createWrapper({ workload: createVllmWorkload() });

      expect(wrapper.vm.metricTypeOptions).not.toContain(EMetricsType.Vllm);
    });

    it("hides vLLM metrics for image-only workloads", () => {
      mockIsClusterVersionSufficient.mockImplementation((_clusterId, version) => version === "3.0");
      const imageOnlyVllmWorkload = {
        ...createVllmWorkload(),
        additionalFields: undefined,
      } as Workload;
      wrapper = createWrapper({ workload: imageOnlyVllmWorkload });

      expect(wrapper.vm.metricTypeOptions).not.toContain(EMetricsType.Vllm);
    });

    it("keeps NIM selection independent from vLLM gating", () => {
      mockIsClusterVersionSufficient.mockImplementation((_clusterId, version) => version === "2.23");
      wrapper = createWrapper({
        workload: {
          ...createVllmWorkload(),
          additionalFields: { nimWorkload: "true", vllmWorkload: "true" },
        } as Workload,
      });

      expect(wrapper.vm.metricTypeOptions).toContain(EMetricsType.NvidiaNim);
      expect(wrapper.vm.metricTypeOptions).not.toContain(EMetricsType.Vllm);
    });

    it("resets vLLM selection when workload updates remove the option", async () => {
      mockIsClusterVersionSufficient.mockImplementation((_clusterId, version) => version === "3.0");
      wrapper = createWrapper({ workload: createVllmWorkload() });
      wrapper.vm.selectedMetricType = EMetricsType.Vllm;
      await wrapper.vm.$nextTick();

      expect(wrapper.find("vllm-workload-widgets-stub").exists()).toBe(true);

      await wrapper.setProps({
        workload: {
          ...mockWorkload,
          id: "regular-workload-id",
        },
      });

      expect(wrapper.vm.metricTypeOptions).not.toContain(EMetricsType.Vllm);
      expect(wrapper.vm.selectedMetricType).toBe(EMetricsType.Default);
      expect(wrapper.find("vllm-workload-widgets-stub").exists()).toBe(false);
    });
  });

  describe("workloadLifespanEndDate Computed Property", () => {
    it("should return liveEndDate for running workload", () => {
      wrapper = createWrapper();
      wrapper.vm.dateLabel = PREDEFINED_DATES.WORKLOADS_LIFESPAN;

      const endDate = wrapper.vm.workloadLifespanEndDate;
      expect(endDate).toBe(wrapper.vm.liveEndDate);
    });

    it("should return phaseUpdatedAt for completed workload", () => {
      mockWorkload.phase = Phase.Completed;
      wrapper = createWrapper();

      const endDate = wrapper.vm.workloadLifespanEndDate;
      expect(endDate.toISOString()).toBe(mockWorkload.phaseUpdatedAt);
    });

    it("should return phaseUpdatedAt for failed workload", () => {
      mockWorkload.phase = Phase.Failed;
      wrapper = createWrapper();

      const endDate = wrapper.vm.workloadLifespanEndDate;
      expect(endDate.toISOString()).toBe(mockWorkload.phaseUpdatedAt);
    });
  });

  describe("updateDateRange Method", () => {
    it("should update date range and increment userActionTimestamp", () => {
      wrapper = createWrapper();
      const initialTimestamp = wrapper.vm.userActionTimestamp;
      const startDate = new Date("2024-01-01");
      const endDate = new Date("2024-01-07");

      wrapper.vm.updateDateRange([startDate, endDate], "Last 7 days");

      expect(wrapper.vm.startDate).toBe(startDate);
      expect(wrapper.vm.endDate).toBe(endDate);
      expect(wrapper.vm.dateLabel).toBe("Last 7 days");
      expect(wrapper.vm.userActionTimestamp).toBeGreaterThan(initialTimestamp);
    });

    it("should update liveEndDate when end date is today", () => {
      wrapper = createWrapper();
      const startDate = new Date("2024-01-01");
      const endDate = new Date(); // End date is today

      wrapper.vm.updateDateRange([startDate, endDate], PREDEFINED_DATES.WORKLOADS_LIFESPAN);

      // liveEndDate should have been updated because end date is today
      expect(wrapper.vm.dateLabel).toBe(PREDEFINED_DATES.WORKLOADS_LIFESPAN);
      expect(wrapper.vm.liveEndDate).toBeInstanceOf(Date);
    });

    it("should not update liveEndDate when end date is not today", () => {
      wrapper = createWrapper();
      const initialLiveEndDate = wrapper.vm.liveEndDate;
      const startDate = new Date("2024-01-01");
      const endDate = new Date("2024-01-07"); // End date is not today

      wrapper.vm.updateDateRange([startDate, endDate], "Last 7 days");

      expect(wrapper.vm.liveEndDate).toBe(initialLiveEndDate);
    });

    it("should increment userActionTimestamp on date range changes", () => {
      wrapper = createWrapper();
      const timestamp1 = wrapper.vm.userActionTimestamp;

      wrapper.vm.updateDateRange([new Date("2024-01-01"), new Date("2024-01-07")], "Last 7 days");

      const timestamp2 = wrapper.vm.userActionTimestamp;
      expect(timestamp2).toBeGreaterThan(timestamp1);
    });
  });

  describe("filterByDates Computed Property", () => {
    it("should return correct filterByDates for regular date range", () => {
      wrapper = createWrapper();
      wrapper.vm.startDate = new Date("2024-01-01");
      wrapper.vm.endDate = new Date("2024-01-07");
      wrapper.vm.dateLabel = "Last 7 days";

      const filterByDates = wrapper.vm.filterByDates;

      expect(filterByDates.dateStart).toBe(wrapper.vm.startDate);
      expect(filterByDates.dateEnd).toBe(wrapper.vm.endDate);
      expect(filterByDates.label).toBe("Last 7 days");
    });

    it("should use workloadLifespanEndDate when end date is today", () => {
      wrapper = createWrapper();
      wrapper.vm.startDate = new Date("2024-01-01");
      wrapper.vm.endDate = new Date(); // End date is today
      wrapper.vm.dateLabel = PREDEFINED_DATES.WORKLOADS_LIFESPAN;

      const filterByDates = wrapper.vm.filterByDates;

      expect(filterByDates.dateStart).toBe(wrapper.vm.startDate);
      expect(filterByDates.dateEnd).toBe(wrapper.vm.workloadLifespanEndDate);
      expect(filterByDates.label).toBe(PREDEFINED_DATES.WORKLOADS_LIFESPAN);
    });

    it("should maintain stable reference when liveEndDate doesn't change", () => {
      wrapper = createWrapper();
      wrapper.vm.dateLabel = "Last 7 days";

      const filterByDates1 = wrapper.vm.filterByDates;
      const filterByDates2 = wrapper.vm.filterByDates;

      expect(filterByDates1.dateEnd.getTime()).toBe(filterByDates2.dateEnd.getTime());
    });
  });

  describe("Edge Cases", () => {
    it("should update userActionTimestamp on each date range change", () => {
      wrapper = createWrapper();
      const initialTimestamp = wrapper.vm.userActionTimestamp;

      wrapper.vm.updateDateRange([new Date("2024-01-01"), new Date("2024-01-07")], "Custom");

      expect(wrapper.vm.userActionTimestamp).toBeGreaterThan(initialTimestamp);
    });

    it("should handle workload phase transitions", async () => {
      wrapper = createWrapper();
      wrapper.vm.dateLabel = PREDEFINED_DATES.WORKLOADS_LIFESPAN;

      expect(wrapper.vm.workloadLifespanEndDate).toBe(wrapper.vm.liveEndDate);

      await wrapper.setProps({
        workload: {
          ...mockWorkload,
          phase: Phase.Completed,
        },
      });

      expect(wrapper.vm.workloadLifespanEndDate.toISOString()).toBe(mockWorkload.phaseUpdatedAt);
    });

    it("should handle missing phaseUpdatedAt gracefully", () => {
      const workloadWithoutPhaseUpdatedAt = {
        ...mockWorkload,
        phase: Phase.Completed,
        phaseUpdatedAt: "",
      } as Workload;

      expect(() => {
        wrapper = createWrapper({ workload: workloadWithoutPhaseUpdatedAt });
      }).not.toThrow();
    });
  });

  describe("loadWorkloadFirstRunDate", () => {
    it("should set workloadFirstRunDate from the first Running phase in history", async () => {
      const firstRunTimestamp = "2024-01-01T06:00:00.000Z";
      (workloadService.getWorkloadHistory as Mock).mockResolvedValue({
        records: [
          {
            meta: { creationTimestamp: "2024-01-01T00:00:00Z", type: HistoryRecordType.PhaseUpdate },
            spec: { phaseUpdate: { phase: Phase.Pending } },
          },
          {
            meta: { creationTimestamp: firstRunTimestamp, type: HistoryRecordType.PhaseUpdate },
            spec: { phaseUpdate: { phase: Phase.Running } },
          },
          {
            meta: { creationTimestamp: "2024-01-02T00:00:00Z", type: HistoryRecordType.PhaseUpdate },
            spec: { phaseUpdate: { phase: Phase.Running } },
          },
        ],
      });

      wrapper = createWrapper();
      await wrapper.vm.$nextTick();
      await vi.waitFor(() => {
        expect(wrapper.vm.workloadFirstRunDate).toBeDefined();
      });

      expect(wrapper.vm.workloadFirstRunDate?.toISOString()).toBe(firstRunTimestamp);
    });

    it("should pick the earliest Running record when history is unsorted", async () => {
      const earliestRunTimestamp = "2024-01-01T06:00:00.000Z";
      (workloadService.getWorkloadHistory as Mock).mockResolvedValue({
        records: [
          {
            meta: { creationTimestamp: "2024-01-02T00:00:00Z", type: HistoryRecordType.PhaseUpdate },
            spec: { phaseUpdate: { phase: Phase.Running } },
          },
          {
            meta: { creationTimestamp: earliestRunTimestamp, type: HistoryRecordType.PhaseUpdate },
            spec: { phaseUpdate: { phase: Phase.Running } },
          },
        ],
      });

      wrapper = createWrapper();
      await vi.waitFor(() => {
        expect(wrapper.vm.workloadFirstRunDate).toBeDefined();
      });

      expect(wrapper.vm.workloadFirstRunDate?.toISOString()).toBe(earliestRunTimestamp);
    });

    it("should keep workloadFirstRunDate undefined when no Running phase exists", async () => {
      (workloadService.getWorkloadHistory as Mock).mockResolvedValue({
        records: [
          {
            meta: { creationTimestamp: "2024-01-01T00:00:00Z", type: HistoryRecordType.PhaseUpdate },
            spec: { phaseUpdate: { phase: Phase.Pending } },
          },
        ],
      });

      wrapper = createWrapper();
      await wrapper.vm.$nextTick();
      await vi.waitFor(() => {
        expect(workloadService.getWorkloadHistory).toHaveBeenCalled();
      });

      expect(wrapper.vm.workloadFirstRunDate).toBeUndefined();
    });

    it("should keep workloadFirstRunDate undefined when history API fails", async () => {
      (workloadService.getWorkloadHistory as Mock).mockRejectedValue(new Error("API error"));

      wrapper = createWrapper();
      await wrapper.vm.$nextTick();
      await vi.waitFor(() => {
        expect(workloadService.getWorkloadHistory).toHaveBeenCalled();
      });

      expect(wrapper.vm.workloadFirstRunDate).toBeUndefined();
    });

    it("should ignore Event records and only use PhaseUpdate records", async () => {
      const runTimestamp = "2024-01-01T06:00:00.000Z";
      (workloadService.getWorkloadHistory as Mock).mockResolvedValue({
        records: [
          {
            meta: { creationTimestamp: "2024-01-01T01:00:00Z", type: HistoryRecordType.Event },
            spec: { event: { message: "some event" } },
          },
          {
            meta: { creationTimestamp: runTimestamp, type: HistoryRecordType.PhaseUpdate },
            spec: { phaseUpdate: { phase: Phase.Running } },
          },
        ],
      });

      wrapper = createWrapper();
      await vi.waitFor(() => {
        expect(wrapper.vm.workloadFirstRunDate).toBeDefined();
      });

      expect(wrapper.vm.workloadFirstRunDate?.toISOString()).toBe(runTimestamp);
    });
  });

  describe("Integration with Silent Reload Feature", () => {
    it("should support explicit reload pattern for user actions", () => {
      wrapper = createWrapper();

      const initialUserActionTimestamp = wrapper.vm.userActionTimestamp;

      wrapper.vm.updateDateRange([new Date("2024-01-01"), new Date("2024-01-07")], "Last 7 days");

      expect(wrapper.vm.userActionTimestamp).toBeGreaterThan(initialUserActionTimestamp);
    });

    it("should have liveEndDate that can be updated independently", () => {
      wrapper = createWrapper();

      const initialLiveEndDate = wrapper.vm.liveEndDate;
      const initialUserActionTimestamp = wrapper.vm.userActionTimestamp;

      wrapper.vm.liveEndDate = new Date();

      expect(wrapper.vm.liveEndDate).not.toBe(initialLiveEndDate);
      expect(wrapper.vm.userActionTimestamp).toBe(initialUserActionTimestamp);
    });
  });
});
