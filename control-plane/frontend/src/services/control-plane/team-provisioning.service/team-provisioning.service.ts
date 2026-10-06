/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { userService } from "@/services/control-plane/user.service/user.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import { rolesService } from "@/services/control-plane/rbac/roles.service/roles.service";
import type { TeamData } from "@/components/onboarding/models/admin-onboarding.model";
import {
  IOnboardingCompletionState,
  TeamIdentificationMethod,
} from "@/components/onboarding/models/admin-onboarding.model";
import type {
  UserCreationRequest,
  UserPostResponse,
  BulkUserCreationRequest,
  BulkUserCreationResponse,
} from "@/swagger-models/identity-manager-client";
import type { ProjectCreationRequest, Project, ResourcesNullable } from "@/swagger-models/org-unit-service-client";
import {
  type AccessRulesBatchCreateFields,
  type AccessRulesBatchCreateResponseFailedInner,
  SubjectType,
  ScopeType,
  type Role,
} from "@/swagger-models/authorization-client";
import { ErrorAlert } from "@/utils/error-alert.util";

// Constants
export const AI_PRACTITIONER_ROLE_NAME = "AI practitioner";
export const SYSTEM_ADMIN_ROLE_NAME = "System administrator";
const MAX_GPU_QUOTA = 8;
const GPU_PER_USER = 1;

// Types
export interface TeamProvisioningResult {
  success: boolean;
  completionState: IOnboardingCompletionState;
  failedSubjectIds: string[];
  createdProjectId?: string;
  createdUsers?: UserPostResponse[];
  error?: string;
}

export interface AdminUserCreationResult {
  success: boolean;
  createdUserId?: string;
  createdUser?: UserPostResponse;
  error?: string;
}

interface UserCreationResults {
  createdUsers: UserPostResponse[];
  failedUsers: string[];
}

interface AccessRuleCreationResults {
  createdAccessRules: number;
  failedSubjectIds: string[];
}

// Service export
export const teamProvisioningService = {
  provisionTeam,
  createAdminUser,
};

/**
 * Main orchestrator for team provisioning
 */
async function provisionTeam(
  clusterId: string,
  teamData: TeamData,
  isSsoConfigured: boolean,
): Promise<TeamProvisioningResult> {
  const context = {
    clusterId,
    teamData,
    failedSubjectIds: [] as string[],
    createdProjectId: undefined as string | undefined,
    createdUsers: [] as UserPostResponse[],
  };

  // Step 1: Create Project (critical - stop if fails)
  const projectResult = await createProjectStep(context);
  if (!projectResult.success) {
    return projectResult;
  }

  // Step 2: Create Access Rules (handles both email and group methods)
  await createAccessRulesStep(context);

  // Step 3: Create Users (only for email-based provisioning)
  if (teamData.identificationMethod === TeamIdentificationMethod.MemberEmails && !isSsoConfigured) {
    await createUsersStep(context);
  }

  // Determine final state and return result
  return buildFinalResult(context);
}

/**
 * Create admin user with system administrator privileges (SSO-aware)
 * If SSO is enabled, only creates access rules; otherwise creates both access rules and local user
 */
async function createAdminUser(
  email: string,
  tenantId: number,
  isSsoConfigured = false,
): Promise<AdminUserCreationResult> {
  try {
    // Step 1: Get system admin role
    const systemAdminRoleId = await getSystemAdminRoleId();

    // Step 2: Create access rule with system admin privileges using email as subjectId
    const accessRuleRequest: AccessRulesBatchCreateFields = {
      payload: {
        roleId: systemAdminRoleId,
        subjects: [
          {
            subjectId: email,
            subjectType: SubjectType.User,
          },
        ],
        scopes: [
          {
            scopeId: tenantId.toString(),
            scopeType: ScopeType.Tenant,
          },
        ],
      },
    };

    await accessRuleService.accessRulesBatchCreate(accessRuleRequest);

    // Step 3: Create local user only if SSO is not enabled
    let createdUserId: string | undefined;
    let createdUser: UserPostResponse | undefined;
    if (!isSsoConfigured) {
      const userRequest: UserCreationRequest = { email: email, notify: true };
      createdUser = await userService.create(userRequest);
      createdUserId = createdUser.username || createdUser.id;
    }

    return {
      success: true,
      createdUserId,
      createdUser,
    };
  } catch (error) {
    new ErrorAlert({
      generalMessage: isSsoConfigured ? "Failed to create admin access rules" : "Failed to create admin user",
    });
    return {
      success: false,
      error:
        error instanceof Error ? error.message : `Failed to create admin ${isSsoConfigured ? "access rules" : "user"}`,
    };
  }
}

/**
 * Step 1: Create project for the team
 */
