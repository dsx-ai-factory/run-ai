import { controlPlaneService } from "@/services/control-plane/control-plane.service/control-plane.service";
import type { Action, PermittedScopes, ResourceType, Permission } from "@/swagger-models/authorization-client";

const ENDPOINT = "api/v1/authorization";

async function getPermissions(): Promise<Permission[]> {
  return controlPlaneService.get(`${ENDPOINT}/permissions`);
}

async function getPermittedScopesForAction(
  resourceType: ResourceType,
  action?: Action,
): Promise<Record<Action, PermittedScopes>> {
  return controlPlaneService.post(`${ENDPOINT}/permitted-scopes`, { resourceType, action });
}

export const permissionService = {
  getPermissions,
  getPermittedScopesForAction,
};
