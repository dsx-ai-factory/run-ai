import { EFilterOperator, type IAssetsFilter } from "@/models/filter.model";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { policyUtil } from "@/utils/policy.util/policy.util";
import { useClusterStore } from "@/stores/cluster.store";

import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { aiApplicationService } from "@/services/cluster/ai-application.service/ai-application.service";
import {
  workloadTemplateService,
  type TTemplateSortBy,
} from "@/services/control-plane/workload-template.service/workload-template.service";
import { dataVolumeService } from "@/services/control-plane/data-volume.service/data-volume.service";
import { environmentService } from "@/services/control-plane/environment.service/environment.service";
import { computeResourceService } from "@/services/control-plane/compute-resource.service/compute-resource.service";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { policyService } from "@/services/control-plane/policy.service/policy.service";
import { auditServiceApi } from "@/services/infra/client-apis/audit-service-api/audit-service-api";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import { userService } from "@/services/control-plane/user.service/user.service";
import { rolesService } from "@/services/control-plane/rbac/roles.service/roles.service";
import { serviceAccountsService } from "@/services/control-plane/service-accounts.service/service-accounts.service";

import {
  DepartmentFilterSortFields,
  ProjectFilterSortFields,
  Verbosity,
} from "@/swagger-models/org-unit-service-client";
import {
  NodepoolSortFilterFields,
  NodeFilterSortFields,
  type Nodepool,
  type Node,
  type DisplayedCluster,
} from "@/swagger-models/cluster-service-client";
import type { ReportFilterAndSortFields } from "@/swagger-models/metrics-service-client";
import {
  WorkloadSortFilterFields,
  type Workload,
  type SupersetTemplate,
  type AIApplication,
} from "@/swagger-models/workloads-client";
import type { DatavolumeNoSharedScopes } from "@/swagger-models/datavolumes-service-client";
import {
  Scope,
  type EnvironmentAsset,
  type ComputeAsset,
  type DatasourceListResponseEntry,
  type CredentialsListResponseEntry,
} from "@/swagger-models/assets-service-client";
import type { PolicyListEntry } from "@/swagger-models/policy-service-client";
import type { AuditLog } from "@/swagger-models/audit-service-client";
import type { AccessRulesSortFilterFields, AccessRule, RoleV1 } from "@/swagger-models/authorization-client";
import type { UsersFilterSortFields, User, ServiceAccount } from "@/swagger-models/identity-manager-client";

const AUTOCOMPLETE_MAX_RESULTS = 100;

type AutocompleteFn = (searchQuery: string, clusterUuid?: string) => Promise<string[]>;

function extractUniqueValues<T>(entities: T[], extractFn: (entity: T) => string | undefined): string[] {
  const uniqueValues = new Set<string>();
  for (const entity of entities) {
    const value = extractFn(entity);
    if (value) uniqueValues.add(value);
  }
  return [...uniqueValues].slice(0, AUTOCOMPLETE_MAX_RESULTS);
}

// Empty string means "All clusters" selected (null filter); undefined falls back to the store.
function resolveClusterFilter(clusterUuid?: string): string | null {
  if (clusterUuid === "") return null;
  const clusterId = clusterUuid ?? useClusterStore().selectedClusterId;
  return filterUtil.getEqualsFilterString("clusterId", clusterId);
}

function createEntityAutocomplete<T>(
  fetchFn: (filterBy: string[], limit: number) => Promise<T[]>,
  fieldName: string,
  extractFn: (entity: T) => string | undefined,
  clusterScoped = true,
): AutocompleteFn {
  return async (searchQuery: string, clusterUuid?: string): Promise<string[]> => {
    const filterBy: string[] = [];
    if (clusterScoped) {
      const clusterFilter = resolveClusterFilter(clusterUuid);
      if (clusterFilter) {
        filterBy.push(clusterFilter);
      }
    }
    if (searchQuery) {
      filterBy.push(`${fieldName}${EFilterOperator.Contains}${searchQuery}`);
    }
    const entities = await fetchFn(filterBy, AUTOCOMPLETE_MAX_RESULTS);
    return extractUniqueValues(entities, extractFn);
  };
}

