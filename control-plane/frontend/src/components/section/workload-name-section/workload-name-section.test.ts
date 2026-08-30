import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import WorkspaceNameSection from "./workload-name-section.vue";
import type { MockedFunction } from "vitest";
import type { DefineComponent } from "vue";
import { Quasar } from "quasar";
import { setActivePinia } from "pinia";
import { createTestingPinia } from "@pinia/testing";
import { capitalizeString } from "@/utils/string.util/string.util";
import { workloadService } from "@/services/cluster/workload.service/workload.service";

vi.mock("@/services/cluster/workload.service/workload.service");

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

describe("workspace-name-section.vue", () => {
  let wsNameSectionCmp: VueWrapper<InstanceType<typeof WorkspaceNameSection>>;

  const mockCountWorkloadsByName = workloadService.countWorkloadsByName as MockedFunction<
    typeof workloadService.countWorkloadsByName
  >;
  mockCountWorkloadsByName.mockResolvedValue(0);

  const entityType = "workspace";
  beforeEach(() => {
    wsNameSectionCmp = mount(WorkspaceNameSection, {
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $route: {
            query: "",
          },
        },
      },
      props: {
        entityType: entityType,
        name: "",
        disableClosing: false,
        projectId: 1,
        clusterId: "3432kj3-jk34j2-k3423",
      },
    });
  });
  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  test("component is loading properly", () => {
    expect(wsNameSectionCmp.html()).toContain(`${capitalizeString(entityType)} name`);
    expect(wsNameSectionCmp.html()).toContain(`${entityType}-name-input`);
  });

  test("emits 'update:name' event when the input value changes", async () => {
    const wsName = "new-value";
    await wsNameSectionCmp.find("input").setValue(wsName);
    const expectedEmits: string[][] = [[wsName]];
    expect(wsNameSectionCmp.emitted("update:name")).toEqual(expectedEmits);
  });

  test("emits 'is-section-invalid' events when new name is an empty string", async () => {
    const expectedEmit: boolean[][] = [[true]]; // Note: in these tests the name is set to an empty string by default (which is invalid)
    expect(wsNameSectionCmp.emitted("is-section-invalid")).toEqual(expectedEmit);
  });

  test("emits 'is-section-invalid' events when new name is contains invalid characters", async () => {
    wsNameSectionCmp = mount(WorkspaceNameSection, {
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $route: {
            query: "",
          },
        },
      },
      props: {
        entityType: "workspace",
        name: "valid-name",
        disableClosing: false,
        projectId: 1,
        clusterId: "231-23de34-34d32",
      },
    });
    const invalidWsName = ".Invalid/name!";
    await wsNameSectionCmp.setProps({ name: invalidWsName });
    const expectedPostChangeEmits: boolean[][] = [[false], [true]];
    expect(wsNameSectionCmp.emitted("is-section-invalid")).toEqual(expectedPostChangeEmits);
  });

  test("sets nameAlreadyExist to true when 'invalid' event is emitted by runai-name-validation", () => {
    expect(wsNameSectionCmp.vm.nameAlreadyExist).toBe(false);

    const runaiNameValidation = wsNameSectionCmp.findComponent<DefineComponent>(".runai-name-validation");
    runaiNameValidation.vm.$emit("invalid", true);
    expect(wsNameSectionCmp.vm.nameAlreadyExist).toBe(true);
  });
});