async function createProjectStep(context: {
  teamData: TeamData;
  clusterId: string;
  createdProjectId: string | undefined;
  createdUsers: UserPostResponse[];
}): Promise<TeamProvisioningResult | { success: true }> {
  try {
    // Calculate member count based on identification method
    const memberCount =
      context.teamData.identificationMethod === TeamIdentificationMethod.GroupName
        ? MAX_GPU_QUOTA // Use max quota for SSO groups since member count is unknown
        : context.teamData.memberEmails.length;
    const project = await createTeamProject(context.teamData.teamName, memberCount, context.clusterId);

    context.createdProjectId = project.id;
    return { success: true };
  } catch (error) {
    return {
      success: false,
      completionState: IOnboardingCompletionState.ProjectCreationFailed,
      failedSubjectIds: [],
      error: error instanceof Error ? error.message : "Project creation failed",
    };
  }
}

/**
 * Step 2: Create access rules (handles both email and group identification methods)
 */
async function createAccessRulesStep(context: {
  clusterId: string;
  createdProjectId: string | undefined;
  failedSubjectIds: string[];
  teamData: TeamData;
  createdUsers: UserPostResponse[];
}): Promise<void> {
  if (!context.createdProjectId) {
    return;
  }

  const { teamData } = context;

  // Validate inputs based on identification method
  if (teamData.identificationMethod === TeamIdentificationMethod.GroupName) {
    if (!teamData.groupName) return;
  } else {
    if (teamData.memberEmails.length === 0) return;
  }

  try {
    let accessRuleResults: AccessRuleCreationResults;

    if (teamData.identificationMethod === TeamIdentificationMethod.GroupName) {
      // Create access rules for SSO group
      accessRuleResults = await createTeamAccessRulesWithGroup(
        teamData.groupName,
        context.createdProjectId,
        context.clusterId,
      );
    } else {
      // Create access rules for individual emails
      accessRuleResults = await createTeamAccessRulesWithEmails(
        teamData.memberEmails,
        context.createdProjectId,
        context.clusterId,
      );
    }

    if (accessRuleResults.failedSubjectIds.length > 0) {
      context.failedSubjectIds.push(...accessRuleResults.failedSubjectIds);
    }
  } catch (error) {
    // If entire batch fails, mark all subjects as failed based on identification method
    if (teamData.identificationMethod === TeamIdentificationMethod.GroupName) {
      context.failedSubjectIds.push(teamData.groupName);
    } else {
      context.failedSubjectIds.push(...teamData.memberEmails);
    }
  }
}

/**
 * Step 3: Create users for team members
 */
async function createUsersStep(context: {
  teamData: TeamData;
  failedSubjectIds: string[];
  createdUsers: UserPostResponse[];
}): Promise<void> {
  try {
    const userResults = await createTeamUsers(context.teamData.memberEmails);
    if (userResults.failedUsers.length > 0) {
      context.failedSubjectIds.push(...userResults.failedUsers);
    }
    context.createdUsers.push(...userResults.createdUsers);
  } catch (error) {
    context.failedSubjectIds.push(...context.teamData.memberEmails);
  }
}

/**
 * Build the final provisioning result
 */
function buildFinalResult(context: {
  teamData: TeamData;
  failedSubjectIds: string[];
  createdProjectId: string | undefined;
  createdUsers: UserPostResponse[];
}): TeamProvisioningResult {
  const uniqueFailedIds = [...new Set(context.failedSubjectIds)];

  let completionState: IOnboardingCompletionState;

  if (uniqueFailedIds.length === 0) {
    completionState = IOnboardingCompletionState.Completed;
  } else {
    // For group-based identification, if the group fails, it's a complete failure
    // For email-based identification, check if all users failed
    if (context.teamData.identificationMethod === TeamIdentificationMethod.GroupName) {
      completionState = IOnboardingCompletionState.AccessRulesCreationFailed;
    } else {
      const totalUsers = context.teamData.memberEmails.length;
      if (uniqueFailedIds.length === totalUsers) {
        completionState = IOnboardingCompletionState.AccessRulesCreationFailed;
      } else {
        completionState = IOnboardingCompletionState.AccessRulesCreationFailed;
      }
    }
  }

  return {
    success: true,
    completionState,
    failedSubjectIds: uniqueFailedIds,
    createdProjectId: context.createdProjectId,
    createdUsers: context.createdUsers,
  };
}

/**
 * Create a project for the team
 */
async function createTeamProject(teamName: string, memberCount: number, clusterId: string): Promise<Project> {
  const gpuQuota = Math.min(memberCount * GPU_PER_USER, MAX_GPU_QUOTA);

  const resources: ResourcesNullable[] = [
    {
      gpu: {
        deserved: gpuQuota,
        limit: gpuQuota,
      },
    },
  ];

  const projectRequest: ProjectCreationRequest = {
    name: teamName,
    clusterId: clusterId,
    resources: resources,
  };

  return await orgUnitService.createProject(projectRequest);
}

