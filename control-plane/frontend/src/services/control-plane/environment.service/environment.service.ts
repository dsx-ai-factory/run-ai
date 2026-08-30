import { assetsServiceApi } from "@/services/infra/client-apis/assets-service-api/assets-service-api";
import type { IAssetsFilter } from "@/models/filter.model";
import { httpService } from "@/services/infra/https.service/http.service";

import { makeId, pick } from "@/utils/common.util";
import {
  ToolType,
  InternalConnectionType,
  ImagePullPolicy,
  Scope,
  type Connection,
  type EnvironmentAsset,
  type EnvironmentCreationRequest,
  type EnvironmentVariableOfAsset,
  type InternalToolInfo,
  type AssetCreationRequest,
  type EnvironmentAssetSpec,
  UidGidSource,
  type EnvironmentUpdateRequest,
  type HttpResponse,
} from "@/swagger-models/assets-service-client";
import type { IUIConnection, IUIEnvironmentAsset } from "@/models/environment.model";
import {
  isFirstLetterIsLowerCase,
  isLowerCaseAndNumbers,
  isNotEmpty,
  isValidDirectoryPath,
  validCommandAndArgsLine,
} from "@/common/form.validators";

export const environmentService = {
  list,
  getEnvironmentModel,
  getConnectionModel,
  save,
  remove,
  update,
  validateEnvironment,
  validateName,
  validateImage,
  validateConnections,
  validateUIConnections,
  validateRuntimeSettings,
  getById,
  prepareConnectionForRequest,
  prepareConnectionForUI,
};

