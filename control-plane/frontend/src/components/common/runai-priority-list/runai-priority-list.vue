<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section ref="listRef" class="runai-priority-list column no-wrap" :aria-label="title" :aid="aid">
    <div class="priority-list-title row items-center no-wrap">
      <span class="text-b-subtitle1">{{ title }}</span>
      <runai-tooltip
        v-if="tooltip"
        :tooltip-text="tooltip"
        width="300px"
        :aria-label="`More information about ${title}`"
      />
    </div>
    <div v-if="subtitle" class="priority-list-subtitle row items-center no-wrap">{{ subtitle }}</div>
    <vue-draggable
      :model-value="items"
      target=".priority-list-items"
      :disabled="disableReorder"
      :animation="150"
      filter=".item-arrows"
      :prevent-on-filter="false"
      ghost-class="priority-list-item--ghost"
      @update:model-value="emit('update:items', $event)"
    >
      <q-list class="priority-list-items">
        <q-item
          v-for="(item, index) in items"
          :key="item"
          clickable
          class="priority-list-item"
          active-class="priority-list-item--active"
          :active="item === modelValue"
          :aid="`${aid}-item-${item}`"
          :aria-current="item === modelValue ? 'true' : undefined"
          :aria-describedby="showDisabledReason ? disabledReasonId : undefined"
          @click="emit('update:model-value', item)"
          @mouseenter="showTitleTooltipIfTruncated($event, item)"
          @focusin="showTitleTooltipIfTruncated($event, item)"
          @mouseleave="hoveredTruncatedItem = null"
          @focusout="hoveredTruncatedItem = null"
        >
          <q-item-section class="item-title-wrapper">
            <div class="row items-center no-wrap">
              <q-badge class="item-number" :aid="`${aid}-item-number-${item}`">{{ index + 1 }}</q-badge>
              <div class="item-title ellipsis-2-lines">{{ item }}</div>
            </div>
            <q-tooltip
              :model-value="hoveredTruncatedItem === item"
              no-parent-event
              anchor="bottom middle"
              self="top middle"
              max-width="300px"
            >
              {{ item }}
            </q-tooltip>
          </q-item-section>
          <q-item-section side class="item-arrows">
            <runai-btn
              icon-button
              icon="ph-bold ph-arrow-up"
              size="xs"
              :aid="`${aid}-move-up-${item}`"
              :aria-label="`Move ${item} up`"
              :disable="disableReorder || index === 0"
              :disabled-tooltip="arrowDisabledTooltip"
              @click.stop="moveWithArrow(item, index, index - 1)"
            />
            <runai-btn
              icon-button
              icon="ph-bold ph-arrow-down"
              size="xs"
              :aid="`${aid}-move-down-${item}`"
              :aria-label="`Move ${item} down`"
              :disable="disableReorder || index === items.length - 1"
              :disabled-tooltip="arrowDisabledTooltip"
              @click.stop="moveWithArrow(item, index, index + 1)"
            />
          </q-item-section>
          <span v-if="item === modelValue" class="item-indicator" />
        </q-item>
      </q-list>
    </vue-draggable>
    <span v-if="showDisabledReason" :id="disabledReasonId" class="priority-list-sr-only">
      {{ disableReorderTooltip }}
    </span>
    <span class="priority-list-sr-only" aria-live="polite" aria-atomic="true">{{ moveAnnouncement }}</span>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, useId } from "vue";
import { VueDraggable } from "vue-draggable-plus";
import { RunaiBtn } from "@/components/common/runai-btn";
import { RunaiTooltip } from "@/components/common/runai-tooltip";

export interface Props {
  title: string;
  items: string[];
  modelValue: string;
  subtitle?: string;
  tooltip?: string;
  disableReorder?: boolean;
  disableReorderTooltip?: string;
  aid?: string;
}

const props = withDefaults(defineProps<Props>(), {
  disableReorder: false,
  aid: "runai-priority-list",
});

const emit = defineEmits<{
  "update:model-value": [item: string];
  "update:items": [items: string[]];
}>();

const hoveredTruncatedItem = ref<string | null>(null);
const listRef = ref<HTMLElement | null>(null);
const moveAnnouncement = ref("");
// Disabled buttons can't take focus, so keyboard users get the reason from the items instead of the tooltip.
const disabledReasonId = useId();
const showDisabledReason = computed(() => props.disableReorder && !!props.disableReorderTooltip);
const arrowDisabledTooltip = computed(() => (props.disableReorder ? props.disableReorderTooltip : undefined));

const move = (from: number, to: number): void => {
  if (props.disableReorder || to < 0 || to >= props.items.length) return;
  const reordered = [...props.items];
  const [moved] = reordered.splice(from, 1);
  reordered.splice(to, 0, moved);
  emit("update:items", reordered);
};

