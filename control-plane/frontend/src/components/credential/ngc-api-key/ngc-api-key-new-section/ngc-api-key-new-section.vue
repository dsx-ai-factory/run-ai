<template>
  <section class="new-secret-details ngc-api-key-new-section">
    <div class="row">
      <q-input
        :class="inputClass"
        :model-value="secretDetails.ngc_api_key"
        label="NGC API Key"
        :rules="[validationRules.validateKeyNotEmpty, validationRules.validateGenericToken]"
        aid="ngc-api-key-input"
        @update:model-value="$emit('update-secret-details', { ...secretDetails, ngc_api_key: $event })"
        no-error-icon
      />
    </div>
  </section>
</template>

<script lang="ts">
import { validationRules } from "@/common/validation-rules";
import type { NgcApiKeyCreationSpec } from "@/swagger-models/assets-service-client";
import { defineComponent, type PropType } from "vue";

export default defineComponent({
  name: "ngc-api-key-new-section",
  props: {
    secretDetails: {
      type: Object as PropType<NgcApiKeyCreationSpec>,
      required: true,
    },
    inputClass: {
      type: String as PropType<string>,
      required: false,
      default: "col-11",
    },
  },
  data() {
    return {
      validationRules,
    };
  },
});
</script>
