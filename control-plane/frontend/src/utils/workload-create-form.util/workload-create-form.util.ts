import type { IUIWorkloadCreation } from "@/models/workload.model";
import {
  ENodePoolsListOrigin,
  type IComputeSectionNodePoolData,
  type IComputeSectionData,
} from "@/components/section/compute-resource-section";
import type { Resources } from "@/swagger-models/org-unit-service-client";
import type { IUIGeneralSectionModel } from "@/components/section/general-section";
import type { IUIWorkloadEnvSectionModel } from "@/components/section/environment-section";

export const workloadCreateFormUtil = {
  getComputeResourceData,
  getEnvironmentRunParamsData,
  getGeneralSettingsData,
};

function getComputeResourceData(
  workload: IUIWorkloadCreation,
  projectNodepools: Array<Resources> | undefined | null = [],
): IComputeSectionData {
  const nodePoolsData: IComputeSectionNodePoolData | null = _getNodePoolsData(workload, projectNodepools || undefined);
  return {
    computeResourceId: workload.assets.compute || null,
    nodeType: workload.specificEnv?.nodeType || null,
    nodePools: nodePoolsData,
    workers: workload.distributed?.numWorkers || null,
    tolerations: workload.specificEnv?.tolerations,
    podAffinity: workload.specificEnv?.podAffinity,
  };
}

function getEnvironmentRunParamsData(workload: IUIWorkloadCreation): IUIWorkloadEnvSectionModel {
  const specificEnv = workload.specificEnv || {};
  return {
    command: specificEnv.command || "",
    args: specificEnv.args || "",
    environmentVariables: specificEnv.environmentVariables || [],
    connections: specificEnv.connections || [],
    runAsGid: specificEnv.runAsGid || null,
    runAsUid: specificEnv.runAsUid || null,
    supplementalGroups: specificEnv.supplementalGroups || null,
    servingPortAccess: specificEnv.servingPortAccess || null,
  };
}

function getGeneralSettingsData(workload: IUIWorkloadCreation): IUIGeneralSectionModel {
  return {
    allowOverQuota: workload.specificEnv?.allowOverQuota || false,
    autoDeletionTimeAfterCompletionSeconds: Number.isInteger(
      workload.specificEnv?.autoDeletionTimeAfterCompletionSeconds,
    )
      ? workload.specificEnv?.autoDeletionTimeAfterCompletionSeconds
      : null,
    annotations: workload.specificEnv?.annotations,
    labels: workload.specificEnv?.labels,
    backoffLimit: Number.isInteger(workload.specificEnv?.backoffLimit) ? workload.specificEnv?.backoffLimit : null,
  };
}

function _getNodePoolsData(
  workload: IUIWorkloadCreation,
  nodePoolsResources: Array<Resources> = [],
): IComputeSectionNodePoolData | null {
  if (!workload.specificEnv || !workload.specificEnv.nodePools) {
    return null;
  }

  const allNodePoolsOptions: Array<string> = nodePoolsResources.map(
    (nodePoolResource: Resources) => nodePoolResource.nodePool?.name || "",
  );

  return {
    defaultNodePools: workload.specificEnv.nodePools,
    allNodePoolsOptions,
    nodePoolsListOrigin: ENodePoolsListOrigin.Project,
  };
}
