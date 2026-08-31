<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-menu fit>
    <q-list>
      <q-item
        :aid="opt.page + '-btn'"
        clickable
        v-close-popup
        v-for="opt in options"
        :key="opt.title"
        @click="$emit('credential-selected', opt.page)"
      >
        <q-item-section class="option">
          {{ opt.title }}
        </q-item-section>
      </q-item>
    </q-list>
  </q-menu>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { credentialTextForUI } from "@/models/credential.model";
import { CREDENTIAL_ROUTE_NAMES } from "@/router/credential.routes/credential.routes.names";

export interface ICredentialOption {
  title: string;
  page: CREDENTIAL_ROUTE_NAMES;
}

export default defineComponent({
  props: {
    /**
     * A set of `CREDENTIAL_ROUTE_NAMES` that determines the visible credential creation options.
     *
     * - If this set contains one or more route names, only the matching options will be displayed.
     * - If the set is empty or not provided, all available creation options will be displayed.
     *
     * @prop {Set<CREDENTIAL_ROUTE_NAMES>} visibleCreateOptions - Specifies the routes to filter visible options.
     */
    visibleCreateOptions: {
      type: Object as PropType<Set<CREDENTIAL_ROUTE_NAMES>>,
      default: () => new Set(),
    },
  },
  data() {
    return {
      allOptions: [
        {
          title: credentialTextForUI.dockerRegistry,
          page: CREDENTIAL_ROUTE_NAMES.CREDENTIAL_DOCKER_REGISTRY_NEW,
        },
        {
          title: credentialTextForUI.genericSecret,
          page: CREDENTIAL_ROUTE_NAMES.CREDENTIAL_GENERIC_SECRET_NEW,
        },
        {
          title: credentialTextForUI.ngcApiKey,
          page: CREDENTIAL_ROUTE_NAMES.CREDENTIAL_NGC_API_KEY_NEW,
        },
      ] as Array<ICredentialOption>,
    };
  },
  computed: {
    options(): Array<ICredentialOption> {
      return this.visibleCreateOptions.size > 0
        ? this.allOptions.filter((currOption: ICredentialOption) => this.visibleCreateOptions.has(currOption.page))
        : this.allOptions;
    },
  },
});
</script>

<style lang="scss" scoped>
.option {
  flex-direction: row;
  gap: 10px;
  justify-content: flex-start;
  align-items: center;
}
</style>
