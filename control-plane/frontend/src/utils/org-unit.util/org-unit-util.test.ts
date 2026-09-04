/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest";
import { orgUnitUtil } from "@/utils/org-unit.util/org-unit.util";
import { EGpuResourceState } from "@/models/org-unit-forms.model";
import {
  defaultResources,
  resourcesOverQuotaWeight0,
  resourcesOverQuotaWeight2,
} from "@/utils/org-unit.util/org-unit-util.mock";
import {
  ProjectFilterSortFields,
  ProjectPhase,
  Units,
  type ProjectStatus,
  type Resources,
  type ResourcesNullableGpu,
  type SchedulingRules,
} from "@/swagger-models/org-unit-service-client";
import {
  EOverQuotaPriority,
  OVER_QUOTA_ENABLED_VALUE,
  RESOURCE_MAX_ALLOWED_INFINITE_VALUE_NUMERIC,
} from "@/models/resource.model";
import { EProjectConditionNotReadyReason, type IProjectTable, type IProjectTableFilterBy } from "@/models/project.model";
import DOCS from "@/constants/docs";
import { AVG_RESOURCE_TIMEFRAME_FILTER_KEY } from "@/models/org-unit.model";

describe("org-unit-util", () => {
  describe("initResourcesOverQuotaWeight", () => {
    it("should init resources over quota weight when isOverQuotaPriorityEnabled disabled", () => {
      const updatedResources = orgUnitUtil.initResourcesOverQuotaWeight(defaultResources as Resources[], false);
      updatedResources.forEach((resource) => {
        expect(resource.gpu.overQuotaWeight).toBe(OVER_QUOTA_ENABLED_VALUE);
      });
    });
    it("should init resources over quota weight when isOverQuotaPriorityEnabled enabled", () => {
      const updatedResources = orgUnitUtil.initResourcesOverQuotaWeight(defaultResources as Resources[], true);
      updatedResources.forEach((resource) => {
        expect(resource.gpu.overQuotaWeight).toBe(EOverQuotaPriority.Medium);
      });
    });
    it("should set overQuotaWeight to None when limit = deserved", () => {
      const updatedResources = orgUnitUtil.initResourcesOverQuotaWeight(
        [
          {
            gpu: {
              deserved: 5,
              limit: 5,
            },
          },
        ],
        true,
      );
      updatedResources.forEach((resource) => {
        expect(resource.gpu.overQuotaWeight).toBe(EOverQuotaPriority.None);
      });
    });
    it("should set limit to -1 when overQuotaWeight is 0", () => {
      const updatedResources = orgUnitUtil.initResourcesOverQuotaWeight(
        [
          {
            gpu: {
              deserved: 5,
              limit: -1,
              overQuotaWeight: EOverQuotaPriority.Lowest,
            },
          },
        ],
        true,
      );
      updatedResources.forEach((resource) => {
        expect(resource.gpu.limit).toBe(RESOURCE_MAX_ALLOWED_INFINITE_VALUE_NUMERIC);
      });
    });
  });
  describe("updateLimitOrOverQuota", () => {
    describe("isOverQuotaPriorityDisabled", () => {
      it("should update limit or over quota when isOverQuotaPriorityEnabled disabled and over quota weight is 2", () => {
        const updatedResources = orgUnitUtil.updateLimitOrOverQuota(resourcesOverQuotaWeight2 as Resources[], false);
        updatedResources.forEach((resource) => {
          expect(resource.gpu.limit).toBe(RESOURCE_MAX_ALLOWED_INFINITE_VALUE_NUMERIC);
        });
      });
      it("should update limit or over quota when isOverQuotaPriorityEnabled disabled and over quota weight is 0", () => {
        const updatedResources = orgUnitUtil.updateLimitOrOverQuota(resourcesOverQuotaWeight0 as Resources[], false);
        updatedResources.forEach((resource) => {
          expect(resource.gpu.limit).toBe(resource.gpu.deserved);
        });
      });
    });
    describe("isOverQuotaPriorityEnabled", () => {
      it("should update limit or over quota when isOverQuotaPriorityEnabled enabled and over quota weight is 2", () => {
        const updatedResources = orgUnitUtil.updateLimitOrOverQuota(resourcesOverQuotaWeight2 as Resources[], true);
        updatedResources.forEach((resource) => {
          expect(resource.gpu.limit).toBe(RESOURCE_MAX_ALLOWED_INFINITE_VALUE_NUMERIC);
        });
      });
      it("should not mutate resources when oqw is -1", () => {
        const updatedResources = orgUnitUtil.updateLimitOrOverQuota(
          [
            {
              gpu: {
                deserved: 5,
                limit: -1,
                overQuotaWeight: -1,
              },
            },
          ],
          true,
        );
        updatedResources.forEach((resource) => {
          expect(resource.gpu.overQuotaWeight).toBe(RESOURCE_MAX_ALLOWED_INFINITE_VALUE_NUMERIC);
        });
      });
      it("should not mutate resources when oqw is 0 and limit is not -1", () => {
        const updatedResources = orgUnitUtil.updateLimitOrOverQuota(
          [
            {
              gpu: {
                deserved: 5,
                limit: 5,
                overQuotaWeight: 0,
              },
            },
          ],
          true,
        );
        updatedResources.forEach((resource) => {
          expect(resource.gpu.overQuotaWeight).toBe(0);
        });
      });
      it("should set limit = deserved when over quota weight set to None", () => {
        const updatedResources = orgUnitUtil.updateLimitOrOverQuota(
          [
            {
              gpu: {
                deserved: 5,
                overQuotaWeight: EOverQuotaPriority.None,
              },
            },
          ],
          true,
        );
        updatedResources.forEach((resource) => {
          expect(resource.gpu.limit).toBe(5);
        });
      });
      it("should set overQuotaWeight = null amd limit to -1 when over quota weight set to EOverQuotaPriority.Lowest", () => {
        const updatedResources = orgUnitUtil.updateLimitOrOverQuota(
          [
            {
              gpu: {
                deserved: 5,
                overQuotaWeight: EOverQuotaPriority.Lowest,
              },
            },
          ],
          true,
        );
        updatedResources.forEach((resource) => {
          expect(resource.gpu.limit).toBe(null);
          expect(resource.gpu.overQuotaWeight).toBe(0);
        });
      });
    });
  });
  describe("SchedulingRules", () => {
    describe("orgUnitUtil.updateSchedulingRules", () => {
      it("should set project value to null if it matches the parent value", () => {
        const projectSR: SchedulingRules = {
          trainingJobTimeLimitSeconds: 3000,
          interactiveJobTimeLimitSeconds: 7200,
        };
        const parentSR: SchedulingRules = {
          trainingJobTimeLimitSeconds: 3000,
          interactiveJobTimeLimitSeconds: 7200,
        };
        const updatedSR = orgUnitUtil.updateSchedulingRules(projectSR, parentSR);
        expect(updatedSR.trainingJobTimeLimitSeconds).toBeNull();
        expect(updatedSR.interactiveJobTimeLimitSeconds).toBeNull();
      });

      it("should set project value to parent value if project value is greater than parent value", () => {
        const projectSR: SchedulingRules = {
          trainingJobTimeLimitSeconds: 4000,
          interactiveJobTimeLimitSeconds: 8000,
        };
        const parentSR: SchedulingRules = {
          trainingJobTimeLimitSeconds: 3000,
          interactiveJobTimeLimitSeconds: 7200,
        };
        const updatedSR = orgUnitUtil.updateSchedulingRules(projectSR, parentSR);
        expect(updatedSR.trainingJobTimeLimitSeconds).toBe(3000);
        expect(updatedSR.interactiveJobTimeLimitSeconds).toBe(7200);
      });

      it("should not change project value if parent value is null or undefined", () => {
        const projectSR: SchedulingRules = {
          trainingJobTimeLimitSeconds: 4000,
          interactiveJobTimeLimitSeconds: 8000,
        };
        const parentSR: SchedulingRules = {
          trainingJobTimeLimitSeconds: null,
          interactiveJobTimeLimitSeconds: undefined,
        };
        const updatedSR = orgUnitUtil.updateSchedulingRules(projectSR, parentSR);
        expect(updatedSR.trainingJobTimeLimitSeconds).toBe(4000);
        expect(updatedSR.interactiveJobTimeLimitSeconds).toBe(8000);
      });

      it("should not change project value if project value is null or undefined", () => {
        const projectSR: SchedulingRules = {
          trainingJobTimeLimitSeconds: null,
          interactiveJobTimeLimitSeconds: undefined,
        };
        const parentSR: SchedulingRules = {
          trainingJobTimeLimitSeconds: 3000,
          interactiveJobTimeLimitSeconds: 7200,
        };
        const updatedSR = orgUnitUtil.updateSchedulingRules(projectSR, parentSR);
        expect(updatedSR.trainingJobTimeLimitSeconds).toBeNull();
        expect(updatedSR.interactiveJobTimeLimitSeconds).toBeUndefined();
      });

      it("should handle empty scheduling rules", () => {
        const projectSR: SchedulingRules = {};
        const parentSR: SchedulingRules = {};
        const updatedSR = orgUnitUtil.updateSchedulingRules(projectSR, parentSR);
        expect(updatedSR).toEqual({});
      });

      it("should not mutate the original object", () => {
        const projectSR: SchedulingRules = {
          trainingJobTimeLimitSeconds: 4000,
          interactiveJobTimeLimitSeconds: 8000,
        };
        const parentSR: SchedulingRules = {
          trainingJobTimeLimitSeconds: 3000,
          interactiveJobTimeLimitSeconds: 7200,
        };
        const updatedSR = orgUnitUtil.updateSchedulingRules(projectSR, parentSR);
        expect(updatedSR).not.toBe(projectSR);
        expect(updatedSR).not.toBe(parentSR);
      });
    });
    describe("orgUnitUtil.initSchedulingRules", () => {
      it("should initialize scheduling rules with parent values if not defined in project", () => {
        const projectSR: SchedulingRules = {
          interactiveJobTimeLimitSeconds: 3600,
        };
        const parentSR: SchedulingRules = {
          interactiveJobTimeLimitSeconds: 7200,
          trainingJobMaxIdleDurationSeconds: 1800,
        };
        const updatedSR = orgUnitUtil.initSchedulingRules(projectSR, parentSR);
        expect(updatedSR.interactiveJobTimeLimitSeconds).toBe(3600);
        expect(updatedSR.trainingJobMaxIdleDurationSeconds).toBe(1800);
      });

      it("should not override project values with parent values", () => {
        const projectSR: SchedulingRules = {
          interactiveJobTimeLimitSeconds: 3600,
          trainingJobMaxIdleDurationSeconds: 3600,
        };
        const parentSR: SchedulingRules = {
          interactiveJobTimeLimitSeconds: 7200,
          trainingJobMaxIdleDurationSeconds: 1800,
        };
        const updatedSR = orgUnitUtil.initSchedulingRules(projectSR, parentSR);
        expect(updatedSR.interactiveJobTimeLimitSeconds).toBe(3600);
        expect(updatedSR.trainingJobMaxIdleDurationSeconds).toBe(3600);
      });

      it("should handle empty project scheduling rules", () => {
        const projectSR: SchedulingRules = {};
        const parentSR: SchedulingRules = {
          interactiveJobTimeLimitSeconds: 7200,
          trainingJobMaxIdleDurationSeconds: 1800,
        };
        const updatedSR = orgUnitUtil.initSchedulingRules(projectSR, parentSR);
        expect(updatedSR.interactiveJobTimeLimitSeconds).toBe(7200);
        expect(updatedSR.trainingJobMaxIdleDurationSeconds).toBe(1800);
      });

      it("should handle empty parent scheduling rules", () => {
        const projectSR: SchedulingRules = {
          interactiveJobTimeLimitSeconds: 3600,
          trainingJobMaxIdleDurationSeconds: 3600,
        };
        const parentSR: SchedulingRules = {};
        const updatedSR = orgUnitUtil.initSchedulingRules(projectSR, parentSR);
        expect(updatedSR.interactiveJobTimeLimitSeconds).toBe(3600);
        expect(updatedSR.trainingJobMaxIdleDurationSeconds).toBe(3600);
      });

      it("should handle both empty project and parent scheduling rules", () => {
        const projectSR: SchedulingRules = {};
        const parentSR: SchedulingRules = {};
        const updatedSR = orgUnitUtil.initSchedulingRules(projectSR, parentSR);
        expect(updatedSR).toEqual({});
      });
    });
  });

  describe("getProjectListFilters", () => {
    it("should return filters with correct allocation and utilization keys for 24 hours", () => {
      const filterBy = {
        sortBy: "averageGpuAllocation",
        avgGpuUtilizationTimeframe: "7d",
        avgGpuAllocationTimeframe: "7d",
        [AVG_RESOURCE_TIMEFRAME_FILTER_KEY]: "7d",
        advancedFilters: [
          {
            minValue: 2,
            maxValue: 0,
            operator: ">=",
            label: "Avg. GPU allocation",
            name: "averageGpuAllocation",
            type: "Numeric",
          },
        ],
      };
      const filters = orgUnitUtil.getProjectListFilters(filterBy as IProjectTableFilterBy);
      expect(filters.sortBy).toBe(ProjectFilterSortFields.AvgGpuAllocation7d);
      expect(filters.filterBy).toStrictEqual(["avgGpuAllocation7d>=2"]);
    });

    it("should return filters with correct allocation and utilization keys for 7 days", () => {
      const filterBy = {
        sortBy: "name",
        avgGpuUtilizationTimeframe: "30d",
        avgGpuAllocationTimeframe: "24h",
        [AVG_RESOURCE_TIMEFRAME_FILTER_KEY]: "7d",
        advancedFilters: [
          {
            minValue: 2,
            maxValue: 0,
            operator: "==",
            label: "Avg. GPU allocation",
            name: "averageGpuAllocation",
            type: "Numeric",
          },
          {
            minValue: 5,
            maxValue: 0,
            operator: "==",
            label: "Avg. GPU utilization",
            name: "averageGpuUtilization",
            type: "Numeric",
          },
        ],
      };
      // TODO: fix this
      const filters = orgUnitUtil.getProjectListFilters(filterBy as IProjectTableFilterBy);
      expect(filters.sortBy).toBe(ProjectFilterSortFields.Name);
      expect(filters.filterBy).toStrictEqual(["avgGpuAllocation24h==2", "avgGpuUtilization30d==5"]);
    });
    it("should return empty filterBy for no advanced filters", () => {
      const filterBy = {
        sortBy: "name",
      };
      const filters = orgUnitUtil.getProjectListFilters(filterBy as IProjectTableFilterBy);
      expect(filters.sortBy).toBe(ProjectFilterSortFields.Name);
      expect(filters.filterBy).toStrictEqual([]);
    });
  });
  describe("getAllocationRatioDisplayValue", () => {
    it("should return percentage when gpuQuota is positive", () => {
      const result = orgUnitUtil.getAllocationRatioDisplayValue(50, { gpuQuota: 100 });
      expect(result).toBe("50.00%");
    });

    it("should return '-' when gpuQuota is negative", () => {
      const result = orgUnitUtil.getAllocationRatioDisplayValue(50, { gpuQuota: -1 });
      expect(result).toBe("-");
    });

    it("should return '0%' when allocatedGpus is 0", () => {
      const result = orgUnitUtil.getAllocationRatioDisplayValue(0, { gpuQuota: 100 });
      expect(result).toBe("0.00%");
    });

    it("should return '-' when totalResources is undefined", () => {
      const result = orgUnitUtil.getAllocationRatioDisplayValue(50);
      expect(result).toBe("-");
    });

    it("should return '-' when gpuQuota is 0", () => {
      const result = orgUnitUtil.getAllocationRatioDisplayValue(50, { gpuQuota: 0 });
      expect(result).toBe("-");
    });
  });
  describe("getStatusColOptions", () => {
    it("should return default options when project is undefined", () => {
      const result = orgUnitUtil.getStatusColOptions();
      expect(result).toEqual({
        status: "-",
        tooltipText: "",
        displayAnimation: false,
      });
    });

    it("should return default options when project status is undefined", () => {
      const result = orgUnitUtil.getStatusColOptions({} as IProjectTable);
      expect(result).toEqual({
        status: "-",
        tooltipText: "",
        displayAnimation: false,
      });
    });

    it("should return options with tooltip text when project is in Deleting phase", () => {
      const project = {
        status: {
          phase: ProjectPhase.Deleting,
          additionalStatusData: {
            conditions: [
              {
                message:
                  "The project couldn't be deleted because the following needs to be deleted first:\nTrainingWorkload job-2025-01-05t08-00-00\n",
                reason: "WorkloadsDeletionHandlerFailed",
                status: "False",
                type: "WorkloadsReady",
              },
            ],
          },
        } as ProjectStatus,
      } as IProjectTable;
      const result = orgUnitUtil.getStatusColOptions(project);
      expect(result.tooltipText).toBe(
        "\n" +
          " The project couldn't be deleted because the following needs to be deleted first:<br><br/>TrainingWorkload job-2025-01-05t08-00-00<br/><br>\n" +
          ' Either delete the above or try <a href="javascript:void(0)" style="text-decoration: underline;" id="status-col-emit-project-deletion">deleting the project anyway</a>',
      );
    });

    it("should return options without tooltip text when project is not in Deleting phase", () => {
      const project = {
        status: {
          phase: ProjectPhase.Ready,
        },
      } as IProjectTable;
      const result = orgUnitUtil.getStatusColOptions(project);
      expect(result.tooltipText).toBe("");
    });
  });

  describe("getNodePoolGpuState", () => {
    it("should return NoGPUResourcesAreSet when gpu is null", () => {
      expect(orgUnitUtil.getNodePoolGpuState(null)).toBe(EGpuResourceState.NoGPUResourcesAreSet);
    });

    it("should return NoGPUResourcesAreSet when gpu is undefined", () => {
      expect(orgUnitUtil.getNodePoolGpuState(undefined)).toBe(EGpuResourceState.NoGPUResourcesAreSet);
    });

    it("should return NoGPUResourcesAreSet when limit is 0", () => {
      const gpu: ResourcesNullableGpu = { deserved: 0, limit: 0 };
      expect(orgUnitUtil.getNodePoolGpuState(gpu)).toBe(EGpuResourceState.NoGPUResourcesAreSet);
    });

    it("should return NoGPUResourcesAreSet when deserved > 0 but limit is 0", () => {
      const gpu: ResourcesNullableGpu = { deserved: 5, limit: 0 };
      expect(orgUnitUtil.getNodePoolGpuState(gpu)).toBe(EGpuResourceState.NoGPUResourcesAreSet);
    });

    it("should return OpportunisticOnly when deserved is 0 and limit is unlimited (null)", () => {
      const gpu: ResourcesNullableGpu = { deserved: 0, limit: null };
      expect(orgUnitUtil.getNodePoolGpuState(gpu)).toBe(EGpuResourceState.OpportunisticOnly);
    });

    it("should return OpportunisticOnly when deserved is 0 and limit is unlimited (-1)", () => {
      const gpu: ResourcesNullableGpu = { deserved: 0, limit: -1 };
      expect(orgUnitUtil.getNodePoolGpuState(gpu)).toBe(EGpuResourceState.OpportunisticOnly);
    });

    it("should return OpportunisticOnly when deserved is 0 and limit is positive", () => {
      const gpu: ResourcesNullableGpu = { deserved: 0, limit: 10 };
      expect(orgUnitUtil.getNodePoolGpuState(gpu)).toBe(EGpuResourceState.OpportunisticOnly);
    });

    it("should return DedicatedGPUs when deserved equals limit", () => {
      const gpu: ResourcesNullableGpu = { deserved: 5, limit: 5 };
      expect(orgUnitUtil.getNodePoolGpuState(gpu)).toBe(EGpuResourceState.DedicatedGPUs);
    });

    it("should return DedicatedAndOpportunisticGPUs when deserved > 0 and limit is unlimited (null)", () => {
      const gpu: ResourcesNullableGpu = { deserved: 5, limit: null };
      expect(orgUnitUtil.getNodePoolGpuState(gpu)).toBe(EGpuResourceState.DedicatedAndOpportunisticGPUs);
    });

    it("should return DedicatedAndOpportunisticGPUs when deserved > 0 and limit is unlimited (-1)", () => {
      const gpu: ResourcesNullableGpu = { deserved: 5, limit: -1 };
      expect(orgUnitUtil.getNodePoolGpuState(gpu)).toBe(EGpuResourceState.DedicatedAndOpportunisticGPUs);
    });

    it("should return DedicatedAndOpportunisticGPUs when deserved > 0 and limit > deserved", () => {
      const gpu: ResourcesNullableGpu = { deserved: 5, limit: 10 };
      expect(orgUnitUtil.getNodePoolGpuState(gpu)).toBe(EGpuResourceState.DedicatedAndOpportunisticGPUs);
    });

    it("should handle undefined deserved as 0", () => {
      const gpu: ResourcesNullableGpu = { limit: 10 };
      expect(orgUnitUtil.getNodePoolGpuState(gpu)).toBe(EGpuResourceState.OpportunisticOnly);
    });

    it("should return DedicatedAndOpportunisticGPUs when deserved is unlimited (-1) and limit is unlimited (-1)", () => {
      const gpu: ResourcesNullableGpu = { deserved: -1, limit: -1 };
      expect(orgUnitUtil.getNodePoolGpuState(gpu)).toBe(EGpuResourceState.DedicatedAndOpportunisticGPUs);
    });

    it("should return DedicatedAndOpportunisticGPUs when deserved is unlimited (-1) and limit is null", () => {
      const gpu: ResourcesNullableGpu = { deserved: -1, limit: null };
      expect(orgUnitUtil.getNodePoolGpuState(gpu)).toBe(EGpuResourceState.DedicatedAndOpportunisticGPUs);
    });

    it("should return DedicatedAndOpportunisticGPUs when deserved is unlimited (-1) and limit is positive", () => {
      const gpu: ResourcesNullableGpu = { deserved: -1, limit: 10 };
      expect(orgUnitUtil.getNodePoolGpuState(gpu)).toBe(EGpuResourceState.DedicatedAndOpportunisticGPUs);
    });
  });

  describe("aggregateGpuStates", () => {
    it("should return NoGPUResourcesAreSet for empty array", () => {
      expect(orgUnitUtil.aggregateGpuStates([])).toBe(EGpuResourceState.NoGPUResourcesAreSet);
    });

    it("should return the single state when array has one element", () => {
      expect(orgUnitUtil.aggregateGpuStates([EGpuResourceState.DedicatedGPUs])).toBe(EGpuResourceState.DedicatedGPUs);
    });

    it("should return the highest priority state (DedicatedAndOpportunistic > Dedicated)", () => {
      const states = [EGpuResourceState.DedicatedGPUs, EGpuResourceState.DedicatedAndOpportunisticGPUs];
      expect(orgUnitUtil.aggregateGpuStates(states)).toBe(EGpuResourceState.DedicatedAndOpportunisticGPUs);
    });

    it("should return DedicatedAndOpportunisticGPUs when mix of Opportunistic and Dedicated", () => {
      const states = [EGpuResourceState.OpportunisticOnly, EGpuResourceState.DedicatedGPUs];
      expect(orgUnitUtil.aggregateGpuStates(states)).toBe(EGpuResourceState.DedicatedAndOpportunisticGPUs);
    });

    it("should return DedicatedAndOpportunisticGPUs when mix includes NoGPUResourcesAreSet", () => {
      const states = [
        EGpuResourceState.NoGPUResourcesAreSet,
        EGpuResourceState.OpportunisticOnly,
        EGpuResourceState.DedicatedGPUs,
      ];
      expect(orgUnitUtil.aggregateGpuStates(states)).toBe(EGpuResourceState.DedicatedAndOpportunisticGPUs);
    });

    it("should return OpportunisticOnly when all are OpportunisticOnly", () => {
      const states = [EGpuResourceState.OpportunisticOnly, EGpuResourceState.OpportunisticOnly];
      expect(orgUnitUtil.aggregateGpuStates(states)).toBe(EGpuResourceState.OpportunisticOnly);
    });

    it("should return DedicatedGPUs when all are DedicatedGPUs", () => {
      const states = [EGpuResourceState.DedicatedGPUs, EGpuResourceState.DedicatedGPUs];
      expect(orgUnitUtil.aggregateGpuStates(states)).toBe(EGpuResourceState.DedicatedGPUs);
    });

    it("should return NoGPUResourcesAreSet when all are NoGPUResourcesAreSet", () => {
      const states = [EGpuResourceState.NoGPUResourcesAreSet, EGpuResourceState.NoGPUResourcesAreSet];
      expect(orgUnitUtil.aggregateGpuStates(states)).toBe(EGpuResourceState.NoGPUResourcesAreSet);
    });

    it("should prefer DedicatedAndOpportunisticGPUs over edge case logic", () => {
      const states = [
        EGpuResourceState.DedicatedAndOpportunisticGPUs,
        EGpuResourceState.OpportunisticOnly,
        EGpuResourceState.DedicatedGPUs,
      ];
      expect(orgUnitUtil.aggregateGpuStates(states)).toBe(EGpuResourceState.DedicatedAndOpportunisticGPUs);
    });
  });

  describe("alignCpuAndMemoryLimitsByGpu", () => {
    it("should set CPU and Memory limits to null when GPU is unlimited (null)", () => {
      const resources: Resources[] = [
        {
          nodePool: { id: "1", name: "default" },
          gpu: { deserved: 10, limit: null },
          cpu: { deserved: 1000, limit: 1000 },
          memory: { deserved: 2048, limit: 2048, units: Units.Mb },
        },
      ];
      const result = orgUnitUtil.alignCpuAndMemoryLimitsByGpu(resources);
      expect(result[0].cpu?.limit).toBeNull();
      expect(result[0].memory?.limit).toBeNull();
    });

    it("should set CPU and Memory limits to null when GPU is unlimited (-1)", () => {
      const resources: Resources[] = [
        {
          nodePool: { id: "1", name: "default" },
          gpu: { deserved: 10, limit: -1 },
          cpu: { deserved: 1000, limit: 1000 },
          memory: { deserved: 2048, limit: 2048, units: Units.Mb },
        },
      ];
      const result = orgUnitUtil.alignCpuAndMemoryLimitsByGpu(resources);
      expect(result[0].cpu?.limit).toBeNull();
      expect(result[0].memory?.limit).toBeNull();
    });

    it("should set CPU and Memory limits to null when GPU limit > deserved (over-quota enabled with specific limit)", () => {
      const resources: Resources[] = [
        {
          nodePool: { id: "1", name: "default" },
          gpu: { deserved: 10, limit: 15 },
          cpu: { deserved: 1000, limit: 1000 },
          memory: { deserved: 2048, limit: 2048, units: Units.Mb },
        },
      ];
      const result = orgUnitUtil.alignCpuAndMemoryLimitsByGpu(resources);
      expect(result[0].cpu?.limit).toBeNull();
      expect(result[0].memory?.limit).toBeNull();
    });

    it("should set CPU and Memory limits to deserved when GPU limit equals deserved (over-quota disabled)", () => {
      const resources: Resources[] = [
        {
          nodePool: { id: "1", name: "default" },
          gpu: { deserved: 10, limit: 10 },
          cpu: { deserved: 1000, limit: null },
          memory: { deserved: 2048, limit: null, units: Units.Mb },
        },
      ];
      const result = orgUnitUtil.alignCpuAndMemoryLimitsByGpu(resources);
      expect(result[0].cpu?.limit).toBe(1000);
      expect(result[0].memory?.limit).toBe(2048);
    });

    it("should set CPU and Memory limits to deserved when GPU limit < deserved", () => {
      const resources: Resources[] = [
        {
          nodePool: { id: "1", name: "default" },
          gpu: { deserved: 10, limit: 5 },
          cpu: { deserved: 1000, limit: null },
          memory: { deserved: 2048, limit: null, units: Units.Mb },
        },
      ];
      const result = orgUnitUtil.alignCpuAndMemoryLimitsByGpu(resources);
      expect(result[0].cpu?.limit).toBe(1000);
      expect(result[0].memory?.limit).toBe(2048);
    });

    it("should handle resources without CPU or Memory", () => {
      const resources: Resources[] = [
        {
          nodePool: { id: "1", name: "default" },
          gpu: { deserved: 10, limit: null },
        },
      ];
      const result = orgUnitUtil.alignCpuAndMemoryLimitsByGpu(resources);
      expect(result[0].cpu).toBeUndefined();
      expect(result[0].memory).toBeUndefined();
    });

    it("should handle multiple resources with different GPU states", () => {
      const resources: Resources[] = [
        {
          nodePool: { id: "1", name: "pool1" },
          gpu: { deserved: 10, limit: null }, // unlimited
          cpu: { deserved: 1000, limit: 1000 },
          memory: { deserved: 2048, limit: 2048, units: Units.Mb },
        },
        {
          nodePool: { id: "2", name: "pool2" },
          gpu: { deserved: 10, limit: 15 }, // over-quota with specific limit
          cpu: { deserved: 2000, limit: 2000 },
          memory: { deserved: 4096, limit: 4096, units: Units.Mb },
        },
        {
          nodePool: { id: "3", name: "pool3" },
          gpu: { deserved: 10, limit: 10 }, // over-quota disabled
          cpu: { deserved: 500, limit: null },
          memory: { deserved: 1024, limit: null, units: Units.Mb },
        },
      ];
      const result = orgUnitUtil.alignCpuAndMemoryLimitsByGpu(resources);

      // Pool 1: GPU unlimited -> CPU/Memory unlimited
      expect(result[0].cpu?.limit).toBeNull();
      expect(result[0].memory?.limit).toBeNull();

      // Pool 2: GPU limit > deserved -> CPU/Memory unlimited
      expect(result[1].cpu?.limit).toBeNull();
      expect(result[1].memory?.limit).toBeNull();

      // Pool 3: GPU limit = deserved -> CPU/Memory = deserved
      expect(result[2].cpu?.limit).toBe(500);
      expect(result[2].memory?.limit).toBe(1024);
    });

    it("should not mutate original resources array (immutability)", () => {
      const resources: Resources[] = [
        {
          nodePool: { id: "1", name: "default" },
          gpu: { deserved: 10, limit: null },
          cpu: { deserved: 1000, limit: 1000 },
          memory: { deserved: 2048, limit: 2048, units: Units.Mb },
        },
      ];
      const result = orgUnitUtil.alignCpuAndMemoryLimitsByGpu(resources);
      expect(resources[0].cpu?.limit).toBe(1000); // Original unchanged
      expect(result[0].cpu?.limit).toBeNull(); // Result modified
      expect(result).not.toBe(resources); // Different array
    });
  });

  describe("ensureLimitsNotBelowDeserved", () => {
    it("should set GPU limit to deserved when limit is less than deserved", () => {
      const resources: Resources[] = [
        {
          nodePool: { id: "1", name: "default" },
          gpu: { deserved: 68, limit: 61 },
        },
      ];
      const result = orgUnitUtil.ensureLimitsNotBelowDeserved(resources);
      expect(result[0].gpu.limit).toBe(68);
      expect(result[0].gpu.deserved).toBe(68);
    });

    it("should not modify GPU limit when limit is greater than deserved", () => {
      const resources: Resources[] = [
        {
          nodePool: { id: "1", name: "default" },
          gpu: { deserved: 50, limit: 100 },
        },
      ];
      const result = orgUnitUtil.ensureLimitsNotBelowDeserved(resources);
      expect(result[0].gpu.limit).toBe(100);
      expect(result[0].gpu.deserved).toBe(50);
    });

    it("should not modify GPU limit when limit equals deserved", () => {
      const resources: Resources[] = [
        {
          nodePool: { id: "1", name: "default" },
          gpu: { deserved: 50, limit: 50 },
        },
      ];
      const result = orgUnitUtil.ensureLimitsNotBelowDeserved(resources);
      expect(result[0].gpu.limit).toBe(50);
    });

    it("should not modify GPU limit when limit is null (unlimited)", () => {
      const resources: Resources[] = [
        {
          nodePool: { id: "1", name: "default" },
          gpu: { deserved: 68, limit: null },
        },
      ];
      const result = orgUnitUtil.ensureLimitsNotBelowDeserved(resources);
      expect(result[0].gpu.limit).toBeNull();
    });

    it("should not modify GPU limit when limit is undefined (unlimited)", () => {
      const resources: Resources[] = [
        {
          nodePool: { id: "1", name: "default" },
          gpu: { deserved: 68, limit: undefined },
        },
      ];
      const result = orgUnitUtil.ensureLimitsNotBelowDeserved(resources);
      expect(result[0].gpu.limit).toBeUndefined();
    });

    it("should not modify GPU limit when limit is -1 (unlimited)", () => {
      const resources: Resources[] = [
        {
          nodePool: { id: "1", name: "default" },
          gpu: { deserved: 68, limit: -1 },
        },
      ];
      const result = orgUnitUtil.ensureLimitsNotBelowDeserved(resources);
      expect(result[0].gpu.limit).toBe(-1);
    });

    it("should set CPU limit to deserved when limit is less than deserved", () => {
      const resources: Resources[] = [
        {
          nodePool: { id: "1", name: "default" },
          gpu: { deserved: 10, limit: 20 },
          cpu: { deserved: 1000, limit: 500 },
        },
      ];
      const result = orgUnitUtil.ensureLimitsNotBelowDeserved(resources);
      expect(result[0].cpu?.limit).toBe(1000);
    });

    it("should set Memory limit to deserved when limit is less than deserved", () => {
      const resources: Resources[] = [
        {
          nodePool: { id: "1", name: "default" },
          gpu: { deserved: 10, limit: 20 },
          memory: { deserved: 2048, limit: 1024, units: Units.Mb },
        },
      ];
      const result = orgUnitUtil.ensureLimitsNotBelowDeserved(resources);
      expect(result[0].memory?.limit).toBe(2048);
    });

    it("should handle multiple resources with mixed limit/deserved values", () => {
      const resources: Resources[] = [
        {
          nodePool: { id: "1", name: "default" },
          gpu: { deserved: 68, limit: 61 }, // limit < deserved
        },
        {
          nodePool: { id: "2", name: "gpu-pool" },
          gpu: { deserved: 20, limit: 50 }, // limit > deserved
        },
        {
          nodePool: { id: "3", name: "cpu-pool" },
          gpu: { deserved: 30, limit: null }, // unlimited
        },
      ];
      const result = orgUnitUtil.ensureLimitsNotBelowDeserved(resources);
      expect(result[0].gpu.limit).toBe(68); // Fixed: now equals deserved
      expect(result[1].gpu.limit).toBe(50); // Unchanged
      expect(result[2].gpu.limit).toBeNull(); // Unchanged (unlimited)
    });

    it("should handle resources with all fields (GPU, CPU, Memory)", () => {
      const resources: Resources[] = [
        {
          nodePool: { id: "1", name: "default" },
          gpu: { deserved: 10, limit: 5 }, // limit < deserved
          cpu: { deserved: 2000, limit: 1000 }, // limit < deserved
          memory: { deserved: 4096, limit: 2048, units: Units.Mb }, // limit < deserved
        },
      ];
      const result = orgUnitUtil.ensureLimitsNotBelowDeserved(resources);
      expect(result[0].gpu.limit).toBe(10);
      expect(result[0].cpu?.limit).toBe(2000);
      expect(result[0].memory?.limit).toBe(4096);
    });

    it("should not modify original resources array (immutability)", () => {
      const resources: Resources[] = [
        {
          nodePool: { id: "1", name: "default" },
          gpu: { deserved: 68, limit: 61 },
        },
      ];
      const result = orgUnitUtil.ensureLimitsNotBelowDeserved(resources);
      expect(resources[0].gpu.limit).toBe(61); // Original unchanged
      expect(result[0].gpu.limit).toBe(68); // Result modified
      expect(result).not.toBe(resources); // Different array
    });
  });

  describe("getStatusColOptions - NotReady tooltip", () => {
    function makeNotReadyProject(phaseMessage: string): IProjectTable {
      return {
        status: { phase: ProjectPhase.NotReady, phaseMessage },
      } as IProjectTable;
    }

    it("returns empty tooltip when phaseMessage is empty", () => {
      const result = orgUnitUtil.getStatusColOptions(makeNotReadyProject(""));
      expect(result.tooltipText).toBe("");
    });

    it("returns empty tooltip for an unknown reason", () => {
      const result = orgUnitUtil.getStatusColOptions(makeNotReadyProject("SomeUnknownReason"));
      expect(result.tooltipText).toBe("");
    });

    it("returns correct tooltip for a single namespace reason", () => {
      const result = orgUnitUtil.getStatusColOptions(makeNotReadyProject("NamespaceNotFound"));
      expect(result.tooltipText).toContain("A namespace couldn't be found for the project");
      expect(result.tooltipText).not.toContain("contact");
    });

    it("returns correct tooltip for a single handler reason", () => {
      const result = orgUnitUtil.getStatusColOptions(makeNotReadyProject("SecretsHandlerFailed"));
      expect(result.tooltipText).toContain("The project couldn't be created (SecretsHandlerFailed).");
      expect(result.tooltipText).not.toContain(DOCS.SUPPORT.LABEL);
    });

    it("shows only RoleBindingsHandlerFailed tooltip when it is the sole reason", () => {
      const result = orgUnitUtil.getStatusColOptions(
        makeNotReadyProject(EProjectConditionNotReadyReason.RoleBindingsHandlerFailed),
      );
      expect(result.tooltipText).toContain("RoleBindingsHandlerFailed");
    });

    it("suppresses RoleBindingsHandlerFailed when combined with another reason", () => {
      const result = orgUnitUtil.getStatusColOptions(
        makeNotReadyProject("NamespaceNotFound, RoleBindingsHandlerFailed"),
      );
      expect(result.tooltipText).toContain("A namespace couldn't be found");
      expect(result.tooltipText).not.toContain("RoleBindingsHandlerFailed");
    });

    it("aggregates multiple namespace reasons without support copy", () => {
      const result = orgUnitUtil.getStatusColOptions(makeNotReadyProject("SecretsHandlerFailed, PvcsHandlerFailed"));
      expect(result.tooltipText).toContain("The project couldn't be created (SecretsHandlerFailed).");
      expect(result.tooltipText).toContain("The project couldn't be created (PvcsHandlerFailed).");
      expect(result.tooltipText).not.toContain(DOCS.SUPPORT.LABEL);
    });

    it("aggregates two handler reasons without a dangling Contact phrase", () => {
      const result = orgUnitUtil.getStatusColOptions(
        makeNotReadyProject("QueuesHandlerFailed, LimitRangeHandlerFailed"),
      );
      expect(result.tooltipText).toContain("The project couldn't be created (QueuesHandlerFailed).");
      expect(result.tooltipText).toContain("The project couldn't be created (LimitRangeHandlerFailed).");
      expect(result.tooltipText).not.toContain("Contact");
      expect(result.tooltipText).not.toContain(DOCS.SUPPORT.LABEL);
    });

    it("retains the clickable support CTA when the support URL is configured", () => {
      const support = DOCS.SUPPORT as { URL: string };
      const originalUrl = support.URL;
      support.URL = "https://support.example.test/contact";

      try {
        const result = orgUnitUtil.getStatusColOptions(makeNotReadyProject("NamespaceNotFound"));
        expect(result.tooltipText).toContain('<a href="https://support.example.test/contact">NVIDIA Run:ai support</a>');
        expect(result.tooltipText).toContain("or contact");
      } finally {
        support.URL = originalUrl;
      }
    });

    it("returns no tooltip for Ready phase", () => {
      const project = { status: { phase: ProjectPhase.Ready, phaseMessage: "" } } as IProjectTable;
      const result = orgUnitUtil.getStatusColOptions(project);
      expect(result.tooltipText).toBeFalsy();
    });
  });
});
