<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="docker-registry-box">
    <div class="row full-width q-mb-md">
      <q-input class="col-2 q-mr-md" disable label="Credential type" stack-label :model-value="'Docker registry'" />
      <q-input class="col-4 q-mr-md" disable label="Credential name" stack-label :model-value="credential.meta.name" />
      <q-input class="flex-1" disable label="Username" stack-label :model-value="credential.spec.dockerRegistry?.user" />
    </div>
    <div class="row q-mb-md">
      <q-input class="col-12" disable label="Docker registry URL" :model-value="credential.spec.dockerRegistry?.url" />
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
  name: "docker-registry-box",
  props: {
    credential: {
      type: Object as PropType<CredentialsListResponseEntry>,

      required: true,
    },
  },
});
</script>
