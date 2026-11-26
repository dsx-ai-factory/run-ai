/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount, flushPromises } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia } from "pinia";
import { Quasar } from "quasar";
import { RunaiSelect } from "@/components/common/runai-select";
import { SubjectAddableMultiSelect } from "@/components/rbac/common/subject-addable-multi-select";
import { ScopeType, SubjectTypeExtended } from "@/swagger-models/authorization-client";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import OrgUnitAdministratorsFields from "./org-unit-administrators-fields.vue";

vi.mock("@/composables/node-pool/use-access-rule-roles.composable", async () => {
  const { ref } = await import("vue");
  return {
    useAccessRuleRoles: () => ({
      roleOptions: ref([
        { label: "Viewer", value: 3 },
        { label: "Department administrator", value: 7 },
      ]),
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
vi.mock("@/services/control-plane/user.service/user.service", () => ({
  default: { getUsersMinimal: vi.fn().mockResolvedValue([]) },
}));

const SCOPE = { scopeId: "dep-1", scopeType: ScopeType.Department, clusterId: "cluster-1" };

const mountFields = () =>
  mount(OrgUnitAdministratorsFields, {
    attachTo: document.body,
    props: { title: "Set the department administrator(s)", tooltip: "Help", defaultRole: "Department administrator" },
    global: { plugins: [Quasar, createPinia()] },
  });

let wrapper: ReturnType<typeof mountFields>;

const addAdministrator = async () => {
  await wrapper.get('[aid="add-administrator"]').trigger("click");
  await flushPromises();
};

const selectSubjects = async (...ids: string[]) => {
  wrapper.findComponent(SubjectAddableMultiSelect).vm.$emit(
    "update:selected-subjects",
    ids.map((id) => ({ label: id, value: id })),
  );
  await flushPromises();
};

describe("OrgUnitAdministratorsFields", () => {
  beforeEach(() => {
    vi.mocked(accessRuleService.createAccessRule)
      .mockReset()
      .mockResolvedValue({} as never);
    vi.mocked(accessRuleService.createAuthenticatedUsersAccessRule)
      .mockReset()
      .mockResolvedValue({} as never);
    wrapper = mountFields();
  });

  afterEach(() => wrapper.unmount());

  it("starts without an administrator, which is valid and saves nothing", async () => {
    expect(wrapper.find(".administrator-rule").exists()).toBe(false);
    expect(wrapper.vm.validate()).toEqual({ valid: true });
    await wrapper.vm.save(SCOPE);
    expect(accessRuleService.createAccessRule).not.toHaveBeenCalled();
  });

  it("opens a rule for users with the default role selected", async () => {
    await addAdministrator();

    const [subject, role] = wrapper.findAllComponents(RunaiSelect);
    expect(subject.props("modelValue")).toMatchObject({ value: SubjectTypeExtended.User });
    expect(role.props("modelValue")).toMatchObject({ label: "Department administrator", value: 7 });
  });

  it("asks for at least one subject", async () => {
    await addAdministrator();

    expect(wrapper.vm.validate()).toEqual({ valid: false, message: "Select at least one subject." });
  });

  it("reports how many administrators are set", async () => {
    await addAdministrator();
    await selectSubjects("ada@run.ai", "grace@run.ai");

    expect(wrapper.emitted("update:count")?.at(-1)).toEqual([2]);

    await wrapper.get('[aid="remove-administrator"]').trigger("click");
    expect(wrapper.emitted("update:count")?.at(-1)).toEqual([0]);
  });

  it("creates one access rule per subject and skips the saved ones on retry", async () => {
    await addAdministrator();
    await selectSubjects("ada@run.ai", "grace@run.ai");
    vi.mocked(accessRuleService.createAccessRule)
      .mockResolvedValueOnce({} as never)
      .mockRejectedValueOnce(new Error("failed"));

    await expect(wrapper.vm.save(SCOPE)).rejects.toThrow("Couldn't add grace@run.ai as administrator");
    await wrapper.vm.save(SCOPE);

    const subjectIds = vi.mocked(accessRuleService.createAccessRule).mock.calls.map(([rule]) => rule.subjectId);
    expect(subjectIds).toEqual(["ada@run.ai", "grace@run.ai", "grace@run.ai"]);
    expect(vi.mocked(accessRuleService.createAccessRule).mock.calls[0][0]).toMatchObject({
      ...SCOPE,
      roleId: 7,
      subjectType: SubjectTypeExtended.User,
    });
  });

  it("keeps adding the other subjects when one fails", async () => {
    await addAdministrator();
    await selectSubjects("ada@run.ai", "grace@run.ai", "alan@run.ai");
    vi.mocked(accessRuleService.createAccessRule)
      .mockResolvedValueOnce({} as never)
      .mockRejectedValueOnce(new Error("failed"))
      .mockResolvedValueOnce({} as never);

    await expect(wrapper.vm.save(SCOPE)).rejects.toThrow("grace@run.ai");

    const subjectIds = vi.mocked(accessRuleService.createAccessRule).mock.calls.map(([rule]) => rule.subjectId);
    expect(subjectIds).toEqual(["ada@run.ai", "grace@run.ai", "alan@run.ai"]);
  });

  it("creates a single rule for all authenticated users", async () => {
    await addAdministrator();
    wrapper
      .findAllComponents(RunaiSelect)[0]
      .vm.$emit("update:modelValue", { label: "All", value: SubjectTypeExtended.AllAuthenticatedUsers });
    await flushPromises();

    expect(wrapper.findComponent(SubjectAddableMultiSelect).exists()).toBe(false);
    expect(wrapper.emitted("update:count")?.at(-1)).toEqual([1]);
    await wrapper.vm.save(SCOPE);
    expect(accessRuleService.createAuthenticatedUsersAccessRule).toHaveBeenCalledWith({ ...SCOPE, roleId: 7 });
  });
});
