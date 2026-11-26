/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { RunaiPriorityList } from "@/components/common/runai-priority-list";
import { EOrgUnitFormTooltip } from "@/models/org-unit-forms.model";
import NodePoolsPriorityList, { type Props } from "./node-pools-priority-list.vue";

const NODE_POOLS = ["default", "a100-pool", "h100-pool"];

const mountList = (props: Partial<Props> = {}) =>
  mount(NodePoolsPriorityList, {
    props: { nodePools: NODE_POOLS, modelValue: "default", ...props },
    slots: { default: '<div aid="node-pool-content">content</div>' },
    global: { plugins: [Quasar] },
  });

describe("NodePoolsPriorityList", () => {
  it("shows the node pools priority list with the selected node pool", () => {
    const wrapper = mountList({ modelValue: "a100-pool" });

    const list = wrapper.findComponent(RunaiPriorityList);
    expect(list.props()).toMatchObject({
      title: "Node pools",
      subtitle: "Priority order",
      tooltip: EOrgUnitFormTooltip.NodePoolsPriorityOrder,
      items: NODE_POOLS,
      modelValue: "a100-pool",
    });
  });

  it("hides the list but keeps the content when there's only one node pool", () => {
    const wrapper = mountList({ nodePools: ["default"] });

    expect(wrapper.findComponent(RunaiPriorityList).exists()).toBe(false);
    expect(wrapper.find('[aid="node-pool-content"]').exists()).toBe(true);
  });

  it("emits the selected node pool", async () => {
    const wrapper = mountList();

    await wrapper.find('[aid="node-pools-priority-list-item-h100-pool"]').trigger("click");

    expect(wrapper.emitted("update:model-value")).toEqual([["h100-pool"]]);
  });

  it("emits the new node pool order", () => {
    const wrapper = mountList();
    const reordered = ["h100-pool", "default", "a100-pool"];

    wrapper.findComponent(RunaiPriorityList).vm.$emit("update:items", reordered);

    expect(wrapper.emitted("update:node-pools")).toEqual([[reordered]]);
  });

  it("passes the disabled reorder state and tooltip to the list", () => {
    const wrapper = mountList({ disableReorder: true, disableReorderTooltip: "Reorder in the GPU quota step" });

    const list = wrapper.findComponent(RunaiPriorityList);
    expect(list.props("disableReorder")).toBe(true);
    expect(list.props("disableReorderTooltip")).toBe("Reorder in the GPU quota step");
  });
});
