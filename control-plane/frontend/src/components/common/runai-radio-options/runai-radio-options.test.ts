import type { DOMWrapper, VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { beforeEach, describe, expect, test } from "vitest";
import { RunaiRadioOptions } from ".";

const options = [
  { label: "Option 1", value: "option1", toolTip: "tooltip1" },
  { label: "Option 2", value: "option2", toolTip: "tooltip2" },
];

describe("runai-radio-options.vue", () => {
  let runaiRadioOptionsCmp: VueWrapper<InstanceType<typeof RunaiRadioOptions>>;

  beforeEach(() => {
    runaiRadioOptionsCmp = mount(RunaiRadioOptions, {
      global: {
        plugins: [Quasar],
      },
      props: {
        modelValue: null,
        options,
      },
    });
  });

  test("component is loading properly", () => {
    expect(runaiRadioOptionsCmp.html()).toContain("runai-radio-options");
  });

  test("component is showing the correct options", async () => {
    expect(runaiRadioOptionsCmp.html()).toContain("Option 1");
    expect(runaiRadioOptionsCmp.html()).toContain("Option 2");
  });

  test("component is showing the title prop", async () => {
    const title = "this is the title";
    await runaiRadioOptionsCmp.setProps({ title });
    expect(runaiRadioOptionsCmp.html()).toContain(title);
  });

  test("component is showing the tooltips if the options contain tooltip", async () => {
    expect(runaiRadioOptionsCmp.findAll(".q-icon.ph-question").length).toBe(2);
  });

  test("component is showing the correct selected value", async () => {
    const modelValue = "option1";
    await runaiRadioOptionsCmp.setProps({ modelValue });

    const radioOptions: Array<DOMWrapper<HTMLInputElement>> = runaiRadioOptionsCmp.findAll(".q-radio");
    expect(radioOptions[0].html()).toContain("q-radio__inner--truthy");
    expect(radioOptions[1].html()).toContain("q-radio__inner--falsy");
  });

  test("component should emit the new selected value", async () => {
    const modelValue = "option1";
    await runaiRadioOptionsCmp.setProps({ modelValue });

    const radioOptions: Array<DOMWrapper<HTMLInputElement>> = runaiRadioOptionsCmp.findAll(".q-radio");
    await radioOptions[1].trigger("click");
    expect(runaiRadioOptionsCmp.emitted("update:model-value")).toEqual([[options[1].value]]);
  });
});
