<template>
  <section class="data-volume-instance-section">
    <div class="row">
      <q-field label="Data volume name" stack-label class="col-6 dv-name-field" disable :loading="loadingVolume">
        <template v-slot:control>
          <div class="self-center full-width no-outline" tabindex="0">{{ dvName }}</div>
        </template>
      </q-field>
      <q-tooltip target=".dv-name-field"> This was defined when the setup was created and can't be modified </q-tooltip>
    </div>
    <div class="q-py-md">{{ containerPathTitle }}</div>
    <div class="row">
      <policy-string-field
        :disable="disable"
        class="col-6"
        aid="mount-path-input"
        :model-value="modelValue.mountPath"
        no-error-icon
        label="Container path"
        stack-label
        placeholder="e.g., /home/local/data"
        input-class="placeholder-italic"
        :rules="[validationRules.validatePathNotEmpty, validationRules.validatePathFormat]"
        @update:model-value="$emit('update:model-value', { ...modelValue, mountPath: $event })"
        :policy-rules="policyRules?.mountPath"
      />
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { PolicyStringField } from "@/components/common/policy-string-field";
// Models
import type { IUIDataVolumeInstance } from "@/models/workload.model";
import type { DataVolumeRules } from "@/swagger-models/assets-service-client";
// Constants
import { assetsUtil } from "@/utils/assets.util/assets.util";
import { validationRules } from "@/common/validation-rules";
import { dataVolumeService } from "@/services/control-plane/data-volume.service/data-volume.service";

export default defineComponent({
  name: "data-volume-instance-section",
  components: {
    PolicyStringField,
  },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: Object as PropType<IUIDataVolumeInstance>,
      required: true,
    },
    policyRules: {
      type: [Object, null] as PropType<DataVolumeRules | null>,
      required: false,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      validationRules,
      loadingVolume: false,
      dvName: "",
    };
  },
  created() {
    this.loadDataVolumes();
  },
  computed: {
    containerPathTitle(): string {
      return assetsUtil.containerPathTitle();
    },
  },
  methods: {
    async loadDataVolumes(): Promise<void> {
      if (!this.modelValue.dataVolumeId) return;
      try {
        this.loadingVolume = true;
        const dv = await dataVolumeService.getDataVolumeById(this.modelValue.dataVolumeId);
        this.dvName = dv.name;
      } catch (error) {
        console.error(error);
      } finally {
        this.loadingVolume = false;
      }
    },
  },
});
</script>
