<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item class="user-credentials-container" label="Credentials" :subheader="subheader" body-no-padding>
    <notification-alert-bar class="q-pa-md" v-if="shouldShowInfoBar" :variant="ENotificationBarVariant.Info">
      <span> Credentials created here will be automatically propagated to all clusters from version 2.22 onwards </span>
    </notification-alert-bar>
    <runai-section
      tooltip-width="350px"
      title="Create credentials to use while submitting your workloads"
      tooltip-text="These credentials are private and available only to you"
      class="q-pa-md"
    >
      <q-skeleton v-if="loadingCredentials" class="q-mb-md" height="200px" square />

      <personal-credential-item
        v-else
        v-for="credential in credentials"
        :key="credential.meta.id"
        :credential="credential"
        @delete-credential="openDeleteCredentialModal"
      />
      <div class="row items-center q-mb-md">
        <runai-btn
          icon="ph-bold ph-plus"
          label="credential"
          class="btn-icon-sm"
          flat
          color="primary"
          aid="add-new-credential"
        >
          <credential-dropdown
            @credential-selected="openCredentialCreationModal"
            :visible-create-options="visibleCreateOptions"
          />
        </runai-btn>
      </div>
    </runai-section>
  </runai-expansion-item>
  <runai-delete-modal
    v-if="credentialToDelete"
    :entity-name="credentialToDelete?.name"
    :deleting="isDeleting"
    custom-description="This can't be undone."
    entity-type="credential"
    @cancel="credentialToDelete = null"
    @delete="deleteCredential"
  />
  <credential-creation-modal
    v-if="credentialToCreate"
    :credential-type="credentialToCreate"
    :loading="isCreating"
    @cancel="credentialToCreate = null"
    @create="createCredential"
  />
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiSection } from "@/components/common/runai-section";
import { NotificationAlertBar, ENotificationBarVariant } from "@/components/common/notification-alert-bar";
import { PersonalCredentialItem } from "./personal-credential-item";
import type { TCredentialCreationRoutes, TCredentialKinds } from "@/models/credential.model";
import { CREDENTIAL_CREATION_ROUTES_TO_CREDENTIAL_KINDS } from "@/models/credential.model";
import { RunaiBtn } from "@/components/common/runai-btn";
import { CredentialDropdown } from "@/components/credential/credential-dropdown";
import { RunaiDeleteModal } from "@/components/common/runai-delete-modal";
import type { CredentialsListResponseEntry, NgcApiKeyCreationRequest } from "@/swagger-models/assets-service-client";
import { AssetKind, Scope } from "@/swagger-models/assets-service-client";
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { CredentialCreationModal } from "./credential-creation-modal";
import { CREDENTIAL_ROUTE_NAMES } from "@/router/credential.routes/credential.routes.names";
import { MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS } from "@/common/version.constant";
import { useClusterStore } from "@/stores/cluster.store";
import type {
  DockerRegistryCreationRequest,
  GenericSecretCreationRequest,
} from "@/swagger-models/assets-service-client";

interface IDeleteCredential {
  kind: TCredentialKinds;
  id: string;
  name: string;
}

