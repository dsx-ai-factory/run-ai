<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="new-secret-details">
    <div class="row q-mb-md q-gutter-lg">
      <q-input
        class="col-5"
        :model-value="secretDetails.user"
        no-error-icon
        label="Username"
        :rules="[validationRules.validateUsernameNotEmpty]"
        aid="docker-registry-username-input"
        @update:model-value="$emit('update-secret-details', { ...secretDetails, user: $event })"
      />

      <q-input
        class="col-5"
        :type="isPassHidden ? 'password' : 'text'"
        :model-value="secretDetails.password"
        no-error-icon
        label="Password"
        :rules="[validationRules.validatePasswordNotEmpty]"
        aid="docker-registry-password-input"
        @update:model-value="$emit('update-secret-details', { ...secretDetails, password: $event })"
      >
        <template v-slot:append>
          <q-icon
            size="20px"
            :name="isPassHidden ? 'ph-fill ph-eye-slash' : 'ph-fill ph-eye'"
            class="cursor-pointer text-black-54"
            @click="isPassHidden = !isPassHidden"
          />
        </template>
      </q-input>
    </div>
    <div class="row">
      <q-input
        class="col-11"
        :model-value="secretDetails.url"
        label="Docker registry URL"
        :rules="[validationRules.validateUrlNotEmpty, validationRules.validateDockerServerFormat]"
        aid="docker-registry-url-input"
        @update:model-value="$emit('update-secret-details', { ...secretDetails, url: $event })"
        no-error-icon
      />
    </div>
  </section>
</template>

<script lang="ts">
import { validationRules } from "@/common/validation-rules";
import type { DockerRegistryCreationSpec } from "@/swagger-models/assets-service-client";
import { defineComponent, type PropType } from "vue";

export default defineComponent({
  name: "docker-registry-new-secret-section",
  props: {
    secretDetails: {
      type: Object as PropType<DockerRegistryCreationSpec>,
      required: true,
    },
  },
  data() {
    return {
      isPassHidden: true as boolean,
      validationRules,
    };
  },
});
</script>
