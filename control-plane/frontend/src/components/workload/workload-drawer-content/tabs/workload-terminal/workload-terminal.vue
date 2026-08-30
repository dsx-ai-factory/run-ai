<template>
  <section class="runai-terminal column no-wrap">
    <workload-pod-container-selector
      :pods="pods"
      :selected-pod-name="selectedPodName"
      :selected-container-name="selectedContainerName"
      :show-pod-status="true"
      @pod-selected="handlePodSelected"
      @container-selected="handleContainerSelected"
    />
    <section class="terminal-body">
      <runai-default-loader v-if="showLoader" />
      <runai-drawer-no-data
        v-else-if="showConnectionError || showDrawerNoData"
        :messages="noDataMessages"
        :icon-name="noDataIcon"
        wrapper-padding-class="q-pt-lg"
      >
        <template v-if="showConnectionError" #additional-content>
          <span class="row justify-center">
            Make sure the pod is running and follow the
            <a :href="troubleshootingUrl" class="text-white troubleshooting-link q-ml-xs" target="_blank"
              >troubleshooting guide</a
            >.
          </span>
        </template>
      </runai-drawer-no-data>
      <runai-terminal v-else-if="workload && websocket" :websocket="websocket" @close="closeTerminal" />
    </section>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch, onUnmounted } from "vue";
import { RunaiTerminal } from "../../../../common/runai-terminal/index";
import { type Workload } from "@/swagger-models/workloads-client";
import { type DisplayedCluster } from "@/swagger-models/cluster-service-client";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { useAuthStore } from "@/stores/auth.store";
import { useClusterStore } from "@/stores/cluster.store";
import { RunaiDrawerNoData } from "@/components/common/runai-drawer-no-data/index";
import { RunaiDefaultLoader } from "@/components/common/runai-default-loader/index";
import { WorkloadPodContainerSelector } from "@/components/common/workload-pod-container-selector";
import { EWorkloadPodPhase, type WorkloadServicePod } from "@/models/workload.model";
import { DOCS } from "@/constants/docs";

export interface Props {
  workload?: Workload | null;
  pods: WorkloadServicePod[];
  loadingPods?: boolean;
  selectedPodName: string;
  selectedContainerName: string;
}

const props = withDefaults(defineProps<Props>(), {
  loadingPods: false,
});

const emit = defineEmits<{
  "pod-selected": [podName: string];
  "container-selected": [containerName: string];
}>();

const websocket = ref<WebSocket>();
const loading = ref(false);
const error = ref("");

// Track the last connected pod/container to avoid unnecessary reconnections
const lastConnectedPod = ref("");
const lastConnectedContainer = ref("");

const POD_NOT_RUNNING_ERROR = "To connect, the pod must be running";

const selectedPod = computed(() => props.pods.find((pod) => pod.name === props.selectedPodName));

const isPodRunning = (pod?: WorkloadServicePod): boolean => pod?.k8sPhase === EWorkloadPodPhase.Running;

const selectedPodIsRunning = computed(() => isPodRunning(selectedPod.value));

const troubleshootingUrl = DOCS.TROUBLESHOOTING.TERMINAL;

const CONNECTION_ERROR_MESSAGES = ["Couldn't connect to the terminal."];

const drawerNoDataMessages = computed<string[]>(() => {
  if (!props.loadingPods && !props.pods.length) {
    return ["No pods available for this workload."];
  }
  if (!selectedPodIsRunning.value) {
    return [POD_NOT_RUNNING_ERROR];
  }
  return [POD_NOT_RUNNING_ERROR];
});

const noDataMessages = computed(() =>
  showConnectionError.value ? CONNECTION_ERROR_MESSAGES : drawerNoDataMessages.value,
);

const noDataIcon = computed(() => (showConnectionError.value ? "ph-bold ph-file-minus" : "ph-bold ph-terminal-window"));

const showLoader = computed(() => {
  if (props.loadingPods) return true;
  if (!props.loadingPods && !props.pods.length) return false;
  if (!selectedPodIsRunning.value) return false;
  return loading.value;
});

const showConnectionError = computed(() => {
  if (props.loadingPods) return false;
  if (!selectedPodIsRunning.value) return false;
  return Boolean(error.value) && error.value !== POD_NOT_RUNNING_ERROR;
});

