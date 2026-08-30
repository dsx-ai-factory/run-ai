import { mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, expect, it } from "vitest";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import ChartAreaZoomControls from "./chart-area-zoom-controls.vue";

describe("ChartAreaZoomControls", () => {
  function mountComponent(props: { externalActive?: boolean; verticalOffset?: number } = {}): VueWrapper {
    return mount(ChartAreaZoomControls, {
      global: { plugins: [Quasar] },
      props,
    });
  }

  it("uses Phosphor icons for zoom and reset", () => {
    const wrapper = mountComponent();
    const buttons = wrapper.findAllComponents({ name: "QBtn" });

    expect(buttons[0].props("icon")).toBe("ph-bold ph-magnifying-glass-plus");
    expect(buttons[1].props("icon")).toBe("ph-bold ph-arrow-counter-clockwise");
  });

  it("positions both tooltips to the left of their buttons", () => {
    const wrapper = mountComponent();
    const tooltips = wrapper.findAllComponents(RunaiTooltipWrapper);

    expect(tooltips).toHaveLength(2);
    tooltips.forEach((tooltip) => {
      expect(tooltip.props("anchor")).toBe("center left");
      expect(tooltip.props("self")).toBe("center right");
    });
  });

  it("toggles zoom and resets its active state", async () => {
    const wrapper = mountComponent();
    const buttons = wrapper.findAllComponents({ name: "QBtn" });

    expect(buttons[0].attributes("aria-pressed")).toBe("false");

    await buttons[0].trigger("click");
    expect(wrapper.emitted("zoom-toggle")).toEqual([[true]]);
    expect(buttons[0].attributes("aria-pressed")).toBe("true");

    await buttons[1].trigger("click");
    expect(wrapper.emitted("reset")).toEqual([[]]);
    expect(buttons[0].attributes("aria-pressed")).toBe("false");
  });

  it("synchronizes the active state from the parent", async () => {
    const wrapper = mountComponent({ externalActive: false });
    const zoomButton = wrapper.findAllComponents({ name: "QBtn" })[0];

    await wrapper.setProps({ externalActive: true });

    expect(zoomButton.attributes("aria-pressed")).toBe("true");
  });

  it("applies an explicit vertical offset without changing the default placement", async () => {
    const wrapper = mountComponent();

    expect(wrapper.attributes("style")).toContain("translateY(-0px)");

    await wrapper.setProps({ verticalOffset: 8 });

    expect(wrapper.attributes("style")).toContain("translateY(-8px)");
  });
});
