<template>
  <runai-drawer-content dense-header>
    <template v-slot:header v-if="!showNoData">
      <section class="row items-center q-gutter-x-md">
        <div class="ellipsis">{{ selectedApplication?.name }}</div>
        <span class="text-capitalize">{{ selectedApplication?.phase }}</span>
      </section>
    </template>

    <template v-slot:subheader v-if="!showNoData">
      <section class="diagram-subheader row items-center justify-between no-wrap full-width">
        <runai-filter-chip
          title="Type"
          :value="resourceTypeChipValue"
          @before-show="handleTypeFilterOpen"
          @close="handleTypeFilterClose"
          @save="handleTypeFilterSave"
        >
          <template #input-component>
            <runai-checkbox-list
              :key="typeFilterListKey"
              class="resource-type-checkbox-list"
              :options="resourceTypeOptions"
              :selected="draftSelectedResourceKinds"
              :show-search="false"
              @changed="handleDraftResourceKindsChanged"
            />
          </template>
        </runai-filter-chip>
        <runai-general-search :search-term="searchTerm" @update-search-term="handleSearchTermChanged" />
      </section>
    </template>

    <template v-slot:body>
      <runai-default-loader v-if="loading" />
      <runai-drawer-no-data
        v-else-if="showNoData"
        :messages="[`Select ${selectedApplicationsAmount > 1 ? 'a single' : 'an'} AI application to view its details.`]"
        icon-name="ph-bold ph-file"
        icon-class="text-icon-default"
      />
      <section v-else class="application-details">
        <ai-application-diagram
          :resources="diagramResources"
          :workloads="workloads"
          :resource-connections="applicationDetails?.resourceConnections"
          :endpoints="applicationDetails?.endpoints"
          :selected-resource-kinds="selectedResourceKinds"
          :search-term="searchTerm"
        />
      </section>
    </template>
  </runai-drawer-content>
</template>

<script setup lang="ts">
import { ref, watch, onUnmounted, computed } from "vue";
import { useQuasar } from "quasar";
import { RunaiDrawerContent } from "@/components/common/runai-drawer-content";
import { RunaiDrawerNoData } from "@/components/common/runai-drawer-no-data";
import { RunaiDefaultLoader } from "@/components/common/runai-default-loader";
import { RunaiCheckboxList } from "@/components/common/runai-checkbox-list";
import type { ICheckboxItem } from "@/components/common/runai-checkbox-list/model";
import { RunaiFilterChip } from "@/components/common/runai-filter-chip";
import { RunaiGeneralSearch } from "@/components/common/runai-general-search";
import type { IApplicationTable } from "@/models/ai-application.model";
import { EAiApplicationErrorMessage } from "@/models/ai-application.model";
import type { Workload } from "@/swagger-models/workloads-client";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { aiApplicationService } from "@/services/cluster/ai-application.service/ai-application.service";
import { intervalUtil } from "@/utils/interval.util";
import { EIntervalLabels } from "@/models/interval.model";
import { ErrorAlert } from "@/utils/error-alert.util";
import AiApplicationDiagram from "./ai-application-diagram.vue";
import { EResourceKind, FILTERABLE_RESOURCE_KINDS } from "./ai-application-diagram.util";

const $q = useQuasar();

export interface Props {
  drawerWidth: string;
  selectedApplicationsAmount: number;
  selectedApplication?: IApplicationTable | null;
}

const props = withDefaults(defineProps<Props>(), {
  selectedApplication: null,
});

const applicationDetails = ref<IApplicationTable | null>(null);
const workloads = ref<Workload[]>([]);
const loadingDetails = ref(false);
const loadingWorkloads = ref(false);
const currentApplicationId = ref<string | null>(null);
const resourceTypeOptions: ICheckboxItem[] = [
  { value: EResourceKind.Workload, label: "Workload" },
  { value: EResourceKind.Endpoint, label: "Endpoint" },
  { value: EResourceKind.Secret, label: "Secret" },
  { value: EResourceKind.PersistentVolumeClaim, label: "PVC" },
  { value: EResourceKind.ConfigMap, label: "ConfigMap" },
];
const selectedResourceKinds = ref<EResourceKind[]>([...FILTERABLE_RESOURCE_KINDS]);
const draftSelectedResourceKinds = ref<string[]>([...FILTERABLE_RESOURCE_KINDS]);
const typeFilterListKey = ref(0);
const searchTerm = ref("");

