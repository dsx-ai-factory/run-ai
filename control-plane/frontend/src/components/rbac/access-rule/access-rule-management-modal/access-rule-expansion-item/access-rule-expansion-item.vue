<template>
  <card-item
    :disabled-delete-button="disabledDeleteButton"
    :delete-button-tooltip="deleteButtonToolTip"
    :permission-config="{ resourceType: ResourceType.AccessRules, action: Action.Delete }"
    tight-spacing
    @delete="$emit('delete')"
  >
    <template #summary>
      <slot name="summary"></slot>
    </template>
  </card-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { errorMessages } from "@/common/error-message.constant";
import { CardItem } from "@/components/common/card-item";
export default defineComponent({
  name: "access-rule-expansion-item",
  components: {
    CardItem,
  },
  emits: ["save", "delete"],
  props: {
    isAdminTenantRole: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    isInheritedFromGroup: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    isInheritedAllAuth: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    accessRuleSubjectId: {
      type: String as PropType<string>,
      required: false,
    },
  },
  computed: {
    Action(): typeof Action {
      return Action;
    },
    ResourceType(): typeof ResourceType {
      return ResourceType;
    },
    disabledDeleteButton(): boolean {
      return this.isAdminTenantRole || this.isInheritedFromGroup || this.isInheritedAllAuth;
    },
    deleteButtonToolTip(): string {
      if (this.isAdminTenantRole) {
        return `This rule can't be deleted because it is your last system-wide administrator access rule.`;
      }
      if (this.isInheritedFromGroup) {
        return `This rule can't be deleted because it is inherited from group ${this.accessRuleSubjectId}. To delete it for the whole group, go to "Access rules."`;
      }
      if (this.isInheritedAllAuth) {
        return `This rule can't be deleted here because it applies to all authenticated users and service accounts. To change it, go to "Access rules."`;
      }
      return `Delete access rule`;
    },
  },
  methods: {
    isRequired(val: string | null): string | boolean {
      return !!val || errorMessages.UNSAVED_RULE;
    },
  },
});
</script>
