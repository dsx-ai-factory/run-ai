<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="workload-details" :key="workload.id">
    <details-panel-view
      :loading="loading"
      :sections="accordionSections"
      v-model:selected-spec="selectedSpec"
      :show-spec-select="showSpecSelect"
      :spec-options="specOptions"
      :empty-messages="['There are no details for this workload yet.']"
    />
  </section>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { Notify } from "quasar";
//models
import type {
  Workload,
  WorkloadDetailed,
  WorkloadNodeAffinity,
  DistributedWorkload,
} from "@/swagger-models/workloads-client";
import type { ISelectOption } from "@/models/global.model";
import type { IDetailsPanelSectionDef } from "@/models/details-panel.model";
import { EWorkloadType, EDistributedReplicaType } from "@/models/workload.model";
//services
import { workloadService, type CrudWorkload } from "@/services/cluster/workload.service/workload.service";
//stores
import { useClusterStore } from "@/stores/cluster.store";
//cmps
import DetailsPanelView from "@/components/workload/workload-drawer-content/tabs/workload-details-tab/details-panel-view.vue";
//utils
import { workloadDetailsUtil, type WorkloadSpec } from "@/utils/workload-details.util";
import { workloadV2ManifestUtil } from "@/utils/workload-v2-manifest.util";
import { alertUtil } from "@/utils/alert.util";
import { SourceApi } from "@/swagger-models/workloads-client";
import { NotFoundError, type HttpErrorResponse } from "@/models/http-response.model";
import { useTopologyLevels } from "@/composables/use-topology-levels.composable/use-topology-levels.composable";

defineOptions({ name: "workload-details-tab" });

const props = withDefaults(
  defineProps<{
    workload: Workload;
    isDeletedViewType?: boolean;
  }>(),
  {
    isDeletedViewType: false,
  },
);

const clusterStore = useClusterStore();

const workloadDetailed = ref<WorkloadDetailed | null>(null);
const workloadCrud = ref<CrudWorkload | null>(null);
const nodeAffinities = ref<WorkloadNodeAffinity[]>([]);
const loading = ref(false);
const selectedSpec = ref<EDistributedReplicaType>(EDistributedReplicaType.Master);

const { topologyLevels, fetchTopologyLevels } = useTopologyLevels();

const isNativeWorkload = computed<boolean>(() => {
  const nativeTypes = [
    EWorkloadType.Workspace,
    EWorkloadType.Training,
    EWorkloadType.Inference,
    EWorkloadType.Distributed,
    EWorkloadType.DistributedInference,
  ];
  return nativeTypes.includes(props.workload.type as EWorkloadType);
});

const isWorkloadsV2Workload = computed<boolean>(() => props.workload.sourceApi === SourceApi.WorkloadsV2);

// Soft-deleted workloads remain in v1; v2 queue rows are hard-deleted after agent confirm.
const useV2Api = computed<boolean>(() => isWorkloadsV2Workload.value && !props.isDeletedViewType);

const isDistributed = computed<boolean>(() => props.workload.type === EWorkloadType.Distributed);
const isDistributedInference = computed<boolean>(() => props.workload.type === EWorkloadType.DistributedInference);
const hasReplicaSelector = computed<boolean>(() => isDistributed.value || isDistributedInference.value);

const specOptions = computed<ISelectOption[]>(() => {
  if (!workloadCrud.value) return [];

  // "Worker" is pluralized as "Workers" in the UI because a workload can have many worker pods,
  // unlike the single master/leader. The underlying value stays as the enum string.
  const workerOption: ISelectOption = { label: "Workers", value: EDistributedReplicaType.Worker };

  if (isDistributed.value) {
    const distributedWorkload = workloadCrud.value as DistributedWorkload;
    const options: ISelectOption[] = [];
    if (distributedWorkload.masterSpec) {
      options.push({ label: "Master", value: EDistributedReplicaType.Master });
    }
    if (distributedWorkload.spec) options.push(workerOption);
    return options;
  }

  if (isDistributedInference.value) {
    const spec = (workloadCrud.value as { spec?: { leader?: unknown; worker?: unknown } }).spec;
    const options: ISelectOption[] = [];
    if (spec?.leader) options.push({ label: "Leader", value: EDistributedReplicaType.Leader });
    if (spec?.worker) options.push(workerOption);
    return options;
  }

  return [];
});

const showSpecSelect = computed<boolean>(() => {
  // Hide the selector when there's only one replica role (e.g., a TFJob with workers and
  // no master, or a single-pod DistributedInference with only a leader). A 1-item dropdown
  // adds no value — the section data is the same regardless of the selection.
  return hasReplicaSelector.value && specOptions.value.length > 1;
});

