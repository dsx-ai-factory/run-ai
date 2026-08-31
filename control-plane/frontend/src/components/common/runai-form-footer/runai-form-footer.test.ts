/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, vi, test, expect, afterEach, beforeEach } from "vitest";
import { mount, type VueWrapper } from "@vue/test-utils";
import RunaiFormFooter from "@/components/common/runai-form-footer/runai-form-footer.vue";
import { Quasar } from "quasar";

let wrapper: VueWrapper<InstanceType<typeof RunaiFormFooter>>;
enum ERunaiFormFooterSelector {
  CancelBtn = '[aid="cancel-form-footer-btn"]',
  SaveBtn = '[aid="save-form-footer-btn"]',
  FormHint = ".form-hint",
  LoadSpinner = ".q-spinner",
}

interface IRunaiFormFooterProps {
  saveBtnLabel: string;
  loading?: boolean;
  isFormValid?: boolean;
  readonly?: boolean;
}

describe("RunaiFormFooter", () => {
  const mountComponent = (props: IRunaiFormFooterProps) => {
    wrapper = mount(RunaiFormFooter, {
      global: {
        plugins: [Quasar],
      },
      props,
    });
  };
  beforeEach(() => {
    vi.useFakeTimers();
  });
  afterEach(() => {
    wrapper.unmount();
    vi.clearAllMocks();
    vi.restoreAllMocks();
  });

  test("should render correctly with default props", () => {
    mountComponent({ saveBtnLabel: "Save" });
    expect(wrapper.html()).toContain("Save");
    expect(wrapper.find(ERunaiFormFooterSelector.CancelBtn).exists()).toBe(true);
    expect(wrapper.find(ERunaiFormFooterSelector.SaveBtn).exists()).toBe(true);
  });

  test("should emit 'save' event when save button is clicked", async () => {
    mountComponent({ saveBtnLabel: "Save" });
    await wrapper.find(ERunaiFormFooterSelector.SaveBtn).trigger("click");
    expect(wrapper.emitted().save).toBeTruthy();
  });

  test("should emit 'cancel' event when cancel button is clicked", async () => {
    mountComponent({ saveBtnLabel: "Save" });
    await wrapper.find(ERunaiFormFooterSelector.CancelBtn).trigger("click");
    expect(wrapper.emitted().cancel).toBeTruthy();
  });

  test("should show loading state on save button", () => {
    mountComponent({ saveBtnLabel: "Save", loading: true });
    expect(wrapper.find(ERunaiFormFooterSelector.LoadSpinner).exists()).toBe(true);
  });

  test("should display form hint when form is invalid and hide it after HINT_TIMEOUT", async () => {
    mountComponent({ saveBtnLabel: "Save", isFormValid: true });

    await wrapper.setProps({ isFormValid: false });

    expect(wrapper.vm.displayFormHint).toBe(true);
    expect(wrapper.find(`${ERunaiFormFooterSelector.FormHint} .q-field__messages`).text()).toBe(
      "Please review and fix the issues in the form",
    );
    vi.runAllTimers();
    expect(wrapper.vm.displayFormHint).toBe(false);
  });

  test("should hide form hint when form is valid", async () => {
    mountComponent({ saveBtnLabel: "Save", isFormValid: true });
    await wrapper.vm.$nextTick();
    expect(wrapper.vm.displayFormHint).toBe(false);
  });

  test("should clear timeout on unmounted", () => {
    mountComponent({ saveBtnLabel: "Save", isFormValid: false });
    const spyClearTimeout = vi.spyOn(wrapper.vm, "clearTimeout");
    wrapper.unmount();
    expect(spyClearTimeout).toHaveBeenCalled();
  });
});
