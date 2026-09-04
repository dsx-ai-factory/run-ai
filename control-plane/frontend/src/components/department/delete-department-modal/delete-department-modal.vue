<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-delete-modal @delete="$emit('delete')" @cancel="$emit('cancel')" :deleting="isDepartmentDeleting">
    <template #title>
      <p class="delete-modal-msg q-mt-md">Permanently delete department {{ currentDepartment.name }}?</p>
    </template>
    <template #description>Once a department is deleted, it cannot be restored.</template>
  </runai-delete-modal>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
//cmps
import { RunaiDeleteModal } from "@/components/common/runai-delete-modal/";
// models
import type { IDepartmentTable } from "@/models/department.model";

export default defineComponent({
  components: {
    RunaiDeleteModal,
  },
  emits: ["cancel", "delete"],
  props: {
    currentDepartment: {
      type: Object as PropType<IDepartmentTable>,
      required: true,
    },
    isDepartmentDeleting: {
      type: Boolean,
      required: true,
    },
  },
});
</script>

<style>
.delete-modal-msg {
  font-size: 16px;
}
</style>
