import type { DOMWrapper, VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, expect, test } from "vitest";
import { VolumeSizeSection } from ".";
import type { TTestEmit } from "@/models/unit-test.model";
import { EMemoryUnitLable, EMemoryUnitValue } from "@/models/global.model";
import type { ClaimSizeCustomization } from "@/swagger-models/k8s-objects-tracker-client";

describe("volume-size-section.vue", () => {
  let volumeSizeSectionCmp: VueWrapper<InstanceType<typeof VolumeSizeSection>>;

  function getClaimSizeInput(): DOMWrapper<HTMLInputElement> {
    const claimSizeInput: DOMWrapper<HTMLInputElement> = volumeSizeSectionCmp.find("[aid=claim-size-input]");
    return claimSizeInput;
  }

  function getUnitSelect(): DOMWrapper<HTMLSelectElement> {
    const unitSelect: DOMWrapper<HTMLSelectElement> = volumeSizeSectionCmp.find("[aid=unit-select]");
    return unitSelect;
  }

  function mountComponent(modelValue = "", claimSizeCustomization?: ClaimSizeCustomization) {
    volumeSizeSectionCmp = mount(VolumeSizeSection, {
      global: {
        plugins: [Quasar],
      },
      props: {
        modelValue,
        claimSizeCustomization,
      },
    });
  }

  test("component is loading properly", () => {
    mountComponent();
    expect(volumeSizeSectionCmp.html()).toContain("Claim size");
  });

  test("component inputs are showing the default values when no modelValue is passed", () => {
    mountComponent();
    const claimSizeInputValue = getClaimSizeInput().element.value;
    const unitSelectValue = getUnitSelect().text();
    expect(claimSizeInputValue).toBe("0");
    expect(unitSelectValue).toBe("");
  });

  test("component inputs are showing the default values when modelValue is passed", async () => {
    mountComponent("40M");
    const claimSizeInputValue = getClaimSizeInput().element.value;
    const unitSelectValue = getUnitSelect().text();
    expect(claimSizeInputValue).toBe("40");
    expect(unitSelectValue).toBe(EMemoryUnitLable.MB);
  });

  test.each([
    { modelValue: "40M", sizeNum: 40, selectedUnit: EMemoryUnitValue.MB },
    { modelValue: "1G", sizeNum: 1, selectedUnit: EMemoryUnitValue.GB },
  ])(
    "should have the correct values in the model depending on the prop",
    async ({ modelValue, sizeNum, selectedUnit }) => {
      mountComponent(modelValue);
      expect(volumeSizeSectionCmp.vm.sizeNum).toBe(sizeNum);
      expect(volumeSizeSectionCmp.vm.selectedUnit).toBe(selectedUnit);
    },
  );

  test("component should emit the correct calculated value when claim size input is changed", async () => {
    mountComponent("20G");
    await getClaimSizeInput().setValue("40");
    const expectedEmits: TTestEmit<string> = [["40G"]];
    expect(volumeSizeSectionCmp.emitted()["update:model-value"]).toEqual(expectedEmits);
  });

  describe("step validation", () => {
    describe("stepBytesSize", () => {
      test("should return 0 when no step is configured", () => {
        mountComponent("1G", { min: "1G" });
        expect(volumeSizeSectionCmp.vm.stepBytesSize).toBe(0);
      });

      test("should return correct bytes for step in GB", () => {
        mountComponent("1G", { step: "1G" });
        expect(volumeSizeSectionCmp.vm.stepBytesSize).toBe(1_000_000_000);
      });

      test("should return correct bytes for step in MB", () => {
        mountComponent("1G", { step: "500M" });
        expect(volumeSizeSectionCmp.vm.stepBytesSize).toBe(500_000_000);
      });

      test("should return correct bytes for step in GiB", () => {
        mountComponent("1G", { step: "1Gi" });
        expect(volumeSizeSectionCmp.vm.stepBytesSize).toBe(1_073_741_824);
      });
    });

    describe("inputStep", () => {
      test("should return undefined when no step is configured", () => {
        mountComponent("1G", { min: "1G" });
        expect(volumeSizeSectionCmp.vm.inputStep).toBeUndefined();
      });

      test("should convert step to selected unit (GB step with GB unit)", () => {
        mountComponent("1G", { step: "1G" });
        expect(volumeSizeSectionCmp.vm.inputStep).toBe(1);
      });

      test("should convert step to selected unit (GB step with MB unit)", () => {
        mountComponent("1000M", { step: "1G" });
        expect(volumeSizeSectionCmp.vm.inputStep).toBe(1000);
      });

      test("should convert step to selected unit (500MB step with GB unit)", () => {
        mountComponent("1G", { step: "500M" });
        expect(volumeSizeSectionCmp.vm.inputStep).toBe(0.5);
      });

      test("should convert step to selected unit (500MB step with MB unit)", () => {
        mountComponent("500M", { step: "500M" });
        expect(volumeSizeSectionCmp.vm.inputStep).toBe(500);
      });
    });

    describe("isValidStep", () => {
      test("should return true when no step is configured", () => {
        mountComponent("1G", { min: "1G" });
        expect(volumeSizeSectionCmp.vm.isValidStep("2G")).toBe(true);
      });

      test("should return true for valid step increment from min", () => {
        mountComponent("1G", { min: "1G", step: "1G" });
        expect(volumeSizeSectionCmp.vm.isValidStep("2G")).toBe(true);
        expect(volumeSizeSectionCmp.vm.isValidStep("3G")).toBe(true);
        expect(volumeSizeSectionCmp.vm.isValidStep("10G")).toBe(true);
      });

      test("should return error message for invalid step increment", () => {
        mountComponent("1G", { min: "1G", step: "1G" });
        const result = volumeSizeSectionCmp.vm.isValidStep("1500M");
        expect(result).not.toBe(true);
        expect(typeof result).toBe("string");
      });

      test("should validate step correctly with different units (1000MB = 1GB)", () => {
        mountComponent("1G", { min: "1G", step: "1G" });
        expect(volumeSizeSectionCmp.vm.isValidStep("2000M")).toBe(true);
        expect(volumeSizeSectionCmp.vm.isValidStep("3000M")).toBe(true);
      });

      test("should validate step correctly with 500MB step", () => {
        mountComponent("500M", { min: "500M", step: "500M" });
        expect(volumeSizeSectionCmp.vm.isValidStep("1G")).toBe(true);
        expect(volumeSizeSectionCmp.vm.isValidStep("1500M")).toBe(true);
        expect(volumeSizeSectionCmp.vm.isValidStep("2G")).toBe(true);
      });

      test("should return error for values not aligned with step", () => {
        mountComponent("500M", { min: "500M", step: "500M" });
        const result = volumeSizeSectionCmp.vm.isValidStep("750M");
        expect(result).not.toBe(true);
      });
    });

    describe("stepErrorMessage", () => {
      test("should include step, min, and max when all are defined", () => {
        mountComponent("1G", { min: "1G", max: "10G", step: "1G" });
        const message = volumeSizeSectionCmp.vm.stepErrorMessage;
        expect(message).toContain("1G");
        expect(message).toContain("10G");
      });

      test("should include step and min when only those are defined", () => {
        mountComponent("1G", { min: "1G", step: "500M" });
        const message = volumeSizeSectionCmp.vm.stepErrorMessage;
        expect(message).toContain("500M");
        expect(message).toContain("1G");
      });

      test("should include only step when min is default", () => {
        mountComponent("1G", { step: "1G" });
        const message = volumeSizeSectionCmp.vm.stepErrorMessage;
        expect(message).toContain("1G");
      });
    });

    describe("sizeFieldValidations with step", () => {
      test("should include step validation when step is configured", () => {
        mountComponent("1G", { step: "1G" });
        const validations = volumeSizeSectionCmp.vm.sizeFieldValidations();
        expect(validations.length).toBeGreaterThan(0);
        expect(validations.some((v) => v === volumeSizeSectionCmp.vm.isValidStep)).toBe(true);
      });

      test("should include both min/max and step validations when all are configured", () => {
        mountComponent("1G", { min: "1G", max: "10G", step: "1G" });
        const validations = volumeSizeSectionCmp.vm.sizeFieldValidations();
        expect(validations).toContain(volumeSizeSectionCmp.vm.isBetweenMinMax);
        expect(validations).toContain(volumeSizeSectionCmp.vm.isValidStep);
      });

      test("should not include step validation when step is not configured", () => {
        mountComponent("1G", { min: "1G", max: "10G" });
        const validations = volumeSizeSectionCmp.vm.sizeFieldValidations();
        expect(validations).not.toContain(volumeSizeSectionCmp.vm.isValidStep);
      });
    });
  });
});
