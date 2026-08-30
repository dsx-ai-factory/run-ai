<template>
  <runai-expansion-item class="nfs-data-mount" default-opened label="Data mount" should-validate>
    <template #subheader>
      <section class="summary row">
        Origin:
        <div class="ellipsis-l">&lrm;{{ pathSummary }}</div>
        <span> > </span> Target:
        <div class="ellipsis-l">&lrm;{{ containerPathSummary }}</div>
      </section>
    </template>
    <section>
      <div class="q-mb-md">
        <div class="q-py-md">Set the data origin</div>
        <q-input
          aid="server-input"
          :model-value="modelValue.server"
          no-error-icon
          label="NFS server (host name or host IP)"
          stack-label
          placeholder="e.g., MyServer"
          input-class="placeholder-italic server-input"
          :rules="[validationRules.validateServerNotEmpty]"
          @update:model-value="$emit('update:modelValue', { ...modelValue, server: $event })"
        />
      </div>

      <div class="q-mb-md">
        <q-input
          aid="path-input"
          :model-value="modelValue.path"
          no-error-icon
          label="NFS path"
          stack-label
          placeholder="e.g., /usr/data"
          input-class="placeholder-italic"
          :rules="[validationRules.validatePathNotEmpty, validationRules.validatePathFormat]"
          @update:model-value="$emit('update:modelValue', { ...modelValue, path: $event })"
        />
      </div>

      <div>
        <div class="q-py-md">{{ containerPathTitle }}</div>
        <q-input
          aid="mount-path-input"
          :model-value="modelValue.mountPath"
          no-error-icon
          label="Container path"
          stack-label
          placeholder="e.g., /home/local/data"
          input-class="placeholder-italic"
          :rules="[validationRules.validatePathNotEmpty, validationRules.validatePathFormat]"
          @update:model-value="$emit('update:modelValue', { ...modelValue, mountPath: $event })"
        />
      </div>
    </section>
    <section>
      <runai-sub-expansion-item label="Restrictions">
        <section class="row items-center q-mb-md">
          <q-toggle
            aid="read-only-toggle"
            :model-value="modelValue.readOnly"
            label="Prevent data modification"
            left-label
            class="text-no-wrap"
            @update:model-value="$emit('update:modelValue', { ...modelValue, readOnly: $event })"
          />
          <runai-tooltip
            class="q-ml-md"
            tooltip-position="right"
            tooltip-text="When enabled, the data will be mounted with read-only permissions."
          />
        </section>
      </runai-sub-expansion-item>
    </section>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiSubExpansionItem } from "@/components/common/runai-sub-expansion-item";
import { RunaiTooltip } from "@/components/common/runai-tooltip";

// Types
import type { NFSAssetSpec } from "@/swagger-models/assets-service-client";
import { assetsUtil } from "@/utils/assets.util/assets.util";
import { validationRules } from "@/common/validation-rules";

export default defineComponent({
  components: {
    RunaiExpansionItem,
    RunaiSubExpansionItem,
    RunaiTooltip,
  },
  emits: ["update:modelValue"],
  props: {
    modelValue: {
      type: Object as PropType<NFSAssetSpec>,
      required: true,
    },
  },
  data() {
    return {
      validationRules,
    };
  },
  computed: {
    pathSummary(): string {
      return `${this.modelValue.path ? this.modelValue.path : "None"}`;
    },
    containerPathSummary(): string {
      return `${this.modelValue.mountPath ? this.modelValue.mountPath : "None"}`;
    },
    containerPathTitle(): string {
      return assetsUtil.containerPathTitle(true);
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
