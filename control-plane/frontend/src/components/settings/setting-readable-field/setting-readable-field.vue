<template>
  <section class="setting-readable-field q-mt-md col-12">
    <div class="q-mb-xs">{{ label }}</div>
    <section class="middle row no-wrap items-center">
      <div class="col-11">
        <pre ref="copyableText" :style="{ 'overflow-y': 'auto' }">
           <slot/>
        </pre>
      </div>
      <q-btn
        class="copy-btn runai-icon-btn--default"
        flat
        round
        icon="ph-fill ph-copy"
        @click="copyText"
        aria-label="Copy to clipboard"
      >
        <q-tooltip class="tool-tip-dark"> Copy to clipboard </q-tooltip>
      </q-btn>
    </section>
  </section>
</template>
<script lang="ts">
import { copyToClipboard } from "quasar";
import type { PropType } from "vue";
import { alertUtil } from "@/utils/alert.util";

export default {
  props: {
    label: {
      type: String as PropType<string>,
      required: true,
    },
  },
  methods: {
    async copyText(): Promise<void> {
      const text: string = (this.$refs.copyableText as HTMLElement).innerText;
      await copyToClipboard(text);
      this.$q.notify(alertUtil.getSuccess("Copied!"));
    },
  },
};
</script>
<style scoped lang="scss">
.setting-readable-field {
  .copy-btn {
    color: $black-54;
  }
  .middle {
    background: #f1f3f4;
    pre {
      font-family: Roboto;
      white-space: pre-line;
      word-break: break-word;
      overflow: hidden;
      text-overflow: ellipsis;
      overflow-wrap: break-word;
    }
  }
  .tool-tip-dark {
    // TODO: check why its not working
    background-color: $positive;
    color: $white;
  }
}
</style>
