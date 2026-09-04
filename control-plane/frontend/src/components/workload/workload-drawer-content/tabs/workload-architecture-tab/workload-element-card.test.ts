/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";

import WorkloadElementCard from "./workload-element-card.vue";
import { EArchitectureNodeKind, type NodeResourceInfo } from "./workload-architecture-diagram.util";

const mountCard = (props: Partial<InstanceType<typeof WorkloadElementCard>["$props"]> = {}) =>
  mount(WorkloadElementCard, {
    global: {
      plugins: [Quasar],
      stubs: { Handle: true },
    },
    props: {
      name: "test-element",
      kind: EArchitectureNodeKind.Element,
      ...props,
    },
  });

describe("workload-element-card.vue", () => {
  describe("rendering", () => {
    it("should render the element name", () => {
      const wrapper = mountCard({ name: "leader-0" });
      expect(wrapper.find(".card-name").text()).toBe("leader-0");
    });

    it("should render the workload icon only for the workload root", () => {
      const workloadCard = mountCard({ kind: EArchitectureNodeKind.Workload });
      expect(workloadCard.find(".card-icon-box").exists()).toBe(true);

      const elementCard = mountCard({ kind: EArchitectureNodeKind.Element });
      expect(elementCard.find(".card-icon-box").exists()).toBe(false);
    });

    it("should render allocation rows when resources are provided", () => {
      const resources: NodeResourceInfo = {
        cpuComputeAllocation: 0.5,
        cpuMemoryAllocation: "512M",
        gpuComputeAllocation: 1,
        gpuMemoryAllocation: "12582M",
      };
      const wrapper = mountCard({ resources });
      const text = wrapper.text();

      expect(text).toContain("CPU compute allocation");
      expect(text).toContain("CPU memory allocation");
      expect(text).toContain("GPU compute allocation");
      expect(text).toContain("GPU memory allocation");
    });

    it("should not render the details section when there is nothing to show", () => {
      const wrapper = mountCard({
        kind: EArchitectureNodeKind.Element,
        elementKind: null,
        kubernetesName: null,
        resources: null,
        podsDisplay: null,
        showExpandAction: false,
      });
      expect(wrapper.find(".card-details").exists()).toBe(false);
    });

    it("should render kubernetes name only for non-workload nodes", () => {
      const elementCard = mountCard({
        kind: EArchitectureNodeKind.Element,
        kubernetesName: "rs-name",
      });
      expect(elementCard.text()).toContain("Name: rs-name");

      const workloadCard = mountCard({
        kind: EArchitectureNodeKind.Workload,
        kubernetesName: "rs-name",
      });
      expect(workloadCard.text()).not.toContain("Name: rs-name");
    });
  });

  describe("selection", () => {
    it("should add 'selected' class when selected prop is true", () => {
      const wrapper = mountCard({ selected: true, selectable: true });
      expect(wrapper.find(".element-card.selected").exists()).toBe(true);
    });

    it("should add 'selectable' class when selectable prop is true", () => {
      const wrapper = mountCard({ selectable: true });
      expect(wrapper.find(".element-card.selectable").exists()).toBe(true);
    });

    it("should not apply selected/selectable classes by default", () => {
      const wrapper = mountCard();
      const card = wrapper.find(".element-card");
      expect(card.classes()).not.toContain("selected");
      expect(card.classes()).not.toContain("selectable");
    });
  });

  describe("interaction", () => {
    it("should emit 'click' when clicked while selectable", async () => {
      const wrapper = mountCard({ selectable: true });
      await wrapper.find(".element-card").trigger("click");

      expect(wrapper.emitted("click")).toHaveLength(1);
    });

    it("should NOT emit 'click' when clicked while not selectable", async () => {
      const wrapper = mountCard({ selectable: false });
      await wrapper.find(".element-card").trigger("click");

      expect(wrapper.emitted("click")).toBeUndefined();
    });

    it("should emit 'click' on Enter and Space when selectable", async () => {
      const wrapper = mountCard({ selectable: true });
      const card = wrapper.find(".element-card");

      await card.trigger("keydown.enter");
      await card.trigger("keydown.space");

      expect(wrapper.emitted("click")).toHaveLength(2);
    });

    it("should emit 'expand-click' from the expand button without bubbling to card click", async () => {
      const wrapper = mountCard({ selectable: true, showExpandAction: true });

      const expandBtn = wrapper.find(".expand-action-btn");
      expect(expandBtn.exists()).toBe(true);

      await expandBtn.trigger("click");

      expect(wrapper.emitted("expand-click")).toHaveLength(1);
      expect(wrapper.emitted("click")).toBeUndefined();
    });

    it("should not render the expand button when showExpandAction is false", () => {
      const wrapper = mountCard({ showExpandAction: false });
      expect(wrapper.find(".expand-action-btn").exists()).toBe(false);
    });
  });

  describe("accessibility", () => {
    it("should expose role='button', tabindex='0' and aria-label when selectable", () => {
      const wrapper = mountCard({ selectable: true, name: "redis-master" });
      const card = wrapper.find(".element-card");

      expect(card.attributes("role")).toBe("button");
      expect(card.attributes("tabindex")).toBe("0");
      expect(card.attributes("aria-label")).toBe("Select redis-master");
    });

    it("should not set role/tabindex/aria-label when not selectable", () => {
      const wrapper = mountCard({ selectable: false });
      const card = wrapper.find(".element-card");

      expect(card.attributes("role")).toBeUndefined();
      expect(card.attributes("tabindex")).toBeUndefined();
      expect(card.attributes("aria-label")).toBeUndefined();
    });

    it("should set aria-label on expand button to 'Maximize view'", () => {
      const wrapper = mountCard({ showExpandAction: true });
      const expandBtn = wrapper.find(".expand-action-btn");
      expect(expandBtn.attributes("aria-label")).toBe("Maximize view");
    });
  });
});
