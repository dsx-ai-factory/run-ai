/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, beforeEach, vi } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import NodePoolTabs from "./node-pool-tabs.vue";
import type { NodePoolOption } from "./node-pool-tabs.vue";
import type { ISelectOption } from "@/models/global.model";

// Type for the component's exposed properties via vm
interface NodePoolTabsVM {
  showTabs: boolean;
  hasRemainingPools: boolean;
  remainingNodePools: NodePoolOption[];
  addButtonDisabledTooltip: string | undefined;
  addNodePoolOptions: ISelectOption[];
  pendingAddUpdate: NodePoolOption[] | null;
  openMenuId: string | null;
  getTabMenuOptions: (index: number) => ISelectOption[];
  handleMoveLeft: (nodePoolId: string) => void;
  handleRemove: (nodePoolId: string) => void;
  handleAddSelection: (nodePoolId: string) => void;
  handleAddMenuHide: () => void;
  handleMenuAction: (action: string, nodePoolId: string) => void;
}

describe("NodePoolTabs", () => {
  let wrapper: VueWrapper;

  const defaultNodePools: NodePoolOption[] = [
    { id: "pool-1", name: "Pool 1" },
    { id: "pool-2", name: "Pool 2" },
  ];

  const allAvailablePools: NodePoolOption[] = [
    { id: "pool-1", name: "Pool 1" },
    { id: "pool-2", name: "Pool 2" },
    { id: "pool-3", name: "Pool 3" },
  ];

  const mountComponent = (props = {}) => {
    return mount(NodePoolTabs, {
      props: {
        nodePools: defaultNodePools,
        modelValue: "pool-1",
        availableNodePools: allAvailablePools,
        ...props,
      },
      global: {
        plugins: [Quasar],
        stubs: {
          RunaiBtn: {
            template: '<button :disabled="disable" :data-aid="aid" @click="$emit(\'click\')"><slot /></button>',
            props: [
              "disable",
              "tooltip",
              "disabledTooltip",
              "aid",
              "round",
              "flat",
              "color",
              "textColor",
              "unelevated",
              "size",
              "icon",
            ],
          },
          RunaiSelectMenu: {
            template: '<div class="mock-select-menu"><slot /></div>',
            props: ["list", "tooltipWidth"],
            emits: ["selected-value", "before-show", "hide"],
          },
        },
      },
      slots: {
        default: '<div class="slot-content">Panel Content</div>',
      },
    });
  };

  const getVM = (w: VueWrapper): NodePoolTabsVM => w.vm as unknown as NodePoolTabsVM;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("Rendering", () => {
    describe("Tabs visibility based on availableNodePools", () => {
      it("shows tabs when availableNodePools has more than 1 pool", () => {
        wrapper = mountComponent();

        const tabsContainer = wrapper.find(".node-pool-tabs-container");
        expect(tabsContainer.exists()).toBe(true);
      });

      it("hides tabs when availableNodePools has only 1 pool", () => {
        const singlePool: NodePoolOption[] = [{ id: "pool-1", name: "Pool 1" }];
        wrapper = mountComponent({
          nodePools: singlePool,
          availableNodePools: singlePool,
        });

        const tabsContainer = wrapper.find(".node-pool-tabs-container");
        expect(tabsContainer.exists()).toBe(false);
      });

      it("always renders panel content (slot) regardless of pool count", () => {
        const singlePool: NodePoolOption[] = [{ id: "pool-1", name: "Pool 1" }];
        wrapper = mountComponent({
          nodePools: singlePool,
          availableNodePools: singlePool,
        });

        const slotContent = wrapper.find(".slot-content");
        expect(slotContent.exists()).toBe(true);
      });
    });

    describe("Tab rendering", () => {
      it("renders correct number of tabs based on nodePools", () => {
        wrapper = mountComponent();

        const tabs = wrapper.findAll(".q-tab");
        expect(tabs.length).toBe(2);
      });

      it("renders add button when there are remaining pools", () => {
        wrapper = mountComponent();

        const addButton = wrapper.find('[data-aid="add-nodepool-btn"]');
        expect(addButton.exists()).toBe(true);
      });
    });
  });

  describe("Props", () => {
    describe("disableListActions", () => {
      it("disables add button when disableListActions is true", () => {
        wrapper = mountComponent({ disableListActions: true });

        const addButton = wrapper.find('[data-aid="add-nodepool-btn"]');
        expect(addButton.attributes("disabled")).toBeDefined();
      });

      it("shows options button when disableListActions is true", () => {
        wrapper = mountComponent({ disableListActions: true });

        // Options button should still be visible (not hidden)
        const optionsButtons = wrapper.findAll('[data-aid="nodepool-options-btn"]');
        expect(optionsButtons.length).toBe(2); // One per tab
      });

      it("does not disable add button when disableListActions is false", () => {
        wrapper = mountComponent({ disableListActions: false });

        const addButton = wrapper.find('[data-aid="add-nodepool-btn"]');
        expect(addButton.attributes("disabled")).toBeUndefined();
      });
    });

    describe("disabledActionsTooltip", () => {
      it("uses add-specific default tooltip for add button when not provided", () => {
        wrapper = mountComponent({ disableListActions: true });

        const vm = getVM(wrapper);
        expect(vm.addButtonDisabledTooltip).toBe("To add node pools, go to the GPU quota step");
      });

      it("uses general default tooltip for menu options when not provided", () => {
        wrapper = mountComponent({ disableListActions: true });

        const vm = getVM(wrapper);
        const options = vm.getTabMenuOptions(1);

        const moveLeftOption = options.find((opt) => opt.value === "move-left");
        const removeOption = options.find((opt) => opt.value === "remove");

        expect(moveLeftOption?.toolTip).toBe("To perform this action, go to the GPU quota step");
        expect(removeOption?.toolTip).toBe("To perform this action, go to the GPU quota step");
      });

      it("uses custom tooltip when provided", () => {
        const customTooltip = "Custom disabled message";
        wrapper = mountComponent({
          disableListActions: true,
          disabledActionsTooltip: customTooltip,
        });

        const vm = getVM(wrapper);
        expect(vm.addButtonDisabledTooltip).toBe(customTooltip);
      });
    });
  });

  describe("Computed properties", () => {
    describe("showTabs", () => {
      it("returns true when availableNodePools > 1", () => {
        wrapper = mountComponent();

        const vm = getVM(wrapper);
        expect(vm.showTabs).toBe(true);
      });

      it("returns false when availableNodePools <= 1", () => {
        const singlePool: NodePoolOption[] = [{ id: "pool-1", name: "Pool 1" }];
        wrapper = mountComponent({
          nodePools: singlePool,
          availableNodePools: singlePool,
        });

        const vm = getVM(wrapper);
        expect(vm.showTabs).toBe(false);
      });
    });

    describe("hasRemainingPools", () => {
      it("returns true when there are pools not in nodePools", () => {
        wrapper = mountComponent();

        const vm = getVM(wrapper);
        expect(vm.hasRemainingPools).toBe(true);
      });

      it("returns false when all pools are in nodePools", () => {
        wrapper = mountComponent({
          nodePools: allAvailablePools,
          availableNodePools: allAvailablePools,
        });

        const vm = getVM(wrapper);
        expect(vm.hasRemainingPools).toBe(false);
      });
    });

    describe("remainingNodePools", () => {
      it("returns pools not in nodePools", () => {
        wrapper = mountComponent();

        const vm = getVM(wrapper);
        expect(vm.remainingNodePools).toEqual([{ id: "pool-3", name: "Pool 3" }]);
      });

      it("returns empty array when all pools are selected", () => {
        wrapper = mountComponent({
          nodePools: allAvailablePools,
          availableNodePools: allAvailablePools,
        });

        const vm = getVM(wrapper);
        expect(vm.remainingNodePools).toEqual([]);
      });
    });

    describe("addButtonDisabledTooltip", () => {
      it("returns add-specific tooltip when disableListActions is true and no custom tooltip", () => {
        wrapper = mountComponent({ disableListActions: true });

        const vm = getVM(wrapper);
        expect(vm.addButtonDisabledTooltip).toBe("To add node pools, go to the GPU quota step");
      });

      it("returns custom tooltip when disableListActions is true and custom tooltip provided", () => {
        const customTooltip = "Custom tooltip for disabled state";
        wrapper = mountComponent({
          disableListActions: true,
          disabledActionsTooltip: customTooltip,
        });

        const vm = getVM(wrapper);
        expect(vm.addButtonDisabledTooltip).toBe(customTooltip);
      });

      it("returns 'no more pools' tooltip when no remaining pools", () => {
        wrapper = mountComponent({
          nodePools: allAvailablePools,
          availableNodePools: allAvailablePools,
          disableListActions: false,
        });

        const vm = getVM(wrapper);
        expect(vm.addButtonDisabledTooltip).toBe("No more node pools to add");
      });

      it("returns undefined when add button is enabled", () => {
        wrapper = mountComponent({ disableListActions: false });

        const vm = getVM(wrapper);
        expect(vm.addButtonDisabledTooltip).toBeUndefined();
      });
    });
  });

  describe("Methods", () => {
    describe("getTabMenuOptions", () => {
      it("disables 'Move to the left' for first tab", () => {
        wrapper = mountComponent();

        const vm = getVM(wrapper);
        const options = vm.getTabMenuOptions(0);

        const moveLeftOption = options.find((opt) => opt.value === "move-left");
        expect(moveLeftOption?.disable).toBe(true);
      });

      it("enables 'Move to the left' for non-first tabs", () => {
        wrapper = mountComponent();

        const vm = getVM(wrapper);
        const options = vm.getTabMenuOptions(1);

        const moveLeftOption = options.find((opt) => opt.value === "move-left");
        expect(moveLeftOption?.disable).toBe(false);
      });

      it("disables 'Remove' when only one node pool", () => {
        const singlePool: NodePoolOption[] = [{ id: "pool-1", name: "Pool 1" }];
        wrapper = mountComponent({
          nodePools: singlePool,
          availableNodePools: allAvailablePools,
        });

        const vm = getVM(wrapper);
        const options = vm.getTabMenuOptions(0);

        const removeOption = options.find((opt) => opt.value === "remove");
        expect(removeOption?.disable).toBe(true);
      });

      it("enables 'Remove' when multiple node pools", () => {
        wrapper = mountComponent();

        const vm = getVM(wrapper);
        const options = vm.getTabMenuOptions(0);

        const removeOption = options.find((opt) => opt.value === "remove");
        expect(removeOption?.disable).toBe(false);
      });

      it("disables all options when disableListActions is true", () => {
        wrapper = mountComponent({ disableListActions: true });

        const vm = getVM(wrapper);
        const options = vm.getTabMenuOptions(1); // Non-first tab

        const moveLeftOption = options.find((opt) => opt.value === "move-left");
        const removeOption = options.find((opt) => opt.value === "remove");

        expect(moveLeftOption?.disable).toBe(true);
        expect(removeOption?.disable).toBe(true);
      });

      it("shows disabled tooltip when disableListActions is true", () => {
        const customTooltip = "Custom message";
        wrapper = mountComponent({
          disableListActions: true,
          disabledActionsTooltip: customTooltip,
        });

        const vm = getVM(wrapper);
        const options = vm.getTabMenuOptions(1);

        const moveLeftOption = options.find((opt) => opt.value === "move-left");
        const removeOption = options.find((opt) => opt.value === "remove");

        expect(moveLeftOption?.toolTip).toBe(customTooltip);
        expect(removeOption?.toolTip).toBe(customTooltip);
      });
    });

    describe("addNodePoolOptions", () => {
      it("returns options for remaining pools", () => {
        wrapper = mountComponent();

        const vm = getVM(wrapper);
        expect(vm.addNodePoolOptions).toEqual([{ label: "Pool 3", value: "pool-3", aid: "add-nodepool-pool-3-btn" }]);
      });
    });
  });

  describe("Events", () => {
    describe("update:node-pools", () => {
      it("emits pool names when handleMoveLeft is called", async () => {
        wrapper = mountComponent();

        const vm = getVM(wrapper);
        vm.handleMoveLeft("pool-2");

        await wrapper.vm.$nextTick();

        const emitted = wrapper.emitted("update:node-pools");
        expect(emitted).toBeTruthy();
        // After moving pool-2 to the left, order should be: pool-2, pool-1
        expect(emitted?.[0]).toEqual([["Pool 2", "Pool 1"]]);
      });

      it("does not emit when trying to move first pool to the left", async () => {
        wrapper = mountComponent();

        const vm = getVM(wrapper);
        vm.handleMoveLeft("pool-1");

        await wrapper.vm.$nextTick();

        expect(wrapper.emitted("update:node-pools")).toBeFalsy();
      });

      it("emits when handleRemove is called", async () => {
        wrapper = mountComponent();

        const vm = getVM(wrapper);
        vm.handleRemove("pool-2");

        await wrapper.vm.$nextTick();

        const emitted = wrapper.emitted("update:node-pools");
        expect(emitted).toBeTruthy();
        expect(emitted?.[0]).toEqual([["Pool 1"]]);
      });

      it("emits when pool is added via handleAddSelection and menu hides", async () => {
        wrapper = mountComponent();

        const vm = getVM(wrapper);

        // Simulate selecting a pool to add
        vm.handleAddSelection("pool-3");

        // Simulate menu hide (which triggers the actual emit)
        vm.handleAddMenuHide();

        await wrapper.vm.$nextTick();

        const emitted = wrapper.emitted("update:node-pools");
        expect(emitted).toBeTruthy();
        expect(emitted?.[0]).toEqual([["Pool 1", "Pool 2", "Pool 3"]]);
      });
    });
  });

  describe("Menu actions", () => {
    describe("handleMenuAction", () => {
      it("moves pool left when move-left action is triggered", async () => {
        wrapper = mountComponent();

        const vm = getVM(wrapper);
        vm.handleMenuAction("move-left", "pool-2");

        await wrapper.vm.$nextTick();

        // Verify the pool was moved by checking the emitted event
        const emitted = wrapper.emitted("update:node-pools");
        expect(emitted).toBeTruthy();
        expect(emitted?.[0]).toEqual([["Pool 2", "Pool 1"]]);
      });

      it("removes pool when remove action is triggered", async () => {
        wrapper = mountComponent();

        const vm = getVM(wrapper);
        vm.handleMenuAction("remove", "pool-2");

        await wrapper.vm.$nextTick();

        // Verify the pool was removed by checking the emitted event
        const emitted = wrapper.emitted("update:node-pools");
        expect(emitted).toBeTruthy();
        expect(emitted?.[0]).toEqual([["Pool 1"]]);
      });
    });

    describe("handleMoveLeft", () => {
      it("swaps pool with previous pool", async () => {
        wrapper = mountComponent();

        const vm = getVM(wrapper);
        vm.handleMoveLeft("pool-2");

        await wrapper.vm.$nextTick();

        const emitted = wrapper.emitted("update:node-pools");
        expect(emitted).toBeTruthy();
        // pool-2 should now be first
        expect(emitted?.[0]).toEqual([["Pool 2", "Pool 1"]]);
      });

      it("does nothing for first pool", async () => {
        wrapper = mountComponent();

        const vm = getVM(wrapper);
        vm.handleMoveLeft("pool-1");

        await wrapper.vm.$nextTick();

        expect(wrapper.emitted("update:node-pools")).toBeFalsy();
      });

      it("does nothing for non-existent pool", async () => {
        wrapper = mountComponent();

        const vm = getVM(wrapper);
        vm.handleMoveLeft("non-existent-pool");

        await wrapper.vm.$nextTick();

        expect(wrapper.emitted("update:node-pools")).toBeFalsy();
      });
    });

    describe("handleRemove", () => {
      it("removes pool from list", async () => {
        wrapper = mountComponent();

        const vm = getVM(wrapper);
        vm.handleRemove("pool-1");

        await wrapper.vm.$nextTick();

        const emitted = wrapper.emitted("update:node-pools");
        expect(emitted).toBeTruthy();
        expect(emitted?.[0]).toEqual([["Pool 2"]]);
      });
    });

    describe("handleAddSelection and handleAddMenuHide", () => {
      it("stores pending update and emits on menu hide", async () => {
        wrapper = mountComponent();

        const vm = getVM(wrapper);

        // Select a pool to add (stores pending)
        vm.handleAddSelection("pool-3");
        expect(vm.pendingAddUpdate).toBeTruthy();

        // Menu hides - should emit
        vm.handleAddMenuHide();

        await wrapper.vm.$nextTick();

        const emitted = wrapper.emitted("update:node-pools");
        expect(emitted).toBeTruthy();
        expect(emitted?.[0]).toEqual([["Pool 1", "Pool 2", "Pool 3"]]);
        expect(vm.pendingAddUpdate).toBeNull();
      });

      it("does not emit if pool not found", async () => {
        wrapper = mountComponent();

        const vm = getVM(wrapper);
        vm.handleAddSelection("non-existent-pool");

        expect(vm.pendingAddUpdate).toBeNull();
      });

      it("clears openMenuId on menu hide", async () => {
        wrapper = mountComponent();

        const vm = getVM(wrapper);
        vm.openMenuId = "some-id";
        vm.handleAddSelection("pool-3");
        vm.handleAddMenuHide();

        expect(vm.openMenuId).toBeNull();
      });
    });
  });

  describe("Edge cases", () => {
    it("handles empty nodePools array", () => {
      wrapper = mountComponent({
        nodePools: [],
        availableNodePools: allAvailablePools,
      });

      const tabs = wrapper.findAll(".q-tab");
      expect(tabs.length).toBe(0);
    });

    it("handles empty availableNodePools array", () => {
      wrapper = mountComponent({
        nodePools: [],
        availableNodePools: [],
      });

      // Tabs should be hidden (showTabs = false)
      const tabsContainer = wrapper.find(".node-pool-tabs-container");
      expect(tabsContainer.exists()).toBe(false);
    });
  });
});
