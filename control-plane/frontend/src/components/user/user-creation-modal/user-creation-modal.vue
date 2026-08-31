<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-base-dialog
    :model-value="modelValue"
    size="dynamic"
    :custom-class="isCreateMode ? 'user-creation-dialog' : 'reset-password-dialog'"
    no-padding
    @close="$emit('close')"
  >
    <template #header>{{ headerText }}</template>
    <template #body>
      <q-stepper
        active-icon="ph-fill ph-pencil-simple"
        class="user-review-stepper"
        v-model="currentStep"
        ref="stepper"
        color="primary"
        animated
        header-class="header-class"
      >
        <q-step :name="EStepName.Create" :title="firstStepTitle" prefix="1" :done="currentStep >= EStepName.Review">
          <section :style="isCreateMode ? 'min-width: 500px' : undefined">
            <q-form class="q-my-md" v-if="isCreateMode" ref="elNewAppForm" @submit.prevent="nextStepAction">
              <user-creation-screen
                v-model="newUser.email"
                :send-invitation="sendInvitation"
                :is-smtp-configured="isSmtpConfigured"
                :first-name="newUser.firstName || ''"
                :last-name="newUser.lastName || ''"
                @update:send-invitation="sendInvitation = $event"
                @update:first-name="newUser.firstName = $event"
                @update:last-name="newUser.lastName = $event"
              ></user-creation-screen>
            </q-form>
            <section v-else class="reset-password-content">
              <p class="text-b-h6">Reset password for {{ selectedUser?.username }}?</p>
              <p class="text-b-subtitle2">This will sign the user out of all active sessions</p>
            </section>
          </section>
        </q-step>
        <q-step :name="EStepName.Review" title="Review" prefix="2" :done="currentStep >= EStepName.AccessRules">
          <user-review-screen
            :action-type="actionType"
            :name="name"
            :password="password"
            :first-name="newUser.firstName || ''"
            :last-name="newUser.lastName || ''"
          ></user-review-screen>
        </q-step>
        <q-step
          v-if="showAccessRulesStep"
          :name="EStepName.AccessRules"
          title="Access rules"
          icon="runai-filled-access-rule"
          :done="false"
        >
          <access-rule-management
            ref="accessRuleManagementRef"
            :form-options="accessRuleFormOptions"
            auto-open-form
            hide-search
            no-inner-padding
          />
        </q-step>
      </q-stepper>
    </template>
    <template #footer-left>
      <q-field
        v-if="isAccessRulesStep"
        class="col-4 form-hint-translate-y-100"
        :model-value="displayFormHint"
        :rules="[isFormIncomplete]"
      />
    </template>
    <template #footer>
      <q-stepper-navigation class="user-review-action-buttons">
        <!-- Step 0: Cancel + Next/Create -->
        <template v-if="currentStep === EStepName.Create">
          <runai-btn class="cancel" flat color="primary" @click="close" label="Cancel" aid="stepper-navigation-cancel" />
          <runai-btn
            class="save"
            @click="nextStepAction"
            color="primary"
            :loading="loading"
            :label="nextBtnLabel"
            aid="stepper-navigation-save"
          />
        </template>
        <!-- Step 1 (Review) with access rules permission: Done (flat) + Add Access Rules (primary) -->
        <template v-else-if="currentStep === EStepName.Review && showAccessRulesStep">
          <runai-btn class="cancel" flat color="primary" @click="close" label="Done" aid="stepper-navigation-done" />
          <runai-btn
            class="save"
            @click="goToAccessRulesStep"
            color="primary"
            label="Add Access Rules"
            aid="stepper-navigation-access-rules"
          />
        </template>
        <!-- Step 1 (Review) without access rules permission: Done (primary) -->
        <template v-else-if="currentStep === EStepName.Review">
          <runai-btn class="save" @click="close" color="primary" label="Done" aid="stepper-navigation-save" />
        </template>
        <!-- Step 2 (Access rules): Close (primary) -->
        <template v-else-if="currentStep === EStepName.AccessRules">
          <runai-btn class="save" @click="close" color="primary" label="Close" aid="stepper-navigation-close" />
        </template>
      </q-stepper-navigation>
    </template>
  </runai-base-dialog>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { QStepper } from "quasar";
