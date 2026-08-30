import type { IOrgTreeNode, IOrgTreeNodeId } from "@/models/org-tree.model";
import { EOrgTreeNodeIcons } from "@/models/org-tree.model";
import { ScopeType, type PermittedScopes } from "@/swagger-models/authorization-client";
import type { AssetMeta } from "@/swagger-models/assets-service-client";
import { deepCopy } from "@/utils/common.util";
import { orgUnitServiceApi } from "@/services/infra/client-apis/org-unit-service-api/org-unit-service-api";
import type { OrgUnit } from "@/swagger-models/org-unit-service-client";

export const orgTreeService = {
  getOrgTreeUnits,
  getOrgTree,
  getNodePathById,
  getExpandedPermittedScopes,
  isAllowedScope,
  getParentClusterUuid,
  getOrgTreeNodeId,
  filterHiddenScopes,
  findMatchingNodePaths,
  sortOrgTree,
};

async function getOrgTreeUnits(): Promise<OrgUnit[]> {
  const orgTreeResponse = await orgUnitServiceApi.orgUnitApi.getOrgTree();
  return orgTreeResponse.data.orgUnits;
}

/*
 *  useful when we need to set scope-input / runai-org-tree selected-scope by entity id (cluster, department, project, etc.)
 */
function getOrgTreeNodeId(orgUnits: Array<OrgUnit>, id: string, type: ScopeType): IOrgTreeNodeId | null {
  const orgUnit: OrgUnit | undefined = orgUnits.find((unit) => unit.id === id && unit.type === type);
  const orgTree: IOrgTreeNode = getOrgTree(orgUnits);
  const path: string | null = getNodePathById(orgTree, { id, type });
  if (!orgUnit || !path) return null;
  return { id: orgUnit.id, type: orgUnit.type, path };
}

function filterHiddenScopes(scopes: IOrgTreeNode[], hiddenScopeTypes: Record<ScopeType, boolean>): IOrgTreeNode[] {
  function shouldHideScope(scope: IOrgTreeNode) {
    return hiddenScopeTypes[scope.type];
  }

  function filterScopes(scope: IOrgTreeNode) {
    if (shouldHideScope(scope)) {
      return null;
    }

    if (scope.children && scope.children.length > 0) {
      scope.children = scope.children
        .map((child) => filterScopes(child))
        .filter((child) => child !== null) as IOrgTreeNode[];
    }

    return scope;
  }

  return deepCopy(scopes)
    .map((scope) => filterScopes(scope))
    .filter((scope) => scope !== null) as IOrgTreeNode[];
}

function getOrgTree(orgUnits: Array<OrgUnit>): IOrgTreeNode {
  const tenantUnit: OrgUnit | undefined = orgUnits.find((unit) => unit.type === ScopeType.Tenant);
  if (!tenantUnit) return {} as IOrgTreeNode;
  // some ids are not unique (departments and projects can have the same id)
  const tenantUniqueKey: string = tenantUnit.id + tenantUnit.type;
  const orgUnitMap: Record<string, OrgUnit> = {};
  const treeNodeMap: Record<string, IOrgTreeNode> = {};
  orgUnits.forEach((unit: OrgUnit) => {
    treeNodeMap[unit.id + unit.type] = _orgUnitToNode(unit);
    orgUnitMap[unit.id + unit.type] = unit;
  });
  for (const node in treeNodeMap) {
    const unit: OrgUnit = orgUnitMap[node];
    if (unit.parentId && unit.parentType) {
      const parent: IOrgTreeNode = treeNodeMap[unit.parentId + unit.parentType];
      if (!parent) continue;
      parent.children.push(treeNodeMap[node]);
    }
  }
  treeNodeMap[tenantUniqueKey].path = tenantUnit.name;
  _setNodesPath(treeNodeMap[tenantUniqueKey].children, treeNodeMap[tenantUniqueKey].path, orgUnitMap);
  return treeNodeMap[tenantUniqueKey];
}

function getNodePathById(node: IOrgTreeNode, nodeId: IOrgTreeNodeId): string | null {
  if (node.id === nodeId.id && node.type === nodeId.type) {
    return node.path;
  }
  for (const child of node.children || []) {
    const result = getNodePathById(child, nodeId);
    if (result) {
      return result;
    }
  }
  return null;
}

function getExpandedPermittedScopes(permittedScopes: PermittedScopes, orgUnits: Array<OrgUnit>): PermittedScopes {
  return orgUnits.reduce(
    (acc: PermittedScopes, unit: OrgUnit) => {
      const { type, id } = unit;
      const { tenants = [], clusters = [], departments = [], projects = [] } = acc;
      if (type === ScopeType.Tenant && permittedScopes.tenants?.includes(id)) {
        return { ...acc, tenants: [...tenants, id] };
      }
      if (
        type === ScopeType.Cluster &&
        (permittedScopes.clusters?.includes(id) || _checkParentPermission(permittedScopes, unit, orgUnits))
      ) {
        return { ...acc, clusters: [...clusters, id] };
      }
      if (
        type === ScopeType.Department &&
        (permittedScopes.departments?.includes(id) || _checkParentPermission(permittedScopes, unit, orgUnits))
      ) {
        return { ...acc, departments: [...departments, id] };
      }
      if (
        type === ScopeType.Project &&
        (permittedScopes.projects?.includes(id) || _checkParentPermission(permittedScopes, unit, orgUnits))
      ) {
        return { ...acc, projects: [...projects, id] };
      }
      return acc;
    },
    { clusters: [], departments: [], projects: [] } as PermittedScopes,
  );
}

