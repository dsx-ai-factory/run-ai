import { beforeEach, describe, expect, it, vi } from "vitest";
import { Quasar } from "quasar";
import { shallowMount, type VueWrapper } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import NewManagementAccessRuleForm from "./new-management-access-rule-form.vue";
import { EAccessRuleModalPage, subjectTypeOptionsWithSSO } from "@/models/access-rule.model";
import { SubjectTypeExtended } from "@/swagger-models/authorization-client";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";

vi.mock("@/services/control-plane/rbac/access-rule.service/access-rule.service", () => ({
  accessRuleService: {
    accessRulesBatchCreate: vi.fn().mockResolvedValue({ succeeded: [], failed: [] }),
    createAuthenticatedUsersAccessRule: vi.fn().mockResolvedValue({ id: 1 }),
  },
}));

const optionFor = (value: SubjectTypeExtended) => subjectTypeOptionsWithSSO.find((o) => o.value === value);

// Scope-anchored page: scope is fixed via options, the admin picks the subject type.
const mountForm = (): VueWrapper<InstanceType<typeof NewManagementAccessRuleForm>> =>
  shallowMount(NewManagementAccessRuleForm, {
    global: {
      plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
      directives: { permission: vi.fn(), clickOutside: vi.fn() },
    },
    props: {
      displayConfig: { subject: true, users: true, roles: true, scope: true },
      isSso: true,
      options: { scopeId: "1001", scopeType: "tenant", page: EAccessRuleModalPage.Project },
    },
  });

describe("NewManagementAccessRuleForm - all-authenticated-users", () => {
  beforeEach(() => vi.clearAllMocks());

  it("onSave calls createAuthenticatedUsersAccessRule for all-authenticated-users", async () => {
    const wrapper = mountForm();
    wrapper.vm.selectedSubjectType = optionFor(SubjectTypeExtended.AllAuthenticatedUsers) ?? null;
    wrapper.vm.selectedRole = { label: "Viewer", value: 5 };
    vi.spyOn(wrapper.vm, "validate").mockResolvedValue(true);

    await wrapper.vm.onSave();

    expect(accessRuleService.createAuthenticatedUsersAccessRule).toHaveBeenCalledWith({
      roleId: 5,
      scopeId: "1001",
      scopeType: "tenant",
    });
    expect(accessRuleService.accessRulesBatchCreate).not.toHaveBeenCalled();
  });

  it("builds a subject-less role+scopes payload for all-authenticated-users", () => {
    const wrapper = mountForm();
    wrapper.vm.selectedSubjectType = optionFor(SubjectTypeExtended.AllAuthenticatedUsers) ?? null;
    wrapper.vm.selectedRole = { label: "Viewer", value: 5 };

    const fields = wrapper.vm.prepareAuthenticatedUsersCreateFields();

    expect(fields).toEqual({ roleId: 5, scopeId: "1001", scopeType: "tenant" });
  });

  it("sends the selected subjectId(s) for a regular subject type", () => {
    const wrapper = mountForm();
    wrapper.vm.selectedSubjectType = optionFor(SubjectTypeExtended.User) ?? null;
    wrapper.vm.selectedRole = { label: "Viewer", value: 5 };
    wrapper.vm.selectedSubjectIds = [{ label: "alice@run.ai", value: "alice@run.ai" as SubjectTypeExtended }];

    const fields = wrapper.vm.prepareAccessRulesCreateFields();

    expect(fields.payload.subjects).toEqual([{ subjectId: "alice@run.ai", subjectType: SubjectTypeExtended.User }]);
  });
});
