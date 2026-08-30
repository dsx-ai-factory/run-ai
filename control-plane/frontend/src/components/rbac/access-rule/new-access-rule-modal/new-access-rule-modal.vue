<template>
  <section class="new-assigment-modal">
    <runai-base-dialog :model-value="true" @close="closeDialog()" size="width-md" no-padding>
      <template #header>New Access Rule</template>
      <template #body>
        <div class="new-assigment-modal-body q-px-lg q-py-md column">
          <q-form ref="newAccessRuleForm">
            <span>To authorize access, select a subject and set their role and scope</span>

            <!--  SUBJECT TYPE-->
            <div class="row q-mt-lg">
              <runai-select
                class="col-6"
                label="Subject"
                aid="subject-select"
                :options="userTypeOptions"
                :model-value="selectedSubjectType"
                @update:model-value="selectedSubjectType = $event"
                :rules="[requiredSubject]"
              />
            </div>

            <!-- SUBJECT ID -->
            <div class="row q-mt-md" v-if="!isAllAuthenticatedUsers">
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
                class="col-6"
                label="Role"
                :options="roleOptions"
                :model-value="selectedRole"
                @update:model-value="selectedRole = $event"
                :rules="[requiredRoleType]"
              />
            </div>

            <!-- SCOPE -->
            <div class="row q-my-md">
              <runai-multi-scope-input
                dense-chip
                class="col-12 items-center dense-chip"
                required
                :selected-scopes="selectedScopes"
                @clear="selectedScopes = []"
                @update:selected-scopes="onScopesSelected"
                counter
                :maxlength="isAllAuthenticatedUsers ? 1 : scopeMaxAmount"
                :hint="scopeHint"
              />
            </div>
          </q-form>
        </div>
      </template>

      <template #footer>
        <runai-form-footer
          save-btn-label="save rule"
          @cancel="$emit('close')"
          @save="onSave"
          :loading="loading"
          :is-form-valid="isFormValid"
        />
      </template>
    </runai-base-dialog>
  </section>
</template>

<script lang="ts">
import { defineComponent, ref, type PropType } from "vue";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { errorMessages } from "@/common/error-message.constant";
import type { ISelectOption } from "@/models/global.model";
import {
  type ISubjectTypeOption,
  SCOPE_MAX_AMOUNT,
  subjectTypeOptions,
  subjectTypeOptionsWithSSO,
} from "@/models/access-rule.model";
import type { SubjectType } from "@/swagger-models/authorization-client";
import { SubjectTypeExtended, ScopeType } from "@/swagger-models/authorization-client";
import { type AccessRulesBatchCreateResponse } from "@/swagger-models/authorization-client";
import type {
  AccessRulesBatchCreateFields,
  AuthenticatedUsersAccessRuleCreationFields,
  RoleV1,
} from "@/swagger-models/authorization-client/api";
import type { IOrgTreeNodeId } from "@/models/org-tree.model";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import { RunaiSelect } from "@/components/common/runai-select";
import { RunaiMultiScopeInput } from "@/components/common/runai-multi-scope-input";
import { SubjectAddableMultiSelect } from "@/components/rbac/common/subject-addable-multi-select";

import { type QForm } from "quasar";
import { RunaiFormFooter } from "@/components/common/runai-form-footer";
import { ErrorAlert } from "@/utils/error-alert.util";

