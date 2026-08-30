<template>
  <section class="git-instance-section">
    <div class="q-mb-md">
      <runai-tooltip-wrapper
        :display-tooltip="allowOnlyContainerPath"
        tooltip-text="This was defined when the setup was created and can't be modified"
        anchor="bottom middle"
        self="bottom middle"
      >
        <div class="q-my-md">Set the repository type</div>
        <runai-radio-options
          class="q-mb-md"
          :model-value="isPrivate"
          :options="[
            { label: 'Public', value: false },
            { label: 'Private', value: true },
          ]"
          @update:model-value="onSelectRepositoryType"
          horizontal
        />
        <div class="credentials-section row q-mb-md" v-if="isPrivate">
          <runai-tooltip-wrapper
            :display-tooltip="allowOnlyContainerPath"
            tooltip-text="This was defined when the setup was created and can't be modified"
            anchor="bottom middle"
            self="bottom middle"
            class="col-6"
          >
            <runai-select
              :disable="allowOnlyContainerPath || disable"
              aid="git-secret-select"
              :model-value="selectedCredentialName || null"
              @update:model-value="onSelectSecret"
              :options="secretOptions"
              no-option-text="No secrets found"
              label="Secret"
              emit-value
              :rules="[isSecretSelected]"
              map-options
            />
            <q-tooltip class="explanatory-tooltip" v-if="disableInfo" max-width="350px">{{ disableInfo }}</q-tooltip>
          </runai-tooltip-wrapper>
        </div>
        <runai-section title="Set the data origin">
          <policy-string-field
            :disable="allowOnlyContainerPath || disable"
            :tooltip="disableInfo"
            aid="repository-input"
            :model-value="modelValue.repository"
            no-error-icon
            label="Repository URL"
            stack-label
            :placeholder="repositoryPlaceholder"
            input-class="placeholder-italic"
            :rules="repositoryRules"
            @update:model-value="$emit('update:model-value', { ...modelValue, repository: $event })"
            :policy-rules="policyRules?.repository"
          />
        </runai-section>
      </runai-tooltip-wrapper>
    </div>

    <div class="row q-mb-md justify-between">
      <runai-tooltip-wrapper
        v-if="!supportGitSyncOneParameter"
        :display-tooltip="allowOnlyContainerPath"
        tooltip-text="This was defined when the setup was created and can't be modified"
        anchor="bottom middle"
        self="bottom middle"
        class="col-6 q-mr-md"
      >
        <policy-string-field
          :disable="allowOnlyContainerPath || disable"
          :tooltip="disableInfo"
          aid="git-branch-input"
          :model-value="modelValue.branch"
          no-error-icon
          label="Git branch"
          stack-label
          placeholder="e.g., main"
          input-class="placeholder-italic"
          hint="If left empty, the repository's default branch will be used"
          @update:model-value="$emit('update:model-value', { ...modelValue, branch: $event === '' ? null : $event })"
          :policy-rules="policyRules?.branch"
        />
      </runai-tooltip-wrapper>

      <runai-tooltip-wrapper
        :display-tooltip="allowOnlyContainerPath"
        tooltip-text="This was defined when the setup was created and can't be modified"
        anchor="bottom middle"
        self="bottom middle"
        class="col"
      >
        <policy-string-field
          :disable="allowOnlyContainerPath || disable"
          :tooltip="disableInfo"
          aid="git-revision-input"
          :model-value="modelValue.revision"
          no-error-icon
          :label="revisionLabel"
          stack-label
          placeholder="e.g., dae86e1950b1277e545cee1805517500..."
          input-class="placeholder-italic"
          hint="If left empty, defaults to 'HEAD' (latest)"
          @update:model-value="$emit('update:model-value', { ...modelValue, revision: $event === '' ? null : $event })"
          :policy-rules="policyRules?.revision"
        />
      </runai-tooltip-wrapper>
    </div>

    <div>
      <div class="q-py-md">{{ containerPathTitle }}</div>
      <div class="row">
        <policy-string-field
          :disable="disable"
          :tooltip="disableInfo"
          class="col-6"
          aid="git-container-input"
          :model-value="modelValue.path"
          no-error-icon
          label="Container path"
          stack-label
          placeholder="e.g., /home/localdata"
          input-class="placeholder-italic"
          :rules="[validationRules.validatePathNotEmpty, validationRules.validatePathFormat]"
          @update:model-value="$emit('update:model-value', { ...modelValue, path: $event })"
          :policy-rules="policyRules?.path"
        />
      </div>
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiSelect } from "@/components/common/runai-select";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { PolicyStringField } from "@/components/common/policy-string-field";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { RunaiSection } from "@/components/common/runai-section";
// Models
import type { IUIGitInstance } from "@/models/workload.model";
import type { ISelectOption } from "@/models/global.model";
import type { GitRules } from "@/swagger-models/assets-service-client";
// Utils
import { assetsUtil } from "@/utils/assets.util/assets.util";
// Constants
import { errorMessages } from "@/common/error-message.constant";
import { validationRules } from "@/common/validation-rules";
import type { SecretListInfo } from "@/swagger-models/k8s-objects-tracker-client";
import { SecretType } from "@/swagger-models/k8s-objects-tracker-client";
import { GitAuthenticationMethod } from "@/swagger-models/workloads-client";
import type { ValidationRule } from "quasar";

