/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterEach, describe, expect, it, vi } from "vitest";
import type { ISelectOption } from "@/models/global.model";
import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { RunaiAddableMultiSelect } from "@/components/common/runai-addable-multi-select/index";
import { Quasar, type QSelect } from "quasar";
import { errorMessages } from "@/common/error-message.constant";
import { unitTestUtil } from "@/utils/unit-test.util";

const DEFAULT_PROPS = {
  options: [],
  selectedOptions: [],
  label: "",
  loading: false,
  newOptionLabel: "option",
  required: false,
  errorMessage: errorMessages.SELECT_A_VALUE,
  disable: false,
};

let wrapper: VueWrapper<InstanceType<typeof RunaiAddableMultiSelect>>;

const getMultiSelectValues = (): string[] => {
  return unitTestUtil.getMultiSelectValues(wrapper as VueWrapper);
};

const clearSelectedOptions = async () => {
  await wrapper.find(".ph-x").trigger("click");
};
describe("RunaiAddableMultiSelect", () => {
  const factory = (
    props: {
      options: ISelectOption[];
      selectedOptions: ISelectOption[];
      label: string;
      loading: boolean;
      newOptionLabel: string;
      required: boolean;
      errorMessage: string;
      disable: boolean;
    } = DEFAULT_PROPS,
  ) => {
    wrapper = mount(RunaiAddableMultiSelect, {
      global: {
        plugins: [Quasar],
      },
      props,
    });
  };
  it("renders options correctly", async () => {
    const options: ISelectOption[] = [
      { label: "option1", value: "option1" },
      { label: "option2", value: "option2" },
    ];

    const selectedOptions: ISelectOption[] = [{ label: "option1", value: "option" }];
    factory({ ...DEFAULT_PROPS, options, selectedOptions });

    const values = getMultiSelectValues();
    expect(values).toEqual(["option1"]);
  });
  it('emits "options-select" event when an option is selected', async () => {
    const options: ISelectOption[] = [
      { label: "option1", value: "option1" },
      { label: "option2", value: "option2" },
    ];
    factory({ ...DEFAULT_PROPS, options });
    wrapper.vm.createNewValue("option3");
    expect(wrapper.emitted("add-option")).toBeTruthy();
  });
  it("shows field feedback for a typed value at the limit and allows adding after removal", async () => {
    const selectedOptions = [
      { label: "first", value: "first" },
      { label: "second", value: "second" },
    ];
    factory({ ...DEFAULT_PROPS, selectedOptions });
    await wrapper.setProps({ maxValues: 2 });
    const input = wrapper.get("input");

    await input.setValue("third");
    await input.trigger("keydown", { key: "Enter", keyCode: 13 });
    await flushPromises();
    expect(wrapper.emitted("add-option")).toBeUndefined();
    expect(wrapper.find(".q-field--error").exists()).toBe(true);
    expect(wrapper.get(".q-field__messages").text()).toContain("Select up to 2 values");
    expect(getMultiSelectValues()).toEqual(["first", "second"]);

    await wrapper.findAllComponents({ name: "QChip" })[1].find(".q-chip__icon--remove").trigger("click");
    expect(wrapper.emitted("options-select")?.[0][0]).toEqual([selectedOptions[0]]);
    await wrapper.setProps({ selectedOptions: [selectedOptions[0]] });
    expect(wrapper.find(".q-field--error").exists()).toBe(false);
    await input.setValue("replacement");
    await input.trigger("keydown", { key: "Enter", keyCode: 13 });
    expect(wrapper.emitted("add-option")?.[0][0]).toBe("replacement");
    expect(wrapper.find(".q-field--error").exists()).toBe(false);
    wrapper.unmount();
  });

  it("clears limit feedback when the typed value is edited", async () => {
    factory({ ...DEFAULT_PROPS, selectedOptions: [{ label: "first", value: "first" }] });
    await wrapper.setProps({ maxValues: 1 });
    const input = wrapper.get("input");
    await input.setValue("second");
    await input.trigger("keydown", { key: "Enter", keyCode: 13 });
    expect(wrapper.find(".q-field--error").exists()).toBe(true);

    await input.setValue("edited");
    await flushPromises();
    expect(wrapper.find(".q-field--error").exists()).toBe(false);
    wrapper.unmount();
  });

  it("retains required-field validation after clearing selection-limit feedback", async () => {
    factory({ ...DEFAULT_PROPS, required: true, selectedOptions: [{ label: "first", value: "first" }] });
    await wrapper.setProps({ maxValues: 1 });
    const input = wrapper.get("input");
    await input.setValue("second");
    await input.trigger("keydown", { key: "Enter", keyCode: 13 });
    expect(wrapper.get(".q-field__messages").text()).toContain("Select up to 1 value");

    await wrapper.setProps({ selectedOptions: [] });
    await (wrapper.findComponent({ name: "QSelect" }).vm as unknown as QSelect).validate();
    await flushPromises();
    expect(wrapper.find(".q-field--error").exists()).toBe(true);
    expect(wrapper.get(".q-field__messages").text()).toBe(errorMessages.SELECT_A_VALUE);
    wrapper.unmount();
  });

  it("clears selected options when clear button is clicked", async () => {
    const options: ISelectOption[] = [
      { label: "option1", value: "option1" },
      { label: "option2", value: "option2" },
    ];
    const selectedOptions: ISelectOption[] = [{ label: "option1", value: "option" }];
    factory({ ...DEFAULT_PROPS, options, selectedOptions });
    await clearSelectedOptions();
    await wrapper.vm.$nextTick();
    expect(wrapper.emitted("options-select")?.[0][0]).toEqual([]);
  });

  it("should emit create new value with callback func", async () => {
    const options: ISelectOption[] = [
      { label: "option1", value: "option1" },
      { label: "option2", value: "option2" },
    ];
    factory({ ...DEFAULT_PROPS, options });
    wrapper.vm.createNewValue("option3");
    expect(wrapper.emitted("add-option")?.[0][0]).toEqual("option3");
    expect(wrapper.emitted("add-option")?.[0][1]).toBeInstanceOf(Function);
  });

  it("clears the active option and closes the popup after adding a new value", () => {
    const options: ISelectOption[] = [{ label: "option1", value: "option1" }];
    factory({ ...DEFAULT_PROPS, options });
    const select = wrapper.findComponent({ name: "QSelect" });
    const selectVm = select.vm as unknown as {
      updateInputValue: (value: string) => void;
      setOptionIndex: (index: number) => void;
      hidePopup: () => void;
    };
    const updateInputValue = vi.spyOn(selectVm, "updateInputValue");
    const setOptionIndex = vi.spyOn(selectVm, "setOptionIndex");
    const hidePopup = vi.spyOn(selectVm, "hidePopup");

    wrapper.vm.createNewValue("new-option");
    const doneAdding = wrapper.emitted("add-option")?.[0][1] as () => void;
    doneAdding();

    expect(updateInputValue).toHaveBeenCalledWith("");
    expect(setOptionIndex).toHaveBeenCalledWith(-1);
    expect(hidePopup).toHaveBeenCalledOnce();
  });

  describe("Policy Rules - canEdit", () => {
    it("should disable the select when policyRules.canEdit is false", () => {
      const options: ISelectOption[] = [
        { label: "option1", value: "option1" },
        { label: "option2", value: "option2" },
      ];
      const policyRules = { canEdit: false };

      wrapper = mount(RunaiAddableMultiSelect, {
        global: {
          plugins: [Quasar],
        },
        props: {
          ...DEFAULT_PROPS,
          options,
          policyRules,
        },
      });

      const qSelect = wrapper.findComponent({ name: "QSelect" });
      expect(qSelect.props("disable")).toBe(true);
    });

    it("should show the disabledState as true when policyRules.canEdit is false", () => {
      const policyRules = { canEdit: false };

      wrapper = mount(RunaiAddableMultiSelect, {
        global: {
          plugins: [Quasar],
        },
        props: {
          ...DEFAULT_PROPS,
          policyRules,
        },
      });

      expect(wrapper.vm.disabledState).toBe(true);
    });

    it("should make chips non-removable when policyRules.canEdit is false", () => {
      const options: ISelectOption[] = [
        { label: "option1", value: "option1" },
        { label: "option2", value: "option2" },
      ];
      const selectedOptions: ISelectOption[] = [{ label: "option1", value: "option1" }];
      const policyRules = { canEdit: false };

      wrapper = mount(RunaiAddableMultiSelect, {
        global: {
          plugins: [Quasar],
        },
        props: {
          ...DEFAULT_PROPS,
          options,
          selectedOptions,
          policyRules,
        },
      });

      const chips = wrapper.findAllComponents({ name: "QChip" });
      expect(chips.length).toBeGreaterThan(0);
      chips.forEach((chip) => {
        expect(chip.props("removable")).toBe(false);
      });
    });

    it("should make chips non-removable when disable prop is true", () => {
      const options: ISelectOption[] = [
        { label: "option1", value: "option1" },
        { label: "option2", value: "option2" },
      ];
      const selectedOptions: ISelectOption[] = [{ label: "option1", value: "option1" }];

      wrapper = mount(RunaiAddableMultiSelect, {
        global: {
          plugins: [Quasar],
        },
        props: {
          ...DEFAULT_PROPS,
          options,
          selectedOptions,
          disable: true,
        },
      });

      const chips = wrapper.findAllComponents({ name: "QChip" });
      expect(chips.length).toBeGreaterThan(0);
      chips.forEach((chip) => {
        expect(chip.props("removable")).toBe(false);
      });
    });

    it("should show tooltip with correct message when policyRules.canEdit is false", () => {
      const policyRules = { canEdit: false };

      wrapper = mount(RunaiAddableMultiSelect, {
        global: {
          plugins: [Quasar],
        },
        props: {
          ...DEFAULT_PROPS,
          policyRules,
        },
      });

      const tooltip = wrapper.findComponent({ name: "QTooltip" });
      expect(tooltip.exists()).toBe(true);
      expect(wrapper.vm.disabledTooltip).toBe(errorMessages.CANT_BE_MODIFIED_POLICY);
    });

    it("should show custom disabledInfo tooltip when provided", () => {
      const customMessage = "Custom disabled message";

      wrapper = mount(RunaiAddableMultiSelect, {
        global: {
          plugins: [Quasar],
        },
        props: {
          ...DEFAULT_PROPS,
          disabledInfo: customMessage,
        },
      });

      expect(wrapper.vm.disabledState).toBe(true);
      expect(wrapper.vm.disabledTooltip).toBe(customMessage);
    });

    it("should be disabled when both disable prop and policyRules.canEdit are false", () => {
      const policyRules = { canEdit: false };

      wrapper = mount(RunaiAddableMultiSelect, {
        global: {
          plugins: [Quasar],
        },
        props: {
          ...DEFAULT_PROPS,
          disable: true,
          policyRules,
        },
      });

      const qSelect = wrapper.findComponent({ name: "QSelect" });
      expect(qSelect.props("disable")).toBe(true);
      expect(wrapper.vm.disabledState).toBe(true);
    });

    it("should allow editing when policyRules.canEdit is true", () => {
      const options: ISelectOption[] = [
        { label: "option1", value: "option1" },
        { label: "option2", value: "option2" },
      ];
      const selectedOptions: ISelectOption[] = [{ label: "option1", value: "option1" }];
      const policyRules = { canEdit: true };

      wrapper = mount(RunaiAddableMultiSelect, {
        global: {
          plugins: [Quasar],
        },
        props: {
          ...DEFAULT_PROPS,
          options,
          selectedOptions,
          policyRules,
        },
      });

      const qSelect = wrapper.findComponent({ name: "QSelect" });
      expect(qSelect.props("disable")).toBe(false);
      expect(wrapper.vm.disabledState).toBe(false);

      const chips = wrapper.findAllComponents({ name: "QChip" });
      chips.forEach((chip) => {
        expect(chip.props("removable")).toBe(true);
      });
    });

    it("should allow editing when policyRules is null", () => {
      const options: ISelectOption[] = [
        { label: "option1", value: "option1" },
        { label: "option2", value: "option2" },
      ];
      const selectedOptions: ISelectOption[] = [{ label: "option1", value: "option1" }];

      wrapper = mount(RunaiAddableMultiSelect, {
        global: {
          plugins: [Quasar],
        },
        props: {
          ...DEFAULT_PROPS,
          options,
          selectedOptions,
          policyRules: null,
        },
      });

      const qSelect = wrapper.findComponent({ name: "QSelect" });
      expect(qSelect.props("disable")).toBe(false);
      expect(wrapper.vm.disabledState).toBe(false);
    });
  });

  describe("serverSideFilter", () => {
    const factoryServerSide = (options: ISelectOption[] = [], extraProps: Record<string, unknown> = {}) => {
      wrapper = mount(RunaiAddableMultiSelect, {
        global: { plugins: [Quasar] },
        props: { ...DEFAULT_PROPS, options, serverSideFilter: true, ...extraProps },
      });
    };

    afterEach(() => {
      wrapper.unmount();
    });

    const callFilter = (val: string) => {
      wrapper.vm.filterFn(val, (fn: () => void) => fn());
    };

    it("assigns all options without client-side text filtering", () => {
      const options = [
        { label: "alice@run.ai", value: "alice@run.ai" },
        { label: "bob@run.ai", value: "bob@run.ai" },
      ];
      factoryServerSide(options);
      // "xyz" matches nothing in client-side filtering, but serverSideFilter skips it
      callFilter("xyz");
      expect(wrapper.vm.filterOptions).toEqual(options);
    });

    it("shows 'add new' option when val is non-empty and no options exist", () => {
      factoryServerSide([]);
      callFilter("newuser@run.ai");
      expect(wrapper.vm.filterOptions).toHaveLength(1);
      expect(wrapper.vm.filterOptions[0].value).toBe("newuser@run.ai");
      expect(wrapper.vm.filterOptions[0].label).toContain("(new");
    });

    it("does not show 'add new' when options exist (server returned results)", () => {
      const options = [{ label: "alice@run.ai", value: "alice@run.ai" }];
      factoryServerSide(options);
      callFilter("alice");
      expect(wrapper.vm.filterOptions).toHaveLength(1);
      expect(wrapper.vm.filterOptions[0].label).toBe("alice@run.ai");
    });

    it("does not show 'add new' when val is empty", () => {
      factoryServerSide([]);
      callFilter("");
      expect(wrapper.vm.filterOptions).toHaveLength(0);
    });

    it("shows plain value without '(new)' label when canSearch=false and no options", () => {
      factoryServerSide([], { canSearch: false });
      callFilter("newuser@run.ai");
      expect(wrapper.vm.filterOptions[0].label).toBe("newuser@run.ai");
      expect(wrapper.vm.filterOptions[0].label).not.toContain("(new");
    });

    it("tracks currentFilterValue so the options watcher can re-run the same filter", () => {
      factoryServerSide([]);
      callFilter("test-query");
      expect(wrapper.vm.currentFilterValue).toBe("test-query");
    });

    it("clears the auto-focused option after filtering", () => {
      factoryServerSide([{ label: "test", value: "test" }]);
      const multiSelectRef = wrapper.vm.$refs.multiSelect as QSelect;
      const setOptionIndex = vi.spyOn(multiSelectRef, "setOptionIndex");

      wrapper.vm.filterFn("te", (applyFilter, afterFilter) => {
        applyFilter();
        afterFilter?.();
      });

      expect(setOptionIndex).toHaveBeenCalledWith(-1);
    });

    it("re-runs filter on the q-select ref when options prop changes", async () => {
      const options = [{ label: "alice@run.ai", value: "alice@run.ai" }];
      factoryServerSide(options);
      callFilter("alice");

      const multiSelectRef = wrapper.vm.$refs.multiSelect as QSelect;
      const filterSpy = vi.spyOn(multiSelectRef, "filter");

      await wrapper.setProps({ options: [...options, { label: "carol@run.ai", value: "carol@run.ai" }] });
      await wrapper.vm.$nextTick();

      expect(filterSpy).toHaveBeenCalledWith("alice");
    });
  });
});
