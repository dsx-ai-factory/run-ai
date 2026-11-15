/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IUIDistributedUISetup } from "./workload.util";
import { workloadUtil } from "./workload.util";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { AssetsIds, SpecificRunParams } from "@/swagger-models/assets-service-client";
import {
  type AssetIdAndKind,
  AssetKind,
  DistributedFramework,
  InternalConnectionType,
} from "@/swagger-models/assets-service-client";
import type { Connection, Workload } from "@/swagger-models/workloads-client";
import type {
  IUIWorkloadAssets,
  IUIWorkloadCreation,
  IUIWorkloadSpecificEnv,
  IWorkloadMeta,
} from "@/models/workload.model";
import { EWorkloadType } from "@/models/workload.model";
import type { SpecificRunParamsKeys } from "./workload.util.models";

import {
  trainingMultiNodeCustomMasterResponseV2,
  advancedFilterPhaseMockRunning,
  advancedFilterPhaseMockStopped,
  advancedFilterPhaseMockPending,
  trainingSingleNodeResponseV2,
  trainingMultiNodeMasterSameAsWorkerResponseV2,
  llmWorkload,
  workspaceWorkload,
  distributedWorkload,
  connectionsMocks,
  nimWorkloadWithNimImage,
  nimWorkloadWithNimWorkloadFlag,
  nonNimInferenceWorkload,
  nonInferenceWorkload,
} from "./workload.util.mocks";
import { deepCopy } from "../common.util";
import type { IAdvancedFilterModel, IEnumBasedFilterModel } from "@/models/filter.model";

import { EFilterOperator, EColumnFilterType } from "@/models/filter.model";
import { filterService } from "@/services/filter.service/filter.service";
import { storageUtil } from "@/utils/storage.util";
import { ETableFilters } from "@/models/table.model";
import { WorkloadSortFilterFields } from "@/swagger-models/workloads-client";
import { STORAGE_KEYS } from "@/models/storage.model";
import { LLM_IMAGE_FOR_HUGGING_FACE, LLM_IMAGE_FOR_TGI } from "@/models/inference.model";
import { GENERIC_WORKLOAD_ICON } from "@/common/icons.constant";

vi.mock("@/utils/storage.util/storage.util.ts");
vi.mock("@/services/infra/unleash.service/unleash.service");

