<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item
    class="policy-yaml-section"
    label="Policy YAML"
    default-opened
    :body-no-padding="true"
    :section-invalid="!existingPolicy"
  >
    <section class="policy-yaml-section-content">
      <!-- Rule builder / YAML editor toggle (Figma: top of the Rules section), supplied by the form. -->
      <div v-if="$slots['mode-toggle']" class="q-mb-md"><slot name="mode-toggle" /></div>
      <section class="existing-policy-section" v-if="existingPolicy">
        <notification-alert-bar
          class="flex-1"
          v-if="hasClusterVersionConflict && !hasAcknowledgedApproval"
          :variant="ENotificationBarVariant.Warning"
        >
          This policy was created using an older cluster version and is currently saved on both the cluster and the
          control plane.<span class="text-bold"
            >Once it is edited, the updated policy will only be saved on the control plane, and the previous version will
            be removed from the cluster.</span
          >
          To proceed, please acknowledge this change.<template v-if="DOCS.SUPPORT.URL">
            For more information
            <a aid="policy-contact-support-link" :href="DOCS.SUPPORT.URL" target="_blank" rel="noopener noreferrer"
              >contact support.</a
            >
          </template>
          <q-btn
            color="primary"
            flat
            label="Acknowledge & Continue Editing"
            aid="acknowledge-continue-btn"
            @click="acknowledgeAndContinue"
          />
        </notification-alert-bar>
        <div class="row items-center justify-between q-mb-md">
          <div>A policy YAML has already been configured for the selected scope and workload type:</div>
          <q-btn
            v-permission="{ resourceType: ResourceType.Policies, action: Action.Update }"
            color="primary"
            flat
            aid="edit-policy-btn"
            label="Edit Policy"
            :loading="verifying"
            :disable="hasClusterVersionConflict && !hasAcknowledgedApproval"
            @click="() => openCodeEditor(true)"
          />
        </div>

        <div class="policy-overview">
          <runai-editor :key="editorKey" read-only :code="policyYamlPreview" height="500px" />
        </div>
      </section>

      <section class="no-policy-section" v-else>
        <div class="q-mb-md">Add a raw YAML configuration file to define policy rules</div>
        <div>
          <slot name="primary-btn">
            <q-btn
              v-permission="{ resourceType: ResourceType.Policies, action: Action.Update }"
              color="primary"
              flat
              class="btn-icon-sm policy-yaml-empty-action"
              icon="ph-bold ph-plus"
              aid="add-policy-btn"
              label="Policy YAML"
              @click="() => openCodeEditor()"
            />
          </slot>
        </div>
        <div class="row">
          <q-field class="col-4 form-hint no-padding" :model-value="!existingPolicy" :rules="[isPolicyExist]"></q-field>
        </div>
      </section>
    </section>
    <runai-editor-modal
      aid="runai-editor-modal"
      v-if="isEditorOpen"
      :invalid-message="errorMessages.ADD_VALID_POLICY_YAML"
      :initial-code="policyYamlPreview"
      :validating="validating"
      :invalid="policyValidationFailed"
      @submit="updatePolicyYaml"
      @changed="policyValidationFailed = false"
      @cancel="closeEditor"
    >
      <template #subheader>
        <div class="row justify-between">
          <span> Type or paste a YAML policy with defaults and rules </span>
          <span class="italic">
            To learn how to configure a policy, see the
            <a :href="DOCS.PLATFORM_MANAGEMENT.POLICIES" target="_blank">Policies and Rules</a>
            guide
          </span>
        </div>
      </template>
    </runai-editor-modal>
  </runai-expansion-item>
</template>
<script lang="ts">
import { errorMessages } from "@/common/error-message.constant";
import { RunaiEditorModal } from "@/components/common/runai-editor-modal";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import {
  type WorkspacePolicyV2,
  type TrainingPolicyV2,
  type DistributedPolicyV2,
  PolicyType,
} from "@/swagger-models/policy-service-client";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { yamlUtil } from "@/utils/yaml.util";
import { defineComponent, type PropType } from "vue";
import { ErrorAlert } from "@/utils/error-alert.util";
import { policyService } from "@/services/control-plane/policy.service/policy.service";
import type { IPolicyMetaAndType } from "@/views/policy/policy-edit.vue";
import type { TPolicyChangeRequest, TPolicyDefaultAndRules, TWorkloadPolicy } from "@/models/policy.model";
import { RunaiEditor } from "@/components/common/runai-editor";
import { NotificationAlertBar, ENotificationBarVariant } from "@/components/common/notification-alert-bar";
import { DOCS } from "@/constants/docs";
import { HttpErrorResponse } from "@/models/http-response.model";

// this string has to look like this make sure not to add gaps or lines in the string
const emptyPolicy = `defaults:
rules:`;

const emptyDistributedPolicy = `defaults:
  worker:
  master:
rules:
  worker:
  master:
`;

