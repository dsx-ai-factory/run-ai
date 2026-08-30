import { describe, it, expect, beforeEach, vi } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { ProjectAllocationRatioWidget } from "./index";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { RankingMetricType, RankingOrderType } from "@/swagger-models/org-unit-service-client";
import { EOrgUnitMetricSourceType, type IOrgUnitMetricSource } from "@/models/org-unit.model";

vi.mock("@/services/control-plane/org-unit.service/org-unit.service");

const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });
const clusterId = "cluster-123";

const mockRanking = [
  { id: "1", name: "proj-a", allocationValue: 0.1 },
  { id: "2", name: "proj-b", allocationValue: 0.3 },
];

class ResizeObserverMock {
  observe(): void {}

  unobserve(): void {}

  disconnect(): void {}
}

describe("ProjectAllocationRatioWidget", () => {
  let wrapper: VueWrapper<InstanceType<typeof ProjectAllocationRatioWidget>>;

  const createComponent = (props?: { orgUnitMetricSource?: IOrgUnitMetricSource | null }) => {
    wrapper = mount(ProjectAllocationRatioWidget, {
      global: { plugins: [Quasar, pinia] },
      props: { clusterId, ...props },
    });
  };

  beforeEach(() => {
    vi.stubGlobal("ResizeObserver", ResizeObserverMock);
    vi.clearAllMocks();
  });

  it("fetches projects ranking with gpuAllocationRatio on mount", async () => {
    vi.mocked(orgUnitService.getProjectsRanking).mockResolvedValue(mockRanking as never);
    createComponent();
    await flushPromises();

    expect(orgUnitService.getProjectsRanking).toHaveBeenCalledWith(
      RankingMetricType.GpuAllocationRatio,
      RankingOrderType.Bottom,
      clusterId,
      undefined,
      undefined,
    );
  });

  it("passes departmentId from orgUnitMetricSource", async () => {
    vi.mocked(orgUnitService.getProjectsRanking).mockResolvedValue(mockRanking as never);
    const source: IOrgUnitMetricSource = { type: EOrgUnitMetricSourceType.Department, id: "dept-1" };
    createComponent({ orgUnitMetricSource: source });
    await flushPromises();

    expect(orgUnitService.getProjectsRanking).toHaveBeenCalledWith(
      RankingMetricType.GpuAllocationRatio,
      RankingOrderType.Bottom,
      clusterId,
      undefined,
      "dept-1",
    );
  });

  it("shows empty state when no data", async () => {
    vi.mocked(orgUnitService.getProjectsRanking).mockResolvedValue([]);
    createComponent();
    await flushPromises();

    expect(wrapper.find('[aid="empty-widget-message"]').exists()).toBe(true);
  });

  it("shows error state on failure", async () => {
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    vi.mocked(orgUnitService.getProjectsRanking).mockRejectedValue({ response: { status: 500 } });
    createComponent();
    await flushPromises();

    expect(wrapper.find(".error-section").exists()).toBe(true);
    errorSpy.mockRestore();
  });

  it("reloads when orgUnitMetricSource changes", async () => {
    vi.mocked(orgUnitService.getProjectsRanking).mockResolvedValue(mockRanking as never);
    createComponent();
    await flushPromises();

    const source: IOrgUnitMetricSource = { type: EOrgUnitMetricSourceType.Department, id: "dept-1" };
    await wrapper.setProps({ orgUnitMetricSource: source });
    await flushPromises();

    expect(orgUnitService.getProjectsRanking).toHaveBeenCalledTimes(2);
  });
});