export default defineComponent({
  name: "git-instance-section",
  components: {
    RunaiSelect,
    RunaiSection,
    RunaiTooltipWrapper,
    PolicyStringField,
    RunaiRadioOptions,
  },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: Object as PropType<IUIGitInstance>,
      required: true,
    },
    secrets: {
      type: Array as PropType<SecretListInfo[]>,
      required: true,
    },
    supportGitSyncOneParameter: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    allowOnlyContainerPath: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<GitRules | null>,
      required: false,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    disableInfo: {
      type: String as PropType<string>,
      required: false,
    },
  },
  data() {
    return {
      errorMessages,
      validationRules,
      isPrivate: false,
    };
  },
  created() {
    this.initIsPrivate();
  },
  computed: {
    revisionLabel(): string {
      return this.supportGitSyncOneParameter ? "Revision (branch, tag or hash)" : "Revision (commit hash or tag)";
    },
    containerPathTitle(): string {
      return assetsUtil.containerPathTitle();
    },
    secretOptions(): ISelectOption[] {
      return this.secrets.map((secret) => ({ label: secret.name, value: secret.name }));
    },
    selectedCredentialName(): string | null {
      if (!this.modelValue.passwordSecret && !this.modelValue.secretRef?.name) return "";
      return this.modelValue.passwordSecret || this.modelValue.secretRef?.name || null;
    },
    repositoryPlaceholder(): string {
      return this.isPrivate
        ? "e.g., git@github.com:kubernetes/kubernetes.git"
        : "e.g., https://github.com/kubernetes/kubernetes.git";
    },
    repositoryRules(): ValidationRule[] {
      return this.isPrivate
        ? [validationRules.validateUrlNotEmpty, validationRules.validatePrivateUrlFormat]
        : [validationRules.validateUrlNotEmpty, validationRules.validateUrlFormat];
    },
  },
  methods: {
    isSecretSelected(val: string): boolean | string {
      return this.isSecretValueValid(val) || errorMessages.SELECT_A_VALUE;
    },
    isSecretValueValid(val: string | null | undefined) {
      return !!val?.length;
    },
    onSelectSecret(val: string): void {
      const selectedSecret = this.secrets.find((c) => c.name === val);

      const isPassword = selectedSecret?.type === SecretType.Password;
      const isSecret = selectedSecret?.type === SecretType.Generic;

      if (!isSecret && !isPassword) {
        return;
      }

      const updatedSpec: IUIGitInstance = {
        ...this.modelValue,
        passwordSecret: undefined,
        secretRef: {
          name: val,
        },
      };
      let authenticationMethod: GitAuthenticationMethod | undefined = undefined;

      if (isPassword) {
        authenticationMethod = GitAuthenticationMethod.Password;
      } else if (isSecret) {
        authenticationMethod = GitAuthenticationMethod.SshKey;
      } else return;

      updatedSpec.secretRef && (updatedSpec.secretRef.authenticationMethod = authenticationMethod);

      this.$emit("update:model-value", updatedSpec);
    },
    onSelectRepositoryType(val: boolean): void {
      this.isPrivate = val;

      const updatedSpec: IUIGitInstance = {
        ...this.modelValue,
        passwordSecret: undefined,
        secretRef: undefined,
      };
      this.$emit("update:model-value", updatedSpec);
    },
    initIsPrivate(): void {
      // set it to private if there is a passwordSecret or secretRef.name
      this.isPrivate = !!this.modelValue.passwordSecret || !!this.modelValue.secretRef?.name;
    },
  },
});
</script>
