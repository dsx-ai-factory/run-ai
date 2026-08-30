<template>
  <section class="runai-multi-scope-input row no-wrap q-gutter-sm items-end">
    <q-field
      :rules="required ? [requiredScope] : undefined"
      class="col"
      label="Scopes"
      :model-value="selectedScopes"
      clearable
      @focus="isOrgTreeModalOpen = true"
      @clear="clearScopes"
      no-error-icon
      clear-icon="ph-bold ph-x"
      :hint="hint"
      :counter="counter"
      :maxlength="maxlength"
    >
      <template v-slot:default>
        <section class="row">
          <q-chip
            class="runai-chip-sm"
            v-for="scope in selectedScopesForDisplay"
            :key="scope.id"
            :aid="`${scope.id}-chip`"
            removable
            @remove="removeScope(scope.id)"
            >{{ scope.path }}</q-chip
          >
        </section>
      </template>
    </q-field>
    <q-btn
      class="scope-btn runai-icon-btn--default"
      color="primary"
      size="sm"
      icon="ph-fill ph-tree-view"
      @click="isOrgTreeModalOpen = true"
    />
    <runai-org-tree-with-pop-over
      v-if="isOrgTreeModalOpen"
      multiple
      :selected="selectedScopes"
      @update:selected-nodes="updateSelectedScopes"
      :disable-tenant="disableTenant"
      :allowed-scopes="allowedScopes"
      @hide="isOrgTreeModalOpen = false"
      :selected-cluster-id="parentClusterId"
      :allow-empty-selection="allowEmptySelection"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Models
import type { IOrgTreeNode, IOrgTreeNodeId } from "@/models/org-tree.model";
import type { PermittedScopes } from "@/swagger-models/authorization-client";
// Components
import { RunaiOrgTreeWithPopOver } from "../runai-org-tree/with-popover";
// Services
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
// Stores
import { useAuthStore } from "@/stores/auth.store";
import { errorMessages } from "@/common/error-message.constant";

export default defineComponent({
  name: "runai-multi-scope-input",
  components: {
    RunaiOrgTreeWithPopOver,
  },
  emits: {
    "update:selected-scopes": (value: IOrgTreeNodeId[]) => value,
  },
  props: {
    selectedScopes: {
      type: Array as PropType<IOrgTreeNodeId[]>,
      required: true,
    },
    disableTenant: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    allowedScopes: {
      type: Object as PropType<PermittedScopes>,
      required: false,
    },
    parentClusterId: {
      type: String as PropType<string>,
      required: false,
    },
    allowEmptySelection: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    required: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    hint: {
      type: String as PropType<string>,
      required: false,
    },
    counter: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    maxlength: {
      type: Number as PropType<number>,
      required: false,
    },
  },
  data() {
    return {
      tree: [] as IOrgTreeNode[],
      isOrgTreeModalOpen: false as boolean,
    };
  },
  created() {
    this.tree = [orgTreeService.getOrgTree(useAuthStore().orgUnitList)];
  },
  computed: {
    selectedScopesForDisplay(): IOrgTreeNodeId[] {
      return this.selectedScopes.map((scope) => {
        return {
          ...scope,
          path: orgTreeService.getNodePathById(this.tree[0], scope) || undefined,
        };
      });
    },
  },
  methods: {
    requiredScope(): boolean | string {
      return this.selectedScopes.length > 0 || errorMessages.SCOPE_TYPE_NOT_EMPTY;
    },
    updateSelectedScopes(selectedScopes: IOrgTreeNodeId[]): void {
      this.$emit("update:selected-scopes", selectedScopes);
    },
    removeScope(id: string): void {
      this.$emit(
        "update:selected-scopes",
        this.selectedScopes.filter((scope) => scope.id !== id),
      );
    },
    clearScopes(): void {
      this.$emit("update:selected-scopes", []);
    },
  },
});
</script>

<style lang="scss" scoped>
.scope-btn {
  width: 40px;
  height: 40px;
}
</style>
