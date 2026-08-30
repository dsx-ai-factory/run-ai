<template>
  <section class="q-pt-md row justify-start">
    <div class="col field-wrapper">
      <q-field
        class="scope-input"
        color="grey-8"
        label="Scope"
        :clearable="clearable"
        @clear="$emit('clear')"
        clear-icon="ph-bold ph-x"
        :model-value="selectedScopePath"
        :rules="[requiredScope]"
        no-error-icon
        ref="scope-input"
        @focus="isOrgTreeModalOpen = true"
        :disable="readonly"
      >
        {{ selectedScopePath }}
      </q-field>
      <q-tooltip v-if="fieldTooltip" anchor="top middle" self="top middle" :delay="300">
        {{ fieldTooltip }}
      </q-tooltip>
    </div>
    <div class="tree-button-wrapper">
      <q-btn
        aid="scope-tree-btn"
        class="tree-button runai-icon-btn--default"
        color="primary"
        icon="ph-fill ph-tree-view"
        size="sm"
        ref="tree-btn"
        @click="isOrgTreeModalOpen = true"
        aria-label="Open scope tree"
      />
    </div>
    <runai-org-tree-with-pop-over
      v-if="isOrgTreeModalOpen"
      @hide="isOrgTreeModalOpen = false"
      :readonly="readonly"
      :tick-strategy="tickStrategy"
      :allowed-scopes="allowedScopes"
      :hidden-scope-types="hiddenScopeTypes"
      :selected="selectedScope ? [selectedScope] : []"
      :placeholder-search="placeholderSearch"
      @update:selected="$emit('update-selected', $event)"
      :title="title"
      cover
      :forced-disabled-scopes="forcedDisabledScopes"
    />
  </section>
  <section class="q-pt-md row justify-start">
    <div class="col-12 q-mr-xl scope-input error">
      <p v-if="error">
        You don't have permissions to set this scope for the role selected above. Try selecting a different role and
        scope combination.
      </p>
    </div>
  </section>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiOrgTreeWithPopOver } from "@/components/common/runai-org-tree/with-popover";
import type { IOrgTreeNode, IOrgTreeNodeId } from "@/models/org-tree.model";
import type { ISelectOption } from "@/models/global.model";
import { errorMessages } from "@/common/error-message.constant";
import type { QBtn } from "quasar";
import type { ScopeType } from "@/swagger-models/authorization-client";
import { type PermittedScopes } from "@/swagger-models/authorization-client";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { useAuthStore } from "@/stores/auth.store";

export default defineComponent({
  name: "scope-input",
  components: { RunaiOrgTreeWithPopOver },
  emits: ["clear", "update-selected"],
  props: {
    selectedScope: {
      type: [Object, null] as PropType<null | IOrgTreeNodeId>,
      required: true,
    },
    error: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    allowedScopes: {
      type: Object as PropType<PermittedScopes>,
      required: false,
    },
    hiddenScopeTypes: {
      type: Object as PropType<Record<ScopeType, boolean>>,
      default: () => ({}),
    },
    tickStrategy: {
      type: String as PropType<"leaf" | "leaf-filtered" | "none" | "strict" | undefined>,
      default: "strict",
    },
    readonly: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    clearable: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    placeholderSearch: {
      type: String as PropType<string>,
      default: `Search for a cluster, department, or project`,
    },
    title: {
      type: String as PropType<string>,
      default: "Scope",
    },
    forcedDisabledScopes: {
      type: Object as PropType<Record<ScopeType, boolean>>,
      default: () => ({}),
    },
    fieldTooltip: {
      type: String as PropType<string>,
      required: false,
    },
  },
  data() {
    return {
      tree: [] as Array<IOrgTreeNode>,
      authStore: useAuthStore(),
      isOrgTreeModalOpen: false as boolean,
    };
  },
  created() {
    this.tree = [orgTreeService.getOrgTree(this.authStore.orgUnitList)];
  },
  computed: {
    selectedScopePath() {
      if (!this.tree.length || !this.selectedScope) return null;
      const selectedScope = {
        ...this.selectedScope,
        id: this.selectedScope.id?.toString() || "",
      };
      return orgTreeService.getNodePathById(this.tree[0], selectedScope);
    },
  },
  methods: {
    requiredScope(val: ISelectOption): boolean | string {
      return !!val || errorMessages.SCOPE_TYPE_NOT_EMPTY;
    },
    openTree(): void {
      if (this.readonly) {
        return;
      }
      (this.$refs["tree-btn"] as QBtn).$el.click();
    },
  },
});
</script>

<style scoped lang="scss">
.tree-button-wrapper {
  align-self: center;
  margin-left: 10px;
}

.tree-button {
  width: 36px;
  height: 36px;
}

.x-button {
  margin-top: 10px;
  color: $black-54;
  width: 40px;
  height: 40px;
}
</style>
