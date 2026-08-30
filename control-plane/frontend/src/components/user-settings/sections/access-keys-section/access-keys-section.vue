<template>
  <runai-expansion-item class="access-keys-container" :subheader="accessKeysSubheader">
    <template #header>
      <span>Access keys <span class="formerly-text">(Formerly applications)</span></span>
    </template>
    <runai-section
      title="Create access keys for access to the NVIDIA Run:ai API"
      tooltip-text="Access keys work like ordinary OAuth tokens and are used to authenticate to the NVIDIA Run:ai API. You can create up to 20."
    >
      <access-key-box
        class="q-mb-md"
        v-for="accessKey in accessKeys"
        :key="accessKey.id"
        :access-key="accessKey"
        @remove-access-key="openDeleteModal(accessKey)"
        @regenerate-password="regeneratePassword(accessKey)"
      />
      <div class="row items-center q-mb-md">
        <runai-btn
          icon="ph-bold ph-plus"
          label="ACCESS KEY"
          class="btn-icon-sm"
          flat
          color="primary"
          @click="addAccessKey"
          aid="aidAddApplication"
          :disable="disableAddAccessKey"
          disabled-tooltip="You can create up to 20 access keys. To add more, first delete some."
        />
      </div>
      <runai-delete-modal
        v-if="isDeleteModalOpen && selectedAccessKey"
        :entity-name="selectedAccessKey.name"
        entity-type="access key"
        @cancel="closeDeleteModal"
        @delete="removeAccessKey"
        :deleting="deleting"
      >
        <template v-slot:description>
          <div>This can't be undone.</div>
        </template>
      </runai-delete-modal>

      <credential-creation-modal
        v-if="isCreationModalOpen"
        entity-type="access-key"
        :action-type="modalActionType"
        :model-value="isCreationModalOpen"
        :credential="selectedAccessKey"
        @close="closeCreationModal"
        @load-credential="getAccessKeys"
      />
    </runai-section>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// types
import type { AccessKey } from "@/swagger-models/identity-manager-client";
import type { TActionTypes } from "@/components/common/credential-creation-modal";

// service
import { accessKeysService } from "@/services/control-plane/access-keys.service/access-keys.service";

// components
import { RunaiDeleteModal } from "@/components/common";
import { CredentialCreationModal } from "@/components/common/credential-creation-modal";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiSection } from "@/components/common/runai-section";
import AccessKeyBox from "@/components/user-settings/sections/access-keys-section/access-key-box/access-key-box.vue";
import { RunaiBtn } from "@/components/common/runai-btn";

// utils
import { alertUtil } from "@/utils/alert.util";

export default defineComponent({
  name: "access-keys-section",
  components: {
    AccessKeyBox,
    RunaiSection,
    CredentialCreationModal,
    RunaiDeleteModal,
    RunaiExpansionItem,
    RunaiBtn,
  },
  data() {
    return {
      isCreationModalOpen: false,
      isDeleteModalOpen: false,
      selectedAccessKey: undefined as AccessKey | undefined,
      deleting: false,
      accessKeys: [] as AccessKey[],
    };
  },
  async created() {
    await this.getAccessKeys();
  },
  computed: {
    accessKeysSubheader(): string {
      const keysCount = this.accessKeys.length;
      return this.accessKeys.length == 0 ? "None" : keysCount == 1 ? "1 access key" : `${keysCount} access keys`;
    },
    disableAddAccessKey(): boolean {
      return this.accessKeys.length >= 20;
    },
    modalActionType(): TActionTypes {
      const isKeySelected = !!this.selectedAccessKey;
      return !isKeySelected ? "create" : "regenerate";
    },
  },
  methods: {
    addAccessKey(): void {
      this.selectedAccessKey = undefined;
      this.isCreationModalOpen = true;
    },
    async getAccessKeys(): Promise<void> {
      this.selectedAccessKey = undefined;
      this.accessKeys = await accessKeysService.getAccessKeys();
    },
    async removeAccessKey(): Promise<void> {
      if (!this.selectedAccessKey) {
        return;
      }
      this.deleting = true;
      try {
        await accessKeysService.removeAccessKey(this.selectedAccessKey?.id);
        this.$q.notify(alertUtil.getSuccess(`Access Key ${this.selectedAccessKey.name} deleted`));
        await this.getAccessKeys();
      } catch (error: unknown) {
        console.error("Failed to delete access key", error);
        this.$q.notify(alertUtil.getError("Failed to delete access key"));
      } finally {
        this.deleting = false;
        this.isDeleteModalOpen = false;
      }
    },
    async openDeleteModal(accessKey: AccessKey): Promise<void> {
      this.selectedAccessKey = accessKey;
      this.isDeleteModalOpen = true;
    },
    closeDeleteModal() {
      this.isDeleteModalOpen = false;
      this.deleting = false;
    },
    async closeCreationModal() {
      this.isCreationModalOpen = false;
      await this.getAccessKeys();
    },
    regeneratePassword(accessKey: AccessKey): void {
      this.selectedAccessKey = accessKey;
      this.isCreationModalOpen = true;
    },
  },
});
</script>
<style lang="scss" scoped>
.access-keys-container {
  margin: auto;
  width: 800px;
  .icon-letter {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    font-size: 40px;
    color: $white;
  }
  .avatar {
    background-color: $black-30;
    border-radius: 50%;
    font-style: normal;

    &.new-background {
      background: $light-navy-70;
    }
  }
}

.formerly-text {
  color: $black-54;
  font-style: italic;
  font-size: 12px;
}
</style>