function createNameAutocomplete<T extends { name: string }>(
  fetchFn: (filterBy: string[], limit: number) => Promise<T[]>,
  fieldName: string,
  nameExtractor: (entity: T) => string = (entity) => entity.name,
  clusterScoped = true,
): AutocompleteFn {
  return async (searchQuery: string, clusterUuid?: string): Promise<string[]> => {
    const filterBy: string[] = [];
    if (clusterScoped) {
      const clusterFilter = resolveClusterFilter(clusterUuid);
      if (clusterFilter) {
        filterBy.push(clusterFilter);
      }
    }
    if (searchQuery) {
      filterBy.push(filterUtil.getContainsFilterString(fieldName, searchQuery));
    }
    const entities = await fetchFn(filterBy, AUTOCOMPLETE_MAX_RESULTS);
    return entities.map(nameExtractor).slice(0, AUTOCOMPLETE_MAX_RESULTS);
  };
}

function resolveClusterUuid(clusterUuid?: string): string | undefined {
  if (clusterUuid === "") return undefined;
  return clusterUuid ?? useClusterStore().selectedClusterId;
}

function clusterEntityFilter<T>(
  getClusterId: (entity: T) => string | null | undefined,
  scopeBypass?: (entity: T) => boolean,
): (entity: T, clusterUuid?: string) => boolean {
  return (entity, clusterUuid) => {
    if (!clusterUuid) return true;
    const entityClusterId = getClusterId(entity);
    if (!entityClusterId) return scopeBypass?.(entity) ?? false;
    return entityClusterId === clusterUuid;
  };
}

function createClientSideAutocomplete<T>(
  fetchFn: (clusterUuid?: string) => Promise<T[]>,
  extractFn: (entity: T) => string | undefined,
  entityFilterFn?: (entity: T, resolvedClusterUuid?: string) => boolean,
): AutocompleteFn {
  return async (searchQuery: string, clusterUuid?: string): Promise<string[]> => {
    const resolved = resolveClusterUuid(clusterUuid);
    const entities = await fetchFn(clusterUuid);
    const uniqueValues = new Set<string>();
    const lowerQuery = searchQuery.toLowerCase();
    for (const entity of entities) {
      if (entityFilterFn && !entityFilterFn(entity, resolved)) continue;
      const value = extractFn(entity);
      if (value && (!searchQuery || value.toLowerCase().includes(lowerQuery))) {
        uniqueValues.add(value);
      }
    }
    return [...uniqueValues].slice(0, AUTOCOMPLETE_MAX_RESULTS);
  };
}

function projects(): AutocompleteFn {
  return createNameAutocomplete(
    (filterBy, limit) => orgUnitService.listProjectsMetadata(filterBy, "name", "asc", 0, limit),
    ProjectFilterSortFields.Name,
  );
}

// Queried via workloads (not the departments API) to avoid permission issues.
function departments(): AutocompleteFn {
  return createEntityAutocomplete(
    (filterBy, limit) =>
      workloadService.getWorkloads(false, 0, limit, undefined, WorkloadSortFilterFields.DepartmentName, filterBy),
    WorkloadSortFilterFields.DepartmentName,
    (w) => w.departmentName,
  );
}

function departmentNames(): AutocompleteFn {
  return createNameAutocomplete(
    (filterBy, limit) =>
      orgUnitService.getDepartments(filterBy, DepartmentFilterSortFields.Name, Verbosity.Brief, "asc", 0, limit),
    DepartmentFilterSortFields.Name,
  );
}

// Queried via the projects API to avoid permission issues.
function departmentNamesViaProjects(): AutocompleteFn {
  return createEntityAutocomplete(
    (filterBy, limit) => orgUnitService.getProjects(filterBy, ProjectFilterSortFields.ParentName, "asc", 0, limit),
    ProjectFilterSortFields.ParentName,
    (p) => p.parent?.name,
  );
}

function nodePoolsMinimal(): AutocompleteFn {
  return createNameAutocomplete(
    (filterBy, limit) => clusterService.getNodepoolsMinimal(filterBy, undefined, undefined, 0, limit),
    NodepoolSortFilterFields.Name,
  );
}

function nodePools(
  fieldName: NodepoolSortFilterFields,
  extractFn: (np: Nodepool) => string | undefined,
): AutocompleteFn {
  return createEntityAutocomplete(
    (filterBy, limit) => clusterService.getNodepools(filterBy, fieldName, undefined, 0, limit),
    fieldName,
    extractFn,
  );
}