const loading = computed(() => loadingDetails.value || loadingWorkloads.value);
const showNoData = computed(
  () => !props.selectedApplication || !props.selectedApplicationsAmount || props.selectedApplicationsAmount > 1,
);

// Filtering services from the diagram resources
const diagramResources = computed(() => (applicationDetails.value?.resources || []).filter((r) => r.kind !== "Service"));
const resourceTypeChipValue = computed(() => {
  if (selectedResourceKinds.value.length === FILTERABLE_RESOURCE_KINDS.length) return "All";
  if (selectedResourceKinds.value.length === 0) return "None";
  return resourceTypeOptions
    .filter((option) => selectedResourceKinds.value.includes(option.value as EResourceKind))
    .map((option) => option.label)
    .join(", ");
});

const handleTypeFilterOpen = (): void => {
  draftSelectedResourceKinds.value = [...selectedResourceKinds.value];
  typeFilterListKey.value += 1;
};

const handleTypeFilterClose = (): void => {
  draftSelectedResourceKinds.value = [...selectedResourceKinds.value];
};

const handleTypeFilterSave = (): void => {
  const validKinds = new Set(FILTERABLE_RESOURCE_KINDS);
  selectedResourceKinds.value = draftSelectedResourceKinds.value.filter((kind): kind is EResourceKind =>
    validKinds.has(kind as EResourceKind),
  );
};

const handleDraftResourceKindsChanged = (selectedKinds: string[]): void => {
  draftSelectedResourceKinds.value = selectedKinds;
};

const handleSearchTermChanged = (value: string): void => {
  searchTerm.value = value;
};

const fetchApplicationDetails = async (applicationId: string): Promise<void> => {
  try {
    loadingDetails.value = true;
    const app = await aiApplicationService.getApplicationById(applicationId);
    applicationDetails.value = app as IApplicationTable;
  } catch (error: unknown) {
    const errorAlert = new ErrorAlert({ generalMessage: EAiApplicationErrorMessage.FailedToLoadApplicationDetails });
    $q.notify(errorAlert.getNotification(error));
    applicationDetails.value = null;
  } finally {
    loadingDetails.value = false;
  }
};

const fetchApplicationWorkloads = async (applicationId: string, showLoading = true): Promise<void> => {
  try {
    if (showLoading) {
      loadingWorkloads.value = true;
    }
    const filterBy = [`aiApplicationId==${applicationId}`];
    const fetchedWorkloads = await workloadService.getWorkloads(false, 0, 100, undefined, undefined, filterBy);
    workloads.value = fetchedWorkloads;
  } catch (error: unknown) {
    const errorAlert = new ErrorAlert({ generalMessage: EAiApplicationErrorMessage.FailedToLoadApplicationWorkloads });
    $q.notify(errorAlert.getNotification(error));
    workloads.value = [];
  } finally {
    if (showLoading) {
      loadingWorkloads.value = false;
    }
  }
};

const startPolling = (applicationId: string): void => {
  intervalUtil.stopInterval(EIntervalLabels.AiApplicationDrawerWorkloads);
  fetchApplicationDetails(applicationId);
  fetchApplicationWorkloads(applicationId, true);
  intervalUtil.startInterval(EIntervalLabels.AiApplicationDrawerWorkloads, () => {
    fetchApplicationWorkloads(applicationId, false);
  });
};

const stopPolling = (): void => {
  intervalUtil.stopInterval(EIntervalLabels.AiApplicationDrawerWorkloads);
};

const resetState = (): void => {
  currentApplicationId.value = null;
  applicationDetails.value = null;
  workloads.value = [];
  stopPolling();
};

watch(
  () => props.selectedApplication,
  (newApplication) => {
    if (newApplication?.id && newApplication.id !== currentApplicationId.value) {
      currentApplicationId.value = newApplication.id;
      startPolling(newApplication.id);
    } else if (!newApplication?.id) {
      resetState();
    }
  },
  { immediate: true },
);

onUnmounted(() => {
  stopPolling();
});
</script>

<style lang="scss">
.resource-type-checkbox-list {
  padding: 0;

  > .q-item {
    min-height: 0;
    padding: 0;
  }

  .q-checkbox {
    min-height: 32px;
  }
}
</style>

<style lang="scss" scoped>
.diagram-subheader {
  padding: 0 14px;
}

.application-details {
  padding: 4px;
  height: 100%;
  display: flex;
  flex-direction: column;
}
</style>
