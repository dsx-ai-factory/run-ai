<template>
  <!--  Batch Delete access rules  -->
  <runai-delete-modal
    :deleting="deleting"
    :custom-title="deletionTitle.main"
    :delete-button-label="deleteButtonLabel"
    :custom-description="customDescription"
    @cancel="$emit('cancel')"
    @delete="confirmBatchDeletion"
  >
    <template v-if="deletionTitle.secondary" #subheader>
      <div class="row q-mb-lg items-center">
        <div>{{ deletionTitle.secondary }}</div>
        <runai-tooltip
          size="xs"
          tooltip-position="right"
          width="100%"
          tooltip-text="You aren't authorized to delete at least one of the selected access rules"
        />
      </div>
    </template>
  </runai-delete-modal>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import { ErrorAlert } from "@/utils/error-alert.util";
import { RunaiDeleteModal } from "@/components/common/runai-delete-modal/";
import { type AccessRule, AccessRulesBatchAction } from "@/swagger-models/authorization-client";
import { RunaiTooltip } from "@/components/common/runai-tooltip";

export interface DeletionTitle {
  main: string;
  secondary?: string;
}

export default defineComponent({
  name: "AccessRuleDeletionModal",
  components: {
    RunaiTooltip,
    RunaiDeleteModal,
  },
  emits: ["cancel", "deletion-success"],
  props: {
    selectedAccessRules: {
      type: Array as PropType<AccessRule[]>,
      required: true,
    },
    deletableAccessRulesIds: {
      type: Array as PropType<number[]>,
      required: true,
    },
  },
  data() {
    return {
      deleting: false,
    };
  },
  computed: {
    selectedAccessRulesIds(): Array<number> {
      return this.selectedAccessRules.map((accessRule: AccessRule) => accessRule.id);
    },
    deletionTitle(): DeletionTitle {
      const total = this.selectedAccessRules.length;
      const deletable = this.deletableAccessRulesIds.length;
      const nonDeletable = total - deletable;

      let mainTitle = "";
      let secondaryTitle = undefined;

      if (total === 1) {
        mainTitle = "Permanently delete 1 access rule?";
      } else if (deletable === total) {
        mainTitle = `Permanently delete ${total} access rules?`;
      } else if (nonDeletable > 0) {
        secondaryTitle = `${nonDeletable} out of ${total} access rules can't be deleted.`;
        if (deletable === 1) {
          mainTitle = `Permanently delete 1 access rule?`;
        } else {
          mainTitle = `Permanently delete ${deletable} access rules?`;
        }
      }

      return { main: mainTitle, secondary: secondaryTitle };
    },
    deleteButtonLabel(): string {
      if (this.selectedAccessRules.length === 1) {
        return "Delete";
      }
      if (this.deletableAccessRulesIds.length === 1) {
        return "Delete 1 access rule";
      }
      return `Delete ${this.deletableAccessRulesIds.length} access rules`;
    },
    customDescription(): string {
      return this.deletableAccessRulesIds.length === 1
        ? "The subject will no longer have access rights to their managed objects (i.e., project, department). This can't be undone."
        : "Subject(s) will no longer have access rights to their managed objects (i.e., project, department). This can't be undone.";
    },
  },
  methods: {
    /** V2 - Batch Delete access rules */
    async confirmBatchDeletion() {
      try {
        this.deleting = true;
        await accessRuleService.accessRulesBatch({
          action: AccessRulesBatchAction.Delete,
          ids: this.deletableAccessRulesIds.map(String).concat("4"),
        });
        this.$emit("deletion-success", this.deletableAccessRulesIds);
      } catch (error: unknown) {
        this.notifyError(error);
      } finally {
        this.deleting = false;
      }
    },
    notifyError(error: unknown) {
      const errorAlert = new ErrorAlert({
        generalMessage: "Access rule couldn't be deleted",
      });
      this.$q.notify(errorAlert.getNotification(error));
    },
  },
});
</script>

<style scoped></style>
