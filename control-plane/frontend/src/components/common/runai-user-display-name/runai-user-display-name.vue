<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <span ref="wrapperRef" class="runai-user-display-name">
    <q-skeleton v-if="loading" type="text" width="120px" />
    <template v-else>
      <span class="runai-user-display-name__email">{{ display.primaryLabel }}</span>
      <span v-if="display.secondaryLabel" class="runai-user-display-name__name">&nbsp;{{ display.secondaryLabel }}</span>
      <q-tooltip v-if="isTruncated" anchor="top middle" self="bottom middle" :offset="[0, 4]">
        <div>{{ display.primaryLabel }}</div>
        <div v-if="display.secondaryLabel">{{ display.secondaryLabel }}</div>
      </q-tooltip>
      <q-resize-observer @resize="checkTruncation" />
    </template>
  </span>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from "vue";
import { useUserDisplayNameStore } from "@/stores/user-display-name.store";

const props = defineProps<{
  username: string;
}>();

const store = useUserDisplayNameStore();
const wrapperRef = ref<HTMLElement | null>(null);
const isTruncated = ref(false);

// Only "pending" counts as loading. If the lookup failed transiently and the
// store didn't cache (so isResolved stays false), the cell renders plain email
// instead of a stuck skeleton — and the next mount re-queues the lookup.
const loading = computed(() => !!props.username && store.isPending(props.username));
const display = computed(
  () => store.getDisplayLabel(props.username) ?? { label: props.username, primaryLabel: props.username },
);

const checkTruncation = async (): Promise<void> => {
  await nextTick();
  const el = wrapperRef.value;
  if (el) isTruncated.value = el.scrollWidth > el.clientWidth;
};

const resolve = (): void => {
  if (props.username) {
    void store.resolveUsernames([props.username]);
  }
};

onMounted(() => {
  resolve();
  void checkTruncation();
});
watch(
  () => props.username,
  () => {
    resolve();
    void checkTruncation();
  },
);
watch([display, loading], () => void checkTruncation());
</script>

<style lang="scss" scoped>
.runai-user-display-name {
  // Single-line wrapper so the email + name flow as one text run; the browser
  // places a single ellipsis at the overflow point instead of one per span.
  display: block;
  max-width: 100%;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;

  &__name {
    color: var(--b-text-sub);
  }
}
</style>
