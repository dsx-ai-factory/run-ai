/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { Quasar, QForm, QInput, Notify } from "quasar";
import RunaiChangePasswordModal from "./runai-change-password-modal.vue";
import { RunaiBtn } from "@/components/common/runai-btn";
import { authService } from "@/services/control-plane/auth.service/auth.service";

vi.mock("@/services/control-plane/auth.service/auth.service", () => ({
  authService: {
    changePassword: vi.fn().mockResolvedValue(undefined),
  },
}));

vi.mock("@/stores/auth.store", () => ({
  useAuthStore: () => ({
    logout: vi.fn(),
    loadUserInfo: vi.fn().mockResolvedValue(undefined),
  }),
}));

describe("runai-change-password-modal.vue", () => {
  const validPassword = "Aa1!aaaa";

  function factory(props = {}): VueWrapper<InstanceType<typeof RunaiChangePasswordModal>> {
    return mount(RunaiChangePasswordModal, {
      props,
      global: {
        plugins: [[Quasar, { plugins: { Notify } }]],
        components: { QForm, QInput, RunaiBtn },
        stubs: {
          RunaiBaseDialog: {
            template: '<div><slot name="header"></slot><slot name="body"></slot><slot name="footer"></slot></div>',
          },
          NotificationAlertBar: { template: '<div class="notification-alert-bar-stub"></div>' },
        },
      },
    });
  }

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders three password inputs and an info banner", () => {
    const wrapper = factory();
    expect(wrapper.findAllComponents(QInput).length).toBe(3);
    expect(wrapper.find(".notification-alert-bar-stub").exists()).toBe(true);
  });

  it("renders the change-password button enabled regardless of form state", () => {
    const wrapper = factory();
    const submitBtn = wrapper.find('[aid="reset-password-button"]');
    expect(submitBtn.exists()).toBe(true);
    expect(submitBtn.attributes("disabled")).toBeUndefined();
  });

  it("emits close on cancel click", async () => {
    const wrapper = factory();
    await wrapper.find('[aid="cancel-button"]').trigger("click");
    expect(wrapper.emitted("close")).toBeTruthy();
  });

  it("does not call changePassword when fields are empty", async () => {
    const wrapper = factory();
    await wrapper.find('[aid="reset-password-button"]').trigger("click");
    await flushPromises();
    expect(authService.changePassword).not.toHaveBeenCalled();
  });

  it("resets the processing flag after a failed validation so a corrected submit can fire", async () => {
    const wrapper = factory();
    // First click with empty fields — validation fails.
    await wrapper.find('[aid="reset-password-button"]').trigger("click");
    await flushPromises();
    expect(authService.changePassword).not.toHaveBeenCalled();

    // Now fill valid input and click again — must not be blocked by a stuck processing flag.
    const [current, next, confirm] = wrapper.findAllComponents(QInput);
    await current.setValue("current-pw");
    await next.setValue(validPassword);
    await confirm.setValue(validPassword);
    await wrapper.find('[aid="reset-password-button"]').trigger("click");
    await flushPromises();
    expect(authService.changePassword).toHaveBeenCalledTimes(1);
  });

  it("does not call changePassword when the new and confirm passwords do not match", async () => {
    const wrapper = factory();
    const [current, next, confirm] = wrapper.findAllComponents(QInput);
    await current.setValue("anything");
    await next.setValue(validPassword);
    await confirm.setValue(`${validPassword}x`);
    await wrapper.find('[aid="reset-password-button"]').trigger("click");
    await flushPromises();
    expect(authService.changePassword).not.toHaveBeenCalled();
  });

  it("calls changePassword and emits close when all fields are valid", async () => {
    const wrapper = factory();
    const [current, next, confirm] = wrapper.findAllComponents(QInput);
    await current.setValue("current-pw");
    await next.setValue(validPassword);
    await confirm.setValue(validPassword);
    await wrapper.find('[aid="reset-password-button"]').trigger("click");
    await flushPromises();
    expect(authService.changePassword).toHaveBeenCalledWith(validPassword, "current-pw");
    expect(wrapper.emitted("close")).toBeTruthy();
  });
});