function _setNodesPath(nodes: Array<IOrgTreeNode>, pathPrefix: string, orgUnitMap: Record<string, OrgUnit>): void {
  nodes.forEach((node: IOrgTreeNode) => {
    // some ids are not unique (departments and projects can have the same id)
    const nodeUniqueKey = node.id + node.type;
    node.path = `${pathPrefix}/${orgUnitMap[nodeUniqueKey].name}`;
    if (node.children.length > 0) {
      _setNodesPath(node.children, node.path, orgUnitMap);
    }
  });
}

function _orgUnitToNode(unit: OrgUnit): IOrgTreeNode {
  const iconMap = {
    [ScopeType.System]: EOrgTreeNodeIcons.Tenant, // Won't be shown in the org tree, but needed for the icon map to avoid ts errors
    [ScopeType.Tenant]: EOrgTreeNodeIcons.Tenant,
    [ScopeType.Cluster]: EOrgTreeNodeIcons.Cluster,
    [ScopeType.Department]: EOrgTreeNodeIcons.Department,
    [ScopeType.Project]: EOrgTreeNodeIcons.Project,
  };
  return {
    id: unit.id,
    path: unit.name,
    type: unit.type,
    children: [],
    icon: iconMap[unit.type],
    label: unit.name,
  };
}

function _checkParentPermission(permittedScopes: PermittedScopes, unit: OrgUnit, orgUnits: Array<OrgUnit>): boolean {
  const { parentType, parentId } = unit;
  if (!parentId || !parentType) return false;
  if (parentType === ScopeType.Tenant && permittedScopes.tenants?.includes(parentId)) return true;
  if (parentType === ScopeType.Cluster && permittedScopes.clusters?.includes(parentId)) return true;
  if (parentType === ScopeType.Department && permittedScopes.departments?.includes(parentId)) return true;
  const parentUnit: OrgUnit | undefined = orgUnits.find((unit) => unit.id === parentId && unit.type === parentType);
  if (!parentUnit) return false;
  // Checking permission for more than 1 level of parent
  return _checkParentPermission(permittedScopes, parentUnit, orgUnits);
}

function isAllowedScope(meta: AssetMeta, allowedScopes: PermittedScopes, orgUnits: Array<OrgUnit>): boolean {
  const expandedAllowedScopes = getExpandedPermittedScopes(allowedScopes, orgUnits);
  const { scope, departmentId, projectId, clusterId, tenantId } = meta;
  switch (scope) {
    case ScopeType.Tenant:
      return !!expandedAllowedScopes.tenants?.includes(tenantId?.toString() ?? "");
    case ScopeType.Project:
      return !!expandedAllowedScopes.projects?.includes(projectId + "");
    case ScopeType.Department:
      return !!expandedAllowedScopes.departments?.includes(departmentId || "");
    case ScopeType.Cluster:
      return !!expandedAllowedScopes.clusters?.includes(clusterId || "");
    default:
      return false;
  }
}

function getParentClusterUuid(orgUnits: Array<OrgUnit>, unitId: string, unitType: ScopeType): string | null {
  if (unitType === ScopeType.Tenant) return null;
  if (unitType === ScopeType.Cluster) return unitId;
  const unit = orgUnits.find((unit) => unit.id === unitId && unit.type === unitType);
  if (!unit || !unit.parentId || !unit.parentType) return null;
  return getParentClusterUuid(orgUnits, unit.parentId, unit.parentType);
}

function findMatchingNodePaths(nodes: IOrgTreeNode[], parentPaths: string[], searchTerm: string): string[] {
  if (!searchTerm) return [];
  const matchingPaths: Set<string> = new Set();
  nodes.forEach((node) => {
    if (node.label?.toLowerCase().includes(searchTerm.toLowerCase())) {
      [...parentPaths, node.path].forEach((parentPath) => {
        matchingPaths.add(parentPath);
      });
    }
    if (node.children.length > 0) {
      const childMatchingPaths = findMatchingNodePaths(node.children, [...parentPaths, node.path], searchTerm);
      childMatchingPaths.forEach((path) => matchingPaths.add(path));
    }
  });
  return Array.from(matchingPaths);
}

export function sortOrgTree(nodes: IOrgTreeNode[]): IOrgTreeNode[] {
  if (!nodes || nodes.length === 0) {
    return nodes;
  }

  nodes.sort((a: IOrgTreeNode, b: IOrgTreeNode) => a.label?.localeCompare(b.label || "") || 0);

  nodes.forEach((node) => {
    if (node.children && node.children.length > 0) {
      node.children = sortOrgTree(node.children);
    }
  });

  return nodes;
}
