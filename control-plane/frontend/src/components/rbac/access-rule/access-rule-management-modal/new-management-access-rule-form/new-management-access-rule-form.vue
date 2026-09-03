<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-card flat bordered>
    <q-card-section>
      <q-form ref="accessRuleForm">
        <!--  SUBJECT TYPE-->
        <div class="row">
          <runai-select
            v-if="displayConfig.subject"
            class="col-5"
            label="Subject"
            aid="subject-select"
            :options="userTypeOptions"
            :model-value="selectedSubjectType"
            @update:model-value="selectedSubjectType = $event"
            :rules="[requiredSubject]"
          />
        </div>
        <!-- SUBJECT ID -->
        <div class="row" v-if="displayConfig.users && selectedSubjectType && !isAllAuthenticatedUsers">
          <subject-addable-multi-select
            class="col-12"
            :subject-type="selectedSubjectType"
            :selected-subjects-options="selectedSubjectIds"
            @update:selected-subjects="selectedSubjectIds = $event"
            @add-subject="addSubject"
          />
        </div>

        <!-- ROLE -->
        <div class="row q-mt-md">
          <runai-select
            class="col-5"
            label="Role"
            :options="roleOptions"
            :model-value="selectedRole"
            @update:model-value="selectedRole = $event"
            :rules="[requiredRoleType]"
          />
        </div>
        <!-- SCOPE -->
        <div class="row q-my-md" v-if="displayConfig.scope">
          <runai-multi-scope-input
            class="col-12 items-center dense-chip"
            required
            :selected-scopes="selectedScopes"
            @clear="selectedScopes = []"
            @update:selected-scopes="onScopesSelected"
            counter
            :maxlength="isAllAuthenticatedUsers ? 1 : scopeMaxAmount"
          />
        </div>
      </q-form>
    </q-card-section>
    <q-separator />
    <q-card-actions>
      <runai-form-footer
        save-btn-label="save rule"
        @cancel="$emit('close')"
        @save="onSave"
        :display-error="displayError"
        :is-form-valid="isFormValid"
        save-btn-flat
      />
    </q-card-actions>
  </q-card>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiSelect } from "@/components/common/runai-select";
import {
  ERoleName,
  type IAccessRuleDisplayConfig,
  type IAccessRuleItem,
  type IAccessRuleManagementFormOptions,
  type ISubjectTypeOption,
  SCOPE_MAX_AMOUNT,
  subjectTypeOptions,
  subjectTypeOptionsWithSSO,
} from "@/models/access-rule.model";
import type { ISelectOption } from "@/models/global.model";
import { useAuthStore } from "@/stores/auth.store";
import type { SubjectType } from "@/swagger-models/authorization-client";
import {
  SubjectTypeExtended,
  type AccessRulesBatchCreateFields,
  type AccessRuleScopesInner,
  type AccessRuleSubjectsInner,
  type AuthenticatedUsersAccessRuleCreationFields,
  ScopeType,
} from "@/swagger-models/authorization-client";
import type { IOrgTreeNodeId } from "@/models/org-tree.model";
import { errorMessages } from "@/common/error-message.constant";
import { SubjectAddableMultiSelect } from "@/components/rbac/common/subject-addable-multi-select";
import { RunaiMultiScopeInput } from "@/components/common/runai-multi-scope-input";
import { RunaiFormFooter } from "@/components/common/runai-form-footer";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import { ErrorAlert } from "@/utils/error-alert.util";

