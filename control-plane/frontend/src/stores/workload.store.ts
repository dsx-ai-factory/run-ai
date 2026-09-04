/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Stores
import { defineStore } from "pinia";
import { useClusterStore } from "./cluster.store";

//api
import {
  type BatchResponse,
  SourceApi,
  type HistoryRecord,
  type Workload,
  type WorkloadCategory,
  type DistributedWorkload,
} from "@/swagger-models/workloads-client";
import { Phase, type Training, type Workspace } from "@/swagger-models/workloads-client";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { workloadUtil } from "@/utils/workload.util/workload.util";
import { filterService } from "@/services/filter.service/filter.service";

// models
import type { IFilterBy } from "@/models/filter.model";
import type { IJobFilter } from "@/models/filter.model";
import { eventHistoryColumns } from "@/table-models/workload.table-model";
import type { IWorkspaceList } from "@/models/workspace.model";
import type { ITrainingList } from "@/models/training.model";
import type { IInferenceList } from "@/models/inference.model";
import type { ITableColumn } from "@/models/table.model";
import type { EWorkloadType } from "@/models/workload.model";
import { type IYamlWorkloadFormState } from "@/models/workload.model";

import { workloadLocalPhaseUpdater } from "@/services/cluster/workload.service/local-status-updater";
import type {
  Inference,
  InferenceV2,
  TrainingV2,
  WorkspaceV2,
  ModelInference,
  ModelInferenceV2,
} from "@/swagger-models/assets-service-client";
import type { WorkloadV2 } from "@/swagger-models/workloads-manager-client";

const getDefaultFilters = (): IJobFilter => {
  return {
    displayedColumns: eventHistoryColumns
      .filter((col: ITableColumn) => col.display)
      .map((col: ITableColumn) => col.name),
  };
};

