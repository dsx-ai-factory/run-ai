/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar, type ValidationRule } from "quasar";
import { describe } from "vitest";
import { errorMessages } from "@/common/error-message.constant";
import { RunaiDurationInput } from ".";
describe("runai-duration-input.vue", () => {
  let runaiDurationInputCmp: VueWrapper<InstanceType<typeof RunaiDurationInput>>;

  function mountComponent(modelValue: number | null, hideSeconds?: boolean, rules?: ValidationRule[]) {
    runaiDurationInputCmp = mount(RunaiDurationInput, {
      global: {
        plugins: [Quasar],
      },
      props: { modelValue, hideSeconds, rules },
    });
  }

  describe("when model value is null", () => {
    beforeEach(() => {
      mountComponent(null);
    });

    test("component is loading properly", () => {
      expect(runaiDurationInputCmp.html()).toContain("days");
    });

    test("seconds input should not exist if hideSeconds prop is true", async () => {
      await runaiDurationInputCmp.setProps({ hideSeconds: true });
      expect(runaiDurationInputCmp.find('[aid="seconds-input"]').exists()).toBe(false);
    });
  });

  describe("When model value is 0", () => {
    beforeEach(() => {
      mountComponent(0);
    });

    test.each(["days", "hours", "minutes", "seconds"])(
      "component should show the time inputs when auto delete time is set",
      async (currInput) => {
        expect(runaiDurationInputCmp.find(`[aid="${currInput}-input"]`).exists()).toBe(true);
      },
    );

    test("all the inputs should show be empty when the model value is 0", () => {
      expect((runaiDurationInputCmp.find(`[aid="days-input"]`).element as HTMLInputElement).value).toBe("");
      expect((runaiDurationInputCmp.find(`[aid="hours-input"]`).element as HTMLInputElement).value).toBe("");
      expect((runaiDurationInputCmp.find(`[aid="minutes-input"]`).element as HTMLInputElement).value).toBe("");
      expect((runaiDurationInputCmp.find(`[aid="seconds-input"]`).element as HTMLInputElement).value).toBe("0");
    });
  });

  test("only the seconds input should show 30 when the model value is set to 30", () => {
    mountComponent(30);
    expect((runaiDurationInputCmp.find(`[aid="days-input"]`).element as HTMLInputElement).value).toBe("");
    expect((runaiDurationInputCmp.find(`[aid="hours-input"]`).element as HTMLInputElement).value).toBe("");
    expect((runaiDurationInputCmp.find(`[aid="minutes-input"]`).element as HTMLInputElement).value).toBe("");
    expect((runaiDurationInputCmp.find(`[aid="seconds-input"]`).element as HTMLInputElement).value).toBe("30");
  });

  test("the minutes and seconds inputs should show the correct values accroding to the modelValue", () => {
    mountComponent(90);
    expect((runaiDurationInputCmp.find(`[aid="days-input"]`).element as HTMLInputElement).value).toBe("");
    expect((runaiDurationInputCmp.find(`[aid="hours-input"]`).element as HTMLInputElement).value).toBe("");
    expect((runaiDurationInputCmp.find(`[aid="minutes-input"]`).element as HTMLInputElement).value).toBe("1");
    expect((runaiDurationInputCmp.find(`[aid="seconds-input"]`).element as HTMLInputElement).value).toBe("30");
  });

  test("component should emit the calculated seconds depending on the inputs", async () => {
    mountComponent(0);
    await runaiDurationInputCmp.find(`[aid="minutes-input"]`).setValue(4);
    expect(runaiDurationInputCmp.emitted("update:model-value")).toEqual([[240]]);
  });

  test("component should emit the calculated seconds depending on the inputs", async () => {
    mountComponent(0);
    await runaiDurationInputCmp.find(`[aid="hours-input"]`).setValue(1);
    expect(runaiDurationInputCmp.emitted("update:model-value")).toEqual([[3600]]);
  });

  test("component should emit the calculated seconds depending on the inputs", async () => {
    mountComponent(0);
    await runaiDurationInputCmp.find(`[aid="days-input"]`).setValue(1);
    expect(runaiDurationInputCmp.emitted("update:model-value")).toEqual([[86400]]);
  });

  describe("built-in non-negative validation", () => {
    test("surfaces VALID_TIME_FRAME error when modelValue is negative and no rules prop is passed", async () => {
      mountComponent(-1);
      const qFieldVm = runaiDurationInputCmp.findComponent({ name: "QField" }).vm as unknown as {
        validate: () => Promise<boolean>;
      };
      const valid = await qFieldVm.validate();
      expect(valid).toBe(false);
      expect(runaiDurationInputCmp.html()).toContain(errorMessages.VALID_TIME_FRAME);
    });

    test("passes validation when modelValue is null", async () => {
      mountComponent(null);
      const qFieldVm = runaiDurationInputCmp.findComponent({ name: "QField" }).vm as unknown as {
        validate: () => Promise<boolean>;
      };
      const valid = await qFieldVm.validate();
      expect(valid).toBe(true);
    });

    test("each q-input shows error state when modelValue is negative", () => {
      mountComponent(-1);
      const inputAids = ["days-input", "hours-input", "minutes-input", "seconds-input"];
      for (const aid of inputAids) {
        const fieldEl = runaiDurationInputCmp.find(`[aid="${aid}"]`).element.closest(".q-field");
        expect(fieldEl?.classList.contains("q-field--error")).toBe(true);
      }
    });

    test("q-inputs do not show error state when modelValue is valid", () => {
      mountComponent(60);
      const inputAids = ["days-input", "hours-input", "minutes-input", "seconds-input"];
      for (const aid of inputAids) {
        const fieldEl = runaiDurationInputCmp.find(`[aid="${aid}"]`).element.closest(".q-field");
        expect(fieldEl?.classList.contains("q-field--error")).toBe(false);
      }
    });

    test("consumer-supplied failing rule wins on error message", async () => {
      const customMessage = "custom-error-message";
      mountComponent(-1, false, [
        (val: number | null | undefined) => (val !== null && val !== undefined ? customMessage : true),
      ]);
      const qFieldVm = runaiDurationInputCmp.findComponent({ name: "QField" }).vm as unknown as {
        validate: () => Promise<boolean>;
      };
      const valid = await qFieldVm.validate();
      expect(valid).toBe(false);
      expect(runaiDurationInputCmp.html()).toContain(customMessage);
      expect(runaiDurationInputCmp.html()).not.toContain(errorMessages.VALID_TIME_FRAME);
    });
  });
});
