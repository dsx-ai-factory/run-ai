import { describe, it, expect, beforeEach, vi } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import QuotaTotalGpuWidget from "./quota-total-gpu-widget.vue";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { EOrgUnitMetricSourceType, type IOrgUnitMetricSource } from "@/models/org-unit.model";

vi.mock("@/services/control-plane/org-unit.service/org-unit.service");
vi.mock("@/services/control-plane/cluster.service/cluster.service");

const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });

const mockTelemetryResponse = { type: "GPU_QUOTA", timestamp: "", values: [{ value: "10" }] };
const mockNodeTelemetryResponse = { type: "TOTAL_GPUS", timestamp: "", values: [{ value: "20" }] };
const mockProject = {
  totalResources: { gpuQuota: 15 },
  status: { quotaStatus: { allocated: { gpu: 8 } } },
  resources: [
    { nodePool: { name: "default" }, gpu: { deserved: 5 } },
    { nodePool: { name: "gpu-pool" }, gpu: { deserved: 10 } },
  ],
};
const mockDepartment = {
  totalResources: { gpuQuota: 30 },
  status: { quotaStatus: { allocated: { gpu: 20 } } },
  resources: [
    { nodePool: { name: "default" }, gpu: { deserved: 12 } },
    { nodePool: { name: "gpu-pool" }, gpu: { deserved: 18 } },
  ],
};

describe("QuotaTotalGpuWidget", () => {
  let wrapper: VueWrapper<InstanceType<typeof QuotaTotalGpuWidget>>;

  function createComponent(props?: { orgUnitMetricSource?: IOrgUnitMetricSource | null }) {
    vi.mocked(orgUnitService.getProjectsTelemetry).mockResolvedValue(mockTelemetryResponse as never);
    vi.mocked(clusterService.getNodeTelemetry).mockResolvedValue(mockNodeTelemetryResponse as never);
    vi.mocked(orgUnitService.getProject).mockResolvedValue(mockProject as never);
    vi.mocked(orgUnitService.getDepartment).mockResolvedValue(mockDepartment as never);

    wrapper = mount(QuotaTotalGpuWidget, {
      global: { plugins: [Quasar, pinia] },
      props: { clusterId: "cluster-1", ...props },
    });
  }

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("shows quota/total and percentage in default mode", async () => {
    createComponent();
    await flushPromises();

    expect(wrapper.find(".stat-data").text()).toContain("/");
    expect(wrapper.find(".percentage").exists()).toBe(true);
    expect(clusterService.getNodeTelemetry).toHaveBeenCalled();
    expect(orgUnitService.getProjectsTelemetry).toHaveBeenCalled();
  });

  it("shows only quota and hides percentage in org-unit mode", async () => {
    const source: IOrgUnitMetricSource = { type: EOrgUnitMetricSourceType.Department, id: "dept-1" };
    createComponent({ orgUnitMetricSource: source });
    await flushPromises();

    expect(wrapper.find(".stat-data").text()).not.toContain("/");
    expect(wrapper.find(".percentage").exists()).toBe(false);
  });

  it("changes title to 'GPU quota' in org-unit mode", async () => {
    const source: IOrgUnitMetricSource = { type: EOrgUnitMetricSourceType.Department, id: "dept-1" };
    createComponent({ orgUnitMetricSource: source });
    await flushPromises();

    expect(wrapper.text()).toContain("GPU quota");
    expect(wrapper.text()).not.toContain("GPU quota / total");
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

  it("does not fetch total GPUs in org-unit mode", async () => {
    const source: IOrgUnitMetricSource = { type: EOrgUnitMetricSourceType.Department, id: "dept-1" };
    createComponent({ orgUnitMetricSource: source });
    await flushPromises();

    expect(clusterService.getNodeTelemetry).not.toHaveBeenCalled();
  });

  it("uses telemetry APIs when no org-unit source", async () => {
    createComponent();
    await flushPromises();

    expect(orgUnitService.getProjectsTelemetry).toHaveBeenCalled();
    expect(clusterService.getNodeTelemetry).toHaveBeenCalled();
    expect(orgUnitService.getProject).not.toHaveBeenCalled();
    expect(orgUnitService.getDepartment).not.toHaveBeenCalled();
  });

  it("filters quota by nodepool in org-unit project mode", async () => {
    const source: IOrgUnitMetricSource = { type: EOrgUnitMetricSourceType.Project, id: "proj-1" };
    vi.mocked(orgUnitService.getProject).mockResolvedValue(mockProject as never);
    wrapper = mount(QuotaTotalGpuWidget, {
      global: { plugins: [Quasar, pinia] },
      props: { clusterId: "cluster-1", nodePoolName: "gpu-pool", orgUnitMetricSource: source },
    });
    await flushPromises();

    expect(wrapper.find(".stat-data").text()).toContain("10");
  });

  it("filters quota by nodepool in org-unit department mode", async () => {
    const source: IOrgUnitMetricSource = { type: EOrgUnitMetricSourceType.Department, id: "dept-1" };
    vi.mocked(orgUnitService.getDepartment).mockResolvedValue(mockDepartment as never);
    wrapper = mount(QuotaTotalGpuWidget, {
      global: { plugins: [Quasar, pinia] },
      props: { clusterId: "cluster-1", nodePoolName: "default", orgUnitMetricSource: source },
    });
    await flushPromises();

    expect(wrapper.find(".stat-data").text()).toContain("12");
  });
});