export default defineComponent({
  name: "new-management-access-rule-form",
  components: {
    RunaiFormFooter,
    RunaiMultiScopeInput,
    SubjectAddableMultiSelect,
    RunaiSelect,
  },
  emits: ["close", "delete-rule", "save"],
  props: {
    displayConfig: {
      type: Object as PropType<IAccessRuleDisplayConfig>,
      required: true,
    },
    subjectOptions: {
      type: Array as PropType<ISelectOption[]>,
      default: () => [],
    },
    roleOptions: {
      type: Array as PropType<ISelectOption[]>,
      default: () => [],
    },
    isInheritedFromGroup: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    isSso: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    options: {
      type: Object as PropType<IAccessRuleManagementFormOptions>,
      required: true,
    },
  },
  data() {
    return {
      authStore: useAuthStore(),
      selectedSubjectType: null as ISelectOption | null,
      selectedSubjectIds: [] as ISubjectTypeOption[],
      selectedScopes: [] as IOrgTreeNodeId[],
      selectedRole: null as ISelectOption | null,
      accessRuleItem: {
        subjectTypeOption: null as ISelectOption | null,
        subjectId: "",
        roleOption: null as ISelectOption | null,
        scopeInput: null as IOrgTreeNodeId | null,
        saved: false,
      } as IAccessRuleItem,
      scopeMaxAmount: SCOPE_MAX_AMOUNT,
      isFormValid: true as boolean,
      loading: false as boolean,
    };
  },
  computed: {
    displayError(): boolean {
      return !this.loading && !this.accessRuleItem.saved;
    },
    userTypeOptions(): ISubjectTypeOption[] {
      return this.isSso ? subjectTypeOptionsWithSSO : subjectTypeOptions;
    },
    isAllAuthenticatedUsers(): boolean {
      return this.selectedSubjectType?.value === SubjectTypeExtended.AllAuthenticatedUsers;
    },
    isAdminTenantRole(): boolean {
      return (
        this.authStore.currentUser.username === this.accessRuleItem.subjectId &&
        this.accessRuleItem.roleName === ERoleName.SystemAdmin &&
        this.accessRuleItem.scopeType === ScopeType.Tenant
      );
    },
    isDirty(): boolean {
      return (
        this.selectedSubjectType !== null ||
        this.selectedSubjectIds.length > 0 ||
        this.selectedRole !== null ||
        this.selectedScopes.length > 0
      );
    },
  },
  methods: {
    onScopesSelected(scopes: IOrgTreeNodeId[]): void {
      // all-authenticated-users maps to a single-scope endpoint; keep only the most recent pick.
      this.selectedScopes = this.isAllAuthenticatedUsers ? scopes.slice(-1) : scopes;
    },
    addSubject(option: ISubjectTypeOption) {
      if (!option.value) return;
      this.selectedSubjectIds.push(option);
    },
    async validate(): Promise<boolean> {
      const form = this.$refs["accessRuleForm"] as HTMLFormElement;
      return await (form as HTMLFormElement).validate();
    },
    requiredRoleType(val: ISelectOption): boolean | string {
      return !!val || errorMessages.ROLE_TYPE_NOT_EMPTY;
    },
    requiredSubject(val: ISelectOption): boolean | string {
      return !!val || errorMessages.SUBJECT_NOT_EMPTY;
    },
    requiredSubjectId(val: string): boolean | string {
      return !!val || errorMessages.SUBJECT_ID_NOT_EMPTY;
    },
    getScopes(): AccessRuleScopesInner[] {
      if (this.options.scopeId && this.options.scopeType) {
        return this.getScopePageScopes();
      }
      return this.selectedScopes.map((scope) => ({ scopeId: scope.id, scopeType: scope.type }));
    },
    getScopePageScopes(): AccessRuleScopesInner[] {
      return [{ scopeId: this.options.scopeId as string, scopeType: this.options.scopeType as ScopeType }];
    },
    getSubjects(): AccessRuleSubjectsInner[] {
      if (this.options.scopeId && this.options.scopeType) {
        return this.getScopePageSubjects();
      }
      return this.getSubjectPageSubjects();
    },
    getScopePageSubjects(): AccessRuleSubjectsInner[] {
      return this.selectedSubjectIds.map((subjectId) => ({
        subjectId: subjectId.value,
        subjectType: this.selectedSubjectType?.value as SubjectType,
      }));
    },
    getSubjectPageSubjects(): AccessRuleSubjectsInner[] {
      return [
        {
          subjectId: this.options.subjectId as string,
          subjectType: this.options.subjectType as SubjectType,
        },
      ];
    },
    prepareAccessRulesCreateFields(): AccessRulesBatchCreateFields {
      return {
        payload: {
          roleId: this.selectedRole?.value as number,
          scopes: this.getScopes(),
          subjects: this.getSubjects(),
        },
      };
    },
    prepareAuthenticatedUsersCreateFields(): AuthenticatedUsersAccessRuleCreationFields {
      const [scope] = this.getScopes();
      // system scope is excluded in the picker; guard at runtime since the API forbids it.
      if (scope.scopeType === ScopeType.System) {
        throw new Error("System scope is not supported for all authenticated users and service accounts");
      }
      return {
        roleId: this.selectedRole?.value as number,
        scopeId: scope.scopeId,
        scopeType: scope.scopeType as AuthenticatedUsersAccessRuleCreationFields["scopeType"],
      };
    },
    notifyCreateError(error: unknown) {
      const errorAlert = new ErrorAlert({
        generalMessage: ErrorAlert.failedCreateMessage("access rules"),
      });
      this.$q.notify(errorAlert.getNotification(error));
    },
    async onSave(): Promise<boolean> {
      this.isFormValid = await this.validate();

      if (!this.isFormValid) return false;

      // Guard the single-scope path: a misconfigured displayConfig could reach here with no scope.
      if (this.isAllAuthenticatedUsers && !this.getScopes().length) return false;

      try {
        this.loading = true;
        const createdAccessRules = this.isAllAuthenticatedUsers
          ? {
              succeeded: [
                await accessRuleService.createAuthenticatedUsersAccessRule(this.prepareAuthenticatedUsersCreateFields()),
              ],
              failed: [],
            }
          : await accessRuleService.accessRulesBatchCreate(this.prepareAccessRulesCreateFields());
        this.$emit("save", createdAccessRules);
        return true;
      } catch (error: unknown) {
        this.notifyCreateError(error);
        return false;
      } finally {
        this.loading = false;
      }
    },
    deleteRule(): void {
      this.$emit("delete-rule");
    },
    async submit(): Promise<void> {
      const success = await this.onSave();
      if (!success) {
        throw new Error("Access rule save failed");
      }
    },
  },
  watch: {
    isAllAuthenticatedUsers(isAll: boolean): void {
      // Single-scope endpoint: clear a multi-scope selection on switch rather than silently keep an arbitrary one.
      if (isAll && this.selectedScopes.length > 1) this.selectedScopes = [];
    },
  },
});
</script>

<style scoped lang="scss"></style>
