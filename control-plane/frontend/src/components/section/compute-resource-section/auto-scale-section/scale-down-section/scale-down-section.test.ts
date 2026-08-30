import { mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, it, expect } from "vitest";
import { ScaleDownSection } from ".";
import {
  SCALE_DOWN_OPTIONS,
  NEVER_SCALE_DOWN_ID,
  CUSTOM_SCALE_DOWN_ID,
  EScaleDownSectionEvents,
} from "./scale-down-section.model";
import type { AutoScalingRules } from "@/swagger-models/assets-service-client";

interface IReplicaScalingProps {
  modelValue: { scaleToZeroRetentionSeconds: number | undefined };
  disabled: boolean;
  isAutoScaleEnabled: boolean;
  policyRules?: AutoScalingRules | null;
}

describe("scale-down-section.vue", () => {
  const baseProps: IReplicaScalingProps = {
    modelValue: { scaleToZeroRetentionSeconds: undefined },
    disabled: false,
    isAutoScaleEnabled: true,
  };

  function getMountedComponent(props: IReplicaScalingProps): VueWrapper<InstanceType<typeof ScaleDownSection>> {
    return mount(ScaleDownSection, {
      global: {
        plugins: [Quasar],
      },
      props,
    });
  }

  describe("Mounted", () => {
    it("renders the section and tooltip", () => {
      const wrapper = getMountedComponent(baseProps);

      expect(wrapper.html()).toContain("Set when the replicas should be automatically scaled down to zero");
      expect(wrapper.findComponent({ name: "runai-tooltip" }).exists()).toBe(true);
    });

    it("disables select if disabled or isAutoScaleEnabled is false", async () => {
      const wrapper = getMountedComponent(baseProps);

      await wrapper.setProps({ disabled: true });
      expect(wrapper.findComponent({ name: "runai-select" }).props("disable")).toBe(true);
      await wrapper.setProps({ disabled: false, isAutoScaleEnabled: false });
      expect(wrapper.findComponent({ name: "runai-select" }).props("disable")).toBe(true);
    });

    it("shows q-tooltip if isAutoScaleEnabled is false", async () => {
      const wrapper = getMountedComponent(baseProps);

      expect(wrapper.findAllComponents({ name: "q-tooltip" }).length).toBe(1);
      await wrapper.setProps({ isAutoScaleEnabled: false });
      const tooltips = wrapper.findAllComponents({ name: "q-tooltip" });
      expect(tooltips.length).toBe(2);
    });

    it("disables 'Never' option if isAutoScaleEnabled is true", async () => {
      const wrapper = getMountedComponent({ ...baseProps, isAutoScaleEnabled: true });
      const neverOption = wrapper.vm.neverOption;
      expect(neverOption.disable).toBe(true);
      expect(neverOption.disabledTooltip).toBe("To enable this, set the minimum number of replicas to at least 1");
    });

    it("enables 'Never' option if isAutoScaleEnabled is false", async () => {
      const wrapper = getMountedComponent({ ...baseProps, isAutoScaleEnabled: false });
      const neverOption = wrapper.vm.neverOption;
      expect(neverOption.disable).toBe(false);
      expect(neverOption.disabledTooltip).toBe("");
    });

    it("initializes with correct custom value if modelValue is custom", async () => {
      const customSeconds = 7 * 60;
      const wrapper = getMountedComponent({ ...baseProps, modelValue: { scaleToZeroRetentionSeconds: customSeconds } });
      if (wrapper.vm.scaleDownSelected) {
        expect(wrapper.vm.scaleDownSelected.value).toBe(CUSTOM_SCALE_DOWN_ID);
      }
      expect(wrapper.vm.customScaleDownSelectedAsMinutes).toBe(7);
    });

    it("initializes with 'Never' if modelValue is undefined or 0", async () => {
      const wrapper = getMountedComponent({
        ...baseProps,
        modelValue: { scaleToZeroRetentionSeconds: 0 },
        isAutoScaleEnabled: true,
      });

      expect(wrapper.vm.scaleDownSelected?.value).toBe(NEVER_SCALE_DOWN_ID);
    });

    it("initializes with correct predefined option if modelValue matches", async () => {
      const wrapper = getMountedComponent({
        ...baseProps,
        modelValue: { scaleToZeroRetentionSeconds: 900 },
        isAutoScaleEnabled: true,
      });
      expect(wrapper.vm.scaleDownSelected?.value).toBe("900");
    });
  });

  describe("Methods", () => {
    describe("getScaleDownOptionById", () => {
      it("getScaleDownOptionById returns correct option for valid IDs", () => {
        const wrapper = getMountedComponent(baseProps);
        for (const opt of SCALE_DOWN_OPTIONS) {
          if (typeof opt.value === "string") {
            const result = wrapper.vm.getScaleDownOptionById(opt.value);
            if (opt.value === NEVER_SCALE_DOWN_ID) {
              expect(result).toEqual({
                ...opt,
                disable: true,
                disabledTooltip: "To enable this, set the minimum number of replicas to at least 1",
              });
            } else {
              expect(result).toEqual(opt);
            }
          }
        }
      });

      it("getScaleDownOptionById throws error for invalid ID", () => {
        const wrapper = getMountedComponent(baseProps);
        expect(() => wrapper.vm.getScaleDownOptionById("invalid-id")).toThrow(
          "Scale down option with id invalid-id not found",
        );
      });
    });

    describe("getNeverOption", () => {
      it("getNeverOption returns the 'Never' option with correct value and label", () => {
        const wrapper = getMountedComponent(baseProps);
        const neverOption = wrapper.vm.getNeverOption();
        expect(neverOption.value).toBe(NEVER_SCALE_DOWN_ID);
        expect(neverOption.label).toBe("Never");
      });

      it("getNeverOption returns the same as getScaleDownOptionById(NEVER_SCALE_DOWN_ID)", () => {
        const wrapper = getMountedComponent(baseProps);
        const neverOption = wrapper.vm.getNeverOption();
        const byIdOption = wrapper.vm.getScaleDownOptionById(NEVER_SCALE_DOWN_ID);
        expect(neverOption).toEqual(byIdOption);
      });
    });

    describe("findScaleDownOptionByRetentionSeconds", () => {
      it("returns neverOption if retentionSeconds is undefined", () => {
        const wrapper = getMountedComponent(baseProps);
        const result = wrapper.vm.findScaleDownOptionByRetentionSeconds(undefined);
        expect(result).toEqual(wrapper.vm.neverOption);
      });

      it("returns neverOption if retentionSeconds is 0", () => {
        const wrapper = getMountedComponent(baseProps);
        const result = wrapper.vm.findScaleDownOptionByRetentionSeconds(0);
        expect(result).toEqual(wrapper.vm.neverOption);
      });

      it("returns the matching predefined option if retentionSeconds matches a predefined value", () => {
        const wrapper = getMountedComponent(baseProps);
        for (const opt of SCALE_DOWN_OPTIONS) {
          if (typeof opt.value === "string" && !isNaN(Number(opt.value))) {
            const result = wrapper.vm.findScaleDownOptionByRetentionSeconds(Number(opt.value));
            expect(result).toEqual(opt);
          }
        }
      });

      it("returns the custom option if retentionSeconds is a positive number not matching a predefined value", () => {
        const wrapper = getMountedComponent(baseProps);
        // 123 is not a predefined value and > 0
        const result = wrapper.vm.findScaleDownOptionByRetentionSeconds(123);
        const customOption = SCALE_DOWN_OPTIONS.find((opt) => opt.value === CUSTOM_SCALE_DOWN_ID);
        expect(result).toEqual(customOption);
      });

      it("returns neverOption for negative value", () => {
        const wrapper = getMountedComponent(baseProps);
        const result = wrapper.vm.findScaleDownOptionByRetentionSeconds(-5);
        expect(result).toEqual(wrapper.vm.neverOption);
      });
    });

    describe("initializeScaleDownSelectedOption", () => {
      it("sets scaleDownSelected to never option for undefined", () => {
        const wrapper = getMountedComponent(baseProps);
        wrapper.vm.initializeScaleDownSelectedOption(undefined);
        expect(wrapper.vm.scaleDownSelected?.value).toBe(NEVER_SCALE_DOWN_ID);
      });

      it("sets scaleDownSelected to never option for 0", () => {
        const wrapper = getMountedComponent(baseProps);
        wrapper.vm.initializeScaleDownSelectedOption(0);
        expect(wrapper.vm.scaleDownSelected?.value).toBe(NEVER_SCALE_DOWN_ID);
      });

      it("sets scaleDownSelected to predefined option for matching value", () => {
        const wrapper = getMountedComponent(baseProps);
        wrapper.vm.initializeScaleDownSelectedOption(900);
        expect(wrapper.vm.scaleDownSelected?.value).toBe("900");
      });

      it("sets scaleDownSelected to custom option for custom value", () => {
        const wrapper = getMountedComponent(baseProps);
        wrapper.vm.initializeScaleDownSelectedOption(123);
        expect(wrapper.vm.scaleDownSelected?.value).toBe(CUSTOM_SCALE_DOWN_ID);
      });

      it("sets customScaleDownSelectedAsMinutes for custom option with retentionSeconds", () => {
        const wrapper = getMountedComponent(baseProps);
        wrapper.vm.initializeScaleDownSelectedOption(300);
        // 300 is a predefined, so should not set custom minutes
        expect(wrapper.vm.customScaleDownSelectedAsMinutes).toBe(0);
        wrapper.vm.initializeScaleDownSelectedOption(123);
        expect(wrapper.vm.customScaleDownSelectedAsMinutes).toBe(2.05); // 123 / 60
      });

      it("sets neverOption.disable and disabledTooltip based on isAutoScaleEnabled", () => {
        const wrapper = getMountedComponent({ ...baseProps, isAutoScaleEnabled: true });
        wrapper.vm.initializeScaleDownSelectedOption(undefined);
        expect(wrapper.vm.neverOption.disable).toBe(true);
        expect(wrapper.vm.neverOption.disabledTooltip).toBe(
          "To enable this, set the minimum number of replicas to at least 1",
        );
        const wrapper2 = getMountedComponent({ ...baseProps, isAutoScaleEnabled: false });
        wrapper2.vm.initializeScaleDownSelectedOption(undefined);
        expect(wrapper2.vm.neverOption.disable).toBe(false);
        expect(wrapper2.vm.neverOption.disabledTooltip).toBe("");
      });
    });

    describe("isValidScaleToZero", () => {
      it("returns true for a value within the allowed range (min)", () => {
        const wrapper = getMountedComponent(baseProps);
        expect(wrapper.vm.isValidScaleToZero("1")).toBe(true);
      });
      it("returns true for a value within the allowed range (max)", () => {
        const wrapper = getMountedComponent(baseProps);
        expect(wrapper.vm.isValidScaleToZero("60")).toBe(true);
      });
      it("returns true for a string number within the range", () => {
        const wrapper = getMountedComponent(baseProps);
        expect(wrapper.vm.isValidScaleToZero("15")).toBe(true);
      });
      it("returns error for a value below the minimum", () => {
        const wrapper = getMountedComponent(baseProps);
        expect(wrapper.vm.isValidScaleToZero("0")).toBe(wrapper.vm.customScaleToZeroMessageError);
      });
      it("returns error for a value above the maximum", () => {
        const wrapper = getMountedComponent(baseProps);
        expect(wrapper.vm.isValidScaleToZero("61")).toBe(wrapper.vm.customScaleToZeroMessageError);
      });
      it("returns error for a non-numeric string", () => {
        const wrapper = getMountedComponent(baseProps);
        expect(wrapper.vm.isValidScaleToZero("abc")).toBe(wrapper.vm.customScaleToZeroMessageError);
      });
      it("returns error for an empty string", () => {
        const wrapper = getMountedComponent(baseProps);
        expect(wrapper.vm.isValidScaleToZero("")).toBe(wrapper.vm.customScaleToZeroMessageError);
      });
      it("returns error for null", () => {
        const wrapper = getMountedComponent(baseProps);
        expect(wrapper.vm.isValidScaleToZero(null as unknown as string)).toBe(wrapper.vm.customScaleToZeroMessageError);
      });
      it("returns error for undefined", () => {
        const wrapper = getMountedComponent(baseProps);
        expect(wrapper.vm.isValidScaleToZero(undefined as unknown as string)).toBe(
          wrapper.vm.customScaleToZeroMessageError,
        );
      });
    });

    describe("initializeScaleDownOptions", () => {
      it("sets scaleDownOptions to default SCALE_DOWN_OPTIONS if no policyRules are provided", () => {
        const wrapper = getMountedComponent(baseProps);
        wrapper.vm.initializeScaleDownOptions();
        expect(wrapper.vm.scaleDownOptions).toEqual(SCALE_DOWN_OPTIONS);
      });

      it("filters only numeric scaleDownOptions by min and max from policyRules, keeps custom and never", () => {
        const min = 300; // 5 minutes
        const max = 900; // 15 minutes
        const wrapper = getMountedComponent({
          ...baseProps,
          policyRules: {
            scaleToZeroRetentionSeconds: { min, max },
          } as AutoScalingRules,
        });
        wrapper.vm.initializeScaleDownOptions();
        const expected = SCALE_DOWN_OPTIONS.filter((opt) => {
          if (opt.value === CUSTOM_SCALE_DOWN_ID || opt.value === NEVER_SCALE_DOWN_ID) {
            return true;
          }
          return Number(opt.value) >= min && Number(opt.value) <= max;
        });
        expect(wrapper.vm.scaleDownOptions).toEqual(expected);
      });

      it("always keeps the custom and never options, even if min/max exclude all numeric options", () => {
        const min = 9999;
        const max = 10000;
        const wrapper = getMountedComponent({
          ...baseProps,
          policyRules: {
            scaleToZeroRetentionSeconds: { min, max },
          } as AutoScalingRules,
        });
        wrapper.vm.initializeScaleDownOptions();
        const hasCustom = wrapper.vm.scaleDownOptions.some((opt) => opt.value === CUSTOM_SCALE_DOWN_ID);
        const hasNever = wrapper.vm.scaleDownOptions.some((opt) => opt.value === NEVER_SCALE_DOWN_ID);
        expect(wrapper.vm.scaleDownOptions.length).toBe(2);
        expect(hasCustom).toBe(true);
        expect(hasNever).toBe(true);
      });

      it("never filters out the never option", () => {
        const min = 9999;
        const max = 10000;
        const wrapper = getMountedComponent({
          ...baseProps,
          policyRules: {
            scaleToZeroRetentionSeconds: { min, max },
          } as AutoScalingRules,
        });
        wrapper.vm.initializeScaleDownOptions();
        const hasNever = wrapper.vm.scaleDownOptions.some((opt) => opt.value === NEVER_SCALE_DOWN_ID);
        expect(hasNever).toBe(true);
      });

      it("never filters out the custom option", () => {
        const min = 9999;
        const max = 10000;
        const wrapper = getMountedComponent({
          ...baseProps,
          policyRules: {
            scaleToZeroRetentionSeconds: { min, max },
          } as AutoScalingRules,
        });
        wrapper.vm.initializeScaleDownOptions();
        const hasCustom = wrapper.vm.scaleDownOptions.some((opt) => opt.value === CUSTOM_SCALE_DOWN_ID);
        expect(hasCustom).toBe(true);
      });

      it("filters only numeric scaleDownOptions by min from policyRules, keeps custom and never", () => {
        const min = 600; // 10 minutes
        const wrapper = getMountedComponent({
          ...baseProps,
          policyRules: {
            scaleToZeroRetentionSeconds: { min },
          } as AutoScalingRules,
        });
        wrapper.vm.initializeScaleDownOptions();
        const expected = SCALE_DOWN_OPTIONS.filter((opt) => {
          if (opt.value === CUSTOM_SCALE_DOWN_ID || opt.value === NEVER_SCALE_DOWN_ID) {
            return true;
          }
          return Number(opt.value) >= min;
        });
        expect(wrapper.vm.scaleDownOptions).toEqual(expected);
      });

      it("filters only numeric scaleDownOptions by max from policyRules, keeps custom and never", () => {
        const max = 600; // 10 minutes
        const wrapper = getMountedComponent({
          ...baseProps,
          policyRules: {
            scaleToZeroRetentionSeconds: { max },
          } as AutoScalingRules,
        });
        wrapper.vm.initializeScaleDownOptions();
        const expected = SCALE_DOWN_OPTIONS.filter((opt) => {
          if (opt.value === CUSTOM_SCALE_DOWN_ID || opt.value === NEVER_SCALE_DOWN_ID) {
            return true;
          }
          return Number(opt.value) <= max;
        });
        expect(wrapper.vm.scaleDownOptions).toEqual(expected);
      });
    });
  });

  describe("Policy Rules", () => {
    const getPolicyRules = (overrides = {}) => ({
      scaleToZeroRetentionSeconds: {
        min: 2,
        max: 10,
        step: 2,
        required: true,
        canEdit: true,
        ...overrides,
      },
    });

    it("disables custom input if canEdit is false", async () => {
      const wrapper = getMountedComponent({
        ...baseProps,
        policyRules: getPolicyRules({ canEdit: false }),
      });
      wrapper.vm.onScaleDownChanged(CUSTOM_SCALE_DOWN_ID);
      await wrapper.vm.$nextTick();
      const inputWrapper = wrapper.findComponent({ name: "policy-number-input" });
      expect(inputWrapper.exists()).toBe(true);
      const input = inputWrapper.find("input");
      expect(input.exists()).toBe(true);
      expect(input.attributes()["disabled"]).to.exist;
    });

    it("enforces min, max, and step from policyRules", async () => {
      const wrapper = getMountedComponent({
        ...baseProps,
        policyRules: getPolicyRules({ min: 3, max: 9, step: 3 }),
      });
      wrapper.vm.onScaleDownChanged(CUSTOM_SCALE_DOWN_ID);
      await wrapper.vm.$nextTick();
      const inputWrapper = wrapper.findComponent({ name: "policy-number-input" });
      expect(inputWrapper.exists()).toBe(true);
      const input = inputWrapper.find("input");
      expect(input.exists()).toBe(true);
      expect(input.attributes()["min"]).toBe("3");
      expect(input.attributes()["max"]).toBe("9");
      expect(input.attributes()["step"]).toBe("3");
    });

    it("shows error if required and value is empty", async () => {
      const wrapper = getMountedComponent({
        ...baseProps,
        policyRules: getPolicyRules({ required: true }),
      });
      wrapper.vm.onScaleDownChanged(CUSTOM_SCALE_DOWN_ID);
      await wrapper.vm.$nextTick();
      const input = wrapper.findComponent({ name: "policy-number-input" });
      // Simulate validation rule
      const rules = input.props("rules") as unknown[];
      const requiredRule = rules.find((r) => typeof r === "function" && r.name === "requiredRule");
      if (typeof requiredRule === "function") {
        expect(requiredRule(undefined)).toBeDefined();
      }
    });

    it("emits correct value when step is enforced", async () => {
      const wrapper = getMountedComponent({
        ...baseProps,
        policyRules: getPolicyRules({ min: 1, step: 2 }),
      });
      wrapper.vm.onScaleDownChanged(CUSTOM_SCALE_DOWN_ID);
      await wrapper.vm.$nextTick();
      // Simulate entering a value that matches the step
      wrapper.vm.updateCustomScaleToZero(4);
      expect(wrapper.vm.customScaleDownSelectedAsMinutes).toBe(4);
      expect(wrapper.emitted()[EScaleDownSectionEvents.UpdateModelValue]).toEqual([
        [{ scaleToZeroRetentionSeconds: 60 }], // initial
        [{ scaleToZeroRetentionSeconds: 240 }], // after update
      ]);
    });

    it("shows error if value is below min or above max", async () => {
      const wrapper = getMountedComponent({
        ...baseProps,
        policyRules: getPolicyRules({ min: 3, max: 9 }),
      });
      wrapper.vm.onScaleDownChanged(CUSTOM_SCALE_DOWN_ID);
      await wrapper.vm.$nextTick();
      const input = wrapper.findComponent({ name: "policy-number-input" });
      const rules = input.props("rules") as unknown[];
      const minRule = rules.find((r) => typeof r === "function" && r.name === "minRule");
      const maxRule = rules.find((r) => typeof r === "function" && r.name === "maxRule");
      if (typeof minRule === "function") {
        expect(minRule(2)).not.toBe(true);
      }
      if (typeof maxRule === "function") {
        expect(maxRule(10)).not.toBe(true);
      }
    });
  });

  describe("Emits", () => {
    describe("updateCustomScaleToZero", () => {
      it("emits correct event and sets customScaleDownSelectedAsMinutes for a valid number", () => {
        const wrapper = getMountedComponent(baseProps);
        wrapper.vm.updateCustomScaleToZero(5);
        expect(wrapper.vm.customScaleDownSelectedAsMinutes).toBe(5);
        const expectedEvent = [[{ scaleToZeroRetentionSeconds: 300 }]];
        expect(wrapper.emitted()[EScaleDownSectionEvents.UpdateModelValue]).toEqual(expectedEvent);
      });

      it("emits correct event for a string number", () => {
        const wrapper = getMountedComponent(baseProps);
        wrapper.vm.updateCustomScaleToZero("7");
        expect(wrapper.vm.customScaleDownSelectedAsMinutes).toBe(7);
        const expectedEvent = [[{ scaleToZeroRetentionSeconds: 420 }]];
        expect(wrapper.emitted()[EScaleDownSectionEvents.UpdateModelValue]).toEqual(expectedEvent);
      });

      it("emits correct event and sets customScaleDownSelectedAsMinutes to 1 for null", () => {
        const wrapper = getMountedComponent(baseProps);
        wrapper.vm.updateCustomScaleToZero(null);
        expect(wrapper.vm.customScaleDownSelectedAsMinutes).toBe(1);
        const expectedEvent = [[{ scaleToZeroRetentionSeconds: 60 }]];
        expect(wrapper.emitted()[EScaleDownSectionEvents.UpdateModelValue]).toEqual(expectedEvent);
      });

      it("emits correct event and sets customScaleDownSelectedAsMinutes to 1 for undefined", () => {
        const wrapper = getMountedComponent(baseProps);
        wrapper.vm.updateCustomScaleToZero(undefined as unknown as number | string | null);
        expect(wrapper.vm.customScaleDownSelectedAsMinutes).toBe(1);
        const expectedEvent = [[{ scaleToZeroRetentionSeconds: 60 }]];
        expect(wrapper.emitted()[EScaleDownSectionEvents.UpdateModelValue]).toEqual(expectedEvent);
      });

      it("emits correct event and sets customScaleDownSelectedAsMinutes to 1 for empty string", () => {
        const wrapper = getMountedComponent(baseProps);
        wrapper.vm.updateCustomScaleToZero("");
        expect(wrapper.vm.customScaleDownSelectedAsMinutes).toBe(1);
        const expectedEvent = [[{ scaleToZeroRetentionSeconds: 60 }]];
        expect(wrapper.emitted()[EScaleDownSectionEvents.UpdateModelValue]).toEqual(expectedEvent);
      });
    });

    describe("onScaleDownChanged", () => {
      it("emits correct event for a predefined option", () => {
        const wrapper = getMountedComponent(baseProps);
        // Use a predefined value (e.g., "900")
        wrapper.vm.onScaleDownChanged("900");
        const expectedEvent = [[{ scaleToZeroRetentionSeconds: 900 }]];
        expect(wrapper.emitted()[EScaleDownSectionEvents.UpdateModelValue]).toEqual(expectedEvent);
        expect(wrapper.vm.customScaleDownSelectedAsMinutes).toBe(0);
      });

      it("emits correct event with undefined for 'Never' option", () => {
        const wrapper = getMountedComponent(baseProps);
        wrapper.vm.onScaleDownChanged(NEVER_SCALE_DOWN_ID);
        const expectedEvent = [[{ scaleToZeroRetentionSeconds: undefined }]];
        expect(wrapper.emitted()[EScaleDownSectionEvents.UpdateModelValue]).toEqual(expectedEvent);
      });

      it("emits correct event and sets customScaleDownSelectedAsMinutes for 'Custom' option", () => {
        const wrapper = getMountedComponent(baseProps);
        wrapper.vm.onScaleDownChanged(CUSTOM_SCALE_DOWN_ID);
        const expectedEvent = [[{ scaleToZeroRetentionSeconds: 60 }]]; // default is 1 minute = 60 seconds
        expect(wrapper.emitted()[EScaleDownSectionEvents.UpdateModelValue]).toEqual(expectedEvent);
        expect(wrapper.vm.customScaleDownSelectedAsMinutes).toBe(1);
      });

      it("does nothing if an invalid option is passed", () => {
        const wrapper = getMountedComponent(baseProps);
        wrapper.vm.onScaleDownChanged("invalid-option");
        // Should not emit any event
        expect(wrapper.emitted()[EScaleDownSectionEvents.UpdateModelValue]).toBeUndefined();
      });
    });
  });
});
