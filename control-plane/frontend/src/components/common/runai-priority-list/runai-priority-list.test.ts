/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest";
import { flushPromises, mount } from "@vue/test-utils";
import { nextTick } from "vue";
import { Quasar, QItem } from "quasar";
import { VueDraggable } from "vue-draggable-plus";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import RunaiPriorityList, { type Props } from "./runai-priority-list.vue";

const ITEMS = ["default", "a100-pool", "h100-pool"];

const mountList = (props: Partial<Props> = {}) =>
  mount(RunaiPriorityList, {
    props: { title: "Node pools", items: ITEMS, modelValue: "default", ...props },
    global: { plugins: [Quasar] },
  });

const moveButton = (wrapper: ReturnType<typeof mountList>, direction: "up" | "down", item: string) =>
  wrapper.find(`[aid="runai-priority-list-move-${direction}-${item}"]`);

describe("RunaiPriorityList", () => {
  it("renders the title, subtitle and help tooltip", () => {
    const wrapper = mountList({ subtitle: "Priority order", tooltip: "Higher pools are tried first" });

    expect(wrapper.find(".priority-list-title").text()).toContain("Node pools");
    expect(wrapper.find(".priority-list-subtitle").text()).toBe("Priority order");
    expect(wrapper.findComponent(RunaiTooltip).props("tooltipText")).toBe("Higher pools are tried first");
    expect(wrapper.find("section").attributes("aria-label")).toBe("Node pools");
    expect(wrapper.find("nav").exists()).toBe(false);
  });

  it("leaves out the subtitle and tooltip when not given", () => {
    const wrapper = mountList();

    expect(wrapper.find(".priority-list-subtitle").exists()).toBe(false);
    expect(wrapper.findComponent(RunaiTooltip).exists()).toBe(false);
  });

  it("lists the items in order with their number and marks the selected one", () => {
    const wrapper = mountList({ modelValue: "a100-pool" });

    const items = wrapper.findAllComponents(QItem);
    expect(items.map((item) => item.find(".item-title").text())).toEqual(ITEMS);
    expect(items.map((item) => item.find(".item-number").text())).toEqual(["1", "2", "3"]);
    expect(items.map((item) => item.classes().includes("priority-list-item--active"))).toEqual([false, true, false]);
    expect(items[1].get(".item-select").attributes("aria-current")).toBe("true");
    expect(items[0].get(".item-select").attributes("aria-current")).toBeUndefined();
    expect(items[1].find(".item-indicator").exists()).toBe(true);
    expect(items[0].find(".item-indicator").exists()).toBe(false);
  });

  it("clamps the item title to two lines", () => {
    const wrapper = mountList();

    expect(wrapper.find(".item-title").classes()).toContain("ellipsis-2-lines");
  });

  it("emits the selected item on click", async () => {
    const wrapper = mountList();

    await wrapper.find('[aid="runai-priority-list-select-h100-pool"]').trigger("click");

    expect(wrapper.emitted("update:model-value")).toEqual([["h100-pool"]]);
  });

  it("emits the new order when an item moves up or down without selecting it", async () => {
    const wrapper = mountList();

    await moveButton(wrapper, "up", "a100-pool").trigger("click");
    await moveButton(wrapper, "down", "a100-pool").trigger("click");

    expect(wrapper.emitted("update:items")).toEqual([
      [["a100-pool", "default", "h100-pool"]],
      [["default", "h100-pool", "a100-pool"]],
    ]);
    expect(wrapper.emitted("update:model-value")).toBeUndefined();
  });

  it("emits the new order when an item is dragged", () => {
    const wrapper = mountList();
    const dragged = ["h100-pool", "default", "a100-pool"];

    wrapper.findComponent(VueDraggable).vm.$emit("update:modelValue", dragged);

    expect(wrapper.emitted("update:items")).toEqual([[dragged]]);
  });

  it("announces the new position of a dragged item", async () => {
    const wrapper = mountList();
    const draggable = wrapper.findComponent(VueDraggable);

    draggable.vm.$emit("start", { oldIndex: 2 });
    draggable.vm.$emit("end", { oldIndex: 2, newIndex: 0 });
    await nextTick();

    expect(wrapper.find('[aria-live="polite"]').text()).toBe("h100-pool moved to position 1 of 3");
  });

  it("gives selecting and moving an item separate controls", () => {
    const wrapper = mountList();

    const row = wrapper.findAllComponents(QItem)[1];
    expect(row.props("clickable")).toBe(false);
    expect(row.attributes("role")).toBe("listitem");
    expect(row.get(".item-select").element.tagName).toBe("BUTTON");
    expect(row.findAll(".item-arrows button")).toHaveLength(2);
  });

  it("does not change the items it was given", async () => {
    const items = [...ITEMS];
    const wrapper = mountList({ items });

    await moveButton(wrapper, "down", "default").trigger("click");

    expect(items).toEqual(ITEMS);
  });

  it("disables moving past the edges of the list", () => {
    const wrapper = mountList();

    expect(moveButton(wrapper, "up", "default").attributes("disabled")).toBeDefined();
    expect(moveButton(wrapper, "down", "h100-pool").attributes("disabled")).toBeDefined();
    expect(moveButton(wrapper, "down", "default").attributes("disabled")).toBeUndefined();
  });

  it("enables dragging by default", () => {
    const wrapper = mountList();

    expect(wrapper.findComponent(VueDraggable).props("disabled")).toBe(false);
  });

  it("disables dragging and every arrow when reordering is disabled", async () => {
    const wrapper = mountList({ disableReorder: true, disableReorderTooltip: "Reorder in the GPU quota step" });

    expect(wrapper.findComponent(VueDraggable).props("disabled")).toBe(true);

    const arrows = wrapper.findAll(".item-arrows button");
    expect(arrows).toHaveLength(ITEMS.length * 2);
    expect(arrows.every((arrow) => arrow.attributes("disabled") !== undefined)).toBe(true);

    await moveButton(wrapper, "up", "a100-pool").trigger("click");
    expect(wrapper.emitted("update:items")).toBeUndefined();
  });

  it("tells keyboard users why reordering is disabled", () => {
    const wrapper = mountList({ disableReorder: true, disableReorderTooltip: "Reorder in the GPU quota step" });

    const reasonId = wrapper.findAllComponents(QItem)[0].get(".item-select").attributes("aria-describedby");
    expect(reasonId).toBeTruthy();
    expect(wrapper.find(`[id="${reasonId}"]`).text()).toBe("Reorder in the GPU quota step");
  });

  it("keeps focus on the moved item and announces its new position", async () => {
    const wrapper = mount(RunaiPriorityList, {
      attachTo: document.body,
      props: {
        title: "Node pools",
        items: ITEMS,
        modelValue: "default",
        "onUpdate:items": (items: string[]): void => {
          void wrapper.setProps({ items });
        },
      },
      global: { plugins: [Quasar] },
    });

    await moveButton(wrapper, "up", "a100-pool").trigger("click");
    await flushPromises();

    expect(document.activeElement).toBe(moveButton(wrapper, "down", "a100-pool").element);
    expect(wrapper.find('[aria-live="polite"]').text()).toBe("a100-pool moved to position 1 of 3");
    wrapper.unmount();
  });
});
