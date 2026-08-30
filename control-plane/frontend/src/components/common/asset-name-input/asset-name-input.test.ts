import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import AssetNameInput from "./asset-name-input.vue";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import type { DefineComponent } from "vue";
import { Quasar, type ValidationRule } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import { Scope } from "@/swagger-models/assets-service-client";

vi.mock("@/services/asset.service");

const pinia = createTestingPinia({ createSpy: vi.fn });

setActivePinia(pinia);

describe("runai-name-input.vue", () => {
  let NameInputCmp: VueWrapper<InstanceType<typeof AssetNameInput>>;

  beforeEach(() => {
    NameInputCmp = mount(AssetNameInput, {
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $route: {
            query: "",
          },
        },
      },
      props: {
        modelValue: "",
        assetKind: "compute",
        scopeModel: {
          scope: Scope.Project,
          projectId: 1,
        },
      },
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  test("sets nameAlreadyExist to true when 'invalid' event is emitted by runai-name-validation", () => {
    const runaiNameValidation = NameInputCmp.findComponent<DefineComponent>(".runai-name-validation");
    runaiNameValidation.vm.$emit("invalid", true);
    expect(NameInputCmp.vm.nameAlreadyExist).toBe(true);
  });

  test("updates 'isNameUnique' when 'scopeModel' prop changes", async () => {
    const prevIsNameUnique: ValidationRule = NameInputCmp.vm.isNameUnique; // function pointer
    await NameInputCmp.setProps({ scopeModel: { scope: Scope.Project, projectId: 3 } });
    await NameInputCmp.vm.$nextTick();
    expect(NameInputCmp.vm.isNameUnique === prevIsNameUnique).toBe(false); // checking that the pointer has changed
  });
});
