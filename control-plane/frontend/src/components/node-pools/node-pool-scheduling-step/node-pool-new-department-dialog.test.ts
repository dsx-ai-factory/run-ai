/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount, flushPromises, type VueWrapper } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";
import { Quasar } from "quasar";
import Dialog from "./node-pool-new-department-dialog.vue";
import DepartmentNameSection from "@/components/department/department-edit-form/department-name-section/department-name-section.vue";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
vi.mock("@/services/control-plane/org-unit.service/org-unit.service", () => ({
  orgUnitService: {
    isDepartmentNameUnique: vi.fn().mockResolvedValue(true),
    createDepartment: vi.fn(),
  },
}));
let wrapper: VueWrapper;
afterEach(() => wrapper?.unmount());
async function render() {
  wrapper = mount(Dialog, {
    props: { clusterId: "cluster", clusterName: "Tokyo" },
    global: { plugins: [Quasar], mocks: { $route: { query: {} } } },
  });
  await flushPromises();
  return wrapper.findComponent(DepartmentNameSection).get("input");
}
describe("new department preview boundary", () => {
  it("shows explicit nonpersisted feedback after a valid Create without writing a department", async () => {
    const name = await render();
    await name.setValue("department-preview");
    await new Promise((resolve) => setTimeout(resolve, 350));
    (document.querySelector('[aid="create-department-preview"]') as HTMLButtonElement).click();
    await flushPromises();
    expect(orgUnitService.isDepartmentNameUnique).toHaveBeenCalled();
    expect(document.querySelector('.footer-left [role="status"]')?.textContent).toContain(
      "Preview only. This department has not been created.",
    );
    expect(orgUnitService.createDepartment).not.toHaveBeenCalled();
  });
  it("focuses the editable Name input when the dialog opens", async () => {
    const name = await render();
    await new Promise((resolve) => setTimeout(resolve, 350));
    expect(document.activeElement).toBe(name.element);
  });
});
