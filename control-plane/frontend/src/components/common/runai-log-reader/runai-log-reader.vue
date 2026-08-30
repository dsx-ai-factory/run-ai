<template>
  <section class="runai-log-reader">
    <div
      class="list-container"
      :class="{ paused: scrollPaused }"
      ref="list-container"
      tabindex="0"
      role="log"
      @scroll="handleScroll"
    >
      <div class="logs-list">
        <div class="log-row" v-for="(line, index) in lines" :key="index">{{ line }}</div>
      </div>
    </div>
    <div class="resume-button-wrapper">
      <q-btn
        v-if="scrollPaused"
        text-color="black"
        label="resume logs"
        class="resume-button"
        @click="resumeAutoScroll"
      />
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { nextTick, type PropType } from "vue";
// models
import type { ILogParams } from "@/models/logs.model";
import type { IWebSocketSubscription } from "@/models/global.model";

// services
import { webSocketsService } from "@/services/infra/websockets.service/websockets.service";
import { workloadService } from "@/services/cluster/workload.service/workload.service";

// stores
import { useClusterStore } from "@/stores/cluster.store";

// constants
import { RESEARCHER } from "@/common/api.constant";
import type { ClustersMinimalData } from "@/swagger-models/cluster-service-client";

// Component types
export type LogParams = { follow: boolean; pod: string; container?: string; tailLines?: number; previous?: boolean };

// WebSocket close codes enum
export enum WebSocketCloseCode {
  INVALID_FRAME_PAYLOAD = 1007, // Data corruption
  POLICY_VIOLATION = 1008, // Auth failure
  TLS_HANDSHAKE_FAIL = 1015, // Security failure
}

// WebSocket close codes that should hide logs (critical errors only)
const CRITICAL_ERROR_CODES = [
  WebSocketCloseCode.INVALID_FRAME_PAYLOAD,
  WebSocketCloseCode.POLICY_VIOLATION,
  WebSocketCloseCode.TLS_HANDSHAKE_FAIL,
];

