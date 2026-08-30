import { environmentService } from "@/services/control-plane/environment.service/environment.service";
import type { IUIWorkloadEnvSectionModel } from "./environment-section.model";
import type {
  Connection,
  EnvironmentAsset,
  EnvironmentAssetSpec,
  EnvironmentVariableOfAsset,
} from "@/swagger-models/assets-service-client";
import type { IUIConnection } from "@/models/environment.model";

export const environmentSectionUtil = {
  getSpecificEnvFromEnvironment,
  mergeEnvironmentVariables,
};

function getSpecificEnvFromEnvironment(
  environment: EnvironmentAsset,
  envVarsFromDefault?: EnvironmentVariableOfAsset[] | null,
): IUIWorkloadEnvSectionModel {
  const spec: EnvironmentAssetSpec = environment.spec;
  const specificEnv: IUIWorkloadEnvSectionModel = {
    args: spec.args || "",
    command: spec.command || undefined,
    runAsUid: spec.runAsUid || null,
    runAsGid: spec.runAsGid || null,
    supplementalGroups: spec.supplementalGroups,
    environmentVariables: mergeEnvironmentVariables(spec.environmentVariables || [], envVarsFromDefault),
    connections:
      spec.connections
        ?.filter((c: Connection) => {
          return c.internalToolInfo?.nodePortInfo?.isCustomPort || c.internalToolInfo?.externalUrlInfo?.isCustomUrl;
        })
        .map((c: Connection) => {
          const uiConnection = environmentService.prepareConnectionForUI(c) as IUIConnection;
          const { name, externalUrl, nodePort } = uiConnection;
          return { name, externalUrl, nodePort };
        }) || [],
  };

  return specificEnv;
}

function mergeEnvironmentVariables(
  envVarFromEnvironment?: EnvironmentVariableOfAsset[] | null,
  envVarFromDefaults?: EnvironmentVariableOfAsset[] | null,
): EnvironmentVariableOfAsset[] {
  const fromEnvironment = envVarFromEnvironment || [];
  const fromEnvironmentNameMap: Record<string, string> = fromEnvironment.reduce((acc: Record<string, string>, item) => {
    item.name && (acc[item.name] = item.name);
    return acc;
  }, {});
  const fromDefaultsNoDuplicates = (envVarFromDefaults || []).filter(
    (env) => !fromEnvironmentNameMap[env.name as string],
  );

  return [...fromEnvironment, ...fromDefaultsNoDuplicates];
}
