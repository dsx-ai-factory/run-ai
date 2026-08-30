<template>
  <div
    class="runai-segmented-button"
    :class="{ 'runai-segmented-button--spread': spread }"
    role="radiogroup"
    :aria-label="ariaLabel"
    :aid="aid"
  >
    <button
      v-for="(option, index) in options"
      :key="String(option.value)"
      ref="segmentRefs"
      type="button"
      role="radio"
      class="runai-segment text-b-subtitle2"
      :class="{
        'runai-segment--selected': isSelected(option),
        'runai-segment--disabled': isDisabled(option),
        'runai-segment--with-description': !!option.description,
      }"
      :aria-checked="isSelected(option)"
      :aria-disabled="isDisabled(option)"
      :disabled="isDisabled(option)"
      :tabindex="tabindexFor(option, index)"
      :aid="optionAid(option)"
      @click="select(option)"
      @keydown="onKeydown($event, index)"
    >
      <runai-svg-icon v-if="option.icon" class="runai-segment__icon" :name="option.icon" size="20" />
      <span class="runai-segment__text">
        <span class="runai-segment__label">{{ option.label }}</span>
        <span v-if="option.description" class="runai-segment__description text-b-body1">{{ option.description }}</span>
      </span>
      <q-tooltip v-if="isDisabled(option) && option.disabledTooltip" :max-width="option.disabledTooltipWidth">
        {{ option.disabledTooltip }}
      </q-tooltip>
      <q-tooltip v-else-if="option.toolTip" :max-width="option.toolTipWidth ?? tooltipWidth">
        {{ option.toolTip }}
      </q-tooltip>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";
import type { ISelectOption } from "@/models/global.model";

export interface ISegmentedButtonOption extends ISelectOption {
  description?: string;
}

type TSegmentValue = string | number | boolean | null;

export interface Props {
  modelValue: TSegmentValue;
  options: ISegmentedButtonOption[];
  spread?: boolean;
  disable?: boolean;
  ariaLabel?: string;
  tooltipWidth?: string;
  aid?: string;
}

const props = withDefaults(defineProps<Props>(), {
  spread: false,
  disable: false,
  ariaLabel: "Options",
  tooltipWidth: undefined,
  aid: undefined,
});

const emit = defineEmits<{
  (event: "update:model-value", value: TSegmentValue): void;
}>();

const segmentRefs = ref<HTMLButtonElement[]>([]);

function isSelected(option: ISegmentedButtonOption): boolean {
  return option.value === props.modelValue;
}
function isDisabled(option: ISegmentedButtonOption): boolean {
  return props.disable || !!option.disable;
}
function optionAid(option: ISegmentedButtonOption): string {
  return option.aid ?? `segment-${option.label}`;
}

// Roving tabindex: only the selected (or first enabled) segment is tabbable; arrows move between them.
const tabStopIndex = computed<number>(() => {
  const selected = props.options.findIndex((option) => isSelected(option) && !isDisabled(option));
  if (selected !== -1) return selected;
  return props.options.findIndex((option) => !isDisabled(option));
});
function tabindexFor(option: ISegmentedButtonOption, index: number): number {
  if (isDisabled(option)) return -1;
  return index === tabStopIndex.value ? 0 : -1;
}

function select(option: ISegmentedButtonOption): void {
  if (isDisabled(option) || isSelected(option)) return;
  emit("update:model-value", option.value ?? null);
}

function nextEnabledIndex(start: number, step: number): number {
  const count = props.options.length;
  for (let i = 1; i <= count; i++) {
    const index = (start + step * i + count * Math.abs(step) * count) % count;
    if (!isDisabled(props.options[index])) return index;
  }
  return start;
}
function focusAndSelect(index: number): void {
  const option = props.options[index];
  if (!option) return;
  segmentRefs.value[index]?.focus();
  select(option);
}
function onKeydown(event: KeyboardEvent, index: number): void {
  switch (event.key) {
    case "ArrowRight":
    case "ArrowDown":
      event.preventDefault();
      focusAndSelect(nextEnabledIndex(index, 1));
      break;
    case "ArrowLeft":
    case "ArrowUp":
      event.preventDefault();
      focusAndSelect(nextEnabledIndex(index, -1));
      break;
    case "Home":
      event.preventDefault();
      focusAndSelect(nextEnabledIndex(-1, 1));
      break;
    case "End":
      event.preventDefault();
      focusAndSelect(nextEnabledIndex(0, -1));
      break;
  }
}
</script>

<style lang="scss" scoped>
.runai-segmented-button {
  display: inline-flex;
  align-items: center;
  gap: var(--space-3xs, 3px);
  padding: var(--space-3xs, 3px);
  background: var(--bg-primary-default);
  border: 1px solid var(--border-subtle);
  border-radius: var(--b-border-radius-radius-m, 10px);
  width: fit-content;

  &--spread {
    display: flex;
    width: 100%;

    .runai-segment {
      flex: 1 1 0;
    }
  }
}

.runai-segment {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-xs, 7px);
  height: 40px;
  padding: 0 var(--space-s, 10px);
  // Transparent border reserves the selected border's space so selecting a segment never shifts the row.
  border: 2px solid transparent;
  border-radius: var(--b-border-radius-radius-m, 10px);
  background: transparent;
  color: var(--b-text-main);
  cursor: pointer;
  white-space: nowrap;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease,
    color 0.15s ease;

  &:hover:not(.runai-segment--selected):not(.runai-segment--disabled) {
    background: var(--bg-primary-hover);
  }

  &:focus-visible:not(.runai-segment--disabled) {
    outline: none;
    border-color: var(--border-strong);
  }

  &--selected {
    background: var(--bg-primary-default);
    border-color: var(--border-pressed);
    color: var(--interactive-secondary-default-text);
    box-shadow: -1px 0 7px 0 rgba(0, 0, 0, 0.15);

    &:hover:not(.runai-segment--disabled) {
      background: var(--interactive-secondary-hover-bg);
    }
    &:focus-visible:not(.runai-segment--disabled) {
      background: var(--bg-primary-hover);
      border-color: var(--border-hover);
    }
  }

  &--disabled {
    cursor: not-allowed;
    color: var(--text-disabled);

    &.runai-segment--selected {
      border-color: var(--border-default);
    }
  }

  &--with-description {
    height: auto;
    min-height: 60px;
    padding: 0 var(--space-m, 16px);
    justify-content: flex-start;
  }

  &__icon {
    flex: 0 0 auto;
  }

  &__text {
    display: flex;
    flex-direction: column;
    align-items: inherit;
    min-width: 0;
  }

  &__description {
    color: var(--text-sub);
    white-space: normal;
  }
}
</style>
