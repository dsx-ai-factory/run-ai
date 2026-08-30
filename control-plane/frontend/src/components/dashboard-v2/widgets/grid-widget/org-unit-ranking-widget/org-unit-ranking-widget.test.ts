import { describe, expect, vi } from "vitest";
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils";
import { OrgUnitRankingWidget } from "@/components/dashboard-v2/widgets/grid-widget/org-unit-ranking-widget/index";
import { Quasar } from "quasar";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import type { ProjectGetRanking } from "@/swagger-models/org-unit-service-client";
import type { ITableColumn } from "@/models/table.model";
import { EOrgUnitProjectsRankingColumnName } from "@/table-models/org-unit.table-model";
const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });
vi.mock("@/services/control-plane/org-unit.service/org-unit.service");
vi.mock("@/services/cluster/workload.service/workload.service");

const MOCK_PROJECTS_RANKING_RECORDS: ProjectGetRanking[] = [
  {
    name: "Project A",
    parentName: "Department A",
    id: "1",
    allocationValue: 80,
    utilizationValue: 70,
    gpuQuota: 10,
  },
  {
    name: "Project B",
    parentName: "Department B",
    id: "2",
    allocationValue: 70,
    utilizationValue: 60,
    gpuQuota: 5,
  },
];
const mockRouter = {
  push: vi.fn(),
};
vi.mock("vue-router", () => ({
  useRoute: vi.fn(),
  useRouter: vi.fn(() => ({
    push: vi.fn(),
  })),
}));
let wrapper: VueWrapper<InstanceType<typeof OrgUnitRankingWidget>>;
describe("org-unit-ranking-widget", () => {
  const mountComponent = (nodePoolName?: string) => {
    wrapper = mount(OrgUnitRankingWidget, {
      global: {
        plugins: [Quasar, pinia],
        directives: {
          permission: vi.fn(),
        },
        mocks: {
          $route: {},
          $router: mockRouter,
        },
      },
      props: {
        clusterId: "test-cluster",
        nodePoolName,
        hasReadProjectsPermission: true,
        hasReadDepartmentsPermission: true,
      },
    });
  };

  it("renders org unit ranking records correctly", async () => {
    vi.mocked(orgUnitService.getProjectsRanking).mockResolvedValue(MOCK_PROJECTS_RANKING_RECORDS);
    mountComponent();
    await flushPromises();
    expect(wrapper.find('[aid="ranking-table"]').exists()).toBe(true);
  });

  it("displays no data message when there are no ranking records", async () => {
    vi.mocked(orgUnitService.getProjectsRanking).mockResolvedValue([]);
    mountComponent();
    await flushPromises();
    expect(wrapper.findComponent(RunaiTableNoData).exists()).toBe(true);
  });

  it("handles error state correctly", async () => {
    vi.mocked(orgUnitService.getProjectsRanking).mockRejectedValue(new Error("Failed to load data"));
    mountComponent();
    await flushPromises();
    expect(wrapper.find(".error-section").exists()).toBe(true);
  });

  it("applies background color and sort to allocation value column", async () => {
    vi.mocked(orgUnitService.getProjectsRanking).mockResolvedValue(MOCK_PROJECTS_RANKING_RECORDS);
    mountComponent();
    await flushPromises();
    expect(wrapper.find(".light-blue-primary-opacity-background").exists()).toBe(true);
    expect(wrapper.find(".sort-desc").text()).toBe("Avg. GPU devices allocation Last 24harrow_upward");
    expect(wrapper.find("[aid='allocationValue']").text()).toBe("80");
    expect(wrapper.find("[aid='allocationValue']").classes()).includes("light-blue-primary-opacity-background");
  });

  it("should not change sort when clicking on the sort icon", async () => {
    vi.mocked(orgUnitService.getProjectsRanking).mockResolvedValue(MOCK_PROJECTS_RANKING_RECORDS);
    mountComponent();
    await flushPromises();
    const sortIcon = wrapper.find(".sort-desc");
    await sortIcon.trigger("click");
    expect(wrapper.find(".sort-desc").text()).toBe("Avg. GPU devices allocation Last 24harrow_upward");
  });

  it("should get ranking records for a specific node pool", async () => {
    vi.mocked(orgUnitService.getProjectsRanking).mockResolvedValue(MOCK_PROJECTS_RANKING_RECORDS);
    mountComponent("test-node-pool");
    await flushPromises();
    expect(orgUnitService.getProjectsRanking).toHaveBeenCalledWith(
      "avgGpuAllocation24h",
      "top",
      "test-cluster",
      undefined,
    );
  });

  it("displays correct entity options when multiple departments exist", async () => {
    vi.mocked(orgUnitService.countDepartments).mockResolvedValue(2); // Mock multiple departments
    mountComponent();
    await flushPromises();
    const entityOptions = wrapper.vm.entityOptions;
    expect(entityOptions.length).toBeGreaterThan(1);
  });

  it("hides parent name column when only one department exists", async () => {
    vi.mocked(orgUnitService.countDepartments).mockResolvedValue(1);
    mountComponent();
    await flushPromises();
    const parentNameColumn = wrapper.find('[aid="parentNameColumn"]');
    expect(parentNameColumn.exists()).toBe(false); // Ensure column is hidden
  });

  it("shows parent name column when multiple departments exist", async () => {
    vi.mocked(orgUnitService.getProjectsRanking).mockResolvedValue(MOCK_PROJECTS_RANKING_RECORDS);
    vi.mocked(orgUnitService.countDepartments).mockResolvedValue(2);
    mountComponent();
    await flushPromises();
    const parentNameColumn = wrapper.vm.columns.find(
      (col: ITableColumn) => col.name === EOrgUnitProjectsRankingColumnName.ParentName,
    );
    expect(parentNameColumn).toBeDefined();
  });

  it("fetches department count correctly", async () => {
    const departmentCountSpy = vi.spyOn(orgUnitService, "countDepartments");
    mountComponent();
    await flushPromises();
    expect(departmentCountSpy).toHaveBeenCalled();
  });
});