// cmps
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { UserCreationScreen } from "../user-creation-screen";
import { UserReviewScreen } from "../user-review-screen";
import { AccessRuleManagement } from "@/components/rbac/access-rule/access-rule-management";
// stores
import { useAppStore } from "@/stores/app.store";
import { usePermissionStore } from "@/stores/permissions.store";
// types
import { ActionTypeEnum, type TActionTypes } from "../action-type.model";
import type { IFilterBy } from "@/models/filter.model";
import type { User, UserCreationRequest, UserPostResponse } from "@/swagger-models/identity-manager-client";
import { Action, ResourceType, SubjectType } from "@/swagger-models/authorization-client";
import { EAccessRuleModalPage } from "@/models/access-rule.model";
import type { IAccessRuleManagementFormOptions } from "@/models/access-rule.model";
// utils
import { HttpErrorResponse, NotFoundError } from "@/models/http-response.model";
import { errorMessages } from "@/common/error-message.constant";
// services
import { alertUtil } from "@/utils/alert.util";
import userService from "@/services/control-plane/user.service/user.service";
import { RunaiBtn } from "@/components/common/runai-btn";
import { useNotificationStore } from "@/stores/notifications.store";

enum EStepName {
  Create = 0,
  Review = 1,
  AccessRules = 2,
}

