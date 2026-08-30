import type { TTestEmit } from "@/models/unit-test.model";
import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, expect, test } from "vitest";
import { RunaiInput } from ".";
describe("runai-input.vue", () => {
  let runaiInputCmp: VueWrapper<InstanceType<typeof RunaiInput>>;

  test("component is loading properly", () => {
    runaiInputCmp = mount(RunaiInput, {
      global: {
        plugins: [Quasar],
      },
      props: {
        defaultValue: null,
        modelValue: null,
      },
    });
    expect(runaiInputCmp.html()).toContain("runai-input");
  });

  describe("Emits behaviors", () => {
    test("should not emit 'update:model-value' when canEdit is false", () => {
      runaiInputCmp = mount(RunaiInput, {
        global: {
          plugins: [Quasar],
        },
        props: {
          modelValue: null,
          defaultValue: null,
          policyRules: { canEdit: false },
          policyDefault: "ubuntu",
        },
      });
      const expectedEmits = undefined;
      expect(runaiInputCmp.emitted("update:model-value")).toEqual(expectedEmits);
    });

    test("should emit 'update:model-value' when canEdit is true", () => {
      const policyDefault = "ubuntu";
      runaiInputCmp = mount(RunaiInput, {
        global: {
          plugins: [Quasar],
        },
        props: {
          modelValue: null,
          defaultValue: null,
          policyRules: { canEdit: true },
          policyDefault,
        },
      });
      const expectedEmits: TTestEmit<string> = [[policyDefault]];
      expect(runaiInputCmp.emitted("update:model-value")).toEqual(expectedEmits);
    });

    test("should emit 'update:model-value' with the defaultValue when no policy is passed", () => {
      const defaultValue = "ubuntu";
      runaiInputCmp = mount(RunaiInput, {
        global: {
          plugins: [Quasar],
        },
        props: {
          modelValue: null,
          defaultValue,
        },
      });
      expect(runaiInputCmp.emitted("update:model-value")).toEqual([[defaultValue]]);
    });
  });

  describe("Rendering behaviors", () => {
    test("Should render the input with the policyDefault value when it is passed", () => {
      runaiInputCmp = mount(RunaiInput, {
        global: {
          plugins: [Quasar],
        },
        props: {
          modelValue: null,
          defaultValue: null,
          policyRules: { canEdit: false },
          policyDefault: "ubuntu",
        },
      });
      expect(runaiInputCmp.find("input").element.value).toBe(runaiInputCmp.vm.policyDefault);
    });

    test("Should render the input with the modelValue when there is no policy", () => {
      runaiInputCmp = mount(RunaiInput, {
        global: {
          plugins: [Quasar],
        },
        props: {
          modelValue: "",
          defaultValue: "",
        },
      });
      expect(runaiInputCmp.find("input").element.value).toBe(runaiInputCmp.vm.modelValue);
    });

    test("Should render the input disabled when policyRules.canEdit is false", () => {
      runaiInputCmp = mount(RunaiInput, {
        global: {
          plugins: [Quasar],
        },
        props: {
          modelValue: null,
          defaultValue: null,
          policyRules: { canEdit: false },
          policyDefault: "ubuntu",
        },
      });
      expect(runaiInputCmp.find("input").attributes("disabled")).toBeDefined();
    });

    test("Should render the input enabled when policyRules.canEdit is true", () => {
      runaiInputCmp = mount(RunaiInput, {
        global: {
          plugins: [Quasar],
        },
        props: {
          modelValue: null,
          defaultValue: null,
          policyRules: { canEdit: true },
          policyDefault: "ubuntu",
        },
      });
      expect(runaiInputCmp.find("input").attributes("disabled")).not.toBeDefined();
    });

    test("Should render the input enabled when there is no policy", () => {
      runaiInputCmp = mount(RunaiInput, {
        global: {
          plugins: [Quasar],
        },
        props: {
          modelValue: null,
          defaultValue: "ubuntu",
        },
      });
      expect(runaiInputCmp.find("input").attributes("disabled")).not.toBeDefined();
    });

    test("Should render the input disabled when passing the disable prop", () => {
      runaiInputCmp = mount(RunaiInput, {
        global: {
          plugins: [Quasar],
        },
        props: {
          modelValue: null,
          defaultValue: null,
          disable: true,
        },
      });
      expect(runaiInputCmp.find("input").attributes("disabled")).toBeDefined();
    });
  });
});
