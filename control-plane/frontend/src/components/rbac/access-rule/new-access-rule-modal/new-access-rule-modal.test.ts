import { beforeEach, describe, expect, it, vi } from "vitest";
import { Quasar } from "quasar";
import { shallowMount, type VueWrapper } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import NewAccessRuleModal from "./new-access-rule-modal.vue";
import { subjectTypeOptionsWithSSO } from "@/models/access-rule.model";
import { SubjectTypeExtended } from "@/swagger-models/authorization-client";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";

vi.mock("@/services/control-plane/rbac/access-rule.service/access-rule.service", () => ({
  accessRuleService: {
    accessRulesBatchCreate: vi.fn().mockResolvedValue({ succeeded: [], failed: [] }),
    createAuthenticatedUsersAccessRule: vi.fn().mockResolvedValue({ id: 1 }),
  },
}));

const optionFor = (value: SubjectTypeExtended) => subjectTypeOptionsWithSSO.find((o) => o.value === value);

const mountModal = (): VueWrapper<InstanceType<typeof NewAccessRuleModal>> =>
  shallowMount(NewAccessRuleModal, {
    global: {
      plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
      directives: { permission: vi.fn(), clickOutside: vi.fn() },
    },
    props: { isSso: true, roles: [] },
  });

describe("NewAccessRuleModal - all-authenticated-users", () => {
  beforeEach(() => vi.clearAllMocks());

  it("onSave calls createAuthenticatedUsersAccessRule for all-authenticated-users", async () => {
    const wrapper = mountModal();
    wrapper.vm.selectedSubjectType = optionFor(SubjectTypeExtended.AllAuthenticatedUsers) ?? null;
    wrapper.vm.selectedRole = { label: "Viewer", value: 5 };
    wrapper.vm.selectedScopes = [{ id: "1001", type: "tenant" }];
    vi.spyOn(wrapper.vm, "validate").mockResolvedValue(true);

    await wrapper.vm.onSave();

    expect(accessRuleService.createAuthenticatedUsersAccessRule).toHaveBeenCalledWith({
      roleId: 5,
      scopeId: "1001",
      scopeType: "tenant",
    });
    expect(accessRuleService.accessRulesBatchCreate).not.toHaveBeenCalled();
  });

  it("isAllAuthenticatedUsers reflects the selected subject type", async () => {
    const wrapper = mountModal();
    wrapper.vm.selectedSubjectType = optionFor(SubjectTypeExtended.AllAuthenticatedUsers) ?? null;
    await wrapper.vm.$nextTick();
    expect(wrapper.vm.isAllAuthenticatedUsers).toBe(true);

    wrapper.vm.selectedSubjectType = optionFor(SubjectTypeExtended.User) ?? null;
    await wrapper.vm.$nextTick();
    expect(wrapper.vm.isAllAuthenticatedUsers).toBe(false);
  });

  it("builds a subject-less role+scopes payload for all-authenticated-users", () => {
    const wrapper = mountModal();
    wrapper.vm.selectedSubjectType = optionFor(SubjectTypeExtended.AllAuthenticatedUsers) ?? null;
    wrapper.vm.selectedRole = { label: "Viewer", value: 5 };
    wrapper.vm.selectedScopes = [{ id: "1001", type: "tenant" }];

    const fields = wrapper.vm.prepareAuthenticatedUsersCreateFields();

    expect(fields).toEqual({ roleId: 5, scopeId: "1001", scopeType: "tenant" });
  });

  it("keeps only the most recent scope on selection for all-authenticated-users", () => {
    const wrapper = mountModal();
    wrapper.vm.selectedSubjectType = optionFor(SubjectTypeExtended.AllAuthenticatedUsers) ?? null;
    wrapper.vm.onScopesSelected([
      { id: "1", type: "tenant" },
      { id: "2", type: "project" },
    ]);
    expect(wrapper.vm.selectedScopes).toEqual([{ id: "2", type: "project" }]);
  });

  it("clears a multi-scope selection when switching to all-authenticated-users", async () => {
    const wrapper = mountModal();
    wrapper.vm.selectedScopes = [
      { id: "1", type: "tenant" },
      { id: "2", type: "project" },
    ];
    wrapper.vm.selectedSubjectType = optionFor(SubjectTypeExtended.AllAuthenticatedUsers) ?? null;
    await wrapper.vm.$nextTick();
    expect(wrapper.vm.selectedScopes).toEqual([]);
  });

  it("rejects a system scope for all-authenticated-users", () => {
    const wrapper = mountModal();
    wrapper.vm.selectedSubjectType = optionFor(SubjectTypeExtended.AllAuthenticatedUsers) ?? null;
    wrapper.vm.selectedRole = { label: "Viewer", value: 5 };
    wrapper.vm.selectedScopes = [{ id: "system", type: "system" }];
    expect(() => wrapper.vm.prepareAuthenticatedUsersCreateFields()).toThrow();
  });

  it("sends the selected subjectId(s) for a regular subject type", () => {
    const wrapper = mountModal();
    wrapper.vm.selectedSubjectType = optionFor(SubjectTypeExtended.User) ?? null;
    wrapper.vm.selectedRole = { label: "Viewer", value: 5 };
    wrapper.vm.selectedScopes = [{ id: "1001", type: "tenant" }];
    wrapper.vm.selectedSubjectIds = [{ label: "alice@run.ai", value: "alice@run.ai" as SubjectTypeExtended }];

    const fields = wrapper.vm.prepareAccessRulesCreateFields();

    expect(fields.payload.subjects).toEqual([{ subjectId: "alice@run.ai", subjectType: SubjectTypeExtended.User }]);
  });
});
