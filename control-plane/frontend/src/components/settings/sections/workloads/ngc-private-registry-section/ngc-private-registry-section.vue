<template>
  <runai-section
    class="ngc-private-registry-section"
    title="NGC private registry"
    tooltip-text="Enable browsing of your organization’s proprietary container image registry, hosted on the NGC private registry service"
  >
    <template #title-right-content>
      <runai-stage-badge class="beta-badge" display-tooltip :badge="EBadgeLabel.Beta" />
    </template>
    <section class="italic">
      For more information, see the
      <a
        href="https://docs.nvidia.com/ngc/latest/ngc-private-registry-user-guide.html#joining-an-org-as-org-owner"
        target="_blank"
        >Joining an Org as Org Owner</a
      >
      documentation.
    </section>
    <section class="q-mt-md">
      <ngc-private-registry-box
        v-for="registry in registries"
        :key="registry.id"
        :name="registry.name"
        :url="registry.url"
        :readonly="isReadOnly"
        @update-registry="updateRegistry(registry.id, $event)"
        @remove-registry="deleteRegistry(registry.id)"
        class="q-mb-md"
      />
    </section>
    <ngc-private-registry-box
      v-if="showAddRegistryBox"
      class="q-mb-md"
      :loading="isCreating"
      :name="registryToAdd.name"
      :url="registryToAdd.url"
      @update-registry="createRegistry($event)"
      @cancel="cancelAddRegistry"
      create-mode
    />
    <setting-readonly-input resource-type="Registries">
      <template v-slot:default="{ readonly }">
        <q-btn
          icon="ph-bold ph-plus"
          label="registry"
          class="btn-icon-sm"
          flat
          color="primary"
          :disable="readonly"
          @click="addRegistryClicked"
        />
      </template>
    </setting-readonly-input>
  </runai-section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
// Components
import { RunaiSection } from "@/components/common/runai-section";
import { RunaiStageBadge } from "@/components/common/runai-stage-badge";
import { NgcPrivateRegistryBox } from "./ngc-private-registry-box";
import SettingReadonlyInput from "@/components/settings/setting-readonly-input/setting-readonly-input.vue";
// Models
import { EBadgeLabel } from "@/models/setting.model";
import {
  ScopeType,
  type RegistryIntegration,
  type RegistryIntegrationCreationRequest,
  type RegistryIntegrationUpdateRequest,
} from "@/swagger-models/catalog-service-client";
// Services
import { catalogService } from "@/services/control-plane/catalog.service/catalog.service";
// Stores
import { useAuthStore } from "@/stores/auth.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { ResourceType } from "@/swagger-models/authorization-client";
import { alertUtil } from "@/utils/alert.util";

const defaultRegistryUrl = "https://api.ngc.nvidia.com/v2/search/org//resources/CONTAINER";
export default defineComponent({
  name: "ngc-private-registry-section",
  components: {
    RunaiSection,
    RunaiStageBadge,
    NgcPrivateRegistryBox,
    SettingReadonlyInput,
  },
  emits: ["create-registry", "remove-registry", "update-registry"],
  data() {
    return {
      authStore: useAuthStore(),
      registries: [] as RegistryIntegration[],
      EBadgeLabel,
      registryToAdd: {
        scopeType: ScopeType.Tenant,
        name: "",
        url: defaultRegistryUrl,
      } as RegistryIntegrationCreationRequest,
      showAddRegistryBox: false,
      isCreating: false,
      isUpdating: false,
      isDeleting: false,
    };
  },
  created() {
    this.loadRegistries();
  },
  computed: {
    isReadOnly(): boolean {
      return usePermissionStore().isReadOnly(ResourceType.Registries);
    },
  },
  methods: {
    async loadRegistries(): Promise<void> {
      try {
        const registries = await catalogService.getRegistries();
        this.registries = registries.filter((registry) => registry.scopeType === ScopeType.Tenant);
      } catch (error: unknown) {
        console.error(error);
      }
    },
    cancelAddRegistry(): void {
      this.showAddRegistryBox = false;
      this.registryToAdd.name = "";
      this.registryToAdd.url = defaultRegistryUrl;
    },
    addRegistryClicked(): void {
      this.showAddRegistryBox = true;
    },
    async createRegistry(registry: RegistryIntegrationCreationRequest): Promise<void> {
      try {
        this.isCreating = true;
        registry.scopeType = ScopeType.Tenant;
        registry.scopeId = this.authStore.tenant.id.toString();
        const newRegistry = await catalogService.createRegistry(registry);
        this.registries.push(newRegistry);
        this.$q.notify(alertUtil.getSuccess(`Settings saved`));
      } catch (error: unknown) {
        console.error(error);
      } finally {
        this.isCreating = false;
      }
      this.showAddRegistryBox = false;

      // Reset the registry to add after creation
      this.registryToAdd.name = "";
      this.registryToAdd.url = defaultRegistryUrl;
    },
    async updateRegistry(registryId: string, registry: RegistryIntegrationUpdateRequest): Promise<void> {
      try {
        const updatedRegistry = await catalogService.updateRegistry(registryId, registry);
        this.registries = this.registries.map((r) => (r.id === registryId ? updatedRegistry : r));
        this.$q.notify(alertUtil.getSuccess(`Settings saved`));
      } catch (error: unknown) {
        console.error(error);
      }
    },
    async deleteRegistry(registryId: string): Promise<void> {
      try {
        await catalogService.deleteRegistry(registryId);
        this.registries = this.registries.filter((r) => r.id !== registryId);
        this.$q.notify(alertUtil.getSuccess(`Settings saved`));
      } catch (error: unknown) {
        console.error(error);
      }
    },
  },
});
</script>

<style scoped lang="scss">
.beta-badge {
  margin-right: 34px; // this margin is used to align the badge with all other badges in the section containing this component.
}
</style>
