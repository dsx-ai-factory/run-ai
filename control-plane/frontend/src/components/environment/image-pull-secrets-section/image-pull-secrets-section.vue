<template>
  <runai-section class="image-pull-secrets-section q-mt-lg" title="Set the credentials for pulling the image">
    <div class="row">
      <runai-select
        aid="image-pull-secrets-select"
        class="col-4"
        :options="options"
        :model-value="imagePullSecretsModel"
        label="Secret name"
        stack-label
        placeholder="Select..."
        no-option-text="No secrets found"
        @update:model-value="updateImagePullSecrets"
        emit-value
        :disable="disable || disabledByPolicy"
        :policy-rules="selectGenericPolicyRules"
      />
    </div>
    <q-tooltip v-if="disabledByPolicy" target=".runai-select">
      This can't be modified based on the policy your administrator set.
    </q-tooltip>
  </runai-section>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";
// Components
import { RunaiSection } from "@/components/common/runai-section";
import { RunaiSelect } from "@/components/common/runai-select";
// Models
import type { ISelectOption } from "@/models/global.model";
import type { ImagePullSecretsRules } from "@/swagger-models/policy-service-client";
import type { ImagePullSecret } from "@/swagger-models/assets-service-client";
import type { IGenericSelectPolicyRules } from "@/models/policy.model";

export default defineComponent({
  name: "image-pull-secrets-section",
  components: {
    RunaiSection,
    RunaiSelect,
  },
  emits: {
    "update-image-pull-secrets": (value: ImagePullSecret[]) => value,
  },
  props: {
    options: {
      type: Array as PropType<ISelectOption[]>,
      required: true,
    },
    modelValue: {
      type: [Array, null] as PropType<ImagePullSecret[] | null>,
      required: true,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<ImagePullSecretsRules | null>,
      required: false,
    },
  },
  computed: {
    imagePullSecretsModel() {
      return this.modelValue?.map((secret: ImagePullSecret) => secret.name)[0] || null;
    },
    selectGenericPolicyRules(): IGenericSelectPolicyRules {
      return {
        sourceOfRule: this.policyRules?.attributes?.name?.sourceOfRule,
        required: this.policyRules?.attributes?.name?.required,
        canEdit: this.policyRules?.attributes?.name?.canEdit,
        options: this.policyRules?.attributes?.name?.options,
        isClosedList: true,
      };
    },
    disabledByPolicy(): boolean {
      return this.policyRules?.attributes?.name?.canEdit === false || this.policyRules?.instances?.canAdd === false;
    },
  },
  methods: {
    updateImagePullSecrets(value: string) {
      this.$emit("update-image-pull-secrets", value ? [{ name: value, userCredential: true }] : []);
    },
  },
});
</script>
