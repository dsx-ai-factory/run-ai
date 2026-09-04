<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-base-dialog
    :aid="`${entityType}-modal-dialog`"
    :model-value="modelValue"
    size="dynamic"
    :custom-class="dialogCustomClass"
    no-padding
    @close="$emit('close')"
  >
    <template #header>
      <span :aid="`${entityType}-creation-header-text`">{{ headerText }}</span>
    </template>
    <template #body>
      <q-stepper
        active-icon="ph-fill ph-pencil-simple"
        class="credential-stepper"
        v-model="currentStep"
        ref="stepper"
        color="primary"
        animated
        header-class="header-class"
      >
        <q-step :name="EStepName.Create" :title="firstStepTitle" icon="" :done="currentStep >= EStepName.Review">
          <section>
            <q-form
              :aid="`first-${entityType}-screen`"
              class="q-my-md"
              v-if="isCreateMode"
              ref="elNewCredentialForm"
              @submit.prevent="nextStepAction"
            >
              <service-accounts-creation-screen v-model="newCredential.name" :is-access-key="isAccessKey" />
            </q-form>
            <p v-else aid="first-regenerate-screen" class="q-my-lg text-b-h6">
              Regenerate the client secret for {{ credential?.name }}?
            </p>
          </section>
        </q-step>
        <q-step
          aid="review-screen"
          :name="EStepName.Review"
          title="Review"
          icon="ph-fill ph-clipboard-text"
          :done="currentStep >= EStepName.AccessRules"
        >
          <service-accounts-view-screen
            :is-create-mode="isCreateMode"
            :is-access-key="isAccessKey"
            :name="name"
            :secret="secret"
            :client="clientId"
          />
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
      <q-stepper-navigation>
        <!-- Step 0: Cancel + Next/Create -->
        <template v-if="currentStep === EStepName.Create">
          <q-btn aid="cancel-btn" flat color="primary" @click="close" label="Cancel" class="q-ml-sm" />
          <q-btn @click="nextStepAction" aid="next-btn" color="primary" :loading="loading" :label="nextBtnLabel" />
        </template>
        <!-- Step 1 (Review) with access rules permission: Done (flat) + Add Access Rules (primary) -->
        <template v-else-if="currentStep === EStepName.Review && showAccessRulesStep">
          <q-btn aid="done-btn" flat color="primary" @click="close" label="Done" class="q-ml-sm" />
          <q-btn @click="goToAccessRulesStep" aid="access-rules-btn" color="primary" label="Add Access Rules" />
        </template>
        <!-- Step 1 (Review) without access rules permission: Done (primary) -->
        <template v-else-if="currentStep === EStepName.Review">
          <q-btn @click="close" aid="next-btn" color="primary" label="Done" />
        </template>
        <!-- Step 2 (Access rules): Close (primary) -->
        <template v-else-if="currentStep === EStepName.AccessRules">
          <q-btn @click="close" aid="close-btn" color="primary" label="Close" />
        </template>
      </q-stepper-navigation>
    </template>
  </runai-base-dialog>
</template>

<script lang="ts">
// cmps
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { ServiceAccountsCreationScreen } from "@/components/service-accounts/service-accounts-creation-screen";
import { ServiceAccountsViewScreen } from "@/components/service-accounts/service-accounts-view-screen";
import { AccessRuleManagement } from "@/components/rbac/access-rule/access-rule-management";
// stores
import { useAppStore } from "@/stores/app.store";
import { usePermissionStore } from "@/stores/permissions.store";
// utils
import { HttpErrorResponse } from "@/models/http-response.model";
import { errorMessages } from "@/common/error-message.constant";
// services
import { alertUtil } from "@/utils/alert.util";
import type { QStepper } from "quasar";
import type { PropType } from "vue";
import { ActionTypeEnum } from "@/components/user/action-type.model";
import { Action, ResourceType, SubjectType } from "@/swagger-models/authorization-client";
import { EAccessRuleModalPage } from "@/models/access-rule.model";
import type { IAccessRuleManagementFormOptions } from "@/models/access-rule.model";
import type {
  AccessKey,
  ServiceAccount,
  AccessKeyCreationRequest,
  ServiceAccountCreationRequest,
  AccessKeyPostResponse,
  ServiceAccountPostResponse,
} from "@/swagger-models/identity-manager-client";
import { accessKeysService } from "@/services/control-plane/access-keys.service/access-keys.service";
import { serviceAccountsService } from "@/services/control-plane/service-accounts.service/service-accounts.service";
import { capitalizeString } from "@/utils/string.util/string.util";
import { ErrorAlert } from "@/utils/error-alert.util";

enum EStepName {
  Create = 0,
  Review = 1,
  AccessRules = 2,
}

export type TActionTypes = "create" | "regenerate";
export type TEntityType = "access-key" | "service-account";

type Credential = AccessKey | ServiceAccount;
type CredentialCreationRequest = AccessKeyCreationRequest | ServiceAccountCreationRequest;
type CredentialPostResponse = AccessKeyPostResponse | ServiceAccountPostResponse;

