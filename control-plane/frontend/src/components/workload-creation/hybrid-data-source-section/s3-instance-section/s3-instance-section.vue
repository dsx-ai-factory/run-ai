<template>
  <section class="s3-instance-section">
    <div class="q-mb-md">
      <div class="q-py-md">Set the data origin</div>
      <runai-tooltip-wrapper
        :display-tooltip="allowOnlyContainerPath"
        tooltip-text="This was defined when the setup was created and can't be modified"
        anchor="bottom middle"
        self="bottom middle"
      >
        <policy-string-field
          :disable="allowOnlyContainerPath || disable"
          aid="s3-service-url-input"
          :model-value="modelValue.url"
          no-error-icon
          label="S3 service URL"
          stack-label
          placeholder="e.g., https://s3.us-west-2.amazonaws.com/"
          input-class="placeholder-italic"
          :rules="[validationRules.validateUrlNotEmpty, validationRules.validateUrlFormat]"
          @update:model-value="$emit('update:model-value', { ...modelValue, url: $event })"
          :policy-rules="policyRules?.url"
        />
      </runai-tooltip-wrapper>
    </div>

    <div class="row">
      <runai-tooltip-wrapper
        :display-tooltip="allowOnlyContainerPath"
        tooltip-text="This was defined when the setup was created and can't be modified"
        anchor="bottom middle"
        self="bottom middle"
        class="col"
      >
        <policy-string-field
          :disable="allowOnlyContainerPath || disable"
          aid="s3-bucket-input"
          :model-value="modelValue.bucket"
          no-error-icon
          label="Bucket name"
          stack-label
          placeholder="e.g., my-bucket"
          input-class="placeholder-italic"
          :rules="[validationRules.validateNameNotEmpty]"
          @update:model-value="$emit('update:model-value', { ...modelValue, bucket: $event })"
          :policy-rules="policyRules?.bucket"
        />
      </runai-tooltip-wrapper>
    </div>

    <div class="credentials-section row q-mb-md">
      <runai-tooltip-wrapper
        :display-tooltip="allowOnlyContainerPath"
        tooltip-text="This was defined when the setup was created and can't be modified"
        anchor="bottom middle"
        self="bottom middle"
        class="col-6"
      >
        <runai-select
          :disable="allowOnlyContainerPath || disable"
          aid="s3-secret-select"
          :model-value="modelValue.accessKeySecret || null"
          @update:model-value="$emit('update:model-value', { ...modelValue, accessKeySecret: $event })"
          :options="secretOptions"
          map-options
          label="Secret"
          emit-value
          :rules="[isSecretSelected]"
        />
      </runai-tooltip-wrapper>
    </div>

    <div>
      <div class="q-py-md">{{ containerPathTitle }}</div>
      <div class="row">
        <policy-string-field
          :disable="disable"
          class="col-6"
          aid="s3-container-path-input"
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
// Models
import type { IUIS3Instance } from "@/models/workload.model";
import type { ISelectOption } from "@/models/global.model";
import type { S3Rules } from "@/swagger-models/policy-service-client";
// Utils
import { assetsUtil } from "@/utils/assets.util/assets.util";
// Constants
import { errorMessages } from "@/common/error-message.constant";
import { validationRules } from "@/common/validation-rules";

export default defineComponent({
  name: "s3-instance-section",
  components: {
    RunaiSelect,
    RunaiTooltipWrapper,
    PolicyStringField,
  },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: Object as PropType<IUIS3Instance>,
      required: true,
    },
    secrets: {
      type: Array as PropType<string[]>,
      required: true,
    },
    allowOnlyContainerPath: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<S3Rules | null>,
      required: false,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      errorMessages,
      validationRules,
    };
  },
  computed: {
    containerPathTitle(): string {
      return assetsUtil.containerPathTitle();
    },
    secretOptions(): ISelectOption[] {
      const options: ISelectOption[] = this.secrets.map((secret) => ({ label: secret, value: secret }));
      options.unshift({ label: "None (for public buckets)", value: null });
      return options;
    },
  },
  methods: {
    isSecretSelected(val: string): boolean | string {
      return this.isSecretValueValid(val) || errorMessages.SELECT_A_VALUE;
    },
    isSecretValueValid(val: string | null | undefined) {
      return val === null || !!val?.length;
    },
  },
});
</script>
