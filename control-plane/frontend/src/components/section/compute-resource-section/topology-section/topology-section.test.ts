/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { QInput, QSelect, Quasar } from "quasar";
import { TopologySection } from ".";
import { PodAffinityType } from "@/swagger-models/assets-service-client";
import type { Component } from "vue";

describe("topology-section.vue", () => {
  function factory(props = {}): VueWrapper<InstanceType<typeof TopologySection>> {
    return mount(TopologySection, {
      props: {
        ...props,
      },
      global: {
        plugins: [Quasar],
        components: { QInput, QSelect },
      },
    });
  }

  describe("Html elements check", () => {
    it("Component exists", () => {
      const wrapper = factory();
      expect(wrapper.exists()).toBe(true);
    });

    it("Component has the correct text", () => {
      const wrapper = factory();
      const titleSpan = wrapper.find("span[aid='topology-section-title']");
      expect(titleSpan.text()).toBe(
        "Set topology (pod affinity) to let the workload be scheduled on nodes with a matching topology",
      );
    });

    it("Component has info tooltip", () => {
      const wrapper = factory();
      const infoTooltip = wrapper.find("[aid='topology-title-tooltip']");
      expect(infoTooltip.exists()).toBe(true);
    });
  });

  describe("Props", () => {
    describe("podAffinity", () => {
      it("add topology button is visible when there is no podAffinity", () => {
        const wrapper = factory();
        const addTopologyButton = wrapper.find("[aid='add-topology-btn']");
        expect(addTopologyButton.exists()).toBe(true);
      });

      it("add topology button is not visible when there is podAffinity", () => {
        const wrapper = factory({ podAffinity: { key: "", type: PodAffinityType.Required } });
        const addTopologyButton = wrapper.find("[aid='add-topology-btn']");
        expect(addTopologyButton.exists()).toBe(false);
      });

      it("topology operator select gets the podAffinity type", () => {
        const wrapper = factory({ podAffinity: { key: "", type: PodAffinityType.Preferred } });
        const operatorSelect: Component = wrapper.findComponent(QSelect);
        expect(operatorSelect.props("modelValue")).toBe(PodAffinityType.Preferred);
      });

      it("topology key input gets the podAffinity key", () => {
        const wrapper = factory({ podAffinity: { key: "key", type: PodAffinityType.Preferred } });
        const keyInput: Component = wrapper.findComponent(QInput);
        expect(keyInput.props("modelValue")).toBe("key");
      });
    });

    describe("readOnly", () => {
      it("add topology button is disabled when readOnly prop is true", () => {
        const wrapper = factory({ readOnly: true });
        const addTopologyButton: Component = wrapper.findComponent("[aid='add-topology-btn']");
        expect(addTopologyButton.props("disable")).toBe(true);
      });

      it("add topology button is enabled when readOnly prop is false", () => {
        const wrapper = factory({ readOnly: false });
        const addTopologyButton: Component = wrapper.findComponent("[aid='add-topology-btn']");
        expect(addTopologyButton.props("disable")).toBe(false);
      });

      it("remove topology button is disabled when readOnly prop is true", () => {
        const wrapper = factory({ readOnly: true, podAffinity: { key: "key", type: PodAffinityType.Required } });
        const removeTopologyButton: Component = wrapper.findComponent("[aid='remove-topology-btn']");
        expect(removeTopologyButton.props("disable")).toBe(true);
      });

      it("remove topology button is enabled when readOnly prop is false", () => {
        const wrapper = factory({ readOnly: false, podAffinity: { key: "key", type: PodAffinityType.Required } });
        const removeTopologyButton: Component = wrapper.findComponent("[aid='remove-topology-btn']");
        expect(removeTopologyButton.props("disable")).toBe(false);
      });

      it("topology operator select is disabled when readOnly prop is true", () => {
        const wrapper = factory({ readOnly: true, podAffinity: { key: "key", type: PodAffinityType.Required } });
        const operatorSelect: Component = wrapper.findComponent(QSelect);
        expect(operatorSelect.props("disable")).toBe(true);
      });

      it("topology operator select is enabled when readOnly prop is false", () => {
        const wrapper = factory({ readOnly: false, podAffinity: { key: "key", type: PodAffinityType.Required } });
        const operatorSelect: Component = wrapper.findComponent(QSelect);
        expect(operatorSelect.props("disable")).toBe(false);
      });

      it("topology key input is disabled when readOnly prop is true", () => {
        const wrapper = factory({ readOnly: true, podAffinity: { key: "key", type: PodAffinityType.Required } });
        const keyInput: Component = wrapper.findComponent(QInput);
        expect(keyInput.props("disable")).toBe(true);
      });

      it("topology key input is enabled when readOnly prop is false", () => {
        const wrapper = factory({ readOnly: false, podAffinity: { key: "key", type: PodAffinityType.Required } });
        const keyInput: Component = wrapper.findComponent(QInput);
        expect(keyInput.props("disable")).toBe(false);
      });
    });
  });

  describe("Emits", () => {
    it("emits update-topology when add topology button is clicked", async () => {
      const wrapper = factory();
      await wrapper.find("[aid='add-topology-btn']").trigger("click");
      const defaultPodAffinity = { key: "", type: PodAffinityType.Required };
      expect(wrapper.emitted("update-topology")).toEqual([[defaultPodAffinity]]);
    });

    it("emits update-topology when topology key is updated", async () => {
      const wrapper = factory({ podAffinity: { key: "", type: PodAffinityType.Required } });
      await wrapper.findComponent(QInput).setValue("key");
      const expectedPodAffinity = { key: "key", type: PodAffinityType.Required };
      expect(wrapper.emitted("update-topology")).toEqual([[expectedPodAffinity]]);
    });

    it("emits update-topology when topology operator is updated", async () => {
      const wrapper = factory({ podAffinity: { key: "", type: PodAffinityType.Required } });
      await wrapper.findComponent(QSelect).setValue(PodAffinityType.Preferred);
      const expectedPodAffinity = { key: "", type: PodAffinityType.Preferred };
      expect(wrapper.emitted("update-topology")).toEqual([[expectedPodAffinity]]);
    });

    it("emits update-topology when remove topology button is clicked", async () => {
      const wrapper = factory({ podAffinity: { key: "key", type: PodAffinityType.Required } });
      await wrapper.find("[aid='remove-topology-btn']").trigger("click");
      expect(wrapper.emitted("update-topology")).toEqual([[null]]);
    });
  });
});