export default defineComponent({
  name: "runai-log-reader",
  emits: ["error", "connecting", "close", "no-data"],
  props: {
    logParams: {
      type: Object as PropType<ILogParams>,
      required: true,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    supportClusterApiLogs: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
  },
  data() {
    return {
      clusterStore: useClusterStore(),
      wsSubscription: null as IWebSocketSubscription | null,
      remainderLine: "" as string,
      lines: [] as string[],
      error: false as boolean,
      connected: false as boolean,
      scrollPaused: false as boolean,
      isProgrammaticScroll: false as boolean,
      isLoading: false as boolean,
      connectionClosed: false as boolean,
      errorOccurred: false as boolean,
    };
  },
  methods: {
    appendLog(event: MessageEvent): void {
      this.connected = true;

      if (event.data instanceof Blob) {
        const reader = new FileReader();
        reader.onload = (e) => {
          const textData = e.target?.result as string;
          this.processLogData(textData);
        };
        reader.readAsText(event.data);
      } else {
        this.processLogData(event.data);
      }
    },
    processLogData(data: string): void {
      let logs: string[] = data.split("\n");
      logs[0] = this.remainderLine + logs[0];
      this.remainderLine = logs[logs.length - 1];
      logs = logs.slice(0, logs.length - 1);

      this.lines = this.lines.concat(logs);
      this.isLoading = false;
      this.$emit("connecting", false);
      this.$emit("no-data", this.lines.length === 0);

      if (!this.scrollPaused) {
        this.autoScroll();
      }
    },
    handleInitLogReader(): void {
      this.wsSubscription?.unsubscribe();
      this.isLoading = true;
      this.$emit("connecting", true);
      this.lines = [];
      this.error = false;
      this.connectionClosed = false;
      this.errorOccurred = false;

      const cluster = this.clusterStore.getClusterById(this.clusterId);
      if (!cluster) {
        console.error("Cluster not found");
        return;
      }

      if (this.supportClusterApiLogs) {
        this.initClusterApiLogReader(cluster);
      } else {
        this.initOldLogReader(cluster);
      }

      if (this.wsSubscription) {
        this.wsSubscription.onMessage = this.appendLog;
        this.wsSubscription.onError = this.onError;
        this.wsSubscription.onClose = this.onClose;
      }
    },
    initClusterApiLogReader(cluster: ClustersMinimalData): void {
      const params: LogParams = { follow: !this.logParams.previous, pod: this.logParams.podName };
      if (this.logParams.container) {
        params.container = this.logParams.container;
      }
      if (this.logParams.tail !== null && this.logParams.tail !== undefined) {
        params.tailLines = this.logParams.tail;
      }
      if (this.logParams.previous) {
        params.previous = true;
      }

      let endpoint: string;

      // Check if it's an external workload
      if (this.logParams.externalWorkloadName) {
        // Use new external workload API pattern
        endpoint = `${cluster.domain}/cluster-api/api/v1/${this.logParams.project}/workloads/external/external/${this.logParams.externalWorkloadName}`;
      } else {
        // Use existing API pattern for non-external workloads
        const framework = workloadService.getFrameworkForUrl(this.logParams.type, this.logParams.distributedFramework);
        endpoint = `${cluster.domain}/cluster-api/api/v1/${
          this.logParams.project
        }/workloads/${this.logParams.type.toLowerCase()}/${framework}/${this.logParams.jobName}`;
      }

      this.wsSubscription = webSocketsService.subscribe(endpoint, "logs", params, cluster.version, true);
    },
    initOldLogReader(cluster: ClustersMinimalData): void {
      this.wsSubscription = webSocketsService.subscribe(
        `${cluster.domain}/${RESEARCHER.v1}/`,
        "logs",
        this.logParams,
        cluster.version,
        false,
      );
    },
    onError(): void {
      this.wsSubscription?.unsubscribe();
      this.errorOccurred = true;
      this.connected = true;
      this.isLoading = false;
      this.$emit("connecting", false);
      // Don't emit error immediately - let onClose handle it based on whether we have logs
    },
    onClose(closeEvent: CloseEvent): void {
      this.connectionClosed = true;
      this.isLoading = false;
      this.$emit("connecting", false);

      // Case 1: We have logs + critical error -> Hide logs (data might be corrupted)
      if (this.lines.length > 0 && CRITICAL_ERROR_CODES.includes(closeEvent.code)) {
        this.error = true;
        this.$emit("error", true);
        return;
      }

      // Case 2: We have logs + normal closure -> Show logs
      if (this.lines.length > 0) {
        this.$emit("close", true);
        return;
      }

      // Case 3: No logs + error occurred -> Show error message
      if (this.errorOccurred) {
        this.error = true;
        this.$emit("error", true);
        return;
      }

      // Case 4: No logs + normal close -> Show no data message
      this.$emit("close", false);
    },
    autoScroll(): void {
      if (this.scrollPaused) return;

      nextTick(() => {
        const listContainer: HTMLElement = this.$refs["list-container"] as HTMLElement;
        if (!listContainer) return;
        this.isProgrammaticScroll = true;
        listContainer.scrollTop = (listContainer as HTMLElement).scrollHeight;
      });
    },
    handleScroll(): void {
      // Ignore the scroll event our own autoScroll() triggers; only pause on user-initiated
      // scrolling (mouse wheel, keyboard, touch, or scrollbar drag all fire "scroll" the same way).
      if (this.isProgrammaticScroll) {
        this.isProgrammaticScroll = false;
        return;
      }
      this.pauseAutoScroll();
    },
    pauseAutoScroll(): void {
      if (this.lines.length) {
        this.scrollPaused = true;
      }
    },
    resumeAutoScroll(): void {
      this.scrollPaused = false;
      this.autoScroll();
    },
    unsubscribe(): void {
      this.wsSubscription?.unsubscribe();
    },
  },
  watch: {
    scrollPaused: function (paused: boolean) {
      if (paused) return;
      this.autoScroll();
    },
    logParams: {
      handler(currLogParams: ILogParams, prevLogParams: ILogParams): void {
        if (
          currLogParams?.podName === prevLogParams?.podName &&
          currLogParams?.container === prevLogParams?.container &&
          currLogParams?.previous === prevLogParams?.previous
        )
          return;

        this.unsubscribe();
        this.handleInitLogReader();
      },
      immediate: true,
      deep: true,
    },
  },
  unmounted() {
    this.wsSubscription?.unsubscribe();
  },
});
</script>

<style lang="scss" scoped>
.runai-log-reader {
  $resume-btn-height: 36px;
  $resume-margin-height: 16px;

  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;

  .list-container {
    width: 100%;
    flex: 1;
    height: calc(100% - #{$resume-btn-height} - #{$resume-margin-height});
    overflow: auto;
  }

  .logs-list {
    width: 100%;
    height: 100%;
  }

  .log-row {
    min-height: 24px;
    font-size: 13px;
    padding: 2px 16px;
    overflow-wrap: break-word;
  }

  .resume-button-wrapper {
    width: 100%;
    height: $resume-btn-height;
    display: flex;
    justify-content: center;
    margin: $resume-margin-height 0;

    .resume-button {
      background-color: $grey-4;
      color: $navy;
      width: 155px;
    }
  }
}
</style>
