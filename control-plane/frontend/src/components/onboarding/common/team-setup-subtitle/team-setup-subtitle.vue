<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <span class="team-setup-subtitle">
    Give your team a name and add members. We'll create a
    <a :href="DOCS.PLATFORM_MANAGEMENT.PROJECTS" target="_blank">project</a> for the team, set
    <a :href="DOCS.AUTHENTICATION.ACCESS_RULES" target="_blank">permissions</a>, and send out invitations.
  </span>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import DOCS from "@/constants/docs";

export default defineComponent({
  name: "team-setup-subtitle",
  data() {
    return {
      DOCS,
    };
  },
});
</script>
