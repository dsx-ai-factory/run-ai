import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import ClusterNew from "./cluster-new.vue";
import { FAKE_ACCESS_TOKEN_A } from "@/tests/fixtures/auth.fixtures";
import { afterEach, beforeEach, describe, expect, test, vi, type MockedFunction } from "vitest";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import { useAuthStore } from "@/stores/auth.store";
import { useClusterStore } from "@/stores/cluster.store";
import { EClusterDistIds, EClusterLocations } from "@/models/cluster.model";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { CLUSTER_ROUTE_NAMES } from "@/router/cluster.routes/cluster.routes.names";
import { ClusterState, type DisplayedCluster } from "@/swagger-models/cluster-service-client";
import { useAppStore } from "@/stores/app.store";

vi.mock("@/services/control-plane/cluster.service/cluster.service");
vi.mock("@/services/cluster/cluster-api.service/cluster-api.service");

const accessToken = FAKE_ACCESS_TOKEN_A;

const pinia = createTestingPinia({ createSpy: vi.fn });
const authStore = useAuthStore(pinia);
const clusterStore = useClusterStore(pinia);
const appStore = useAppStore(pinia);

const clusters: DisplayedCluster[] = [
  {
    createdAt: "2023-06-14T11:19:48.755Z",
    domain: "https://dev-some-sh-no-cluster.runailabs.com",
    name: "some-test",
    tenantId: 1,
    uuid: "35458207-de64-49f0-8f9b-79adf7ca2072",
    version: "2.11.0",
    status: {
      state: ClusterState.Connected,
    },
    updatedAt: "2023-06-14T11:19:48.755Z",
    lastLiveness: "2023-06-14T11:19:48.755Z",
  },
];

