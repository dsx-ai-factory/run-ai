import type { Permission, RoleV1 } from "@/swagger-models/authorization-client";
import { authorizationApi } from "@/services/infra/client-apis/authorization-api/authorization-api";

export const rolesService = {
  list,
  isPermittedRole,
};
async function list(): Promise<RoleV1[]> {
  const response = await authorizationApi.rolesApi.getRolesV1();
  return response.data;
}

function isPermittedRole(role: RoleV1, permissions: Array<Permission>): boolean {
  // Check that the permissions of the current user contain at least all the resource types of the role
  for (const permission of role.permissions) {
    const rolePermission = permissions.find((p) => p.resourceType === permission.resourceType);
    if (!rolePermission) {
      return false;
    }
    // Check that the role's permission object has at least all the actions that the first permission object has
    for (const action of permission.actions) {
      if (!rolePermission.actions.includes(action)) {
        return false;
      }
    }
  }
  // If all checks pass, the role is valid
  return true;
}

export default rolesService;