const isFocusable = (element: HTMLElement | null | undefined): element is HTMLElement =>
  !!element && !element.hasAttribute("disabled") && element.getAttribute("aria-disabled") !== "true";

// Moving an item to either end disables the arrow that was just pressed, which would drop focus to the page.
const focusMovedItem = (item: string, direction: "up" | "down"): void => {
  const opposite = direction === "up" ? "down" : "up";
  const candidates = [
    `${props.aid}-move-${direction}-${item}`,
    `${props.aid}-move-${opposite}-${item}`,
    `${props.aid}-item-${item}`,
  ].map((aid) => listRef.value?.querySelector<HTMLElement>(`[aid="${aid}"]`));
  candidates.find(isFocusable)?.focus();
};

const moveWithArrow = async (item: string, from: number, to: number): Promise<void> => {
  move(from, to);
  await nextTick();
  focusMovedItem(item, to < from ? "up" : "down");
  moveAnnouncement.value = `${item} moved to position ${to + 1} of ${props.items.length}`;
};

const showTitleTooltipIfTruncated = (event: Event, item: string): void => {
  const title = (event.currentTarget as HTMLElement).querySelector<HTMLElement>(".item-title");
  const isTruncated = !!title && title.scrollHeight > title.clientHeight;
  hoveredTruncatedItem.value = isTruncated ? item : null;
};
</script>

<style scoped lang="scss">
.priority-list-sr-only {
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

.runai-priority-list {
  width: 201px;
  flex: 0 0 201px;
  align-self: stretch;
  padding-bottom: var(--space-m);
  background: var(--bg-primary-default);
  border-right: 1px solid var(--border-subtle);
  // Figma's inner strokes take no layout space, so they're drawn as inset shadows to keep content where Figma puts it.
  box-shadow: inset 1px 0 0 var(--border-subtle);
}

.priority-list-title,
.priority-list-subtitle {
  gap: var(--space-s);
  padding: 0 var(--space-xl) 0 var(--space-m);
  box-shadow: inset 0 -1px 0 var(--border-subtle);
}

.priority-list-title :deep(.q-btn) {
  width: 32px;
  min-width: 32px;
  height: 32px;
  min-height: 32px;
}

.priority-list-title {
  height: 48px;
  color: var(--b-text-main);
}

.priority-list-subtitle {
  height: 34px;
  color: var(--b-text-sub);
  font-size: 12px;
  font-weight: 500;
  line-height: 14px;
  letter-spacing: 0.72px;
  text-transform: uppercase;
}

.priority-list-items {
  padding: 0;
}

.priority-list-item {
  position: relative;
  height: 60px;
  min-height: 60px;
  gap: var(--space-s);
  padding: var(--space-s);
  box-shadow: inset 0 -1px 0 var(--border-subtle);

  &--ghost {
    opacity: 0.4;
  }
}

.item-title-wrapper {
  min-width: 0;

  > .row {
    gap: var(--space-m);
  }
}

.item-number {
  flex-shrink: 0;
  justify-content: center;
  width: 18px;
  height: 18px;
  min-height: 18px;
  padding: 0;
  border-radius: 4px;
  background: var(--b-bg-secondary-default);
  color: var(--text-tertiary);
  font-size: 10px;
  font-weight: 700;
  line-height: 15px;
}

.item-title {
  flex: 1;
  min-width: 0;
  color: var(--b-text-sub);
  font-size: 14px;
  font-weight: 400;
  line-height: 21px;
  word-break: break-word;
}

.priority-list-item--active {
  .item-number {
    background: var(--bg-primary-pressed);
    color: var(--text-brand);
  }

  .item-title {
    color: var(--text-brand);
    font-weight: 500;
    line-height: 20px;
  }
}

.item-indicator {
  position: absolute;
  top: 50%;
  right: 0;
  width: 4px;
  height: 59px;
  transform: translateY(-50%);
  background: var(--interactive-primary-default-bg);
}

.item-arrows.q-item__section--side {
  flex-direction: column;
  justify-content: center;
  gap: var(--space-2xs);
  padding: 0;
  cursor: default;

  :deep(.q-btn) {
    width: 12px;
    min-width: 12px;
    height: 12px;
    min-height: 12px;
    padding: 0;
    color: $black-54;
  }

  :deep(.q-btn.disabled) {
    opacity: 1 !important;
    color: var(--icon-disabled);
  }

  :deep(.q-btn .q-icon) {
    font-size: 12px;
  }

  :deep(.q-btn .q-btn__content) {
    min-height: 12px;
  }
}
</style>
