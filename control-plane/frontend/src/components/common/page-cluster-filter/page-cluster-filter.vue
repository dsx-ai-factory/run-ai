<template>
  <runai-filter-chip
    title="Cluster"
    :value="clusterName"
    @save="onClusterFilterChanged"
    :icon-right="outerChipIcon"
    :save-btn-label="saveFilterBtnLabel"
    :save-btn-icon-right="saveFilterBtnIcon"
    @before-show="initSelectedClusterId"
    @close="selectedClusterId = filterValue"
  >
    <template v-slot:input-component>
      <div v-if="showClusterSearch" class="q-pl-sm q-pr-md q-pb-sm cluster-search-input">
        <q-input v-model="searchTerm" type="search" dense placeholder="Search..." aria-label="Search clusters">
          <template v-slot:prepend>
            <q-icon name="ph-bold ph-magnifying-glass" size="12px" />
          </template>
        </q-input>
      </div>
      <runai-radio-options
        class="cluster-radio-buttons"
        v-model="selectedClusterId"
        :options="filteredOptions"
        option-max-width="200px"
        :min-chars-to-show-overflow-tooltip="30"
        scrollable-y
      ></runai-radio-options>
    </template>
  </runai-filter-chip>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// stores
import { useClusterStore } from "@/stores/cluster.store";
// cmps
import { RunaiFilterChip } from "@/components/common/runai-filter-chip";
// type
import type { ISelectOption } from "@/models/global.model";

import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { ALL_CLUSTER_FILTER_LABEL, ALL_CLUSTER_FILTER_VALUE, GLOBAL_FILTER_ICON } from "@/models/filter.model";
import type { ClustersMinimalData } from "@/swagger-models/cluster-service-client";

const allOption: ISelectClusterOption = { label: ALL_CLUSTER_FILTER_LABEL, value: ALL_CLUSTER_FILTER_VALUE };
const MIN_CLUSTERS_AMOUNT_TO_SHOW_SEARCH = 10;
export interface ISelectClusterOption extends ISelectOption {
  value: string;
}

export default defineComponent({
  components: {
    RunaiFilterChip,
    RunaiRadioOptions,
  },
  emits: ["update-filter"],
  props: {
    filterValue: {
      type: String as PropType<string>,
      required: true,
    },
    hideAllOption: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    applyGlobally: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    allowedClusterIds: {
      type: Array as PropType<string[]>,
      default: null,
    },
  },
  data() {
    return {
      clusterStore: useClusterStore(),
      selectedClusterId: "" as ISelectClusterOption["value"],
      searchTerm: "",
      thumbStyle: {
        right: "4px",
        borderRadius: "4px",
        backgroundColor: "#00000089",
        width: "7px",
        opacity: 1,
      },
    };
  },
  created() {
    this.initSelectedClusterId();
  },
  computed: {
    hideGlobalFilter(): boolean {
      return !this.applyGlobally || this.selectedClusterId === ALL_CLUSTER_FILTER_VALUE || !this.isMultiCluster;
    },
    saveFilterBtnLabel(): string | undefined {
      return this.hideGlobalFilter ? undefined : "Apply to all pages";
    },
    saveFilterBtnIcon(): string | undefined {
      return this.hideGlobalFilter ? undefined : GLOBAL_FILTER_ICON;
    },
    outerChipIcon(): string | undefined {
      return this.clusterName === ALL_CLUSTER_FILTER_LABEL || !this.isMultiCluster ? undefined : GLOBAL_FILTER_ICON;
    },
    isMultiCluster(): boolean {
      return this.clusterStore.isMultiCluster;
    },
    options(): ISelectClusterOption[] {
      let options = this.getClusterOptions();
      options = this.moveSelectedClusterToTop(options);
      if (!this.hideAllOption) {
        options.unshift(allOption);
      }
      return options;
    },
    clusterName(): string {
      if (!this.applyGlobally) {
        return this.options.find((option) => option.value === this.filterValue)?.label || "";
      }
      if (!this.filterValue && !this.hideAllOption) {
        return ALL_CLUSTER_FILTER_LABEL;
      }
      return this.clusterStore.selectedClusterName;
    },
    showClusterSearch(): boolean {
      return this.clustersAmount >= MIN_CLUSTERS_AMOUNT_TO_SHOW_SEARCH;
    },
    filteredOptions(): ISelectClusterOption[] {
      if (!this.searchTerm) return this.options;
      return this.options.filter((o) => o.label.toLowerCase().includes(this.searchTerm.toLowerCase()));
    },
    clustersAmount(): number {
      return this.clusterStore.clustersAmount;
    },
  },
  methods: {
    initSelectedClusterId(): void {
      this.searchTerm = "";
      if (!this.applyGlobally) {
        this.selectedClusterId = this.filterValue;
        return;
      }
      if (this.filterValue) {
        this.selectedClusterId = this.clusterStore.selectedClusterId;
      } else {
        this.selectedClusterId = this.hideAllOption ? this.clusterStore.selectedClusterId : ALL_CLUSTER_FILTER_VALUE;
        this.onClusterFilterChanged();
      }
    },
    setGlobalClusterId(clusterId: string): void {
      this.clusterStore.setSelectedClusterById(clusterId);
    },
    onClusterFilterChanged(): void {
      if (this.applyGlobally && this.selectedClusterId !== ALL_CLUSTER_FILTER_VALUE) {
        this.setGlobalClusterId(this.selectedClusterId);
      }

      this.$emit("update-filter", this.selectedClusterId);
    },
    getClusterOptions(): ISelectClusterOption[] {
      let clusters = this.clusterStore.clusterList;

      // Filter by allowed cluster IDs if provided
      if (this.allowedClusterIds && this.allowedClusterIds.length > 0) {
        clusters = clusters.filter((c: ClustersMinimalData) => this.allowedClusterIds.includes(c.uuid));
      }

      return clusters
        .map((c: ClustersMinimalData) => ({ label: c.name, value: c.uuid }))
        .sort((a, b) => a.label.localeCompare(b.label));
    },
    moveSelectedClusterToTop(options: ISelectClusterOption[]): ISelectClusterOption[] {
      const selectedClusterIndex = options.findIndex((option) => option.value === this.filterValue);
      if (selectedClusterIndex < 0) return options;
      const [selectedCluster] = options.splice(selectedClusterIndex, 1);
      options.unshift(selectedCluster);
      return options;
    },
  },
  watch: {
    filterValue() {
      this.initSelectedClusterId();
    },
  },
});
</script>

<style lang="scss" scoped>
.cluster-search-input {
  position: sticky;
  top: 0;
  background-color: var(--bg-primary-default);
  z-index: 10;
}
</style>