const showDrawerNoData = computed(() => {
  if (!props.loadingPods && !props.pods.length) return true;
  if (!selectedPodIsRunning.value) return true;
  if (error.value === POD_NOT_RUNNING_ERROR) return true;
  return false;
});
const createWebSocket = (workload: Workload, cluster: DisplayedCluster, podName: string, containerName: string) => {
  const accessToken = useAuthStore().accessToken;
  const protocols = [`auth.bearer.${accessToken}`, "Bearer"];

  // Build URL
  let baseUrl: string;
  if (workload.externalWorkloadName) {
    baseUrl = `${cluster.domain}/cluster-api/api/v1/${
      workload.projectName
    }/workloads/external/external/${workload.externalWorkloadName}`;
  } else {
    const framework = workloadService.getFrameworkForUrl(workload.type, workload.distributedFramework);
    baseUrl = `${cluster.domain}/cluster-api/api/v1/${
      workload.projectName
    }/workloads/${workload.type.toLowerCase()}/${framework}/${workload.name}`;
  }

  // Build query params
  const params = new URLSearchParams({
    pod: podName,
    container: containerName,
    stdin: "true",
    stdout: "true",
    stderr: "true",
    tty: "true",
    command: "WyIvYmluL2Jhc2giXQ==", // base64 ["/bin/bash"]
  });

  const wsUrl = `${baseUrl}/exec?${params}`.replace("http://", "ws://").replace("https://", "wss://");

  const ws = new WebSocket(wsUrl, protocols);
  ws.binaryType = "arraybuffer";

  ws.onopen = () => {
    loading.value = false;
    lastConnectedPod.value = podName;
    lastConnectedContainer.value = containerName;
  };

  ws.onerror = () => {
    // if a pod status is not running, show the error message "To connect, the pod must be running"
    const pod = props.pods.find((p) => p.name === podName);
    if (!pod) {
      console.error("Selected pod not found");
      error.value = "Selected pod not found";
      loading.value = false;
      return;
    }
    if (pod.k8sPhase.toLowerCase() !== "running") {
      error.value = POD_NOT_RUNNING_ERROR;
      loading.value = false;
      return;
    }
    error.value = "Failed to connect to terminal";
    loading.value = false;
  };

  return ws;
};

const connectTerminal = (podName: string, containerName: string) => {
  closeTerminal();
  error.value = "";

  const workload = props.workload;
  if (!workload) return;

  const pod = props.pods.find((p) => p.name === podName);
  if (!pod) {
    console.error("Selected pod not found");
    error.value = "Selected pod not found";
    return;
  }

  // Don't attempt to connect if the selected pod is not running.
  if (!isPodRunning(pod)) {
    error.value = POD_NOT_RUNNING_ERROR;
    return;
  }

  // Skip if already connected to the same pod/container
  if (podName === lastConnectedPod.value && containerName === lastConnectedContainer.value && websocket.value) {
    return;
  }

  try {
    loading.value = true;
    const cluster = useClusterStore().getClusterById(workload.clusterId);
    if (!cluster) {
      console.error("Cluster not found");
      error.value = "Cluster not found";
      loading.value = false;
      return;
    }

    websocket.value = createWebSocket(workload, cluster as DisplayedCluster, podName, containerName);
  } catch (err) {
    console.error("Error connecting terminal:", err);
    error.value = "Failed to connect terminal";
    loading.value = false;
  }
};

const closeTerminal = () => {
  if (websocket.value) {
    websocket.value.close();
    websocket.value = undefined;
  }
};

const handlePodSelected = (podName: string) => {
  emit("pod-selected", podName);
};

const handleContainerSelected = (containerName: string) => {
  emit("container-selected", containerName);
};

// Watch for selection changes from parent to connect terminal
watch(
  () => ({ podName: props.selectedPodName, containerName: props.selectedContainerName }),
  (newSelection) => {
    if (!newSelection.podName || !newSelection.containerName) {
      return;
    }

    // Connect terminal with the new selection
    connectTerminal(newSelection.podName, newSelection.containerName);
  },
  { immediate: true },
);

// Watch for workload changes to reset terminal
watch(
  () => props.workload?.id,
  (newId, oldId) => {
    if (newId === oldId) return;

    closeTerminal();
    error.value = "";
    lastConnectedPod.value = "";
    lastConnectedContainer.value = "";
  },
);

// Cleanup on unmount
onUnmounted(() => {
  closeTerminal();
});
</script>

<style scoped>
.runai-terminal {
  height: 100%;
  width: 100%;
}

.terminal-body {
  height: 100%;
  width: 100%;
  color: var(--text-inverse);
  background-color: var(--bg-code-surface);
}

.error {
  color: red;
  padding: 20px;
}

.troubleshooting-link {
  text-decoration: underline;
}
</style>
