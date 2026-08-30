<template>
  <runai-expansion-item class="s3-data-mount" default-opened label="Data mount" should-validate>
    <template #subheader>
      <section class="summary row">
        Origin:
        <div class="ellipsis-l">&lrm;{{ serviceUrlSummary }}</div>
        <span> > </span> Target:
        <div class="ellipsis-l">&lrm;{{ containerPathSummary }}</div>
      </section>
    </template>
    <section>
      <div class="q-mb-md">
        <div class="q-py-md">Set the data origin</div>
        <q-input
          aid="s3-service-url-input"
          :model-value="modelValue.url"
          no-error-icon
          label="S3 service URL"
          stack-label
          placeholder="e.g., https://s3.us-west-2.amazonaws.com/"
          input-class="placeholder-italic"
          :rules="[validationRules.validateUrlNotEmpty, validationRules.validateUrlFormat]"
          @update:model-value="$emit('update:model-value', { ...modelValue, url: $event })"
        />
      </div>

      <div class="credentials-section row q-mb-md">
        <runai-select
          aid="s3-credentials-select"
          class="col-8"
          :model-value="accessKeyAssetId"
          @update:model-value="$emit('update:model-value', { ...modelValue, accessKeyAssetId: $event })"
          :loading="loadingCredentials"
          :options="credentialOptions"
          map-options
          label="Credentials"
          emit-value
          :rules="[isCredSelected]"
        />
      </div>
      <div class="row">
        <q-input
          aid="s3-bucket-input"
          class="col-8"
          :model-value="modelValue.bucket"
          no-error-icon
          label="Bucket name"
          stack-label
          placeholder="e.g., my-bucket"
          input-class="placeholder-italic"
          :rules="[validationRules.validateNameNotEmpty]"
          @update:model-value="$emit('update:model-value', { ...modelValue, bucket: $event })"
        />
      </div>
      <div>
        <div class="q-py-md">{{ containerPathTitle }}</div>
        <q-input
          aid="s3-container-path-input"
          :model-value="modelValue.path"
          no-error-icon
          label="Container path"
          stack-label
          placeholder="e.g., /home/localdata"
          input-class="placeholder-italic"
          :rules="[validationRules.validatePathNotEmpty, validationRules.validatePathFormat]"
          @update:model-value="$emit('update:model-value', { ...modelValue, path: $event })"
        />
      </div>
    </section>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiSelect } from "@/components/common/runai-select";

// Types
import { errorMessages } from "@/common/error-message.constant";
import type { AccessKey, S3AssetSpec } from "@/swagger-models/assets-service-client";
import type { ISelectOption } from "@/models/global.model";
import { assetsUtil } from "@/utils/assets.util/assets.util";
import { validationRules } from "@/common/validation-rules";

export default defineComponent({
  components: {
    RunaiExpansionItem,
    RunaiSelect,
  },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: Object as PropType<S3AssetSpec>,
      required: true,
    },
    credentials: {
      type: Array as PropType<AccessKey[]>,
      required: true,
    },
    loadingCredentials: {
      type: Boolean as PropType<boolean>,
      required: true,
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
      return assetsUtil.containerPathTitle(true);
    },
    accessKeyAssetId(): string | null {
      if (this.modelValue.accessKeyAssetId === undefined) return "";
      return this.modelValue.accessKeyAssetId;
    },
    serviceUrlSummary(): string {
      return `${this.modelValue.url ? this.modelValue.url : "None"}`;
    },
    containerPathSummary(): string {
      return `${this.modelValue.path ? this.modelValue.path : "None"}`;
    },
    credentialOptions(): ISelectOption[] {
      const mappedOptions: ISelectOption[] = this.credentials.map((cred: AccessKey) => {
        return { label: cred.meta.name, value: cred.meta.id };
      });
      return [{ label: "None (for public buckets)", value: null }, ...mappedOptions];
    },
  },
  methods: {
    isCredSelected(val: string): boolean | string {
      return this.isCredValueValid(val) || errorMessages.SELECT_CREDENTIAL;
    },
    isCredValueValid(val: string | null | undefined) {
      return val === null || !!val?.length;
    },
  },
});
</script>
<style lang="scss" scoped>
.summary {
  white-space: pre;
}

.ellipsis-l {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100px;
  direction: rtl;
  text-align: left;
}
</style>
