/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import RunaiVerifyButton from "./runai-verify-button.vue";
import { ERunaiVerifyButtonStatus } from "./runai-verify-button.model";

interface IRunaiVerifyButtonProps {
  verifyStatus: ERunaiVerifyButtonStatus;
  validationMessage?: string;
  disabled?: boolean;
}

describe("runai-verify-button.vue", () => {
  function factory(props: IRunaiVerifyButtonProps): VueWrapper<InstanceType<typeof RunaiVerifyButton>> {
    return mount(RunaiVerifyButton, {
      global: {
        plugins: [Quasar],
      },
      props: props,
    });
  }

  describe("Html elements check", () => {
    it("renders with default state", () => {
      const wrapper = factory({
        verifyStatus: ERunaiVerifyButtonStatus.None,
        validationMessage: "",
      });

      expect(wrapper.find(".verify-button").text()).toContain("Verify");
      expect(wrapper.find(".validation").text()).toBe("");
      expect(wrapper.find(".verify-button").classes()).toContain("pending");
    });

    it('displays "Verified" status correctly', () => {
      const wrapper = factory({
        verifyStatus: ERunaiVerifyButtonStatus.Verified,
        validationMessage: "Verified successfully",
      });

      expect(wrapper.find(".verify-button").text()).toContain("Verified");
      expect(wrapper.find(".validation").text()).toBe("Verified successfully");
      expect(wrapper.find(".verify-button").classes()).toContain("verified");
      expect(wrapper.find(".validation").classes()).toContain("verified-message");
    });

    it('displays "Verifying" status correctly', () => {
      const wrapper = factory({
        verifyStatus: ERunaiVerifyButtonStatus.Verifying,
        validationMessage: "Verification in progress",
      });

      expect(wrapper.find(".verify-button").text()).toContain("VERIFYING");
      expect(wrapper.find(".validation").text()).toBe("Verification in progress");
      expect(wrapper.find(".verify-button").classes()).toContain("pending");
    });

    it('displays "Failed" status correctly', () => {
      const wrapper = factory({
        verifyStatus: ERunaiVerifyButtonStatus.Failed,
        validationMessage: "Verification failed",
      });

      expect(wrapper.find(".verify-button").text()).toContain("Failed");
      expect(wrapper.find(".validation").text()).toBe("Verification failed");
      expect(wrapper.find(".verify-button").classes()).toContain("failed");
      expect(wrapper.find(".validation").classes()).toContain("non-verified-message");
    });
  });

  describe("Props", () => {
    it("disables the button when disabled prop is true", () => {
      const wrapper = factory({
        verifyStatus: ERunaiVerifyButtonStatus.None,
        disabled: true,
      });

      expect(wrapper.find(".verify-button").classes("disabled")).toBe(true);
    });
  });

  describe("Events", () => {
    it('emits "verify" event on button click', async () => {
      const wrapper = factory({
        verifyStatus: ERunaiVerifyButtonStatus.None,
      });

      await wrapper.find(".verify-button").trigger("click");
      expect(wrapper.emitted()).toHaveProperty("verify");
    });
  });
});
