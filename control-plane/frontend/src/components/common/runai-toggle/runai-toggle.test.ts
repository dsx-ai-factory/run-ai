/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe } from "vitest";
import { RunaiToggle } from ".";
describe("runai-toggle.vue", () => {
  let runaiToggleCmp: VueWrapper<InstanceType<typeof RunaiToggle>>;

  describe("Emits behaviors", () => {
    test("should not emit 'update:model-value' when canEdit is false", () => {
      runaiToggleCmp = mount(RunaiToggle, {
        global: {
          plugins: [Quasar],
        },
        props: {
          modelValue: null,
          defaultValue: false,
          policyRules: { canEdit: false },
          policyDefault: true,
        },
      });
      const expectedEmits = undefined;
      expect(runaiToggleCmp.emitted("update:model-value")).toEqual(expectedEmits);
    });
  });

  describe("Rendering behaviors", () => {
    test("Should render the toggle on when there is no policy and the modelValue is true", () => {
      runaiToggleCmp = mount(RunaiToggle, {
        global: {
          plugins: [Quasar],
        },
        props: {
          modelValue: true,
          defaultValue: true,
        },
      });
      expect(runaiToggleCmp.find(".q-toggle").attributes("aria-checked")).toBe("true");
    });

    test("Should render the toggle off when there is no policy and the modelValue is false", () => {
      runaiToggleCmp = mount(RunaiToggle, {
        global: {
          plugins: [Quasar],
        },
        props: {
          modelValue: false,
          defaultValue: false,
        },
      });
      expect(runaiToggleCmp.find(".q-toggle").attributes("aria-checked")).toBe("false");
    });

    test("Should render the toggle disabled when policyRules.canEdit is false", () => {
      runaiToggleCmp = mount(RunaiToggle, {
        global: {
          plugins: [Quasar],
        },
        props: {
          modelValue: null,
          defaultValue: true,
          policyRules: { canEdit: false },
          policyDefault: false,
        },
      });
      expect(runaiToggleCmp.find(".q-toggle").attributes("aria-disabled")).toBe("true");
    });

    test("Should render the toggle enabled when policyRules.canEdit is true", () => {
      runaiToggleCmp = mount(RunaiToggle, {
        global: {
          plugins: [Quasar],
        },
        props: {
          modelValue: null,
          defaultValue: true,
          policyRules: { canEdit: true },
          policyDefault: false,
        },
      });
      expect(runaiToggleCmp.find(".q-toggle").attributes("aria-disabled")).toBeFalsy();
    });

    test("Should render the toggle enabled when there is no policy", () => {
      runaiToggleCmp = mount(RunaiToggle, {
        global: {
          plugins: [Quasar],
        },
        props: {
          modelValue: null,
          defaultValue: true,
        },
      });
      expect(runaiToggleCmp.find(".q-toggle").attributes("aria-disabled")).toBeFalsy();
    });

    test("Should render the toggle disabled when passing the disable prop", () => {
      runaiToggleCmp = mount(RunaiToggle, {
        global: {
          plugins: [Quasar],
        },
        props: {
          modelValue: null,
          defaultValue: true,
          disable: true,
        },
      });
      expect(runaiToggleCmp.find(".q-toggle").attributes("aria-disabled")).toBe("true");
    });
  });
});
