<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-base-dialog :model-value="true" class="connection-access-modal" size="width-md" @close="$emit('close')">
    <template #header>
      <div class="connection-access-modal-header">Access to {{ toolName }}</div>
    </template>

    <template #body>
      <main class="connection-access-modal-body">
        <div class="q-my-sm">Set who is authorized to access the {{ connectionType }}</div>
        <q-form :ref="formRef" @submit.prevent>
          <div class="options-container q-mb-md">
            <runai-radio-options
              :options="accessOptions"
              :model-value="selectedAccess"
              @update:model-value="onSelectedOptionChanged"
              :policy-rules="radioOptionsRules"
            />
          </div>

          <div class="users-list" v-if="selectedAccess === 'specific-users'">
            <runai-section-title
              :title="`Add the users and service accounts who can access the ${connectionType}`"
              :tooltip-text="`Enter a valid username or email address. If you remove yourself, you will lose access to the ${connectionType}.`"
            />
            <subject-addable-multi-select
              :subject-type="userSubjectType"
              :selected-subjects-options="selectedUserOptions"
              @update:selected-subjects="onUserOptionsChanged"
              @add-subject="onAddUserOption"
            />
          </div>

          <div class="groups-list" v-if="selectedAccess === 'groups'">
            <runai-section-title
              :title="`Add the groups that can access the ${connectionType}`"
              :tooltip-text="`Enter group names as they appear in your identity provider. You must be a member of one of the groups listed to have access to the ${connectionType}.`"
            />
            <access-input-list
              input-label="Group name"
              add-item-btn-label="Group"
              :items="groups"
              :validation-rules="[groupNotEmpty, isNameExist]"
              @changed="onGroupsChanged"
              :policy-rules="policyRules?.authorizedGroups"
            ></access-input-list>
          </div>
        </q-form>
      </main>
    </template>

    <template #footer>
      <q-btn label="Cancel" color="primary" flat @click="$emit('close')" aid="cancel-connection-access-role" />
      <q-btn label="Save" color="primary" @click="onSave" aid="save-connection-access-role" />
    </template>
  </runai-base-dialog>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// cmps
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { AccessInputList } from "@/components/environment/connection-access/access-input-list";
import { RunaiSectionTitle } from "@/components/common/runai-section-title";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { SubjectAddableMultiSelect } from "@/components/rbac/common/subject-addable-multi-select";

// models
import { type IConnectionAccess, EToolLabel } from "./connection-access-modal.model";
import type { IAccessInput } from "@/components/environment/connection-access/access-input-list";
import { EAccessOptions } from "./connection-access-modal.model";
import type { ISelectOption } from "@/models/global.model";
import {
  ServingPortAccessAuthorizationTypeEnum,
  type SupersetServingPortRules,
} from "@/swagger-models/policy-service-client";
import type { IGenericRadioOptionsPolicyRules } from "@/models/policy.model";
import { SubjectType } from "@/swagger-models/authorization-client";

// utils
import { convertNamesToItems } from "./connection-access-modal.util";

import { errorMessages } from "@/common/error-message.constant";
import type { StringOption, StringRules } from "@/swagger-models/assets-service-client";

