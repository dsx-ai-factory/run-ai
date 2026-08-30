import { describe, it, expect, beforeEach, vi } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { PendingWorkloadsByNodepoolWidget } from "./index";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { mockPendingByNodepoolTelemetry, mockEmptyTelemetry } from "./pending-workloads-by-nodepool-widget.mock";
import { Phase, WorkloadTelemetryType } from "@/swagger-models/workloads-client";
import { EWorkloadTelemetryGroupBy } from "@/models/workload.model";
import { EOrgUnitMetricSourceType, type IOrgUnitMetricSource } from "@/models/org-unit.model";

vi.mock("@/services/cluster/workload.service/workload.service");

let wrapper: VueWrapper<InstanceType<typeof PendingWorkloadsByNodepoolWidget>>;
const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });
const clusterId = "cluster-123";

describe("PendingWorkloadsByNodepoolWidget", () => {
  const createComponent = (props?: { nodePoolName?: string; orgUnitMetricSource?: IOrgUnitMetricSource | null }) => {
    wrapper = mount(PendingWorkloadsByNodepoolWidget, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        clusterId,
        hasDepartments: true,
        ...props,
      },
    });
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders and fetches telemetry on mount with ClusterId and RequestedNodepools groupBy", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockResolvedValue(mockPendingByNodepoolTelemetry);
    createComponent();
    await flushPromises();

    expect(workloadService.getWorkloadsTelemetry).toHaveBeenCalledWith(
      WorkloadTelemetryType.WorkloadsCount,
      clusterId,
      undefined,
      undefined,
      undefined,
      [EWorkloadTelemetryGroupBy.ClusterId, EWorkloadTelemetryGroupBy.RequestedNodepools],
      [Phase.Pending],
    );
  });

  it("passes nodePoolName to telemetry API", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockResolvedValue(mockPendingByNodepoolTelemetry);
    createComponent({ nodePoolName: "gpu-pool" });
    await flushPromises();

    expect(workloadService.getWorkloadsTelemetry).toHaveBeenCalledWith(
      WorkloadTelemetryType.WorkloadsCount,
      clusterId,
      "gpu-pool",
      undefined,
      undefined,
      [EWorkloadTelemetryGroupBy.ClusterId, EWorkloadTelemetryGroupBy.RequestedNodepools],
      [Phase.Pending],
    );
  });

  it("passes departmentId from orgUnitMetricSource", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockResolvedValue(mockPendingByNodepoolTelemetry);
    const source: IOrgUnitMetricSource = { type: EOrgUnitMetricSourceType.Department, id: "dept-1" };
    createComponent({ orgUnitMetricSource: source });
    await flushPromises();

    expect(workloadService.getWorkloadsTelemetry).toHaveBeenCalledWith(
      WorkloadTelemetryType.WorkloadsCount,
      clusterId,
      undefined,
      "dept-1",
      undefined,
      [EWorkloadTelemetryGroupBy.ClusterId, EWorkloadTelemetryGroupBy.RequestedNodepools],
      [Phase.Pending],
    );
  });

  it("shows empty state when no data", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockResolvedValue(mockEmptyTelemetry);
    createComponent();
    await flushPromises();

    expect(wrapper.find('[aid="empty-widget-message"]').exists()).toBe(true);
  });

  it("shows error state on fetch failure", async () => {
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    vi.mocked(workloadService.getWorkloadsTelemetry).mockRejectedValue({ response: { status: 500 } });
    createComponent();
    await flushPromises();

    expect(wrapper.find(".error-section").exists()).toBe(true);
    errorSpy.mockRestore();
  });

  it("reloads on nodePoolName change", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockResolvedValue(mockPendingByNodepoolTelemetry);
    createComponent();
    await flushPromises();

    await wrapper.setProps({ nodePoolName: "cpu-pool" });
    await flushPromises();

    expect(workloadService.getWorkloadsTelemetry).toHaveBeenCalledTimes(2);
  });

  it("reloads on orgUnitMetricSource change", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockResolvedValue(mockPendingByNodepoolTelemetry);
    createComponent();
    await flushPromises();

    const source: IOrgUnitMetricSource = { type: EOrgUnitMetricSourceType.Department, id: "dept-1" };
    await wrapper.setProps({ orgUnitMetricSource: source });
    await flushPromises();

    expect(workloadService.getWorkloadsTelemetry).toHaveBeenCalledTimes(2);
  });

  it("shows loading skeleton initially", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockResolvedValue(mockPendingByNodepoolTelemetry);
    createComponent();

    expect(wrapper.findComponent(".chart-widget-skeleton").exists()).toBe(true);
    await flushPromises();
    expect(wrapper.findComponent(".chart-widget-skeleton").exists()).toBe(false);
  });
});
