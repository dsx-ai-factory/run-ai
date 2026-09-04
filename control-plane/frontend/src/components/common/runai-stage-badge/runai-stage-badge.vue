<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-badge
    :aid="badgeAid"
    class="badge"
    outline
    align="middle"
    :class="[badgeColor, { 'badge-default-cursor': useDefaultCursor }]"
    :style="badgeStyle"
  >
    {{ badge }}
    <q-tooltip v-if="displayTooltip" max-width="300px" anchor="bottom end">{{ badgeTooltip }}</q-tooltip>
  </q-badge>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";
import { EBadgeLabel } from "@/models/setting.model";

export interface BadgeConfig {
  color?: string;
  tooltip?: string;
  padding?: string;
}

const badgeConfig: Record<EBadgeLabel, BadgeConfig> = {
  [EBadgeLabel.Experimental]: {
    color: "lighter-green",
    tooltip:
      "This feature is in early development. It may not be stable and could be removed or changed significantly in future versions. Use with caution.",
    padding: "1px 10px",
  },
  [EBadgeLabel.Beta]: {
    color: "runai-orange",
    tooltip:
      "This feature is still being developed for official release in a future version and may have some limitations. Use with caution.",
    padding: "1px 30px",
  },
  [EBadgeLabel.Legacy]: {
    color: "text-black-54",
    tooltip:
      "This feature is scheduled to be removed in future versions. We recommend using alternatives if available. Use only if necessary.",
    padding: "1px 24px",
  },
  [EBadgeLabel.New]: {
    color: "text-secondary",
    tooltip: "This is a new feature.",
  },
};

export default defineComponent({
  name: "runai-stage-badge",
  props: {
    badge: {
      type: String as PropType<EBadgeLabel>,
      required: true,
    },
    displayTooltip: {
      type: Boolean,
      default: false,
    },
    color: {
      type: String,
      required: false,
    },
    badgeWidth: {
      type: String as PropType<string>,
      default: "90px",
    },
    useDefaultCursor: {
      type: Boolean,
      default: false,
    },
  },
  computed: {
    badgeAid(): string {
      return `badge-${this.badge}`;
    },
    badgeStyle(): { padding: string; width: string; color?: string } {
      const config = badgeConfig[this.badge] || {};
      return { padding: config.padding || "", ...(this.color ? { color: this.color } : {}), width: this.badgeWidth };
    },
    badgeColor(): string {
      if (this.color) return "";
      return badgeConfig[this.badge]?.color || "";
    },
    badgeTooltip(): string {
      return badgeConfig[this.badge]?.tooltip || "";
    },
  },
});
</script>

<style scoped lang="scss">
.badge {
  font-size: 12px;
  cursor: pointer;
}
.badge-default-cursor {
  cursor: default;
}
.lighter-green {
  color: $lighter-green !important;
}
.runai-orange {
  color: $runai-orange !important;
}
</style>
