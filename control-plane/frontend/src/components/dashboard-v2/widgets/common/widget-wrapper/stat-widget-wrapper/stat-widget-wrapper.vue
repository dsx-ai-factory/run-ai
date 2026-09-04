<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-card class="stat-box-wrapper">
    <div class="title-tooltip-container">
      <div class="text-b-main w-sm-100 title" :aid="aidForTitle">
        {{ options.title }}
        <runai-tooltip-wrapper use-html v-if="options.tooltipText" :tooltip-text="options.tooltipText">
          <template #tooltip-content>
            <div v-dompurify-html="options.tooltipText" />
          </template>
          <q-icon :name="helpIcon" size="20px" class="text-icon-default"
        /></runai-tooltip-wrapper>
      </div>
    </div>

    <div v-if="!loading && !error" class="stat-data"><slot name="data"></slot></div>

    <div v-if="loading">
      <q-skeleton height="80px" square></q-skeleton>
    </div>
    <div v-else-if="error" class="error-section text-b-sub">Data is not available right now</div>
    <div v-else class="link-and-time-frame-container">
      <widget-link-btn
        v-if="options.linkText"
        :link-text="options.linkText"
        :entity-type="options.entityType"
        @click="$emit('link-clicked')"
        class="q-mr-auto"
        min-height="auto"
      />

      <div v-if="options.timeFrame">
        <q-icon name="ph-bold ph-calendar text-black-30 q-mr-xs" size="16px" class="dash-icon" />
        <span class="text-black-54 time-frame-text">{{ options.timeFrame }}</span>
      </div>
    </div>
  </q-card>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
//model
import type { IWidgetWrapperOptions } from "@/models/chart.model";
//cmps
import { WidgetLinkBtn } from "@/components/dashboard-v2/widgets/common/buttons/widget-link-btn";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { runaiIcons, toQuasarIconName } from "@/plugins/runai-icons.catalog";

export default defineComponent({
  name: "stat-widget-wrapper",
  components: { RunaiTooltipWrapper, WidgetLinkBtn },
  emits: ["link-clicked"],
  props: {
    options: {
      type: Object as PropType<IWidgetWrapperOptions>,
      required: true,
    },
    loading: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    error: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
  },
  data() {
    return { helpIcon: toQuasarIconName(runaiIcons.object.help) };
  },
  computed: {
    aidForTitle(): string {
      return this.options.title.toLowerCase().trim().replace(/ /g, "-");
    },
  },
});
</script>
<style scoped lang="scss">
.stat-box-wrapper {
  width: 100%;
  padding: 16px;
  height: 169px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;

  .title-tooltip-container {
    display: flex;
    align-items: start;
    height: 33px;
    .title {
      font-size: 14px;
    }
  }

  .stat-data {
    font-size: 40px;
    font-weight: 500;
  }

  .time-frame-text {
    font-size: 12px;
  }

  .link-and-time-frame-container {
    display: flex;
    flex-wrap: wrap;
    align-items: end;
    height: 32px;
  }

  .error-section {
    height: 40px;
    color: $black-54;
    margin-top: 13px;
  }
}

@media screen and (min-width: 1400px) {
  .stat-box-wrapper {
    .stat-data {
      font-size: 45px;
    }
  }
}

@media screen and (min-width: 1800px) {
  .stat-box-wrapper {
    .stat-data {
      font-size: 55px;
    }
  }
}
</style>
