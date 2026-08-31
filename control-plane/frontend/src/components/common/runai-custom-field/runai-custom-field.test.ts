/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { beforeEach, describe, expect, test } from "vitest";
import { RunaiCustomField } from ".";
import type { ICustomField } from "./runai-custom-field.vue";

describe("runai-custom-field.vue", () => {
  let runaiCustomFieldCmp: VueWrapper<InstanceType<typeof RunaiCustomField>>;

  const defaultField: ICustomField = {
    key: "test-field",
    label: "Test Field",
    value: null,
    format: "text",
    required: false,
  };

  function mountComponent(props = { field: defaultField, disable: false }) {
    runaiCustomFieldCmp = mount(RunaiCustomField, {
      global: {
        plugins: [Quasar],
      },
      props,
    });
  }

  beforeEach(() => {
    mountComponent();
  });

  test("component is loading properly", () => {
    expect(runaiCustomFieldCmp.exists()).toBe(true);
  });

  describe("Computed properties", () => {
    describe("toggleWithIcon", () => {
      test("returns true when tooltip exists and format is boolean", () => {
        mountComponent({
          field: { ...defaultField, format: "boolean", tooltip: "Test tooltip" },
          disable: false,
        });
        expect(runaiCustomFieldCmp.vm.toggleWithIcon).toBe(true);
      });

      test("returns false when tooltip exists but format is not boolean", () => {
        mountComponent({
          field: { ...defaultField, format: "text", tooltip: "Test tooltip" },
          disable: false,
        });
        expect(runaiCustomFieldCmp.vm.toggleWithIcon).toBe(false);
      });

      test("returns false when tooltip does not exist", () => {
        mountComponent({
          field: { ...defaultField, format: "boolean" },
          disable: false,
        });
        expect(runaiCustomFieldCmp.vm.toggleWithIcon).toBe(false);
      });
    });

    describe("classes", () => {
      test("returns col-11 for number format", () => {
        mountComponent({
          field: { ...defaultField, format: "number" },
          disable: false,
        });
        expect(runaiCustomFieldCmp.vm.classes).toBe("col-11");
      });

      test("returns col-11 for text format", () => {
        mountComponent({
          field: { ...defaultField, format: "text" },
          disable: false,
        });
        expect(runaiCustomFieldCmp.vm.classes).toBe("col-11");
      });

      test("returns empty string for boolean format", () => {
        mountComponent({
          field: { ...defaultField, format: "boolean" },
          disable: false,
        });
        expect(runaiCustomFieldCmp.vm.classes).toBe("");
      });
    });

    describe("componentToRender", () => {
      test("returns policy-number-input for number format", () => {
        mountComponent({
          field: { ...defaultField, format: "number" },
          disable: false,
        });
        expect(runaiCustomFieldCmp.vm.componentToRender).toBe("policy-number-input");
      });

      test("returns policy-string-field for text format", () => {
        mountComponent({
          field: { ...defaultField, format: "text" },
          disable: false,
        });
        expect(runaiCustomFieldCmp.vm.componentToRender).toBe("policy-string-field");
      });

      test("returns runai-toggle for boolean format", () => {
        mountComponent({
          field: { ...defaultField, format: "boolean" },
          disable: false,
        });
        expect(runaiCustomFieldCmp.vm.componentToRender).toBe("runai-toggle");
      });

      test("returns no component would be rendered if type is invalid", () => {
        mountComponent({
          field: { ...defaultField, format: "unknown" as ICustomField["format"] },
          disable: false,
        });
        expect(runaiCustomFieldCmp.vm.componentToRender).toBe("");
        expect(runaiCustomFieldCmp.html()).toBe("<!--v-if-->");
      });
    });

    describe("policyRules", () => {
      test("returns correct rules for number format", () => {
        const numberField: ICustomField = {
          ...defaultField,
          format: "number",
          required: true,
          min: 0,
          max: 100,
          step: 1,
        };
        mountComponent({
          field: numberField,
          disable: false,
        });
        expect(runaiCustomFieldCmp.vm.policyRules).toEqual({
          required: true,
          min: 0,
          max: 100,
          step: 1,
        });
      });

      test("returns correct rules for text format", () => {
        const textField: ICustomField = {
          ...defaultField,
          format: "text",
          required: true,
          options: ["option1", "option2"],
        };
        mountComponent({
          field: textField,
          disable: false,
        });
        expect(runaiCustomFieldCmp.vm.policyRules).toEqual({
          required: true,
          options: ["option1", "option2"],
        });
      });

      test("returns correct rules for boolean format", () => {
        const booleanField: ICustomField = {
          ...defaultField,
          format: "boolean",
          required: true,
        };
        mountComponent({
          field: booleanField,
          disable: false,
        });
        expect(runaiCustomFieldCmp.vm.policyRules).toEqual({
          required: true,
        });
      });

      test("returns null for unknown format", () => {
        mountComponent({
          field: { ...defaultField, format: "unknown" as ICustomField["format"] },
          disable: false,
        });
        expect(runaiCustomFieldCmp.vm.policyRules).toBeNull();
      });
    });
  });

  describe("Field format handling", () => {
    test("renders number input when format is number", () => {
      mountComponent({
        field: { ...defaultField, format: "number", value: 42 },
        disable: false,
      });
      expect(runaiCustomFieldCmp.findComponent({ name: "policy-number-input" }).exists()).toBe(true);
    });

    test("renders text input when format is text", () => {
      mountComponent({
        field: { ...defaultField, format: "text", value: "test" },
        disable: false,
      });
      expect(runaiCustomFieldCmp.findComponent({ name: "policy-string-field" }).exists()).toBe(true);
    });

    test("renders toggle when format is boolean", () => {
      mountComponent({
        field: { ...defaultField, format: "boolean", value: true },
        disable: false,
      });
      expect(runaiCustomFieldCmp.findComponent({ name: "runai-toggle" }).exists()).toBe(true);
    });
  });

  describe("Tooltip behavior", () => {
    test("shows tooltip when tooltip is provided ", () => {
      mountComponent({
        field: { ...defaultField, format: "boolean", tooltip: "Test tooltip" },
        disable: false,
      });
      expect(runaiCustomFieldCmp.findComponent({ name: "runai-tooltip" }).exists()).toBe(true);
    });

    test("does not show tooltip if tooltip is not provided", () => {
      mountComponent({
        field: { ...defaultField, format: "text", tooltip: "" },
        disable: false,
      });
      expect(runaiCustomFieldCmp.findComponent({ name: "runai-tooltip" }).exists()).toBe(false);
    });
  });

  describe("Value updates", () => {
    test("emits update event with correct value for text field", async () => {
      mountComponent({
        field: { ...defaultField, format: "text", value: "initial" },
        disable: false,
      });
      const newValue = "updated";
      runaiCustomFieldCmp.vm.updateValue(newValue);
      const emitted = runaiCustomFieldCmp.emitted("update");
      expect(emitted).toBeTruthy();
      expect(emitted?.[0]).toEqual([newValue]);
    });

    test("emits update event with parsed number for number field", async () => {
      mountComponent({
        field: { ...defaultField, format: "number", value: 0 },
        disable: false,
      });
      const newValue = "42";
      runaiCustomFieldCmp.vm.updateValue(newValue);
      const emitted = runaiCustomFieldCmp.emitted("update");
      expect(emitted).toBeTruthy();
      expect(emitted?.[0]).toEqual([42]);
    });

    test("emits update event with boolean value for boolean field", async () => {
      mountComponent({
        field: { ...defaultField, format: "boolean", value: false },
        disable: false,
      });
      const newValue = true;
      runaiCustomFieldCmp.vm.updateValue(newValue);
      const emitted = runaiCustomFieldCmp.emitted("update");
      expect(emitted).toBeTruthy();
      expect(emitted?.[0]).toEqual([newValue]);
    });
  });

  describe("Validation rules", () => {
    test("applies regex validation when regex is provided", () => {
      const regex = "^[A-Z]+$";
      mountComponent({
        field: { ...defaultField, regex },
        disable: false,
      });
      const result = runaiCustomFieldCmp.vm.regexRule("ABC");
      expect(result).toBe(true);
      const invalidResult = runaiCustomFieldCmp.vm.regexRule("abc");
      expect(invalidResult).toBe("Invalid input");
    });

    test("applies custom invalid message when provided", () => {
      const regex = "^[A-Z]+$";
      const customMessage = "Only uppercase letters allowed";
      mountComponent({
        field: { ...defaultField, regex, invalidMessage: customMessage },
        disable: false,
      });
      const result = runaiCustomFieldCmp.vm.regexRule("abc");
      expect(result).toBe(customMessage);
    });

    test("returns true when regex is not provided", () => {
      mountComponent({
        field: { ...defaultField },
        disable: false,
      });
      const result = runaiCustomFieldCmp.vm.regexRule("any value");
      expect(result).toBe(true);
    });

    test("returns true when field is not required and value is empty", () => {
      const regex = "^[A-Z]+$";
      mountComponent({
        field: { ...defaultField, regex, required: false },
        disable: false,
      });
      const result = runaiCustomFieldCmp.vm.regexRule("");
      expect(result).toBe(true);
    });

    test("applies regex validation when field is required and value is empty", () => {
      const regex = "^[A-Z]+$";
      mountComponent({
        field: { ...defaultField, regex, required: true },
        disable: false,
      });
      const result = runaiCustomFieldCmp.vm.regexRule("");
      expect(result).toBe("Invalid input");
    });
  });
});
