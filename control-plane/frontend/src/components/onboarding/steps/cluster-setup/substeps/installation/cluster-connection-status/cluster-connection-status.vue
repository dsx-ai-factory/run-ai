<template>
  <div class="cluster-connection-status">
    <div class="connection-content">
      <div class="connection-message" :class="isClusterConnected ? 'q-ml-md' : ''">
        <div>{{ connectionMessage }}</div>
      </div>
      <div class="connection-animation-and-stats">
        <div class="connection-animation">
          <vue3-lottie
            v-if="shouldShowLoading"
            :animation-data="clusterConnect1"
            :height="189"
            :width="256"
            :speed="1.0"
          />
          <vue3-lottie v-else :animation-data="clusterConnect2" :height="189" :width="256" :speed="1.0" :loop="false" />
        </div>

        <!-- CLUSTER STATS -->
        <div v-if="isClusterConnected" class="cluster-stats">
          <div class="stat" v-if="gpuCount !== null">
            <span class="stat-number">{{ gpuCount }}</span> GPUs
          </div>
          <q-separator class="cluster-stats-separator" v-if="gpuCount !== null && nodeCount !== null" />
          <div class="stat" v-if="nodeCount !== null">
            <span class="stat-number">{{ nodeCount }}</span> Nodes
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, computed, onMounted, onUnmounted, type PropType } from "vue";
import { Vue3Lottie } from "vue3-lottie";
import { ClusterState, NodeTelemetryType } from "@/swagger-models/cluster-service-client";
import { intervalUtil } from "@/utils/interval.util";
import { EIntervalLabels } from "@/models/interval.model";

// Import animations
import clusterConnect1 from "@/assets/icons/animation/cluster/clusterConnect1.json";
import clusterConnect2 from "@/assets/icons/animation/cluster/clusterConnect2.json";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { widgetUtil } from "@/utils/widget.util";
import { ErrorAlert } from "@/utils/error-alert.util";

export default defineComponent({
  name: "cluster-connection-status",
  components: { Vue3Lottie },
  emits: {
    "update:is-cluster-connected": (isConnected: boolean) => typeof isConnected === "boolean",
  },
  props: {
    clusterUuid: {
      type: String as PropType<string>,
      required: false,
    },
    isClusterConnected: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  setup(props, { emit }) {
    const shouldShowLoading = ref(true);

    const nodeCount = ref<number | null>(null);
    const gpuCount = ref<number | null>(null);
    const loadingCounts = ref(false);
    const errorCounts = ref<string | null>(null);

    const connectionMessage = computed(() => {
      if (props.isClusterConnected) {
        return "Cluster connected!";
      }
      return "Waiting for cluster to connect";
    });

    const fetchCounts = async (clusterUuid: string) => {
      loadingCounts.value = true;
      errorCounts.value = null;
      try {
        // Fetch total nodes
        const nodeResp = await clusterService.getNodeTelemetry(NodeTelemetryType.TotalGpuNodes, clusterUuid);
        nodeCount.value = widgetUtil.sumAllTelemetryValues(nodeResp.values);
        // Fetch total GPUs
        const gpuResp = await clusterService.getNodeTelemetry(NodeTelemetryType.TotalGpus, clusterUuid);
        gpuCount.value = widgetUtil.sumAllTelemetryValues(gpuResp.values);
      } catch (e) {
        // On error, just hide the counts (do not show error message)
        new ErrorAlert({
          generalMessage: "Failed to fetch cluster telemetry",
        });
        nodeCount.value = null;
        gpuCount.value = null;
        errorCounts.value = null;
      } finally {
        loadingCounts.value = false;
      }
    };

    const checkClusterStatus = async () => {
      if (!props.clusterUuid || props.isClusterConnected) {
        return;
      }
      try {
        const cluster = await clusterService.getClusterByUuid(props.clusterUuid);

        if (cluster && cluster.status?.state === ClusterState.Connected) {
          emit("update:is-cluster-connected", true);

          shouldShowLoading.value = false;
          await fetchCounts(cluster.uuid);
        }
      } catch (error) {
        new ErrorAlert({
          generalMessage: "Failed to check cluster status",
        });
      }
    };

    const startStatusChecking = () => {
      // Check immediately
      checkClusterStatus();

      // Then check periodically
      intervalUtil.startInterval(EIntervalLabels.CheckClusterStatus, checkClusterStatus);
    };

    const stopStatusChecking = () => {
      intervalUtil.stopInterval(EIntervalLabels.CheckClusterStatus);
    };

    onMounted(() => {
      startStatusChecking();
    });

    onUnmounted(() => {
      stopStatusChecking();
    });

    return {
      shouldShowLoading,
      connectionMessage,
      clusterConnect1,
      clusterConnect2,
      nodeCount,
      gpuCount,
      loadingCounts,
      errorCounts,
    };
  },
});
</script>

<style scoped lang="scss">
.cluster-connection-status {
  display: flex;
  margin-top: 20px;
  max-height: 120px;

  .connection-content {
    display: flex;
    flex-direction: column;

    .connection-message {
      font-weight: 500;
      font-size: 14px;
      line-height: 30px;
    }
    .connection-animation-and-stats {
      display: flex;
      align-items: center;
      gap: 32px;
      position: relative;
      left: -40px;
      bottom: 24px;
    }
    .connection-animation {
      position: static;
      left: 0;
    }
    .cluster-stats {
      position: absolute;
      left: 225px;
      width: 100%;
      bottom: 68px;
      display: flex;
      flex-direction: column;
      justify-content: center;
      gap: 12px;
      font-size: 14px;
      color: $black-87;
      min-width: 120px;
      .cluster-stats-separator {
        max-width: 78px;
      }
      .stat {
        font-weight: 400;
        .stat-number {
          font-size: 20px;
          font-weight: 700;
        }
      }
    }
  }
}
</style>
