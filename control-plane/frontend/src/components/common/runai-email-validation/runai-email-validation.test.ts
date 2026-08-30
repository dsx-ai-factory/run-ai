import type { DOMWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { describe, it, expect } from "vitest";
import { RunaiEmailValidation } from "./index";
import { Quasar } from "quasar";
import { errorMessages } from "@/common/error-message.constant";

describe("EmailValidationInput.vue", () => {
  it("renders the input with the correct placeholder", () => {
    const wrapper = mount(RunaiEmailValidation, {
      global: {
        plugins: [Quasar],
      },
      props: {
        modelValue: "",
        placeholder: "Enter your email",
      },
    });

    const input: DOMWrapper<HTMLInputElement> = wrapper.find(".runai-email-validation input");
    expect(input.element.placeholder).toBe("Enter your email");
  });

  it("renders the input with the default placeholder when no placeholder is provided", () => {
    const wrapper = mount(RunaiEmailValidation, {
      global: {
        plugins: [Quasar],
      },
      props: {
        modelValue: "",
      },
    });

    const input: DOMWrapper<HTMLInputElement> = wrapper.find(".runai-email-validation input");
    expect(input.element.placeholder).toBe("Enter email address");
  });

  it('emits "update:modelValue" on input change', async () => {
    const wrapper = mount(RunaiEmailValidation, {
      global: {
        plugins: [Quasar],
      },
      props: {
        modelValue: "",
      },
    });

    const input = wrapper.find(".runai-email-validation input");
    await input.setValue("test@example.com");

    expect(wrapper.emitted("update:modelValue")).toBeTruthy();
    expect(wrapper.emitted("update:modelValue")?.[0]).toEqual(["test@example.com"]);
  });

  it('emits "invalid" when there is an error', async () => {
    const wrapper = mount(RunaiEmailValidation, {
      global: {
        plugins: [Quasar],
      },
      props: {
        modelValue: "",
        rules: [(val: string) => !!val || errorMessages.EMAIL_NOT_EMPTY],
      },
    });

    const input = wrapper.find(".runai-email-validation input");
    await input.setValue(""); // Triggering error

    expect(wrapper.emitted("invalid")).toBeTruthy();
  });

  it("validates with multiple rules and emits invalid correctly", async () => {
    const wrapper = mount(RunaiEmailValidation, {
      global: {
        plugins: [Quasar],
      },
      props: {
        modelValue: "",
        rules: [
          (val: string) => !!val || errorMessages.EMAIL_NOT_EMPTY,
          (val: string) => /.+@.+\..+/.test(val) || errorMessages.EMAIL_NOT_VALID,
        ],
      },
    });

    const input = wrapper.find(".runai-email-validation input");
    await input.setValue("invalid-email"); // Triggering second rule

    expect(wrapper.emitted("invalid")).toBeTruthy();
  });

  it("reacts to changes in modelValue prop", async () => {
    const wrapper = mount(RunaiEmailValidation, {
      global: {
        plugins: [Quasar],
      },
      props: {
        modelValue: "",
      },
    });

    await wrapper.setProps({ modelValue: "new@example.com" });
    const input: DOMWrapper<HTMLInputElement> = wrapper.find(".runai-email-validation input");
    expect(input.element.value).toBe("new@example.com");
  });

  it("does not emit 'invalid' with an empty rules array", async () => {
    const wrapper = mount(RunaiEmailValidation, {
      global: {
        plugins: [Quasar],
      },
      props: {
        modelValue: "",
        rules: [],
      },
    });

    const input = wrapper.find(".runai-email-validation input");
    await input.setValue("test@example.com");

    const invalidEvents: [][] | undefined = wrapper.emitted("invalid");
    expect(invalidEvents).toBeTruthy();
    expect(invalidEvents?.[0]).toEqual([false]);
  });

  it("keeps component state consistent with prop changes", async () => {
    const wrapper = mount(RunaiEmailValidation, {
      global: {
        plugins: [Quasar],
      },
      props: {
        modelValue: "",
      },
    });

    await wrapper.setProps({ modelValue: "new@example.com" });
    const input: DOMWrapper<HTMLInputElement> = wrapper.find(".runai-email-validation input");
    expect(input.element.value).toBe("new@example.com");

    await wrapper.setProps({ modelValue: "" });
    expect(input.element.value).toBe("");
  });
});
