import { describe, it, expect, beforeEach, vi } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { PendingWorkloadsWidget } from "@/components/dashboard-v2/widgets/stat-widgets/pending-workloads-widget/index";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { EOrgUnitMetricSourceType, type IOrgUnitMetricSource } from "@/models/org-unit.model";

vi.mock("@/services/cluster/workload.service/workload.service");

const mockRouter = {
  push: vi.fn(),
};

let wrapper: VueWrapper<InstanceType<typeof PendingWorkloadsWidget>>;
const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });
const clusterId = "34234-322-sdfs323";

describe("PendingWorkloadsWidget component", () => {
  const createComponent = (props?: { nodePoolName?: string; orgUnitMetricSource?: IOrgUnitMetricSource | null }) => {
    wrapper = mount(PendingWorkloadsWidget, {
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $route: {},
          $router: mockRouter,
        },
      },
      props: {
        ...props,
        clusterId,
      },
    });

    wrapper.vm.$q.notify = vi.fn(); // disabling notify to prevent it from failing tests
  };

  beforeEach(async () => {
    vi.clearAllMocks();
  });

  const getStatData = () => {
    return wrapper.find(".stat-data").text();
  };

  const clickOnGridLink = async () => {
    await wrapper.find(".widget-link-btn").trigger("click");
  };

  it("should render PendingWorkloadsWidget component", async () => {
    vi.mocked(workloadService.getWorkloadsCount).mockImplementation(() => Promise.resolve(144));
    createComponent();
    await flushPromises();
    expect(getStatData()).toBe("144");
  });

  it("should render PendingWorkloadsWidget component with filter", async () => {
    vi.mocked(workloadService.getWorkloadsCount).mockImplementation(() => Promise.resolve(16));
    createComponent({ nodePoolName: "Node pool" });
    await flushPromises();
    expect(getStatData()).toBe("16");
  });

  it("should link to nodes table when clicked", async () => {
    vi.mocked(workloadService.getWorkloadsCount).mockImplementation(() => Promise.resolve(16));
    createComponent();
    await flushPromises();
    await clickOnGridLink();
    expect(mockRouter.push).toHaveBeenCalledWith({ name: "workload-index" });
  });

  it("should trigger get telemetry data on filter change", async () => {
    vi.mocked(workloadService.getWorkloadsCount).mockImplementation(() => Promise.resolve(16));
    createComponent();
    await flushPromises();
    await wrapper.setProps({ nodePoolName: "Node pool" });
    expect(workloadService.getWorkloadsCount).toHaveBeenCalledTimes(2);
  });

  it("should display loading skeleton when loading", async () => {
    vi.mocked(workloadService.getWorkloadsCount).mockImplementation(() => Promise.resolve(16));
    createComponent();
    expect(wrapper.findComponent(".q-skeleton").exists()).toEqual(true);
    await flushPromises();
    expect(wrapper.findComponent(".q-skeleton").exists()).toEqual(false);
  });

  it("should display error message when telemetry fetch fails", async () => {
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    vi.mocked(workloadService.getWorkloadsCount).mockRejectedValueOnce({ response: { status: 500 } });
    createComponent();
    await flushPromises();
    expect(wrapper.find(".error-section").exists()).toEqual(true);
    expect(wrapper.find(".error-section").text()).eql("Data is not available right now");
    errorSpy.mockRestore();
  });

  it("should include departmentId filter when orgUnitMetricSource is department", async () => {
    vi.mocked(workloadService.getWorkloadsCount).mockResolvedValue(5);
    const source: IOrgUnitMetricSource = { type: EOrgUnitMetricSourceType.Department, id: "dept-123" };
    createComponent({ orgUnitMetricSource: source });
    await flushPromises();

    const filterArg = vi.mocked(workloadService.getWorkloadsCount).mock.calls[0][0] as string[];
    expect(filterArg).toEqual(
      expect.arrayContaining([expect.stringContaining("departmentId"), expect.stringContaining("dept-123")]),
    );
    expect(filterArg.join(",")).not.toContain("projectId");
  });

  it("should include projectId filter when orgUnitMetricSource is project", async () => {
    vi.mocked(workloadService.getWorkloadsCount).mockResolvedValue(3);
    const source: IOrgUnitMetricSource = { type: EOrgUnitMetricSourceType.Project, id: "proj-456" };
    createComponent({ orgUnitMetricSource: source });
    await flushPromises();

    const filterArg = vi.mocked(workloadService.getWorkloadsCount).mock.calls[0][0] as string[];
    expect(filterArg).toEqual(
      expect.arrayContaining([expect.stringContaining("projectId"), expect.stringContaining("proj-456")]),
    );
    expect(filterArg.join(",")).not.toContain("departmentId");
  });

  it("should not include org-unit filters when orgUnitMetricSource is null", async () => {
    vi.mocked(workloadService.getWorkloadsCount).mockResolvedValue(10);
    createComponent({ orgUnitMetricSource: null });
    await flushPromises();

    const filterArg = vi.mocked(workloadService.getWorkloadsCount).mock.calls[0][0] as string[];
    const joined = filterArg.join(",");
    expect(joined).not.toContain("departmentId");
    expect(joined).not.toContain("projectId");
  });

  it("should reload when orgUnitMetricSource changes", async () => {
    vi.mocked(workloadService.getWorkloadsCount).mockResolvedValue(10);
    createComponent();
    await flushPromises();

    const source: IOrgUnitMetricSource = { type: EOrgUnitMetricSourceType.Department, id: "dept-1" };
    await wrapper.setProps({ orgUnitMetricSource: source });
    await flushPromises();

    expect(workloadService.getWorkloadsCount).toHaveBeenCalledTimes(2);
  });
});
