<template>
  <div class="slack-integration-card">
    <p>Nvidia Run:ai is connected to your Slack workspace:</p>
    <div v-if="workspaceName" class="workspace-name-container">
      <q-icon name="ph-bold ph-slack-logo" class="slack-icon" />
      <span class="workspace-name">{{ workspaceName }}</span>
    </div>
    <div class="slack-integration-card-footer">
      <div class="slack-integration-status">Status: <span class="status-text">Connected</span></div>
      <runai-btn
        class="disconnect-btn"
        label="Disconnect"
        flat
        aid="disconnect-slack-integration-btn"
        :disable="readonly"
        @click="disconnectSlackIntegration"
      />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// components
import { RunaiBtn } from "@/components/common/runai-btn";

export default defineComponent({
  name: "slack-integration-card",
  components: { RunaiBtn },
  emits: ["disconnect"],
  props: {
    workspaceName: {
      type: String,
      required: true,
    },
    readonly: {
      type: Boolean,
      default: false,
    },
  },
  methods: {
    disconnectSlackIntegration(): void {
      this.$emit("disconnect");
    },
  },
});
</script>

<style scoped lang="scss">
.slack-integration-card {
  background-color: $body-background-color;
  padding: 16px;
  width: 360px;

  .workspace-name-container {
    display: flex;
    align-items: center;
    gap: 4px;
    margin: 12px 0;

    .slack-icon {
      font-size: 14px;
      color: $black-54;
      margin-bottom: 2px;
    }

    .workspace-name {
      font-weight: 600;
    }
  }

  .slack-integration-card-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .slack-integration-status {
      .status-text {
        color: $success;
        font-weight: 600;
      }
    }

    .disconnect-btn {
      color: $black-54;
    }
  }
}
</style>