export default defineComponent({
  name: "connection-access-modal",
  components: {
    RunaiBaseDialog,
    AccessInputList,
    RunaiSectionTitle,
    RunaiRadioOptions,
    SubjectAddableMultiSelect,
  },
  emits: ["save", "close"],
  props: {
    toolName: {
      type: String as PropType<string>,
      required: true,
    },
    access: {
      type: Object as PropType<IConnectionAccess>,
      required: true,
      validator: (access: IConnectionAccess) => {
        if (!Object.values(EAccessOptions).includes(access.accessOption)) {
          throw new Error(
            `Invalid prop: "access" => accessOption: "${access.accessOption}". 
            Expected one of: ${Object.values(EAccessOptions).join(", ")}`,
          );
        }
        return true;
      },
    },
    defaultUser: {
      type: String as PropType<string>,
      required: true,
    },
    multiUsers: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    enablePublicAccess: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    connectionType: {
      type: String as PropType<"tool" | "endpoint">,
      required: true,
    },
    policyRules: {
      type: [Object, null] as PropType<SupersetServingPortRules | null>,
      required: false,
    },
  },
  data() {
    return {
      selectedAccess: EAccessOptions.EVERYONE as string,
      accessOptions: [] as ISelectOption[],
      users: undefined as IAccessInput[] | undefined,
      selectedUserOptions: [] as ISelectOption[],
      groups: undefined as IAccessInput[] | undefined,
      formRef: "connection-access-modal-form-el" as string,
      radioOptionsRules: undefined as IGenericRadioOptionsPolicyRules | undefined,
      userSubjectType: { label: "User", value: SubjectType.User } as ISelectOption,
    };
  },
  created() {
    this.initPolicyRules();
    this.buildAccessOptions();

    // Prioritize accessOption if it exists - it represents the user's explicit choice
    if (this.access.accessOption) {
      this.selectedAccess = this.access.accessOption;
      // Initialize the appropriate list based on the selected option
      switch (this.access.accessOption) {
        case EAccessOptions.GROUPS:
          this.onGroupsSelected();
          break;
        case EAccessOptions.SPECIFIC_USERS:
          this.onUsersSelected();
          break;
        case EAccessOptions.PRIVATE:
          this.onPrivateSelected();
          break;
        // For EVERYONE or PUBLIC, keep users and groups as undefined
      }
    } else if (this.access.authorizedGroups) {
      // Fallback: if no explicit option, determine from data
      this.onGroupsSelected();
    } else if (this.access.authorizedUsers) {
      this.multiUsers ? this.onUsersSelected() : this.onPrivateSelected();
    }
  },
  methods: {
    onSelectedOptionChanged(selected: string): void {
      this.selectedAccess = selected;
      switch (this.selectedAccess) {
        case EAccessOptions.GROUPS:
          this.onGroupsSelected();
          break;
        case EAccessOptions.SPECIFIC_USERS:
          this.onUsersSelected();
          break;
        case EAccessOptions.PRIVATE:
          this.onPrivateSelected();
          break;
        default:
          this.groups = undefined;
          this.users = undefined;
          this.selectedUserOptions = [];
          break;
      }
    },
    onGroupsSelected(): void {
      this.selectedAccess = EAccessOptions.GROUPS;
      const authorizedGroups = this.access.authorizedGroups || [];
      this.groups = convertNamesToItems(authorizedGroups, EToolLabel.GroupName);
      this.users = undefined;
      this.selectedUserOptions = [];
    },
    onUsersSelected(): void {
      this.selectedAccess = EAccessOptions.SPECIFIC_USERS;
      const authorizedUsers = this.access.authorizedUsers || [this.defaultUser];
      this.selectedUserOptions = authorizedUsers.map((email) => ({ label: email, value: email }));
      this.users = undefined;
      this.groups = undefined;
    },
    onPrivateSelected(): void {
      this.selectedAccess = EAccessOptions.PRIVATE;
      this.groups = undefined;
      this.users = [this.defaultUser].map((user) => ({ text: user, label: "", removable: false }));
    },
    onGroupsChanged(groups: IAccessInput[]): void {
      this.groups = groups;
      this.selectedUserOptions = [];
    },
    onUserOptionsChanged(options: ISelectOption[]): void {
      this.selectedUserOptions = options;
      this.groups = undefined;
    },
    onAddUserOption(option: ISelectOption): void {
      if (option.value && !this.selectedUserOptions.some((o) => o.value === option.value)) {
        this.selectedUserOptions.push(option);
      }
    },
    async isSelectionIsValid(): Promise<boolean> {
      const formEl = this.$refs[this.formRef] as HTMLFormElement;
      return await formEl?.validate();
    },
    async onSave(): Promise<void> {
      const isAllowedToSave = await this.isSelectionIsValid();
      if (!isAllowedToSave) return;

      const authorizedUsers =
        this.selectedAccess === EAccessOptions.SPECIFIC_USERS
          ? this.selectedUserOptions.map((opt) => opt.value as string)
          : this.users?.map((user) => user.text);

      this.$emit("save", {
        authorizedUsers,
        authorizedGroups: this.groups?.map((group) => group.text),
        accessOption: this.selectedAccess,
      });
    },
    groupNotEmpty(val: string): boolean | string {
      return val?.length ? true : errorMessages.NAME_NOT_EMPTY;
    },
    isNameExist(val: string): boolean | string {
      if (!this.groups?.length) return true;
      const isExist = this.groups.filter((group) => group.text === val);
      return isExist.length <= 1 || errorMessages.NAME_ALREADY_EXIST;
    },
    initPolicyRules(): void {
      if (!this.policyRules?.authorizationType) return;

      this.radioOptionsRules = {
        ...this.policyRules?.authorizationType,
        options: this.getOptionsRules(this.policyRules?.authorizationType),
      };
    },
    findAuthorizationType(
      authorizationTypeOptions: Array<StringOption>,
      authorizationType: ServingPortAccessAuthorizationTypeEnum,
    ): StringOption | undefined {
      return authorizationTypeOptions.find((option) => option.value === authorizationType);
    },
    getOptionsRules(
      authorizationType: StringRules | undefined,
    ): Array<{ displayed: string | null | undefined; value: string }> | undefined {
      if (!authorizationType?.options) return;

      const options: Array<{ displayed: string | null | undefined; value: string }> = [];

      const publicOption = this.findAuthorizationType(
        authorizationType.options,
        ServingPortAccessAuthorizationTypeEnum.Public,
      );

      if (publicOption) {
        options.push({
          displayed: publicOption.displayed,
          value: EAccessOptions.PUBLIC,
        });
      }

      const authUsersOption = this.findAuthorizationType(
        authorizationType.options,
        ServingPortAccessAuthorizationTypeEnum.AuthenticatedUsers,
      );

      if (authUsersOption) {
        options.push({
          displayed: authUsersOption.displayed,
          value: EAccessOptions.EVERYONE,
        });
      }

      const authUsersOrGroupsOption = this.findAuthorizationType(
        authorizationType.options,
        ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups,
      );

      if (authUsersOrGroupsOption) {
        options.push({
          displayed: authUsersOrGroupsOption.displayed,
          value: EAccessOptions.SPECIFIC_USERS,
        });

        options.push({
          displayed: authUsersOrGroupsOption.displayed,
          value: EAccessOptions.GROUPS,
        });
      }

      return options;
    },
    buildAccessOptions(): void {
      // Start with base options
      this.accessOptions = [
        {
          label: `All authenticated users and service accounts`,
          value: EAccessOptions.EVERYONE,
          aid: EAccessOptions.EVERYONE,
        },
      ];

      // Add Public option if enabled
      if (this.enablePublicAccess) {
        this.accessOptions.unshift({
          label: "Public",
          value: EAccessOptions.PUBLIC,
          aid: EAccessOptions.PUBLIC,
        });
      }

      // Add user/group options
      if (this.multiUsers) {
        this.accessOptions.push({
          label: "Specific group(s)",
          value: EAccessOptions.GROUPS,
          toolTip: "Groups from your identity provider",
          aid: EAccessOptions.GROUPS,
        });
        this.accessOptions.push({
          label: "Specific users and service accounts",
          value: EAccessOptions.SPECIFIC_USERS,
          aid: EAccessOptions.SPECIFIC_USERS,
        });
      } else {
        this.accessOptions.push({
          label: "Private",
          value: EAccessOptions.PRIVATE,
          toolTip: "When selected, only the workload creator will have access to the tool",
          aid: EAccessOptions.PRIVATE,
        });
      }

      // Apply policy constraints
      this.applyPolicyConstraints();
    },
    applyPolicyConstraints(): void {
      const authorizedUsersRule = this.policyRules?.authorizedUsers;
      const authorizedGroupsRule = this.policyRules?.authorizedGroups;

      // Check constraints for users and groups
      const usersCanEdit = authorizedUsersRule?.canEdit !== false;
      const usersRequired = authorizedUsersRule?.required === true;
      const groupsCanEdit = authorizedGroupsRule?.canEdit !== false;
      const groupsRequired = authorizedGroupsRule?.required === true;

      // Determine which options should be available based on policy rules
      const allowedOptions = new Set<string>();

      // If users are required or can't be edited (locked to users only)
      if (usersRequired || !usersCanEdit) {
        allowedOptions.add(EAccessOptions.SPECIFIC_USERS);
      }

      // If groups are required or can't be edited (locked to groups only)
      if (groupsRequired || !groupsCanEdit) {
        allowedOptions.add(EAccessOptions.GROUPS);
      }

      // If no specific constraints, allow all options
      if (allowedOptions.size === 0) {
        return; // No policy constraints to apply
      }

      // Apply constraints to disable non-allowed options
      this.accessOptions = this.accessOptions.map((option) => {
        const optionValue = String(option.value); // Convert to string for type safety
        if (allowedOptions.has(optionValue)) {
          // This option is allowed by policy
          return option;
        } else {
          // Disable this option with policy tooltip
          return {
            ...option,
            disable: true,
            toolTip: "Can't be modified due to policy restrictions",
          };
        }
      });

      // Set default selection based on policy requirements
      if (usersRequired && allowedOptions.has(EAccessOptions.SPECIFIC_USERS)) {
        this.selectedAccess = EAccessOptions.SPECIFIC_USERS;
      } else if (groupsRequired && allowedOptions.has(EAccessOptions.GROUPS)) {
        this.selectedAccess = EAccessOptions.GROUPS;
      } else if (allowedOptions.size === 1) {
        // Only one option available, select it
        const firstOption = Array.from(allowedOptions)[0];
        if (firstOption) {
          this.selectedAccess = firstOption;
        }
      }
    },
  },
});
</script>

<style lang="scss">
.connection-access-modal {
  .access-option {
    width: 100%;
  }
}
</style>