describe("workload.util", () => {
  describe("getUIWorkloadAssets", () => {
    it("should return the full workload assets object match to the UI", () => {
      expect(workloadUtil.getUIWorkloadAssets({ environment: "123" })).toStrictEqual({
        environment: "123",
        compute: null,
        datasources: undefined,
        uiVolumes: undefined,
      });

      expect(workloadUtil.getUIWorkloadAssets({ environment: "123", compute: "125" })).toStrictEqual({
        environment: "123",
        compute: "125",
        datasources: undefined,
        uiVolumes: undefined,
      });

      expect(
        workloadUtil.getUIWorkloadAssets({
          environment: "123",
          compute: "125",
          datasources: [{ id: "123", kind: "git" }],
        }),
      ).toStrictEqual({
        environment: "123",
        compute: "125",
        datasources: [{ id: "123", kind: "git" }],
        uiVolumes: undefined,
      });
    });
  });

  describe("getUIWorkloadSpecificEnv", () => {
    it("should return the full specific env params object match to the UI", () => {
      expect(workloadUtil.getUIWorkloadSpecificEnv({})).toStrictEqual({
        args: undefined,
        command: undefined,
        runAsUid: null,
        runAsGid: null,
        supplementalGroups: undefined,
        nodeType: null,
        allowOverQuota: null,
        nodePools: null,
        autoDeletionTimeAfterCompletionSeconds: null,
        connections: undefined,
        environmentVariables: undefined,
        annotations: undefined,
        labels: undefined,
        backoffLimit: 6,
        tolerations: undefined,
        podAffinity: undefined,
        parallelism: undefined,
        completions: undefined,
        terminationGracePeriodSeconds: undefined,
      });
    });
  });

  describe("getWorkloadAssetsFromUIWorkloadAssets", () => {
    it("should convert the ui workload assets to assets backend model", () => {
      const baseObj: IUIWorkloadAssets = { environment: "123", compute: "125" };
      const datasources: Array<AssetIdAndKind> = [{ id: "123", kind: AssetKind.Git }];
      const workloadVolumesIds: Array<string> = ["123"];
      expect(workloadUtil.getWorkloadAssetsFromUIWorkloadAssets(baseObj)).toStrictEqual(baseObj);
      expect(workloadUtil.getWorkloadAssetsFromUIWorkloadAssets({ ...baseObj, datasources })).toStrictEqual({
        ...baseObj,
        datasources,
      });
      expect(workloadUtil.getWorkloadAssetsFromUIWorkloadAssets(baseObj, workloadVolumesIds)).toStrictEqual({
        ...baseObj,
        workloadVolumes: workloadVolumesIds,
      });
    });
  });

  describe("getWorkloadSpecificEnvFromUIWorkloadSpecificEnv", () => {
    const baseObj = {
      args: null,
      command: null,
      runAsGid: undefined,
      runAsUid: undefined,
      supplementalGroups: undefined,
    };
    it("should convert the an empty ui specific env to specific env backend model", () => {
      expect(workloadUtil.getWorkloadSpecificEnvFromUIWorkloadSpecificEnv({})).toStrictEqual(baseObj);
    });

    it("should convert an empty string for command and args to null", () => {
      expect(
        workloadUtil.getWorkloadSpecificEnvFromUIWorkloadSpecificEnv({
          args: "",
          command: "",
          runAsGid: 123,
          runAsUid: 123,
        }),
      ).toStrictEqual({
        args: null,
        command: null,
        runAsGid: 123,
        runAsUid: 123,
        supplementalGroups: undefined,
      });
    });

    it("should convert UI itemized fields to the backend model", () => {
      expect(
        workloadUtil.getWorkloadSpecificEnvFromUIWorkloadSpecificEnv({
          environmentVariables: [{ name: "env-key-1", value: "env-value-1", exclude: true }],
          annotations: [{ name: "annotation-key-1", value: "annotation-value-1", exclude: false }],
          labels: [{ name: "label-key-1", value: "label-value-1" }],
        }),
      ).toStrictEqual({
        ...baseObj,
        environmentVariables: [{ name: "env-key-1", value: "env-value-1", exclude: true }],
        annotations: [{ name: "annotation-key-1", value: "annotation-value-1", exclude: false }],
        labels: [{ name: "label-key-1", value: "label-value-1" }],
      });
    });

    it("should convert supplementalGroups string to undefined", () => {
      expect(
        workloadUtil.getWorkloadSpecificEnvFromUIWorkloadSpecificEnv({
          supplementalGroups: "123, 456",
        }),
      ).toStrictEqual({
        ...baseObj,
        supplementalGroups: "123, 456",
      });
    });

    it("should set the autoDeletionTimeAfterCompletionSeconds as number or null ", () => {
      expect(
        workloadUtil.getWorkloadSpecificEnvFromUIWorkloadSpecificEnv({
          autoDeletionTimeAfterCompletionSeconds: 34,
        }),
      ).toStrictEqual({
        ...baseObj,
        autoDeletionTimeAfterCompletionSeconds: 34,
      });

      expect(
        workloadUtil.getWorkloadSpecificEnvFromUIWorkloadSpecificEnv({
          autoDeletionTimeAfterCompletionSeconds: null,
        }),
      ).toStrictEqual({
        ...baseObj,
        autoDeletionTimeAfterCompletionSeconds: null,
      });
    });
  });

  describe("convertWorkloadToWorkloadUI", () => {
    it("should convert a backend single node workload to workload UI model", () => {
      const expectedResult: IUIWorkloadCreation = {
        name: "w100",
        namespace: "",
        clusterId: "aefb179a-4bee-420f-bb00-1da188a155fd",
        projectId: 8,
        assets: {
          environment: "55c1bdad-b01f-4635-a883-4a20f9c6763",
          compute: "23d90236-4687-47cc-9d0f-b6cffaa59bfb",
          datasources: [],
          uiVolumes: undefined,
        },
        specificEnv: {
          args: undefined,
          command: undefined,
          runAsUid: null,
          runAsGid: null,
          supplementalGroups: undefined,
          nodeType: null,
          allowOverQuota: null,
          nodePools: null,
          autoDeletionTimeAfterCompletionSeconds: null,
          connections: [],
          environmentVariables: [{ name: "env-key-1", value: "env-value-1" }],
          annotations: undefined,
          labels: undefined,
          backoffLimit: 6,
          tolerations: undefined,
          podAffinity: undefined,
          parallelism: undefined,
          completions: undefined,
          terminationGracePeriodSeconds: undefined,
        },
      };
      expect(workloadUtil.convertWorkloadToWorkloadUI(trainingSingleNodeResponseV2)).toStrictEqual(expectedResult);
    });

    it("should convert a backend multi node workload to workload UI model", () => {
      const expectedResult: IUIWorkloadCreation = {
        name: trainingMultiNodeCustomMasterResponseV2.meta.name,
        namespace: "",
        clusterId: trainingMultiNodeCustomMasterResponseV2.meta.clusterId,
        projectId: +trainingMultiNodeCustomMasterResponseV2.meta.projectId,
        assets: {
          environment: trainingMultiNodeCustomMasterResponseV2.assets.environment.id,
          compute: trainingMultiNodeCustomMasterResponseV2.assets.compute?.id || "",
          datasources: [],
          uiVolumes: undefined,
        },
        specificEnv: {
          args: undefined,
          command: undefined,
          runAsUid: null,
          runAsGid: null,
          supplementalGroups: undefined,
          nodeType: null,
          allowOverQuota: null,
          nodePools: ["default"],
          autoDeletionTimeAfterCompletionSeconds:
            trainingMultiNodeCustomMasterResponseV2.specificRunParams?.autoDeletionTimeAfterCompletionSeconds,
          connections: [],
          environmentVariables: trainingMultiNodeCustomMasterResponseV2.specificRunParams?.environmentVariables || [],
          annotations: undefined,
          labels: undefined,
          backoffLimit: 6,
          tolerations: undefined,
          podAffinity: undefined,
          parallelism: undefined,
          completions: undefined,
          terminationGracePeriodSeconds: undefined,
        },
        distributed: {
          noMaster: undefined,
          numWorkers: trainingMultiNodeCustomMasterResponseV2.distributed?.numWorkers,
          distFramework: trainingMultiNodeCustomMasterResponseV2.distributed?.framework,
          master: {
            assets: {
              environment: trainingMultiNodeCustomMasterResponseV2.distributed?.master?.assets?.environment.id || "",
              compute: trainingMultiNodeCustomMasterResponseV2.distributed?.master?.assets?.compute?.id || "",
              datasources: undefined,
              uiVolumes: undefined,
            },
            specificEnv: {
              args: undefined,
              command: undefined,
              runAsUid: null,
              runAsGid: null,
              supplementalGroups: undefined,
              nodeType: null,
              allowOverQuota: null,
              nodePools: ["default"],
              autoDeletionTimeAfterCompletionSeconds:
                trainingMultiNodeCustomMasterResponseV2.distributed?.master?.specificRunParams
                  ?.autoDeletionTimeAfterCompletionSeconds,
              connections: [],
              environmentVariables:
                trainingMultiNodeCustomMasterResponseV2.distributed?.master?.specificRunParams?.environmentVariables ||
                [],
              annotations: undefined,
              labels: undefined,
              backoffLimit: trainingMultiNodeCustomMasterResponseV2.distributed?.master?.specificRunParams?.backoffLimit,
              tolerations: undefined,
              podAffinity: undefined,
              parallelism: undefined,
              completions: undefined,
              terminationGracePeriodSeconds: undefined,
            },
          },
        },
        enableEditingMaster: true,
      };
      expect(workloadUtil.convertWorkloadToWorkloadUI(trainingMultiNodeCustomMasterResponseV2)).toStrictEqual(
        expectedResult,
      );
    });

    it("should convert a backend multi node workload to workload UI model without the master when noMaster sets to true", () => {
      const expectedResult: IUIWorkloadCreation = {
        name: trainingMultiNodeMasterSameAsWorkerResponseV2.meta.name,
        namespace: "",
        clusterId: trainingMultiNodeMasterSameAsWorkerResponseV2.meta.clusterId,
        projectId: +trainingMultiNodeMasterSameAsWorkerResponseV2.meta.projectId,
        assets: {
          environment: trainingMultiNodeMasterSameAsWorkerResponseV2.assets.environment.id,
          compute: trainingMultiNodeMasterSameAsWorkerResponseV2.assets.compute?.id || "",
          datasources: [],
          uiVolumes: undefined,
        },
        specificEnv: {
          args: undefined,
          command: undefined,
          runAsUid: null,
          runAsGid: null,
          supplementalGroups: undefined,
          nodeType: null,
          allowOverQuota: null,
          nodePools: ["default"],
          autoDeletionTimeAfterCompletionSeconds: 2592000,
          connections: [],
          environmentVariables: [{ name: "env-key-1", value: "env-value-1" }],
          annotations: undefined,
          labels: undefined,
          backoffLimit: 6,
          tolerations: undefined,
          podAffinity: undefined,
          parallelism: undefined,
          completions: undefined,
          terminationGracePeriodSeconds: undefined,
        },
        distributed: {
          numWorkers: 1,
          distFramework: DistributedFramework.Mpi,
          noMaster: true,
          master: null,
        },
        enableEditingMaster: undefined,
      };

      const trainingMultiNodeResponseNoMaster = deepCopy(trainingMultiNodeMasterSameAsWorkerResponseV2);
      if (trainingMultiNodeResponseNoMaster.distributed) {
        trainingMultiNodeResponseNoMaster.distributed.master = undefined;
      }

      expect(workloadUtil.convertWorkloadToWorkloadUI(trainingMultiNodeResponseNoMaster)).toStrictEqual(expectedResult);
    });

    it("should sets the enableEditingMaster to true when master is NOT equal to workers", () => {
      const trainingMultiNodeResponseDiffMaster = deepCopy(trainingMultiNodeCustomMasterResponseV2);

      const workloadUI = workloadUtil.convertWorkloadToWorkloadUI(trainingMultiNodeResponseDiffMaster);
      expect(workloadUI.enableEditingMaster).toBe(true);
    });

    it("should sets the enableEditingMaster to false when master is equal to workers", () => {
      const workloadUI = workloadUtil.convertWorkloadToWorkloadUI(trainingMultiNodeMasterSameAsWorkerResponseV2);
      expect(workloadUI.enableEditingMaster).toBe(false);
    });
  });

  describe("processCommonSpecificEnv", () => {
    it("should convert ui specificEnv to SpecificRunParams correctly", () => {
      const specificEnv: IUIWorkloadSpecificEnv = {
        args: "arg1 arg2",
        command: "cmd",
        runAsUid: 1000,
        runAsGid: 2000,
        supplementalGroups: "3000,4000",
        environmentVariables: [{ name: "env1", value: "value1", exclude: false }],
        annotations: [{ name: "anno1", value: "value1", exclude: true }],
        labels: [{ name: "label1", value: "value1", exclude: false }],
        allowOverQuota: true,
        connections: null,
        nodeType: "my-node-type",
        nodePools: ["node-pool-1", "node-pool-2"],
        autoDeletionTimeAfterCompletionSeconds: 54,
        backoffLimit: 32,
      };

      const result = workloadUtil.processCommonSpecificEnv(specificEnv, []);

      expect(result).toEqual({
        args: "arg1 arg2",
        command: "cmd",
        runAsUid: specificEnv.runAsUid,
        runAsGid: specificEnv.runAsGid,
        supplementalGroups: specificEnv.supplementalGroups,
        environmentVariables: [{ name: "env1", value: "value1", exclude: false }],
        annotations: [{ name: "anno1", value: "value1", exclude: true }],
        labels: [{ name: "label1", value: "value1", exclude: false }],
        allowOverQuota: true,
        connections: null,
        nodeType: "my-node-type",
        nodePools: ["node-pool-1", "node-pool-2"],
        autoDeletionTimeAfterCompletionSeconds: 54,
        backoffLimit: 32,
      });
    });

    it("should omit specified fields and process the remaining fields correctly", () => {
      const specificEnv: IUIWorkloadSpecificEnv = {
        args: "arg1 arg2",
        command: "cmd",
        runAsUid: 1000,
        runAsGid: 2000,
        supplementalGroups: "3000,4000",
        environmentVariables: [{ name: "env1", value: "value1", exclude: false }],
        annotations: [{ name: "anno1", value: "value1", exclude: true }],
        labels: [{ name: "label1", value: "value1", exclude: false }],
        autoScaleData: {
          minReplicas: 1,
          maxReplicas: 2,
          thresholdMetric: undefined,
          thresholdValue: undefined,
          scaleToZeroRetentionSeconds: undefined,
        },
        allowOverQuota: true,
        connections: null,
        nodeType: "my-node-type",
        nodePools: ["node-pool-1", "node-pool-2"],
        autoDeletionTimeAfterCompletionSeconds: 54,
        backoffLimit: 32,
      };

      const omitFields: SpecificRunParamsKeys = ["autoScaleData", "allowOverQuota", "backoffLimit"];
      const result = workloadUtil.processCommonSpecificEnv(specificEnv, omitFields);

      expect(result).toEqual({
        args: "arg1 arg2",
        command: "cmd",
        runAsUid: specificEnv.runAsUid,
        runAsGid: specificEnv.runAsGid,
        supplementalGroups: specificEnv.supplementalGroups,
        environmentVariables: [{ name: "env1", value: "value1", exclude: false }],
        annotations: [{ name: "anno1", value: "value1", exclude: true }],
        labels: [{ name: "label1", value: "value1", exclude: false }],
        connections: null,
        nodeType: "my-node-type",
        nodePools: ["node-pool-1", "node-pool-2"],
        autoDeletionTimeAfterCompletionSeconds: 54,
      });
    });

    it("should set args and command to null, runAsUid, runAsGid and supplementalGroups to undefined if they are not provided", () => {
      const specificEnv = {};
      const omitFields: SpecificRunParamsKeys = [];

      const result = workloadUtil.processCommonSpecificEnv(specificEnv, omitFields);

      expect(result.args).toBeNull();
      expect(result.command).toBeNull();
      expect(result.runAsUid).toBeUndefined();
      expect(result.runAsGid).toBeUndefined();
      expect(result.supplementalGroups).toBeUndefined();
      expect(Object.keys(result).length).toBe(5);
    });
  });

  describe("getWorkloadSpecificEnvFromUIWorkloadSpecificEnv", () => {
    it("should call processCommonSpecificEnv with the correct omitFields", () => {
      const specificEnv: IUIWorkloadSpecificEnv = {
        args: "arg1 arg2",
        command: "cmd",
        runAsUid: 1000,
        runAsGid: 2000,
        supplementalGroups: "3000,4000",
        environmentVariables: [{ name: "env1", value: "value1", exclude: false }],
        annotations: [{ name: "anno1", value: "value1", exclude: true }],
        labels: [{ name: "label1", value: "value1", exclude: false }],
        allowOverQuota: true,
        connections: null,
        nodeType: "my-node-type",
        nodePools: ["node-pool-1", "node-pool-2"],
        autoDeletionTimeAfterCompletionSeconds: 54,
        backoffLimit: 32,
        tolerations: undefined,
      };
      const result = workloadUtil.getWorkloadSpecificEnvFromUIWorkloadSpecificEnv(specificEnv);

      expect(result).toEqual({
        args: "arg1 arg2",
        command: "cmd",
        runAsUid: specificEnv.runAsUid,
        runAsGid: specificEnv.runAsGid,
        supplementalGroups: specificEnv.supplementalGroups,
        environmentVariables: [{ name: "env1", value: "value1", exclude: false }],
        annotations: [{ name: "anno1", value: "value1", exclude: true }],
        labels: [{ name: "label1", value: "value1", exclude: false }],
        allowOverQuota: true,
        connections: null,
        nodeType: "my-node-type",
        nodePools: ["node-pool-1", "node-pool-2"],
        autoDeletionTimeAfterCompletionSeconds: 54,
        backoffLimit: 32,
        tolerations: undefined,
      });

      expect(result.autoScaling).toBeUndefined();
    });

    it("should add autoScaling if autoScaleData is provided", () => {
      const specificEnv: IUIWorkloadSpecificEnv = {
        args: "arg1 arg2",
        command: "cmd",
        runAsUid: 1000,
        runAsGid: 2000,
        supplementalGroups: "3000,4000",
        environmentVariables: [{ name: "env1", value: "value1", exclude: false }],
        annotations: [{ name: "anno1", value: "value1", exclude: true }],
        labels: [{ name: "label1", value: "value1", exclude: false }],
        allowOverQuota: true,
        connections: null,
        nodeType: "my-node-type",
        nodePools: ["node-pool-1", "node-pool-2"],
        autoDeletionTimeAfterCompletionSeconds: 54,
        backoffLimit: 32,
        autoScaleData: {
          minReplicas: 1,
          maxReplicas: 2,
          thresholdMetric: undefined,
          thresholdValue: undefined,
          scaleToZeroRetentionSeconds: undefined,
        },
        tolerations: undefined,
      };
      const result = workloadUtil.getWorkloadSpecificEnvFromUIWorkloadSpecificEnv(specificEnv);

      expect(result).toEqual({
        args: "arg1 arg2",
        command: "cmd",
        runAsUid: specificEnv.runAsUid,
        runAsGid: specificEnv.runAsGid,
        supplementalGroups: specificEnv.supplementalGroups,
        environmentVariables: [{ name: "env1", value: "value1", exclude: false }],
        annotations: [{ name: "anno1", value: "value1", exclude: true }],
        labels: [{ name: "label1", value: "value1", exclude: false }],
        allowOverQuota: true,
        connections: null,
        nodeType: "my-node-type",
        nodePools: ["node-pool-1", "node-pool-2"],
        autoDeletionTimeAfterCompletionSeconds: 54,
        backoffLimit: 32,
        tolerations: undefined,
      });
    });
  });

  describe("adjustWorkloadPhaseFilter", () => {
    const testAdjustWorkloadPhaseFilter = (input: IAdvancedFilterModel[], expectedSelectedValues: string[]) => {
      const expected = [
        {
          selectedValues: expectedSelectedValues,
          label: "Status",
          name: "phase",
        } as IEnumBasedFilterModel,
      ];
      const result = workloadUtil.adjustWorkloadPhaseFilter(input);
      expect(result).toEqual(expected);
    };

    it("should adjust selected filter with running phase to contain additional hidden phases", () => {
      testAdjustWorkloadPhaseFilter(advancedFilterPhaseMockRunning, ["Running", "Resuming", "Updating", "Terminating"]);
    });

    it("should adjust selected filter with Stopped phase to contain additional hidden phases", () => {
      testAdjustWorkloadPhaseFilter(advancedFilterPhaseMockStopped, ["Stopped", "Stopping", "Terminating"]);
    });

    it("should always add Terminating phase to the selected filter", () => {
      testAdjustWorkloadPhaseFilter(advancedFilterPhaseMockPending, ["Pending", "Terminating"]);
    });

    it("should not add any additional hidden phases when selected filter is empty", () => {
      const result = workloadUtil.adjustWorkloadPhaseFilter([]);
      expect(result).toEqual([]);
    });
  });

  describe("applyUserCreatedWorkloadsFilter", () => {
    beforeEach(() => {
      filterService.setColumnAdvancedFilter = vi.fn();
      filterService.saveFilters = vi.fn();
    });
    afterEach(() => {
      vi.resetAllMocks();
    });
    it("should apply filter for first login and filter is required", () => {
      const spySaveFilter = vi.spyOn(storageUtil, "save");

      const userName = "testUser";
      const filterBy = {};
      const submittedByFilter = {
        term: userName,
        value: EFilterOperator.Equals,
        label: "Created by",
        name: WorkloadSortFilterFields.SubmittedBy,
        type: EColumnFilterType.FreeText,
      };

      storageUtil.get = vi.fn().mockReturnValue(true);
      filterService.loadFilters = vi.fn().mockReturnValue(filterBy);
      filterService.setColumnAdvancedFilter = vi.fn();
      filterService.saveFilters = vi.fn();

      workloadUtil.applyUserCreatedWorkloadsFilter(true, userName);

      expect(filterService.loadFilters).toHaveBeenCalledWith(window.location, ETableFilters.WORKLOAD, {});
      expect(filterService.setColumnAdvancedFilter).toHaveBeenCalledWith(
        filterBy,
        submittedByFilter,
        ETableFilters.WORKLOAD,
      );
      expect(filterService.saveFilters).toHaveBeenCalledWith(ETableFilters.WORKLOAD, filterBy);
      expect(spySaveFilter).toHaveBeenCalledWith(STORAGE_KEYS.IS_FIRST_LOGIN, false);
    });

    it("should apply filter for first login and filter is not required", () => {
      const userName = "testUser";
      const filterBy = { advancedFilters: ["filter1", "filter2"] };

      storageUtil.get = vi.fn().mockReturnValue(true);
      filterService.loadFilters = vi.fn().mockReturnValue(filterBy);
      filterService.saveFilters = vi.fn();

      workloadUtil.applyUserCreatedWorkloadsFilter(false, userName);

      expect(filterService.loadFilters).toHaveBeenCalledWith(window.location, ETableFilters.WORKLOAD, {});
      expect(filterService.saveFilters).toHaveBeenCalledWith(ETableFilters.WORKLOAD, { advancedFilters: [] });
    });

    it("should not apply filter for non-first login", () => {
      const userName = "testUser";

      storageUtil.get = vi.fn().mockReturnValue(false);
      filterService.loadFilters = vi.fn();
      filterService.saveFilters = vi.fn();

      workloadUtil.applyUserCreatedWorkloadsFilter(true, userName);

      expect(filterService.loadFilters).not.toHaveBeenCalled();
      expect(filterService.saveFilters).not.toHaveBeenCalled();
    });
  });

  describe("getWorkloadIcon", () => {
    it("should return distributed framework icon if exists", () => {
      const result = workloadUtil.getWorkloadIcon(distributedWorkload);
      expect(result).toBe("pytorch");
    });

    it("should return Hugging Face icon for inference workload with old LLM image when is it a assets based HF workload", () => {
      const result = workloadUtil.getWorkloadIcon(llmWorkload);
      expect(result).toBe("hugging-face");
    });

    it("should return VLLM icon for inference workload with LLM image when is it a flat based HF workload", () => {
      const result = workloadUtil.getWorkloadIcon({
        ...llmWorkload,
        environments: undefined,
        images: [LLM_IMAGE_FOR_HUGGING_FACE],
      });
      expect(result).toBe("vllm");
    });

    it("should return TGI icon for inference workload with TGI image when is it a flat based HF workload", () => {
      const result = workloadUtil.getWorkloadIcon({
        ...llmWorkload,
        environments: undefined,
        images: [LLM_IMAGE_FOR_TGI],
      });
      expect(result).toBe("tgi");
    });

    it("should return the generic icon when the legacy Hugging Face signature does not match", () => {
      const workload = {
        ...llmWorkload,
        environmentVariables: {
          RUNAI_MODEL: "someModel",
          RUNAI_MODEL_NAME: "someModelName",
        },
      };

      const result = workloadUtil.getWorkloadIcon(workload);
      expect(result).toBe(GENERIC_WORKLOAD_ICON);
    });

    it("should return toolType icon if exists", () => {
      const result = workloadUtil.getWorkloadIcon(workspaceWorkload);
      expect(result).toBe("tensorboard");
    });

    it("should return custom-tool if no distributed framework or toolType exists", () => {
      const workload = {
        ...llmWorkload,
        connections: [],
        environmentVariables: {},
        type: "custom",
      };

      const result = workloadUtil.getWorkloadIcon(workload);
      expect(result).toBe("custom-tool");
    });
  });

  describe("_getWorkloadUI", () => {
    const meta: IWorkloadMeta = {
      name: "Test Workload",
      namespace: "default",
      clusterId: "cluster123",
      projectId: 123,
    };

    const defaultAssets: AssetsIds = {
      environment: "env-id-1",
      compute: "compute-id-1",
      datasources: [
        {
          id: "datasource1",
          kind: AssetKind.Nfs,
        },
      ],
    };

    const defaultSpecificEnv: SpecificRunParams = {
      args: undefined,
      command: undefined,
      runAsUid: null,
      runAsGid: null,
      supplementalGroups: undefined,
      nodeType: null,
      allowOverQuota: null,
      nodePools: null,
      autoDeletionTimeAfterCompletionSeconds: null,
      connections: undefined,
      environmentVariables: undefined,
      annotations: undefined,
      labels: undefined,
      backoffLimit: 6,
      tolerations: undefined,
      podAffinity: undefined,
      parallelism: undefined,
      completions: undefined,
    };

    const distributedSetup: IUIDistributedUISetup = {
      distributedData: {
        numWorkers: 4,
        master: {
          assets: {
            environment: {
              id: "e123",
              name: "my-env",
            },
          },
          specificRunParams: {},
        },
      },
      noMaster: false,
      enableEditingMaster: true,
    };

    it("should create workload UI without distributed setup", () => {
      const currSpecificEnv: SpecificRunParams = {
        command: "start",
        runAsUid: 1000,
        runAsGid: 2000,
      };
      const result: IUIWorkloadCreation = workloadUtil._getWorkloadUI(meta, defaultAssets, currSpecificEnv);

      expect(result).toEqual({
        name: meta.name,
        namespace: meta.namespace,
        clusterId: meta.clusterId,
        projectId: meta.projectId,
        assets: defaultAssets,
        specificEnv: {
          ...defaultSpecificEnv,
          command: currSpecificEnv.command,
          runAsUid: currSpecificEnv.runAsUid,
          runAsGid: currSpecificEnv.runAsGid,
        },
      });
    });

    it("should create workload UI with distributed setup", () => {
      const mockSpecificEnv = { specificEnv: { ...defaultSpecificEnv } };
      const mockAssets = { assets: defaultAssets };

      const result: IUIWorkloadCreation = workloadUtil._getWorkloadUI(
        meta,
        mockAssets.assets,
        mockSpecificEnv.specificEnv,
        undefined,
        distributedSetup,
      );

      expect(result).toEqual({
        name: meta.name,
        namespace: meta.namespace,
        clusterId: meta.clusterId,
        projectId: meta.projectId,
        assets: {
          ...mockAssets.assets,
          uiVolumes: undefined,
        },
        specificEnv: mockSpecificEnv.specificEnv,
        distributed: {
          numWorkers: distributedSetup.distributedData?.numWorkers,
          distFramework: undefined,
          master: {
            assets: {
              compute: null,
              datasources: undefined,
              environment: distributedSetup.distributedData?.master?.assets?.environment.id,
              uiVolumes: undefined,
            },
            specificEnv: defaultSpecificEnv,
          },
          noMaster: distributedSetup.noMaster,
        },
        enableEditingMaster: distributedSetup.enableEditingMaster,
      });
    });

    it("should handle case where distributed setup has no master", () => {
      const distributedSetupNoMaster: IUIDistributedUISetup = {
        distributedData: {
          numWorkers: 4,
          master: null,
        },
        noMaster: true,
        enableEditingMaster: true,
      };

      const mockSpecificEnv = { specificEnv: { ...defaultSpecificEnv } };
      const mockAssets = { assets: defaultAssets };

      const result: IUIWorkloadCreation = workloadUtil._getWorkloadUI(
        meta,
        mockAssets.assets,
        mockSpecificEnv.specificEnv,
        undefined,
        distributedSetupNoMaster,
      );

      expect(result).toEqual({
        name: meta.name,
        namespace: meta.namespace,
        clusterId: meta.clusterId,
        projectId: meta.projectId,
        assets: {
          ...mockAssets.assets,
          uiVolumes: undefined,
        },
        specificEnv: mockSpecificEnv.specificEnv,
        distributed: {
          numWorkers: distributedSetup.distributedData?.numWorkers,
          distFramework: undefined,
          master: null,
          noMaster: distributedSetupNoMaster.noMaster,
        },
        enableEditingMaster: distributedSetup.enableEditingMaster,
      });
    });

    it("should handle case where distributed setup is not provided", () => {
      const result: IUIWorkloadCreation = workloadUtil._getWorkloadUI(
        meta,
        defaultAssets,
        defaultSpecificEnv,
        undefined,
        undefined,
      );

      expect(result).not.toHaveProperty("distributed");
    });
  });

  describe("_getWorkloadConnectionsForInference", () => {
    const servingPortUrl = "http://i100.runai-p100.svc.cluster.local";

    it("should add new connections if no ServingPort connection exists and workload.urls has more items than connections", () => {
      const connections: Connection[] = [
        {
          name: "ChatbotUI",
          toolType: "chatbot-ui",
          connectionType: "ExternalUrl",
          url: "https://cluster.example.com/p100/i100",
        },
      ];
      const result = workloadUtil._getWorkloadConnectionsForInference(connectionsMocks.inference, connections);

      expect(result).toHaveLength(2);
      expect(result[1]).toMatchObject({
        url: servingPortUrl,
        name: "Serving Endpoint",
        toolType: "custom",
        connectionType: InternalConnectionType.ServingPort,
      });
    });

    it("should not add new ServingPort connections if one already exists", () => {
      const currWorkload = {
        ...connectionsMocks.inference,
      };

      let connections: Connection[] = [];
      if (currWorkload.environments?.length) {
        currWorkload.environments[0].connections?.push({
          url: servingPortUrl,
          name: "Serving Endpoint",
          toolType: "custom",
          connectionType: InternalConnectionType.ServingPort,
        });
        connections = currWorkload.environments[0].connections || [];
      }

      const result = workloadUtil._getWorkloadConnectionsForInference(connectionsMocks.inference, connections);

      expect(connections.length).toBe(2);
      expect(result).toHaveLength(2); // No new connections should be added
      expect(result[0]).toMatchObject(connections[0]);
      expect(result[1]).toMatchObject(connections[1]);
    });

    it("should set the URL if clusterDomain is provided for connections of type ExternalUrl", () => {
      const connections = [
        {
          url: "initial-url",
          name: "External Connection",
          toolType: "custom",
          connectionType: InternalConnectionType.ExternalUrl,
        },
      ];
      const clusterDomain = "https://example.com";
      const result = workloadUtil._getWorkloadConnectionsForInference(
        connectionsMocks.inference,
        connections,
        clusterDomain,
      );

      expect(result[0].url).toBe(
        `${clusterDomain}/${connectionsMocks.inference.projectName}/${connectionsMocks.inference.name}`,
      );
    });

    it("should keep the added ServingPort connection when clusterDomain is provided", () => {
      const connections: Connection[] = [
        {
          name: "ChatbotUI",
          toolType: "chatbot-ui",
          connectionType: "ExternalUrl",
          url: "https://cluster.example.com/p100/i100",
        },
      ];
      const clusterDomain = "https://example.com";
      const result = workloadUtil._getWorkloadConnectionsForInference(
        connectionsMocks.inference,
        connections,
        clusterDomain,
      );

      expect(result).toHaveLength(2);
      expect(result[0].url).toBe(
        `${clusterDomain}/${connectionsMocks.inference.projectName}/${connectionsMocks.inference.name}`,
      );
      expect(result[1]).toMatchObject({
        url: servingPortUrl,
        name: "Serving Endpoint",
        toolType: "custom",
        connectionType: InternalConnectionType.ServingPort,
      });
    });

    it("should not add new connections if workload.urls has fewer or equal items as connections", () => {
      const currWorkload = {
        ...connectionsMocks.inference,
        urls: [],
      };
      const connections = [
        {
          url: "url1",
          name: "Initial Connection",
          toolType: "custom",
          connectionType: InternalConnectionType.ExternalUrl,
        },
      ];
      const result = workloadUtil._getWorkloadConnectionsForInference(currWorkload, connections);
      expect(result).toHaveLength(1);
      expect(result[0]).toMatchObject(connections[0]);
    });

    it("should not update URLs for external connections if clusterDomain is not provided", () => {
      const connections = [
        {
          url: "initial-url",
          name: "External Connection",
          toolType: "custom",
          connectionType: InternalConnectionType.ExternalUrl,
        },
      ];
      const result = workloadUtil._getWorkloadConnectionsForInference(connectionsMocks.inference, connections);

      expect(result[0].url).toBe("initial-url");
    });

    it("should not add a ServingPort connection when several workload URLs are unclaimed", () => {
      const currWorkload = {
        ...connectionsMocks.inference,
        urls: ["https://x/tool", "https://x/serve"],
      };
      const connections: Connection[] = [{ name: "ChatbotUI", toolType: "chatbot-ui", connectionType: "ExternalUrl" }];

      const result = workloadUtil._getWorkloadConnectionsForInference(currWorkload, connections);

      expect(result).toEqual(connections);
    });

    it("should not add a ServingPort connection when the workload has no URLs", () => {
      const currWorkload = {
        ...connectionsMocks.inference,
        urls: undefined,
      };
      const connections: Connection[] = [];

      const result = workloadUtil._getWorkloadConnectionsForInference(currWorkload, connections);

      expect(result).toEqual([]);
    });

    it("should not mutate the input connections when clusterDomain is provided", () => {
      const connections: Connection[] = [
        {
          name: "ChatbotUI",
          toolType: "chatbot-ui",
          connectionType: "ExternalUrl",
          url: "https://cluster.example.com/p100/i100",
        },
      ];
      const connectionsBefore = deepCopy(connections);

      workloadUtil._getWorkloadConnectionsForInference(connectionsMocks.inference, connections, "https://example.com");

      expect(connections).toEqual(connectionsBefore);
    });

    it("should return the original list unchanged if no conditions are met", () => {
      const currWorkload = {
        ...connectionsMocks.inference,
        urls: [],
      };
      const connections = [
        {
          url: "initial-url",
          name: "Initial Connection",
          toolType: "custom",
          connectionType: InternalConnectionType.ExternalUrl,
        },
      ];
      const result = workloadUtil._getWorkloadConnectionsForInference(currWorkload, connections);

      expect(result).toEqual(connections);
    });
  });

  describe("getUnclaimedServingUrl", () => {
    const toolUrl = "https://x/tool";
    const serveUrl = "https://x/serve";

    it("should return the only URL no connection claims", () => {
      const connections: Connection[] = [
        { name: "ChatbotUI", toolType: "chatbot-ui", connectionType: "ExternalUrl", url: toolUrl },
        { name: "ServingPort", toolType: "custom", connectionType: "ServingPort", containerPort: 8000 },
      ];

      expect(workloadUtil.getUnclaimedServingUrl([toolUrl, serveUrl], connections)).toBe(serveUrl);
    });

    it("should treat known URLs as claimed", () => {
      const connections: Connection[] = [{ name: "ChatbotUI", toolType: "chatbot-ui", connectionType: "ExternalUrl" }];

      expect(workloadUtil.getUnclaimedServingUrl([toolUrl, serveUrl], connections, [toolUrl])).toBe(serveUrl);
    });

    it("should return undefined when several URLs are unclaimed", () => {
      const connections: Connection[] = [{ name: "ChatbotUI", toolType: "chatbot-ui", connectionType: "ExternalUrl" }];

      expect(workloadUtil.getUnclaimedServingUrl([toolUrl, serveUrl], connections)).toBeUndefined();
    });

    it("should return undefined when there are no URLs", () => {
      expect(workloadUtil.getUnclaimedServingUrl(undefined, [])).toBeUndefined();
    });
  });

  describe("getWorkloadConnectionsForGrid", () => {
    describe("non inference workload", () => {
      it("should reduce and filter connections from environments if they are present", () => {
        const result = workloadUtil.getWorkloadConnectionsForGrid(connectionsMocks.workspace);

        expect(result).toHaveLength(1);
        const environments = connectionsMocks.workspace.environments || [];
        const connections = environments[0].connections || [];
        expect(result[0].url).toBe(connections[0].url);
      });

      it("should use externalConnections if environments are not present", () => {
        const currWorkload: Workload = {
          ...connectionsMocks.workspace,
          environments: undefined,
          externalConnections: [
            {
              url: "external-url1",
              connectionType: InternalConnectionType.NodePort,
              name: "",
              toolType: "",
            },
          ],
        };

        const result = workloadUtil.getWorkloadConnectionsForGrid(currWorkload);

        expect(result).toEqual(currWorkload.externalConnections);
      });

      it("should prefer environments[].connections[] over endpoints[] when both are present", () => {
        const currWorkload: Workload = {
          ...connectionsMocks.workspace,
          endpoints: [{ host: "10.62.206.9", port: 32670, url: "https://cluster-2.example.com/p100/w100" }],
        };

        const result = workloadUtil.getWorkloadConnectionsForGrid(currWorkload);

        expect(result).toHaveLength(1);
        expect(result[0].name).toBe("TensorBoard");
        expect(result[0].toolType).toBe("tensorboard");
        expect(result[0].connectionType).toBe("ExternalUrl");
        expect(result[0].containerPort).toBe(6006);
      });

      it("should fall back to endpoints[] when environments are absent", () => {
        const currWorkload: Workload = {
          ...connectionsMocks.workspace,
          environments: undefined,
          externalConnections: undefined,
          endpoints: [{ host: "10.62.206.9", port: 32670, url: "https://example/x" }],
        };

        const result = workloadUtil.getWorkloadConnectionsForGrid(currWorkload);

        expect(result).toHaveLength(1);
        expect(result[0].url).toBe("https://example/x");
        expect(result[0].containerPort).toBe(32670);
        expect(result[0].name).toBe("10.62.206.9");
      });
    });

    describe("an inference workload", () => {
      it("should filter out ServingPort connections without URLs in old clusters", () => {
        const workload: Workload = {
          ...connectionsMocks.inference,
          environments: [
            {
              connections: [
                {
                  url: "",
                  connectionType: InternalConnectionType.ServingPort,
                  toolType: "-ui",
                  name: "should-be-filtered-out",
                }, // Should be filtered out
              ],
              id: "1",
              name: "env-1",
            },
          ],
          urls: [],
        };
        const result = workloadUtil.getWorkloadConnectionsForGrid(workload);
        expect(result).toHaveLength(0);
      });

      it("should return an empty array if no connections or environments are present", () => {
        const workload = {
          ...connectionsMocks.inference,
          environments: [],
          externalConnections: [],
          urls: [],
        };
        const result = workloadUtil.getWorkloadConnectionsForGrid(workload);
        expect(result).toEqual([]);
      });
    });
  });

  describe("flattenWorkloadPhaseIssues", () => {
    it("should return an empty array when issues string is empty", () => {
      const result = workloadUtil.flattenWorkloadPhaseIssues("");
      expect(result).toEqual([]);
    });

    it("should return an array of issues when issues string is not empty", () => {
      const issues =
        '{"network":null,"storage":null,"compute":[{"kind":"TFJob","name":"runai-idle-gpu-0-1740563647","description":"[TFJob:runai-idle-gpu-0-1740563647] TFJob runai-alon/runai-idle-gpu-0-1740563647 has failed because 1 Chief replica(s) failed."}],"internal":null}';
      const result = workloadUtil.flattenWorkloadPhaseIssues(issues);
      expect(result).toEqual([
        {
          kind: "TFJob",
          name: "runai-idle-gpu-0-1740563647",
          description:
            "[TFJob:runai-idle-gpu-0-1740563647] TFJob runai-alon/runai-idle-gpu-0-1740563647 has failed because 1 Chief replica(s) failed.",
          parentName: "Compute",
        },
      ]);
    });

    it("should return an array of issues when issues string is not empty and has multiple issues", () => {
      const issues =
        '{"network":null,"storage":null,"compute":[{"kind":"TFJob","name":"runai-idle-gpu-0-1740563647","description":"[TFJob:runai-idle-gpu-0-1740563647] TFJob runai-alon/runai-idle-gpu-0-1740563647 has failed because 1 Chief replica(s) failed."},{"kind":"TFJob","name":"runai-idle-gpu-0-1740563648","description":"[TFJob:runai-idle-gpu-0-1740563648] TFJob runai-alon/runai-idle-gpu-0-1740563648 has failed because 1 Chief replica(s) failed."}],"internal":null}';
      const result = workloadUtil.flattenWorkloadPhaseIssues(issues);
      expect(result).toEqual([
        {
          kind: "TFJob",
          name: "runai-idle-gpu-0-1740563647",
          description:
            "[TFJob:runai-idle-gpu-0-1740563647] TFJob runai-alon/runai-idle-gpu-0-1740563647 has failed because 1 Chief replica(s) failed.",
          parentName: "Compute",
        },
        {
          kind: "TFJob",
          name: "runai-idle-gpu-0-1740563648",
          description:
            "[TFJob:runai-idle-gpu-0-1740563648] TFJob runai-alon/runai-idle-gpu-0-1740563648 has failed because 1 Chief replica(s) failed.",
          parentName: "Compute",
        },
      ]);
    });

    it("should return an array of issues when issues string is not empty and has multiple issues with different parent names", () => {
      const issues =
        '{"network":null,"compute":null,"Storage":[{"kind":"TFJob","name":null,"description":"[TFJob:runai-idle-gpu-0-1740563647] TFJob runai-alon/runai-idle-gpu-0-1740563647 has failed because 1 Chief replica(s) failed."}],"internal":[{"kind":"TFJob","name":"runai-idle-gpu-0-1740563648","description":"[TFJob:runai-idle-gpu-0-1740563648] TFJob runai-alon/runai-idle-gpu-0-1740563648 has failed because 1 Chief replica(s) failed."}]}';
      const result = workloadUtil.flattenWorkloadPhaseIssues(issues);
      expect(result).toEqual([
        {
          kind: "TFJob",
          name: null,
          description:
            "[TFJob:runai-idle-gpu-0-1740563647] TFJob runai-alon/runai-idle-gpu-0-1740563647 has failed because 1 Chief replica(s) failed.",
          parentName: "Storage",
        },
        {
          kind: "TFJob",
          name: "runai-idle-gpu-0-1740563648",
          description:
            "[TFJob:runai-idle-gpu-0-1740563648] TFJob runai-alon/runai-idle-gpu-0-1740563648 has failed because 1 Chief replica(s) failed.",
          parentName: "Internal",
        },
      ]);
    });

    it("should handle case where issues string is not a valid JSON", () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      const issues = '{"key": "value",}';
      const result = workloadUtil.flattenWorkloadPhaseIssues(issues);
      expect(errorSpy).toHaveBeenCalled();
      expect(result).toEqual([]);
      errorSpy.mockRestore();
    });
  });

  describe("isNIMWorkload", () => {
    it("should return true for inference workload with NIM image prefix", () => {
      const result = workloadUtil.isNIMWorkload(nimWorkloadWithNimImage);
      expect(result).toBe(true);
    });

    it("should return true for inference workload with nimWorkload flag set to 'true'", () => {
      const result = workloadUtil.isNIMWorkload(nimWorkloadWithNimWorkloadFlag);
      expect(result).toBe(true);
    });

    it("should return false when an explicit vLLM marker accompanies a NIM image", () => {
      const workload = {
        ...nimWorkloadWithNimImage,
        additionalFields: {
          vllmWorkload: "true",
        },
      } as Workload;

      expect(workloadUtil.isNIMWorkload(workload)).toBe(false);
    });

    it("should prefer an explicit boolean NIM marker when both engine markers are true", () => {
      const workload = {
        ...nonNimInferenceWorkload,
        additionalFields: {
          nimWorkload: true,
          vllmWorkload: true,
        },
      } as Workload;

      expect(workloadUtil.isNIMWorkload(workload)).toBe(true);
      expect(workloadUtil.isVLLMWorkload(workload)).toBe(false);
    });

    it("should return false for non-inference workload with NIM image", () => {
      const result = workloadUtil.isNIMWorkload(nonInferenceWorkload);
      expect(result).toBe(true);
    });

    it("should return false for inference workload without NIM image or flag", () => {
      const result = workloadUtil.isNIMWorkload(nonNimInferenceWorkload);
      expect(result).toBe(false);
    });

    it("should return false for inference workload with nimWorkload flag set to 'false'", () => {
      const workload = {
        ...nonNimInferenceWorkload,
        additionalFields: {
          nimWorkload: "false",
        },
      };
      const result = workloadUtil.isNIMWorkload(workload);
      expect(result).toBe(false);
    });

    it("should return false for inference workload with nimWorkload flag set to invalid value", () => {
      const workload = {
        ...nonNimInferenceWorkload,
        additionalFields: {
          nimWorkload: "invalid",
        },
      };
      const result = workloadUtil.isNIMWorkload(workload);
      expect(result).toBe(false);
    });

    it("should return false for inference workload with empty images array", () => {
      const workload = {
        ...nonNimInferenceWorkload,
        images: [],
      };
      const result = workloadUtil.isNIMWorkload(workload);
      expect(result).toBe(false);
    });

    it("should return false for inference workload with undefined images", () => {
      const workload = {
        ...nonNimInferenceWorkload,
        images: undefined,
      };
      const result = workloadUtil.isNIMWorkload(workload);
      expect(result).toBe(false);
    });

    it("should return true for inference workload with image that contains NIM prefix anywhere in the name", () => {
      const workload = {
        ...nonNimInferenceWorkload,
        images: ["some-registry.com/nvcr.io/nim/model:latest"],
      };
      const result = workloadUtil.isNIMWorkload(workload);
      expect(result).toBe(true);
    });

    it("should return true for inference workload with image that starts with NIM prefix", () => {
      const workload = {
        ...nonNimInferenceWorkload,
        images: ["nvcr.io/nim/model:latest"],
      };
      const result = workloadUtil.isNIMWorkload(workload);
      expect(result).toBe(true);
    });

    it("should return true for inference workload with image that starts with NIM prefix and has additional path", () => {
      const workload = {
        ...nonNimInferenceWorkload,
        images: ["nvcr.io/nim/mistralai/mistral-7b-instruct:latest"],
      };
      const result = workloadUtil.isNIMWorkload(workload);
      expect(result).toBe(true);
    });
  });

  describe("isVLLMWorkload", () => {
    const vllmInferenceWorkload = {
      ...nonNimInferenceWorkload,
      type: EWorkloadType.Inference,
      images: ["vllm/vllm-openai:v0.10.0"],
    } as Workload;

    it.each([
      [{ additionalFields: { "run.ai/vllm-workload": "true" } }, true],
      [{ additionalFields: { vllmWorkload: "true" } }, true],
      [{ additionalFields: { vllmWorkload: true } }, true],
      [{ additionalFields: {}, images: ["registry.example.com/vllm/vllm-openai:v0.10.0"] }, false],
      [{ images: ["registry.example.com/vllm/vllm-openai:v0.10.0"] }, false],
      [{ additionalFields: { vllmWorkload: "false" } }, false],
      [{ additionalFields: { "run.ai/vllm-workload": "false" } }, false],
      [
        {
          additionalFields: { vllmWorkload: "false" },
          images: ["registry.example.com/vllm/vllm-openai:v0.10.0"],
        },
        false,
      ],
      [{ additionalFields: { nimWorkload: "true", vllmWorkload: "true" } }, false],
      [{ additionalFields: { nimWorkload: true, vllmWorkload: true } }, false],
      [{ additionalFields: { "run.ai/nim-workload": "true" } }, false],
      [{ images: ["nvcr.io/nim/vllm/vllm-openai:latest"] }, false],
    ])("shows vLLM metrics only for workloads with the explicit metadata marker", (overrides, expected) => {
      expect(workloadUtil.isVLLMWorkload({ ...vllmInferenceWorkload, ...overrides } as Workload)).toBe(expected);
    });

    it("does not classify non-inference workloads", () => {
      expect(
        workloadUtil.isVLLMWorkload({
          ...vllmInferenceWorkload,
          type: EWorkloadType.Training,
          additionalFields: { vllmWorkload: "true" },
        } as Workload),
      ).toBe(false);
    });
  });

  describe("_getNimInferenceIcon", () => {
    it("should return NvidiaNim icon for NIM workload with NIM image", () => {
      const result = workloadUtil._getNimInferenceIcon(nimWorkloadWithNimImage);
      expect(result).toBe("nvidianim");
    });

    it("should return NvidiaNim icon for NIM workload with nimWorkload flag", () => {
      const result = workloadUtil._getNimInferenceIcon(nimWorkloadWithNimWorkloadFlag);
      expect(result).toBe("nvidianim");
    });

    it("should return undefined for non-NIM inference workload", () => {
      const result = workloadUtil._getNimInferenceIcon(nonNimInferenceWorkload);
      expect(result).toBeUndefined();
    });

    it("should return undefined for non-inference workload", () => {
      const result = workloadUtil._getNimInferenceIcon({
        ...nonInferenceWorkload,
        images: ["ubuntu"],
      });
      expect(result).toBeUndefined();
    });

    it("should return undefined for inference workload with nimWorkload flag set to 'false'", () => {
      const workload = {
        ...nonNimInferenceWorkload,
        additionalFields: {
          nimWorkload: "false",
        },
      };
      const result = workloadUtil._getNimInferenceIcon(workload);
      expect(result).toBeUndefined();
    });

    it("should return undefined for inference workload with empty images array", () => {
      const workload = {
        ...nonNimInferenceWorkload,
        images: [],
      };
      const result = workloadUtil._getNimInferenceIcon(workload);
      expect(result).toBeUndefined();
    });

    it("should return undefined for inference workload with undefined images", () => {
      const workload = {
        ...nonNimInferenceWorkload,
        images: undefined,
      };
      const result = workloadUtil._getNimInferenceIcon(workload);
      expect(result).toBeUndefined();
    });
  });
});
