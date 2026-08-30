/* eslint-disable */

import type { PolicyMeta } from "@/swagger-models/policy-service-client";
import {
  PolicyType,
  PolicySyncStatusOfClusterStatusEnum,
  ScopeType,
  type PolicyListEntry,
} from "@/swagger-models/policy-service-client";

import { tableUtil } from "../table.util/table.util";
import type { IStatusColOptions } from "@/models/table.model";
import { capitalizeString } from "../string.util/string.util";
import type { TPolicyScopes } from "@/models/policy.model";
import { errorMessages } from "@/common/error-message.constant";
import type { ISelectOption } from "@/models/global.model";
import {
  PolicyRuleEnum,
  type ComplianceInfoReason,
  type UnifiedPolicyInfoPerReplicaType,
} from "@/swagger-models/workloads-client";
import type {
  InferenceSpecificRunParams,
  PolicyInfo,
  WorkspaceSpecificRunParams,
} from "@/swagger-models/assets-service-client";

import { makeId } from "../common.util";
import type { IUISuperWorkloadSpec, IUIWorkloadStorageFields } from "@/models/workload.model";

export const policyUtil = {
  createPolicyIdFromPolicyTypeAndMeta,
  createPolicyId,
  resolveSelectedPolicyId,
  createPolicyIdFromPolicyListEntry,
  getPolicyTypeScopeAndScopeId,
  isGvkPolicyId,
  isGvkPolicyListEntry,
  gvkKindFromPolicyType,
  getPolicyStatusColOptions,
  getQueryObjectFromPolicyId,
  getPolicyScopePropertiesFromPolicyId,
  getOptionsWithRules,
  cleanupItemizeExcludeDuplicates,
  cleanupLockedItems,
  cleanupItemizeBeforeSubmission,
  cleanupFlatItemizeBeforeSubmission,
  isNonManageableRuleConflict,
};

export const policyIdSeperator = "_";

type IListPolicyQueryPolicyType = "Workspace" | "Training" | "Distributed" | "Inference" | "DistributedInference";
export interface IListPolicyQuery {
  policyType?: IListPolicyQueryPolicyType;
  scope?: TPolicyScopes;
  departmentId?: string;
  projectId?: string;
  clusterId?: string;
  includeFallbackPolicies?: boolean;
  includeGvkPolicies?: boolean;
}

export interface IPolicyTypeScopeAndScopeId {
  type: PolicyType;
  scope: TPolicyScopes;
  scopeId: string | undefined;
}

export interface IPolicyScopeProperties {
  scope: ScopeType;
  projectId?: number | null;
  departmentId?: string | null;
  clusterId?: string | null;
}

function createPolicyIdFromPolicyListEntry(policy: PolicyListEntry) {
  if (!policy.type) return policyIdSeperator;
  return createPolicyIdFromPolicyTypeAndMeta(policy.type, policy.meta);
}

function createPolicyIdFromPolicyTypeAndMeta(policyType: PolicyType, policyMeta?: PolicyMeta): string {
  if (!policyMeta) return policyIdSeperator;
  const { scope, projectId, departmentId, clusterId } = policyMeta;
  switch (scope) {
    case ScopeType.Tenant:
      return policyUtil.createPolicyId(policyType, ScopeType.Tenant);
    case ScopeType.Cluster:
      return policyUtil.createPolicyId(policyType, ScopeType.Cluster, String(clusterId));
    case ScopeType.Department:
      return policyUtil.createPolicyId(policyType, ScopeType.Department, String(departmentId));
    case ScopeType.Project:
      return policyUtil.createPolicyId(policyType, ScopeType.Project, String(projectId));
    default:
      return policyIdSeperator;
  }
}

function createPolicyId(policyType: PolicyType, policyScope: ScopeType, scopeId?: string | number): string {
  return !!scopeId
    ? `${policyType}${policyIdSeperator}${policyScope}${policyIdSeperator}${scopeId}`
    : `${policyType}${policyIdSeperator}${policyScope}`;
}

function resolveSelectedPolicyId(
  selectedPolicy: PolicyType | string,
  policyScope: ScopeType,
  scopeId?: string | number,
): string {
  const isNativePolicyType = Object.values(PolicyType).includes(selectedPolicy as PolicyType);
  return isNativePolicyType ? createPolicyId(selectedPolicy as PolicyType, policyScope, scopeId) : selectedPolicy;
}

