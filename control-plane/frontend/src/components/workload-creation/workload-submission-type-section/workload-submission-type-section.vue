<template>
  <runai-expansion-item
    class="workload-submission-type-section"
    aid="workload-submission-type-section"
    label="Submission"
    default-opened
    disable-closing
  >
    <section>
      <runai-section title="Select how to submit your workload">
        <runai-radio-options
          :model-value="submissionType"
          :options="submissionOptions"
          @update:model-value="$emit('update', $event)"
          horizontal
        />
      </runai-section>

      <template v-if="isHybridSubmission">
        <template v-if="shouldPresentPreview">
          <div class="video-container q-mt-md">
            <video style="max-width: 500px" ref="video-element" muted loop autoplay>
              <source src="/videos/hybrid-submission-intro.mp4" type="video/mp4" />
              Your browser does not support the video tag.
            </video>
          </div>
        </template>
        <section class="info-section q-my-md">
          <ol class="q-mb-md">
            <li>● Select an existing setup or start from scratch</li>
            <li>● Review internal fields' details</li>
            <li class="q-mb-md">● Modify fields without having to create a new setup</li>
          </ol>
          <div>Check it out and get to work!</div>
        </section>

        <q-checkbox
          v-if="shouldPresentPreview"
          class="minus-ml-9px"
          :model-value="dontShowMeAgainValue"
          @update:model-value="onDontShowMeAgainChange"
          label="Don't show the introduction again"
        />
      </template>
      <!-- </template> -->
      <template v-else>
        <section class="info-section q-mt-md">Continue with the original submission</section>
      </template>
    </section>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { RunaiSection } from "@/components/common/runai-section";
import { storageUtil } from "@/utils/storage.util";
import { STORAGE_KEYS } from "@/models/storage.model";
import { ESubmissionOptions, hybridSubmissionOptions } from "@/models/workload-hybrid.model";

export default defineComponent({
  components: {
    RunaiExpansionItem,
    RunaiRadioOptions,
    RunaiSection,
  },
  emits: ["update"],
  props: {
    submissionType: {
      type: String as PropType<ESubmissionOptions>,
      required: true,
    },
    disableClosing: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    defaultOpen: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
  },
  data() {
    return {
      shouldPresentPreview: true,
      dontShowMeAgainValue: false,
      submissionOptions: hybridSubmissionOptions,
    };
  },
  created() {
    if (storageUtil.get(STORAGE_KEYS.HYBRID_PREVIEW)) this.shouldPresentPreview = false;
  },
  computed: {
    isHybridSubmission(): boolean {
      return this.submissionType === ESubmissionOptions.Hybrid;
    },
  },
  methods: {
    onDontShowMeAgainChange(value: boolean): void {
      this.dontShowMeAgainValue = value;
      this.hidePreviewForFutureUse(value);
    },
    hidePreviewForFutureUse(val: boolean): void {
      storageUtil.save(STORAGE_KEYS.HYBRID_PREVIEW, val);
    },
  },
});
</script>
<style lang="scss" scoped>
.workload-submission-type-section {
  .video-container video {
    border: 2px solid $black-12;
  }
  .info-section {
    color: $black-54;
    font-style: italic;
  }
  .minus-ml-9px {
    margin-left: -9px;
  }
}
</style>
