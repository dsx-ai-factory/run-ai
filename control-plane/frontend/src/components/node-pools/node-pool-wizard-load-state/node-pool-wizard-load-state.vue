<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <runai-form-wrapper :form-state="{}" full-width>
    <runai-default-loader v-if="loading" />
    <q-banner v-else-if="error" class="bg-red-1 text-negative q-ma-lg" role="alert">
      {{ error }}
      <template #action>
        <runai-btn :aid="`retry-${step}-pool`" label="Retry" flat color="primary" @click="emit('retry')" />
      </template>
    </q-banner>
  </runai-form-wrapper>
</template>
<script setup lang="ts">
import { RunaiBtn } from "@/components/common/runai-btn";
import { RunaiDefaultLoader } from "@/components/common/runai-default-loader";
import RunaiFormWrapper from "@/components/common/runai-form-wrapper/runai-form-wrapper.vue";

defineProps<{ step: "scheduling" | "topology"; loading: boolean; error: string }>();
const emit = defineEmits<{ retry: [] }>();
</script>
