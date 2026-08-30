<template>
  <figure class="runai-resources-chart" role="img" :aria-label="accessibleDescription">
    <v-chart
      aid="runai-resources-chart"
      ref="chartRef"
      :option="echartsOptions"
      :init-options="{ renderer: 'svg' }"
      autoresize
    />
  </figure>
</template>

<script lang="ts">
import "@/plugins/echarts";
import { defineComponent, type PropType } from "vue";
import VChart from "vue-echarts";
import type { EChartsOption } from "echarts";
import { chartUtil } from "@/utils/chart.util";

type TooltipPositionSize = {
  contentSize: [number, number];
  viewSize: [number, number];
};

const TOOLTIP_GAP = 10;
// The widest resource series uses a 70px bar. Anchor outside its edge
// instead of outside the entire card so the tooltip stays close without
// obscuring any of the overlapping bars.
const WIDEST_BAR_HALF_WIDTH = 35;

export default defineComponent({
  components: { VChart },
  props: {
    overQuotaValue: {
      type: Number as PropType<number>,
      required: true,
    },
    quotaValue: {
      type: Number as PropType<number>,
      required: true,
    },
    allocatedValue: {
      type: Number as PropType<number>,
      required: true,
    },
    showToolTip: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  computed: {
    accessibleDescription(): string {
      return `Project resources. ${this.allocatedValue} allocated, ${this.quotaValue} quota, ${this.overQuotaValue} over quota.`;
    },
    echartsOptions(): EChartsOption {
      const options = chartUtil.getResourcesOptions(this.overQuotaValue, this.quotaValue, this.allocatedValue);

      return {
        ...options,
        tooltip: {
          ...(options.tooltip as object),
          position: (
            _point: [number, number],
            _params: unknown,
            _dom: HTMLElement,
            _rect: unknown,
            size: TooltipPositionSize,
          ): [number, number] => {
            const chart = this.$refs.chartRef as (InstanceType<typeof VChart> & { $el?: HTMLElement }) | undefined;
            const chartElement = chart?.$el;
            const chartBounds = chartElement?.getBoundingClientRect();
            const viewportWidth = typeof window === "undefined" ? Number.POSITIVE_INFINITY : window.innerWidth;
            const viewportHeight = typeof window === "undefined" ? Number.POSITIVE_INFINITY : window.innerHeight;

            if (!chartBounds || chartBounds.width === 0 || chartBounds.height === 0) {
              return [
                size.viewSize[0] / 2 + WIDEST_BAR_HALF_WIDTH + TOOLTIP_GAP,
                Math.max(0, (size.viewSize[1] - size.contentSize[1]) / 2),
              ];
            }

            const scaleX = chartBounds.width / size.viewSize[0];
            const barCenterX = chartBounds.left + (size.viewSize[0] / 2) * scaleX;
            const rightBarEdge = barCenterX + WIDEST_BAR_HALF_WIDTH * scaleX;
            const leftBarEdge = barCenterX - WIDEST_BAR_HALF_WIDTH * scaleX;
            const fitsOnRight = rightBarEdge + TOOLTIP_GAP + size.contentSize[0] <= viewportWidth;
            const viewportX = fitsOnRight ? rightBarEdge + TOOLTIP_GAP : leftBarEdge - size.contentSize[0] - TOOLTIP_GAP;
            const centeredViewportY = chartBounds.top + (chartBounds.height - size.contentSize[1]) / 2;
            const viewportY = Math.max(
              TOOLTIP_GAP,
              Math.min(centeredViewportY, viewportHeight - size.contentSize[1] - TOOLTIP_GAP),
            );

            // ECharts expects chart-local coordinates even when the tooltip is
            // appended to body. Convert the desired viewport position back to
            // that coordinate space, accounting for the card's hover scaling.
            const x = ((viewportX - chartBounds.left) * size.viewSize[0]) / chartBounds.width;
            const y = ((viewportY - chartBounds.top) * size.viewSize[1]) / chartBounds.height;

            return [x, y];
          },
        } as EChartsOption["tooltip"],
      };
    },
  },
  watch: {
    showToolTip(newVal: boolean) {
      const chart = this.$refs.chartRef as InstanceType<typeof VChart> | undefined;
      if (!chart) return;

      if (newVal) {
        chart.chart?.dispatchAction({
          type: "showTip",
          seriesIndex: 0,
          dataIndex: 0,
        });
      } else {
        chart.chart?.dispatchAction({ type: "hideTip" });
      }
    },
  },
});
</script>

<style lang="scss">
.runai-resources-chart {
  height: 105px;
}
</style>
