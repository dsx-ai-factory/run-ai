import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, it, expect } from "vitest";
import { RunaiSegmentedButton, type ISegmentedButtonOption } from ".";

const OPTIONS: ISegmentedButtonOption[] = [
  { label: "Rule builder", value: "rule-builder" },
  { label: "YAML editor", value: "yaml" },
];

function mountControl(modelValue: string | null, options: ISegmentedButtonOption[] = OPTIONS) {
  return mount(RunaiSegmentedButton, {
    global: { plugins: [Quasar] },
    props: { modelValue, options },
  });
}

function segments(wrapper: ReturnType<typeof mountControl>) {
  return wrapper.findAll('button[role="radio"]');
}

describe("runai-segmented-button.vue", () => {
  it("renders one radio segment per option with a radiogroup wrapper", () => {
    const wrapper = mountControl("yaml");
    expect(wrapper.attributes("role")).toBe("radiogroup");
    const buttons = segments(wrapper);
    expect(buttons).toHaveLength(2);
    expect(buttons.map((b) => b.text())).toEqual(["Rule builder", "YAML editor"]);
  });

  it("marks the option matching modelValue as selected (class + aria-checked + tab stop)", () => {
    const buttons = segments(mountControl("yaml"));
    expect(buttons[1].classes()).toContain("runai-segment--selected");
    expect(buttons[1].attributes("aria-checked")).toBe("true");
    expect(buttons[1].attributes("tabindex")).toBe("0");
    expect(buttons[0].attributes("aria-checked")).toBe("false");
    expect(buttons[0].attributes("tabindex")).toBe("-1");
  });

  it("emits the value when an unselected segment is clicked", async () => {
    const wrapper = mountControl("yaml");
    await segments(wrapper)[0].trigger("click");
    expect(wrapper.emitted("update:model-value")?.[0]).toEqual(["rule-builder"]);
  });

  it("does not emit when the already-selected segment is clicked", async () => {
    const wrapper = mountControl("yaml");
    await segments(wrapper)[1].trigger("click");
    expect(wrapper.emitted("update:model-value")).toBeUndefined();
  });

  it("does not emit when a disabled segment is clicked, and marks it disabled", async () => {
    const wrapper = mountControl("yaml", [
      { label: "Rule builder", value: "rule-builder", disable: true },
      { label: "YAML editor", value: "yaml" },
    ]);
    const buttons = segments(wrapper);
    expect(buttons[0].classes()).toContain("runai-segment--disabled");
    expect(buttons[0].attributes("disabled")).toBeDefined();
    await buttons[0].trigger("click");
    expect(wrapper.emitted("update:model-value")).toBeUndefined();
  });

  it("moves the selection with the arrow keys, skipping disabled segments", async () => {
    const wrapper = mountControl("yaml");
    await segments(wrapper)[1].trigger("keydown", { key: "ArrowLeft" });
    expect(wrapper.emitted("update:model-value")?.[0]).toEqual(["rule-builder"]);
  });

  it("renders an option description (stretch layout)", () => {
    const wrapper = mountControl("a", [
      { label: "A", value: "a", description: "first choice" },
      { label: "B", value: "b" },
    ]);
    const buttons = segments(wrapper);
    expect(buttons[0].classes()).toContain("runai-segment--with-description");
    expect(buttons[0].text()).toContain("first choice");
  });
});