function nodes(fieldName: NodeFilterSortFields, extractFn: (node: Node) => string | undefined): AutocompleteFn {
  return async (searchQuery: string, clusterUuid?: string): Promise<string[]> => {
    const filterBy: string[] = [];
    const resolvedCluster = clusterUuid ?? useClusterStore().selectedClusterId;
    if (resolvedCluster) {
      filterBy.push(filterUtil.getEqualsFilterString(NodeFilterSortFields.ClusterUuid, resolvedCluster));
    }
    if (searchQuery) {
      filterBy.push(`${fieldName}${EFilterOperator.Contains}${searchQuery}`);
    }
    const entities = await clusterService.getTenantNodes(filterBy, fieldName, undefined, 0, AUTOCOMPLETE_MAX_RESULTS);
    return extractUniqueValues(entities, extractFn);
  };
}

function clusters(extractFn: (cluster: DisplayedCluster) => string | undefined): AutocompleteFn {
  return createClientSideAutocomplete(() => clusterService.list(), extractFn);
}

function reports(
  fieldName: ReportFilterAndSortFields,
  extractFn: (report: { name: string; createdBy: string }) => string | undefined,
): AutocompleteFn {
  return createEntityAutocomplete(
    (filterBy, limit) => orgUnitService.getReports(filterBy, fieldName, undefined, 0, limit),
    fieldName,
    extractFn,
    false,
  );
}

// Audit API requires an explicit date range (defaults to the last 90 days here).
function audits(fieldName: string, extractFn: (log: AuditLog) => string | undefined): AutocompleteFn {
  return createEntityAutocomplete(
    async (filterBy, limit) => {
      const now = new Date();
      const start = new Date(now.getTime() - 90 * 24 * 60 * 60 * 1000).toISOString();
      const end = now.toISOString();
      const response = await auditServiceApi.auditLogsApi.getAuditLogs(
        start,
        end,
        0,
        limit,
        undefined,
        undefined,
        filterBy,
      );
      return response.data.audit_logs ?? [];
    },
    fieldName,
    extractFn,
    false,
  );
}

function workloads(
  fieldName: WorkloadSortFilterFields,
  extractFn: (workload: Workload) => string | undefined,
): AutocompleteFn {
  return createEntityAutocomplete(
    (filterBy, limit) => workloadService.getWorkloads(false, 0, limit, undefined, fieldName, filterBy),
    fieldName,
    extractFn,
  );
}

// NOTE: departmentName is not returned by the AI apps API as a filterable field

function aiApplications(fieldName: string, extractFn: (app: AIApplication) => string | undefined): AutocompleteFn {
  return createEntityAutocomplete(
    (filterBy, limit) => aiApplicationService.getApplications(0, limit, undefined, fieldName, filterBy),
    fieldName,
    extractFn,
  );
}

function templates(fieldName: string, extractFn: (t: SupersetTemplate) => string | undefined): AutocompleteFn {
  return async (searchQuery: string, clusterUuid?: string): Promise<string[]> => {
    const filterBy: string[] = [];
    const clusterFilter = resolveClusterFilter(clusterUuid);
    if (clusterFilter) {
      filterBy.push(clusterFilter);
    }
    if (searchQuery) {
      filterBy.push(`${fieldName}${EFilterOperator.Contains}${searchQuery}`);
    }
    const includeAncestors = filterBy.some((f) => f.includes("clusterId=="));
    const entities = await workloadTemplateService.templatesList(
      0,
      AUTOCOMPLETE_MAX_RESULTS,
      undefined,
      fieldName as TTemplateSortBy,
      filterBy,
      undefined,
      includeAncestors,
    );
    return extractUniqueValues(entities, extractFn);
  };
}

function dataVolumes(
  fieldName: string,
  extractFn: (dv: DatavolumeNoSharedScopes) => string | undefined,
): AutocompleteFn {
  return createEntityAutocomplete(
    (filterBy, limit) => dataVolumeService.listOriginatedDataVolumes(0, limit, undefined, undefined, filterBy),
    fieldName,
    extractFn,
    false,
  );
}

