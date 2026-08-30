//@ts-nocheck
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import { controlPlaneService } from "@/services/control-plane/control-plane.service/control-plane.service";
import { serviceAccountsService } from "@/services/control-plane/service-accounts.service/service-accounts.service";
import {
  appsMock,
  accessRulesMock,
  usersMock,
} from "@/services/control-plane/rbac/access-rule.service/access-rule.service.mock";
import type { AccessRuleCreationFields } from "@/swagger-models/authorization-client";
import { ScopeType, SubjectType } from "@/swagger-models/authorization-client";

vi.mock("@/services/control-plane/control-plane.service/control-plane.service");
vi.mock("@/services/control-plane/service-accounts.service/service-accounts.service");
describe("accessRuleService", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    setActivePinia(createPinia());
  });

  it("Delete assignment", async () => {
    controlPlaneService.get = vi.fn();
    await accessRuleService.deleteAccessRule(5);
    expect(controlPlaneService.delete).toBeCalledWith("api/v1/authorization/access-rules/5");
  });

  it("New assignment", async () => {
    controlPlaneService.post = vi.fn();
    const assignmentCreationRequest: AccessRuleCreationFields = {
      roleId: 2,
      scopeId: 1,
      scopeType: ScopeType.Cluster,
      subjectId: "123",
      subjectType: "subjectType",
    };
    await accessRuleService.createAccessRule(assignmentCreationRequest);
    expect(controlPlaneService.post).toBeCalledWith("api/v1/authorization/access-rules", assignmentCreationRequest);
  });

  it("Get subject ids for user type", async () => {
    controlPlaneService.get.mockResolvedValue(usersMock);
    const accessRules = await accessRuleService.getSubjectIds(false, SubjectType.User, "user2");
    expect(controlPlaneService.get).toBeCalledWith("api/v1/users");
    expect(accessRules.length).toBe(2);
  });

  it("Get subject ids for service account type", async () => {
    serviceAccountsService.list = vi.fn().mockResolvedValue(appsMock);
    const accessRules = await accessRuleService.getSubjectIds(false, SubjectType.ServiceAccount, "app1");
    expect(serviceAccountsService.list).toBeCalled();
    expect(accessRules.length).toBe(1);
  });

  it("Get subject ids for service account type, sso and check uniqueness", async () => {
    serviceAccountsService.list = vi.fn().mockResolvedValue(appsMock);
    controlPlaneService.get.mockResolvedValueOnce(accessRulesMock);
    const accessRules = await accessRuleService.getSubjectIds(true, SubjectType.ServiceAccount, "app1");
    expect(controlPlaneService.get).toBeCalledWith(
      "api/v1/authorization/access-rules?subjectType=service-account&subjectIdFilter=app1",
    );
    expect(serviceAccountsService.list).toBeCalled();
    expect(accessRules.length).toBe(3);
  });

  it("Get subject ids for group type, sso", async () => {
    controlPlaneService.get.mockResolvedValueOnce(accessRulesMock);
    const accessRules = await accessRuleService.getSubjectIds(true, SubjectType.Group, "group1");
    expect(controlPlaneService.get).toBeCalledWith(
      "api/v1/authorization/access-rules?subjectType=group&subjectIdFilter=group1",
    );
    expect(accessRules.length).toBe(2);
  });

  it("Get subject ids for user type, sso", async () => {
    controlPlaneService.get.mockResolvedValueOnce(usersMock).mockResolvedValueOnce(accessRulesMock);
    const accessRules = await accessRuleService.getSubjectIds(true, SubjectType.User, "user1");
    expect(controlPlaneService.get).toBeCalledWith(
      "api/v1/authorization/access-rules?subjectType=user&subjectIdFilter=user1",
    );
    expect(controlPlaneService.get).toBeCalledWith("api/v1/users");
    expect(accessRules.length).toBe(3);
  });
});