function getPolicyTypeScopeAndScopeId(policyId: string): IPolicyTypeScopeAndScopeId {
  const [type, scope, scopeId] = policyId.split(policyIdSeperator);
  return {
    type: type as PolicyType,
    scope: scope as TPolicyScopes,
    scopeId,
  };
}

/** GVK ids are backend UUIDs; native ids always contain `policyIdSeperator`, so its absence marks a GVK id. */
function isGvkPolicyId(policyId: string): boolean {
  return !!policyId && !policyId.includes(policyIdSeperator);
}

function isGvkPolicyListEntry(policy: PolicyListEntry): boolean {
  return !!policy.type && policy.type.includes("/");
}

function gvkKindFromPolicyType(type: string | undefined): string {
  return type?.split("/")[1] ?? type ?? "";
}

function getQueryObjectFromPolicyId(policyId: string): IListPolicyQuery {
  const policyTypeAndScope: IPolicyTypeScopeAndScopeId = getPolicyTypeScopeAndScopeId(policyId);

  const policyType = capitalizeString(policyTypeAndScope.type) as IListPolicyQueryPolicyType;

  switch (policyTypeAndScope.scope) {
    case ScopeType.Tenant:
      return {
        policyType,
        scope: policyTypeAndScope.scope,
      };
    case ScopeType.Cluster:
      return {
        policyType,
        scope: policyTypeAndScope.scope,
        clusterId: policyTypeAndScope.scopeId,
      };
    case ScopeType.Department:
      return {
        policyType,
        scope: policyTypeAndScope.scope,
        departmentId: policyTypeAndScope.scopeId,
      };
    case ScopeType.Project:
      return {
        policyType,
        scope: policyTypeAndScope.scope,
        projectId: policyTypeAndScope.scopeId,
      };
  }
}
function getPolicyScopePropertiesFromPolicyId(policyId: string): IPolicyScopeProperties {
  const policyTypeAndScope: IPolicyTypeScopeAndScopeId = getPolicyTypeScopeAndScopeId(policyId);

  switch (policyTypeAndScope.scope) {
    case ScopeType.Tenant:
      return {
        scope: policyTypeAndScope.scope,
      };
    case ScopeType.Cluster:
      return {
        scope: policyTypeAndScope.scope,
        clusterId: String(policyTypeAndScope.scopeId) || undefined,
      };
    case ScopeType.Department:
      return {
        scope: policyTypeAndScope.scope,
        departmentId: String(policyTypeAndScope.scopeId) || undefined,
      };
    case ScopeType.Project:
      return {
        scope: policyTypeAndScope.scope,
        projectId: Number(policyTypeAndScope.scopeId) || undefined,
      };
  }
}

function getPolicyStatusColOptions(
  status?: PolicySyncStatusOfClusterStatusEnum,
  errorMessage?: string,
): IStatusColOptions {
  if (!status) {
    return {
      status: "-",
      tooltipText: "",
      displayAnimation: false,
      filterKey: "status",
    };
  }
  return tableUtil.getStatusColOptions(policyStatusMap[status], errorMessage);
}

export const policyStatusMap: Record<PolicySyncStatusOfClusterStatusEnum, IStatusColOptions> = {
  [PolicySyncStatusOfClusterStatusEnum.Ready]: {
    status: "Ready",
    color: "success",
    displayAnimation: false,
    filterKey: "status",
  },
  [PolicySyncStatusOfClusterStatusEnum.NotReady]: {
    status: "Not Ready",
    color: "negative",
    displayAnimation: false,
    filterKey: "status",
  },
  [PolicySyncStatusOfClusterStatusEnum.Failed]: {
    status: "Failed",
    color: "negative",
    displayAnimation: false,
    filterKey: "status",
  },
  [PolicySyncStatusOfClusterStatusEnum.Deleted]: {
    status: "Deleted",
    displayAnimation: false,
    filterKey: "status",
  },
  [PolicySyncStatusOfClusterStatusEnum.DeletionFailed]: {
    status: "Failed",
    color: "negative",
    displayAnimation: false,
    filterKey: "status",
  },
  [PolicySyncStatusOfClusterStatusEnum.Applying]: {
    status: "Updating...",
    displayAnimation: true,
    filterKey: "status",
  },
  [PolicySyncStatusOfClusterStatusEnum.Deleting]: {
    status: "Deleting...",
    displayAnimation: true,
    filterKey: "status",
  },
  [PolicySyncStatusOfClusterStatusEnum.PendingDeletion]: {
    status: "Deleting...",
    displayAnimation: true,
    filterKey: "status",
  },
};

