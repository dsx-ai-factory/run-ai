<template>
  <q-banner class="nim-info-box bg-grey-3 q-mb-md">
    <div class="row no-wrap items-center">
      <q-icon name="ph-bold ph-question" size="16px" color="icon-default" />
      <div class="q-ml-md nim-info-message">
        NVIDIA NIM requires an image that is pulled from the
        <a :href="ngcCatalog" target="_blank" rel="noopener">NGC catalog</a>. Make sure a Docker registry credential is
        configured. For more information, see
        <a :href="deployNimInference" target="_blank" rel="noopener">Deploy inference workloads with NVIDIA NIM</a>.
      </div>
      <runai-tooltip
        tooltip-text="Hide this message"
        icon="ph-bold ph-x"
        tooltip-position="top"
        size="sm"
        @click="$emit('close')"
      />
    </div>
  </q-banner>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { DOCS } from "@/constants/docs";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
export default defineComponent({
  components: {
    RunaiTooltip,
  },
  emits: {
    close: (): boolean => true,
  },
  computed: {
    ngcCatalog(): string {
      return DOCS.EXTERNAL.NIM.NGC_CATALOG;
    },
    deployNimInference(): string {
      return DOCS.EXTERNAL.NIM.DEPLOY_INFERENCE;
    },
  },
});
</script>

<style scoped lang="scss">
.nim-info-box {
  .nim-info-message {
    font-size: 12px;
    font-style: normal;
    font-weight: 400;
    line-height: 20px;
  }
}
</style>
