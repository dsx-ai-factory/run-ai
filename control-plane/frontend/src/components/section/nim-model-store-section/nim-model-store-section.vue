<template>
  <runai-expansion-item
    :default-opened="defaultOpen"
    label="Model store"
    :disable-closing="disableClosing"
    class="nim-model-store-section"
    aid="nim-model-store-section"
    should-validate
  >
    <template #subheader>
      <span>{{ summary }}</span>
    </template>
    <section class="main-section">
      <nim-model-storage
        :data-sources="dataSources"
        :cluster-id="clusterId"
        :selected-data-sources="modelStoreData.selectedDataSources"
        :created-data-source-id="createdDataSourceId"
        @create-new="$emit('create-new', $event)"
        @datasource-changed="onDataSourceChanged"
        :required="required"
        :policy-rules="policyRules"
      />
    </section>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// components
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { NimModelStorage } from "./nim-model-storage";

// models
import type { INimModelStoreSection } from "./nim-model-store-section.model";
import type { DatasourceListResponseEntry, PolicyInfo } from "@/swagger-models/assets-service-client";

// others

export default defineComponent({
  name: "nim-model-store-section",
  components: {
    RunaiExpansionItem,
    NimModelStorage,
  },
  emits: ["model-store-data-changed", "create-new"],
  props: {
    disableClosing: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    defaultOpen: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    modelStoreData: {
      type: Object as PropType<INimModelStoreSection>,
      required: true,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    dataSources: {
      type: Array as PropType<Array<DatasourceListResponseEntry>>,
      required: true,
    },
    createdDataSourceId: {
      type: String as PropType<string>,
      required: false,
    },
    required: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<PolicyInfo | null>,
      required: false,
    },
  },
  computed: {
    summary(): string {
      const selectedDataSourceAsset = this.modelStoreData.selectedDataSources[0];
      if (!selectedDataSourceAsset) return "None";

      const selectedDataSource = this.dataSources.find(
        (ds: DatasourceListResponseEntry) => ds.meta.id === selectedDataSourceAsset.id,
      );

      return selectedDataSource?.meta.name || "None";
    },
  },
  methods: {
    onDataSourceChanged(selectedDataSourceCards: string[]): void {
      this.$emit("model-store-data-changed", {
        ...this.modelStoreData,
        selectedDataSources: selectedDataSourceCards,
      });
    },
  },
});
</script>

<style scoped lang="scss"></style>
