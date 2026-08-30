import { AccessRulePhase, SubjectType, ScopeType } from "@/swagger-models/authorization-client";
import type { IAccessRuleTable } from "@/table-models/access-rule.table-model";
import type { DisplayedCluster } from "@/swagger-models/cluster-service-client";

export const accessRulesMock: IAccessRuleTable[] = [
  {
    id: 1,
    subjectId: "user1",
    subjectType: SubjectType.User,
    roleId: 1,
    scopeId: "scope1",
    scopeType: ScopeType.Cluster,
    roleName: "Admin",
    scopeName: "Scope A",
    createdAt: "2023-01-01T00:00:00Z",
    updatedAt: "2023-01-02T00:00:00Z",
    tenantId: 1,
    createdBy: "creator1",
    status: {
      phase: AccessRulePhase.NotReady,
      details: "Cluster ID: 12345",
    },
  },
  {
    id: 2,
    subjectId: "user2",
    subjectType: SubjectType.User,
    roleId: 2,
    scopeId: "scope2",
    scopeType: ScopeType.Cluster,
    roleName: "User",
    scopeName: "Scope B",
    createdAt: "2023-01-01T00:00:00Z",
    updatedAt: "2023-01-02T00:00:00Z",
    tenantId: 2,
    createdBy: "creator2",
    status: {
      phase: AccessRulePhase.Ready,
    },
  },
];

export const clustersMock: DisplayedCluster[] = [
  { uuid: "12345", name: "Cluster A", tenantId: 1, createdAt: "2023-01-01T00:00:00Z" },
  { uuid: "67890", name: "Cluster B", tenantId: 2, createdAt: "2023-01-01T00:00:00Z" },
  { uuid: "67892", name: "Cluster B", tenantId: 2, createdAt: "2023-01-01T00:00:00Z" },
  { uuid: "13215", name: "Cluster C", tenantId: 2, createdAt: "2023-01-01T00:00:00Z" },
];
