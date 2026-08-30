<template>
  <section class="service-accounts-view-screen">
    <p class="text-b-subtitle2 credential-header">Copy the client ID and client secret and store them securely</p>
    <div class="row items-center no-wrap credential-row">
      <span class="field-type">{{ isAccessKey ? "Access key:" : "Service account:" }} </span>
      <div class="row items-center flex-1 credential-value">
        <span class="flex-1 text-b-subtitle2">{{ name }}</span>
      </div>
    </div>
    <div class="row items-center no-wrap credential-row">
      <span class="field-type">Client ID: </span>
      <div class="row items-center flex-1 credential-value">
        <span class="flex-1 text-b-subtitle2">{{ client }}</span>
        <runai-copy-to-clipboard-btn :text-to-copy="client" />
      </div>
    </div>
    <div class="row items-center no-wrap credential-row">
      <span class="field-type">{{ !isCreateMode ? "New client secret: " : "Client secret: " }} </span>
      <div class="row items-center flex-1 credential-value">
        <span class="flex-1 text-b-subtitle2 value-text">{{ isPassHidden ? "••••••••" : secret }}</span>
        <div class="row items-center credential-icons">
          <runai-btn
            aid="toggle-secret-visibility-btn"
            aria-label="Toggle client secret visibility"
            class="text-icon-default runai-icon-btn--default"
            flat
            size="10px"
            round
            :icon="isPassHidden ? 'ph-fill ph-eye-slash' : 'ph-fill ph-eye'"
            :tooltip="isPassHidden ? 'Show secret' : 'Hide secret'"
            @click="handleToggleSecretVisibility"
          />
          <runai-copy-to-clipboard-btn :text-to-copy="secret" />
        </div>
      </div>
    </div>
  </section>
</template>
<script lang="ts">
import { RunaiCopyToClipboardBtn } from "@/components/common/runai-copy-to-clipboard-btn/";
import { RunaiBtn } from "@/components/common/runai-btn";
import { defineComponent } from "vue";

export default defineComponent({
  components: { RunaiCopyToClipboardBtn, RunaiBtn },
  props: {
    isCreateMode: {
      type: Boolean,
      required: true,
    },
    isAccessKey: {
      type: Boolean,
      default: false,
    },
    name: {
      type: String,
      required: true,
    },
    client: {
      type: String,
      required: true,
    },
    secret: {
      type: String,
      required: true,
    },
  },
  data() {
    return {
      isPassHidden: true as boolean,
    };
  },
  methods: {
    handleToggleSecretVisibility(): void {
      this.isPassHidden = !this.isPassHidden;
    },
  },
});
</script>
<style lang="scss" scoped>
.service-accounts-view-screen {
  margin: -24px;

  .credential-header {
    height: 60px;
    padding: 16px;
    display: flex;
    align-items: center;
  }
  .credential-row {
    height: 60px;
    padding: 16px;
    gap: 20px;
  }
  .field-type {
    flex: 1 0 0;
    color: $black-54;
    white-space: nowrap;
  }
  .credential-value {
    gap: 16px;
    min-width: 0;

    span {
      overflow: hidden;
      white-space: nowrap;
      text-overflow: ellipsis;
      min-width: 0;
    }
  }
  .credential-icons {
    gap: 5px;
  }
}
</style>
