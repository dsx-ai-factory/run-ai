/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import GenericSecretInputPair from "./generic-secret-input-pair.vue";
import { beforeEach, describe, expect, test } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import type { GenericSecretInputPairProps } from "./generic-secret-input-pair.vue";
import { validationRules } from "@/common/validation-rules";

describe("generic-secret-input-pair.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof GenericSecretInputPair>>;

  const defaultProps: { secretPairDetails: GenericSecretInputPairProps } = {
    secretPairDetails: {
      identifierInput: {
        value: "test-identifier",
        label: "Identifier",
        rules: [validationRules.validateKeyNotEmpty],
      },
      secretInput: {
        value: "test-secret",
        label: "Secret",
        rules: [validationRules.validateSecretNotEmpty],
      },
    },
  };

  beforeEach(() => {
    wrapper = mount(GenericSecretInputPair, {
      global: {
        plugins: [Quasar],
      },
      props: defaultProps,
    });
  });

  describe("Component Rendering", () => {
    test("renders both identifier and secret inputs", () => {
      const identifierInput = wrapper.find("[aid=generic-pair-identifier-input]");
      const secretInput = wrapper.find("[aid=generic-pair-secret-input]");

      expect(identifierInput.exists()).toBe(true);
      expect(secretInput.exists()).toBe(true);
    });

    test("renders secret input as password type by default", () => {
      const secretInput = wrapper.find("[aid=generic-pair-secret-input]");
      expect(secretInput.attributes("type")).toBe("password");
    });

    test("has eye icon for password visibility toggle", () => {
      const eyeIcon = wrapper.find(".q-icon");
      expect(eyeIcon.exists()).toBe(true);
    });
  });

  describe("Password Visibility Toggle", () => {
    test("toggles password visibility when eye icon is clicked", async () => {
      const eyeIcon = wrapper.find(".q-icon");
      const secretInput = wrapper.find("[aid=generic-pair-secret-input]");

      expect(secretInput.attributes("type")).toBe("password");
      expect(wrapper.vm.isPassHidden).toBe(true);

      await eyeIcon.trigger("click");
      await wrapper.vm.$nextTick();

      expect(wrapper.vm.isPassHidden).toBe(false);
      expect(secretInput.attributes("type")).toBe("text");
    });

    test("changes eye icon when password visibility is toggled", async () => {
      const eyeIcon = wrapper.find(".q-icon");

      await eyeIcon.trigger("click");
      await wrapper.vm.$nextTick();

      // The icon name should change from eye-slash to eye
      expect(wrapper.vm.isPassHidden).toBe(false);
    });
  });

  describe("Identifier Input Updates", () => {
    test("emits update event when identifier value changes", async () => {
      const identifierInput = wrapper.find("[aid=generic-pair-identifier-input]");
      await identifierInput.setValue("new-identifier");

      const emitted = wrapper.emitted("update");
      expect(emitted).toBeTruthy();
      expect(emitted?.[0]).toEqual([
        {
          identifier: "new-identifier",
          secret: "test-secret",
        },
      ]);
    });

    test("emits empty string when identifier is cleared", async () => {
      const identifierInput = wrapper.find("[aid=generic-pair-identifier-input]");
      await identifierInput.setValue("");

      const emitted = wrapper.emitted("update");
      expect(emitted).toBeTruthy();
      expect(emitted?.[emitted.length - 1]).toEqual([
        {
          identifier: "",
          secret: "test-secret",
        },
      ]);
    });

    test("handles null identifier value", async () => {
      wrapper.vm.updateIdentifierInputValue(null);
      await wrapper.vm.$nextTick();

      const emitted = wrapper.emitted("update");
      expect(emitted?.[0]).toEqual([
        {
          identifier: "",
          secret: "test-secret",
        },
      ]);
    });
  });

  describe("Secret Input Updates", () => {
    test("emits update event when secret value changes", async () => {
      const secretInput = wrapper.find("[aid=generic-pair-secret-input]");
      await secretInput.setValue("new-secret");

      const emitted = wrapper.emitted("update");
      expect(emitted).toBeTruthy();
      expect(emitted?.[0]).toEqual([
        {
          identifier: "test-identifier",
          secret: "new-secret",
        },
      ]);
    });

    test("emits empty string when secret is cleared", async () => {
      const secretInput = wrapper.find("[aid=generic-pair-secret-input]");
      await secretInput.setValue("");

      const emitted = wrapper.emitted("update");
      expect(emitted).toBeTruthy();
      expect(emitted?.[emitted.length - 1]).toEqual([
        {
          identifier: "test-identifier",
          secret: "",
        },
      ]);
    });

    test("handles null secret value", async () => {
      wrapper.vm.updateSecretInputValue(null);
      await wrapper.vm.$nextTick();

      const emitted = wrapper.emitted("update");
      expect(emitted?.[0]).toEqual([
        {
          identifier: "test-identifier",
          secret: "",
        },
      ]);
    });
  });

  describe("Multiple Updates", () => {
    test("handles multiple consecutive updates to identifier", async () => {
      wrapper.vm.updateIdentifierInputValue("first");
      await wrapper.vm.$nextTick();

      wrapper.vm.updateIdentifierInputValue("second");
      await wrapper.vm.$nextTick();

      wrapper.vm.updateIdentifierInputValue("third");
      await wrapper.vm.$nextTick();

      const emitted = wrapper.emitted("update");
      expect(emitted).toHaveLength(3);
      expect(emitted?.[2]).toEqual([
        {
          identifier: "third",
          secret: "test-secret",
        },
      ]);
    });

    test("handles multiple consecutive updates to secret", async () => {
      wrapper.vm.updateSecretInputValue("first-secret");
      await wrapper.vm.$nextTick();

      wrapper.vm.updateSecretInputValue("second-secret");
      await wrapper.vm.$nextTick();

      const emitted = wrapper.emitted("update");
      expect(emitted).toHaveLength(2);
      expect(emitted?.[1]).toEqual([
        {
          identifier: "test-identifier",
          secret: "second-secret",
        },
      ]);
    });

    test("maintains correct state when both inputs are updated", async () => {
      wrapper.vm.updateIdentifierInputValue("new-id");
      await wrapper.vm.$nextTick();

      wrapper.vm.updateSecretInputValue("new-secret");
      await wrapper.vm.$nextTick();

      const emitted = wrapper.emitted("update");
      expect(emitted).toHaveLength(2);

      // First emission - identifier changed
      expect(emitted?.[0]).toEqual([
        {
          identifier: "new-id",
          secret: "test-secret",
        },
      ]);

      // Second emission - secret changed (with original identifier from props)
      expect(emitted?.[1]).toEqual([
        {
          identifier: "test-identifier",
          secret: "new-secret",
        },
      ]);
    });
  });
});
