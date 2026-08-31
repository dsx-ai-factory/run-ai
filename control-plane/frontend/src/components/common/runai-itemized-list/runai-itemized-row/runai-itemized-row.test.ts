/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";

import RunaiItemizedRow from "./runai-itemized-row.vue";

import type { IItemizedListItem } from "../runai-itemized-list.model";
import type { TTestEmit } from "@/models/unit-test.model";
import { errorMessages } from "@/common/error-message.constant";

interface IRunaiItemizedRowProps {
  item: IItemizedListItem;
  disabled?: boolean;
  disabledTooltip?: string;
  shouldValidateValue?: boolean;
  itemNameLabel?: string; // default: "Name",
  itemValueLabel?: string; // default: "Value",
  nameInputValidationMessage?: string; // errorMessages.NAME_NOT_EMPTY
  valueInputValidationMessage?: string; // errorMessages.VALUE_NOT_EMPTY
  stackInputsLabel?: boolean;
  preventDelete?: boolean;
}

describe("runai-itemized-row.vue", () => {
  const item = { id: "100", name: "Test Item", value: "Test Value" };
  const nameInputSelector = '[aid="itemized-name-input"]input';
  const valueInputSelector = '[aid="itemized-value-input"]input';

  const factory = (props: IRunaiItemizedRowProps): VueWrapper<InstanceType<typeof RunaiItemizedRow>> => {
    return mount(RunaiItemizedRow, {
      global: {
        plugins: [Quasar],
      },
      props: {
        ...props,
      },
    });
  };

  it("renders correctly with default props", () => {
    const wrapper = factory({ item });
    expect(wrapper.exists()).toBe(true);
    expect(wrapper.find('[aid="itemized-name-input"]').exists()).toBe(true);
    expect(wrapper.find('[aid="itemized-value-input"]').exists()).toBe(true);
  });

  it('emits "changed" event when name input is changed', async () => {
    const wrapper = factory({ item });
    const input = wrapper.find(nameInputSelector);
    await input.setValue("Updated Name");
    expect(wrapper.emitted().changed).toBeTruthy();
    const emitResult: TTestEmit<IItemizedListItem[]> = wrapper.emitted("changed") || [];
    expect(emitResult[0][0]).toEqual({ ...item, name: "Updated Name" });
  });

  it('emits "changed" event when value input is changed', async () => {
    const wrapper = factory({ item });
    const input = wrapper.find(valueInputSelector);
    await input.setValue("Updated Value");
    expect(wrapper.emitted().changed).toBeTruthy();
    const emitResult: TTestEmit<IItemizedListItem[]> = wrapper.emitted("changed") || [];
    expect(emitResult[0][0]).toEqual({ ...item, value: "Updated Value" });
  });

  it('emits "remove" event when remove button is clicked', async () => {
    const wrapper = factory({ item, preventDelete: false, disabled: false });
    const button = wrapper.find('[aid="remove-itemized-button"]');
    await button.trigger("click");
    expect(wrapper.emitted().remove).toBeTruthy();
    const emitResult: TTestEmit<IItemizedListItem[]> = wrapper.emitted("remove") || [];
    expect(emitResult[0][0]).toEqual(item);
  });

  it("does not render remove button when preventDelete or disabled is true", () => {
    let wrapper = factory({ item, preventDelete: true });
    expect(wrapper.find('[aid="remove-itemized-button"]').exists()).toBe(false);

    wrapper = factory({ item, disabled: true });
    expect(wrapper.find('[aid="remove-itemized-button"]').exists()).toBe(false);
  });

  it("displays disabledTooltip when disabled", () => {
    const wrapper = factory({ item, disabled: true, disabledTooltip: "Tooltip text" });
    const tooltip = wrapper.findComponent({ name: "QTooltip" });
    expect(tooltip.exists()).toBe(true);
    expect(wrapper.vm.shouldHaveTooltip).toBe(true);
  });

  it("should not displays tooltip when disabledTooltip is set but disabled is false", () => {
    const wrapper = factory({ item, disabled: false, disabledTooltip: "Tooltip text" });
    const tooltips = wrapper.findAllComponents({ name: "QTooltip" });
    const disabledTooltip = tooltips.find((t) => t.text() === "Tooltip text");
    expect(disabledTooltip).toBeUndefined();
    expect(wrapper.vm.shouldHaveTooltip).toBe(false);
  });

  it("should not displays tooltip when disabled is true but disabledTooltip is empty", () => {
    const wrapper = factory({ item, disabled: true, disabledTooltip: "" });
    const tooltips = wrapper.findAllComponents({ name: "QTooltip" });
    const disabledTooltip = tooltips.find((t) => t.text() === "");
    expect(disabledTooltip).toBeUndefined();
    expect(wrapper.vm.shouldHaveTooltip).toBe(false);
  });

  it("validates name input correctly", async () => {
    const nameInputValidationMessage = "Name cannot be empty";
    const wrapper = factory({ item, nameInputValidationMessage });
    expect(wrapper.vm.isEmptyNameField("")).toBe(nameInputValidationMessage);
  });

  it("validates name input correctly with default message", async () => {
    const wrapper = factory({ item });
    expect(wrapper.vm.isEmptyNameField("")).toBe(errorMessages.NAME_NOT_EMPTY);
  });

  it("validates value input correctly", async () => {
    const valueInputValidationMessage = "Value cannot be empty";
    const wrapper = factory({ item, valueInputValidationMessage });
    expect(wrapper.vm.isEmptyValueField("")).toBe(valueInputValidationMessage);
  });

  it("validates value input correctly with default message", async () => {
    const wrapper = factory({ item });
    expect(wrapper.vm.isEmptyValueField("")).toBe(errorMessages.VALUE_NOT_EMPTY);
  });

  it("computes nameInputPlaceholder correctly", () => {
    const wrapper = factory({ item, itemNameLabel: "Custom Name" });
    expect(wrapper.vm.nameInputPlaceholder).toBe("Enter a custom name");
    const input = wrapper.find('[aid="itemized-name-input"]');
    expect(input.attributes("placeholder")).toBe("Enter a custom name");
  });

  it("computes nameInputPlaceholder correctly when itemNameLabel is not set", () => {
    const wrapper = factory({ item });
    expect(wrapper.vm.nameInputPlaceholder).toBe("Enter a name");
    const input = wrapper.find('[aid="itemized-name-input"]');
    expect(input.attributes("placeholder")).toBe("Enter a name");
  });

  it("computes valueInputPlaceholder correctly", () => {
    const wrapper = factory({ item, itemValueLabel: "Custom Value" });
    expect(wrapper.vm.valueInputPlaceholder).toBe("Enter a custom value");
    const input = wrapper.find('[aid="itemized-value-input"]');
    expect(input.attributes("placeholder")).toBe("Enter a custom value");
  });

  it("computes valueInputPlaceholder correctly when itemValueLabel is not set", () => {
    const wrapper = factory({ item });
    expect(wrapper.vm.valueInputPlaceholder).toBe("Enter a value");
    const input = wrapper.find('[aid="itemized-value-input"]');
    expect(input.attributes("placeholder")).toBe("Enter a value");
  });
});
