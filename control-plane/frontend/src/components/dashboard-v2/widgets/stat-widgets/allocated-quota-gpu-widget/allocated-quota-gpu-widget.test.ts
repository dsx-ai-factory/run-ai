import { describe, it, expect, beforeEach, vi } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import AllocatedQuotaGpuWidget from "./allocated-quota-gpu-widget.vue";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { EOrgUnitMetricSourceType, type IOrgUnitMetricSource } from "@/models/org-unit.model";

vi.mock("@/services/control-plane/org-unit.service/org-unit.service");

const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });

const mockTelemetryResponse = { type: "GPU_ALLOCATION", timestamp: "", values: [{ value: "8" }] };
const mockProject = {
  totalResources: { gpuQuota: 15 },
  status: {
    quotaStatus: { allocated: { gpu: 8 } },
    nodePoolQuotaStatuses: [
      { nodePoolName: "default", allocated: { gpu: 3 } },
      { nodePoolName: "gpu-pool", allocated: { gpu: 5 } },
    ],
  },
  resources: [
    { nodePool: { name: "default" }, gpu: { deserved: 5 } },
    { nodePool: { name: "gpu-pool" }, gpu: { deserved: 10 } },
  ],
};
const mockDepartment = {
  totalResources: { gpuQuota: 30 },
  status: {
    quotaStatus: { allocated: { gpu: 20 } },
    nodePoolQuotaStatuses: [
      { nodePoolName: "default", allocated: { gpu: 8 } },
      { nodePoolName: "gpu-pool", allocated: { gpu: 12 } },
    ],
  },
  resources: [
    { nodePool: { name: "default" }, gpu: { deserved: 12 } },
    { nodePool: { name: "gpu-pool" }, gpu: { deserved: 18 } },
  ],
};

describe("AllocatedQuotaGpuWidget", () => {
  let wrapper: VueWrapper<InstanceType<typeof AllocatedQuotaGpuWidget>>;

  function createComponent(props?: { orgUnitMetricSource?: IOrgUnitMetricSource | null }) {
    vi.mocked(orgUnitService.getProjectsTelemetry).mockResolvedValue(mockTelemetryResponse as never);
    vi.mocked(orgUnitService.getProject).mockResolvedValue(mockProject as never);
    vi.mocked(orgUnitService.getDepartment).mockResolvedValue(mockDepartment as never);

    wrapper = mount(AllocatedQuotaGpuWidget, {
      global: { plugins: [Quasar, pinia] },
      props: { clusterId: "cluster-1", ...props },
    });
  }

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("uses telemetry APIs by default", async () => {
    createComponent();
    await flushPromises();

    expect(orgUnitService.getProjectsTelemetry).toHaveBeenCalled();
    expect(orgUnitService.getProject).not.toHaveBeenCalled();
    expect(orgUnitService.getDepartment).not.toHaveBeenCalled();
  });

  it("calls getDepartment for department source", async () => {
    const source: IOrgUnitMetricSource = { type: EOrgUnitMetricSourceType.Department, id: "dept-1" };
    createComponent({ orgUnitMetricSource: source });
    await flushPromises();

    expect(orgUnitService.getDepartment).toHaveBeenCalledWith("dept-1");
    expect(orgUnitService.getProject).not.toHaveBeenCalled();
    expect(orgUnitService.getProjectsTelemetry).not.toHaveBeenCalled();
  });

  it("calls getProject for project source", async () => {
    const source: IOrgUnitMetricSource = { type: EOrgUnitMetricSourceType.Project, id: "proj-1" };
    createComponent({ orgUnitMetricSource: source });
    await flushPromises();

    expect(orgUnitService.getProject).toHaveBeenCalledWith("proj-1");
    expect(orgUnitService.getDepartment).not.toHaveBeenCalled();
    expect(orgUnitService.getProjectsTelemetry).not.toHaveBeenCalled();
  });

  it("extracts allocated and quota from department response", async () => {
    const source: IOrgUnitMetricSource = { type: EOrgUnitMetricSourceType.Department, id: "dept-1" };
    createComponent({ orgUnitMetricSource: source });
    await flushPromises();

    expect(wrapper.find(".stat-data").text()).toContain("20");
    expect(wrapper.find(".stat-data").text()).toContain("30");
  });

  it("reloads when orgUnitMetricSource changes", async () => {
    createComponent();
    await flushPromises();
    vi.mocked(orgUnitService.getProjectsTelemetry).mockClear();

    const source: IOrgUnitMetricSource = { type: EOrgUnitMetricSourceType.Department, id: "dept-1" };
    await wrapper.setProps({ orgUnitMetricSource: source });
    await flushPromises();

    expect(orgUnitService.getDepartment).toHaveBeenCalledWith("dept-1");
  });

  it("filters by nodepool in org-unit project mode", async () => {
    const source: IOrgUnitMetricSource = { type: EOrgUnitMetricSourceType.Project, id: "proj-1" };
    vi.mocked(orgUnitService.getProject).mockResolvedValue(mockProject as never);
    wrapper = mount(AllocatedQuotaGpuWidget, {
      global: { plugins: [Quasar, pinia] },
      props: { clusterId: "cluster-1", nodePoolName: "gpu-pool", orgUnitMetricSource: source },
    });
    await flushPromises();

    expect(wrapper.find(".stat-data").text()).toContain("5");
    expect(wrapper.find(".stat-data").text()).toContain("10");
  });

  it("filters by nodepool in org-unit department mode", async () => {
    const source: IOrgUnitMetricSource = { type: EOrgUnitMetricSourceType.Department, id: "dept-1" };
    vi.mocked(orgUnitService.getDepartment).mockResolvedValue(mockDepartment as never);
    wrapper = mount(AllocatedQuotaGpuWidget, {
      global: { plugins: [Quasar, pinia] },
      props: { clusterId: "cluster-1", nodePoolName: "default", orgUnitMetricSource: source },
    });
    await flushPromises();

    expect(wrapper.find(".stat-data").text()).toContain("8");
    expect(wrapper.find(".stat-data").text()).toContain("12");
  });
});
