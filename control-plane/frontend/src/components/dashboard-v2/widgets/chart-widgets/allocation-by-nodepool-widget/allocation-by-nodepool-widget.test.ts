import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { AllocationByNodepoolWidget } from "./index";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { OrgUnitTelemetryType } from "@/swagger-models/org-unit-service-client";
import { EOrgUnitMetricSourceType, type IOrgUnitMetricSource } from "@/models/org-unit.model";

vi.mock("@/services/control-plane/org-unit.service/org-unit.service");

const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });
const clusterId = "cluster-123";

const mockTelemetry = {
  type: "GPU_ALLOCATION",
  timestamp: "2024-03-01T00:00:00.000Z",
  values: [
    { value: "8", groups: [{ key: "NodepoolId", value: "np-1", name: "default" }] },
    { value: "3", groups: [{ key: "NodepoolId", value: "np-2", name: "gpu-pool" }] },
  ],
};

const mockEmptyTelemetry = { type: "GPU_ALLOCATION", timestamp: "2024-03-01T00:00:00.000Z", values: [] };

describe("AllocationByNodepoolWidget", () => {
  let wrapper: VueWrapper<InstanceType<typeof AllocationByNodepoolWidget>>;

  const createComponent = (props?: { nodePoolName?: string; orgUnitMetricSource?: IOrgUnitMetricSource | null }) => {
    wrapper = mount(AllocationByNodepoolWidget, {
      global: { plugins: [Quasar, pinia] },
      props: { clusterId, hasDepartments: true, ...props },
    });
  };

  afterEach(() => {
    wrapper?.unmount();
  });

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("fetches GPU_ALLOCATION telemetry with NodepoolId groupBy on mount", async () => {
    vi.mocked(orgUnitService.getProjectsTelemetry).mockResolvedValue(mockTelemetry as never);
    createComponent();
    await flushPromises();

    expect(orgUnitService.getProjectsTelemetry).toHaveBeenCalledWith(
      OrgUnitTelemetryType.GpuAllocation,
      clusterId,
      undefined,
      undefined,
      undefined,
      ["ClusterId", "NodepoolId"],
    );
  });

  it("passes departmentId from orgUnitMetricSource", async () => {
    vi.mocked(orgUnitService.getProjectsTelemetry).mockResolvedValue(mockTelemetry as never);
    const source: IOrgUnitMetricSource = { type: EOrgUnitMetricSourceType.Department, id: "dept-1" };
    createComponent({ orgUnitMetricSource: source });
    await flushPromises();

    expect(orgUnitService.getProjectsTelemetry).toHaveBeenCalledWith(
      OrgUnitTelemetryType.GpuAllocation,
      clusterId,
      undefined,
      "dept-1",
      undefined,
      ["ClusterId", "NodepoolId"],
    );
  });

  it("shows empty state when no data", async () => {
    vi.mocked(orgUnitService.getProjectsTelemetry).mockResolvedValue(mockEmptyTelemetry as never);
    createComponent();
    await flushPromises();

    expect(wrapper.find('[aid="empty-widget-message"]').exists()).toBe(true);
  });

  it("shows error state on fetch failure", async () => {
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    vi.mocked(orgUnitService.getProjectsTelemetry).mockRejectedValue({ response: { status: 500 } });
    createComponent();
    await flushPromises();

    expect(wrapper.find(".error-section").exists()).toBe(true);
    errorSpy.mockRestore();
  });

  it("reloads on filter change", async () => {
    vi.mocked(orgUnitService.getProjectsTelemetry).mockResolvedValue(mockTelemetry as never);
    createComponent();
    await flushPromises();

    await wrapper.setProps({ nodePoolName: "gpu-pool" });
    await flushPromises();

    expect(orgUnitService.getProjectsTelemetry).toHaveBeenCalledTimes(2);
  });

  it("ignores an older response after the node-pool filter starts a newer request", async () => {
    let resolveFirstRequest!: (value: typeof mockEmptyTelemetry) => void;
    const firstRequest = new Promise<typeof mockEmptyTelemetry>((resolve) => {
      resolveFirstRequest = resolve;
    });
    vi.mocked(orgUnitService.getProjectsTelemetry)
      .mockImplementationOnce(() => firstRequest as never)
      .mockResolvedValueOnce(mockTelemetry as never);

    createComponent();
    await vi.waitFor(() => expect(orgUnitService.getProjectsTelemetry).toHaveBeenCalledTimes(1));
    await wrapper.setProps({ nodePoolName: "gpu-pool" });
    await flushPromises();

    resolveFirstRequest(mockEmptyTelemetry);
    await flushPromises();

    expect(wrapper.find('[aid="empty-widget-message"]').exists()).toBe(false);
    expect(wrapper.find(".error-section").exists()).toBe(false);
  });
});