export default defineComponent({
  components: {
    RunaiExpansionItem,
    RunaiEditorModal,
    RunaiEditor,
    NotificationAlertBar,
  },
  emits: ["update-policy"],
  props: {
    existingPolicy: {
      type: [Object, null] as PropType<TPolicyDefaultAndRules | null>,
      required: true,
    },
    policyMetaAndType: {
      type: [Object, null] as PropType<IPolicyMetaAndType | null>,
      required: true,
    },
  },
  data() {
    return {
      Action: Action,
      ResourceType: ResourceType,
      isEditorOpen: false,
      editorKey: 101 as number,
      validating: false as boolean,
      policyValidationFailed: false as boolean,
      verifying: false as boolean,
      userApprovalForClusterDeletion: false as boolean,
      hasClusterVersionConflict: false as boolean,
      hasAcknowledgedApproval: false as boolean,
      errorMessages: errorMessages,
      DOCS,
      ENotificationBarVariant,
    };
  },
  async created() {
    if (this.existingPolicy) {
      const shouldVerify = await this.verifyClusterDeletion();
      if (shouldVerify) {
        this.hasClusterVersionConflict = true;
      }
    }
  },
  computed: {
    policyYamlPreview(): string {
      if (!this.existingPolicy) {
        if (this.policyMetaAndType?.policyType === PolicyType.Distributed) {
          return emptyDistributedPolicy;
        }
        return emptyPolicy;
      }

      return yamlUtil.objectToYaml(this.existingPolicy)?.trimStart() || "";
    },
  },
  methods: {
    isPolicyExist(val: boolean): boolean | string {
      return !val ? true : errorMessages.ADD_VALID_POLICY_YAML;
    },
    async openCodeEditor(checkVerification?: boolean): Promise<void> {
      // If we haven't checked yet and this is an edit operation, check for verification
      if (checkVerification && !this.hasClusterVersionConflict) {
        const shouldVerify = await this.verifyClusterDeletion();
        if (shouldVerify) {
          this.hasClusterVersionConflict = true;
          return;
        }
      }

      // Only open editor if approval is not needed or has been acknowledged
      if (!this.hasClusterVersionConflict || this.hasAcknowledgedApproval) {
        this.isEditorOpen = true;
      }
    },
    async updatePolicyYaml(updatedPolicy: string): Promise<void> {
      this.policyValidationFailed = false;
      const policyToUpdate: TPolicyDefaultAndRules = yamlUtil.yamlToObject(updatedPolicy) as TPolicyDefaultAndRules;
      const isValid: boolean = await this.validatePolicy(policyToUpdate);
      if (!isValid) {
        this.policyValidationFailed = true;
        return;
      }
      this.$emit("update-policy", policyToUpdate);
      this.closeEditor();
    },

    async validatePolicy(updatedPolicy: TPolicyDefaultAndRules): Promise<boolean> {
      if (!this.policyMetaAndType?.policyType || !updatedPolicy) return false;
      try {
        this.validating = true;
        const policyToUpdate = {
          meta: {
            ...this.policyMetaAndType.policyMeta,
          },
          policy: updatedPolicy,
        } as TWorkloadPolicy;
        await policyService.validatePolicy(
          this.policyMetaAndType.policyType,
          this.getPreparedPolicyForRequest(policyToUpdate),
          this.userApprovalForClusterDeletion,
        );

        return true;
      } catch (error: unknown) {
        console.error("Error validating policy", error);
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedValidateMessage("Policy"),
        });
        const removePattern = /line.*?:/g;
        this.$q.notify({ ...errorAlert.getNotification(error, removePattern), timeout: 0 });
        return false;
      } finally {
        this.validating = false;
      }
    },
    closeEditor(): void {
      this.isEditorOpen = false;
      this.policyValidationFailed = false;
    },
    getPreparedPolicyForRequest(updatedPolicy: TWorkloadPolicy): TPolicyChangeRequest {
      return this.policyMetaAndType?.policyType === PolicyType.Distributed
        ? policyService.prepareDistributedDataForUpdateRequest(updatedPolicy as DistributedPolicyV2)
        : policyService.prepareDataForUpdateRequest(updatedPolicy as WorkspacePolicyV2 | TrainingPolicyV2);
    },
    async verifyClusterDeletion(): Promise<boolean> {
      if (!this.policyMetaAndType?.policyType || !this.existingPolicy) return false;
      try {
        this.verifying = true;
        const policyToVerify = {
          meta: {
            ...this.policyMetaAndType.policyMeta,
          },
          policy: this.existingPolicy,
        } as TWorkloadPolicy;
        await policyService.validatePolicy(
          this.policyMetaAndType.policyType,
          this.getPreparedPolicyForRequest(policyToVerify),
        );

        return false;
      } catch (error: unknown) {
        // special message from the backend that would indicate that we need the user approval
        if (
          error instanceof HttpErrorResponse &&
          error.statusCode === 412 &&
          error.message.includes("needs your approval")
        ) {
          return true;
        }
        console.warn("unexpected error during verification", error);
        return false;
      } finally {
        this.verifying = false;
      }
    },
    acknowledgeAndContinue(): void {
      this.hasAcknowledgedApproval = true;
      this.userApprovalForClusterDeletion = true;
    },
  },
  watch: {
    policyYamlPreview: {
      handler(): void {
        this.editorKey++;
      },
    },
  },
});
</script>
<style lang="scss" scoped>
.policy-yaml-section-content {
  // Match the any-workload (GVK) YAML section: 20px block / 16px inline (Figma).
  padding: var(--space-l, 20px) var(--space-m, 16px);
}

.preview-yaml {
  height: 300px;
  overflow: auto;
  background: var(--bg-secondary-default);
  font-family: Roboto;
}
</style>
