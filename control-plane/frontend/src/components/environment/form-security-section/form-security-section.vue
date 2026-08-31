<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item label="Security" class="form-security-section" aid="form-security-section">
    <template #subheader>{{ summary }} </template>
    <form-security-settings
      @update="$emit('update', $event)"
      :is-s-s-o="isSSO"
      :security-model="securityModel"
      show-override-toggle
    />
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { FormSecuritySettings } from "@/components/environment/form-security-section/form-security-settings";
// Models
import { UID_GID_SOURCE_LABELS, type ISecurityModel } from "@/models/environment.model";
import { UidGidSource } from "@/swagger-models/assets-service-client";

export default defineComponent({
  components: {
    RunaiExpansionItem,
    FormSecuritySettings,
  },
  props: {
    securityModel: {
      type: Object as PropType<ISecurityModel>,
      required: true,
    },
    isSSO: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
  },
  computed: {
    summary(): string {
      switch (this.securityModel.uidGidSource) {
        case UidGidSource.FromTheImage:
          return UID_GID_SOURCE_LABELS.fromTheImage;
        case UidGidSource.FromIdpToken:
          return UID_GID_SOURCE_LABELS.fromIdpToken;
        case UidGidSource.Custom:
          return UID_GID_SOURCE_LABELS.custom;
        default:
          return "";
      }
    },
  },
});
</script>

<style lang="scss">
.form-security-section {
  .q-radio {
    margin-right: 20px;
  }
  .linux-capabilities-section {
    .q-chip__content {
      padding-left: 5px;
      padding-right: 5px;
    }
  }
}
</style>