export default {
  components: { ServiceAccountsCreationScreen, ServiceAccountsViewScreen, RunaiBaseDialog, AccessRuleManagement },
  emits: ["close", "load-credential"],
  props: {
    entityType: {
      type: String as PropType<TEntityType>,
      required: true,
    },
    actionType: {
      type: String as PropType<TActionTypes>,
      required: true,
    },
    modelValue: {
      type: Boolean,
      required: true,
    },
    credential: {
      type: Object as PropType<Credential>,
      required: false,
    },
  },
  data() {
    return {
      credentialForm: null as HTMLFormElement | null,
      loading: false as boolean,
      currentStep: 0 as number,
      name: "" as string,
      secret: "" as string,
      clientId: "" as string,
      displayFormHint: false as boolean,
      newCredential: {
        name: "",
      } as CredentialCreationRequest,
      localActionType: this.actionType as TActionTypes,
    };
  },
  setup() {
    return { EStepName };
  },
  created() {
    useAppStore().setPageLoading(false);
  },
  mounted() {
    this.$nextTick(() => {
      this.credentialForm = this.$refs.elNewCredentialForm as HTMLFormElement;
    });
  },
  computed: {
    isAccessKey(): boolean {
      return this.entityType === "access-key";
    },
    entityDisplayName(): string {
      return this.isAccessKey ? "Access Key" : "Service Account";
    },
    entityDisplayNameLower(): string {
      return this.isAccessKey ? "access key" : "service account";
    },
    entityDisplayNameSentenceCase(): string {
      return capitalizeString(this.entityDisplayNameLower);
    },
    service() {
      return this.isAccessKey ? accessKeysService : serviceAccountsService;
    },
    headerText(): string {
      return this.isCreateMode ? `New ${this.entityDisplayName}` : "Regenerate Client Secret";
    },
    isCreateMode(): boolean {
      return this.localActionType === ActionTypeEnum.CREATE;
    },
    canManageAccessRules(): boolean {
      return usePermissionStore().hasPermission(ResourceType.AccessRules, Action.Create);
    },
    showAccessRulesStep(): boolean {
      return this.isCreateMode && this.canManageAccessRules && !this.isAccessKey;
    },
    isAccessRulesStep(): boolean {
      return this.currentStep === EStepName.AccessRules;
    },
    isLastStep(): boolean {
      return this.currentStep > 0;
    },
    nextBtnLabel(): string {
      return this.isCreateMode
        ? [ActionTypeEnum.CREATE, ActionTypeEnum.DONE][this.currentStep]
        : [ActionTypeEnum.REGENERATE, ActionTypeEnum.DONE][this.currentStep];
    },
    firstStepTitle(): string {
      return this.isCreateMode ? "Create" : "Regenerate";
    },
    dialogCustomClass(): string {
      return `credential-creation-dialog credential-creation-dialog--${this.entityType}`;
    },
    accessRuleFormOptions(): IAccessRuleManagementFormOptions {
      return {
        page: EAccessRuleModalPage.Application,
        subjectId: this.name,
        subjectType: SubjectType.ServiceAccount,
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
    close(): void {
      this.$emit("close");
    },
    goToAccessRulesStep(): void {
      this.currentStep = EStepName.AccessRules;
    },
    async nextStepAction(): Promise<void> {
      if (this.currentStep === EStepName.Create) {
        this.isCreateMode ? await this.createCredential() : await this.regenerateSecret();
      }
    },
    async validate(): Promise<boolean> {
      return await (this.credentialForm as HTMLFormElement).validate();
    },
    async createCredential(): Promise<void> {
      try {
        const isValid = await this.validate();
        if (!isValid) return;

        this.loading = true;

        const newCredential: CredentialPostResponse = this.isAccessKey
          ? await accessKeysService.createAccessKey({ name: this.newCredential.name })
          : await serviceAccountsService.createServiceAccount({ name: this.newCredential.name });

        this.name = newCredential.name;
        this.secret = newCredential.clientSecret;
        this.clientId = newCredential.clientId || "";
        this.$emit("load-credential", newCredential.id);
        this.$q.notify(alertUtil.getSuccess(`${this.entityDisplayNameSentenceCase} ${newCredential.name} created`));
        (this.$refs.stepper as QStepper).next();
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.$q.notify(alertUtil.getError(ErrorAlert.failedCreateMessage(this.entityDisplayNameLower)));
        }
      } finally {
        this.loading = false;
      }
    },
    async regenerateSecret(): Promise<void> {
      if (!this.credential) return;

      try {
        this.loading = true;

        const secretResponse = this.isAccessKey
          ? await accessKeysService.generateAccessKeySecret(this.credential.id)
          : await serviceAccountsService.regenerateSecret(this.credential.id);

        this.name = this.credential.name;
        this.clientId = this.credential.clientId;
        this.secret = secretResponse.clientSecret || "";
        (this.$refs.stepper as QStepper).next();
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.$q.notify(alertUtil.getError("Failed to regenerate client secret"));
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
};
</script>

<style lang="scss">
.credential-stepper {
  .q-stepper__dot {
    .q-icon.ph-fill.ph-pencil-simple {
      font-size: 12px;
    }
  }

  .q-stepper__tab .q-icon {
    font-size: var(--icon-size-small) !important;
  }

  &.q-stepper--horizontal {
    .q-stepper__line {
      &::before,
      &::after {
        top: 49%;
      }
    }
  }

  .header-class {
    display: flex;
    justify-content: center;

    div.col-grow {
      flex: unset;
      min-width: 24px;
    }
  }
}
.footer {
  .q-stepper__nav {
    padding-top: unset;
  }
}

.credential-creation-dialog--access-key {
  width: 500px;
}

.credential-creation-dialog--service-account {
  width: 766px;
}
</style>
