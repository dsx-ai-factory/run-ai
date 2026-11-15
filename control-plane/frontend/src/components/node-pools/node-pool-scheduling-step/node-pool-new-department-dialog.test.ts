/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount, flushPromises, type VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia } from "pinia";
import { Quasar, Notify } from "quasar";
import Dialog from "./node-pool-new-department-dialog.vue";
import DepartmentNameSection from "@/components/department/department-edit-form/department-name-section/department-name-section.vue";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
vi.mock("@/services/control-plane/org-unit.service/org-unit.service", () => ({
  orgUnitService: {
    isDepartmentNameUnique: vi.fn().mockResolvedValue(true),
    createDepartment: vi.fn(),
  },
}));
vi.mock("@/stores/permissions.store", () => ({ usePermissionStore: () => ({ hasPermission: () => true }) }));
vi.mock("@/composables/access-rule/use-access-rule-roles.composable", async () => {
  const { ref } = await import("vue");
  return {
    useAccessRuleRoles: () => ({
      roleOptions: [{ label: "Department administrator", value: 7 }],
      canAssignRoles: ref(true),
      loadRoles: vi.fn(),
      validateRole: (role: unknown) => !!role,
      loading: ref(false),
    }),
  };
});
vi.mock("@/services/control-plane/rbac/access-rule.service/access-rule.service", () => ({
  accessRuleService: {
    createAccessRule: vi.fn(),
    createAuthenticatedUsersAccessRule: vi.fn(),
    getSubjects: vi.fn().mockResolvedValue([]),
  },
}));
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import { RunaiSelect } from "@/components/common/runai-select";
import { SubjectAddableMultiSelect } from "@/components/rbac/common/subject-addable-multi-select";
vi.mock("@/services/control-plane/user.service/user.service", () => ({
  default: { getUsersMinimal: vi.fn().mockResolvedValue([]) },
  userService: { getUsersMinimal: vi.fn().mockResolvedValue([]) },
}));
let wrapper: VueWrapper;
beforeEach(() => {
  vi.mocked(orgUnitService.createDepartment)
    .mockReset()
    .mockResolvedValue({ id: "new", name: "department-preview", clusterId: "cluster" } as never);
  vi.mocked(accessRuleService.createAccessRule)
    .mockReset()
    .mockResolvedValue({} as never);
});
afterEach(() => wrapper?.unmount());
async function render() {
  wrapper = mount(Dialog, {
    props: { clusterId: "cluster", clusterName: "Tokyo" },
    global: { plugins: [[Quasar, { plugins: { Notify } }], createPinia()], mocks: { $route: { query: {} } } },
  });
  await flushPromises();
  return wrapper.findComponent(DepartmentNameSection).get("input");
}
describe("new department creation", () => {
  it("creates a real department and emits it for table refresh", async () => {
    const name = await render();
    await name.setValue("department-preview");
    await new Promise((resolve) => setTimeout(resolve, 350));
    (document.querySelector('[aid="create-department"]') as HTMLButtonElement).click();
    await flushPromises();
    expect(orgUnitService.createDepartment).toHaveBeenCalledWith({
      name: "department-preview",
      description: "",
      clusterId: "cluster",
      resources: [],
    });
    expect(wrapper.emitted("created")?.[0][0]).toMatchObject({ id: "new" });
    expect(wrapper.emitted("close")).toHaveLength(1);
  });
  it("preserves the form on creation failure", async () => {
    vi.mocked(orgUnitService.createDepartment).mockRejectedValue(new Error("Creation failed"));
    const name = await render();
    await name.setValue("department-preview");
    await new Promise((resolve) => setTimeout(resolve, 350));
    (document.querySelector('[aid="create-department"]') as HTMLButtonElement).click();
    await flushPromises();
    expect(wrapper.emitted("close")).toBeUndefined();
    expect((name.element as HTMLInputElement).value).toBe("department-preview");
    expect(document.querySelector(".department-save-error")?.textContent).toContain("Creation failed");
  });
  it("retries only incomplete role assignments without creating another department", async () => {
    const name = await render();
    await name.setValue("department-preview");
    await new Promise((resolve) => setTimeout(resolve, 350));
    (document.querySelector('[aid="add-department-administrator"]') as HTMLButtonElement).click();
    await flushPromises();
    const selects = wrapper.findAllComponents(RunaiSelect);
    selects[0].vm.$emit("update:modelValue", { label: "User", value: "user" });
    await flushPromises();
    wrapper.findComponent(SubjectAddableMultiSelect).vm.$emit("update:selected-subjects", [
      { label: "a@test.com", value: "a@test.com" },
      { label: "b@test.com", value: "b@test.com" },
    ]);
    selects[1].vm.$emit("update:modelValue", { label: "Department administrator", value: 7 });
    vi.mocked(accessRuleService.createAccessRule)
      .mockResolvedValueOnce({} as never)
      .mockRejectedValueOnce(new Error("Role failed"));
    await flushPromises();
    (document.querySelector('[aid="create-department"]') as HTMLButtonElement).click();
    await flushPromises();
    expect(wrapper.emitted("close")).toBeUndefined();
    expect(document.querySelector(".department-save-error")?.textContent).toContain("Department created");
    (document.querySelector('[aid="create-department"]') as HTMLButtonElement).click();
    await flushPromises();
    expect(orgUnitService.createDepartment).toHaveBeenCalledTimes(1);
    expect(accessRuleService.createAccessRule).toHaveBeenCalledTimes(3);
    expect(vi.mocked(accessRuleService.createAccessRule).mock.calls[2][0]).toMatchObject({
      subjectId: "b@test.com",
      roleId: 7,
      scopeId: "new",
      scopeType: "department",
    });
    expect(wrapper.emitted("close")).toHaveLength(1);
  });
  it("does not revalidate the name while editing description", async () => {
    const name = await render();
    await name.setValue("department-preview");
    await new Promise((resolve) => setTimeout(resolve, 400));
    await flushPromises();
    vi.mocked(orgUnitService.isDepartmentNameUnique).mockClear();
    wrapper.findComponent(DepartmentNameSection).vm.$emit("update:description", "Description draft");
    await flushPromises();
    await new Promise((resolve) => setTimeout(resolve, 400));
    expect(orgUnitService.isDepartmentNameUnique).not.toHaveBeenCalled();
  });
  it("focuses the editable Name input when the dialog opens", async () => {
    const name = await render();
    await new Promise((resolve) => setTimeout(resolve, 350));
    expect(document.activeElement).toBe(name.element);
  });
});
