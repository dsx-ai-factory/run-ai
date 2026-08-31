<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-btn
    class="runai-action-button q-pt-sm runai-icon-btn--default"
    size="sm"
    stack
    flat
    :disable="disable"
    :icon="icon"
    :label="label"
    :aid="aid"
    @click="emit('click')"
  >
    <slot></slot>
  </q-btn>
</template>

<script lang="ts">
export type TBtnAction = { icon: string; label: string };

const btnActionMap = {
  delete: { icon: "ph-fill ph-trash", label: "delete" },
  connect: { icon: "ph-fill ph-arrow-square-out", label: "connect" },
  stop: { icon: "ph-fill ph-stop", label: "stop" },
  activate: { icon: "ph-fill ph-play", label: "activate" },
  run: { icon: "ph-fill ph-play", label: "run" },
  edit: { icon: "ph-fill ph-pencil-simple", label: "edit" },
  review: { icon: "ph-fill ph-file", label: "review" },
  duplicate: { icon: "ph-fill ph-copy", label: "make a copy" },
  showDetails: { icon: "ph-fill ph-file-text", label: "show details" },
  hideDetails: { icon: "ph-fill ph-file-text", label: "hide details" },
  rename: { icon: "ph-fill ph-pencil-simple", label: "rename" },
  assign: {
    // Custom glyph from the runai-glyph SVG sprite (catalog: entity.accessRule).
    icon: "runai-filled-access-rule",
    label: "access rules",
  },
  reset: { icon: "ph-fill ph-lock", label: "reset password" },
  install: { icon: "ph-fill ph-download", label: "get installation instructions" },
  regenerate: { icon: "ph-fill ph-lock", label: "regenerate client secret" },
  remove: { icon: "ph-bold ph-x", label: "remove" },
  viewPolicy: { icon: "ph-fill ph-clipboard-text", label: "view policy" },
  downloadCsv: { icon: "ph-fill ph-arrow-line-down", label: "download csv" },
  update: { icon: "ph-fill ph-pencil-simple", label: "update" },
  topology: { icon: "runai-filled-topology", label: "network topology" },
  moveToRank: { icon: "ph-fill ph-arrows-down-up", label: "move to rank" },
} as const satisfies Record<string, TBtnAction>;

export type TAction = keyof typeof btnActionMap;
export type TBtnActionMap = Record<TAction, TBtnAction>;
</script>

<script setup lang="ts">
import { computed } from "vue";

interface Props {
  btnAction: TAction;
  disable?: boolean;
  aid?: string;
}

const props = withDefaults(defineProps<Props>(), { disable: false });

const emit = defineEmits<{ (e: "click"): void }>();

const label = computed(() => btnActionMap[props.btnAction].label);
const icon = computed(() => btnActionMap[props.btnAction].icon);
</script>

<style lang="scss" scoped>
.runai-action-button {
  margin-right: var(--space-m);

  :deep(.q-btn__label) {
    font-size: 10px;
    font-weight: 500;
    line-height: 14px;
    text-transform: uppercase;
    color: inherit;
  }
}
</style>
