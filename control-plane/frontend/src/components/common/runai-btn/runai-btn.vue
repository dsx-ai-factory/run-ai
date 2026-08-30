<template>
  <q-btn
    class="runai-btn"
    :class="{ 'runai-btn--icon': iconButton }"
    :round="iconButton"
    :flat="iconButton || undefined"
    :aid="aid"
    v-bind="$attrs"
    @click.stop
    :disable="disabled"
  >
    <slot></slot>
    <q-tooltip
      :anchor="tooltipAnchor"
      :self="tooltipSelf"
      :offset="tooltipOffset"
      style="white-space: nowrap"
      v-if="showInfoTooltip"
    >
      {{ tooltip }}
    </q-tooltip>
  </q-btn>
  <q-tooltip :target="`.runai-btn[aid='${aid}']`" v-if="$attrs.disable && disabledTooltip">
    {{ disabledTooltip }}
  </q-tooltip>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// Full Quasar QTooltip anchor/self positions (vertical + horizontal pairs).
type QTooltipAnchor =
  | "top left"
  | "top middle"
  | "top right"
  | "top start"
  | "top end"
  | "center left"
  | "center middle"
  | "center right"
  | "center start"
  | "center end"
  | "bottom left"
  | "bottom middle"
  | "bottom right"
  | "bottom start"
  | "bottom end";

export default defineComponent({
  name: "runai-btn",
  props: {
    aid: {
      type: String,
      required: true,
    },
    // Renders a round, flat, icon-only button with the default icon color
    // (replaces the former runai-icon-button). Pair with an `icon` attr.
    iconButton: {
      type: Boolean,
      default: false,
    },
    disabledTooltip: {
      type: String as PropType<string>,
      required: false,
    },
    tooltip: {
      type: String as PropType<string>,
      required: false,
    },
    tooltipAnchor: {
      type: String as PropType<QTooltipAnchor>,
      required: false,
    },
    tooltipSelf: {
      type: String as PropType<QTooltipAnchor>,
      required: false,
    },
    tooltipOffset: {
      type: Array as PropType<number[]>,
      required: false,
    },
  },
  computed: {
    disabled(): boolean {
      return !!this.$attrs.disable || !!this.$attrs.loading;
    },
    showInfoTooltip(): boolean {
      // Showing the information tooltip only if the button is not disabled and got no disabled tooltip
      // to avoid multiple tooltips at the same time
      return !!this.tooltip && !this.$attrs.disable && !this.disabledTooltip;
    },
  },
});
</script>

<style lang="scss" scoped>
.runai-btn--icon {
  color: var(--icon-default);
}
</style>
