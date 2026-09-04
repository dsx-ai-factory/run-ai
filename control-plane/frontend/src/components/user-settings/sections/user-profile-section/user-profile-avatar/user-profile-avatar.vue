<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="row items-center q-mb-md user-profile-avatar" aid="user-profile-avatar">
    <div class="col-2">
      <runai-avatar size="large" :initial="initials" />
    </div>
    <div class="col-10">
      <div class="row email">{{ email }}</div>
      <div v-if="fullName" class="row full-name">{{ fullName }}</div>
    </div>
  </div>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiAvatar } from "@/components/common/runai-avatar";

export default defineComponent({
  name: "user-profile-avatar",
  components: {
    RunaiAvatar,
  },
  props: {
    email: {
      type: String as PropType<string>,
      required: true,
    },
    firstName: {
      type: String as PropType<string>,
      default: "",
    },
    lastName: {
      type: String as PropType<string>,
      default: "",
    },
  },
  computed: {
    initials(): string {
      return (
        `${this.firstName?.[0] ?? ""}${this.lastName?.[0] ?? ""}`.toUpperCase() || this.email?.[0]?.toUpperCase() || ""
      );
    },
    fullName(): string {
      return [this.firstName, this.lastName].filter(Boolean).join(" ");
    },
  },
});
</script>

<style scoped lang="scss">
.user-profile-avatar {
  .email {
    font-size: 16px;
  }
  .full-name {
    font-size: 14px;
    color: var(--b-text-sub);
  }
}
</style>
