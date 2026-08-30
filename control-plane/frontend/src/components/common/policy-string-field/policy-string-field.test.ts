import { mount } from "@vue/test-utils";
import { PolicyStringField } from "@/components/common/policy-string-field";
import { describe, it, expect } from "vitest";
import { QInput, QSelect, Quasar } from "quasar";
import RunaiSelect from "../runai-select/runai-select.vue";
import RunaiTooltipWrapper from "../runai-tooltip-wrapper/runai-tooltip-wrapper.vue";

describe("policy-string-field.vue", () => {
  const baseProps = {
    modelValue: "",
    policyRules: null,
    label: "Test Label",
    disable: false,
    placeholder: "Enter value",
  };

  function mountComponent(props = {}) {
    return mount(PolicyStringField, {
      global: {
        plugins: [Quasar],
      },
      components: {
        QInput,
        QSelect,
        RunaiSelect,
        RunaiTooltipWrapper,
      },
      props: {
        ...baseProps,
        ...props,
      },
    });
  }

  describe("Html elements check", () => {
    it("renders q-input when no options are available", () => {
      const wrapper = mountComponent({ policyRules: { options: null } });

      expect(wrapper.find(".policy-string-field-input").exists()).toBe(true);
      expect(wrapper.find(".policy-string-field-select").exists()).toBe(false);
    });

    it("renders runai-select when options are available", () => {
      const wrapper = mountComponent({ policyRules: { options: [{ value: "1", displayed: "Option 1" }] } });

      expect(wrapper.find(".policy-string-field-select").exists()).toBe(true);
      expect(wrapper.find(".policy-string-field-input").exists()).toBe(false);
    });
  });

  describe("Computed", () => {
    describe("disableInput", () => {
      it("disables input when policyRules canEdit is false", () => {
        const wrapper = mountComponent({ policyRules: { canEdit: false } });
        expect(wrapper.vm.disableInput).toBe(true);
        const input = wrapper.findComponent(QInput);
        expect(input.props().disable).toBe(true);
      });
    });
    describe("showExplanatoryTooltip", () => {
      it("showExplanatoryTooltip is true when tooltip prop is provided", () => {
        const wrapper = mountComponent({ tooltip: "This is a tooltip" });
        expect(wrapper.vm.showExplanatoryTooltip).toBe(true);
      });

      it("showExplanatoryTooltip is false when tooltip prop is not provided", () => {
        const wrapper = mountComponent({ disable: true, tooltip: null });
        expect(wrapper.vm.showExplanatoryTooltip).toBe(false);
        expect(wrapper.findComponent(".explanatory-tooltip").exists()).toBe(false);
      });
    });
  });

  describe("Props", () => {
    describe("rules", () => {
      it("applies custom rules if provided", () => {
        const customRule = vi.fn((val) => val !== "" || "Required field");
        const wrapper = mountComponent({ rules: [customRule] });
        const input = wrapper.findComponent(QInput);
        expect(input.props().rules.includes(customRule)).toBe(true);
      });
    });
    describe("policyRules", () => {
      it("applies required rule when policyRules.required is true", () => {
        const wrapper = mountComponent({ policyRules: { required: true } });
        const input = wrapper.findComponent(QInput);
        expect(input.props().rules.includes(wrapper.vm.requiredRule)).toBe(true);
      });
    });
  });

  describe("emits", () => {
    describe("update:model-value", () => {
      it("emits update:model-value when input value changes", async () => {
        const wrapper = mountComponent();
        const input = wrapper.find(".policy-string-field-input input");
        await input.setValue("new value");

        expect(wrapper.emitted()["update:model-value"][0]).toEqual(["new value"]);
      });
    });
  });
});
