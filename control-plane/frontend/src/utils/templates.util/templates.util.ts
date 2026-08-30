import { Scope } from "@/swagger-models/assets-service-client";
import type { PermittedScopes } from "@/swagger-models/authorization-client";
import type { TemplateMetaFields } from "@/swagger-models/workloads-client";

export const templatesUtil = {
  isPermittedTemplate,
};

function isPermittedTemplate(
  templateMeta: TemplateMetaFields,
  permissions: PermittedScopes,
  tenantId?: string,
): boolean {
  const { scopeType, scopeId } = templateMeta;

  switch (scopeType) {
    case Scope.System:
      // System templates are validated at tenant level (backend converts system scope to tenant for auth)
      return !!tenantId && !!permissions.tenants?.includes(tenantId);
    case Scope.Tenant:
      return !!permissions.tenants?.includes(scopeId || "");
    case Scope.Cluster:
      return !!permissions.clusters?.includes(scopeId || "");
    case Scope.Department:
      return !!permissions.departments?.includes(scopeId || "");
    case Scope.Project:
      return !!permissions.projects?.includes(scopeId || "");
    default:
      return false;
  }
}
