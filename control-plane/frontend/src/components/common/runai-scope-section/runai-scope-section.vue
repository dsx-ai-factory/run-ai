<template>
  <runai-expansion-item
    label="Scope"
    :default-opened="defaultOpened"
    :hide-expend-icon="hideExpansionIcon"
    :disable-opening="readOnly"
    aid="runai-scope-section"
    should-validate
  >
    <template #subheader>
      <div class="row items-center justify-between q-pr-sm">
        <span aid="scope-section-subheader">{{ summary }}</span>
        <runai-btn
          v-if="readOnly && backTooltip"
          icon="ph-bold ph-pencil-simple"
          flat
          round
          size="sm"
          @click="$emit('on-back')"
          aid="scope-section-back-btn"
        >
          <q-tooltip anchor="center right" self="center left">
            <span class="text-no-wrap">{{ backTooltip }}</span>
          </q-tooltip>
        </runai-btn>
      </div>
    </template>
    <section v-show="!readOnly">
      <div class="q-mb-md" v-if="kindOfAsset">Set the scope for this {{ kindOfAsset }}</div>
      <div class="row items-center">
        <q-field
          aid="scope-field"
          class="col"
          ref="scope-input"
          color="grey-8"
          :model-value="scope"
          label="Scope"
          no-error-icon
          clearable
          @clear="resetScope"
          @focus="isOrgTreeModalOpen = true"
          clear-icon="ph-bold ph-x"
          :rules="[isValidScope]"
        >
          {{ scope }}
        </q-field>
        <q-btn
          aid="scope-tree-btn"
          class="scope-tree-btn q-ml-md runai-icon-btn--default"
          ref="tree-btn"
          color="primary"
          size="sm"
          icon="ph-fill ph-tree-view"
          @click="isOrgTreeModalOpen = true"
        />
      </div>
      <runai-org-tree-with-pop-over
        v-if="isOrgTreeModalOpen"
        :selected="selectedScope ? [selectedScope] : []"
        @update:selected="changedScope"
        @hide="validateScope"
        :allowed-scopes="allowedScopes"
        :scopes-supported-versions="scopesSupportedVersions"
        disable-tenant
        :forced-disabled-scopes="forcedDisabledScopes"
        :selected-cluster-id="selectedClusterId"
      />
      <notification-alert-bar class="flex-1" v-if="informationText">
        <span v-dompurify-html="informationText"></span>
      </notification-alert-bar>

      <slot />
    </section>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiOrgTreeWithPopOver } from "@/components/common/runai-org-tree/with-popover";
import { NotificationAlertBar } from "../notification-alert-bar";
import { RunaiBtn } from "@/components/common/runai-btn";
// Constants
import { errorMessages } from "@/common/error-message.constant";
import type { IOrgTreeNode, IOrgTreeNodeId } from "@/models/org-tree.model";
import { ScopeType, type PermittedScopes } from "@/swagger-models/authorization-client";
import { Scope } from "@/swagger-models/assets-service-client";

import type { QBtn, QField } from "quasar";
import { useAuthStore } from "@/stores/auth.store";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import type { IScopeModel, TTreeDisabledScopes, TTreeSupportedVersions } from "@/models/global.model";
import { useAppStore } from "@/stores/app.store";