describe("cluster-new.vue", () => {
  let clusterNewCmp: VueWrapper<InstanceType<typeof ClusterNew>>;

  function mockWindowObject(): void {
    global.window = Object.create(window);
    global.window.location = {
      // @ts-expect-error
      ancestorOrigins: [""],
      hash: "",
      host: "dummy.com",
      port: "80",
      protocol: "http:",
      hostname: "dummy.com",
      href: "http://dummy.com?page=1&name=testing",
      origin: "http://dummy.com",
      pathname: "",
    };
  }

  function mountComponent(query = {}, params = {}) {
    mockOperatorVersions();
    clusterNewCmp = mount(ClusterNew, {
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $router: [],
          $route: {
            query: query,
            params: params,
          },
        },
      },
    });
    clusterNewCmp.vm.$q.notify = vi.fn(); // disabling notify to prevent it from failing tests
  }

  const operatorVersionsMap = {
    "2.16": "2.16.79",
    "2.17": "2.17.75",
    "2.18": "2.18.105",
    "2.19": "2.19.138",
    "2.20": "2.20.68",
    "2.21": "2.21.53",
    "2.22": "2.22.64",
    "2.23": "2.23.17",
  } as Record<string, string>;

  function mockOperatorVersions() {
    (clusterService.getOperatorVersions as MockedFunction<typeof clusterService.getOperatorVersions>).mockResolvedValue(
      operatorVersionsMap,
    );
  }

  async function setValidFormModel() {
    await clusterNewCmp.setData({
      clusterFormModel: {
        name: "some-name",
        version: "2.23",
        prerequisites: {
          location: EClusterLocations.CONTROL_PLANE,
          clusterUrl: "",
          k8sDistribution: EClusterDistIds.VANILLA,
        },
      },
    });
  }

  async function clickContiuneButton() {
    return clusterNewCmp.find("[aid=continue-btn]").trigger("click");
  }

  beforeEach(() => {
    authStore.$patch({ accessToken: accessToken });
    clusterStore.$patch({ clusters: clusters });
    appStore.$patch({ config: { isMultiTenant: false } }); // TODO: add tests to support isMultiTenant

    mockWindowObject();
  });
  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  test("component is loading properly", async () => {
    mountComponent();
    await flushPromises();
    expect(clusterNewCmp.text()).toContain("Run:ai version");
    expect(clusterNewCmp.text()).toContain("Cluster name");
  });

  test("initializes with correct data when there is an id in the $route.params.id", async () => {
    const selectedCluster: DisplayedCluster = clusters[0];

    vi.mocked(clusterStore.getClusterById).mockImplementation(() => {
      return clusters[0];
    });

    mountComponent({}, { id: selectedCluster.uuid });
    await flushPromises();
    expect(clusterNewCmp.vm.clusterFormModel.name).toEqual(selectedCluster.name);
    expect(clusterNewCmp.vm.createdCluster).toEqual(selectedCluster);
    expect(clusterNewCmp.vm.isClusterEditMode).toEqual(true);
  });

  test("createCluster should be called when pressing continue button", async () => {
    mountComponent();
    await clusterNewCmp.vm.$nextTick();

    const spyCreateCluster = vi.spyOn(clusterNewCmp.vm, "createCluster");
    clickContiuneButton();

    expect(spyCreateCluster).toHaveBeenCalledOnce();
    spyCreateCluster.mockRestore();
  });

  test("when installing an existing cluster createCluster method should not run when pressing Continue button", async () => {
    const selectedCluster: DisplayedCluster = clusters[0];
    vi.mocked(clusterStore.getClusterById).mockImplementation(() => {
      return clusters[0];
    });

    mountComponent({}, { id: selectedCluster.uuid });
    await flushPromises();

    const spyCreateCluster = vi.spyOn(clusterNewCmp.vm, "createCluster");
    await clickContiuneButton();

    expect(spyCreateCluster).to.not.toHaveBeenCalled();
    spyCreateCluster.mockRestore();
  });

  test("calls setOperatorVersionMap on created lifecycle", async () => {
    mountComponent();
    await clusterNewCmp.vm.$nextTick();
    expect(clusterService.getOperatorVersions).toHaveBeenCalled();
    await clusterNewCmp.vm.$nextTick();
    expect(clusterNewCmp.vm.operatorVersionsMap).toEqual(operatorVersionsMap);
  });

  test("calls validate method when continue button is pressed and createCluster is running", async () => {
    mountComponent();
    clusterNewCmp.vm.validate = vi.fn(() => Promise.resolve(true));

    await clickContiuneButton();

    expect(clusterNewCmp.vm.validate).toHaveBeenCalled();
  });

  test("calls validate method, shows hint when validateAndNotify fails", async () => {
    mountComponent();
    clusterNewCmp.vm.validate = vi.fn(() => Promise.resolve(false));
    const spyShowHint = vi.spyOn(clusterNewCmp.vm, "showHint");

    await clickContiuneButton();

    expect(clusterNewCmp.vm.displayFormHint).toBe(true);
    expect(clusterNewCmp.vm.validate).toHaveBeenCalled();
    expect(spyShowHint).toHaveBeenCalled();

    spyShowHint.mockClear();
  });

  test("component will render the creation step and hide the installation step on once its loaded", async () => {
    expect(clusterNewCmp.find(".cluster-installation").exists()).toBeFalsy();
    expect(clusterNewCmp.find(".cluster-creation").exists()).toBeTruthy();
  });

  test("component will render the 2nd step after submitted a valid form in the first step", async () => {
    mountComponent();
    expect(clusterNewCmp.find(".cluster-installation").exists()).toBeFalsy();
    await setValidFormModel();

    await clickContiuneButton();
    await flushPromises();
    expect(clusterNewCmp.vm.currentStep).toBe(2);
    expect(clusterNewCmp.find(".cluster-installation").exists()).toBeTruthy();
  });

  test("component will delete the cluster it created if the continue button is pressed and there is no clusterUuid in the $route.query", async () => {
    mountComponent();
    const spyDeleteLastCreatedCluster = vi.spyOn(clusterNewCmp.vm, "deleteLastCreatedCluster");
    await clusterNewCmp.setData({ createdCluster: { uuid: "some-id" } });
    await setValidFormModel();
    await clickContiuneButton();
    await flushPromises();

    expect(spyDeleteLastCreatedCluster).toHaveBeenCalled();
  });

  test("calls goBack method when back button is pressed returning to the initial step", async () => {
    mountComponent();
    await clusterNewCmp.setData({ currentStep: 2 });
    await clusterNewCmp.find("[aid=back-btn]").trigger("click");
    expect(clusterNewCmp.vm.currentStep).toBe(1);
  });

  test("calls leave method when leave button is pressed", async () => {
    mountComponent();
    await clusterNewCmp.setData({ currentStep: 2, createdCluster: { uuid: "some-uuid" } });
    await clusterNewCmp.find("[aid=leave-btn]").trigger("click");
    expect(clusterNewCmp.vm.$router).toContainEqual({
      name: CLUSTER_ROUTE_NAMES.CLUSTER_INDEX,
      query: { createdEntityId: "some-uuid" },
    });
  });

  test("sets version to latest (2.23) when cluster version (2.24.x) is not in operatorVersionsMap", async () => {
    const clusterWithNewerVersion: DisplayedCluster = {
      ...clusters[0],
      uuid: "cluster-with-newer-version",
      version: "2.24.20000040-e5ca97a",
    };

    vi.mocked(clusterStore.getClusterById).mockImplementation(() => {
      return clusterWithNewerVersion;
    });

    mountComponent({}, { id: clusterWithNewerVersion.uuid });
    await flushPromises();

    // Since "2.24" is not in operatorVersionsMap (max is 2.23), it should default to the latest version (2.23)
    expect(clusterNewCmp.vm.clusterFormModel.version).toBe("2.23");
    expect(clusterNewCmp.vm.sortedVersions[0]).toBe("2.23"); // Verify 2.23 is the latest
  });

  test("sets version to cluster version when it exists in operatorVersionsMap", async () => {
    const clusterWithExistingVersion: DisplayedCluster = {
      ...clusters[0],
      uuid: "cluster-with-existing-version",
      version: "2.20.42",
    };

    vi.mocked(clusterStore.getClusterById).mockImplementation(() => {
      return clusterWithExistingVersion;
    });

    mountComponent({}, { id: clusterWithExistingVersion.uuid });
    await flushPromises();

    // Since "2.20" exists in operatorVersionsMap, it should use that version
    expect(clusterNewCmp.vm.clusterFormModel.version).toBe("2.20");
  });
});
