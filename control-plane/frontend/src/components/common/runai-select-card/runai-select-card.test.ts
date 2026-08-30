import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { RunaiSelectCard } from "./index";

describe("runai-select-card", () => {
  it("should display content as default slot", async () => {
    const slotContent = "the content inside the card";
    const cmp = mount(RunaiSelectCard, {
      global: {
        plugins: [Quasar],
      },
      slots: {
        default: slotContent,
      },
      props: {
        value: true,
      },
    });

    expect(cmp.html()).toContain(slotContent);
  });

  it("should not have 'is-checked' class", async () => {
    const cmp = mount(RunaiSelectCard, {
      global: {
        plugins: [Quasar],
      },
      props: {
        value: false,
      },
    });

    const wrapperEl = cmp.find(".card-content");
    expect(wrapperEl.classes("is-checked")).toBe(false);
  });

  it("should have 'is-checked' class", async () => {
    const cmp = mount(RunaiSelectCard, {
      global: {
        plugins: [Quasar],
      },
      props: {
        value: true,
      },
    });

    const wrapperEl = cmp.find(".card-content");
    expect(wrapperEl.classes("is-checked")).toBe(true);
    const inputEl = cmp.find("input");
    expect(inputEl.attributes("type")).toBe("checkbox");
  });

  it("should emit 'clicked' event when clicking on the cmp", async () => {
    const cmp = mount(RunaiSelectCard, {
      global: {
        plugins: [Quasar],
      },
      props: {
        value: false,
      },
    });

    const inputEl = cmp.find("input");
    expect(inputEl.attributes("type")).toBe("checkbox");
    await inputEl.trigger("click");
    expect(cmp.emitted()).toHaveProperty("clicked");
  });
});
