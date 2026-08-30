import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";

import ModelNameSection from "./model-name-section.vue";

import type { IUIWorkloadModelNameSectionModel } from "./model-name-section.model";
import type { TTestEmit } from "@/models/unit-test.model";
import { EWorkloadFormType } from "@/models/workload.model";

describe("model-name-section.vue", () => {
  function getMountedCmp(name: string): VueWrapper<InstanceType<typeof ModelNameSection>> {
    return mount(ModelNameSection, {
      global: {
        plugins: [Quasar],
      },
      props: {
        entityType: EWorkloadFormType.Inference,
        disableClosing: false,
        modelName: {
          name,
        },
      },
    });
  }

  test("emits 'model-name-changed' event when the input name value changes", async () => {
    const name = "new/value";
    const modelNameSectionCmp = getMountedCmp("");
    await modelNameSectionCmp.find("input").setValue(name);
    const expectedEmits: TTestEmit<IUIWorkloadModelNameSectionModel> = [[{ name }]];
    expect(modelNameSectionCmp.emitted("model-name-changed")).toEqual(expectedEmits);
  });

  test("emits 'is-section-invalid' event when name value is valid", async () => {
    const validName = "some_name";
    const modelNameSectionCmp = getMountedCmp(validName);
    const expectedValidatorEvent: TTestEmit<boolean> = [[false]];
    expect(modelNameSectionCmp.emitted("is-section-invalid")).toEqual(expectedValidatorEvent);
  });

  test("emits 'is-section-invalid' as true when name is empty", async () => {
    const invalidName = "";
    const modelNameSectionCmp = getMountedCmp(invalidName);
    const expectedEmit: TTestEmit<boolean> = [[true]];
    expect(modelNameSectionCmp.emitted("is-section-invalid")).toEqual(expectedEmit);
  });
});