export default defineComponent({
  name: "personal-credentials-section",
  components: {
    RunaiExpansionItem,
    RunaiSection,
    NotificationAlertBar,
    PersonalCredentialItem,
    RunaiBtn,
    CredentialDropdown,
    RunaiDeleteModal,
    CredentialCreationModal,
  },
  data() {
    return {
      isDeleting: false,
      credentials: [] as CredentialsListResponseEntry[],
      loadingCredentials: false,
      credentialToDelete: null as IDeleteCredential | null,
      credentialToCreate: null as TCredentialKinds | null,
      visibleCreateOptions: new Set([
        CREDENTIAL_ROUTE_NAMES.CREDENTIAL_DOCKER_REGISTRY_NEW,
        CREDENTIAL_ROUTE_NAMES.CREDENTIAL_GENERIC_SECRET_NEW,
        CREDENTIAL_ROUTE_NAMES.CREDENTIAL_NGC_API_KEY_NEW,
      ]),
      isCreating: false,
      clusterStore: useClusterStore(),
    };
  },
  created() {
    this.getPersonalCredentials();
  },
  computed: {
    subheader(): string {
      return this.credentials?.length ? `${this.credentials.length} credentials` : "None";
    },
    shouldShowInfoBar(): boolean {
      return this.clusterStore.clusterList.some((cluster) => {
        const version = cluster.version;
        if (!version) return true;
        return !this.clusterStore.isClusterVersionSufficient(cluster.uuid, MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS);
      });
    },
    ENotificationBarVariant() {
      return ENotificationBarVariant;
    },
  },
  methods: {
    async getPersonalCredentials(): Promise<void> {
      try {
        this.loadingCredentials = true;
        const creds = await credentialService.listCredentials({ scope: Scope.Tenant, privateOnly: true });
        this.credentials = creds.filter(
          (c) =>
            c.meta.kind === AssetKind.GenericSecret ||
            c.meta.kind === AssetKind.DockerRegistry ||
            c.meta.kind === AssetKind.NgcApiKey,
        );
      } catch (e: unknown) {
        console.error(e);
        this.$q.notify({
          message: "Failed to load personal credentials",
          color: "negative",
        });
        this.credentials = [];
      } finally {
        this.loadingCredentials = false;
      }
    },
    openCredentialCreationModal(credentialType: TCredentialCreationRoutes): void {
      if (!credentialType) return;
      this.credentialToCreate = CREDENTIAL_CREATION_ROUTES_TO_CREDENTIAL_KINDS[credentialType];
    },
    openDeleteCredentialModal(credential: IDeleteCredential): void {
      this.credentialToDelete = credential;
    },
    async deleteCredential(): Promise<void> {
      if (!this.credentialToDelete) return;
      try {
        const idToRemove = this.credentialToDelete.id;
        this.isDeleting = true;
        await credentialService.remove(idToRemove, this.credentialToDelete.kind);
        this.$q.notify({
          message: `Credential ${this.credentialToDelete?.name} deleted`,
          color: "positive",
        });
        this.credentials = this.credentials.filter((c) => c.meta.id !== idToRemove);
        this.credentialToDelete = null;
      } catch (e: unknown) {
        console.error(e);
        this.$q.notify({
          message: "Failed to delete credential",
          color: "negative",
        });
      } finally {
        this.isDeleting = false;
      }
    },
    async createCredential(
      credentialModel: DockerRegistryCreationRequest | GenericSecretCreationRequest | NgcApiKeyCreationRequest,
    ): Promise<void> {
      if (!credentialModel || !this.credentialToCreate) return;

      try {
        this.isCreating = true;
        switch (this.credentialToCreate) {
          case AssetKind.DockerRegistry:
            await credentialService.createDockerRegistry(credentialModel as DockerRegistryCreationRequest);
            break;
          case AssetKind.GenericSecret: {
            const preparedGenericSecret = credentialService.prepareGenericSecretForServerRequest(
              credentialModel as GenericSecretCreationRequest,
            );
            await credentialService.createGeneralCredential(preparedGenericSecret);
            break;
          }
          case AssetKind.NgcApiKey:
            await credentialService.createNgcApiKey(credentialModel as NgcApiKeyCreationRequest);
            break;
          default:
            throw new Error(`Unsupported credential type: ${this.credentialToCreate}`);
        }

        this.$q.notify({
          message: `Credential ${credentialModel.meta.name} created`,
          color: "positive",
        });

        this.credentialToCreate = null;
        await this.getPersonalCredentials();
      } catch (e: unknown) {
        console.error(e);
        this.$q.notify({
          message: "Failed to create credential",
          color: "negative",
        });
      } finally {
        this.isCreating = false;
      }
    },
  },
});
</script>

<style lang="scss" scoped>
.user-credentials-container {
  margin: auto;
  width: 800px;
}
</style>
