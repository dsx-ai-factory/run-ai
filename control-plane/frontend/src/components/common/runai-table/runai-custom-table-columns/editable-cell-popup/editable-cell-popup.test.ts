/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount } from "@vue/test-utils";
import { QBtn, QMenu, Quasar } from "quasar";
import { describe, expect, it, vi } from "vitest";
import EditableCellPopup from "./editable-cell-popup.vue";
import OrgUnitEditableWeightCol from "../org-unit-editable-weight-col.vue";
import OrgUnitGpuEditableQuotaCol from "../org-unit-gpu-editable-quota-col.vue";
import { createTestingPinia } from "@pinia/testing";
import { ECustomCell } from "@/models/table.model";
import { EQuotaEntity } from "@/models/resource.model";

describe("editable cell popup", () => {
  it("preserves the org-unit weight cell's active appearance while invalid input disables Apply", async () => {
    const wrapper = mount(OrgUnitEditableWeightCol, {
      props: {
        row: { id: "a", orgUnitId: "a", weight: 2 },
        column: { name: "weight", label: "Weight", field: "weight" },
      },
      global: { plugins: [Quasar], stubs: { QMenu: { template: "<div><slot /></div>" } } },
    });
    await wrapper.get(".cell-trigger").trigger("click");
    await wrapper.get("input").setValue("-2");
    expect(wrapper.findComponent(EditableCellPopup).props("state")).toBe("error");
    expect(wrapper.classes()).toContain("cell--active");
    expect(wrapper.classes()).not.toContain("cell--error");
    const apply = wrapper.findAllComponents(QBtn).filter((button) => button.props("label") === "APPLY")[0];
    expect(apply.attributes("disabled")).toBeDefined();
    wrapper.unmount();
  });

  it.each([
    { quota: -2, state: "error", applyDisabled: true },
    { quota: 20, state: "warning", applyDisabled: false },
  ])(
    "preserves the org-unit quota cell's active appearance for $state validation",
    async ({ quota, state, applyDisabled }) => {
      const wrapper = mount(OrgUnitGpuEditableQuotaCol, {
        props: {
          row: { id: "a", name: "a", gpuQuota: 5, quotaDistribution: 5, allocatedNonPreemptibleGpu: 0 },
          column: {
            name: "quota",
            label: "GPUs",
            field: "gpuQuota",
            customCellConfig: {
              [ECustomCell.ORG_UNIT_GPU_EDITABLE_QUOTA_COL]: {
                availableQuota: 10,
                totalQuota: 100,
                nodePoolName: "default",
                nodePoolId: "pool",
                entityType: EQuotaEntity.project,
                onSave: vi.fn(),
              },
            },
          },
        },
        global: {
          plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
          stubs: { QMenu: { template: "<div><slot /></div>" } },
        },
      });
      await wrapper.get(".cell-trigger").trigger("click");
      await wrapper.get("input").setValue(String(quota));
      expect(wrapper.findComponent(EditableCellPopup).props("state")).toBe(state);
      expect(wrapper.classes()).toContain("cell--active");
      expect(wrapper.classes()).not.toContain(`cell--${state}`);
      const apply = wrapper.findAllComponents(QBtn).filter((button) => button.props("label") === "APPLY")[0];
      expect(apply.attributes("disabled") !== undefined).toBe(applyDisabled);
      wrapper.unmount();
    },
  );
  it("preserves the existing opened-after-show lifecycle by default", async () => {
    const wrapper = mount(EditableCellPopup, { props: { title: "Weight" }, global: { plugins: [Quasar] } });
    const menu = wrapper.findComponent(QMenu);
    await wrapper.get(".cell-trigger").trigger("click");
    expect(wrapper.emitted("opened")).toBeUndefined();
    menu.vm.$emit("show");
    expect(wrapper.emitted("opened")).toHaveLength(1);
    wrapper.unmount();
  });
  it("exposes before-show separately without changing the opened lifecycle", async () => {
    const wrapper = mount(EditableCellPopup, {
      props: { title: "Weight" },
      global: { plugins: [Quasar] },
    });
    await wrapper.get(".cell-trigger").trigger("click");
    expect(wrapper.emitted("before-show")).toHaveLength(1);
    expect(wrapper.emitted("opened")).toBeUndefined();
    wrapper.findComponent(QMenu).vm.$emit("show");
    expect(wrapper.emitted("opened")).toHaveLength(1);
    wrapper.unmount();
  });
  it("disables Apply during invalid or loading input and enables it when submission becomes valid", async () => {
    const wrapper = mount(EditableCellPopup, {
      props: { title: "Weight", isSubmitDisabled: true },
      global: {
        plugins: [Quasar],
        stubs: { QMenu: { template: "<div><slot /></div>" } },
      },
    });
    const apply = wrapper.findAllComponents(QBtn).filter((button) => button.props("label") === "APPLY")[0];
    expect(apply.attributes("disabled")).toBeDefined();
    await apply.trigger("click");
    expect(wrapper.emitted("apply")).toBeUndefined();

    await wrapper.setProps({ isSubmitDisabled: false });
    expect(apply.attributes("disabled")).toBeUndefined();
    await apply.trigger("click");
    expect(wrapper.emitted("apply")).toHaveLength(1);
    wrapper.unmount();
  });
});
