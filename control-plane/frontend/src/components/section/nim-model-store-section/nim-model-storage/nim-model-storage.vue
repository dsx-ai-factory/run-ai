<template>
  <runai-section
    title="Select the data source that will serve as a model store"
    tooltip-text="If the model is already stored on the selected data source, it will be loaded from there automatically. Otherwise, it will be stored on the selected data source during the first workload deployment."
    tooltip-width="445px"
  >
    <data-source-gallery
      :entity-type="entityType"
      :loading="loading"
      :data-sources="filteredDataSources"
      :selected-data-sources="selectedDataSources"
      :cluster-id="clusterId"
      :created-data-source-id="createdDataSourceId"
      :allowed-created-types="allowedCreatedTypes"
      :policy-rules="policyRules"
      @create-new="$emit('create-new', $event)"
      :container-path-override-allowed="false"
      :multi-select="false"
      @datasource-changed="$emit('datasource-changed', $event)"
    />
    <div class="col-12 q-ml-sm" v-if="required">
      <q-field class="hidden-field gallery-message" :model-value="selectedDataSources" :rules="[isRequired]"></q-field>
    </div>
  </runai-section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// components
import { RunaiSection } from "@/components/common/runai-section";
import { DataSourceGallery } from "@/components/section/data-source-section/data-source-gallery";

// models
import type { AssetIdAndKind, DatasourceListResponseEntry, PolicyInfo } from "@/swagger-models/assets-service-client";
import { AssetKind } from "@/swagger-models/assets-service-client";
import { EWorkloadFormType } from "@/models/workload.model";
import type { ICardListItem } from "@/components/common/runai-card-list/runai-card-list.model";

// stores
import { useAppStore } from "@/stores/app.store";
import { errorMessages } from "@/common/error-message.constant";
import { dataSourceUtil } from "@/utils/data-source.util";
import type { TDataSourceKinds } from "@/models/data-source.model";

export default defineComponent({
  name: "nim-model-storage",
  components: {
    RunaiSection,
    DataSourceGallery,
  },
  emits: ["datasource-changed", "create-new"],
  props: {
    dataSources: {
      type: Array as PropType<Array<DatasourceListResponseEntry>>,
      required: true,
    },
    selectedDataSources: {
      type: Array as PropType<Array<AssetIdAndKind>>,
      required: true,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    createdDataSourceId: {
      type: String as PropType<string>,
      required: false,
    },
    loading: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<PolicyInfo | null>,
      required: false,
    },
    required: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
  },
  data() {
    return {
      appStore: useAppStore(),
      entityType: EWorkloadFormType.Inference as EWorkloadFormType,
    };
  },
  computed: {
    shouldBeOpened(): boolean {
      return this.required || this.selectedDataSources.length > 0;
    },
    allowedCreatedTypes(): Set<TDataSourceKinds> {
      const allowedDsTypes = new Set<TDataSourceKinds>([AssetKind.Nfs, AssetKind.Pvc, AssetKind.HostPath]);
      return dataSourceUtil.filterDataSourcesByPolicyRules<TDataSourceKinds>(
        allowedDsTypes,
        this.policyRules?.rules?.assets?.datasources,
      );
    },
    filteredDataSources(): Array<DatasourceListResponseEntry> {
      return this.dataSources.filter((dataSource: DatasourceListResponseEntry) => {
        return this.allowedCreatedTypes.has(dataSource.meta.kind as TDataSourceKinds);
      });
    },
  },
  methods: {
    isRequired(val: Array<ICardListItem>): boolean | string {
      return val[0] && val[0].id ? true : errorMessages.SELECT_DATA_SOURCE;
    },
  },
});
</script>

<style scoped lang="scss"></style>
