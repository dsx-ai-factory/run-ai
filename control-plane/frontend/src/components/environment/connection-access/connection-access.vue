<template>
  <div class="connection-access row">
    <q-input
      :model-value="placeholderLabel"
      no-error-icon
      label="Access"
      stack-label
      :disable="true"
      class="flex-1 connection-access-label"
    >
      <template v-slot:prepend>
        <q-icon class="connection-access-icon q-mt-md" name="runai-access" size="24px" />
      </template>
    </q-input>
    <span>
      <q-btn
        class="edit-btn runai-icon-btn--default"
        round
        flat
        size="sm"
        icon="ph-bold ph-pencil-simple"
        @click="showModal = true"
        aid="edit-access-button"
        :disabled="disabled"
      ></q-btn>
      <q-tooltip v-if="accessBtnTooltip">{{ accessBtnTooltip }}</q-tooltip>
    </span>

    <connection-access-modal
      v-if="showModal"
      :default-user="userEmail"
      :tool-name="toolName"
      :access="access"
      :multi-users="isMultiUsers"
      @close="onClose"
      @save="saveAccess"
      :enable-public-access="enablePublicAccess"
      :connection-type="connectionType"
      :policy-rules="policyRules"
    />
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { ConnectionAccessModal } from "./connection-access-modal";
import { type IConnectionAccess, EAccessOptions } from "./connection-access-modal/connection-access-modal.model";

// stores
import { useAuthStore } from "@/stores/auth.store";

// models
import type { SupersetServingPortRules } from "@/swagger-models/policy-service-client";

export default defineComponent({
  name: "connection-access",
  components: {
    ConnectionAccessModal,
  },
  emits: ["save"],
  props: {
    access: {
      type: Object as PropType<IConnectionAccess>,
      required: true,
    },
    toolName: {
      type: String as PropType<string>,
      required: true,
    },
    isMultiUsers: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    enablePublicAccess: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    connectionType: {
      type: String as PropType<"tool" | "endpoint">,
      required: true,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    disabledTooltip: {
      type: String as PropType<string>,
      required: false,
    },
    policyRules: {
      type: [Object, null] as PropType<SupersetServingPortRules | null>,
      required: false,
    },
  },
  data() {
    return {
      authStore: useAuthStore(),
      showModal: false as boolean,
      accessRoles: {} as IConnectionAccess,
    };
  },
  created() {
    this.accessRoles.authorizedGroups = this.access.authorizedGroups;
    this.accessRoles.authorizedUsers = this.access.authorizedUsers;
    this.accessRoles.accessOption = this.access.accessOption;
  },
  computed: {
    userEmail(): string {
      return this.authStore.userEmail;
    },
    accessBtnTooltip(): string | undefined {
      if (!this.disabled) return `Click to change who can access this ${this.connectionType}`;
      if (this.disabled && this.disabledTooltip) return this.disabledTooltip;

      return undefined;
    },
    placeholderLabel(): string {
      if (this.access.accessOption === EAccessOptions.GROUPS && this.access.authorizedGroups && this.isMultiUsers) {
        return `Groups (${this.access?.authorizedGroups?.length || 0})`;
      } else if (this.access.accessOption === EAccessOptions.SPECIFIC_USERS && this.access.authorizedUsers) {
        return this.isMultiUsers ? `Users (${this.access?.authorizedUsers?.length || 0})` : `Private`;
      } else if (this.access.accessOption === EAccessOptions.PUBLIC && this.enablePublicAccess) {
        return "Public";
      } else {
        return "All authenticated users and service accounts";
      }
    },
  },
  methods: {
    onClose(): void {
      this.showModal = false;
    },
    saveAccess(access: IConnectionAccess): void {
      this.showModal = false;
      this.accessRoles = access;
      this.$emit("save", access);
    },
  },
});
</script>
<style lang="scss" scoped>
.connection-access {
  width: 100%;

  .connection-access-label {
    margin-bottom: 20px;

    :deep(.q-field__prepend) {
      opacity: 1;
    }
  }

  .connection-access-icon {
    color: #707070;
  }

  .edit-btn {
    color: $black-54;
    margin-top: 23px;
  }
}
</style>
