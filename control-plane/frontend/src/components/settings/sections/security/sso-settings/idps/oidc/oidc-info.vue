<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="oidc-info">
    <setting-readable-field label="Redirect URL" default-opened v-if="idp?.redirectUri">
      <div class="row no-wrap items-center">
        <div class="col-9 q-pt-md q-pl-md">
          {{ idp.redirectUri }}
        </div>
      </div>
    </setting-readable-field>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { Idp } from "@/swagger-models/identity-manager-client";
import SettingReadableField from "@/components/settings/setting-readable-field/setting-readable-field.vue";

export default defineComponent({
  name: "oidc-info",
  components: {
    SettingReadableField,
  },
  props: {
    idp: {
      type: Object as PropType<Idp>,
      required: true,
    },
  },
});
</script>
