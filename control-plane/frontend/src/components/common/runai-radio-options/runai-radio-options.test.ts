/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { DOMWrapper, VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { QTooltip, Quasar } from "quasar";
import { afterEach, beforeEach, describe, expect, test } from "vitest";
import { RunaiRadioOptions } from ".";
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";

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

describe("separate option tooltips", () => {
  let wrapper: VueWrapper<InstanceType<typeof RunaiRadioOptions>>;

  function mountOptions(props: Record<string, unknown> = {}) {
    wrapper = mount(RunaiRadioOptions, {
      attachTo: document.body,
      global: { plugins: [Quasar] },
      props: { modelValue: "option1", options, separateTooltips: true, ...props },
    });
    return wrapper;
  }

  afterEach(() => wrapper?.unmount());

  test("keeps option help outside radios with clean labels and persistent descriptions", async () => {
    mountOptions({ title: "Placement" });

    expect(wrapper.get('[role="radiogroup"]').attributes("aria-label")).toBe("Placement");
    const radios = wrapper.findAll('[role="radio"]');
    expect(radios.map((radio) => radio.attributes("aria-label"))).toEqual(["Option 1", "Option 2"]);
    expect(wrapper.find('[role="radio"] button').exists()).toBe(false);
    const descriptionId = radios[0].attributes("aria-describedby");
    expect(descriptionId).toBeTruthy();
    expect(document.getElementById(descriptionId ?? "")?.textContent).toContain("tooltip1");
    expect(wrapper.findAll("button").map((button) => button.attributes("aria-label"))).toEqual([
      "More information about Option 1",
      "More information about Option 2",
    ]);
    await wrapper.setProps({ modelValue: "option2" });
    expect(radios[0].attributes("aria-describedby")).toBe(descriptionId);
  });

  test("clicking or pressing Enter on help does not select its option", async () => {
    mountOptions();
    const help = wrapper.findAll("button")[1];

    await help.trigger("click");
    await help.trigger("keydown", { key: "Enter", keyCode: 13 });
    await help.trigger("keyup", { key: "Enter", keyCode: 13 });
    await help.trigger("keydown", { key: "ArrowDown", keyCode: 40 });

    expect(wrapper.emitted("update:model-value")).toBeUndefined();
    expect(wrapper.findAll('[role="radio"]')[0].attributes("aria-checked")).toBe("true");
  });

  test.each([
    ["ArrowDown", 40, "option2", 1],
    ["ArrowRight", 39, "option2", 1],
    ["ArrowUp", 38, "option3", 2],
    ["ArrowLeft", 37, "option3", 2],
  ])("%s selects and focuses the next enabled radio with wrapping", async (key, keyCode, value, index) => {
    mountOptions({ options: [...options, { label: "Option 3", value: "option3" }] });
    const radios = wrapper.findAll('[role="radio"]');
    (radios[0].element as HTMLElement).focus();

    await radios[0].trigger("keydown", { key, keyCode });

    expect(wrapper.emitted("update:model-value")).toEqual([[value]]);
    expect(document.activeElement).toBe(radios.at(index)?.element);
    await wrapper.setProps({ modelValue: value });
    expect(radios.map((radio) => radio.attributes("tabindex"))).toEqual(
      index === 1 ? ["-1", "0", "-1"] : ["-1", "-1", "0"],
    );
  });

  test("skips disabled options and keeps only an enabled selected radio in the tab order", async () => {
    mountOptions({
      options: [options[0], { ...options[1], disable: true }, { label: "Option 3", value: "option3" }],
    });
    const radios = wrapper.findAll('[role="radio"]');
    expect(radios.map((radio) => radio.attributes("tabindex"))).toEqual(["0", "-1", "-1"]);

    await radios[0].trigger("keydown", { key: "ArrowDown", keyCode: 40 });

    expect(wrapper.emitted("update:model-value")).toEqual([["option3"]]);
    expect(document.activeElement).toBe(radios[2].element);
    await wrapper.setProps({ modelValue: "option2" });
    expect(radios.map((radio) => radio.attributes("tabindex"))).toEqual(["0", "-1", "-1"]);
  });

  test("uses the first enabled radio as the tab stop before any selection", () => {
    mountOptions({ modelValue: null, options: [{ ...options[0], disable: true }, options[1]] });
    expect(wrapper.get('[role="radiogroup"]').attributes("aria-label")).toBe("Options");
    expect(wrapper.findAll('[role="radio"]').map((radio) => radio.attributes("tabindex"))).toEqual(["-1", "0"]);
  });

  test("wraps forward from the last radio while keeping each help button focusable", async () => {
    mountOptions({ modelValue: "option2" });
    const radios = wrapper.findAll('[role="radio"]');

    await radios[1].trigger("keydown", { key: "ArrowRight", keyCode: 39 });

    expect(wrapper.emitted("update:model-value")).toEqual([["option1"]]);
    expect(document.activeElement).toBe(radios[0].element);
    for (const help of wrapper.findAll("button")) {
      expect(help.attributes("tabindex")).not.toBe("-1");
      help.element.focus();
      expect(document.activeElement).toBe(help.element);
    }
  });

  test("does not select or focus a radio when every option is disabled", async () => {
    mountOptions({ options: options.map((option) => ({ ...option, disable: true })) });

    await wrapper.findAll('[role="radio"]')[0].trigger("keydown", { key: "ArrowDown", keyCode: 40 });

    expect(wrapper.emitted("update:model-value")).toBeUndefined();
    expect(wrapper.findAll('[role="radio"]').map((radio) => radio.attributes("tabindex"))).toEqual(["-1", "-1"]);
  });

  test("preserves each option's aid and tooltip width", async () => {
    mountOptions({
      options: [
        {
          label: "Option 1",
          value: "option1",
          aid: "custom-option",
          toolTip: "Selection help",
          toolTipWidth: "310px",
        },
      ],
    });
    expect(wrapper.get('[aid="custom-option"]').find('[role="radio"]').exists()).toBe(true);
    const tooltip = wrapper.findComponent({ name: "QTooltip" });
    expect(tooltip.props("maxWidth")).toBe("310px");

    await tooltip.vm.show();
    await wrapper.vm.$nextTick();

    const popup = document.body.querySelector('[role="tooltip"]');
    expect(popup?.textContent).toContain("Selection help");
  });

  test.each([{ disable: true }, { policyRules: { canEdit: false } }])(
    "does not change selection for disabled or readonly groups: %j",
    async (props) => {
      mountOptions(props);
      const radios = wrapper.findAll('[role="radio"]');

      await radios[0].trigger("keydown", { key: "ArrowDown", keyCode: 40 });
      await radios[1].trigger("click");
      await radios[1].trigger("keyup", { key: " ", keyCode: 32 });

      expect(wrapper.emitted("update:model-value")).toBeUndefined();
      expect(radios.map((radio) => radio.attributes("tabindex"))).toEqual(["-1", "-1"]);
      expect(radios.map((radio) => radio.attributes("aria-disabled"))).toEqual(["true", "true"]);
    },
  );

  test("arrow navigation respects policy-disabled options", async () => {
    mountOptions({
      policyRules: { options: [{ value: "option1" }, { value: "option3" }] },
      options: [...options, { label: "Option 3", value: "option3" }],
    });

    await wrapper.findAll('[role="radio"]')[0].trigger("keydown", { key: "ArrowDown", keyCode: 40 });

    expect(wrapper.emitted("update:model-value")).toEqual([["option3"]]);
  });

  test.each([
    ["Enter", 13],
    [" ", 32],
  ])("preserves selection with %s on a radio", async (key, keyCode) => {
    mountOptions();
    const radio = wrapper.findAll('[role="radio"]')[1];

    await radio.trigger("keydown", { key, keyCode });
    await radio.trigger("keyup", { key, keyCode });

    expect(wrapper.emitted("update:model-value")).toEqual([["option2"]]);
  });
});

test.each([false, true])(
  "preserves icons, overflow text and disabled help with separateTooltips=%s",
  async (separateTooltips) => {
    const option = {
      label: "Unavailable cluster",
      value: "unavailable",
      disable: true,
      icon: "eks",
      svg: "gke",
      disabledTooltip: "Cluster is offline",
      disabledTooltipWidth: "210px",
    };
    const wrapper = mount(RunaiRadioOptions, {
      attachTo: document.body,
      global: { plugins: [Quasar], stubs: { RunaiSvgIcon: true } },
      props: { modelValue: null, options: [option], minCharsToShowOverflowTooltip: 5, separateTooltips },
    });
    try {
      const radio = wrapper.get('[role="radio"]');
      expect(radio.text()).toContain("Unavailable cluster");
      expect(radio.attributes("aria-disabled")).toBe("true");
      expect(wrapper.findAllComponents(RunaiSvgIcon).map((icon) => icon.props("name"))).toEqual(["eks", "gke"]);
      const [overflow, disabled] = wrapper.findAllComponents(QTooltip);
      overflow.vm.show();
      await wrapper.vm.$nextTick();
      expect(document.body.querySelector('[role="tooltip"]')?.textContent).toContain("Unavailable cluster");
      disabled.vm.show();
      await wrapper.vm.$nextTick();
      expect(disabled.props("maxWidth")).toBe("210px");
      expect(document.body.querySelectorAll('[role="tooltip"]')[1]?.textContent).toContain("Cluster is offline");
    } finally {
      wrapper.unmount();
    }
  },
);
