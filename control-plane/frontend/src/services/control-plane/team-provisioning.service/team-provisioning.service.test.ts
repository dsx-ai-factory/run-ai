import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { teamProvisioningService, AI_PRACTITIONER_ROLE_NAME, SYSTEM_ADMIN_ROLE_NAME } from "./team-provisioning.service";
import { userService } from "@/services/control-plane/user.service/user.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import { rolesService } from "@/services/control-plane/rbac/roles.service/roles.service";
import { TeamIdentificationMethod } from "@/components/onboarding/models/admin-onboarding.model";

vi.mock("@/services/control-plane/user.service/user.service");
vi.mock("@/services/control-plane/org-unit.service/org-unit.service");
vi.mock("@/services/control-plane/rbac/access-rule.service/access-rule.service");
vi.mock("@/services/control-plane/rbac/roles.service/roles.service");

describe("teamProvisioningService", () => {
  const mockClusterId = "cluster-123";
  const mockProjectId = "project-456";
  const mockAIPractitionerRoleId = 10;
  const mockSystemAdminRoleId = 1;

  const mockRoles = [
    { id: mockSystemAdminRoleId, name: SYSTEM_ADMIN_ROLE_NAME },
    { id: mockAIPractitionerRoleId, name: AI_PRACTITIONER_ROLE_NAME },
    { id: 3, name: "Other role" },
  ];

  beforeEach(() => {
    vi.resetAllMocks();
    setActivePinia(createPinia());

    // Default mock implementations
    rolesService.list = vi.fn().mockResolvedValue(mockRoles);
    orgUnitService.createProject = vi.fn().mockResolvedValue({ id: mockProjectId });
    accessRuleService.accessRulesBatchCreate = vi.fn().mockResolvedValue({
      succeeded: [{ id: 1 }],
      failed: [],
    });
    userService.createBatch = vi.fn().mockResolvedValue({
      results: [],
    });
  });

  describe("AI_PRACTITIONER_ROLE_NAME constant", () => {
    it("should have the correct value", () => {
      expect(AI_PRACTITIONER_ROLE_NAME).toBe("AI practitioner");
    });
  });

  describe("provisionTeam", () => {
    it("should use AI practitioner role when creating access rules with emails", async () => {
      const teamData = {
        teamName: "Test Team",
        memberEmails: ["user1@test.com", "user2@test.com"],
        identificationMethod: TeamIdentificationMethod.MemberEmails,
        letSomeoneElseFill: false,
        delegateEmail: "",
        groupName: "",
      };

      await teamProvisioningService.provisionTeam(mockClusterId, teamData, false);

      // Verify roles were fetched
      expect(rolesService.list).toHaveBeenCalled();

      // Verify access rules were created with AI practitioner role
      expect(accessRuleService.accessRulesBatchCreate).toHaveBeenCalledWith(
        expect.objectContaining({
          payload: expect.objectContaining({
            roleId: mockAIPractitionerRoleId,
          }),
        }),
      );
    });

    it("should use AI practitioner role when creating access rules with SSO group", async () => {
      const teamData = {
        teamName: "Test Team",
        memberEmails: [],
        identificationMethod: TeamIdentificationMethod.GroupName,
        letSomeoneElseFill: false,
        delegateEmail: "",
        groupName: "test-sso-group",
      };

      await teamProvisioningService.provisionTeam(mockClusterId, teamData, true);

      // Verify access rules were created with AI practitioner role
      expect(accessRuleService.accessRulesBatchCreate).toHaveBeenCalledWith(
        expect.objectContaining({
          payload: expect.objectContaining({
            roleId: mockAIPractitionerRoleId,
          }),
        }),
      );
    });

    it("should fallback to any practitioner role if exact AI practitioner role not found", async () => {
      const fallbackPractitionerRoleId = 99;
      const rolesWithoutExactMatch = [
        { id: mockSystemAdminRoleId, name: SYSTEM_ADMIN_ROLE_NAME },
        { id: fallbackPractitionerRoleId, name: "Senior AI practitioner" }, // Contains "practitioner"
        { id: 3, name: "Other role" },
      ];
      rolesService.list = vi.fn().mockResolvedValue(rolesWithoutExactMatch);

      const teamData = {
        teamName: "Test Team",
        memberEmails: ["user1@test.com"],
        identificationMethod: TeamIdentificationMethod.MemberEmails,
        letSomeoneElseFill: false,
        delegateEmail: "",
        groupName: "",
      };

      await teamProvisioningService.provisionTeam(mockClusterId, teamData, false);

      // Verify access rules were created with fallback practitioner role
      expect(accessRuleService.accessRulesBatchCreate).toHaveBeenCalledWith(
        expect.objectContaining({
          payload: expect.objectContaining({
            roleId: fallbackPractitionerRoleId,
          }),
        }),
      );
    });
  });

  describe("createAdminUser", () => {
    it("should use system admin role when creating admin user", async () => {
      const adminEmail = "admin@test.com";
      const tenantId = 123;

      await teamProvisioningService.createAdminUser(adminEmail, tenantId, false);

      // Verify roles were fetched
      expect(rolesService.list).toHaveBeenCalled();

      // Verify access rules were created with system admin role
      expect(accessRuleService.accessRulesBatchCreate).toHaveBeenCalledWith(
        expect.objectContaining({
          payload: expect.objectContaining({
            roleId: mockSystemAdminRoleId,
          }),
        }),
      );
    });
  });
});
