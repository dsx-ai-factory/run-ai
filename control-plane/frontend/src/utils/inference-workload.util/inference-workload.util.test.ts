/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach } from "vitest";

import { ERunaiCredentialSource } from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";
import type { IAutoScaleData } from "@/components/section/compute-resource-section/auto-scale-section/auto-scale-section.model";
import { SCALE_TO_ZERO_RETENTION_IN_MINUTES_DEFAULT_VALUE } from "@/components/section/compute-resource-section/auto-scale-section/scale-down-section/scale-down-section.model";
import {
  DEFAULT_THRESHOLD_VALUE,
  METRIC_THROUGHPUT,
} from "@/components/section/compute-resource-section/auto-scale-section/threshold-metric-section/threshold-metric-section.model";
import {
  EAccessOptions,
  type IConnectionAccess,
} from "@/components/environment/connection-access/connection-access-modal/connection-access-modal.model";
import { ENimAccessModelMethod } from "@/components/workload-creation/hybrid-nim-model-section/nim-access-model/nim-access-model.model";

import type { IUIVolume } from "@/models/data-source.model";
import {
  ARGS_PREFIX_FOR_TGI,
  ARGS_PREFIX_FOR_VLLM,
  EInferenceArchitecture,
  EInferenceType,
  EServerType,
  EHuggingFaceEnvVariableNames,
  ENimEnvVariableNames,
  HF_HOME_DEFAULT_PATH,
  HOME_DEFAULT_PATH,
  LLM_IMAGE_FOR_HUGGING_FACE,
  LLM_IMAGE_FOR_TGI,
  MAX_REPLICAS_DEFAULT_VALUE,
  MIN_REPLICAS_DEFAULT_VALUE,
  NIM_WORKLOAD_LABEL,
  SERVING_PORT_SERVER_PORT,
} from "@/models/inference.model";
import type {
  IUIInferenceCreation,
  IUIStorageFields,
  IUISuperWorkloadCreationRequest,
  IUISuperWorkloadSpec,
  IUIWorkloadAssets,
  IUIWorkloadCreation,
  IUIWorkloadSpecificEnv,
} from "@/models/workload.model";

import {
  AssetKind,
  ServingPortAccessAuthorizationTypeEnum,
  type AutoScaling,
  type EnvironmentAsset,
  type EnvironmentVariable,
  type EnvironmentVariableOfAsset,
  type HostPathInstance,
  type InferenceCreationRequestV2,
  type InferenceSpecificRunParams,
  type InferenceV2,
  type NfsInstance,
  type PvcInstance,
  type SpecificRunParams,
} from "@/swagger-models/assets-service-client";
import type {
  Inference as InferenceFlat,
  InferenceSpecSpec,
  UnifiedPolicyInfoPerReplicaType,
  UnifiedWorkload,
} from "@/swagger-models/workloads-client";
import type { WorkspaceSpecSpecAllOfCompute } from "@/swagger-models/workloads-client/api";

