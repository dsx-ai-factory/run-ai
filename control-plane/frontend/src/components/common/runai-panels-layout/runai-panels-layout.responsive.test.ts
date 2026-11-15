/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import { nextTick, reactive } from "vue";
import { Quasar } from "quasar";
import type * as QuasarModule from "quasar";
import { afterEach, describe, expect, it, vi } from "vitest";
import RunaiPanelsLayout from "./runai-panels-layout.vue";
import RunaiWizardLayout from "../runai-wizard-layout/runai-wizard-layout.vue";

// Quasar's Node build has a non-reactive Screen plugin. Model the browser's
// reactive viewport while retaining the real layout and Quasar components.
const Screen = reactive({ width: 1600 });
vi.mock("quasar", async (importOriginal) => {
  const quasar = await importOriginal<typeof QuasarModule>();
  return { ...quasar, useQuasar: () => ({ ...quasar.useQuasar(), screen: Screen }) };
});

describe("responsive panel layout", () => {
  const wrappers: ReturnType<typeof mount>[] = [];
  const initialWidth = Screen.width;
  afterEach(() => {
    wrappers.splice(0).forEach((wrapper) => wrapper.unmount());
    Screen.width = initialWidth;
  });

  function render(props = {}) {
    const wrapper = mount(RunaiPanelsLayout, {
      attachTo: document.body,
      props,
      slots: {
        left: '<template #left="{ collapsed }"><button>{{ collapsed ? "Icon" : "Step title" }}</button></template>',
        default: "<input aria-label='Name' />",
        right: "<input aria-label='Summary value' />",
      },
      global: { plugins: [Quasar] },
    });
    wrappers.push(wrapper);
    return wrapper;
  }

  it("hides the right panel below 1280 without collapsing the node pool steps", async () => {
    Screen.width = 1280;
    const wrapper = render({ hideRightOnSmallScreen: true });
    expect(wrapper.get(".panel-right").isVisible()).toBe(true);
    Screen.width = 1279;
    await nextTick();
    expect(wrapper.get(".panel-right").isVisible()).toBe(false);
    expect(wrapper.get(".panel-left").text()).toBe("Step title");
    Screen.width = 1280;
    await nextTick();
    expect(wrapper.get(".panel-right").isVisible()).toBe(true);
  });

  it("collapses independently of right panel visibility and expands on hover", async () => {
    Screen.width = 1279;
    const wrapper = render({ collapseLeftOnSmallScreen: true });
    expect(wrapper.get(".panel-left").text()).toBe("Icon");
    expect(wrapper.get(".panel-right").isVisible()).toBe(true);
    await wrapper.get(".panel-left").trigger("mouseenter");
    expect(wrapper.get(".panel-left").text()).toBe("Step title");
    await wrapper.get(".panel-left").trigger("mouseleave");
    expect(wrapper.get(".panel-left").text()).toBe("Icon");
  });

  it("expands for keyboard focus and keeps the rail expanded while focused", async () => {
    Screen.width = 1024;
    const wrapper = render({ collapseLeftOnSmallScreen: true });
    await wrapper.get(".panel-left button").trigger("focusin");
    await wrapper.get(".panel-left").trigger("mouseenter");
    await wrapper.get(".panel-left").trigger("mouseleave");
    expect(wrapper.get(".panel-left").text()).toBe("Step title");
    await wrapper.get(".panel-left button").trigger("focusout");
    expect(wrapper.get(".panel-left").text()).toBe("Icon");
  });

  it("preserves summary state across resizing and does not opt existing consumers into collapsing", async () => {
    Screen.width = 1600;
    const wrapper = render();
    await wrapper.get(".panel-right input").setValue("kept");
    Screen.width = 1024;
    await nextTick();
    expect(wrapper.get(".panel-left").text()).toBe("Step title");
    expect(wrapper.get(".panel-right").isVisible()).toBe(true);
    await wrapper.setProps({ hideRightOnSmallScreen: true });
    expect(wrapper.get(".panel-right").isVisible()).toBe(false);
    Screen.width = 1600;
    await nextTick();
    expect((wrapper.get(".panel-right input").element as HTMLInputElement).value).toBe("kept");
  });

  it("keeps step status and disabled navigation in the wizard's mini rail", async () => {
    Screen.width = 1279;
    const wrapper = mount(RunaiWizardLayout, {
      attachTo: document.body,
      props: {
        steps: [
          { name: "first", title: "Cluster and name", done: true },
          { name: "second", title: "Scheduling mode", done: false },
          { name: "third", title: "Topology", done: false, disabled: true, hasError: true },
        ],
        activeStep: "second",
        collapseLeftOnSmallScreen: true,
        hideRightOnSmallScreen: true,
      },
      slots: { default: "<input />", "right-sidebar": "Summary" },
      global: { plugins: [Quasar] },
    });
    wrappers.push(wrapper);
    expect(wrapper.get(".wizard-step-title").isVisible()).toBe(false);
    expect(wrapper.findAll(".wizard-step-indicator")).toHaveLength(3);
    expect(wrapper.find(".wizard-step--completed .ph-check").exists()).toBe(true);
    expect(wrapper.find(".wizard-step--error .ph-warning-circle").exists()).toBe(true);
    await wrapper.get('[aid="wizard-step-item-third"]').trigger("click");
    expect(wrapper.emitted("step-click")).toBeUndefined();
    await wrapper.get('[aid="wizard-step-item-first"]').trigger("click");
    expect(wrapper.emitted("step-click")?.[0]?.[0]).toMatchObject({ name: "first" });
    await wrapper.get(".panel-left").trigger("mouseenter");
    expect(wrapper.get(".wizard-step-title").isVisible()).toBe(true);
    await wrapper.setProps({ orientation: "horizontal" });
    expect(wrapper.find(".panel-left").exists()).toBe(false);
    expect(wrapper.get(".wizard-step-title").isVisible()).toBe(true);
    expect(wrapper.get(".wizard-layout-horizontal-body").text()).toContain("Summary");
  });

  it("lets keyboard users reveal disabled steps and gives a transparent overlay solid backing", async () => {
    Screen.width = 1024;
    const wrapper = mount(RunaiWizardLayout, {
      attachTo: document.body,
      props: {
        steps: [{ name: "first", title: "Cluster and name", done: false, disabled: true }],
        activeStep: "first",
        collapseLeftOnSmallScreen: true,
        transparentStepsPanel: true,
        disableNavigation: true,
      },
      global: { plugins: [Quasar] },
    });
    wrappers.push(wrapper);
    const rail = wrapper.get(".panel-left");
    expect(rail.attributes("tabindex")).toBe("0");
    expect(wrapper.get(".wizard-step-title").isVisible()).toBe(false);
    await rail.trigger("focusin");
    expect(wrapper.get(".wizard-step-title").isVisible()).toBe(true);
    expect(wrapper.find(".panel-left .runai-panel-container--transparent").exists()).toBe(false);
    await wrapper.get('[aid="wizard-step-item-first"]').trigger("click");
    expect(wrapper.emitted("step-click")).toBeUndefined();
    await rail.trigger("focusout");
    expect(wrapper.find(".panel-left .runai-panel-container--transparent").exists()).toBe(true);
    Screen.width = 1280;
    await nextTick();
    expect(rail.attributes("tabindex")).toBeUndefined();
    expect(wrapper.find(".panel-left .runai-panel-container--transparent").exists()).toBe(true);
  });
});
