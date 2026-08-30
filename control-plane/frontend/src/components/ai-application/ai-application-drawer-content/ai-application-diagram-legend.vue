<template>
  <div class="diagram-legend flex items-center bg-white" :class="{ collapsed: isCollapsed }">
    <div v-if="!isCollapsed" class="legend-items flex items-center q-gutter-md">
      <div v-for="item in legendItems" :key="item.kind" class="legend-item flex items-center q-gutter-x-xs">
        <div class="legend-icon flex items-center justify-center" :class="item.iconClass">
          <i v-if="item.phIcon" :class="item.phIcon" />
          <runai-svg-icon v-else-if="item.svgIcon" :name="item.svgIcon" :size="item.svgSize || '24px'" />
        </div>
        <span class="text-caption text-b-main">{{ item.label }}</span>
      </div>
    </div>

    <runai-btn
      aid="ai-application-graph-legend-toggle"
      class="collapse-btn"
      flat
      round
      size="xs"
      :icon="isCollapsed ? 'ph-bold ph-caret-right' : 'ph-bold ph-caret-left'"
      color="black-54"
      @click="toggleCollapsed"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { RunaiBtn } from "@/components/common/runai-btn";
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";
import { EResourceKind } from "./ai-application-diagram.util";

export interface LegendItem {
  kind: EResourceKind;
  label: string;
  phIcon?: string;
  svgIcon?: string;
  svgSize?: string;
  iconClass?: string;
}

const legendItems: LegendItem[] = [
  { kind: EResourceKind.Workload, label: "Workload", phIcon: "ph-bold ph-laptop", iconClass: "workload-icon" },
  { kind: EResourceKind.Secret, label: "Secret", svgIcon: "secret-volume" },
  { kind: EResourceKind.PersistentVolumeClaim, label: "PVC", svgIcon: "pvc" },
  { kind: EResourceKind.ConfigMap, label: "ConfigMap", svgIcon: "config-map" },
  {
    kind: EResourceKind.Endpoint,
    label: "Endpoint",
    svgIcon: "serving-endpoint",
    svgSize: "16px",
    iconClass: "endpoint-icon",
  },
];

const isCollapsed = ref(false);

const toggleCollapsed = (): void => {
  isCollapsed.value = !isCollapsed.value;
};
</script>

<style lang="scss" scoped>
.diagram-legend {
  position: absolute;
  top: 10px;
  left: 10px;
  z-index: 5;
  height: 50px;
  border-radius: 5px;
  padding: 10px 4px 10px 10px;
  box-shadow: -1px 0 7px 0 rgba(0, 0, 0, 0.15);

  &.collapsed {
    padding: 10px;
  }
}

.legend-icon {
  width: 30px;
  height: 30px;
  border-radius: 4px;
  background: $body-background-color;

  i {
    font-size: 14px;
    color: var(--b-text-sub);
  }

  &.workload-icon {
    background: $diagram-workload-icon-bg;
  }

  &.endpoint-icon {
    background: $diagram-endpoint-icon-bg;
  }
}

.legend-item {
  span {
    font-size: 14px;
  }
}

.collapse-btn {
  .diagram-legend:not(.collapsed) & {
    margin-left: 7px;
  }
}
</style>