interface IRuleOpt<T> {
  value: T | null;
  displayed?: string | null;
}

interface ISelectOptionWithRules<T> extends Omit<ISelectOption, "value"> {
  value: T | null;
}

function getOptionsWithRules<T>(
  options: ISelectOption[],
  ruleOptions?: IRuleOpt<T>[] | null,
  isClosedList?: boolean,
  currentValue?: T | null,
): ISelectOption[] {
  if (!ruleOptions?.length) return options;

  if (isClosedList) {
    return options.map((option: ISelectOption) => {
      // looking for the option in the rules options to get the displayed name and know what needs to be disabled
      const foundInRulesOptions = ruleOptions?.find((rule) => rule.value === option.value);
      const opt = {
        ...option,
        label: foundInRulesOptions?.displayed || option.label,
      };
      if (!foundInRulesOptions) {
        opt.disable = true;
        opt.disabledTooltip = errorMessages.CANT_BE_SELECTED_POLICY;
      }
      return opt;
    });
  } else {
    const optionsMap = options.reduce(
      (acc, option) => {
        option.value && (acc[option.value.toString()] = option);
        return acc;
      },
      {} as Record<string, ISelectOption>,
    );

    // in case of open list we return the rules options only everything else is ignored
    const mappedOptions = ruleOptions.map((ruleOption: IRuleOpt<T>) => {
      const option = optionsMap[ruleOption.value?.toString() || ""];
      return {
        value: ruleOption.value,
        label: ruleOption.displayed || option?.label || ruleOption.value,
      } as ISelectOption;
    });

    // if the current value is not in the options we add it to the options with a disabled state.
    // we need to be sure not to add the same value twice
    // this is done so users would still be able to see the current value in the dropdown
    // the invalid value would be displayed if its already selected, and would be disabled in the list of options.
    if (currentValue && !mappedOptions.some((option) => option.value === currentValue)) {
      mappedOptions.push({
        value: currentValue,
        label: currentValue.toString(),
        disable: true,
        disabledTooltip: errorMessages.CANT_BE_SELECTED_POLICY,
      } as ISelectOption);
    }

    return mappedOptions;
  }
}

function cleanupItemizeExcludeDuplicates<T extends { exclude?: boolean | null }>(
  items: T[],
  mainKey: keyof T,
): Array<T> {
  const nameMap: { [key: string]: T } = {};
  for (const item of items) {
    const mainKeyValue = item[mainKey] as string;
    // add if not exists, if exclude exists in the map replace it.
    if (!nameMap[mainKeyValue] || (!item.exclude && nameMap[mainKeyValue].exclude)) {
      nameMap[mainKeyValue] = item;
    } else if (!nameMap[mainKeyValue].exclude && !item.exclude) {
      // if we have a duplicate we will add it with a new id
      nameMap[makeId(6)] = item;
    }
  }
  return Object.values(nameMap);
}

function cleanupLockedItems<T>(items: T[], mainKey: keyof T, locked: string[]): T[] {
  return items.filter((item) => !locked.includes(item[mainKey] as string));
}

function cleanupItemizeBeforeSubmission<T extends WorkspaceSpecificRunParams & InferenceSpecificRunParams>(
  specificRunParams: T,
  policyRules?: PolicyInfo | null,
): T {
  const annoNoDups = cleanupItemizeExcludeDuplicates(specificRunParams.annotations || [], "name");
  specificRunParams.annotations = cleanupLockedItems(
    annoNoDups || [],
    "name",
    policyRules?.rules?.specificRunParams?.annotations?.instances?.locked || [],
  );

  const labelNoDups = cleanupItemizeExcludeDuplicates(specificRunParams.labels || [], "name");
  specificRunParams.labels = cleanupLockedItems(
    labelNoDups || [],
    "name",
    policyRules?.rules?.specificRunParams?.labels?.instances?.locked || [],
  );

  const envVarNoDups = cleanupItemizeExcludeDuplicates(specificRunParams.environmentVariables || [], "name");
  specificRunParams.environmentVariables = cleanupLockedItems(
    envVarNoDups || [],
    "name",
    policyRules?.rules?.specificRunParams?.environmentVariables?.instances?.locked || [],
  );

  const tolNoDups = cleanupItemizeExcludeDuplicates(specificRunParams.tolerations || [], "name");
  specificRunParams.tolerations = cleanupLockedItems(
    tolNoDups || [],
    "name",
    policyRules?.rules?.specificRunParams?.tolerations?.instances?.locked || [],
  );

  return specificRunParams;
}