async function createTeamUsers(memberEmails: string[]): Promise<UserCreationResults> {
  const bulkUserCreationRequest: BulkUserCreationRequest = {
    users: memberEmails.map((email) => ({ email })),
    notify: true,
  };

  try {
    const bulkResponse: BulkUserCreationResponse = await userService.createBatch(bulkUserCreationRequest);

    const createdUsers: UserPostResponse[] = [];
    const failedUsers: string[] = [];

    bulkResponse.results.forEach((result, index) => {
      const email = memberEmails[index];
      if (result.success && result.user) {
        createdUsers.push(result.user);
      } else {
        failedUsers.push(email);
      }
    });
    return { createdUsers, failedUsers };
  } catch (error) {
    return {
      createdUsers: [],
      failedUsers: [...memberEmails],
    };
  }
}

/**
 * Create access rules for team members using email addresses as subjectId (before users exist)
 */
async function createTeamAccessRulesWithEmails(
  memberEmails: string[],
  projectId: string,
  clusterId: string,
): Promise<AccessRuleCreationResults> {
  try {
    const aiPractitionerRoleId = await getAIPractitionerRoleId();

    const batchRequest: AccessRulesBatchCreateFields = {
      payload: {
        roleId: aiPractitionerRoleId,
        clusterId: clusterId,
        subjects: memberEmails.map((email) => ({
          subjectId: email,
          subjectType: SubjectType.User,
        })),
        scopes: [
          {
            scopeId: projectId,
            scopeType: ScopeType.Project,
          },
        ],
      },
    };

    const result = await accessRuleService.accessRulesBatchCreate(batchRequest);
    const failedSubjectIds = extractFailedSubjectIds(result.failed || []);
    const successCount = result.succeeded?.length || 0;

    return {
      createdAccessRules: successCount,
      failedSubjectIds,
    };
  } catch (error) {
    return {
      createdAccessRules: 0,
      failedSubjectIds: memberEmails,
    };
  }
}

/**
 * Create access rules for SSO groups
 */
async function createTeamAccessRulesWithGroup(
  groupName: string,
  projectId: string,
  clusterId: string,
): Promise<AccessRuleCreationResults> {
  try {
    const aiPractitionerRoleId = await getAIPractitionerRoleId();

    const batchRequest: AccessRulesBatchCreateFields = {
      payload: {
        roleId: aiPractitionerRoleId,
        clusterId: clusterId,
        subjects: [
          {
            subjectId: groupName,
            subjectType: SubjectType.Group, // This is the key difference - using Group instead of User
          },
        ],
        scopes: [
          {
            scopeId: projectId,
            scopeType: ScopeType.Project,
          },
        ],
      },
    };

    const result = await accessRuleService.accessRulesBatchCreate(batchRequest);
    const failedSubjectIds = extractFailedSubjectIds(result.failed || []);
    const successCount = result.succeeded?.length || 0;

    return {
      createdAccessRules: successCount,
      failedSubjectIds,
    };
  } catch (error) {
    return {
      createdAccessRules: 0,
      failedSubjectIds: [groupName],
    };
  }
}

/**
 * Extract failed subject IDs from batch creation result
 */
function extractFailedSubjectIds(failedRules: AccessRulesBatchCreateResponseFailedInner[]): string[] {
  if (!failedRules?.length) return [];

  return failedRules.map((failedRule) => failedRule.payload?.subjectId).filter((id): id is string => Boolean(id));
}

/**
 * Get the AI practitioner role ID with fallback logic
 */
async function getAIPractitionerRoleId(): Promise<number> {
  try {
    const roles = await rolesService.list();

    // Try to find exact AI practitioner role
    const aiPractitionerRole = roles.find((role: Role) => role.name === AI_PRACTITIONER_ROLE_NAME);
    if (aiPractitionerRole) {
      return aiPractitionerRole.id;
    }

    // Fallback to any practitioner role
    const fallbackRole = roles.find((role: Role) => role.name.toLowerCase().includes("practitioner"));

    if (!fallbackRole) {
      throw new Error("No practitioner role found");
    }

    return fallbackRole.id;
  } catch (error) {
    throw new Error("Failed to get practitioner role");
  }
}

/**
 * Get the system administrator role ID
 */
async function getSystemAdminRoleId(): Promise<number> {
  try {
    const roles = await rolesService.list();
    const systemAdminRole = roles.find((role: Role) => role.name === SYSTEM_ADMIN_ROLE_NAME);

    if (!systemAdminRole) {
      throw new Error("System administrator role not found");
    }

    return systemAdminRole.id;
  } catch (error) {
    throw new Error("Failed to get system administrator role");
  }
}
