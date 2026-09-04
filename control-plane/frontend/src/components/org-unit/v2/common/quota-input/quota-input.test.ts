/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import { describe, it, expect, vi, beforeAll } from "vitest";
import { Quasar } from "quasar";
import QuotaInput, { type Props, type UnitOption } from "./quota-input.vue";
import { EQuotaInputState } from "@/models/org-unit.model";
import { Units } from "@/swagger-models/org-unit-service-client";

describe("QuotaInput", () => {
  // Suppress QSelect internal warning during tests
  beforeAll(() => {
    const originalWarn = console.warn;
    vi.spyOn(console, "warn").mockImplementation((...args) => {
      const msg = args[0];
      if (typeof msg === "string" && msg.includes('Component emitted event "update:model-value"')) {
        return;
      }
      originalWarn(...args);
    });
  });

  const mountQuotaInput = (props: Partial<Props> = {}) => {
    return mount(QuotaInput, {
      global: {
        plugins: [Quasar],
      },
      props: {
        modelValue: props.modelValue ?? null,
        ...props,
      },
    });
  };

  it("renders correctly with default props", () => {
    const wrapper = mountQuotaInput({
      modelValue: null,
    });
    expect(wrapper.exists()).toBe(true);
    expect(wrapper.find("input").exists()).toBe(true);
  });

  it("renders unit selection when unitOptions are provided", () => {
    const unitOptions: UnitOption[] = [
      { label: "MiB", value: Units.Mib },
      { label: "GiB", value: Units.Gb },
    ];
    const wrapper = mountQuotaInput({
      modelValue: 10,
      unit: Units.Mib,
      unitOptions,
    });

    const select = wrapper.findComponent({ name: "QSelect" });
    expect(select.exists()).toBe(true);
  });

  it("does not render unit selection if no options provided", () => {
    const wrapper = mountQuotaInput({
      modelValue: 10,
      unit: "MiB",
      unitOptions: [],
    });
    const select = wrapper.findComponent({ name: "QSelect" });
    expect(select.exists()).toBe(false);
  });

  it("displays message with correct color for error state", () => {
    const message = "Error message";
    const wrapper = mountQuotaInput({
      modelValue: 10,
      state: EQuotaInputState.Error,
      message,
    });
    const messageEl = wrapper.find(".quota-message");
    expect(messageEl.exists()).toBe(true);
    expect(messageEl.text()).toBe(message);
  });

  it("displays message with correct color for warning state", () => {
    const message = "Warning message";
    const wrapper = mountQuotaInput({
      modelValue: 10,
      state: EQuotaInputState.Warning,
      message,
    });
    const messageEl = wrapper.find(".quota-message");
    expect(messageEl.exists()).toBe(true);
    expect(messageEl.text()).toBe(message);
  });

  it("applies error state to input", () => {
    const wrapper = mountQuotaInput({
      modelValue: 10,
      state: EQuotaInputState.Error,
      label: "Test Input",
    });
    const input = wrapper.findComponent({ name: "QInput" });
    expect(input.props("error")).toBe(true);
  });

  it("emits update:model-value event on input", async () => {
    const wrapper = mountQuotaInput({
      modelValue: null,
    });
    const input = wrapper.find("input");
    await input.setValue(100);
    expect(wrapper.emitted("update:model-value")).toBeTruthy();
    expect(wrapper.emitted("update:model-value")?.[0]).toEqual([100]);
  });

  it("emits update:unit event on unit change", async () => {
    const unitOptions: UnitOption[] = [
      { label: "MiB", value: Units.Mib },
      { label: "GiB", value: Units.Gb },
    ];
    const wrapper = mountQuotaInput({
      modelValue: 10,
      unit: "MiB",
      unitOptions,
    });

    const select = wrapper.findComponent({ name: "QSelect" });
    expect(select.exists()).toBe(true);

    // Simulate unit change
    await select.vm.$emit("update:model-value", "GiB");
    expect(wrapper.emitted("update:unit")).toBeTruthy();
    expect(wrapper.emitted("update:unit")?.[0]).toEqual(["GiB"]);
  });

  describe("preventNull functionality", () => {
    it("allows null values when preventNull is false (default)", async () => {
      const wrapper = mountQuotaInput({
        modelValue: null,
        preventNull: false,
      });

      const input = wrapper.find("input");
      await input.setValue("");

      const emittedEvents = wrapper.emitted("update:model-value");
      expect(emittedEvents).toBeTruthy();
      expect(emittedEvents?.[0]).toEqual([null]);
    });

    it("converts null to defaultValue when preventNull is true", async () => {
      const wrapper = mountQuotaInput({
        modelValue: null,
        preventNull: true,
        defaultValue: 0,
      });

      const input = wrapper.find("input");
      await input.setValue("");

      const emittedEvents = wrapper.emitted("update:model-value");
      expect(emittedEvents).toBeTruthy();
      expect(emittedEvents?.[0]).toEqual([0]);
    });

    it("converts empty string to defaultValue when preventNull is true", async () => {
      const wrapper = mountQuotaInput({
        modelValue: 10,
        preventNull: true,
        defaultValue: 0,
      });

      const input = wrapper.find("input");
      await input.setValue("");

      const emittedEvents = wrapper.emitted("update:model-value");
      expect(emittedEvents).toBeTruthy();
      expect(emittedEvents?.[emittedEvents.length - 1]).toEqual([0]);
    });

    it("converts null to custom defaultValue when preventNull is true", async () => {
      const wrapper = mountQuotaInput({
        modelValue: null,
        preventNull: true,
        defaultValue: 5,
      });

      const input = wrapper.find("input");
      await input.setValue("");

      const emittedEvents = wrapper.emitted("update:model-value");
      expect(emittedEvents).toBeTruthy();
      expect(emittedEvents?.[0]).toEqual([5]);
    });

    it("does not convert null to defaultValue on blur when preventNull is false", async () => {
      const wrapper = mountQuotaInput({
        modelValue: null,
        preventNull: false,
      });

      const input = wrapper.find("input");
      await input.trigger("blur");

      const emittedEvents = wrapper.emitted("update:model-value");
      if (emittedEvents) {
        const lastEvent = emittedEvents[emittedEvents.length - 1];
        expect(lastEvent).toEqual([null]);
      }
    });

    it("allows valid numbers when preventNull is true", async () => {
      const wrapper = mountQuotaInput({
        modelValue: 10,
        preventNull: true,
        defaultValue: 0,
      });

      const input = wrapper.find("input");
      await input.setValue(50);

      const emittedEvents = wrapper.emitted("update:model-value");
      expect(emittedEvents).toBeTruthy();
      const lastEvent = emittedEvents?.[emittedEvents.length - 1];
      expect(lastEvent).toEqual([50]);
    });

    it("allows zero value when preventNull is true", async () => {
      const wrapper = mountQuotaInput({
        modelValue: 10,
        preventNull: true,
        defaultValue: 0,
      });

      const input = wrapper.find("input");
      await input.setValue(0);

      const emittedEvents = wrapper.emitted("update:model-value");
      expect(emittedEvents).toBeTruthy();
      const lastEvent = emittedEvents?.[emittedEvents.length - 1];
      expect(lastEvent).toEqual([0]);
    });

    it("allows negative values when preventNull is true", async () => {
      const wrapper = mountQuotaInput({
        modelValue: 10,
        preventNull: true,
        defaultValue: 0,
      });

      const input = wrapper.find("input");
      await input.setValue(-1);

      const emittedEvents = wrapper.emitted("update:model-value");
      expect(emittedEvents).toBeTruthy();
      const lastEvent = emittedEvents?.[emittedEvents.length - 1];
      expect(lastEvent).toEqual([-1]);
    });

    it("handles decimal values when preventNull is true", async () => {
      const wrapper = mountQuotaInput({
        modelValue: 10,
        preventNull: true,
        defaultValue: 0,
      });

      const input = wrapper.find("input");
      await input.setValue(10.5);

      const emittedEvents = wrapper.emitted("update:model-value");
      expect(emittedEvents).toBeTruthy();
      const lastEvent = emittedEvents?.[emittedEvents.length - 1];
      expect(lastEvent).toEqual([10.5]);
    });
  });

  describe("Input Validation", () => {
    describe("Keypress validation", () => {
      it("allows typing digits", async () => {
        const wrapper = mountQuotaInput({ modelValue: null });
        const input = wrapper.find("input");

        const event = new KeyboardEvent("keypress", { key: "5" });
        const preventDefaultSpy = vi.spyOn(event, "preventDefault");

        await input.element.dispatchEvent(event);

        expect(preventDefaultSpy).not.toHaveBeenCalled();
      });

      it("allows typing decimal point when none exists", async () => {
        const wrapper = mountQuotaInput({ modelValue: null });
        const input = wrapper.find("input");
        input.element.value = "10";

        const event = new KeyboardEvent("keypress", { key: "." });
        const preventDefaultSpy = vi.spyOn(event, "preventDefault");

        await input.element.dispatchEvent(event);

        expect(preventDefaultSpy).not.toHaveBeenCalled();
      });

      it("blocks typing second decimal point", async () => {
        const wrapper = mountQuotaInput({ modelValue: null });
        const input = wrapper.find("input");
        input.element.value = "10.5";

        const event = new KeyboardEvent("keypress", { key: "." });
        const preventDefaultSpy = vi.spyOn(event, "preventDefault");

        await input.element.dispatchEvent(event);

        expect(preventDefaultSpy).toHaveBeenCalled();
      });

      it("allows typing minus sign when none exists", async () => {
        const wrapper = mountQuotaInput({ modelValue: null });
        const input = wrapper.find("input");
        input.element.value = "";

        const event = new KeyboardEvent("keypress", { key: "-" });
        const preventDefaultSpy = vi.spyOn(event, "preventDefault");

        await input.element.dispatchEvent(event);

        expect(preventDefaultSpy).not.toHaveBeenCalled();
      });

      it("blocks typing second minus sign", async () => {
        const wrapper = mountQuotaInput({ modelValue: null });
        const input = wrapper.find("input");
        input.element.value = "-10";

        const event = new KeyboardEvent("keypress", { key: "-" });
        const preventDefaultSpy = vi.spyOn(event, "preventDefault");

        await input.element.dispatchEvent(event);

        expect(preventDefaultSpy).toHaveBeenCalled();
      });

      it("blocks typing letters", async () => {
        const wrapper = mountQuotaInput({ modelValue: null });
        const input = wrapper.find("input");

        const event = new KeyboardEvent("keypress", { key: "a" });
        const preventDefaultSpy = vi.spyOn(event, "preventDefault");

        await input.element.dispatchEvent(event);

        expect(preventDefaultSpy).toHaveBeenCalled();
      });
    });

    describe("Paste validation", () => {
      const createPasteEvent = (text: string) => {
        return {
          clipboardData: {
            getData: () => text,
          },
          preventDefault: vi.fn(),
        } as unknown as ClipboardEvent;
      };

      it("allows pasting valid positive integer", () => {
        const wrapper = mountQuotaInput({ modelValue: null });
        const qInput = wrapper.findComponent({ name: "QInput" });
        const pasteEvent = createPasteEvent("123");

        qInput.vm.$emit("paste", pasteEvent);

        expect(pasteEvent.preventDefault).not.toHaveBeenCalled();
      });

      it("allows pasting valid negative integer", () => {
        const wrapper = mountQuotaInput({ modelValue: null });
        const qInput = wrapper.findComponent({ name: "QInput" });
        const pasteEvent = createPasteEvent("-1");

        qInput.vm.$emit("paste", pasteEvent);

        expect(pasteEvent.preventDefault).not.toHaveBeenCalled();
      });

      it("allows pasting valid decimal", () => {
        const wrapper = mountQuotaInput({ modelValue: null });
        const qInput = wrapper.findComponent({ name: "QInput" });
        const pasteEvent = createPasteEvent("10.5");

        qInput.vm.$emit("paste", pasteEvent);

        expect(pasteEvent.preventDefault).not.toHaveBeenCalled();
      });

      it("allows pasting valid negative decimal", () => {
        const wrapper = mountQuotaInput({ modelValue: null });
        const qInput = wrapper.findComponent({ name: "QInput" });
        const pasteEvent = createPasteEvent("-2.75");

        qInput.vm.$emit("paste", pasteEvent);

        expect(pasteEvent.preventDefault).not.toHaveBeenCalled();
      });

      it("blocks pasting multiple minus signs", () => {
        const wrapper = mountQuotaInput({ modelValue: null });
        const qInput = wrapper.findComponent({ name: "QInput" });
        const pasteEvent = createPasteEvent("-333-222");

        qInput.vm.$emit("paste", pasteEvent);

        expect(pasteEvent.preventDefault).toHaveBeenCalled();
      });

      it("blocks pasting multiple decimal points", () => {
        const wrapper = mountQuotaInput({ modelValue: null });
        const qInput = wrapper.findComponent({ name: "QInput" });
        const pasteEvent = createPasteEvent("12.5.3");

        qInput.vm.$emit("paste", pasteEvent);

        expect(pasteEvent.preventDefault).toHaveBeenCalled();
      });

      it("blocks pasting letters", () => {
        const wrapper = mountQuotaInput({ modelValue: null });
        const qInput = wrapper.findComponent({ name: "QInput" });
        const pasteEvent = createPasteEvent("abc");

        qInput.vm.$emit("paste", pasteEvent);

        expect(pasteEvent.preventDefault).toHaveBeenCalled();
      });

      it("blocks pasting double minus at start", () => {
        const wrapper = mountQuotaInput({ modelValue: null });
        const qInput = wrapper.findComponent({ name: "QInput" });
        const pasteEvent = createPasteEvent("--1");

        qInput.vm.$emit("paste", pasteEvent);

        expect(pasteEvent.preventDefault).toHaveBeenCalled();
      });
    });
  });
});