function cleanupFlatItemizeBeforeSubmission<T extends IUISuperWorkloadSpec>(
  flatSpec: T,
  policy?: UnifiedPolicyInfoPerReplicaType | null,
): T {
  // Clean up itemized fields (duplicates and locked items)
  const annoNoDups = cleanupItemizeExcludeDuplicates(flatSpec.annotations || [], "name");
  flatSpec.annotations = cleanupLockedItems(
    annoNoDups || [],
    "name",
    policy?.rules?.annotations?.instances?.locked || [],
  );

  const labelNoDups = cleanupItemizeExcludeDuplicates(flatSpec.labels || [], "name");
  flatSpec.labels = cleanupLockedItems(labelNoDups || [], "name", policy?.rules?.labels?.instances?.locked || []);

  const envVarNoDups = cleanupItemizeExcludeDuplicates(flatSpec.environmentVariables || [], "name");
  flatSpec.environmentVariables = cleanupLockedItems(
    envVarNoDups || [],
    "name",
    policy?.rules?.environmentVariables?.instances?.locked || [],
  );

  const tolNoDups = cleanupItemizeExcludeDuplicates(flatSpec.tolerations || [], "name");
  flatSpec.tolerations = cleanupLockedItems(
    tolNoDups || [],
    "name",
    policy?.rules?.tolerations?.instances?.locked || [],
  );

  if (flatSpec.compute?.extendedResources) {
    const eResNoDups = cleanupItemizeExcludeDuplicates(flatSpec.compute.extendedResources || [], "resource");
    flatSpec.compute.extendedResources = cleanupLockedItems(
      eResNoDups || [],
      "resource",
      policy?.rules?.compute?.extendedResources?.instances?.locked || [],
    );
  }

  cleanupStorageLockedItems(flatSpec.storage, policy);
  return flatSpec;
}

// cleanupStorageLockedItems is a function that cleans up the locked items in the storage fields of the workload spec
// we should not send the locked items to the backend
function cleanupStorageLockedItems(
  storage: IUIWorkloadStorageFields | undefined,
  policy?: UnifiedPolicyInfoPerReplicaType | null,
) {
  if (!storage) return;

  if (storage.pvc) {
    storage.pvc = cleanupLockedItems(storage.pvc || [], "name", policy?.rules?.storage?.pvc?.instances?.locked || []);
  }

  if (storage.volumes) {
    storage.volumes = cleanupLockedItems(
      storage.volumes || [],
      "name",
      policy?.rules?.storage?.pvc?.instances?.locked || [],
    );
  }

  if (storage.nfs) {
    storage.nfs = cleanupLockedItems(storage.nfs || [], "name", policy?.rules?.storage?.nfs?.instances?.locked || []);
  }

  if (storage.s3) {
    storage.s3 = cleanupLockedItems(storage.s3 || [], "name", policy?.rules?.storage?.s3?.instances?.locked || []);
  }

  if (storage.hostPath) {
    storage.hostPath = cleanupLockedItems(
      storage.hostPath || [],
      "name",
      policy?.rules?.storage?.hostPath?.instances?.locked || [],
    );
  }

  if (storage.git) {
    storage.git = cleanupLockedItems(storage.git || [], "name", policy?.rules?.storage?.git?.instances?.locked || []);
  }

  if (storage.configMapVolume) {
    storage.configMapVolume = cleanupLockedItems(
      storage.configMapVolume || [],
      "name",
      policy?.rules?.storage?.configMapVolume?.instances?.locked || [],
    );
  }

  if (storage.secretVolume) {
    storage.secretVolume = cleanupLockedItems(
      storage.secretVolume || [],
      "name",
      policy?.rules?.storage?.secretVolume?.instances?.locked || [],
    );
  }

  if (storage.emptyDirVolume) {
    storage.emptyDirVolume = cleanupLockedItems(
      storage.emptyDirVolume || [],
      "name",
      policy?.rules?.storage?.emptyDirVolume?.instances?.locked || [],
    );
  }
}

function isNonManageableRuleConflict(complianceIssue: ComplianceInfoReason): boolean {
  return (
    complianceIssue.rule === PolicyRuleEnum.CanEdit ||
    complianceIssue.rule === PolicyRuleEnum.CanAdd ||
    complianceIssue.rule === PolicyRuleEnum.Blocked
  );
}
