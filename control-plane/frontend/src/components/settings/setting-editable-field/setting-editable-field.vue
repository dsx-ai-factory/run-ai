<template>
  <div class="setting-editable-field row justify-between no-wrap items-center q-pa-sm q-mt-md">
    <span class="ellipsis col-6" :class="{ 'text-subtitle1': useLargerLabel }">{{ label }}</span>
    <div class="buttons row no-wrap q-gutter-sm" v-if="!hideButtons">
      <slot name="start"></slot>
      <runai-btn
        :aid="`edit-${label}`"
        :icon="editReadIcon"
        class="q-pa-sm icon runai-icon-btn--default"
        flat
        size="sm"
        round
        @click="editClicked"
        :tooltip="toolTipMessage"
        color="black-54"
      />
      <runai-btn
        v-if="!isReadOnly"
        :aid="`delete-${label}`"
        icon="ph-bold ph-x"
        class="q-pa-sm icon runai-icon-btn--default"
        flat
        size="sm"
        rounded
        @click="isDeleteModalOpen = true"
        :disable="deleteDisabled"
        :tooltip="deleteToolTip"
        color="black-54"
      />
      <slot name="end"></slot>
    </div>
  </div>
  <setting-delete-modal
    v-if="isDeleteModalOpen"
    :delete-confirm-message="deleteToolTip || ''"
    :delete-confirm-message-subtitle="deleteSubTitle"
    @cancel="isDeleteModalOpen = false"
    @delete="deleteSetting"
  />
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import SettingDeleteModal from "@/components/settings/setting-delete-modal/setting-delete-modal.vue";
import { usePermissionStore } from "@/stores/permissions.store";
import { ResourceType } from "@/swagger-models/authorization-client";
import { RunaiBtn } from "@/components/common/runai-btn";

export default defineComponent({
  name: "setting-editable-field",
  components: { SettingDeleteModal, RunaiBtn },
  emits: ["edit", "delete"],
  props: {
    label: {
      type: String as PropType<string>,
      required: true,
    },
    editToolTip: {
      type: String as PropType<string>,
    },
    deleteDisabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    deleteToolTip: {
      type: String as PropType<string>,
    },
    deleteSubTitle: {
      type: String as PropType<string>,
      default: "",
    },
    hideButtons: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    useLargerLabel: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    resourceType: {
      type: String as PropType<keyof typeof ResourceType>,
      default: ResourceType.Settings,
    },
  },
  data() {
    return {
      isDeleteModalOpen: false as boolean,
      isBoxOpen: false as boolean,
    };
  },
  computed: {
    isReadOnly(): boolean {
      return usePermissionStore().isReadOnly(ResourceType[this.resourceType]);
    },
    editReadIcon(): string {
      return this.isReadOnly ? "ph-fill ph-eye" : "ph-bold ph-pencil-simple";
    },
    toolTipMessage(): string {
      return this.isReadOnly ? this.editToolTip?.replace("Edit", "View") || "View" : this.editToolTip || "";
    },
  },
  methods: {
    editClicked(): void {
      this.$emit("edit");
    },
    deleteSetting(): void {
      this.isDeleteModalOpen = false;
      this.$emit("delete");
    },
  },
});
</script>
<style lang="scss" scoped>
.setting-editable-field {
  border: 1px solid $black-12;
}
.icon {
  color: $black-54;

  :deep(.q-icon) {
    font-size: 20px;
  }
}
</style>
