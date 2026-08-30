<template>
  <q-btn
    dark-percentage
    :percentage="downloadProgress"
    color="grey-4"
    text-color="positive"
    class="btn-icon-sm download-btn"
    icon="ph-bold ph-download-simple"
    :aid="aid"
    :label="label"
    no-caps
    @click="$emit('download')"
    :loading="loading"
  >
    <template v-slot:loading>
      <q-spinner class="download-spinner" color="primary" />
      DOWNLOADING...
    </template>
  </q-btn>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

export default defineComponent({
  name: "download-button-wtih-progress",
  emits: ["download"],
  props: {
    downloadProgress: {
      type: Number as PropType<number>,
      required: true,
    },
    loading: {
      type: Boolean,
      required: true,
    },
    label: {
      type: String,
      required: true,
    },
    aid: {
      type: String,
      required: true,
    },
  },
});
</script>

<style lang="scss" scoped>
.download-btn {
  width: 180px;
  .download-spinner {
    margin-right: 8px;
  }
}
</style>
