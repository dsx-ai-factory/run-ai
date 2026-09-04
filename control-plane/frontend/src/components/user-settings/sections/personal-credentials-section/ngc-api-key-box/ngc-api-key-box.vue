<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="ngc-api-key-box">
    <div class="row full-width">
      <q-input class="col-2 q-mr-md" disable label="Credential type" stack-label :model-value="'NGC API key'" />
      <q-input class="col-4" disable label="Credential name" stack-label :model-value="credential.meta.name" />
    </div>
    <div class="row">
      <q-input
        class="col-12"
        disable
        label="Description"
        :model-value="credential.meta.description"
        :maxlength="250"
        counter
        autogrow
      />
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { CredentialsListResponseEntry } from "@/swagger-models/assets-service-client";

export default defineComponent({
  name: "ngc-api-key-box",
  props: {
    credential: {
      type: Object as PropType<CredentialsListResponseEntry>,
      required: true,
    },
  },
});
</script>
