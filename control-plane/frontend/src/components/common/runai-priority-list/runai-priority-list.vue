<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="runai-priority-list column no-wrap" :aria-label="title" :aid="aid">
    <div class="priority-list-title row items-center no-wrap">
      <span class="text-b-subtitle1">{{ title }}</span>
      <runai-tooltip
        v-if="tooltip"
        :tooltip-text="tooltip"
        width="300px"
        :aria-label="`More information about ${title}`"
      />
    </div>
    <div v-if="subtitle" class="priority-list-subtitle text-b-overline-caps text-b-sub row items-center no-wrap">
      {{ subtitle }}
    </div>
    <vue-draggable
      :model-value="items"
      target=".priority-list-items"
      :disabled="disableReorder"
      :animation="150"
      force-fallback
      filter=".item-arrows"
      :prevent-on-filter="false"
      ghost-class="priority-list-item--ghost"
      @start="itemsAtDragStart = [...items]"
      @end="announceDraggedItem"
      @update:model-value="emit('update:items', $event)"
    >
      <q-list class="priority-list-items" role="list">
        <q-item
          v-for="(item, index) in items"
          :key="item"
          :ref="(row) => setRowRef(item, row)"
          role="listitem"
          class="priority-list-item"
          :class="{ 'priority-list-item--active': item === modelValue }"
          :aid="`${aid}-item-${item}`"
          @mouseenter="showTitleTooltipIfTruncated($event, item)"
          @focusin="showTitleTooltipIfTruncated($event, item)"
          @mouseleave="hoveredTruncatedItem = null"
          @focusout="hoveredTruncatedItem = null"
        >
          <button
            type="button"
            class="item-select"
            :aid="`${aid}-select-${item}`"
            :aria-current="item === modelValue ? 'true' : undefined"
            :aria-describedby="showDisabledReason ? disabledReasonId : undefined"
            @click="emit('update:model-value', item)"
          >
            <q-badge class="item-number" :aid="`${aid}-item-number-${item}`">{{ index + 1 }}</q-badge>
            <span
              class="item-title ellipsis-2-lines"
              :class="item === modelValue ? 'text-b-subtitle1' : 'text-b-subtitle2 text-b-sub'"
            >
              {{ item }}
            </span>
          </button>
          <q-tooltip
            :model-value="hoveredTruncatedItem === item"
            no-parent-event
            anchor="bottom middle"
            self="top middle"
            max-width="300px"
          >
            {{ item }}
          </q-tooltip>
          <div class="item-arrows column no-wrap">
            <runai-btn
              icon-button
              :icon="moveUpIcon"
              :aid="`${aid}-move-up-${item}`"
              :aria-label="`Move ${item} up`"
              :disable="disableReorder || index === 0"
              :disabled-tooltip="arrowDisabledTooltip"
              @click="moveWithArrow(item, index, index - 1)"
            />
            <runai-btn
              icon-button
              :icon="moveDownIcon"
              :aid="`${aid}-move-down-${item}`"
              :aria-label="`Move ${item} down`"
              :disable="disableReorder || index === items.length - 1"
              :disabled-tooltip="arrowDisabledTooltip"
              @click="moveWithArrow(item, index, index + 1)"
            />
          </div>
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
import { computed, nextTick, ref, useId, type ComponentPublicInstance } from "vue";
import { VueDraggable, type DraggableEvent } from "vue-draggable-plus";
import { RunaiBtn } from "@/components/common/runai-btn";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { runaiIcons, toQuasarIconName } from "@/plugins/runai-icons.catalog";

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

const moveUpIcon = toQuasarIconName(runaiIcons.action.sortUp);
const moveDownIcon = toQuasarIconName(runaiIcons.action.sortDown);

const hoveredTruncatedItem = ref<string | null>(null);
const moveAnnouncement = ref("");
const itemsAtDragStart = ref<string[]>([]);
const rowRefs = new Map<string, HTMLElement>();
// Disabled buttons can't take focus, so keyboard users get the reason from the items instead of the tooltip.
const disabledReasonId = useId();
const showDisabledReason = computed(() => props.disableReorder && !!props.disableReorderTooltip);
const arrowDisabledTooltip = computed(() => (props.disableReorder ? props.disableReorderTooltip : undefined));

const setRowRef = (item: string, row: Element | ComponentPublicInstance | null): void => {
  const element = row && "$el" in row ? row.$el : row;
  if (element instanceof HTMLElement) rowRefs.set(item, element);
  else rowRefs.delete(item);
};

const announceMove = (item: string, to: number): void => {
  moveAnnouncement.value = `${item} moved to position ${to + 1} of ${props.items.length}`;
};

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
  const row = rowRefs.get(item);
  if (!row) return;
  const [moveUp, moveDown] = Array.from(row.querySelectorAll<HTMLElement>(".item-arrows button"));
  const pressed = direction === "up" ? moveUp : moveDown;
  const opposite = direction === "up" ? moveDown : moveUp;
  [pressed, opposite, row.querySelector<HTMLElement>(".item-select")].find(isFocusable)?.focus();
};

const moveWithArrow = async (item: string, from: number, to: number): Promise<void> => {
  move(from, to);
  await nextTick();
  focusMovedItem(item, to < from ? "up" : "down");
  announceMove(item, to);
};

const announceDraggedItem = (event: DraggableEvent): void => {
  const { oldIndex, newIndex } = event;
  if (oldIndex === undefined || newIndex === undefined || oldIndex === newIndex) return;
  const item = itemsAtDragStart.value.at(oldIndex);
  if (item) announceMove(item, newIndex);
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
}

.priority-list-subtitle {
  height: 34px;
}

.priority-list-items {
  padding: 0;
}

.priority-list-item {
  position: relative;
  align-items: center;
  height: 60px;
  min-height: 60px;
  gap: var(--space-s);
  padding: var(--space-s);
  box-shadow: inset 0 -1px 0 var(--border-subtle);

  &:hover {
    background: var(--bg-primary-hover);
  }

  &--ghost {
    opacity: 0.4;
  }
}

.item-select {
  display: flex;
  flex: 1;
  align-items: center;
  align-self: stretch;
  gap: var(--space-m);
  min-width: 0;
  margin: 0;
  padding: 0;
  border: 0;
  background: none;
  color: inherit;
  font: inherit;
  text-align: left;
  cursor: pointer;

  &:focus-visible {
    outline: 2px solid var(--border-strong);
    outline-offset: 2px;
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
  word-break: break-word;

  // Figma's text-b-subtitle2 uses a 21px line height; the app's shared class still uses 19.5px.
  &.text-b-subtitle2 {
    line-height: 21px;
  }
}

.priority-list-item--active {
  .item-number {
    background: var(--bg-primary-pressed);
    color: var(--text-brand);
  }

  .item-title {
    color: var(--text-brand);
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

.item-arrows {
  flex-shrink: 0;
  gap: var(--space-2xs);

  :deep(.q-btn) {
    position: relative;
    width: 12px;
    min-width: 12px;
    height: 12px;
    min-height: 12px;
    padding: 0;
    color: var(--icon-default);

    // The arrows keep Figma's 12px size; this invisible area gives each one a larger click target without moving them.
    &::after {
      content: "";
      position: absolute;
      inset: -6px -6px -2px;
    }

    .q-icon {
      font-size: 12px;
    }

    .q-btn__content {
      min-height: 12px;
    }
  }

  :deep(.q-btn:last-child::after) {
    inset: -2px -6px -6px;
  }

  :deep(.q-btn.disabled) {
    opacity: 1 !important;
    color: var(--icon-disabled);
  }
}
</style>