function dataVolumesClientSide(extractFn: (dv: DatavolumeNoSharedScopes) => string | undefined): AutocompleteFn {
  return createClientSideAutocomplete(
    () => dataVolumeService.listOriginatedDataVolumes(0, AUTOCOMPLETE_MAX_RESULTS, undefined, undefined),
    extractFn,
  );
}
function environments(extractFn: (env: EnvironmentAsset) => string | undefined): AutocompleteFn {
  return createClientSideAutocomplete((clusterUuid?: string) => {
    const resolved = resolveClusterUuid(clusterUuid);
    const filter: IAssetsFilter = { includeDescendants: true };
    if (resolved) {
      filter.clusterUuid = resolved;
    }
    return environmentService.list(filter);
  }, extractFn);
}

const systemOrTenantBypass = <T extends { meta: { scope?: string } }>(entity: T) =>
  entity.meta.scope === Scope.System || entity.meta.scope === Scope.Tenant;

const tenantBypass = <T extends { meta: { scope?: string } }>(entity: T) => entity.meta.scope === Scope.Tenant;

function computeResources(extractFn: (cr: ComputeAsset) => string | undefined): AutocompleteFn {
  return createClientSideAutocomplete(
    () => computeResourceService.list(),
    extractFn,
    clusterEntityFilter((cr) => cr.meta.clusterId, systemOrTenantBypass),
  );
}

function dataSources(extractFn: (ds: DatasourceListResponseEntry) => string | undefined): AutocompleteFn {
  return createClientSideAutocomplete(
    () => dataSourceService.list(),
    extractFn,
    clusterEntityFilter((ds) => ds.meta.clusterId, systemOrTenantBypass),
  );
}

function credentials(extractFn: (cred: CredentialsListResponseEntry) => string | undefined): AutocompleteFn {
  return createClientSideAutocomplete(
    () => credentialService.listCredentials({}),
    extractFn,
    clusterEntityFilter((cred) => cred.meta.clusterId, tenantBypass),
  );
}

// Client-side: the policy list API has no server-side filterBy support.
function policies(extractFn: (policy: PolicyListEntry) => string | undefined): AutocompleteFn {
  return createClientSideAutocomplete<PolicyListEntry>(
    // Tag GVK rows (identified by a `group/Kind/Version` type) so extractors classify them like the grid.
    async () => {
      const entries = await policyService.listPolicies({ includeGvkPolicies: true });
      return entries.map((policy) =>
        policyUtil.isGvkPolicyListEntry(policy)
          ? ({ ...policy, isGvk: true, gvkKind: policyUtil.gvkKindFromPolicyType(policy.type) } as PolicyListEntry)
          : policy,
      );
    },
    extractFn,
    clusterEntityFilter((policy) => policy.meta?.clusterId),
  );
}

function accessRules(
  fieldName: AccessRulesSortFilterFields,
  extractFn: (rule: AccessRule) => string | undefined,
): AutocompleteFn {
  return createEntityAutocomplete(
    (filterBy, limit) => accessRuleService.getAccessRulesV2(filterBy, fieldName, "asc", 0, limit),
    fieldName,
    extractFn,
    false,
  );
}

function users(fieldName: UsersFilterSortFields, extractFn: (user: User) => string | undefined): AutocompleteFn {
  return createEntityAutocomplete(
    (filterBy, limit) => userService.getUsers(filterBy, fieldName, "asc", 0, limit),
    fieldName,
    extractFn,
    false,
  );
}

function roles(extractFn: (role: RoleV1) => string | undefined): AutocompleteFn {
  return createClientSideAutocomplete(() => rolesService.list(), extractFn);
}

function serviceAccounts(extractFn: (sa: ServiceAccount) => string | undefined): AutocompleteFn {
  return createClientSideAutocomplete(() => serviceAccountsService.list(), extractFn);
}

export const autocompleteUtil = {
  createEntityAutocomplete,
  createNameAutocomplete,
  createClientSideAutocomplete,

  projects,
  departments,
  departmentNames,
  departmentNamesViaProjects,
  nodePoolsMinimal,
  nodePools,
  nodes,
  clusters,
  reports,
  audits,
  workloads,
  aiApplications,
  templates,
  dataVolumes,
  dataVolumesClientSide,
  environments,
  computeResources,
  dataSources,
  credentials,
  policies,
  accessRules,
  users,
  roles,
  serviceAccounts,

  AUTOCOMPLETE_MAX_RESULTS,
};