export default defineComponent({
  components: { RunaiBtn, UserCreationScreen, UserReviewScreen, RunaiBaseDialog, AccessRuleManagement },
  emits: ["close", "load-users"],
  props: {
    actionType: {
      type: String as PropType<TActionTypes>,
      required: true,
    },
    modelValue: {
      type: Boolean,
      required: true,
    },
    selectedUser: {
      type: Object as PropType<User>,
      required: false,
    },
    filterBy: {
      type: Object as PropType<IFilterBy>,
      required: false,
    },
  },
  data() {
    return {
      name: "" as string,
      password: "" as string,
      currentStep: 0 as number,
      loading: false as boolean,
      appForm: null as HTMLFormElement | null,
      sendInvitation: false as boolean,
      isSmtpConfigured: false as boolean,
      displayFormHint: false as boolean,
      newUser: {
        email: "",
        firstName: "",
        lastName: "",
      } as UserCreationRequest,
    };
  },
  setup() {
    return { EStepName };
  },
  async created() {
    useAppStore().setPageLoading(false);
    await this.checkSmtpConfiguration();
  },
  mounted() {
    this.$nextTick(() => {
      this.appForm = this.$refs.elNewAppForm as HTMLFormElement;
    });
  },
  computed: {
    headerText(): string {
      return this.isCreateMode ? "New User" : "Reset Password";
    },
    isCreateMode(): boolean {
      return this.actionType === ActionTypeEnum.CREATE;
    },
    canManageAccessRules(): boolean {
      return usePermissionStore().hasPermission(ResourceType.AccessRules, Action.Create);
    },
    showAccessRulesStep(): boolean {
      return this.isCreateMode && this.canManageAccessRules;
    },
    isAccessRulesStep(): boolean {
      return this.currentStep === EStepName.AccessRules;
    },
    nextBtnLabel(): string {
      return this.isCreateMode
        ? [ActionTypeEnum.CREATE, ActionTypeEnum.DONE][this.currentStep]
        : [ActionTypeEnum.RESET, ActionTypeEnum.DONE][this.currentStep];
    },
    firstStepTitle(): string {
      return this.isCreateMode ? "Create" : "Reset";
    },
    accessRuleFormOptions(): IAccessRuleManagementFormOptions {
      return {
        page: EAccessRuleModalPage.User,
        subjectId: this.name,
        subjectType: SubjectType.User,
      };
    },
  },
  methods: {
    updateFormHint(): void {
      if (!this.isAccessRulesStep) {
        this.displayFormHint = false;
        return;
      }
      const accessRuleManagement = this.$refs.accessRuleManagementRef as
        | InstanceType<typeof AccessRuleManagement>
        | undefined;
      this.displayFormHint = !!(accessRuleManagement?.isFormOpen && accessRuleManagement?.isFormDirty);
    },
    isFormIncomplete(val: boolean): boolean | string {
      return !val ? true : errorMessages.FORM_INCOMPLETE;
    },
    async checkSmtpConfiguration(): Promise<void> {
      try {
        const notificationStore = useNotificationStore();
        await notificationStore.loadNotificationChannels();
        this.isSmtpConfigured = notificationStore.isEmailChannelAvailable;
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse && error.statusCode !== NotFoundError.statusCode) {
          console.warn("Failed to check SMTP configuration", error);
        }
        this.isSmtpConfigured = false;
      }
    },
    close(): void {
      this.$emit("close");
    },
    goToAccessRulesStep(): void {
      this.currentStep = EStepName.AccessRules;
    },
    async nextStepAction(): Promise<void> {
      if (this.currentStep === EStepName.Create) {
        const valid = this.appForm ? await this.validate() : true;
        if (valid) {
          this.isCreateMode ? await this.createUser() : await this.resetPassword();
        }
      }
    },
    async validate(): Promise<boolean> {
      return await (this.appForm as HTMLFormElement).validate();
    },
    async createUser(): Promise<void> {
      this.loading = true;
      const isUserExist = await userService.isUserExist(this.newUser.email);
      if (isUserExist) {
        this.$q.notify(alertUtil.getError(`User ${this.newUser.email} already exists.`));
        this.loading = false;
        return;
      }
      try {
        const newCreatedUser: UserPostResponse = await userService.create({
          email: this.newUser.email,
          notify: this.sendInvitation,
          firstName: this.newUser.firstName?.trim() || undefined,
          lastName: this.newUser.lastName?.trim() || undefined,
        });
        this.name = newCreatedUser.username;
        this.password = newCreatedUser.tempPassword;
        this.$emit("load-users", newCreatedUser);
        this.$q.notify(alertUtil.getSuccess(`User ${newCreatedUser.username} created`));
        (this.$refs.stepper as QStepper).next();
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.$q.notify(alertUtil.getError("Failed to create user"));
        }
      } finally {
        this.loading = false;
      }
    },
    async resetPassword(): Promise<void> {
      this.loading = true;
      try {
        if (this.selectedUser) {
          this.password = await userService.resetPassword(this.selectedUser.id);
          this.name = this.selectedUser.username;
          this.newUser.firstName = this.selectedUser.firstName || "";
          this.newUser.lastName = this.selectedUser.lastName || "";
          this.$emit("load-users");
          (this.$refs.stepper as QStepper).next();
        }
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.$q.notify(alertUtil.getError("Failed to reset user password"));
        }
      } finally {
        this.loading = false;
      }
    },
  },
  watch: {
    currentStep(): void {
      this.updateFormHint();
    },
  },
});
</script>
<style lang="scss">
.user-review-stepper {
  .header-class {
    display: flex;
    justify-content: center;

    div.col-grow {
      flex: unset;
    }
  }

  .q-stepper__tab .q-icon {
    font-size: var(--icon-size-small) !important;
  }
}

.user-review-action-buttons {
  padding-top: 5px;

  .save,
  .cancel {
    padding: 0 20px;
  }
}

.user-creation-dialog {
  width: 600px;
}

.reset-password-dialog {
  width: 500px;
}

.reset-password-content {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 5px 16px;
}
</style>