export default defineComponent({
  name: "new-access-rule-modal",
  components: {
    RunaiFormFooter,
    SubjectAddableMultiSelect,
    RunaiMultiScopeInput,
    RunaiBaseDialog,
    RunaiSelect,
  },
  props: {
    isSso: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    roles: {
      type: Array as PropType<RoleV1[]>,
      required: true,
    },
  },
  data() {
    return {
      selectedSubjectType: subjectTypeOptions[0] as ISelectOption | null,
      selectedRole: null as ISelectOption | null,
      selectedScopes: [] as IOrgTreeNodeId[],
      selectedSubjectIds: [] as ISubjectTypeOption[],
      loading: false,
      isFormValid: true,
      newAccessRuleForm: ref<QForm | null>(null),
      scopeMaxAmount: SCOPE_MAX_AMOUNT,
    };
  },
  computed: {
    userTypeOptions(): ISubjectTypeOption[] {
      return this.isSso ? subjectTypeOptionsWithSSO : subjectTypeOptions;
    },
    isAllAuthenticatedUsers(): boolean {
      return this.selectedSubjectType?.value === SubjectTypeExtended.AllAuthenticatedUsers;
    },
    scopeHint(): string {
      return this.isAllAuthenticatedUsers ? "Select a single scope" : `Select up to ${SCOPE_MAX_AMOUNT} scopes`;
    },
    roleOptions(): ISelectOption[] {
      return this.roles.map((role) => ({
        label: role.name,
        value: role.id,
        roleStageBadge: role.name,
      }));
    },
  },
  methods: {
    async validate(): Promise<boolean> {
      const isValid = await (this.$refs["newAccessRuleForm"] as QForm).validate();
      this.isFormValid = isValid;
      return isValid;
    },
    onScopesSelected(scopes: IOrgTreeNodeId[]): void {
      // all-authenticated-users maps to a single-scope endpoint; keep only the most recent pick.
      this.selectedScopes = this.isAllAuthenticatedUsers ? scopes.slice(-1) : scopes;
    },
    addSubject(option: ISubjectTypeOption): void {
      if (!option.value) return;
      this.selectedSubjectIds.push(option);
    },
    async onSave(): Promise<void> {
      this.isFormValid = true;
      const isValid = await this.validate();

      if (!isValid || !this.selectedScopes.length) return;

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
        this.closeDialog(createdAccessRules);
      } catch (error: unknown) {
        this.notifyError(error);
      } finally {
        this.loading = false;
      }
    },
    prepareAuthenticatedUsersCreateFields(): AuthenticatedUsersAccessRuleCreationFields {
      const [scope] = this.selectedScopes;
      // system scope is excluded in the picker; guard at runtime since the API forbids it.
      if (scope.type === ScopeType.System) {
        throw new Error("System scope is not supported for all authenticated users and service accounts");
      }
      return {
        roleId: this.selectedRole?.value as number,
        scopeId: scope.id,
        scopeType: scope.type as AuthenticatedUsersAccessRuleCreationFields["scopeType"],
      };
    },
    prepareAccessRulesCreateFields(): AccessRulesBatchCreateFields {
      return {
        payload: {
          roleId: this.selectedRole?.value as number,
          subjects: this.selectedSubjectIds.map((subjectId) => ({
            subjectId: subjectId.value,
            subjectType: this.selectedSubjectType?.value as SubjectType,
          })),
          scopes: this.selectedScopes.map((scope) => ({ scopeId: scope.id, scopeType: scope.type })),
        },
      };
    },
    closeDialog(createdAccessRules?: AccessRulesBatchCreateResponse) {
      this.$emit("close", createdAccessRules);
    },
    notifyError(error: unknown) {
      const errorAlert = new ErrorAlert({
        generalMessage: ErrorAlert.failedCreateMessage("access rule"),
      });
      this.$q.notify(errorAlert.getNotification(error));
    },
    requiredSubject(val: ISelectOption): boolean | string {
      return !!val || errorMessages.SUBJECT_NOT_EMPTY;
    },
    requiredRoleType(val: ISelectOption): boolean | string {
      return !!val || errorMessages.ROLE_TYPE_NOT_EMPTY;
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

<style lang="scss">
.new-assigment-modal-body {
  .column {
    flex-wrap: unset;
  }
  .q-field--standard.q-field--readonly .q-field__control:before {
    border-bottom-style: solid;
  }
}
</style>
<style lang="scss" scoped>
.new-assigment-modal-body {
  .tree-button-wrapper {
    align-self: center;
    margin-left: 10px;
  }
  .tree-button {
    width: 40px;
    height: 40px;
  }
  .x-button {
    margin-top: 10px;
    color: $black-54;
    width: 40px;
    height: 40px;
  }
  .user-auto-complete {
    width: 300px;
  }
  .error {
    color: $negative;
  }
}
</style>
