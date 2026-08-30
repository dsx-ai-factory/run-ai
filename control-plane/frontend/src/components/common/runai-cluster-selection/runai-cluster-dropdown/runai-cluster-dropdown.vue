<template>
  <runai-tooltip-wrapper :display-tooltip="displayTooltip" :tooltip-text="disabledTooltipText">
    <runai-select
      v-bind="$attrs"
      label="Cluster"
      aid="cluster-dropdown"
      :options="clusterOptions"
      :model-value="selectedClusterOption"
      @update:model-value="updateSelectedCluster"
      :rules="[isClusterSelected, ...rules]"
      :disable="disabled"
      max-width-item-tooltip="600px"
      unclearable
    />
  </runai-tooltip-wrapper>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { ValidationRule } from "quasar";

// cmps
import { RunaiSelect } from "@/components/common/runai-select";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";

// stores
import { useClusterStore } from "@/stores/cluster.store";

// models
import type { ClustersMinimalData } from "@/swagger-models/cluster-service-client";
import type { ISelectOption } from "@/models/global.model";

// constants
import { errorMessages } from "@/common/error-message.constant";
import { clusterUtil } from "@/utils/cluster.util";

export enum ERunaiClusterDropdownEvents {
  UpdateClusterId = "update:cluster-id",
}

export default defineComponent({
  name: "runai-cluster-dropdown",
  components: { RunaiTooltipWrapper, RunaiSelect },
  emits: {
    [ERunaiClusterDropdownEvents.UpdateClusterId]: (clusterId: string, payload: { disabled: boolean }) => {
      return typeof clusterId === "string" && typeof payload.disabled === "boolean";
    },
  },
  props: {
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    rules: {
      type: Array as PropType<ValidationRule[]>,
      default: () => [],
    },
    disabledTooltipText: {
      type: String as PropType<string>,
      required: false,
    },
    /**
     * A mapping of cluster IDs to their corresponding disable reasons.
     *
     * @param {Record<string, string>} disabledClusterOptions - An object where:
     *   - The key is the cluster uuid.
     *   - The value is a string explaining why the cluster is disabled.
     *   - Defaults to an empty object `{}` - no disable.
     */
    disabledClusterOptions: {
      type: Object as PropType<Record<string, string>>,
      default: () => ({}),
    },
  },
  data() {
    return {
      clusterStore: useClusterStore(),
    };
  },
  computed: {
    displayTooltip(): boolean {
      return this.disabled && !!this.disabledTooltipText;
    },
    clusterOptions(): Array<ISelectOption> {
      return this.clusterStore.clusterList.map((cluster: ClustersMinimalData) => {
        let disable = false;
        let disabledTooltip;

        if (this.disabledClusterOptions[cluster.uuid]) {
          disable = true;
          disabledTooltip = this.disabledClusterOptions[cluster.uuid];
        }

        return {
          label: clusterUtil.formatClusterNameWithVersion(cluster),
          value: cluster.uuid,
          disable,
          disabledTooltip,
        };
      });
    },
    selectedClusterOption(): ISelectOption | null {
      return this.clusterOptions.find((option) => option.value === this.clusterId) || null;
    },
  },
  methods: {
    updateSelectedCluster(clusterOption: ISelectOption | null): void {
      const clusterId = clusterOption?.value ? clusterOption.value.toString() : "";
      this.$emit(ERunaiClusterDropdownEvents.UpdateClusterId, clusterId, { disabled: clusterOption?.disable || false });
    },
    isClusterSelected(clusterOption: ISelectOption): boolean | string {
      return !!clusterOption || errorMessages.SELECT_CLUSTER;
    },
  },
});
</script>

<style scoped lang="scss"></style>
