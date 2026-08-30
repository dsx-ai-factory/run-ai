<template>
  <div class="resource-card" :class="cardClass">
    <handle
      v-if="type === EResourceCardType.Workload"
      :id="WORKLOAD_TOP_HANDLE_ID"
      type="source"
      :position="Position.Top"
      class="connection-handle"
    />
    <handle
      v-if="type === EResourceCardType.Workload"
      :id="WORKLOAD_BOTTOM_HANDLE_ID"
      type="source"
      :position="Position.Bottom"
      class="connection-handle"
    />
    <handle
      v-if="type !== EResourceCardType.Workload"
      :id="RESOURCE_TOP_HANDLE_ID"
      type="target"
      :position="Position.Top"
      class="connection-handle"
    />
    <div class="flex items-start gap-5">
      <div class="card-icon-box flex items-center justify-center">
        <runai-svg-icon
          v-if="RESOURCE_ICON_CONFIG[type].svgIcon"
          :name="RESOURCE_ICON_CONFIG[type].svgIcon!"
          size="24px"
        />
        <i v-else-if="RESOURCE_ICON_CONFIG[type].phIcon" :class="RESOURCE_ICON_CONFIG[type].phIcon!"></i>
      </div>
      <div class="flex column gap-4 flex-1 min-width-0">
        <span class="card-name text-weight-medium">
          <runai-highlighted-text :text="name" :search-term="searchTerm" />
        </span>
        <div v-if="phase" class="flex items-center justify-between">
          <workload-status-dot :status="phase" />
          <i class="ph-bold ph-caret-down expand-chevron" :class="{ expanded }" aria-hidden="true" />
        </div>
      </div>
    </div>
    <slot name="details"></slot>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { Handle, Position } from "@vue-flow/core";
import { RunaiHighlightedText } from "@/components/common/runai-highlighted-text";
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";
import { WorkloadStatusDot } from "@/components/common/workload-status-dot";
import {
  EResourceCardType,
  RESOURCE_TOP_HANDLE_ID,
  WORKLOAD_BOTTOM_HANDLE_ID,
  WORKLOAD_TOP_HANDLE_ID,
} from "./ai-application-diagram.util";

const RESOURCE_ICON_CONFIG: Record<EResourceCardType, { svgIcon: string | null; phIcon: string | null }> = {
  [EResourceCardType.Workload]: { svgIcon: null, phIcon: "ph-bold ph-laptop" },
  [EResourceCardType.Secret]: { svgIcon: "secret-volume", phIcon: null },
  [EResourceCardType.Pvc]: { svgIcon: "pvc", phIcon: null },
  [EResourceCardType.ConfigMap]: { svgIcon: "config-map", phIcon: null },
  [EResourceCardType.Endpoint]: { svgIcon: "serving-endpoint", phIcon: null },
};

export interface Props {
  name: string;
  type: EResourceCardType;
  phase?: string | null;
  expanded?: boolean;
  searchTerm?: string;
}

const props = withDefaults(defineProps<Props>(), {
  phase: null,
  expanded: false,
  searchTerm: "",
});

const cardClass = computed(() => ({
  [`${props.type}-card`]: true,
  expanded: props.expanded,
  pointer: props.type === EResourceCardType.Workload,
}));
</script>

<style lang="scss" scoped>
.resource-card {
  position: relative;
  width: 230px;
  padding: 10px 7px;
  border-radius: 4px;
  border: 1px solid $black-12;
  background: $white;
  color: $black-87;

  &.workload-card:hover {
    border-color: $black-default-chart;
  }
}

:deep(.connection-handle) {
  opacity: 0;
  pointer-events: none;
}

.card-icon-box {
  width: 24px;
  height: 24px;
  border-radius: 4px;
  flex-shrink: 0;
  background: $body-background-color;

  i {
    font-size: 12px;
    color: $black-54;
  }

  .workload-card & {
    background: $diagram-workload-icon-bg;

    i {
      color: $black-87;
    }
  }
}

.min-width-0 {
  min-width: 0;
}

.card-name {
  font-size: 14px;
}

.expand-chevron {
  font-size: 16px;
  color: $black-54;
  transition: transform 0.2s ease;

  &.expanded {
    transform: rotate(180deg);
  }
}
</style>
