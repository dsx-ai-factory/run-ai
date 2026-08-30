<template>
  <q-card
    :class="['chart-widget-wrapper', sizeClass, isGrid ? 'grid-widget-container' : '']"
    :bordered="!hideBorder"
    :flat="flat"
  >
    <div v-if="!$slots.header" class="chart-widget-header row items-center q-py-xs q-pl-md">
      <div class="title-tooltip-container">
        <h3 class="chart-widget-title text-b-main" :id="titleId" :aid="aidForTitle">
          {{ options.title }}
        </h3>
        <runai-tooltip-wrapper use-html v-if="options.tooltipText" :tooltip-text="options.tooltipText" class="q-ml-xs">
          <template #tooltip-content>
            <div v-dompurify-html="options.tooltipText" />
          </template>
          <q-icon class="text-icon-default" :name="helpIcon" size="20px"
        /></runai-tooltip-wrapper>
      </div>

      <widget-link-btn
        class="q-mr-md"
        v-if="options.linkText"
        :link-text="options.linkText"
        :entity-type="options.entityType"
        @click="$emit('link-clicked')"
      />

      <div v-if="options.timeFrame && !hideTimeFrame" class="timeframe-container text-black-54 q-mr-xs flex">
        <q-icon name="ph-bold ph-calendar q-mr-sm items-center text-black-30" size="16px" class="dash-icon" />
        <span class="text-black-54 time-frame-text">{{ options.timeFrame }}</span>
      </div>
      <runai-actions-menu v-if="!hideActions" round hide-label large-icon :action-items="actionMenuItems" />
    </div>
    <template v-else>
      <slot name="header" :title-id="titleId" />
    </template>

    <q-separator v-if="!hideSeparator" />
    <q-card
      class="chart-semantic-region col flex-1"
      :class="{ 'no-padding-top-and-bottom': bodyNoPadding }"
      role="region"
      aria-roledescription="chart"
      :tabindex="isGrid ? undefined : 0"
      :aria-labelledby="titleId"
      :aria-describedby="`${descriptionId} ${summaryId}`"
    >
      <p :id="descriptionId" class="visually-hidden">{{ accessibleDescriptionText }}</p>
      <p :id="summaryId" class="visually-hidden">{{ accessibleSummaryText }}</p>
      <div class="visually-hidden" aria-live="polite" aria-atomic="true">{{ announcement }}</div>
      <div class="loading-section q-px-sm row items-center q-py-md" v-if="loading">
        <q-skeleton class="chart-widget-skeleton" :height="skeletonHeight" width="100%" square></q-skeleton>
      </div>
      <div v-else-if="error" class="centered-container error-section text-b-sub" aid="error-widget-message">
        Data is not available right now
      </div>
      <div v-else-if="empty" class="centered-container empty-section" aid="empty-widget-message">
        {{ emptyMessage }}
      </div>
      <template v-else>
        <div class="breadcrumbs-container" v-if="breadcrumbs">
          <transition name="fade">
            <chart-breadcrumbs :breadcrumbs="breadcrumbs" @breadcrumb-clicked="$emit('breadcrumb-clicked', $event)" />
          </transition>
        </div>
        <div :class="slotWrapperClass" style="width: 100%">
          <slot></slot>
        </div>
      </template>
    </q-card>
  </q-card>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
//model
import type { IWidgetWrapperOptions } from "@/models/chart.model";
import { EWidgetEmptyMessages } from "@/models/chart.model";
//cmps
import { RunaiActionsMenu } from "@/components/common/runai-page-actions/runai-actions-menu";
import { ChartBreadcrumbs } from "@/components/dashboard-v2/widgets/common/chart-breadcrumbs";
import { WidgetLinkBtn } from "@/components/dashboard-v2/widgets/common/buttons/widget-link-btn";
import { stringUtil } from "@/utils/string.util";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { runaiIcons, toQuasarIconName } from "@/plugins/runai-icons.catalog";

