import { describe, expect, it } from "vitest";
import { type Permission, ResourceType, ResourceTypeGroupId } from "@/swagger-models/authorization-client";
import { accessRuleUtil } from "@/utils/rbac.util/access-rule.util/access-rule.util";
import { AccessRulePhase, ScopeType } from "@/swagger-models/authorization-client";
import { clustersMock } from "./access-rule.util.mock";
import type { IAccessRuleTable } from "@/table-models/access-rule.table-model";
import DOCS from "@/constants/docs";

describe("accessRuleUtil", () => {
  describe("sortAndGroupPermissionsByGroupId", () => {
    it("should sort permissions by resourceType", () => {
      const permissions = [
        { resourceType: ResourceType.Users, groupId: ResourceTypeGroupId.Iam, displayName: "Users" },
        { resourceType: ResourceType.Tenant, groupId: ResourceTypeGroupId.Organization, displayName: "Account" },
      ];
      const result: Permission[] = accessRuleUtil.sortAndGroupPermissionsByGroupId(permissions as Permission[]);
      expect(result[0].resourceType).toBe(ResourceType.Tenant);
      expect(result[1].resourceType).toBe(ResourceType.Users);
    });

    it("should group permissions by groupId", () => {
      const permissions = [
        { resourceType: ResourceType.Users, groupId: ResourceTypeGroupId.Iam, displayName: "Users" },
        { resourceType: ResourceType.Roles, groupId: ResourceTypeGroupId.Iam, displayName: "Roles" },
        { resourceType: ResourceType.Department, groupId: ResourceTypeGroupId.Organization, displayName: "Departments" },
      ];
      const result: Permission[] = accessRuleUtil.sortAndGroupPermissionsByGroupId(permissions as Permission[]);
      expect(result[0].groupId).toBe(ResourceTypeGroupId.Organization);
      expect(result[1].groupId).toBe(ResourceTypeGroupId.Iam);
      expect(result[2].groupId).toBe(ResourceTypeGroupId.Iam);
    });

    it("should order groups according to ResourceTypeGroupIdOrder", () => {
      const permissions = [
        { resourceType: ResourceType.Users, groupId: ResourceTypeGroupId.Iam, displayName: "Users" },
        { resourceType: ResourceType.Department, groupId: ResourceTypeGroupId.Organization, displayName: "Departments" },
      ];
      const result: Permission[] = accessRuleUtil.sortAndGroupPermissionsByGroupId(permissions as Permission[]);
      expect(result[0].groupId).toBe(ResourceTypeGroupId.Organization);
      expect(result[1].groupId).toBe(ResourceTypeGroupId.Iam);
    });
  });

  describe("addStatusExplanationToAccessRules", () => {
    function createAccessRule(overrides = {}) {
      return {
        scopeType: ScopeType.Cluster,
        status: {
          phase: AccessRulePhase.NotReady,
          details: "Cluster ID: 12345",
        },
        ...overrides,
      };
    }

    const testCases = [
      {
        description: "should add explanation to access rules with NotReady status and details",
        accessRules: [createAccessRule()],
        expectedExplanation: "Issues were found while enforcing the access rule <br> in the cluster.",
      },
      {
        description: "should not add explanation if status phase is not NotReady",
        accessRules: [createAccessRule({ status: { phase: AccessRulePhase.Ready } })],
        expectedExplanation: undefined,
      },
      {
        description: "should handle access rules without status details",
        accessRules: [createAccessRule({ status: { phase: AccessRulePhase.Ready, details: undefined } })],
        expectedExplanation: undefined,
      },
      {
        description: "should handle access rules without status",
        accessRules: [createAccessRule({ status: undefined })],
        expectedExplanation: undefined,
      },
      {
        description: "should handle access rules with missing status details",
        accessRules: [createAccessRule({ status: { phase: AccessRulePhase.NotReady, details: undefined } })],
        expectedExplanation: undefined,
      },
      {
        description: "should add explanation for multiple clusters",
        accessRules: [
          createAccessRule({ status: { details: "NotReady: [12345, 67890]", phase: AccessRulePhase.NotReady } }),
        ],
        expectedExplanation:
          "Issues were found while enforcing the access rule in cluster(s) <br> Cluster A, Cluster B.",
      },
      {
        description: "should add explanation for a single cluster",
        accessRules: [createAccessRule()],
        expectedExplanation: "Issues were found while enforcing the access rule <br> in the cluster.",
      },
      {
        description: "should add explanation for multi cluster issues found",
        accessRules: [
          createAccessRule({ status: { details: "NotReady: [12345, 67890, 67892]", phase: AccessRulePhase.NotReady } }),
        ],
        expectedExplanation:
          "Issues were found while enforcing the access rule in cluster(s) <br> Cluster A, Cluster B, Cluster B.",
      },
      {
        description: "should add explanation for multi cluster issues found with + more indication",
        accessRules: [
          createAccessRule({
            status: { details: "NotReady: [12345, 67890, 67892, 13215]", phase: AccessRulePhase.NotReady },
          }),
        ],
        expectedExplanation:
          "Issues were found while enforcing the access rule in cluster(s) <br> Cluster A, Cluster B, Cluster B + 1 more.",
      },
    ];
    it.each(testCases)("$description", ({ accessRules, expectedExplanation }) => {
      const result = accessRuleUtil.addStatusExplanationToAccessRules(accessRules as IAccessRuleTable[], clustersMock);
      const explanation = result[0].status?.explanation;
      expect(explanation).toBe(expectedExplanation);
    });

    it("adds a clickable support CTA when the support URL is configured", () => {
      const support = DOCS.SUPPORT as { URL: string };
      const originalUrl = support.URL;
      support.URL = "https://support.example.test/contact";

      try {
        const result = accessRuleUtil.addStatusExplanationToAccessRules(
          [createAccessRule()] as IAccessRuleTable[],
          clustersMock,
        );
        expect(result[0].status?.explanation).toContain(
          '<a href="https://support.example.test/contact" target="_blank" rel="noopener noreferrer">',
        );
      } finally {
        support.URL = originalUrl;
      }
    });
  });
});
