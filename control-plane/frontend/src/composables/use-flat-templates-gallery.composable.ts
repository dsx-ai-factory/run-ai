/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { ICardQuery, TCardSortValues } from "@/components/common/runai-card-list/runai-card-list.model";
import {
  type TTemplateSortBy,
  workloadTemplateService,
} from "@/services/control-plane/workload-template.service/workload-template.service";
import {
  type DistributedTemplatesListResponseItem,
  type TrainingTemplatesListResponseItem,
  type WorkspaceTemplatesListResponseItem,
  WorkloadType,
} from "@/swagger-models/workloads-client";
import { ref } from "vue";

const mapSortByToTemplateSortBy: Record<TCardSortValues, TTemplateSortBy | undefined> = {
  name: "name",
  creation: "createdAt",
  recentUsage: undefined,
};

const templateListByType: Record<WorkloadType, CallableFunction> = {
  [WorkloadType.Training]: workloadTemplateService.trainingTemplatesList,
  [WorkloadType.Distributed]: workloadTemplateService.distributedTemplatesList,
  [WorkloadType.Workspace]: workloadTemplateService.workspaceTemplatesList,
  [WorkloadType.Inference]: workloadTemplateService.inferenceTemplatesList,
  [WorkloadType.DistributedInference]: () => {
    throw new Error("Distributed inference templates are not supported");
  },
};

const listBaseFilterBy: Record<WorkloadType, string> = {
  [WorkloadType.Training]: "workloadType!=Workspace,workloadType!=Distributed,workloadType!=Inference",
  [WorkloadType.Distributed]: "workloadType!=Training,workloadType!=Workspace,workloadType!=Inference",
  [WorkloadType.Workspace]: "workloadType!=Training,workloadType!=Distributed,workloadType!=Inference",
  [WorkloadType.Inference]: "workloadType!=Training,workloadType!=Distributed,workloadType!=Workspace",
  [WorkloadType.DistributedInference]: "",
};

export function useFlatTemplatesGallery() {
  const listQuery = ref<ICardQuery>({
    search: "",
    sortBy: "creation",
    filterBy: "",
  });

  const offset = ref<number | undefined>(undefined);
  const totalCount = ref<number>(0);
  const templates = ref<
    TrainingTemplatesListResponseItem[] | DistributedTemplatesListResponseItem[] | WorkspaceTemplatesListResponseItem[]
  >([]);

  const setTemplates = (
    newTemplates:
      | TrainingTemplatesListResponseItem[]
      | DistributedTemplatesListResponseItem[]
      | WorkspaceTemplatesListResponseItem[],
  ) => {
    templates.value = newTemplates;
  };

  const setListQuery = (newQuery: ICardQuery) => {
    listQuery.value = newQuery;
  };

  const list = async (workloadType: WorkloadType, projectId: string, applyOffset = false, amountRequired?: number) => {
    const filterBy = `${listBaseFilterBy[workloadType]},projectId==${projectId}`;

    if (!applyOffset) {
      totalCount.value = await workloadTemplateService.templatesCount(
        [filterBy],
        listQuery.value.search?.toString(),
        true,
      );
    }

    // this logic is mostly due to issues caused by the emptyCard we add to the list
    // if there is a filter we ask for 6 cards, otherwise we ask for 5 as we have to present the emptyCard as well
    // if an offset (pages higher than 1) is provided we ask for 6 cards as well as the emptycard is present only on the first page of the gallery
    // amountRequired is a special case, we use it (when we have only loaded the part of the batch and user asks not to see the next batch directly but pages which skip it directly so we need to ask for the amount in between those pages)
    // i.e: if we have only loaded page 1 and 2 but suddenly user asks to see page 4, then we need to ask for the total amount of 3 AND 4 to fill the gap.
    const limit = amountRequired || (listQuery.value.search?.toString() || applyOffset ? 6 : 5);

    const listFunction = templateListByType[workloadType];

    if (!listFunction) {
      throw new Error(`List function for workload type ${workloadType} not found`);
    }

    const response = await listFunction(
      applyOffset ? offset.value : undefined,
      limit,
      "asc",
      mapSortByToTemplateSortBy[listQuery.value.sortBy as TCardSortValues],
      [filterBy],
      listQuery.value.search?.toString(),
      true,
      projectId,
    );

    if (applyOffset) {
      setTemplates([...templates.value, ...(response.entries || [])]);
    } else {
      setTemplates(response.entries || []);
    }

    offset.value = response.pagination?.nextOffset || undefined;
    // list was successfully loaded, saving the last filterBy used
    listQuery.value.filterBy = filterBy;
  };

  return {
    list,
    templates,
    totalCount,
    listQuery,
    offset,
    setTemplates,
    setListQuery,
  };
}
