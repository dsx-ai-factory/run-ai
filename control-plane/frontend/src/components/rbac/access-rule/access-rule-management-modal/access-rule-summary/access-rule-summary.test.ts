import { describe, expect, it } from "vitest";
import { Quasar } from "quasar";
import { mount, type VueWrapper } from "@vue/test-utils";
import AccessRuleSummary from "./access-rule-summary.vue";
import { EAccessRuleModalPage } from "@/models/access-rule.model";
import { SubjectTypeExtended, type AccessRule } from "@/swagger-models/authorization-client";

const baseRule = {
  roleName: "Viewer",
  scopeType: "tenant",
  scopeName: "acme",
  subjectId: "all-authenticated-users",
  subjectType: SubjectTypeExtended.AllAuthenticatedUsers,
} as AccessRule;

const mountSummary = (props: Record<string, unknown>): VueWrapper =>
  mount(AccessRuleSummary, {
    global: { plugins: [Quasar] },
    props: { page: EAccessRuleModalPage.User, accessRule: baseRule, ...props },
  });

describe("access-rule-summary", () => {
  // Subject page (User / Application)
  it("renders role, scope and subject for an all-authenticated-users rule", () => {
    const text = mountSummary({}).text();
    expect(text).toContain("Role: Viewer");
    expect(text).toContain("Subject: All authenticated users and service accounts");
    // the storage sentinel must never be surfaced
    expect(text).not.toContain("all-authenticated-users");
  });

  it("renders the subject as 'Group - <name>' for a group rule", () => {
    const text = mountSummary({
      accessRule: { ...baseRule, subjectType: SubjectTypeExtended.Group, subjectId: "dev-team" },
    }).text();
    expect(text).toContain("Subject: Group - dev-team");
  });

  it("renders the subject as 'Service account - <name>' for a service account rule", () => {
    const text = mountSummary({
      accessRule: { ...baseRule, subjectType: SubjectTypeExtended.ServiceAccount, subjectId: "my-sa" },
    }).text();
    expect(text).toContain("Subject: Service account - my-sa");
  });

  it("renders the subject id for a user rule", () => {
    const text = mountSummary({
      accessRule: { ...baseRule, subjectType: SubjectTypeExtended.User, subjectId: "alice@run.ai" },
    }).text();
    expect(text).toContain("Subject: alice@run.ai");
  });

  // Scope page (Project / Department)
  it("renders subject and role on a scope page for an all-authenticated-users rule", () => {
    const text = mountSummary({ page: EAccessRuleModalPage.Project, accessRule: baseRule }).text();
    expect(text).toContain("Subject: All authenticated users and service accounts");
    expect(text).toContain("Role: Viewer");
    expect(text).not.toContain("all-authenticated-users");
  });

  it("renders the subject id on a scope page for a user rule", () => {
    const text = mountSummary({
      page: EAccessRuleModalPage.Project,
      accessRule: { ...baseRule, subjectType: SubjectTypeExtended.User, subjectId: "alice@run.ai" },
    }).text();
    expect(text).toContain("Subject: alice@run.ai");
  });
});
