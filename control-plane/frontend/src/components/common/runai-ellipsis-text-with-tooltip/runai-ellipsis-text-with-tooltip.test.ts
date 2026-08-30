import { describe, expect, it } from "vitest";
import { Quasar } from "quasar";
import { mount } from "@vue/test-utils";
import RunaiEllipsisTextWithTooltip from "./runai-ellipsis-text-with-tooltip.vue";

describe("runai-ellipsis-text-with-tooltip", () => {
  it("renders the provided text", () => {
    const wrapper = mount(RunaiEllipsisTextWithTooltip, {
      global: { plugins: [Quasar] },
      props: { text: "Subject: All authenticated users and service accounts" },
    });
    expect(wrapper.text()).toContain("Subject: All authenticated users and service accounts");
  });

  it("does not render a tooltip when the text is not truncated", () => {
    // jsdom has no layout, so scrollWidth === clientWidth === 0 → not truncated
    const wrapper = mount(RunaiEllipsisTextWithTooltip, {
      global: { plugins: [Quasar] },
      props: { text: "short" },
    });
    expect(wrapper.findComponent({ name: "QTooltip" }).exists()).toBe(false);
  });
});