export const useWorkloadStore = defineStore("WorkloadStore", {
  state: () => ({
    workloads: [] as Array<Workload>,
    workloadHistoryRecords: [] as Array<HistoryRecord>,
    workloadHistoryFilterBy: getDefaultFilters() as IJobFilter,
    loading: false as boolean,
    namespaces: {} as Record<string, string | undefined>,
    clusterStore: useClusterStore(),
    workloadAdded: null as null | Workload,
    workloadCategories: [] as WorkloadCategory[],
    // YAML workload form state for preservation during navigation
    yamlFormState: null as IYamlWorkloadFormState | null,
  }),
  getters: {
    workloadList(): Array<Workload> {
      if (!this.workloadAdded) return this.workloads;
      return this.workloads;
    },
    allWorkloads(): Array<Workload> {
      return this.workloads;
    },
    allWorkloadMapById(): Map<string, Workload> {
      const map = new Map<string, Workload>();
      this.workloads.forEach((w) => {
        if (w.id) map.set(w.id, w);
      });
      return map;
    },
    lastCreatedWorkload(): Workload | null {
      return this.workloadAdded;
    },
  },
  actions: {
    async loadWorkloads(filterBy: IFilterBy = {}, deleted = false): Promise<Workload[]> {
      const workloads: Workload[] = await this.loadWorkloadsList(filterBy, deleted);
      await this.updateWorkloadAdded(workloads);
      this.workloads = workloads;
      return this.workloadList;
    },
    async loadWorkloadsList(filterBy: IFilterBy = {}, deleted = false): Promise<Workload[]> {
      return await workloadService.list(this.clusterStore.selectedClusterId, filterBy, deleted);
    },
    async deleteWorkloads(workloadIds: string[]): Promise<BatchResponse> {
      const response: BatchResponse = await workloadService.deleteMultipleWorkloads(workloadIds);
      if (response.succeeded?.length) {
        response.succeeded.forEach((workloadId: string) => {
          const workload: Workload | undefined = this.allWorkloadMapById.get(workloadId);
          if (workload) {
            this.setWorkloadPhase(workload, Phase.Deleting);
          }
        });
      }
      return response;
    },
    async deleteWorkload(clusterId: string, workloadToDelete: Workload): Promise<void> {
      try {
        // Check if workload has SourceAPI field set to WorkloadsV2
        if (workloadToDelete.sourceApi === SourceApi.WorkloadsV2) {
          await workloadService.deleteWorkloadManagerV2(workloadToDelete.id);
          return;
        }

        await workloadService.deleteWorkloadV2(workloadToDelete.id, workloadToDelete.type);

        this.setWorkloadPhase(workloadToDelete, Phase.Deleting);
      } catch (e: unknown) {
        console.error(e);
        throw e;
      }
    },
    removeWorkloadAdded(): void {
      this.workloadAdded = null;
    },
    setWorkloadAdded(
      workloadAdded:
        | IWorkspaceList
        | ITrainingList
        | IInferenceList
        | Workload
        | WorkspaceV2
        | TrainingV2
        | Inference
        | InferenceV2
        | ModelInference
        | ModelInferenceV2
        | Training
        | Workspace
        | WorkloadV2
        | DistributedWorkload,
      workloadType?: EWorkloadType,
    ): void {
      this.workloadAdded = workloadUtil.convertToWorkload(workloadAdded, workloadType);
    },
    async updateWorkloadAdded(workloads: Workload[]): Promise<void> {
      if (this.workloadAdded != null) {
        const workloadAdded: Workload | undefined = workloads.find(
          (workload: Workload) => workload.id === this.workloadAdded?.id,
        );
        if (workloadAdded) {
          this.workloadAdded = workloadAdded;
        } else {
          try {
            this.workloadAdded = await workloadService.getWorkloadById(this.workloadAdded.id);
          } catch (e) {
            console.error(e);
          }
        }
      }
    },
    isTopRow(workload: Workload): boolean {
      if (this.workloadAdded == null) return false;
      return this.workloadAdded.id === workload.id;
    },
    updatePhase(workload: Workload, phase: Phase): void {
      workload.phaseUpdatedAt = "";
      workload.phase = phase;
      workload.phaseMessage = "";
    },
    setWorkloadPhase(workloadToUpdate: Workload, phase: Phase): void {
      workloadLocalPhaseUpdater.updateLocalPhases(workloadToUpdate, phase);
      if (this.isTopRow(workloadToUpdate)) {
        this.updatePhase(this.workloadAdded as Workload, phase);
      } else {
        const workload: Workload | undefined = this.workloads.find(
          (workload: Workload) => workload.id === workloadToUpdate.id,
        );
        if (workload) {
          this.updatePhase(workload, phase);
        }
      }
    },
    async loadWorkloadHistory(workloadId: string, clusterUuid: string): Promise<void> {
      this.workloadHistoryRecords = (await workloadService.getWorkloadHistory(workloadId, clusterUuid)).records;
    },
    getWorkloadHistoryList(withFilter = false): HistoryRecord[] {
      if (!withFilter) return this.workloadHistoryRecords;
      let historyRecords: HistoryRecord[] = filterService.filterBySearchTermAndByColumns(
        this.workloadHistoryRecords,
        this.workloadHistoryFilterBy,
        eventHistoryColumns,
      );

      if (this.workloadHistoryFilterBy.type) {
        historyRecords = historyRecords.filter((historyRecord: HistoryRecord) => {
          return historyRecord.meta.type === this.workloadHistoryFilterBy.type;
        });
      }
      return historyRecords;
    },
    setWorkloadHistoryFilterBy(filterBy: IJobFilter): void {
      this.workloadHistoryFilterBy = filterBy;
    },
    /** Workload categories **/
    async loadWorkloadCategories(): Promise<void> {
      this.workloadCategories = await workloadService.listCategories();
    },
    /** YAML form state management **/
    setYamlFormState(state: IYamlWorkloadFormState): void {
      this.yamlFormState = state;
    },
    getYamlFormState(): IYamlWorkloadFormState | null {
      return this.yamlFormState;
    },
    clearYamlFormState(): void {
      this.yamlFormState = null;
    },
  },
});