export default defineComponent({
  name: "chart-widget-wrapper",
  components: { RunaiTooltipWrapper, WidgetLinkBtn, ChartBreadcrumbs, RunaiActionsMenu },
  emits: ["export-csv", "link-clicked", "breadcrumb-clicked"],
  props: {
    options: {
      type: Object as PropType<IWidgetWrapperOptions>,
      required: true,
    },
    breadcrumbs: {
      type: Array as PropType<string[]>,
      required: false,
    },
    bodyNoPadding: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    hideActions: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hideBorder: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    flat: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hideTimeFrame: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    loading: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    error: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    empty: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    hideSeparator: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    slotWrapperClass: {
      type: String as PropType<string>,
      default: "q-pt-md",
    },
    size: {
      type: String as PropType<"sm" | "medium" | "large" | "auto">,
      default: "medium",
    },
    isGrid: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    skeletonHeight: {
      type: String as PropType<string>,
      default: "230px",
    },
    accessibleDescription: {
      type: String as PropType<string>,
      default: "",
    },
    accessibleSummary: {
      type: String as PropType<string>,
      default: "",
    },
    announcement: {
      type: String as PropType<string>,
      default: "",
    },
  },
  data() {
    return {
      helpIcon: toQuasarIconName(runaiIcons.object.help),
      actionMenuItems: [
        {
          label: "Download as CSV",
          action: () => this.$emit("export-csv"),
        },
      ],
    };
  },
  computed: {
    aidForTitle(): string {
      return stringUtil.slug(this.options.title);
    },
    titleId(): string {
      return `chart-card-title-${this.aidForTitle}-${this.$.uid}`;
    },
    descriptionId(): string {
      return `${this.titleId}-description`;
    },
    summaryId(): string {
      return `${this.titleId}-summary`;
    },
    accessibleDescriptionText(): string {
      if (this.accessibleDescription) return this.accessibleDescription;
      return `${this.options.title} chart${this.options.timeFrame ? ` for ${this.options.timeFrame}` : ""}.`;
    },
    accessibleSummaryText(): string {
      if (this.accessibleSummary) return this.accessibleSummary;
      if (this.loading) return `${this.options.title} is loading.`;
      if (this.error) return `${this.options.title} data is not available right now.`;
      if (this.empty) return `${this.options.title} has no data to display.`;
      return `Shows ${this.options.title.toLowerCase()}${this.options.timeFrame ? ` for ${this.options.timeFrame}` : ""}.`;
    },
    sizeClass(): string {
      return `size-${this.size}`;
    },
    emptyMessage(): string {
      return this.options.emptyMessage || EWidgetEmptyMessages.Default;
    },
  },
});
</script>

<style scoped lang="scss">
.chart-widget-wrapper {
  display: flex;
  flex-direction: column;
  position: relative;
  container-type: inline-size;

  &.size-sm {
    height: 200px;
  }

  &.size-medium {
    height: 320px;
  }

  &.size-large {
    height: 350px;
  }
  &.size-auto {
    height: auto;
  }
}
.chart-widget-header {
  min-height: 50px;
}
.title-tooltip-container {
  display: flex;
  align-items: center;
  flex-wrap: nowrap;
  margin-right: auto;
}

.chart-widget-title {
  margin: 0;
  font: inherit;
  letter-spacing: inherit;
}

.chart-semantic-region:focus-visible {
  outline: 2px solid var(--border-strong, #707070);
  outline-offset: -2px;
}

@container (max-width: 500px) {
  .chart-widget-header {
    align-items: flex-start;
    padding-top: 10px;
    padding-bottom: 10px;
  }

  .title-tooltip-container {
    min-width: 0;
  }

  .title-tooltip-container .text-b-main {
    white-space: normal;
  }
}

.breadcrumbs-container {
  position: absolute;
  top: 12px;
  z-index: 2;
}

.centered-container {
  display: flex;
  height: 100%;
  width: 100%;
  justify-content: center;
  align-items: center;
}

.time-frame-text {
  font-size: 12px;
}

.loading-section {
  height: 100%;
  width: 100%;
}

.no-padding-top-and-bottom {
  padding: 0 16px;
  display: flex;
  align-items: center;
}

.timeframe-container:last-child {
  margin-right: 20px;
}

.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
</style>

<style lang="scss">
.chart-widget-heading--visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.grid-widget-container {
  .q-table__bottom--nodata {
    min-height: 215px;
  }

  .q-table tbody td {
    font-size: 12px;
  }

  .q-table tbody td {
    height: 45px;
  }
  .q-table thead th {
    text-wrap: wrap;
  }
  .q-table thead tr {
    height: 41px;
  }
  .q-td {
    padding: 0 16px;
  }
  .q-table__sort-icon {
    position: absolute;
  }
}
</style>
