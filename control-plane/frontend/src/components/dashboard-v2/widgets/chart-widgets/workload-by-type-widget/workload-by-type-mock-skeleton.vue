<template>
  <div class="workload-by-type-mock-skeleton full-width q-pa-md" style="height: 240px">
    <div class="column full-width full-height justify-end">
      <div class="row items-end justify-center full-width" style="flex: 1">
        <div class="col row items-end justify-center full-height" v-for="i in 4" :key="i">
          <q-skeleton type="rect" :style="{ height: getHeight(i), width: '100%', maxWidth: '60px' }" />
        </div>
      </div>
      <div class="chart-axis-labels">
        <span aid="workload-by-type-drilldown-label" class="text-weight-bold cluster-name-label">
          {{ clusterName }}
        </span>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

export default defineComponent({
  name: "workload-by-type-mock-skeleton",
  props: {
    clusterName: {
      type: String as PropType<string>,
      required: true,
    },
  },
  methods: {
    getHeight(index: number): string {
      // Generate different heights for visual variety
      const heights = ["60%", "80%", "70%", "50%"];
      return heights[index - 1];
    },
  },
});
</script>

<style lang="scss" scoped>
.workload-by-type-mock-skeleton {
  .chart-axis-labels {
    margin: auto;
    .cluster-name-label {
      color: rgb(6, 84, 254);
      cursor: pointer;
      font-size: 11px;
      text-decoration: underline;
    }
  }
}
</style>
