<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="credential-gallery">
    <runai-form-card-section
      :loading="loading"
      empty-message="Looks like you don't have any credentials yet..."
      :main-message="cardsMainTitle"
      entity-name="credential"
      :cards-list="credentialCards"
      :selected-cards-ids="selectedCredentials"
      icon-name="ph-bold ph-key"
      search-name="credentials"
      :multi-select="false"
      @selected-card-changed="onSelectedCredential"
      :sort-options="{ name: true, creation: true }"
      default-sort-option="name"
    >
      <template #add-new>
        <q-btn
          flat
          color="primary"
          :disabled="sectionDisabled || disableCreateButton"
          aid="credential-gallery-add-new-btn"
        >
          <q-icon left name="ph-bold ph-plus" size="16px"></q-icon>
          <div>new credential</div>
          <credential-dropdown @credential-selected="createCredential" :visible-create-options="allowedCreatedTypes" />
        </q-btn>
      </template>
    </runai-form-card-section>
    <div class="col-12 q-ml-sm" v-if="required">
      <q-field class="hidden-field gallery-message" :model-value="selectedCredentials" :rules="[isRequired]"></q-field>
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// models
import type { PolicyInfo, CredentialsListResponseEntry } from "@/swagger-models/assets-service-client";
import { ResourceType, Action } from "@/swagger-models/authorization-client";
import { TCardCmpName, type ICardListItem } from "@/components/common/runai-card-list";

// stores
import { usePermissionStore } from "@/stores/permissions.store";
import { useClusterStore } from "@/stores/cluster.store";

// components
import { RunaiFormCardSection } from "@/components/common";
import { CredentialDropdown } from "@/components/credential/credential-dropdown";

// routes
import type { CREDENTIAL_ROUTE_NAMES } from "@/router/credential.routes/credential.routes.names";

// constants
import { errorMessages } from "@/common/error-message.constant";

export default defineComponent({
  name: "credentials-gallery",
  components: {
    RunaiFormCardSection,
    CredentialDropdown,
  },
  emits: ["credentials-changed", "create-new", "is-section-invalid"],
  props: {
    loading: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    /**
     * The credential list that are shown in the gallery.
     */
    credentials: {
      type: Array as PropType<Array<CredentialsListResponseEntry>>,
      required: true,
    },
    /**
     * The credential types that are allowed to create for the workload type.
     * This is used to filter the credentials that are shown in create new credential button.
     * If you want to filter the credential cards from the gallery, you need to filter the credentials prop.
     */
    allowedCreatedTypes: {
      type: Set as PropType<Set<CREDENTIAL_ROUTE_NAMES>>,
      required: true,
    },
    selectedCredentials: {
      type: Array as PropType<Array<string>>,
      required: true,
    },
    cardsMainTitle: {
      type: String as PropType<string>,
      required: false,
    },
    sectionDisabled: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<PolicyInfo | null>,
      required: false,
    },
    required: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
  },
  data() {
    return {
      permissionStore: usePermissionStore(),
      clusterStore: useClusterStore(),
    };
  },
  computed: {
    disableCreateButton(): boolean {
      return !this.permissionStore.hasPermission(ResourceType.Credentials, Action.Create);
    },
    lockedCards(): Array<ICardListItem> {
      return this.credentialCards.filter((card) => card.locked);
    },
    credentialCards(): Array<ICardListItem> {
      return this.credentials.map((credential: CredentialsListResponseEntry) => ({
        id: credential.meta.id,
        cardName: TCardCmpName.CREDENTIAL,
        data: credential,
        searchValues: [credential.meta.name].filter((val) => !!val),
        locked: false,
        disabled: false,
        tooltip: "",
        showDisabledInfo: false,
        sortInfo: {
          name: credential.meta.name,
          createdAt: credential.meta.createdAt,
        },
      }));
    },
  },
  methods: {
    onSelectedCredential(selectedItems: Array<string>): void {
      this.$emit("credentials-changed", selectedItems);
    },
    createCredential(pageName: string): void {
      this.$emit("create-new", pageName);
    },
    isRequired(val: Array<ICardListItem>): boolean | string {
      return val.length > 0 || errorMessages.SELECT_CREDENTIAL;
    },
  },
});
</script>

<style lang="scss" scoped></style>