// api calls
async function list(filterBy: IAssetsFilter = {}, filterByName?: string): Promise<Array<EnvironmentAsset>> {
  const filters: IAssetsFilter = pick(
    filterBy,
    "sortBy",
    "page",
    "rowsPerPage",
    "projectId",
    "departmentId",
    "scope",
    "usageInfo",
    "complyToProject",
    "complyToWorkloadType",
    "isWorkspace",
    "isTraining",
    "isInference",
    "isDistributed",
    "isDistributedInference",
    "distributedFramework",
    "clusterUuid",
    "complyToReplicaType",
    "statusInfo",
    "includeDescendants",
  );
  try {
    return assetsServiceApi.environmentApi
      .listEnvironmentAssets(
        filterByName || undefined,
        filters.scope,
        filters.projectId ? Number(filters.projectId) : undefined,
        filters.departmentId?.toString(),
        filters.clusterUuid?.toString(),
        filters.usageInfo,
        filters.complyToProject,
        filters.complyToWorkloadType as "Workspace" | "Training" | "Distributed" | "Inference" | "DistributedInference",
        filters.distributedFramework || undefined,
        filters.isDistributed,
        filters.isTraining,
        filters.isWorkspace,
        filters.isInference,
        filters.complyToReplicaType,
        filters.statusInfo,
        filters.includeDescendants,
        filters.isDistributedInference,
      )
      .then((res) => res.data.entries);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getById(environmentId: string): Promise<EnvironmentAsset> {
  try {
    const response = await assetsServiceApi.environmentApi.getEnvironmentAssetById(environmentId, true);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function save(environment: EnvironmentCreationRequest): Promise<EnvironmentAsset> {
  try {
    const response = await assetsServiceApi.environmentApi.createEnvironmentAsset(environment);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function update(environmentId: string, environment: EnvironmentUpdateRequest): Promise<EnvironmentAsset> {
  try {
    const response = await assetsServiceApi.environmentApi.updateEnvironmentAssetById(environmentId, environment);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function remove(environmentId: string): Promise<HttpResponse> {
  try {
    const response = await assetsServiceApi.environmentApi.deleteEnvironmentAssetById(environmentId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

function getConnectionModel(): IUIConnection {
  return {
    id: makeId(),
    name: "",
    toolType: null,
    connectionType: InternalConnectionType.ExternalUrl,
    containerPort: 0,
    isCustomUrl: false,
  };
}

function getEnvironmentModel(isSSO: boolean): IUIEnvironmentAsset {
  return {
    meta: {
      name: "",
      description: null,
      scope: null,
      projectId: null,
      workloadSupportedTypes: {
        workspace: true,
        training: true,
        inference: false,
      },
    },
    spec: {
      command: null,
      args: null,
      environmentVariables: [],
      image: "",
      imagePullPolicy: ImagePullPolicy.IfNotPresent,
      workingDir: null,
      connections: [],
      uidGidSource: isSSO ? UidGidSource.FromIdpToken : UidGidSource.FromTheImage,
      overrideUidGidInWorkspace: isSSO ? true : false,
      capabilities: [],
    },
  };
}

function validateEnvironment(environment: EnvironmentCreationRequest): boolean {
  const { meta, spec }: { meta: AssetCreationRequest; spec: EnvironmentAssetSpec } = environment;
  if (!validateScope(meta.scope, meta.projectId, meta.departmentId, meta.clusterId)) return false;
  if (!validateName(meta.name)) return false;
  if (!validateImage(spec.image || "")) return false;
  if (!validateConnections(spec.connections)) return false;
  if (
    !validateRuntimeSettings(
      spec.command,
      spec.args,
      spec.environmentVariables as Array<EnvironmentVariableOfAsset>,
      spec.workingDir,
    )
  )
    return false;
  return true;
}
function validateScope(
  scope: Scope,
  projectId: number | null | undefined,
  departmentId: string | null | undefined,
  clusterId: string | null | undefined,
): boolean {
  if (scope === Scope.Project) return !!projectId;
  else if (scope === Scope.Cluster) return !!clusterId;
  else if (scope === Scope.Department) return !!departmentId;
  else if (scope === Scope.Tenant) return !projectId && !departmentId;
  return false;
}
function validateName(name: string): boolean {
  // TODO :: Add api uniqueness validation.
  return isNotEmpty(name) && isFirstLetterIsLowerCase(name) && isLowerCaseAndNumbers(name);
}
function validateImage(image: string): boolean {
  // TODO :: add image regex
  return isNotEmpty(image);
}
function validateConnections(connections: Array<Connection> = []): boolean {
  return connections.every((tool: Connection) => {
    if (tool.isExternal) {
      return tool.name && tool.externalToolInfo?.externalUrl;
    } else {
      const { containerPort, connectionType, toolType } = tool.internalToolInfo as InternalToolInfo;
      return tool.name && containerPort && connectionType && toolType;
    }
  });
}
function validateUIConnections(connections: Array<IUIConnection> = []): boolean {
  return connections.every((tool: IUIConnection) => {
    return tool.containerPort + "" && tool.name && tool.toolType;
  });
}

function validateRuntimeSettings(
  command: string | null = null,
  args: string | null = null,
  envVariables: Array<EnvironmentVariableOfAsset> = [],
  workingDir: string | null = null,
): boolean {
  const isValidDirPath: boolean = workingDir === null || workingDir.length === 0 || isValidDirectoryPath(workingDir);
  const isValidEnvVar: boolean = envVariables.every((envVar: EnvironmentVariableOfAsset) =>
    isNotEmpty(envVar.name || ""),
  );
  return validCommandAndArgsLine(command) && validCommandAndArgsLine(args) && isValidEnvVar && isValidDirPath;
}

function prepareConnectionForRequest(connection: IUIConnection): Connection {
  const { name, toolType, connectionType, protocol, containerPort, isCustomPort, isCustomUrl, externalUrl } = connection;
  if (toolType === ToolType.Wandb || toolType === ToolType.Comet) {
    if (!name || !externalUrl) {
      throw new Error("Some of the fields are missing");
    }

    return {
      name,
      isExternal: true,
      externalToolInfo: {
        toolType,
        externalUrl: externalUrl,
      },
    };
  } else {
    const internalTool: Connection = {
      name,
      isExternal: false,
      internalToolInfo: {
        toolType: toolType,
        connectionType: connectionType,
        containerPort,
      } as InternalToolInfo,
    };
    if (connectionType === InternalConnectionType.ServingPort && internalTool.internalToolInfo) {
      internalTool.internalToolInfo.servingPortInfo = { protocol };
    }

    if (connectionType === InternalConnectionType.ExternalUrl) {
      internalTool.internalToolInfo = {
        ...internalTool.internalToolInfo,
        externalUrlInfo: {
          isCustomUrl,
          externalUrl: externalUrl || null,
        },
      } as InternalToolInfo;
    } else if (connectionType === InternalConnectionType.NodePort) {
      internalTool.internalToolInfo = {
        ...internalTool.internalToolInfo,
        nodePortInfo: {
          isCustomPort,
        },
      } as InternalToolInfo;
    }
    return internalTool;
  }
}

function prepareConnectionForUI(connection: Connection): IUIConnection | void {
  if (connection.isExternal) {
    return {
      id: makeId(),
      name: connection.name,
      toolType: connection.externalToolInfo?.toolType || ToolType.Wandb,
      externalUrl: connection.externalToolInfo?.externalUrl || null,
    };
  }
  if (connection.internalToolInfo) {
    const { connectionType, containerPort, toolType, externalUrlInfo, nodePortInfo, servingPortInfo } =
      connection.internalToolInfo;
    const resConnextion: IUIConnection = {
      id: makeId(),
      name: connection.name,
      connectionType: connectionType,
      containerPort: containerPort,
      toolType: toolType,
      isCustomUrl: externalUrlInfo?.isCustomUrl,
      isCustomPort: nodePortInfo?.isCustomPort,
    };

    if (servingPortInfo) {
      resConnextion.protocol = servingPortInfo.protocol;
    }
    return resConnextion;
  }
}
