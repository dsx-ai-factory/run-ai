<template>
  <section class="volume-instance-section">
    <volume-form
      :volume="modelValue"
      :storage-classes="storageClasses"
      :policy-rules="policyRules"
      :show-volume-persistency-section="showVolumePersistencySection"
      :loading-storage-class="loadingStorageClass"
      @update-volume="$emit('update:model-value', $event)"
      :disable="disable"
      :disable-info="disableInfo"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components

import { VolumeForm } from "@/components/section/volume-section/volume-box/volume-form";
// Models
import type { IUIPvcInstance } from "@/models/workload.model";
import type { StorageClassV2 } from "@/swagger-models/k8s-objects-tracker-client";
import type { PvcRules } from "@/swagger-models/assets-service-client";

export default defineComponent({
  name: "volume-instance-section",
  components: {
    VolumeForm,
  },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: Object as PropType<IUIPvcInstance>,
      required: true,
    },
    storageClasses: {
      type: Array as PropType<Array<StorageClassV2>>,
      required: true,
    },
    showVolumePersistencySection: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    loadingStorageClass: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<PvcRules | null>,
      required: false,
      default: () => ({}),
    },
    disableInfo: {
      type: String as PropType<string>,
      required: false,
    },
  },
});
</script>
