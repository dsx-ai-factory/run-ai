import { type AccessRule, ScopeType, SubjectType } from "@/swagger-models/authorization-client";
import type { IAccessRuleTableModalOptions } from "@/models/access-rule.model";
import { accessRulesMiniTableColumns } from "@/table-models/access-rule.table-model";

export const mockAccessRules: AccessRule[] = [
  {
    id: 1,
    roleName: "Admin",
    roleId: 1,
    scopeType: ScopeType.Project,
    scopeId: "project-1",
    scopeName: "Test Project",
    subjectType: SubjectType.User,
    subjectId: "user-1",
    subjectName: "John Doe",
    createdAt: "2023-01-01T00:00:00Z",
    updatedAt: "2023-01-01T00:00:00Z",
    tenantId: 1,
    createdBy: "admin",
  } as AccessRule,
];

export const mockModalOptionsServerSide: IAccessRuleTableModalOptions = {
  header: "Test Access Rules",
  columns: accessRulesMiniTableColumns,
  scopeType: ScopeType.Project,
  scopeId: "project-1",
  loading: false,
};

export const mockModalOptionsWithSubject: IAccessRuleTableModalOptions = {
  header: "Test Access Rules for User",
  columns: accessRulesMiniTableColumns,
  subjectType: SubjectType.User,
  subjectId: "user-123",
  loading: false,
};
export const mockModalOptionsBackwardCompatibility: IAccessRuleTableModalOptions = {
  header: "Test Access Rules (Backward Compatibility)",
  columns: accessRulesMiniTableColumns,
  accessRules: mockAccessRules,
  loading: false,
};