const currentSpec = computed<WorkloadSpec | undefined>(() => {
  const crudWorkload = workloadCrud.value ?? undefined;
  const selectedReplicaType = hasReplicaSelector.value ? selectedSpec.value : undefined;

  return workloadDetailsUtil.getSpecSection(crudWorkload, selectedReplicaType);
});

const currentNodeAffinity = computed<WorkloadNodeAffinity | undefined>(() => {
  if (!nodeAffinities.value.length) return undefined;
  if (nodeAffinities.value.length === 1) return nodeAffinities.value[0];
  return (
    nodeAffinities.value.find((entry) => entry.specName?.toLowerCase() === selectedSpec.value?.toLowerCase()) ??
    nodeAffinities.value[0]
  );
});

function getClusterName(): string {
  if (!workloadDetailed.value?.clusterId) return "";
  const cluster = clusterStore.getClusterById(workloadDetailed.value.clusterId);
  return cluster?.name || workloadDetailed.value.clusterId;
}

// The accordion-section model consumed by the shared details-panel-view. Built here (rather
// than in a separate "view" component) since this is the only consumer of the data this tab
// already fetched — no reuse to justify the extra layer.
const accordionSections = computed<IDetailsPanelSectionDef[]>(() => {
  if (!workloadDetailed.value) return [];
  return workloadDetailsUtil.buildSearchableSections({
    workloadDetailed: workloadDetailed.value,
    workloadCrud: workloadCrud.value ?? undefined,
    currentSpec: currentSpec.value,
    clusterName: getClusterName(),
    topologyLevels: topologyLevels.value,
    nodeAffinity: currentNodeAffinity.value?.nodeAffinity,
  });
});

async function fetchWorkloadDetailed(): Promise<boolean> {
  try {
    workloadDetailed.value = await workloadService.getWorkloadById(props.workload.id);
    return true;
  } catch (error) {
    return false;
  }
}

async function fetchWorkloadV2Details(): Promise<boolean> {
  try {
    const response = await workloadService.getWorkloadV2ById(props.workload.id);
    workloadCrud.value =
      (workloadV2ManifestUtil.getCrudWorkloadFromManifest(
        response.workload.rawManifest,
        props.workload.type,
      ) as CrudWorkload) ?? null;
    if (!workloadDetailed.value) {
      workloadDetailed.value = props.workload as WorkloadDetailed;
    }
    return true;
  } catch (error) {
    if ((error as HttpErrorResponse)?.statusCode === NotFoundError.statusCode) {
      return fetchWorkloadDetailed();
    }
    return false;
  }
}

async function fetchWorkloadCrud(): Promise<boolean> {
  try {
    workloadCrud.value = await workloadService.getWorkloadCrudById(props.workload.id, props.workload.type);
    return true;
  } catch (error) {
    return false;
  }
}

async function fetchWorkloadDetails(): Promise<void> {
  loading.value = true;
  nodeAffinities.value = [];
  workloadCrud.value = null;
  workloadDetailed.value = null;

  // Supported (external) workloads have no CRUD/asset representation — the CRUD endpoint
  // throws for non-native types. Fetch it only for native workloads and render everything
  // else from WorkloadDetailed alone.
  // Deleted view: always use v1 get-by-id (soft-deleted rows); v2 get-by-id returns 404.
  const detailedPromise = useV2Api.value ? Promise.resolve(false) : fetchWorkloadDetailed();
  const crudPromise =
    !isWorkloadsV2Workload.value && isNativeWorkload.value ? fetchWorkloadCrud() : Promise.resolve(false);
  const workloadV2Promise = useV2Api.value ? fetchWorkloadV2Details() : Promise.resolve(false);

  const [hasDetailed, hasCrud, hasWorkloadV2] = await Promise.all([detailedPromise, crudPromise, workloadV2Promise]);

  // Show error only if everything we tried to load failed.
  if (!hasDetailed && !hasCrud && !hasWorkloadV2) {
    Notify.create(alertUtil.getError("Failed to load workload details"));
  }

  // Snap selectedSpec to the first available replica option — the data-level default
  // (Master) is invalid for DistributedInference, which only has Leader/Worker.
  const options = specOptions.value;
  if (options.length > 0 && !options.some((o) => o.value === selectedSpec.value)) {
    selectedSpec.value = options[0].value as EDistributedReplicaType;
  }

  const detailed = workloadDetailed.value as WorkloadDetailed | null;
  nodeAffinities.value = detailed?.nodeAffinities ?? [];
  await fetchTopologyLevels(detailed?.topology?.name, detailed?.clusterId);

  loading.value = false;
}

watch(() => props.workload.id, fetchWorkloadDetails, { immediate: true });
</script>

<style scoped lang="scss">
.workload-details {
  height: 100%;
}

:deep(.item-details-item) {
  margin-bottom: 24px;
}
</style>