import { deepCopy, omit } from "@/utils/common.util/common.util";
import { inferenceWorkloadUtil } from "@/utils/inference-workload.util/inference-workload.util";
import { workloadHybridUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import { workloadUtil } from "@/utils/workload.util/workload.util";

import {
  mockBaseInferenceWorkload,
  mockCluster,
  mockInferenceWorkload,
  mockInferenceWorkloadUI,
} from "./inference-workload.util.mocks";

vi.mock("@/utils/workload.util/workload.util");

describe("inference-workload.util", () => {
  const mockUIWorkloadCreation: IUIWorkloadCreation = {
    name: "",
    projectId: -1,
    namespace: "",
    clusterId: "",
    assets: {
      environment: "",
      compute: "",
    },
    specificEnv: {
      backoffLimit: 6,
    },
  };

  describe("getEmptyUINimInferenceCreation", () => {
    it("defaults model access to credential selection", () => {
      vi.mocked(workloadUtil.getEmptyUIWorkloadCreation).mockReturnValue(mockUIWorkloadCreation);

      const result = inferenceWorkloadUtil.getEmptyUINimInferenceCreation();

      expect(result.accessModelMethod).toBe(ENimAccessModelMethod.FromSecret);
      expect(result.accessFromStorage).toBe(false);
    });
  });

  describe("getHybridInferenceModel", () => {
    it("clears copied plaintext tokens while preserving credential references", () => {
      vi.spyOn(workloadHybridUtil, "getHybridBaseWorkloadModel").mockReturnValue({
        spec: {
          environmentVariables: [
            { name: EHuggingFaceEnvVariableNames.HF_TOKEN, value: "manual-hf" },
            { name: ENimEnvVariableNames.NgcApiKey, value: "manual-ngc" },
            {
              name: EHuggingFaceEnvVariableNames.HF_TOKEN,
              secret: { name: "hf-secret", key: EHuggingFaceEnvVariableNames.HF_TOKEN },
            },
            {
              name: ENimEnvVariableNames.NgcApiKey,
              userCredential: { name: "ngc-credential", key: ENimEnvVariableNames.NgcApiKey },
            },
            { name: "NORMAL_VAR", value: "visible" },
          ],
        },
      } as IUISuperWorkloadCreationRequest);

      const result = inferenceWorkloadUtil.getHybridInferenceModel({} as UnifiedWorkload, true);

      expect(result.spec.environmentVariables).toEqual([
        { name: EHuggingFaceEnvVariableNames.HF_TOKEN, value: "" },
        { name: ENimEnvVariableNames.NgcApiKey, value: "" },
        {
          name: EHuggingFaceEnvVariableNames.HF_TOKEN,
          secret: { name: "hf-secret", key: EHuggingFaceEnvVariableNames.HF_TOKEN },
        },
        {
          name: ENimEnvVariableNames.NgcApiKey,
          userCredential: { name: "ngc-credential", key: ENimEnvVariableNames.NgcApiKey },
        },
        { name: "NORMAL_VAR", value: "visible" },
      ]);
    });
  });

  describe("getEmptyUIInferenceCreation", () => {
    beforeEach(() => {
      vi.clearAllMocks();
    });

    it("should return an IUIInferenceCreation object with default values when no specificEnv is provided", () => {
      vi.mocked(workloadUtil.getEmptyUIWorkloadCreation).mockReturnValue(mockUIWorkloadCreation);

      const result = inferenceWorkloadUtil.getEmptyUIInferenceCreation();

      expect(result).toEqual({
        ...mockUIWorkloadCreation,
        specificEnv: {
          autoScaleData: {
            minReplicas: 1,
            maxReplicas: 1,
            thresholdMetric: undefined,
            thresholdValue: undefined,
            scaleToZeroRetentionSeconds: undefined,
          },
          servingPortAccess: {
            access: {
              authorizedUsers: null,
              authorizedGroups: null,
              accessOption: EAccessOptions.PUBLIC,
            },
            port: 8000,
            protocol: "http",
          },
        },
        serverType: EServerType.Nim,
        architecture: EInferenceArchitecture.Standard,
      });
    });

    it("should include specificEnv properties for asset based if provided ", () => {
      const specificEnv: IUIWorkloadSpecificEnv = {
        environmentVariables: [{ name: "TEST_VAR", value: "test" }],
      };

      vi.mocked(workloadUtil.getEmptyUIWorkloadCreation).mockReturnValue({
        ...mockUIWorkloadCreation,
        specificEnv,
      });

      const result = inferenceWorkloadUtil.getEmptyUIInferenceCreation(specificEnv);

      expect(result).toEqual({
        ...mockUIWorkloadCreation,
        specificEnv: {
          ...omit(specificEnv, ["backoffLimit"]),
          autoScaleData: {
            minReplicas: 1,
            maxReplicas: 1,
            thresholdMetric: undefined,
            thresholdValue: undefined,
            scaleToZeroRetentionSeconds: undefined,
          },
          servingPortAccess: {
            access: {
              authorizedUsers: null,
              authorizedGroups: null,
              accessOption: EAccessOptions.PUBLIC,
            },
            port: 8000,
            protocol: "http",
          },
        },
        serverType: EServerType.Nim,
        architecture: EInferenceArchitecture.Standard,
      });
    });

    it("should call workloadUtil.getEmptyUIWorkloadCreation with provided specificEnv", () => {
      const specificEnv: IUIWorkloadSpecificEnv = {
        environmentVariables: [{ name: "SAMPLE_VAR", value: "sample" }],
      };

      inferenceWorkloadUtil.getEmptyUIInferenceCreation(specificEnv);
      expect(workloadUtil.getEmptyUIWorkloadCreation).toHaveBeenCalledWith(specificEnv);
    });
  });

  describe("getHuggingFaceTokenValue", () => {
    it("should return the value of HF_TOKEN if it exists in environment variables", () => {
      const environmentVariables = [
        { name: "OTHER_VAR", value: "otherValue" },
        { name: EHuggingFaceEnvVariableNames.HF_TOKEN, value: "testTokenValue" },
      ];
      const result = inferenceWorkloadUtil.getHuggingFaceTokenValue(environmentVariables);
      expect(result).toBe("testTokenValue");
    });

    it("should return an empty string if HF_TOKEN does not exist in environment variables", () => {
      const environmentVariables = [
        { name: "OTHER_VAR", value: "otherValue" },
        { name: "ANOTHER_VAR", value: "anotherValue" },
      ];
      const result = inferenceWorkloadUtil.getHuggingFaceTokenValue(environmentVariables);
      expect(result).toBe("");
    });

    it("should return an empty string if environment variables array is empty", () => {
      const result = inferenceWorkloadUtil.getHuggingFaceTokenValue([]);
      expect(result).toBe("");
    });

    it("should return an empty string if environment variables array is undefined", () => {
      const result = inferenceWorkloadUtil.getHuggingFaceTokenValue(undefined);
      expect(result).toBe("");
    });

    it("should return an empty string if HF_TOKEN exists but has no value", () => {
      const environmentVariables = [{ name: EHuggingFaceEnvVariableNames.HF_TOKEN, value: "" }];
      const result = inferenceWorkloadUtil.getHuggingFaceTokenValue(environmentVariables);
      expect(result).toBe("");
    });
  });

  describe("getEnvironmentVariableValue", () => {
    it("should return the value of the provided key environment variable if it exists", () => {
      const environmentVariables: EnvironmentVariableOfAsset[] = [
        { name: "OTHER_ENV", value: "other-value" },
        { name: "NimModelProfile", value: "profile-value" },
      ];
      const result = inferenceWorkloadUtil.getEnvironmentVariableValue("NimModelProfile", environmentVariables);
      expect(result).toBe("profile-value");
    });

    it("should return an empty string if provider key is not present", () => {
      const environmentVariables: EnvironmentVariableOfAsset[] = [{ name: "OTHER_ENV", value: "other-value" }];
      const result = inferenceWorkloadUtil.getEnvironmentVariableValue("key_not_exist", environmentVariables);
      expect(result).toBe("");
    });

    it("should return an empty string if the input array is empty", () => {
      const result = inferenceWorkloadUtil.getEnvironmentVariableValue("some_key", []);
      expect(result).toBe("");
    });

    it("should return an empty string if no input is provided", () => {
      const result = inferenceWorkloadUtil.getEnvironmentVariableValue("some_key");
      expect(result).toBe("");
    });

    it("should handle multiple keys entries by returning the first match", () => {
      const environmentVariables: EnvironmentVariableOfAsset[] = [
        { name: "NimModelProfile", value: "first-value" },
        { name: "NimModelProfile", value: "second-value" },
      ];
      const result = inferenceWorkloadUtil.getEnvironmentVariableValue("NimModelProfile", environmentVariables);
      expect(result).toBe("first-value");
    });

    it("should handle an environment variable with an empty value", () => {
      const environmentVariables: EnvironmentVariableOfAsset[] = [{ name: "NimModelProfile", value: "" }];
      const result = inferenceWorkloadUtil.getEnvironmentVariableValue("NimModelProfile", environmentVariables);
      expect(result).toBe("");
    });
  });

  describe("inferCredentialSourceFromEnvVar", () => {
    it("returns undefined when env is undefined", () => {
      expect(inferenceWorkloadUtil.inferCredentialSourceFromEnvVar(undefined)).toBeUndefined();
    });

    it("returns undefined when env has a plain value (manual token)", () => {
      const env: EnvironmentVariable = { name: "HF_TOKEN", value: "tok" };
      expect(inferenceWorkloadUtil.inferCredentialSourceFromEnvVar(env)).toBeUndefined();
    });

    it("returns SharedSecret when env references a Kubernetes secret", () => {
      const env: EnvironmentVariable = {
        name: "HF_TOKEN",
        secret: { name: "sec-a", key: "HF_TOKEN" },
      };
      expect(inferenceWorkloadUtil.inferCredentialSourceFromEnvVar(env)).toBe(ERunaiCredentialSource.SharedSecret);
    });

    it("returns MyCredentials when env references a user credential", () => {
      const env: EnvironmentVariable = {
        name: "HF_TOKEN",
        userCredential: { name: "uc-a", key: "HF_TOKEN" },
      };
      expect(inferenceWorkloadUtil.inferCredentialSourceFromEnvVar(env)).toBe(ERunaiCredentialSource.MyCredentials);
    });
  });

  describe("excludeEnvironmentVariablesByName", () => {
    it("should exclude environment variables with the specified key name", () => {
      const envVariables: EnvironmentVariableOfAsset[] = [
        { name: "API_KEY", value: "12345" },
        { name: "DB_PASSWORD", value: "password" },
        { name: "API_KEY", value: "67890" },
      ];

      const result = inferenceWorkloadUtil.excludeEnvironmentVariablesByName("API_KEY", envVariables);

      expect(result).toEqual([{ name: "DB_PASSWORD", value: "password" }]);
    });

    it("should return the original array if the key name is not found", () => {
      const envVariables: EnvironmentVariableOfAsset[] = [
        { name: "DB_PASSWORD", value: "password" },
        { name: "HOST", value: "localhost" },
      ];

      const result = inferenceWorkloadUtil.excludeEnvironmentVariablesByName("API_KEY", envVariables);

      expect(result).toEqual(envVariables);
    });

    it("should return an empty array if all variables match the key name", () => {
      const envVariables: EnvironmentVariableOfAsset[] = [
        { name: "API_KEY", value: "12345" },
        { name: "API_KEY", value: "67890" },
      ];

      const result = inferenceWorkloadUtil.excludeEnvironmentVariablesByName("API_KEY", envVariables);

      expect(result).toEqual([]);
    });

    it("should return an empty array if the input array is empty", () => {
      const envVariables: EnvironmentVariableOfAsset[] = [];

      const result = inferenceWorkloadUtil.excludeEnvironmentVariablesByName("API_KEY", envVariables);

      expect(result).toEqual([]);
    });

    it("should handle the case where no key name is provided", () => {
      const envVariables: EnvironmentVariableOfAsset[] = [{ name: "DB_PASSWORD", value: "password" }];

      // Expect all variables to remain, as no key name matches `undefined`
      const result = inferenceWorkloadUtil.excludeEnvironmentVariablesByName(
        undefined as unknown as string,
        envVariables,
      );

      expect(result).toEqual(envVariables);
    });

    it("excludes variables whose names are in the keyNames array", () => {
      const environmentVariables: EnvironmentVariable[] = [
        { name: "VAR1", value: "value1" },
        { name: "VAR2", value: "value2" },
        { name: "VAR3", value: "value3" },
      ];

      const keyNames = ["VAR1", "VAR3"];

      const result = inferenceWorkloadUtil.excludeEnvironmentVariablesByName(keyNames, environmentVariables);

      expect(result).toEqual([{ name: "VAR2", value: "value2" }]);
    });

    it("returns all variables if no names match in the keyNames array", () => {
      const environmentVariables: EnvironmentVariable[] = [
        { name: "VAR1", value: "value1" },
        { name: "VAR2", value: "value2" },
        { name: "VAR3", value: "value3" },
      ];

      const keyNames = ["NON_EXISTENT_VAR"];

      const result = inferenceWorkloadUtil.excludeEnvironmentVariablesByName(keyNames, environmentVariables);

      expect(result).toEqual(environmentVariables);
    });

    it("returns an empty array if all variable names match in the keyNames array", () => {
      const environmentVariables: EnvironmentVariable[] = [
        { name: "VAR1", value: "value1" },
        { name: "VAR2", value: "value2" },
      ];

      const keyNames = ["VAR1", "VAR2"];

      const result = inferenceWorkloadUtil.excludeEnvironmentVariablesByName(keyNames, environmentVariables);

      expect(result).toEqual([]);
    });

    it("handles an empty environmentVariables array gracefully", () => {
      const environmentVariables: EnvironmentVariable[] = [];
      const keyNames = ["VAR1", "VAR2"];

      const result = inferenceWorkloadUtil.excludeEnvironmentVariablesByName(keyNames, environmentVariables);

      expect(result).toEqual([]);
    });

    it("handles a keyNames array with no elements gracefully", () => {
      const environmentVariables: EnvironmentVariable[] = [
        { name: "VAR1", value: "value1" },
        { name: "VAR2", value: "value2" },
      ];

      const keyNames: string[] = [];

      const result = inferenceWorkloadUtil.excludeEnvironmentVariablesByName(keyNames, environmentVariables);

      expect(result).toEqual(environmentVariables);
    });
  });

  describe("getInferenceRequestV2SpecificRunParams", () => {
    let defaultSpecificEnv: IUIWorkloadSpecificEnv;

    beforeEach(() => {
      vi.resetAllMocks();
      defaultSpecificEnv = {
        args: "arg1 arg2",
        command: "cmd",
        runAsUid: 1000,
        runAsGid: 2000,
        supplementalGroups: "3000,4000",
        environmentVariables: [{ name: "env1", value: "value1", exclude: false }],
        annotations: [{ name: "anno1", value: "value1", exclude: true }],
        labels: [{ name: "label1", value: "value1", exclude: false }],
        connections: null,
        nodeType: "my-node-type",
        nodePools: ["node-pool-1", "node-pool-2"],
        autoDeletionTimeAfterCompletionSeconds: 54,
        tolerations: undefined,
      };
    });

    it("should ignore backoffLimit and allowOverQuota fields if they provided", () => {
      const specificEnv: IUIWorkloadSpecificEnv = {
        ...defaultSpecificEnv,
        allowOverQuota: true,
        backoffLimit: 6,
      };

      vi.mocked(workloadUtil.processCommonSpecificEnv).mockReturnValue(omit(defaultSpecificEnv, ["servingPortAccess"]));
      const processCommonSpecificEnvSpied = vi.spyOn(workloadUtil, "processCommonSpecificEnv");

      const result: InferenceSpecificRunParams =
        inferenceWorkloadUtil.getInferenceRequestV2SpecificRunParams(specificEnv);

      const expectedSpecificEnv: InferenceSpecificRunParams = omit(defaultSpecificEnv, ["servingPortAccess"]);

      expect(result).toEqual(expectedSpecificEnv);
      expect(processCommonSpecificEnvSpied).toHaveBeenCalledWith(specificEnv, [
        "allowOverQuota",
        "backoffLimit",
        "autoScaleData",
        "servingPortAccess",
      ]);
    });

    it("should convert autoScaleData field to autoScaling when autoScaleData is provided", () => {
      const autoScaleData: IAutoScaleData = {
        minReplicas: 1,
        maxReplicas: 5,
        thresholdMetric: METRIC_THROUGHPUT,
        thresholdValue: 50,
        scaleToZeroRetentionSeconds: undefined,
      };
      const specificEnv: IUIWorkloadSpecificEnv = {
        ...defaultSpecificEnv,
        autoScaleData,
      };

      vi.mocked(workloadUtil.processCommonSpecificEnv).mockReturnValue(omit(defaultSpecificEnv, ["servingPortAccess"]));
      const processCommonSpecificEnvSpied = vi.spyOn(workloadUtil, "processCommonSpecificEnv");

      const result: InferenceSpecificRunParams =
        inferenceWorkloadUtil.getInferenceRequestV2SpecificRunParams(specificEnv);

      const expectedSpecificEnv: SpecificRunParams = {
        ...omit(defaultSpecificEnv, ["servingPortAccess"]),
        autoScaling: autoScaleData,
      };

      expect(result).toEqual(expectedSpecificEnv);
      expect(processCommonSpecificEnvSpied).toHaveBeenCalledWith(specificEnv, [
        "allowOverQuota",
        "backoffLimit",
        "autoScaleData",
        "servingPortAccess",
      ]);
    });

    it("should add servingPortAccess field when supportingServingPortAccess is true", () => {
      const specificEnv: IUIWorkloadSpecificEnv = {
        ...defaultSpecificEnv,
        servingPortAccess: {
          access: {
            authorizedUsers: null,
            authorizedGroups: ["rnd"],
            accessOption: EAccessOptions.GROUPS,
          },
        },
      };

      vi.mocked(workloadUtil.processCommonSpecificEnv).mockReturnValue(omit(defaultSpecificEnv, ["servingPortAccess"]));
      const processCommonSpecificEnvSpied = vi.spyOn(workloadUtil, "processCommonSpecificEnv");

      const result: InferenceSpecificRunParams = inferenceWorkloadUtil.getInferenceRequestV2SpecificRunParams(
        specificEnv,
        true,
      );

      expect(result).toEqual({
        ...specificEnv,
        servingPortAccess: {
          authorizationType: ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups,
          authorizedUsers: null,
          authorizedGroups: ["rnd"],
        },
      });

      expect(processCommonSpecificEnvSpied).toHaveBeenCalledWith(specificEnv, [
        "allowOverQuota",
        "backoffLimit",
        "autoScaleData",
        "servingPortAccess",
      ]);
    });

    it("should not add servingPortAccess field when supportingServingPortAccess is false", () => {
      const specificEnv: IUIWorkloadSpecificEnv = {
        ...defaultSpecificEnv,
        servingPortAccess: {
          access: {
            authorizedUsers: null,
            authorizedGroups: ["rnd"],
            accessOption: EAccessOptions.GROUPS,
          },
        },
      };

      vi.mocked(workloadUtil.processCommonSpecificEnv).mockReturnValue(omit(defaultSpecificEnv, ["servingPortAccess"]));
      const processCommonSpecificEnvSpied = vi.spyOn(workloadUtil, "processCommonSpecificEnv");

      const result: InferenceSpecificRunParams = inferenceWorkloadUtil.getInferenceRequestV2SpecificRunParams(
        specificEnv,
        false,
      );
      expect(result).not.toHaveProperty("servingPortAccess");
      expect(processCommonSpecificEnvSpied).toHaveBeenCalledWith(specificEnv, [
        "allowOverQuota",
        "backoffLimit",
        "autoScaleData",
        "servingPortAccess",
      ]);
    });
  });

  describe("convertInferenceToWorkloadUI", () => {
    it("should correctly convert workload to inference UI format with autoScaling and servingPortAccess", () => {
      const expectedInferenceUI: IUIInferenceCreation = {
        ...mockBaseInferenceWorkload,
        inferenceType: EInferenceType.AssetsBased,
        serverType: EServerType.Custom,
        templateId: undefined,
        specificEnv: {
          ...mockBaseInferenceWorkload.specificEnv,
          autoScaleData: {
            minReplicas: 1,
            maxReplicas: 1,
            thresholdMetric: undefined,
            thresholdValue: undefined,
            scaleToZeroRetentionSeconds: undefined,
          },
          servingPortAccess: {
            access: {
              accessOption: EAccessOptions.PUBLIC,
              authorizedGroups: undefined,
              authorizedUsers: undefined,
            },
          },
        },
        architecture: EInferenceArchitecture.Standard,
      };

      vi.mocked(workloadUtil.convertWorkloadToWorkloadUI).mockReturnValue(mockBaseInferenceWorkload);

      const result = inferenceWorkloadUtil.convertInferenceToWorkloadUI(
        mockInferenceWorkload,
        EInferenceType.AssetsBased,
        true,
      );
      expect(result).toEqual(expectedInferenceUI);
    });

    it("should correctly convert workload to inference UI format without servingPortAccess", () => {
      const mockInferenceUI: IUIInferenceCreation = {
        ...mockBaseInferenceWorkload,
        inferenceType: EInferenceType.AssetsBased,
        serverType: EServerType.Custom,
        templateId: undefined,
        specificEnv: {
          ...mockBaseInferenceWorkload.specificEnv,
          autoScaleData: {
            minReplicas: 1,
            maxReplicas: 1,
            thresholdMetric: undefined,
            thresholdValue: undefined,
            scaleToZeroRetentionSeconds: undefined,
          },
        },
        architecture: EInferenceArchitecture.Standard,
      };

      vi.mocked(workloadUtil.convertWorkloadToWorkloadUI).mockReturnValue(mockBaseInferenceWorkload);
      const result = inferenceWorkloadUtil.convertInferenceToWorkloadUI(
        mockInferenceWorkload,
        EInferenceType.AssetsBased,
        false,
      );

      expect(result).toEqual(mockInferenceUI);
    });

    it("should correctly convert workload to inference UI format without servingPortAccess when supportingServingPortAccess is false", () => {
      const mockInferenceUI: IUIInferenceCreation = {
        ...mockBaseInferenceWorkload,
        inferenceType: EInferenceType.AssetsBased,
        serverType: EServerType.Custom,
        templateId: undefined,
        specificEnv: {
          ...mockBaseInferenceWorkload.specificEnv,
          autoScaleData: {
            minReplicas: 1,
            maxReplicas: 1,
            thresholdMetric: undefined,
            thresholdValue: undefined,
            scaleToZeroRetentionSeconds: undefined,
          },
        },
        architecture: EInferenceArchitecture.Standard,
      };

      vi.mocked(workloadUtil.convertWorkloadToWorkloadUI).mockReturnValue(mockBaseInferenceWorkload);
      const result = inferenceWorkloadUtil.convertInferenceToWorkloadUI(
        {
          ...mockInferenceWorkload,
          specificRunParams: {
            ...mockBaseInferenceWorkload.specificEnv,
            autoScaling: {
              minReplicas: 1,
              maxReplicas: 1,
              thresholdMetric: undefined,
              thresholdValue: undefined,
              scaleToZeroRetentionSeconds: undefined,
            },
            servingPortAccess: {
              authorizationType: ServingPortAccessAuthorizationTypeEnum.Public,
              authorizedGroups: undefined,
              authorizedUsers: undefined,
            },
          },
        },
        EInferenceType.AssetsBased,
        false,
      );

      expect(result).toEqual(mockInferenceUI);
    });

    it("should correctly convert workload to inference UI format auto scaling is not provided", () => {
      const mockInferenceUI: IUIInferenceCreation = {
        ...mockBaseInferenceWorkload,
        inferenceType: EInferenceType.AssetsBased,
        serverType: EServerType.Custom,
        templateId: undefined,
        specificEnv: {
          ...mockBaseInferenceWorkload.specificEnv,
          autoScaleData: {
            minReplicas: 1,
            maxReplicas: 1,
            thresholdMetric: undefined,
            thresholdValue: undefined,
            scaleToZeroRetentionSeconds: undefined,
          },
        },
        architecture: EInferenceArchitecture.Standard,
      };
      vi.mocked(workloadUtil.convertWorkloadToWorkloadUI).mockReturnValue(mockBaseInferenceWorkload);

      const result = inferenceWorkloadUtil.convertInferenceToWorkloadUI(
        {
          ...mockInferenceWorkload,
          specificRunParams: {
            ...mockBaseInferenceWorkload.specificEnv,
            autoScaling: undefined,
            servingPortAccess: {
              authorizationType: ServingPortAccessAuthorizationTypeEnum.Public,
              authorizedGroups: undefined,
              authorizedUsers: undefined,
            },
          },
        },
        EInferenceType.AssetsBased,
        false,
      );

      expect(result).toEqual(mockInferenceUI);
    });

    it("should correctly convert workload to inference UI format when uiVolumes array provided", () => {
      const uiVolumes: Array<IUIVolume> = [
        {
          claimName: "claim-o5xg03jniz",
          path: "/home",
          ephemeral: false,
          claimInfo: {
            size: "1G",
            storageClass: "standard",
            accessModes: {
              readWriteOnce: true,
            },
            volumeMode: "Filesystem",
          },
        },
      ];

      const currentWorkload: InferenceV2 = {
        ...mockInferenceWorkload,
        specificRunParams: {
          ...mockBaseInferenceWorkload.specificEnv,
          autoScaling: undefined,
          servingPortAccess: {
            authorizationType: ServingPortAccessAuthorizationTypeEnum.Public,
            authorizedGroups: undefined,
            authorizedUsers: undefined,
          },
        },
      };

      const returnInferenceToWorkloadUI: IUIInferenceCreation = {
        ...mockBaseInferenceWorkload,
        assets: {
          ...mockBaseInferenceWorkload.assets,
          uiVolumes,
        },
        inferenceType: EInferenceType.AssetsBased,
        serverType: EServerType.Custom,
        templateId: undefined,
        architecture: EInferenceArchitecture.Standard,
      };

      vi.mocked(workloadUtil.convertWorkloadToWorkloadUI).mockReturnValue(returnInferenceToWorkloadUI);
      const convertWorkloadToWorkloadUISpied = vi.spyOn(workloadUtil, "convertWorkloadToWorkloadUI");

      const result = inferenceWorkloadUtil.convertInferenceToWorkloadUI(
        currentWorkload,
        EInferenceType.AssetsBased,
        false,
        uiVolumes,
      );

      expect(convertWorkloadToWorkloadUISpied).toHaveBeenCalledWith(currentWorkload, uiVolumes);
      expect(result.assets.uiVolumes).toEqual(uiVolumes);
    });

    it("should correctly convert workload to inference UI format when uiVolumes array not provided", () => {
      const currentWorkload: InferenceV2 = {
        ...mockInferenceWorkload,
        specificRunParams: {
          ...mockBaseInferenceWorkload.specificEnv,
          autoScaling: undefined,
          servingPortAccess: {
            authorizationType: ServingPortAccessAuthorizationTypeEnum.Public,
            authorizedGroups: undefined,
            authorizedUsers: undefined,
          },
        },
      };

      const returnInferenceToWorkloadUI: IUIInferenceCreation = {
        ...mockBaseInferenceWorkload,
        assets: {
          ...mockBaseInferenceWorkload.assets,
        },
        inferenceType: EInferenceType.AssetsBased,
        serverType: EServerType.Custom,
        templateId: undefined,
        architecture: EInferenceArchitecture.Standard,
      };

      vi.mocked(workloadUtil.convertWorkloadToWorkloadUI).mockReturnValue(returnInferenceToWorkloadUI);
      const convertWorkloadToWorkloadUISpied = vi.spyOn(workloadUtil, "convertWorkloadToWorkloadUI");

      const result = inferenceWorkloadUtil.convertInferenceToWorkloadUI(
        currentWorkload,
        EInferenceType.AssetsBased,
        false,
      );

      expect(convertWorkloadToWorkloadUISpied).toHaveBeenCalledWith(currentWorkload, undefined);
      expect(result.assets.uiVolumes).toBeUndefined();
    });

    it("should add modelName to inference UI format when it is provided", () => {
      const currentWorkload: InferenceV2 = {
        ...mockInferenceWorkload,
        specificRunParams: {
          ...mockBaseInferenceWorkload.specificEnv,
          autoScaling: undefined,
          servingPortAccess: {
            authorizationType: ServingPortAccessAuthorizationTypeEnum.Public,
            authorizedGroups: undefined,
            authorizedUsers: undefined,
          },
        },
      };

      const returnInferenceToWorkloadUI: IUIInferenceCreation = {
        ...mockBaseInferenceWorkload,
        assets: {
          ...mockBaseInferenceWorkload.assets,
        },
        inferenceType: EInferenceType.AssetsBased,
        serverType: EServerType.Custom,
        templateId: undefined,
        architecture: EInferenceArchitecture.Standard,
      };

      vi.mocked(workloadUtil.convertWorkloadToWorkloadUI).mockReturnValue(returnInferenceToWorkloadUI);
      const convertWorkloadToWorkloadUISpied = vi.spyOn(workloadUtil, "convertWorkloadToWorkloadUI");

      const result = inferenceWorkloadUtil.convertInferenceToWorkloadUI(
        currentWorkload,
        EInferenceType.AssetsBased,
        false,
        undefined,
        "someModel",
      );

      expect(convertWorkloadToWorkloadUISpied).toHaveBeenCalledWith(currentWorkload, undefined);
      expect(result.modelName).toBe("someModel");
    });

    it("should not add modelName to inference UI format when it is not provided", () => {
      const currentWorkload: InferenceV2 = {
        ...mockInferenceWorkload,
        specificRunParams: {
          ...mockBaseInferenceWorkload.specificEnv,
          autoScaling: undefined,
          servingPortAccess: {
            authorizationType: ServingPortAccessAuthorizationTypeEnum.Public,
            authorizedGroups: undefined,
            authorizedUsers: undefined,
          },
        },
      };

      const returnInferenceToWorkloadUI: IUIInferenceCreation = {
        ...mockBaseInferenceWorkload,
        assets: {
          ...mockBaseInferenceWorkload.assets,
        },
        inferenceType: EInferenceType.AssetsBased,
        serverType: EServerType.Custom,
        templateId: undefined,
        architecture: EInferenceArchitecture.Standard,
      };

      vi.mocked(workloadUtil.convertWorkloadToWorkloadUI).mockReturnValue(returnInferenceToWorkloadUI);
      const convertWorkloadToWorkloadUISpied = vi.spyOn(workloadUtil, "convertWorkloadToWorkloadUI");

      const result = inferenceWorkloadUtil.convertInferenceToWorkloadUI(
        currentWorkload,
        EInferenceType.AssetsBased,
        false,
        undefined,
      );

      expect(convertWorkloadToWorkloadUISpied).toHaveBeenCalledWith(currentWorkload, undefined);
      expect(result.modelName).toBeUndefined();
    });
  });

  describe("getInferenceCreationRequestV2", () => {
    const mockAssets: IUIWorkloadAssets = {
      environment: "4c2f601df492",
      compute: "832a901021e7",
      datasources: [{ id: "26f6314a0f0f", kind: AssetKind.Git }],
    };

    const mockWorkloadVolumes: Array<string> = ["9339", "4154"];

    const uiWorkloadCreation: IUIWorkloadCreation = {
      name: "i100",
      projectId: 100,
      namespace: "",
      clusterId: "ead7a57f",
      assets: mockAssets,
      specificEnv: {
        backoffLimit: 6,
        autoScaleData: mockInferenceWorkloadUI.specificEnv.autoScaleData,
        servingPortAccess: {
          access: {
            accessOption: EAccessOptions.PUBLIC,
          },
        },
      },
    };

    beforeEach(() => {
      vi.mocked(workloadUtil.processCommonSpecificEnv).mockReturnValue({
        args: null,
        command: null,
        runAsUid: undefined,
        runAsGid: undefined,
        supplementalGroups: undefined,
        autoScaling: {
          minReplicas: 1,
          maxReplicas: 2,
          thresholdMetric: undefined,
          thresholdValue: undefined,
          scaleToZeroRetentionSeconds: undefined,
        },
      });
    });

    it("should return a valid InferenceCreationRequestV2 object when all inputs are set", () => {
      vi.mocked(workloadUtil.getWorkloadAssetsFromUIWorkloadAssets).mockReturnValue({
        environment: mockAssets.environment || "",
        compute: mockAssets.compute,
        datasources: mockAssets.datasources,
        workloadVolumes: mockWorkloadVolumes,
      });

      const getWorkloadAssetsFromUIWorkloadAssetsSpied = vi.spyOn(workloadUtil, "getWorkloadAssetsFromUIWorkloadAssets");
      const result = inferenceWorkloadUtil.getInferenceCreationRequestV2(uiWorkloadCreation, true, mockWorkloadVolumes);

      expect(result).toEqual({
        meta: {
          name: uiWorkloadCreation.name,
          projectId: uiWorkloadCreation.projectId.toString(),
          clusterId: uiWorkloadCreation.clusterId,
        },
        assets: {
          environment: mockAssets.environment,
          compute: mockAssets.compute,
          datasources: mockAssets.datasources,
          workloadVolumes: mockWorkloadVolumes,
        },
        specificRunParams: {
          args: null,
          autoScaling: {
            maxReplicas: 1,
            minReplicas: 1,
            scaleToZeroRetentionSeconds: undefined,
            thresholdMetric: undefined,
            thresholdValue: undefined,
          },
          command: null,
          runAsGid: undefined,
          runAsUid: undefined,
          supplementalGroups: undefined,
          servingPortAccess: {
            authorizationType: undefined,
            authorizedGroups: undefined,
            authorizedUsers: undefined,
          },
        },
      });

      expect(getWorkloadAssetsFromUIWorkloadAssetsSpied).toHaveBeenCalledWith(mockAssets, mockWorkloadVolumes);
    });

    it("should throw an error if assets.environment is missing", () => {
      const invalidInference = {
        ...uiWorkloadCreation,
        assets: { environment: "", compute: "compute-1" },
      };

      expect(() => inferenceWorkloadUtil.getInferenceCreationRequestV2(invalidInference)).toThrow(
        "Environment details are missing",
      );
    });

    it("should handle default workloadVolumes when not provided", () => {
      vi.mocked(workloadUtil.getWorkloadAssetsFromUIWorkloadAssets).mockReturnValue({
        environment: mockAssets.environment || "",
        compute: mockAssets.compute,
        datasources: mockAssets.datasources,
      });

      const getWorkloadAssetsFromUIWorkloadAssetsSpied = vi.spyOn(workloadUtil, "getWorkloadAssetsFromUIWorkloadAssets");
      const result = inferenceWorkloadUtil.getInferenceCreationRequestV2(uiWorkloadCreation, false);

      expect(result.assets.workloadVolumes).toBeUndefined();
      expect(getWorkloadAssetsFromUIWorkloadAssetsSpied).toHaveBeenCalledWith(mockAssets, undefined);
    });

    it("should exclude servingPortAccess if supportingServingPortAccess is false", () => {
      vi.mocked(workloadUtil.getWorkloadAssetsFromUIWorkloadAssets).mockReturnValue({
        environment: mockAssets.environment || "",
        compute: mockAssets.compute,
        datasources: mockAssets.datasources,
      });

      vi.mocked(workloadUtil.processCommonSpecificEnv).mockReturnValue({
        args: null,
        command: null,
        runAsUid: undefined,
        runAsGid: undefined,
        supplementalGroups: undefined,
        autoScaling: {
          minReplicas: 1,
          maxReplicas: 2,
          thresholdMetric: undefined,
          thresholdValue: undefined,
          scaleToZeroRetentionSeconds: undefined,
        },
      });

      const result: InferenceCreationRequestV2 = inferenceWorkloadUtil.getInferenceCreationRequestV2(uiWorkloadCreation);

      expect(result.specificRunParams).toBeDefined();
      expect(result.specificRunParams?.servingPortAccess).toBeUndefined();
      expect(workloadUtil.processCommonSpecificEnv).toHaveBeenCalledWith(
        uiWorkloadCreation.specificEnv,
        expect.any(Array),
      );
    });
  });

  describe("_getAuthorizationType", () => {
    it("should return Public when servingPortAccess is null", () => {
      const result = inferenceWorkloadUtil._getAuthorizationType(null);
      expect(result).toBe(undefined);
    });

    it("should return Public when servingPortAccess is undefined", () => {
      const result = inferenceWorkloadUtil._getAuthorizationType(undefined);
      expect(result).toBe(undefined);
    });

    it("should return Public when accessOption is EAccessOptions.PUBLIC", () => {
      const result = inferenceWorkloadUtil._getAuthorizationType({ accessOption: EAccessOptions.PUBLIC });
      expect(result).toBe(undefined);
    });

    it("should return AuthenticatedUsers when accessOption is EAccessOptions.EVERYONE", () => {
      const result = inferenceWorkloadUtil._getAuthorizationType({ accessOption: EAccessOptions.EVERYONE });
      expect(result).toBe(ServingPortAccessAuthorizationTypeEnum.AuthenticatedUsers);
    });

    it("should return AuthorizedUsersOrGroups when accessOption is EAccessOptions.GROUPS", () => {
      const result = inferenceWorkloadUtil._getAuthorizationType({ accessOption: EAccessOptions.GROUPS });
      expect(result).toBe(ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups);
    });

    it("should return AuthorizedUsersOrGroups when accessOption is EAccessOptions.SPECIFIC_USERS", () => {
      const result = inferenceWorkloadUtil._getAuthorizationType({ accessOption: EAccessOptions.SPECIFIC_USERS });
      expect(result).toBe(ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups);
    });
  });

  describe("_servingPortAccess", () => {
    beforeEach(() => {
      vi.clearAllMocks();
    });

    it("should return default values when servingPortAccess is null", () => {
      const result = inferenceWorkloadUtil.servingPortAccessAuthorization(null);

      expect(result).toEqual({
        authorizationType: undefined,
        authorizedGroups: undefined,
        authorizedUsers: undefined,
      });
    });

    it("should return default values when servingPortAccess is undefined", () => {
      const result = inferenceWorkloadUtil.servingPortAccessAuthorization(undefined);

      expect(result).toEqual({
        authorizationType: undefined,
        authorizedGroups: undefined,
        authorizedUsers: undefined,
      });
    });

    it("should return correct values when authorizedGroups is provided", () => {
      const servingPortAccess: IConnectionAccess = {
        accessOption: EAccessOptions.GROUPS,
        authorizedGroups: ["group1", "group2"],
        authorizedUsers: null,
      };

      const result = inferenceWorkloadUtil.servingPortAccessAuthorization(servingPortAccess);

      expect(result).toEqual({
        authorizationType: ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups,
        authorizedGroups: ["group1", "group2"],
        authorizedUsers: null,
      });
    });

    it("should return correct values when authorizedUsers is provided", () => {
      const servingPortAccess: IConnectionAccess = {
        accessOption: EAccessOptions.SPECIFIC_USERS,
        authorizedGroups: null,
        authorizedUsers: ["user1", "user2"],
      };

      const result = inferenceWorkloadUtil.servingPortAccessAuthorization(servingPortAccess);

      expect(result).toEqual({
        authorizationType: ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups,
        authorizedGroups: null,
        authorizedUsers: ["user1", "user2"],
      });
    });

    it("should return correct values when EVERYONE is provided", () => {
      const servingPortAccess: IConnectionAccess = {
        accessOption: EAccessOptions.EVERYONE,
      };

      const result = inferenceWorkloadUtil.servingPortAccessAuthorization(servingPortAccess);

      expect(result).toEqual({
        authorizationType: ServingPortAccessAuthorizationTypeEnum.AuthenticatedUsers,
        authorizedGroups: undefined,
        authorizedUsers: undefined,
      });
    });
  });

  describe("_getServingPortAccessOption", () => {
    it("should return PUBLIC when servingPortAccess is null", () => {
      const result = inferenceWorkloadUtil._getServingPortAccessOption(null);
      expect(result).toBe(EAccessOptions.PUBLIC);
    });

    it("should return PUBLIC when servingPortAccess is undefined", () => {
      const result = inferenceWorkloadUtil._getServingPortAccessOption(undefined);
      expect(result).toBe(EAccessOptions.PUBLIC);
    });

    it("should return PUBLIC when authorizationType is Public", () => {
      const result = inferenceWorkloadUtil._getServingPortAccessOption({
        authorizationType: ServingPortAccessAuthorizationTypeEnum.Public,
      });
      expect(result).toBe(EAccessOptions.PUBLIC);
    });

    it("should return EVERYONE when authorizationType is AuthenticatedUsers", () => {
      const result = inferenceWorkloadUtil._getServingPortAccessOption({
        authorizationType: ServingPortAccessAuthorizationTypeEnum.AuthenticatedUsers,
      });
      expect(result).toBe(EAccessOptions.EVERYONE);
    });

    it("should return GROUPS when authorizationType is AuthorizedUsersOrGroups and authorizedGroups is provided", () => {
      const result = inferenceWorkloadUtil._getServingPortAccessOption({
        authorizationType: ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups,
        authorizedGroups: ["group1", "group2"],
      });
      expect(result).toBe(EAccessOptions.GROUPS);
    });

    it("should return SPECIFIC_USERS when authorizationType is AuthorizedUsersOrGroups and authorizedGroups is not provided", () => {
      const result = inferenceWorkloadUtil._getServingPortAccessOption({
        authorizationType: ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups,
      });
      expect(result).toBe(EAccessOptions.SPECIFIC_USERS);
    });
  });

  describe("servingPortInternalToolInfo", () => {
    beforeEach(() => {
      vi.clearAllMocks();
    });

    it("should return default values when servingPortAccess is null", () => {
      const result = inferenceWorkloadUtil.servingPortInternalToolInfo(null);

      expect(result).toEqual({
        accessOption: EAccessOptions.PUBLIC,
        authorizedGroups: undefined,
        authorizedUsers: undefined,
      });
    });

    it("should return default values when servingPortAccess is undefined", () => {
      const result = inferenceWorkloadUtil.servingPortInternalToolInfo(undefined);

      expect(result).toEqual({
        accessOption: EAccessOptions.PUBLIC,
        authorizedGroups: undefined,
        authorizedUsers: undefined,
      });
    });

    it("should return correct values when authorizedGroups is provided", () => {
      const servingPortAccess = {
        authorizationType: ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups,
        authorizedGroups: ["group1", "group2"],
        authorizedUsers: null,
      };

      const result = inferenceWorkloadUtil.servingPortInternalToolInfo(servingPortAccess);

      expect(result).toEqual({
        accessOption: EAccessOptions.GROUPS,
        authorizedGroups: ["group1", "group2"],
        authorizedUsers: null,
      });
    });

    it("should return correct values when authorizedUsers is provided", () => {
      const servingPortAccess = {
        authorizationType: ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups,
        authorizedGroups: null,
        authorizedUsers: ["user1", "user2"],
      };

      const result = inferenceWorkloadUtil.servingPortInternalToolInfo(servingPortAccess);

      expect(result).toEqual({
        accessOption: EAccessOptions.SPECIFIC_USERS,
        authorizedGroups: null,
        authorizedUsers: ["user1", "user2"],
      });
    });

    it("should return correct values when AuthenticatedUsers is provided", () => {
      const servingPortAccess = {
        authorizationType: ServingPortAccessAuthorizationTypeEnum.AuthenticatedUsers,
      };

      const result = inferenceWorkloadUtil.servingPortInternalToolInfo(servingPortAccess);
      expect(result).toEqual({
        accessOption: EAccessOptions.EVERYONE,
        authorizedGroups: undefined,
        authorizedUsers: undefined,
      });
    });
  });

  describe("_getFlatWorkloadData", () => {
    const flatCompute: WorkspaceSpecSpecAllOfCompute = {
      gpuDevicesRequest: 0,
      cpuCoreRequest: 0.1,
      cpuMemoryRequest: "100M",
    };

    const environment: EnvironmentAsset = {
      meta: {
        name: "e100",
        scope: "tenant",
        id: "",
        kind: "environment",
        createdBy: "",
        createdAt: "",
        updatedBy: "",
        updatedAt: "",
      },
      spec: {
        image: "test-image",
        imagePullPolicy: "Always",
        workingDir: "/home",
        createHomeDir: true,
      },
    };

    it("should generate a workload request with specificEnv and environment", () => {
      const workload: IUIInferenceCreation = {
        name: "test-workload",
        projectId: 1,
        clusterId: "cluster-1",
        specificEnv: {
          command: "python",
          args: "script.py",
          nodeType: "GPU",
          annotations: [{ name: "annotation1", value: "value1", exclude: false }],
          labels: [{ name: "label1", value: "value1" }],
        },
        inferenceType: EInferenceType.Nim,
        serverType: EServerType.Nim,
        architecture: EInferenceArchitecture.Standard,
        namespace: "",
        assets: {
          environment: null,
          compute: null,
        },
      };

      const result = inferenceWorkloadUtil._getFlatWorkloadData(workload, flatCompute, environment);

      expect(result).toEqual({
        name: "test-workload",
        projectId: "1",
        clusterId: "cluster-1",
        spec: {
          command: "python",
          args: "script.py",
          nodeType: "GPU",
          nodePools: undefined,
          podAffinity: undefined,
          environmentVariables: undefined,
          annotations: [{ name: "annotation1", value: "value1", exclude: false }],
          labels: [{ name: "label1", value: "value1" }],
          tolerations: undefined,
          compute: flatCompute,
          image: "test-image",
          imagePullPolicy: "Always",
          workingDir: "/home",
          createHomeDir: true,
        },
      });
    });

    it("should generate a workload request without specificEnv or environment", () => {
      const workload: IUIInferenceCreation = {
        name: "test-workload",
        projectId: 2,
        clusterId: "cluster-2",
        inferenceType: EInferenceType.Nim,
        serverType: EServerType.Nim,
        namespace: "",
        assets: {
          environment: null,
          compute: null,
        },
        specificEnv: {},
        architecture: EInferenceArchitecture.Standard,
      };

      const result = inferenceWorkloadUtil._getFlatWorkloadData(workload, flatCompute);

      expect(result).toEqual({
        name: "test-workload",
        projectId: "2",
        clusterId: "cluster-2",
        spec: {
          command: undefined,
          args: undefined,
          nodeType: undefined,
          nodePools: undefined,
          podAffinity: undefined,
          environmentVariables: undefined,
          annotations: undefined,
          labels: undefined,
          tolerations: undefined,
          compute: flatCompute,
        },
      });
    });

    it("should handle missing fields gracefully", () => {
      const workload: IUIInferenceCreation = {
        name: "test-workload",
        projectId: 3,
        clusterId: "cluster-3",
        specificEnv: {},
        inferenceType: EInferenceType.AssetsBased,
        serverType: EServerType.Custom,
        namespace: "",
        assets: {
          environment: null,
          compute: null,
        },
        architecture: EInferenceArchitecture.Standard,
      };

      const result = inferenceWorkloadUtil._getFlatWorkloadData(workload, flatCompute);

      expect(result).toEqual({
        name: "test-workload",
        projectId: "3",
        clusterId: "cluster-3",
        spec: {
          command: undefined,
          args: undefined,
          nodeType: undefined,
          nodePools: undefined,
          podAffinity: undefined,
          environmentVariables: undefined,
          annotations: undefined,
          labels: undefined,
          tolerations: undefined,
          compute: flatCompute,
        },
      });
    });
  });

  describe("_getFlatInferenceWorkloadData", () => {
    const flatCompute: WorkspaceSpecSpecAllOfCompute = {
      gpuDevicesRequest: 0,
      cpuCoreRequest: 0.1,
      cpuMemoryRequest: "100M",
    };

    const environment: EnvironmentAsset = {
      meta: {
        name: "e100",
        scope: "tenant",
        id: "",
        kind: "environment",
        createdBy: "",
        createdAt: "",
        updatedBy: "",
        updatedAt: "",
      },
      spec: {
        image: "test-image",
        imagePullPolicy: "Always",
        workingDir: "/home",
        createHomeDir: true,
      },
    };

    it("should add autoscaling data and serving port", () => {
      const workload: IUIInferenceCreation = {
        name: "test-workload",
        projectId: 1,
        clusterId: "cluster-1",
        serverType: EServerType.Custom,
        specificEnv: {
          autoScaleData: {
            minReplicas: 1,
            maxReplicas: 10,
            thresholdMetric: undefined,
            thresholdValue: undefined,
            scaleToZeroRetentionSeconds: undefined,
          },
          servingPortAccess: { access: { accessOption: EAccessOptions.PUBLIC } },
        },
        inferenceType: EInferenceType.Nim,
        architecture: EInferenceArchitecture.Standard,
        namespace: "",
        assets: {
          environment: null,
          compute: null,
        },
      };

      const result = inferenceWorkloadUtil._getFlatInferenceWorkloadData(workload, flatCompute, environment);
      expect(result).toStrictEqual({
        name: workload.name,
        projectId: workload.projectId.toString(),
        clusterId: workload.clusterId,
        spec: {
          command: undefined,
          args: undefined,
          nodeType: undefined,
          nodePools: undefined,
          podAffinity: undefined,
          environmentVariables: undefined,
          annotations: undefined,
          labels: undefined,
          tolerations: undefined,
          compute: flatCompute,
          image: "test-image",
          imagePullPolicy: "Always",
          workingDir: "/home",
          createHomeDir: true,
          autoscaling: {
            minReplicas: 1,
            maxReplicas: 10,
            metric: undefined,
            metricThreshold: undefined,
            scaleToZeroRetentionSeconds: undefined,
          },
        },
      });
    });

    it("should handle missing environment gracefully", () => {
      const workload: IUIInferenceCreation = {
        name: "test-workload",
        projectId: 4,
        clusterId: "cluster-4",
        specificEnv: {
          autoScaleData: {
            minReplicas: 1,
            maxReplicas: 10,
            thresholdMetric: undefined,
            thresholdValue: undefined,
            scaleToZeroRetentionSeconds: undefined,
          },
          servingPortAccess: { access: { accessOption: EAccessOptions.PUBLIC } },
        },
        inferenceType: EInferenceType.Nim,
        architecture: EInferenceArchitecture.Standard,
        namespace: "",
        assets: {
          environment: null,
          compute: null,
        },
        serverType: EServerType.VLLM,
      };

      const result = inferenceWorkloadUtil._getFlatInferenceWorkloadData(workload, flatCompute);

      expect(result.spec).not.toHaveProperty("image");
      expect(result.spec).not.toHaveProperty("imagePullPolicy");
      expect(result.spec).not.toHaveProperty("workingDir");
      expect(result.spec).not.toHaveProperty("createHomeDir");
    });
  });

  describe("_addNimEnvironmentVariables", () => {
    it("should add default NIM environment variables when input is empty", () => {
      const input: EnvironmentVariable[] = [];
      const result = inferenceWorkloadUtil._addNimEnvironmentVariables(input);
      expect(result).toEqual([
        { name: ENimEnvVariableNames.OutlinesCacheDir, value: "/tmp/outlines" },
        { name: ENimEnvVariableNames.NimServerPort, value: "8000" },
        { name: ENimEnvVariableNames.NimJsonlLogging, value: "1" },
        { name: ENimEnvVariableNames.NimLogLevel, value: "INFO" },
      ]);
    });

    it("should replace existing NIM environment variables", () => {
      const input: EnvironmentVariable[] = [
        { name: ENimEnvVariableNames.OutlinesCacheDir, value: "/custom/path" },
        { name: ENimEnvVariableNames.NimServerPort, value: "9000" },
      ];

      const result = inferenceWorkloadUtil._addNimEnvironmentVariables(input);
      expect(result).toEqual([
        { name: ENimEnvVariableNames.OutlinesCacheDir, value: "/tmp/outlines" },
        { name: ENimEnvVariableNames.NimServerPort, value: "8000" },
        { name: ENimEnvVariableNames.NimJsonlLogging, value: "1" },
        { name: ENimEnvVariableNames.NimLogLevel, value: "INFO" },
      ]);
    });

    it("should retain unrelated environment variables", () => {
      const input: EnvironmentVariable[] = [{ name: "UNRELATED_VAR", value: "value" }];

      const result = inferenceWorkloadUtil._addNimEnvironmentVariables(input);

      expect(result).toEqual([
        { name: "UNRELATED_VAR", value: "value" },
        { name: ENimEnvVariableNames.OutlinesCacheDir, value: "/tmp/outlines" },
        { name: ENimEnvVariableNames.NimServerPort, value: "8000" },
        { name: ENimEnvVariableNames.NimJsonlLogging, value: "1" },
        { name: ENimEnvVariableNames.NimLogLevel, value: "INFO" },
      ]);
    });
  });

  describe("_fillDatasourceData", () => {
    const nfsSource: NfsInstance = { path: "/data/nfs", name: "nfs100", readOnly: true };
    const kind = "nfs";
    const envVarName = "ENV_HOME";

    it("should add the data source to the correct storage kind", () => {
      const nimWorkloadSpec: InferenceSpecSpec = { storage: {} };

      const result = inferenceWorkloadUtil._fillDatasourceData(nimWorkloadSpec, nfsSource, kind, envVarName);

      expect(result.storage?.nfs).toEqual([nfsSource]);
    });

    it("should replace existing environment variable with the new one", () => {
      const nimWorkloadSpec: InferenceSpecSpec = {
        environmentVariables: [{ name: envVarName, value: "/old/path" }],
      };

      const result = inferenceWorkloadUtil._fillDatasourceData(nimWorkloadSpec, nfsSource, kind, envVarName);
      expect(result.environmentVariables).toEqual([{ name: envVarName, value: "/data/nfs" }]);
    });

    it("should not add env variable if no datasource is found", () => {
      const nimWorkloadSpec: InferenceSpecSpec = { environmentVariables: [] };
      const result = inferenceWorkloadUtil._fillDatasourceData(nimWorkloadSpec, undefined, "nfs", envVarName);

      expect(result.environmentVariables).toEqual([]);
    });

    it("should not add storage if no kind provided", () => {
      const nimWorkloadSpec: InferenceSpecSpec = { environmentVariables: [] };
      const result = inferenceWorkloadUtil._fillDatasourceData(nimWorkloadSpec, nfsSource, undefined, envVarName);

      expect(result.storage).toBeUndefined();
    });

    it("should add env variable from pvc path", () => {
      const nimWorkloadSpec: InferenceSpecSpec = { environmentVariables: [] };
      const pvcSource: PvcInstance = { path: "/data/pvc", name: "pvc100", readOnly: true };
      const kind = "pvc";
      const result = inferenceWorkloadUtil._fillDatasourceData(nimWorkloadSpec, pvcSource, kind, envVarName);

      expect(result.environmentVariables).toEqual([{ name: envVarName, value: "/data/pvc" }]);
    });

    it("should add env variable from hostPath path", () => {
      const nimWorkloadSpec: InferenceSpecSpec = { environmentVariables: [] };
      const hostSource: HostPathInstance = { path: "/data/host", name: "host100", readOnly: true };
      const kind = "hostPath";

      const result = inferenceWorkloadUtil._fillDatasourceData(nimWorkloadSpec, hostSource, kind, envVarName);

      expect(result.environmentVariables).toEqual([{ name: envVarName, value: "/data/host" }]);
    });

    it("should replace new data source with the same kind existing storage", () => {
      const nimWorkloadSpec: InferenceSpecSpec = { storage: { nfs: [{ name: "nfs100", path: "/data/nfs" }] } };

      const result = inferenceWorkloadUtil._fillDatasourceData(nimWorkloadSpec, nfsSource, kind, envVarName);
      expect(result.storage?.nfs).toEqual([nfsSource]);
    });

    it("should replace new data source with different kind existing storage", () => {
      const nimWorkloadSpec: InferenceSpecSpec = {
        storage: { hostPath: [{ name: "host-path", path: "/data/host-path" }] },
      };

      const result = inferenceWorkloadUtil._fillDatasourceData(nimWorkloadSpec, nfsSource, kind, envVarName);
      expect(result.storage).toHaveProperty("nfs");
      expect(result.storage).not.toHaveProperty("hostPath");
    });
  });

  describe("getHFInferenceWorkloadCreationRequest", () => {
    const modelName = "some-model";
    const workload: IUIInferenceCreation = {
      name: "test-workload",
      projectId: 1,
      clusterId: "cluster-1",
      specificEnv: {
        command: "python",
        args: "script.py",
        nodeType: "GPU",
        environmentVariables: [{ name: "SOME_VAR", value: "test" }],
        annotations: [{ name: "annotation1", value: "value1", exclude: false }],
        labels: [{ name: "label1", value: "value1" }],
      },
      inferenceType: EInferenceType.Nim,
      architecture: EInferenceArchitecture.Standard,
      namespace: "",
      assets: {
        environment: null,
        compute: null,
      },
      modelName,
      serverType: EServerType.VLLM,
      image: LLM_IMAGE_FOR_HUGGING_FACE,
    };
    const token = "hf_someToken";
    const flatCompute: WorkspaceSpecSpecAllOfCompute = {
      gpuDevicesRequest: 0,
      cpuCoreRequest: 0.1,
      cpuMemoryRequest: "100M",
    };

    it("should set hugging-face specific image", () => {
      const result = inferenceWorkloadUtil.getHFInferenceWorkloadCreationRequest(workload, flatCompute);
      expect(result.spec?.image).toBe(LLM_IMAGE_FOR_HUGGING_FACE);
    });

    it("should set servingPort parameters correctly", () => {
      const result = inferenceWorkloadUtil.getHFInferenceWorkloadCreationRequest(workload, flatCompute);
      expect(result.spec?.servingPort).toEqual({ container: SERVING_PORT_SERVER_PORT, protocol: "http" });
    });

    it("should throw an error if serverType is missing", () => {
      const workloadWithoutServerType = { ...workload, serverType: undefined } as unknown as IUIInferenceCreation;

      expect(() =>
        inferenceWorkloadUtil.getHFInferenceWorkloadCreationRequest(workloadWithoutServerType, flatCompute),
      ).toThrow("Missing server type: The workload object must include a valid 'serverType' property.");
    });

    it("should set the correct model flag based on server type", () => {
      const tgiWorkload = { ...workload, serverType: EServerType.TGI, image: "tgi-image" };
      const vllmWorkload = { ...workload, serverType: EServerType.VLLM, image: "inference-image" };

      const tgiRequest = inferenceWorkloadUtil.getHFInferenceWorkloadCreationRequest(tgiWorkload, flatCompute);
      expect(tgiRequest.spec?.args).toContain(`--model-id ${modelName} ${workload.specificEnv.args}`);

      const vllmRequest = inferenceWorkloadUtil.getHFInferenceWorkloadCreationRequest(vllmWorkload, flatCompute);
      expect(vllmRequest.spec?.args).toContain(`--model ${modelName} ${workload.specificEnv.args}`);
    });

    it("should include the correct server type image", () => {
      const currWorkload = { ...workload, serverType: EServerType.TGI, image: "custom-image" };

      const result = inferenceWorkloadUtil.getHFInferenceWorkloadCreationRequest(currWorkload, flatCompute);
      expect(result.spec?.image).toBe("custom-image");
    });

    it("should set hf token environment variable when token is provided", () => {
      const result = inferenceWorkloadUtil.getHFInferenceWorkloadCreationRequest(
        { ...workload, modelName, token },
        flatCompute,
      );

      const runaiModelParam = result.spec?.environmentVariables?.find(
        (env: EnvironmentVariable) => env.name === EHuggingFaceEnvVariableNames.HF_TOKEN,
      );

      expect(runaiModelParam?.name).toBe(EHuggingFaceEnvVariableNames.HF_TOKEN);
      expect(runaiModelParam?.value).toBe(token);
    });

    it("should set hf token environment variable when generic secret is provided", () => {
      const modelName = "some-model";
      const genericSecretName = "hf-secret-name";

      const result = inferenceWorkloadUtil.getHFInferenceWorkloadCreationRequest(
        { ...workload, modelName },
        flatCompute,
        { genericSecretName },
      );

      const runaiModelParam = result.spec?.environmentVariables?.find(
        (env: EnvironmentVariable) => env.name === EHuggingFaceEnvVariableNames.HF_TOKEN,
      );

      expect(runaiModelParam?.name).toBe(EHuggingFaceEnvVariableNames.HF_TOKEN);
      expect(runaiModelParam?.secret?.name).toBe(genericSecretName);
      expect(runaiModelParam?.secret?.key).toBe(EHuggingFaceEnvVariableNames.HF_TOKEN);
    });

    it("should fill datasource data if a dataSource is provided in hugging face", () => {
      const nfsSource: NfsInstance = { path: "/data/nfs", name: "nfs100", readOnly: true };
      const kind = "nfs";

      const result = inferenceWorkloadUtil.getHFInferenceWorkloadCreationRequest(workload, flatCompute, {
        storage: { dataSource: nfsSource, kind },
      });
      expect(Object.keys(result.spec?.storage || {}).length === 1).toBeTruthy();
      expect(result.spec?.storage?.nfs).toBeTruthy();

      const hfHomeEnvPath = result.spec?.environmentVariables?.find(
        (env) => env.name === EHuggingFaceEnvVariableNames.HF_HOME,
      );
      expect(hfHomeEnvPath?.value).toBeDefined();
      expect(hfHomeEnvPath?.value).toBe(nfsSource.path);
      expect(hfHomeEnvPath?.value).not.toBe(HF_HOME_DEFAULT_PATH);
    });

    it("should handle the absence of dataSource gracefully in hugging face", () => {
      const result = inferenceWorkloadUtil.getHFInferenceWorkloadCreationRequest(workload, flatCompute);
      expect(Object.keys(result.spec?.storage || {}).length === 0).toBeTruthy();
      expect(result.spec?.storage).toBeUndefined();

      const hfHomeEnvPath = result.spec?.environmentVariables?.find(
        (env) => env.name === EHuggingFaceEnvVariableNames.HF_HOME,
      );
      expect(hfHomeEnvPath?.value).toBeDefined();
      expect(hfHomeEnvPath?.value).toBe(HF_HOME_DEFAULT_PATH);
    });
  });

  describe("isLatencyEnableOnCluster", () => {
    it("returns false when dependenciesStatus is undefined", () => {
      const currCluster = {
        ...mockCluster,
        dependenciesStatus: undefined,
      };
      expect(inferenceWorkloadUtil.isLatencyEnableOnCluster(currCluster)).toBe(false);
    });

    it("returns false when knative dependency is missing", () => {
      const currCluster = deepCopy(mockCluster);
      currCluster.dependenciesStatus = { required: {}, optional: {} };
      expect(inferenceWorkloadUtil.isLatencyEnableOnCluster(currCluster)).toBe(false);
    });

    it("returns false when knative components are undefined", () => {
      const currCluster = deepCopy(mockCluster);
      currCluster.dependenciesStatus = {
        required: {},
        optional: {
          knative: {
            available: true,
            components: undefined,
          },
        },
      };
      expect(inferenceWorkloadUtil.isLatencyEnableOnCluster(currCluster)).toBe(false);
    });

    it("returns false when hpa component is not available", () => {
      const currCluster = deepCopy(mockCluster);
      currCluster.dependenciesStatus = {
        required: {},
        optional: {
          knative: {
            available: true,
            components: {
              hpa: { available: false },
            },
          },
        },
      };

      expect(inferenceWorkloadUtil.isLatencyEnableOnCluster(currCluster)).toBe(false);
    });

    it("returns true when hpa component is available", () => {
      const currCluster = deepCopy(mockCluster);
      currCluster.dependenciesStatus = {
        required: {},
        optional: {
          knative: {
            available: true,
            components: {
              hpa: { available: true },
            },
          },
        },
      };
      expect(inferenceWorkloadUtil.isLatencyEnableOnCluster(currCluster)).toBe(true);
    });
  });

  describe("isAutoScaleEnableOnCluster", () => {
    it("returns false when dependenciesStatus is undefined", () => {
      const currCluster = {
        ...mockCluster,
        dependenciesStatus: undefined,
      };
      expect(inferenceWorkloadUtil.isAutoScaleEnableOnCluster(currCluster)).toBe(false);
    });

    it("returns false when knative dependency is missing", () => {
      const currCluster = deepCopy(mockCluster);
      currCluster.dependenciesStatus = { required: {}, optional: {} };
      expect(inferenceWorkloadUtil.isAutoScaleEnableOnCluster(currCluster)).toBe(false);
    });

    it("returns false when knative components are undefined", () => {
      const currCluster = deepCopy(mockCluster);
      currCluster.dependenciesStatus = {
        required: {},
        optional: {
          knative: {
            available: true,
            components: undefined,
          },
        },
      };
      expect(inferenceWorkloadUtil.isAutoScaleEnableOnCluster(currCluster)).toBe(false);
    });

    it("returns false when knative component is not available", () => {
      const currCluster = deepCopy(mockCluster);
      currCluster.dependenciesStatus = {
        required: {},
        optional: {
          knative: {
            available: true,
            components: {
              knative: { available: true },
            },
          },
        },
      };

      expect(inferenceWorkloadUtil.isAutoScaleEnableOnCluster(currCluster)).toBe(false);
    });

    it("returns false when enable-scale-to-zero is not exist", () => {
      const currCluster = deepCopy(mockCluster);
      currCluster.dependenciesStatus = {
        required: {},
        optional: {
          knative: {
            available: true,
            components: {
              knative: {
                available: true,
                components: {},
              },
            },
          },
        },
      };

      expect(inferenceWorkloadUtil.isAutoScaleEnableOnCluster(currCluster)).toBe(false);
    });

    it("returns false when enable-scale-to-zero is not available", () => {
      const currCluster = deepCopy(mockCluster);
      currCluster.dependenciesStatus = {
        required: {},
        optional: {
          knative: {
            available: true,
            components: {
              knative: {
                available: true,
                components: {
                  "enable-scale-to-zero": {
                    available: false,
                  },
                },
              },
            },
          },
        },
      };

      expect(inferenceWorkloadUtil.isAutoScaleEnableOnCluster(currCluster)).toBe(false);
    });

    it("returns true when hpa component is available", () => {
      const currCluster = deepCopy(mockCluster);
      currCluster.dependenciesStatus = {
        required: {},
        optional: {
          knative: {
            available: true,
            components: {
              knative: {
                available: true,
                components: {
                  "enable-scale-to-zero": {
                    available: true,
                  },
                },
              },
            },
          },
        },
      };
      expect(inferenceWorkloadUtil.isAutoScaleEnableOnCluster(currCluster)).toBe(true);
    });
  });

  describe("getInferenceSupportedDataSources", () => {
    it("should return base sources plus cluster 2.19 supported sources when isCluster219 is true", () => {
      const options = { isVersion219OrHigher: true, modelStoreSupportOnly: false };
      const result = inferenceWorkloadUtil.getInferenceSupportedDataSources(options);

      expect(result).toEqual([
        AssetKind.Pvc,
        AssetKind.Git,
        AssetKind.ConfigMap,
        AssetKind.HostPath,
        AssetKind.Nfs,
        AssetKind.DataVolume,
      ]);
    });

    it("should return only base sources when isVersion219OrHigher is false", () => {
      const options = { isVersion219OrHigher: false, modelStoreSupportOnly: false };
      const result = inferenceWorkloadUtil.getInferenceSupportedDataSources(options);

      expect(result).toEqual([AssetKind.Pvc, AssetKind.Git, AssetKind.ConfigMap]);
    });

    it("should return only nim supported sources when modelStoreSupportOnly is true", () => {
      const options = { isVersion219OrHigher: true, modelStoreSupportOnly: true };
      const result = inferenceWorkloadUtil.getInferenceSupportedDataSources(options);

      expect(result).toEqual([AssetKind.Pvc, AssetKind.HostPath, AssetKind.Nfs]);
    });
  });

  describe("prepareUpdateInferenceFlatRequest", () => {
    it("should return only changed spec fields", () => {
      const inference: InferenceFlat = {
        name: "",
        requestedName: "",
        workloadId: "",
        projectId: "",
        clusterId: "",
        createdBy: "",
        createdAt: "",
        desiredPhase: "Running",
        spec: {
          command: "old-command",
          args: "old-args",
          image: "old-image",
          imagePullPolicy: "Always",
          workingDir: "/old/path",
          createHomeDir: false,
          imagePullSecrets: [
            {
              name: "my-dr",
              userCredential: true,
            },
          ],
          environmentVariables: [{ name: "VAR1", value: "old-value" }],
          autoscaling: { minReplicas: 1, maxReplicas: 5 },
        },
      };

      const modelUi: IUISuperWorkloadCreationRequest = {
        name: "",
        projectId: "",
        clusterId: "",
        spec: {
          command: "new-command",
          args: "old-args", // unchanged
          image: "new-image",
          imagePullPolicy: "Never",
          workingDir: "/new/path",
          createHomeDir: true,
          imagePullSecrets: [
            {
              name: "my-dr1",
              userCredential: true,
            },
          ],
          environmentVariables: [{ name: "VAR1", value: "new-value" }],
          autoscaling: { minReplicas: 2, maxReplicas: 10 },
        },
      };

      const result = inferenceWorkloadUtil.prepareUpdateInferenceFlatRequest(inference, modelUi);

      expect(result).toEqual({
        command: "new-command",
        image: "new-image",
        imagePullPolicy: "Never",
        workingDir: "/new/path",
        createHomeDir: true,
        imagePullSecrets: [
          {
            name: "my-dr1",
            userCredential: true,
          },
        ],
        environmentVariables: [{ name: "VAR1", value: "new-value" }],
        autoscaling: { minReplicas: 2, maxReplicas: 10 },
      });
    });

    it("should return undefined in the object if no fields changed", () => {
      const inference: InferenceFlat = {
        name: "",
        requestedName: "",
        workloadId: "",
        projectId: "",
        clusterId: "",
        createdBy: "",
        createdAt: "",
        desiredPhase: "Running",
        spec: {
          command: "same-command",
          args: "same-args",
          image: "same-image",
          imagePullPolicy: "Always",
          workingDir: "/same/path",
          createHomeDir: false,
          environmentVariables: [{ name: "VAR1", value: "same-value" }],
          autoscaling: { minReplicas: 1, maxReplicas: 5 },
        },
      };

      const modelUi: IUISuperWorkloadCreationRequest = {
        name: "",
        projectId: "",
        clusterId: "",
        spec: {
          command: "same-command",
          args: "same-args",
          image: "same-image",
          imagePullPolicy: "Always",
          workingDir: "/same/path",
          createHomeDir: false,
          environmentVariables: [{ name: "VAR1", value: "same-value" }],
          autoscaling: { minReplicas: 1, maxReplicas: 5 },
        },
      };

      const result = inferenceWorkloadUtil.prepareUpdateInferenceFlatRequest(inference, modelUi);

      expect(result).toEqual({
        command: undefined,
        args: undefined,
        image: undefined,
        imagePullPolicy: undefined,
        probes: undefined,
        workingDir: undefined,
        createHomeDir: undefined,
        environmentVariables: undefined,
        autoscaling: undefined,
      });
    });
  });

  describe("_handleHuggingFaceEnvironmentVariables", () => {
    it("should add HF_TOKEN with value if token is provided", () => {
      const envs = inferenceWorkloadUtil._handleHuggingFaceEnvironmentVariables([], "my-token");
      expect(envs).toEqual([
        { name: EHuggingFaceEnvVariableNames.HF_TOKEN, value: "my-token" },
        { name: EHuggingFaceEnvVariableNames.HF_HOME, value: HF_HOME_DEFAULT_PATH },
      ]);
    });

    it("should add HF_TOKEN with credential if genericSecretName is provided", () => {
      const genericSecretName = "my-secret";
      const envs = inferenceWorkloadUtil._handleHuggingFaceEnvironmentVariables([], undefined, {
        genericSecretName,
      });
      expect(envs).toEqual([
        {
          name: EHuggingFaceEnvVariableNames.HF_TOKEN,
          secret: {
            name: genericSecretName,
            key: EHuggingFaceEnvVariableNames.HF_TOKEN,
          },
        },
        { name: EHuggingFaceEnvVariableNames.HF_HOME, value: HF_HOME_DEFAULT_PATH },
      ]);
    });

    it("should remove existing HF_TOKEN before adding new one", () => {
      const envs = inferenceWorkloadUtil._handleHuggingFaceEnvironmentVariables(
        [
          { name: EHuggingFaceEnvVariableNames.HF_TOKEN, value: "old" },
          { name: "OTHER", value: "val" },
        ],
        "new-token",
      );
      expect(envs).toEqual([
        { name: "OTHER", value: "val" },
        { name: EHuggingFaceEnvVariableNames.HF_TOKEN, value: "new-token" },
        { name: EHuggingFaceEnvVariableNames.HF_HOME, value: HF_HOME_DEFAULT_PATH },
      ]);
    });

    it("should add HF_HOME if not present", () => {
      const envs = inferenceWorkloadUtil._handleHuggingFaceEnvironmentVariables([], "token");
      expect(envs.some((e) => e.name === EHuggingFaceEnvVariableNames.HF_HOME)).toBe(true);
    });

    it("should not add HF_HOME if already present", () => {
      const envs = inferenceWorkloadUtil._handleHuggingFaceEnvironmentVariables(
        [{ name: EHuggingFaceEnvVariableNames.HF_HOME, value: "/custom/path" }],
        "token",
      );
      expect(envs.filter((e) => e.name === EHuggingFaceEnvVariableNames.HF_HOME).length).toBe(1);
      expect(envs.find((e) => e.name === EHuggingFaceEnvVariableNames.HF_HOME)?.value).toBe("/custom/path");
    });

    it("should handle empty input array", () => {
      const envs = inferenceWorkloadUtil._handleHuggingFaceEnvironmentVariables([]);
      expect(envs).toEqual([{ name: EHuggingFaceEnvVariableNames.HF_HOME, value: HF_HOME_DEFAULT_PATH }]);
    });

    it("should prefer genericSecretName over token if both are provided", () => {
      const envs = inferenceWorkloadUtil._handleHuggingFaceEnvironmentVariables([], "token", {
        genericSecretName: "secret",
      });
      expect(envs).toEqual([
        {
          name: EHuggingFaceEnvVariableNames.HF_TOKEN,
          secret: {
            name: "secret",
            key: EHuggingFaceEnvVariableNames.HF_TOKEN,
          },
        },
        { name: EHuggingFaceEnvVariableNames.HF_HOME, value: HF_HOME_DEFAULT_PATH },
      ]);
    });

    // on openshift, we need to add the home dir env variable, since the default path is not writable
    // and lead to permission error on /.cache since the default HOME is "/"
    it("should add HOME env variable if addHomeDir is true", () => {
      const envs = inferenceWorkloadUtil._handleHuggingFaceEnvironmentVariables([], undefined, { addHomeDir: true });
      expect(envs).toEqual([
        { name: EHuggingFaceEnvVariableNames.HF_HOME, value: HF_HOME_DEFAULT_PATH },
        { name: EHuggingFaceEnvVariableNames.HOME, value: HOME_DEFAULT_PATH },
      ]);
    });

    it("should not add HOME env variable if addHomeDir is false or undefined", () => {
      const envs1 = inferenceWorkloadUtil._handleHuggingFaceEnvironmentVariables([], undefined, { addHomeDir: false });
      const envs2 = inferenceWorkloadUtil._handleHuggingFaceEnvironmentVariables([]);
      expect(envs1).toEqual([{ name: EHuggingFaceEnvVariableNames.HF_HOME, value: HF_HOME_DEFAULT_PATH }]);
      expect(envs2).toEqual([{ name: EHuggingFaceEnvVariableNames.HF_HOME, value: HF_HOME_DEFAULT_PATH }]);
    });

    it("should add HOME env variable along with HF_TOKEN and HF_HOME if addHomeDir is true and token is provided", () => {
      const envs = inferenceWorkloadUtil._handleHuggingFaceEnvironmentVariables([], "token", { addHomeDir: true });
      expect(envs).toEqual([
        { name: EHuggingFaceEnvVariableNames.HF_TOKEN, value: "token" },
        { name: EHuggingFaceEnvVariableNames.HF_HOME, value: HF_HOME_DEFAULT_PATH },
        { name: EHuggingFaceEnvVariableNames.HOME, value: HOME_DEFAULT_PATH },
      ]);
    });

    it("should add HOME env variable along with HF_TOKEN as secret and HF_HOME if addHomeDir and genericSecretName are provided", () => {
      const envs = inferenceWorkloadUtil._handleHuggingFaceEnvironmentVariables([], "token", {
        genericSecretName: "secret",
        addHomeDir: true,
      });
      expect(envs).toEqual([
        {
          name: EHuggingFaceEnvVariableNames.HF_TOKEN,
          secret: {
            name: "secret",
            key: EHuggingFaceEnvVariableNames.HF_TOKEN,
          },
        },
        { name: EHuggingFaceEnvVariableNames.HF_HOME, value: HF_HOME_DEFAULT_PATH },
        { name: EHuggingFaceEnvVariableNames.HOME, value: HOME_DEFAULT_PATH },
      ]);
    });
  });

  describe("normalizeAutoScaleData", () => {
    const { normalizeAutoScaleData } = inferenceWorkloadUtil;

    it("should default minReplicas and maxReplicas to 1 if not provided", () => {
      expect(normalizeAutoScaleData({})).toMatchObject({ minReplicas: 1, maxReplicas: 1 });
      expect(normalizeAutoScaleData(undefined)).toMatchObject({ minReplicas: 1, maxReplicas: 1 });
      expect(normalizeAutoScaleData(null)).toMatchObject({ minReplicas: 1, maxReplicas: 1 });
    });

    it("should set thresholdValue to DEFAULT_THRESHOLD_VALUE if thresholdMetric is set but thresholdValue is undefined", () => {
      expect(normalizeAutoScaleData({ minReplicas: 2, maxReplicas: 3, thresholdMetric: "latency" }).thresholdValue).toBe(
        DEFAULT_THRESHOLD_VALUE,
      );
    });

    it("should clear scaleToZeroRetentionSeconds if minReplicas > 0", () => {
      expect(
        normalizeAutoScaleData({ minReplicas: 2, maxReplicas: 3, scaleToZeroRetentionSeconds: 100 })
          .scaleToZeroRetentionSeconds,
      ).toBeUndefined();
    });

    it("should clear thresholdMetric and thresholdValue if minReplicas == maxReplicas", () => {
      expect(
        normalizeAutoScaleData({ minReplicas: 2, maxReplicas: 2, thresholdMetric: "latency", thresholdValue: 10 }),
      ).toMatchObject({ thresholdMetric: undefined, thresholdValue: undefined });
    });

    it("should clear thresholdMetric and thresholdValue if minReplicas is 0 and maxReplicas is 1", () => {
      expect(
        normalizeAutoScaleData({ minReplicas: 0, maxReplicas: 1, thresholdMetric: "latency", thresholdValue: 10 }),
      ).toMatchObject({ thresholdMetric: undefined, thresholdValue: undefined });
    });

    it("should clear thresholdMetric and thresholdValue if minReplicas or maxReplicas is null or undefined", () => {
      expect(
        normalizeAutoScaleData({
          minReplicas: undefined,
          maxReplicas: 1,
          thresholdMetric: "latency",
          thresholdValue: 10,
        }),
      ).toMatchObject({ thresholdMetric: undefined, thresholdValue: undefined });
      expect(
        normalizeAutoScaleData({
          minReplicas: 1,
          maxReplicas: undefined,
          thresholdMetric: "latency",
          thresholdValue: 10,
        }),
      ).toMatchObject({ thresholdMetric: undefined, thresholdValue: undefined });
      expect(
        normalizeAutoScaleData({ minReplicas: null, maxReplicas: 1, thresholdMetric: "latency", thresholdValue: 10 }),
      ).toMatchObject({ thresholdMetric: undefined, thresholdValue: undefined });
      expect(
        normalizeAutoScaleData({ minReplicas: 1, maxReplicas: null, thresholdMetric: "latency", thresholdValue: 10 }),
      ).toMatchObject({ thresholdMetric: undefined, thresholdValue: undefined });
    });

    it("should preserve scaleToZeroRetentionSeconds if minReplicas is 0", () => {
      expect(
        normalizeAutoScaleData({ minReplicas: 0, maxReplicas: 2, scaleToZeroRetentionSeconds: 123 })
          .scaleToZeroRetentionSeconds,
      ).toBe(123);
    });
  });

  describe("prepareDistributedInferenceFlatRequest", () => {
    function makeDistributedModel(replicas?: number): IUISuperWorkloadCreationRequest {
      return {
        name: "dist-inf",
        projectId: "project-1",
        clusterId: "cluster-1",
        spec: {
          replicas,
          servingPort: {
            protocol: "http",
            container: 8080,
          },
          leader: {
            image: "leader:latest",
          },
          worker: {
            image: "worker:latest",
          },
        },
      } as IUISuperWorkloadCreationRequest;
    }

    it("should include the selected replicas in the distributed inference creation request", () => {
      const result = inferenceWorkloadUtil.prepareDistributedInferenceFlatRequest(makeDistributedModel(3), 2);

      expect(result.spec).toBeDefined();
      expect(result.spec?.replicas).toBe(3);
      expect(result.spec?.workers).toBe(2);
    });

    it("should default replicas to one when no value exists on the UI model", () => {
      const result = inferenceWorkloadUtil.prepareDistributedInferenceFlatRequest(makeDistributedModel(), 2);

      expect(result.spec).toBeDefined();
      expect(result.spec?.replicas).toBe(1);
    });
  });

  describe("_getFlatInferenceSpecData", () => {
    it("validates required fields and returns a valid SupersetSpec", () => {
      const workloadSpec: IUISuperWorkloadSpec = {
        image: "nginx:latest",
        servingPort: {
          protocol: "http",
          container: 3030,
        },
        autoscaling: { minReplicas: 1, maxReplicas: 3 },
      };

      const result = inferenceWorkloadUtil._getFlatInferenceSpecData(workloadSpec);

      expect(result).toEqual({
        annotations: undefined,
        args: undefined,
        autoDeletionTimeAfterCompletionSeconds: undefined,
        autoscaling: {
          maxReplicas: 3,
          minReplicas: 1,
        },
        command: undefined,
        compute: undefined,
        createHomeDir: undefined,
        environmentVariables: undefined,
        exposedUrls: undefined,
        image: "nginx:latest",
        imagePullPolicy: undefined,
        labels: undefined,
        nodePools: undefined,
        nodeType: undefined,
        podAffinity: undefined,
        ports: undefined,
        priorityClass: undefined,
        probes: undefined,
        relatedUrls: undefined,
        restartPolicy: undefined,
        security: undefined,
        servingPort: {
          container: 3030,
          protocol: "http",
        },
        stdin: undefined,
        storage: undefined,
        terminateAfterPreemption: undefined,
        tolerations: undefined,
        tty: undefined,
        workingDir: undefined,
      });
    });

    it("throws an error if servingPort is invalid (using the validation function)", () => {
      const workloadSpec: IUISuperWorkloadSpec = {
        image: "nginx:latest",
        servingPort: undefined,
        autoscaling: { minReplicas: 1, maxReplicas: 3 },
      };

      expect(() => inferenceWorkloadUtil._getFlatInferenceSpecData(workloadSpec)).toThrowError(
        "Invalid servingPort configuration",
      );
    });

    it("throws an error if servingPort is missing protocol", () => {
      const workloadSpec: IUISuperWorkloadSpec = {
        image: "nginx:latest",
        servingPort: {
          container: 300,
        },
        autoscaling: { minReplicas: 1, maxReplicas: 3 },
      };

      expect(() => inferenceWorkloadUtil._getFlatInferenceSpecData(workloadSpec)).toThrowError(
        "Invalid servingPort configuration",
      );
    });

    it("throws an error if servingPort is missing container", () => {
      const workloadSpec: IUISuperWorkloadSpec = {
        image: "nginx:latest",
        servingPort: {
          protocol: "http",
        },
        autoscaling: { minReplicas: 1, maxReplicas: 3 },
      };

      expect(() => inferenceWorkloadUtil._getFlatInferenceSpecData(workloadSpec)).toThrowError(
        "Invalid servingPort configuration",
      );
    });

    it("throws an error if autoscaling are invalid (using the validation function)", () => {
      const workloadSpec: IUISuperWorkloadSpec = {
        image: "nginx:latest",
        servingPort: {
          protocol: "http",
          container: 3030,
        },
        autoscaling: undefined,
      };

      expect(() => inferenceWorkloadUtil._getFlatInferenceSpecData(workloadSpec)).toThrowError(
        "Invalid autoscaling configuration",
      );
    });

    it("throws an error if autoscaling is missing minReplicas field (using the validation function)", () => {
      const workloadSpec: IUISuperWorkloadSpec = {
        image: "nginx:latest",
        servingPort: {
          protocol: "http",
          container: 3030,
        },
        autoscaling: {
          maxReplicas: 5,
        },
      };

      expect(() => inferenceWorkloadUtil._getFlatInferenceSpecData(workloadSpec)).toThrowError(
        "Invalid autoscaling configuration",
      );
    });

    it("throws an error if autoscaling is missing maxReplicas field (using the validation function)", () => {
      const workloadSpec: IUISuperWorkloadSpec = {
        image: "nginx:latest",
        servingPort: {
          protocol: "http",
          container: 3030,
        },
        autoscaling: {
          minReplicas: 5,
        },
      };

      expect(() => inferenceWorkloadUtil._getFlatInferenceSpecData(workloadSpec)).toThrowError(
        "Invalid autoscaling configuration",
      );
    });

    it("returns environmentVariables as is if present", () => {
      const envVars = [
        { name: "FOO", value: "bar" },
        { name: "BAR", value: "baz" },
      ];
      const workloadSpec: IUISuperWorkloadSpec = {
        image: "nginx:latest",
        servingPort: {
          protocol: "http",
          container: 3030,
        },
        autoscaling: { minReplicas: 1, maxReplicas: 3 },
        environmentVariables: envVars,
      };
      const result = inferenceWorkloadUtil._getFlatInferenceSpecData(workloadSpec);
      expect(result.environmentVariables).toEqual(envVars);
    });

    it("returns undefined for environmentVariables if not present", () => {
      const workloadSpec: IUISuperWorkloadSpec = {
        image: "nginx:latest",
        servingPort: {
          protocol: "http",
          container: 3030,
        },
        autoscaling: { minReplicas: 1, maxReplicas: 3 },
      };
      const result = inferenceWorkloadUtil._getFlatInferenceSpecData(workloadSpec);
      expect(result.environmentVariables).toBeUndefined();
    });
  });

  describe("getHFModelNameFromArgs", () => {
    const { getHFModelNameFromArgs } = inferenceWorkloadUtil;

    it("should extract model name for TGI prefix", () => {
      const modelName = "my-tgi-model";
      const args = `${ARGS_PREFIX_FOR_TGI} ${modelName}`;
      expect(getHFModelNameFromArgs(args)).toBe(modelName);
    });

    it("should extract model name for VLLM prefix", () => {
      const modelName = "my-vllm-model";
      const args = `${ARGS_PREFIX_FOR_VLLM} ${modelName}`;
      expect(getHFModelNameFromArgs(args)).toBe(modelName);
    });

    it("should trim whitespace from extracted model name", () => {
      const modelName = "   spaced-model   ";
      const args = `${ARGS_PREFIX_FOR_VLLM}   ${modelName}`;
      expect(getHFModelNameFromArgs(args)).toBe(modelName.trim());
    });

    it("should return empty string if no known prefix is present", () => {
      expect(getHFModelNameFromArgs("--unknown-prefix model")).toBe("");
      expect(getHFModelNameFromArgs("")).toBe("");
    });

    it("should handle args with only the prefix and no model name", () => {
      expect(getHFModelNameFromArgs(ARGS_PREFIX_FOR_TGI)).toBe("");
      expect(getHFModelNameFromArgs(ARGS_PREFIX_FOR_VLLM)).toBe("");
    });
  });

  describe("getHFArgsForModel", () => {
    const { getHFArgsForModel } = inferenceWorkloadUtil;

    it("should return TGI args string when image is TGI", () => {
      const modelName = "tgi-model";
      const result = getHFArgsForModel(LLM_IMAGE_FOR_TGI, modelName);
      expect(result).toBe(`${ARGS_PREFIX_FOR_TGI} ${modelName}`);
    });

    it("should return VLLM args string when image is VLLM", () => {
      const modelName = "vllm-model";
      const result = getHFArgsForModel(LLM_IMAGE_FOR_HUGGING_FACE, modelName);
      expect(result).toBe(`${ARGS_PREFIX_FOR_VLLM} ${modelName}`);
    });

    it("should handle model names with extra whitespace", () => {
      const modelName = "   spaced-model   ";
      const result = getHFArgsForModel(LLM_IMAGE_FOR_HUGGING_FACE, modelName);
      expect(result).toBe(`${ARGS_PREFIX_FOR_VLLM} ${modelName}`);
    });

    it("should handle empty model name", () => {
      expect(getHFArgsForModel(LLM_IMAGE_FOR_TGI, "")).toBe(`${ARGS_PREFIX_FOR_TGI} `);
      expect(getHFArgsForModel(LLM_IMAGE_FOR_HUGGING_FACE, "")).toBe(`${ARGS_PREFIX_FOR_VLLM} `);
    });
  });

  describe("adjustNimWorkloadProperties", () => {
    function makeBaseModel(overrides = {}): IUISuperWorkloadCreationRequest {
      return {
        spec: {
          environmentVariables: [] as EnvironmentVariableOfAsset[],
          compute: { gpuDevicesRequest: 1 },
          ...overrides,
        },
      } as IUISuperWorkloadCreationRequest;
    }

    it("adds NIM_CACHE_PATH from nfs storage", () => {
      const model = makeBaseModel({
        storage: { nfs: [{ path: "/mnt/nfs" }] },
      });
      const result = inferenceWorkloadUtil.adjustNimWorkloadProperties({ ...model });
      expect(result.spec.environmentVariables).toContainEqual({
        name: ENimEnvVariableNames.NimCachePath,
        value: "/mnt/nfs",
      } as EnvironmentVariableOfAsset);
    });

    it("adds NIM_CACHE_PATH from pvc storage", () => {
      const model = makeBaseModel({
        storage: { pvc: [{ path: "/mnt/pvc" }] },
      });
      const result = inferenceWorkloadUtil.adjustNimWorkloadProperties({ ...model });
      expect(result.spec.environmentVariables).toContainEqual({
        name: ENimEnvVariableNames.NimCachePath,
        value: "/mnt/pvc",
      } as EnvironmentVariableOfAsset);
    });

    it("adds NIM_CACHE_PATH from hostPath storage", () => {
      const model = makeBaseModel({
        storage: { hostPath: [{ path: "/mnt/host" }] },
      });
      const result = inferenceWorkloadUtil.adjustNimWorkloadProperties({ ...model });
      expect(result.spec.environmentVariables).toContainEqual({
        name: ENimEnvVariableNames.NimCachePath,
        value: "/mnt/host",
      } as EnvironmentVariableOfAsset);
    });

    it("removes old NIM_CACHE_PATH and sets new one", () => {
      const model = makeBaseModel({
        storage: { nfs: [{ path: "/mnt/nfs" }] },
        environmentVariables: [
          { name: ENimEnvVariableNames.NimCachePath, value: "/old/path" },
          { name: "OTHER", value: "foo" },
        ] as EnvironmentVariableOfAsset[],
      });
      // Place envs at the right place
      model.spec.environmentVariables = [
        { name: ENimEnvVariableNames.NimCachePath, value: "/old/path" },
        { name: "OTHER", value: "foo" },
      ] as EnvironmentVariableOfAsset[];
      const result = inferenceWorkloadUtil.adjustNimWorkloadProperties({ ...model });
      expect(result.spec.environmentVariables).toContainEqual({
        name: ENimEnvVariableNames.NimCachePath,
        value: "/mnt/nfs",
      } as EnvironmentVariableOfAsset);
      expect(result.spec.environmentVariables).not.toContainEqual({
        name: ENimEnvVariableNames.NimCachePath,
        value: "/old/path",
      } as EnvironmentVariableOfAsset);
      expect(result.spec.environmentVariables).toContainEqual({
        name: "OTHER",
        value: "foo",
      } as EnvironmentVariableOfAsset);
    });

    it("does nothing if no storage is present", () => {
      const model = makeBaseModel({
        environmentVariables: [{ name: "FOO", value: "bar" }] as EnvironmentVariableOfAsset[],
      });
      model.spec.environmentVariables = [{ name: "FOO", value: "bar" }] as EnvironmentVariableOfAsset[];
      const result = inferenceWorkloadUtil.adjustNimWorkloadProperties({ ...model });
      expect(result.spec.environmentVariables).toEqual([{ name: "FOO", value: "bar" }] as EnvironmentVariableOfAsset[]);
    });

    it("always sets compute.largeShmRequest to true and preserves other compute fields", () => {
      const model = makeBaseModel({
        compute: { gpuDevicesRequest: 2 },
      });
      const result = inferenceWorkloadUtil.adjustNimWorkloadProperties({ ...model });
      expect(result.spec.compute?.largeShmRequest).toBe(true);
      expect(result.spec.compute?.gpuDevicesRequest).toBe(2);
    });

    it("orders environment variables with NIM base vars first then others", () => {
      const model = makeBaseModel({
        environmentVariables: [
          { name: ENimEnvVariableNames.NimCachePath, value: "/cache" },
          { name: ENimEnvVariableNames.NgcApiKey, value: "key" },
          { name: ENimEnvVariableNames.NimModelProfile, value: "profile-1" },
          { name: ENimEnvVariableNames.NimServerPort, value: "8000" },
          { name: ENimEnvVariableNames.OutlinesCacheDir, value: "/tmp/outlines" },
        ] as EnvironmentVariableOfAsset[],
      });
      model.spec.environmentVariables = [
        { name: ENimEnvVariableNames.NimCachePath, value: "/cache" },
        { name: ENimEnvVariableNames.NgcApiKey, value: "key" },
        { name: ENimEnvVariableNames.NimModelProfile, value: "profile-1" },
        { name: ENimEnvVariableNames.NimServerPort, value: "8000" },
        { name: ENimEnvVariableNames.OutlinesCacheDir, value: "/tmp/outlines" },
      ] as EnvironmentVariableOfAsset[];
      const result = inferenceWorkloadUtil.adjustNimWorkloadProperties({ ...model });
      const names = result.spec.environmentVariables?.map((e) => e.name) ?? [];
      expect(names.indexOf(ENimEnvVariableNames.OutlinesCacheDir)).toBeLessThan(
        names.indexOf(ENimEnvVariableNames.NimServerPort),
      );
      expect(names.indexOf(ENimEnvVariableNames.NimServerPort)).toBeLessThan(
        names.indexOf(ENimEnvVariableNames.NimModelProfile),
      );
      expect(names.indexOf(ENimEnvVariableNames.NimModelProfile)).toBeLessThan(
        names.indexOf(ENimEnvVariableNames.NgcApiKey),
      );
      expect(names.indexOf(ENimEnvVariableNames.NgcApiKey)).toBeLessThan(
        names.indexOf(ENimEnvVariableNames.NimCachePath),
      );
    });

    it("orders env vars after _cleanStorageAndSetEnvVar adds NIM_CACHE_PATH", () => {
      const model = makeBaseModel({
        storage: { nfs: [{ path: "/mnt/nfs" }] },
        environmentVariables: [
          { name: ENimEnvVariableNames.NgcApiKey, value: "key" },
          { name: ENimEnvVariableNames.NimServerPort, value: "8000" },
        ] as EnvironmentVariableOfAsset[],
      });
      model.spec.environmentVariables = [
        { name: ENimEnvVariableNames.NgcApiKey, value: "key" },
        { name: ENimEnvVariableNames.NimServerPort, value: "8000" },
      ] as EnvironmentVariableOfAsset[];
      const result = inferenceWorkloadUtil.adjustNimWorkloadProperties({ ...model });
      const names = result.spec.environmentVariables?.map((e) => e.name) ?? [];
      expect(names).toHaveLength(3);
      expect(names[0]).toBe(ENimEnvVariableNames.NimServerPort);
      expect(names[1]).toBe(ENimEnvVariableNames.NgcApiKey);
      expect(names[2]).toBe(ENimEnvVariableNames.NimCachePath);
      expect(result.spec.environmentVariables).toContainEqual({
        name: ENimEnvVariableNames.NimCachePath,
        value: "/mnt/nfs",
      } as EnvironmentVariableOfAsset);
    });
  });

  describe("getAutoScalingData", () => {
    describe("minReplicas", () => {
      // data from default value
      it("returns the correct autoscaling.minReplicas data when no policy is provided and no autoscaling is provided", () => {
        const result: AutoScaling = inferenceWorkloadUtil.getAutoScalingData(undefined, null);
        expect(result.minReplicas).toBe(MIN_REPLICAS_DEFAULT_VALUE);
      });

      // data from copy and edit
      it("returns the correct autoscaling.minReplicas data when no policy is provided and no autoscaling is provided", () => {
        const result: AutoScaling = inferenceWorkloadUtil.getAutoScalingData({ minReplicas: 4, maxReplicas: 6 }, null);
        expect(result.minReplicas).toBe(4);
      });

      // data from policy.defaults
      it("returns the correct autoscaling.minReplicas data when policy provided with default value", () => {
        const policy: UnifiedPolicyInfoPerReplicaType = {
          defaults: {
            autoscaling: {
              minReplicas: 5,
            },
          },
        };
        const autoScaling: AutoScaling = { minReplicas: policy.defaults?.autoscaling?.minReplicas, maxReplicas: 6 };
        const result: AutoScaling = inferenceWorkloadUtil.getAutoScalingData(autoScaling, policy);

        expect(result.minReplicas).toBe(policy.defaults?.autoscaling?.minReplicas);
      });

      // data from policy.rules
      it("returns the correct autoscaling.minReplicas data when policy provided with default value", () => {
        const policy: UnifiedPolicyInfoPerReplicaType = {
          rules: {
            autoscaling: {
              minReplicas: {
                min: 7,
              },
            },
          },
        };
        const result: AutoScaling = inferenceWorkloadUtil.getAutoScalingData(undefined, policy);
        expect(result.minReplicas).toBe(policy.rules?.autoscaling?.minReplicas?.min);
      });
    });

    describe("maxReplicas", () => {
      // data from default value
      it("returns the correct autoscaling.maxReplicas data when no policy is provided and no autoscaling is provided", () => {
        const result: AutoScaling = inferenceWorkloadUtil.getAutoScalingData(undefined, null);
        expect(result.maxReplicas).toBe(MAX_REPLICAS_DEFAULT_VALUE);
      });

      // data from copy and edit
      it("returns the correct autoscaling.maxReplicas data when no policy is provided and no autoscaling is provided", () => {
        const result: AutoScaling = inferenceWorkloadUtil.getAutoScalingData({ minReplicas: 4, maxReplicas: 6 }, null);
        expect(result.maxReplicas).toBe(6);
      });

      // data from policy.defaults
      it("returns the correct autoscaling.maxReplicas data when policy provided with default value", () => {
        const policy: UnifiedPolicyInfoPerReplicaType = {
          defaults: {
            autoscaling: {
              maxReplicas: 5,
            },
          },
        };
        const autoscaling: AutoScaling = { minReplicas: 1, maxReplicas: policy.defaults?.autoscaling?.maxReplicas };
        const result: AutoScaling = inferenceWorkloadUtil.getAutoScalingData(autoscaling, policy);

        expect(result.maxReplicas).toBe(policy.defaults?.autoscaling?.maxReplicas);
      });

      // data from policy.rules
      it("returns the correct autoscaling.maxReplicas data when policy provided with default value", () => {
        const policy: UnifiedPolicyInfoPerReplicaType = {
          rules: {
            autoscaling: {
              maxReplicas: {
                min: 7,
              },
            },
          },
        };
        const result: AutoScaling = inferenceWorkloadUtil.getAutoScalingData(undefined, policy);
        expect(result.maxReplicas).toBe(policy.rules?.autoscaling?.maxReplicas?.min);
      });
    });

    describe("scaleToZeroRetentionSeconds", () => {
      // copy and edit and there is a value
      it("returns the correct autoscaling.scaleToZeroRetentionSeconds data when no policy is provided and no autoscaling is provided", () => {
        const result: AutoScaling = inferenceWorkloadUtil.getAutoScalingData(undefined, null);
        expect(result.minReplicas).toBe(1);
        expect(result.maxReplicas).toBe(1);
        expect(result.scaleToZeroRetentionSeconds).toBe(undefined);
      });

      // data from the policy.defaults
      it("returns the correct autoscaling.scaleToZeroRetentionSeconds data when policy provided with default value", () => {
        const policy: UnifiedPolicyInfoPerReplicaType = {
          defaults: {
            autoscaling: { scaleToZeroRetentionSeconds: 60 },
          },
        };
        const result: AutoScaling = inferenceWorkloadUtil.getAutoScalingData(undefined, policy);
        expect(result.minReplicas).toBe(1);
        expect(result.maxReplicas).toBe(1);
        expect(result.scaleToZeroRetentionSeconds).toBe(undefined);
      });

      it("returns the correct autoscaling.scaleToZeroRetentionSeconds data when policy provided with default value and minReplicas is > 0", () => {
        const policy: UnifiedPolicyInfoPerReplicaType = {
          defaults: {
            autoscaling: { scaleToZeroRetentionSeconds: 60 },
          },
        };
        const result: AutoScaling = inferenceWorkloadUtil.getAutoScalingData({ minReplicas: 1, maxReplicas: 2 }, policy);
        expect(result.minReplicas).toBe(1);
        expect(result.maxReplicas).toBe(2);
        expect(result.scaleToZeroRetentionSeconds).toBe(undefined);
      });

      it("returns the correct autoscaling.scaleToZeroRetentionSeconds data when policy provided with default value and minReplicas is 0", () => {
        const policy: UnifiedPolicyInfoPerReplicaType = {
          defaults: {
            autoscaling: { scaleToZeroRetentionSeconds: 60 },
          },
        };
        const result: AutoScaling = inferenceWorkloadUtil.getAutoScalingData(
          { minReplicas: 0, maxReplicas: 2, scaleToZeroRetentionSeconds: 42 },
          policy,
        );
        expect(result.minReplicas).toBe(0);
        expect(result.maxReplicas).toBe(2);
        expect(result.scaleToZeroRetentionSeconds).toBe(42);
      });

      it("returns the correct autoscaling.scaleToZeroRetentionSeconds data when policy provided with rules min value and minReplicas is 0", () => {
        const policy: UnifiedPolicyInfoPerReplicaType = {
          rules: {
            autoscaling: {
              scaleToZeroRetentionSeconds: {
                min: 100,
              },
            },
          },
        };
        const result: AutoScaling = inferenceWorkloadUtil.getAutoScalingData({ minReplicas: 0, maxReplicas: 2 }, policy);
        expect(result.minReplicas).toBe(0);
        expect(result.maxReplicas).toBe(2);
        expect(result.scaleToZeroRetentionSeconds).toBe(policy.rules?.autoscaling?.scaleToZeroRetentionSeconds?.min);
      });

      it("returns the correct autoscaling.scaleToZeroRetentionSeconds data when no policy is provided and minReplicas is 0 and maxReplicas is 2", () => {
        const result: AutoScaling = inferenceWorkloadUtil.getAutoScalingData({ minReplicas: 0, maxReplicas: 2 }, null);
        expect(result.minReplicas).toBe(0);
        expect(result.maxReplicas).toBe(2);
        expect(result.scaleToZeroRetentionSeconds).toBe(SCALE_TO_ZERO_RETENTION_IN_MINUTES_DEFAULT_VALUE * 60);
      });
    });

    describe("metric and metricThreshold", () => {
      const THRESHOLD_DEFAULT_METRIC = "throughput";
      it("returns the correct autoscaling.metric and autoscaling.metricThreshold data when no policy is provided and no autoscaling is provided", () => {
        const result: AutoScaling = inferenceWorkloadUtil.getAutoScalingData(undefined, null);
        expect(result.metric).toBeUndefined();
        expect(result.metricThreshold).toBeUndefined();
      });

      it("returns the correct autoscaling.metric and autoscaling.metricThreshold data when no policy is provided and autoscaling min and max are equal", () => {
        const result: AutoScaling = inferenceWorkloadUtil.getAutoScalingData({ minReplicas: 5, maxReplicas: 5 }, null);
        expect(result.metric).toBeUndefined();
        expect(result.metricThreshold).toBeUndefined();
      });

      it("returns the correct autoscaling.metric and autoscaling.metricThreshold data when policy is provided and autoscaling.min is 0", () => {
        const result: AutoScaling = inferenceWorkloadUtil.getAutoScalingData(
          { minReplicas: 0, maxReplicas: 1, metric: "cpu", metricThreshold: 0.5 },
          null,
        );
        expect(result.metric).toBeUndefined();
        expect(result.metricThreshold).toBeUndefined();
      });

      it("returns the correct autoscaling.metric and autoscaling.metricThreshold data when no policy is provided and autoscaling max is greater than min", () => {
        const result: AutoScaling = inferenceWorkloadUtil.getAutoScalingData({ minReplicas: 5, maxReplicas: 10 }, null);
        expect(result.metric).toBe(THRESHOLD_DEFAULT_METRIC);
        expect(result.metricThreshold).toBe(DEFAULT_THRESHOLD_VALUE);
      });

      it("returns the correct autoscaling.metric and autoscaling.metricThreshold data when policy is provided and autoscaling.min is 0 and autoscaling.max greater than 1", () => {
        const policy: UnifiedPolicyInfoPerReplicaType = {
          rules: {
            autoscaling: {
              maxReplicas: {
                canEdit: false,
                max: 12,
                min: 4,
              },
            },
          },
        };
        const result: AutoScaling = inferenceWorkloadUtil.getAutoScalingData(
          { minReplicas: 0, maxReplicas: 2, scaleToZeroRetentionSeconds: 80 },
          policy,
        );
        expect(result.metric).not.toBeUndefined();
        expect(result.metricThreshold).not.toBeUndefined();
      });

      it("returns the correct autoscaling.metric and autoscaling.metricThreshold as undefined when policy is provided and autoscaling.min is 0 and autoscaling.max is 1", () => {
        const policy: UnifiedPolicyInfoPerReplicaType = {
          rules: {
            autoscaling: {
              maxReplicas: {
                canEdit: false,
                max: 12,
                min: 4,
              },
            },
          },
        };
        const result: AutoScaling = inferenceWorkloadUtil.getAutoScalingData(
          { minReplicas: 0, maxReplicas: 1, scaleToZeroRetentionSeconds: 80 },
          policy,
        );
        expect(result.metric).toBeUndefined();
        expect(result.metricThreshold).toBeUndefined();
      });

      it("returns the correct autoscaling.metric and autoscaling.metricThreshold data when no policy is provided and autoscaling max is greater than min and metric and metricThreshold are provided", () => {
        const result: AutoScaling = inferenceWorkloadUtil.getAutoScalingData(
          { minReplicas: 5, maxReplicas: 10, metric: "cpu", metricThreshold: 0.5 },
          null,
        );
        expect(result.metric).toBe("cpu");
        expect(result.metricThreshold).toBe(0.5);
      });

      it("returns the correct autoscaling.metric and autoscaling.metricThreshold data when policy is provided and autoscaling max greater then min and min metricThreshold is greater then default", () => {
        const metricThreshold = 120;
        const policy: UnifiedPolicyInfoPerReplicaType = {
          rules: {
            autoscaling: {
              metricThreshold: {
                min: metricThreshold,
              },
            },
          },
        };
        const result: AutoScaling = inferenceWorkloadUtil.getAutoScalingData({ minReplicas: 2, maxReplicas: 5 }, policy);

        expect(result.metric).toBe(THRESHOLD_DEFAULT_METRIC);
        expect(result.metricThreshold).toBe(metricThreshold);
      });

      it("returns the correct autoscaling.metric and autoscaling.metricThreshold data when policy is provided and autoscaling max greater then min and min metricThreshold is less then default", () => {
        const policy: UnifiedPolicyInfoPerReplicaType = {
          rules: {
            autoscaling: {
              metricThreshold: {
                min: 50,
              },
            },
          },
        };
        const result: AutoScaling = inferenceWorkloadUtil.getAutoScalingData({ minReplicas: 2, maxReplicas: 5 }, policy);

        expect(result.metric).toBe(THRESHOLD_DEFAULT_METRIC);
        expect(result.metricThreshold).toBe(DEFAULT_THRESHOLD_VALUE);
      });

      it("returns the correct autoscaling.metric and autoscaling.metricThreshold data when policy is provided and no autoscaling is provided", () => {
        const policy: UnifiedPolicyInfoPerReplicaType = {
          defaults: {
            autoscaling: { metric: "cpu", metricThreshold: 0.5 },
          },
        };
        const result: AutoScaling = inferenceWorkloadUtil.getAutoScalingData(undefined, policy);
        expect(result.metric).toBeUndefined();
        expect(result.metricThreshold).toBeUndefined();
      });
    });
  });

  describe("_addNimEnvironmentVariablesIfNeeded", () => {
    it("should add all required NIM environment variables if none exist", () => {
      const input: EnvironmentVariableOfAsset[] = [];
      const result = inferenceWorkloadUtil._addNimEnvironmentVariablesIfNeeded([...input]);
      expect(result).toEqual(
        expect.arrayContaining([
          { name: ENimEnvVariableNames.OutlinesCacheDir, value: "/tmp/outlines" },
          { name: ENimEnvVariableNames.NimServerPort, value: String(SERVING_PORT_SERVER_PORT) },
          { name: ENimEnvVariableNames.NimJsonlLogging, value: "1" },
          { name: ENimEnvVariableNames.NimLogLevel, value: "INFO" },
        ]),
      );
      expect(result.length).toBe(4);
    });

    it("should not add variables that already exist", () => {
      const input: EnvironmentVariableOfAsset[] = [
        { name: ENimEnvVariableNames.OutlinesCacheDir, value: "/custom/path" },
        { name: ENimEnvVariableNames.NimServerPort, value: "9000" },
      ];
      const result = inferenceWorkloadUtil._addNimEnvironmentVariablesIfNeeded([...input]);
      expect(result).toEqual(
        expect.arrayContaining([
          { name: ENimEnvVariableNames.OutlinesCacheDir, value: "/custom/path" },
          { name: ENimEnvVariableNames.NimServerPort, value: "9000" },
          { name: ENimEnvVariableNames.NimJsonlLogging, value: "1" },
          { name: ENimEnvVariableNames.NimLogLevel, value: "INFO" },
        ]),
      );
      expect(result.length).toBe(4);
    });

    it("should not duplicate any NIM environment variable if all already exist", () => {
      const input: EnvironmentVariableOfAsset[] = [
        { name: ENimEnvVariableNames.OutlinesCacheDir, value: "/tmp/outlines" },
        { name: ENimEnvVariableNames.NimServerPort, value: String(SERVING_PORT_SERVER_PORT) },
        { name: ENimEnvVariableNames.NimJsonlLogging, value: "1" },
        { name: ENimEnvVariableNames.NimLogLevel, value: "INFO" },
      ];
      const result = inferenceWorkloadUtil._addNimEnvironmentVariablesIfNeeded([...input]);
      expect(result).toEqual(input);
      expect(result.length).toBe(4);
    });

    it("should preserve unrelated environment variables", () => {
      const input: EnvironmentVariableOfAsset[] = [{ name: "OTHER_ENV", value: "other-value" }];
      const result = inferenceWorkloadUtil._addNimEnvironmentVariablesIfNeeded([...input]);
      expect(result).toEqual(
        expect.arrayContaining([
          { name: "OTHER_ENV", value: "other-value" },
          { name: ENimEnvVariableNames.OutlinesCacheDir, value: "/tmp/outlines" },
          { name: ENimEnvVariableNames.NimServerPort, value: String(SERVING_PORT_SERVER_PORT) },
          { name: ENimEnvVariableNames.NimJsonlLogging, value: "1" },
          { name: ENimEnvVariableNames.NimLogLevel, value: "INFO" },
        ]),
      );
      expect(result.length).toBe(5);
    });

    it("should use policy defaults.servingPort.container for NimServerPort when provided", () => {
      const policy: UnifiedPolicyInfoPerReplicaType = {
        defaults: { servingPort: { container: 8080 } },
      };
      const input: EnvironmentVariableOfAsset[] = [];
      const result = inferenceWorkloadUtil._addNimEnvironmentVariablesIfNeeded([...input], policy);
      const nimServerPort = result.find((e) => e.name === ENimEnvVariableNames.NimServerPort);
      expect(nimServerPort).toEqual({ name: ENimEnvVariableNames.NimServerPort, value: "8080" });
    });

    it("should use SERVING_PORT_SERVER_PORT for NimServerPort when policy has no servingPort default", () => {
      const policy: UnifiedPolicyInfoPerReplicaType = { defaults: {} };
      const input: EnvironmentVariableOfAsset[] = [];
      const result = inferenceWorkloadUtil._addNimEnvironmentVariablesIfNeeded([...input], policy);
      const nimServerPort = result.find((e) => e.name === ENimEnvVariableNames.NimServerPort);
      expect(nimServerPort).toEqual({
        name: ENimEnvVariableNames.NimServerPort,
        value: String(SERVING_PORT_SERVER_PORT),
      });
    });
  });

  describe("setInferenceNIMUIDefaults", () => {
    const baseModel: IUISuperWorkloadCreationRequest = {
      name: "test",
      projectId: "p1",
      clusterId: "c1",
      spec: {
        environmentVariables: [],
        compute: {},
        security: {},
      },
    };

    it("should add all required NIM environment variables if none exist", () => {
      const model = JSON.parse(JSON.stringify(baseModel));
      const result = inferenceWorkloadUtil.setInferenceNIMUIDefaults(model, false);
      expect(result.spec.environmentVariables).toEqual(
        expect.arrayContaining([
          { name: ENimEnvVariableNames.OutlinesCacheDir, value: "/tmp/outlines" },
          { name: ENimEnvVariableNames.NimServerPort, value: String(SERVING_PORT_SERVER_PORT) },
          { name: ENimEnvVariableNames.NimJsonlLogging, value: "1" },
          { name: ENimEnvVariableNames.NimLogLevel, value: "INFO" },
        ]),
      );
    });

    it("should not overwrite existing NIM environment variables", () => {
      const model = JSON.parse(JSON.stringify(baseModel));
      model.spec.environmentVariables = [
        { name: ENimEnvVariableNames.OutlinesCacheDir, value: "/custom/path" },
        { name: ENimEnvVariableNames.NimServerPort, value: "9000" },
      ];
      const result = inferenceWorkloadUtil.setInferenceNIMUIDefaults(model, false);
      expect(result.spec.environmentVariables).toEqual(
        expect.arrayContaining([
          { name: ENimEnvVariableNames.OutlinesCacheDir, value: "/custom/path" },
          { name: ENimEnvVariableNames.NimServerPort, value: "9000" },
          { name: ENimEnvVariableNames.NimJsonlLogging, value: "1" },
          { name: ENimEnvVariableNames.NimLogLevel, value: "INFO" },
        ]),
      );
    });

    it("should set runAsGid and runAsUid to 1000 in security", () => {
      const model = JSON.parse(JSON.stringify(baseModel));
      const result = inferenceWorkloadUtil.setInferenceNIMUIDefaults(model, false);
      expect(result.spec.security).toMatchObject({ runAsGid: 1000, runAsUid: 1000 });
    });

    it("should set servingPort protocol to http and container to SERVING_PORT_SERVER_PORT if missing", () => {
      const model = JSON.parse(JSON.stringify(baseModel));
      const result = inferenceWorkloadUtil.setInferenceNIMUIDefaults(model, false);
      expect(result.spec.servingPort).toMatchObject({ protocol: "http", container: SERVING_PORT_SERVER_PORT });
    });

    it("should not overwrite existing servingPort.container", () => {
      const model = JSON.parse(JSON.stringify(baseModel));
      model.spec.servingPort = { protocol: "http", container: 1234 };
      const result = inferenceWorkloadUtil.setInferenceNIMUIDefaults(model, false);
      expect(result.spec.servingPort).toMatchObject({ protocol: "http", container: 1234 });
    });

    it("should preserve unrelated fields in the model", () => {
      const model = JSON.parse(JSON.stringify(baseModel));
      model.spec.labels = [{ name: "label1", value: "v1" }];
      const result = inferenceWorkloadUtil.setInferenceNIMUIDefaults(model, false);
      expect(result.spec.labels).toEqual([
        { name: "label1", value: "v1" },
        { name: NIM_WORKLOAD_LABEL, value: "true" },
      ]);
    });

    it("should not overwrite existing servingConfiguration if showRequestTimeout is true", () => {
      const model = JSON.parse(JSON.stringify(baseModel));
      model.spec.servingConfiguration = { initializationTimeoutSeconds: 1, requestTimeoutSeconds: 2 };
      const result = inferenceWorkloadUtil.setInferenceNIMUIDefaults(model, false);
      expect(result.spec.servingConfiguration).toEqual({ initializationTimeoutSeconds: 1, requestTimeoutSeconds: 2 });
    });

    it("should pass through policy and isSSO to setInferenceUIDefaults", () => {
      // This test is a smoke test to ensure no error is thrown and the function returns a model
      const model = JSON.parse(JSON.stringify(baseModel));
      const policy: UnifiedPolicyInfoPerReplicaType = { rules: {}, defaults: {} };
      const result = inferenceWorkloadUtil.setInferenceNIMUIDefaults(model, true, policy);
      expect(result).toBeTruthy();
    });

    it("should add run.ai/nim-workload label if not present", () => {
      const model = JSON.parse(JSON.stringify(baseModel));
      const result = inferenceWorkloadUtil.setInferenceNIMUIDefaults(model, false);
      expect(result.spec.labels).toContainEqual({ name: NIM_WORKLOAD_LABEL, value: "true" });
      expect(result.spec.labels).toHaveLength(1);
    });

    it("should not add run.ai/nim-workload label if present", () => {
      const model = JSON.parse(JSON.stringify(baseModel));
      model.spec.labels = [{ name: NIM_WORKLOAD_LABEL, value: "true" }];
      const result = inferenceWorkloadUtil.setInferenceNIMUIDefaults(model, false);
      expect(result.spec.labels).toContainEqual({ name: NIM_WORKLOAD_LABEL, value: "true" });
      expect(result.spec.labels).toHaveLength(1);
    });
  });

  describe("setDistributedUIDefaults", () => {
    const baseModel: IUISuperWorkloadCreationRequest = {
      name: "distributed-nim",
      projectId: "p1",
      clusterId: "c1",
      spec: {
        leader: {},
        worker: {},
      },
    };

    it("adds the NIM identity label to the leader spec for NIM workloads", () => {
      const result = inferenceWorkloadUtil.setDistributedUIDefaults(
        JSON.parse(JSON.stringify(baseModel)),
        false,
        undefined,
        EServerType.Nim,
      );

      expect(result.spec.leader?.labels).toContainEqual({ name: NIM_WORKLOAD_LABEL, value: "true" });
    });

    it("does not duplicate an existing NIM identity label", () => {
      const model = JSON.parse(JSON.stringify(baseModel));
      model.spec.leader.labels = [{ name: NIM_WORKLOAD_LABEL, value: "true" }];

      const result = inferenceWorkloadUtil.setDistributedUIDefaults(model, false, undefined, EServerType.Nim);

      expect(result.spec.leader?.labels).toEqual([{ name: NIM_WORKLOAD_LABEL, value: "true" }]);
    });

    it("does not add the NIM identity label to custom workloads", () => {
      const result = inferenceWorkloadUtil.setDistributedUIDefaults(
        JSON.parse(JSON.stringify(baseModel)),
        false,
        undefined,
        EServerType.Custom,
      );

      expect(result.spec.leader?.labels).toBeUndefined();
    });
  });

  describe("_ensureStoragePathEnvVar", () => {
    const ENV_NAME = "MY_PATH_VAR";

    it("adds env var from nfs storage", () => {
      const envs: Array<EnvironmentVariableOfAsset | EnvironmentVariable> = [];
      const storage: IUIStorageFields = { nfs: [{ path: "/mnt/nfs", id: "nfs1", type: "nfs" }] };
      const result = inferenceWorkloadUtil._ensureStoragePathEnvVar(envs, storage, ENV_NAME);
      expect(result).toContainEqual({ name: ENV_NAME, value: "/mnt/nfs" });
    });

    it("adds env var from pvc storage", () => {
      const envs: Array<EnvironmentVariableOfAsset | EnvironmentVariable> = [];
      const storage: IUIStorageFields = { pvc: [{ path: "/mnt/pvc", id: "pvc1", type: "pvc" }] };
      const result = inferenceWorkloadUtil._ensureStoragePathEnvVar(envs, storage, ENV_NAME);
      expect(result).toContainEqual({ name: ENV_NAME, value: "/mnt/pvc" });
    });

    it("adds env var from hostPath storage", () => {
      const envs: Array<EnvironmentVariableOfAsset | EnvironmentVariable> = [];
      const storage: IUIStorageFields = { hostPath: [{ path: "/mnt/host", id: "host1", type: "hostPath" }] };
      const result = inferenceWorkloadUtil._ensureStoragePathEnvVar(envs, storage, ENV_NAME);
      expect(result).toContainEqual({ name: ENV_NAME, value: "/mnt/host" });
    });

    it("removes previous instances of the env var", () => {
      const envs: Array<EnvironmentVariableOfAsset | EnvironmentVariable> = [
        { name: ENV_NAME, value: "/old/path" },
        { name: "OTHER", value: "foo" },
      ];
      const storage: IUIStorageFields = { nfs: [{ path: "/mnt/nfs", id: "nfs1", type: "nfs" }] };
      const result = inferenceWorkloadUtil._ensureStoragePathEnvVar(envs, storage, ENV_NAME);
      expect(result).toContainEqual({ name: ENV_NAME, value: "/mnt/nfs" });
      expect(result).not.toContainEqual({ name: ENV_NAME, value: "/old/path" });
      expect(result).toContainEqual({ name: "OTHER", value: "foo" });
    });

    it("does not add env var if no storage is present", () => {
      const envs: Array<EnvironmentVariableOfAsset | EnvironmentVariable> = [{ name: "FOO", value: "bar" }];
      const result = inferenceWorkloadUtil._ensureStoragePathEnvVar(envs, undefined, ENV_NAME);
      expect(result).toEqual([{ name: "FOO", value: "bar" }]);
    });

    it("does not add env var if storage has no path", () => {
      const envs: Array<EnvironmentVariableOfAsset | EnvironmentVariable> = [{ name: "FOO", value: "bar" }];
      const storage: IUIStorageFields = { nfs: [{ id: "nfs1", type: "nfs" }] };
      const result = inferenceWorkloadUtil._ensureStoragePathEnvVar(envs, storage, ENV_NAME);
      expect(result).toEqual([{ name: "FOO", value: "bar" }]);
    });

    it("handles empty environmentVariables array", () => {
      const storage: IUIStorageFields = { pvc: [{ path: "/mnt/pvc", id: "pvc1", type: "pvc" }] };
      const result = inferenceWorkloadUtil._ensureStoragePathEnvVar([], storage, ENV_NAME);
      expect(result).toContainEqual({ name: ENV_NAME, value: "/mnt/pvc" });
    });

    it("handles environmentVariables with multiple previous instances", () => {
      const envs: Array<EnvironmentVariableOfAsset | EnvironmentVariable> = [
        { name: ENV_NAME, value: "/old/path1" },
        { name: ENV_NAME, value: "/old/path2" },
        { name: "FOO", value: "bar" },
      ];
      const storage: IUIStorageFields = { hostPath: [{ path: "/mnt/host", id: "host1", type: "hostPath" }] };
      const result = inferenceWorkloadUtil._ensureStoragePathEnvVar(envs, storage, ENV_NAME);
      expect(result.filter((e): e is { name: string; value: string } => e.name === ENV_NAME)).toEqual([
        { name: ENV_NAME, value: "/mnt/host" },
      ]);
      expect(result).toContainEqual({ name: "FOO", value: "bar" });
    });
  });

  describe("removeEmptyStorageArrays", () => {
    it("removes empty arrays from storage", () => {
      const storage = {
        nfs: [],
        pvc: [{ path: "/mnt/pvc", id: "pvc1", type: AssetKind.Pvc }],
        hostPath: [],
      };
      const result = inferenceWorkloadUtil.removeEmptyStorageArrays(storage);
      expect(result).toEqual({ pvc: [{ path: "/mnt/pvc", id: "pvc1", type: AssetKind.Pvc }] });
    });

    it("leaves non-empty arrays intact", () => {
      const storage = {
        nfs: [{ path: "/mnt/nfs", id: "nfs1", type: AssetKind.Nfs }],
        pvc: [{ path: "/mnt/pvc", id: "pvc1", type: AssetKind.Pvc }],
      };
      const result = inferenceWorkloadUtil.removeEmptyStorageArrays(storage);
      expect(result).toEqual(storage);
    });

    it("returns undefined if input is undefined", () => {
      expect(inferenceWorkloadUtil.removeEmptyStorageArrays(undefined)).toBeUndefined();
    });

    it("returns an empty object if all arrays are empty", () => {
      const storage = { nfs: [], pvc: [] };
      const result = inferenceWorkloadUtil.removeEmptyStorageArrays(storage);
      expect(result).toEqual({});
    });

    it("does not mutate the original object", () => {
      const storage = {
        nfs: [],
        pvc: [{ path: "/mnt/pvc", id: "pvc1", type: AssetKind.Pvc }],
      };
      const copy = { ...storage };
      inferenceWorkloadUtil.removeEmptyStorageArrays(storage);
      expect(storage).toEqual(copy);
    });
  });

  describe("_orderHfEnvironmentVariables", () => {
    it("returns empty array for empty input", () => {
      const result = inferenceWorkloadUtil._orderHfEnvironmentVariables([]);
      expect(result).toEqual([]);
    });

    it("orders HF_HOME before HF_TOKEN when both are present", () => {
      const envVars: EnvironmentVariableOfAsset[] = [
        { name: EHuggingFaceEnvVariableNames.HF_TOKEN, value: "token" },
        { name: EHuggingFaceEnvVariableNames.HF_HOME, value: HF_HOME_DEFAULT_PATH },
      ];
      const result = inferenceWorkloadUtil._orderHfEnvironmentVariables(envVars);
      expect(result.map((e) => e.name)).toEqual([
        EHuggingFaceEnvVariableNames.HF_HOME,
        EHuggingFaceEnvVariableNames.HF_TOKEN,
      ]);
    });

    it("puts vars not in the order list after HF_HOME and HF_TOKEN, preserving their relative order", () => {
      const envVars: EnvironmentVariableOfAsset[] = [
        { name: "OTHER_B", value: "b" },
        { name: EHuggingFaceEnvVariableNames.HF_TOKEN, value: "token" },
        { name: "OTHER_A", value: "a" },
        { name: EHuggingFaceEnvVariableNames.HF_HOME, value: HF_HOME_DEFAULT_PATH },
      ];
      const result = inferenceWorkloadUtil._orderHfEnvironmentVariables(envVars);
      expect(result.map((e) => e.name)).toEqual([
        EHuggingFaceEnvVariableNames.HF_HOME,
        EHuggingFaceEnvVariableNames.HF_TOKEN,
        "OTHER_B",
        "OTHER_A",
      ]);
    });

    it("leaves single var unchanged when only HF_HOME is present", () => {
      const envVars: EnvironmentVariableOfAsset[] = [{ name: EHuggingFaceEnvVariableNames.HF_HOME, value: "/custom" }];
      const result = inferenceWorkloadUtil._orderHfEnvironmentVariables(envVars);
      expect(result).toEqual(envVars);
    });

    it("leaves single var unchanged when only HF_TOKEN is present", () => {
      const envVars: EnvironmentVariableOfAsset[] = [{ name: EHuggingFaceEnvVariableNames.HF_TOKEN, value: "t" }];
      const result = inferenceWorkloadUtil._orderHfEnvironmentVariables(envVars);
      expect(result).toEqual(envVars);
    });

    it("deduplicates by name keeping last occurrence", () => {
      const envVars: EnvironmentVariableOfAsset[] = [
        { name: EHuggingFaceEnvVariableNames.HF_TOKEN, value: "first" },
        { name: EHuggingFaceEnvVariableNames.HF_TOKEN, value: "second" },
      ];
      const result = inferenceWorkloadUtil._orderHfEnvironmentVariables(envVars);
      expect(result).toHaveLength(1);
      expect(result[0]).toMatchObject({ name: EHuggingFaceEnvVariableNames.HF_TOKEN, value: "second" });
    });

    it("skips entries without name", () => {
      const envVars = [
        { name: "", value: "empty" },
        { name: EHuggingFaceEnvVariableNames.HF_HOME, value: HF_HOME_DEFAULT_PATH },
      ] as EnvironmentVariableOfAsset[];
      const result = inferenceWorkloadUtil._orderHfEnvironmentVariables(envVars);
      expect(result.map((e) => e.name)).toEqual([EHuggingFaceEnvVariableNames.HF_HOME]);
    });
  });

  describe("_orderNimEnvironmentVariables", () => {
    it("returns empty array for empty input", () => {
      const result = inferenceWorkloadUtil._orderNimEnvironmentVariables([]);
      expect(result).toEqual([]);
    });

    it("orders known NIM vars in the correct order when all are present", () => {
      const envVars: EnvironmentVariableOfAsset[] = [
        { name: ENimEnvVariableNames.NimCachePath, value: "/test/sample" },
        { name: ENimEnvVariableNames.NgcApiKey, value: "key" },
        { name: ENimEnvVariableNames.NimModelProfile, value: "profile-1" },
        { name: ENimEnvVariableNames.NimServerPort, value: "8000" },
        { name: ENimEnvVariableNames.OutlinesCacheDir, value: "/tmp/outlines" },
        { name: ENimEnvVariableNames.NimLogLevel, value: "INFO" },
        { name: ENimEnvVariableNames.NimJsonlLogging, value: "1" },
      ];
      const result = inferenceWorkloadUtil._orderNimEnvironmentVariables(envVars);
      expect(result.map((e) => e.name)).toEqual([
        ENimEnvVariableNames.OutlinesCacheDir,
        ENimEnvVariableNames.NimServerPort,
        ENimEnvVariableNames.NimJsonlLogging,
        ENimEnvVariableNames.NimLogLevel,
        ENimEnvVariableNames.NimModelProfile,
        ENimEnvVariableNames.NgcApiKey,
        ENimEnvVariableNames.NimCachePath,
      ]);
    });

    it("puts vars not in the order list after the ordered ones, preserving their relative order", () => {
      const envVars: EnvironmentVariableOfAsset[] = [
        { name: ENimEnvVariableNames.NimModelProfile, value: "profile-1" },
        { name: ENimEnvVariableNames.NimCachePath, value: "/cache" },
        { name: ENimEnvVariableNames.NimServerPort, value: "8000" },
        { name: ENimEnvVariableNames.OutlinesCacheDir, value: "/tmp/outlines" },
      ];
      const result = inferenceWorkloadUtil._orderNimEnvironmentVariables(envVars);
      expect(result.map((e) => e.name)).toEqual([
        ENimEnvVariableNames.OutlinesCacheDir,
        ENimEnvVariableNames.NimServerPort,
        ENimEnvVariableNames.NimModelProfile,
        ENimEnvVariableNames.NimCachePath,
      ]);
    });

    it("leaves single var unchanged", () => {
      const envVars: EnvironmentVariableOfAsset[] = [{ name: ENimEnvVariableNames.NgcApiKey, value: "key" }];
      const result = inferenceWorkloadUtil._orderNimEnvironmentVariables(envVars);
      expect(result).toEqual(envVars);
    });

    it("handles partial set of ordered names", () => {
      const envVars: EnvironmentVariableOfAsset[] = [
        { name: ENimEnvVariableNames.NimLogLevel, value: "INFO" },
        { name: ENimEnvVariableNames.NimServerPort, value: "9000" },
      ];
      const result = inferenceWorkloadUtil._orderNimEnvironmentVariables(envVars);
      expect(result.map((e) => e.name)).toEqual([ENimEnvVariableNames.NimServerPort, ENimEnvVariableNames.NimLogLevel]);
    });

    it("deduplicates by name keeping last occurrence", () => {
      const envVars: EnvironmentVariableOfAsset[] = [
        { name: ENimEnvVariableNames.NimServerPort, value: "8000" },
        { name: ENimEnvVariableNames.NimServerPort, value: "9000" },
      ];
      const result = inferenceWorkloadUtil._orderNimEnvironmentVariables(envVars);
      expect(result).toHaveLength(1);
      expect(result[0]).toMatchObject({ name: ENimEnvVariableNames.NimServerPort, value: "9000" });
    });

    it("skips entries without name", () => {
      const envVars = [
        { name: "", value: "empty" },
        { name: ENimEnvVariableNames.NimServerPort, value: "8000" },
      ] as EnvironmentVariableOfAsset[];
      const result = inferenceWorkloadUtil._orderNimEnvironmentVariables(envVars);
      expect(result.map((e) => e.name)).toEqual([ENimEnvVariableNames.NimServerPort]);
    });
  });

  describe("_cleanStorageAndSetEnvVar", () => {
    const ENV_NAME = "MY_PATH_VAR";
    function makeModel(storage: IUIStorageFields | undefined, envs: Array<{ name: string; value: string }> = []) {
      return {
        spec: {
          storage,
          environmentVariables: envs,
        },
      } as IUISuperWorkloadCreationRequest;
    }

    it("removes empty storage arrays and sets env var from nfs", () => {
      const model = makeModel({ nfs: [{ path: "/mnt/nfs", id: "nfs1", type: AssetKind.Nfs }], pvc: [] });
      inferenceWorkloadUtil._cleanStorageAndSetEnvVar(model, ENV_NAME);
      expect(model.spec.storage).toEqual({ nfs: [{ path: "/mnt/nfs", id: "nfs1", type: AssetKind.Nfs }] });
      expect(model.spec.environmentVariables).toContainEqual({ name: ENV_NAME, value: "/mnt/nfs" });
    });

    it("removes previous envs and sets new one from pvc", () => {
      const model = makeModel({ pvc: [{ path: "/mnt/pvc", id: "pvc1", type: AssetKind.Pvc }], nfs: [] }, [
        { name: ENV_NAME, value: "/old/path" },
        { name: "OTHER", value: "foo" },
      ]);
      inferenceWorkloadUtil._cleanStorageAndSetEnvVar(model, ENV_NAME);
      expect(model.spec.storage).toEqual({ pvc: [{ path: "/mnt/pvc", id: "pvc1", type: AssetKind.Pvc }] });
      expect(model.spec.environmentVariables).toContainEqual({ name: ENV_NAME, value: "/mnt/pvc" });
      expect(model.spec.environmentVariables).not.toContainEqual({ name: ENV_NAME, value: "/old/path" });
      expect(model.spec.environmentVariables).toContainEqual({ name: "OTHER", value: "foo" });
    });

    it("does not add env var if no storage is present", () => {
      const model = makeModel(undefined, [{ name: "FOO", value: "bar" }]);
      inferenceWorkloadUtil._cleanStorageAndSetEnvVar(model, ENV_NAME);
      expect(model.spec.environmentVariables).toEqual([{ name: "FOO", value: "bar" }]);
    });

    it("does not add env var if storage has no path", () => {
      const model = makeModel({ nfs: [{ id: "nfs1", type: AssetKind.Nfs }] }, [{ name: "FOO", value: "bar" }]);
      inferenceWorkloadUtil._cleanStorageAndSetEnvVar(model, ENV_NAME);
      expect(model.spec.environmentVariables).toEqual([{ name: "FOO", value: "bar" }]);
    });

    it("handles empty environmentVariables array", () => {
      const model = makeModel({ pvc: [{ path: "/mnt/pvc", id: "pvc1", type: AssetKind.Pvc }] }, []);
      inferenceWorkloadUtil._cleanStorageAndSetEnvVar(model, ENV_NAME);
      expect(model.spec.environmentVariables).toContainEqual({ name: ENV_NAME, value: "/mnt/pvc" });
    });
  });

  describe("_ensureDefaultEnvVar", () => {
    const ENV_NAME = "MY_DEFAULT_VAR";
    const ENV_VALUE = "default-value";

    it("adds the variable if missing", () => {
      const envs: EnvironmentVariableOfAsset[] = [{ name: "FOO", value: "bar" }];
      const result = inferenceWorkloadUtil._ensureDefaultEnvVar(envs, ENV_NAME, ENV_VALUE);
      expect(result).toContainEqual({ name: ENV_NAME, value: ENV_VALUE });
      expect(result).toContainEqual({ name: "FOO", value: "bar" });
    });

    it("does not add the variable if already present", () => {
      const envs: EnvironmentVariableOfAsset[] = [
        { name: ENV_NAME, value: ENV_VALUE },
        { name: "FOO", value: "bar" },
      ];
      const result = inferenceWorkloadUtil._ensureDefaultEnvVar(envs, ENV_NAME, ENV_VALUE);
      expect(result.filter((e): e is EnvironmentVariableOfAsset => e.name === ENV_NAME)).toHaveLength(1);
      expect(result).toContainEqual({ name: ENV_NAME, value: ENV_VALUE });
    });

    it("works with an empty array", () => {
      const result = inferenceWorkloadUtil._ensureDefaultEnvVar([], ENV_NAME, ENV_VALUE);
      expect(result).toEqual([{ name: ENV_NAME, value: ENV_VALUE }]);
    });

    it("does not duplicate the variable if called multiple times", () => {
      let envs: EnvironmentVariableOfAsset[] = [];
      envs = inferenceWorkloadUtil._ensureDefaultEnvVar(envs, ENV_NAME, ENV_VALUE);
      envs = inferenceWorkloadUtil._ensureDefaultEnvVar(envs, ENV_NAME, ENV_VALUE);
      expect(envs.filter((e): e is EnvironmentVariableOfAsset => e.name === ENV_NAME)).toHaveLength(1);
    });
  });
});

describe("adjustHFWorkloadProperties", () => {
  const makeModel = (
    storage: IUIStorageFields | undefined,
    envs: EnvironmentVariableOfAsset[] = [],
  ): IUISuperWorkloadCreationRequest => ({
    name: "test",
    projectId: "1",
    clusterId: "cluster-1",
    spec: {
      storage,
      environmentVariables: envs,
    },
  });

  it("sets HF_HOME from nfs storage", () => {
    const model = makeModel({ nfs: [{ path: "/mnt/nfs", id: "nfs1", type: AssetKind.Nfs }] });
    const result = inferenceWorkloadUtil.adjustHFWorkloadProperties(model);
    expect(
      result.spec.environmentVariables?.some(
        (e) => e.name === EHuggingFaceEnvVariableNames.HF_HOME && e.value === "/mnt/nfs",
      ),
    ).toBe(true);
  });

  it("sets HF_HOME from pvc storage", () => {
    const model = makeModel({ pvc: [{ path: "/mnt/pvc", id: "pvc1", type: AssetKind.Pvc }] });
    const result = inferenceWorkloadUtil.adjustHFWorkloadProperties(model);
    expect(
      result.spec.environmentVariables?.some(
        (e) => e.name === EHuggingFaceEnvVariableNames.HF_HOME && e.value === "/mnt/pvc",
      ),
    ).toBe(true);
  });

  it("sets HF_HOME from hostPath storage", () => {
    const model = makeModel({ hostPath: [{ path: "/mnt/host", id: "host1", type: AssetKind.HostPath }] });
    const result = inferenceWorkloadUtil.adjustHFWorkloadProperties(model);
    expect(
      result.spec.environmentVariables?.some(
        (e) => e.name === EHuggingFaceEnvVariableNames.HF_HOME && e.value === "/mnt/host",
      ),
    ).toBe(true);
  });

  it("adds default HF_HOME if not present and no storage", () => {
    const model = makeModel(undefined, [{ name: "FOO", value: "bar" }]);
    const result = inferenceWorkloadUtil.adjustHFWorkloadProperties(model);
    expect(
      result.spec.environmentVariables?.some(
        (e) => e.name === EHuggingFaceEnvVariableNames.HF_HOME && e.value === HF_HOME_DEFAULT_PATH,
      ),
    ).toBe(true);
    expect(result.spec.environmentVariables?.some((e) => e.name === "FOO" && e.value === "bar")).toBe(true);
  });

  it("does not duplicate HF_HOME if already present", () => {
    const model = makeModel(undefined, [
      { name: EHuggingFaceEnvVariableNames.HF_HOME, value: "/custom/path" },
      { name: "FOO", value: "bar" },
    ]);
    const result = inferenceWorkloadUtil.adjustHFWorkloadProperties(model);
    const hfHomeVars =
      result.spec.environmentVariables?.filter((e) => e.name === EHuggingFaceEnvVariableNames.HF_HOME) || [];
    expect(hfHomeVars).toHaveLength(1);
    expect(hfHomeVars[0]).toEqual({ name: EHuggingFaceEnvVariableNames.HF_HOME, value: "/custom/path" });
    expect(result.spec.environmentVariables?.some((e) => e.name === "FOO" && e.value === "bar")).toBe(true);
  });

  it("preserves other environment variables", () => {
    const model = makeModel({ nfs: [{ path: "/mnt/nfs", id: "nfs1", type: AssetKind.Nfs }] }, [
      { name: "FOO", value: "bar" },
      { name: "BAR", value: "baz" },
    ]);
    const result = inferenceWorkloadUtil.adjustHFWorkloadProperties(model);
    expect(result.spec.environmentVariables?.some((e) => e.name === "FOO" && e.value === "bar")).toBe(true);
    expect(result.spec.environmentVariables?.some((e) => e.name === "BAR" && e.value === "baz")).toBe(true);
    expect(
      result.spec.environmentVariables?.some(
        (e) => e.name === EHuggingFaceEnvVariableNames.HF_HOME && e.value === "/mnt/nfs",
      ),
    ).toBe(true);
  });

  it("orders environment variables with HF_HOME before HF_TOKEN", () => {
    const model = makeModel(undefined, [
      { name: EHuggingFaceEnvVariableNames.HF_TOKEN, value: "token" },
      { name: EHuggingFaceEnvVariableNames.HF_HOME, value: HF_HOME_DEFAULT_PATH },
      { name: "OTHER", value: "foo" },
    ]);
    const result = inferenceWorkloadUtil.adjustHFWorkloadProperties(model);
    const names = result.spec.environmentVariables?.map((e) => e.name) ?? [];
    expect(names.indexOf(EHuggingFaceEnvVariableNames.HF_HOME)).toBeLessThan(
      names.indexOf(EHuggingFaceEnvVariableNames.HF_TOKEN),
    );
    expect(names.indexOf(EHuggingFaceEnvVariableNames.HF_TOKEN)).toBeLessThan(names.indexOf("OTHER"));
  });
});
