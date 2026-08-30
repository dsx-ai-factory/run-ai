<template>
  <runai-expansion-item
    label="Cluster"
    default-opened
    :prevent-top-padding-when-open="preventTopPaddingWhenOpen"
    :disable-closing="disableClosing"
    aid="cluster-selection-section"
    should-validate
  >
    <template #subheader>
      <span aid="cluster-section-subheader"> {{ clusterName || "None" }}</span>
    </template>
    <div class="q-pt-sm q-pb-md">Set under which cluster to create this {{ entityName }}</div>
    <div class="row">
      <div class="col-10">
        <runai-cluster-dropdown
          :disabled="disabled"
          :cluster-id="clusterId"
          @update:cluster-id="onClusterChanged"
          disabled-tooltip-text="This can't be modified"
          :disabled-cluster-options="disabledClusterOptions"
        />
      </div>
    </div>
  </runai-expansion-item>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";

//store
import { useClusterStore } from "@/stores/cluster.store";

//cmps
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiClusterDropdown } from "@/components/common/runai-cluster-selection/runai-cluster-dropdown";
import { clusterUtil } from "@/utils/cluster.util";

export enum ERunaiClusterSectionEvents {
  UpdateClusterId = "update:cluster-id",
}

export default defineComponent({
  name: "runai-cluster-section",
  components: { RunaiClusterDropdown, RunaiExpansionItem },
  emits: {
    [ERunaiClusterSectionEvents.UpdateClusterId]: (clusterId: string, payload: { disabled: boolean }) => {
      return typeof clusterId === "string" && typeof payload.disabled === "boolean";
    },
  },
  props: {
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    entityName: {
      type: String as PropType<string>,
      required: true,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      default: false,
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
    preventTopPaddingWhenOpen: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    disableClosing: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      clusterStore: useClusterStore(),
    };
  },
  computed: {
    clusterName(): string | null {
      const cluster = clusterUtil.findClusterById(this.clusterStore.clusterList, this.clusterId);
      return clusterUtil.formatClusterNameWithVersion(cluster);
    },
  },
  methods: {
    onClusterChanged(clusterId: string, payload: { disabled: boolean }): void {
      this.$emit(ERunaiClusterSectionEvents.UpdateClusterId, clusterId, payload);
    },
  },
  watch: {
    clusterId: {
      handler(clusterId: string): void {
        //Viper:TODO - we need to remove this line and setSelectedClusterById only upon form submit
        this.clusterStore.setSelectedClusterById(clusterId);
      },
    },
  },
});
</script>

<style scoped lang="scss"></style>