export default defineComponent({
  components: {
    RunaiExpansionItem,
    RunaiOrgTreeWithPopOver,
    NotificationAlertBar,
    RunaiBtn,
  },
  emits: ["changed-scope", "on-back"],
  props: {
    readOnly: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    kindOfAsset: {
      type: String as PropType<string>,
      required: false,
    },
    scopeModel: {
      type: Object as PropType<IScopeModel>,
      required: true,
    },
    allowedScopes: {
      type: Object as PropType<PermittedScopes>,
      required: false,
    },
    scopesSupportedVersions: {
      type: Object as PropType<TTreeSupportedVersions>,
      default: () => ({}),
    },
    forcedDisabledScopes: {
      type: Object as PropType<TTreeDisabledScopes>,
      default: () => ({}),
    },
    selectedClusterId: {
      type: String as PropType<string>,
      required: true,
    },
    informationText: {
      type: String as PropType<string>,
      required: false,
    },
    // relevant only if readOnly is true
    backTooltip: {
      type: String as PropType<string>,
      required: false,
    },
    // relevant only if readOnly is false (readOnly forces the section to be closed)
    initClosed: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      appStore: useAppStore(),
      authStore: useAuthStore(),
      tree: [] as Array<IOrgTreeNode>,
      isOrgTreeModalOpen: false as boolean,
    };
  },
  created() {
    // Need the tree here to get the path of the selected scope before the tree component is created.
    this.tree = [orgTreeService.getOrgTree(this.authStore.orgUnitList)];
  },
  computed: {
    scope(): string | null {
      if (!this.selectedScope || !this.tree.length) return null;
      // Setting the value of the scope input if the scope is set.
      return orgTreeService.getNodePathById(this.tree[0], this.selectedScope);
    },
    selectedScope(): IOrgTreeNodeId | null {
      if (!this.scopeModel.scope) return null;
      const details: IOrgTreeNodeId = { type: this.scopeModel.scope as ScopeType, id: "" };
      switch (this.scopeModel.scope) {
        case ScopeType.Tenant:
          if (!this.authStore.tenant.id) return null;
          details.id = this.authStore.tenant.id.toString();
          break;
        case ScopeType.Cluster:
          if (!this.scopeModel.clusterId) return null;
          details.id = this.scopeModel.clusterId?.toString();
          break;
        case ScopeType.Department:
          if (!this.scopeModel.departmentId) return null;
          details.id = this.scopeModel.departmentId?.toString();
          break;
        case ScopeType.Project:
          if (!this.scopeModel.projectId) return null;
          details.id = this.scopeModel.projectId?.toString();
          break;
      }
      return details;
    },
    summary(): string {
      return this.scope || "None";
    },
    hideExpansionIcon(): boolean {
      if (this.readOnly) return true;
      return false;
    },
    defaultOpened(): boolean {
      if (this.readOnly) return false;
      if (this.initClosed) return false;
      return true;
    },
  },
  methods: {
    validateScope(): void {
      this.isOrgTreeModalOpen = false;
      (this.$refs["scope-input"] as QField).validate();
    },
    isValidScope(val: string | null): boolean | string {
      return val ? true : errorMessages.SCOPE_TYPE_NOT_EMPTY;
    },
    changedScope(selectedScope: IOrgTreeNodeId): void {
      const parentClusterUuid = orgTreeService.getParentClusterUuid(
        this.authStore.orgUnitList,
        selectedScope.id,
        selectedScope.type,
      );
      const updatedScope: IScopeModel = { scope: selectedScope.type, parentClusterUuid };
      switch (selectedScope.type) {
        case ScopeType.Tenant:
          break;
        case Scope.Cluster:
          updatedScope.clusterId = selectedScope.id;
          break;
        case Scope.Department:
          updatedScope.departmentId = selectedScope.id;
          break;
        case Scope.Project:
          updatedScope.projectId = Number(selectedScope.id);
          break;
      }

      this.$emit("changed-scope", updatedScope);
      // Trigger validation immediately after scope changes to ensure it shows as valid
      this.$nextTick(() => {
        const field = this.$refs["scope-input"] as QField;
        if (field) {
          field.validate();
        }
      });
    },
    resetScope(): void {
      this.$emit("changed-scope", { scope: null });
    },
    openTree(): void {
      (this.$refs["tree-btn"] as QBtn).$el.click();
      this.$nextTick(() => {
        (this.$refs["scope-input"] as QField).resetValidation();
      });
    },
  },
});
</script>
<style lang="scss" scoped>
.scope-tree-btn {
  width: 36px;
  height: 36px;
}
</style>
