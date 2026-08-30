import { afterEach, beforeEach, describe, expect, type MockedFunction, test, vi } from "vitest";
import { Quasar } from "quasar";
import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";

// services
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
// utils
import { unitTestUtil } from "@/utils/unit-test.util";
// components
import NodeListModal from "@/components/node/node-list-modal/node-list-modal.vue";
// mocks
import { mockNodes } from "@/components/node/node-list-modal/node-list-modal.mock";
// models
import type { INodeModalOptions } from "@/models/node.model";
import { NodeFilterSortFields } from "@/swagger-models/cluster-service-client";

const clusterId = "test-cluster-123";
const nodePoolName = "gpu-pool";

vi.mock("@/services/control-plane/cluster.service/cluster.service");

describe("node-list-modal component", () => {
  let wrapper: VueWrapper<InstanceType<typeof NodeListModal>>;

  const createComponent = (props: { options: INodeModalOptions }): void => {
    wrapper = mount(NodeListModal, {
      global: {
        plugins: [Quasar],
        directives: {
          permission: vi.fn(),
          clickOutside: vi.fn(),
        },
      },
      props: props,
    });
    wrapper.vm.$q.notify = vi.fn(); // disabling notify to prevent it from failing tests
  };

  // Mock API calls
  const mockCountTenantNodes = clusterService.countTenantNodes as MockedFunction<typeof clusterService.countTenantNodes>;
  const mockGetTenantNodes = clusterService.getTenantNodes as MockedFunction<typeof clusterService.getTenantNodes>;
  const mockGetNodeTelemetries = clusterService.getNodeTelemetries as MockedFunction<
    typeof clusterService.getNodeTelemetries
  >;

  const getTenantNodesSpy = vi.spyOn(clusterService, "getTenantNodes");
  const countTenantNodesSpy = vi.spyOn(clusterService, "countTenantNodes");

  beforeEach(async () => {
    mockCountTenantNodes.mockResolvedValue(mockNodes.length);
    mockGetNodeTelemetries.mockResolvedValue([
      {
        type: "READY_GPUS",
        timestamp: "2025-10-26T12:11:24.942236434Z",
        values: [
          {
            value: "16",
            groups: [
              {
                key: "Node",
                value: "runai-cluster-worker-0-0",
              },
              {
                key: "ClusterId",
                value: "b651d788-cb71-49b7-a46e-a7eb873c5a1c",
              },
            ],
          },
        ],
      },
    ]);
    mockGetTenantNodes.mockResolvedValue(mockNodes);

    createComponent({
      options: {
        clusterId: clusterId,
        nodePoolName: nodePoolName,
      },
    });

    await flushPromises();
  });

  afterEach(() => {
    wrapper.unmount();
    vi.clearAllMocks();
  });

  test("Modal should render with correct header", () => {
    expect(wrapper.findComponent(".runai-base-dialog-header").text()).toContain(
      `Nodes Associated with Node pool ${nodePoolName}`,
    );
  });

  test("Modal should load nodes with correct filters on mount", async () => {
    expect(getTenantNodesSpy).toHaveBeenCalledWith(
      [`${NodeFilterSortFields.ClusterUuid}==${clusterId}`, `${NodeFilterSortFields.NodePool}==${nodePoolName}`],
      undefined,
      "asc",
      0,
      20,
      undefined,
    );
  });

  test("Modal should count nodes with correct filters", async () => {
    expect(countTenantNodesSpy).toHaveBeenCalledWith(
      [`${NodeFilterSortFields.ClusterUuid}==${clusterId}`, `${NodeFilterSortFields.NodePool}==${nodePoolName}`],
      "",
    );
  });

  test("Table should display node data correctly", () => {
    const modalEl = unitTestUtil.getModalElement();
    const tableRows = modalEl.findAll(".runai-table-row");

    expect(tableRows.length).toBe(mockNodes.length);

    // Check if node names are rendered
    modalEl.findAll('[aid="name"]').forEach((el, index) => {
      expect(el.text()).toBe(mockNodes[index].name);
    });
  });

  test("Search functionality should trigger reload with search term", async () => {
    const modalEl = unitTestUtil.getModalElement();
    const searchInput = modalEl.find("[aid=runai-list-modal-search]");

    // Clear previous calls
    getTenantNodesSpy.mockClear();

    // Set search term
    await searchInput.setValue("node-gpu-1");
    await flushPromises();

    // Should call getTenantNodes with search term
    expect(getTenantNodesSpy).toHaveBeenCalledWith(
      [`${NodeFilterSortFields.ClusterUuid}==${clusterId}`, `${NodeFilterSortFields.NodePool}==${nodePoolName}`],
      undefined,
      "asc",
      0,
      20,
      "node-gpu-1",
    );
  });

  test("Pagination should trigger reload without recounting", async () => {
    getTenantNodesSpy.mockClear();
    countTenantNodesSpy.mockClear();

    wrapper.vm.tableFilter = {
      ...wrapper.vm.tableFilter,
      page: 2,
    };

    await wrapper.vm.updateFilter(wrapper.vm.tableFilter);
    await flushPromises();

    expect(getTenantNodesSpy).toHaveBeenCalled();
    expect(countTenantNodesSpy).toHaveBeenCalled();
  });

  test("Close button should emit close event", async () => {
    const modalEl = unitTestUtil.getModalElement();
    const closeButton = modalEl.find(".footer > button");

    await closeButton.trigger("click");

    expect(wrapper.emitted("close")).toBeTruthy();
    expect(wrapper.emitted("close")).toHaveLength(1);
  });

  test("Modal should handle empty node list", async () => {
    wrapper.unmount();
    mockGetTenantNodes.mockResolvedValue([]);
    mockCountTenantNodes.mockResolvedValue(0);

    createComponent({
      options: {
        clusterId: clusterId,
        nodePoolName: nodePoolName,
      },
    });

    await flushPromises();

    const modalEl = unitTestUtil.getModalElement();
    const noDataMessage = modalEl.find(".empty-state-message");

    expect(noDataMessage.exists()).toBe(true);
  });

  test("Modal should show loading state during data fetch", async () => {
    wrapper.unmount();

    // Create a promise that we control
    // eslint-disable-next-line @typescript-eslint/no-empty-function
    let resolvePromise: (value: typeof mockNodes) => void = () => {};
    const pendingPromise = new Promise<typeof mockNodes>((resolve) => {
      resolvePromise = resolve;
    });

    mockGetTenantNodes.mockReturnValue(pendingPromise);

    createComponent({
      options: {
        clusterId: clusterId,
        nodePoolName: nodePoolName,
      },
    });

    expect(wrapper.vm.loading).toBe(true);

    resolvePromise(mockNodes);
    await flushPromises();

    expect(wrapper.vm.loading).toBe(false);
  });

  test("Clear search should reload nodes without search term", async () => {
    const modalEl = unitTestUtil.getModalElement();
    const searchInput = modalEl.find("[aid=runai-list-modal-search]");

    await searchInput.setValue("node-gpu-1");
    await flushPromises();

    getTenantNodesSpy.mockClear();

    await modalEl.find("[aid=runai-list-modal-search-clear]").trigger("click");
    await flushPromises();

    expect(getTenantNodesSpy).toHaveBeenCalledWith(
      [`${NodeFilterSortFields.ClusterUuid}==${clusterId}`, `${NodeFilterSortFields.NodePool}==${nodePoolName}`],
      undefined,
      "asc",
      0,
      20,
      undefined,
    );
  });
});
