<template>
  <runai-expansion-item class="scopes-section" label="Scopes" default-opened should-validate>
    <runai-section
      title="Set the scopes for this data volume"
      tooltip-text="The selected scopes will be able to mount the data volume with read-only permissions"
    >
      <runai-multi-scope-input
        :selected-scopes="selectedScopes"
        @update:selected-scopes="updateSelectedScopes"
        disable-tenant
        :allowed-scopes="allowedScopes"
        :parent-cluster-id="parentClusterId"
        :allow-empty-selection="allowEmptySelection"
      />
    </runai-section>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiSection } from "@/components/common/runai-section";
import { RunaiMultiScopeInput } from "@/components/common/runai-multi-scope-input";
// Models
import type { SharedScopeBase } from "@/swagger-models/datavolumes-service-client";
import { ScopeType } from "@/swagger-models/datavolumes-service-client";
import type { IOrgTreeNodeId } from "@/models/org-tree.model";
import type { PermittedScopes } from "@/swagger-models/authorization-client";

export default defineComponent({
  name: "scopes-section",
  components: {
    RunaiExpansionItem,
    RunaiSection,
    RunaiMultiScopeInput,
  },
  emits: {
    "update:model-value": (scopes: Array<SharedScopeBase>) => scopes,
  },
  props: {
    modelValue: {
      type: Array as PropType<Array<SharedScopeBase>>,
      required: true,
    },
    allowedScopes: {
      type: Object as PropType<PermittedScopes>,
      required: true,
    },
    parentClusterId: {
      type: String as PropType<string>,
      required: false,
    },
    allowEmptySelection: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  computed: {
    selectedScope(): IOrgTreeNodeId {
      if (!this.modelValue || this.modelValue.length === 0) {
        return { id: "", type: ScopeType.Project };
      }
      return {
        id: this.modelValue[0].scopeId,
        type: this.modelValue[0].scopeType,
      };
    },
    selectedScopes(): IOrgTreeNodeId[] {
      return this.modelValue.map((scope) => ({ id: scope.scopeId, type: scope.scopeType }));
    },
  },
  methods: {
    updateSharedScopes(scope: IOrgTreeNodeId | undefined) {
      if (!scope) {
        this.$emit("update:model-value", []);
        return;
      }

      this.$emit("update:model-value", [{ scopeId: scope.id, scopeType: scope.type as ScopeType }]);
    },
    updateSelectedScopes(selectedScopes: IOrgTreeNodeId[]) {
      this.$emit(
        "update:model-value",
        selectedScopes.map((scope) => ({ scopeId: scope.id, scopeType: scope.type as ScopeType })),
      );
    },
  },
});
</script>
