<template>
  <runai-expansion-item label="GPU network acceleration">
    <template #subheader>
      <span class="node-pool-network-subheader">{{ networkSummary }}</span>
    </template>
    <q-card-section class="q-px-sm q-py-md">
      <network-topology-selection
        aid="node-pool-network-topology-selection"
        :network-topology-id="networkTopologyId"
        :cluster-id="clusterId"
        @select="updateNetworkTopologyId"
      />
      <!-- MNNVL DISCOVERY METHOD -->
      <runai-section
        class="q-mb-md"
        title="Define whether MNNVL is present on this node pool:"
        tooltip-text="If this node pool is configured as Multi-Node NVLink (MNNVL) capable, NVIDIA Run:ai leverages its high-bandwidth GPU interconnects to accelerate workloads."
      >
        <runai-radio-options
          aid="node-pool-network-discovery-radio"
          :options="discoveryMethodOptions"
          :model-value="discoveryMethod"
          @update:model-value="discoveryMethodChanged"
        />
      </runai-section>

      <!-- MNNVL LABEL -->
      <runai-section
        v-if="showAccelerationLabelKeyInput"
        class="q-mt-md"
        title="Set the node's label used to detect and enable MNNVL acceleration"
      >
        <div class="row">
          <div class="col-8">
            <runai-name-validation
              aid="node-pool-network-label-input"
              :model-value="gpuNetworkAccelerationLabelKey"
              stack-label
              dense
              :counter="false"
              :maxlength="100"
              :placeholder="gpuNetworkAccelerationLabelKeyPlaceholder"
              label="Label"
              class="q-mb-md"
              :rules="[labelNotEmpty]"
              @update:model-value="$emit(EEmitName.UpdateGpuNetworkAccelerationLabelKey, $event)"
            />
          </div>
        </div>
      </runai-section>

      <div>
        For more information, see the
        <a target="_blank" :href="DOCS.TOPOLOGIES.GB200"> Using GB200 NVL72 and Multi-Node NVLink Domains guide </a>
      </div>
    </q-card-section>
  </runai-expansion-item>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, type PropType, ref } from "vue";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { NodepoolGPUNetworkAccelerationDetection } from "@/swagger-models/cluster-service-client";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import {
  DEFAULT_GPU_NETWORK_ACCELERATION_LABEL_KEY,
  NODE_POOL_NETWORK_DISCOVERY_OPTIONS,
} from "@/models/node-pool.model";
import { RunaiNameValidation } from "@/components/common/runai-name-validation";
import { isNotEmpty } from "@/common/form.validators";
import { errorMessages } from "@/common/error-message.constant";
import { RunaiSection } from "@/components/common/runai-section";
import { NetworkTopologySelection } from "./network-topology-selection";
import { DOCS } from "@/constants/docs";

export enum EEmitName {
  UpdateGpuNetworkAccelerationDetection = "update:gpu-network-acceleration-detection",
  UpdateGpuNetworkAccelerationLabelKey = "update:gpu-network-acceleration-label-key",
  UpdateNetworkTopologyId = "update:network-topology-id",
}

export default defineComponent({
  name: "node-pool-network-section",
  components: {
    RunaiSection,
    RunaiNameValidation,
    RunaiRadioOptions,
    RunaiExpansionItem,
    NetworkTopologySelection,
  },
  emits: {
    [EEmitName.UpdateGpuNetworkAccelerationDetection]: (payload: NodepoolGPUNetworkAccelerationDetection) => payload,
    [EEmitName.UpdateGpuNetworkAccelerationLabelKey]: (payload: string) => payload,
    [EEmitName.UpdateNetworkTopologyId]: (payload: string | null) => typeof payload === "string" || payload === null,
  },
  props: {
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    gpuNetworkAccelerationLabelKey: {
      type: String as PropType<string>,
      default: "",
    },
    gpuNetworkAccelerationDetection: {
      type: [String, null] as PropType<NodepoolGPUNetworkAccelerationDetection | null>,
      required: false,
    },
    networkTopologyId: {
      type: [String as PropType<string>, null] as PropType<string | null>,
      default: null,
    },
  },
  setup(props, { emit }) {
    const discoveryMethod = ref<NodepoolGPUNetworkAccelerationDetection>(
      props.gpuNetworkAccelerationDetection || NodepoolGPUNetworkAccelerationDetection.Auto,
    );

    const networkSummary = computed(() => {
      const detectionStr = NODE_POOL_NETWORK_DISCOVERY_OPTIONS.find(
        (option) => option.value === props.gpuNetworkAccelerationDetection,
      )?.label;

      return discoveryMethod.value === NodepoolGPUNetworkAccelerationDetection.DontUse
        ? detectionStr
        : `${detectionStr} - Label: ${props.gpuNetworkAccelerationLabelKey}`;
    });

    const gpuNetworkAccelerationLabelKeyPlaceholder = DEFAULT_GPU_NETWORK_ACCELERATION_LABEL_KEY;
    const discoveryMethodOptions = computed(() => NODE_POOL_NETWORK_DISCOVERY_OPTIONS);

    const showAccelerationLabelKeyInput = computed(
      () => discoveryMethod.value !== NodepoolGPUNetworkAccelerationDetection.DontUse,
    );

    onMounted(() => {
      if (
        props.gpuNetworkAccelerationDetection === NodepoolGPUNetworkAccelerationDetection.Auto ||
        !props.gpuNetworkAccelerationDetection
      ) {
        discoveryMethod.value = NodepoolGPUNetworkAccelerationDetection.Auto;
        emit(EEmitName.UpdateGpuNetworkAccelerationDetection, NodepoolGPUNetworkAccelerationDetection.Auto);
        emit(EEmitName.UpdateGpuNetworkAccelerationLabelKey, DEFAULT_GPU_NETWORK_ACCELERATION_LABEL_KEY);
      }
    });

    const discoveryMethodChanged = (value: NodepoolGPUNetworkAccelerationDetection) => {
      discoveryMethod.value = value;
      emit(EEmitName.UpdateGpuNetworkAccelerationDetection, value);
    };

    const labelNotEmpty = (val: string): boolean | string => isNotEmpty(val) || errorMessages.LABEL_NOT_EMPTY;

    const updateNetworkTopologyId = (value: string | null) => {
      emit(EEmitName.UpdateNetworkTopologyId, value);
    };

    return {
      discoveryMethod,
      discoveryMethodOptions,
      showAccelerationLabelKeyInput,
      networkSummary,
      EEmitName,
      labelNotEmpty,
      discoveryMethodChanged,
      gpuNetworkAccelerationLabelKeyPlaceholder,
      updateNetworkTopologyId,
      DOCS,
    };
  },
});
</script>

<style scoped lang="scss"></style>
